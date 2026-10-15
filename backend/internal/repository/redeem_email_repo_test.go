package repository

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

var redeemEmailBatchColumnNames = []string{"id", "status", "segments", "reference_time", "coupon_value", "code_expires_at",
	"total_count", "pending_count", "sending_count", "sent_count", "failed_count", "skipped_count",
	"created_by", "started_at", "finished_at", "created_at", "updated_at"}

var redeemEmailRecipientColumnNames = []string{"id", "email", "username", "last_used_at", "segment"}

func newRedeemEmailRepoMock(t *testing.T) (*redeemEmailRepository, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return NewRedeemEmailRepository(db).(*redeemEmailRepository), mock
}

func redeemEmailCreateParams() service.CreateRedeemEmailBatchParams {
	now := time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)
	return service.CreateRedeemEmailBatchParams{
		CreatedBy:          9,
		IdempotencyKeyHash: strings.Repeat("a", 64),
		RequestHash:        strings.Repeat("b", 64),
		Segments:           []string{service.RedeemEmailSegmentInactive1To2Months, service.RedeemEmailSegmentNeverUsed},
		Cutoffs:            service.NewRedeemEmailCutoffs(now),
		UserIDs:            []int64{3, 7},
		CouponValue:        5,
		Now:                now,
	}
}

func redeemEmailBatchRow(p service.CreateRedeemEmailBatchParams, extra ...any) *sqlmock.Rows {
	cols := redeemEmailBatchColumnNames
	if len(extra) > 0 {
		cols = append(append([]string(nil), cols...), "request_hash")
	}
	values := []driver.Value{int64(11), "pending", "{inactive_1_2m,never_used}", p.Now, "5.00000000", nil,
		2, 2, 0, 0, 0, 0, p.CreatedBy, nil, nil, p.Now, p.Now}
	for _, v := range extra {
		values = append(values, v)
	}
	return sqlmock.NewRows(cols).AddRow(values...)
}

func TestRedeemEmailRepoSearchPassesCutoffsAndSegments(t *testing.T) {
	repo, mock := newRedeemEmailRepoMock(t)
	p := redeemEmailCreateParams()
	last := p.Now.AddDate(0, -1, -3)
	mock.ExpectQuery(regexp.QuoteMeta("WHERE s.segment = ANY($4::text[])")).
		WithArgs(p.Cutoffs.OneMonth, p.Cutoffs.TwoMonths, p.Cutoffs.ThreeMonths, sqlmock.AnyArg(), "%a\\_b%", 1001).
		WillReturnRows(sqlmock.NewRows(redeemEmailRecipientColumnNames).
			AddRow(int64(3), "a@example.com", "a", nil, "never_used").
			AddRow(int64(7), "b@example.com", "", last, "inactive_1_2m"))

	got, err := repo.SearchRecipients(context.Background(), p.Segments, " a_b ", p.Cutoffs, 0)
	require.NoError(t, err)
	require.Len(t, got, 2)
	require.Equal(t, "never_used", got[0].Segment)
	require.Nil(t, got[0].LastUsedAt)
	require.Equal(t, "inactive_1_2m", got[1].Segment)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRedeemEmailRepoCreateBatchReplaysSameRequest(t *testing.T) {
	repo, mock := newRedeemEmailRepoMock(t)
	p := redeemEmailCreateParams()
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("SELECT pg_advisory_xact_lock($1)")).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("FROM redeem_email_batches WHERE created_by").WithArgs(p.CreatedBy, p.IdempotencyKeyHash).
		WillReturnRows(redeemEmailBatchRow(p, p.RequestHash))
	mock.ExpectRollback()

	batch, replayed, err := repo.CreateBatch(context.Background(), p)
	require.NoError(t, err)
	require.True(t, replayed)
	require.Equal(t, int64(11), batch.ID)
	require.Equal(t, []string{"inactive_1_2m", "never_used"}, batch.Segments)
	require.InDelta(t, 5.0, batch.CouponValue, 1e-9)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRedeemEmailRepoCreateBatchRejectsDifferentRequestForSameKey(t *testing.T) {
	repo, mock := newRedeemEmailRepoMock(t)
	p := redeemEmailCreateParams()
	mock.ExpectBegin()
	mock.ExpectExec("pg_advisory_xact_lock").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("FROM redeem_email_batches WHERE created_by").WillReturnRows(redeemEmailBatchRow(p, strings.Repeat("c", 64)))
	mock.ExpectRollback()

	_, _, err := repo.CreateBatch(context.Background(), p)
	require.ErrorIs(t, err, service.ErrRedeemEmailBatchConflict)
	require.NoError(t, mock.ExpectationsWereMet())
}

func expectRedeemEmailRecipientLock(mock sqlmock.Sqlmock, p service.CreateRedeemEmailBatchParams, rows *sqlmock.Rows) {
	mock.ExpectExec("FOR SHARE OF u").WillReturnResult(sqlmock.NewResult(0, int64(len(p.UserIDs))))
	mock.ExpectQuery(regexp.QuoteMeta("u.id = ANY($5::bigint[])")).
		WithArgs(p.Cutoffs.OneMonth, p.Cutoffs.TwoMonths, p.Cutoffs.ThreeMonths, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(rows)
}

func TestRedeemEmailRepoCreateBatchRejectsRecipientsThatLeftSelectedSegments(t *testing.T) {
	repo, mock := newRedeemEmailRepoMock(t)
	p := redeemEmailCreateParams()
	mock.ExpectBegin()
	mock.ExpectExec("pg_advisory_xact_lock").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("FROM redeem_email_batches WHERE created_by").WillReturnError(sql.ErrNoRows)
	expectRedeemEmailRecipientLock(mock, p, sqlmock.NewRows(redeemEmailRecipientColumnNames).
		AddRow(int64(3), "a@example.com", "a", nil, "never_used"))
	mock.ExpectRollback()

	_, _, err := repo.CreateBatch(context.Background(), p)
	require.ErrorIs(t, err, service.ErrRedeemEmailRecipientIneligible)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRedeemEmailRepoCreateBatchGeneratesOneCodePerRecipientInOneTransaction(t *testing.T) {
	repo, mock := newRedeemEmailRepoMock(t)
	p := redeemEmailCreateParams()
	mock.ExpectBegin()
	mock.ExpectExec("pg_advisory_xact_lock").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("FROM redeem_email_batches WHERE created_by").WillReturnError(sql.ErrNoRows)
	expectRedeemEmailRecipientLock(mock, p, sqlmock.NewRows(redeemEmailRecipientColumnNames).
		AddRow(int64(3), "a@example.com", "", nil, "never_used").
		AddRow(int64(7), "b@example.com", "bob", p.Now.AddDate(0, -1, -1), "inactive_1_2m"))
	mock.ExpectQuery("INSERT INTO redeem_email_batches").
		WithArgs(sqlmock.AnyArg(), p.Now, "5.00000000", nil, 2, p.CreatedBy, p.IdempotencyKeyHash, p.RequestHash).
		WillReturnRows(redeemEmailBatchRow(p))
	// 第一轮仅返回一行，模拟撞码被 ON CONFLICT 忽略，需要第二轮补发。
	mock.ExpectQuery("INSERT INTO redeem_codes").WithArgs(sqlmock.AnyArg(), "5.00000000", "redeem_email_batch:11", p.Now, nil).
		WillReturnRows(sqlmock.NewRows([]string{"id", "code"}).AddRow(int64(100), "CODEA"))
	mock.ExpectQuery("INSERT INTO redeem_codes").WillReturnRows(sqlmock.NewRows([]string{"id", "code"}).AddRow(int64(101), "CODEB"))
	mock.ExpectExec("INSERT INTO redeem_email_items").WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectCommit()

	batch, replayed, err := repo.CreateBatch(context.Background(), p)
	require.NoError(t, err)
	require.False(t, replayed)
	require.Equal(t, 2, batch.TotalCount)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRedeemEmailRepoDeleteBatchRejectsActiveBatch(t *testing.T) {
	repo, mock := newRedeemEmailRepoMock(t)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT status FROM redeem_email_batches").WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("running"))
	mock.ExpectRollback()

	require.ErrorIs(t, repo.DeleteBatch(context.Background(), 5), service.ErrRedeemEmailBatchNotDeletable)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRedeemEmailRepoGetBatchNotFound(t *testing.T) {
	repo, mock := newRedeemEmailRepoMock(t)
	mock.ExpectQuery("FROM redeem_email_batches WHERE id").WillReturnError(sql.ErrNoRows)

	_, err := repo.GetBatch(context.Background(), 404)
	require.ErrorIs(t, err, service.ErrRedeemEmailBatchNotFound)
}

func TestRedeemEmailRepoCompleteAttemptIgnoresStaleToken(t *testing.T) {
	repo, mock := newRedeemEmailRepoMock(t)
	token := uuid.New()
	mock.ExpectBegin()
	mock.ExpectQuery("UPDATE redeem_email_items SET status").WithArgs(int64(1), token.String(), "sent", true, sqlmock.AnyArg(), "", "").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	require.NoError(t, repo.CompleteAttempt(context.Background(), service.RedeemEmailAttemptResult{ItemID: 1, AttemptToken: token, Status: "sent", CompletedAt: time.Now()}))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRedeemEmailRepoCompleteAttemptCountsSkippedSeparately(t *testing.T) {
	repo, mock := newRedeemEmailRepoMock(t)
	token := uuid.New()
	now := time.Now()
	mock.ExpectBegin()
	mock.ExpectQuery("UPDATE redeem_email_items SET status").
		WithArgs(int64(1), token.String(), "skipped", false, now, "RECIPIENT_UNSUBSCRIBED", "unsubscribed").
		WillReturnRows(sqlmock.NewRows([]string{"batch_id"}).AddRow(int64(9)))
	mock.ExpectExec("UPDATE redeem_email_batches SET sending_count").WithArgs(int64(9), 0, 0, 1, now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	require.NoError(t, repo.CompleteAttempt(context.Background(), service.RedeemEmailAttemptResult{
		ItemID: 1, AttemptToken: token, Status: "skipped", ErrorCode: "RECIPIENT_UNSUBSCRIBED", ErrorMessage: "unsubscribed", CompletedAt: now,
	}))
	require.NoError(t, mock.ExpectationsWereMet())

	require.ErrorIs(t, repo.CompleteAttempt(context.Background(), service.RedeemEmailAttemptResult{Status: "pending"}), service.ErrRedeemEmailItemStatusInvalid)
}

func TestRedeemEmailRepoClaimNextReturnsNilWhenQueueEmpty(t *testing.T) {
	repo, mock := newRedeemEmailRepoMock(t)
	mock.ExpectQuery("WITH candidate AS").WillReturnError(sql.ErrNoRows)

	claim, err := repo.ClaimNext(context.Background(), time.Now(), 0)
	require.NoError(t, err)
	require.Nil(t, claim)
}

func TestRedeemEmailRepoSQLKeepsEligibilityAndCastContracts(t *testing.T) {
	for _, sqlText := range []string{redeemEmailSearchSQL, redeemEmailSelectedUsersSQL} {
		for _, want := range []string{"u.deleted_at IS NULL", "u.status = 'active'", "u.role = 'user'", "NOT LIKE '%.invalid'",
			"WHEN r.last_used_at IS NULL THEN 'never_used'",
			"WHEN r.last_used_at >= $1::timestamptz THEN 'active'",
			"WHEN r.last_used_at >= $2::timestamptz THEN 'inactive_1_2m'",
			"WHEN r.last_used_at >= $3::timestamptz THEN 'inactive_2_3m'",
			"ELSE 'inactive_3m_plus'",
			"s.segment = ANY($4::text[])"} {
			require.Contains(t, sqlText, want)
		}
		require.NotContains(t, sqlText, "%s")
	}
	require.Contains(t, redeemEmailSearchSQL, "LIMIT $6")
	require.Contains(t, redeemEmailCompleteBatchSQL, "skipped_count = skipped_count + $4")
	require.Contains(t, redeemEmailClaimNextSQL, "FOR UPDATE OF i SKIP LOCKED")
	require.Contains(t, redeemEmailClaimNextSQL, "make_interval(secs => $3::double precision)")
	require.Contains(t, redeemEmailRecoverExpiredSQL, "DELIVERY_OUTCOME_UNKNOWN")
}
