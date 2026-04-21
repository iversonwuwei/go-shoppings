package middleware

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"wechat-mall-saas/internal/pkg/ctxkeys"
	apperr "wechat-mall-saas/internal/pkg/errors"
	"wechat-mall-saas/internal/pkg/response"
	"wechat-mall-saas/internal/service"
)

// Tenant 从 Header X-Tenant-ID 读取租户 ID，注入 TenantInfo 到请求 context
func Tenant(svc *service.TenantService, required bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.GetHeader("X-Tenant-ID")
		if idStr == "" {
			if required {
				response.Fail(c, apperr.ErrTenantRequired)
				c.Abort()
				return
			}
			c.Next()
			return
		}
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil || id == 0 {
			response.Fail(c, apperr.ErrTenantRequired)
			c.Abort()
			return
		}
		info, err := svc.LoadContext(c.Request.Context(), id)
		if err != nil {
			response.Fail(c, err)
			c.Abort()
			return
		}
		// 封禁 / 待审核 => 拒绝
		if info.Status == service.TenantStatusBanned || info.Status == service.TenantStatusPending {
			response.Fail(c, apperr.ErrTenantInvalid)
			c.Abort()
			return
		}
		// 过期 + 非欠费：继续；但写操作会在 service 层 RequireFeature/CheckLimit 处拒绝
		ctx := ctxkeys.WithTenant(c.Request.Context(), info)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
