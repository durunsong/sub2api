package handler

import (
	"context"
	"encoding/json"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	pkghttputil "github.com/Wei-Shaw/sub2api/internal/pkg/httputil"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newVideoStudioForTest(t *testing.T, relay http.Handler) (*ImageStudioHandler, *studioUserRepo, *studioGroupRepo) {
	h, users, groups, _ := newStudioForTest(t, "", relay)
	groups.groups[0].Platform = service.PlatformGrok
	h.gateway = newGatewayModelsHandlerForTest(&gatewayModelsAccountRepoStub{byGroup: map[int64][]service.Account{1: {{ID: 1, Platform: service.PlatformGrok, Credentials: map[string]any{"model_mapping": map[string]any{"grok-imagine-video": "grok-imagine-video"}}}}}})
	return h, users, groups
}

func TestVideoStudioCatalogHonorsMediaPermissionAndVisibility(t *testing.T) {
	h, users, groups := newVideoStudioForTest(t, nil)
	catalog, err := h.videoCatalog(context.Background(), 7, false)
	require.NoError(t, err)
	require.Len(t, catalog.Models, 1)
	groups.groups[0].AllowImageGeneration = false
	catalog, err = h.videoCatalog(context.Background(), 7, false)
	require.NoError(t, err)
	require.Empty(t, catalog.Models)
	groups.groups[0].AllowImageGeneration = true
	require.NoError(t, h.settings.SetVideoStudioConfig(context.Background(), service.VideoStudioConfig{Enabled: true, Models: []service.ImageStudioModelConfig{{GroupID: 1, Model: "grok-imagine-video", Enabled: false}}}))
	catalog, err = h.videoCatalog(context.Background(), 7, false)
	require.NoError(t, err)
	require.Empty(t, catalog.Models)
	catalog, err = h.videoCatalog(context.Background(), 0, true)
	require.NoError(t, err)
	require.Len(t, catalog.Models, 1)
	users.user.AllowedGroups = nil
	catalog, err = h.videoCatalog(context.Background(), 7, false)
	require.NoError(t, err)
	require.Empty(t, catalog.Models)
}

func TestVideoStudioPrivateRelayAndAuthorizedRecovery(t *testing.T) {
	calls := 0
	relay := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		require.True(t, ctxkey.NoContentPersistence(r.Context()))
		require.True(t, strings.HasPrefix(r.Header.Get("Authorization"), "Bearer internal:images:"))
		require.Empty(t, r.URL.RawQuery)
		require.Empty(t, r.Header.Get("Cookie"))
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			require.Equal(t, "/v1/videos/generations", r.URL.Path)
			_, _ = io.WriteString(w, `{"request_id":"task-1"}`)
		} else {
			require.Equal(t, "/v1/videos/task-1", r.URL.Path)
			_, _ = io.WriteString(w, `{"status":"done","video":{"url":"https://private.example/video"}}`)
		}
	})
	h, users, _ := newVideoStudioForTest(t, relay)
	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 7}) })
	router.POST("/generate", h.RequestBodyLimit(), h.GenerateVideo)
	router.GET("/status", h.VideoStatus)
	body := `{"model_id":"` + studioModelID(1, "grok-imagine-video") + `","prompt":"camera moves","seconds":6,"size":"1280x720","resolution":"720","mode":"frames"}`
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest("POST", "/generate?secret=drop", strings.NewReader(body)))
	require.Equal(t, 200, w.Code, w.Body.String())
	require.Equal(t, 1, calls)
	require.NoError(t, h.settings.SetVideoStudioConfig(context.Background(), service.VideoStudioConfig{Enabled: false}))
	w = httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest("GET", "/status?group_id=1&task_id=task-1", nil))
	require.Equal(t, 200, w.Code, w.Body.String())
	require.NotContains(t, w.Body.String(), "private.example")
	require.Equal(t, 2, calls)
	users.user.AllowedGroups = nil
	w = httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest("GET", "/status?group_id=1&task_id=task-1", nil))
	require.Equal(t, 403, w.Code)
	require.Equal(t, 2, calls)
}

func TestVideoStudioPayloadPreservesModesAndRejectsUnsafeInput(t *testing.T) {
	input := videoStudioRequest{ModelID: "1:grok-imagine-video", Prompt: "camera pans", Seconds: 6, Resolution: "720", Size: "1280x720", Mode: "frames", References: []string{testStudioPNG, testStudioPNG}}
	body, err := buildVideoStudioRequest("grok-imagine-video", input)
	require.NoError(t, err)
	var data map[string]any
	require.NoError(t, json.Unmarshal(body, &data))
	require.Equal(t, float64(6), data["duration"])
	require.Equal(t, "16:9", data["aspect_ratio"])
	require.Equal(t, "720p", data["resolution"])
	require.NotNil(t, data["last_frame"])
	require.Nil(t, data["reference_images"])
	input.Mode = "reference"
	body, err = buildVideoStudioRequest("grok-imagine-video", input)
	require.NoError(t, err)
	require.Contains(t, string(body), "reference_images")
	require.NotContains(t, string(body), "last_frame")
	for _, change := range []func(*videoStudioRequest){func(v *videoStudioRequest) { v.Seconds = 0 }, func(v *videoStudioRequest) { v.Seconds = 31 }, func(v *videoStudioRequest) { v.Mode = "unknown" }, func(v *videoStudioRequest) { v.References = []string{"http://localhost/private"} }, func(v *videoStudioRequest) { v.Prompt = " " }, func(v *videoStudioRequest) { v.Size = "999999x1" }} {
		invalid := input
		change(&invalid)
		_, err := buildVideoStudioRequest("grok-imagine-video", invalid)
		require.Error(t, err)
	}
}
func TestVideoStudioNormalizesTaskWithoutEchoingProviderContent(t *testing.T) {
	for _, raw := range []string{`{"request_id":"task-1"}`, `{"id":"task-1","status":"pending","progress":25}`} {
		result, err := normalizeVideoStudioStatus([]byte(raw), "")
		require.NoError(t, err)
		require.Equal(t, "pending", result.Status)
		require.Equal(t, "task-1", result.ID)
	}
	result, err := normalizeVideoStudioStatus([]byte(`{"status":"done","video":{"url":"https://upstream/private.mp4","respect_moderation":true}}`), "task-1")
	require.NoError(t, err)
	require.Equal(t, "completed", result.Status)
	serialized, _ := json.Marshal(result)
	require.NotContains(t, string(serialized), "upstream")
	result, err = normalizeVideoStudioStatus([]byte(`{"status":"failed","error":{"message":"SECRET PROMPT"}}`), "task-1")
	require.NoError(t, err)
	require.Equal(t, "failed", result.Status)
	serialized, _ = json.Marshal(result)
	require.NotContains(t, string(serialized), "SECRET")
	_, err = normalizeVideoStudioStatus([]byte(`{"request_id":"../secrets"}`), "")
	require.Error(t, err)
}

func TestVideoStudioPendingDoesNotCompleteEarly(t *testing.T) {
	result, err := normalizeVideoStudioStatus([]byte(`{"status":"pending","video":{"url":"https://provider/partial.mp4"}}`), "task")
	require.NoError(t, err)
	require.Equal(t, "pending", result.Status)
	result, err = normalizeVideoStudioStatus([]byte(`{"status":"done","video":{"url":"https://provider/blocked.mp4","respect_moderation":false}}`), "task")
	require.NoError(t, err)
	require.Equal(t, "failed", result.Status)
}

func TestVideoStudioContentRejectsIncompleteSuccess(t *testing.T) {
	for _, test := range []struct {
		name                                  string
		status                                int
		length, contentRange, readError, body string
		expected                              int
	}{
		{"complete", 200, "3", "", "", "abc", 200},
		{"truncated", 200, "10", "", "", "abc", 502},
		{"invalid length", 200, "bad", "", "", "abc", 502},
		{"partial", 206, "3", "bytes 0-2/100", "", "abc", 502},
		{"chunked read failed", 200, "", "", "true", "abc", 502},
		{"empty", 200, "0", "", "", "", 502},
	} {
		t.Run(test.name, func(t *testing.T) {
			h, _, _ := newVideoStudioForTest(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "video/mp4")
				w.Header().Set("Content-Length", test.length)
				w.Header().Set("Content-Range", test.contentRange)
				w.Header().Set("X-Sub2API-Content-Read-Error", test.readError)
				w.WriteHeader(test.status)
				_, _ = io.WriteString(w, test.body)
			}))
			router := gin.New()
			router.Use(func(c *gin.Context) { c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 7}) })
			router.GET("/content", h.VideoContent)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, httptest.NewRequest("GET", "/content?group_id=1&task_id=task-1", nil))
			require.Equal(t, test.expected, w.Code, w.Body.String())
		})
	}
}

func TestVideoStudioOutputMemoryBudgetIsSharedAndReleased(t *testing.T) {
	budget := pkghttputil.NewBodyMemoryBudget(20)
	ctx, release := pkghttputil.WithRequestBodyMemoryBudget(context.Background(), budget)
	capture := &videoStudioCapture{limit: 100, budgetContext: ctx}
	_, err := capture.Write([]byte("123456"))
	require.NoError(t, err)
	otherCtx, otherRelease := pkghttputil.WithRequestBodyMemoryBudget(context.Background(), budget)
	defer otherRelease()
	other := &videoStudioCapture{limit: 100, budgetContext: otherCtx}
	_, err = other.Write([]byte("12345"))
	require.ErrorIs(t, err, pkghttputil.ErrRequestBodyMemoryBudget)
	require.True(t, other.overflow)
	require.Zero(t, other.body.Len())
	release()
	_, err = other.Write([]byte("12345"))
	require.NoError(t, err)
}
