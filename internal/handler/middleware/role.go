package middleware

import (
	"github.com/gin-gonic/gin"

	"wechat-mall-saas/internal/pkg/ctxkeys"
	apperr "wechat-mall-saas/internal/pkg/errors"
	"wechat-mall-saas/internal/pkg/response"
)

// RequireRole 限制只有指定角色之一可以访问（role 为空视为未授权）
// super 角色视作超级管理员，永远放行
func RequireRole(roles ...string) gin.HandlerFunc {
	allow := make(map[string]struct{}, len(roles)+1)
	for _, r := range roles {
		allow[r] = struct{}{}
	}
	allow["super"] = struct{}{}
	return func(c *gin.Context) {
		a := ctxkeys.GetAdmin(c.Request.Context())
		if a == nil {
			response.Fail(c, apperr.ErrUnauthorized)
			c.Abort()
			return
		}
		if _, ok := allow[a.Role]; !ok {
			response.FailCode(c, 10003, "权限不足")
			c.Abort()
			return
		}
		c.Next()
	}
}
