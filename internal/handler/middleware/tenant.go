package middleware

import (
	"strconv"
	"strings"

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
		// 待审核 => 拒绝（尚未通过平台审批）
		if info.Status == service.TenantStatusPending {
			response.Fail(c, apperr.ErrTenantInvalid)
			c.Abort()
			return
		}
		// 已封禁 => 仅允许访问订阅付费相关路由，便于自助续订解封
		if info.Status == service.TenantStatusBanned {
			if !strings.Contains(c.Request.URL.Path, "/admin/subscription/") {
				response.FailCode(c, 30021, "租户已封禁，请先完成续订付费")
				c.Abort()
				return
			}
		}
		// Overdue（3~5 天欠费宽限期）：放行读取，写操作由 RequireFeature 拦截
		ctx := ctxkeys.WithTenant(c.Request.Context(), info)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
