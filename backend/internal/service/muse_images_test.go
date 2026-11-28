//go:build unit

package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestMuseImagesAPIKeyOnlyAndNoResponsesConversion(t *testing.T) {
	for _, model := range []string{"muse-image", "muse-image-1.0"} {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		body := []byte(`{"model":"` + model + `","prompt":"draw","response_format":"b64_json"}`)
		c.Request = httptest.NewRequest("POST", "/v1/images/generations", bytes.NewReader(body))
		parsed, err := (&OpenAIGatewayService{}).ParseOpenAIImagesRequest(c, body)
		require.NoError(t, err)
		require.Equal(t, OpenAIImagesCapabilityAPIKey, parsed.RequiredCapability)
		require.Equal(t, OpenAIImagesCapabilityAPIKey, (&OpenAIImagesRequest{}).RequiredCapabilityForModel(model))
		require.False(t, isOpenAIImageGenerationModel(model))
		for _, typ := range []string{AccountTypeOAuth, AccountTypeSetupToken} {
			require.False(t, (&Account{Platform: PlatformOpenAI, Type: typ}).SupportsOpenAIImageCapability(parsed.RequiredCapability))
		}
	}
}

func TestMuseImagesForwardAndExplicitPrice(t *testing.T) {
	for _, priced := range []bool{false, true} {
		t.Run(map[bool]string{false: "missing_price", true: "configured"}[priced], func(t *testing.T) {
			body := []byte(`{"model":"muse-image","prompt":"draw","n":1,"response_format":"b64_json"}`)
			c, w := gin.CreateTestContext(httptest.NewRecorder())
			_ = w
			c.Request = httptest.NewRequest("POST", "/v1/images/generations", bytes.NewReader(body))
			c.Request.Header.Set("Content-Type", "application/json")
			group := &Group{ImageCountBilling: true}
			if priced {
				price := 0.2
				group.ImagePrice1K = &price
				group.ImagePrice2K = &price
				group.ImagePrice4K = &price
			}
			c.Set("api_key", &APIKey{Group: group})
			var pngBytes bytes.Buffer
			require.NoError(t, png.Encode(&pngBytes, image.NewRGBA(image.Rect(0, 0, 2, 3))))
			encoded := base64.StdEncoding.EncodeToString(pngBytes.Bytes())
			upstream := &httpUpstreamRecorder{resp: &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(`{"data":[{"b64_json":"` + encoded + `","size":"4096x4096"}]}`))}}
			svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}
			parsed, err := svc.ParseOpenAIImagesRequest(c, body)
			require.NoError(t, err)
			account := &Account{ID: 1, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Credentials: map[string]any{"api_key": "fixture-key", "base_url": "https://muse.example/v1", "model_mapping": map[string]any{"muse-image": "muse-image"}}}
			result, err := svc.ForwardImages(context.Background(), c, account, body, parsed, "")
			if !priced {
				require.ErrorContains(t, err, "price")
				require.Nil(t, upstream.lastReq)
				return
			}
			require.NoError(t, err)
			require.Equal(t, 1, result.ImageCount)
			require.Equal(t, []string{"2x3"}, result.ImageOutputSizes)
			require.Equal(t, "muse-image", gjson.GetBytes(upstream.lastBody, "model").String())
			require.Equal(t, "https://muse.example/v1/images/generations", upstream.lastReq.URL.String())
			require.True(t, HTTPUpstreamRedirectsDisabled(upstream.lastReq.Context()))
		})
	}
}

func TestMuseImagesRejectUnsupportedAndUnmoderatedInputs(t *testing.T) {
	for _, field := range []string{`"n":2`, `"stream":true`, `"mask":{}`, `"quality":"high"`, `"async":true`, `"extra":"hidden prompt"`, `"prompt":"second prompt"`, `"images":[{"url":"https://example.com/ref.png"}]`, `"videos":["https://example.com/ref.mp4"]`} {
		body := []byte(`{"model":"muse-image","prompt":"draw",` + field + `}`)
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("POST", "/v1/images/generations", bytes.NewReader(body))
		_, err := (&OpenAIGatewayService{}).ParseOpenAIImagesRequest(c, body)
		require.Error(t, err, field)
	}
}

func TestMuseImagesReferencesReachModeration(t *testing.T) {
	for _, upload := range []bool{false, true} {
		body := []byte(`{"model":"muse-image","prompt":"edit","images":[{"image_url":"https://example.com/ref.png"}]}`)
		contentType := "application/json"
		if upload {
			var b bytes.Buffer
			w := multipart.NewWriter(&b)
			require.NoError(t, w.WriteField("model", "muse-image"))
			require.NoError(t, w.WriteField("prompt", "edit"))
			part, err := w.CreateFormFile("image[]", "reference.png")
			require.NoError(t, err)
			_, err = part.Write(b64BackfillPNGBytes)
			require.NoError(t, err)
			require.NoError(t, w.Close())
			body = b.Bytes()
			contentType = w.FormDataContentType()
		}
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("POST", "/v1/images/edits", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", contentType)
		parsed, err := (&OpenAIGatewayService{}).ParseOpenAIImagesRequest(c, body)
		require.NoError(t, err)
		require.NotEmpty(t, parsed.ModerationBody())
		if upload {
			require.Len(t, parsed.Uploads, 1)
			require.Contains(t, string(parsed.ModerationBody()), "data:image/png;base64,")
		} else {
			require.Equal(t, []string{"https://example.com/ref.png"}, parsed.InputImageURLs)
		}
	}
}

func TestMuseImagesAliasUploadReachesModerationBeforeMapping(t *testing.T) {
	var pngBytes bytes.Buffer
	require.NoError(t, png.Encode(&pngBytes, image.NewRGBA(image.Rect(0, 0, 2, 3))))
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	require.NoError(t, w.WriteField("model", "gpt-image-custom"))
	require.NoError(t, w.WriteField("prompt", "edit"))
	part, err := w.CreateFormFile("image[]", "reference.png")
	require.NoError(t, err)
	_, err = part.Write(pngBytes.Bytes())
	require.NoError(t, err)
	require.NoError(t, w.Close())
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/v1/images/edits", bytes.NewReader(b.Bytes()))
	c.Request.Header.Set("Content-Type", w.FormDataContentType())
	parsed, err := (&OpenAIGatewayService{}).ParseOpenAIImagesRequest(c, b.Bytes())
	require.NoError(t, err)
	input := ExtractContentModerationInput(ContentModerationProtocolOpenAIImages, parsed.ModerationBody())
	require.Equal(t, []string{"data:image/png;base64," + base64.StdEncoding.EncodeToString(pngBytes.Bytes())}, input.Images)
}

func TestMuseImagesStudioCompositeChainedAlias(t *testing.T) {
	var pngBytes bytes.Buffer
	require.NoError(t, png.Encode(&pngBytes, image.NewRGBA(image.Rect(0, 0, 2, 3))))
	encoded := base64.StdEncoding.EncodeToString(pngBytes.Bytes())
	for _, upload := range []bool{false, true} {
		body := []byte(`{"model":"provider-image","prompt":"draw"}`)
		path, contentType := "/v1/images/generations", "application/json"
		if upload {
			var b bytes.Buffer
			w := multipart.NewWriter(&b)
			require.NoError(t, w.WriteField("model", "studio-image"))
			require.NoError(t, w.WriteField("prompt", "edit"))
			part, err := w.CreateFormFile("image[]", "reference.png")
			require.NoError(t, err)
			_, err = part.Write(pngBytes.Bytes())
			require.NoError(t, err)
			require.NoError(t, w.Close())
			body, path, contentType = b.Bytes(), "/v1/images/edits", w.FormDataContentType()
		}
		for _, public := range []string{"studio-image", "unrelated-image"} {
			ctx := WithImageStudioModel(context.Background(), "studio-image", "muse-image")
			ctx = WithCompositeRouteDecision(ctx, CompositeRouteDecision{Matched: true, PublicModel: public, TargetPlatform: PlatformOpenAI, UpstreamModel: "provider-image"})
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest("POST", path, bytes.NewReader(body)).WithContext(ctx)
			c.Request.Header.Set("Content-Type", contentType)
			parsed, err := (&OpenAIGatewayService{}).ParseOpenAIImagesRequest(c, body)
			if public != "studio-image" {
				require.Error(t, err)
				continue
			}
			require.NoError(t, err)
			require.Equal(t, "provider-image", parsed.Model)
			require.Equal(t, OpenAIImagesCapabilityAPIKey, parsed.RequiredCapability)
			if upload {
				require.Contains(t, string(parsed.ModerationBody()), "data:image/png;base64,")
			}
			price := 0.1
			c.Set("api_key", &APIKey{Group: &Group{ImageCountBilling: true, ImagePrice1K: &price, ImagePrice2K: &price, ImagePrice4K: &price}})
			upstream := &httpUpstreamRecorder{resp: &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(`{"data":[{"b64_json":"` + encoded + `"}]}`))}}
			svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}
			account := &Account{ID: 1, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Credentials: map[string]any{"api_key": "fixture-key", "base_url": "https://muse.example/v1", "model_mapping": map[string]any{"provider-image": "muse-image"}}}
			result, err := svc.ForwardImages(ctx, c, account, body, parsed, "")
			require.NoError(t, err)
			require.Equal(t, 1, result.ImageCount)
			require.Contains(t, string(upstream.lastBody), "muse-image")
			require.NotContains(t, string(upstream.lastBody), "provider-image")
			require.NotContains(t, string(upstream.lastBody), "studio-image")
		}
	}
}

func TestMuseImagesRejectMixedMultipartReference(t *testing.T) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	require.NoError(t, w.WriteField("model", "muse-image"))
	require.NoError(t, w.WriteField("prompt", "edit"))
	part, err := w.CreateFormFile("image[]", "reference.png")
	require.NoError(t, err)
	_, err = part.Write(b64BackfillPNGBytes)
	require.NoError(t, err)
	require.NoError(t, w.WriteField("image[]", "https://example.com/unmoderated.png"))
	require.NoError(t, w.Close())
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/v1/images/edits", bytes.NewReader(b.Bytes()))
	c.Request.Header.Set("Content-Type", w.FormDataContentType())
	_, err = (&OpenAIGatewayService{}).ParseOpenAIImagesRequest(c, b.Bytes())
	require.Error(t, err)
}

func TestMuseImagesServerFailureIsNotRetriedOrEchoed(t *testing.T) {
	price := 0.1
	body := []byte(`{"model":"muse-image","prompt":"private"}`)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest("POST", "/v1/images/generations", bytes.NewReader(body))
	c.Set("api_key", &APIKey{Group: &Group{ImageCountBilling: true, ImagePrice1K: &price, ImagePrice2K: &price, ImagePrice4K: &price}})
	upstream := &httpUpstreamRecorder{resp: &http.Response{StatusCode: 500, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(`{"error":"private prompt"}`))}}
	svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}
	parsed, err := svc.ParseOpenAIImagesRequest(c, body)
	require.NoError(t, err)
	_, err = svc.ForwardImages(c.Request.Context(), c, &Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Credentials: map[string]any{"base_url": "https://muse.example/v1", "api_key": "test", "model_mapping": map[string]any{"muse-image": "muse-image"}}}, body, parsed, "")
	require.Error(t, err)
	var failover *UpstreamFailoverError
	require.NotErrorAs(t, err, &failover)
	require.NotContains(t, rec.Body.String(), "private")
}

func TestMuseImagesURLDimensionsAndDownloadFailure(t *testing.T) {
	var pngBytes bytes.Buffer
	require.NoError(t, png.Encode(&pngBytes, image.NewRGBA(image.Rect(0, 0, 2, 3))))
	for _, status := range []int{200, 503} {
		upstream := &httpUpstreamRecorder{resp: b64BackfillImageResponse(status, "image/png", pngBytes.Bytes())}
		svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}
		body, err := svc.normalizeMuseImageResponse(context.Background(), b64BackfillAccount(false), []byte(`{"data":[{"url":"https://cdn.example.com/a.png","size":"4096x4096"}]}`))
		if status == 200 {
			require.NoError(t, err)
			require.Equal(t, "2x3", gjson.GetBytes(body, "data.0.size").String())
			require.NotEmpty(t, gjson.GetBytes(body, "data.0.b64_json").String())
		} else {
			require.ErrorContains(t, err, "dimensions could not be verified")
			require.Nil(t, body)
		}
		require.Empty(t, upstream.lastReq.Header.Get("Authorization"))
		require.True(t, HTTPUpstreamRedirectsDisabled(upstream.lastReq.Context()))
	}
}

func TestMuseImagesRejectUnmoderatableJSONReference(t *testing.T) {
	for _, reference := range []string{"aGVsbG8=", "file:///tmp/image.png", "data:video/mp4;base64,aGVsbG8="} {
		body := []byte(`{"model":"muse-image","prompt":"edit","images":[{"image_url":"` + reference + `"}]}`)
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("POST", "/v1/images/edits", bytes.NewReader(body))
		_, err := (&OpenAIGatewayService{}).ParseOpenAIImagesRequest(c, body)
		require.Error(t, err, reference)
	}
}
