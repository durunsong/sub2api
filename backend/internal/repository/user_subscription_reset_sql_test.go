package repository

import (
	"strings"
	"testing"
)

func TestResetDailyQuotaSQL_AllowsActiveOneTimeWithoutRestart(t *testing.T) {
	if strings.Contains(resetDailyQuotaSQL, "$9 = FALSE AND status = 'active' AND expires_at > $6 AND expires_at > starts_at") {
		t.Fatal("production SQL still rejects active one-time subscriptions")
	}
	if !strings.Contains(resetDailyQuotaSQL, "$9 = FALSE AND status = 'active' AND expires_at > $6") {
		t.Fatal("production SQL must allow every active unexpired subscription")
	}
}
