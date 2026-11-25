//go:build integration

package repository

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestUsageBalanceFloor(t *testing.T) {
	for _, batch := range []bool{false, true} {
		for _, initial := range []float64{1, 0, -2, 3, 0.00000001} {
			t.Run(fmt.Sprintf("batch=%v/balance=%g", batch, initial), func(t *testing.T) {
				ctx := context.Background()
				client := testEntClient(t)
				repo := NewUsageBillingRepository(client, integrationDB)
				user := mustCreateUser(t, client, &service.User{Email: uuid.NewString() + "@test.com", PasswordHash: "hash", Balance: initial})
				key := mustCreateApiKey(t, client, &service.APIKey{UserID: user.ID, Key: "sk-" + uuid.NewString(), Name: "floor", Quota: 10})
				cmd := &service.UsageBillingCommand{RequestID: uuid.NewString(), APIKeyID: key.ID, UserID: user.ID, BalanceCost: 3, APIKeyQuotaCost: 3}
				var result *service.UsageBillingApplyResult
				if batch {
					outcomes, err := repo.ApplyBatch(ctx, []*service.UsageBillingCommand{cmd})
					require.NoError(t, err)
					require.NoError(t, outcomes[0].Err)
					result = outcomes[0].Result
				} else {
					var err error
					result, err = repo.Apply(ctx, cmd)
					require.NoError(t, err)
				}
				wantBalance := 0.0
				if initial < 0 {
					wantBalance = initial
				}
				require.Equal(t, wantBalance, *result.NewBalance)
				var balance, charged, absorbed, requested, before, after, quota float64
				require.NoError(t, integrationDB.QueryRowContext(ctx, "SELECT balance FROM users WHERE id=$1", user.ID).Scan(&balance))
				require.Equal(t, wantBalance, balance)
				require.NoError(t, integrationDB.QueryRowContext(ctx, "SELECT requested_amount, charged_amount, absorbed_amount, balance_before, balance_after FROM usage_balance_settlements WHERE request_id=$1 AND api_key_id=$2", cmd.RequestID, key.ID).Scan(&requested, &charged, &absorbed, &before, &after))
				wantCharged := initial
				if wantCharged < 0 {
					wantCharged = 0
				}
				require.Equal(t, 3.0, requested)
				require.Equal(t, wantCharged, charged)
				require.InDelta(t, 3-wantCharged, absorbed, 1e-10)
				require.Equal(t, initial, before)
				require.Equal(t, wantBalance, after)
				require.NoError(t, integrationDB.QueryRowContext(ctx, "SELECT quota_used FROM api_keys WHERE id=$1", key.ID).Scan(&quota))
				require.Equal(t, 3.0, quota)
				duplicate, err := repo.Apply(ctx, cmd)
				require.NoError(t, err)
				require.False(t, duplicate.Applied)
				var count int
				require.NoError(t, integrationDB.QueryRowContext(ctx, "SELECT COUNT(*) FROM usage_balance_settlements WHERE user_id=$1", user.ID).Scan(&count))
				require.Equal(t, 1, count)
			})
		}
	}
}

func TestUsageBalanceFloor_ConcurrentChargesAndRecharge(t *testing.T) {
	ctx := context.Background()
	client := testEntClient(t)
	repo := NewUsageBillingRepository(client, integrationDB)
	user := mustCreateUser(t, client, &service.User{Email: uuid.NewString() + "@test.com", PasswordHash: "hash", Balance: 1})
	key := mustCreateApiKey(t, client, &service.APIKey{UserID: user.ID, Key: "sk-" + uuid.NewString(), Name: "concurrent"})
	var wg sync.WaitGroup
	start := make(chan struct{})
	errs := make(chan error, 21)
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, err := repo.Apply(ctx, &service.UsageBillingCommand{RequestID: uuid.NewString(), APIKeyID: key.ID, UserID: user.ID, BalanceCost: 0.25})
			errs <- err
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-start
		_, err := integrationDB.ExecContext(ctx, "UPDATE users SET balance=balance+2 WHERE id=$1", user.ID)
		errs <- err
	}()
	close(start)
	wg.Wait()
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}
	var balance, charged, absorbed float64
	require.NoError(t, integrationDB.QueryRowContext(ctx, "SELECT balance FROM users WHERE id=$1", user.ID).Scan(&balance))
	require.GreaterOrEqual(t, balance, 0.0)
	require.NoError(t, integrationDB.QueryRowContext(ctx, "SELECT SUM(charged_amount),SUM(absorbed_amount) FROM usage_balance_settlements WHERE user_id=$1", user.ID).Scan(&charged, &absorbed))
	require.Equal(t, 3.0, balance+charged)
	require.Equal(t, 5.0, charged+absorbed)
}

func TestUsageBalanceFloor_BatchImageRounding(t *testing.T) {
	ctx := context.Background()
	client := testEntClient(t)
	user := mustCreateUser(t, client, &service.User{Email: uuid.NewString() + "@test.com", PasswordHash: "hash", Balance: 0.5})
	tx := testTx(t)
	_, err := reserveUsageBillingBatchImageBalance(ctx, tx, &service.BatchImageBalanceHoldCommand{UserID: user.ID, HoldAmount: 0.5})
	require.NoError(t, err)
	_, err = captureUsageBillingBatchImageBalance(ctx, tx, &service.BatchImageBalanceHoldCommand{UserID: user.ID, HoldAmount: 0.5, ActualAmount: 0.500000009})
	require.ErrorIs(t, err, service.ErrBatchImageSettlementCostExceedsHold)
	var balance, frozen float64
	require.NoError(t, tx.QueryRowContext(ctx, "SELECT balance,frozen_balance FROM users WHERE id=$1", user.ID).Scan(&balance, &frozen))
	require.Zero(t, balance)
	require.Equal(t, 0.5, frozen)
}

func TestUsageBalanceFloor_Legacy(t *testing.T) {
	ctx := context.Background()
	client := testEntClient(t)
	repo := NewUserRepository(client, integrationDB)
	user := mustCreateUser(t, client, &service.User{Email: uuid.NewString() + "@test.com", PasswordHash: "hash", Balance: 1})
	require.NoError(t, repo.DeductBalance(ctx, user.ID, 3))
	got, err := repo.GetByID(ctx, user.ID)
	require.NoError(t, err)
	require.Zero(t, got.Balance)
	var charged, absorbed float64
	require.NoError(t, integrationDB.QueryRowContext(ctx, "SELECT charged_amount,absorbed_amount FROM usage_balance_settlements WHERE user_id=$1", user.ID).Scan(&charged, &absorbed))
	require.Equal(t, 1.0, charged)
	require.Equal(t, 2.0, absorbed)
}

func TestUsageBalanceFloor_RollsBackLedgerAndWallet(t *testing.T) {
	for _, batch := range []bool{false, true} {
		t.Run(fmt.Sprintf("batch=%v", batch), func(t *testing.T) {
			ctx := context.Background()
			client := testEntClient(t)
			repo := NewUsageBillingRepository(client, integrationDB)
			user := mustCreateUser(t, client, &service.User{Email: uuid.NewString() + "@test.com", PasswordHash: "hash", Balance: 1})
			key := mustCreateApiKey(t, client, &service.APIKey{UserID: user.ID, Key: "sk-" + uuid.NewString(), Name: "rollback", Quota: 10})
			cmd := &service.UsageBillingCommand{RequestID: uuid.NewString(), APIKeyID: key.ID, UserID: user.ID, BalanceCost: 3, APIKeyQuotaCost: 1e30}
			var err error
			if batch {
				_, err = repo.ApplyBatch(ctx, []*service.UsageBillingCommand{cmd})
			} else {
				_, err = repo.Apply(ctx, cmd)
			}
			require.ErrorContains(t, err, "numeric field overflow")
			var balance float64
			var count int
			require.NoError(t, integrationDB.QueryRowContext(ctx, "SELECT balance FROM users WHERE id=$1", user.ID).Scan(&balance))
			require.Equal(t, 1.0, balance)
			require.NoError(t, integrationDB.QueryRowContext(ctx, "SELECT COUNT(*) FROM usage_balance_settlements WHERE user_id=$1", user.ID).Scan(&count))
			require.Zero(t, count)
			require.NoError(t, integrationDB.QueryRowContext(ctx, "SELECT COUNT(*) FROM usage_billing_dedup WHERE request_id=$1 AND api_key_id=$2", cmd.RequestID, key.ID).Scan(&count))
			require.Zero(t, count)
			cmd.APIKeyQuotaCost, cmd.RequestFingerprint = 3, ""
			result, err := repo.Apply(ctx, cmd)
			require.NoError(t, err)
			require.True(t, result.Applied)
			require.Zero(t, *result.NewBalance)
		})
	}
}

func TestUsageBalanceFloor_ConcurrentBatchAndSingle(t *testing.T) {
	ctx := context.Background()
	client := testEntClient(t)
	repo := NewUsageBillingRepository(client, integrationDB)
	user := mustCreateUser(t, client, &service.User{Email: uuid.NewString() + "@test.com", PasswordHash: "hash", Balance: 1})
	key := mustCreateApiKey(t, client, &service.APIKey{UserID: user.ID, Key: "sk-" + uuid.NewString(), Name: "concurrent-batch"})
	var wg sync.WaitGroup
	start := make(chan struct{})
	errs := make(chan error, 10)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			cmd := &service.UsageBillingCommand{RequestID: uuid.NewString(), APIKeyID: key.ID, UserID: user.ID, BalanceCost: 0.5}
			var err error
			if i%2 == 0 {
				_, err = repo.ApplyBatch(ctx, []*service.UsageBillingCommand{cmd, {RequestID: uuid.NewString(), APIKeyID: key.ID, UserID: user.ID, BalanceCost: 0.5}})
			} else {
				_, err = repo.Apply(ctx, cmd)
			}
			errs <- err
		}(i)
	}
	close(start)
	wg.Wait()
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}
	var balance, charged, absorbed float64
	var count int
	require.NoError(t, integrationDB.QueryRowContext(ctx, "SELECT balance FROM users WHERE id=$1", user.ID).Scan(&balance))
	require.Zero(t, balance)
	require.NoError(t, integrationDB.QueryRowContext(ctx, "SELECT COUNT(*),SUM(charged_amount),SUM(absorbed_amount) FROM usage_balance_settlements WHERE user_id=$1", user.ID).Scan(&count, &charged, &absorbed))
	require.Equal(t, 15, count)
	require.Equal(t, 1.0, charged)
	require.Equal(t, 6.5, absorbed)
}
