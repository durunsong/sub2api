package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type ipBanRepository struct {
	db *sql.DB
}

func NewIPBanRepository(db *sql.DB) service.IPBanRepository {
	return &ipBanRepository{db: db}
}

func (r *ipBanRepository) Create(ctx context.Context, ban *service.IPBan) error {
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO ip_bans (pattern, status, reason, source, created_by, expires_at, hit_count)
		VALUES ($1, $2, NULLIF($3, ''), $4, $5, $6, $7)
		RETURNING id, pattern, status, reason, source, created_by, expires_at, last_hit_at, hit_count, created_at, updated_at
	`, ban.Pattern, ban.Status, ban.Reason, ban.Source, nullableInt64(ban.CreatedBy), nullableTime(ban.ExpiresAt), ban.HitCount)
	created, err := scanIPBan(row)
	if err != nil {
		return translatePersistenceError(err, nil, service.ErrIPBanAlreadyExists)
	}
	*ban = *created
	return nil
}

func (r *ipBanRepository) GetByID(ctx context.Context, id int64) (*service.IPBan, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, pattern, status, reason, source, created_by, expires_at, last_hit_at, hit_count, created_at, updated_at
		FROM ip_bans
		WHERE id = $1 AND deleted_at IS NULL
	`, id)
	ban, err := scanIPBan(row)
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrIPBanNotFound, nil)
	}
	return ban, nil
}

func (r *ipBanRepository) List(ctx context.Context, params pagination.PaginationParams, filters service.IPBanListFilters) ([]service.IPBan, *pagination.PaginationResult, error) {
	where := "WHERE deleted_at IS NULL"
	args := make([]any, 0, 4)
	if filters.Search != "" {
		args = append(args, "%"+filters.Search+"%")
		where += fmt.Sprintf(" AND (pattern ILIKE $%d OR reason ILIKE $%d)", len(args), len(args))
	}
	if filters.Status != "" {
		args = append(args, filters.Status)
		where += fmt.Sprintf(" AND status = $%d", len(args))
	}
	var total int64
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM ip_bans "+where, args...).Scan(&total); err != nil {
		return nil, nil, err
	}

	orderBy := ipBanOrderBy(params)
	args = append(args, params.Limit(), params.Offset())
	query := fmt.Sprintf(`
		SELECT id, pattern, status, reason, source, created_by, expires_at, last_hit_at, hit_count, created_at, updated_at
		FROM ip_bans
		%s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, where, orderBy, len(args)-1, len(args))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	out, err := scanIPBanRows(rows)
	if err != nil {
		return nil, nil, err
	}
	return out, paginationResultFromTotal(total, params), nil
}

func (r *ipBanRepository) Update(ctx context.Context, ban *service.IPBan) error {
	row := r.db.QueryRowContext(ctx, `
		UPDATE ip_bans
		SET pattern = $2,
			status = $3,
			reason = NULLIF($4, ''),
			source = $5,
			created_by = $6,
			expires_at = $7,
			updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, pattern, status, reason, source, created_by, expires_at, last_hit_at, hit_count, created_at, updated_at
	`, ban.ID, ban.Pattern, ban.Status, ban.Reason, ban.Source, nullableInt64(ban.CreatedBy), nullableTime(ban.ExpiresAt))
	updated, err := scanIPBan(row)
	if err != nil {
		return translatePersistenceError(err, service.ErrIPBanNotFound, service.ErrIPBanAlreadyExists)
	}
	*ban = *updated
	return nil
}

func (r *ipBanRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, "UPDATE ip_bans SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL", id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return service.ErrIPBanNotFound
	}
	return err
}

func (r *ipBanRepository) ListActive(ctx context.Context, now time.Time) ([]service.IPBan, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, pattern, status, reason, source, created_by, expires_at, last_hit_at, hit_count, created_at, updated_at
		FROM ip_bans
		WHERE deleted_at IS NULL
			AND status = $1
			AND (expires_at IS NULL OR expires_at > $2)
		ORDER BY id DESC
	`, service.IPBanStatusActive, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanIPBanRows(rows)
}

func (r *ipBanRepository) RecordHit(ctx context.Context, id int64, at time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE ip_bans
		SET hit_count = hit_count + 1,
			last_hit_at = $2,
			updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`, id, at)
	return err
}

type ipBanScanner interface {
	Scan(dest ...any) error
}

func scanIPBan(scanner ipBanScanner) (*service.IPBan, error) {
	var ban service.IPBan
	var reason sql.NullString
	var createdBy sql.NullInt64
	var expiresAt sql.NullTime
	var lastHitAt sql.NullTime
	if err := scanner.Scan(
		&ban.ID,
		&ban.Pattern,
		&ban.Status,
		&reason,
		&ban.Source,
		&createdBy,
		&expiresAt,
		&lastHitAt,
		&ban.HitCount,
		&ban.CreatedAt,
		&ban.UpdatedAt,
	); err != nil {
		return nil, err
	}
	if reason.Valid {
		ban.Reason = reason.String
	}
	if createdBy.Valid {
		ban.CreatedBy = &createdBy.Int64
	}
	if expiresAt.Valid {
		ban.ExpiresAt = &expiresAt.Time
	}
	if lastHitAt.Valid {
		ban.LastHitAt = &lastHitAt.Time
	}
	return &ban, nil
}

func scanIPBanRows(rows *sql.Rows) ([]service.IPBan, error) {
	out := make([]service.IPBan, 0)
	for rows.Next() {
		ban, err := scanIPBan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *ban)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func nullableInt64(value *int64) any {
	if value == nil {
		return nil
	}
	return *value
}

func nullableTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return *value
}

func ipBanOrderBy(params pagination.PaginationParams) string {
	column := "created_at"
	switch params.SortBy {
	case "pattern", "status", "source", "hit_count", "last_hit_at", "expires_at", "created_at", "updated_at", "id":
		column = params.SortBy
	}
	order := params.NormalizedSortOrder(pagination.SortOrderDesc)
	return column + " " + order + ", id DESC"
}
