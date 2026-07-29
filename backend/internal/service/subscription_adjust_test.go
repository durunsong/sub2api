//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type adjustmentUserSubRepoStub struct {
	*subscriptionUserSubRepoStub
}

func (s *adjustmentUserSubRepoStub) AdjustExpiryAndMonthlyWindow(
	_ context.Context,
	subscriptionID int64,
	newExpiresAt time.Time,
	newMonthlyWindowStart *time.Time,
	reactivate bool,
) error {
	sub := s.byID[subscriptionID]
	if sub == nil {
		return ErrSubscriptionNotFound
	}
	sub.ExpiresAt = newExpiresAt
	sub.MonthlyWindowStart = newMonthlyWindowStart
	if reactivate {
		sub.Status = SubscriptionStatusActive
	}
	return nil
}

func TestAdjustSubscription_ShiftsMonthlyResetWithExpiry(t *testing.T) {
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
	repo := &adjustmentUserSubRepoStub{subscriptionUserSubRepoStub: baseRepo}
	svc := NewSubscriptionService(groupRepoNoop{}, repo, nil, nil, nil)

	got, err := svc.AdjustSubscription(context.Background(), 1, 60, true)

	require.NoError(t, err)
	require.Equal(t, originalExpiry.AddDate(0, 0, 60), got.ExpiresAt)
	require.NotNil(t, got.MonthlyWindowStart)
	require.Equal(t, monthlyStart.AddDate(0, 0, 60), *got.MonthlyWindowStart)
	require.Equal(t, dailyStart, *got.DailyWindowStart)
	require.Equal(t, weeklyStart, *got.WeeklyWindowStart)
}

func TestAdjustSubscription_DefaultKeepsMonthlyResetDate(t *testing.T) {
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
	repo := &adjustmentUserSubRepoStub{subscriptionUserSubRepoStub: baseRepo}
	svc := NewSubscriptionService(groupRepoNoop{}, repo, nil, nil, nil)

	got, err := svc.AdjustSubscription(context.Background(), 2, 60, false)

	require.NoError(t, err)
	require.Equal(t, originalExpiry.AddDate(0, 0, 60), got.ExpiresAt)
	require.NotNil(t, got.MonthlyWindowStart)
	require.Equal(t, monthlyStart, *got.MonthlyWindowStart)
}
