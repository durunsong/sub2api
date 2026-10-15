//go:build integration

package repository

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestRedeemEmailRepoLifecycleIntegration(t *testing.T) {
	ctx := context.Background()
	repo := NewRedeemEmailRepository(integrationDB)
	suffix := time.Now().UnixNano()
	admin, err := integrationEntClient.User.Create().SetEmail(fmt.Sprintf("redeem-admin-%d@example.com", suffix)).
		SetPasswordHash("x").SetRole("admin").Save(ctx)
	require.NoError(t, err)
	u1, err := integrationEntClient.User.Create().SetEmail(fmt.Sprintf("redeem-a-%d@example.com", suffix)).
		SetPasswordHash("x").SetUsername("alice").Save(ctx)
	require.NoError(t, err)
	u2, err := integrationEntClient.User.Create().SetEmail(fmt.Sprintf("redeem-b-%d@example.com", suffix)).
		SetPasswordHash("x").Save(ctx)
	require.NoError(t, err)
	disabled, err := integrationEntClient.User.Create().SetEmail(fmt.Sprintf("redeem-c-%d@example.com", suffix)).
		SetPasswordHash("x").SetStatus("disabled").Save(ctx)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = integrationDB.Exec(`DELETE FROM redeem_codes WHERE id IN (SELECT i.redeem_code_id FROM redeem_email_items i
			JOIN redeem_email_batches b ON b.id = i.batch_id WHERE b.created_by = $1)`, admin.ID)
		_, _ = integrationDB.Exec(`DELETE FROM redeem_email_batches WHERE created_by = $1`, admin.ID)
		_, _ = integrationDB.Exec(`DELETE FROM users WHERE id = ANY($1::bigint[])`, pq.Array([]int64{admin.ID, u1.ID, u2.ID, disabled.ID}))
	})

	now := time.Now().UTC()
	cutoffs := service.NewRedeemEmailCutoffs(now)
	search := fmt.Sprintf("-%d@", suffix)
	withNever := []string{service.RedeemEmailSegmentInactive1To2Months, service.RedeemEmailSegmentNeverUsed}
	found, err := repo.SearchRecipients(ctx, withNever, search, cutoffs, 10)
	require.NoError(t, err)
	ids := map[int64]bool{}
	for _, f := range found {
		ids[f.UserID] = true
		require.Equal(t, service.RedeemEmailSegmentNeverUsed, f.Segment)
	}
	require.True(t, ids[u1.ID] && ids[u2.ID])
	require.False(t, ids[disabled.ID] || ids[admin.ID], "disabled users and admins are not eligible")

	none, err := repo.SearchRecipients(ctx, service.RedeemEmailDefaultSegments(), search, cutoffs, 10)
	require.NoError(t, err)
	require.Empty(t, none, "never-used users are excluded unless the never_used segment is selected")

	expires := now.AddDate(0, 0, 30)
	p := service.CreateRedeemEmailBatchParams{
		CreatedBy: admin.ID, IdempotencyKeyHash: strings.Repeat("d", 64), RequestHash: strings.Repeat("e", 64),
		Segments: withNever, Cutoffs: cutoffs, UserIDs: []int64{u2.ID, u1.ID}, CouponValue: 1.5, CodeExpiresAt: &expires, Now: now,
	}
	batch, replayed, err := repo.CreateBatch(ctx, p)
	require.NoError(t, err)
	require.False(t, replayed)
	require.Equal(t, 2, batch.TotalCount)
	require.Equal(t, 2, batch.PendingCount)
	require.Equal(t, withNever, batch.Segments)

	again, replayed, err := repo.CreateBatch(ctx, p)
	require.NoError(t, err)
	require.True(t, replayed)
	require.Equal(t, batch.ID, again.ID)

	conflict := p
	conflict.RequestHash = strings.Repeat("f", 64)
	_, _, err = repo.CreateBatch(ctx, conflict)
	require.ErrorIs(t, err, service.ErrRedeemEmailBatchConflict)

	ineligible := p
	ineligible.IdempotencyKeyHash = strings.Repeat("9", 64)
	ineligible.UserIDs = []int64{disabled.ID}
	_, _, err = repo.CreateBatch(ctx, ineligible)
	require.ErrorIs(t, err, service.ErrRedeemEmailRecipientIneligible)

	items, page, err := repo.ListItems(ctx, batch.ID, "", pagination.PaginationParams{Page: 1, PageSize: 10})
	require.NoError(t, err)
	require.EqualValues(t, 2, page.Total)
	require.NotEqual(t, items[0].RedeemCode, items[1].RedeemCode)
	var codeValue float64
	var notes string
	require.NoError(t, integrationDB.QueryRow(`SELECT value, notes FROM redeem_codes WHERE id = $1`, *items[0].RedeemCodeID).Scan(&codeValue, &notes))
	require.InDelta(t, 1.5, codeValue, 1e-9)
	require.Equal(t, fmt.Sprintf("redeem_email_batch:%d", batch.ID), notes)

	first, err := repo.ClaimNext(ctx, now, 45*time.Second)
	require.NoError(t, err)
	require.NotNil(t, first)
	require.InDelta(t, 1.5, first.CouponValue, 1e-9)
	ok, err := repo.IsCodeAvailable(ctx, first.Item.RedeemCodeID, now)
	require.NoError(t, err)
	require.True(t, ok)
	require.NoError(t, repo.CompleteAttempt(ctx, service.RedeemEmailAttemptResult{ItemID: first.Item.ID, AttemptToken: *first.Item.AttemptToken, Status: service.RedeemEmailItemStatusSent, CompletedAt: now}))

	second, err := repo.ClaimNext(ctx, now, time.Second)
	require.NoError(t, err)
	require.NotNil(t, second)
	recovered, err := repo.RecoverExpiredSending(ctx, now.Add(time.Minute), 10)
	require.NoError(t, err)
	require.Equal(t, 1, recovered)

	got, err := repo.GetBatch(ctx, batch.ID)
	require.NoError(t, err)
	require.Equal(t, service.RedeemEmailBatchStatusPartialFailed, got.Status)
	require.Equal(t, 1, got.SentCount)
	require.Equal(t, 1, got.FailedCount)

	retried, err := repo.RetryFailed(ctx, batch.ID, now)
	require.NoError(t, err)
	require.Equal(t, 1, retried.PendingCount)
	require.Equal(t, 0, retried.FailedCount)
	require.ErrorIs(t, repo.DeleteBatch(ctx, batch.ID), service.ErrRedeemEmailBatchNotDeletable)

	third, err := repo.ClaimNext(ctx, now, 45*time.Second)
	require.NoError(t, err)
	require.NotNil(t, third)
	require.NoError(t, repo.CompleteAttempt(ctx, service.RedeemEmailAttemptResult{ItemID: third.Item.ID, AttemptToken: *third.Item.AttemptToken,
		Status: service.RedeemEmailItemStatusSkipped, ErrorCode: "RECIPIENT_UNSUBSCRIBED", CompletedAt: now}))
	done, err := repo.GetBatch(ctx, batch.ID)
	require.NoError(t, err)
	require.Equal(t, service.RedeemEmailBatchStatusSucceeded, done.Status, "skipped recipients do not count as failures")
	require.Equal(t, 1, done.SkippedCount)
}
