package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"wechat-mall-saas/internal/pkg/ctxkeys"
	"wechat-mall-saas/internal/pkg/logger"
)

func Logging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		if logger.L == nil {
			return
		}
		var tenantID uint64
		if t := ctxkeys.GetTenant(c.Request.Context()); t != nil {
			tenantID = t.ID
		}
		logger.L.Info("access",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("cost", time.Since(start)),
			zap.Uint64("tenant_id", tenantID),
			zap.String("ip", c.ClientIP()),
		)
	}
}
