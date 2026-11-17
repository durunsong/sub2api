package handler

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	middleware "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestImageStudioRequestRejectsRemoteReferencesAndUnsupportedControls(t *testing.T) {
	model := imageStudioModel{Model: "gpt-image-2", Protocol: "images", Sizes: []string{"1024x1024"}, Qualities: []string{"auto"}, MaxImages: 1, MaxReferences: 3, SupportsMask: true}
	for _, input := range []imageStudioGenerateRequest{
		{Prompt: "hello", References: []string{"https://example.com/private.png"}},
		{Prompt: "hello", Size: "99999x99999"},
		{Prompt: "hello", N: 2},
		{Prompt: "hello", Mask: "data:image/png;base64,aGVsbG8="},
	} {
		_, _, _, err := buildImageStudioRequest(model, input)
		require.Error(t, err)
	}
}

func TestImageStudioGeminiRequestPreservesReferenceAndImageModality(t *testing.T) {
	model := imageStudioModel{Model: "gemini-3-pro-image", Protocol: "gemini", Sizes: []string{"1K", "2K", "4K"}, MaxImages: 1, MaxReferences: 3}
	body, path, contentType, err := buildImageStudioRequest(model, imageStudioGenerateRequest{Prompt: "edit", Resolution: "2K", AspectRatio: "16:9", References: []string{testStudioPNG}})
	require.NoError(t, err)
	require.Equal(t, "/v1beta/models/gemini-3-pro-image:generateContent", path)
	require.Equal(t, "application/json", contentType)
	var parsed map[string]any
	require.NoError(t, json.Unmarshal(body, &parsed))
	require.Equal(t, []any{"TEXT", "IMAGE"}, parsed["generationConfig"].(map[string]any)["responseModalities"])
	parts := parsed["contents"].([]any)[0].(map[string]any)["parts"].([]any)
	require.Equal(t, "image/png", parts[1].(map[string]any)["inlineData"].(map[string]any)["mimeType"])
}

func TestImageStudioResponseDoesNotExposeRemoteURLs(t *testing.T) {
	_, err := normalizeImageStudioResponse([]byte(`{"data":[{"url":"http://127.0.0.1/secret"}]}`))
	require.Error(t, err)
	result, err := normalizeImageStudioResponse([]byte(`{"candidates":[{"content":{"parts":[{"text":"done"},{"inlineData":{"mimeType":"image/png","data":"` + testStudioPNGData + `"}}]}}]}`))
	require.NoError(t, err)
	require.Len(t, result.Images, 1)
	require.Equal(t, "done", result.Text)
}

const testStudioPNGData = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+jRZkAAAAASUVORK5CYII="
const testStudioPNG = "data:image/png;base64," + testStudioPNGData

// Repository fakes exercise the real authorization/catalog services.
type studioSettingsRepo struct {
	service.SettingRepository
	config string
}

func (r *studioSettingsRepo) GetValue(context.Context, string) (string, error) { return r.config, nil }
func (r *studioSettingsRepo) Set(_ context.Context, _ string, value string) error {
	r.config = value
	return nil
}

type studioUserRepo struct {
	service.UserRepository
	user *service.User
}

func (r *studioUserRepo) GetByID(context.Context, int64) (*service.User, error) { return r.user, nil }

type studioGroupRepo struct {
	service.GroupRepository
	groups []service.Group
}

func (r *studioGroupRepo) ListActive(context.Context) ([]service.Group, error) { return r.groups, nil }
func (r *studioGroupRepo) GetByID(_ context.Context, id int64) (*service.Group, error) {
	for i := range r.groups {
		if r.groups[i].ID == id {
			return &r.groups[i], nil
		}
	}
	return nil, service.ErrGroupNotFound
}

type studioSubRepo struct {
	service.UserSubscriptionRepository
	subs []service.UserSubscription
}

func (r *studioSubRepo) ListActiveByUserID(context.Context, int64) ([]service.UserSubscription, error) {
	return r.subs, nil
}
func (r *studioSubRepo) GetActiveByUserIDAndGroupID(_ context.Context, _ int64, id int64) (*service.UserSubscription, error) {
	for i := range r.subs {
		if r.subs[i].GroupID == id {
			return &r.subs[i], nil
		}
	}
	return nil, errors.New("no subscription")
}

type studioKeyRepo struct {
	service.APIKeyRepository
	keys map[string]*service.APIKey
}

func (r *studioKeyRepo) GetByKey(_ context.Context, key string) (*service.APIKey, error) {
	if k := r.keys[key]; k != nil {
		return k, nil
	}
	return nil, service.ErrAPIKeyNotFound
}
func (r *studioKeyRepo) Create(_ context.Context, key *service.APIKey) error {
	key.ID = int64(len(r.keys) + 1)
	r.keys[key.Key] = key
	return nil
}

func newStudioForTest(t *testing.T, configJSON string, relay http.Handler) (*ImageStudioHandler, *studioUserRepo, *studioGroupRepo, *studioKeyRepo) {
	t.Helper()
	users := &studioUserRepo{user: &service.User{ID: 7, Status: service.StatusActive, RestrictPublicGroups: true, AllowedGroups: []int64{1}}}
	groups := &studioGroupRepo{groups: []service.Group{
		{ID: 1, Name: "allowed", Status: service.StatusActive, Platform: service.PlatformOpenAI, AllowImageGeneration: true, RateMultiplier: 2},
		{ID: 2, Name: "private", Status: service.StatusActive, Platform: service.PlatformOpenAI, AllowImageGeneration: true, IsExclusive: true},
		{ID: 3, Name: "subscription", Status: service.StatusActive, Platform: service.PlatformGemini, AllowImageGeneration: true, SubscriptionType: service.SubscriptionTypeSubscription},
	}}
	keys := &studioKeyRepo{keys: map[string]*service.APIKey{}}
	cfg := &config.Config{}
	cfg.JWT.Secret = "test-only-signing-secret"
	svc := service.NewAPIKeyService(keys, users, groups, &studioSubRepo{}, nil, nil, cfg)
	settings := service.NewSettingService(&studioSettingsRepo{config: configJSON}, cfg)
	accounts := &gatewayModelsAccountRepoStub{byGroup: map[int64][]service.Account{1: {{ID: 1, Platform: service.PlatformOpenAI, Credentials: map[string]any{"model_mapping": map[string]any{"gpt-image-2": "gpt-image-2", "gpt-4": "gpt-4"}}}}, 2: {{ID: 2, Platform: service.PlatformOpenAI, Credentials: map[string]any{"model_mapping": map[string]any{"gpt-image-2": "gpt-image-2"}}}}}}
	return NewImageStudioHandler(svc, settings, newGatewayModelsHandlerForTest(accounts), relay), users, groups, keys
}
func TestImageStudioCatalogEnforcesGroupsModelsAndAdminHiding(t *testing.T) {
	h, _, groups, _ := newStudioForTest(t, "", nil)
	catalog, err := h.catalog(context.Background(), 7, false)
	require.NoError(t, err)
	require.True(t, catalog.Enabled)
	require.Len(t, catalog.Models, 1)
	require.Equal(t, int64(1), catalog.Models[0].GroupID)
	require.Equal(t, "gpt-image-2", catalog.Models[0].Model)
	require.NoError(t, h.settings.SetImageStudioConfig(context.Background(), service.ImageStudioConfig{Enabled: true, Models: []service.ImageStudioModelConfig{{GroupID: 1, Model: "gpt-image-2", Enabled: false}}}))
	catalog, err = h.catalog(context.Background(), 7, false)
	require.NoError(t, err)
	require.Empty(t, catalog.Models)
	require.NoError(t, h.settings.SetImageStudioConfig(context.Background(), service.ImageStudioConfig{Enabled: true}))
	groups.groups[0].ModelAllowlist = service.GroupModelAllowlist{Enabled: true, Models: []string{"gpt-4"}}
	catalog, err = h.catalog(context.Background(), 7, false)
	require.NoError(t, err)
	require.Empty(t, catalog.Models)
}
func TestImageStudioGenerateUsesFixedGatewayPathAndPrivateAccountingIdentity(t *testing.T) {
	calls := 0
	relay := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		require.Equal(t, "/v1/images/generations", r.URL.Path)
		require.Empty(t, r.URL.RawQuery)
		require.True(t, ctxkey.NoContentPersistence(r.Context()))
		require.True(t, strings.HasPrefix(r.Header.Get("Authorization"), "Bearer internal:images:"))
		require.NotContains(t, r.Header.Get("Authorization"), "panel-jwt")
		var native map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&native))
		require.Equal(t, "gpt-image-2", native["model"])
		require.Equal(t, "b64_json", native["response_format"])
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"data":[{"b64_json":"`+testStudioPNGData+`"}]}`)
	})
	h, users, _, keys := newStudioForTest(t, "", relay)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/generate", func(c *gin.Context) {
		c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 7})
		c.Next()
	}, h.RequestBodyLimit(), h.Generate)
	body := `{"model_id":"` + studioModelID(1, "gpt-image-2") + `","prompt":"a cat"}`
	req := httptest.NewRequest(http.MethodPost, "/generate?key=not-forwarded", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer panel-jwt")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, 200, w.Code, w.Body.String())
	require.Len(t, keys.keys, 1)
	require.Equal(t, 1, calls)
	users.user.AllowedGroups = nil
	w = httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/generate", strings.NewReader(body)))
	require.Equal(t, 403, w.Code)
	require.Equal(t, 1, calls)
}

func TestImageStudioCatalogDiscoversMappedNanoBananaAlias(t *testing.T) {
	h, _, _, _ := newStudioForTest(t, "", nil)
	h.gateway = newGatewayModelsHandlerForTest(&gatewayModelsAccountRepoStub{byGroup: map[int64][]service.Account{1: {{ID: 1, Platform: service.PlatformOpenAI, Credentials: map[string]any{"model_mapping": map[string]any{"nano-banana": "gemini-3-pro-image"}}}}}})
	catalog, err := h.catalog(context.Background(), 7, false)
	require.NoError(t, err)
	require.Len(t, catalog.Models, 1)
	require.Equal(t, "nano-banana", catalog.Models[0].Model)
	require.Equal(t, "images", catalog.Models[0].Protocol)
}

func TestImageStudioCustomDimensionsRespectGPTImage2Limits(t *testing.T) {
	model := studioModel(service.Group{Platform: service.PlatformOpenAI}, "gpt-image-2", "images")
	for _, size := range []string{"2048x2048", "3840x2160", "2160x3840", "1536x864"} {
		_, _, _, err := buildImageStudioRequest(model, imageStudioGenerateRequest{Prompt: "hello", Size: size})
		require.NoError(t, err, size)
	}
	for _, size := range []string{"3841x2160", "3840x3840", "1024x256", "800x800", "1025x1024", "9999999999999999999999x1024"} {
		_, _, _, err := buildImageStudioRequest(model, imageStudioGenerateRequest{Prompt: "hello", Size: size})
		require.Error(t, err, size)
	}
	old := studioModel(service.Group{Platform: service.PlatformOpenAI}, "gpt-image-1", "images")
	_, _, _, err := buildImageStudioRequest(old, imageStudioGenerateRequest{Prompt: "hello", Size: "3840x2160"})
	require.Error(t, err)
}

func TestImageStudioGeminiCapabilitiesRespectNativeResolution(t *testing.T) {
	for _, name := range []string{"gemini-2.5-flash-image", "gemini-2.5-flash-image-preview", "gemini-3.1-flash-lite-image"} {
		model := studioModel(service.Group{ID: 1, Platform: service.PlatformGemini}, name, "gemini")
		require.Equal(t, []string{"1K"}, model.Sizes, name)
		body, _, _, err := buildImageStudioRequest(model, imageStudioGenerateRequest{Prompt: "draw", Size: "1K", Resolution: "1K", AspectRatio: "16:9"})
		require.NoError(t, err)
		var payload map[string]any
		require.NoError(t, json.Unmarshal(body, &payload))
		imageConfig := payload["generationConfig"].(map[string]any)["imageConfig"].(map[string]any)
		require.NotContains(t, imageConfig, "imageSize", name)
		require.Equal(t, "16:9", imageConfig["aspectRatio"])
		_, _, _, err = buildImageStudioRequest(model, imageStudioGenerateRequest{Prompt: "draw", Size: "2K"})
		require.Error(t, err, name)
	}
	for _, name := range []string{"gemini-3-pro-image-preview", "gemini-3.1-flash-image"} {
		model := studioModel(service.Group{ID: 1, Platform: service.PlatformGemini}, name, "gemini")
		require.Equal(t, []string{"1K", "2K", "4K"}, model.Sizes)
		body, _, _, err := buildImageStudioRequest(model, imageStudioGenerateRequest{Prompt: "draw", Size: "4K"})
		require.NoError(t, err)
		require.Contains(t, string(body), `"imageSize":"4K"`)
	}
}
func TestImageStudioBodyLimitFitsThreeMaximumReferences(t *testing.T) {
	image := testStudioPNG + strings.Repeat("A", base64.StdEncoding.EncodedLen(imageStudioMaxImage)-len(testStudioPNGData))
	body, err := json.Marshal(imageStudioGenerateRequest{ModelID: strings.Repeat("m", 200), Prompt: strings.Repeat("🎨", 16000), References: []string{image, image, image}})
	require.NoError(t, err)
	require.LessOrEqual(t, len(body), imageStudioMaxBody, "advertised reference capacity must fit after base64 and JSON overhead")
}

func TestImageStudioConfigPreservesUnchangedInactiveOverrides(t *testing.T) {
	h, _, _, _ := newStudioForTest(t, "", nil)
	stale := service.ImageStudioModelConfig{GroupID: 99, Model: "gpt-image-2", Enabled: true, Label: "Old group"}
	require.NoError(t, h.settings.SetImageStudioConfig(context.Background(), service.ImageStudioConfig{Enabled: true, Models: []service.ImageStudioModelConfig{stale}}))
	for _, tc := range []struct {
		name   string
		models []service.ImageStudioModelConfig
		want   int
	}{
		{"preserve stale while changing live", []service.ImageStudioModelConfig{stale, {GroupID: 1, Model: "gpt-image-2", Enabled: false}}, http.StatusOK},
		{"reject new unavailable", []service.ImageStudioModelConfig{{GroupID: 98, Model: "gpt-image-2", Enabled: true}}, http.StatusBadRequest},
		{"reject changed unavailable", []service.ImageStudioModelConfig{{GroupID: 99, Model: "gpt-image-2", Enabled: false}}, http.StatusBadRequest},
		{"remove stale", nil, http.StatusOK},
	} {
		t.Run(tc.name, func(t *testing.T) {
			payload, err := json.Marshal(service.ImageStudioConfig{Enabled: false, Models: tc.models})
			require.NoError(t, err)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPut, "/", strings.NewReader(string(payload)))
			c.Request.Header.Set("Content-Type", "application/json")
			h.SetConfig(c)
			require.Equal(t, tc.want, w.Code, w.Body.String())
			if tc.want == http.StatusOK {
				saved, err := h.settings.GetImageStudioConfig(context.Background())
				require.NoError(t, err)
				require.False(t, saved.Enabled)
				require.Equal(t, len(tc.models), len(saved.Models))
			}
		})
	}
}
