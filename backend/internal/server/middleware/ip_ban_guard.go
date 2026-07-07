package middleware

import (
	"github.com/Wei-Shaw/sub2api/internal/config"
	ippkg "github.com/Wei-Shaw/sub2api/internal/pkg/ip"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// IPBanGuard blocks globally banned client IPs/CIDRs for normal API responses.
func IPBanGuard(ipBanService *service.IPBanService, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if checkIPBan(c, ipBanService, cfg, func(c *gin.Context, status int, message string) {
			response.ErrorFrom(c, service.ErrIPBanned)
		}) {
			return
		}
		c.Next()
	}
}

// GatewayIPBanGuard blocks globally banned IPs while preserving gateway error shape.
func GatewayIPBanGuard(ipBanService *service.IPBanService, cfg *config.Config, writeError GatewayErrorWriter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if writeError == nil {
			writeError = AnthropicErrorWriter
		}
		if checkIPBan(c, ipBanService, cfg, writeError) {
			return
		}
		c.Next()
	}
}

func checkIPBan(c *gin.Context, ipBanService *service.IPBanService, cfg *config.Config, writeError GatewayErrorWriter) bool {
	if ipBanService == nil {
		return false
	}
	clientIP := getIPBanClientIP(c, cfg)
	_, banned, err := ipBanService.Check(c.Request.Context(), clientIP)
	if err != nil {
		AbortWithError(c, 500, "INTERNAL_ERROR", "Failed to check IP ban status")
		return true
	}
	if !banned {
		return false
	}
	service.MarkOpsClientBusinessLimited(c, service.OpsClientBusinessLimitedReasonIPRestriction)
	writeError(c, 403, service.ErrIPBanned.Message)
	c.Abort()
	return true
}

func getIPBanClientIP(c *gin.Context, cfg *config.Config) string {
	if cfg != nil && cfg.TrustForwardedIPForAPIKeyACL() {
		return ippkg.GetClientIP(c)
	}
	return ippkg.GetTrustedClientIP(c)
}
