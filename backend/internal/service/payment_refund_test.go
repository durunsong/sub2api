//go:build unit

package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/paymentauditlog"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestValidateRefundRequestRejectsLegacyGuessedProviderInstance(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)

	user, err := client.User.Create().
		SetEmail("refund-legacy@example.com").
		SetPasswordHash("hash").
		SetUsername("refund-legacy-user").
		Save(ctx)
	require.NoError(t, err)

	_, err = client.PaymentProviderInstance.Create().
		SetProviderKey(payment.TypeAlipay).
		SetName("alipay-refund-instance").
		SetConfig("{}").
		SetSupportedTypes("alipay").
		SetEnabled(true).
		SetAllowUserRefund(true).
		SetRefundEnabled(true).
		Save(ctx)
	require.NoError(t, err)

	order, err := client.PaymentOrder.Create().
		SetUserID(user.ID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetAmount(88).
		SetPayAmount(88).
		SetFeeRate(0).
		SetRechargeCode("REFUND-LEGACY-ORDER").
		SetOutTradeNo("sub2_refund_legacy_order").
		SetPaymentType(payment.TypeAlipay).
		SetPaymentTradeNo("trade-legacy-refund").
		SetOrderType(payment.OrderTypeBalance).
		SetStatus(OrderStatusCompleted).
		SetExpiresAt(time.Now().Add(time.Hour)).
		SetPaidAt(time.Now()).
		SetClientIP("127.0.0.1").
		SetSrcHost("api.example.com").
		Save(ctx)
	require.NoError(t, err)

	svc := &PaymentService{
		entClient: client,
	}

	_, err = svc.validateRefundRequest(ctx, order.ID, user.ID)
	require.Error(t, err)
	require.Equal(t, "USER_REFUND_DISABLED", infraerrors.Reason(err))
}

func TestPrepareRefundRejectsLegacyGuessedProviderInstance(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)

	user, err := client.User.Create().
		SetEmail("refund-legacy-admin@example.com").
		SetPasswordHash("hash").
		SetUsername("refund-legacy-admin-user").
		Save(ctx)
	require.NoError(t, err)

	_, err = client.PaymentProviderInstance.Create().
		SetProviderKey(payment.TypeAlipay).
		SetName("alipay-refund-admin-instance").
		SetConfig("{}").
		SetSupportedTypes("alipay").
		SetEnabled(true).
		SetAllowUserRefund(true).
		SetRefundEnabled(true).
		Save(ctx)
	require.NoError(t, err)

	order, err := client.PaymentOrder.Create().
		SetUserID(user.ID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetAmount(188).
		SetPayAmount(188).
		SetFeeRate(0).
		SetRechargeCode("REFUND-LEGACY-ADMIN-ORDER").
		SetOutTradeNo("sub2_refund_legacy_admin_order").
		SetPaymentType(payment.TypeAlipay).
		SetPaymentTradeNo("trade-legacy-admin-refund").
		SetOrderType(payment.OrderTypeBalance).
		SetStatus(OrderStatusCompleted).
		SetExpiresAt(time.Now().Add(time.Hour)).
		SetPaidAt(time.Now()).
		SetClientIP("127.0.0.1").
		SetSrcHost("api.example.com").
		Save(ctx)
	require.NoError(t, err)

	svc := &PaymentService{
		entClient: client,
	}

	plan, result, err := svc.PrepareRefund(ctx, order.ID, 0, "", false, false)
	require.Nil(t, plan)
	require.Nil(t, result)
	require.Error(t, err)
	require.Equal(t, "REFUND_DISABLED", infraerrors.Reason(err))
}

func TestPrepDeductBalanceRequiresForceWhenBalanceIsInsufficient(t *testing.T) {
	for _, tc := range []struct {
		name        string
		balance     float64
		force       bool
		wantDeduct  float64
		wantWarning bool
	}{
		{name: "insufficient balance", balance: 40, wantWarning: true},
		{name: "forced insufficient balance", balance: 40, force: true, wantDeduct: 40},
		{name: "equal balance", balance: 100, wantDeduct: 100},
	} {
		t.Run(tc.name, func(t *testing.T) {
			plan := &RefundPlan{RefundAmount: 100}
			svc := &PaymentService{userRepo: &mockUserRepo{getByIDUser: &User{Balance: tc.balance}}}

			result := svc.prepDeduct(context.Background(), &dbent.PaymentOrder{
				UserID:    1,
				OrderType: payment.OrderTypeBalance,
			}, plan, tc.force)

			if tc.wantWarning {
				require.NotNil(t, result)
				require.False(t, result.Success)
				require.True(t, result.RequireForce)
				require.Equal(t, "user balance is insufficient for deduction, use force", result.Warning)
				require.Zero(t, plan.BalanceToDeduct)
				return
			}
			require.Nil(t, result)
			require.Equal(t, payment.DeductionTypeBalance, plan.DeductionType)
			require.Equal(t, tc.wantDeduct, plan.BalanceToDeduct)
		})
	}
}

type recordingBalanceCacheInvalidator struct {
	calls  []int64
	err    error
	onCall func(context.Context, int64)
}

func (r *recordingBalanceCacheInvalidator) InvalidateUserBalance(ctx context.Context, userID int64) error {
	r.calls = append(r.calls, userID)
	if r.onCall != nil {
		r.onCall(ctx, userID)
	}
	return r.err
}

func TestExecuteRefundUsesActualAvailableBalanceDeduction(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	user, err := client.User.Create().
		SetEmail("refund-execute-clamp@example.com").
		SetPasswordHash("hash").
		SetUsername("refund-execute-clamp").
		Save(ctx)
	require.NoError(t, err)
	order, err := client.PaymentOrder.Create().
		SetUserID(user.ID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetAmount(100).
		SetPayAmount(100).
		SetFeeRate(0).
		SetRechargeCode("REFUND-EXECUTE-CLAMP").
		SetOutTradeNo("refund_execute_clamp").
		SetPaymentType(payment.TypeStripe).
		SetPaymentTradeNo("").
		SetOrderType(payment.OrderTypeBalance).
		SetStatus(OrderStatusCompleted).
		SetExpiresAt(time.Now().Add(time.Hour)).
		SetPaidAt(time.Now()).
		SetClientIP("127.0.0.1").
		SetSrcHost("api.example.com").
		Save(ctx)
	require.NoError(t, err)

	deducted := false
	repo := &mockUserRepo{deductAvailableBalanceFn: func(_ context.Context, id int64, amount float64) (float64, error) {
		require.Equal(t, user.ID, id)
		require.Equal(t, 100.0, amount)
		deducted = true
		return 25, nil
	}}
	plan := &RefundPlan{
		OrderID: order.ID, Order: order, RefundAmount: 100, GatewayAmount: 100,
		Reason: "concurrent spend", Force: true, DeductionType: payment.DeductionTypeBalance, BalanceToDeduct: 100,
	}

	cache := &recordingBalanceCacheInvalidator{onCall: func(context.Context, int64) { require.True(t, deducted) }}
	result, err := (&PaymentService{entClient: client, userRepo: repo, balanceCacheInvalidator: cache}).ExecuteRefund(ctx, plan)
	require.NoError(t, err)
	require.Equal(t, []int64{user.ID}, cache.calls)
	require.True(t, result.Success)
	require.Equal(t, 25.0, plan.BalanceToDeduct)
	require.Equal(t, 25.0, result.BalanceDeducted)
	audit, err := client.PaymentAuditLog.Query().
		Where(paymentauditlog.OrderIDEQ(strconv.FormatInt(order.ID, 10)), paymentauditlog.ActionEQ("REFUND_SUCCESS")).
		Only(ctx)
	require.NoError(t, err)
	require.Contains(t, audit.Detail, `"balanceDeducted":25`)
}

func TestExecuteRefundIgnoresBalanceCacheInvalidationFailure(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	user, err := client.User.Create().
		SetEmail("refund-cache-failure@example.com").
		SetPasswordHash("hash").
		SetUsername("refund-cache-failure").
		Save(ctx)
	require.NoError(t, err)
	order, err := client.PaymentOrder.Create().
		SetUserID(user.ID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetAmount(50).
		SetPayAmount(50).
		SetFeeRate(0).
		SetRechargeCode("REFUND-CACHE-FAILURE").
		SetOutTradeNo("refund_cache_failure").
		SetPaymentType(payment.TypeStripe).
		SetPaymentTradeNo("").
		SetOrderType(payment.OrderTypeBalance).
		SetStatus(OrderStatusCompleted).
		SetExpiresAt(time.Now().Add(time.Hour)).
		SetPaidAt(time.Now()).
		SetClientIP("127.0.0.1").
		SetSrcHost("api.example.com").
		Save(ctx)
	require.NoError(t, err)

	var logs bytes.Buffer
	previousLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(previousLogger) })

	cache := &recordingBalanceCacheInvalidator{err: errors.New("redis unavailable")}
	svc := &PaymentService{
		entClient: client,
		userRepo: &mockUserRepo{deductAvailableBalanceFn: func(context.Context, int64, float64) (float64, error) {
			return 50, nil
		}},
		balanceCacheInvalidator: cache,
	}
	result, err := svc.ExecuteRefund(ctx, &RefundPlan{
		OrderID: order.ID, Order: order, RefundAmount: 50, GatewayAmount: 50,
		Reason: "cache failure", DeductionType: payment.DeductionTypeBalance, BalanceToDeduct: 50,
	})

	require.NoError(t, err)
	require.True(t, result.Success)
	require.Equal(t, []int64{user.ID}, cache.calls)
	require.Contains(t, logs.String(), "level=ERROR")
	require.Contains(t, logs.String(), "msg=payment_refund.balance_cache_invalidation_failed")
	require.Contains(t, logs.String(), "operation=deduct")
	require.Contains(t, logs.String(), fmt.Sprintf("order_id=%d", order.ID))
	require.Contains(t, logs.String(), fmt.Sprintf("user_id=%d", user.ID))
	require.Contains(t, logs.String(), `error="redis unavailable"`)
}

func TestGwRefundRejectsAlipayMerchantIdentitySnapshotMismatch(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)

	user, err := client.User.Create().
		SetEmail("refund-snapshot-mismatch@example.com").
		SetPasswordHash("hash").
		SetUsername("refund-snapshot-mismatch-user").
		Save(ctx)
	require.NoError(t, err)

	inst, err := client.PaymentProviderInstance.Create().
		SetProviderKey(payment.TypeAlipay).
		SetName("alipay-refund-mismatch-instance").
		SetConfig(encryptWebhookProviderConfig(t, map[string]string{
			"appId":      "runtime-alipay-app",
			"privateKey": "runtime-private-key",
		})).
		SetSupportedTypes("alipay").
		SetEnabled(true).
		SetRefundEnabled(true).
		Save(ctx)
	require.NoError(t, err)

	instID := strconv.FormatInt(inst.ID, 10)
	order, err := client.PaymentOrder.Create().
		SetUserID(user.ID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetAmount(88).
		SetPayAmount(88).
		SetFeeRate(0).
		SetRechargeCode("REFUND-SNAPSHOT-MISMATCH-ORDER").
		SetOutTradeNo("sub2_refund_snapshot_mismatch_order").
		SetPaymentType(payment.TypeAlipay).
		SetPaymentTradeNo("trade-refund-snapshot-mismatch").
		SetOrderType(payment.OrderTypeBalance).
		SetStatus(OrderStatusCompleted).
		SetExpiresAt(time.Now().Add(time.Hour)).
		SetPaidAt(time.Now()).
		SetClientIP("127.0.0.1").
		SetSrcHost("api.example.com").
		SetProviderInstanceID(instID).
		SetProviderKey(payment.TypeAlipay).
		SetProviderSnapshot(map[string]any{
			"schema_version":       2,
			"provider_instance_id": instID,
			"provider_key":         payment.TypeAlipay,
			"merchant_app_id":      "expected-alipay-app",
		}).
		Save(ctx)
	require.NoError(t, err)

	svc := &PaymentService{
		entClient:    client,
		loadBalancer: newWebhookProviderTestLoadBalancer(client),
	}

	_, err = svc.gwRefund(ctx, &RefundPlan{
		OrderID:       order.ID,
		Order:         order,
		RefundAmount:  order.Amount,
		GatewayAmount: order.Amount,
		Reason:        "snapshot mismatch",
	})
	require.ErrorContains(t, err, "alipay app_id mismatch")
}

func TestCalculateGatewayRefundAmountUsesCurrencyPrecision(t *testing.T) {
	require.InDelta(t, 6.173, calculateGatewayRefundAmount(100, 12.345, 50, "KWD"), 1e-12)
	require.InDelta(t, 12.345, calculateGatewayRefundAmount(100, 12.345, 100, "KWD"), 1e-12)
	require.InDelta(t, 52, calculateGatewayRefundAmount(100, 103, 50, "JPY"), 1e-12)
}

func TestFormatGatewayRefundAmountUsesOrderCurrency(t *testing.T) {
	order := &dbent.PaymentOrder{
		ProviderSnapshot: map[string]any{
			"currency": "KWD",
		},
	}

	require.Equal(t, "12.345", formatGatewayRefundAmount(12.345, order))
}

func TestValidateRefundProviderResponseAcceptsPending(t *testing.T) {
	require.NoError(t, validateRefundProviderResponse(&payment.RefundResponse{Status: payment.ProviderStatusPending}))
	require.NoError(t, validateRefundProviderResponse(&payment.RefundResponse{Status: payment.ProviderStatusSuccess}))
	require.Error(t, validateRefundProviderResponse(&payment.RefundResponse{Status: payment.ProviderStatusFailed}))
	require.Error(t, validateRefundProviderResponse(nil))
}

func TestFinishRefundPendingMarksOrderPendingAndRollsBackDeduction(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)

	user, err := client.User.Create().
		SetEmail("refund-pending@example.com").
		SetPasswordHash("hash").
		SetUsername("refund-pending-user").
		Save(ctx)
	require.NoError(t, err)

	order, err := client.PaymentOrder.Create().
		SetUserID(user.ID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetAmount(100).
		SetPayAmount(100).
		SetFeeRate(0).
		SetRechargeCode("REFUND-PENDING-ORDER").
		SetOutTradeNo("sub2_refund_pending_order").
		SetPaymentType(payment.TypeStripe).
		SetPaymentTradeNo("pi_refund_pending").
		SetOrderType(payment.OrderTypeBalance).
		SetStatus(OrderStatusRefunding).
		SetExpiresAt(time.Now().Add(time.Hour)).
		SetPaidAt(time.Now()).
		SetClientIP("127.0.0.1").
		SetSrcHost("api.example.com").
		Save(ctx)
	require.NoError(t, err)

	var rolledBack float64
	userRepo := &mockUserRepo{}
	userRepo.updateBalanceFn = func(ctx context.Context, id int64, amount float64) error {
		require.Equal(t, user.ID, id)
		rolledBack += amount
		return nil
	}
	svc := &PaymentService{
		entClient: client,
		userRepo:  userRepo,
	}
	plan := &RefundPlan{
		OrderID:         order.ID,
		Order:           order,
		RefundAmount:    40,
		GatewayAmount:   40,
		Reason:          "gateway accepted but not final",
		Force:           true,
		DeductionType:   payment.DeductionTypeBalance,
		BalanceToDeduct: 40,
	}

	result, err := svc.finishRefund(ctx, plan, &payment.RefundResponse{Status: payment.ProviderStatusPending})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.False(t, result.Success)
	require.Contains(t, result.Warning, "pending confirmation")
	require.Equal(t, 40.0, rolledBack)
	require.Zero(t, plan.BalanceToDeduct)

	reloaded, err := client.PaymentOrder.Get(ctx, order.ID)
	require.NoError(t, err)
	require.Equal(t, OrderStatusRefundPending, reloaded.Status)
	require.Equal(t, 40.0, reloaded.RefundAmount)
	require.NotNil(t, reloaded.RefundReason)
	require.Equal(t, "gateway accepted but not final", *reloaded.RefundReason)
	require.Nil(t, reloaded.RefundAt)

	pendingAudits, err := client.PaymentAuditLog.Query().
		Where(paymentauditlog.OrderIDEQ(strconv.FormatInt(order.ID, 10)), paymentauditlog.ActionEQ("REFUND_PENDING")).
		Count(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, pendingAudits)
	successAudits, err := client.PaymentAuditLog.Query().
		Where(paymentauditlog.OrderIDEQ(strconv.FormatInt(order.ID, 10)), paymentauditlog.ActionEQ("REFUND_SUCCESS")).
		Count(ctx)
	require.NoError(t, err)
	require.Zero(t, successAudits)
}

func TestFinishRefundSuccessStatusesFinalize(t *testing.T) {
	for _, status := range []string{payment.ProviderStatusSuccess, payment.ProviderStatusRefunded} {
		t.Run(status, func(t *testing.T) {
			ctx := context.Background()
			client := newPaymentConfigServiceTestClient(t)

			user, err := client.User.Create().
				SetEmail("refund-success-" + status + "@example.com").
				SetPasswordHash("hash").
				SetUsername("refund-success-" + status).
				Save(ctx)
			require.NoError(t, err)

			order, err := client.PaymentOrder.Create().
				SetUserID(user.ID).
				SetUserEmail(user.Email).
				SetUserName(user.Username).
				SetAmount(100).
				SetPayAmount(100).
				SetFeeRate(0).
				SetRechargeCode("REFUND-SUCCESS-" + status).
				SetOutTradeNo("sub2_refund_success_" + status).
				SetPaymentType(payment.TypeStripe).
				SetPaymentTradeNo("pi_refund_success_" + status).
				SetOrderType(payment.OrderTypeBalance).
				SetStatus(OrderStatusRefunding).
				SetExpiresAt(time.Now().Add(time.Hour)).
				SetPaidAt(time.Now()).
				SetClientIP("127.0.0.1").
				SetSrcHost("api.example.com").
				Save(ctx)
			require.NoError(t, err)

			svc := &PaymentService{entClient: client}
			plan := &RefundPlan{
				OrderID:         order.ID,
				Order:           order,
				RefundAmount:    100,
				GatewayAmount:   100,
				Reason:          "final success",
				DeductionType:   payment.DeductionTypeBalance,
				BalanceToDeduct: 100,
			}

			result, err := svc.finishRefund(ctx, plan, &payment.RefundResponse{Status: status})
			require.NoError(t, err)
			require.NotNil(t, result)
			require.True(t, result.Success)
			require.Equal(t, 100.0, result.BalanceDeducted)

			reloaded, err := client.PaymentOrder.Get(ctx, order.ID)
			require.NoError(t, err)
			require.Equal(t, OrderStatusRefunded, reloaded.Status)
			require.NotNil(t, reloaded.RefundAt)

			successAudits, err := client.PaymentAuditLog.Query().
				Where(paymentauditlog.OrderIDEQ(strconv.FormatInt(order.ID, 10)), paymentauditlog.ActionEQ("REFUND_SUCCESS")).
				Count(ctx)
			require.NoError(t, err)
			require.Equal(t, 1, successAudits)
			pendingAudits, err := client.PaymentAuditLog.Query().
				Where(paymentauditlog.OrderIDEQ(strconv.FormatInt(order.ID, 10)), paymentauditlog.ActionEQ("REFUND_PENDING")).
				Count(ctx)
			require.NoError(t, err)
			require.Zero(t, pendingAudits)
		})
	}
}

func TestQueryAndFinalizeRefundFinalizesProviderStatuses(t *testing.T) {
	for _, tc := range []struct {
		name       string
		status     string
		wantStatus string
		wantDeduct float64
		available  float64
	}{
		{name: "success", status: payment.ProviderStatusSuccess, wantStatus: OrderStatusRefunded, wantDeduct: 100, available: 100},
		{name: "success clamps current balance", status: payment.ProviderStatusSuccess, wantStatus: OrderStatusRefunded, wantDeduct: 35, available: 35},
		{name: "failed", status: payment.ProviderStatusFailed, wantStatus: OrderStatusRefundFailed},
		{name: "pending", status: payment.ProviderStatusPending, wantStatus: OrderStatusRefundPending},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			client := newPaymentConfigServiceTestClient(t)
			order := createPendingRefundOrderForTest(t, ctx, client, "query-finalize-"+tc.name)

			var deducted float64
			svc := &PaymentService{
				entClient:    client,
				loadBalancer: &captureLoadBalancer{},
				userRepo: &mockUserRepo{deductAvailableBalanceFn: func(ctx context.Context, id int64, amount float64) (float64, error) {
					deducted += tc.available
					return tc.available, nil
				}},
			}
			restore := replacePaymentProviderFactoryForTest(t, &refundQueryProviderTestDouble{
				refundResponse: &payment.RefundResponse{RefundID: "rf_test", Status: tc.status},
			})
			defer restore()

			result, err := svc.QueryAndFinalizeRefund(ctx, order.ID)
			require.NoError(t, err)
			require.NotNil(t, result)
			require.Equal(t, tc.status == payment.ProviderStatusSuccess, result.Success)
			require.Equal(t, tc.wantDeduct, deducted)
			if tc.status == payment.ProviderStatusSuccess {
				require.Equal(t, tc.wantDeduct, result.BalanceDeducted)
				audit, err := client.PaymentAuditLog.Query().
					Where(paymentauditlog.OrderIDEQ(strconv.FormatInt(order.ID, 10)), paymentauditlog.ActionEQ("REFUND_SUCCESS")).
					Only(ctx)
				require.NoError(t, err)
				require.Contains(t, audit.Detail, fmt.Sprintf(`"balanceDeducted":%v`, tc.wantDeduct))
			}

			reloaded, err := client.PaymentOrder.Get(ctx, order.ID)
			require.NoError(t, err)
			require.Equal(t, tc.wantStatus, reloaded.Status)
		})
	}
}

func TestFinalizePendingRefundSuccessRejectsStaleCallerBeforeSecondDeduction(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	order := createPendingRefundOrderForTest(t, ctx, client, "finalize-stale")

	deductions := 0
	cache := &recordingBalanceCacheInvalidator{onCall: func(cacheCtx context.Context, userID int64) {
		require.Nil(t, dbent.TxFromContext(cacheCtx))
		reloaded, reloadErr := client.PaymentOrder.Get(cacheCtx, order.ID)
		require.NoError(t, reloadErr)
		require.Equal(t, OrderStatusRefunded, reloaded.Status)
	}}
	svc := &PaymentService{
		entClient: client,
		userRepo: &mockUserRepo{deductAvailableBalanceFn: func(ctx context.Context, id int64, amount float64) (float64, error) {
			require.NotNil(t, dbent.TxFromContext(ctx))
			deductions++
			return amount, nil
		}},
		balanceCacheInvalidator: cache,
	}

	first, err := svc.finalizePendingRefundSuccess(ctx, svc.refundFinalizePlan(order))
	require.NoError(t, err)
	require.True(t, first.Success)

	second, err := svc.finalizePendingRefundSuccess(ctx, svc.refundFinalizePlan(order))
	require.Nil(t, second)
	require.Error(t, err)
	require.Equal(t, "CONFLICT", infraerrors.Reason(err))
	require.Equal(t, 1, deductions)
	require.Equal(t, []int64{order.UserID}, cache.calls)

	successAudits, err := client.PaymentAuditLog.Query().
		Where(paymentauditlog.OrderIDEQ(strconv.FormatInt(order.ID, 10)), paymentauditlog.ActionEQ("REFUND_SUCCESS")).
		Count(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, successAudits)
}

func TestFinalizePendingRefundSuccessRollsBackPostDeductionFailure(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	order := createPendingRefundOrderForTest(t, ctx, client, "finalize-rollback")
	_, err := client.User.UpdateOneID(order.UserID).SetBalance(100).Save(ctx)
	require.NoError(t, err)

	cache := &recordingBalanceCacheInvalidator{}
	svc := &PaymentService{
		entClient: client,
		userRepo: &mockUserRepo{deductAvailableBalanceFn: func(ctx context.Context, id int64, amount float64) (float64, error) {
			tx := dbent.TxFromContext(ctx)
			require.NotNil(t, tx)
			if _, updateErr := tx.Client().User.UpdateOneID(id).AddBalance(-amount).Save(ctx); updateErr != nil {
				return 0, updateErr
			}
			return 0, errors.New("injected failure after deduction")
		}},
		balanceCacheInvalidator: cache,
	}

	result, err := svc.finalizePendingRefundSuccess(ctx, svc.refundFinalizePlan(order))
	require.Nil(t, result)
	require.ErrorContains(t, err, "injected failure after deduction")

	user, err := client.User.Get(ctx, order.UserID)
	require.NoError(t, err)
	require.Equal(t, 100.0, user.Balance)
	reloaded, err := client.PaymentOrder.Get(ctx, order.ID)
	require.NoError(t, err)
	require.Equal(t, OrderStatusRefundPending, reloaded.Status)
	successAudits, err := client.PaymentAuditLog.Query().
		Where(paymentauditlog.OrderIDEQ(strconv.FormatInt(order.ID, 10)), paymentauditlog.ActionEQ("REFUND_SUCCESS")).
		Count(ctx)
	require.NoError(t, err)
	require.Zero(t, successAudits)
	require.Empty(t, cache.calls)
}

func TestRollbackRefundInvalidatesRestoredBalance(t *testing.T) {
	restored := false
	cache := &recordingBalanceCacheInvalidator{onCall: func(context.Context, int64) { require.True(t, restored) }}
	repo := &mockUserRepo{updateBalanceFn: func(context.Context, int64, float64) error {
		restored = true
		return nil
	}}
	svc := &PaymentService{userRepo: repo, balanceCacheInvalidator: cache}
	plan := &RefundPlan{
		OrderID:         9,
		Order:           &dbent.PaymentOrder{UserID: 42},
		DeductionType:   payment.DeductionTypeBalance,
		BalanceToDeduct: 12,
	}

	require.True(t, svc.RollbackRefund(context.Background(), plan, errors.New("gateway failed")))
	require.Equal(t, []int64{42}, cache.calls)
}

func TestRollbackRefundDoesNotInvalidateWhenBalanceRestoreFails(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	cache := &recordingBalanceCacheInvalidator{}
	repo := &mockUserRepo{updateBalanceErr: errors.New("database unavailable")}
	svc := &PaymentService{entClient: client, userRepo: repo, balanceCacheInvalidator: cache}
	plan := &RefundPlan{
		OrderID:         10,
		Order:           &dbent.PaymentOrder{UserID: 43},
		DeductionType:   payment.DeductionTypeBalance,
		BalanceToDeduct: 12,
	}

	require.False(t, svc.RollbackRefund(ctx, plan, errors.New("gateway failed")))
	require.Empty(t, cache.calls)
}

func TestQueryAndFinalizeRefundUnsupportedProviderReturnsClearError(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	order := createPendingRefundOrderForTest(t, ctx, client, "query-finalize-unsupported")
	svc := &PaymentService{entClient: client, loadBalancer: &captureLoadBalancer{}}
	restore := replacePaymentProviderFactoryForTest(t, refundProviderTestDouble{})
	defer restore()

	result, err := svc.QueryAndFinalizeRefund(ctx, order.ID)
	require.Nil(t, result)
	require.Error(t, err)
	require.Equal(t, "REFUND_QUERY_UNSUPPORTED", infraerrors.Reason(err))
}

func createPendingRefundOrderForTest(t *testing.T, ctx context.Context, client *dbent.Client, suffix string) *dbent.PaymentOrder {
	t.Helper()

	user, err := client.User.Create().
		SetEmail(suffix + "@example.com").
		SetPasswordHash("hash").
		SetUsername(suffix).
		Save(ctx)
	require.NoError(t, err)

	inst, err := client.PaymentProviderInstance.Create().
		SetProviderKey(payment.TypeStripe).
		SetName(suffix + "-provider").
		SetConfig("{}").
		SetSupportedTypes("stripe").
		SetEnabled(true).
		SetRefundEnabled(true).
		Save(ctx)
	require.NoError(t, err)

	order, err := client.PaymentOrder.Create().
		SetUserID(user.ID).
		SetUserEmail(user.Email).
		SetUserName(user.Username).
		SetAmount(100).
		SetPayAmount(100).
		SetFeeRate(0).
		SetRechargeCode("REFUND-" + suffix).
		SetOutTradeNo("sub2_" + suffix).
		SetPaymentType(payment.TypeStripe).
		SetPaymentTradeNo("pi_" + suffix).
		SetOrderType(payment.OrderTypeBalance).
		SetStatus(OrderStatusRefundPending).
		SetRefundAmount(100).
		SetRefundReason("pending refund").
		SetExpiresAt(time.Now().Add(time.Hour)).
		SetPaidAt(time.Now()).
		SetClientIP("127.0.0.1").
		SetSrcHost("api.example.com").
		SetProviderInstanceID(strconv.FormatInt(inst.ID, 10)).
		Save(ctx)
	require.NoError(t, err)

	_, err = client.PaymentAuditLog.Create().
		SetOrderID(strconv.FormatInt(order.ID, 10)).
		SetAction("REFUND_PENDING").
		SetOperator("admin").
		SetDetail(`{"refundID":"rf_test","deductionRollbackOK":true}`).
		Save(ctx)
	require.NoError(t, err)
	return order
}

func replacePaymentProviderFactoryForTest(t *testing.T, prov payment.Provider) func() {
	t.Helper()
	original := createPaymentProviderFromInstance
	createPaymentProviderFromInstance = func(providerKey, instanceID string, config map[string]string) (payment.Provider, error) {
		return prov, nil
	}
	return func() { createPaymentProviderFromInstance = original }
}

type refundProviderTestDouble struct{}

func (refundProviderTestDouble) Name() string { return "refund-test" }
func (refundProviderTestDouble) ProviderKey() string {
	return payment.TypeStripe
}
func (refundProviderTestDouble) SupportedTypes() []payment.PaymentType {
	return []payment.PaymentType{payment.TypeStripe}
}
func (refundProviderTestDouble) CreatePayment(context.Context, payment.CreatePaymentRequest) (*payment.CreatePaymentResponse, error) {
	return nil, nil
}
func (refundProviderTestDouble) QueryOrder(context.Context, string) (*payment.QueryOrderResponse, error) {
	return nil, nil
}
func (refundProviderTestDouble) VerifyNotification(context.Context, string, map[string]string) (*payment.PaymentNotification, error) {
	return nil, nil
}
func (refundProviderTestDouble) Refund(context.Context, payment.RefundRequest) (*payment.RefundResponse, error) {
	return nil, nil
}

type refundQueryProviderTestDouble struct {
	refundProviderTestDouble
	refundResponse *payment.RefundResponse
}

func (p *refundQueryProviderTestDouble) QueryRefund(context.Context, payment.RefundQueryRequest) (*payment.RefundResponse, error) {
	return p.refundResponse, nil
}

func TestPrepareRefundVoidsUnconsumedSourceCardInsteadOfDeductingSubscription(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	user, err := client.User.Create().SetEmail("refund-card@example.com").SetPasswordHash("hash").SetUsername("refund-card").Save(ctx)
	require.NoError(t, err)
	inst, err := client.PaymentProviderInstance.Create().SetProviderKey(payment.TypeStripe).SetName("refund-card-provider").SetConfig("{}").SetSupportedTypes("stripe").SetEnabled(true).SetRefundEnabled(true).Save(ctx)
	require.NoError(t, err)
	group, err := client.Group.Create().SetName("refund-card-group").SetPlatform("openai").SetSubscriptionType("subscription").Save(ctx)
	require.NoError(t, err)
	sub, err := client.UserSubscription.Create().SetUserID(user.ID).SetGroupID(group.ID).SetStartsAt(time.Now()).SetExpiresAt(time.Now().Add(30 * 24 * time.Hour)).SetStatus(SubscriptionStatusActive).SetManualResetCredits(1).Save(ctx)
	require.NoError(t, err)
	days := 30
	order, err := client.PaymentOrder.Create().SetUserID(user.ID).SetUserEmail(user.Email).SetUserName(user.Username).SetAmount(100).SetPayAmount(100).SetFeeRate(0).SetRechargeCode("REFUND-CARD").SetOutTradeNo("refund_card").SetPaymentType(payment.TypeStripe).SetPaymentTradeNo("").SetOrderType(payment.OrderTypeSubscription).SetStatus(OrderStatusCompleted).SetExpiresAt(time.Now().Add(time.Hour)).SetPaidAt(time.Now()).SetClientIP("127.0.0.1").SetSrcHost("api.example.com").SetProviderInstanceID(strconv.FormatInt(inst.ID, 10)).SetSubscriptionGroupID(group.ID).SetSubscriptionDays(days).Save(ctx)
	require.NoError(t, err)
	_, err = client.ExecContext(ctx, `CREATE TABLE user_subscription_reset_cards (id INTEGER PRIMARY KEY AUTOINCREMENT, user_subscription_id INTEGER NOT NULL, validity_days INTEGER NOT NULL, source_type TEXT NOT NULL, source_reference TEXT NOT NULL, consumed_at DATETIME, voided_at DATETIME)`)
	require.NoError(t, err)
	_, err = client.ExecContext(ctx, `INSERT INTO user_subscription_reset_cards (user_subscription_id, validity_days, source_type, source_reference) VALUES ($1,$2,'payment_order',$3)`, sub.ID, days, strconv.FormatInt(order.ID, 10))
	require.NoError(t, err)
	subRepo := &resetCardRepoStub{subscriptionUserSubRepoStub: newSubscriptionUserSubRepoStub()}
	subRepo.seed(&UserSubscription{ID: sub.ID, UserID: user.ID, GroupID: group.ID, StartsAt: sub.StartsAt, ExpiresAt: sub.ExpiresAt, Status: sub.Status, ManualResetCredits: 1})
	subSvc := NewSubscriptionService(groupRepoNoop{}, subRepo, nil, client, nil)
	svc := &PaymentService{entClient: client, subscriptionSvc: subSvc}
	plan, early, err := svc.PrepareRefund(ctx, order.ID, 0, "", false, true)
	require.NoError(t, err)
	require.Nil(t, early)
	require.Zero(t, plan.SubDaysToDeduct)
	require.True(t, plan.ResetCardToVoid)
}

func TestPrepareRefundConsumedSourceCardRequiresForceWithoutSubscriptionDeduction(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	_, err := client.ExecContext(ctx, `CREATE TABLE user_subscription_reset_cards (id INTEGER PRIMARY KEY AUTOINCREMENT, user_subscription_id INTEGER NOT NULL, source_type TEXT NOT NULL, source_reference TEXT NOT NULL, consumed_at DATETIME, voided_at DATETIME)`)
	require.NoError(t, err)
	_, err = client.ExecContext(ctx, `INSERT INTO user_subscription_reset_cards (user_subscription_id, source_type, source_reference, consumed_at) VALUES (10,'payment_order','99',CURRENT_TIMESTAMP)`)
	require.NoError(t, err)
	days, groupID := 30, int64(3)
	order := &dbent.PaymentOrder{ID: 99, UserID: 7, OrderType: payment.OrderTypeSubscription, SubscriptionDays: &days, SubscriptionGroupID: &groupID}
	svc := &PaymentService{entClient: client}
	plan := &RefundPlan{}
	result := svc.prepDeduct(ctx, order, plan, false)
	require.NotNil(t, result)
	require.True(t, result.RequireForce)
	require.Zero(t, plan.SubDaysToDeduct)
	plan = &RefundPlan{}
	require.Nil(t, svc.prepDeduct(ctx, order, plan, true))
	require.Zero(t, plan.SubDaysToDeduct)
	require.False(t, plan.ResetCardToVoid)
}

type countingRefundProvider struct {
	refundProviderTestDouble
	refunds            int
	queries            int
	voidedDuringRefund bool
	inspectVoided      func() bool
	refundErr          error
	refundResponse     *payment.RefundResponse
	queryResponse      *payment.RefundResponse
}

func (p *countingRefundProvider) Refund(context.Context, payment.RefundRequest) (*payment.RefundResponse, error) {
	p.refunds++
	if p.inspectVoided != nil {
		p.voidedDuringRefund = p.inspectVoided()
	}
	if p.refundErr != nil {
		return nil, p.refundErr
	}
	if p.refundResponse != nil {
		return p.refundResponse, nil
	}
	return &payment.RefundResponse{Status: payment.ProviderStatusSuccess, RefundID: "rf_reset_card"}, nil
}

func (p *countingRefundProvider) QueryRefund(context.Context, payment.RefundQueryRequest) (*payment.RefundResponse, error) {
	p.queries++
	if p.queryResponse != nil {
		return p.queryResponse, nil
	}
	return &payment.RefundResponse{Status: payment.ProviderStatusSuccess, RefundID: "rf_reset_card"}, nil
}

type resetCardRefundFixture struct {
	client *dbent.Client
	svc    *PaymentService
	order  *dbent.PaymentOrder
	cardID int64
	plan   *RefundPlan
}

func createResetCardRefundFixture(t *testing.T, suffix, tradeNo, status string) resetCardRefundFixture {
	t.Helper()
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	user, err := client.User.Create().SetEmail(suffix + "@example.com").SetPasswordHash("hash").SetUsername(suffix).Save(ctx)
	require.NoError(t, err)
	inst, err := client.PaymentProviderInstance.Create().SetProviderKey(payment.TypeStripe).SetName(suffix + "-provider").SetConfig("{}").SetSupportedTypes("stripe").SetEnabled(true).SetRefundEnabled(true).Save(ctx)
	require.NoError(t, err)
	group, err := client.Group.Create().SetName(suffix + "-group").SetPlatform("openai").SetSubscriptionType("subscription").Save(ctx)
	require.NoError(t, err)
	sub, err := client.UserSubscription.Create().SetUserID(user.ID).SetGroupID(group.ID).SetStartsAt(time.Now()).SetExpiresAt(time.Now().Add(30 * 24 * time.Hour)).SetStatus(SubscriptionStatusActive).SetManualResetCredits(1).Save(ctx)
	require.NoError(t, err)
	days := 30
	order, err := client.PaymentOrder.Create().
		SetUserID(user.ID).SetUserEmail(user.Email).SetUserName(user.Username).
		SetAmount(100).SetPayAmount(100).SetFeeRate(0).
		SetRechargeCode("REFUND-" + suffix).SetOutTradeNo("refund_" + suffix).
		SetPaymentType(payment.TypeStripe).SetPaymentTradeNo(tradeNo).
		SetOrderType(payment.OrderTypeSubscription).SetStatus(status).
		SetExpiresAt(time.Now().Add(time.Hour)).SetPaidAt(time.Now()).
		SetClientIP("127.0.0.1").SetSrcHost("api.example.com").
		SetProviderInstanceID(strconv.FormatInt(inst.ID, 10)).
		SetSubscriptionGroupID(group.ID).SetSubscriptionDays(days).
		Save(ctx)
	require.NoError(t, err)
	_, err = client.ExecContext(ctx, `CREATE TABLE user_subscription_reset_cards (id INTEGER PRIMARY KEY AUTOINCREMENT, user_subscription_id INTEGER NOT NULL, validity_days INTEGER NOT NULL, source_type TEXT NOT NULL, source_reference TEXT NOT NULL, consumed_at DATETIME, voided_at DATETIME)`)
	require.NoError(t, err)
	_, err = client.ExecContext(ctx, `INSERT INTO user_subscription_reset_cards (user_subscription_id, validity_days, source_type, source_reference) VALUES ($1,$2,'payment_order',$3)`, sub.ID, days, strconv.FormatInt(order.ID, 10))
	require.NoError(t, err)
	cardID := scanResetCardID(t, client, ctx, order.ID)
	svc := &PaymentService{entClient: client, loadBalancer: &captureLoadBalancer{}}
	return resetCardRefundFixture{
		client: client,
		svc:    svc,
		order:  order,
		cardID: cardID,
		plan: &RefundPlan{
			OrderID:         order.ID,
			Order:           order,
			RefundAmount:    100,
			GatewayAmount:   100,
			Reason:          "void after gateway",
			DeductBalance:   true,
			DeductionType:   payment.DeductionTypeSubscription,
			ResetCardID:     cardID,
			ResetCardToVoid: true,
		},
	}
}

func scanResetCardID(t *testing.T, client *dbent.Client, ctx context.Context, orderID int64) int64 {
	t.Helper()
	rows, err := client.QueryContext(ctx, `SELECT id FROM user_subscription_reset_cards WHERE source_reference = $1`, strconv.FormatInt(orderID, 10))
	require.NoError(t, err)
	defer rows.Close()
	require.True(t, rows.Next())
	var cardID int64
	require.NoError(t, rows.Scan(&cardID))
	require.NoError(t, rows.Err())
	return cardID
}

func resetCardVoided(t *testing.T, client *dbent.Client, cardID int64) bool {
	t.Helper()
	rows, err := client.QueryContext(context.Background(), `SELECT voided_at IS NOT NULL FROM user_subscription_reset_cards WHERE id = $1`, cardID)
	require.NoError(t, err)
	defer rows.Close()
	require.True(t, rows.Next())
	var voided bool
	require.NoError(t, rows.Scan(&voided))
	require.NoError(t, rows.Err())
	return voided
}

func TestExecuteRefundVoidsResetCardOnlyAfterGatewaySuccess(t *testing.T) {
	fx := createResetCardRefundFixture(t, "void-after-gw", "pi_void_after_gw", OrderStatusCompleted)
	prov := &countingRefundProvider{inspectVoided: func() bool { return resetCardVoided(t, fx.client, fx.cardID) }}
	restore := replacePaymentProviderFactoryForTest(t, prov)
	defer restore()

	result, err := fx.svc.ExecuteRefund(context.Background(), fx.plan)
	require.NoError(t, err)
	require.True(t, result.Success)
	require.Equal(t, 1, prov.refunds)
	require.False(t, prov.voidedDuringRefund)
	require.True(t, resetCardVoided(t, fx.client, fx.cardID))
	require.True(t, fx.plan.ResetCardVoided)

	reloaded, err := fx.client.PaymentOrder.Get(context.Background(), fx.order.ID)
	require.NoError(t, err)
	require.Equal(t, OrderStatusRefunded, reloaded.Status)
	require.True(t, fx.svc.hasAuditLog(context.Background(), fx.order.ID, refundGatewaySucceededAudit))
	require.True(t, fx.svc.hasAuditLog(context.Background(), fx.order.ID, refundResetCardVoidedAudit))
}

func TestExecuteRefundGatewayFailureLeavesResetCardAvailable(t *testing.T) {
	fx := createResetCardRefundFixture(t, "gw-fail-card", "pi_gw_fail_card", OrderStatusCompleted)
	prov := &countingRefundProvider{refundErr: errors.New("gateway rejected refund")}
	restore := replacePaymentProviderFactoryForTest(t, prov)
	defer restore()

	result, err := fx.svc.ExecuteRefund(context.Background(), fx.plan)
	require.NoError(t, err)
	require.False(t, result.Success)
	require.Contains(t, result.Warning, "gateway failed")
	require.Equal(t, 1, prov.refunds)
	require.False(t, fx.plan.ResetCardVoided)
	require.False(t, resetCardVoided(t, fx.client, fx.cardID))

	reloaded, err := fx.client.PaymentOrder.Get(context.Background(), fx.order.ID)
	require.NoError(t, err)
	require.Equal(t, OrderStatusCompleted, reloaded.Status)
	require.False(t, fx.svc.hasAuditLog(context.Background(), fx.order.ID, refundGatewaySucceededAudit))
	require.False(t, fx.svc.hasAuditLog(context.Background(), fx.order.ID, refundResetCardVoidedAudit))
}

func TestExecuteRefundRefundingReentryFinalizesWithoutSecondGatewayCall(t *testing.T) {
	fx := createResetCardRefundFixture(t, "refunding-resume", "pi_refunding_resume", OrderStatusRefunding)
	fx.svc.writeAuditLog(context.Background(), fx.order.ID, refundGatewaySucceededAudit, "admin", map[string]any{"status": payment.ProviderStatusSuccess})
	prov := &countingRefundProvider{refundErr: errors.New("Refund must not be called on REFUNDING resume")}
	restore := replacePaymentProviderFactoryForTest(t, prov)
	defer restore()

	result, err := fx.svc.ExecuteRefund(context.Background(), fx.plan)
	require.NoError(t, err)
	require.True(t, result.Success)
	require.Zero(t, prov.refunds)
	require.Zero(t, prov.queries)
	require.True(t, resetCardVoided(t, fx.client, fx.cardID))

	reloaded, err := fx.client.PaymentOrder.Get(context.Background(), fx.order.ID)
	require.NoError(t, err)
	require.Equal(t, OrderStatusRefunded, reloaded.Status)
	successAudits, err := fx.client.PaymentAuditLog.Query().
		Where(paymentauditlog.OrderIDEQ(strconv.FormatInt(fx.order.ID, 10)), paymentauditlog.ActionEQ("REFUND_SUCCESS")).
		Count(context.Background())
	require.NoError(t, err)
	require.Equal(t, 1, successAudits)
}

func TestQueryAndFinalizeRefundAcceptsRefundingAndDoesNotCreateAnotherRefund(t *testing.T) {
	fx := createResetCardRefundFixture(t, "query-refunding", "pi_query_refunding", OrderStatusRefunding)
	prov := &countingRefundProvider{
		refundErr:     errors.New("Refund must not be called when querying a REFUNDING order"),
		queryResponse: &payment.RefundResponse{Status: payment.ProviderStatusSuccess, RefundID: "rf_query"},
	}
	restore := replacePaymentProviderFactoryForTest(t, prov)
	defer restore()

	result, err := fx.svc.QueryAndFinalizeRefund(context.Background(), fx.order.ID)
	require.NoError(t, err)
	require.True(t, result.Success)
	require.Zero(t, prov.refunds)
	require.Equal(t, 1, prov.queries)
	require.True(t, resetCardVoided(t, fx.client, fx.cardID))

	reloaded, err := fx.client.PaymentOrder.Get(context.Background(), fx.order.ID)
	require.NoError(t, err)
	require.Equal(t, OrderStatusRefunded, reloaded.Status)
}
