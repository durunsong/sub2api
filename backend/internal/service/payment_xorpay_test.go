//go:build unit

package service

import (
	"context"
	"crypto/md5"
	"fmt"
	"net/url"
	"strconv"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/stretchr/testify/require"
)

func TestXorPayWechatSource(t *testing.T) {
	for _, source := range []string{"xorpay", "xorpay_wxpay", " XORPAY_WXPAY "} {
		require.Equal(t, "xorpay_wxpay", NormalizeVisibleMethodSource(payment.TypeWxpay, source))
		key, ok := VisibleMethodProviderKeyForSource(payment.TypeWxpay, source)
		require.True(t, ok)
		require.Equal(t, payment.TypeXorPay, key)
	}
	require.Empty(t, NormalizeVisibleMethodSource(payment.TypeAlipay, "xorpay_wxpay"))
	for _, supported := range []string{"xorpay", "", "alipay"} {
		require.NotContains(t, enabledVisibleMethodsForProvider(payment.TypeXorPay, supported), payment.TypeWxpay)
	}
	require.Contains(t, enabledVisibleMethodsForProvider(payment.TypeXorPay, "xorpay,wxpay"), payment.TypeWxpay)
}

func TestXorPayWechatRoutingAndLimits(t *testing.T) {
	for _, source := range []string{"xorpay_wxpay", "official_wxpay", "easypay_wxpay"} {
		t.Run(source, func(t *testing.T) {
			ctx := context.Background()
			client := newPaymentConfigServiceTestClient(t)
			instances := make([]*dbent.PaymentProviderInstance, 0, 3)
			for i, key := range []string{payment.TypeWxpay, payment.TypeEasyPay, payment.TypeXorPay} {
				types, limits := "wxpay", `{"wxpay":{"singleMin":1,"singleMax":100}}`
				if key == payment.TypeXorPay {
					types, limits = "xorpay,wxpay", `{"wxpay":{"singleMin":5,"singleMax":50}}`
				}
				inst, err := client.PaymentProviderInstance.Create().SetProviderKey(key).SetName(key).SetConfig("{}").SetSupportedTypes(types).SetLimits(limits).SetEnabled(true).SetSortOrder(i).Save(ctx)
				require.NoError(t, err)
				instances = append(instances, inst)
			}
			svc := &PaymentConfigService{entClient: client, settingRepo: &paymentConfigSettingRepoStub{values: map[string]string{
				SettingPaymentVisibleMethodWxpaySource: source,
			}}}
			inner := &captureLoadBalancer{}
			_, err := newVisibleMethodLoadBalancer(inner, svc).SelectInstance(ctx, "", payment.TypeWxpay, payment.StrategyRoundRobin, 10)
			require.NoError(t, err)
			want := map[string]string{"xorpay_wxpay": payment.TypeXorPay, "official_wxpay": payment.TypeWxpay, "easypay_wxpay": payment.TypeEasyPay}[source]
			require.Equal(t, want, inner.lastProviderKey)
			require.Equal(t, payment.TypeWxpay, inner.lastPaymentType)
			limits, err := svc.GetAvailableMethodLimits(ctx)
			require.NoError(t, err)
			require.Contains(t, limits.Methods, payment.TypeWxpay)
			wantMax := 100.0
			if want == payment.TypeXorPay {
				wantMax = 50
			}
			require.Equal(t, wantMax, limits.Methods[payment.TypeWxpay].SingleMax)
			require.Contains(t, limits.Methods, payment.TypeXorPay, "existing Alipay remains available")
			available := buildVisibleMethodSourceAvailability(instances)
			require.True(t, available["xorpay_wxpay"])
			require.Contains(t, applyVisibleMethodRoutingToEnabledTypes([]string{payment.TypeXorPay}, map[string]string{
				SettingPaymentVisibleMethodWxpaySource: source, SettingPaymentVisibleMethodWxpayEnabled: "true",
			}, available), payment.TypeWxpay)
			paymentSvc := &PaymentService{configService: svc}
			require.Equal(t, want == payment.TypeWxpay, paymentSvc.usesOfficialWxpayVisibleMethod(ctx))
		})
	}
}

func TestXorPayAlipayOnlyDoesNotExposeWechat(t *testing.T) {
	inst := &dbent.PaymentProviderInstance{ProviderKey: payment.TypeXorPay, SupportedTypes: "xorpay", Enabled: true}
	require.False(t, providerSupportsVisibleMethod(inst, payment.TypeWxpay))
	require.False(t, buildVisibleMethodSourceAvailability([]*dbent.PaymentProviderInstance{inst})["xorpay_wxpay"])
	inst.SupportedTypes, inst.Enabled = "xorpay,wxpay", false
	require.False(t, providerSupportsVisibleMethod(inst, payment.TypeWxpay))
}

func TestXorPayWechatSingleInstanceNeedsNoLegacySourceSetting(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	_, err := client.PaymentProviderInstance.Create().SetProviderKey(payment.TypeXorPay).
		SetName("XorPay").SetConfig("{}").SetSupportedTypes("xorpay,wxpay").SetEnabled(true).Save(ctx)
	require.NoError(t, err)
	for _, source := range []string{"", "official_wxpay"} {
		svc := &PaymentConfigService{entClient: client, settingRepo: &paymentConfigSettingRepoStub{values: map[string]string{
			SettingPaymentVisibleMethodWxpaySource: source,
		}}}
		inner := &captureLoadBalancer{}
		_, err := newVisibleMethodLoadBalancer(inner, svc).SelectInstance(ctx, "", payment.TypeWxpay, payment.StrategyRoundRobin, 10)
		require.NoError(t, err)
		require.Equal(t, payment.TypeXorPay, inner.lastProviderKey)
		limits, err := svc.GetAvailableMethodLimits(ctx)
		require.NoError(t, err)
		require.Contains(t, limits.Methods, payment.TypeWxpay)
	}
}

func TestXorPayWechatNotificationFulfillment(t *testing.T) {
	ctx := context.Background()
	client := newPaymentConfigServiceTestClient(t)
	ensurePaymentAuditOrderActionUniqueIndex(t, ctx, client)
	inst, err := client.PaymentProviderInstance.Create().
		SetProviderKey(payment.TypeXorPay).SetName("XorPay").SetSupportedTypes("xorpay,wxpay").SetEnabled(true).
		SetConfig(encryptWebhookProviderConfig(t, map[string]string{
			"aid": "test-aid", "appSecret": "test-secret", "notifyUrl": "https://merchant.example/api/v1/payment/webhook/xorpay",
		})).Save(ctx)
	require.NoError(t, err)
	order := createPaymentFulfillmentSubscriptionOrder(t, ctx, client, OrderStatusPending, time.Now())
	sel := &payment.InstanceSelection{ProviderKey: payment.TypeXorPay, InstanceID: strconv.FormatInt(inst.ID, 10), SupportedTypes: "xorpay,wxpay"}
	order, err = client.PaymentOrder.UpdateOneID(order.ID).
		SetPaymentType(payment.TypeWxpay).SetProviderKey(payment.TypeXorPay).SetProviderInstanceID(sel.InstanceID).
		SetProviderSnapshot(buildPaymentOrderProviderSnapshot(sel, CreateOrderRequest{PaymentType: payment.TypeWxpay})).
		SetOrderType(payment.OrderTypeBalance).ClearPlanID().ClearSubscriptionGroupID().ClearSubscriptionDays().Save(ctx)
	require.NoError(t, err)
	redeemRepo := &paymentFulfillmentRedeemRepo{}
	credited := 0.0
	userRepo := &mockUserRepo{getByIDUser: &User{ID: order.UserID}}
	userRepo.updateBalanceFn = func(_ context.Context, id int64, amount float64) error {
		require.Equal(t, order.UserID, id)
		credited += amount
		return nil
	}
	svc := &PaymentService{
		entClient: client, loadBalancer: newWebhookProviderTestLoadBalancer(client), registry: payment.NewRegistry(),
		userRepo:      userRepo,
		redeemService: NewRedeemService(redeemRepo, userRepo, nil, &paymentFulfillmentRedeemCacheStub{}, nil, client, nil, nil),
	}
	// In-flight callbacks must use the pinned XorPay instance even after it is disabled.
	_, err = client.PaymentProviderInstance.UpdateOneID(inst.ID).SetEnabled(false).Save(ctx)
	require.NoError(t, err)
	prov, err := svc.GetWebhookProvider(ctx, payment.TypeXorPay, order.OutTradeNo)
	require.NoError(t, err)
	require.Equal(t, payment.TypeXorPay, prov.ProviderKey())
	verify := func(price string) *payment.PaymentNotification {
		form := url.Values{"aoid": {"wx-aoid"}, "order_id": {order.OutTradeNo}, "pay_price": {price}, "pay_time": {"2026-11-26 12:00:00"}}
		form.Set("sign", fmt.Sprintf("%x", md5.Sum([]byte("wx-aoid"+order.OutTradeNo+price+form.Get("pay_time")+"test-secret"))))
		n, verifyErr := prov.VerifyNotification(ctx, form.Encode(), nil)
		require.NoError(t, verifyErr)
		return n
	}
	require.ErrorContains(t, svc.HandlePaymentNotification(ctx, verify("1.00"), prov.ProviderKey()), "amount mismatch")
	unchanged, err := client.PaymentOrder.Get(ctx, order.ID)
	require.NoError(t, err)
	require.Equal(t, OrderStatusPending, unchanged.Status)
	require.Zero(t, credited)
	n := verify(fmt.Sprintf("%.2f", order.PayAmount))
	require.NoError(t, svc.HandlePaymentNotification(ctx, n, prov.ProviderKey()))
	require.NoError(t, svc.HandlePaymentNotification(ctx, n, prov.ProviderKey()))
	require.Equal(t, order.Amount, credited)
	require.Len(t, redeemRepo.useCalls, 1)
	completed, err := client.PaymentOrder.Get(ctx, order.ID)
	require.NoError(t, err)
	require.Equal(t, OrderStatusCompleted, completed.Status)
	require.Equal(t, "wx-aoid", completed.PaymentTradeNo)
}

func TestXorPayWechatSkipsOfficialOAuth(t *testing.T) {
	svc := &PaymentService{}
	req := CreateOrderRequest{PaymentType: payment.TypeWxpay, IsMobile: true, IsWeChatBrowser: true}
	sel := &payment.InstanceSelection{ProviderKey: payment.TypeXorPay}
	require.NoError(t, svc.validateSelectedCreateOrderInstance(context.Background(), req, sel))
	result, err := svc.maybeBuildWeChatOAuthRequiredResponseForSelection(context.Background(), req, 10, 10, 0, sel)
	require.NoError(t, err)
	require.Nil(t, result)
}

func TestXorPayRegistrationPreservesLegacyWechatProvider(t *testing.T) {
	for _, xorpayFirst := range []bool{false, true} {
		t.Run(fmt.Sprintf("xorpay_first_%t", xorpayFirst), func(t *testing.T) {
			ctx := context.Background()
			client := newPaymentConfigServiceTestClient(t)
			keys := []string{payment.TypeWxpay, payment.TypeXorPay}
			if xorpayFirst {
				keys[0], keys[1] = keys[1], keys[0]
			}
			for _, key := range keys {
				config := ""
				if key == payment.TypeWxpay {
					config = encryptValidWebhookWxpayConfig(t, "legacy")
				} else {
					config = encryptWebhookProviderConfig(t, map[string]string{"aid": "test-aid", "appSecret": "test-secret", "notifyUrl": "https://merchant.example/notify"})
				}
				_, err := client.PaymentProviderInstance.Create().SetProviderKey(key).SetName(key).SetSupportedTypes("wxpay").SetConfig(config).SetEnabled(true).Save(ctx)
				require.NoError(t, err)
			}
			svc := &PaymentService{entClient: client, loadBalancer: newWebhookProviderTestLoadBalancer(client), registry: payment.NewRegistry()}
			svc.RefreshProviders(ctx)
			require.Equal(t, payment.TypeWxpay, svc.registry.GetProviderKey(payment.TypeWxpay))
			require.Equal(t, payment.TypeXorPay, svc.registry.GetProviderKey(payment.TypeXorPay))
			order := createPaymentFulfillmentSubscriptionOrder(t, ctx, client, OrderStatusCompleted, time.Now())
			order, err := client.PaymentOrder.UpdateOneID(order.ID).SetPaymentType(payment.TypeWxpay).Save(ctx)
			require.NoError(t, err)
			require.Nil(t, order.ProviderInstanceID)
			require.Nil(t, order.ProviderKey)
			require.Empty(t, order.ProviderSnapshot)
			prov, err := svc.getOrderProvider(ctx, order)
			require.NoError(t, err)
			require.Equal(t, payment.TypeWxpay, prov.ProviderKey())
			require.NoError(t, svc.HandlePaymentNotification(ctx, &payment.PaymentNotification{
				OrderID: order.OutTradeNo, TradeNo: order.PaymentTradeNo, Amount: order.PayAmount, Status: payment.NotificationStatusSuccess,
			}, payment.TypeWxpay))
		})
	}
}
