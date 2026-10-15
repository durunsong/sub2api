package service

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

func TestNewRedeemEmailCutoffsUseCalendarMonthsInUTC(t *testing.T) {
	now := time.Date(2026, 10, 15, 8, 30, 0, 0, time.FixedZone("UTC+8", 8*3600))
	c := NewRedeemEmailCutoffs(now)
	require.True(t, time.Date(2026, 9, 15, 0, 30, 0, 0, time.UTC).Equal(c.OneMonth))
	require.True(t, time.Date(2026, 8, 15, 0, 30, 0, 0, time.UTC).Equal(c.TwoMonths))
	require.True(t, time.Date(2026, 7, 15, 0, 30, 0, 0, time.UTC).Equal(c.ThreeMonths))
	require.Equal(t, time.UTC, c.OneMonth.Location())
}

func TestNormalizeRedeemEmailSegments(t *testing.T) {
	got, err := NormalizeRedeemEmailSegments([]string{"never_used", "inactive_1_2m", " never_used "})
	require.NoError(t, err)
	require.Equal(t, []string{RedeemEmailSegmentInactive1To2Months, RedeemEmailSegmentNeverUsed}, got)

	for _, bad := range [][]string{nil, {}, {"inactive_4m"}, {"inactive_1_2m", "bogus"}} {
		_, err := NormalizeRedeemEmailSegments(bad)
		require.ErrorIs(t, err, ErrRedeemEmailFilterInvalid, "segments=%v", bad)
	}
	require.Equal(t, []string{"inactive_1_2m", "inactive_2_3m", "inactive_3m_plus"}, RedeemEmailDefaultSegments())
	require.NotContains(t, RedeemEmailDefaultSegments(), RedeemEmailSegmentNeverUsed)
}

func TestNormalizeRedeemEmailCouponValue(t *testing.T) {
	for _, v := range []float64{0.01, 5, 0.12345678, 1000000} {
		got, err := NormalizeRedeemEmailCouponValue(v)
		require.NoError(t, err, "value=%v", v)
		require.Equal(t, v, got)
	}
	for _, v := range []float64{0, -1, math.NaN(), math.Inf(1), 0.123456789, 1000000.01} {
		_, err := NormalizeRedeemEmailCouponValue(v)
		require.ErrorIs(t, err, ErrRedeemEmailCouponInvalid, "value=%v", v)
	}
}

func TestHashRedeemEmailCreateRequestIgnoresOrder(t *testing.T) {
	base := CreateRedeemEmailBatchInput{
		Filter:      RedeemEmailRecipientFilter{Segments: []string{"inactive_2_3m", "inactive_1_2m"}},
		UserIDs:     []int64{3, 1, 2},
		CouponValue: 5,
	}
	reordered := base
	reordered.UserIDs = []int64{2, 3, 1}
	reordered.Filter = RedeemEmailRecipientFilter{Segments: []string{"inactive_1_2m", "inactive_2_3m"}}
	changed := base
	changed.Filter = RedeemEmailRecipientFilter{Segments: []string{"inactive_1_2m"}}

	h1, err := HashRedeemEmailCreateRequest(base)
	require.NoError(t, err)
	h2, err := HashRedeemEmailCreateRequest(reordered)
	require.NoError(t, err)
	h3, err := HashRedeemEmailCreateRequest(changed)
	require.NoError(t, err)
	require.Equal(t, h1, h2)
	require.NotEqual(t, h1, h3)
	require.Len(t, h1, 64)
}

func TestFormatRedeemEmailCouponValue(t *testing.T) {
	require.Equal(t, "$5.00", FormatRedeemEmailCouponValue(5))
	require.Equal(t, "$0.12345678", FormatRedeemEmailCouponValue(0.12345678))
	require.Equal(t, "$1.50", FormatRedeemEmailCouponValue(1.5))
	require.Equal(t, "$0.125", FormatRedeemEmailCouponValue(0.125))
}

func TestRedeemInactiveCouponIsOptionalNotificationEvent(t *testing.T) {
	info, ok := notificationEmailEventDefinitions[NotificationEmailEventRedeemInactiveCoupon]
	require.True(t, ok)
	require.True(t, info.Optional)
	require.Equal(t, "marketing", info.Category)
	for _, p := range []string{"coupon_value", "redeem_code", "expires_at", "unsubscribe_url", "recipient_name"} {
		require.Contains(t, info.Placeholders, p)
	}
	require.Contains(t, notificationEmailEventOrder, NotificationEmailEventRedeemInactiveCoupon)

	svc := NewNotificationEmailService(newNotificationEmailMemorySettingRepo(), nil)
	for _, locale := range notificationEmailLocales {
		tmpl, err := svc.GetTemplate(context.Background(), NotificationEmailEventRedeemInactiveCoupon, locale)
		require.NoError(t, err)
		require.False(t, tmpl.IsCustom)
		require.Contains(t, tmpl.HTML, "{{redeem_code}}")
		require.Contains(t, tmpl.HTML, "{{unsubscribe_url}}")
	}
}

type redeemEmailRepoStub struct {
	searchSegments []string
	searchText     string
	searchCutoffs  RedeemEmailCutoffs
	searchLimit    int
	searchResult   []RedeemEmailRecipient
	createParams   *CreateRedeemEmailBatchParams
	createCalls    int
	createReplay   bool
	createErr      error
	retryCalls     int
	deleteCalls    int
	itemsStatus    string
	batchToReturn  *RedeemEmailBatch
}

func (s *redeemEmailRepoStub) SearchRecipients(_ context.Context, segments []string, search string, cutoffs RedeemEmailCutoffs, limit int) ([]RedeemEmailRecipient, error) {
	s.searchSegments, s.searchText, s.searchCutoffs, s.searchLimit = segments, search, cutoffs, limit
	return s.searchResult, nil
}
func (s *redeemEmailRepoStub) CreateBatch(_ context.Context, p CreateRedeemEmailBatchParams) (*RedeemEmailBatch, bool, error) {
	s.createCalls++
	s.createParams = &p
	if s.createErr != nil {
		return nil, false, s.createErr
	}
	return &RedeemEmailBatch{ID: 7, TotalCount: len(p.UserIDs)}, s.createReplay, nil
}
func (s *redeemEmailRepoStub) ListBatches(context.Context, pagination.PaginationParams) ([]RedeemEmailBatch, *pagination.PaginationResult, error) {
	return nil, &pagination.PaginationResult{}, nil
}
func (s *redeemEmailRepoStub) GetBatch(context.Context, int64) (*RedeemEmailBatch, error) {
	return s.batchToReturn, nil
}
func (s *redeemEmailRepoStub) ListItems(_ context.Context, _ int64, status string, _ pagination.PaginationParams) ([]RedeemEmailItem, *pagination.PaginationResult, error) {
	s.itemsStatus = status
	return nil, &pagination.PaginationResult{}, nil
}
func (s *redeemEmailRepoStub) RetryFailed(context.Context, int64, time.Time) (*RedeemEmailBatch, error) {
	s.retryCalls++
	return &RedeemEmailBatch{ID: 1}, nil
}
func (s *redeemEmailRepoStub) DeleteBatch(context.Context, int64) error {
	s.deleteCalls++
	return nil
}

type redeemEmailSMTPStub struct{ cfgErr error }

func (s *redeemEmailSMTPStub) GetSMTPConfig(context.Context) (*SMTPConfig, error) {
	if s.cfgErr != nil {
		return nil, s.cfgErr
	}
	return &SMTPConfig{Host: "smtp.example.com", From: "noreply@example.com"}, nil
}

func newRedeemEmailServiceForTest(repo *redeemEmailRepoStub, smtp *redeemEmailSMTPStub, now time.Time) *RedeemEmailService {
	s := NewRedeemEmailService(repo, smtp, NewNotificationEmailService(newNotificationEmailMemorySettingRepo(), nil))
	s.now = func() time.Time { return now }
	return s
}

func validRedeemEmailCreateInput() CreateRedeemEmailBatchInput {
	days := 30
	return CreateRedeemEmailBatchInput{
		CreatedBy:         9,
		IdempotencyKey:    "campaign-key-001",
		Filter:            RedeemEmailRecipientFilter{Segments: RedeemEmailDefaultSegments()},
		UserIDs:           []int64{11, 12},
		CouponValue:       5,
		CodeExpiresInDays: &days,
	}
}

func TestRedeemEmailServiceSearchComputesCutoffsServerSide(t *testing.T) {
	now := time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)
	repo := &redeemEmailRepoStub{}
	svc := newRedeemEmailServiceForTest(repo, &redeemEmailSMTPStub{}, now)

	_, err := svc.SearchRecipients(context.Background(), RedeemEmailRecipientFilter{Segments: []string{"never_used", "inactive_3m_plus"}, Search: "  ab  "})
	require.NoError(t, err)
	require.Equal(t, NewRedeemEmailCutoffs(now), repo.searchCutoffs)
	require.Equal(t, []string{"inactive_3m_plus", "never_used"}, repo.searchSegments)
	require.Equal(t, "ab", repo.searchText)
	require.Equal(t, RedeemEmailMaxRecipients+1, repo.searchLimit)

	_, err = svc.SearchRecipients(context.Background(), RedeemEmailRecipientFilter{})
	require.ErrorIs(t, err, ErrRedeemEmailFilterInvalid)
}

func TestRedeemEmailServiceSearchMarksUnsubscribedRecipients(t *testing.T) {
	ctx := context.Background()
	settings := newNotificationEmailMemorySettingRepo()
	require.NoError(t, settings.Set(ctx, notificationEmailPreferenceKey(NotificationEmailEventRedeemInactiveCoupon, "Gone@Example.com"), "unsubscribed"))
	repo := &redeemEmailRepoStub{searchResult: []RedeemEmailRecipient{{UserID: 1, Email: "stay@example.com"}, {UserID: 2, Email: "gone@example.com"}}}
	svc := NewRedeemEmailService(repo, &redeemEmailSMTPStub{}, NewNotificationEmailService(settings, nil))

	rows, err := svc.SearchRecipients(ctx, RedeemEmailRecipientFilter{Segments: RedeemEmailDefaultSegments()})
	require.NoError(t, err)
	require.False(t, rows[0].Unsubscribed)
	require.True(t, rows[1].Unsubscribed)
}

func TestRedeemEmailServiceSearchRejectsTooManyRecipients(t *testing.T) {
	repo := &redeemEmailRepoStub{searchResult: make([]RedeemEmailRecipient, RedeemEmailMaxRecipients+1)}
	svc := newRedeemEmailServiceForTest(repo, &redeemEmailSMTPStub{}, time.Now())
	_, err := svc.SearchRecipients(context.Background(), RedeemEmailRecipientFilter{Segments: RedeemEmailDefaultSegments()})
	require.ErrorIs(t, err, ErrRedeemEmailRecipientLimitExceeded)
}

func TestRedeemEmailServiceCreateBatchBuildsParams(t *testing.T) {
	now := time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)
	repo := &redeemEmailRepoStub{}
	svc := newRedeemEmailServiceForTest(repo, &redeemEmailSMTPStub{}, now)

	batch, replayed, err := svc.CreateBatch(context.Background(), validRedeemEmailCreateInput())
	require.NoError(t, err)
	require.False(t, replayed)
	require.Equal(t, int64(7), batch.ID)
	p := repo.createParams
	require.NotNil(t, p)
	require.Equal(t, NewRedeemEmailCutoffs(now), p.Cutoffs)
	require.Equal(t, RedeemEmailDefaultSegments(), p.Segments)
	require.NotNil(t, p.CodeExpiresAt)
	require.Equal(t, now.AddDate(0, 0, 30), *p.CodeExpiresAt)
	require.Len(t, p.IdempotencyKeyHash, 64)
	require.Len(t, p.RequestHash, 64)
	require.Equal(t, []int64{11, 12}, p.UserIDs)
}

func TestRedeemEmailServiceCreateBatchRejectsInvalidInputWithoutTouchingRepo(t *testing.T) {
	now := time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)
	tooMany := make([]int64, RedeemEmailMaxRecipients+1)
	for i := range tooMany {
		tooMany[i] = int64(i + 1)
	}
	zeroDays, bigDays := 0, 3651
	cases := map[string]struct {
		mutate func(*CreateRedeemEmailBatchInput)
		want   error
	}{
		"short key":      {func(in *CreateRedeemEmailBatchInput) { in.IdempotencyKey = "short" }, ErrRedeemEmailIdempotencyKeyInvalid},
		"no users":       {func(in *CreateRedeemEmailBatchInput) { in.UserIDs = nil }, ErrRedeemEmailRecipientInvalid},
		"duplicate user": {func(in *CreateRedeemEmailBatchInput) { in.UserIDs = []int64{1, 1} }, ErrRedeemEmailRecipientInvalid},
		"bad user id":    {func(in *CreateRedeemEmailBatchInput) { in.UserIDs = []int64{0} }, ErrRedeemEmailRecipientInvalid},
		"too many":       {func(in *CreateRedeemEmailBatchInput) { in.UserIDs = tooMany }, ErrRedeemEmailRecipientLimitExceeded},
		"zero value":     {func(in *CreateRedeemEmailBatchInput) { in.CouponValue = 0 }, ErrRedeemEmailCouponInvalid},
		"zero days":      {func(in *CreateRedeemEmailBatchInput) { in.CodeExpiresInDays = &zeroDays }, ErrRedeemEmailCouponInvalid},
		"too many days":  {func(in *CreateRedeemEmailBatchInput) { in.CodeExpiresInDays = &bigDays }, ErrRedeemEmailCouponInvalid},
		"no segments":    {func(in *CreateRedeemEmailBatchInput) { in.Filter.Segments = nil }, ErrRedeemEmailFilterInvalid},
		"bad segment":    {func(in *CreateRedeemEmailBatchInput) { in.Filter.Segments = []string{"inactive_6m"} }, ErrRedeemEmailFilterInvalid},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			repo := &redeemEmailRepoStub{}
			svc := newRedeemEmailServiceForTest(repo, &redeemEmailSMTPStub{}, now)
			in := validRedeemEmailCreateInput()
			tc.mutate(&in)
			_, _, err := svc.CreateBatch(context.Background(), in)
			require.ErrorIs(t, err, tc.want)
			require.Zero(t, repo.createCalls)
		})
	}
}

func TestRedeemEmailServiceCreateBatchRequiresSMTP(t *testing.T) {
	repo := &redeemEmailRepoStub{}
	svc := newRedeemEmailServiceForTest(repo, &redeemEmailSMTPStub{cfgErr: ErrEmailNotConfigured}, time.Now())
	_, _, err := svc.CreateBatch(context.Background(), validRedeemEmailCreateInput())
	require.ErrorIs(t, err, ErrRedeemEmailSMTPNotConfigured)
	require.Zero(t, repo.createCalls)

	defaults := svc.Defaults(context.Background())
	require.False(t, defaults.SMTPConfigured)
	require.Equal(t, RedeemEmailMaxRecipients, defaults.MaxRecipients)
	require.Equal(t, NotificationEmailEventRedeemInactiveCoupon, defaults.TemplateEvent)
	require.Equal(t, RedeemEmailDefaultSegments(), defaults.DefaultSegments)
	require.Len(t, defaults.SegmentOptions, 4)
}

func TestRedeemEmailServiceCreateBatchNeverExpires(t *testing.T) {
	repo := &redeemEmailRepoStub{}
	svc := newRedeemEmailServiceForTest(repo, &redeemEmailSMTPStub{}, time.Now())
	in := validRedeemEmailCreateInput()
	in.CodeExpiresInDays = nil
	_, _, err := svc.CreateBatch(context.Background(), in)
	require.NoError(t, err)
	require.Nil(t, repo.createParams.CodeExpiresAt)
}

func TestRedeemEmailServicePreviewUsesSavedSettingsTemplate(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)
	settings := newNotificationEmailMemorySettingRepo()
	notification := NewNotificationEmailService(settings, nil)
	svc := NewRedeemEmailService(&redeemEmailRepoStub{}, &redeemEmailSMTPStub{}, notification)
	svc.now = func() time.Time { return now }

	days := 30
	official, err := svc.Preview(ctx, RedeemEmailPreviewInput{Locale: "zh", CouponValue: 2.5, CodeExpiresInDays: &days})
	require.NoError(t, err)
	require.Contains(t, official.Subject, "$2.50")
	require.Contains(t, official.HTML, "2026-11-14 00:00 UTC")
	require.Contains(t, official.HTML, "PREVIEW-CODE-0000")

	_, err = notification.UpdateTemplate(ctx, NotificationEmailEventRedeemInactiveCoupon, "en", "Custom {{coupon_value}}", "<p>{{redeem_code}} until {{expires_at}}</p>")
	require.NoError(t, err)
	custom, err := svc.Preview(ctx, RedeemEmailPreviewInput{Locale: "en", CouponValue: 5})
	require.NoError(t, err)
	require.Equal(t, "Custom $5.00", custom.Subject)
	require.Equal(t, "<p>PREVIEW-CODE-0000 until Never expires</p>", custom.HTML)

	_, err = svc.Preview(ctx, RedeemEmailPreviewInput{Locale: "en", CouponValue: 0})
	require.ErrorIs(t, err, ErrRedeemEmailCouponInvalid)
}

func TestRedeemEmailServiceListItemsValidatesStatus(t *testing.T) {
	repo := &redeemEmailRepoStub{}
	svc := newRedeemEmailServiceForTest(repo, &redeemEmailSMTPStub{}, time.Now())
	_, _, err := svc.ListItems(context.Background(), 1, "skipped", pagination.PaginationParams{})
	require.NoError(t, err)
	require.Equal(t, "skipped", repo.itemsStatus)
	_, _, err = svc.ListItems(context.Background(), 1, "bogus", pagination.PaginationParams{})
	require.Error(t, err)
}

func TestClassifyRedeemEmailDeliveryError(t *testing.T) {
	require.Equal(t, "SMTP_AUTH_FAILED", ClassifyRedeemEmailDeliveryError(errors.New("535 auth failed")).Code)
	require.Equal(t, "SMTP_RECIPIENT_REJECTED", ClassifyRedeemEmailDeliveryError(errors.New("rcpt to rejected")).Code)
	require.Equal(t, "SMTP_TIMEOUT", ClassifyRedeemEmailDeliveryError(context.DeadlineExceeded).Code)
	require.Equal(t, "SMTP_DELIVERY_FAILED", ClassifyRedeemEmailDeliveryError(errors.New("boom")).Code)
	require.Equal(t, "TEMPLATE_INVALID", ClassifyRedeemEmailDeliveryError(notificationEmailTemplateErr(errors.New("x"))).Code)
	require.Equal(t, "SMTP_NOT_CONFIGURED", ClassifyRedeemEmailDeliveryError(notificationEmailConfigErr(errors.New("x"))).Code)
}
