package service

import (
	"context"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/claude"
	"github.com/Wei-Shaw/sub2api/internal/pkg/geminicli"
	"github.com/Wei-Shaw/sub2api/internal/pkg/openai"
)

// ImageStudioWorkflowCapabilities uses current schedulable accounts, rather than
// treating an empty/failed public model listing as proof that defaults are usable.
func (s *GatewayService) ImageStudioWorkflowCapabilities(ctx context.Context, group *Group, models []string) (map[string]bool, error) {
	accounts, err := s.accountRepo.ListSchedulableByGroupID(ctx, group.ID)
	if err != nil {
		return nil, err
	}
	out := make(map[string]bool)
	for _, model := range models {
		requestedModel, targetPlatform := model, group.Platform
		if group.Platform == PlatformComposite {
			decision, matched, err := s.resolveCompositeRouteDecision(ctx, group, model, CompositeRouteEndpointChatCompletions)
			if err != nil {
				return nil, err
			}
			if !matched {
				continue
			}
			requestedModel, targetPlatform = decision.UpstreamModel, decision.TargetPlatform
		}
		found, text, references := false, true, true
		for i := range accounts {
			account := &accounts[i]
			if account.Platform == PlatformTypeSafe || (account.Platform != targetPlatform && !mixedListingAccountAllowed(targetPlatform, account)) || !account.IsModelSupported(requestedModel) {
				continue
			}
			canonical := account.GetMappedModel(requestedModel)
			if group.Platform == PlatformComposite {
				if owner, known := DetectModelPlatform(canonical); known && owner != account.Platform && !(owner == PlatformAnthropic && (account.Platform == PlatformKiro || account.Platform == PlatformAntigravity)) {
					continue
				}
			}
			isText := studioWorkflowModelCapabilities(canonical)
			vision := studioWorkflowReferenceCapability(account, canonical)
			found = true
			text = text && isText
			references = references && vision
		}
		if found && text {
			out[model] = references
		}
	}
	return out, nil
}

func studioWorkflowModelCapabilities(model string) bool {
	m := strings.ToLower(strings.TrimSpace(model))
	if i := strings.LastIndex(m, "/"); i >= 0 {
		m = m[i+1:]
	}
	for _, excluded := range []string{"image", "imagine", "video", "audio", "voice", "embedding", "rerank", "moderation", "whisper", "tts", "dall-e", "sora", "veo", "imagen", "live", "realtime", "computer-use"} {
		if strings.Contains(m, excluded) {
			return false
		}
	}
	// Aliases must resolve to a known text family. Unknown vision support stays false.
	text := false
	for _, prefix := range []string{"gpt-", "chatgpt-", "o1", "o3", "o4", "claude-", "gemini-", "grok-", "deepseek-", "glm-", "kimi-", "minimax-", "qwen", "llama", "mistral", "gemma"} {
		if strings.HasPrefix(m, prefix) {
			text = true
			break
		}
	}
	return text
}

// Reuse gateway/account modality metadata and curated native catalogs. A family
// prefix alone is insufficient to promise vision for an unknown future model.
func studioWorkflowReferenceCapability(account *Account, model string) bool {
	if account == nil {
		return false
	}
	text := studioWorkflowModelCapabilities(model)
	if !text {
		return false
	}
	if _, ok := account.GetUpstreamModelMetadata(model); ok {
		return accountCodexModelSupportsImageInput(account, model)
	}
	m := strings.ToLower(strings.TrimSpace(model))
	if strings.Contains(m, "spark") {
		return false
	}
	switch account.Platform {
	case PlatformOpenAI, PlatformDeepseek, PlatformOpenCodeGo:
		known := false
		for _, id := range openai.DefaultModelIDs() {
			if m == id {
				known = true
				break
			}
		}
		switch m {
		case "gpt-4o", "gpt-4o-mini", "gpt-4.1", "gpt-4.1-mini", "gpt-4.1-nano", "gpt-4-turbo", "gpt-4-vision-preview":
			known = true
		}
		return known && accountCodexModelSupportsImageInput(account, model)
	case PlatformGrok:
		return accountCodexModelSupportsImageInput(account, model)
	case PlatformAnthropic, PlatformKiro, PlatformAntigravity:
		for _, id := range claude.DefaultModelIDs() {
			if m == id {
				return true
			}
		}
		if account.Platform != PlatformAntigravity {
			return false
		}
		fallthrough
	case PlatformGemini:
		for _, entry := range append(append([]geminicli.Model{}, geminicli.DefaultModels...), geminicli.GoogleOneModels...) {
			if m == entry.ID {
				return true
			}
		}
	}
	return false
}

// ImageStudioWorkflowNativeModelIDs supplies workflow-only defaults for native
// providers; the public gateway's historical fallback is intentionally unchanged.
func ImageStudioWorkflowNativeModelIDs(platform string) []string {
	switch platform {
	case PlatformDeepseek:
		return append([]string(nil), deepseekServableModels...)
	case PlatformKimi:
		// Existing native catalog: frontend/src/composables/useModelWhitelist.ts.
		return []string{"kimi-latest", "kimi-for-coding", "kimi-k2"}
	case PlatformZhipu:
		return []string{"glm-4.7", "glm-5.3", "glm-5.3-flash"}
	case PlatformMiniMax:
		return []string{"MiniMax-M3", "MiniMax-M2.7", "MiniMax-M2.5"}
	default:
		return nil
	}
}
