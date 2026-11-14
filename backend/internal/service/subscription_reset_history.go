package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

// SubscriptionResetHistory contains only public reset facts, never card source
// references, idempotency keys, or administrator identity.
type SubscriptionResetHistory struct {
	ID           int64     `json:"id"`
	Type         string    `json:"type"`
	ResetAt      time.Time `json:"reset_at"`
	Daily        bool      `json:"daily"`
	Weekly       bool      `json:"weekly"`
	Monthly      bool      `json:"monthly"`
	ValidityDays int       `json:"validity_days,omitempty"`
}

type SubscriptionResetHistoryRepository interface {
	ListResetHistory(context.Context, int64, pagination.PaginationParams) ([]SubscriptionResetHistory, *pagination.PaginationResult, error)
}

// ListResetHistory requires ownership when userID is provided. Only the admin
// handler passes nil, and can inspect history after a subscription is revoked.
func (s *SubscriptionService) ListResetHistory(ctx context.Context, subscriptionID int64, userID *int64, page, pageSize int) ([]SubscriptionResetHistory, *pagination.PaginationResult, error) {
	if subscriptionID <= 0 || (userID != nil && *userID <= 0) {
		return nil, nil, ErrSubscriptionNotFound
	}
	var sub *UserSubscription
	var err error
	if userID == nil {
		sub, err = s.userSubRepo.GetByIDIncludeDeleted(ctx, subscriptionID)
	} else {
		sub, err = s.userSubRepo.GetByID(ctx, subscriptionID)
	}
	if err != nil {
		return nil, nil, err
	}
	if sub == nil || (userID != nil && (sub.UserID != *userID || sub.DeletedAt != nil)) {
		return nil, nil, ErrSubscriptionNotFound
	}
	repo, ok := s.userSubRepo.(SubscriptionResetHistoryRepository)
	if !ok {
		return nil, nil, fmt.Errorf("subscription repository does not support reset history")
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return repo.ListResetHistory(ctx, subscriptionID, pagination.PaginationParams{Page: page, PageSize: pageSize})
}
