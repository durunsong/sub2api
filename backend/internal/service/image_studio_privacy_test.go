package service

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestEphemeralUpstreamBodiesDoNotReachGlobalLogging(t *testing.T) {
	sink, cleanup := captureStructuredLog(t)
	defer cleanup()
	ctx := ctxkey.WithNoContentPersistence(context.Background())
	cfg := &config.Config{}
	cfg.Gateway.LogUpstreamErrorBody = true
	cfg.Gateway.LogUpstreamErrorBodyMaxBytes = 2048
	body := []byte(`{"error":{"message":"private-creative-canary"}}`)
	for _, platform := range []string{PlatformOpenAI, PlatformGemini, PlatformAntigravity} {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodPost, "/images", nil).WithContext(ctx)
		account := &Account{ID: 1, Platform: platform, Type: AccountTypeAPIKey}
		if platform == PlatformAntigravity {
			_ = (&AntigravityGatewayService{settingService: &SettingService{cfg: cfg}}).writeMappedClaudeError(c, account, 400, "request-1", body)
		} else if platform == PlatformGemini {
			(&GeminiMessagesCompatService{cfg: cfg}).recordGeminiNativeUpstreamError(c, account, 400, body, "request-1")
		} else {
			_, _ = (&OpenAIGatewayService{cfg: cfg}).handleOpenAIImagesErrorResponse(ctx, &http.Response{StatusCode: 400, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(string(body)))}, c, account)
		}
	}
	require.False(t, sink.ContainsMessage("private-creative-canary"))
}

func TestEphemeralAntigravityMalformedOutputDoesNotReachLogs(t *testing.T) {
	sink, cleanup := captureStructuredLog(t)
	defer cleanup()
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/images", nil).WithContext(ctxkey.WithNoContentPersistence(context.Background()))
	svc := &AntigravityGatewayService{settingService: &SettingService{cfg: &config.Config{}}}
	frames := `data: {"response":{"candidates":[{"finishReason":"MALFORMED_FUNCTION_CALL","content":{"parts":[{"text":"private-malformed-canary"}]}}]}}` + "\n\n"
	_, _ = svc.handleGeminiStreamToNonStreaming(c, &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(frames))}, time.Now())
	require.False(t, sink.ContainsMessage("private-malformed-canary"))
}

func TestAntigravityRetryParserDoesNotLogUntrustedDuration(t *testing.T) {
	sink, cleanup := captureStructuredLog(t)
	defer cleanup()
	parseAntigravitySmartRetryInfo([]byte(`{"error":{"status":"RESOURCE_EXHAUSTED","details":[{"@type":"type.googleapis.com/google.rpc.RetryInfo","retryDelay":"private-duration-canary"}]}}`))
	require.False(t, sink.ContainsMessage("private-duration-canary"))
}

func TestEphemeralOpsContextOmitsUpstreamEcho(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/images/generations", nil).WithContext(ctxkey.WithNoContentPersistence(context.Background()))
	setOpsUpstreamError(c, 400, "private prompt", "private image")
	appendOpsUpstreamError(c, OpsUpstreamErrorEvent{UpstreamStatusCode: 400, Message: "private prompt", Detail: "private image", UpstreamResponseBody: "private body"})
	require.Empty(t, c.GetString(OpsUpstreamErrorMessageKey))
	require.Empty(t, c.GetString(OpsUpstreamErrorDetailKey))
	value, _ := c.Get(OpsUpstreamErrorsKey)
	encoded, err := json.Marshal(value)
	require.NoError(t, err)
	require.NotContains(t, string(encoded), "private")
	require.Contains(t, string(encoded), "400")
	MarkOpsCyberPolicy(c, CyberPolicyMark{Message: "private prompt", Body: "private image", UpstreamStatus: 400, UpstreamInTok: 12})
	mark := GetOpsCyberPolicy(c)
	require.Empty(t, mark.Message)
	require.Empty(t, mark.Body)
	require.Equal(t, 12, mark.UpstreamInTok)
}

func TestEphemeralModerationKeepsBlockAndObserveSemanticsWithoutContentQueue(t *testing.T) {
	for _, mode := range []string{ContentModerationModeObserve, ContentModerationModePreBlock} {
		t.Run(mode, func(t *testing.T) {
			var upstreamCalls atomic.Int64
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				upstreamCalls.Add(1)
				_ = json.NewEncoder(w).Encode(moderationAPIResponse{Results: []moderationAPIResult{{CategoryScores: map[string]float64{"sexual": 0.99, "private creative prompt": 1.0}}}})
			}))
			defer server.Close()
			cfg := defaultContentModerationConfig()
			cfg.Enabled, cfg.RecordNonHits, cfg.Mode = true, true, mode
			cfg.BaseURL, cfg.APIKeys = server.URL, []string{"test-key"}
			svc, repo := newContentModerationModelFilterTestService(t, cfg)
			decision, err := svc.Check(ctxkey.WithNoContentPersistence(context.Background()), ContentModerationCheckInput{
				Protocol: ContentModerationProtocolOpenAIImages, Body: []byte(`{"prompt":"private creative prompt"}`),
			})
			require.NoError(t, err)
			require.Equal(t, int64(1), upstreamCalls.Load(), "audit must complete in the request, including observe mode")
			require.Equal(t, mode == ContentModerationModePreBlock, decision.Blocked)
			require.Zero(t, svc.asyncEnqueued.Load())
			logs := requireContentModerationLogCount(t, repo, 1)
			encoded, err := json.Marshal(logs)
			require.NoError(t, err)
			require.NotContains(t, string(encoded), "private creative prompt")
			require.Empty(t, logs[0].InputExcerpt)
			require.Empty(t, logs[0].Error)
			require.Nil(t, logs[0].EngineMeta)
		})
	}
}
