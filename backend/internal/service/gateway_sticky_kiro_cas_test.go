package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/stretchr/testify/require"
)

type forkStickyTTLCache struct {
	GatewayCache
	ttl time.Duration
}

func (c *forkStickyTTLCache) SetSessionAccountID(_ context.Context, _ int64, _ string, _ int64, ttl time.Duration) error {
	c.ttl = ttl
	return nil
}

type forkStickyCompareTTLCache struct{ *forkStickyTTLCache }

func (c *forkStickyCompareTTLCache) CompareAndBindSessionAccountID(_ context.Context, _ int64, _ string, _ int64, _ []int64, ttl time.Duration) (bool, error) {
	c.ttl = ttl
	return true, nil
}

func TestStickyWritebackPreservesKiroTTL(t *testing.T) {
	for _, compare := range []bool{false, true} {
		t.Run(map[bool]string{false: "fallback", true: "compare"}[compare], func(t *testing.T) {
			cache := &forkStickyTTLCache{}
			svc := &GatewayService{cache: cache}
			if compare {
				svc.cache = &forkStickyCompareTTLCache{cache}
			}
			group := &Group{ID: 42, Platform: PlatformKiro, Status: StatusActive, Hydrated: true, KiroStickySessionTTLSeconds: 7200}
			ctx := context.WithValue(context.Background(), ctxkey.Group, group)
			require.NoError(t, svc.BindStickySessionUnlessRebound(ctx, &group.ID, "session", 1, []int64{2}))
			require.Equal(t, 2*time.Hour, cache.ttl)
		})
	}
}
