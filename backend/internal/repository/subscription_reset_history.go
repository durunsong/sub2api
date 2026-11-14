package repository

import (
	"context"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

const subscriptionResetHistorySQL = `
	SELECT id, 'admin' AS type, reset_at, daily, weekly, monthly, 0 AS validity_days
	FROM user_subscription_quota_resets WHERE user_subscription_id = $1
	UNION ALL
	SELECT id, 'reset_card' AS type, consumed_at AS reset_at, TRUE, TRUE, TRUE, validity_days
	FROM user_subscription_reset_cards
	WHERE user_subscription_id = $1 AND consumed_at IS NOT NULL`

func (r *userSubscriptionRepository) ListResetHistory(ctx context.Context, subscriptionID int64, params pagination.PaginationParams) ([]service.SubscriptionResetHistory, *pagination.PaginationResult, error) {
	client := clientFromContext(ctx, r.client)
	rows, err := client.QueryContext(ctx, `SELECT COUNT(*) FROM (`+subscriptionResetHistorySQL+`) AS history`, subscriptionID)
	if err != nil {
		return nil, nil, err
	}
	var total int64
	if rows.Next() {
		err = rows.Scan(&total)
	}
	rowsErr := rows.Err()
	_ = rows.Close()
	if err != nil {
		return nil, nil, err
	}
	if rowsErr != nil {
		return nil, nil, rowsErr
	}
	rows, err = client.QueryContext(ctx, `SELECT * FROM (`+subscriptionResetHistorySQL+`) AS history
		ORDER BY reset_at DESC, type DESC, id DESC LIMIT $2 OFFSET $3`, subscriptionID, params.Limit(), params.Offset())
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = rows.Close() }()
	items := make([]service.SubscriptionResetHistory, 0)
	for rows.Next() {
		var item service.SubscriptionResetHistory
		if err := rows.Scan(&item.ID, &item.Type, &item.ResetAt, &item.Daily, &item.Weekly, &item.Monthly, &item.ValidityDays); err != nil {
			return nil, nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	return items, &pagination.PaginationResult{
		Total: total, Page: params.Page, PageSize: params.Limit(), Pages: int((total + int64(params.Limit()) - 1) / int64(params.Limit())),
	}, nil
}
