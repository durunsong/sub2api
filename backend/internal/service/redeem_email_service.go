package service

import (
	"context"
	"errors"
	"net"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

const redeemEmailSearchMaxLength = 100

type RedeemEmailSMTPConfigProvider interface {
	GetSMTPConfig(ctx context.Context) (*SMTPConfig, error)
}

type RedeemEmailNotificationReader interface {
	PreviewTemplate(ctx context.Context, input NotificationEmailPreviewInput) (NotificationEmailPreview, error)
	IsUnsubscribed(ctx context.Context, email, event string) (bool, error)
}

type redeemEmailServiceRepository interface {
	SearchRecipients(context.Context, []string, string, RedeemEmailCutoffs, int) ([]RedeemEmailRecipient, error)
	CreateBatch(context.Context, CreateRedeemEmailBatchParams) (*RedeemEmailBatch, bool, error)
	ListBatches(context.Context, pagination.PaginationParams) ([]RedeemEmailBatch, *pagination.PaginationResult, error)
	GetBatch(context.Context, int64) (*RedeemEmailBatch, error)
	ListItems(context.Context, int64, string, pagination.PaginationParams) ([]RedeemEmailItem, *pagination.PaginationResult, error)
	RetryFailed(context.Context, int64, time.Time) (*RedeemEmailBatch, error)
	DeleteBatch(context.Context, int64) error
}

type RedeemEmailDefaults struct {
	TemplateEvent        string   `json:"template_event"`
	SegmentOptions       []string `json:"segment_options"`
	DefaultSegments      []string `json:"default_segments"`
	RatePerSecond        int      `json:"rate_per_second"`
	SMTPConfigured       bool     `json:"smtp_configured"`
	MaxRecipients        int      `json:"max_recipients"`
	DefaultExpiresInDays int      `json:"default_expires_in_days"`
	MaxExpiresInDays     int      `json:"max_expires_in_days"`
}

type RedeemEmailPreviewInput struct {
	Locale            string
	CouponValue       float64
	CodeExpiresInDays *int
}

type RedeemEmailService struct {
	repo      redeemEmailServiceRepository
	smtp      RedeemEmailSMTPConfigProvider
	previewer RedeemEmailNotificationReader
	now       func() time.Time
}

func NewRedeemEmailService(repo redeemEmailServiceRepository, smtp RedeemEmailSMTPConfigProvider, previewer RedeemEmailNotificationReader) *RedeemEmailService {
	return &RedeemEmailService{repo: repo, smtp: smtp, previewer: previewer, now: time.Now}
}

func ProvideRedeemEmailService(repo RedeemEmailRepository, emailService *EmailService, notificationEmail *NotificationEmailService) *RedeemEmailService {
	return NewRedeemEmailService(repo, emailService, notificationEmail)
}

func (s *RedeemEmailService) smtpConfigured(ctx context.Context) bool {
	cfg, err := s.smtp.GetSMTPConfig(ctx)
	return err == nil && cfg != nil && strings.TrimSpace(cfg.Host) != ""
}

func (s *RedeemEmailService) Defaults(ctx context.Context) RedeemEmailDefaults {
	return RedeemEmailDefaults{
		TemplateEvent:        NotificationEmailEventRedeemInactiveCoupon,
		SegmentOptions:       RedeemEmailSegmentOptions(),
		DefaultSegments:      RedeemEmailDefaultSegments(),
		RatePerSecond:        1,
		SMTPConfigured:       s.smtpConfigured(ctx),
		MaxRecipients:        RedeemEmailMaxRecipients,
		DefaultExpiresInDays: 30,
		MaxExpiresInDays:     RedeemEmailMaxExpiresInDays,
	}
}

func normalizeRedeemEmailSearch(search string) string {
	search = strings.TrimSpace(search)
	if r := []rune(search); len(r) > redeemEmailSearchMaxLength {
		search = string(r[:redeemEmailSearchMaxLength])
	}
	return search
}

func (s *RedeemEmailService) SearchRecipients(ctx context.Context, f RedeemEmailRecipientFilter) ([]RedeemEmailRecipient, error) {
	segments, err := NormalizeRedeemEmailSegments(f.Segments)
	if err != nil {
		return nil, err
	}
	rows, err := s.repo.SearchRecipients(ctx, segments, normalizeRedeemEmailSearch(f.Search), NewRedeemEmailCutoffs(s.now()), RedeemEmailMaxRecipients+1)
	if err != nil {
		return nil, err
	}
	if len(rows) > RedeemEmailMaxRecipients {
		return nil, ErrRedeemEmailRecipientLimitExceeded
	}
	for i := range rows {
		unsubscribed, err := s.previewer.IsUnsubscribed(ctx, rows[i].Email, NotificationEmailEventRedeemInactiveCoupon)
		if err != nil {
			return nil, err
		}
		rows[i].Unsubscribed = unsubscribed
	}
	return rows, nil
}

func validateRedeemEmailExpiresInDays(days *int) error {
	if days != nil && (*days < 1 || *days > RedeemEmailMaxExpiresInDays) {
		return ErrRedeemEmailCouponInvalid
	}
	return nil
}

// Preview 用 /admin/settings 中保存的 redeem.inactive_coupon 模板和本次券面信息渲染示例邮件。
func (s *RedeemEmailService) Preview(ctx context.Context, in RedeemEmailPreviewInput) (NotificationEmailPreview, error) {
	value, err := NormalizeRedeemEmailCouponValue(in.CouponValue)
	if err != nil {
		return NotificationEmailPreview{}, err
	}
	if err := validateRedeemEmailExpiresInDays(in.CodeExpiresInDays); err != nil {
		return NotificationEmailPreview{}, err
	}
	expiresAt := redeemEmailNeverExpiresLabel(in.Locale)
	if in.CodeExpiresInDays != nil {
		expiresAt = FormatRedeemEmailExpiresAt(s.now().UTC().AddDate(0, 0, *in.CodeExpiresInDays))
	}
	preview, err := s.previewer.PreviewTemplate(ctx, NotificationEmailPreviewInput{
		Event:  NotificationEmailEventRedeemInactiveCoupon,
		Locale: in.Locale,
		Variables: map[string]string{
			"coupon_value": FormatRedeemEmailCouponValue(value),
			"expires_at":   expiresAt,
		},
	})
	if err != nil {
		return NotificationEmailPreview{}, ErrRedeemEmailTemplateInvalid.WithCause(err)
	}
	return preview, nil
}

func FormatRedeemEmailExpiresAt(t time.Time) string {
	return t.UTC().Format("2006-01-02 15:04 UTC")
}

func (s *RedeemEmailService) CreateBatch(ctx context.Context, in CreateRedeemEmailBatchInput) (*RedeemEmailBatch, bool, error) {
	if !validRedeemEmailIdempotencyKey(in.IdempotencyKey) {
		return nil, false, ErrRedeemEmailIdempotencyKeyInvalid
	}
	now := s.now().UTC()
	segments, err := NormalizeRedeemEmailSegments(in.Filter.Segments)
	if err != nil {
		return nil, false, err
	}
	if len(in.UserIDs) == 0 {
		return nil, false, ErrRedeemEmailRecipientInvalid
	}
	if len(in.UserIDs) > RedeemEmailMaxRecipients {
		return nil, false, ErrRedeemEmailRecipientLimitExceeded
	}
	seen := make(map[int64]struct{}, len(in.UserIDs))
	for _, id := range in.UserIDs {
		if _, dup := seen[id]; dup || id <= 0 {
			return nil, false, ErrRedeemEmailRecipientInvalid
		}
		seen[id] = struct{}{}
	}
	value, err := NormalizeRedeemEmailCouponValue(in.CouponValue)
	if err != nil {
		return nil, false, err
	}
	in.CouponValue = value
	if err := validateRedeemEmailExpiresInDays(in.CodeExpiresInDays); err != nil {
		return nil, false, err
	}
	var expiresAt *time.Time
	if in.CodeExpiresInDays != nil {
		at := now.AddDate(0, 0, *in.CodeExpiresInDays)
		expiresAt = &at
	}
	if _, err := s.smtp.GetSMTPConfig(ctx); err != nil {
		if errors.Is(err, ErrEmailNotConfigured) {
			return nil, false, ErrRedeemEmailSMTPNotConfigured
		}
		return nil, false, err
	}
	requestHash, err := HashRedeemEmailCreateRequest(in)
	if err != nil {
		return nil, false, err
	}
	return s.repo.CreateBatch(ctx, CreateRedeemEmailBatchParams{
		CreatedBy:          in.CreatedBy,
		IdempotencyKeyHash: RedeemEmailHashKey(in.IdempotencyKey),
		RequestHash:        requestHash,
		Segments:           segments,
		Cutoffs:            NewRedeemEmailCutoffs(now),
		UserIDs:            append([]int64(nil), in.UserIDs...),
		CouponValue:        value,
		CodeExpiresAt:      expiresAt,
		Now:                now,
	})
}

func (s *RedeemEmailService) ListBatches(ctx context.Context, p pagination.PaginationParams) ([]RedeemEmailBatch, *pagination.PaginationResult, error) {
	return s.repo.ListBatches(ctx, p)
}

func (s *RedeemEmailService) GetBatch(ctx context.Context, id int64) (*RedeemEmailBatch, error) {
	return s.repo.GetBatch(ctx, id)
}

func (s *RedeemEmailService) ListItems(ctx context.Context, batchID int64, status string, p pagination.PaginationParams) ([]RedeemEmailItem, *pagination.PaginationResult, error) {
	switch status {
	case "", RedeemEmailItemStatusPending, RedeemEmailItemStatusSending, RedeemEmailItemStatusSent, RedeemEmailItemStatusFailed, RedeemEmailItemStatusSkipped:
	default:
		return nil, nil, ErrRedeemEmailItemStatusInvalid
	}
	return s.repo.ListItems(ctx, batchID, status, p)
}

func (s *RedeemEmailService) RetryFailed(ctx context.Context, batchID int64) (*RedeemEmailBatch, error) {
	return s.repo.RetryFailed(ctx, batchID, s.now().UTC())
}

func (s *RedeemEmailService) DeleteBatch(ctx context.Context, batchID int64) error {
	return s.repo.DeleteBatch(ctx, batchID)
}

type RedeemEmailDeliveryFailure struct{ Code, Message string }

func ClassifyRedeemEmailDeliveryError(err error) RedeemEmailDeliveryFailure {
	if err == nil {
		return RedeemEmailDeliveryFailure{}
	}
	var tmplErr notificationEmailTemplateError
	if errors.As(err, &tmplErr) {
		return RedeemEmailDeliveryFailure{"TEMPLATE_INVALID", "email template in settings is invalid"}
	}
	var cfgErr notificationEmailConfigError
	if errors.As(err, &cfgErr) || errors.Is(err, ErrEmailNotConfigured) {
		return RedeemEmailDeliveryFailure{"SMTP_NOT_CONFIGURED", "SMTP is not configured"}
	}
	lower := strings.ToLower(err.Error())
	var ne net.Error
	switch {
	case strings.Contains(lower, "auth"):
		return RedeemEmailDeliveryFailure{"SMTP_AUTH_FAILED", "SMTP authentication failed"}
	case strings.Contains(lower, "rcpt") || strings.Contains(lower, "recipient"):
		return RedeemEmailDeliveryFailure{"SMTP_RECIPIENT_REJECTED", "SMTP rejected the recipient"}
	case errors.As(err, &ne) && ne.Timeout(), errors.Is(err, context.DeadlineExceeded), strings.Contains(lower, "timeout"):
		return RedeemEmailDeliveryFailure{"SMTP_TIMEOUT", "SMTP delivery timed out"}
	default:
		return RedeemEmailDeliveryFailure{"SMTP_DELIVERY_FAILED", "SMTP delivery failed"}
	}
}
