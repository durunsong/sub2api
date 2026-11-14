package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSubscriptionTermQuotaDoesNotRefillBeforeExpiry(t *testing.T) {
	start := time.Date(2026, 11, 14, 3, 17, 0, 0, time.UTC)
	expiry := time.Date(2027, 2, 12, 3, 17, 0, 0, time.UTC)
	for _, elapsed := range []time.Duration{time.Hour, 30 * 24 * time.Hour, 60 * 24 * time.Hour, 90*24*time.Hour - time.Second} {
		t.Run(fmt.Sprint(elapsed), func(t *testing.T) {
			for _, anchor := range []time.Time{start, startOfDay(start), start.Add(30 * 24 * time.Hour)} {
				sub := &UserSubscription{ID: 1, Status: SubscriptionStatusActive, StartsAt: start, ExpiresAt: expiry, MonthlyWindowStart: &anchor, MonthlyUsageUSD: 10000.01, MonthlyUsageTokens: 1568040383}
				svc := NewSubscriptionService(groupRepoNoop{}, userSubRepoNoop{}, nil, nil, nil)
				t.Cleanup(svc.Stop)
				svc.now = func() time.Time { return start.Add(elapsed) }
				require.Equal(t, expiry, *sub.MonthlyResetTime())
				normalized := []UserSubscription{*sub}
				normalizeExpiredWindowsAt(normalized, svc.now())
				require.Equal(t, sub.MonthlyUsageUSD, normalized[0].MonthlyUsageUSD)
				require.Equal(t, sub.MonthlyUsageTokens, normalized[0].MonthlyUsageTokens)
				require.Equal(t, anchor, *normalized[0].MonthlyWindowStart)
				needsMaintenance, err := svc.ValidateAndCheckLimits(sub, &Group{MonthlyLimitUSD: ptrFloat64(10000)})
				require.ErrorIs(t, err, ErrMonthlyLimitExceeded)
				require.False(t, needsMaintenance)
				require.NoError(t, svc.CheckAndResetWindows(context.Background(), sub))
				require.Equal(t, 10000.01, sub.MonthlyUsageUSD)
				progress := svc.calculateProgress(sub, &Group{MonthlyLimitUSD: ptrFloat64(10000)})
				require.Equal(t, expiry, progress.Monthly.ResetsAt)
				require.Equal(t, 10000.01, progress.Monthly.UsedUSD)
			}
		})
	}
}

func TestSubscriptionTermQuotaExpiresWithoutRefilling(t *testing.T) {
	start := time.Date(2026, 11, 14, 3, 17, 0, 0, time.UTC)
	for _, days := range []int{7, 30, 90, 180} {
		t.Run(fmt.Sprint(days), func(t *testing.T) {
			expiry := start.AddDate(0, 0, days)
			sub := &UserSubscription{ID: 1, Status: SubscriptionStatusActive, StartsAt: start, ExpiresAt: expiry,
				MonthlyWindowStart: &start, MonthlyUsageUSD: 42, MonthlyUsageTokens: 123}
			svc := NewSubscriptionService(groupRepoNoop{}, userSubRepoNoop{}, nil, nil, nil)
			t.Cleanup(svc.Stop)
			svc.now = func() time.Time { return expiry }
			require.Equal(t, expiry, *sub.MonthlyResetTime())
			_, err := svc.ValidateAndCheckLimits(sub, &Group{})
			require.ErrorIs(t, err, ErrSubscriptionExpired)
			require.NoError(t, svc.CheckAndResetWindows(context.Background(), sub))
			require.Equal(t, 42.0, sub.MonthlyUsageUSD)
			require.EqualValues(t, 123, sub.MonthlyUsageTokens)
		})
	}
}
