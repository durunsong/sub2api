package middleware

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestImageStudioAuditPreservesMetadataWithoutCreativeBody(t *testing.T) {
	for _, path := range []string{"/api/v1/image-studio/generate", "/api/v1/image-studio/workflow", "/api/v1/image-studio/videos/generate"} {
		for _, status := range []int{http.StatusOK, http.StatusForbidden} {
			t.Run(fmt.Sprintf("%s/status=%d", path, status), func(t *testing.T) {
				repo := &auditCaptureRepository{}
				audit := service.NewAuditLogService(repo, nil)
				audit.Start()
				router := gin.New()
				router.Use(gin.HandlerFunc(NewAuditLogMiddleware(audit)))
				reached := false
				router.POST(path, func(c *gin.Context) {
					// Simulate an access ban/body-limit rejection before the studio handler.
					if status != http.StatusOK {
						c.AbortWithStatus(status)
						return
					}
					c.Next()
				}, func(c *gin.Context) {
					reached = true
					body, err := io.ReadAll(c.Request.Body)
					require.NoError(t, err)
					require.Contains(t, string(body), "private prompt")
					c.Request = c.Request.WithContext(ctxkey.WithNoContentPersistence(c.Request.Context()))
					c.Status(http.StatusOK)
				})
				request := httptest.NewRequest(http.MethodPost, path+"?prompt=private-query", bytes.NewBufferString(`{"prompt":"private prompt","references":["private image"]}`))
				request.Header.Set("Content-Type", "application/json")
				router.ServeHTTP(httptest.NewRecorder(), request)
				audit.Stop()
				repo.mu.Lock()
				defer repo.mu.Unlock()
				require.Equal(t, status == http.StatusOK, reached)
				require.Len(t, repo.logs, 1)
				require.NotContains(t, repo.logs[0].RequestBody, "private")
				require.NotContains(t, repo.logs[0].Extra, "query")
				require.Equal(t, status, repo.logs[0].StatusCode)
			})
		}
	}
}

func TestEphemeralRequestLoggerPreservesPrivacyAndCorrelation(t *testing.T) {
	initMiddlewareTestLogger(t)
	for _, private := range []bool{false, true} {
		t.Run(fmt.Sprintf("private=%t", private), func(t *testing.T) {
			router := gin.New()
			router.Use(RequestLogger())
			router.GET("/relay", func(c *gin.Context) {
				require.Equal(t, "privacy-correlation", c.Request.Context().Value(ctxkey.RequestID))
				require.Equal(t, !private, logger.FromContext(c.Request.Context()).Core().Enabled(zap.ErrorLevel))
				c.Status(http.StatusOK)
			})
			request := httptest.NewRequest(http.MethodGet, "/relay", nil)
			request.Header.Set(requestIDHeader, "privacy-correlation")
			if private {
				request = request.WithContext(ctxkey.WithNoContentPersistence(request.Context()))
			}
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			require.Equal(t, "privacy-correlation", response.Header().Get(requestIDHeader))
		})
	}
}
