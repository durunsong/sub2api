//go:build unit

package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAssignOrExtendSubscription_ActiveDailyCardRestartsFromNowAndGrantsResetCredit(t *testing.T) {
	groupRepo := &subscriptionGroupRepoStub{
		group: &Group{ID: 1, SubscriptionType: SubscriptionTypeSubscription},
	}
	subRepo := newSubscriptionUserSubRepoStub()
	boughtAt := time.Date(2026, 7, 12, 10, 0, 0, 0, time.UTC)
	windowStart := startOfDay(boughtAt)
	subRepo.seed(&UserSubscription{
		ID:               100,
		UserID:           200,
		GroupID:          1,
		StartsAt:         boughtAt,
		ExpiresAt:        boughtAt.AddDate(0, 0, 1), // tomorrow 10:00
		Status:           SubscriptionStatusActive,
		DailyWindowStart: &windowStart,
		DailyUsageUSD:    50,
		DailyUsageTokens: 1234,
		Notes:            "first",
	})
	svc := NewSubscriptionService(groupRepo, subRepo, nil, nil, nil)

	before := time.Now()
	renewed, reused, err := svc.AssignOrExtendSubscription(context.Background(), &AssignSubscriptionInput{
		UserID:       200,
		GroupID:      1,
		ValidityDays: 1,
		Notes:        "second",
	})
	after := time.Now()

	require.NoError(t, err)
	require.True(t, reused)
	require.Equal(t, int64(100), renewed.ID)
	require.Equal(t, 50.0, renewed.DailyUsageUSD, "再买日卡不应自动清零用量")
	require.Equal(t, int64(1234), renewed.DailyUsageTokens)
	require.Equal(t, 1, renewed.ManualResetCredits, "再买应赠送 1 次手动重置")
	require.True(t, renewed.HasOneTimeDailyQuota(), "重算后仍应是一次性日额度")
	require.False(t, renewed.StartsAt.Before(before))
	require.False(t, renewed.StartsAt.After(after))
	require.WithinDuration(t, renewed.StartsAt.AddDate(0, 0, 1), renewed.ExpiresAt, time.Second)
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

func TestUserResetDailyQuota_ConsumesCreditAtomically(t *testing.T) {
	groupRepo := &subscriptionGroupRepoStub{
		group: &Group{ID: 1, SubscriptionType: SubscriptionTypeSubscription},
	}
	subRepo := newSubscriptionUserSubRepoStub()
	now := time.Now()
	windowStart := startOfDay(now)
	subRepo.seed(&UserSubscription{
		ID:                 102,
		UserID:             202,
		GroupID:            1,
		StartsAt:           now,
		ExpiresAt:          now.Add(24 * time.Hour),
		Status:             SubscriptionStatusActive,
		DailyWindowStart:   &windowStart,
		DailyUsageUSD:      50,
		DailyUsageTokens:   99,
		ManualResetCredits: 1,
	})
	svc := NewSubscriptionService(groupRepo, subRepo, nil, nil, nil)

	got, err := svc.UserResetDailyQuota(context.Background(), 202, 102)
	require.NoError(t, err)
	require.Equal(t, 0, got.ManualResetCredits)
	require.Equal(t, 0.0, got.DailyUsageUSD)
	require.Equal(t, int64(0), got.DailyUsageTokens)

	_, err = svc.UserResetDailyQuota(context.Background(), 202, 102)
	require.True(t, errors.Is(err, ErrManualResetNoCredits), "次数用尽后不可再重置")
}

func TestUserResetDailyQuota_RejectsForeignOrInactive(t *testing.T) {
	groupRepo := &subscriptionGroupRepoStub{
		group: &Group{ID: 1, SubscriptionType: SubscriptionTypeSubscription},
	}
	subRepo := newSubscriptionUserSubRepoStub()
	now := time.Now()
	subRepo.seed(&UserSubscription{
		ID:                 103,
		UserID:             203,
		GroupID:            1,
		StartsAt:           now.Add(-48 * time.Hour),
		ExpiresAt:          now.Add(-1 * time.Hour),
		Status:             SubscriptionStatusExpired,
		ManualResetCredits: 2,
	})
	svc := NewSubscriptionService(groupRepo, subRepo, nil, nil, nil)

	_, err := svc.UserResetDailyQuota(context.Background(), 999, 103)
	require.True(t, errors.Is(err, ErrSubscriptionNotFound))

	_, err = svc.UserResetDailyQuota(context.Background(), 203, 103)
	require.True(t, errors.Is(err, ErrManualResetNotAllowed))
}
