package admin

import (
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/accessban"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	servermiddleware "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// IPBanHandler handles global access ban management.
type IPBanHandler struct {
	ipBanService *service.IPBanService
}

func NewIPBanHandler(ipBanService *service.IPBanService) *IPBanHandler {
	return &IPBanHandler{ipBanService: ipBanService}
}

type CreateIPBanRequest struct {
	RuleType  string `json:"rule_type"`
	Pattern   string `json:"pattern" binding:"required"`
	UAPattern string `json:"ua_pattern"`
	Reason    string `json:"reason"`
	Source    string `json:"source"`
	ExpiresAt *int64 `json:"expires_at"`
}

type UpdateIPBanRequest struct {
	RuleType  *string `json:"rule_type"`
	Pattern   *string `json:"pattern"`
	UAPattern *string `json:"ua_pattern"`
	Reason    *string `json:"reason"`
	Status    *string `json:"status" binding:"omitempty,oneof=active inactive"`
	ExpiresAt *int64  `json:"expires_at"`
}

func (h *IPBanHandler) List(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	search := strings.TrimSpace(c.Query("search"))
	if len(search) > 100 {
		search = search[:100]
	}
	filters := service.IPBanListFilters{
		Search:   search,
		Status:   strings.TrimSpace(c.Query("status")),
		RuleType: accessban.NormalizeRuleType(c.Query("rule_type")),
	}
	params := pagination.PaginationParams{
		Page:      page,
		PageSize:  pageSize,
		SortBy:    c.DefaultQuery("sort_by", "created_at"),
		SortOrder: c.DefaultQuery("sort_order", "desc"),
	}

	bans, result, err := h.ipBanService.List(c.Request.Context(), params, filters)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, bans, result.Total, page, pageSize)
}

func (h *IPBanHandler) GetByID(c *gin.Context) {
	id, ok := parseIPBanID(c)
	if !ok {
		return
	}
	ban, err := h.ipBanService.GetByID(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, ban)
}

func (h *IPBanHandler) Create(c *gin.Context) {
	var req CreateIPBanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	input := service.CreateIPBanInput{
		RuleType:  req.RuleType,
		Pattern:   req.Pattern,
		UAPattern: req.UAPattern,
		Reason:    req.Reason,
		Source:    req.Source,
	}
	if subject, ok := servermiddleware.GetAuthSubjectFromContext(c); ok {
		input.CreatedBy = &subject.UserID
	}
	if req.ExpiresAt != nil && *req.ExpiresAt > 0 {
		t := time.Unix(*req.ExpiresAt, 0)
		input.ExpiresAt = &t
	}

	ban, err := h.ipBanService.Create(c.Request.Context(), input)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, ban)
}

func (h *IPBanHandler) Update(c *gin.Context) {
	id, ok := parseIPBanID(c)
	if !ok {
		return
	}
	var req UpdateIPBanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	input := service.UpdateIPBanInput{
		RuleType:  req.RuleType,
		Pattern:   req.Pattern,
		UAPattern: req.UAPattern,
		Reason:    req.Reason,
		Status:    req.Status,
	}
	if req.ExpiresAt != nil {
		var expiresAt *time.Time
		if *req.ExpiresAt > 0 {
			t := time.Unix(*req.ExpiresAt, 0)
			expiresAt = &t
		}
		input.ExpiresAt = &expiresAt
	}

	ban, err := h.ipBanService.Update(c.Request.Context(), id, input)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, ban)
}

func (h *IPBanHandler) Delete(c *gin.Context) {
	id, ok := parseIPBanID(c)
	if !ok {
		return
	}
	if err := h.ipBanService.Delete(c.Request.Context(), id); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"message": "Access ban rule deleted successfully"})
}

func parseIPBanID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid access ban rule ID")
		return 0, false
	}
	return id, true
}
