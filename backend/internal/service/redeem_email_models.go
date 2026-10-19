package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/google/uuid"
)

const (
	RedeemEmailMaxRecipients    = 1000
	RedeemEmailMaxCouponValue   = 1000000
	RedeemEmailMaxExpiresInDays = 3650
	redeemEmailCouponScale      = 1e8
)

var (
	ErrRedeemEmailFilterInvalid          = infraerrors.BadRequest("REDEEM_EMAIL_FILTER_INVALID", "segments must be a non-empty subset of inactive_1_2m, inactive_2_3m, inactive_3m_plus, never_used, or exactly zero_balance")
	ErrRedeemEmailCouponInvalid          = infraerrors.BadRequest("REDEEM_EMAIL_COUPON_INVALID", "coupon value must be positive with at most 8 decimals, and expiry must be 1-3650 days")
	ErrRedeemEmailRecipientInvalid       = infraerrors.BadRequest("REDEEM_EMAIL_RECIPIENT_INVALID", "recipients are invalid")
	ErrRedeemEmailRecipientIneligible    = infraerrors.BadRequest("REDEEM_EMAIL_RECIPIENT_INELIGIBLE", "some recipients no longer match the inactive filter")
	ErrRedeemEmailRecipientLimitExceeded = infraerrors.BadRequest("REDEEM_EMAIL_RECIPIENT_LIMIT_EXCEEDED", "recipient limit exceeded")
	ErrRedeemEmailTemplateInvalid        = infraerrors.BadRequest("REDEEM_EMAIL_TEMPLATE_INVALID", "the redeem.inactive_coupon email template saved in settings is invalid")
	ErrRedeemEmailSMTPNotConfigured      = infraerrors.ServiceUnavailable("REDEEM_EMAIL_SMTP_NOT_CONFIGURED", "SMTP is not configured")
	ErrRedeemEmailBatchConflict          = infraerrors.Conflict("REDEEM_EMAIL_BATCH_CONFLICT", "idempotency key conflicts with another request")
	ErrRedeemEmailBatchNotFound          = infraerrors.NotFound("REDEEM_EMAIL_BATCH_NOT_FOUND", "redeem email batch not found")
	ErrRedeemEmailBatchNotDeletable      = infraerrors.Conflict("REDEEM_EMAIL_BATCH_NOT_DELETABLE", "active redeem email batch cannot be deleted")
	ErrRedeemEmailIdempotencyKeyInvalid  = infraerrors.BadRequest("REDEEM_EMAIL_IDEMPOTENCY_KEY_INVALID", "idempotency key must be 8 to 128 characters")
	ErrRedeemEmailItemStatusInvalid      = infraerrors.BadRequest("REDEEM_EMAIL_ITEM_STATUS_INVALID", "invalid item status")
)

const (
	RedeemEmailBatchStatusPending       = "pending"
	RedeemEmailBatchStatusRunning       = "running"
	RedeemEmailBatchStatusSucceeded     = "succeeded"
	RedeemEmailBatchStatusPartialFailed = "partial_failed"
	RedeemEmailBatchStatusFailed        = "failed"

	RedeemEmailItemStatusPending = "pending"
	RedeemEmailItemStatusSending = "sending"
	RedeemEmailItemStatusSent    = "sent"
	RedeemEmailItemStatusFailed  = "failed"
	RedeemEmailItemStatusSkipped = "skipped"
)

// 未活跃分段按最后一次 usage_logs 时间划分为互斥区间；多选取并集。
const (
	RedeemEmailSegmentInactive1To2Months  = "inactive_1_2m"
	RedeemEmailSegmentInactive2To3Months  = "inactive_2_3m"
	RedeemEmailSegmentInactive3MonthsPlus = "inactive_3m_plus"
	RedeemEmailSegmentNeverUsed           = "never_used"
	// RedeemEmailSegmentZeroBalance 与未活跃分段互斥：不看最后使用时间，只匹配钱包余额恰好为 0 的用户。
	RedeemEmailSegmentZeroBalance = "zero_balance"
)

var redeemEmailSegmentOrder = []string{
	RedeemEmailSegmentInactive1To2Months,
	RedeemEmailSegmentInactive2To3Months,
	RedeemEmailSegmentInactive3MonthsPlus,
	RedeemEmailSegmentNeverUsed,
}

func RedeemEmailSegmentOptions() []string {
	return append(append([]string(nil), redeemEmailSegmentOrder...), RedeemEmailSegmentZeroBalance)
}

func RedeemEmailDefaultSegments() []string {
	return append([]string(nil), redeemEmailSegmentOrder[:3]...)
}

// NormalizeRedeemEmailSegments 去重并按固定顺序排列，保证哈希与存储稳定。
// zero_balance 不能与未活跃分段取并集，只能单独选择。
func NormalizeRedeemEmailSegments(segments []string) ([]string, error) {
	selected := make(map[string]bool, len(segments))
	for _, s := range segments {
		selected[strings.TrimSpace(s)] = true
	}
	if selected[RedeemEmailSegmentZeroBalance] {
		if len(selected) != 1 {
			return nil, ErrRedeemEmailFilterInvalid
		}
		return []string{RedeemEmailSegmentZeroBalance}, nil
	}
	out := make([]string, 0, len(selected))
	for _, s := range redeemEmailSegmentOrder {
		if selected[s] {
			out = append(out, s)
			delete(selected, s)
		}
	}
	if len(out) == 0 || len(selected) > 0 {
		return nil, ErrRedeemEmailFilterInvalid
	}
	return out, nil
}

// RedeemEmailCutoffs 是分段边界：最后使用时间 t 落在 [TwoMonths, OneMonth) 为 1–2 个月，依此类推。
type RedeemEmailCutoffs struct {
	OneMonth    time.Time
	TwoMonths   time.Time
	ThreeMonths time.Time
}

func NewRedeemEmailCutoffs(reference time.Time) RedeemEmailCutoffs {
	ref := reference.UTC()
	return RedeemEmailCutoffs{
		OneMonth:    ref.AddDate(0, -1, 0),
		TwoMonths:   ref.AddDate(0, -2, 0),
		ThreeMonths: ref.AddDate(0, -3, 0),
	}
}

type RedeemEmailRecipientFilter struct {
	Segments []string `json:"segments"`
	Search   string   `json:"search,omitempty"`
}

type RedeemEmailRecipient struct {
	UserID     int64      `json:"user_id"`
	Email      string     `json:"email"`
	Username   string     `json:"username"`
	LastUsedAt *time.Time `json:"last_used_at"`
	Segment    string     `json:"segment"`
	// Unsubscribed 表示已退订 redeem.inactive_coupon；仅用于提示，发送时 Worker 仍会再次检查并跳过。
	Unsubscribed bool `json:"unsubscribed"`
}

type CreateRedeemEmailBatchInput struct {
	CreatedBy         int64
	IdempotencyKey    string
	Filter            RedeemEmailRecipientFilter
	UserIDs           []int64
	CouponValue       float64
	CodeExpiresInDays *int
}

// CreateRedeemEmailBatchParams 是已校验的建批参数；仓储必须在同一事务中复核资格并为每位用户生成兑换码。
type CreateRedeemEmailBatchParams struct {
	CreatedBy          int64
	IdempotencyKeyHash string
	RequestHash        string
	Segments           []string
	Cutoffs            RedeemEmailCutoffs
	UserIDs            []int64
	CouponValue        float64
	CodeExpiresAt      *time.Time
	Now                time.Time
}

type RedeemEmailBatch struct {
	ID            int64      `json:"id"`
	Status        string     `json:"status"`
	Segments      []string   `json:"segments"`
	ReferenceTime time.Time  `json:"reference_time"`
	CouponValue   float64    `json:"coupon_value"`
	CodeExpiresAt *time.Time `json:"code_expires_at"`
	TotalCount    int        `json:"total_count"`
	PendingCount  int        `json:"pending_count"`
	SendingCount  int        `json:"sending_count"`
	SentCount     int        `json:"sent_count"`
	FailedCount   int        `json:"failed_count"`
	SkippedCount  int        `json:"skipped_count"`
	CreatedBy     int64      `json:"created_by"`
	StartedAt     *time.Time `json:"started_at"`
	FinishedAt    *time.Time `json:"finished_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type RedeemEmailItem struct {
	ID             int64      `json:"id"`
	BatchID        int64      `json:"batch_id"`
	Position       int        `json:"position"`
	UserID         *int64     `json:"user_id"`
	RecipientName  string     `json:"recipient_name"`
	RecipientEmail string     `json:"recipient_email"`
	RedeemCodeID   *int64     `json:"redeem_code_id"`
	RedeemCode     string     `json:"redeem_code"`
	Status         string     `json:"status"`
	AttemptCount   int        `json:"attempt_count"`
	AttemptToken   *uuid.UUID `json:"-"`
	LeaseExpiresAt *time.Time `json:"-"`
	SentAt         *time.Time `json:"sent_at"`
	LastAttemptAt  *time.Time `json:"last_attempt_at"`
	ErrorCode      *string    `json:"error_code"`
	ErrorMessage   *string    `json:"error_message"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type RedeemEmailClaim struct {
	Item          RedeemEmailItem
	CouponValue   float64
	CodeExpiresAt *time.Time
}

// RedeemEmailAttemptResult 的 Status 只能是 sent / failed / skipped。
type RedeemEmailAttemptResult struct {
	ItemID       int64
	AttemptToken uuid.UUID
	Status       string
	ErrorCode    string
	ErrorMessage string
	CompletedAt  time.Time
}

type RedeemEmailRepository interface {
	SearchRecipients(ctx context.Context, segments []string, search string, cutoffs RedeemEmailCutoffs, limit int) ([]RedeemEmailRecipient, error)
	CreateBatch(ctx context.Context, params CreateRedeemEmailBatchParams) (*RedeemEmailBatch, bool, error)
	ListBatches(ctx context.Context, params pagination.PaginationParams) ([]RedeemEmailBatch, *pagination.PaginationResult, error)
	GetBatch(ctx context.Context, id int64) (*RedeemEmailBatch, error)
	ListItems(ctx context.Context, batchID int64, status string, params pagination.PaginationParams) ([]RedeemEmailItem, *pagination.PaginationResult, error)
	RetryFailed(ctx context.Context, batchID int64, now time.Time) (*RedeemEmailBatch, error)
	DeleteBatch(ctx context.Context, batchID int64) error
	ClaimNext(ctx context.Context, now time.Time, lease time.Duration) (*RedeemEmailClaim, error)
	IsCodeAvailable(ctx context.Context, codeID *int64, now time.Time) (bool, error)
	CompleteAttempt(ctx context.Context, result RedeemEmailAttemptResult) error
	RecoverExpiredSending(ctx context.Context, now time.Time, limit int) (int, error)
}

func NormalizeRedeemEmailCouponValue(v float64) (float64, error) {
	if math.IsNaN(v) || math.IsInf(v, 0) || v <= 0 || v > RedeemEmailMaxCouponValue {
		return 0, ErrRedeemEmailCouponInvalid
	}
	scaled := v * redeemEmailCouponScale
	if math.Abs(scaled-math.Round(scaled)) > 1e-6 {
		return 0, ErrRedeemEmailCouponInvalid
	}
	return math.Round(scaled) / redeemEmailCouponScale, nil
}

func FormatRedeemEmailCouponValue(v float64) string {
	s := strconv.FormatFloat(v, 'f', 8, 64)
	s = strings.TrimRight(s, "0")
	if dot := strings.IndexByte(s, '.'); dot >= 0 && len(s)-dot-1 < 2 {
		s += strings.Repeat("0", 2-(len(s)-dot-1))
	}
	return "$" + s
}

func RedeemEmailRecipientName(username, email string) string {
	if name := strings.TrimSpace(username); name != "" {
		return name
	}
	trimmed := strings.TrimSpace(email)
	if at := strings.LastIndex(trimmed, "@"); at > 0 {
		return trimmed[:at]
	}
	return trimmed
}

func RedeemEmailHashKey(v string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(v)))
	return hex.EncodeToString(sum[:])
}

func validRedeemEmailIdempotencyKey(key string) bool {
	trimmed := strings.TrimSpace(key)
	return len(trimmed) >= 8 && len(trimmed) <= 128
}

func RedeemEmailBatchCanDelete(status string) bool {
	switch status {
	case RedeemEmailBatchStatusSucceeded, RedeemEmailBatchStatusPartialFailed, RedeemEmailBatchStatusFailed:
		return true
	}
	return false
}

type redeemEmailCanonicalCreateRequest struct {
	Segments          []string `json:"segments"`
	UserIDs           []int64  `json:"user_ids"`
	CouponValue       string   `json:"coupon_value"`
	CodeExpiresInDays *int     `json:"code_expires_in_days"`
}

// HashRedeemEmailCreateRequest 对影响发券结果的字段做规范化哈希；筛选基准时间不参与，重放以原批次为准。
func HashRedeemEmailCreateRequest(in CreateRedeemEmailBatchInput) (string, error) {
	segments, err := NormalizeRedeemEmailSegments(in.Filter.Segments)
	if err != nil {
		return "", err
	}
	ids := append([]int64(nil), in.UserIDs...)
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	raw, err := json.Marshal(redeemEmailCanonicalCreateRequest{
		Segments:          segments,
		UserIDs:           ids,
		CouponValue:       strconv.FormatFloat(in.CouponValue, 'f', 8, 64),
		CodeExpiresInDays: in.CodeExpiresInDays,
	})
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), nil
}
