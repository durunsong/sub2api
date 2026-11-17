package handler

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	pkghttputil "github.com/Wei-Shaw/sub2api/internal/pkg/httputil"
	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ImageStudioHandler struct {
	keys        *service.APIKeyService
	settings    *service.SettingService
	gateway     *GatewayHandler
	relay       http.Handler
	inputBudget *pkghttputil.BodyMemoryBudget
}

func NewImageStudioHandler(keys *service.APIKeyService, settings *service.SettingService, gateway *GatewayHandler, relay http.Handler) *ImageStudioHandler {
	return &ImageStudioHandler{keys: keys, settings: settings, gateway: gateway, relay: relay, inputBudget: pkghttputil.NewBodyMemoryBudget(128 << 20)}
}

func (h *ImageStudioHandler) RequestBodyLimit() gin.HandlerFunc {
	return middleware.RequestBodyLimit(imageStudioMaxBody, h.inputBudget)
}

type imageStudioModel struct {
	ID                 string                     `json:"id"`
	GroupID            int64                      `json:"group_id"`
	GroupName          string                     `json:"group_name"`
	Model              string                     `json:"model"`
	Label              string                     `json:"label"`
	Protocol           string                     `json:"protocol"`
	Sizes              []string                   `json:"sizes"`
	Qualities          []string                   `json:"qualities"`
	MaxImages          int                        `json:"max_images"`
	MaxReferences      int                        `json:"max_references"`
	SupportsCustomSize bool                       `json:"supports_custom_size"`
	SizeMultiple       int                        `json:"size_multiple"`
	SupportsMask       bool                       `json:"supports_mask"`
	Pricing            service.ImageStudioPricing `json:"pricing"`
	platform           string
	canonicalModel     string
}

type imageStudioCatalog struct {
	Enabled bool               `json:"enabled"`
	Models  []imageStudioModel `json:"models"`
}

func studioModelID(groupID int64, model string) string {
	return fmt.Sprintf("%d:%s", groupID, base64.RawURLEncoding.EncodeToString([]byte(model)))
}
func imageStudioProtocol(platform, model string) string {
	m := strings.ToLower(model)
	switch platform {
	case service.PlatformGemini, service.PlatformAntigravity:
		if strings.HasPrefix(m, "gemini-") && (strings.Contains(m, "-image") || strings.Contains(m, "nano-banana")) {
			return "gemini"
		}
	case service.PlatformOpenAI:
		if strings.HasPrefix(m, "gpt-image-") || (strings.HasPrefix(m, "gemini-") && strings.Contains(m, "-image")) {
			return "images"
		}
	case service.PlatformGrok:
		if (strings.Contains(m, "imagine") || strings.Contains(m, "image")) && !strings.Contains(m, "video") {
			return "images"
		}
	case service.PlatformComposite:
		if strings.HasPrefix(m, "gemini-") && strings.Contains(m, "-image") {
			return "gemini"
		}
		if strings.HasPrefix(m, "gpt-image-") || strings.HasPrefix(m, "grok-imagine-image") || m == "grok-imagine" {
			return "images"
		}
	}
	return ""
}
func studioModel(group service.Group, model, protocol string) imageStudioModel {
	m := imageStudioModel{ID: studioModelID(group.ID, model), GroupID: group.ID, GroupName: group.Name, Model: model, Label: model, Protocol: protocol, MaxImages: 1, MaxReferences: 3, Sizes: []string{"1024x1024", "1536x1024", "1024x1536"}, Qualities: []string{}, platform: group.Platform}

	if protocol == "gemini" {
		// Older Nano Banana and Flash Lite do not support 2K/4K output.
		// Unrecognized families keep the native default until capabilities are known.
		m.Sizes = []string{"1K"}
		lower := strings.ToLower(model)
		if strings.HasPrefix(lower, "gemini-3-pro-image") || strings.HasPrefix(lower, "gemini-3.1-flash-image") || strings.HasPrefix(lower, "gemini-nano-banana-2.1") {
			m.Sizes = append(m.Sizes, "2K", "4K")
		}
		return m
	}
	if strings.HasPrefix(strings.ToLower(model), "gpt-image-") {
		m.SupportsMask = true
		m.Qualities = []string{"auto", "low", "medium", "high"}
		m.Sizes = append(m.Sizes, "auto")
		if strings.HasPrefix(strings.ToLower(model), "gpt-image-2") {
			m.SupportsCustomSize = true
			m.SizeMultiple = 16
			m.Sizes = append(m.Sizes, "2048x2048", "2048x1152", "1152x2048", "3840x2160", "2160x3840")
		}
		if strings.HasPrefix(strings.ToLower(model), "gpt-image-2.5") || strings.HasPrefix(strings.ToLower(model), "gpt-image-2.6") {
			m.Qualities = append(m.Qualities, "xhigh", "max")
		}
	}
	if strings.HasPrefix(model, "dall-e-") {
		m.MaxReferences = 0
		m.Sizes = []string{"1024x1024"}
	}
	return m
}
func (h *ImageStudioHandler) catalog(ctx context.Context, userID int64, admin bool) (imageStudioCatalog, error) {
	cfg, err := h.settings.GetImageStudioConfig(ctx)
	out := imageStudioCatalog{Enabled: cfg.Enabled, Models: []imageStudioModel{}}
	if err != nil {
		return out, err
	}
	if !cfg.Enabled && !admin {
		return out, nil
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
		if !g.AllowImageGeneration {
			continue
		}
		switch g.Platform {
		case service.PlatformOpenAI, service.PlatformGrok, service.PlatformGemini, service.PlatformAntigravity, service.PlatformComposite:
		default:
			continue
		}
		models := h.gateway.gatewayService.GetAvailableModels(ctx, &g.ID, g.Platform)
		if g.Platform == service.PlatformComposite {
			models = h.gateway.compositeAvailableModels(ctx, &g.ID, "", false)
		} else if models == nil {
			models = defaultModelIDsForPlatform(g.Platform)
		}
		// Only aliases already advertised by the actual group can be configured.
		for _, name := range models {
			if strings.ContainsAny(name, "/\\?#\r\n\x00*") || !g.ModelAllowlist.Allows(name) {
				continue
			}
			canonical := name
			protocol := imageStudioProtocol(g.Platform, name)
			if protocol == "" {
				if mapped := h.gateway.gatewayService.ImageStudioMappedModel(ctx, g.ID, g.Platform, name); mapped != "" {
					canonical = mapped
					protocol = imageStudioProtocol(g.Platform, mapped)
				}
			}
			override, configured := overrides[studioModelID(g.ID, name)]
			if configured && !override.Enabled && !admin {
				continue
			}
			if configured && override.Protocol != "" {
				protocol = override.Protocol
			}
			if protocol == "" {
				continue
			}
			if protocol == "gemini" && g.Platform != service.PlatformGemini && g.Platform != service.PlatformAntigravity && g.Platform != service.PlatformComposite {
				continue
			}
			if protocol == "images" && g.Platform != service.PlatformOpenAI && g.Platform != service.PlatformGrok && g.Platform != service.PlatformComposite {
				continue
			}
			m := studioModel(g, canonical, protocol)
			m.ID = studioModelID(g.ID, name)
			m.Model = name
			m.Label = name
			m.canonicalModel = canonical
			m.Pricing = h.gateway.gatewayService.ImageStudioPricing(ctx, userID, &g, name, canonical)
			if configured && override.Label != "" {
				m.Label = override.Label
			}

			out.Models = append(out.Models, m)
		}
	}
	sort.Slice(out.Models, func(i, j int) bool {
		if out.Models[i].GroupID == out.Models[j].GroupID {
			return out.Models[i].Model < out.Models[j].Model
		}
		return out.Models[i].GroupID < out.Models[j].GroupID
	})
	return out, nil
}
func (h *ImageStudioHandler) Catalog(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "authentication required")
		return
	}
	catalog, err := h.catalog(c.Request.Context(), subject.UserID, false)
	if err != nil {
		response.InternalError(c, "Image studio catalog is unavailable")
		return
	}
	response.Success(c, catalog)
}
func (h *ImageStudioHandler) AdminCatalog(c *gin.Context) {
	catalog, err := h.catalog(c.Request.Context(), 0, true)
	if err != nil {
		response.InternalError(c, "Image studio catalog is unavailable")
		return
	}
	response.Success(c, catalog)
}
func (h *ImageStudioHandler) GetConfig(c *gin.Context) {
	cfg, err := h.settings.GetImageStudioConfig(c.Request.Context())
	if err != nil {
		response.InternalError(c, "Image studio configuration is unavailable")
		return
	}
	response.Success(c, cfg)
}
func (h *ImageStudioHandler) SetConfig(c *gin.Context) {
	var cfg service.ImageStudioConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		response.BadRequest(c, "Invalid image studio configuration")
		return
	}
	groups, err := h.keys.ImageStudioGroups(c.Request.Context(), 0, true)
	if err != nil {
		response.InternalError(c, "Groups are unavailable")
		return
	}
	allowed := map[int64]bool{}
	for _, g := range groups {
		allowed[g.ID] = true
	}
	previous, err := h.settings.GetImageStudioConfig(c.Request.Context())
	if err != nil {
		response.InternalError(c, "Image studio configuration is unavailable")
		return
	}
	unchanged := make(map[service.ImageStudioModelConfig]bool, len(previous.Models))
	for _, m := range previous.Models {
		unchanged[m] = true
	}
	for _, m := range cfg.Models {
		// Retired groups must not block unrelated edits; new or changed overrides still require an active group.
		if !allowed[m.GroupID] && !unchanged[m] {
			response.BadRequest(c, "Unknown or inactive group")
			return
		}
	}
	if err = h.settings.SetImageStudioConfig(c.Request.Context(), cfg); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, cfg)
}
func (h *ImageStudioHandler) Generate(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "authentication required")
		return
	}
	raw, err := pkghttputil.ReadRequestBodyWithPrealloc(c.Request)
	if err != nil || len(raw) > imageStudioMaxBody {
		response.Error(c, http.StatusRequestEntityTooLarge, "Image request is too large")
		return
	}
	var input imageStudioGenerateRequest
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&input); err != nil {
		response.BadRequest(c, "Invalid image request")
		return
	}
	if decoder.Decode(new(any)) != io.EOF {
		response.BadRequest(c, "Invalid image request")
		return
	}
	catalog, err := h.catalog(c.Request.Context(), subject.UserID, false)
	if err != nil {
		response.InternalError(c, "Image studio catalog is unavailable")
		return
	}
	var model *imageStudioModel
	for i := range catalog.Models {
		if catalog.Models[i].ID == input.ModelID {
			model = &catalog.Models[i]
			break
		}
	}
	if model == nil || !catalog.Enabled {
		response.Forbidden(c, "This image model is unavailable")
		return
	}
	body, path, contentType, err := buildImageStudioRequest(*model, input)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	key, err := h.keys.EnsureImageStudioKey(c.Request.Context(), subject.UserID, model.GroupID)
	if err != nil {
		if !response.ErrorFrom(c, err) {
			response.InternalError(c, "Image studio authorization failed")
		}
		return
	}
	if model.platform == service.PlatformAntigravity && model.Protocol == "gemini" {
		path = "/antigravity" + path
	}
	ctx := ctxkey.WithNoContentPersistence(service.WithImageStudioCredential(c.Request.Context(), key.Key))
	ctx = service.WithImageStudioModel(ctx, model.Model, model.canonicalModel)
	ctx = logger.IntoContext(ctx, zap.NewNop())
	req := c.Request.Clone(ctx)
	req.URL.Path = path
	req.URL.RawPath = ""
	req.URL.RawQuery = ""
	req.RequestURI = path
	req.Body = io.NopCloser(bytes.NewReader(body))
	req.GetBody = nil
	req.ContentLength = int64(len(body))
	// Forward only metadata needed by the existing gateway; never forward panel JWTs.
	req.Header = make(http.Header)
	req.Header.Set("Authorization", "Bearer "+key.Key)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("User-Agent", c.Request.UserAgent())
	for _, name := range []string{"X-Forwarded-For", "X-Real-IP", "Forwarded"} {
		if v := c.GetHeader(name); v != "" {
			req.Header.Set(name, v)
		}
	}
	capture := &imageStudioResponseWriter{header: make(http.Header)}
	h.relay.ServeHTTP(capture, req)
	if capture.overflow {
		response.Error(c, http.StatusBadGateway, "Generated output exceeded the response limit; generation may already be billed")
		return
	}
	if capture.status >= 400 {
		// Upstream messages may echo private content. The panel receives only a stable category.
		message := "Image generation failed"
		switch capture.status {
		case 400:
			message = "The provider rejected the image parameters"
		case 401, 403:
			message = "Image generation is not authorized or the balance is insufficient"
		case 429:
			message = "Image generation capacity or quota is exhausted"
		case 499:
			message = "Image generation was cancelled"
		}
		if retry := capture.header.Get("Retry-After"); retry != "" {
			c.Header("Retry-After", retry)
		}
		response.Error(c, capture.status, message)
		return
	}
	output, err := normalizeImageStudioResponse(capture.body.Bytes())
	if err != nil {
		response.Error(c, http.StatusBadGateway, err.Error())
		return
	}
	response.Success(c, output)
}
