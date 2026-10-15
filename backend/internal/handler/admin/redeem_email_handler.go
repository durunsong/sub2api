package admin

import (
	"context"
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

type redeemEmailBatchService interface {
	Defaults(ctx context.Context) service.RedeemEmailDefaults
	SearchRecipients(ctx context.Context, f service.RedeemEmailRecipientFilter) ([]service.RedeemEmailRecipient, error)
	Preview(ctx context.Context, in service.RedeemEmailPreviewInput) (service.NotificationEmailPreview, error)
	CreateBatch(ctx context.Context, in service.CreateRedeemEmailBatchInput) (*service.RedeemEmailBatch, bool, error)
	ListBatches(ctx context.Context, p pagination.PaginationParams) ([]service.RedeemEmailBatch, *pagination.PaginationResult, error)
	GetBatch(ctx context.Context, id int64) (*service.RedeemEmailBatch, error)
	ListItems(ctx context.Context, batchID int64, status string, p pagination.PaginationParams) ([]service.RedeemEmailItem, *pagination.PaginationResult, error)
	RetryFailed(ctx context.Context, batchID int64) (*service.RedeemEmailBatch, error)
	DeleteBatch(ctx context.Context, batchID int64) error
}

// RedeemEmailHandler 处理「未活跃用户余额券群发」管理接口。
type RedeemEmailHandler struct {
	svc redeemEmailBatchService
}

func NewRedeemEmailHandler(svc *service.RedeemEmailService) *RedeemEmailHandler {
	return &RedeemEmailHandler{svc: svc}
}

type RedeemEmailRecipientSearchRequest struct {
	Segments []string `json:"segments" binding:"required,min=1,max=4,dive,oneof=inactive_1_2m inactive_2_3m inactive_3m_plus never_used"`
	Search   string   `json:"search" binding:"max=100"`
}

// RedeemEmailCouponRequest 中 never_expires 与 code_expires_in_days 二选一，避免缺字段时被误当作永久有效。
type RedeemEmailCouponRequest struct {
	CouponValue       float64 `json:"coupon_value" binding:"required,gt=0"`
	CodeExpiresInDays *int    `json:"code_expires_in_days" binding:"omitempty,min=1,max=3650"`
	NeverExpires      bool    `json:"never_expires"`
}

type RedeemEmailPreviewRequest struct {
	RedeemEmailCouponRequest
	Locale string `json:"locale" binding:"omitempty,oneof=zh en"`
}

type CreateRedeemEmailBatchRequest struct {
	RedeemEmailCouponRequest
	Segments []string `json:"segments" binding:"required,min=1,max=4,dive,oneof=inactive_1_2m inactive_2_3m inactive_3m_plus never_used"`
	UserIDs  []int64  `json:"user_ids" binding:"required,min=1,max=1000"`
}

func (r RedeemEmailCouponRequest) expiresInDays() (*int, bool) {
	if r.NeverExpires {
		return nil, r.CodeExpiresInDays == nil
	}
	return r.CodeExpiresInDays, r.CodeExpiresInDays != nil
}

// GET /api/v1/admin/redeem-codes/email-batches/defaults
func (h *RedeemEmailHandler) Defaults(c *gin.Context) {
	response.Success(c, h.svc.Defaults(c.Request.Context()))
}

// POST /api/v1/admin/redeem-codes/email-batches/recipients/search
func (h *RedeemEmailHandler) SearchRecipients(c *gin.Context) {
	var req RedeemEmailRecipientSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	items, err := h.svc.SearchRecipients(c.Request.Context(), service.RedeemEmailRecipientFilter{Segments: req.Segments, Search: req.Search})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"items": items, "total": len(items), "max_recipients": service.RedeemEmailMaxRecipients})
}

// POST /api/v1/admin/redeem-codes/email-batches/preview
func (h *RedeemEmailHandler) Preview(c *gin.Context) {
	var req RedeemEmailPreviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	days, ok := req.expiresInDays()
	if !ok {
		response.ErrorFrom(c, service.ErrRedeemEmailCouponInvalid)
		return
	}
	rendered, err := h.svc.Preview(c.Request.Context(), service.RedeemEmailPreviewInput{
		Locale: req.Locale, CouponValue: req.CouponValue, CodeExpiresInDays: days,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, rendered)
}

// POST /api/v1/admin/redeem-codes/email-batches
// 需要 Idempotency-Key 请求头；同 key 同请求重放原批次（响应头 Idempotency-Replayed: true），同 key 不同请求返回 409。
func (h *RedeemEmailHandler) Create(c *gin.Context) {
	middleware2.SetAuditAction(c, "admin.redeem_email_batches.create")
	var req CreateRedeemEmailBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}
	days, ok := req.expiresInDays()
	if !ok {
		response.ErrorFrom(c, service.ErrRedeemEmailCouponInvalid)
		return
	}
	batch, replayed, err := h.svc.CreateBatch(c.Request.Context(), service.CreateRedeemEmailBatchInput{
		CreatedBy:         subject.UserID,
		IdempotencyKey:    strings.TrimSpace(c.GetHeader("Idempotency-Key")),
		Filter:            service.RedeemEmailRecipientFilter{Segments: req.Segments},
		UserIDs:           req.UserIDs,
		CouponValue:       req.CouponValue,
		CodeExpiresInDays: days,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	if replayed {
		c.Header("Idempotency-Replayed", "true")
	}
	response.Accepted(c, batch)
}

// GET /api/v1/admin/redeem-codes/email-batches
func (h *RedeemEmailHandler) List(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	items, result, err := h.svc.ListBatches(c.Request.Context(), pagination.PaginationParams{Page: page, PageSize: pageSize})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, items, result.Total, page, pageSize)
}

// GET /api/v1/admin/redeem-codes/email-batches/:batch_id
func (h *RedeemEmailHandler) Get(c *gin.Context) {
	id, ok := parseRedeemEmailBatchID(c)
	if !ok {
		return
	}
	batch, err := h.svc.GetBatch(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, batch)
}

// GET /api/v1/admin/redeem-codes/email-batches/:batch_id/items
func (h *RedeemEmailHandler) ListItems(c *gin.Context) {
	id, ok := parseRedeemEmailBatchID(c)
	if !ok {
		return
	}
	page, pageSize := response.ParsePagination(c)
	items, result, err := h.svc.ListItems(c.Request.Context(), id, strings.TrimSpace(c.Query("status")), pagination.PaginationParams{Page: page, PageSize: pageSize})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, items, result.Total, page, pageSize)
}

// POST /api/v1/admin/redeem-codes/email-batches/:batch_id/retry-failed
func (h *RedeemEmailHandler) RetryFailed(c *gin.Context) {
	middleware2.SetAuditAction(c, "admin.redeem_email_batches.retry_failed")
	id, ok := parseRedeemEmailBatchID(c)
	if !ok {
		return
	}
	batch, err := h.svc.RetryFailed(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, batch)
}

// DELETE /api/v1/admin/redeem-codes/email-batches/:batch_id
// 仅删除已结束批次的发送记录；已生成的兑换码保留在兑换码列表中（备注 redeem_email_batch:<id>）。
func (h *RedeemEmailHandler) Delete(c *gin.Context) {
	middleware2.SetAuditAction(c, "admin.redeem_email_batches.delete")
	id, ok := parseRedeemEmailBatchID(c)
	if !ok {
		return
	}
	if err := h.svc.DeleteBatch(c.Request.Context(), id); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"message": "Redeem email batch deleted successfully"})
}

func parseRedeemEmailBatchID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("batch_id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid batch ID")
		return 0, false
	}
	return id, true
}
