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

func TestAuditLogRepository_PersistsDailyResetSuccessExtraThroughServiceFlush(t *testing.T) {
	db := integrationDB
	repo := NewAuditLogRepository(db)
	auditService := service.NewAuditLogService(repo, nil)
	auditService.Start()

	requestID := fmt.Sprintf("daily-reset-it-%d", time.Now().UnixNano())
	actorID := int64(7)
	auditService.Record(&service.AuditLog{
		ActorUserID: &actorID,
		Action:      "user.subscription.daily_reset",
		Method:      "POST",
		Path:        "/api/v1/subscriptions/42/reset-daily",
		RequestID:   requestID,
		StatusCode:  500,
		Extra: map[string]any{
			"user_id": int64(7), "subscription_id": int64(42),
			"credits_before": 2, "credits_after": 1,
			"result": "success", "error_code": "RESPONSE_RELOAD_FAILED",
		},
	})
	auditService.Stop()
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM audit_logs WHERE request_id = $1", requestID)
	})

	var result, errorCode string
	var before, after int
	err := db.QueryRowContext(context.Background(), `
		SELECT extra->>'result', extra->>'error_code',
		       (extra->>'credits_before')::int, (extra->>'credits_after')::int
		FROM audit_logs WHERE request_id = $1
	`, requestID).Scan(&result, &errorCode, &before, &after)
	require.NoError(t, err)
	require.Equal(t, "success", result)
	require.Equal(t, "RESPONSE_RELOAD_FAILED", errorCode)
	require.Equal(t, 2, before)
	require.Equal(t, 1, after)
}
