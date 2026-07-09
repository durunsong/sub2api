package service

import (
	"context"

	"github.com/Wei-Shaw/sub2api/internal/pkg/kirocooldown"
	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
)

func (s *GatewayService) isKiroRuntimeSchedulable(ctx context.Context, account *Account) bool {
	if account == nil || account.Platform != PlatformKiro || account.Type != AccountTypeOAuth || s == nil || s.kiroCooldownStore == nil {
		return true
	}
	state, err := s.getKiroCooldownState(ctx, buildKiroAccountKey(account))
	if err != nil {
		return true
	}
	return state == nil || !state.Active
}

func (s *GatewayService) tryRecoverKiroCooldownPool(ctx context.Context, accounts []Account, requestedModel string, excludedIDs map[int64]struct{}, allowMixedScheduling bool) bool {
	if s == nil || s.kiroCooldownStore == nil || ctx.Value(kiroCooldownRecoveryAttemptedKey) == true {
		return false
	}
	tokenKeys := s.kiroTransientCooldownRecoveryKeys(ctx, accounts, requestedModel, excludedIDs, allowMixedScheduling)
	if len(tokenKeys) == 0 {
		return false
	}
	cleared, err := s.kiroCooldownStore.ClearEarliestTransientCooldown(ctx, tokenKeys)
	if err != nil {
		logger.LegacyPrintf("service.gateway", "Kiro cooldown pool recovery failed: %v", err)
		return false
	}
	if cleared {
		logger.LegacyPrintf("service.gateway", "Kiro cooldown pool recovery cleared one transient cooldown")
	}
	return cleared
}

func (s *GatewayService) kiroTransientCooldownRecoveryKeys(ctx context.Context, accounts []Account, requestedModel string, excludedIDs map[int64]struct{}, allowMixedScheduling bool) []string {
	tokenKeys := make([]string, 0, len(accounts))
	eligible := 0
	for i := range accounts {
		acc := &accounts[i]
		if acc == nil || acc.Platform != PlatformKiro || acc.Type != AccountTypeOAuth {
			if allowMixedScheduling {
				continue
			}
			return nil
		}
		if _, excluded := excludedIDs[acc.ID]; excluded {
			continue
		}
		if !acc.IsSchedulable() {
			continue
		}
		if requestedModel != "" && !s.isModelSupportedByAccountWithContext(ctx, acc, requestedModel) {
			continue
		}
		if !s.isAccountSchedulableForQuota(acc) ||
			!s.isAccountSchedulableForWindowCost(ctx, acc, false) ||
			!s.isAccountSchedulableForRPM(ctx, acc, false) {
			continue
		}
		eligible++
		state, err := s.getKiroCooldownState(ctx, buildKiroAccountKey(acc))
		if err != nil || state == nil || !state.Active {
			return nil
		}
		if state.Reason != kirocooldown.CooldownReason429 {
			return nil
		}
		tokenKeys = append(tokenKeys, buildKiroAccountKey(acc))
	}
	if eligible == 0 || len(tokenKeys) != eligible {
		return nil
	}
	return tokenKeys
}
