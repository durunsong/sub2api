package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

const (
	MuseVideoEndpointCreate  GrokMediaEndpoint = "muse_video_create"
	MuseVideoEndpointStatus  GrokMediaEndpoint = "muse_video_status"
	MuseVideoEndpointContent GrokMediaEndpoint = "muse_video_content"
	museVideoContentMaxBytes                   = 128 << 20
)

func (e GrokMediaEndpoint) IsMuseVideo() bool {
	return e == MuseVideoEndpointCreate || e == MuseVideoEndpointStatus || e == MuseVideoEndpointContent
}

func MuseVideoTaskKey(id string) string { return "muse_" + strings.TrimSpace(id) }

func museVideoUpstreamID(id string) (string, error) {
	if !strings.HasPrefix(strings.TrimSpace(id), "muse_") {
		return "", fmt.Errorf("invalid Muse task ID")
	}
	id = strings.TrimPrefix(strings.TrimSpace(id), "muse_")
	if !strings.HasPrefix(id, "task_") || len(id) <= len("task_") || validateUpstreamPathSegment("Muse task ID", id) != nil {
		return "", fmt.Errorf("invalid Muse task ID")
	}
	return id, nil
}

func museVideoTaskIDFromUpstream(id string) (string, error) {
	if !strings.HasPrefix(id, "task_") || len(id) <= len("task_") || validateUpstreamPathSegment("Muse task ID", id) != nil {
		return "", fmt.Errorf("invalid Muse task ID")
	}
	return MuseVideoTaskKey(id), nil
}

// ValidateMuseImageReference keeps every upstream-accepted reference visible to moderation.
func ValidateMuseImageReference(value string) error {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > 16<<20 {
		return fmt.Errorf("invalid Muse image reference")
	}
	if strings.HasPrefix(value, "data:") {
		if err := validateAccountTestDataURL(value, "image/"); err == nil {
			return nil
		}
		return fmt.Errorf("invalid Muse image data URL")
	}
	if !strings.HasPrefix(value, "https://") && !strings.HasPrefix(value, "http://") {
		return fmt.Errorf("Muse image reference must be an HTTP image URL")
	}
	parsed, err := url.Parse(value)
	if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Hostname() == "" || parsed.User != nil {
		return fmt.Errorf("Muse image reference must be an HTTP image URL")
	}
	return nil
}

func ParseMuseVideoRequest(body []byte) (GrokMediaRequestInfo, error) {
	var info GrokMediaRequestInfo
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil || raw == nil {
		return info, fmt.Errorf("request body must be a JSON object")
	}
	for key := range raw {
		switch key {
		case "model", "prompt", "duration", "resolution", "size", "aspect_ratio", "images":
		default:
			return info, fmt.Errorf("unsupported Muse video parameter: %s", key)
		}
	}
	if rawImages, ok := raw["images"]; ok {
		var images []map[string]json.RawMessage
		if err := json.Unmarshal(rawImages, &images); err != nil || images == nil {
			return info, fmt.Errorf("images must be an array of URL objects")
		}
		for _, image := range images {
			if len(image) != 1 || image["url"] == nil {
				return info, fmt.Errorf("images must contain only URL objects")
			}
		}
	}
	var request struct {
		Model       string `json:"model"`
		Prompt      string `json:"prompt"`
		Duration    int    `json:"duration"`
		Resolution  string `json:"resolution"`
		Size        string `json:"size"`
		AspectRatio string `json:"aspect_ratio"`
		Images      []struct {
			URL string `json:"url"`
		} `json:"images"`
	}
	if err := json.Unmarshal(body, &request); err != nil {
		return info, fmt.Errorf("invalid Muse video request")
	}
	info.Model, info.Prompt = strings.TrimSpace(request.Model), strings.TrimSpace(request.Prompt)
	if info.Model == "" || info.Prompt == "" {
		return info, fmt.Errorf("model and prompt are required")
	}
	if request.Duration <= 0 || request.Duration > VideoBillingMaxDurationSeconds {
		return info, fmt.Errorf("duration must be between 1 and 15 seconds")
	}
	if tier, ok := LookupVideoBillingResolution(request.Resolution); !ok {
		return info, fmt.Errorf("resolution must be 480p, 720p or 1080p")
	} else {
		info.Resolution = tier
	}
	info.DurationSeconds = request.Duration
	info.Size, info.AspectRatio = strings.TrimSpace(request.Size), strings.TrimSpace(request.AspectRatio)
	if info.Size != "" && info.Size != "auto" && info.Size != "16:9" && info.Size != "9:16" && info.Size != "1280x720" && info.Size != "720x1280" && info.Size != "1920x1080" && info.Size != "1080x1920" && info.Size != "854x480" && info.Size != "480x854" {
		return info, fmt.Errorf("unsupported size")
	}
	if info.AspectRatio != "" && info.AspectRatio != "16:9" && info.AspectRatio != "9:16" && info.AspectRatio != "auto" {
		return info, fmt.Errorf("unsupported aspect_ratio")
	}
	if len(request.Images) > 10 {
		return info, fmt.Errorf("too many reference images")
	}
	for _, image := range request.Images {
		imageURL := strings.TrimSpace(image.URL)
		if err := ValidateMuseImageReference(imageURL); err != nil {
			return info, err
		}
		info.InputImageURLs = append(info.InputImageURLs, imageURL)
	}
	return info, nil
}

func museVideoURL(base string, endpoint GrokMediaEndpoint, id string) (string, error) {
	base = strings.TrimRight(base, "/")
	if !strings.HasSuffix(base, "/v1") {
		base += "/v1"
	}
	base += "/videos"
	if endpoint != MuseVideoEndpointCreate {
		upstreamID, err := museVideoUpstreamID(id)
		if err != nil {
			return "", err
		}
		base += "/" + upstreamID
	}
	return base, nil
}

func museVideoContentURL(base, raw string) (string, error) {
	provider, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("invalid Muse provider URL")
	}
	media, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || media == nil || media.User != nil || media.Fragment != "" || media.RawQuery != "" {
		return "", fmt.Errorf("invalid Muse media URL")
	}
	if media.IsAbs() && (media.Scheme != provider.Scheme || !strings.EqualFold(media.Host, provider.Host)) {
		return "", fmt.Errorf("Muse media origin mismatch")
	}
	if !media.IsAbs() && (media.Host != "" || !strings.HasPrefix(media.Path, "/")) {
		return "", fmt.Errorf("invalid Muse media path")
	}
	if provider.Scheme != "https" && provider.Scheme != "http" {
		return "", fmt.Errorf("invalid Muse provider scheme")
	}
	path := media.EscapedPath()
	if !strings.HasPrefix(path, "/v1/media/") || strings.Contains(path, "%") || strings.Contains(path, "\\") {
		return "", fmt.Errorf("invalid Muse media path")
	}
	name := strings.TrimPrefix(path, "/v1/media/")
	if name == "" || strings.Contains(name, "/") || strings.Contains(name, "..") {
		return "", fmt.Errorf("invalid Muse media name")
	}
	return provider.Scheme + "://" + provider.Host + path, nil
}

func museVideoStatus(body []byte, taskID, proxyURL string) ([]byte, bool, string, error) {
	if !gjson.ValidBytes(body) || !gjson.ParseBytes(body).IsObject() {
		return nil, false, "", fmt.Errorf("invalid Muse status response")
	}
	upstreamID, err := museVideoUpstreamID(taskID)
	if err != nil || gjson.GetBytes(body, "id").String() != upstreamID {
		return nil, false, "", fmt.Errorf("Muse task ID mismatch")
	}
	status := strings.ToLower(strings.TrimSpace(gjson.GetBytes(body, "status").String()))
	switch status {
	case "queued", "processing", "failed", "completed", "succeeded", "success", "done":
	default:
		return nil, false, "", fmt.Errorf("invalid Muse task status")
	}
	upstreamURL := strings.TrimSpace(gjson.GetBytes(body, "result.url").String())
	if upstreamURL == "" {
		upstreamURL = strings.TrimSpace(gjson.GetBytes(body, "url").String())
	}
	completed := status == "completed" || status == "succeeded" || status == "success" || status == "done"
	if completed && upstreamURL == "" {
		return nil, false, "", fmt.Errorf("Muse task has no result URL")
	}
	if completed {
		status = "completed"
	}
	out := map[string]any{"id": taskID, "task_id": taskID, "object": "video.task", "status": status}
	if completed {
		out["progress"] = 100
		out["url"] = proxyURL
		out["video"] = map[string]string{"url": proxyURL}
		out["result"] = map[string]string{"url": proxyURL}
	}
	clean, err := json.Marshal(out)
	return clean, completed, upstreamURL, err
}

func (s *OpenAIGatewayService) ForwardMuseVideo(ctx context.Context, c *gin.Context, account *Account, endpoint GrokMediaEndpoint, taskID string, body []byte) (*OpenAIForwardResult, error) {
	if account == nil || !account.SupportsOpenAIEndpointCapability(OpenAIEndpointCapabilityMuseVideo) || !endpoint.IsMuseVideo() {
		return nil, fmt.Errorf("Muse video requires a mapped OpenAI API key account")
	}
	base, err := s.validateUpstreamBaseURL(account.GetCredential("base_url"))
	if err != nil {
		return nil, err
	}
	target, err := museVideoURL(base, endpoint, taskID)
	if err != nil {
		return nil, err
	}
	info := GrokMediaRequestInfo{}
	if endpoint == MuseVideoEndpointCreate {
		info, err = ParseMuseVideoRequest(body)
		if err != nil {
			return nil, &GrokMediaInputValidationError{Reason: err.Error()}
		}
		mapped, matched := account.ResolveMappedModel(info.Model)
		if !matched || mapped != "muse-video" {
			return nil, &GrokMediaInputValidationError{Reason: "Muse video model requires an explicit account mapping"}
		}
		var request map[string]any
		if err = json.Unmarshal(body, &request); err != nil {
			return nil, err
		}
		request["model"] = mapped
		body, err = json.Marshal(request)
		if err != nil {
			return nil, err
		}
	}
	proxy := ""
	if account.ProxyID != nil && account.Proxy != nil {
		proxy = account.Proxy.URL()
	}
	token := account.GetCredential("api_key")
	request := func(method, rawURL string, body []byte, accept string, authenticated bool) (*http.Response, error) {
		req, err := http.NewRequestWithContext(WithHTTPUpstreamRedirectsDisabled(ctx), method, rawURL, bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		if authenticated {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		req.Header.Set("Accept", accept)
		if method == http.MethodPost {
			req.Header.Set("Content-Type", "application/json")
		}
		if method == http.MethodGet && accept == "*/*" && c != nil {
			if rangeHeader := strings.TrimSpace(c.GetHeader("Range")); rangeHeader != "" {
				req.Header.Set("Range", rangeHeader)
			}
		}
		if authenticated {
			account.ApplyHeaderOverrides(req.Context(), req.Header)
		}
		return s.httpUpstream.Do(req, proxy, account.ID, account.Concurrency)
	}
	started := time.Now()
	method := http.MethodGet
	if endpoint == MuseVideoEndpointCreate {
		method = http.MethodPost
	}
	resp, err := request(method, target, body, "application/json", true)
	SetOpsLatencyMs(c, OpsUpstreamLatencyMsKey, time.Since(started).Milliseconds())
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	responseBody, err := ReadUpstreamResponseBody(resp.Body, s.cfg, c, openAITooLargeError)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("Muse upstream rejected request")
	}
	result := &OpenAIForwardResult{Model: info.Model, BillingModel: "muse-video", UpstreamModel: "muse-video", Duration: time.Since(started), VideoDurationSeconds: info.DurationSeconds, VideoResolution: info.Resolution}
	if endpoint == MuseVideoEndpointCreate {
		id := strings.TrimSpace(gjson.GetBytes(responseBody, "id").String())
		publicID, err := museVideoTaskIDFromUpstream(id)
		if err != nil {
			return nil, fmt.Errorf("Muse create response missing task ID")
		}
		result.ResponseID = publicID
		clean, _ := json.Marshal(map[string]any{"id": result.ResponseID, "task_id": result.ResponseID, "object": "video.task", "status": "queued"})
		c.Data(resp.StatusCode, "application/json", clean)
		return result, nil
	}
	proxyURL := grokMediaContentProxyURL(c, taskID)
	clean, completed, mediaURL, err := museVideoStatus(responseBody, taskID, proxyURL)
	if err != nil {
		return nil, err
	}
	result.ResponseID = taskID
	if completed {
		contentURL, err := museVideoContentURL(base, mediaURL)
		if err != nil {
			return nil, err
		}
		result.VideoCount = 1
		if endpoint == MuseVideoEndpointContent {
			contentResp, err := request(http.MethodGet, contentURL, nil, "*/*", false)
			if err != nil {
				return nil, err
			}
			defer func() { _ = contentResp.Body.Close() }()
			if contentResp.StatusCode != http.StatusOK && contentResp.StatusCode != http.StatusPartialContent {
				return nil, fmt.Errorf("Muse media download failed")
			}
			if contentResp.ContentLength > museVideoContentMaxBytes {
				return nil, fmt.Errorf("Muse media exceeds proxy limit")
			}
			mediaBytes, err := io.ReadAll(io.LimitReader(contentResp.Body, museVideoContentMaxBytes+1))
			if err != nil || len(mediaBytes) > museVideoContentMaxBytes {
				return nil, fmt.Errorf("Muse media exceeds proxy limit")
			}
			limited := *contentResp
			limited.Body = io.NopCloser(bytes.NewReader(mediaBytes))
			limited.ContentLength = int64(len(mediaBytes))
			if err := writeGrokMediaContentResponse(c, &limited); err != nil {
				return nil, err
			}
			return result, nil
		}
	}
	if endpoint == MuseVideoEndpointContent {
		return nil, fmt.Errorf("Muse video is not complete")
	}
	c.Data(resp.StatusCode, "application/json", clean)
	return result, nil
}
