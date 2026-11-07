//go:build unit

package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	dbuser "github.com/Wei-Shaw/sub2api/ent/user"
	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/accessban"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestLoginOrRegisterVerifiedEmailOAuth_BlacklistedExistingUserCanStillSignIn(t *testing.T) {
	svc, repo, _ := newAuthServiceForEmailBindWithRefreshCache(t, map[string]string{
		service.SettingKeyRegistrationEnabled:        "true",
		service.SettingKeyRegistrationEmailBlacklist: `["kept.user@gmail.com"]`,
	}, nil, nil, newEmailBindRefreshTokenCacheStub())
	ctx := context.Background()

	existing := &service.User{
		Email:        "kept.user@gmail.com",
		PasswordHash: "hash",
		Role:         service.RoleUser,
		Status:       service.StatusActive,
		Concurrency:  1,
	}
	require.NoError(t, repo.Create(ctx, existing))

	tokens, user, err := svc.LoginOrRegisterVerifiedEmailOAuth(ctx, service.EmailOAuthIdentityInput{
		ProviderType:    "github",
		ProviderSubject: "gh-kept-user",
		Email:           "kept.user@gmail.com",
		EmailVerified:   true,
		Username:        "kept",
	})
	require.NoError(t, err)
	require.NotNil(t, tokens)
	require.Equal(t, existing.ID, user.ID)
}

func TestLoginOrRegisterVerifiedEmailOAuth_BlacklistedNewEmailRejected(t *testing.T) {
	svc, _, client := newAuthServiceForEmailBindWithRefreshCache(t, map[string]string{
		service.SettingKeyRegistrationEnabled:        "true",
		service.SettingKeyRegistrationEmailBlacklist: `["blocked.user@gmail.com"]`,
	}, nil, nil, newEmailBindRefreshTokenCacheStub())
	ctx := context.Background()

	_, _, err := svc.LoginOrRegisterVerifiedEmailOAuth(ctx, service.EmailOAuthIdentityInput{
		ProviderType:    "google",
		ProviderSubject: "google-blocked-user",
		Email:           "blockeduser+x@googlemail.com",
		EmailVerified:   true,
	})
	require.ErrorIs(t, err, service.ErrEmailBlocked)

	count, err := client.User.Query().Where(dbuser.EmailEQ("blockeduser+x@googlemail.com")).Count(ctx)
	require.NoError(t, err)
	require.Zero(t, count)
}

// The exact registration blacklist intentionally allows existing OAuth users,
// but Fork email format and Access Ban rules still guard every OAuth login.
type emailOAuthAccessBanRepo struct {
	service.IPBanRepository
	rules []service.IPBan
	err   error
}

func (r *emailOAuthAccessBanRepo) ListActive(context.Context, time.Time) ([]service.IPBan, error) {
	return r.rules, r.err
}
func (r *emailOAuthAccessBanRepo) RecordHit(context.Context, int64, time.Time) error { return nil }

func TestLoginOrRegisterVerifiedEmailOAuth_PreservesForkEmailGuards(t *testing.T) {
	for _, tc := range []struct {
		name, email      string
		rule             service.IPBan
		repoErr, wantErr error
		existing         bool
	}{
		{name: "local part over 64 characters", email: strings.Repeat("a", 65) + "@example.com", wantErr: service.ErrInvalidEmailFormat},
		{name: "single label domain", email: "user@localhost", wantErr: service.ErrInvalidEmailFormat},
		{name: "blocked new domain", email: "user@blocked.example.com", rule: service.IPBan{RuleType: accessban.RuleTypeEmailSuffix, Pattern: "@blocked.example.com"}, wantErr: service.ErrEmailBanned},
		{name: "blocked existing domain", email: "user@blocked.example.com", rule: service.IPBan{RuleType: accessban.RuleTypeEmailSuffix, Pattern: "@blocked.example.com"}, existing: true, wantErr: service.ErrEmailBanned},
		{name: "blocked email regex", email: "user+tag@example.com", rule: service.IPBan{RuleType: accessban.RuleTypeEmailRegex, Pattern: `\+[^@]+@example\.com$`}, wantErr: service.ErrEmailBanned},
		{name: "ban storage unavailable", email: "user@example.com", repoErr: errors.New("unavailable"), wantErr: service.ErrServiceUnavailable},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, repo, client := newAuthServiceForEmailBindWithRefreshCache(t, nil, nil, nil, nil)
			cfg := &config.Config{JWT: config.JWTConfig{Secret: "test-oauth-email-policy", ExpireHour: 1}}
			settings := service.NewSettingService(&emailBindSettingRepoStub{values: map[string]string{service.SettingKeyRegistrationEnabled: "true"}}, cfg)
			tc.rule.ID = 1
			tc.rule.Status = service.IPBanStatusActive
			banRepo := &emailOAuthAccessBanRepo{rules: []service.IPBan{tc.rule}, err: tc.repoErr}
			svc := service.NewAuthService(client, repo, nil, newEmailBindRefreshTokenCacheStub(), cfg, settings, nil, nil, nil, nil, nil, nil, nil, service.NewIPBanService(banRepo))
			ctx := context.Background()
			if tc.existing {
				require.NoError(t, repo.Create(ctx, &service.User{Email: tc.email, PasswordHash: "hash", Role: service.RoleUser, Status: service.StatusActive, Concurrency: 1}))
			}
			before, err := client.User.Query().Count(ctx)
			require.NoError(t, err)
			identitiesBefore, err := client.AuthIdentity.Query().Count(ctx)
			require.NoError(t, err)
			tokens, user, err := svc.LoginOrRegisterVerifiedEmailOAuth(ctx, service.EmailOAuthIdentityInput{
				ProviderType: "github", ProviderSubject: "fork-guard-user", Email: tc.email, EmailVerified: true,
			})
			require.ErrorIs(t, err, tc.wantErr)
			require.Nil(t, tokens)
			require.Nil(t, user)
			after, err := client.User.Query().Count(ctx)
			require.NoError(t, err)
			require.Equal(t, before, after, "rejected OAuth must not create an account")
			identities, err := client.AuthIdentity.Query().Count(ctx)
			require.NoError(t, err)
			require.Equal(t, identitiesBefore, identities, "rejected OAuth must not bind an identity")
		})
	}
}
