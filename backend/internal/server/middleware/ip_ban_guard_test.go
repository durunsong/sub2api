//go:build unit

package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/accessban"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type stubIPBanRepo struct {
	active []service.IPBan
	hits   []int64
}

func (s *stubIPBanRepo) Create(context.Context, *service.IPBan) error { return nil }
func (s *stubIPBanRepo) GetByID(context.Context, int64) (*service.IPBan, error) {
	return nil, service.ErrIPBanNotFound
}
func (s *stubIPBanRepo) List(context.Context, pagination.PaginationParams, service.IPBanListFilters) ([]service.IPBan, *pagination.PaginationResult, error) {
	return nil, nil, nil
}
func (s *stubIPBanRepo) Update(context.Context, *service.IPBan) error { return nil }
func (s *stubIPBanRepo) Delete(context.Context, int64) error          { return nil }
func (s *stubIPBanRepo) ListActive(context.Context, time.Time) ([]service.IPBan, error) {
	return append([]service.IPBan(nil), s.active...), nil
}
func (s *stubIPBanRepo) RecordHit(_ context.Context, id int64, _ time.Time) error {
	s.hits = append(s.hits, id)
	return nil
}

func TestGatewayIPBanGuardUsesForwardedClientIPWithoutTrustedProxies(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &stubIPBanRepo{
		active: []service.IPBan{{ID: 1, RuleType: accessban.RuleTypeIP, Pattern: "1.2.3.4", Status: service.IPBanStatusActive}},
	}
	svc := service.NewIPBanService(repo)
	cfg := &config.Config{RunMode: config.RunModeSimple}

	router := gin.New()
	router.Use(GatewayIPBanGuard(svc, cfg, AnthropicErrorWriter))
	router.GET("/t", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/t", nil)
	req.RemoteAddr = "9.9.9.9:12345"
	req.Header.Set("X-Forwarded-For", "1.2.3.4")
	req.Header.Set("X-Real-IP", "1.2.3.4")
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusForbidden, w.Code)
	require.Contains(t, w.Body.String(), service.ErrIPBanned.Message)
	require.Equal(t, []int64{1}, repo.hits)
}

func TestGatewayIPBanGuardBlocksUARule(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &stubIPBanRepo{
		active: []service.IPBan{{ID: 2, RuleType: accessban.RuleTypeUA, Pattern: "evilbot", Status: service.IPBanStatusActive}},
	}
	svc := service.NewIPBanService(repo)
	cfg := &config.Config{RunMode: config.RunModeSimple}

	router := gin.New()
	router.Use(GatewayIPBanGuard(svc, cfg, AnthropicErrorWriter))
	router.GET("/t", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/t", nil)
	req.RemoteAddr = "9.9.9.9:12345"
	req.Header.Set("User-Agent", "evilbot/1.0")
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusForbidden, w.Code)
	require.Contains(t, w.Body.String(), service.ErrClientAccessBanned.Message)
	require.Equal(t, []int64{2}, repo.hits)
}
