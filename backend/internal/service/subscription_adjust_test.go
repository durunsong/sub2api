//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAdjustSubscription_LegacyAlignPreservesQuotaAnchor(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	dailyStart := now.Add(-6 * time.Hour)
	weeklyStart := now.Add(-2 * 24 * time.Hour)
	monthlyStart := now.Add(-3 * 24 * time.Hour)
	originalExpiry := now.AddDate(0, 0, 30)

	baseRepo := newSubscriptionUserSubRepoStub()
	baseRepo.seed(&UserSubscription{
		ID:                 1,
		UserID:             10,
		GroupID:            20,
		Status:             SubscriptionStatusActive,
		ExpiresAt:          originalExpiry,
		DailyWindowStart:   &dailyStart,
		WeeklyWindowStart:  &weeklyStart,
		MonthlyWindowStart: &monthlyStart,
	})
	svc := NewSubscriptionService(groupRepoNoop{}, baseRepo, nil, nil, nil)
	t.Cleanup(svc.Stop)

	got, err := svc.AdjustSubscription(context.Background(), 1, 60, true)

	require.NoError(t, err)
	require.Equal(t, originalExpiry.AddDate(0, 0, 60), got.ExpiresAt)
	require.NotNil(t, got.MonthlyWindowStart)
	require.Equal(t, monthlyStart, *got.MonthlyWindowStart)
	require.Equal(t, got.ExpiresAt, *got.MonthlyResetTime())
	require.Equal(t, dailyStart, *got.DailyWindowStart)
	require.Equal(t, weeklyStart, *got.WeeklyWindowStart)
}

func TestAdjustSubscription_DefaultMovesQuotaEndWithExpiry(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	monthlyStart := now.Add(-3 * 24 * time.Hour)
	originalExpiry := now.AddDate(0, 0, 30)

	baseRepo := newSubscriptionUserSubRepoStub()
	baseRepo.seed(&UserSubscription{
		ID:                 2,
		UserID:             10,
		GroupID:            20,
		Status:             SubscriptionStatusActive,
		ExpiresAt:          originalExpiry,
		MonthlyWindowStart: &monthlyStart,
	})
	svc := NewSubscriptionService(groupRepoNoop{}, baseRepo, nil, nil, nil)
	t.Cleanup(svc.Stop)

	got, err := svc.AdjustSubscription(context.Background(), 2, 60, false)

	require.NoError(t, err)
	require.Equal(t, originalExpiry.AddDate(0, 0, 60), got.ExpiresAt)
	require.NotNil(t, got.MonthlyWindowStart)
	require.Equal(t, monthlyStart, *got.MonthlyWindowStart)
	require.Equal(t, got.ExpiresAt, *got.MonthlyResetTime())
}

func TestAdjustSubscription_LegacyAlignWithoutDaysPreservesAnchor(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	monthlyStart := now.Add(-4 * 24 * time.Hour)
	originalExpiry := now.AddDate(0, 0, 87)

	baseRepo := newSubscriptionUserSubRepoStub()
	baseRepo.seed(&UserSubscription{
		ID:                 3,
		UserID:             10,
		GroupID:            20,
		Status:             SubscriptionStatusActive,
		ExpiresAt:          originalExpiry,
		MonthlyWindowStart: &monthlyStart,
	})
	svc := NewSubscriptionService(groupRepoNoop{}, baseRepo, nil, nil, nil)
	t.Cleanup(svc.Stop)

	got, err := svc.AdjustSubscription(context.Background(), 3, 0, true)

	require.NoError(t, err)
	require.Equal(t, originalExpiry, got.ExpiresAt)
	require.NotNil(t, got.MonthlyWindowStart)
	require.Equal(t, monthlyStart, *got.MonthlyWindowStart)
	require.Equal(t, got.ExpiresAt, *got.MonthlyResetTime())
}

func TestAdjustSubscription_ReactivatesWithoutRefillingTermQuota(t *testing.T) {
	now := time.Date(2026, 11, 14, 3, 17, 0, 0, time.UTC)
	anchor := now.Add(-91 * 24 * time.Hour)
	for _, legacyAlign := range []bool{false, true} {
		repo := newSubscriptionUserSubRepoStub()
		repo.seed(&UserSubscription{ID: 4, UserID: 10, GroupID: 20, Status: SubscriptionStatusExpired,
			ExpiresAt: now.Add(-24 * time.Hour), MonthlyWindowStart: &anchor, MonthlyUsageUSD: 9000, MonthlyUsageTokens: 1200})
		svc := NewSubscriptionService(groupRepoNoop{}, repo, nil, nil, nil)
		t.Cleanup(svc.Stop)
		svc.now = func() time.Time { return now }
		got, err := svc.AdjustSubscription(context.Background(), 4, 90, legacyAlign)
		require.NoError(t, err)
		require.Equal(t, SubscriptionStatusActive, got.Status)
		require.Equal(t, now.AddDate(0, 0, 90), got.ExpiresAt)
		require.Equal(t, got.ExpiresAt, *got.MonthlyResetTime())
		require.Equal(t, anchor, *got.MonthlyWindowStart)
		require.Equal(t, 9000.0, got.MonthlyUsageUSD)
		require.EqualValues(t, 1200, got.MonthlyUsageTokens)
	}
}
