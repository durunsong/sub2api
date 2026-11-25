package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveOldBalance_UsesActualSettlement(t *testing.T) {
	before, after := 1.0, 0.0
	p := &postUsageBillingParams{User: &User{Balance: 100}, Cost: &CostBreakdown{ActualCost: 3}}
	result := &UsageBillingApplyResult{OldBalance: &before, NewBalance: &after, BalanceCharged: 1, BalanceAbsorbed: 2}
	require.Equal(t, 1.0, resolveOldBalance(p, result), "notification must not invent a $3 starting balance")
}

func TestResolveBalanceCharge_RoundedZeroCost(t *testing.T) {
	p := &postUsageBillingParams{Cost: &CostBreakdown{ActualCost: 0.000000004}}
	cmd := &UsageBillingCommand{BalanceCost: p.Cost.ActualCost}
	cmd.Normalize()
	require.Zero(t, cmd.BalanceCost)
	// No wallet update means the repository returns no balance snapshot.
	result := &UsageBillingApplyResult{Applied: true}
	require.Zero(t, resolveBalanceCharge(p, result), "cache and notifications must not deduct a cost rounded to zero")
}

func TestBalanceNotificationAmounts_KeepActualCostSeparate(t *testing.T) {
	for _, tc := range []struct {
		name                             string
		before, after, charged, absorbed float64
	}{
		{"partial charge", 1, 0, 1, 2},
		{"already exhausted", 0, 0, 0, 3},
		{"historical debt", -2, -2, 0, 3},
		{"fully paid", 5, 2, 3, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := &postUsageBillingParams{User: &User{Balance: 100}, Cost: &CostBreakdown{ActualCost: 3}}
			result := &UsageBillingApplyResult{OldBalance: &tc.before, NewBalance: &tc.after, BalanceCharged: tc.charged, BalanceAbsorbed: tc.absorbed}
			require.Equal(t, tc.charged, resolveBalanceCharge(p, result))
			require.Equal(t, tc.after, resolveOldBalance(p, result)-resolveBalanceCharge(p, result))
			require.Equal(t, 3.0, p.Cost.ActualCost, "usage and quota accounting must retain the full cost")
		})
	}
}
