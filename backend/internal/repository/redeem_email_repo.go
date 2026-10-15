package repository

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type redeemEmailRepository struct {
	db *sql.DB
}

func NewRedeemEmailRepository(db *sql.DB) service.RedeemEmailRepository {
	return &redeemEmailRepository{db: db}
}

const redeemEmailBaseUserSQL = `u.deleted_at IS NULL
	AND u.status = 'active'
	AND u.role = 'user'
	AND BTRIM(u.email) <> ''
	AND LOWER(u.email) NOT LIKE '%.invalid'`

// redeemEmailSegmentedUsersSQL 按最后一次 usage_logs 时间把用户归入互斥分段：
// $1/$2/$3 为 1/2/3 个月前的边界，$4 为选中分段；%s 处拼接额外的用户过滤条件。
// 'active'（近 1 个月有用量）不在可选分段中，因此永远不会入选。
const redeemEmailSegmentedUsersSQL = `SELECT s.id, s.email, s.username, s.last_used_at, s.segment FROM (
	SELECT r.id, r.email, r.username, r.last_used_at,
		CASE WHEN r.last_used_at IS NULL THEN 'never_used'
			WHEN r.last_used_at >= $1::timestamptz THEN 'active'
			WHEN r.last_used_at >= $2::timestamptz THEN 'inactive_1_2m'
			WHEN r.last_used_at >= $3::timestamptz THEN 'inactive_2_3m'
			ELSE 'inactive_3m_plus' END AS segment
	FROM (
		SELECT u.id, BTRIM(u.email) AS email, COALESCE(u.username, '') AS username,
			(SELECT MAX(l.created_at) FROM usage_logs l WHERE l.user_id = u.id) AS last_used_at
		FROM users u
		WHERE ` + redeemEmailBaseUserSQL + ` AND %s
	) r
) s
WHERE s.segment = ANY($4::text[])`

var (
	redeemEmailSearchSQL = strings.Replace(redeemEmailSegmentedUsersSQL, "%s",
		`($5::text = '' OR u.email ILIKE $5::text ESCAPE '\' OR u.username ILIKE $5::text ESCAPE '\')`, 1) + `
ORDER BY s.last_used_at ASC NULLS FIRST, s.id ASC
LIMIT $6`
	redeemEmailSelectedUsersSQL = strings.Replace(redeemEmailSegmentedUsersSQL, "%s", `u.id = ANY($5::bigint[])`, 1) + `
ORDER BY s.id ASC`
)

func redeemEmailCutoffArgs(c service.RedeemEmailCutoffs, segments []string) []any {
	return []any{c.OneMonth, c.TwoMonths, c.ThreeMonths, pq.Array(segments)}
}

func scanRedeemEmailRecipients(rows *sql.Rows) ([]service.RedeemEmailRecipient, error) {
	defer func() { _ = rows.Close() }()
	out := make([]service.RedeemEmailRecipient, 0)
	for rows.Next() {
		var x service.RedeemEmailRecipient
		if err := rows.Scan(&x.UserID, &x.Email, &x.Username, &x.LastUsedAt, &x.Segment); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func (r *redeemEmailRepository) SearchRecipients(ctx context.Context, segments []string, search string, cutoffs service.RedeemEmailCutoffs, limit int) ([]service.RedeemEmailRecipient, error) {
	if limit <= 0 || limit > service.RedeemEmailMaxRecipients+1 {
		limit = service.RedeemEmailMaxRecipients + 1
	}
	pattern := ""
	if s := strings.TrimSpace(search); s != "" {
		pattern = "%" + escapeLikeWildcards(s) + "%"
	}
	rows, err := r.db.QueryContext(ctx, redeemEmailSearchSQL, append(redeemEmailCutoffArgs(cutoffs, segments), pattern, limit)...)
	if err != nil {
		return nil, err
	}
	return scanRedeemEmailRecipients(rows)
}

const redeemEmailBatchColumns = `id, status, segments, reference_time, coupon_value, code_expires_at,
	total_count, pending_count, sending_count, sent_count, failed_count, skipped_count,
	created_by, started_at, finished_at, created_at, updated_at`

type redeemEmailRowScanner interface {
	Scan(dest ...any) error
}

func scanRedeemEmailBatch(row redeemEmailRowScanner, extra ...any) (*service.RedeemEmailBatch, error) {
	var b service.RedeemEmailBatch
	dest := []any{&b.ID, &b.Status, pq.Array(&b.Segments), &b.ReferenceTime, &b.CouponValue, &b.CodeExpiresAt,
		&b.TotalCount, &b.PendingCount, &b.SendingCount, &b.SentCount, &b.FailedCount, &b.SkippedCount,
		&b.CreatedBy, &b.StartedAt, &b.FinishedAt, &b.CreatedAt, &b.UpdatedAt}
	if err := row.Scan(append(dest, extra...)...); err != nil {
		return nil, err
	}
	return &b, nil
}

const redeemEmailItemColumns = `id, batch_id, position, user_id, recipient_name, recipient_email, redeem_code_id, redeem_code,
	status, attempt_count, sent_at, last_attempt_at, error_code, error_message, created_at, updated_at`

func scanRedeemEmailItem(row redeemEmailRowScanner) (service.RedeemEmailItem, error) {
	var x service.RedeemEmailItem
	err := row.Scan(&x.ID, &x.BatchID, &x.Position, &x.UserID, &x.RecipientName, &x.RecipientEmail, &x.RedeemCodeID, &x.RedeemCode,
		&x.Status, &x.AttemptCount, &x.SentAt, &x.LastAttemptAt, &x.ErrorCode, &x.ErrorMessage, &x.CreatedAt, &x.UpdatedAt)
	return x, err
}

const redeemEmailCodeInsertMaxRounds = 5

func (r *redeemEmailRepository) CreateBatch(ctx context.Context, p service.CreateRedeemEmailBatchParams) (*service.RedeemEmailBatch, bool, error) {
	if len(p.UserIDs) == 0 {
		return nil, false, service.ErrRedeemEmailRecipientInvalid
	}
	if len(p.UserIDs) > service.RedeemEmailMaxRecipients {
		return nil, false, service.ErrRedeemEmailRecipientLimitExceeded
	}
	if len(p.IdempotencyKeyHash) != 64 || len(p.RequestHash) != 64 || p.CreatedBy <= 0 {
		return nil, false, service.ErrRedeemEmailIdempotencyKeyInvalid
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, false, err
	}
	defer func() { _ = tx.Rollback() }()

	lockKey := "redeem-email-create:" + strconv.FormatInt(p.CreatedBy, 10) + ":" + p.IdempotencyKeyHash
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1)`, advisoryLockHash(lockKey)); err != nil {
		return nil, false, err
	}
	var existingHash string
	existing, err := scanRedeemEmailBatch(tx.QueryRowContext(ctx,
		`SELECT `+redeemEmailBatchColumns+`, request_hash FROM redeem_email_batches WHERE created_by = $1 AND idempotency_key_hash = $2`,
		p.CreatedBy, p.IdempotencyKeyHash), &existingHash)
	if err == nil {
		if existingHash != p.RequestHash {
			return nil, false, service.ErrRedeemEmailBatchConflict
		}
		return existing, true, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, false, err
	}

	recipients, err := lockEligibleRedeemEmailRecipients(ctx, tx, p)
	if err != nil {
		return nil, false, err
	}

	total := len(recipients)
	batch, err := scanRedeemEmailBatch(tx.QueryRowContext(ctx, `INSERT INTO redeem_email_batches
		(status, segments, reference_time, coupon_value, code_expires_at,
		 total_count, pending_count, created_by, idempotency_key_hash, request_hash, created_at, updated_at)
		VALUES ('pending', $1::text[], $2, $3::numeric, $4, $5, $5, $6, $7, $8, $2, $2)
		RETURNING `+redeemEmailBatchColumns,
		pq.Array(p.Segments), p.Now, redeemEmailNumeric(p.CouponValue), p.CodeExpiresAt,
		total, p.CreatedBy, p.IdempotencyKeyHash, p.RequestHash))
	if err != nil {
		if isUniqueViolation(err) {
			return nil, false, service.ErrRedeemEmailBatchConflict
		}
		return nil, false, err
	}

	codes, codeIDs, err := insertRedeemEmailBalanceCodes(ctx, tx, batch.ID, total, p)
	if err != nil {
		return nil, false, err
	}
	positions := make([]int64, total)
	userIDs := make([]int64, total)
	names := make([]string, total)
	emails := make([]string, total)
	for i, rc := range recipients {
		positions[i] = int64(i)
		userIDs[i] = rc.UserID
		names[i] = service.RedeemEmailRecipientName(rc.Username, rc.Email)
		emails[i] = strings.TrimSpace(rc.Email)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO redeem_email_items
		(batch_id, position, user_id, recipient_name, recipient_email, redeem_code_id, redeem_code, status, created_at, updated_at)
		SELECT $1, x.position, x.user_id, x.recipient_name, x.recipient_email, x.redeem_code_id, x.redeem_code, 'pending', $8, $8
		FROM unnest($2::int[], $3::bigint[], $4::text[], $5::text[], $6::bigint[], $7::text[])
			AS x(position, user_id, recipient_name, recipient_email, redeem_code_id, redeem_code)`,
		batch.ID, pq.Array(positions), pq.Array(userIDs), pq.Array(names), pq.Array(emails), pq.Array(codeIDs), pq.Array(codes), p.Now); err != nil {
		return nil, false, err
	}
	if err := tx.Commit(); err != nil {
		return nil, false, err
	}
	return batch, false, nil
}

// lockEligibleRedeemEmailRecipients 先锁定所选用户行防止并发改状态，再以与搜索相同的分段谓词复核。
func lockEligibleRedeemEmailRecipients(ctx context.Context, tx *sql.Tx, p service.CreateRedeemEmailBatchParams) ([]service.RedeemEmailRecipient, error) {
	if _, err := tx.ExecContext(ctx, `SELECT u.id FROM users u WHERE u.id = ANY($1::bigint[]) ORDER BY u.id ASC FOR SHARE OF u`,
		pq.Array(p.UserIDs)); err != nil {
		return nil, err
	}
	rows, err := tx.QueryContext(ctx, redeemEmailSelectedUsersSQL, append(redeemEmailCutoffArgs(p.Cutoffs, p.Segments), pq.Array(p.UserIDs))...)
	if err != nil {
		return nil, err
	}
	out, err := scanRedeemEmailRecipients(rows)
	if err != nil {
		return nil, err
	}
	if len(out) != len(p.UserIDs) {
		return nil, service.ErrRedeemEmailRecipientIneligible
	}
	return out, nil
}

// insertRedeemEmailBalanceCodes 为每位收件人生成独立余额码；ON CONFLICT DO NOTHING 让偶发撞码只需补发而不中止事务。
func insertRedeemEmailBalanceCodes(ctx context.Context, tx *sql.Tx, batchID int64, n int, p service.CreateRedeemEmailBatchParams) ([]string, []int64, error) {
	notes := "redeem_email_batch:" + strconv.FormatInt(batchID, 10)
	codes := make([]string, 0, n)
	ids := make([]int64, 0, n)
	for round := 0; round < redeemEmailCodeInsertMaxRounds && len(codes) < n; round++ {
		want := n - len(codes)
		candidates := make([]string, want)
		for i := range candidates {
			c, err := service.GenerateRedeemCode()
			if err != nil {
				return nil, nil, err
			}
			candidates[i] = c
		}
		rows, err := tx.QueryContext(ctx, `INSERT INTO redeem_codes (code, type, value, status, notes, created_at, expires_at)
			SELECT c, 'balance', $2::numeric, 'unused', $3, $4, $5 FROM unnest($1::text[]) AS t(c)
			ON CONFLICT (code) DO NOTHING
			RETURNING id, code`, pq.Array(candidates), redeemEmailNumeric(p.CouponValue), notes, p.Now, p.CodeExpiresAt)
		if err != nil {
			return nil, nil, err
		}
		for rows.Next() {
			var id int64
			var code string
			if err := rows.Scan(&id, &code); err != nil {
				_ = rows.Close()
				return nil, nil, err
			}
			ids = append(ids, id)
			codes = append(codes, code)
		}
		if err := rows.Close(); err != nil {
			return nil, nil, err
		}
		if err := rows.Err(); err != nil {
			return nil, nil, err
		}
	}
	if len(codes) != n {
		return nil, nil, errors.New("redeem email: failed to generate unique redeem codes")
	}
	return codes, ids, nil
}

func redeemEmailNumeric(v float64) string {
	return strconv.FormatFloat(v, 'f', 8, 64)
}

func (r *redeemEmailRepository) ListBatches(ctx context.Context, p pagination.PaginationParams) ([]service.RedeemEmailBatch, *pagination.PaginationResult, error) {
	var total int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM redeem_email_batches`).Scan(&total); err != nil {
		return nil, nil, err
	}
	rows, err := r.db.QueryContext(ctx, `SELECT `+redeemEmailBatchColumns+` FROM redeem_email_batches
		ORDER BY created_at DESC, id DESC OFFSET $1 LIMIT $2`, p.Offset(), p.Limit())
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = rows.Close() }()
	out := make([]service.RedeemEmailBatch, 0)
	for rows.Next() {
		b, err := scanRedeemEmailBatch(rows)
		if err != nil {
			return nil, nil, err
		}
		out = append(out, *b)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	return out, paginationResultFromTotal(total, p), nil
}

func (r *redeemEmailRepository) GetBatch(ctx context.Context, id int64) (*service.RedeemEmailBatch, error) {
	b, err := scanRedeemEmailBatch(r.db.QueryRowContext(ctx, `SELECT `+redeemEmailBatchColumns+` FROM redeem_email_batches WHERE id = $1`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrRedeemEmailBatchNotFound
	}
	return b, err
}

func (r *redeemEmailRepository) ListItems(ctx context.Context, batchID int64, status string, p pagination.PaginationParams) ([]service.RedeemEmailItem, *pagination.PaginationResult, error) {
	if _, err := r.GetBatch(ctx, batchID); err != nil {
		return nil, nil, err
	}
	var total int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM redeem_email_items WHERE batch_id = $1 AND ($2::text = '' OR status = $2::text)`,
		batchID, status).Scan(&total); err != nil {
		return nil, nil, err
	}
	rows, err := r.db.QueryContext(ctx, `SELECT `+redeemEmailItemColumns+` FROM redeem_email_items
		WHERE batch_id = $1 AND ($2::text = '' OR status = $2::text)
		ORDER BY position ASC, id ASC OFFSET $3 LIMIT $4`, batchID, status, p.Offset(), p.Limit())
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = rows.Close() }()
	out := make([]service.RedeemEmailItem, 0)
	for rows.Next() {
		x, err := scanRedeemEmailItem(rows)
		if err != nil {
			return nil, nil, err
		}
		out = append(out, x)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	return out, paginationResultFromTotal(total, p), nil
}

// RetryFailed 只把失败明细放回队列；重复调用时已无失败明细，天然幂等。兑换码沿用原码，Worker 发送前会复核其可用性。
func (r *redeemEmailRepository) RetryFailed(ctx context.Context, batchID int64, now time.Time) (*service.RedeemEmailBatch, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	var locked int64
	if err := tx.QueryRowContext(ctx, `SELECT id FROM redeem_email_batches WHERE id = $1 FOR UPDATE`, batchID).Scan(&locked); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, service.ErrRedeemEmailBatchNotFound
		}
		return nil, err
	}
	res, err := tx.ExecContext(ctx, `UPDATE redeem_email_items SET status = 'pending', attempt_token = NULL, lease_expires_at = NULL,
		error_code = NULL, error_message = NULL, updated_at = $2
		WHERE batch_id = $1 AND status = 'failed'`, batchID, now)
	if err != nil {
		return nil, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if n > 0 {
		if _, err := tx.ExecContext(ctx, `UPDATE redeem_email_batches SET pending_count = pending_count + $2, failed_count = failed_count - $2,
			status = CASE WHEN sending_count > 0 THEN 'running' ELSE 'pending' END, finished_at = NULL, updated_at = $3
			WHERE id = $1`, batchID, n, now); err != nil {
			return nil, err
		}
	}
	b, err := scanRedeemEmailBatch(tx.QueryRowContext(ctx, `SELECT `+redeemEmailBatchColumns+` FROM redeem_email_batches WHERE id = $1`, batchID))
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return b, nil
}

func (r *redeemEmailRepository) DeleteBatch(ctx context.Context, batchID int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	var status string
	if err := tx.QueryRowContext(ctx, `SELECT status FROM redeem_email_batches WHERE id = $1 FOR UPDATE`, batchID).Scan(&status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return service.ErrRedeemEmailBatchNotFound
		}
		return err
	}
	if !service.RedeemEmailBatchCanDelete(status) {
		return service.ErrRedeemEmailBatchNotDeletable
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM redeem_email_batches WHERE id = $1`, batchID); err != nil {
		return err
	}
	return tx.Commit()
}

// redeemEmailClaimNextSQL 固定参数类型：$1 为时钟、$3 为租约秒数，缺少类型转换时 PostgreSQL 无法推断区间运算。
const redeemEmailClaimNextSQL = `
WITH candidate AS (
	SELECT i.id
	FROM redeem_email_items i
	JOIN redeem_email_batches b ON b.id = i.batch_id
	WHERE i.status = 'pending'
	ORDER BY b.created_at ASC, b.id ASC, i.position ASC, i.id ASC
	LIMIT 1
	FOR UPDATE OF i SKIP LOCKED
), claimed AS (
	UPDATE redeem_email_items i SET status = 'sending', attempt_count = attempt_count + 1,
		attempt_token = $2::uuid,
		lease_expires_at = $1::timestamptz + make_interval(secs => $3::double precision),
		last_attempt_at = $1::timestamptz,
		error_code = NULL, error_message = NULL, updated_at = $1::timestamptz
	FROM candidate c WHERE i.id = c.id
	RETURNING i.id, i.batch_id, i.position, i.user_id, i.recipient_name, i.recipient_email, i.redeem_code_id, i.redeem_code,
		i.attempt_count, i.attempt_token, i.lease_expires_at, i.last_attempt_at
), batch_updated AS (
	UPDATE redeem_email_batches b SET status = 'running', started_at = COALESCE(started_at, $1::timestamptz),
		pending_count = pending_count - 1, sending_count = sending_count + 1, updated_at = $1::timestamptz
	FROM claimed c WHERE b.id = c.batch_id
	RETURNING b.id, b.coupon_value, b.code_expires_at
)
SELECT c.id, c.batch_id, c.position, c.user_id, c.recipient_name, c.recipient_email, c.redeem_code_id, c.redeem_code,
	c.attempt_count, c.attempt_token, c.lease_expires_at, c.last_attempt_at,
	b.coupon_value, b.code_expires_at
FROM claimed c JOIN batch_updated b ON b.id = c.batch_id`

func (r *redeemEmailRepository) ClaimNext(ctx context.Context, now time.Time, lease time.Duration) (*service.RedeemEmailClaim, error) {
	if lease <= 0 {
		lease = 45 * time.Second
	}
	var c service.RedeemEmailClaim
	err := r.db.QueryRowContext(ctx, redeemEmailClaimNextSQL, now, uuid.New().String(), lease.Seconds()).Scan(
		&c.Item.ID, &c.Item.BatchID, &c.Item.Position, &c.Item.UserID, &c.Item.RecipientName, &c.Item.RecipientEmail,
		&c.Item.RedeemCodeID, &c.Item.RedeemCode, &c.Item.AttemptCount, &c.Item.AttemptToken, &c.Item.LeaseExpiresAt, &c.Item.LastAttemptAt,
		&c.CouponValue, &c.CodeExpiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	c.Item.Status = service.RedeemEmailItemStatusSending
	return &c, nil
}

func (r *redeemEmailRepository) IsCodeAvailable(ctx context.Context, codeID *int64, now time.Time) (bool, error) {
	if codeID == nil {
		return false, nil
	}
	var available bool
	err := r.db.QueryRowContext(ctx, `SELECT status = 'unused' AND (expires_at IS NULL OR expires_at > $2) FROM redeem_codes WHERE id = $1`,
		*codeID, now).Scan(&available)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return available, err
}

const redeemEmailCompleteItemSQL = `UPDATE redeem_email_items SET status = $3,
	sent_at = CASE WHEN $4::boolean THEN $5::timestamptz ELSE NULL END,
	lease_expires_at = NULL, error_code = NULLIF($6::text, ''), error_message = NULLIF($7::text, ''), updated_at = $5::timestamptz
WHERE id = $1 AND attempt_token = $2::uuid AND status = 'sending'
RETURNING batch_id`

// 已退订跳过的明细不算失败：全部为发送成功或跳过时批次为 succeeded。
const redeemEmailCompleteBatchSQL = `UPDATE redeem_email_batches SET sending_count = sending_count - 1,
	sent_count = sent_count + $2, failed_count = failed_count + $3, skipped_count = skipped_count + $4,
	status = CASE WHEN pending_count = 0 AND sending_count = 1 THEN
		CASE WHEN failed_count + $3 = 0 THEN 'succeeded' WHEN sent_count + $2 = 0 THEN 'failed' ELSE 'partial_failed' END
		ELSE 'running' END,
	finished_at = CASE WHEN pending_count = 0 AND sending_count = 1 THEN $5::timestamptz ELSE NULL END,
	updated_at = $5::timestamptz
WHERE id = $1`

func (r *redeemEmailRepository) CompleteAttempt(ctx context.Context, p service.RedeemEmailAttemptResult) error {
	var sentDelta, failedDelta, skippedDelta int
	switch p.Status {
	case service.RedeemEmailItemStatusSent:
		sentDelta = 1
	case service.RedeemEmailItemStatusFailed:
		failedDelta = 1
	case service.RedeemEmailItemStatusSkipped:
		skippedDelta = 1
	default:
		return service.ErrRedeemEmailItemStatusInvalid
	}
	sent := p.Status == service.RedeemEmailItemStatusSent
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	var batchID int64
	err = tx.QueryRowContext(ctx, redeemEmailCompleteItemSQL, p.ItemID, p.AttemptToken.String(), p.Status, sent, p.CompletedAt,
		p.ErrorCode, p.ErrorMessage).Scan(&batchID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, redeemEmailCompleteBatchSQL, batchID, sentDelta, failedDelta, skippedDelta, p.CompletedAt); err != nil {
		return err
	}
	return tx.Commit()
}

// 租约过期说明 SMTP 结果未知（可能已投递），因此标记失败而不自动重发，交由管理员确认后重试。
const redeemEmailRecoverExpiredSQL = `WITH expired AS (
	SELECT id FROM redeem_email_items
	WHERE status = 'sending' AND lease_expires_at < $1::timestamptz
	ORDER BY lease_expires_at, id LIMIT $2 FOR UPDATE SKIP LOCKED
), updated AS (
	UPDATE redeem_email_items i SET status = 'failed', lease_expires_at = NULL,
		error_code = 'DELIVERY_OUTCOME_UNKNOWN', error_message = 'delivery outcome is unknown; verify before retrying',
		updated_at = $1::timestamptz
	FROM expired e WHERE i.id = e.id
	RETURNING i.batch_id
)
UPDATE redeem_email_batches b SET sending_count = sending_count - x.n, failed_count = failed_count + x.n,
	status = CASE WHEN pending_count = 0 AND sending_count = x.n THEN
		CASE WHEN sent_count = 0 THEN 'failed' ELSE 'partial_failed' END ELSE status END,
	finished_at = CASE WHEN pending_count = 0 AND sending_count = x.n THEN $1::timestamptz ELSE finished_at END,
	updated_at = $1::timestamptz
FROM (SELECT batch_id, COUNT(*)::int AS n FROM updated GROUP BY batch_id) x
WHERE b.id = x.batch_id`

func (r *redeemEmailRepository) RecoverExpiredSending(ctx context.Context, now time.Time, limit int) (int, error) {
	if limit <= 0 {
		limit = 100
	}
	res, err := r.db.ExecContext(ctx, redeemEmailRecoverExpiredSQL, now, limit)
	if err != nil {
		return 0, err
	}
	n, err := res.RowsAffected()
	return int(n), err
}
