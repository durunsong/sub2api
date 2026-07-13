//go:build unit

package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAssignOrExtendSubscription_ActiveDailyCardGrantsPendingResetWithoutRestartingClock(t *testing.T) {
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
	require.Equal(t, 50.0, renewed.DailyUsageUSD, "再买日卡不应自动清零用量")
	require.Equal(t, int64(1234), renewed.DailyUsageTokens)
	require.Equal(t, 1, renewed.ManualResetCredits, "再买应发放 1 次付费重置机会")
	require.Equal(t, boughtAt, renewed.StartsAt, "再买不改 starts_at，24h 应从点击重置起算")
	require.Equal(t, oldExpires, renewed.ExpiresAt, "再买不改 expires_at")
	require.Equal(t, "first\nsecond", renewed.Notes)
}

func TestAssignOrExtendSubscription_ActiveMultiDayGrantsResetCreditAndExtends(t *testing.T) {
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
	require.Equal(t, 1, renewed.ManualResetCredits)
}

func TestUserResetDailyQuota_StartsFresh24hFromClick(t *testing.T) {
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

	before := time.Now()
	got, err := svc.UserResetDailyQuota(context.Background(), 202, 102)
	after := time.Now()

	require.NoError(t, err)
	require.Equal(t, 0, got.ManualResetCredits)
	require.Equal(t, 0.0, got.DailyUsageUSD)
	require.Equal(t, int64(0), got.DailyUsageTokens)
	require.False(t, got.StartsAt.Before(before))
	require.False(t, got.StartsAt.After(after))
	require.WithinDuration(t, got.StartsAt.AddDate(0, 0, 1), got.ExpiresAt, time.Second)
	require.True(t, got.IsActive())

	_, err = svc.UserResetDailyQuota(context.Background(), 202, 102)
	require.True(t, errors.Is(err, ErrManualResetNoCredits), "次数用尽后不可再重置")
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
	require.Equal(t, 0, got.ManualResetCredits)
	require.Equal(t, 0.0, got.DailyUsageUSD)
	require.Equal(t, SubscriptionStatusActive, got.Status)
	require.True(t, got.IsActive())
	require.True(t, got.HasOneTimeDailyQuota())
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
	require.True(t, errors.Is(err, ErrManualResetNotAllowed), "多日卡过期后不能仅靠重置次数复活")
}
