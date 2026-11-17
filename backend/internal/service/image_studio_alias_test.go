package service

import (
	"bytes"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"net/http/httptest"
	"testing"
)

func TestImageStudioAliasRequiresPrivateModelResolution(t *testing.T) {
	body := []byte(`{"model":"nano-banana","prompt":"edit image"}`)
	svc := &OpenAIGatewayService{}
	for _, trusted := range []bool{false, true} {
		ctx := context.Background()
		if trusted {
			ctx = WithImageStudioModel(ctx, "nano-banana", "gemini-3-pro-image")
		}
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("POST", "/v1/images/generations", bytes.NewReader(body)).WithContext(ctx)
		c.Request.Header.Set("Content-Type", "application/json")
		parsed, err := svc.ParseOpenAIImagesRequest(c, body)
		if !trusted {
			require.Error(t, err)
			continue
		}
		require.NoError(t, err)
		require.Equal(t, "nano-banana", parsed.Model)
		require.Equal(t, OpenAIImagesCapabilityAPIKey, parsed.RequiredCapability)
	}
}

func TestImageStudioAliasesReachMappedImagesUpstream(t *testing.T) {
	for _, tc := range []struct {
		name, canonical string
		oauth           bool
	}{{"nano-banana", "gemini-3-pro-image", false}, {"studio-gpt", "gpt-image-2", true}} {
		t.Run(tc.name, func(t *testing.T) {
			body := []byte(`{"model":"` + tc.name + `","prompt":"draw"}`)
			ctx := WithImageStudioModel(context.Background(), tc.name, tc.canonical)
			c, _ := newOpenAIImagesTestContext(t, body)
			c.Request = c.Request.WithContext(ctx)
			upstream := &httpUpstreamRecorder{resp: openAIImagesJSONResponse()}
			svc := newOpenAIImagesTestService(upstream)
			account := newOpenAIImagesAPIKeyAccount()
			if tc.oauth {
				account = directImagesTestAccount()
			}
			account.Credentials["model_mapping"] = map[string]any{tc.name: tc.canonical}
			parsed, err := svc.ParseOpenAIImagesRequest(c, body)
			require.NoError(t, err)
			result, err := svc.ForwardImages(ctx, c, account, body, parsed, "")
			require.NoError(t, err)
			require.Equal(t, tc.canonical, result.UpstreamModel)
			require.Equal(t, tc.canonical, gjson.GetBytes(upstream.lastBody, "model").String())
			account.Credentials["model_mapping"] = map[string]any{tc.name: "gpt-4-text"}
			_, err = svc.ForwardImages(ctx, c, account, body, parsed, "")
			require.Error(t, err, "actual non-image target must remain forbidden")
			_, err = svc.ForwardImages(context.Background(), c, account, body, parsed, "")
			require.Error(t, err, "external alias must remain forbidden")
		})
	}
}
