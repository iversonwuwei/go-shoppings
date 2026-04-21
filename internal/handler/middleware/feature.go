package middleware

import (
	"github.com/gin-gonic/gin"

	"wechat-mall-saas/internal/pkg/ctxkeys"
	apperr "wechat-mall-saas/internal/pkg/errors"
	"wechat-mall-saas/internal/pkg/response"
)

// RequireFeature 要求当前租户所属套餐包含指定功能 code。
// 需在 middleware.Tenant 之后使用。
// - 未加载租户上下文 -> ErrTenantRequired
// - 套餐已过期        -> ErrPlanExpired
// - 套餐未开通该功能  -> ErrFeatureDisabled
func RequireFeature(code string) gin.HandlerFunc {
	return func(c *gin.Context) {
		t := ctxkeys.GetTenant(c.Request.Context())
		if t == nil || t.ID == 0 {
			response.Fail(c, apperr.ErrTenantRequired)
			c.Abort()
			return
		}
		if t.Expired {
			response.Fail(c, apperr.ErrPlanExpired)
			c.Abort()
			return
		}
		for _, f := range t.Features {
			if f == code {
				c.Next()
				return
			}
		}
		response.Fail(c, apperr.ErrFeatureDisabled)
		c.Abort()
	}
}
