package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type loginBanUserRepo struct {
	UserRepository
	user *User
	err  error
}

func (r loginBanUserRepo) GetByEmail(context.Context, string) (*User, error) { return r.user, r.err }

type loginBanRepo struct {
	memoryIPBanRepo
	created *IPBan
	err     error
}

func (r *loginBanRepo) UpsertAutomatic(ctx context.Context, ban *IPBan) error {
	if r.err != nil {
		return r.err
	}
	r.created = ban
	r.rules = []IPBan{*ban}
	return nil
}

func TestLoginWithClientIP_AutomaticBan(t *testing.T) {
	hash, err := (&AuthService{}).HashPassword("correct-password")
	require.NoError(t, err)
	for _, tc := range []struct {
		name, email, ip, password string
		user                      *User
		repoErr                   error
		wantBan                   bool
	}{
		{"missing admin", "admin@sub2api.local", "43.255.119.7", "wrong", nil, ErrUserNotFound, true},
		{"mixed case admin", "OpsADMIN@example.com", "43.255.119.7", "wrong", nil, ErrUserNotFound, true},
		{"wrong admin password", "admin@example.com", "43.255.119.7", "wrong", &User{PasswordHash: hash, Status: StatusActive}, nil, true},
		{"ordinary user", "user@example.com", "43.255.119.7", "wrong", nil, ErrUserNotFound, false},
		{"admin in domain only", "user@admin.example.com", "43.255.119.7", "wrong", nil, ErrUserNotFound, false},
		{"database failure", "admin@example.com", "43.255.119.7", "wrong", nil, errors.New("database unavailable"), false},
		{"successful login", "admin@example.com", "43.255.119.7", "correct-password", &User{PasswordHash: hash, Status: StatusActive}, nil, false},
		{"inactive user", "admin@example.com", "43.255.119.7", "correct-password", &User{PasswordHash: hash, Status: "disabled"}, nil, false},
		{"private proxy", "admin@example.com", "172.18.0.1", "wrong", nil, ErrUserNotFound, false},
		{"loopback", "admin@example.com", "127.0.0.1", "wrong", nil, ErrUserNotFound, false},
		{"invalid address", "admin@example.com", "unknown", "wrong", nil, ErrUserNotFound, false},
		{"IPv6", "admin@example.com", "2606:4700:4700::1111", "wrong", nil, ErrUserNotFound, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &loginBanRepo{}
			bans := NewIPBanService(repo)
			svc := &AuthService{userRepo: loginBanUserRepo{user: tc.user, err: tc.repoErr}, accessBanService: bans, cfg: &config.Config{}}
			_, _, loginErr := svc.LoginWithClientIP(context.Background(), tc.email, tc.password, tc.ip)
			if tc.name == "successful login" {
				require.NoError(t, loginErr)
			}
			if !tc.wantBan {
				require.Nil(t, repo.created)
				return
			}
			require.ErrorIs(t, loginErr, ErrInvalidCredentials)
			require.NotNil(t, repo.created)
			require.Equal(t, IPBanSourceAdminLogin, repo.created.Source)
			require.Equal(t, tc.ip, repo.created.Pattern)
			require.WithinDuration(t, time.Now().Add(24*time.Hour), *repo.created.ExpiresAt, 5*time.Second)
			_, blocked, err := bans.CheckClient(context.Background(), tc.ip, "browser")
			require.NoError(t, err)
			require.True(t, blocked, "new ban must invalidate the cached empty rules")
		})
	}
}

func TestLoginWithClientIP_BanFailurePreservesLoginError(t *testing.T) {
	svc := &AuthService{userRepo: loginBanUserRepo{err: ErrUserNotFound}, accessBanService: NewIPBanService(&loginBanRepo{err: errors.New("write failed")})}
	_, _, err := svc.LoginWithClientIP(context.Background(), "admin@example.com", "wrong", "43.255.119.7")
	require.ErrorIs(t, err, ErrInvalidCredentials)
}
