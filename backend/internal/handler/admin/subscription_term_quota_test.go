//go:build unit

package admin

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type termAdjustmentRepo struct {
	service.UserSubscriptionRepository
	sub *service.UserSubscription
}

func (r *termAdjustmentRepo) GetByID(context.Context, int64) (*service.UserSubscription, error) {
	return r.sub, nil
}
func (r *termAdjustmentRepo) GetByIDForUpdate(context.Context, int64) (*service.UserSubscription, error) {
	return r.sub, nil
}
func (r *termAdjustmentRepo) ExtendExpiry(_ context.Context, _ int64, expiry time.Time) error {
	r.sub.ExpiresAt = expiry
	return nil
}

func TestSubscriptionExtendAcceptsLegacyAlignmentFlagsWithoutRebasingQuota(t *testing.T) {
	gin.SetMode(gin.TestMode)
	previous := service.DefaultIdempotencyCoordinator()
	t.Cleanup(func() { service.SetDefaultIdempotencyCoordinator(previous) })
	service.SetDefaultIdempotencyCoordinator(service.NewIdempotencyCoordinator(newMemoryIdempotencyRepoStub(), service.DefaultIdempotencyConfig()))
	for _, flag := range []string{"align_monthly_reset", "shift_monthly_reset"} {
		t.Run(flag, func(t *testing.T) {
			anchor := time.Now().UTC().Add(-time.Hour).Truncate(time.Second)
			expiry := anchor.Add(90 * 24 * time.Hour)
			repo := &termAdjustmentRepo{sub: &service.UserSubscription{ID: 1, UserID: 2, GroupID: 3,
				Status: service.SubscriptionStatusActive, StartsAt: anchor, ExpiresAt: expiry,
				MonthlyWindowStart: &anchor, MonthlyUsageUSD: 123.45, MonthlyUsageTokens: 678}}
			svc := service.NewSubscriptionService(nil, repo, nil, nil, nil)
			t.Cleanup(svc.Stop)
			h := NewSubscriptionHandler(svc)
			r := gin.New()
			r.POST("/subscriptions/:id/extend", h.Extend)
			body, err := json.Marshal(map[string]any{"days": 0, flag: true})
			require.NoError(t, err)
			req := httptest.NewRequest("POST", "/subscriptions/1/extend", strings.NewReader(string(body)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Idempotency-Key", "term-quota-"+flag)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			require.Equal(t, 200, w.Code, w.Body.String())
			require.Equal(t, expiry, repo.sub.ExpiresAt)
			require.Equal(t, anchor, *repo.sub.MonthlyWindowStart)
			require.Equal(t, expiry, *repo.sub.MonthlyResetTime())
			require.Equal(t, 123.45, repo.sub.MonthlyUsageUSD)
			require.EqualValues(t, 678, repo.sub.MonthlyUsageTokens)
		})
	}
}
