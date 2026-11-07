//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/adminaccess"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/stretchr/testify/require"
)

// Kiro API Key 账号现已支持用量查询(与 OAuth 同路径,直连 q.{region}.amazonaws.com)。
// 缺少 api_key 凭据时优雅降级:返回非空 UsageInfo(内含错误),而非整体报错。
func TestAccountUsageService_GetUsage_KiroAPIKeySupported(t *testing.T) {
	account := &Account{
		ID:       9101,
		Platform: PlatformKiro,
		Type:     AccountTypeAPIKey,
	}
	repo := &mockAccountRepoForGemini{accountsByID: map[int64]*Account{account.ID: account}}
	svc := NewAccountUsageService(repo, nil, nil, nil, nil, nil, nil, NewUsageCache(), nil, nil)

	usage, err := svc.GetUsage(context.Background(), account.ID)
	require.NoError(t, err)
	require.NotNil(t, usage)
	require.Equal(t, "active", usage.Source)
}

func TestAccountUsageService_GetPassiveUsage_KiroAPIKeySupported(t *testing.T) {
	account := &Account{
		ID:       9102,
		Platform: PlatformKiro,
		Type:     AccountTypeAPIKey,
	}
	repo := &mockAccountRepoForGemini{accountsByID: map[int64]*Account{account.ID: account}}
	svc := NewAccountUsageService(repo, nil, nil, nil, nil, nil, nil, NewUsageCache(), nil, nil)

	usage, err := svc.GetPassiveUsage(context.Background(), account.ID)
	require.NoError(t, err)
	require.NotNil(t, usage)
	require.Equal(t, "passive", usage.Source)
}

func TestAccountUsageService_GetPassiveUsage_KiroRequiresScopedAccess(t *testing.T) {
	account := &Account{ID: 9103, Platform: PlatformKiro, Type: AccountTypeAPIKey}
	repo := &mockAccountRepoForGemini{accountsByID: map[int64]*Account{account.ID: account}}
	svc := NewAccountUsageService(repo, nil, nil, nil, nil, nil, nil, NewUsageCache(), nil, nil)
	ctx := adminaccess.WithPrincipal(context.Background(), &adminaccess.Principal{})
	usage, err := svc.GetPassiveUsage(ctx, account.ID)
	require.Nil(t, usage)
	require.Equal(t, 404, infraerrors.Code(err))
}
