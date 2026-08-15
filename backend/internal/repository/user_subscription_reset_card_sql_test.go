package repository

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResetCardSubscriptionLockIsSeparateFromConsumptionSnapshot(t *testing.T) {
	lockSQL := strings.Join(strings.Fields(lockResetCardSubscriptionSQL), " ")
	consumeSQL := strings.Join(strings.Fields(consumeResetCardSQL), " ")
	require.Contains(t, lockSQL, "FROM user_subscriptions")
	require.Contains(t, lockSQL, "FOR UPDATE")
	require.NotContains(t, lockSQL, "SKIP LOCKED")
	require.Contains(t, lockSQL, "user_id = $2")
	require.Contains(t, lockSQL, "deleted_at IS NULL")
	require.Contains(t, lockSQL, "status IN ('active', 'expired')")
	require.NotContains(t, consumeSQL, "locked_subscription")
	require.NotContains(t, consumeSQL, "FROM user_subscriptions")
	require.Contains(t, consumeSQL, "c.user_subscription_id = $1")
	require.Contains(t, consumeSQL, "c.validity_days = $2")
}

func TestConsumeResetCardSQLUsesDatabaseObservedConsumptionTime(t *testing.T) {
	normalized := strings.Join(strings.Fields(consumeResetCardSQL), " ")
	require.Contains(t, normalized, "SELECT CURRENT_TIMESTAMP AS now")
	require.Contains(t, normalized, "GREATEST(clock.now, c.created_at)")
	require.Contains(t, normalized, "LEAST(GREATEST(clock.now, c.created_at) + make_interval(days => $2), $3)")
	require.NotContains(t, normalized, "GREATEST($3, c.created_at)")
	require.Contains(t, normalized, "RETURNING c.id, sc.validity_days, c.consumed_at, sc.expires_at")
}

func TestConsumeResetCardSQLReplaysIdempotencyKeyAndFiltersVoidedCards(t *testing.T) {
	normalized := strings.Join(strings.Fields(consumeResetCardSQL), " ")
	require.Contains(t, normalized, "consume_idempotency_key = $4")
	require.Contains(t, normalized, "c.voided_at IS NULL")
	require.Contains(t, normalized, "already_consumed")
}

func TestConsumeResetCardSQLClearsAllUsageAndMaintainsLegacyMirror(t *testing.T) {
	normalized := strings.Join(strings.Fields(consumeResetCardSQL), " ")
	for _, fragment := range []string{
		"daily_usage_usd = 0", "weekly_usage_usd = 0", "monthly_usage_usd = 0",
		"daily_usage_tokens = 0", "weekly_usage_tokens = 0", "monthly_usage_tokens = 0",
		"manual_reset_credits = GREATEST(manual_reset_credits - 1, 0)",
	} {
		require.Contains(t, normalized, fragment)
	}
}

func TestListAvailableResetCardGroupsSQLIsBatchAggregate(t *testing.T) {
	normalized := strings.Join(strings.Fields(listAvailableResetCardGroupsSQL), " ")
	require.Contains(t, normalized, "user_subscription_id = ANY($1)")
	require.Contains(t, normalized, "consumed_at IS NULL")
	require.Contains(t, normalized, "voided_at IS NULL")
	require.Contains(t, normalized, "GROUP BY user_subscription_id, validity_days")
}
