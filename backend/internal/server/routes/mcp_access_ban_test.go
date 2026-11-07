package routes

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/accessban"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type mcpAccessBanRepository struct {
	service.IPBanRepository
	rules []service.IPBan
	err   error
}

func (r *mcpAccessBanRepository) ListActive(context.Context, time.Time) ([]service.IPBan, error) {
	return r.rules, r.err
}
func (r *mcpAccessBanRepository) RecordHit(context.Context, int64, time.Time) error { return nil }

func TestMCPRoutesEnforceAccessBanBeforeProtocolHandlers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	endpoints := []struct{ method, path string }{
		{http.MethodPost, "/api/v1/mcp"},
		{http.MethodGet, "/api/v1/mcp"},
		{http.MethodDelete, "/api/v1/mcp"},
		{http.MethodPost, "/api/v1/mcp/oauth/register"},
		{http.MethodPost, "/api/v1/mcp/oauth/token"},
		{http.MethodPost, "/api/v1/mcp/oauth/revoke"},
		{http.MethodGet, "/.well-known/oauth-protected-resource/api/v1/mcp"},
		{http.MethodGet, "/.well-known/oauth-protected-resource"},
		{http.MethodGet, "/.well-known/oauth-authorization-server"},
		{http.MethodGet, "/api/v1/user/mcp/config"},
		{http.MethodGet, "/api/v1/user/mcp/grants"},
		{http.MethodDelete, "/api/v1/user/mcp/grants/grant-1"},
		{http.MethodGet, "/api/v1/user/mcp/authorize"},
		{http.MethodPost, "/api/v1/user/mcp/authorize"},
	}
	for _, tc := range []struct {
		name   string
		rule   service.IPBan
		err    error
		status int
	}{
		{name: "IP", rule: service.IPBan{RuleType: accessban.RuleTypeIP, Pattern: "43.255.119.7"}, status: http.StatusForbidden},
		{name: "UA", rule: service.IPBan{RuleType: accessban.RuleTypeUA, Pattern: "blocked-agent"}, status: http.StatusForbidden},
		{name: "IP and UA", rule: service.IPBan{RuleType: accessban.RuleTypeIPUA, Pattern: "43.255.119.7", UAPattern: "blocked-agent"}, status: http.StatusForbidden},
		{name: "automatic login ban uses trusted IP", rule: service.IPBan{RuleType: accessban.RuleTypeIP, Pattern: "43.255.119.7", Source: service.IPBanSourceAdminLogin}, status: http.StatusForbidden},
		{name: "repository failure closes access", err: errors.New("unavailable"), status: http.StatusInternalServerError},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tc.rule.ID = 1
			tc.rule.Status = service.IPBanStatusActive
			repo := &mcpAccessBanRepository{rules: []service.IPBan{tc.rule}, err: tc.err}
			guard := middleware.IPBanGuard(service.NewIPBanService(repo), nil)
			router := gin.New()
			require.NoError(t, router.SetTrustedProxies(nil))
			h := mcpRouteHandlers()
			v1 := router.Group("/api/v1")
			RegisterUserRoutes(v1, h, middleware.JWTAuthMiddleware(mcpRouteNext), middleware.AuditLogMiddleware(mcpRouteNext), nil, nil, guard)
			RegisterMCPRoutes(v1, h, middleware.AuditLogMiddleware(mcpRouteNext), middleware.StepUpAuthMiddleware(mcpRouteNext), nil, nil, nil, nil, guard)
			RegisterMCPOAuthRoutes(router, v1, h, nil, guard)
			for _, endpoint := range endpoints {
				t.Run(endpoint.method+" "+endpoint.path, func(t *testing.T) {
					req := httptest.NewRequest(endpoint.method, endpoint.path, nil)
					req.RemoteAddr = "43.255.119.7:1234"
					req.Header.Set("User-Agent", "blocked-agent")
					if tc.rule.Source == service.IPBanSourceAdminLogin {
						req.Header.Set("X-Forwarded-For", "8.8.8.8")
					}
					w := httptest.NewRecorder()
					router.ServeHTTP(w, req)
					require.Equal(t, tc.status, w.Code, w.Body.String())
				})
			}
		})
	}
}
