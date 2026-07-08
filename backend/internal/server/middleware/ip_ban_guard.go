package middleware

import (
	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/accessban"
	ippkg "github.com/Wei-Shaw/sub2api/internal/pkg/ip"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// IPBanGuard blocks globally banned client IPs/CIDRs/UA combinations for normal API responses.
func IPBanGuard(ipBanService *service.IPBanService, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		ban, blocked, err := evaluateAccessBan(c, ipBanService, cfg)
		if err != nil {
			AbortWithError(c, 500, "INTERNAL_ERROR", "Failed to check access ban status")
			return
		}
		if !blocked {
			c.Next()
			return
		}
		if ban != nil && accessban.NormalizeRuleType(ban.RuleType) == accessban.RuleTypeIP {
			response.ErrorFrom(c, service.ErrIPBanned)
			return
		}
		response.ErrorFrom(c, service.ErrClientAccessBanned)
	}
}

// GatewayIPBanGuard blocks globally banned client signals while preserving gateway error shape.
func GatewayIPBanGuard(ipBanService *service.IPBanService, cfg *config.Config, writeError GatewayErrorWriter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if writeError == nil {
			writeError = AnthropicErrorWriter
		}
		ban, blocked, err := evaluateAccessBan(c, ipBanService, cfg)
		if err != nil {
			AbortWithError(c, 500, "INTERNAL_ERROR", "Failed to check access ban status")
			return
		}
		if !blocked {
			c.Next()
			return
		}
		service.MarkOpsClientBusinessLimited(c, service.OpsClientBusinessLimitedReasonIPRestriction)
		writeError(c, 403, clientAccessBanMessage(ban))
		c.Abort()
	}
}

func evaluateAccessBan(c *gin.Context, ipBanService *service.IPBanService, cfg *config.Config) (*service.IPBan, bool, error) {
	if ipBanService == nil {
		return nil, false, nil
	}
	_ = cfg
	clientIP := ippkg.GetClientIP(c)
	userAgent := c.GetHeader("User-Agent")
	return ipBanService.CheckClient(c.Request.Context(), clientIP, userAgent)
}

func clientAccessBanMessage(ban *service.IPBan) string {
	if ban != nil && accessban.NormalizeRuleType(ban.RuleType) == accessban.RuleTypeIP {
		return service.ErrIPBanned.Message
	}
	return service.ErrClientAccessBanned.Message
}
