package handler

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// SubscriptionSummaryItem represents a subscription item in summary
type SubscriptionSummaryItem struct {
	ID              int64   `json:"id"`
	GroupID         int64   `json:"group_id"`
	GroupName       string  `json:"group_name"`
	Status          string  `json:"status"`
	DailyUsedUSD    float64 `json:"daily_used_usd,omitempty"`
	DailyLimitUSD   float64 `json:"daily_limit_usd,omitempty"`
	WeeklyUsedUSD   float64 `json:"weekly_used_usd,omitempty"`
	WeeklyLimitUSD  float64 `json:"weekly_limit_usd,omitempty"`
	MonthlyUsedUSD  float64 `json:"monthly_used_usd,omitempty"`
	MonthlyLimitUSD float64 `json:"monthly_limit_usd,omitempty"`
	ExpiresAt       *string `json:"expires_at,omitempty"`
}

// SubscriptionProgressInfo represents subscription with progress info
type SubscriptionProgressInfo struct {
	Subscription *dto.UserSubscription         `json:"subscription"`
	Progress     *service.SubscriptionProgress `json:"progress"`
}

// SubscriptionHandler handles user subscription operations
type SubscriptionHandler struct {
	subscriptionService *service.SubscriptionService
}

// NewSubscriptionHandler creates a new user subscription handler
func NewSubscriptionHandler(subscriptionService *service.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{
		subscriptionService: subscriptionService,
	}
}

// List handles listing current user's subscriptions
// GET /api/v1/subscriptions
func (h *SubscriptionHandler) List(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}

	subscriptions, err := h.subscriptionService.ListUserSubscriptions(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	out := make([]dto.UserSubscription, 0, len(subscriptions))
	for i := range subscriptions {
		out = append(out, *dto.UserSubscriptionFromService(&subscriptions[i]))
	}
	response.Success(c, out)
}

// GetActive handles getting current user's active subscriptions
// GET /api/v1/subscriptions/active
func (h *SubscriptionHandler) GetActive(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}

	subscriptions, err := h.subscriptionService.ListActiveUserSubscriptions(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	out := make([]dto.UserSubscription, 0, len(subscriptions))
	for i := range subscriptions {
		out = append(out, *dto.UserSubscriptionFromService(&subscriptions[i]))
	}
	response.Success(c, out)
}

// GetProgress handles getting subscription progress for current user
// GET /api/v1/subscriptions/progress
func (h *SubscriptionHandler) GetProgress(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}

	// Get all active subscriptions with progress
	subscriptions, err := h.subscriptionService.ListActiveUserSubscriptions(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	result := make([]SubscriptionProgressInfo, 0, len(subscriptions))
	for i := range subscriptions {
		sub := &subscriptions[i]
		progress, err := h.subscriptionService.GetSubscriptionProgress(c.Request.Context(), sub.ID)
		if err != nil {
			// Skip subscriptions with errors
			continue
		}
		result = append(result, SubscriptionProgressInfo{
			Subscription: dto.UserSubscriptionFromService(sub),
			Progress:     progress,
		})
	}

	response.Success(c, result)
}

// GetSummary handles getting a summary of current user's subscription status
// GET /api/v1/subscriptions/summary
func (h *SubscriptionHandler) GetSummary(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}

	// Get all active subscriptions
	subscriptions, err := h.subscriptionService.ListActiveUserSubscriptions(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	var totalUsed float64
	items := make([]SubscriptionSummaryItem, 0, len(subscriptions))

	for _, sub := range subscriptions {
		item := SubscriptionSummaryItem{
			ID:             sub.ID,
			GroupID:        sub.GroupID,
			Status:         sub.Status,
			DailyUsedUSD:   sub.DailyUsageUSD,
			WeeklyUsedUSD:  sub.WeeklyUsageUSD,
			MonthlyUsedUSD: sub.MonthlyUsageUSD,
		}

		// Add group info if preloaded
		if sub.Group != nil {
			item.GroupName = sub.Group.Name
			if sub.Group.DailyLimitUSD != nil {
				item.DailyLimitUSD = *sub.Group.DailyLimitUSD
			}
			if sub.Group.WeeklyLimitUSD != nil {
				item.WeeklyLimitUSD = *sub.Group.WeeklyLimitUSD
			}
			if sub.Group.MonthlyLimitUSD != nil {
				item.MonthlyLimitUSD = *sub.Group.MonthlyLimitUSD
			}
		}

		// Format expiration time
		if !sub.ExpiresAt.IsZero() {
			formatted := sub.ExpiresAt.Format("2006-01-02T15:04:05Z07:00")
			item.ExpiresAt = &formatted
		}

		// Track total usage (use monthly as the most comprehensive)
		totalUsed += sub.MonthlyUsageUSD

		items = append(items, item)
	}

	summary := struct {
		ActiveCount   int                       `json:"active_count"`
		TotalUsedUSD  float64                   `json:"total_used_usd"`
		Subscriptions []SubscriptionSummaryItem `json:"subscriptions"`
	}{
		ActiveCount:   len(subscriptions),
		TotalUsedUSD:  totalUsed,
		Subscriptions: items,
	}

	response.Success(c, summary)
}

// ConsumeResetCard consumes a purchased reset card with the requested validity.
// POST /api/v1/subscriptions/:id/reset-cards/consume
func (h *SubscriptionHandler) ConsumeResetCard(c *gin.Context) {
	middleware2.SetAuditAction(c, "user.subscription.reset_card.consume")
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}
	var req struct {
		ValidityDays int `json:"validity_days"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.ValidityDays <= 0 || req.ValidityDays > service.MaxValidityDays {
		subscriptionID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
		setResetCardAudit(c, subject.UserID, subscriptionID, req.ValidityDays, "failed", "INVALID_VALIDITY_DAYS")
		response.ErrorFrom(c, infraerrors.BadRequest("INVALID_VALIDITY_DAYS", "validity_days must be a positive integer"))
		return
	}
	h.consumeResetCard(c, req.ValidityDays)
}

// ResetDaily is the compatibility endpoint for consuming a one-day reset card.
// POST /api/v1/subscriptions/:id/reset-daily
func (h *SubscriptionHandler) ResetDaily(c *gin.Context) {
	h.consumeResetCard(c, 1)
}

var resetCardIdempotencyKeyPattern = regexp.MustCompile(`^[A-Za-z0-9._:-]{8,128}$`)

func (h *SubscriptionHandler) consumeResetCard(c *gin.Context, validityDays int) {
	middleware2.SetAuditAction(c, "user.subscription.reset_card.consume")
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}
	idempotencyKey := strings.TrimSpace(c.GetHeader("Idempotency-Key"))
	if !resetCardIdempotencyKeyPattern.MatchString(idempotencyKey) {
		response.ErrorFrom(c, infraerrors.BadRequest("INVALID_IDEMPOTENCY_KEY", "Idempotency-Key must be 8-128 URL-safe characters"))
		return
	}
	subscriptionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || subscriptionID <= 0 {
		setResetCardAudit(c, subject.UserID, -1, validityDays, "failed", "INVALID_SUBSCRIPTION_ID")
		response.ErrorFrom(c, infraerrors.BadRequest("INVALID_SUBSCRIPTION_ID", "invalid subscription ID"))
		return
	}
	updated, err := h.subscriptionService.ConsumeResetCard(c.Request.Context(), service.ConsumeResetCardInput{UserID: subject.UserID, SubscriptionID: subscriptionID, ValidityDays: validityDays, IdempotencyKey: idempotencyKey})
	if err != nil {
		errorCode := infraerrors.Reason(err)
		if errorCode == "" {
			errorCode = "INTERNAL_ERROR"
		}
		result := "failed"
		if errors.Is(err, service.ErrResetCardResponseReloadFailed) {
			result = "success"
		}
		setResetCardAudit(c, subject.UserID, subscriptionID, validityDays, result, errorCode)
		response.ErrorFrom(c, err)
		return
	}
	if err := h.subscriptionService.AttachResetCards(c.Request.Context(), updated); err != nil {
		errorCode := infraerrors.Reason(err)
		if errorCode == "" {
			errorCode = "INTERNAL_ERROR"
		}
		setResetCardAudit(c, subject.UserID, subscriptionID, validityDays, "success", errorCode)
		response.ErrorFrom(c, err)
		return
	}
	setResetCardAudit(c, subject.UserID, subscriptionID, validityDays, "success", "")
	response.Success(c, dto.UserSubscriptionFromService(updated))
}

func setResetCardAudit(c *gin.Context, userID, subscriptionID int64, validityDays int, result, errorCode string) {
	keyHash := sha256.Sum256([]byte(strings.TrimSpace(c.GetHeader("Idempotency-Key"))))
	middleware2.SetAuditExtra(c, map[string]any{
		"user_id": userID, "subscription_id": subscriptionID, "validity_days": validityDays,
		"result": result, "error_code": errorCode,
		"idempotency_key_hash": fmt.Sprintf("%x", keyHash[:8]),
	})
}
