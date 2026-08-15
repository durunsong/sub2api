package migrations

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResetCardMigrationConstrainsConsumptionChronology(t *testing.T) {
	data, err := os.ReadFile("224_create_user_subscription_reset_cards.sql")
	require.NoError(t, err)
	require.Contains(t, strings.Join(strings.Fields(string(data)), " "), "consumed_at IS NULL OR consumed_at >= created_at")
}

func TestResetCardMigrationSupportsConsumeIdempotencyAndVoiding(t *testing.T) {
	data, err := os.ReadFile("224_create_user_subscription_reset_cards.sql")
	require.NoError(t, err)
	normalized := strings.Join(strings.Fields(string(data)), " ")
	require.Contains(t, normalized, "consume_idempotency_key VARCHAR(128)")
	require.Contains(t, normalized, "voided_at TIMESTAMPTZ")
	require.Contains(t, normalized, "CREATE UNIQUE INDEX IF NOT EXISTS idx_user_subscription_reset_cards_consume_idempotency ON user_subscription_reset_cards (user_subscription_id, consume_idempotency_key) WHERE consume_idempotency_key IS NOT NULL")
	require.Contains(t, normalized, "WHERE consumed_at IS NULL AND voided_at IS NULL")
}

func TestResetCardBackfillIsReplaySafeAndPreservesLegacyCounter(t *testing.T) {
	data, err := os.ReadFile("225_backfill_user_subscription_reset_cards.sql")
	require.NoError(t, err)
	normalized := strings.Join(strings.Fields(string(data)), " ")
	require.Contains(t, normalized, "ON CONFLICT")
	require.NotContains(t, strings.ToUpper(normalized), "UPDATE USER_SUBSCRIPTIONS")
}
