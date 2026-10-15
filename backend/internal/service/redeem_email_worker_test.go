package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type redeemEmailWorkerRepoStub struct {
	claims       []*RedeemEmailClaim
	available    bool
	availableErr error
	completed    []RedeemEmailAttemptResult
	recoverCalls int
	claimCalls   int
}

func (s *redeemEmailWorkerRepoStub) ClaimNext(context.Context, time.Time, time.Duration) (*RedeemEmailClaim, error) {
	s.claimCalls++
	if len(s.claims) == 0 {
		return nil, nil
	}
	c := s.claims[0]
	s.claims = s.claims[1:]
	return c, nil
}
func (s *redeemEmailWorkerRepoStub) IsCodeAvailable(context.Context, *int64, time.Time) (bool, error) {
	return s.available, s.availableErr
}
func (s *redeemEmailWorkerRepoStub) CompleteAttempt(_ context.Context, r RedeemEmailAttemptResult) error {
	s.completed = append(s.completed, r)
	return nil
}
func (s *redeemEmailWorkerRepoStub) RecoverExpiredSending(context.Context, time.Time, int) (int, error) {
	s.recoverCalls++
	return 0, nil
}

type redeemEmailFakeClock struct{ now time.Time }

func (c *redeemEmailFakeClock) Now() time.Time { return c.now }
func (c *redeemEmailFakeClock) Sleep(ctx context.Context, d time.Duration) bool {
	c.now = c.now.Add(d)
	return ctx.Err() == nil
}

type redeemEmailNotifierStub struct {
	locale         string
	unsubscribed   bool
	unsubscribeErr error
	sendErr        error
	sent           []NotificationEmailSendInput
}

func (s *redeemEmailNotifierStub) IsUnsubscribed(context.Context, string, string) (bool, error) {
	return s.unsubscribed, s.unsubscribeErr
}
func (s *redeemEmailNotifierStub) ResolveRecipientLocale(context.Context, int64, string) string {
	if s.locale == "" {
		return notificationEmailLocaleChinese
	}
	return s.locale
}
func (s *redeemEmailNotifierStub) Send(_ context.Context, in NotificationEmailSendInput) error {
	s.sent = append(s.sent, in)
	return s.sendErr
}

type redeemEmailLeaderStub struct{ held bool }

func (l *redeemEmailLeaderStub) TryAcquireLeaderLock(context.Context, string, string, time.Duration) (bool, error) {
	return !l.held, nil
}
func (l *redeemEmailLeaderStub) ReleaseLeaderLock(context.Context, string, string) error { return nil }

func redeemEmailTestClaim() *RedeemEmailClaim {
	token := uuid.New()
	codeID := int64(55)
	userID := int64(42)
	expires := time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)
	return &RedeemEmailClaim{
		Item: RedeemEmailItem{
			ID: 1, BatchID: 2, UserID: &userID, RecipientName: "", RecipientEmail: "tom@example.com",
			RedeemCodeID: &codeID, RedeemCode: "CODE123", AttemptToken: &token,
		},
		CouponValue:   5,
		CodeExpiresAt: &expires,
	}
}

func newRedeemEmailWorkerForTest(repo *redeemEmailWorkerRepoStub, notifier *redeemEmailNotifierStub, leader *redeemEmailLeaderStub) (*RedeemEmailWorker, *redeemEmailFakeClock) {
	clock := &redeemEmailFakeClock{now: time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)}
	return NewRedeemEmailWorker(repo, notifier, leader, nil, RedeemEmailWorkerOptions{}, clock), clock
}

func TestRedeemEmailWorkerSendsNotificationEventAndCompletes(t *testing.T) {
	claim := redeemEmailTestClaim()
	repo := &redeemEmailWorkerRepoStub{claims: []*RedeemEmailClaim{claim}, available: true}
	notifier := &redeemEmailNotifierStub{locale: notificationEmailDefaultLocale}
	w, clock := newRedeemEmailWorkerForTest(repo, notifier, &redeemEmailLeaderStub{})
	start := clock.now

	require.True(t, w.RunOnce(context.Background()))
	require.Equal(t, 1, repo.recoverCalls)
	require.Len(t, notifier.sent, 1)
	in := notifier.sent[0]
	require.Equal(t, NotificationEmailEventRedeemInactiveCoupon, in.Event)
	require.Equal(t, notificationEmailDefaultLocale, in.Locale)
	require.Equal(t, "tom@example.com", in.RecipientEmail)
	require.Equal(t, "tom", in.RecipientName)
	require.Equal(t, int64(42), in.UserID)
	require.Empty(t, in.SourceType, "batch sends must not write per-recipient delivery keys to settings")
	require.Equal(t, map[string]string{"coupon_value": "$5.00", "redeem_code": "CODE123", "expires_at": "2026-11-14 00:00 UTC"}, in.Variables)
	require.Len(t, repo.completed, 1)
	require.Equal(t, RedeemEmailItemStatusSent, repo.completed[0].Status)
	require.Equal(t, *claim.Item.AttemptToken, repo.completed[0].AttemptToken)
	require.GreaterOrEqual(t, clock.now.Sub(start), time.Second, "new leader must wait one send interval")
}

func TestRedeemEmailWorkerSkipsUnsubscribedRecipient(t *testing.T) {
	repo := &redeemEmailWorkerRepoStub{claims: []*RedeemEmailClaim{redeemEmailTestClaim()}, available: true}
	notifier := &redeemEmailNotifierStub{unsubscribed: true}
	w, _ := newRedeemEmailWorkerForTest(repo, notifier, &redeemEmailLeaderStub{})

	require.True(t, w.RunOnce(context.Background()))
	require.Empty(t, notifier.sent)
	require.Len(t, repo.completed, 1)
	require.Equal(t, RedeemEmailItemStatusSkipped, repo.completed[0].Status)
	require.Equal(t, "RECIPIENT_UNSUBSCRIBED", repo.completed[0].ErrorCode)
}

func TestRedeemEmailWorkerFailsWhenUnsubscribeCheckErrors(t *testing.T) {
	repo := &redeemEmailWorkerRepoStub{claims: []*RedeemEmailClaim{redeemEmailTestClaim()}, available: true}
	notifier := &redeemEmailNotifierStub{unsubscribeErr: errors.New("db down")}
	w, _ := newRedeemEmailWorkerForTest(repo, notifier, &redeemEmailLeaderStub{})

	require.True(t, w.RunOnce(context.Background()))
	require.Empty(t, notifier.sent)
	require.Equal(t, RedeemEmailItemStatusFailed, repo.completed[0].Status)
	require.Equal(t, "UNSUBSCRIBE_CHECK_FAILED", repo.completed[0].ErrorCode)
}

func TestRedeemEmailWorkerFailsUnavailableCodeWithoutSending(t *testing.T) {
	repo := &redeemEmailWorkerRepoStub{claims: []*RedeemEmailClaim{redeemEmailTestClaim()}, available: false}
	notifier := &redeemEmailNotifierStub{}
	w, _ := newRedeemEmailWorkerForTest(repo, notifier, &redeemEmailLeaderStub{})

	require.True(t, w.RunOnce(context.Background()))
	require.Empty(t, notifier.sent)
	require.Len(t, repo.completed, 1)
	require.Equal(t, RedeemEmailItemStatusFailed, repo.completed[0].Status)
	require.Equal(t, "REDEEM_CODE_UNAVAILABLE", repo.completed[0].ErrorCode)
}

func TestRedeemEmailWorkerClassifiesDeliveryFailures(t *testing.T) {
	for want, err := range map[string]error{
		"SMTP_RECIPIENT_REJECTED": notificationEmailDeliveryErr(errors.New("550 recipient rejected")),
		"TEMPLATE_INVALID":        notificationEmailTemplateErr(errors.New("bad placeholder")),
		"SMTP_NOT_CONFIGURED":     notificationEmailDeliveryErr(ErrEmailNotConfigured),
	} {
		repo := &redeemEmailWorkerRepoStub{claims: []*RedeemEmailClaim{redeemEmailTestClaim()}, available: true}
		w, _ := newRedeemEmailWorkerForTest(repo, &redeemEmailNotifierStub{sendErr: err}, &redeemEmailLeaderStub{})

		require.True(t, w.RunOnce(context.Background()))
		require.Len(t, repo.completed, 1)
		require.Equal(t, RedeemEmailItemStatusFailed, repo.completed[0].Status)
		require.Equal(t, want, repo.completed[0].ErrorCode)
	}
}

func TestRedeemEmailWorkerSkipsWhenLeaderHeldByPeer(t *testing.T) {
	repo := &redeemEmailWorkerRepoStub{claims: []*RedeemEmailClaim{redeemEmailTestClaim()}, available: true}
	w, _ := newRedeemEmailWorkerForTest(repo, &redeemEmailNotifierStub{}, &redeemEmailLeaderStub{held: true})

	require.False(t, w.RunOnce(context.Background()))
	require.Zero(t, repo.claimCalls)
	require.Zero(t, repo.recoverCalls)
}

func TestRedeemEmailWorkerNeverExpiresLabelFollowsRecipientLocale(t *testing.T) {
	for locale, want := range map[string]string{notificationEmailLocaleChinese: "长期有效", notificationEmailDefaultLocale: "Never expires"} {
		claim := redeemEmailTestClaim()
		claim.CodeExpiresAt = nil
		repo := &redeemEmailWorkerRepoStub{claims: []*RedeemEmailClaim{claim}, available: true}
		notifier := &redeemEmailNotifierStub{locale: locale}
		w, _ := newRedeemEmailWorkerForTest(repo, notifier, &redeemEmailLeaderStub{})

		require.True(t, w.RunOnce(context.Background()))
		require.Equal(t, want, notifier.sent[0].Variables["expires_at"])
	}
}

func TestRedeemEmailWorkerRuntimeStartStop(t *testing.T) {
	repo := &redeemEmailWorkerRepoStub{}
	w, _ := newRedeemEmailWorkerForTest(repo, &redeemEmailNotifierStub{}, &redeemEmailLeaderStub{})
	w.clock = realRedeemEmailWorkerClock{}
	w.opts.EmptyBackoff = 10 * time.Millisecond
	rt := NewRedeemEmailWorkerRuntime(w)
	rt.Start()
	require.True(t, rt.Running())
	rt.Stop()
	require.False(t, rt.Running())
}
