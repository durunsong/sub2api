package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type resetCardPublishCache struct {
	BillingCache
	invalidated bool
	published   string
}

func (c *resetCardPublishCache) InvalidateSubscriptionCache(context.Context, int64, int64) error {
	c.invalidated = true
	return nil
}
func (c *resetCardPublishCache) PublishSubscriptionCacheInvalidation(_ context.Context, key string) error {
	c.published = key
	return nil
}
func (c *resetCardPublishCache) SubscribeSubscriptionCacheInvalidation(context.Context, func(string)) error {
	return nil
}

type resetCardRepoStub struct {
	*subscriptionUserSubRepoStub
	cards []UserSubscriptionResetCard
	calls []string
}

func (r *resetCardRepoStub) GrantResetCard(_ context.Context, subscriptionID int64, validityDays int, source PurchaseSource) (bool, error) {
	for i := range r.cards {
		card := &r.cards[i]
		if card.UserSubscriptionID == subscriptionID && card.SourceType == source.Type && card.SourceReference == source.Reference && card.SourceSequence == 1 {
			return false, nil
		}
	}
	r.cards = append(r.cards, UserSubscriptionResetCard{ID: int64(len(r.cards) + 1), UserSubscriptionID: subscriptionID, ValidityDays: validityDays, SourceType: source.Type, SourceReference: source.Reference, SourceSequence: 1})
	r.byID[subscriptionID].ManualResetCredits++
	return true, nil
}

func (r *resetCardRepoStub) LockResetCardSubscription(_ context.Context, subscriptionID, userID int64) error {
	r.calls = append(r.calls, "lock")
	sub := r.byID[subscriptionID]
	if sub == nil || sub.UserID != userID || sub.DeletedAt != nil ||
		(sub.Status != SubscriptionStatusActive && sub.Status != SubscriptionStatusExpired) {
		return ErrResetCardNotFound
	}
	return nil
}

func (r *resetCardRepoStub) ConsumeResetCard(_ context.Context, request ConsumeResetCardRequest) (ConsumeResetCardResult, error) {
	r.calls = append(r.calls, "consume")
	for i := range r.cards {
		card := &r.cards[i]
		if card.UserSubscriptionID != request.SubscriptionID || card.ValidityDays != request.ValidityDays || card.ConsumedAt != nil {
			continue
		}
		consumedAt := request.Now
		card.ConsumedAt = &consumedAt
		sub := r.byID[request.SubscriptionID]
		actual := request.Now
		if card.CreatedAt.After(actual) {
			actual = card.CreatedAt
		}
		expiresAt := actual.AddDate(0, 0, request.ValidityDays)
		if expiresAt.After(request.MaxExpiresAt) {
			expiresAt = request.MaxExpiresAt
		}
		sub.StartsAt, sub.ExpiresAt, sub.Status = actual, expiresAt, SubscriptionStatusActive
		sub.DailyWindowStart, sub.WeeklyWindowStart, sub.MonthlyWindowStart = &actual, &actual, &actual
		sub.DailyUsageUSD, sub.WeeklyUsageUSD, sub.MonthlyUsageUSD = 0, 0, 0
		sub.DailyUsageTokens, sub.WeeklyUsageTokens, sub.MonthlyUsageTokens = 0, 0, 0
		sub.ManualResetCredits--
		return ConsumeResetCardResult{CardID: card.ID, ValidityDays: card.ValidityDays, ConsumedAt: actual, MutationApplied: true}, nil
	}
	return ConsumeResetCardResult{}, ErrResetCardNotFound
}

func TestConsumeResetCardLocksSubscriptionBeforeConsuming(t *testing.T) {
	now := time.Date(2026, 8, 16, 2, 0, 0, 0, time.UTC)
	repo := &resetCardRepoStub{
		subscriptionUserSubRepoStub: newSubscriptionUserSubRepoStub(),
		cards:                       []UserSubscriptionResetCard{{ID: 1, UserSubscriptionID: 1, ValidityDays: 7}},
	}
	repo.seed(&UserSubscription{ID: 1, UserID: 2, GroupID: 3, Status: SubscriptionStatusActive, ManualResetCredits: 1})
	svc := NewSubscriptionService(groupRepoNoop{}, repo, nil, nil, nil)
	svc.now = func() time.Time { return now }

	_, err := svc.ConsumeResetCard(context.Background(), ConsumeResetCardInput{
		UserID: 2, SubscriptionID: 1, ValidityDays: 7, IdempotencyKey: "test-lock-before-consume",
	})

	require.NoError(t, err)
	require.Equal(t, []string{"lock", "consume"}, repo.calls)
}

func TestAdminActiveAssignmentGrantsCardWithoutChangingTerm(t *testing.T) {
	now := time.Date(2026, 8, 16, 2, 0, 0, 0, time.UTC)
	start := now.Add(-3 * 24 * time.Hour)
	expires := now.Add(23 * 24 * time.Hour)
	repo := &resetCardRepoStub{subscriptionUserSubRepoStub: newSubscriptionUserSubRepoStub()}
	repo.seed(&UserSubscription{ID: 1, UserID: 2, GroupID: 3, StartsAt: start, ExpiresAt: expires, Status: SubscriptionStatusActive, MonthlyUsageUSD: 1.16})
	svc := NewSubscriptionService(&subscriptionGroupRepoStub{group: &Group{ID: 3, SubscriptionType: SubscriptionTypeSubscription}}, repo, nil, nil, nil)
	svc.now = func() time.Time { return now }

	sub, err := svc.AssignSubscription(context.Background(), &AssignSubscriptionInput{UserID: 2, GroupID: 3, ValidityDays: 30, AssignedBy: 9, Notes: "admin"})
	require.NoError(t, err)
	require.Equal(t, start, sub.StartsAt)
	require.Equal(t, expires, sub.ExpiresAt)
	require.Equal(t, float64(1.16), sub.MonthlyUsageUSD)
	require.Equal(t, 1, sub.ManualResetCredits)
	require.Len(t, repo.cards, 1)
	require.Equal(t, 30, repo.cards[0].ValidityDays)
	require.Equal(t, PurchaseSourceAssignment, repo.cards[0].SourceType)
}

func TestPaidActiveRepurchaseGrantsCardWithoutChangingTermOrUsage(t *testing.T) {
	now := time.Date(2026, 8, 16, 2, 0, 0, 0, time.UTC)
	start := now.Add(-48 * time.Hour)
	expires := now.Add(20 * 24 * time.Hour)
	repo := &resetCardRepoStub{subscriptionUserSubRepoStub: newSubscriptionUserSubRepoStub()}
	repo.seed(&UserSubscription{ID: 1, UserID: 2, GroupID: 3, StartsAt: start, ExpiresAt: expires, Status: SubscriptionStatusActive, DailyUsageUSD: 4, WeeklyUsageUSD: 5, MonthlyUsageUSD: 6})
	svc := NewSubscriptionService(&subscriptionGroupRepoStub{group: &Group{ID: 3, SubscriptionType: SubscriptionTypeSubscription}}, repo, nil, nil, nil)
	svc.now = func() time.Time { return now }
	sub, reused, err := svc.AssignOrExtendSubscription(context.Background(), &AssignSubscriptionInput{UserID: 2, GroupID: 3, ValidityDays: 17, PurchaseSource: &PurchaseSource{Type: PurchaseSourcePayment, Reference: "9001"}})
	require.NoError(t, err)
	require.True(t, reused)
	require.Equal(t, start, sub.StartsAt)
	require.Equal(t, expires, sub.ExpiresAt)
	require.Equal(t, float64(4), sub.DailyUsageUSD)
	require.Equal(t, float64(5), sub.WeeklyUsageUSD)
	require.Equal(t, float64(6), sub.MonthlyUsageUSD)
	require.Equal(t, 1, sub.ManualResetCredits)
	require.Len(t, repo.cards, 1)
	require.Equal(t, 17, repo.cards[0].ValidityDays)
}

func TestPurchaseSourceIsIdempotent(t *testing.T) {
	now := time.Date(2026, 8, 16, 2, 0, 0, 0, time.UTC)
	repo := &resetCardRepoStub{subscriptionUserSubRepoStub: newSubscriptionUserSubRepoStub()}
	repo.seed(&UserSubscription{ID: 1, UserID: 2, GroupID: 3, StartsAt: now.Add(-time.Hour), ExpiresAt: now.Add(24 * time.Hour), Status: SubscriptionStatusActive})
	svc := NewSubscriptionService(&subscriptionGroupRepoStub{group: &Group{ID: 3, SubscriptionType: SubscriptionTypeSubscription}}, repo, nil, nil, nil)
	svc.now = func() time.Time { return now }
	input := &AssignSubscriptionInput{UserID: 2, GroupID: 3, ValidityDays: 30, PurchaseSource: &PurchaseSource{Type: PurchaseSourceRedeem, Reference: "77"}}
	_, _, err := svc.AssignOrExtendSubscription(context.Background(), input)
	require.NoError(t, err)
	_, _, err = svc.AssignOrExtendSubscription(context.Background(), input)
	require.NoError(t, err)
	require.Len(t, repo.cards, 1)
	require.Equal(t, 1, repo.byID[1].ManualResetCredits)
}

func TestConsumeResetCardRestartsExpiredSubscriptionAndClearsEveryWindow(t *testing.T) {
	now := time.Date(2026, 8, 16, 2, 0, 0, 0, time.UTC)
	oldWindow := now.Add(-10 * 24 * time.Hour)
	repo := &resetCardRepoStub{subscriptionUserSubRepoStub: newSubscriptionUserSubRepoStub(), cards: []UserSubscriptionResetCard{{ID: 9, UserSubscriptionID: 1, ValidityDays: 17, SourceType: PurchaseSourcePayment, SourceReference: "9001", SourceSequence: 1}}}
	repo.seed(&UserSubscription{ID: 1, UserID: 2, GroupID: 3, StartsAt: now.Add(-40 * 24 * time.Hour), ExpiresAt: now.Add(-10 * 24 * time.Hour), Status: SubscriptionStatusExpired, DailyWindowStart: &oldWindow, WeeklyWindowStart: &oldWindow, MonthlyWindowStart: &oldWindow, DailyUsageUSD: 1, WeeklyUsageUSD: 2, MonthlyUsageUSD: 3, DailyUsageTokens: 4, WeeklyUsageTokens: 5, MonthlyUsageTokens: 6, ManualResetCredits: 1})
	svc := NewSubscriptionService(&subscriptionGroupRepoStub{group: &Group{ID: 3, SubscriptionType: SubscriptionTypeSubscription}}, repo, nil, nil, nil)
	svc.now = func() time.Time { return now }
	sub, err := svc.ConsumeResetCard(context.Background(), ConsumeResetCardInput{IdempotencyKey: "test-reset-key", UserID: 2, SubscriptionID: 1, ValidityDays: 17})
	require.NoError(t, err)
	require.Equal(t, now, sub.StartsAt)
	require.Equal(t, now.AddDate(0, 0, 17), sub.ExpiresAt)
	require.Equal(t, SubscriptionStatusActive, sub.Status)
	require.Equal(t, now, *sub.DailyWindowStart)
	require.Equal(t, now, *sub.WeeklyWindowStart)
	require.Equal(t, now, *sub.MonthlyWindowStart)
	require.Zero(t, sub.DailyUsageUSD)
	require.Zero(t, sub.WeeklyUsageUSD)
	require.Zero(t, sub.MonthlyUsageUSD)
	require.Zero(t, sub.DailyUsageTokens)
	require.Zero(t, sub.WeeklyUsageTokens)
	require.Zero(t, sub.MonthlyUsageTokens)
	require.Zero(t, sub.ManualResetCredits)
}

func (r *resetCardRepoStub) ListAvailableResetCardGroups(_ context.Context, subscriptionIDs []int64) ([]ResetCardGroup, error) {
	wanted := make(map[int64]struct{}, len(subscriptionIDs))
	for _, id := range subscriptionIDs {
		wanted[id] = struct{}{}
	}
	counts := make(map[[2]int64]int)
	for _, card := range r.cards {
		if card.ConsumedAt != nil {
			continue
		}
		if _, ok := wanted[card.UserSubscriptionID]; !ok {
			continue
		}
		counts[[2]int64{card.UserSubscriptionID, int64(card.ValidityDays)}]++
	}
	groups := make([]ResetCardGroup, 0, len(counts))
	for key, count := range counts {
		groups = append(groups, ResetCardGroup{SubscriptionID: key[0], ValidityDays: int(key[1]), AvailableCount: count})
	}
	return groups, nil
}

func TestExpiredPurchaseReopensTermWithoutGrantingCard(t *testing.T) {
	now := time.Date(2026, 8, 16, 2, 0, 0, 0, time.UTC)
	repo := &resetCardRepoStub{subscriptionUserSubRepoStub: newSubscriptionUserSubRepoStub()}
	repo.seed(&UserSubscription{ID: 1, UserID: 2, GroupID: 3, StartsAt: now.AddDate(0, 0, -40), ExpiresAt: now.Add(-time.Hour), Status: SubscriptionStatusExpired, DailyUsageUSD: 9})
	svc := NewSubscriptionService(&subscriptionGroupRepoStub{group: &Group{ID: 3, SubscriptionType: SubscriptionTypeSubscription}}, repo, nil, nil, nil)
	svc.now = func() time.Time { return now }

	sub, _, err := svc.AssignOrExtendSubscription(context.Background(), &AssignSubscriptionInput{UserID: 2, GroupID: 3, ValidityDays: 30, PurchaseSource: &PurchaseSource{Type: PurchaseSourcePayment, Reference: "9002"}})

	require.NoError(t, err)
	require.Equal(t, now, sub.StartsAt)
	require.Equal(t, now.AddDate(0, 0, 30), sub.ExpiresAt)
	require.Zero(t, sub.DailyUsageUSD)
	require.Empty(t, repo.cards)
	require.Zero(t, sub.ManualResetCredits)
}

func TestConsumeResetCardSelectsRequestedValidity(t *testing.T) {
	now := time.Date(2026, 8, 16, 2, 0, 0, 0, time.UTC)
	repo := &resetCardRepoStub{subscriptionUserSubRepoStub: newSubscriptionUserSubRepoStub(), cards: []UserSubscriptionResetCard{
		{ID: 1, UserSubscriptionID: 7, ValidityDays: 7},
		{ID: 2, UserSubscriptionID: 7, ValidityDays: 30},
	}}
	repo.seed(&UserSubscription{ID: 7, UserID: 2, GroupID: 3, StartsAt: now.Add(-time.Hour), ExpiresAt: now.Add(time.Hour), Status: SubscriptionStatusActive, ManualResetCredits: 2})
	svc := NewSubscriptionService(groupRepoNoop{}, repo, nil, nil, nil)
	svc.now = func() time.Time { return now }

	sub, err := svc.ConsumeResetCard(context.Background(), ConsumeResetCardInput{IdempotencyKey: "test-reset-key", UserID: 2, SubscriptionID: 7, ValidityDays: 30})

	require.NoError(t, err)
	require.Nil(t, repo.cards[0].ConsumedAt)
	require.NotNil(t, repo.cards[1].ConsumedAt)
	require.Equal(t, now.AddDate(0, 0, 30), sub.ExpiresAt)
	require.Equal(t, 1, sub.ManualResetCredits)
}

func TestConsumeResetCardHidesForeignMissingAndForbiddenStates(t *testing.T) {
	now := time.Date(2026, 8, 16, 2, 0, 0, 0, time.UTC)
	deletedAt := now
	for _, tc := range []struct {
		name string
		sub  *UserSubscription
	}{
		{name: "missing"},
		{name: "foreign", sub: &UserSubscription{ID: 8, UserID: 99, Status: SubscriptionStatusActive}},
		{name: "suspended", sub: &UserSubscription{ID: 8, UserID: 2, Status: SubscriptionStatusSuspended}},
		{name: "revoked", sub: &UserSubscription{ID: 8, UserID: 2, Status: SubscriptionStatusActive, DeletedAt: &deletedAt}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &resetCardRepoStub{subscriptionUserSubRepoStub: newSubscriptionUserSubRepoStub(), cards: []UserSubscriptionResetCard{{ID: 1, UserSubscriptionID: 8, ValidityDays: 30}}}
			if tc.sub != nil {
				repo.seed(tc.sub)
			}
			svc := NewSubscriptionService(groupRepoNoop{}, repo, nil, nil, nil)
			svc.now = func() time.Time { return now }

			_, err := svc.ConsumeResetCard(context.Background(), ConsumeResetCardInput{IdempotencyKey: "test-reset-key", UserID: 2, SubscriptionID: 8, ValidityDays: 30})

			require.ErrorIs(t, err, ErrResetCardNotFound)
			require.Nil(t, repo.cards[0].ConsumedAt)
		})
	}
}

func TestListAvailableResetCardGroupsAggregatesBatchAndSingleSubscription(t *testing.T) {
	repo := &resetCardRepoStub{subscriptionUserSubRepoStub: newSubscriptionUserSubRepoStub(), cards: []UserSubscriptionResetCard{
		{ID: 1, UserSubscriptionID: 1, ValidityDays: 7},
		{ID: 2, UserSubscriptionID: 1, ValidityDays: 7},
		{ID: 3, UserSubscriptionID: 1, ValidityDays: 30},
		{ID: 4, UserSubscriptionID: 2, ValidityDays: 7},
	}}
	svc := NewSubscriptionService(groupRepoNoop{}, repo, nil, nil, nil)

	batch, err := svc.ListAvailableResetCardGroups(context.Background(), []int64{1, 2})
	require.NoError(t, err)
	require.Len(t, batch, 3)
	single, err := svc.ListAvailableResetCardGroups(context.Background(), []int64{1})
	require.NoError(t, err)
	require.Len(t, single, 2)
}

type resetCardReloadFailureRepo struct {
	*resetCardRepoStub
	getCalls int
}

func (r *resetCardReloadFailureRepo) GetByID(ctx context.Context, id int64) (*UserSubscription, error) {
	r.getCalls++
	if r.getCalls > 1 {
		return nil, errors.New("reload failed")
	}
	return r.resetCardRepoStub.GetByID(ctx, id)
}

func TestConsumeResetCardReportsReloadFailureAfterAppliedMutation(t *testing.T) {
	now := time.Date(2026, 8, 16, 2, 0, 0, 0, time.UTC)
	base := &resetCardRepoStub{subscriptionUserSubRepoStub: newSubscriptionUserSubRepoStub(), cards: []UserSubscriptionResetCard{{ID: 1, UserSubscriptionID: 1, ValidityDays: 7}}}
	base.seed(&UserSubscription{ID: 1, UserID: 2, GroupID: 3, StartsAt: now.Add(-time.Hour), ExpiresAt: now.Add(time.Hour), Status: SubscriptionStatusActive, ManualResetCredits: 1})
	repo := &resetCardReloadFailureRepo{resetCardRepoStub: base}
	svc := NewSubscriptionService(groupRepoNoop{}, repo, nil, nil, nil)
	svc.now = func() time.Time { return now }

	_, err := svc.ConsumeResetCard(context.Background(), ConsumeResetCardInput{IdempotencyKey: "test-reset-key", UserID: 2, SubscriptionID: 1, ValidityDays: 7})

	require.ErrorIs(t, err, ErrResetCardResponseReloadFailed)
	require.NotNil(t, base.cards[0].ConsumedAt, "single-statement mutation already committed before reload")
}

func TestConsumeResetCardPublishesCrossInstanceInvalidation(t *testing.T) {
	now := time.Date(2026, 8, 16, 2, 0, 0, 0, time.UTC)
	repo := &resetCardRepoStub{subscriptionUserSubRepoStub: newSubscriptionUserSubRepoStub(), cards: []UserSubscriptionResetCard{{ID: 1, UserSubscriptionID: 1, ValidityDays: 7}}}
	repo.seed(&UserSubscription{ID: 1, UserID: 2, GroupID: 3, StartsAt: now.Add(-time.Hour), ExpiresAt: now.Add(time.Hour), Status: SubscriptionStatusActive, ManualResetCredits: 1})
	cache := &resetCardPublishCache{}
	billing := NewBillingCacheService(cache, nil, nil, nil, nil, nil, &config.Config{}, nil)
	t.Cleanup(billing.Stop)
	svc := NewSubscriptionService(groupRepoNoop{}, repo, billing, nil, nil)
	svc.now = func() time.Time { return now }
	_, err := svc.ConsumeResetCard(context.Background(), ConsumeResetCardInput{UserID: 2, SubscriptionID: 1, ValidityDays: 7, IdempotencyKey: "publish-key"})
	require.NoError(t, err)
	require.True(t, cache.invalidated)
	require.Equal(t, "sub:2:3", cache.published)
}

type resetCardMismatchedReplayRepo struct {
	*resetCardRepoStub
}

func (r *resetCardMismatchedReplayRepo) ConsumeResetCard(context.Context, ConsumeResetCardRequest) (ConsumeResetCardResult, error) {
	return ConsumeResetCardResult{CardID: 1, ValidityDays: 7, ConsumedAt: time.Now(), MutationApplied: false}, nil
}

func TestConsumeResetCardRejectsIdempotencyReplayWithDifferentValidity(t *testing.T) {
	now := time.Date(2026, 8, 16, 2, 0, 0, 0, time.UTC)
	base := &resetCardRepoStub{subscriptionUserSubRepoStub: newSubscriptionUserSubRepoStub()}
	base.seed(&UserSubscription{ID: 1, UserID: 2, GroupID: 3, StartsAt: now.Add(-time.Hour), ExpiresAt: now.Add(time.Hour), Status: SubscriptionStatusActive, ManualResetCredits: 1})
	repo := &resetCardMismatchedReplayRepo{resetCardRepoStub: base}
	svc := NewSubscriptionService(groupRepoNoop{}, repo, nil, nil, nil)
	svc.now = func() time.Time { return now }

	_, err := svc.ConsumeResetCard(context.Background(), ConsumeResetCardInput{UserID: 2, SubscriptionID: 1, ValidityDays: 30, IdempotencyKey: "replayed-key"})

	require.ErrorIs(t, err, ErrResetCardIdempotencyConflict)
	require.Equal(t, 1, base.byID[1].ManualResetCredits, "conflicting replay must not consume another card")
}
