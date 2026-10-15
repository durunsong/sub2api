package migrations

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRedeemEmailCampaignsMigration(t *testing.T) {
	content, err := FS.ReadFile("246_redeem_email_campaigns.sql")
	require.NoError(t, err)

	sql := strings.Join(strings.Fields(string(content)), " ")
	for _, want := range []string{
		"CREATE TABLE IF NOT EXISTS redeem_email_batches",
		"coupon_value NUMERIC(20, 8) NOT NULL CHECK (coupon_value > 0)",
		"segments <@ ARRAY['inactive_1_2m', 'inactive_2_3m', 'inactive_3m_plus', 'never_used']::TEXT[]",
		"reference_time TIMESTAMPTZ NOT NULL",
		"created_by BIGINT NOT NULL REFERENCES users(id)",
		"UNIQUE (created_by, idempotency_key_hash)",
		"CHECK (pending_count + sending_count + sent_count + failed_count + skipped_count = total_count)",
		"CHECK (status IN ('pending', 'sending', 'sent', 'failed', 'skipped'))",
		"CREATE TABLE IF NOT EXISTS redeem_email_items",
		"batch_id BIGINT NOT NULL REFERENCES redeem_email_batches(id) ON DELETE CASCADE",
		"user_id BIGINT REFERENCES users(id) ON DELETE SET NULL",
		"redeem_code_id BIGINT UNIQUE REFERENCES redeem_codes(id) ON DELETE SET NULL",
		"UNIQUE (batch_id, position)",
		"UNIQUE (batch_id, user_id)",
		"idx_redeem_email_items_claim",
		"idx_redeem_email_items_sending_lease",
	} {
		require.Contains(t, sql, want)
	}
	require.NotContains(t, strings.ToUpper(sql), "DROP TABLE")
}
