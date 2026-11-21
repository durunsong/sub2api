//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/Wei-Shaw/sub2api/migrations"
	"github.com/stretchr/testify/require"
)

func TestRedeemEmailActiveAudiencesIntegration(t *testing.T) {
	ctx := context.Background()
	tx := testTx(t)
	// Transaction-local tables exercise the real predicates without touching other fixtures.
	_, err := tx.Exec(`CREATE TEMP TABLE users (
		id BIGINT PRIMARY KEY, email TEXT, username TEXT, status TEXT DEFAULT 'active',
		role TEXT DEFAULT 'user', deleted_at TIMESTAMPTZ, balance NUMERIC DEFAULT 0
	) ON COMMIT DROP;
	CREATE TEMP TABLE usage_logs (user_id BIGINT, created_at TIMESTAMPTZ) ON COMMIT DROP;
	INSERT INTO users (id, email, username)
	SELECT n, 'audience-' || n || '@example.com', 'audience-' || n FROM generate_series(1, 10) n;
	UPDATE users SET status = 'disabled' WHERE id = 7;
	UPDATE users SET role = 'admin' WHERE id = 8;
	UPDATE users SET deleted_at = '2026-03-01Z' WHERE id = 9;
	UPDATE users SET email = 'audience-10@example.invalid' WHERE id = 10;
	INSERT INTO usage_logs (user_id, created_at) VALUES
	(1, '2026-03-01 12:00:00Z'), (1, '2025-01-01 00:00:00Z'),
	(2, '2026-02-22 12:00:00Z'), (3, '2026-02-22 11:59:59.999999Z'),
	(4, '2026-01-30 12:00:00Z'), (5, '2026-01-30 11:59:59.999999Z'),
	(7, '2026-03-01 12:00:00Z'), (8, '2026-03-01 12:00:00Z'),
	(9, '2026-03-01 12:00:00Z'), (10, '2026-03-01 12:00:00Z')`)
	require.NoError(t, err)
	now := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	cutoffs := service.NewRedeemEmailCutoffs(now)
	for _, tc := range []struct {
		name     string
		segments []string
		want     []int64
	}{
		{"seven days inclusive", []string{"active_7d"}, []int64{1, 2}},
		{"thirty days inclusive", []string{"active_30d"}, []int64{1, 2, 3, 4}},
		{"overlapping active windows", []string{"active_7d", "active_30d"}, []int64{1, 2, 3, 4}},
		{"calendar month overlap", []string{"inactive_1_2m", "active_30d"}, []int64{1, 2, 3, 4, 5}},
		{"include never used explicitly", []string{"never_used", "active_7d"}, []int64{1, 2, 6}},
		{"old defaults unchanged", service.RedeemEmailDefaultSegments(), []int64{4, 5}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rows, err := tx.QueryContext(ctx, redeemEmailSearchSQL, append(redeemEmailCutoffArgs(cutoffs, tc.segments), "", 1001)...)
			require.NoError(t, err)
			got, err := scanRedeemEmailRecipients(rows)
			require.NoError(t, err)
			ids := make([]int64, 0, len(got))
			for _, recipient := range got {
				ids = append(ids, recipient.UserID)
				require.Contains(t, tc.segments, recipient.Segment)
			}
			require.ElementsMatch(t, tc.want, ids, "overlapping segments must not duplicate recipients")
		})
	}
	for _, segment := range []string{"active_7d", "active_30d"} {
		p := service.CreateRedeemEmailBatchParams{Segments: []string{segment}, Cutoffs: cutoffs, UserIDs: []int64{1, 2}}
		got, err := lockEligibleRedeemEmailRecipients(ctx, tx, p)
		require.NoError(t, err)
		require.Len(t, got, 2)
		// When time advances, boundary users no longer qualify for that window.
		p.Cutoffs = service.NewRedeemEmailCutoffs(now.Add(31 * 24 * time.Hour))
		_, err = lockEligibleRedeemEmailRecipients(ctx, tx, p)
		require.ErrorIs(t, err, service.ErrRedeemEmailRecipientIneligible)
	}
}

func TestRedeemEmailActiveSegmentsMigrationIntegration(t *testing.T) {
	ctx := context.Background()
	tx := testTx(t)
	_, err := tx.Exec(`CREATE TEMP TABLE redeem_email_batches (
		segments TEXT[] NOT NULL CONSTRAINT redeem_email_batches_segments_check CHECK (
			cardinality(segments) > 0
			AND segments <@ ARRAY['inactive_1_2m', 'inactive_2_3m', 'inactive_3m_plus', 'never_used', 'zero_balance']::TEXT[]
			AND (NOT ('zero_balance' = ANY (segments)) OR cardinality(segments) = 1)
		)
	) ON COMMIT DROP;
	INSERT INTO redeem_email_batches VALUES ('{inactive_1_2m}'), ('{zero_balance}')`)
	require.NoError(t, err)
	content, err := migrations.FS.ReadFile("253_redeem_email_active_segments.sql")
	require.NoError(t, err)
	_, err = tx.ExecContext(ctx, string(content))
	require.NoError(t, err)
	for i, segments := range []string{"{active_7d}", "{active_30d}", "{active_7d,active_30d}", "{active_30d,inactive_1_2m}", "{never_used}", "{zero_balance}", "{}", "{bogus}", "{zero_balance,active_7d}", "{zero_balance,active_30d}"} {
		t.Run(segments, func(t *testing.T) {
			_, err = tx.Exec("SAVEPOINT validate_segments")
			require.NoError(t, err)
			_, err = tx.ExecContext(ctx, "INSERT INTO redeem_email_batches (segments) VALUES ($1::text[])", segments)
			if i < 6 {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, "redeem_email_batches_segments_check")
			}
			_, err = tx.Exec("ROLLBACK TO SAVEPOINT validate_segments")
			require.NoError(t, err)
		})
	}
	var count int
	require.NoError(t, tx.QueryRow("SELECT COUNT(*) FROM redeem_email_batches").Scan(&count))
	require.Equal(t, 2, count, "existing batches must survive the migration")
}
