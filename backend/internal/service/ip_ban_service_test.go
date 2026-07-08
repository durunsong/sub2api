package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/accessban"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type memoryIPBanRepo struct {
	rules []IPBan
}

func (m *memoryIPBanRepo) Create(context.Context, *IPBan) error { return nil }
func (m *memoryIPBanRepo) GetByID(context.Context, int64) (*IPBan, error) {
	return nil, ErrIPBanNotFound
}
func (m *memoryIPBanRepo) List(context.Context, pagination.PaginationParams, IPBanListFilters) ([]IPBan, *pagination.PaginationResult, error) {
	return nil, nil, nil
}
func (m *memoryIPBanRepo) Update(context.Context, *IPBan) error { return nil }
func (m *memoryIPBanRepo) Delete(context.Context, int64) error  { return nil }
func (m *memoryIPBanRepo) ListActive(context.Context, time.Time) ([]IPBan, error) {
	return append([]IPBan(nil), m.rules...), nil
}
func (m *memoryIPBanRepo) RecordHit(context.Context, int64, time.Time) error { return nil }

func TestIPBanService_CheckEmail(t *testing.T) {
	repo := &memoryIPBanRepo{
		rules: []IPBan{{
			ID:       1,
			RuleType: accessban.RuleTypeEmailSuffix,
			Pattern:  "@365.liout.com",
			Status:   IPBanStatusActive,
		}},
	}
	svc := NewIPBanService(repo)
	_, banned, err := svc.CheckEmail(context.Background(), "user@365.liout.com")
	require.NoError(t, err)
	require.True(t, banned)

	_, banned, err = svc.CheckEmail(context.Background(), "user@mail.365.liout.com")
	require.NoError(t, err)
	require.False(t, banned)
}

func TestIPBanService_CheckEmail_Regex(t *testing.T) {
	repo := &memoryIPBanRepo{
		rules: []IPBan{{
			ID:       3,
			RuleType: accessban.RuleTypeEmailRegex,
			Pattern:  `\+[^@]+@hotmail\.com$`,
			Status:   IPBanStatusActive,
		}},
	}
	svc := NewIPBanService(repo)
	_, banned, err := svc.CheckEmail(context.Background(), "cppttlf4390v+baxxsxjh9zj@hotmail.com")
	require.NoError(t, err)
	require.True(t, banned)

	_, banned, err = svc.CheckEmail(context.Background(), "user@hotmail.com")
	require.NoError(t, err)
	require.False(t, banned)
}

func TestIPBanService_CheckClient_UAOnly(t *testing.T) {
	repo := &memoryIPBanRepo{
		rules: []IPBan{{
			ID:       2,
			RuleType: accessban.RuleTypeUA,
			Pattern:  "evilbot",
			Status:   IPBanStatusActive,
		}},
	}
	svc := NewIPBanService(repo)
	_, banned, err := svc.CheckClient(context.Background(), "1.1.1.1", "evilbot/1.0")
	require.NoError(t, err)
	require.True(t, banned)
}
