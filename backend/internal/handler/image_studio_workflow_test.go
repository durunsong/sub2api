package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	middleware "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestImageStudioWorkflowCatalogUsesAuthorizedTextAccounts(t *testing.T) {
	h, users, groups, _ := newStudioForTest(t, "", nil)
	groups.groups[0].AllowImageGeneration = false
	h.gateway = newGatewayModelsHandlerForTest(&gatewayModelsAccountRepoStub{byGroup: map[int64][]service.Account{1: {{ID: 1, Platform: service.PlatformOpenAI, Type: service.AccountTypeAPIKey, Credentials: map[string]any{"model_mapping": map[string]any{"gpt-4o": "gpt-4o", "writer": "gpt-4o", "picture": "gpt-image-2", "gpt-audio": "gpt-audio", "text-embedding-3-small": "text-embedding-3-small"}}}}, 2: {{ID: 2, Platform: service.PlatformOpenAI, Credentials: map[string]any{"model_mapping": map[string]any{"gpt-4o": "gpt-4o"}}}}}})
	catalog, err := h.workflowCatalog(context.Background(), 7, false)
	require.NoError(t, err)
	require.Len(t, catalog.Models, 2)
	require.Equal(t, "gpt-4o", catalog.Models[0].Model)
	require.True(t, catalog.Models[0].SupportsReferences)
	require.Equal(t, "writer", catalog.Models[1].Model)
	require.NoError(t, h.settings.SetImageStudioConfig(context.Background(), service.ImageStudioConfig{Enabled: true, Models: []service.ImageStudioModelConfig{{GroupID: 1, Model: "writer", Label: "Writer", Enabled: false}}}))
	catalog, err = h.workflowCatalog(context.Background(), 7, false)
	require.NoError(t, err)
	require.Len(t, catalog.Models, 1)
	admin, err := h.workflowCatalog(context.Background(), 0, true)
	require.NoError(t, err)
	require.Len(t, admin.Models, 3)
	require.Equal(t, "Writer", admin.Models[1].Label)
	users.user.AllowedGroups = nil
	catalog, err = h.workflowCatalog(context.Background(), 7, false)
	require.NoError(t, err)
	require.Empty(t, catalog.Models)
}
func workflowRouter(h *ImageStudioHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/workflow", func(c *gin.Context) { c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 7}) }, h.RequestBodyLimit(), h.Workflow)
	return r
}
func workflowRequest(t *testing.T, r http.Handler, payload any) *httptest.ResponseRecorder {
	t.Helper()
	b, err := json.Marshal(payload)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/workflow?ignored=secret", strings.NewReader(string(b)))
	req.Header.Set("Authorization", "Bearer panel-jwt")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
func TestImageStudioWorkflowUsesFixedPrivateGatewayAndRechecksAuthorization(t *testing.T) {
	calls := 0
	relay := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		require.Equal(t, "/v1/chat/completions", r.URL.Path)
		require.Empty(t, r.URL.RawQuery)
		require.True(t, ctxkey.NoContentPersistence(r.Context()))
		require.True(t, strings.HasPrefix(r.Header.Get("Authorization"), "Bearer internal:images:"))
		require.NotContains(t, r.Header.Get("Authorization"), "panel-jwt")
		var body map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		require.Equal(t, "gpt-4", body["model"])
		require.Equal(t, false, body["stream"])
		require.NotContains(t, body, "tools")
		require.NotContains(t, body, "url")
		_, _ = io.WriteString(w, `{"choices":[{"message":{"content":"{\"prompts\":[\"one\",\"two\"]}"},"finish_reason":"stop"}]}`)
	})
	h, users, _, keys := newStudioForTest(t, "", relay)
	r := workflowRouter(h)
	payload := map[string]any{"model_id": studioModelID(1, "gpt-4"), "mode": "series", "prompt": "private brief", "count": 2}
	w := workflowRequest(t, r, payload)
	require.Equal(t, 200, w.Code, w.Body.String())
	require.Contains(t, w.Body.String(), `"prompts":["one","two"]`)
	require.Len(t, keys.keys, 1)
	require.Equal(t, 1, calls)
	users.user.AllowedGroups = nil
	w = workflowRequest(t, r, payload)
	require.Equal(t, 403, w.Code)
	require.Equal(t, 1, calls)
}
func TestImageStudioWorkflowRejectsInvalidInputBeforeForwarding(t *testing.T) {
	calls := 0
	h, _, _, _ := newStudioForTest(t, "", http.HandlerFunc(func(http.ResponseWriter, *http.Request) { calls++ }))
	r := workflowRouter(h)
	for _, edit := range []map[string]any{
		{"url": "https://evil.test"}, {"messages": []any{}}, {"mode": "other"}, {"count": 0}, {"count": 21}, {"count": nil}, {"prompt": " "}, {"prompt": strings.Repeat("中", 16001)},
		{"references": []string{"https://evil.test/p.png"}}, {"references": []string{testStudioPNG}}, {"references": []string{testStudioPNG, testStudioPNG, testStudioPNG, testStudioPNG}},
	} {
		input := map[string]any{"model_id": studioModelID(1, "gpt-4"), "mode": "series", "prompt": "test", "count": 2}
		for k, v := range edit {
			input[k] = v
		}
		w := workflowRequest(t, r, input)
		require.Equal(t, 400, w.Code, w.Body.String())
	}
	require.Zero(t, calls)
}
func TestImageStudioWorkflowOutputIsStrictAndBounded(t *testing.T) {
	for _, bad := range []string{`hello`, `{"prompts":[]}`, `{"prompts":["one"]}`, `{"prompts":["one"," "]}`, `{"prompts":["one","two"],"tools":[]}`, `{"prompts":null}`, `{"prompts":["one","two"]} trailing`, `{"prompts":["one","two"],"prompts":["x","y"]}`, `{"prompts":["` + strings.Repeat("x", 10001) + `","two"]}`} {
		_, err := normalizeStudioWorkflowOutput(bad, "series", 2)
		require.Error(t, err, bad[:min(len(bad), 100)])
	}
	result, err := normalizeStudioWorkflowOutput("```json\n{\"prompts\":[\"one\",\"two\"]}\n```", "series", 2)
	require.NoError(t, err)
	require.Equal(t, []string{"one", "two"}, result.Prompts)
	draft := `{"draft":{"title":"Poster","prompt":"Make {{subject}}","category":"art","tags":["poster"],"seriesPrompts":[]}}`
	result, err = normalizeStudioWorkflowOutput(draft, "workflow", 0)
	require.NoError(t, err)
	require.Equal(t, "Make {{subject}}", result.Draft.Prompt)
	for _, bad := range []string{`{"draft":null}`, strings.Replace(draft, `"Poster"`, `""`, 1), strings.Replace(draft, `"art"`, `"`+strings.Repeat("x", 121)+`"`, 1), strings.Replace(draft, `"tags":["poster"]`, `"tags":null`, 1), strings.Replace(draft, `"seriesPrompts":[]`, `"seriesPrompts":[""]`, 1), strings.Replace(draft, `"category":"art",`, "", 1)} {
		_, err = normalizeStudioWorkflowOutput(bad, "workflow", 0)
		require.Error(t, err)
	}
}
func TestImageStudioWorkflowErrorsAreStableAndCancellationPropagates(t *testing.T) {
	h, _, _, _ := newStudioForTest(t, "", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(502)
		_, _ = io.WriteString(w, "private-output-canary")
	}))
	payload := map[string]any{"model_id": studioModelID(1, "gpt-4"), "mode": "series", "prompt": "test", "count": 2}
	w := workflowRequest(t, workflowRouter(h), payload)
	require.Equal(t, 502, w.Code)
	require.NotContains(t, w.Body.String(), "private-output-canary")
	body, _ := json.Marshal(payload)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	h.relay = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cancel()
		require.ErrorIs(t, r.Context().Err(), context.Canceled)
	})
	req := httptest.NewRequest(http.MethodPost, "/workflow", strings.NewReader(string(body))).WithContext(ctx)
	w = httptest.NewRecorder()
	workflowRouter(h).ServeHTTP(w, req)
	require.Equal(t, 499, w.Code)
}

func TestImageStudioWorkflowSubmissionRejectsHiddenAndDisabledModels(t *testing.T) {
	calls := 0
	h, _, _, _ := newStudioForTest(t, "", http.HandlerFunc(func(http.ResponseWriter, *http.Request) { calls++ }))
	r := workflowRouter(h)
	payload := map[string]any{"model_id": studioModelID(1, "gpt-4"), "mode": "workflow", "prompt": "brief"}
	for _, cfg := range []service.ImageStudioConfig{
		{Enabled: true, Models: []service.ImageStudioModelConfig{{GroupID: 1, Model: "gpt-4", Enabled: false}}},
		{Enabled: false},
	} {
		require.NoError(t, h.settings.SetImageStudioConfig(context.Background(), cfg))
		w := workflowRequest(t, r, payload)
		require.Equal(t, 403, w.Code)
	}
	require.Zero(t, calls)
}
func TestImageStudioWorkflowReferencesAreValidatedAndUsedForUnderstanding(t *testing.T) {
	model := imageStudioWorkflowModel{Model: "gpt-4o", SupportsReferences: true}
	body, err := buildStudioWorkflowRequest(model, imageStudioWorkflowRequest{Mode: "workflow", Prompt: "describe this", References: []string{testStudioPNG}})
	require.NoError(t, err)
	require.Contains(t, string(body), `"type":"image_url"`)
	require.Contains(t, string(body), testStudioPNG)
	require.Contains(t, string(body), "{{variable}}")
	for _, refs := range [][]string{{"https://evil.test/image.png"}, {"data:image/png;base64,aGVsbG8="}, {"data:image/svg+xml;base64,PHN2Zz4="}, {testStudioPNG, testStudioPNG, testStudioPNG, testStudioPNG}, {"data:image/png;base64," + strings.Repeat("A", 14<<20)}} {
		_, err = buildStudioWorkflowRequest(model, imageStudioWorkflowRequest{Mode: "workflow", Prompt: "brief", References: refs})
		require.Error(t, err)
	}
}
func TestImageStudioWorkflowOutputRejectsNullAndExtraneousEnvelopes(t *testing.T) {
	for _, input := range []struct {
		mode, text string
		count      int
	}{
		{"workflow", `{"draft":{"title":"T","prompt":"P","category":null,"tags":[],"seriesPrompts":[]}}`, 0},
		{"workflow", `{"draft":{"title":"T","prompt":"P","category":"","tags":[],"seriesPrompts":[]},"prompts":null}`, 0},
		{"series", `{"prompts":["P"],"draft":null}`, 1},
	} {
		_, err := normalizeStudioWorkflowOutput(input.text, input.mode, input.count)
		require.Error(t, err)
	}
}

func TestImageStudioWorkflowAcceptsEmptyToolCallsAndRejectsActualCalls(t *testing.T) {
	for _, tc := range []struct {
		calls  string
		status int
	}{{"null", 200}, {"[]", 200}, {`[{"id":"call_1","type":"function","function":{"name":"run","arguments":"{}"}}]`, 502}} {
		h, _, _, _ := newStudioForTest(t, "", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = io.WriteString(w, `{"choices":[{"message":{"content":"{\"prompts\":[\"one\"]}","tool_calls":`+tc.calls+`},"finish_reason":"stop"}]}`)
		}))
		w := workflowRequest(t, workflowRouter(h), map[string]any{"model_id": studioModelID(1, "gpt-4"), "mode": "series", "prompt": "brief", "count": 1})
		require.Equal(t, tc.status, w.Code, w.Body.String())
	}
}

func TestImageStudioWorkflowSlashAliasCanSubmitAndBeHidden(t *testing.T) {
	calls := 0
	h, _, _, _ := newStudioForTest(t, "", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		require.Equal(t, "/v1/chat/completions", r.URL.Path)
		var body map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		require.Equal(t, "team/writer", body["model"])
		_, _ = io.WriteString(w, `{"choices":[{"message":{"content":"{\"prompts\":[\"one\"]}"},"finish_reason":"stop"}]}`)
	}))
	h.gateway = newGatewayModelsHandlerForTest(&gatewayModelsAccountRepoStub{byGroup: map[int64][]service.Account{1: {{ID: 1, Platform: service.PlatformOpenAI, Type: service.AccountTypeAPIKey, Credentials: map[string]any{"model_mapping": map[string]any{"team/writer": "gpt-4o", "team/picture": "gpt-image-2"}}}}}})
	catalog, err := h.workflowCatalog(context.Background(), 7, false)
	require.NoError(t, err)
	require.Len(t, catalog.Models, 1)
	require.Equal(t, "team/writer", catalog.Models[0].Model)
	require.True(t, catalog.Models[0].SupportsReferences)
	images, err := h.catalog(context.Background(), 7, false)
	require.NoError(t, err)
	require.Empty(t, images.Models)
	payload := map[string]any{"model_id": studioModelID(1, "team/writer"), "mode": "series", "prompt": "brief", "count": 1}
	require.Equal(t, 200, workflowRequest(t, workflowRouter(h), payload).Code)
	require.NoError(t, h.settings.SetImageStudioConfig(context.Background(), service.ImageStudioConfig{Enabled: true, Models: []service.ImageStudioModelConfig{{GroupID: 1, Model: "team/writer", Label: "Team writer", Enabled: false}}}))
	catalog, err = h.workflowCatalog(context.Background(), 7, false)
	require.NoError(t, err)
	require.Empty(t, catalog.Models)
	admin, err := h.workflowCatalog(context.Background(), 0, true)
	require.NoError(t, err)
	require.Len(t, admin.Models, 1)
	require.Equal(t, "Team writer", admin.Models[0].Label)
	require.Equal(t, 403, workflowRequest(t, workflowRouter(h), payload).Code)
	require.Equal(t, 1, calls)
}

func TestImageStudioWorkflowCompositeExplicitAliasUsesTarget(t *testing.T) {
	calls := 0
	h, _, groups, _ := newStudioForTest(t, "", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		var body map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		require.Equal(t, "writer", body["model"])
		_, _ = io.WriteString(w, `{"choices":[{"message":{"content":"{\"prompts\":[\"one\"]}"},"finish_reason":"stop"}]}`)
	}))
	groups.groups[0].Platform = service.PlatformComposite
	routes := []service.CompositeModelRoute{}
	for _, tc := range []struct {
		name, target, platform, endpoint string
		enabled                          bool
	}{
		{"writer", "gpt-4o", service.PlatformOpenAI, service.CompositeRouteEndpointChatCompletions, true},
		{"picture", "gpt-image-2", service.PlatformOpenAI, service.CompositeRouteEndpointChatCompletions, true},
		{"wrong-platform", "gpt-4o", service.PlatformGemini, service.CompositeRouteEndpointChatCompletions, true},
		{"wrong-endpoint", "gpt-4o", service.PlatformOpenAI, service.CompositeRouteEndpointResponses, true},
		{"disabled", "gpt-4o", service.PlatformOpenAI, service.CompositeRouteEndpointChatCompletions, false},
	} {
		routes = append(routes, service.CompositeModelRoute{GroupID: 1, PublicModel: tc.name, UpstreamModel: tc.target, TargetPlatform: tc.platform, Endpoint: tc.endpoint, Enabled: tc.enabled, MatchType: service.CompositeRouteMatchExact})
	}
	h.gateway = newGatewayModelsHandlerForTest(&gatewayModelsAccountRepoStub{byGroup: map[int64][]service.Account{1: {{ID: 1, Platform: service.PlatformOpenAI, Type: service.AccountTypeAPIKey, Credentials: map[string]any{"model_mapping": map[string]any{"gpt-4o": "gpt-4o", "gpt-image-2": "gpt-image-2"}}}}}}, &gatewayModelsRouteRepoStub{routes: routes})
	catalog, err := h.workflowCatalog(context.Background(), 7, false)
	require.NoError(t, err)
	var names []string
	for _, model := range catalog.Models {
		names = append(names, model.Model)
		require.True(t, model.SupportsReferences)
	}
	require.Equal(t, []string{"gpt-4o", "writer"}, names)
	groups.groups[0].ModelAllowlist = service.GroupModelAllowlist{Enabled: true, Models: []string{"writer"}}
	catalog, err = h.workflowCatalog(context.Background(), 7, false)
	require.NoError(t, err)
	require.Len(t, catalog.Models, 1)
	require.Equal(t, "writer", catalog.Models[0].Model)
	payload := map[string]any{"model_id": studioModelID(1, "writer"), "mode": "series", "prompt": "brief", "count": 1}
	require.Equal(t, 200, workflowRequest(t, workflowRouter(h), payload).Code)
	require.Equal(t, 1, calls)
	require.NoError(t, h.settings.SetImageStudioConfig(context.Background(), service.ImageStudioConfig{Enabled: true, Models: []service.ImageStudioModelConfig{{GroupID: 1, Model: "writer", Enabled: false}}}))
	require.Equal(t, 403, workflowRequest(t, workflowRouter(h), payload).Code)
	require.Equal(t, 1, calls)
}

func TestImageStudioWorkflowUnmappedNativeTextAccountsUseOwnCatalog(t *testing.T) {
	for _, tc := range []struct{ platform, model string }{{service.PlatformDeepseek, "deepseek-flash"}, {service.PlatformKimi, "kimi-latest"}, {service.PlatformZhipu, "glm-4.7"}, {service.PlatformMiniMax, "MiniMax-M3"}} {
		t.Run(tc.platform, func(t *testing.T) {
			calls := 0
			h, _, groups, _ := newStudioForTest(t, "", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				require.Equal(t, "/v1/chat/completions", r.URL.Path)
				var body map[string]any
				require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
				require.Equal(t, tc.model, body["model"])
				_, _ = io.WriteString(w, `{"choices":[{"message":{"content":"{\"prompts\":[\"one\"]}"},"finish_reason":"stop"}]}`)
			}))
			groups.groups[0].Platform = tc.platform
			h.gateway = newGatewayModelsHandlerForTest(&gatewayModelsAccountRepoStub{byGroup: map[int64][]service.Account{1: {{ID: 1, Platform: tc.platform, Type: service.AccountTypeAPIKey}}}})
			catalog, err := h.workflowCatalog(context.Background(), 7, false)
			require.NoError(t, err)
			require.NotEmpty(t, catalog.Models)
			var names []string
			for _, m := range catalog.Models {
				names = append(names, m.Model)
				require.NotContains(t, m.Model, "claude")
				require.False(t, m.SupportsReferences)
			}
			require.Contains(t, names, tc.model)
			payload := map[string]any{"model_id": studioModelID(1, tc.model), "mode": "series", "prompt": "brief", "count": 1}
			w := workflowRequest(t, workflowRouter(h), payload)
			require.Equal(t, 200, w.Code, w.Body.String())
			require.Equal(t, 1, calls)
		})
	}
}

func TestImageStudioWorkflowCompositeAmbiguousAliasIsUnavailable(t *testing.T) {
	h, _, groups, _ := newStudioForTest(t, "", nil)
	groups.groups[0].Platform = service.PlatformComposite
	h.gateway = newGatewayModelsHandlerForTest(&gatewayModelsAccountRepoStub{byGroup: map[int64][]service.Account{1: {
		{ID: 1, Platform: service.PlatformOpenAI, Type: service.AccountTypeAPIKey, Credentials: map[string]any{"model_mapping": map[string]any{"shared-writer": "gpt-4o"}}},
		{ID: 2, Platform: service.PlatformGemini, Type: service.AccountTypeAPIKey, Credentials: map[string]any{"model_mapping": map[string]any{"shared-writer": "gemini-2.5-pro"}}},
	}}}, &gatewayModelsRouteRepoStub{})
	catalog, err := h.workflowCatalog(context.Background(), 7, false)
	require.NoError(t, err)
	require.Empty(t, catalog.Models)
}
