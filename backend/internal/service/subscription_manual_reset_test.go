//go:build unit

package service

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestAssignOrExtendSubscription_OrdinaryDailyAssignmentExtendsWithoutGrantingResetCard(t *testing.T) {
	groupRepo := &subscriptionGroupRepoStub{
		group: &Group{ID: 1, SubscriptionType: SubscriptionTypeSubscription},
	}
	subRepo := newSubscriptionUserSubRepoStub()
	// ponytail: 用相对时间，避免硬编码日期过期后把「未过期再买」误判成续订清零
	boughtAt := time.Now().UTC().Add(-6 * time.Hour)
	oldExpires := boughtAt.AddDate(0, 0, 1)
	windowStart := startOfDay(boughtAt)
	subRepo.seed(&UserSubscription{
		ID:               100,
		UserID:           200,
		GroupID:          1,
		StartsAt:         boughtAt,
		ExpiresAt:        oldExpires,
		Status:           SubscriptionStatusActive,
		DailyWindowStart: &windowStart,
		DailyUsageUSD:    50,
		DailyUsageTokens: 1234,
		Notes:            "first",
	})
	svc := NewSubscriptionService(groupRepo, subRepo, nil, nil, nil)

	renewed, reused, err := svc.AssignOrExtendSubscription(context.Background(), &AssignSubscriptionInput{
		UserID:       200,
		GroupID:      1,
		ValidityDays: 1,
		Notes:        "second",
	})

	require.NoError(t, err)
	require.True(t, reused)
	require.Equal(t, int64(100), renewed.ID)
	require.Equal(t, 50.0, renewed.DailyUsageUSD, "普通分配不应清零用量")
	require.Equal(t, int64(1234), renewed.DailyUsageTokens)
	require.Zero(t, renewed.ManualResetCredits, "普通分配不得凭空增加重置卡权益")
	require.Equal(t, boughtAt, renewed.StartsAt)
	require.Equal(t, oldExpires.AddDate(0, 0, 1), renewed.ExpiresAt)
	require.Equal(t, "first\nsecond", renewed.Notes)
}

func TestAssignOrExtendSubscription_OrdinaryMultiDayAssignmentExtendsWithoutResetCredit(t *testing.T) {
	groupRepo := &subscriptionGroupRepoStub{
		group: &Group{ID: 1, SubscriptionType: SubscriptionTypeSubscription},
	}
	subRepo := newSubscriptionUserSubRepoStub()
	start := time.Now().Add(-12 * time.Hour)
	oldExpires := start.AddDate(0, 0, 30)
	subRepo.seed(&UserSubscription{
		ID:        101,
		UserID:    201,
		GroupID:   1,
		StartsAt:  start,
		ExpiresAt: oldExpires,
		Status:    SubscriptionStatusActive,
	})
	svc := NewSubscriptionService(groupRepo, subRepo, nil, nil, nil)

	renewed, reused, err := svc.AssignOrExtendSubscription(context.Background(), &AssignSubscriptionInput{
		UserID:       201,
		GroupID:      1,
		ValidityDays: 30,
	})

	require.NoError(t, err)
	require.True(t, reused)
	require.Equal(t, start, renewed.StartsAt, "多日卡再买不应改 starts_at")
	require.WithinDuration(t, oldExpires.AddDate(0, 0, 30), renewed.ExpiresAt, time.Second)
	require.Zero(t, renewed.ManualResetCredits)
}

func TestUserResetDailyQuota_ActiveOneTimeCardDoesNotRestartTerm(t *testing.T) {
	groupRepo := &subscriptionGroupRepoStub{
		group: &Group{ID: 1, SubscriptionType: SubscriptionTypeSubscription},
	}
	subRepo := newSubscriptionUserSubRepoStub()
	boughtAt := time.Now().UTC().Add(-6 * time.Hour)
	windowStart := startOfDay(boughtAt)
	subRepo.seed(&UserSubscription{
		ID:                 102,
		UserID:             202,
		GroupID:            1,
		StartsAt:           boughtAt,
		ExpiresAt:          boughtAt.AddDate(0, 0, 1),
		Status:             SubscriptionStatusActive,
		DailyWindowStart:   &windowStart,
		DailyUsageUSD:      50,
		DailyUsageTokens:   99,
		ManualResetCredits: 1,
	})
	svc := NewSubscriptionService(groupRepo, subRepo, nil, nil, nil)

	got, err := svc.UserResetDailyQuota(context.Background(), 202, 102)

	require.NoError(t, err)
	require.Equal(t, 1, got.CreditsBefore)
	require.Equal(t, 0, got.CreditsAfter)
	require.Equal(t, 0, got.Subscription.ManualResetCredits)
	require.Equal(t, 0.0, got.Subscription.DailyUsageUSD)
	require.Equal(t, int64(0), got.Subscription.DailyUsageTokens)
	require.Equal(t, boughtAt, got.Subscription.StartsAt, "有效日卡重置不得改 starts_at")
	require.Equal(t, boughtAt.AddDate(0, 0, 1), got.Subscription.ExpiresAt, "有效日卡重置不得改 expires_at")
	require.True(t, got.Subscription.IsActive())

	_, err = svc.UserResetDailyQuota(context.Background(), 202, 102)
	require.True(t, errors.Is(err, ErrManualResetNoCredits), "次数用尽后不可再重置")
}

func TestUserResetDailyQuota_ExpiredOneTimeCardWithoutCreditsIsRejected(t *testing.T) {
	now := time.Now()
	subRepo := newSubscriptionUserSubRepoStub()
	subRepo.seed(&UserSubscription{
		ID: 105, UserID: 205, GroupID: 1,
		StartsAt: now.Add(-48 * time.Hour), ExpiresAt: now.Add(-24 * time.Hour),
		Status: SubscriptionStatusExpired, ManualResetCredits: 0,
	})
	svc := NewSubscriptionService(groupRepoNoop{}, subRepo, nil, nil, nil)
	svc.now = func() time.Time { return now }

	_, err := svc.UserResetDailyQuota(context.Background(), 205, 105)

	require.ErrorIs(t, err, ErrManualResetNoCredits)
}
func TestUserResetDailyQuota_RedeemPendingCreditAfterExpiry(t *testing.T) {
	groupRepo := &subscriptionGroupRepoStub{
		group: &Group{ID: 1, SubscriptionType: SubscriptionTypeSubscription},
	}
	subRepo := newSubscriptionUserSubRepoStub()
	start := time.Now().Add(-48 * time.Hour)
	subRepo.seed(&UserSubscription{
		ID:                 103,
		UserID:             203,
		GroupID:            1,
		StartsAt:           start,
		ExpiresAt:          start.Add(24 * time.Hour), // already expired; still one-time span
		Status:             SubscriptionStatusExpired,
		DailyUsageUSD:      50,
		ManualResetCredits: 1, // paid pending activation from repurchase
	})
	svc := NewSubscriptionService(groupRepo, subRepo, nil, nil, nil)

	got, err := svc.UserResetDailyQuota(context.Background(), 203, 103)
	require.NoError(t, err)
	require.Equal(t, 0, got.Subscription.ManualResetCredits)
	require.Equal(t, 0.0, got.Subscription.DailyUsageUSD)
	require.Equal(t, SubscriptionStatusActive, got.Subscription.Status)
	require.True(t, got.Subscription.IsActive())
	require.True(t, got.Subscription.HasOneTimeDailyQuota())
}

func TestUserResetDailyQuota_NotFoundPathsUseFixedUnknownCredits(t *testing.T) {
	now := time.Now()
	deletedAt := now.Add(-time.Minute)
	for _, tc := range []struct {
		name string
		sub  *UserSubscription
	}{
		{name: "foreign", sub: &UserSubscription{ID: 106, UserID: 999, StartsAt: now.Add(-time.Hour), ExpiresAt: now.Add(time.Hour), Status: SubscriptionStatusActive, ManualResetCredits: 73}},
		{name: "soft deleted", sub: &UserSubscription{ID: 106, UserID: 206, StartsAt: now.Add(-time.Hour), ExpiresAt: now.Add(time.Hour), Status: SubscriptionStatusActive, ManualResetCredits: 89, DeletedAt: &deletedAt}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := newSubscriptionUserSubRepoStub()
			repo.seed(tc.sub)
			svc := NewSubscriptionService(groupRepoNoop{}, repo, nil, nil, nil)
			svc.now = func() time.Time { return now }

			_, err := svc.UserResetDailyQuota(context.Background(), 206, 106)

			require.ErrorIs(t, err, ErrSubscriptionNotFound)
			status := infraerrors.FromError(err)
			require.Equal(t, "-1", status.Metadata["credits_before"])
			require.Equal(t, "-1", status.Metadata["credits_after"])
		})
	}
}
func TestUserResetDailyQuota_RejectsForeignOrMultiDayExpired(t *testing.T) {
	groupRepo := &subscriptionGroupRepoStub{
		group: &Group{ID: 1, SubscriptionType: SubscriptionTypeSubscription},
	}
	subRepo := newSubscriptionUserSubRepoStub()
	now := time.Now()
	subRepo.seed(&UserSubscription{
		ID:                 104,
		UserID:             204,
		GroupID:            1,
		StartsAt:           now.AddDate(0, 0, -40),
		ExpiresAt:          now.Add(-time.Hour),
		Status:             SubscriptionStatusExpired,
		ManualResetCredits: 2,
	})
	svc := NewSubscriptionService(groupRepo, subRepo, nil, nil, nil)

	_, err := svc.UserResetDailyQuota(context.Background(), 999, 104)
	require.True(t, errors.Is(err, ErrSubscriptionNotFound))

	_, err = svc.UserResetDailyQuota(context.Background(), 204, 104)
	require.True(t, errors.Is(err, ErrManualResetStateNotAllowed), "过期状态的普通月卡不能靠重置复活")
}

type atomicResetRepoStub struct {
	userSubRepoNoop
	snapshots  []*UserSubscription
	result     ManualDailyResetResult
	err        error
	request    ManualDailyResetRequest
	getCalls   int
	resetCalls int
	getErrors  []error
}

func (r *atomicResetRepoStub) GetByID(_ context.Context, id int64) (*UserSubscription, error) {
	r.getCalls++
	index := r.getCalls - 1
	if index < len(r.getErrors) && r.getErrors[index] != nil {
		return nil, r.getErrors[index]
	}
	if index >= len(r.snapshots) {
		index = len(r.snapshots) - 1
	}
	if index < 0 || r.snapshots[index] == nil || r.snapshots[index].ID != id {
		return nil, ErrSubscriptionNotFound
	}
	cp := *r.snapshots[index]
	return &cp, nil
}

func (r *atomicResetRepoStub) ResetDailyQuota(_ context.Context, request ManualDailyResetRequest) (ManualDailyResetResult, error) {
	r.resetCalls++
	r.request = request
	return r.result, r.err
}

func TestUserResetDailyQuota_ReclassifiesChangedSnapshotAfterAtomicMiss(t *testing.T) {
	now := time.Now()
	initial := &UserSubscription{ID: 200, UserID: 20, StartsAt: now.Add(-24 * time.Hour), ExpiresAt: now.Add(24 * time.Hour), Status: SubscriptionStatusActive, ManualResetCredits: 1}
	tests := []struct {
		name   string
		latest *UserSubscription
		want   error
	}{
		{name: "deleted or foreign", latest: nil, want: ErrSubscriptionNotFound},
		{name: "no credits", latest: &UserSubscription{ID: 200, UserID: 20, StartsAt: initial.StartsAt, ExpiresAt: initial.ExpiresAt, Status: SubscriptionStatusActive}, want: ErrManualResetNoCredits},
		{name: "suspended", latest: &UserSubscription{ID: 200, UserID: 20, StartsAt: initial.StartsAt, ExpiresAt: initial.ExpiresAt, Status: SubscriptionStatusSuspended, ManualResetCredits: 1}, want: ErrManualResetStateNotAllowed},
		{name: "expired monthly", latest: &UserSubscription{ID: 200, UserID: 20, StartsAt: now.Add(-31 * 24 * time.Hour), ExpiresAt: now.Add(-time.Hour), Status: SubscriptionStatusActive, ManualResetCredits: 1}, want: ErrManualResetSubscriptionExpired},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &atomicResetRepoStub{snapshots: []*UserSubscription{initial, tt.latest}, err: ErrManualResetConflict}
			svc := NewSubscriptionService(groupRepoNoop{}, repo, nil, nil, nil)
			svc.now = func() time.Time { return now }
			_, err := svc.UserResetDailyQuota(context.Background(), 20, 200)
			require.ErrorIs(t, err, tt.want)
			if tt.latest != nil {
				status := infraerrors.FromError(err)
				wantCredits := strconv.Itoa(tt.latest.ManualResetCredits)
				require.Equal(t, wantCredits, status.Metadata["credits_before"])
				require.Equal(t, wantCredits, status.Metadata["credits_after"])
			}
			require.Equal(t, 2, repo.getCalls)
			require.Equal(t, 1, repo.resetCalls)
		})
	}
}

func TestUserResetDailyQuota_StillEligibleAfterAtomicMissReturnsConflict(t *testing.T) {
	now := time.Now()
	sub := &UserSubscription{ID: 200, UserID: 20, StartsAt: now.Add(-24 * time.Hour), ExpiresAt: now.Add(24 * time.Hour), Status: SubscriptionStatusActive, ManualResetCredits: 1}
	repo := &atomicResetRepoStub{snapshots: []*UserSubscription{sub, sub}, err: ErrManualResetConflict}
	svc := NewSubscriptionService(groupRepoNoop{}, repo, nil, nil, nil)
	svc.now = func() time.Time { return now }

	_, err := svc.UserResetDailyQuota(context.Background(), 20, 200)

	require.ErrorIs(t, err, ErrManualResetConflict)
	require.Equal(t, 2, repo.getCalls)
	require.Equal(t, 1, repo.resetCalls)
}
func TestUserResetDailyQuota_AppliedMutationSurvivesReloadFailure(t *testing.T) {
	now := time.Now()
	sub := &UserSubscription{ID: 300, UserID: 30, StartsAt: now.Add(-time.Hour), ExpiresAt: now.Add(time.Hour), Status: SubscriptionStatusActive, ManualResetCredits: 2}
	reloadErr := errors.New("reload failed")
	repo := &atomicResetRepoStub{
		snapshots: []*UserSubscription{sub},
		result:    ManualDailyResetResult{MutationApplied: true, CreditsBefore: 2, CreditsAfter: 1},
	}
	repo.getErrors = []error{nil, reloadErr}
	svc := NewSubscriptionService(groupRepoNoop{}, repo, nil, nil, nil)
	svc.now = func() time.Time { return now }

	result, err := svc.UserResetDailyQuota(context.Background(), 30, 300)

	require.ErrorIs(t, err, ErrManualResetResponseReloadFailed)
	require.NotNil(t, result)
	require.True(t, result.MutationApplied)
	require.Equal(t, 2, result.CreditsBefore)
	require.Equal(t, 1, result.CreditsAfter)
	require.Nil(t, result.Subscription)
}
func TestManualDailyResetRequestContractCarriesAtomicEligibility(t *testing.T) {
	now := time.Now()
	req := ManualDailyResetRequest{
		SubscriptionID: 1,
		UserID:         2,
		Now:            now,
		WindowStart:    startOfDay(now),
		RestartTerm:    true,
		NewStartsAt:    now,
		NewExpiresAt:   now.Add(24 * time.Hour),
	}
	require.Equal(t, int64(1), req.SubscriptionID)
	require.Equal(t, int64(2), req.UserID)
}
