package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestIPBanUpsertAutomaticPreservesManualOverrides(t *testing.T) {
	for _, writeErr := range []error{nil, errors.New("database write failed")} {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		t.Cleanup(func() { _ = db.Close() })
		expires := time.Now().Add(24 * time.Hour)
		ban := &service.IPBan{RuleType: "ip", Pattern: "43.255.119.7", Status: "active", Reason: "automatic", Source: service.IPBanSourceAdminLogin, ExpiresAt: &expires}
		// Conflict handling must preserve manual/disabled/permanent rules and not extend an active ban on every attempt.
		expected := mock.ExpectExec(`(?s)INSERT INTO ip_bans .*ON CONFLICT .*WHERE deleted_at IS NULL.*DO UPDATE.*WHERE ip_bans.source = EXCLUDED.source AND ip_bans.status = 'active'.*ip_bans.expires_at IS NOT NULL AND ip_bans.expires_at <= NOW\(\)`).
			WithArgs(ban.RuleType, ban.Pattern, ban.Status, ban.Reason, ban.Source, expires)
		if writeErr == nil {
			expected.WillReturnResult(sqlmock.NewResult(0, 1))
		} else {
			expected.WillReturnError(writeErr)
		}
		err = NewIPBanRepository(db).UpsertAutomatic(context.Background(), ban)
		require.ErrorIs(t, err, writeErr)
		require.NoError(t, mock.ExpectationsWereMet())
	}
}
