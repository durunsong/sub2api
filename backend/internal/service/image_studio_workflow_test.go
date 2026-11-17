package service

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestImageStudioWorkflowUnknownVisionAndKnownMediaAreConservative(t *testing.T) {
	for _, m := range []string{"gpt-6-unknown-future", "gpt-5.3-codex-spark", "deepseek-chat"} {
		require.False(t, studioWorkflowReferenceCapability(&Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey}, m), m)
	}
	for _, m := range []string{"gpt-image-2", "grok-imagine-image", "grok-imagine-video", "text-embedding-3-small", "gpt-audio", "sora-2", "gemini-live", "unknown-custom"} {
		text := studioWorkflowModelCapabilities(m)
		require.False(t, text, m)
	}
}
func TestImageStudioWorkflowCatalogDoesNotInventModelsWithoutAccounts(t *testing.T) {
	svc := &GatewayService{accountRepo: &studioWorkflowAccountRepo{}}
	models, err := svc.ImageStudioWorkflowCapabilities(context.Background(), &Group{ID: 1, Platform: PlatformOpenAI}, []string{"gpt-6.1-sol"})
	require.NoError(t, err)
	require.Empty(t, models)
}

type studioWorkflowAccountRepo struct{ AccountRepository }

func (*studioWorkflowAccountRepo) ListSchedulableByGroupID(context.Context, int64) ([]Account, error) {
	return nil, nil
}

func TestImageStudioWorkflowLegacyAndUnknownNativeModelsDoNotPromiseVision(t *testing.T) {
	for _, tc := range []struct{ platform, model string }{{PlatformAnthropic, "claude-2"}, {PlatformAnthropic, "claude-unknown-future"}, {PlatformGemini, "gemini-1.0-pro"}, {PlatformGemini, "gemini-unknown-future"}} {
		require.False(t, studioWorkflowReferenceCapability(&Account{Platform: tc.platform, Type: AccountTypeAPIKey}, tc.model), tc.model)
	}
}
