package routes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/handler"
	servermiddleware "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type gatewayBannedClientRepo struct{ service.IPBanRepository }

func (gatewayBannedClientRepo) ListActive(context.Context, time.Time) ([]service.IPBan, error) {
	return []service.IPBan{{ID: 1, RuleType: "ua", Pattern: "blocked-client", Status: service.IPBanStatusActive}}, nil
}

func (gatewayBannedClientRepo) RecordHit(context.Context, int64, time.Time) error { return nil }

func TestGatewayRoutesAccessBanPrecedesAdmission(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterGatewayRoutes(router, &handler.Handlers{
		Gateway: &handler.GatewayHandler{}, OpenAIGateway: &handler.OpenAIGatewayHandler{},
		AsyncImage: handler.NewAsyncImageHandler(nil, nil),
	}, servermiddleware.APIKeyAuthMiddleware(func(c *gin.Context) {
		t.Error("banned client reached authentication")
		c.AbortWithStatus(http.StatusUnauthorized)
	}), nil, nil, nil, nil, nil, &config.Config{}, service.NewIPBanService(gatewayBannedClientRepo{}))

	for _, path := range []string{
		"/v1/messages", "/responses", "/responses/compact", "/chat/completions",
		"/messages/count_tokens", "/images/generations", "/videos/generations", "/x_search",
		"/backend-api/codex/responses", "/backend-api/codex/realtime/calls",
		"/v1beta/models/gemini-2.5-pro:generateContent", "/antigravity/v1/messages",
		"/antigravity/v1beta/models/gemini-2.5-pro:generateContent",
	} {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{"model":"unlisted-model"}`))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("User-Agent", "blocked-client/1.0")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			require.Equal(t, http.StatusForbidden, w.Code, w.Body.String())
			require.Contains(t, w.Body.String(), "Access denied")
		})
	}
}

func TestGatewayRoutesKiroModelAllowlist(t *testing.T) {
	router := newGatewayRoutesTestRouterWithGroup(allowlistGroup(service.PlatformKiro, true, "claude-*"))
	for _, path := range []string{"/v1/messages", "/v1/responses", "/responses", "/chat/completions"} {
		req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{"model":"gpt-5.5"}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusNotFound, w.Code, path)
		require.Contains(t, w.Body.String(), "not available for this group", path)
	}
}
