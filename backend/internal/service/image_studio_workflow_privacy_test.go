package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"github.com/Wei-Shaw/sub2api/internal/pkg/openai_compat"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestEphemeralChatProtocolFallbackDoesNotLogProviderContent(t *testing.T) {
	logs, cleanup := captureStructuredLog(t)
	defer cleanup()
	body := []byte(`{"model":"gpt-5.4","stream":false,"messages":[{"role":"user","content":"private-brief-canary"}]}`)
	ctx := logger.IntoContext(ctxkey.WithNoContentPersistence(context.Background()), zap.NewNop())
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil).WithContext(ctx)
	upstream := &httpUpstreamRecorder{responses: []*http.Response{
		{StatusCode: 404, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(`{"error":{"message":"private-provider-canary"}}`))},
		{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(`{"choices":[{"message":{"content":"ok"},"finish_reason":"stop"}],"usage":{"prompt_tokens":1,"completion_tokens":2}}`))},
	}}
	account := structuredInputRetryTestAccount()
	account.Extra = map[string]any{openai_compat.ExtraKeyResponsesSupported: nil}
	svc := &OpenAIGatewayService{cfg: structuredInputRetryTestConfig(), httpUpstream: upstream}
	_, _ = svc.ForwardAsChatCompletions(ctx, c, account, body, "", "")
	require.False(t, logs.ContainsFieldValue("upstream_message", "private-provider-canary"))
}
func TestEphemeralAnthropicChatStreamErrorDoesNotLogProviderContent(t *testing.T) {
	logs, cleanup := captureStructuredLog(t)
	defer cleanup()
	ctx := ctxkey.WithNoContentPersistence(context.Background())
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil).WithContext(ctx)
	svc := &GatewayService{cfg: &config.Config{}}
	_ = svc.sseStreamErrorEventFailover(ctx, c, &Account{ID: 1, Platform: PlatformAnthropic}, &http.Response{Header: http.Header{}}, `{"type":"error","error":{"type":"permission_error","message":"private-anthropic-canary"}}`, "claude-sonnet-4", false)
	require.False(t, logs.ContainsMessage("private-anthropic-canary"))
}
func TestEphemeralKiroChatErrorDoesNotLogProviderContent(t *testing.T) {
	logs, cleanup := captureStructuredLog(t)
	defer cleanup()
	ctx := ctxkey.WithNoContentPersistence(context.Background())
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil).WithContext(ctx)
	svc := &GatewayService{cfg: &config.Config{}}
	_ = svc.handleKiroHTTPError(ctx, &http.Response{StatusCode: 400, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(`{"message":"private-kiro-canary"}`))}, c, &Account{ID: 1, Platform: PlatformKiro}, "claude-sonnet-4", nil)
	require.False(t, logs.ContainsFieldValue("body_excerpt", "private-kiro-canary"))
}
func TestEphemeralChatBufferedReadErrorDoesNotLogProviderContent(t *testing.T) {
	logs, cleanup := captureStructuredLog(t)
	defer cleanup()
	ctx := logger.IntoContext(ctxkey.WithNoContentPersistence(context.Background()), zap.NewNop())
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil).WithContext(ctx)
	svc := &OpenAIGatewayService{cfg: &config.Config{}}
	_, _ = svc.handleChatBufferedStreamingResponse(&http.Response{Header: http.Header{}, Body: io.NopCloser(&studioErrorReader{})}, c, &Account{ID: 1, Platform: PlatformOpenAI}, "gpt-5.4", "gpt-5.4", "gpt-5.4", time.Now())
	require.False(t, logs.ContainsFieldValue("error", "private-read-canary"))
}

type studioErrorReader struct{}

func (*studioErrorReader) Read([]byte) (int, error) { return 0, studioReadError{} }

type studioReadError struct{}

func (studioReadError) Error() string { return "private-read-canary" }

func TestEphemeralKiroAuthQuarantineDoesNotPersistProviderContent(t *testing.T) {
	repo := &studioWorkflowTempRepo{}
	svc := &GatewayService{accountRepo: repo}
	svc.markKiroAuthTemporarilyUnavailable(ctxkey.WithNoContentPersistence(context.Background()), &Account{ID: 1, Platform: PlatformKiro}, 403, "private-auth-canary")
	require.NotContains(t, repo.reason, "private-auth-canary")
	require.Contains(t, repo.reason, "403")
}

type studioWorkflowTempRepo struct {
	AccountRepository
	reason string
}

func (r *studioWorkflowTempRepo) SetTempUnschedulable(_ context.Context, _ int64, _ time.Time, reason string) error {
	r.reason = reason
	return nil
}

func TestEphemeralKiro429DebugDoesNotReadOrLogContent(t *testing.T) {
	logs, cleanup := captureStructuredLog(t)
	defer cleanup()
	resp := &http.Response{StatusCode: 429, Header: http.Header{}, Body: io.NopCloser(strings.NewReader("private-429-canary"))}
	dumpKiro429ResponseForDebug(ctxkey.WithNoContentPersistence(context.Background()), resp, 1, "https://example.test", "fixture")
	require.False(t, logs.ContainsFieldValue("body_sample", "private-429-canary"))
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "private-429-canary", string(body))
}
func TestEphemeralGeminiAndAntigravityChatDoNotLogErrorContent(t *testing.T) {
	for _, platform := range []string{PlatformGemini, PlatformAntigravity} {
		t.Run(platform, func(t *testing.T) {
			logs, cleanup := captureStructuredLog(t)
			defer cleanup()
			ctx := logger.IntoContext(ctxkey.WithNoContentPersistence(context.Background()), zap.NewNop())
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil).WithContext(ctx)
			upstream := &httpUpstreamRecorder{responses: []*http.Response{{StatusCode: 400, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(`{"error":{"message":"private-chat-native-canary","status":"INVALID_ARGUMENT"}}`))}}}
			cfg := &config.Config{}
			cfg.Gateway.LogUpstreamErrorBody = true
			cfg.Gateway.LogUpstreamErrorBodyMaxBytes = 2048
			if platform == PlatformGemini {
				svc := &GeminiMessagesCompatService{cfg: cfg, httpUpstream: upstream}
				_, _ = svc.ForwardAsChatCompletions(ctx, c, &Account{ID: 1, Platform: PlatformGemini, Type: AccountTypeAPIKey, Credentials: map[string]any{"api_key": "fixture-key"}}, []byte(`{"model":"gemini-3-flash","stream":false,"messages":[{"role":"user","content":"private-brief-canary"}]}`))
			} else {
				svc := newAntigravityCompatService(cfg.Gateway, upstream)
				_, _ = svc.ForwardAsChatCompletions(ctx, c, newAntigravityCompatAccount(AccountTypeOAuth), []byte(`{"model":"claude-sonnet-4-5","stream":false,"messages":[{"role":"user","content":"private-brief-canary"}]}`), nil)
			}
			require.Len(t, upstream.requests, 1)
			logs.mu.Lock()
			defer logs.mu.Unlock()
			for _, event := range logs.events {
				encoded, _ := json.Marshal(event)
				require.NotContains(t, string(encoded), "private-chat-native-canary")
				require.NotContains(t, string(encoded), "private-brief-canary")
			}
		})
	}
}

func TestEphemeralModerationFailureDoesNotRetainProviderContentInKeyHealth(t *testing.T) {
	for _, statusCode := range []int{http.StatusBadRequest, http.StatusForbidden} {
		for _, ephemeral := range []bool{false, true} {
			t.Run(fmt.Sprintf("http_%d/ephemeral_%t", statusCode, ephemeral), func(t *testing.T) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(statusCode)
					_, _ = io.WriteString(w, `{"error":{"message":"private-moderation-health-canary"}}`)
				}))
				defer server.Close()
				cfg := defaultContentModerationConfig()
				cfg.BaseURL, cfg.APIKeys, cfg.RetryCount = server.URL, []string{"fixture-moderation-key"}, 0
				svc, _ := newContentModerationModelFilterTestService(t, cfg)
				ctx := context.Background()
				if ephemeral {
					ctx = ctxkey.WithNoContentPersistence(ctx)
				}
				_, err := svc.callModeration(ctx, cfg, "private-input-canary")
				require.Error(t, err)
				statuses := svc.apiKeyStatuses(cfg.APIKeys, cfg.Engine)
				require.Len(t, statuses, 1)
				state := statuses[0]
				require.NotEmpty(t, state.LastError)
				require.Equal(t, statusCode, state.LastHTTPStatus)
				require.True(t, state.LastTested)
				if statusCode == http.StatusBadRequest {
					require.Equal(t, "error", state.Status)
					require.Zero(t, state.FailureCount)
					require.Nil(t, state.FrozenUntil)
				} else {
					require.Equal(t, "frozen", state.Status)
					require.Equal(t, 1, state.FailureCount)
					require.NotNil(t, state.FrozenUntil)
				}
				serialized, err := json.Marshal(statuses)
				require.NoError(t, err)
				if ephemeral {
					require.NotContains(t, string(serialized), "private-moderation-health-canary")
					require.NotContains(t, string(serialized), "private-input-canary")
				} else {
					require.Contains(t, state.LastError, "private-moderation-health-canary")
				}
			})
		}
	}
}
