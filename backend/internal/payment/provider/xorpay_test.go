package provider

import (
	"context"
	"crypto/md5"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/stretchr/testify/require"
)

func TestXorPayCreatePayment(t *testing.T) {
	for _, tc := range []struct{ method, upstream, qr string }{
		{payment.TypeXorPay, "alipay", "https://qr.alipay.com/test"},
		{payment.TypeWxpay, "native", "weixin://wxpay/bizpayurl?pr=test"},
	} {
		t.Run(tc.method, func(t *testing.T) {
			var form url.Values
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodPost, r.Method)
				require.Equal(t, "/api/pay/test-aid", r.URL.Path)
				require.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))
				require.NoError(t, r.ParseForm())
				form = r.PostForm
				fmt.Fprintf(w, `{"status":"ok","aoid":"test-aoid","info":{"qr":%q}}`, tc.qr)
			}))
			defer server.Close()
			x, err := NewXorPay("1", map[string]string{"aid": "test-aid", "appSecret": "test-secret", "notifyUrl": "https://merchant.example/api/v1/payment/webhook/xorpay", "apiBase": server.URL})
			require.NoError(t, err)
			require.Contains(t, x.SupportedTypes(), tc.method)
			result, err := x.CreatePayment(context.Background(), payment.CreatePaymentRequest{
				OrderID: "order-123", Amount: "50", PaymentType: tc.method, Subject: "算力充值 & 订阅", IsMobile: true,
			})
			require.NoError(t, err)
			require.Equal(t, tc.upstream, form.Get("pay_type"))
			require.Equal(t, "50.00", form.Get("price"))
			require.Equal(t, "order-123", form.Get("order_id"))
			wantSign := fmt.Sprintf("%x", md5.Sum([]byte("算力充值 & 订阅"+tc.upstream+"50.00order-123https://merchant.example/api/v1/payment/webhook/xorpaytest-secret")))
			require.Equal(t, wantSign, form.Get("sign"))
			require.Empty(t, form.Get("appSecret"))
			require.Equal(t, "test-aoid", result.TradeNo)
			require.Equal(t, tc.qr, result.QRCode)
			require.Empty(t, result.PayURL)
			require.Nil(t, result.OAuth)
			require.Nil(t, result.JSAPI)
		})
	}
}

func TestXorPayRejectsUnsupportedPaymentType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("unsupported payment type must not reach XorPay")
		fmt.Fprint(w, `{"status":"ok","info":{"qr":"unexpected"}}`)
	}))
	defer server.Close()
	x, err := NewXorPay("1", map[string]string{"aid": "test-aid", "appSecret": "test-secret", "notifyUrl": "https://merchant.example/notify", "apiBase": server.URL})
	require.NoError(t, err)
	_, err = x.CreatePayment(context.Background(), payment.CreatePaymentRequest{PaymentType: payment.TypeStripe, Amount: "50.00"})
	require.ErrorContains(t, err, "unsupported payment type")
}

func TestXorPayCreateErrors(t *testing.T) {
	for _, tc := range []struct{ name, body, want string }{
		{"fee", `{"status":"fee_error","info":"insufficient fee"}`, "status=fee_error"},
		{"wechat", `{"status":"wechat_api_error","info":{"code":"DENIED"}}`, "status=wechat_api_error"},
		{"empty qr", `{"status":"ok","info":{}}`, "empty qr"},
		{"invalid response", `invalid`, "parse"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, tc.body) }))
			defer server.Close()
			x, err := NewXorPay("1", map[string]string{"aid": "test-aid", "appSecret": "test-secret", "notifyUrl": "https://merchant.example/notify", "apiBase": server.URL})
			require.NoError(t, err)
			_, err = x.CreatePayment(context.Background(), payment.CreatePaymentRequest{PaymentType: payment.TypeWxpay, Amount: "50.00"})
			require.ErrorContains(t, err, tc.want)
		})
	}
}

func TestXorPayVerifyNotification(t *testing.T) {
	x, err := NewXorPay("1", map[string]string{"aid": "test-aid", "appSecret": "test-secret", "notifyUrl": "https://merchant.example/notify"})
	require.NoError(t, err)
	form := url.Values{"aoid": {"test-aoid"}, "order_id": {"order-123"}, "pay_price": {"50.00"}, "pay_time": {"2026-11-26 12:00:00"}}
	form.Set("sign", fmt.Sprintf("%x", md5.Sum([]byte("test-aoidorder-12350.002026-11-26 12:00:00test-secret"))))
	n, err := x.VerifyNotification(context.Background(), form.Encode(), nil)
	require.NoError(t, err)
	require.Equal(t, "order-123", n.OrderID)
	require.Equal(t, "test-aoid", n.TradeNo)
	require.Equal(t, 50.0, n.Amount)
	require.Equal(t, payment.NotificationStatusSuccess, n.Status)
	require.Equal(t, map[string]string{"aid": "test-aid"}, n.Metadata)
	form.Set("pay_price", "500.00")
	_, err = x.VerifyNotification(context.Background(), form.Encode(), nil)
	require.ErrorContains(t, err, "invalid signature")
}

func TestXorPayQueryAndRefund(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/query2/test-aid":
			require.Equal(t, http.MethodGet, r.Method)
			require.Equal(t, "order-123", r.URL.Query().Get("order_id"))
			require.Equal(t, fmt.Sprintf("%x", md5.Sum([]byte("order-123test-secret"))), r.URL.Query().Get("sign"))
			fmt.Fprint(w, `{"status":"success"}`)
		case "/api/refund/test-aoid":
			require.Equal(t, http.MethodPost, r.Method)
			require.NoError(t, r.ParseForm())
			require.Equal(t, "5.00", r.PostForm.Get("price"))
			require.Equal(t, fmt.Sprintf("%x", md5.Sum([]byte("5.00test-secret"))), r.PostForm.Get("sign"))
			fmt.Fprint(w, `{"status":"ok"}`)
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	x, err := NewXorPay("1", map[string]string{"aid": "test-aid", "appSecret": "test-secret", "notifyUrl": "https://merchant.example/notify", "apiBase": server.URL})
	require.NoError(t, err)
	query, err := x.QueryOrder(context.Background(), "order-123")
	require.NoError(t, err)
	require.Equal(t, payment.ProviderStatusPaid, query.Status)
	require.Zero(t, query.Amount, "query must not fabricate a paid amount")
	refund, err := x.Refund(context.Background(), payment.RefundRequest{TradeNo: "test-aoid", OrderID: "order-123", Amount: "5"})
	require.NoError(t, err)
	require.Equal(t, payment.ProviderStatusSuccess, refund.Status)
}
