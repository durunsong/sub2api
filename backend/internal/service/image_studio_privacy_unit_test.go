//go:build unit

package service

import (
	"context"
	"net/http"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/stretchr/testify/require"
)

func TestEphemeralAccountErrorsPreserveQuarantineWithoutCreativeContent(t *testing.T) {
	for _, status := range []int{401, 403} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			repo := &rateLimitAccountRepoStub{}
			svc := NewRateLimitService(repo, nil, &config.Config{}, nil, nil)
			svc.SetOpenAI403CounterCache(&openAI403CounterCacheStub{counts: []int64{1}})
			account := &Account{ID: 1, Platform: PlatformOpenAI, Type: AccountTypeAPIKey}
			blocked := svc.HandleUpstreamError(ctxkey.WithNoContentPersistence(context.Background()), account, status, http.Header{}, []byte(`{"error":{"message":"private prompt and image"}}`))
			require.True(t, blocked)
			require.NotContains(t, repo.lastErrorMsg+repo.lastTempReason, "private")
			require.Equal(t, 1, repo.setErrorCalls+repo.tempCalls)
		})
	}
}

func TestEphemeralCustomRulePreservesCooldownWithoutEcho(t *testing.T) {
	repo := &rateLimitAccountRepoStub{}
	svc := NewRateLimitService(repo, nil, &config.Config{}, nil, nil)
	ctx := ctxkey.WithNoContentPersistence(context.Background())
	blocked := svc.triggerTempUnschedulable(ctx, &Account{ID: 1, Platform: PlatformOpenAI, Type: AccountTypeAPIKey}, TempUnschedulableRule{DurationMinutes: 1}, 2, 500, "private keyword", []byte("private response"))
	require.True(t, blocked)
	require.Equal(t, 1, repo.tempCalls)
	require.NotContains(t, repo.lastTempReason, "private")
	require.Contains(t, repo.lastTempReason, "500")
}
