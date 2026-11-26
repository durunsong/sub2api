package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// Video configuration is independent: saving either tab cannot overwrite the other.
type VideoStudioConfig struct {
	Enabled bool                     `json:"enabled"`
	Models  []ImageStudioModelConfig `json:"models"`
}

func (s *SettingService) GetVideoStudioConfig(ctx context.Context) (VideoStudioConfig, error) {
	cfg := VideoStudioConfig{Enabled: true, Models: []ImageStudioModelConfig{}}
	raw, err := s.settingRepo.GetValue(ctx, "video_studio_config")
	if errors.Is(err, ErrSettingNotFound) || err == nil && strings.TrimSpace(raw) == "" {
		return cfg, nil
	}
	if err != nil {
		return cfg, err
	}
	if err = json.Unmarshal([]byte(raw), &cfg); err != nil {
		return cfg, fmt.Errorf("invalid video studio configuration")
	}
	if cfg.Models == nil {
		cfg.Models = []ImageStudioModelConfig{}
	}
	return cfg, nil
}
func (s *SettingService) SetVideoStudioConfig(ctx context.Context, cfg VideoStudioConfig) error {
	if len(cfg.Models) > 500 {
		return fmt.Errorf("too many model overrides")
	}
	seen := map[string]bool{}
	for i := range cfg.Models {
		m := &cfg.Models[i]
		m.Model = strings.TrimSpace(m.Model)
		m.Label = strings.TrimSpace(m.Label)
		if m.GroupID <= 0 || m.Model == "" || len(m.Model) > 200 || len(m.Label) > 100 || strings.ContainsAny(m.Model, "\\?#\r\n\x00*") || m.Protocol != "" && m.Protocol != "grok" {
			return fmt.Errorf("invalid video model")
		}
		id := fmt.Sprintf("%d:%s", m.GroupID, m.Model)
		if seen[id] {
			return fmt.Errorf("duplicate video model")
		}
		seen[id] = true
	}
	raw, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	return s.settingRepo.Set(ctx, "video_studio_config", string(raw))
}
