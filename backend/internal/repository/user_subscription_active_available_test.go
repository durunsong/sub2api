package repository

import (
	"context"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	_ "github.com/Wei-Shaw/sub2api/ent/runtime"
	"github.com/Wei-Shaw/sub2api/ent/usersubscription"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
)

func TestListActiveAvailableBuildsQuotaAndWindowPredicates(t *testing.T) {
	var capturedSQL string
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(captureEntQueryMatcher{actual: &capturedSQL}))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	driver := entsql.OpenDB(dialect.Postgres, db)
	client := dbent.NewClient(dbent.Driver(driver))
	t.Cleanup(func() { _ = client.Close() })
	repo := NewUserSubscriptionRepository(client)

	mock.ExpectQuery("count active subscriptions with quota remaining").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery("list active subscriptions with quota remaining").
		WillReturnRows(sqlmock.NewRows(usersubscription.Columns))

	subs, page, err := repo.List(
		context.Background(),
		pagination.PaginationParams{Page: 1, PageSize: 20},
		nil,
		nil,
		service.SubscriptionFilterActiveAvailable,
		"",
		"created_at",
		"desc",
	)
	require.NoError(t, err)
	require.Empty(t, subs)
	require.Zero(t, page.Total)
	require.NoError(t, mock.ExpectationsWereMet())

	normalized := normalizeSQLWhitespace(capturedSQL)
	for _, fragment := range []string{
		`"status" = $1`,
		`"expires_at" > $2`,
		"EXISTS",
		`"daily_limit_usd"`,
		`"weekly_limit_usd"`,
		`"monthly_limit_usd"`,
		`"daily_usage_usd"`,
		`"weekly_usage_usd"`,
		`"monthly_usage_usd"`,
		"AT TIME ZONE",
		"make_interval",
	} {
		require.True(t, strings.Contains(normalized, fragment), "missing %q in query: %s", fragment, normalized)
	}
}
