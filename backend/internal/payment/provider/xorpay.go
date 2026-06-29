package provider

import (
	"context"
	"crypto/hmac"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/shopspring/decimal"
)

// XorPay constants.
const (
	xorpayDefaultAPIBase  = "https://xorpay.com"
	xorpayHTTPTimeout     = 15 * time.Second
	xorpayMaxResponseSize = 1 << 20 // 1MB
	xorpayMaxErrorSummary = 512
	xorpayStatusOK        = "ok"
	xorpayPayTypeAlipay   = "alipay"

	// Order query (GET /api/query2/{aid}) status values.
	xorpayQueryStatusNotExist = "not_exist"
	xorpayQueryStatusNew      = "new"
	xorpayQueryStatusPayed    = "payed"
	xorpayQueryStatusFeeError = "fee_error"
	xorpayQueryStatusSuccess  = "success"
	xorpayQueryStatusExpire   = "expire"
)

// XorPay implements payment.Provider for the XorPay aggregation platform.
// It currently supports Alipay face-to-face (扫码) payments: the create call
// returns an Alipay QR URL that the frontend renders for the user to scan.
type XorPay struct {
	instanceID string
	config     map[string]string // aid, appSecret, notifyUrl, returnUrl, apiBase
	httpClient *http.Client
}

// NewXorPay creates a new XorPay provider instance.
// Required config keys: aid, appSecret, notifyUrl. Optional: apiBase (defaults
// to https://xorpay.com), returnUrl (XorPay does not use it; kept for parity).
func NewXorPay(instanceID string, config map[string]string) (*XorPay, error) {
	for _, k := range []string{"aid", "appSecret", "notifyUrl"} {
		if strings.TrimSpace(config[k]) == "" {
			return nil, fmt.Errorf("xorpay config missing required key: %s", k)
		}
	}
	cfg := cloneStringMap(config)
	cfg["apiBase"] = normalizeXorPayAPIBase(cfg["apiBase"])
	return &XorPay{
		instanceID: instanceID,
		config:     cfg,
		httpClient: &http.Client{Timeout: xorpayHTTPTimeout},
	}, nil
}

func normalizeXorPayAPIBase(apiBase string) string {
	base := strings.TrimSpace(apiBase)
	if base == "" {
		return xorpayDefaultAPIBase
	}
	return strings.TrimRight(base, "/")
}

func (x *XorPay) apiBase() string {
	if x == nil {
		return xorpayDefaultAPIBase
	}
	return normalizeXorPayAPIBase(x.config["apiBase"])
}

func (x *XorPay) Name() string        { return "XorPay" }
func (x *XorPay) ProviderKey() string { return payment.TypeXorPay }
func (x *XorPay) SupportedTypes() []payment.PaymentType {
	return []payment.PaymentType{payment.TypeXorPay}
}

// MerchantIdentityMetadata exposes the non-sensitive aid for snapshot/audit use.
// The app secret is never included.
func (x *XorPay) MerchantIdentityMetadata() map[string]string {
	if x == nil {
		return nil
	}
	aid := strings.TrimSpace(x.config["aid"])
	if aid == "" {
		return nil
	}
	return map[string]string{"aid": aid}
}

// CreatePayment creates an Alipay scan-to-pay order via XorPay and returns the
// Alipay QR URL for the frontend to render.
//
// Endpoint: POST {apiBase}/api/pay/{aid}
// Sign:     MD5(name + pay_type + price + order_id + notify_url + app_secret)
func (x *XorPay) CreatePayment(ctx context.Context, req payment.CreatePaymentRequest) (*payment.CreatePaymentResponse, error) {
	notifyURL := strings.TrimSpace(x.config["notifyUrl"])
	if req.NotifyURL != "" {
		notifyURL = req.NotifyURL
	}
	if notifyURL == "" {
		return nil, fmt.Errorf("xorpay create: notify url is not configured")
	}

	// Normalize the price to two decimals so the value sent, the value signed,
	// and the pay_price returned in the callback all use the same format.
	price := strings.TrimSpace(req.Amount)
	if d, err := decimal.NewFromString(price); err == nil {
		price = d.StringFixed(2)
	}

	payType := xorpayPayTypeAlipay
	params := map[string]string{
		"name":       req.Subject,
		"pay_type":   payType,
		"price":      price,
		"order_id":   req.OrderID,
		"notify_url": notifyURL,
		"sign":       xorPayConcatSign(req.Subject, payType, price, req.OrderID, notifyURL, x.config["appSecret"]),
	}

	endpoint := x.apiBase() + "/api/pay/" + url.PathEscape(strings.TrimSpace(x.config["aid"]))
	body, status, err := x.postForm(ctx, endpoint, params)
	if err != nil {
		return nil, fmt.Errorf("xorpay create request: %w", err)
	}
	if status < http.StatusOK || status >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("xorpay create HTTP %d: %s", status, summarizeXorPayResponse(body))
	}

	var resp struct {
		Status string          `json:"status"`
		AOID   string          `json:"aoid"`
		Info   json.RawMessage `json:"info"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("xorpay create parse: %s", summarizeXorPayResponse(body))
	}
	if !strings.EqualFold(strings.TrimSpace(resp.Status), xorpayStatusOK) {
		return nil, fmt.Errorf("xorpay create failed: status=%s info=%s", resp.Status, summarizeXorPayInfo(resp.Info))
	}

	var info struct {
		QR string `json:"qr"`
	}
	if len(resp.Info) > 0 {
		_ = json.Unmarshal(resp.Info, &info)
	}
	if strings.TrimSpace(info.QR) == "" {
		return nil, fmt.Errorf("xorpay create: empty qr in response")
	}

	return &payment.CreatePaymentResponse{
		TradeNo: strings.TrimSpace(resp.AOID),
		QRCode:  strings.TrimSpace(info.QR),
	}, nil
}

// QueryOrder queries the order status by our order_id (the out_trade_no we sent).
//
// Endpoint: GET {apiBase}/api/query2/{aid}?order_id=...&sign=...
// Sign:     MD5(order_id + app_secret)
//
// Note: XorPay's query API only returns a status string (no amount). The webhook
// remains the authoritative fulfillment path because it carries pay_price.
func (x *XorPay) QueryOrder(ctx context.Context, tradeNo string) (*payment.QueryOrderResponse, error) {
	orderID := strings.TrimSpace(tradeNo)
	if orderID == "" {
		return nil, fmt.Errorf("xorpay query: missing order id")
	}
	q := url.Values{}
	q.Set("order_id", orderID)
	q.Set("sign", xorPayConcatSign(orderID, x.config["appSecret"]))
	endpoint := x.apiBase() + "/api/query2/" + url.PathEscape(strings.TrimSpace(x.config["aid"])) + "?" + q.Encode()

	body, status, err := x.get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("xorpay query request: %w", err)
	}
	if status < http.StatusOK || status >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("xorpay query HTTP %d: %s", status, summarizeXorPayResponse(body))
	}
	var resp struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("xorpay query parse: %s", summarizeXorPayResponse(body))
	}

	return &payment.QueryOrderResponse{
		Status:   xorPayQueryProviderStatus(resp.Status),
		Metadata: x.MerchantIdentityMetadata(),
	}, nil
}

// VerifyNotification parses and verifies an async payment notification.
//
// Body: application/x-www-form-urlencoded with aoid, order_id, pay_price,
// pay_time, more, detail, sign.
// Sign: MD5(aoid + order_id + pay_price + pay_time + app_secret)
func (x *XorPay) VerifyNotification(_ context.Context, rawBody string, _ map[string]string) (*payment.PaymentNotification, error) {
	values, err := url.ParseQuery(rawBody)
	if err != nil {
		return nil, fmt.Errorf("xorpay parse notify: %w", err)
	}
	aoid := values.Get("aoid")
	orderID := values.Get("order_id")
	payPrice := values.Get("pay_price")
	payTime := values.Get("pay_time")
	sign := strings.TrimSpace(values.Get("sign"))
	if sign == "" {
		return nil, fmt.Errorf("xorpay notify missing sign")
	}
	expected := xorPayConcatSign(aoid, orderID, payPrice, payTime, x.config["appSecret"])
	if !hmac.Equal([]byte(strings.ToLower(expected)), []byte(strings.ToLower(sign))) {
		return nil, fmt.Errorf("xorpay notify invalid signature: order_id=%s", orderID)
	}

	amount, err := strconv.ParseFloat(strings.TrimSpace(payPrice), 64)
	if err != nil {
		return nil, fmt.Errorf("xorpay notify invalid pay_price: order_id=%s", orderID)
	}

	// XorPay only sends a callback when the order is successfully paid.
	return &payment.PaymentNotification{
		TradeNo:  strings.TrimSpace(aoid),
		OrderID:  strings.TrimSpace(orderID),
		Amount:   amount,
		Status:   payment.NotificationStatusSuccess,
		RawData:  rawBody,
		Metadata: x.MerchantIdentityMetadata(),
	}, nil
}

// Refund requests a refund for a previously paid order.
//
// Endpoint: POST {apiBase}/api/refund/{aoid}
// Sign:     MD5(price + app_secret)
func (x *XorPay) Refund(ctx context.Context, req payment.RefundRequest) (*payment.RefundResponse, error) {
	aoid := strings.TrimSpace(req.TradeNo)
	if aoid == "" {
		return nil, fmt.Errorf("xorpay refund missing aoid (payment trade no)")
	}
	price := strings.TrimSpace(req.Amount)
	if d, err := decimal.NewFromString(price); err == nil {
		price = d.StringFixed(2)
	}
	params := map[string]string{
		"price": price,
		"sign":  xorPayConcatSign(price, x.config["appSecret"]),
	}
	endpoint := x.apiBase() + "/api/refund/" + url.PathEscape(aoid)
	body, status, err := x.postForm(ctx, endpoint, params)
	if err != nil {
		return nil, fmt.Errorf("xorpay refund request: %w", err)
	}
	if status < http.StatusOK || status >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("xorpay refund HTTP %d: %s", status, summarizeXorPayResponse(body))
	}
	var resp struct {
		Status string          `json:"status"`
		Info   json.RawMessage `json:"info"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("xorpay refund parse: %s", summarizeXorPayResponse(body))
	}
	if !strings.EqualFold(strings.TrimSpace(resp.Status), xorpayStatusOK) {
		return nil, fmt.Errorf("xorpay refund failed: status=%s info=%s", resp.Status, summarizeXorPayInfo(resp.Info))
	}
	refundID := aoid
	if req.OrderID != "" {
		refundID = aoid + "-" + req.OrderID
	}
	return &payment.RefundResponse{RefundID: refundID, Status: payment.ProviderStatusSuccess}, nil
}

func xorPayQueryProviderStatus(status string) string {
	switch strings.TrimSpace(strings.ToLower(status)) {
	case xorpayQueryStatusSuccess, xorpayQueryStatusPayed:
		return payment.ProviderStatusPaid
	case xorpayQueryStatusExpire:
		return payment.ProviderStatusFailed
	case xorpayQueryStatusNew, xorpayQueryStatusNotExist, xorpayQueryStatusFeeError:
		return payment.ProviderStatusPending
	default:
		return payment.ProviderStatusPending
	}
}

func (x *XorPay) postForm(ctx context.Context, endpoint string, params map[string]string) ([]byte, int, error) {
	form := url.Values{}
	for k, v := range params {
		form.Set(k, v)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return x.do(req)
}

func (x *XorPay) get(ctx context.Context, endpoint string) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, 0, err
	}
	return x.do(req)
}

func (x *XorPay) do(req *http.Request) ([]byte, int, error) {
	client := x.httpClient
	if client == nil {
		client = &http.Client{Timeout: xorpayHTTPTimeout}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(resp.Body, xorpayMaxResponseSize))
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return body, resp.StatusCode, nil
}

// xorPayConcatSign computes MD5 over the concatenation of the given values
// (pure value concatenation, no separators, no field names), per XorPay docs.
func xorPayConcatSign(parts ...string) string {
	hash := md5.Sum([]byte(strings.Join(parts, "")))
	return hex.EncodeToString(hash[:])
}

func summarizeXorPayResponse(body []byte) string {
	summary := strings.Join(strings.Fields(string(body)), " ")
	if summary == "" {
		return "<empty>"
	}
	if len(summary) > xorpayMaxErrorSummary {
		return summary[:xorpayMaxErrorSummary] + "..."
	}
	return summary
}

// summarizeXorPayInfo renders the polymorphic `info` field (string on error,
// object on success) for error messages without leaking secrets.
func summarizeXorPayInfo(raw json.RawMessage) string {
	if len(raw) == 0 {
		return "<empty>"
	}
	var asString string
	if err := json.Unmarshal(raw, &asString); err == nil {
		return summarizeXorPayResponse([]byte(asString))
	}
	return summarizeXorPayResponse(raw)
}

// Ensure interface compliance.
var (
	_ payment.Provider                 = (*XorPay)(nil)
	_ payment.MerchantIdentityProvider = (*XorPay)(nil)
)
