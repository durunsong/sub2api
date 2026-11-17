package service

import (
	"context"
	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestImageStudioPriceUsesSettlementRatesAndModes(t *testing.T) {
	bill := NewBillingService(&config.Config{}, nil)
	rate := 0.5
	price := 0.2
	svc := &GatewayService{billingService: bill, userGroupRateRepo: &userGroupRateRepoHotpathStub{rate: &rate}}
	svc.resolver = NewModelPricingResolver(nil, bill)
	group := &Group{ID: 3, RateMultiplier: 2, ImagePrice1K: &price, ImagePrice2K: &price, ImagePrice4K: &price}
	quote := svc.ImageStudioPricing(context.Background(), 7, group, "gpt-image-2")
	require.Equal(t, "image", quote.Mode)
	require.InDelta(t, 0.1, *quote.ImagePrice1K, 1e-10)
	require.Equal(t, rate, quote.RateMultiplier)
	group.ImageRateIndependent = true
	group.ImageRateMultiplier = 3
	quote = svc.ImageStudioPricing(context.Background(), 7, group, "gpt-image-2")
	require.InDelta(t, 0.6, *quote.ImagePrice1K, 1e-10)
	group.ModelPricing = []ChannelModelPricing{{Models: []string{"gpt-image-2"}, BillingMode: BillingModeToken}}
	quote = svc.ImageStudioPricing(context.Background(), 7, group, "gpt-image-2")
	require.Equal(t, "token", quote.Mode)
	require.Nil(t, quote.ImagePrice1K)
	require.Equal(t, rate, quote.RateMultiplier)
	group.ImageCountBilling = true
	quote = svc.ImageStudioPricing(context.Background(), 7, group, "gpt-image-2")
	require.Equal(t, "image", quote.Mode)
	require.InDelta(t, 0.6, *quote.ImagePrice1K, 1e-10)
}

func TestImageStudioAliasPriceUsesCanonicalSettlementModel(t *testing.T) {
	bill := NewBillingService(&config.Config{}, nil)
	svc := &GatewayService{billingService: bill, resolver: NewModelPricingResolver(nil, bill)}
	group := &Group{ID: 3, RateMultiplier: 2, ModelPricing: []ChannelModelPricing{{Models: []string{"gemini-3-pro-image"}, BillingMode: BillingModeToken}}}
	quote := svc.ImageStudioPricing(context.Background(), 0, group, "nano-banana", "gemini-3-pro-image")
	require.Equal(t, "token", quote.Mode)
	require.Nil(t, quote.ImagePrice1K)
	price := 0.8
	group.ModelPricing = []ChannelModelPricing{{Models: []string{"gemini-3-pro-image"}, BillingMode: BillingModeImage, PerRequestPrice: &price}}
	quote = svc.ImageStudioPricing(context.Background(), 0, group, "nano-banana", "gemini-3-pro-image")
	require.Equal(t, "image", quote.Mode)
	require.InDelta(t, 1.6, *quote.ImagePrice1K, 1e-10)
}
