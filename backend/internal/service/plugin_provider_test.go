package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	pluginv1 "github.com/Wei-Shaw/sub2api/pkg/pluginapi/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestProvidePluginManager_AccountDirectoryCapability(t *testing.T) {
	manager := ProvidePluginManager(nil, nil, &config.Config{}, PluginHostInfo{}, newFakePluginKVStore(), &OpenAIGatewayService{})
	installation := &PluginInstallation{PluginKey: "test.directory", Manifest: PluginManifest{
		Capabilities: []PluginCapability{{ID: PluginCapabilityOpenAIOAuthOutbound, Platform: PlatformOpenAI, AccountType: AccountTypeOAuth}},
	}}
	accounts, err := manager.buildHostServices(installation).ListAccounts(context.Background(), &pluginv1.ListAccountsRequest{})
	require.NoError(t, err)
	require.Empty(t, accounts.AccountIds)

	installation.Manifest.Capabilities = nil
	_, err = manager.buildHostServices(installation).ListAccounts(context.Background(), &pluginv1.ListAccountsRequest{})
	require.Equal(t, codes.Unavailable, status.Code(err))
}
