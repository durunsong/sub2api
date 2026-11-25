package repository

import (
	"context"
	"database/sql"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
)

// Lock before computing available funds, including against concurrent top-ups.
// Historical debt is preserved; new consumption never increases it.
// The ledger and wallet update share a statement and the billing transaction.
const balanceSettlementSQL = `
 WITH target AS (
  SELECT id, balance FROM users WHERE id = $1 AND deleted_at IS NULL FOR UPDATE
 ), items AS (
  SELECT cost, request_id, api_key_id, ord,
   COALESCE(SUM(cost) OVER (ORDER BY ord ROWS BETWEEN UNBOUNDED PRECEDING AND 1 PRECEDING), 0) AS spent
  FROM unnest($2::numeric[], $3::text[], $4::bigint[]) WITH ORDINALITY AS input(cost, request_id, api_key_id, ord)
 ), updated AS (
  UPDATE users u SET balance = target.balance - LEAST((SELECT SUM(cost) FROM items), GREATEST(target.balance, 0)), updated_at = NOW()
  FROM target WHERE u.id = target.id
  RETURNING target.balance AS original_balance
 ), amounts AS (
  SELECT items.*,
   original_balance - LEAST(spent, GREATEST(original_balance, 0)) AS before,
   original_balance - LEAST(spent + cost, GREATEST(original_balance, 0)) AS after
  FROM items CROSS JOIN updated
 ), recorded AS (
  INSERT INTO usage_balance_settlements
   (user_id, request_id, api_key_id, requested_amount, charged_amount, absorbed_amount, balance_before, balance_after)
  SELECT $1, request_id, api_key_id, cost, before - after, cost - (before - after), before, after FROM amounts
  RETURNING request_id, api_key_id, balance_before, balance_after, charged_amount, absorbed_amount
 )
 SELECT balance_before, balance_after, charged_amount, absorbed_amount
 FROM recorded JOIN items USING (request_id, api_key_id) ORDER BY ord
`

func settleUsageBillingBalances(ctx context.Context, tx *sql.Tx, cmds []*service.UsageBillingCommand) ([]*service.UsageBillingApplyResult, error) {
	costs := make([]string, len(cmds))
	requests := make([]string, len(cmds))
	keys := make([]int64, len(cmds))
	for i, cmd := range cmds {
		costs[i] = usageBillingAmount(cmd.BalanceCost).String()
		requests[i] = cmd.RequestID
		keys[i] = cmd.APIKeyID
	}
	rows, err := tx.QueryContext(ctx, balanceSettlementSQL, cmds[0].UserID, pq.Array(costs), pq.Array(requests), pq.Array(keys))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	results := make([]*service.UsageBillingApplyResult, 0, len(cmds))
	for rows.Next() {
		var before, after, charged, absorbed decimal.Decimal
		if err := rows.Scan(&before, &after, &charged, &absorbed); err != nil {
			return nil, err
		}
		oldBalance, newBalance := before.InexactFloat64(), after.InexactFloat64()
		results = append(results, &service.UsageBillingApplyResult{Applied: true, OldBalance: &oldBalance, NewBalance: &newBalance, BalanceCharged: charged.InexactFloat64(), BalanceAbsorbed: absorbed.InexactFloat64(), BalanceOverdrafted: absorbed.IsPositive()})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(results) != len(cmds) {
		return nil, service.ErrUserNotFound
	}
	return results, nil
}
