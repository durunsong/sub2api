package service

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	redeemEmailWorkerLeaderLockKey = "redeem-email:worker:leader"
	redeemEmailCompleteTimeout     = 5 * time.Second
)

type RedeemEmailWorkerOptions struct {
	SendInterval time.Duration
	EmptyBackoff time.Duration
	Lease        time.Duration
	LeaderTTL    time.Duration
	RecoverLimit int
}

func DefaultRedeemEmailWorkerOptions() RedeemEmailWorkerOptions {
	return RedeemEmailWorkerOptions{SendInterval: time.Second, EmptyBackoff: 2 * time.Second, Lease: 45 * time.Second, LeaderTTL: 2 * time.Minute, RecoverLimit: 100}
}

type redeemEmailWorkerRepository interface {
	ClaimNext(context.Context, time.Time, time.Duration) (*RedeemEmailClaim, error)
	IsCodeAvailable(context.Context, *int64, time.Time) (bool, error)
	CompleteAttempt(context.Context, RedeemEmailAttemptResult) error
	RecoverExpiredSending(context.Context, time.Time, int) (int, error)
}

type redeemEmailWorkerClock interface {
	Now() time.Time
	Sleep(context.Context, time.Duration) bool
}

type realRedeemEmailWorkerClock struct{}

func (realRedeemEmailWorkerClock) Now() time.Time { return time.Now() }
func (realRedeemEmailWorkerClock) Sleep(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

// RedeemEmailNotifier 由 NotificationEmailService 实现：按收件人语言渲染设置页模板并投递。
type RedeemEmailNotifier interface {
	IsUnsubscribed(ctx context.Context, email, event string) (bool, error)
	ResolveRecipientLocale(ctx context.Context, userID int64, email string) string
	Send(ctx context.Context, input NotificationEmailSendInput) error
}

type RedeemEmailWorker struct {
	repo       redeemEmailWorkerRepository
	notifier   RedeemEmailNotifier
	lockCache  LeaderLockCache
	db         *sql.DB
	opts       RedeemEmailWorkerOptions
	clock      redeemEmailWorkerClock
	instanceID string
	lastSend   time.Time
}

func NewRedeemEmailWorker(repo redeemEmailWorkerRepository, notifier RedeemEmailNotifier, cache LeaderLockCache, db *sql.DB, o RedeemEmailWorkerOptions, c redeemEmailWorkerClock) *RedeemEmailWorker {
	d := DefaultRedeemEmailWorkerOptions()
	if o.SendInterval <= 0 {
		o.SendInterval = d.SendInterval
	}
	if o.EmptyBackoff <= 0 {
		o.EmptyBackoff = d.EmptyBackoff
	}
	if o.Lease <= 0 {
		o.Lease = d.Lease
	}
	if o.LeaderTTL <= 0 {
		o.LeaderTTL = d.LeaderTTL
	}
	if o.RecoverLimit <= 0 {
		o.RecoverLimit = d.RecoverLimit
	}
	if c == nil {
		c = realRedeemEmailWorkerClock{}
	}
	return &RedeemEmailWorker{repo: repo, notifier: notifier, lockCache: cache, db: db, opts: o, clock: c, instanceID: uuid.NewString()}
}

// RunOnce 在持有全局 leader 锁时至多投递一封邮件；返回 true 表示处理了一条明细。
func (w *RedeemEmailWorker) RunOnce(ctx context.Context) bool {
	release, ok := tryAcquireSingletonLeaderLock(ctx, w.lockCache, w.db, redeemEmailWorkerLeaderLockKey, w.instanceID, w.opts.LeaderTTL)
	if !ok {
		return false
	}
	defer release()
	// 新当选的 leader 先等待一个完整间隔，使实例切换时仍保持全局限速。
	acquiredAt := w.clock.Now()
	if _, err := w.repo.RecoverExpiredSending(ctx, acquiredAt, w.opts.RecoverLimit); err != nil {
		logger.L().Error("redeem_email.recover_expired_failed", zap.Error(err))
		return false
	}
	limitFrom := acquiredAt
	if w.lastSend.After(limitFrom) {
		limitFrom = w.lastSend
	}
	if wait := w.opts.SendInterval - w.clock.Now().Sub(limitFrom); wait > 0 && !w.clock.Sleep(ctx, wait) {
		return false
	}
	if ctx.Err() != nil {
		return false
	}
	claim, err := w.repo.ClaimNext(ctx, w.clock.Now(), w.opts.Lease)
	if err != nil {
		logger.L().Error("redeem_email.claim_next_failed", zap.Error(err))
		return false
	}
	if claim == nil || claim.Item.AttemptToken == nil {
		return false
	}
	result := RedeemEmailAttemptResult{ItemID: claim.Item.ID, AttemptToken: *claim.Item.AttemptToken}
	available, err := w.repo.IsCodeAvailable(ctx, claim.Item.RedeemCodeID, w.clock.Now())
	if err != nil {
		logger.L().Error("redeem_email.code_availability_check_failed", zap.Int64("item_id", claim.Item.ID), zap.Error(err))
		return false
	}
	if !available {
		result.Status = RedeemEmailItemStatusFailed
		result.ErrorCode, result.ErrorMessage = "REDEEM_CODE_UNAVAILABLE", "redeem code is used, disabled, expired or deleted"
		result.CompletedAt = w.clock.Now()
		w.completeAttempt(ctx, result)
		return true
	}
	unsubscribed, err := w.notifier.IsUnsubscribed(ctx, claim.Item.RecipientEmail, NotificationEmailEventRedeemInactiveCoupon)
	switch {
	case err != nil:
		result.Status = RedeemEmailItemStatusFailed
		result.ErrorCode, result.ErrorMessage = "UNSUBSCRIBE_CHECK_FAILED", "failed to read unsubscribe preference"
	case unsubscribed:
		result.Status = RedeemEmailItemStatusSkipped
		result.ErrorCode, result.ErrorMessage = "RECIPIENT_UNSUBSCRIBED", "recipient unsubscribed from coupon emails"
	default:
		w.deliver(ctx, claim, &result)
	}
	result.CompletedAt = w.clock.Now()
	w.completeAttempt(ctx, result)
	return true
}

func (w *RedeemEmailWorker) deliver(ctx context.Context, claim *RedeemEmailClaim, result *RedeemEmailAttemptResult) {
	var userID int64
	if claim.Item.UserID != nil {
		userID = *claim.Item.UserID
	}
	locale := w.notifier.ResolveRecipientLocale(ctx, userID, claim.Item.RecipientEmail)
	expires := redeemEmailNeverExpiresLabel(locale)
	if claim.CodeExpiresAt != nil {
		expires = FormatRedeemEmailExpiresAt(*claim.CodeExpiresAt)
	}
	w.lastSend = w.clock.Now()
	err := w.notifier.Send(ctx, NotificationEmailSendInput{
		Event:          NotificationEmailEventRedeemInactiveCoupon,
		Locale:         locale,
		RecipientEmail: claim.Item.RecipientEmail,
		RecipientName:  RedeemEmailRecipientName(claim.Item.RecipientName, claim.Item.RecipientEmail),
		UserID:         userID,
		Variables: map[string]string{
			"coupon_value": FormatRedeemEmailCouponValue(claim.CouponValue),
			"redeem_code":  claim.Item.RedeemCode,
			"expires_at":   expires,
		},
	})
	if err == nil {
		result.Status = RedeemEmailItemStatusSent
		return
	}
	f := ClassifyRedeemEmailDeliveryError(err)
	result.Status = RedeemEmailItemStatusFailed
	result.ErrorCode, result.ErrorMessage = f.Code, f.Message
}

func (w *RedeemEmailWorker) completeAttempt(ctx context.Context, result RedeemEmailAttemptResult) {
	completeCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), redeemEmailCompleteTimeout)
	defer cancel()
	if err := w.repo.CompleteAttempt(completeCtx, result); err != nil {
		logger.L().Error("redeem_email.complete_attempt_failed",
			zap.Int64("item_id", result.ItemID),
			zap.String("status", result.Status),
			zap.String("error_code", result.ErrorCode),
			zap.Error(err),
		)
	}
}

func (w *RedeemEmailWorker) Run(ctx context.Context) {
	for ctx.Err() == nil {
		if !w.RunOnce(ctx) && !w.clock.Sleep(ctx, w.opts.EmptyBackoff) {
			return
		}
	}
}

type RedeemEmailWorkerRuntime struct {
	worker *RedeemEmailWorker
	mu     sync.Mutex
	cancel context.CancelFunc
	done   chan struct{}
}

func NewRedeemEmailWorkerRuntime(w *RedeemEmailWorker) *RedeemEmailWorkerRuntime {
	return &RedeemEmailWorkerRuntime{worker: w}
}

func (r *RedeemEmailWorkerRuntime) Start() {
	if r == nil || r.worker == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.cancel != nil {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	r.cancel, r.done = cancel, done
	go func() {
		defer close(done)
		r.worker.Run(ctx)
	}()
}

func (r *RedeemEmailWorkerRuntime) Stop() {
	if r == nil {
		return
	}
	r.mu.Lock()
	cancel, done := r.cancel, r.done
	r.cancel, r.done = nil, nil
	r.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	if done != nil {
		<-done
	}
}

func (r *RedeemEmailWorkerRuntime) Running() bool {
	if r == nil {
		return false
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.cancel != nil
}

func ProvideRedeemEmailWorkerRuntime(repo RedeemEmailRepository, notificationEmail *NotificationEmailService, cache LeaderLockCache, db *sql.DB) *RedeemEmailWorkerRuntime {
	r := NewRedeemEmailWorkerRuntime(NewRedeemEmailWorker(repo, notificationEmail, cache, db, DefaultRedeemEmailWorkerOptions(), nil))
	r.Start()
	return r
}
