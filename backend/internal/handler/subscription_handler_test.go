package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type resetDailyHandlerRepo struct {
	service.UserSubscriptionRepository
	subscription *service.UserSubscription
	getCalls     int
	reloadErr    error
}

func (r *resetDailyHandlerRepo) GetByID(_ context.Context, id int64) (*service.UserSubscription, error) {
	r.getCalls++
	if r.getCalls > 1 && r.reloadErr != nil {
		return nil, r.reloadErr
	}
	if r.subscription == nil || r.subscription.ID != id {
		return nil, service.ErrSubscriptionNotFound
	}
	copy := *r.subscription
	return &copy, nil
}

func (r *resetDailyHandlerRepo) ResetDailyQuota(_ context.Context, request service.ManualDailyResetRequest) (service.ManualDailyResetResult, error) {
	if r.subscription.ManualResetCredits <= 0 {
		return service.ManualDailyResetResult{CreditsBefore: -1, CreditsAfter: -1}, service.ErrManualResetConflict
	}
	before := r.subscription.ManualResetCredits
	r.subscription.ManualResetCredits--
	r.subscription.DailyUsageUSD = 0
	r.subscription.DailyUsageTokens = 0
	return service.ManualDailyResetResult{MutationApplied: true, CreditsBefore: before, CreditsAfter: before - 1}, nil
}

func newResetDailyHandlerContext(userID int64) (*gin.Context, *httptest.ResponseRecorder) {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/subscriptions/42/reset-daily", nil)
	c.Params = gin.Params{{Key: "id", Value: "42"}}
	c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: userID})
	return c, recorder
}

type resetDailyHTTPEnvelope struct {
	Code     int               `json:"code"`
	Reason   string            `json:"reason"`
	Metadata map[string]string `json:"metadata"`
}

func performResetDailyHTTP(t *testing.T, repo *resetDailyHandlerRepo, userID int64) (int, resetDailyHTTPEnvelope, map[string]any) {
	t.Helper()
	handler := NewSubscriptionHandler(service.NewSubscriptionService(nil, repo, nil, nil, nil))
	c, recorder := newResetDailyHandlerContext(userID)
	handler.ResetDaily(c)
	var body resetDailyHTTPEnvelope
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	extra, ok := c.Get("audit_extra")
	require.True(t, ok)
	auditExtra, ok := extra.(map[string]any)
	require.True(t, ok)
	return recorder.Code, body, auditExtra
}

func TestSubscriptionHandlerResetDaily_HidesCrossUserSubscriptionExistence(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Now().UTC()
	foreign := &resetDailyHandlerRepo{subscription: &service.UserSubscription{
		ID: 42, UserID: 999, Status: service.SubscriptionStatusActive,
		StartsAt: now.Add(-time.Hour), ExpiresAt: now.Add(time.Hour), ManualResetCredits: 37,
	}}
	missing := &resetDailyHandlerRepo{}

	foreignStatus, foreignBody, foreignAudit := performResetDailyHTTP(t, foreign, 7)
	missingStatus, missingBody, missingAudit := performResetDailyHTTP(t, missing, 7)

	require.Equal(t, missingStatus, foreignStatus)
	require.Equal(t, missingBody.Reason, foreignBody.Reason)
	require.Equal(t, missingBody.Metadata, foreignBody.Metadata)
	require.Equal(t, map[string]string{"credits_before": "-1", "credits_after": "-1"}, foreignBody.Metadata)
	require.Equal(t, missingAudit["credits_before"], foreignAudit["credits_before"])
	require.Equal(t, missingAudit["credits_after"], foreignAudit["credits_after"])
	require.Equal(t, -1, foreignAudit["credits_before"])
	require.Equal(t, -1, foreignAudit["credits_after"])
}

func TestSubscriptionHandlerResetDaily_HidesSoftDeletedSubscriptionExistence(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Now().UTC()
	deletedAt := now.Add(-time.Minute)
	deleted := &resetDailyHandlerRepo{subscription: &service.UserSubscription{
		ID: 42, UserID: 7, Status: service.SubscriptionStatusActive,
		StartsAt: now.Add(-time.Hour), ExpiresAt: now.Add(time.Hour), ManualResetCredits: 51, DeletedAt: &deletedAt,
	}}
	missing := &resetDailyHandlerRepo{}

	deletedStatus, deletedBody, deletedAudit := performResetDailyHTTP(t, deleted, 7)
	missingStatus, missingBody, missingAudit := performResetDailyHTTP(t, missing, 7)

	require.Equal(t, missingStatus, deletedStatus)
	require.Equal(t, missingBody.Reason, deletedBody.Reason)
	require.Equal(t, missingBody.Metadata, deletedBody.Metadata)
	require.Equal(t, missingAudit, deletedAudit)
}
func TestSubscriptionHandlerResetDailyReturnsZeroCreditAndSuccessAuditFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Now().UTC()
	repo := &resetDailyHandlerRepo{subscription: &service.UserSubscription{ID: 42, UserID: 7, GroupID: 3, Status: service.SubscriptionStatusActive, StartsAt: now.Add(-time.Hour), ExpiresAt: now.Add(time.Hour), ManualResetCredits: 1}}
	handler := NewSubscriptionHandler(service.NewSubscriptionService(nil, repo, nil, nil, nil))
	c, recorder := newResetDailyHandlerContext(7)
	handler.ResetDaily(c)
	require.Equal(t, http.StatusOK, recorder.Code)
	var body struct {
		Data struct {
			ManualResetCredits int `json:"manual_reset_credits"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	require.Zero(t, body.Data.ManualResetCredits)
	require.Equal(t, "user.subscription.daily_reset", c.GetString("audit_action"))
	require.Equal(t, map[string]any{"user_id": int64(7), "subscription_id": int64(42), "credits_before": 1, "credits_after": 0, "result": "success", "error_code": ""}, c.MustGet("audit_extra"))
}

func TestSubscriptionHandlerResetDailyAuditsSuccessWhenReloadFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Now().UTC()
	repo := &resetDailyHandlerRepo{
		subscription: &service.UserSubscription{ID: 42, UserID: 7, GroupID: 3, Status: service.SubscriptionStatusActive, StartsAt: now.Add(-time.Hour), ExpiresAt: now.Add(time.Hour), ManualResetCredits: 2},
		reloadErr:    errors.New("reload failed"),
	}
	handler := NewSubscriptionHandler(service.NewSubscriptionService(nil, repo, nil, nil, nil))
	c, recorder := newResetDailyHandlerContext(7)

	handler.ResetDaily(c)

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
	require.Equal(t, map[string]any{"user_id": int64(7), "subscription_id": int64(42), "credits_before": 2, "credits_after": 1, "result": "success", "error_code": "RESPONSE_RELOAD_FAILED"}, c.MustGet("audit_extra"))
}
func TestSubscriptionHandlerResetDailyWritesStableFailureAuditFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Now().UTC()
	repo := &resetDailyHandlerRepo{subscription: &service.UserSubscription{ID: 42, UserID: 7, GroupID: 3, Status: service.SubscriptionStatusActive, StartsAt: now.Add(-time.Hour), ExpiresAt: now.Add(time.Hour), ManualResetCredits: 0}}
	handler := NewSubscriptionHandler(service.NewSubscriptionService(nil, repo, nil, nil, nil))
	c, recorder := newResetDailyHandlerContext(7)
	handler.ResetDaily(c)
	require.Equal(t, http.StatusConflict, recorder.Code)
	require.Equal(t, "user.subscription.daily_reset", c.GetString("audit_action"))
	require.Equal(t, map[string]any{"user_id": int64(7), "subscription_id": int64(42), "credits_before": 0, "credits_after": 0, "result": "failed", "error_code": "MANUAL_RESET_NO_CREDITS"}, c.MustGet("audit_extra"))
}
func TestSubscriptionHandlerResetDailyAuditsConcurrentNoCreditAsZeroToZero(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Now().UTC()
	repo := &resetDailyHandlerRepo{subscription: &service.UserSubscription{ID: 42, UserID: 7, GroupID: 3, Status: service.SubscriptionStatusActive, StartsAt: now.Add(-time.Hour), ExpiresAt: now.Add(time.Hour), ManualResetCredits: 0}}
	handler := NewSubscriptionHandler(service.NewSubscriptionService(nil, repo, nil, nil, nil))
	c, recorder := newResetDailyHandlerContext(7)

	handler.ResetDaily(c)

	require.Equal(t, http.StatusConflict, recorder.Code)
	require.Equal(t, map[string]any{"user_id": int64(7), "subscription_id": int64(42), "credits_before": 0, "credits_after": 0, "result": "failed", "error_code": "MANUAL_RESET_NO_CREDITS"}, c.MustGet("audit_extra"))
}
func TestSubscriptionHandlerResetDailyAuditsInvalidIDWithUnknownValues(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewSubscriptionHandler(nil)
	c, recorder := newResetDailyHandlerContext(7)
	c.Params[0].Value = "invalid"

	handler.ResetDaily(c)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	require.Equal(t, map[string]any{
		"user_id": int64(7), "subscription_id": int64(-1),
		"credits_before": -1, "credits_after": -1,
		"result": "failed", "error_code": "INVALID_SUBSCRIPTION_ID",
	}, c.MustGet("audit_extra"))
}
