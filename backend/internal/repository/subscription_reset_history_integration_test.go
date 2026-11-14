//go:build integration

package repository

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func (s *UserSubscriptionRepoSuite) TestResetHistory_MergesOldCardsAndAdminResetsWithPagination() {
	u := s.mustCreateUser("reset-history@test.com", "")
	g := s.mustCreateGroup("reset-history")
	sub := s.mustCreateSubscription(u.ID, g.ID, nil)
	old := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	_, err := s.client.ExecContext(s.ctx, `INSERT INTO user_subscription_reset_cards
		(user_subscription_id, validity_days, source_type, source_reference, created_at, consumed_at, voided_at)
		VALUES ($1, 7, 'payment', 'private-order', $2, $2, $2),
		($1, 30, 'assignment', 'unused', $2, NULL, NULL)`, sub.ID, old)
	s.Require().NoError(err)
	s.Require().NoError(s.repo.IncrementUsage(s.ctx, sub.ID, 2.5, 123))
	now := old.Add(24 * time.Hour)
	s.Require().NoError(s.repo.ResetUsageWindows(s.ctx, sub.ID, true, false, true, now, now))
	items, page, err := s.repo.ListResetHistory(s.ctx, sub.ID, pagination.PaginationParams{Page: 1, PageSize: 1})
	s.Require().NoError(err)
	s.Require().EqualValues(2, page.Total)
	s.Require().Equal(2, page.Pages)
	s.Require().Len(items, 1)
	s.Require().Equal("admin", items[0].Type)
	s.Require().True(items[0].Daily)
	s.Require().False(items[0].Weekly)
	s.Require().True(items[0].Monthly)
	s.Require().Zero(items[0].ValidityDays)
	s.Require().True(now.Equal(items[0].ResetAt))
	items, _, err = s.repo.ListResetHistory(s.ctx, sub.ID, pagination.PaginationParams{Page: 2, PageSize: 1})
	s.Require().NoError(err)
	s.Require().Len(items, 1)
	s.Require().Equal("reset_card", items[0].Type)
	s.Require().Equal(7, items[0].ValidityDays)
	s.Require().True(items[0].Daily && items[0].Weekly && items[0].Monthly)
	updated, err := s.repo.GetByID(s.ctx, sub.ID)
	s.Require().NoError(err)
	s.Require().Zero(updated.DailyUsageUSD)
	s.Require().Zero(updated.MonthlyUsageTokens)
	s.Require().Equal(2.5, updated.WeeklyUsageUSD)
	s.Require().EqualValues(123, updated.WeeklyUsageTokens)
	items, page, err = s.repo.ListResetHistory(s.ctx, sub.ID, pagination.PaginationParams{Page: 3, PageSize: 1})
	s.Require().NoError(err)
	s.Require().Empty(items)
	s.Require().NotNil(items)
	s.Require().EqualValues(2, page.Total)
}

func (s *UserSubscriptionRepoSuite) TestResetHistory_InsertFailureRollsBackQuota() {
	u := s.mustCreateUser("reset-rollback@test.com", "")
	g := s.mustCreateGroup("reset-rollback")
	sub := s.mustCreateSubscription(u.ID, g.ID, nil)
	s.Require().NoError(s.repo.IncrementUsage(s.ctx, sub.ID, 3.25, 50))
	// Force an insert failure in this transaction only; the UPDATE must roll back too.
	_, err := s.client.ExecContext(s.ctx, `ALTER TABLE user_subscription_quota_resets ADD CONSTRAINT test_reject_reset CHECK (user_subscription_id <> `+strconv.FormatInt(sub.ID, 10)+`)`)
	s.Require().NoError(err)
	_, err = s.client.ExecContext(s.ctx, "SAVEPOINT reset_attempt")
	s.Require().NoError(err)
	now := time.Now()
	s.Require().Error(s.repo.ResetUsageWindows(s.ctx, sub.ID, true, true, true, now, now))
	_, err = s.client.ExecContext(s.ctx, "ROLLBACK TO SAVEPOINT reset_attempt")
	s.Require().NoError(err)
	updated, err := s.repo.GetByID(s.ctx, sub.ID)
	s.Require().NoError(err)
	s.Require().Equal(3.25, updated.DailyUsageUSD)
	s.Require().EqualValues(50, updated.MonthlyUsageTokens)
	items, _, err := s.repo.ListResetHistory(s.ctx, sub.ID, pagination.DefaultPagination())
	s.Require().NoError(err)
	s.Require().Empty(items)
}

func TestResetHistory_ConsumptionReplayAndBulkReset(t *testing.T) {
	svc, userID, subscriptionID := createConcurrentResetCardFixture(t, 7)
	t.Cleanup(func() {
		_, err := integrationDB.ExecContext(context.Background(), "DELETE FROM user_subscription_quota_resets WHERE user_subscription_id = $1", subscriptionID)
		require.NoError(t, err)
	})
	ctx := context.Background()
	input := service.ConsumeResetCardInput{UserID: userID, SubscriptionID: subscriptionID, ValidityDays: 7, IdempotencyKey: "history-replay-key"}
	_, err := svc.ConsumeResetCard(ctx, input)
	require.NoError(t, err)
	_, err = svc.ConsumeResetCard(ctx, input)
	require.NoError(t, err)
	result, err := svc.BulkSubscriptionAction(ctx, &service.BulkSubscriptionActionInput{SubscriptionIDs: []int64{subscriptionID, subscriptionID, -1}, Action: "reset_quota", Weekly: true})
	require.Error(t, err)
	require.Nil(t, result)
	result, err = svc.BulkSubscriptionAction(ctx, &service.BulkSubscriptionActionInput{SubscriptionIDs: []int64{subscriptionID, subscriptionID}, Action: "reset_quota", Weekly: true})
	require.NoError(t, err)
	require.Equal(t, 1, result.SuccessCount)
	items, page, err := svc.ListResetHistory(ctx, subscriptionID, &userID, 1, 20)
	require.NoError(t, err)
	require.EqualValues(t, 2, page.Total)
	require.Len(t, items, 2)
	require.Equal(t, "admin", items[0].Type)
	require.Equal(t, "reset_card", items[1].Type)
}
