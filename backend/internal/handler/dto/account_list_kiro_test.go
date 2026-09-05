package dto

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestAccountListItemPreservesKiroRuntimeAndQuota(t *testing.T) {
	resetAt := time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)
	account := &service.Account{
		ID: 42, Platform: service.PlatformKiro,
		KiroQuotaState: "exhausted", KiroQuotaReason: "credits exhausted", KiroQuotaResetAt: &resetAt,
		KiroRuntimeState: "cooldown", KiroRuntimeReason: "upstream 429", KiroRuntimeResetAt: &resetAt,
		Credentials: map[string]any{"access_token": "test-secret"},
	}
	full := AccountFromService(account)
	lite := AccountListItemFromAccount(full)
	fullJSON, err := json.Marshal(full)
	require.NoError(t, err)
	liteJSON, err := json.Marshal(lite)
	require.NoError(t, err)
	var fullFields, liteFields map[string]any
	require.NoError(t, json.Unmarshal(fullJSON, &fullFields))
	require.NoError(t, json.Unmarshal(liteJSON, &liteFields))
	for _, key := range []string{
		"kiro_quota_state", "kiro_quota_reason", "kiro_quota_reset_at",
		"kiro_runtime_state", "kiro_runtime_reason", "kiro_runtime_reset_at",
	} {
		require.Contains(t, liteFields, key)
		require.Equal(t, fullFields[key], liteFields[key], key)
	}
	require.NotContains(t, string(liteJSON), "test-secret")
	require.NotContains(t, liteFields, "groups")
}
