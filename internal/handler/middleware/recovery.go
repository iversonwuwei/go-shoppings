package middleware

import (
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"wechat-mall-saas/internal/pkg/logger"
	"wechat-mall-saas/internal/pkg/response"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				if logger.L != nil {
					logger.L.Error("panic recovered",
						zap.Any("err", r),
						zap.String("path", c.Request.URL.Path),
						zap.ByteString("stack", debug.Stack()),
					)
				}
				response.FailCode(c, 50000, "internal server error")
				c.Abort()
			}
		}()
		c.Next()
	}
}
