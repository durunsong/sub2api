//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestConsumeResetCard_NinetyDayQuotaEndsWithTerm(t *testing.T) {
	svc, userID, subscriptionID := createConcurrentResetCardFixture(t, 90)
	t.Cleanup(func() {
		_, err := integrationDB.ExecContext(context.Background(), "DELETE FROM user_subscription_quota_resets WHERE user_subscription_id = $1", subscriptionID)
		require.NoError(t, err)
	})
	ctx := context.Background()
	repo := NewUserSubscriptionRepository(integrationEntClient)
	require.NoError(t, repo.IncrementUsage(ctx, subscriptionID, 9795.45, 1568040383))
	var before, after time.Time
	require.NoError(t, integrationDB.QueryRowContext(ctx, "SELECT clock_timestamp()").Scan(&before))
	input := service.ConsumeResetCardInput{
		UserID: userID, SubscriptionID: subscriptionID, ValidityDays: 90,
		IdempotencyKey: "ninety-day-reset-timing",
	}
	sub, err := svc.ConsumeResetCard(ctx, input)
	require.NoError(t, err)
	require.False(t, sub.StartsAt.Before(before))
	require.NoError(t, integrationDB.QueryRowContext(ctx, "SELECT clock_timestamp()").Scan(&after))
	require.False(t, sub.StartsAt.After(after))
	require.Equal(t, 90*24*time.Hour, sub.ExpiresAt.Sub(sub.StartsAt))
	require.Equal(t, 90*24*time.Hour, sub.MonthlyResetTime().Sub(sub.StartsAt))
	require.Equal(t, 7*24*time.Hour, sub.WeeklyResetTime().Sub(sub.StartsAt))
	for _, start := range []*time.Time{sub.DailyWindowStart, sub.WeeklyWindowStart, sub.MonthlyWindowStart} {
		require.NotNil(t, start)
		require.True(t, start.Equal(sub.StartsAt))
	}
	require.Zero(t, sub.DailyUsageUSD)
	require.Zero(t, sub.WeeklyUsageUSD)
	require.Zero(t, sub.MonthlyUsageUSD)
	require.Zero(t, sub.DailyUsageTokens)
	require.Zero(t, sub.WeeklyUsageTokens)
	require.Zero(t, sub.MonthlyUsageTokens)
	require.Zero(t, sub.ManualResetCredits)

	// Retrying the same consumption must preserve both the term and new usage.
	require.NoError(t, repo.IncrementUsage(ctx, subscriptionID, 1.25, 100))
	replayed, err := svc.ConsumeResetCard(ctx, input)
	require.NoError(t, err)
	require.True(t, replayed.StartsAt.Equal(sub.StartsAt))
	require.True(t, replayed.ExpiresAt.Equal(sub.ExpiresAt))
	require.True(t, replayed.MonthlyWindowStart.Equal(*sub.MonthlyWindowStart))
	require.Equal(t, 1.25, replayed.MonthlyUsageUSD)
	require.EqualValues(t, 100, replayed.MonthlyUsageTokens)
	require.Zero(t, replayed.ManualResetCredits)

	// Both admin entry points clear quota without restarting the paid term.
	reset, err := svc.AdminResetQuota(ctx, subscriptionID, false, false, true)
	require.NoError(t, err)
	require.True(t, reset.ExpiresAt.Equal(sub.ExpiresAt))
	require.True(t, reset.MonthlyResetTime().Equal(sub.ExpiresAt))
	require.Zero(t, reset.MonthlyUsageUSD)
	require.Zero(t, reset.MonthlyUsageTokens)
	require.Equal(t, 1.25, reset.WeeklyUsageUSD)
	require.NoError(t, repo.IncrementUsage(ctx, subscriptionID, 2.5, 200))
	bulk, err := svc.BulkSubscriptionAction(ctx, &service.BulkSubscriptionActionInput{SubscriptionIDs: []int64{subscriptionID}, Action: "reset_quota", Monthly: true})
	require.NoError(t, err)
	require.Equal(t, 1, bulk.SuccessCount)
	reset, err = repo.GetByID(ctx, subscriptionID)
	require.NoError(t, err)
	require.True(t, reset.ExpiresAt.Equal(sub.ExpiresAt))
	require.True(t, reset.MonthlyResetTime().Equal(sub.ExpiresAt))
	require.Zero(t, reset.MonthlyUsageUSD)
	require.Zero(t, reset.MonthlyUsageTokens)
	history, page, err := svc.ListResetHistory(ctx, subscriptionID, &userID, 1, 20)
	require.NoError(t, err)
	require.EqualValues(t, 3, page.Total)
	require.Len(t, history, 3)

	// Adjusting the expiry must not manufacture a future quota anchor or refill usage.
	require.NoError(t, repo.IncrementUsage(ctx, subscriptionID, 3.5, 300))
	adjusted, err := svc.AdjustSubscription(ctx, subscriptionID, 60, true)
	require.NoError(t, err)
	require.True(t, adjusted.ExpiresAt.Equal(reset.ExpiresAt.AddDate(0, 0, 60)))
	require.True(t, adjusted.MonthlyWindowStart.Equal(*reset.MonthlyWindowStart))
	require.True(t, adjusted.MonthlyResetTime().Equal(adjusted.ExpiresAt))
	require.Equal(t, 3.5, adjusted.MonthlyUsageUSD)
	require.EqualValues(t, 300, adjusted.MonthlyUsageTokens)
}

func (s *UserSubscriptionRepoSuite) TestNinetyDayTermQuotaRemainsExhaustedAfterThirtyDays() {
	u := s.mustCreateUser("term-quota@test.com", "")
	g := s.mustCreateGroup("term-quota")
	_, err := s.client.Group.UpdateOneID(g.ID).SetMonthlyLimitUsd(10000).Save(s.ctx)
	s.Require().NoError(err)
	now := time.Now().UTC().Truncate(time.Microsecond)
	start := now.Add(-61 * 24 * time.Hour)
	expiry := start.Add(90 * 24 * time.Hour)
	sub := s.mustCreateSubscription(u.ID, g.ID, func(c *dbent.UserSubscriptionCreate) {
		c.SetStartsAt(start).SetExpiresAt(expiry).SetMonthlyWindowStart(start).
			SetMonthlyUsageUsd(10000.01).SetMonthlyUsageTokens(1568040383)
	})
	svc := service.NewSubscriptionService(nil, s.repo, nil, nil, nil)
	s.T().Cleanup(svc.Stop)
	active, page, err := svc.List(s.ctx, 1, 20, &u.ID, nil, service.SubscriptionFilterActiveAvailable, "", "", "")
	s.Require().NoError(err)
	s.Require().Empty(active, "day 61 must not grant a third quota")
	s.Require().Zero(page.Total)
	all, _, err := svc.List(s.ctx, 1, 20, &u.ID, nil, service.SubscriptionStatusActive, "", "", "")
	s.Require().NoError(err)
	s.Require().Len(all, 1)
	s.Require().Equal(10000.01, all[0].MonthlyUsageUSD)
	s.Require().EqualValues(1568040383, all[0].MonthlyUsageTokens)
	progress, err := svc.GetSubscriptionProgress(s.ctx, sub.ID)
	s.Require().NoError(err)
	s.Require().True(progress.Monthly.ResetsAt.Equal(expiry))
	s.Require().Equal(10000.01, progress.Monthly.UsedUSD)
	needsMaintenance, err := svc.ValidateAndCheckLimits(&all[0], all[0].Group)
	s.Require().ErrorIs(err, service.ErrMonthlyLimitExceeded)
	s.Require().False(needsMaintenance)
	s.Require().NoError(svc.CheckAndResetWindows(s.ctx, &all[0]))
	persisted, err := s.repo.GetByID(s.ctx, sub.ID)
	s.Require().NoError(err)
	s.Require().Equal(10000.01, persisted.MonthlyUsageUSD)
	s.Require().EqualValues(1568040383, persisted.MonthlyUsageTokens)
	groups, err := s.repo.ListAvailableResetCardGroups(s.ctx, []int64{sub.ID})
	s.Require().NoError(err)
	s.Require().Empty(groups)
	_, historyPage, err := s.repo.ListResetHistory(s.ctx, sub.ID, pagination.DefaultPagination())
	s.Require().NoError(err)
	s.Require().Zero(historyPage.Total)
}
