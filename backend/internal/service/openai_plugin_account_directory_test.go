package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolvePluginOutboundIdentityRejectsDisabledAccount(t *testing.T) {
	repo := schedulerTestOpenAIAccountRepo{accounts: []Account{{
		ID: 7, Platform: PlatformOpenAI, Type: AccountTypeOAuth,
		Status: StatusActive, Schedulable: false,
		Credentials: map[string]any{"access_token": "test-plugin-token"},
	}}}
	gateway := &OpenAIGatewayService{accountRepo: repo}
	ctx := context.Background()

	// A scheduling pause does not revoke an otherwise active account's identity.
	identity, err := gateway.ResolvePluginOutboundIdentity(ctx, 7)
	require.NoError(t, err)
	require.NotNil(t, identity)
	require.Equal(t, "test-plugin-token", identity.Token)

	for _, accountStatus := range []string{StatusDisabled, StatusError} {
		t.Run(accountStatus, func(t *testing.T) {
			// The plugin retains the ID; the repository now returns the new status.
			repo.accounts[0].Status = accountStatus
			identity, err := gateway.ResolvePluginOutboundIdentity(ctx, 7)
			require.NoError(t, err)
			require.Nil(t, identity, "inactive accounts must not expose credentials to plugins")
		})
	}
}
