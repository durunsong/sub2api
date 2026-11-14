//go:build unit

package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestAPIKeyAuthRejectsExhaustedNinetyDayTermAfterDaySixty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{RunMode: config.RunModeStandard}
	now := time.Now()
	start := now.Add(-61 * 24 * time.Hour)
	limit := 10000.0
	group := &service.Group{ID: 42, Name: "90-day", Status: service.StatusActive, Hydrated: true,
		SubscriptionType: service.SubscriptionTypeSubscription, MonthlyLimitUSD: &limit}
	user := &service.User{ID: 7, Role: service.RoleUser, Status: service.StatusActive, Concurrency: 3}
	key := &service.APIKey{ID: 100, Key: "sk-term-test", UserID: user.ID, GroupID: &group.ID,
		Status: service.StatusActive, User: user, Group: group}
	keySvc := service.NewAPIKeyService(&stubApiKeyRepo{getByKey: func(context.Context, string) (*service.APIKey, error) { return key, nil }}, nil, nil, nil, nil, nil, cfg)
	sub := &service.UserSubscription{ID: 55, UserID: user.ID, GroupID: group.ID, Status: service.SubscriptionStatusActive,
		StartsAt: start, ExpiresAt: start.Add(90 * 24 * time.Hour), DailyWindowStart: &now, WeeklyWindowStart: &now,
		MonthlyWindowStart: &start, MonthlyUsageUSD: 10000.01}
	getSub := func() (*service.UserSubscription, error) { clone := *sub; return &clone, nil }
	monthlyResets := 0
	repo := &stubUserSubscriptionRepo{
		getActive:    func(context.Context, int64, int64) (*service.UserSubscription, error) { return getSub() },
		getByID:      func(context.Context, int64) (*service.UserSubscription, error) { return getSub() },
		resetMonthly: func(context.Context, int64, time.Time) error { monthlyResets++; return nil },
	}
	svc := service.NewSubscriptionService(nil, repo, nil, nil, cfg)
	t.Cleanup(svc.Stop)
	for _, middleware := range []gin.HandlerFunc{gin.HandlerFunc(NewAPIKeyAuthMiddleware(keySvc, svc, cfg)), APIKeyAuthWithSubscriptionGoogle(keySvc, svc, cfg)} {
		r := gin.New()
		r.Use(middleware)
		r.GET("/t", func(c *gin.Context) {
			t.Error("exhausted subscription reached gateway handler")
			c.Status(http.StatusOK)
		})
		req := httptest.NewRequest(http.MethodGet, "/t", nil)
		req.Header.Set("x-api-key", key.Key)
		req.Header.Set("x-goog-api-key", key.Key)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusTooManyRequests, w.Code, w.Body.String())
		require.Contains(t, w.Body.String(), "monthly usage limit exceeded")
	}
	require.Zero(t, monthlyResets)
}
