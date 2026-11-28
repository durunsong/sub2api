package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

const ImageStudioKeyPrefix = "internal:images:"
const imageStudioSettingKey = "image_studio_config"

type imageStudioAuthContextKey struct{}

func WithImageStudioCredential(ctx context.Context, key string) context.Context {
	return context.WithValue(ctx, imageStudioAuthContextKey{}, key)
}

func IsImageStudioKey(key string) bool { return strings.HasPrefix(key, ImageStudioKeyPrefix) }

type ImageStudioModelConfig struct {
	GroupID  int64  `json:"group_id"`
	Model    string `json:"model"`
	Enabled  bool   `json:"enabled"`
	Label    string `json:"label"`
	Protocol string `json:"protocol,omitempty"`
}
type ImageStudioConfig struct {
	Enabled bool                     `json:"enabled"`
	Models  []ImageStudioModelConfig `json:"models"`
}

func (s *SettingService) GetImageStudioConfig(ctx context.Context) (ImageStudioConfig, error) {
	cfg := ImageStudioConfig{Enabled: true, Models: []ImageStudioModelConfig{}}
	raw, err := s.settingRepo.GetValue(ctx, imageStudioSettingKey)
	if errors.Is(err, ErrSettingNotFound) {
		return cfg, nil
	}
	if err != nil {
		return cfg, err
	}
	if strings.TrimSpace(raw) == "" {
		return cfg, nil
	}
	if err = json.Unmarshal([]byte(raw), &cfg); err != nil {
		return cfg, fmt.Errorf("invalid image studio configuration")
	}
	if cfg.Models == nil {
		cfg.Models = []ImageStudioModelConfig{}
	}
	return cfg, nil
}
func (s *SettingService) SetImageStudioConfig(ctx context.Context, cfg ImageStudioConfig) error {
	if len(cfg.Models) > 500 {
		return fmt.Errorf("too many model overrides")
	}
	seen := map[string]bool{}
	for i := range cfg.Models {
		m := &cfg.Models[i]
		m.Model = strings.TrimSpace(m.Model)
		m.Label = strings.TrimSpace(m.Label)
		if m.GroupID <= 0 || m.Model == "" || len(m.Model) > 200 || len(m.Label) > 100 || strings.ContainsAny(m.Model, "\\?#\r\n\x00") {
			return fmt.Errorf("invalid studio model")
		}
		if m.Protocol != "" && m.Protocol != "images" && m.Protocol != "gemini" {
			return fmt.Errorf("invalid image protocol")
		}
		id := fmt.Sprintf("%d:%s", m.GroupID, m.Model)
		if seen[id] {
			return fmt.Errorf("duplicate studio model")
		}
		seen[id] = true
	}
	raw, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	return s.settingRepo.Set(ctx, imageStudioSettingKey, string(raw))
}

func (s *APIKeyService) ImageStudioGroups(ctx context.Context, userID int64, admin bool) ([]Group, error) {
	if admin {
		return s.groupRepo.ListActive(ctx)
	}
	return s.GetAvailableGroups(ctx, userID)
}

// EnsureImageStudioKey creates an accounting identity, never a client credential.
// The existing unique key constraint makes concurrent first requests converge.
func (s *APIKeyService) EnsureImageStudioKey(ctx context.Context, userID, groupID int64) (*APIKey, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	group, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if !user.IsActive() || group.Status != StatusActive || !s.canUserBindGroup(ctx, user, group) {
		return nil, ErrGroupNotAllowed
	}
	if s.cfg == nil || s.cfg.JWT.Secret == "" {
		return nil, errors.New("image studio identity unavailable")
	}
	mac := hmac.New(sha256.New, []byte(s.cfg.JWT.Secret))
	fmt.Fprintf(mac, "sub2api:image-studio:v1:%d:%d", userID, groupID)
	credential := ImageStudioKeyPrefix + hex.EncodeToString(mac.Sum(nil))
	key, err := s.apiKeyRepo.GetByKey(ctx, credential)
	if errors.Is(err, ErrAPIKeyNotFound) {
		key = &APIKey{UserID: userID, GroupID: &groupID, Key: credential, Name: "Image Studio", Status: StatusActive}
		err = s.apiKeyRepo.Create(ctx, key)
		if errors.Is(err, ErrAPIKeyExists) {
			key, err = s.apiKeyRepo.GetByKey(ctx, credential)
		}
	}
	if err != nil {
		return nil, err
	}
	if key.UserID != userID || key.GroupID == nil || *key.GroupID != groupID {
		return nil, ErrGroupNotAllowed
	}
	key.User = user
	key.Group = group
	return key, nil
}

type imageStudioModelContextKey struct{}
type imageStudioModelResolution struct{ Public, Canonical string }

func WithImageStudioModel(ctx context.Context, public, canonical string) context.Context {
	return context.WithValue(ctx, imageStudioModelContextKey{}, imageStudioModelResolution{Public: public, Canonical: canonical})
}

// imageStudioValidationModel resolves only server-selected studio aliases.
// Forwarding must still validate the account's actual mapped target separately.
func imageStudioValidationModel(ctx context.Context, model string) string {
	if resolved, ok := ctx.Value(imageStudioModelContextKey{}).(imageStudioModelResolution); ok && resolved.Canonical != "" {
		if resolved.Public == model {
			return resolved.Canonical
		}
		// Composite middleware can rewrite the same authorized studio selection.
		public, _ := RequestedPublicModelFromContext(ctx)
		upstream, _ := ResolvedUpstreamModelFromContext(ctx)
		if public == resolved.Public && upstream == model {
			return resolved.Canonical
		}
	}
	return model
}

// ImageStudioMappedModel exposes an alias only when every eligible account maps
// it to the same image family. Text mappings and ambiguous aliases stay hidden.
func (s *GatewayService) ImageStudioMappedModel(ctx context.Context, groupID int64, platform, model string) (string, bool) {
	if s == nil || s.accountRepo == nil {
		return "", true
	}
	if platform == PlatformComposite {
		decision, matched, err := s.resolveCompositeRouteDecision(ctx, &Group{ID: groupID, Platform: platform}, model, CompositeRouteEndpointImages)
		if err != nil || !matched {
			return "", false
		}
		platform, model = decision.TargetPlatform, decision.UpstreamModel
	}
	accounts, err := s.accountRepo.ListSchedulableByGroupID(ctx, groupID)
	if err != nil {
		return "", false
	}
	canonical := ""
	ambiguous, hasMuse := false, IsMuseImageModel(model)
	for i := range accounts {
		account := &accounts[i]
		if account.Platform != platform || !account.IsModelSupported(model) {
			continue
		}
		mapped, matched := account.ResolveMappedModel(model)
		hasMuse = hasMuse || IsMuseImageModel(mapped)
		if !matched || (!IsGPTImageGenerationModel(mapped) && !isGeminiCompatibleImageModel(mapped) && !isGrokImageGenerationModel(mapped) && !IsMuseImageModel(mapped)) {
			ambiguous = true
			continue
		}
		if canonical != "" && canonical != mapped {
			ambiguous = true
		}
		canonical = mapped
	}
	if ambiguous {
		return "", !hasMuse
	}
	return canonical, true
}

// ImageStudioPricing uses the same image/token decision and image cost resolver
// as settlement. Prices already include the user/image multiplier and must not
// be multiplied again. Actual output and pricing time determine settlement.
type ImageStudioPricing struct {
	Currency       string   `json:"currency"`
	Mode           string   `json:"mode"`
	RateMultiplier float64  `json:"rate_multiplier"`
	ImagePrice1K   *float64 `json:"image_price_1k"`
	ImagePrice2K   *float64 `json:"image_price_2k"`
	ImagePrice4K   *float64 `json:"image_price_4k"`
	Estimated      bool     `json:"estimated"`
}

func (s *GatewayService) ImageStudioPricing(ctx context.Context, userID int64, group *Group, model string, canonicalModels ...string) ImageStudioPricing {
	key := &APIKey{GroupID: &group.ID, Group: group}
	canonical := model
	if len(canonicalModels) > 0 && strings.TrimSpace(canonicalModels[0]) != "" {
		canonical = canonicalModels[0]
	}
	fields := ChannelUsageFields{OriginalModel: model, ChannelMappedModel: model}
	if s.channelService != nil {
		fields = s.channelService.ResolveChannelMapping(ctx, group.ID, model).ToUsageFields(model, canonical)
	}
	model = s.recordUsageBillingModel(ctx, key, &ForwardResult{Model: model, UpstreamModel: canonical}, fields)
	var user *User
	if userID > 0 {
		user = &User{ID: userID}
	}
	tokenRate, imageRate, _ := s.recordUsageMultipliers(ctx, key, user, time.Time{})
	quote := ImageStudioPricing{Currency: "USD", Mode: "image", RateMultiplier: imageRate, Estimated: true}
	if IsMuseImageModel(canonical) && ValidateMuseImagePricing(group) != nil {
		return quote
	}
	resolved := s.resolveChannelPricing(ctx, model, key)
	if !imageCountBillingApplies(key, true) && resolved != nil && resolved.Mode == BillingModeToken {
		quote.Mode = "token"
		quote.RateMultiplier = tokenRate
		return quote
	}
	for _, tier := range []struct {
		size   string
		target **float64
	}{{"1K", &quote.ImagePrice1K}, {"2K", &quote.ImagePrice2K}, {"4K", &quote.ImagePrice4K}} {
		if s.billingService == nil {
			if price := group.GetImagePrice(tier.size); price != nil {
				value := *price * imageRate
				*tier.target = &value
			}
			continue
		}
		cost := s.calculateImageCost(ctx, &ForwardResult{ImageCount: 1, ImageEndpoint: true, ImageSize: tier.size}, key, model, imageRate)
		if cost != nil && cost.BillingMode != string(BillingModeToken) {
			value := cost.ActualCost
			*tier.target = &value
		}
	}
	return quote
}
