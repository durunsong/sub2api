//go:build unit

package handler

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	middleware "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestMuseVideoHandlerLifecycleAndOwnership(t *testing.T) {
	h, slots, bindings, upstream := newGrokMediaSlotHandler(t, false, false, "muse")
	newContext := func(method, id string, priced bool) (*gin.Context, *httptest.ResponseRecorder) {
		c, w := grokMediaSlotContext(context.Background(), method == http.MethodPost)
		key, _ := middleware.GetAPIKeyFromContext(c)
		key.Group.Platform = service.PlatformOpenAI
		if priced {
			key.Group.VideoModelPrices = map[string]map[string]float64{"muse-video": {"720p": 0.02}}
		}
		body := ""
		if method == http.MethodPost {
			body = `{"model":"public-muse-video","prompt":"waves","duration":6,"resolution":"720p"}`
		}
		c.Request = httptest.NewRequest(method, "/v1/videos/"+id, strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "request_id", Value: id}}
		return c, w
	}

	c, w := newContext(http.MethodPost, "", false)
	h.MuseVideoGeneration(c)
	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Zero(t, upstream.calls)
	c, w = newContext(http.MethodPost, "", true)
	key, _ := middleware.GetAPIKeyFromContext(c)
	key.Group.Platform = service.PlatformComposite
	c.Request = c.Request.WithContext(service.WithResolvedTargetPlatform(c.Request.Context(), service.PlatformGrok))
	h.MuseVideoGeneration(c)
	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Zero(t, upstream.calls)

	upstream.call = func(req *http.Request, _ int64) (*http.Response, error) {
		if req.Method == http.MethodPost {
			return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(`{"id":"task_123","status":"queued"}`))}, nil
		}
		return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(`{"id":"task_123","status":"processing"}`))}, nil
	}
	c, w = newContext(http.MethodPost, "", true)
	h.MuseVideoGeneration(c)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	require.Contains(t, w.Body.String(), "muse_task_123")
	require.Len(t, bindings.pending, 1)
	require.Equal(t, 1, upstream.calls)
	slots.assertReleased(t)

	c, w = newContext(http.MethodGet, "muse_task_123", false)
	h.MuseVideoStatus(c)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	require.Contains(t, w.Body.String(), "processing")
	require.Equal(t, 2, upstream.calls)
	slots.assertReleased(t)

	for _, other := range []string{"user", "key", "task"} {
		c, w = newContext(http.MethodGet, "muse_task_123", false)
		key, _ := middleware.GetAPIKeyFromContext(c)
		switch other {
		case "user":
			c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 11, Concurrency: 5})
		case "key":
			key.ID = 21
		case "task":
			c.Params = gin.Params{{Key: "request_id", Value: "muse_task_other"}}
		}
		before := upstream.calls
		h.MuseVideoStatus(c)
		require.Equal(t, http.StatusNotFound, w.Code, other+": "+w.Body.String())
		require.Equal(t, before, upstream.calls)
	}

	upstream.call = func(*http.Request, int64) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusServiceUnavailable, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(`{"error":"private prompt"}`))}, nil
	}
	c, w = newContext(http.MethodPost, "", true)
	before := upstream.calls
	h.MuseVideoGeneration(c)
	require.Equal(t, http.StatusBadGateway, w.Code)
	require.Equal(t, before+1, upstream.calls)
	require.NotContains(t, w.Body.String(), "private prompt")

	c, _ = newContext(http.MethodGet, "muse_task_123", false)
	key, _ = middleware.GetAPIKeyFromContext(c)
	subject, _ := middleware.GetAuthSubjectFromContext(c)
	result := &service.OpenAIForwardResult{ResponseID: "muse_task_123", VideoCount: 1}
	for i := 0; i < 2; i++ {
		billed := prepareGrokVideoCompletionBilling(context.Background(), h, zap.NewNop(), key, subject, result.ResponseID, result)
		if i == 0 {
			require.NotNil(t, billed)
			require.Equal(t, "muse-video", billed.BillingModel)
			require.Equal(t, 6, billed.VideoDurationSeconds)
			require.Equal(t, "720p", billed.VideoResolution)
		} else {
			require.Nil(t, billed)
		}
	}
	require.Len(t, bindings.billed, 1)
}
