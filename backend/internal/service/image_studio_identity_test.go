package service

import (
	"context"
	"errors"
	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
)

type studioIdentityRepo struct {
	APIKeyRepository
	mu      sync.Mutex
	key     *APIKey
	creates int
}

func (r *studioIdentityRepo) GetByKey(_ context.Context, _ string) (*APIKey, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.key == nil {
		return nil, ErrAPIKeyNotFound
	}
	copy := *r.key
	return &copy, nil
}
func (r *studioIdentityRepo) GetByID(ctx context.Context, _ int64) (*APIKey, error) {
	return r.GetByKey(ctx, "")
}
func (r *studioIdentityRepo) GetKeyAndOwnerID(ctx context.Context, _ int64) (string, int64, error) {
	k, e := r.GetByKey(ctx, "")
	if e != nil {
		return "", 0, e
	}
	return k.Key, k.UserID, nil
}
func (r *studioIdentityRepo) Create(_ context.Context, k *APIKey) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.key != nil {
		return ErrAPIKeyExists
	}
	k.ID = 41
	copy := *k
	r.key = &copy
	r.creates++
	return nil
}

type studioIdentityUsers struct {
	UserRepository
	user User
}

func (r *studioIdentityUsers) GetByID(context.Context, int64) (*User, error) {
	u := r.user
	return &u, nil
}

type studioIdentityGroups struct {
	GroupRepository
	group Group
}

func (r *studioIdentityGroups) GetByID(context.Context, int64) (*Group, error) {
	g := r.group
	return &g, nil
}

type studioIdentitySubscriptions struct{ UserSubscriptionRepository }

func (*studioIdentitySubscriptions) GetActiveByUserIDAndGroupID(context.Context, int64, int64) (*UserSubscription, error) {
	return nil, errors.New("expired")
}
func TestImageStudioIdentityConcurrentCreationAndManagementIsolation(t *testing.T) {
	repo := &studioIdentityRepo{}
	cfg := &config.Config{}
	cfg.JWT.Secret = "test-only-secret"
	svc := NewAPIKeyService(repo, &studioIdentityUsers{user: User{ID: 7, Status: StatusActive}}, &studioIdentityGroups{group: Group{ID: 3, Status: StatusActive}}, &studioIdentitySubscriptions{}, nil, nil, cfg)
	var wg sync.WaitGroup
	results := make(chan *APIKey, 12)
	errs := make(chan error, 12)
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			k, e := svc.EnsureImageStudioKey(context.Background(), 7, 3)
			results <- k
			errs <- e
		}()
	}
	wg.Wait()
	close(results)
	close(errs)
	for e := range errs {
		require.NoError(t, e)
	}
	for k := range results {
		require.Equal(t, int64(41), k.ID)
		require.True(t, IsImageStudioKey(k.Key))
		require.Equal(t, float64(0), k.Quota)
	}
	require.Equal(t, 1, repo.creates)
	_, err := svc.GetByID(context.Background(), 41)
	require.ErrorIs(t, err, ErrAPIKeyNotFound)
	_, err = svc.Update(context.Background(), 41, 7, UpdateAPIKeyRequest{})
	require.ErrorIs(t, err, ErrAPIKeyNotFound)
	require.ErrorIs(t, svc.Delete(context.Background(), 41, 7), ErrAPIKeyNotFound)
}
func TestImageStudioIdentityRequiresCurrentGroupPermission(t *testing.T) {
	cfg := &config.Config{}
	cfg.JWT.Secret = "test-only-secret"
	for _, group := range []Group{{ID: 3, Status: StatusActive, IsExclusive: true}, {ID: 3, Status: StatusActive, SubscriptionType: SubscriptionTypeSubscription}, {ID: 3, Status: StatusDisabled}} {
		repo := &studioIdentityRepo{}
		svc := NewAPIKeyService(repo, &studioIdentityUsers{user: User{ID: 7, Status: StatusActive}}, &studioIdentityGroups{group: group}, &studioIdentitySubscriptions{}, nil, nil, cfg)
		_, err := svc.EnsureImageStudioKey(context.Background(), 7, 3)
		require.ErrorIs(t, err, ErrGroupNotAllowed)
		require.Zero(t, repo.creates)
	}
}

func TestImageStudioIdentityCannotBeReboundByAdmin(t *testing.T) {
	repo := &studioIdentityRepo{key: &APIKey{ID: 41, Key: ImageStudioKeyPrefix + "private"}}
	svc := &adminServiceImpl{apiKeyRepo: repo}
	_, err := svc.AdminUpdateAPIKeyGroupID(context.Background(), 41, nil)
	require.ErrorIs(t, err, ErrAPIKeyNotFound)
	_, err = svc.AdminResetAPIKeyRateLimitUsage(context.Background(), 41)
	require.ErrorIs(t, err, ErrAPIKeyNotFound)

}
