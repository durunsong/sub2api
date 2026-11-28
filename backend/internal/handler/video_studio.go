package handler

// Video parameter mapping follows basketikun/infinite-canvas xai.ts at
// 4c67f1ece4f4edaef81e58e964eb78abb1e9809d (MIT; see docs/VIDEO_STUDIO_UPSTREAM.md).
import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	pkghttputil "github.com/Wei-Shaw/sub2api/internal/pkg/httputil"
	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type videoStudioModel struct {
	ID        string `json:"id"`
	GroupID   int64  `json:"group_id"`
	GroupName string `json:"group_name"`
	Model     string `json:"model"`
	Label     string `json:"label"`
	Protocol  string `json:"protocol"`
}
type videoStudioCatalog struct {
	Enabled bool               `json:"enabled"`
	Models  []videoStudioModel `json:"models"`
}
type videoStudioRequest struct {
	ModelID    string   `json:"model_id"`
	Prompt     string   `json:"prompt"`
	Seconds    int      `json:"seconds"`
	Resolution string   `json:"resolution"`
	Size       string   `json:"size"`
	Mode       string   `json:"mode"`
	References []string `json:"references"`
}
type videoStudioStatus struct {
	ID       string  `json:"id"`
	Status   string  `json:"status"`
	Progress float64 `json:"progress,omitempty"`
}

var videoOutputBudget = pkghttputil.NewBodyMemoryBudget(512 << 20)

var videoTaskID = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]{0,199}$`)

func buildVideoStudioRequest(model string, v videoStudioRequest) ([]byte, error) {
	return buildVideoStudioProtocolRequest(model, videoStudioProtocol(model), v)
}

func videoStudioProtocol(model string) string {
	m := strings.ToLower(strings.TrimSpace(model))
	if m == "muse-video" {
		return "muse"
	}
	if strings.HasPrefix(m, "grok-") && strings.Contains(m, "video") {
		return "grok"
	}
	return ""
}

func buildVideoStudioProtocolRequest(model, protocol string, v videoStudioRequest) ([]byte, error) {
	if strings.TrimSpace(v.Prompt) == "" || len(v.Prompt) > 32000 || v.Seconds < 4 || v.Seconds > 30 || v.Mode != "frames" && v.Mode != "reference" || len(v.References) > 9 || v.Mode == "frames" && len(v.References) > 2 {
		return nil, errors.New("invalid video parameters")
	}
	resolution, err := strconv.Atoi(strings.TrimSuffix(v.Resolution, "p"))
	if err != nil || resolution < 1 || resolution > 4320 {
		return nil, errors.New("invalid video resolution")
	}
	if protocol == "muse" && (v.Mode != "reference" || v.Seconds > 15 || resolution != 480 && resolution != 720 && resolution != 1080) {
		return nil, errors.New("Muse requires reference mode, 4-15 seconds and 480p, 720p or 1080p resolution")
	}
	ratio := "16:9"
	if v.Size != "auto" {
		var w, h int
		if n, err := fmt.Sscanf(v.Size, "%dx%d", &w, &h); err != nil || n != 2 || w < 1 || h < 1 || w > 8192 || h > 8192 || fmt.Sprintf("%dx%d", w, h) != v.Size {
			return nil, errors.New("invalid video size")
		}
		// Match upstream inferVideoRatio for custom dimensions.
		best := 1e9
		for _, p := range [][2]int{{1, 1}, {3, 4}, {4, 3}, {16, 9}, {9, 16}, {21, 9}} {
			diff := float64(w)/float64(h) - float64(p[0])/float64(p[1])
			if diff < 0 {
				diff = -diff
			}
			if diff < best {
				best = diff
				ratio = fmt.Sprintf("%d:%d", p[0], p[1])
			}
		}
	}
	body := map[string]any{"model": model, "prompt": v.Prompt, "duration": v.Seconds, "resolution": fmt.Sprintf("%dp", resolution), "aspect_ratio": ratio, "generate_audio": true}
	if protocol == "muse" && ratio != "16:9" && ratio != "9:16" {
		return nil, errors.New("Muse supports 16:9 or 9:16 video aspect ratios")
	}
	refs := []map[string]string{}
	total := 0
	for _, raw := range v.References {
		data, _, err := decodeStudioImage(raw)
		if err != nil {
			return nil, err
		}
		total += len(data)
		if total > 30<<20 {
			return nil, errors.New("video references exceed memory limit")
		}
		refs = append(refs, map[string]string{"url": raw})
	}
	if protocol == "muse" {
		delete(body, "generate_audio")
		delete(body, "aspect_ratio")
		body["size"] = ratio
		if len(refs) > 0 {
			body["images"] = refs
		}
	} else if v.Mode == "reference" && len(refs) > 0 {
		body["reference_images"] = refs
	} else if len(refs) > 0 {
		body["image"] = refs[0]
		if len(refs) > 1 {
			body["last_frame"] = refs[1]
		}
	}
	return json.Marshal(body)
}
func normalizeVideoStudioStatus(raw []byte, fallbackID string) (videoStudioStatus, error) {
	var v struct {
		ID        string  `json:"id"`
		RequestID string  `json:"request_id"`
		Status    string  `json:"status"`
		Progress  float64 `json:"progress"`
		Video     struct {
			URL     string `json:"url"`
			Respect *bool  `json:"respect_moderation"`
		} `json:"video"`
		URL      string `json:"url"`
		VideoURL string `json:"video_url"`
	}
	out := videoStudioStatus{Status: "pending"}
	if json.Unmarshal(raw, &v) != nil {
		return out, errors.New("invalid video response")
	}
	out.ID = v.RequestID
	if out.ID == "" {
		out.ID = v.ID
	}
	if out.ID == "" {
		out.ID = fallbackID
	}
	if !videoTaskID.MatchString(out.ID) {
		return out, errors.New("invalid video task ID")
	}
	if v.Progress >= 0 && v.Progress <= 100 {
		out.Progress = v.Progress
	}
	switch strings.ToLower(v.Status) {
	case "failed", "error", "cancelled", "canceled", "expired", "rejected":
		out.Status = "failed"
	case "done", "completed", "succeeded", "success":
		if v.Video.Respect != nil && !*v.Video.Respect {
			out.Status = "failed"
		} else if v.Video.URL != "" || v.URL != "" || v.VideoURL != "" {
			out.Status = "completed"
		} else {
			return out, errors.New("video response contains no output")
		}
	}
	return out, nil
}
func (h *ImageStudioHandler) videoCatalog(ctx context.Context, userID int64, admin bool) (videoStudioCatalog, error) {
	cfg, err := h.settings.GetVideoStudioConfig(ctx)
	out := videoStudioCatalog{Enabled: cfg.Enabled, Models: []videoStudioModel{}}
	if err != nil || !cfg.Enabled && !admin {
		return out, err
	}
	groups, err := h.keys.ImageStudioGroups(ctx, userID, admin)
	if err != nil {
		return out, err
	}
	overrides := map[string]service.ImageStudioModelConfig{}
	for _, m := range cfg.Models {
		overrides[studioModelID(m.GroupID, m.Model)] = m
	}
	for _, g := range groups {
		if !service.GroupAllowsImageGeneration(&g) || g.Platform != service.PlatformGrok && g.Platform != service.PlatformComposite && g.Platform != service.PlatformOpenAI {
			continue
		}
		models := h.gateway.gatewayService.GetAvailableModels(ctx, &g.ID, g.Platform)
		if g.Platform == service.PlatformComposite {
			models = h.gateway.compositeAvailableModels(ctx, &g.ID, "", false)
		} else if models == nil {
			models = defaultModelIDsForPlatform(g.Platform)
		}
		for _, name := range models {
			if !g.ModelAllowlist.Allows(name) || strings.ContainsAny(name, "\\?#\r\n\x00*") {
				continue
			}
			id := studioModelID(g.ID, name)
			setting, configured := overrides[id]
			protocol := videoStudioProtocol(name)
			if configured && setting.Protocol != "" {
				protocol = setting.Protocol
			}
			if protocol == "" || protocol == "muse" && g.Platform == service.PlatformGrok || protocol == "grok" && g.Platform == service.PlatformOpenAI {
				continue
			}
			if configured && !setting.Enabled && !admin {
				continue
			}
			label := name
			if configured && setting.Label != "" {
				label = setting.Label
			}
			out.Models = append(out.Models, videoStudioModel{ID: id, GroupID: g.ID, GroupName: g.Name, Model: name, Label: label, Protocol: protocol})
		}
	}
	sort.Slice(out.Models, func(i, j int) bool { return out.Models[i].ID < out.Models[j].ID })
	return out, nil
}
func (h *ImageStudioHandler) VideoCatalog(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "authentication required")
		return
	}
	out, err := h.videoCatalog(c.Request.Context(), subject.UserID, false)
	if err != nil {
		response.InternalError(c, "Video catalog is unavailable")
		return
	}
	response.Success(c, out)
}
func (h *ImageStudioHandler) AdminVideoCatalog(c *gin.Context) {
	out, err := h.videoCatalog(c.Request.Context(), 0, true)
	if err != nil {
		response.InternalError(c, "Video catalog is unavailable")
		return
	}
	response.Success(c, out)
}
func (h *ImageStudioHandler) GetVideoConfig(c *gin.Context) {
	out, err := h.settings.GetVideoStudioConfig(c.Request.Context())
	if err != nil {
		response.InternalError(c, "Video configuration is unavailable")
		return
	}
	response.Success(c, out)
}
func (h *ImageStudioHandler) SetVideoConfig(c *gin.Context) {
	var cfg service.VideoStudioConfig
	if c.ShouldBindJSON(&cfg) != nil {
		response.BadRequest(c, "Invalid video configuration")
		return
	}
	groups, err := h.keys.ImageStudioGroups(c.Request.Context(), 0, true)
	if err != nil {
		response.InternalError(c, "Groups are unavailable")
		return
	}
	previous, err := h.settings.GetVideoStudioConfig(c.Request.Context())
	if err != nil {
		response.InternalError(c, "Video configuration is unavailable")
		return
	}
	allowed := map[int64]bool{}
	for _, g := range groups {
		allowed[g.ID] = g.Platform == service.PlatformGrok || g.Platform == service.PlatformComposite || g.Platform == service.PlatformOpenAI
	}
	unchanged := map[service.ImageStudioModelConfig]bool{}
	for _, m := range previous.Models {
		unchanged[m] = true
	}
	for _, m := range cfg.Models {
		if !allowed[m.GroupID] && !unchanged[m] {
			response.BadRequest(c, "Unknown or unsupported video group")
			return
		}
	}
	if err = h.settings.SetVideoStudioConfig(c.Request.Context(), cfg); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, cfg)
}
func (h *ImageStudioHandler) GenerateVideo(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "authentication required")
		return
	}
	raw, err := pkghttputil.ReadRequestBodyWithPrealloc(c.Request)
	if err != nil || len(raw) > imageStudioMaxBody {
		response.Error(c, 413, "Video request is too large")
		return
	}
	var input videoStudioRequest
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&input) != nil || decoder.Decode(new(any)) != io.EOF {
		response.BadRequest(c, "Invalid video request")
		return
	}
	catalog, err := h.videoCatalog(c.Request.Context(), subject.UserID, false)
	if err != nil {
		response.InternalError(c, "Video catalog is unavailable")
		return
	}
	var model *videoStudioModel
	for i := range catalog.Models {
		if catalog.Models[i].ID == input.ModelID {
			model = &catalog.Models[i]
			break
		}
	}
	if model == nil || !catalog.Enabled {
		response.Forbidden(c, "This video model is unavailable")
		return
	}
	body, err := buildVideoStudioProtocolRequest(model.Model, model.Protocol, input)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	key, err := h.keys.EnsureImageStudioKey(c.Request.Context(), subject.UserID, model.GroupID)
	if err != nil {
		response.Forbidden(c, "Video generation is not authorized")
		return
	}
	path := "/v1/videos/generations"
	if model.Protocol == "muse" {
		path = "/v1/videos"
	}
	captured := h.relayVideo(c.Request, key, path, body, 1<<20)
	defer captured.release()
	h.respondVideo(c, captured, "")
}
func (h *ImageStudioHandler) VideoStatus(c *gin.Context)  { h.videoLookup(c, false) }
func (h *ImageStudioHandler) VideoContent(c *gin.Context) { h.videoLookup(c, true) }
func (h *ImageStudioHandler) videoLookup(c *gin.Context, content bool) {
	c.Header("Cache-Control", "no-store")
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "authentication required")
		return
	}
	group, err := strconv.ParseInt(c.Query("group_id"), 10, 64)
	id := c.Query("task_id")
	if err != nil || group < 1 || !videoTaskID.MatchString(id) {
		response.BadRequest(c, "Invalid video task")
		return
	}
	// Hiding a model prevents new submissions, but must not orphan already billed tasks.
	key, err := h.keys.EnsureImageStudioKey(c.Request.Context(), subject.UserID, group)
	if err != nil {
		response.Forbidden(c, "Video task is not authorized")
		return
	}
	path := "/v1/videos/" + id
	limit := 1 << 20
	if content {
		path += "/content"
		limit = 256 << 20
	}
	capture := h.relayVideo(c.Request, key, path, nil, limit)
	defer capture.release()
	if !content {
		h.respondVideo(c, capture, id)
		return
	}
	if videoRelayError(c, capture) {
		return
	}
	if capture.status != http.StatusOK || capture.body.Len() == 0 || capture.header.Get("Content-Range") != "" || capture.header.Get("X-Sub2API-Content-Read-Error") != "" {
		response.Error(c, 502, "Video download was incomplete; retry fetching the original task")
		return
	}
	if declared := capture.header.Get("Content-Length"); declared != "" {
		expected, err := strconv.ParseInt(declared, 10, 64)
		if err != nil || expected != int64(capture.body.Len()) {
			response.Error(c, 502, "Video download was incomplete; retry fetching the original task")
			return
		}
	}
	kind := strings.Split(capture.header.Get("Content-Type"), ";")[0]
	if kind != "video/mp4" && kind != "video/webm" && kind != "application/octet-stream" {
		response.Error(c, 502, "Video content is unavailable")
		return
	}
	if kind == "application/octet-stream" {
		kind = "video/mp4"
	}
	c.Header("X-Content-Type-Options", "nosniff")
	c.Data(200, kind, capture.body.Bytes())
}

type videoStudioCapture struct {
	budgetContext context.Context
	release       func()
	header        http.Header
	body          bytes.Buffer
	status        int
	limit         int
	overflow      bool
}

func (w *videoStudioCapture) Header() http.Header { return w.header }
func (w *videoStudioCapture) WriteHeader(status int) {
	if w.status == 0 {
		w.status = status
	}
}
func (w *videoStudioCapture) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = 200
	}
	if w.body.Len()+len(b) > w.limit {
		w.overflow = true
		return 0, errors.New("video response limit")
	}
	if w.budgetContext != nil {
		// bytes.Buffer may retain up to twice the written bytes during growth.
		if err := pkghttputil.ReserveRequestBodyMemory(w.budgetContext, int64(len(b))*2); err != nil {
			w.overflow = true
			return 0, err
		}
	}
	return w.body.Write(b)
}
func (h *ImageStudioHandler) relayVideo(original *http.Request, key *service.APIKey, path string, body []byte, limit int) *videoStudioCapture {
	ctx := ctxkey.WithNoContentPersistence(service.WithImageStudioCredential(original.Context(), key.Key))
	ctx = logger.IntoContext(ctx, zap.NewNop())
	req := original.Clone(ctx)
	req.URL.Path = path
	req.URL.RawPath = ""
	req.URL.RawQuery = ""
	req.RequestURI = path
	req.Method = http.MethodGet
	if body != nil {
		req.Method = http.MethodPost
	}
	req.Body = io.NopCloser(bytes.NewReader(body))
	req.GetBody = nil
	req.ContentLength = int64(len(body))
	req.Header = make(http.Header)
	req.Header.Set("Authorization", "Bearer "+key.Key)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", original.UserAgent())
	for _, name := range []string{"X-Forwarded-For", "X-Real-IP", "Forwarded"} {
		if v := original.Header.Get(name); v != "" {
			req.Header.Set(name, v)
		}
	}
	budgetContext, release := pkghttputil.WithRequestBodyMemoryBudget(context.Background(), videoOutputBudget)
	capture := &videoStudioCapture{header: make(http.Header), limit: limit, budgetContext: budgetContext, release: release}
	h.relay.ServeHTTP(capture, req)
	return capture
}
func videoRelayError(c *gin.Context, w *videoStudioCapture) bool {
	if w.overflow {
		response.Error(c, 502, "Video response exceeded the download limit; the task may already be billed")
		return true
	}
	if w.status < 200 || w.status >= 300 {
		status := w.status
		if status < 400 || status > 599 {
			status = 502
		}
		response.Error(c, status, "Video request failed; check model access, quota or retry task lookup")
		return true
	}
	return false
}
func (h *ImageStudioHandler) respondVideo(c *gin.Context, w *videoStudioCapture, id string) {
	if videoRelayError(c, w) {
		return
	}
	out, err := normalizeVideoStudioStatus(w.body.Bytes(), id)
	if err != nil {
		response.Error(c, 502, "Video response could not be read; do not automatically resubmit")
		return
	}
	response.Success(c, out)
}
