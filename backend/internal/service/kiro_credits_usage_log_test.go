//go:build unit

package service

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	kiropkg "github.com/Wei-Shaw/sub2api/internal/pkg/kiro"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestKiroUsageToClaudeMapsKiroCredits(t *testing.T) {
	usage := kiroUsageToClaude(kiropkg.Usage{
		InputTokens:  12,
		OutputTokens: 7,
		KiroCredits:  0.17,
	}, 0)

	require.Equal(t, 12, usage.InputTokens)
	require.Equal(t, 7, usage.OutputTokens)
	require.InDelta(t, 0.17, usage.KiroCredits, 0.000001)
}

func TestBuildRecordUsageLogPersistsPositiveKiroCredits(t *testing.T) {
	svc := &GatewayService{}
	log := svc.buildRecordUsageLog(
		context.Background(),
		&recordUsageCoreInput{},
		&ForwardResult{
			RequestID: "req_kiro_credits",
			Model:     "claude-sonnet-4-5",
			Usage: ClaudeUsage{
				InputTokens:           12,
				OutputTokens:          7,
				CacheCreation5mTokens: 3,
				CacheCreation1hTokens: 5,
				KiroCredits:           0.17,
			},
			Duration: time.Second,
		},
		&APIKey{ID: 10},
		&User{ID: 20},
		&Account{ID: 30, Platform: PlatformKiro},
		nil,
		"claude-sonnet-4-5",
		1,
		1,
		1,
		1,
		false,
		&CostBreakdown{},
	)

	require.NotNil(t, log.KiroCredits)
	require.InDelta(t, 0.17, *log.KiroCredits, 0.000001)
	require.Equal(t, 8, log.CacheCreationTokens)
	require.Equal(t, 27, log.TotalTokens())
}

func TestBuildRecordUsageLogOmitsZeroKiroCredits(t *testing.T) {
	svc := &GatewayService{}
	log := svc.buildRecordUsageLog(
		context.Background(),
		&recordUsageCoreInput{},
		&ForwardResult{
			RequestID: "req_kiro_zero_credits",
			Model:     "claude-sonnet-4-5",
			Usage: ClaudeUsage{
				InputTokens:  12,
				OutputTokens: 7,
			},
			Duration: time.Second,
		},
		&APIKey{ID: 10},
		&User{ID: 20},
		&Account{ID: 30, Platform: PlatformKiro},
		nil,
		"claude-sonnet-4-5",
		1,
		1,
		1,
		1,
		false,
		&CostBreakdown{},
	)

	require.Nil(t, log.KiroCredits)
}

func TestBuildRecordUsageLogOmitsNonKiroCredits(t *testing.T) {
	log := (&GatewayService{}).buildRecordUsageLog(context.Background(), &recordUsageCoreInput{},
		&ForwardResult{Usage: ClaudeUsage{KiroCredits: 0.17}}, &APIKey{}, &User{}, &Account{Platform: PlatformAnthropic},
		nil, "claude-sonnet-4-5", 1, 1, 1, 1, false, &CostBreakdown{})
	require.Nil(t, log.KiroCredits)
}

func TestKiroChatCredits(t *testing.T) {
	svc := &GatewayService{}
	stream := strings.Replace(anthropicStreamWithMidStreamText(), `"usage":{"output_tokens":42}`, `"usage":{"output_tokens":42,"_sub2api_kiro_credits":0.17}`, 1)
	for _, route := range []string{"buffered", "streaming"} {
		t.Run(route, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
			response := &http.Response{Header: http.Header{}, Body: io.NopCloser(strings.NewReader(stream))}
			handle := svc.handleCCBufferedFromAnthropic
			if route == "streaming" {
				handle = svc.handleCCStreamingFromAnthropic
			}
			result, err := handle(response, ctx, "claude-sonnet-4-6", "claude-sonnet-4-6", nil, time.Now())
			require.NoError(t, err)
			require.InDelta(t, 0.17, result.Usage.KiroCredits, 1e-9)
			require.NotContains(t, recorder.Body.String(), "kiro_credits")
		})
	}
}
