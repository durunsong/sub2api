//go:build unit

package service

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func museVideoTestAccount() *Account {
	return &Account{ID: 19, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Credentials: map[string]any{
		"api_key": "muse-secret", "base_url": "https://muse.example/v1",
		"model_mapping": map[string]any{"muse-video": "muse-video"},
	}}
}

func TestMuseVideoRequestValidationAndCapability(t *testing.T) {
	account := museVideoTestAccount()
	require.True(t, account.SupportsOpenAIEndpointCapability(OpenAIEndpointCapabilityMuseVideo))
	delete(account.Credentials, "model_mapping")
	require.False(t, account.SupportsOpenAIEndpointCapability(OpenAIEndpointCapabilityMuseVideo))
	for _, raw := range []string{`{}`, `{"model":"muse-video","prompt":"x","last_frame":{"url":"https://x"}}`, `{"model":"muse-video","prompt":"x","images":[{"url":""}]}`, `{"model":"muse-video","prompt":"x","videos":["https://x"]}`, `{"model":"muse-video","prompt":"x","duration":30,"resolution":"720p"}`} {
		_, err := ParseMuseVideoRequest([]byte(raw))
		require.Error(t, err, raw)
	}
	account.Credentials["model_mapping"] = map[string]any{"public-muse-video": "muse-video"}
	require.True(t, account.SupportsOpenAIEndpointCapability(OpenAIEndpointCapabilityMuseVideo))
	_, err := ParseMuseVideoRequest([]byte(`{"model":"public-muse-video","prompt":"waves","duration":6,"resolution":"720p"}`))
	require.NoError(t, err)
	info, err := ParseMuseVideoRequest([]byte(`{"model":"muse-video","prompt":"waves","duration":6,"resolution":"720p","aspect_ratio":"9:16","images":[{"url":"https://example.com/a.png"},{"url":"https://example.com/b.png"}]}`))
	require.NoError(t, err)
	require.Equal(t, []string{"https://example.com/a.png", "https://example.com/b.png"}, info.InputImageURLs)
	require.Equal(t, 6, info.DurationSeconds)
	require.Contains(t, string(info.ModerationBody()), "https://example.com/b.png")
}

func TestMuseImageReferenceValidation(t *testing.T) {
	for _, value := range []string{"aW1hZ2U=", "data:image/png;base64,%%%", "data:text/plain;base64,YQ==", "https:///missing-host", "ftp://example.com/x.png", "HTTP://example.com/x.png"} {
		require.Error(t, ValidateMuseImageReference(value), value)
	}
	for _, value := range []string{"https://example.com/x.png", "data:image/png;base64,YQ=="} {
		require.NoError(t, ValidateMuseImageReference(value), value)
	}
	_, err := ParseMuseVideoRequest([]byte(`{"model":"muse-video","prompt":"waves","duration":6,"resolution":"720p","images":[{"url":"aW1hZ2U="}]}`))
	require.Error(t, err)
}

func TestMuseVideoForwardCreateAndStatus(t *testing.T) {
	upstream := &grokMediaContentUpstreamStub{response: grokMediaContentStatusResponse(`{"id":"task_123","status":"queued"}`)}
	svc := &OpenAIGatewayService{httpUpstream: upstream}
	c, w := grokMediaContentTestContext(http.MethodPost, "/v1/videos", nil)
	result, err := svc.ForwardMuseVideo(context.Background(), c, museVideoTestAccount(), MuseVideoEndpointCreate, "", []byte(`{"model":"muse-video","prompt":"waves","duration":6,"resolution":"720p","images":[{"url":"https://example.com/a.png"}]}`))
	require.NoError(t, err)
	require.Equal(t, "muse_task_123", result.ResponseID)
	require.Equal(t, "https://muse.example/v1/videos", upstream.request.URL.String())
	require.Equal(t, "Bearer muse-secret", upstream.request.Header.Get("Authorization"))
	forwarded, err := io.ReadAll(upstream.request.Body)
	require.NoError(t, err)
	require.Equal(t, "https://example.com/a.png", gjson.GetBytes(forwarded, "images.0.url").String())
	require.Equal(t, "muse_task_123", gjson.Get(w.Body.String(), "id").String())

	upstream = &grokMediaContentUpstreamStub{response: grokMediaContentStatusResponse(`{"id":"task_123","status":"completed","prompt":"private","result":{"url":"/v1/media/video.mp4","filename":"secret.mp4"}}`)}
	svc = &OpenAIGatewayService{httpUpstream: upstream}
	c, w = grokMediaContentTestContext(http.MethodGet, "/v1/videos/muse_task_123", nil)
	result, err = svc.ForwardMuseVideo(context.Background(), c, museVideoTestAccount(), MuseVideoEndpointStatus, "muse_task_123", nil)
	require.NoError(t, err)
	require.Equal(t, 1, result.VideoCount)
	require.Equal(t, "muse_task_123", gjson.Get(w.Body.String(), "id").String())
	require.Equal(t, "/v1/videos/muse_task_123/content", gjson.Get(w.Body.String(), "result.url").String())
	require.False(t, strings.Contains(w.Body.String(), "private"))
	require.False(t, strings.Contains(w.Body.String(), "filename"))
	require.Equal(t, "/v1/videos/task_123", upstream.request.URL.Path)
}

func TestMuseVideoRejectsUnsafeContentURL(t *testing.T) {
	for _, raw := range []string{"https://evil.example/v1/media/x.mp4", "http://muse.example/v1/media/x.mp4", "/v1/media/../secret", "//evil.example/v1/media/x.mp4"} {
		_, err := museVideoContentURL("https://muse.example/v1", raw)
		require.Error(t, err, raw)
	}
	url, err := museVideoContentURL("https://muse.example/v1", "/v1/media/video.mp4")
	require.NoError(t, err)
	require.Equal(t, "https://muse.example/v1/media/video.mp4", url)
}

func TestMuseVideoStatusRejectsOtherTask(t *testing.T) {
	_, _, _, err := museVideoStatus([]byte(`{"id":"task_other","status":"completed","result":{"url":"/v1/media/video.mp4"}}`), "muse_task_123", "/v1/videos/muse_task_123/content")
	require.Error(t, err)
}

func TestMuseVideoContentKeepsCredentialFromPublicMedia(t *testing.T) {
	upstream := &grokMediaContentUpstreamStub{responses: []*http.Response{
		grokMediaContentStatusResponse(`{"id":"task_123","status":"completed","result":{"url":"/v1/media/video.mp4"}}`),
		{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": []string{"video/mp4"}}, Body: io.NopCloser(strings.NewReader("video")), ContentLength: -1},
	}}
	svc := &OpenAIGatewayService{httpUpstream: upstream}
	c, w := grokMediaContentTestContext(http.MethodGet, "/v1/videos/muse_task_123/content", nil)
	result, err := svc.ForwardMuseVideo(context.Background(), c, museVideoTestAccount(), MuseVideoEndpointContent, "muse_task_123", nil)
	require.NoError(t, err)
	require.Equal(t, "video", w.Body.String())
	require.Equal(t, 1, result.VideoCount)
	require.Equal(t, "Bearer muse-secret", upstream.requests[0].Header.Get("Authorization"))
	require.Empty(t, upstream.requests[1].Header.Get("Authorization"))
	require.Equal(t, "https://muse.example/v1/media/video.mp4", upstream.requests[1].URL.String())
}
