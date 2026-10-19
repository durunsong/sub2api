package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type redeemEmailHandlerSvcStub struct {
	createIn     service.CreateRedeemEmailBatchInput
	replayed     bool
	createErr    error
	searchFilter service.RedeemEmailRecipientFilter
	previewIn    service.RedeemEmailPreviewInput
	deleteErr    error
}

func (s *redeemEmailHandlerSvcStub) Defaults(context.Context) service.RedeemEmailDefaults {
	return service.RedeemEmailDefaults{TemplateEvent: service.NotificationEmailEventRedeemInactiveCoupon}
}
func (s *redeemEmailHandlerSvcStub) SearchRecipients(_ context.Context, f service.RedeemEmailRecipientFilter) ([]service.RedeemEmailRecipient, error) {
	s.searchFilter = f
	return []service.RedeemEmailRecipient{{UserID: 1, Email: "a@example.com", Segment: "never_used"}}, nil
}
func (s *redeemEmailHandlerSvcStub) Preview(_ context.Context, in service.RedeemEmailPreviewInput) (service.NotificationEmailPreview, error) {
	s.previewIn = in
	return service.NotificationEmailPreview{Subject: "[Demo] coupon", HTML: "<p>ok</p>"}, nil
}
func (s *redeemEmailHandlerSvcStub) CreateBatch(_ context.Context, in service.CreateRedeemEmailBatchInput) (*service.RedeemEmailBatch, bool, error) {
	s.createIn = in
	if s.createErr != nil {
		return nil, false, s.createErr
	}
	return &service.RedeemEmailBatch{ID: 3, Status: "pending", TotalCount: len(in.UserIDs)}, s.replayed, nil
}
func (s *redeemEmailHandlerSvcStub) ListBatches(context.Context, pagination.PaginationParams) ([]service.RedeemEmailBatch, *pagination.PaginationResult, error) {
	return []service.RedeemEmailBatch{}, &pagination.PaginationResult{}, nil
}
func (s *redeemEmailHandlerSvcStub) GetBatch(context.Context, int64) (*service.RedeemEmailBatch, error) {
	return nil, service.ErrRedeemEmailBatchNotFound
}
func (s *redeemEmailHandlerSvcStub) ListItems(context.Context, int64, string, pagination.PaginationParams) ([]service.RedeemEmailItem, *pagination.PaginationResult, error) {
	return []service.RedeemEmailItem{}, &pagination.PaginationResult{}, nil
}
func (s *redeemEmailHandlerSvcStub) RetryFailed(context.Context, int64) (*service.RedeemEmailBatch, error) {
	return &service.RedeemEmailBatch{ID: 3}, nil
}
func (s *redeemEmailHandlerSvcStub) DeleteBatch(context.Context, int64) error { return s.deleteErr }

func newRedeemEmailTestRouter(svc *redeemEmailHandlerSvcStub) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := &RedeemEmailHandler{svc: svc}
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(string(middleware2.ContextKeyUser), middleware2.AuthSubject{UserID: 42})
		c.Next()
	})
	g := r.Group("/email-batches")
	g.GET("/defaults", h.Defaults)
	g.POST("/recipients/search", h.SearchRecipients)
	g.POST("/preview", h.Preview)
	g.POST("", h.Create)
	g.GET("/:batch_id", h.Get)
	g.DELETE("/:batch_id", h.Delete)
	return r
}

func doRedeemEmailRequest(r *gin.Engine, method, path string, body any, headers map[string]string) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func validRedeemEmailCreateBody() map[string]any {
	return map[string]any{
		"segments":             []string{"inactive_2_3m", "never_used"},
		"user_ids":             []int64{5, 6},
		"coupon_value":         3.5,
		"code_expires_in_days": 30,
	}
}

func TestRedeemEmailHandlerCreatePassesActorKeyAndSegments(t *testing.T) {
	svc := &redeemEmailHandlerSvcStub{replayed: true}
	w := doRedeemEmailRequest(newRedeemEmailTestRouter(svc), http.MethodPost, "/email-batches", validRedeemEmailCreateBody(),
		map[string]string{"Idempotency-Key": "  key-12345678  "})

	require.Equal(t, http.StatusAccepted, w.Code, w.Body.String())
	require.Equal(t, "true", w.Header().Get("Idempotency-Replayed"))
	require.Equal(t, int64(42), svc.createIn.CreatedBy)
	require.Equal(t, "key-12345678", svc.createIn.IdempotencyKey)
	require.Equal(t, []string{"inactive_2_3m", "never_used"}, svc.createIn.Filter.Segments)
	require.Equal(t, 30, *svc.createIn.CodeExpiresInDays)
}

func TestRedeemEmailHandlerCreateRequiresExplicitExpiryChoice(t *testing.T) {
	svc := &redeemEmailHandlerSvcStub{}
	body := validRedeemEmailCreateBody()
	delete(body, "code_expires_in_days")
	w := doRedeemEmailRequest(newRedeemEmailTestRouter(svc), http.MethodPost, "/email-batches", body, map[string]string{"Idempotency-Key": "key-12345678"})
	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Zero(t, svc.createIn.CreatedBy, "service must not be called")

	body["never_expires"] = true
	w = doRedeemEmailRequest(newRedeemEmailTestRouter(svc), http.MethodPost, "/email-batches", body, map[string]string{"Idempotency-Key": "key-12345678"})
	require.Equal(t, http.StatusAccepted, w.Code, w.Body.String())
	require.Nil(t, svc.createIn.CodeExpiresInDays)
}

func TestRedeemEmailHandlerRejectsInvalidSegmentsAndMapsServiceErrors(t *testing.T) {
	for _, segments := range []any{[]string{}, []string{"inactive_6m"}, nil} {
		body := validRedeemEmailCreateBody()
		body["segments"] = segments
		svc := &redeemEmailHandlerSvcStub{}
		w := doRedeemEmailRequest(newRedeemEmailTestRouter(svc), http.MethodPost, "/email-batches", body, nil)
		require.Equal(t, http.StatusBadRequest, w.Code, "segments=%v", segments)
		require.Zero(t, svc.createIn.CreatedBy)
	}

	svc := &redeemEmailHandlerSvcStub{createErr: service.ErrRedeemEmailBatchConflict}
	w := doRedeemEmailRequest(newRedeemEmailTestRouter(svc), http.MethodPost, "/email-batches", validRedeemEmailCreateBody(), map[string]string{"Idempotency-Key": "key-12345678"})
	require.Equal(t, http.StatusConflict, w.Code)
}

func TestRedeemEmailHandlerSearchPassesSegments(t *testing.T) {
	svc := &redeemEmailHandlerSvcStub{}
	r := newRedeemEmailTestRouter(svc)
	w := doRedeemEmailRequest(r, http.MethodPost, "/email-batches/recipients/search",
		map[string]any{"segments": []string{"inactive_3m_plus", "never_used"}, "search": "bob"}, nil)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	require.Equal(t, []string{"inactive_3m_plus", "never_used"}, svc.searchFilter.Segments)
	require.Equal(t, "bob", svc.searchFilter.Search)
	require.Contains(t, w.Body.String(), `"segment":"never_used"`)

	w = doRedeemEmailRequest(r, http.MethodPost, "/email-batches/recipients/search", map[string]any{"segments": []string{"bogus"}}, nil)
	require.Equal(t, http.StatusBadRequest, w.Code)

	w = doRedeemEmailRequest(r, http.MethodPost, "/email-batches/recipients/search", map[string]any{"segments": []string{"zero_balance"}}, nil)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	require.Equal(t, []string{"zero_balance"}, svc.searchFilter.Segments)

	w = doRedeemEmailRequest(r, http.MethodPost, "/email-batches/recipients/search", map[string]any{"segments": []string{"zero_balance", "never_used"}}, nil)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	require.Equal(t, []string{"zero_balance", "never_used"}, svc.searchFilter.Segments)
}

func TestRedeemEmailHandlerPreviewPassesLocaleAndCoupon(t *testing.T) {
	svc := &redeemEmailHandlerSvcStub{}
	r := newRedeemEmailTestRouter(svc)
	w := doRedeemEmailRequest(r, http.MethodPost, "/email-batches/preview", map[string]any{
		"coupon_value": 2, "code_expires_in_days": 7, "locale": "en",
	}, nil)

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	require.Equal(t, "en", svc.previewIn.Locale)
	require.InDelta(t, 2.0, svc.previewIn.CouponValue, 1e-9)
	require.Equal(t, 7, *svc.previewIn.CodeExpiresInDays)
	require.Contains(t, w.Body.String(), "[Demo] coupon")

	w = doRedeemEmailRequest(r, http.MethodPost, "/email-batches/preview", map[string]any{"coupon_value": 2, "never_expires": true, "locale": "fr"}, nil)
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRedeemEmailHandlerBatchErrors(t *testing.T) {
	r := newRedeemEmailTestRouter(&redeemEmailHandlerSvcStub{deleteErr: service.ErrRedeemEmailBatchNotDeletable})
	require.Equal(t, http.StatusBadRequest, doRedeemEmailRequest(r, http.MethodGet, "/email-batches/abc", nil, nil).Code)
	require.Equal(t, http.StatusNotFound, doRedeemEmailRequest(r, http.MethodGet, "/email-batches/9", nil, nil).Code)
	require.Equal(t, http.StatusConflict, doRedeemEmailRequest(r, http.MethodDelete, "/email-batches/9", nil, nil).Code)
}
