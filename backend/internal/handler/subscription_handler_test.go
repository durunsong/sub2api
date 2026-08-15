package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type resetCardAPIRepo struct {
	service.UserSubscriptionRepository
	subscriptions map[int64]*service.UserSubscription
	cards         []service.UserSubscriptionResetCard
	listCalls     int
}

func (r *resetCardAPIRepo) GetByID(_ context.Context, id int64) (*service.UserSubscription, error) {
	sub := r.subscriptions[id]
	if sub == nil {
		return nil, service.ErrSubscriptionNotFound
	}
	cp := *sub
	return &cp, nil
}
func (r *resetCardAPIRepo) ListByUserID(_ context.Context, userID int64) ([]service.UserSubscription, error) {
	out := []service.UserSubscription{}
	for _, sub := range r.subscriptions {
		if sub.UserID == userID {
			out = append(out, *sub)
		}
	}
	return out, nil
}
func (r *resetCardAPIRepo) GrantResetCard(context.Context, int64, int, service.PurchaseSource) (bool, error) {
	return false, nil
}
func (r *resetCardAPIRepo) LockResetCardSubscription(_ context.Context, subscriptionID, userID int64) error {
	sub := r.subscriptions[subscriptionID]
	if sub == nil || sub.UserID != userID || sub.DeletedAt != nil ||
		(sub.Status != service.SubscriptionStatusActive && sub.Status != service.SubscriptionStatusExpired) {
		return service.ErrResetCardNotFound
	}
	return nil
}
func (r *resetCardAPIRepo) ConsumeResetCard(_ context.Context, req service.ConsumeResetCardRequest) (service.ConsumeResetCardResult, error) {
	for i := range r.cards {
		card := &r.cards[i]
		if card.UserSubscriptionID == req.SubscriptionID && card.ValidityDays == req.ValidityDays && card.ConsumedAt == nil {
			card.ConsumedAt = &req.Now
			sub := r.subscriptions[req.SubscriptionID]
			expiresAt := req.Now.AddDate(0, 0, req.ValidityDays)
			if expiresAt.After(req.MaxExpiresAt) {
				expiresAt = req.MaxExpiresAt
			}
			sub.StartsAt, sub.ExpiresAt, sub.Status = req.Now, expiresAt, service.SubscriptionStatusActive
			sub.ManualResetCredits--
			return service.ConsumeResetCardResult{CardID: card.ID, ValidityDays: card.ValidityDays, ConsumedAt: req.Now, MutationApplied: true}, nil
		}
	}
	return service.ConsumeResetCardResult{}, service.ErrResetCardNotFound
}
func (r *resetCardAPIRepo) ListAvailableResetCardGroups(_ context.Context, ids []int64) ([]service.ResetCardGroup, error) {
	r.listCalls++
	wanted := map[int64]bool{}
	for _, id := range ids {
		wanted[id] = true
	}
	counts := map[[2]int64]int{}
	for _, card := range r.cards {
		if wanted[card.UserSubscriptionID] && card.ConsumedAt == nil {
			counts[[2]int64{card.UserSubscriptionID, int64(card.ValidityDays)}]++
		}
	}
	out := []service.ResetCardGroup{}
	for key, count := range counts {
		out = append(out, service.ResetCardGroup{SubscriptionID: key[0], ValidityDays: int(key[1]), AvailableCount: count})
	}
	return out, nil
}

func resetCardAPIContext(method, path string, body []byte, userID int64) (*gin.Context, *httptest.ResponseRecorder) {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(method, path, bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.Header.Set("Idempotency-Key", "reset-card-test-key")
	c.Params = gin.Params{{Key: "id", Value: "42"}}
	c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: userID})
	return c, rec
}
func resetCardTestSub(now time.Time, id, userID int64, credits int) *service.UserSubscription {
	return &service.UserSubscription{ID: id, UserID: userID, GroupID: 3, Status: service.SubscriptionStatusActive, StartsAt: now.Add(-time.Hour), ExpiresAt: now.Add(time.Hour), ManualResetCredits: credits}
}

func TestSubscriptionHandlerListAggregatesResetCardsInOneBatch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Now().UTC()
	repo := &resetCardAPIRepo{subscriptions: map[int64]*service.UserSubscription{1: resetCardTestSub(now, 1, 7, 3), 2: resetCardTestSub(now, 2, 7, 1)}, cards: []service.UserSubscriptionResetCard{{ID: 1, UserSubscriptionID: 1, ValidityDays: 7}, {ID: 2, UserSubscriptionID: 1, ValidityDays: 7}, {ID: 3, UserSubscriptionID: 1, ValidityDays: 30}, {ID: 4, UserSubscriptionID: 2, ValidityDays: 30}}}
	h := NewSubscriptionHandler(service.NewSubscriptionService(nil, repo, nil, nil, nil))
	c, rec := resetCardAPIContext(http.MethodGet, "/api/v1/subscriptions", nil, 7)
	h.List(c)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, 1, repo.listCalls)
	var body struct {
		Data []struct {
			ID         int64 `json:"id"`
			ResetCards struct {
				Total int `json:"total"`
			} `json:"reset_cards"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	totals := map[int64]int{}
	for _, sub := range body.Data {
		totals[sub.ID] = sub.ResetCards.Total
	}
	require.Equal(t, map[int64]int{1: 3, 2: 1}, totals)
}

func TestSubscriptionHandlerConsumeResetCardReturnsRemainingGroups(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Now().UTC()
	repo := &resetCardAPIRepo{subscriptions: map[int64]*service.UserSubscription{42: resetCardTestSub(now, 42, 7, 2)}, cards: []service.UserSubscriptionResetCard{{ID: 1, UserSubscriptionID: 42, ValidityDays: 30}, {ID: 2, UserSubscriptionID: 42, ValidityDays: 7}}}
	h := NewSubscriptionHandler(service.NewSubscriptionService(nil, repo, nil, nil, nil))
	c, rec := resetCardAPIContext(http.MethodPost, "/api/v1/subscriptions/42/reset-cards/consume", []byte(`{"validity_days":30}`), 7)
	h.ConsumeResetCard(c)
	require.Equal(t, http.StatusOK, rec.Code)
	var body struct {
		Data struct {
			ResetCards struct {
				Total int `json:"total"`
			} `json:"reset_cards"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, 1, body.Data.ResetCards.Total)
	require.Equal(t, "user.subscription.reset_card.consume", c.GetString("audit_action"))
	auditExtra := c.MustGet("audit_extra").(map[string]any)
	require.NotContains(t, auditExtra, "idempotency_key")
	require.NotContains(t, fmt.Sprint(auditExtra), "reset-card-test-key")
	require.Equal(t, map[string]any{"user_id": int64(7), "subscription_id": int64(42), "validity_days": 30, "result": "success", "error_code": "", "idempotency_key_hash": "34fa89a0977fb956"}, auditExtra)
}

func TestSubscriptionHandlerConsumeResetCardRequiresValidIdempotencyKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSubscriptionHandler(nil)
	for _, key := range []string{"", "bad key", string(make([]byte, 129))} {
		c, rec := resetCardAPIContext(http.MethodPost, "/api/v1/subscriptions/42/reset-cards/consume", []byte(`{"validity_days":30}`), 7)
		c.Request.Header.Set("Idempotency-Key", key)
		h.ConsumeResetCard(c)
		require.Equal(t, http.StatusBadRequest, rec.Code)
	}
}

func TestSubscriptionHandlerResetDailyRequiresIdempotencyKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSubscriptionHandler(nil)
	c, rec := resetCardAPIContext(http.MethodPost, "/api/v1/subscriptions/42/reset-daily", nil, 7)
	c.Request.Header.Del("Idempotency-Key")
	h.ResetDaily(c)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestSubscriptionHandlerConsumeResetCardRejectsInvalidDays(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewSubscriptionHandler(nil)
	for _, raw := range []string{`{"validity_days":0}`, `{"validity_days":-1}`, `{}`} {
		c, rec := resetCardAPIContext(http.MethodPost, "/api/v1/subscriptions/42/reset-cards/consume", []byte(raw), 7)
		h.ConsumeResetCard(c)
		require.Equal(t, http.StatusBadRequest, rec.Code)
	}
}

func TestSubscriptionHandlerConsumeResetCardHidesForeignAndMissingCard(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Now().UTC()
	repos := []*resetCardAPIRepo{
		{subscriptions: map[int64]*service.UserSubscription{42: resetCardTestSub(now, 42, 99, 1)}, cards: []service.UserSubscriptionResetCard{{ID: 1, UserSubscriptionID: 42, ValidityDays: 30}}},
		{subscriptions: map[int64]*service.UserSubscription{42: resetCardTestSub(now, 42, 7, 0)}},
	}
	reasons := []string{}
	for _, repo := range repos {
		h := NewSubscriptionHandler(service.NewSubscriptionService(nil, repo, nil, nil, nil))
		c, rec := resetCardAPIContext(http.MethodPost, "/api/v1/subscriptions/42/reset-cards/consume", []byte(`{"validity_days":30}`), 7)
		h.ConsumeResetCard(c)
		require.Equal(t, http.StatusConflict, rec.Code)
		var body struct {
			Reason string `json:"reason"`
		}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		reasons = append(reasons, body.Reason)
	}
	require.Equal(t, []string{"RESET_CARD_NOT_FOUND", "RESET_CARD_NOT_FOUND"}, reasons)
}

func TestSubscriptionHandlerResetDailyMapsToOneDayCard(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Now().UTC()
	repo := &resetCardAPIRepo{subscriptions: map[int64]*service.UserSubscription{42: resetCardTestSub(now, 42, 7, 1)}, cards: []service.UserSubscriptionResetCard{{ID: 1, UserSubscriptionID: 42, ValidityDays: 1}}}
	h := NewSubscriptionHandler(service.NewSubscriptionService(nil, repo, nil, nil, nil))
	c, rec := resetCardAPIContext(http.MethodPost, "/api/v1/subscriptions/42/reset-daily", nil, 7)
	h.ResetDaily(c)
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotNil(t, repo.cards[0].ConsumedAt)
	require.Equal(t, "user.subscription.reset_card.consume", c.GetString("audit_action"))
	require.Equal(t, 1, c.MustGet("audit_extra").(map[string]any)["validity_days"])
}
