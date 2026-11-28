package handler

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestMuseImageStudioCapabilities(t *testing.T) {
	require.Equal(t, "images", imageStudioProtocol(service.PlatformOpenAI, "muse-image"))
	require.Equal(t, "images", imageStudioProtocol(service.PlatformComposite, "muse-image"))
	model := studioModel(service.Group{ID: 1, Platform: service.PlatformOpenAI}, "muse-image", "images")
	require.False(t, model.SupportsMask)
	require.Empty(t, model.Qualities)
	require.Contains(t, model.Sizes, "auto")
	body, path, _, err := buildImageStudioRequest(model, imageStudioGenerateRequest{Prompt: "draw", N: 1, Size: "auto"})
	require.NoError(t, err)
	require.Equal(t, "/v1/images/generations", path)
	require.Contains(t, string(body), "b64_json")
}

func TestMuseImageStudioGPTAliasUsesMappedCapabilities(t *testing.T) {
	h, _, _, _ := newStudioForTest(t, "", nil)
	h.gateway = newGatewayModelsHandlerForTest(&gatewayModelsAccountRepoStub{byGroup: map[int64][]service.Account{1: {{ID: 1, Platform: service.PlatformOpenAI, Type: service.AccountTypeAPIKey, Credentials: map[string]any{"base_url": "https://muse.example/v1", "model_mapping": map[string]any{"gpt-image-custom": "muse-image"}}}}}})
	catalog, err := h.catalog(context.Background(), 7, false)
	require.NoError(t, err)
	require.Len(t, catalog.Models, 1)
	require.Equal(t, "gpt-image-custom", catalog.Models[0].Model)
	require.Equal(t, "muse-image", catalog.Models[0].canonicalModel)
	require.False(t, catalog.Models[0].SupportsMask)
	require.Empty(t, catalog.Models[0].Qualities)
}

func TestMuseImageStudioHidesAmbiguousFamily(t *testing.T) {
	for _, otherTarget := range []string{"gpt-image-2", ""} {
		for _, museFirst := range []bool{true, false} {
			h, _, _, _ := newStudioForTest(t, "", nil)
			accounts := []service.Account{
				{ID: 1, Platform: service.PlatformOpenAI, Type: service.AccountTypeAPIKey, Credentials: map[string]any{"base_url": "https://muse.example/v1", "model_mapping": map[string]any{"gpt-image-custom": "muse-image"}}},
				{ID: 2, Platform: service.PlatformOpenAI, Type: service.AccountTypeAPIKey, Credentials: map[string]any{}},
			}
			if otherTarget != "" {
				accounts[1].Credentials["model_mapping"] = map[string]any{"gpt-image-custom": otherTarget}
			}
			if !museFirst {
				accounts[0], accounts[1] = accounts[1], accounts[0]
			}
			h.gateway = newGatewayModelsHandlerForTest(&gatewayModelsAccountRepoStub{byGroup: map[int64][]service.Account{1: accounts}})
			catalog, err := h.catalog(context.Background(), 7, false)
			require.NoError(t, err)
			for _, model := range catalog.Models {
				require.NotEqual(t, "gpt-image-custom", model.Model, "mixed Muse/GPT alias must not advertise unsupported controls")
			}
		}
	}
}

func TestMuseImageStudioCompositeAliasUsesRoutedCapabilities(t *testing.T) {
	for _, routeModel := range []string{"", "muse-image", "provider-image"} {
		t.Run("route_"+routeModel, func(t *testing.T) {
			h, _, groups, _ := newStudioForTest(t, "", nil)
			groups.groups[0].Platform = service.PlatformComposite
			mapping := map[string]any{"gpt-image-custom": "muse-image", "muse-image": "muse-image", "provider-image": "muse-image"}
			h.gateway = newGatewayModelsHandlerForTest(&gatewayModelsAccountRepoStub{byGroup: map[int64][]service.Account{1: {
				{ID: 1, Platform: service.PlatformOpenAI, Type: service.AccountTypeAPIKey, Credentials: map[string]any{"base_url": "https://muse.example/v1", "model_mapping": mapping}},
				{ID: 2, Platform: service.PlatformGrok, Type: service.AccountTypeAPIKey, Credentials: map[string]any{"model_mapping": map[string]any{"gpt-image-custom": "grok-imagine-image"}}},
			}}}, &gatewayModelsRouteRepoStub{routes: []service.CompositeModelRoute{{PublicModel: "gpt-image-custom", TargetPlatform: service.PlatformOpenAI, UpstreamModel: routeModel, MatchType: service.CompositeRouteMatchExact, Endpoint: service.CompositeRouteEndpointImages, Enabled: true}}})
			catalog, err := h.catalog(context.Background(), 7, false)
			require.NoError(t, err)
			found := false
			for _, model := range catalog.Models {
				if model.Model != "gpt-image-custom" {
					continue
				}
				found = true
				require.Equal(t, "muse-image", model.canonicalModel)
				require.False(t, model.SupportsMask)
				require.Empty(t, model.Qualities)
			}
			require.True(t, found)
		})
	}
}

func TestMuseVideoStudioExplicitProtocolOverridesRecognizedAlias(t *testing.T) {
	for _, platform := range []string{service.PlatformOpenAI, service.PlatformComposite} {
		t.Run(platform, func(t *testing.T) {
			h, _, groups := newVideoStudioForTest(t, nil)
			groups.groups[0].Platform = platform
			h.gateway = newGatewayModelsHandlerForTest(&gatewayModelsAccountRepoStub{byGroup: map[int64][]service.Account{1: {{ID: 1, Platform: service.PlatformOpenAI, Type: service.AccountTypeAPIKey, Credentials: map[string]any{"base_url": "https://muse.example/v1", "model_mapping": map[string]any{"grok-imagine-video": "muse-video"}}}}}}, &gatewayModelsRouteRepoStub{})
			require.NoError(t, h.settings.SetVideoStudioConfig(context.Background(), service.VideoStudioConfig{Enabled: true, Models: []service.ImageStudioModelConfig{{GroupID: 1, Model: "grok-imagine-video", Protocol: "muse", Enabled: true}}}))
			catalog, err := h.videoCatalog(context.Background(), 7, false)
			require.NoError(t, err)
			require.Len(t, catalog.Models, 1)
			require.Equal(t, "muse", catalog.Models[0].Protocol)
		})
	}
}

func TestMuseVideoStudioCatalog(t *testing.T) {
	h, _, groups := newVideoStudioForTest(t, nil)
	groups.groups[0].Platform = service.PlatformOpenAI
	h.gateway = newGatewayModelsHandlerForTest(&gatewayModelsAccountRepoStub{byGroup: map[int64][]service.Account{1: {{ID: 1, Platform: service.PlatformOpenAI, Type: service.AccountTypeAPIKey, Credentials: map[string]any{"base_url": "https://muse.example/v1", "model_mapping": map[string]any{"muse-video": "muse-video"}}}}}})
	catalog, err := h.videoCatalog(context.Background(), 7, false)
	require.NoError(t, err)
	require.Len(t, catalog.Models, 1)
	require.Equal(t, "muse", catalog.Models[0].Protocol)
	require.NoError(t, h.settings.SetVideoStudioConfig(context.Background(), service.VideoStudioConfig{Enabled: true, Models: []service.ImageStudioModelConfig{{GroupID: 1, Model: "muse-video", Protocol: "muse", Enabled: true}}}))
}

func TestMuseVideoStudioPayload(t *testing.T) {
	input := videoStudioRequest{Prompt: "pan", Seconds: 6, Resolution: "720", Size: "1280x720", Mode: "reference", References: []string{testStudioPNG, testStudioPNG}}
	body, err := buildVideoStudioRequest("muse-video", input)
	require.NoError(t, err)
	var payload map[string]any
	require.NoError(t, json.Unmarshal(body, &payload))
	require.Len(t, payload["images"], 2)
	require.Equal(t, "16:9", payload["size"])
	require.NotContains(t, payload, "generate_audio")
	require.NotContains(t, payload, "last_frame")
	aliasBody, err := buildVideoStudioProtocolRequest("cinema", "muse", input)
	require.NoError(t, err)
	require.Contains(t, string(aliasBody), `"model":"cinema"`)
	input.Seconds = 16
	_, err = buildVideoStudioRequest("muse-video", input)
	require.Error(t, err)
	input.Seconds = 6
	input.Size = "720x720"
	_, err = buildVideoStudioRequest("muse-video", input)
	require.Error(t, err)
	input.Size = "1280x720"
	input.Mode = "frames"
	_, err = buildVideoStudioRequest("muse-video", input)
	require.Error(t, err)
}
