package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

func TestEphemeralOpsMiddlewareDoesNotCaptureImageResponse(t *testing.T) {
	router := gin.New()
	router.Use(OpsErrorLoggerMiddleware(nil))
	router.POST("/v1/images/generations", func(c *gin.Context) {
		_, capture := c.Writer.(*opsCaptureWriter)
		require.False(t, capture)
		require.False(t, requestLogger(c, "image-studio").Core().Enabled(zapcore.ErrorLevel))
		c.JSON(http.StatusBadGateway, gin.H{"error": "private upstream echo"})
	})
	request := httptest.NewRequest(http.MethodPost, "/v1/images/generations", nil).WithContext(ctxkey.WithNoContentPersistence(context.Background()))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	require.Equal(t, http.StatusBadGateway, response.Code)
}
