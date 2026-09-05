//go:build unit

package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestForwardKiroMessagesPreservesUpstreamRequestID(t *testing.T) {
	for _, stream := range []bool{false, true} {
		for _, header := range []string{"X-Request-Id", "X-Amzn-Requestid", "X-Custom-Trace"} {
			t.Run(fmt.Sprintf("stream=%t/header=%s", stream, header), func(t *testing.T) {
				body := []byte(fmt.Sprintf(`{"model":"claude-sonnet-4-6","stream":%t,"messages":[{"role":"user","content":"hello"}]}`, stream))
				parsed, err := ParseGatewayRequest(NewRequestBodyRef(body), PlatformAnthropic)
				require.NoError(t, err)
				responseHeaders := make(http.Header)
				responseHeaders.Set(header, "direct-upstream-id")
				upstream := &queuedHTTPUpstream{responses: []*http.Response{{
					StatusCode: http.StatusOK, Header: responseHeaders,
					Body: io.NopCloser(strings.NewReader("")),
				}}}
				svc := &GatewayService{
					cfg: &config.Config{}, httpUpstream: upstream,
					kiroCooldownStore:   &stubKiroCooldownStore{},
					tlsFPProfileService: &TLSFingerprintProfileService{},
				}
				account := &Account{
					ID: 42, Platform: PlatformKiro, Type: AccountTypeAPIKey,
					Credentials: map[string]any{"api_key": "test-kiro-key"},
					Extra:       map[string]any{AccountExtraUpstreamRequestIDHeader: header},
				}
				rec := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(rec)
				c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewReader(body))
				result, err := svc.forwardKiroMessages(context.Background(), c, account, parsed, time.Now())
				require.NoError(t, err)
				require.NotNil(t, result)
				id := usageUpstreamRequestIDPtr(account, result.UpstreamHeaders, false)
				require.NotNil(t, id)
				require.Equal(t, "direct-upstream-id", *id)
				require.Equal(t, "direct-upstream-id", responseHeaders.Get(header))
			})
		}
	}
}
