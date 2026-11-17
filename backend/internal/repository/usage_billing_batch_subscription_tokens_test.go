//go:build unit

package repository

import (
	"context"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

func TestUsageBillingBatchSubscriptionsPreserveTokensAndTermQuota(t *testing.T) {
	for _, cost := range []float64{0, 0.25} {
		t.Run(map[bool]string{true: "free", false: "paid"}[cost == 0], func(t *testing.T) {
			var statement string
			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherFunc(func(expected, actual string) error {
				statement = actual
				return sqlmock.QueryMatcherRegexp.Match(expected, actual)
			})))
			require.NoError(t, err)
			defer db.Close()
			mock.ExpectBegin()
			tx, err := db.BeginTx(context.Background(), nil)
			require.NoError(t, err)
			defer tx.Rollback()
			// Only newly claimed requests count, including free requests. Updates must
			// accumulate monthly fields without restoring the upstream 30-day reset.
			mock.ExpectExec("UPDATE user_subscriptions").WithArgs(cost*2, int64(210), int64(7)).WillReturnResult(sqlmock.NewResult(0, 1))
			mock.ExpectExec("UPDATE user_subscriptions").WithArgs(0.0, int64(20), int64(9)).WillReturnResult(sqlmock.NewResult(0, 1))
			id, other := int64(7), int64(9)
			cmds := []*service.UsageBillingCommand{
				{SubscriptionID: &other, InputTokens: 20},
				{SubscriptionID: &id, SubscriptionCost: cost, InputTokens: 100, OutputTokens: 10, CacheCreationTokens: 20, CacheReadTokens: 30},
				{SubscriptionID: &id, SubscriptionCost: cost, InputTokens: 50},
				{SubscriptionID: &id, SubscriptionCost: 999, InputTokens: 999},
				{InputTokens: 999},
			}
			err = applyUsageBillingBatchSubscriptions(context.Background(), tx, cmds, []bool{true, true, true, false, true})
			require.NoError(t, err)
			require.Contains(t, statement, "monthly_usage_usd = us.monthly_usage_usd + $1")
			require.Contains(t, statement, "monthly_usage_tokens = us.monthly_usage_tokens + $2")
			require.False(t, strings.Contains(statement, "monthly_window_start"))
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
