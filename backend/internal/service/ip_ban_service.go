package service

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/accessban"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	ippkg "github.com/Wei-Shaw/sub2api/internal/pkg/ip"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

const (
	IPBanStatusActive   = "active"
	IPBanStatusInactive = "inactive"
)

var (
	ErrIPBanNotFound       = infraerrors.NotFound("IP_BAN_NOT_FOUND", "access ban rule not found")
	ErrInvalidIPBanPattern = infraerrors.BadRequest("INVALID_IP_BAN_PATTERN", "invalid IP or CIDR pattern")
	ErrInvalidUABanPattern = infraerrors.BadRequest("INVALID_UA_BAN_PATTERN", "invalid User-Agent pattern")
	ErrInvalidEmailBanPattern = infraerrors.BadRequest("INVALID_EMAIL_BAN_PATTERN", "invalid email suffix pattern")
	ErrInvalidEmailRegexBanPattern = infraerrors.BadRequest("INVALID_EMAIL_REGEX_BAN_PATTERN", "invalid email regex pattern")
	ErrInvalidAccessBanRuleType = infraerrors.BadRequest("INVALID_ACCESS_BAN_RULE_TYPE", "invalid access ban rule type")
	ErrIPBanAlreadyExists  = infraerrors.Conflict("IP_BAN_ALREADY_EXISTS", "access ban rule already exists")
	ErrIPBanned            = infraerrors.Forbidden("IP_BANNED", "Your IP is banned")
	ErrEmailBanned         = infraerrors.Forbidden("EMAIL_BANNED", "This email address is not allowed")
	ErrClientAccessBanned  = infraerrors.Forbidden("CLIENT_ACCESS_BANNED", "Access denied")
)

type IPBan struct {
	ID        int64      `json:"id"`
	RuleType  string     `json:"rule_type"`
	Pattern   string     `json:"pattern"`
	UAPattern string     `json:"ua_pattern,omitempty"`
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
	Search   string
	Status   string
	RuleType string
}

type CreateIPBanInput struct {
	RuleType  string
	Pattern   string
	UAPattern string
	Reason    string
	Source    string
	CreatedBy *int64
	ExpiresAt *time.Time
}

type UpdateIPBanInput struct {
	RuleType  *string
	Pattern   *string
	UAPattern *string
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
	ruleType, pattern, uaPattern, err := normalizeAccessBanInput(input.RuleType, input.Pattern, input.UAPattern)
	if err != nil {
		return nil, err
	}
	source := strings.TrimSpace(input.Source)
	if source == "" {
		source = "manual"
	}
	ban := &IPBan{
		RuleType:  ruleType,
		Pattern:   pattern,
		UAPattern: uaPattern,
		Status:    IPBanStatusActive,
		Reason:    strings.TrimSpace(input.Reason),
		Source:    source,
		CreatedBy: input.CreatedBy,
		ExpiresAt: input.ExpiresAt,
	}
	if err := s.repo.Create(ctx, ban); err != nil {
		return nil, err
	}
	s.invalidateCache()
	return ban, nil
}

func (s *IPBanService) GetByID(ctx context.Context, id int64) (*IPBan, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *IPBanService) List(ctx context.Context, params pagination.PaginationParams, filters IPBanListFilters) ([]IPBan, *pagination.PaginationResult, error) {
	filters.Search = strings.TrimSpace(filters.Search)
	filters.Status = strings.TrimSpace(filters.Status)
	filters.RuleType = accessban.NormalizeRuleType(filters.RuleType)
	return s.repo.List(ctx, params, filters)
}

func (s *IPBanService) Update(ctx context.Context, id int64, input UpdateIPBanInput) (*IPBan, error) {
	ban, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if input.RuleType != nil {
		ruleType := accessban.NormalizeRuleType(*input.RuleType)
		if ruleType == "" {
			return nil, ErrInvalidAccessBanRuleType
		}
		ban.RuleType = ruleType
	}
	if input.Pattern != nil || input.UAPattern != nil || input.RuleType != nil {
		pattern := ban.Pattern
		uaPattern := ban.UAPattern
		if input.Pattern != nil {
			pattern = *input.Pattern
		}
		if input.UAPattern != nil {
			uaPattern = *input.UAPattern
		}
		ruleType, normalizedPattern, normalizedUA, err := normalizeAccessBanInput(ban.RuleType, pattern, uaPattern)
		if err != nil {
			return nil, err
		}
		ban.RuleType = ruleType
		ban.Pattern = normalizedPattern
		ban.UAPattern = normalizedUA
	}
	if input.Reason != nil {
		ban.Reason = strings.TrimSpace(*input.Reason)
	}
	if input.Status != nil {
		status := strings.TrimSpace(*input.Status)
		if status != IPBanStatusActive && status != IPBanStatusInactive {
			return nil, infraerrors.BadRequest("INVALID_IP_BAN_STATUS", "invalid access ban status")
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

// CheckClient matches IP / UA / IP+UA rules for gateway and auth client guards.
func (s *IPBanService) CheckClient(ctx context.Context, clientIP, userAgent string) (*IPBan, bool, error) {
	clientIP = strings.TrimSpace(clientIP)
	userAgent = strings.TrimSpace(userAgent)
	if s == nil || s.repo == nil {
		return nil, false, nil
	}
	now := time.Now()
	bans, err := s.listActiveCached(ctx, now)
	if err != nil {
		return nil, false, err
	}
	for i := range bans {
		switch accessban.NormalizeRuleType(bans[i].RuleType) {
		case accessban.RuleTypeIP, accessban.RuleTypeUA, accessban.RuleTypeIPUA:
		default:
			continue
		}
		if clientRuleNeedsIP(bans[i].RuleType) && clientIP == "" {
			continue
		}
		if !accessban.MatchesClient(bans[i].RuleType, bans[i].Pattern, bans[i].UAPattern, clientIP, userAgent) {
			continue
		}
		_ = s.repo.RecordHit(ctx, bans[i].ID, now)
		return &bans[i], true, nil
	}
	return nil, false, nil
}

// CheckEmail matches email suffix ban rules for registration/login flows.
func (s *IPBanService) CheckEmail(ctx context.Context, email string) (*IPBan, bool, error) {
	email = strings.TrimSpace(email)
	if email == "" || s == nil || s.repo == nil {
		return nil, false, nil
	}
	now := time.Now()
	bans, err := s.listActiveCached(ctx, now)
	if err != nil {
		return nil, false, err
	}
	for i := range bans {
		ruleType := accessban.NormalizeRuleType(bans[i].RuleType)
		matched := false
		switch ruleType {
		case accessban.RuleTypeEmailSuffix:
			matched = accessban.MatchesEmailSuffix(email, bans[i].Pattern)
		case accessban.RuleTypeEmailRegex:
			matched = accessban.MatchesEmailRegex(email, bans[i].Pattern)
		default:
			continue
		}
		if !matched {
			continue
		}
		_ = s.repo.RecordHit(ctx, bans[i].ID, now)
		return &bans[i], true, nil
	}
	return nil, false, nil
}

func clientRuleNeedsIP(ruleType string) bool {
	switch accessban.NormalizeRuleType(ruleType) {
	case accessban.RuleTypeIP, accessban.RuleTypeIPUA:
		return true
	default:
		return false
	}
}

func normalizeAccessBanInput(ruleType, pattern, uaPattern string) (string, string, string, error) {
	ruleType = accessban.NormalizeRuleType(ruleType)
	if ruleType == "" {
		ruleType = accessban.RuleTypeIP
	}
	pattern = strings.TrimSpace(pattern)
	uaPattern = strings.TrimSpace(uaPattern)

	switch ruleType {
	case accessban.RuleTypeIP:
		if !ippkg.ValidateIPPattern(pattern) {
			return "", "", "", ErrInvalidIPBanPattern
		}
		return ruleType, pattern, "", nil
	case accessban.RuleTypeUA:
		if !accessban.ValidateUAPattern(pattern) {
			return "", "", "", ErrInvalidUABanPattern
		}
		return ruleType, pattern, "", nil
	case accessban.RuleTypeIPUA:
		if !ippkg.ValidateIPPattern(pattern) {
			return "", "", "", ErrInvalidIPBanPattern
		}
		if !accessban.ValidateUAPattern(uaPattern) {
			return "", "", "", ErrInvalidUABanPattern
		}
		return ruleType, pattern, uaPattern, nil
	case accessban.RuleTypeEmailSuffix:
		normalized, err := normalizeRegistrationEmailSuffix(pattern)
		if err != nil || normalized == "" {
			return "", "", "", ErrInvalidEmailBanPattern
		}
		return ruleType, normalized, "", nil
	case accessban.RuleTypeEmailRegex:
		if !accessban.ValidateEmailRegexPattern(pattern) {
			return "", "", "", ErrInvalidEmailRegexBanPattern
		}
		return ruleType, pattern, "", nil
	default:
		return "", "", "", ErrInvalidAccessBanRuleType
	}
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
