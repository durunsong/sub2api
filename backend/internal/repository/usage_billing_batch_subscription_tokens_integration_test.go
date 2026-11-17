//go:build integration

package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestUsageBillingBatchForkTermQuotaAndTokens(t *testing.T) {
	ctx := context.Background()
	client := testEntClient(t)
	repo := NewUsageBillingRepository(client, integrationDB)
	for _, cost := range []float64{0, 0.25} {
		t.Run(fmt.Sprintf("cost_%g", cost), func(t *testing.T) {
			f := newBatchBillingFixture(t, client, 10)
			started := time.Now().UTC().Add(-40 * 24 * time.Hour).Truncate(time.Second)
			expires := started.Add(90 * 24 * time.Hour)
			_, err := integrationDB.ExecContext(ctx, `UPDATE user_subscriptions SET starts_at=$1, monthly_window_start=$1, expires_at=$2, monthly_usage_usd=5, monthly_usage_tokens=1000 WHERE id=$3`, started, expires, f.subscriptionID)
			require.NoError(t, err)
			cmds := []*service.UsageBillingCommand{
				{RequestID: fmt.Sprintf("fork-term-%d-a", f.userID), APIKeyID: f.keyA, UserID: f.userID, SubscriptionID: &f.subscriptionID, SubscriptionCost: cost, InputTokens: 100, OutputTokens: 10, CacheCreationTokens: 20, CacheReadTokens: 30},
				{RequestID: fmt.Sprintf("fork-term-%d-b", f.userID), APIKeyID: f.keyB, UserID: f.userID, SubscriptionID: &f.subscriptionID, InputTokens: 50},
			}
			for attempt := 0; attempt < 2; attempt++ {
				outcomes, err := repo.ApplyBatch(ctx, cmds)
				require.NoError(t, err)
				require.Len(t, outcomes, 2)
				for _, outcome := range outcomes {
					require.NoError(t, outcome.Err)
					require.Equal(t, attempt == 0, outcome.Result.Applied)
				}
				var monthlyCost float64
				var dailyTokens, weeklyTokens, monthlyTokens int64
				var gotStart, gotExpires time.Time
				err = integrationDB.QueryRowContext(ctx, `SELECT monthly_usage_usd, daily_usage_tokens, weekly_usage_tokens, monthly_usage_tokens, monthly_window_start, expires_at FROM user_subscriptions WHERE id=$1`, f.subscriptionID).Scan(&monthlyCost, &dailyTokens, &weeklyTokens, &monthlyTokens, &gotStart, &gotExpires)
				require.NoError(t, err)
				require.Equal(t, 5+cost, monthlyCost)
				require.Equal(t, int64(210), dailyTokens)
				require.Equal(t, int64(210), weeklyTokens)
				require.Equal(t, int64(1210), monthlyTokens)
				require.True(t, started.Equal(gotStart))
				require.True(t, expires.Equal(gotExpires))
			}
		})
	}
}
