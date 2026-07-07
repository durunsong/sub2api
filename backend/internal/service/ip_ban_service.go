package service

import (
	"context"
	"strings"
	"sync"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	ippkg "github.com/Wei-Shaw/sub2api/internal/pkg/ip"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

const (
	IPBanStatusActive   = "active"
	IPBanStatusInactive = "inactive"
)

var (
	ErrIPBanNotFound       = infraerrors.NotFound("IP_BAN_NOT_FOUND", "IP ban rule not found")
	ErrInvalidIPBanPattern = infraerrors.BadRequest("INVALID_IP_BAN_PATTERN", "invalid IP or CIDR pattern")
	ErrIPBanAlreadyExists  = infraerrors.Conflict("IP_BAN_ALREADY_EXISTS", "IP ban rule already exists")
	ErrIPBanned            = infraerrors.Forbidden("IP_BANNED", "Your IP is banned")
)

type IPBan struct {
	ID        int64      `json:"id"`
	Pattern   string     `json:"pattern"`
	Status    string     `json:"status"`
	Reason    string     `json:"reason,omitempty"`
	Source    string     `json:"source"`
	CreatedBy *int64     `json:"created_by,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	LastHitAt *time.Time `json:"last_hit_at,omitempty"`
	HitCount  int64      `json:"hit_count"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type IPBanListFilters struct {
	Search string
	Status string
}

type CreateIPBanInput struct {
	Pattern   string
	Reason    string
	Source    string
	CreatedBy *int64
	ExpiresAt *time.Time
}

type UpdateIPBanInput struct {
	Pattern   *string
	Reason    *string
	Status    *string
	ExpiresAt **time.Time
}

type IPBanRepository interface {
	Create(ctx context.Context, ban *IPBan) error
	GetByID(ctx context.Context, id int64) (*IPBan, error)
	List(ctx context.Context, params pagination.PaginationParams, filters IPBanListFilters) ([]IPBan, *pagination.PaginationResult, error)
	Update(ctx context.Context, ban *IPBan) error
	Delete(ctx context.Context, id int64) error
	ListActive(ctx context.Context, now time.Time) ([]IPBan, error)
	RecordHit(ctx context.Context, id int64, at time.Time) error
}

type IPBanService struct {
	repo     IPBanRepository
	cacheTTL time.Duration

	cacheMu      sync.RWMutex
	cachedBans   []IPBan
	cacheExpires time.Time
}

func NewIPBanService(repo IPBanRepository) *IPBanService {
	return &IPBanService{repo: repo, cacheTTL: 5 * time.Second}
}

func (s *IPBanService) Create(ctx context.Context, input CreateIPBanInput) (*IPBan, error) {
	pattern := strings.TrimSpace(input.Pattern)
	if !ippkg.ValidateIPPattern(pattern) {
		return nil, ErrInvalidIPBanPattern
	}
	source := strings.TrimSpace(input.Source)
	if source == "" {
		source = "manual"
	}
	ban := &IPBan{Pattern: pattern, Status: IPBanStatusActive, Reason: strings.TrimSpace(input.Reason), Source: source, CreatedBy: input.CreatedBy, ExpiresAt: input.ExpiresAt}
	if err := s.repo.Create(ctx, ban); err != nil {
		return nil, err
	}
	s.invalidateCache()
	return ban, nil
}

func (s *IPBanService) GetByID(ctx context.Context, id int64) (*IPBan, error) { return s.repo.GetByID(ctx, id) }

func (s *IPBanService) List(ctx context.Context, params pagination.PaginationParams, filters IPBanListFilters) ([]IPBan, *pagination.PaginationResult, error) {
	filters.Search = strings.TrimSpace(filters.Search)
	filters.Status = strings.TrimSpace(filters.Status)
	return s.repo.List(ctx, params, filters)
}

func (s *IPBanService) Update(ctx context.Context, id int64, input UpdateIPBanInput) (*IPBan, error) {
	ban, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if input.Pattern != nil {
		pattern := strings.TrimSpace(*input.Pattern)
		if !ippkg.ValidateIPPattern(pattern) {
			return nil, ErrInvalidIPBanPattern
		}
		ban.Pattern = pattern
	}
	if input.Reason != nil {
		ban.Reason = strings.TrimSpace(*input.Reason)
	}
	if input.Status != nil {
		status := strings.TrimSpace(*input.Status)
		if status != IPBanStatusActive && status != IPBanStatusInactive {
			return nil, infraerrors.BadRequest("INVALID_IP_BAN_STATUS", "invalid IP ban status")
		}
		ban.Status = status
	}
	if input.ExpiresAt != nil {
		ban.ExpiresAt = *input.ExpiresAt
	}
	if err := s.repo.Update(ctx, ban); err != nil {
		return nil, err
	}
	s.invalidateCache()
	return ban, nil
}

func (s *IPBanService) Delete(ctx context.Context, id int64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	s.invalidateCache()
	return nil
}

func (s *IPBanService) Check(ctx context.Context, clientIP string) (*IPBan, bool, error) {
	clientIP = strings.TrimSpace(clientIP)
	if clientIP == "" || s == nil || s.repo == nil {
		return nil, false, nil
	}
	now := time.Now()
	bans, err := s.listActiveCached(ctx, now)
	if err != nil {
		return nil, false, err
	}
	for i := range bans {
		if ippkg.MatchesPattern(clientIP, bans[i].Pattern) {
			_ = s.repo.RecordHit(ctx, bans[i].ID, now)
			return &bans[i], true, nil
		}
	}
	return nil, false, nil
}

func (s *IPBanService) listActiveCached(ctx context.Context, now time.Time) ([]IPBan, error) {
	s.cacheMu.RLock()
	if now.Before(s.cacheExpires) {
		cached := cloneIPBans(s.cachedBans)
		s.cacheMu.RUnlock()
		return cached, nil
	}
	s.cacheMu.RUnlock()

	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()
	if now.Before(s.cacheExpires) {
		return cloneIPBans(s.cachedBans), nil
	}
	bans, err := s.repo.ListActive(ctx, now)
	if err != nil {
		return nil, err
	}
	s.cachedBans = cloneIPBans(bans)
	s.cacheExpires = now.Add(s.cacheTTL)
	return cloneIPBans(bans), nil
}

func (s *IPBanService) invalidateCache() {
	if s == nil {
		return
	}
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()
	s.cachedBans = nil
	s.cacheExpires = time.Time{}
}

func cloneIPBans(in []IPBan) []IPBan {
	if len(in) == 0 {
		return nil
	}
	out := make([]IPBan, len(in))
	copy(out, in)
	return out
}
