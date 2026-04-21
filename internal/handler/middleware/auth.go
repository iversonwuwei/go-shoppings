package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"wechat-mall-saas/internal/pkg/ctxkeys"
	apperr "wechat-mall-saas/internal/pkg/errors"
	"wechat-mall-saas/internal/pkg/jwtx"
	"wechat-mall-saas/internal/pkg/response"
)

func bearer(c *gin.Context) string {
	h := c.GetHeader("Authorization")
	if strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer ")
	}
	return ""
}

// AdminAuth 管理员 JWT 鉴权
func AdminAuth(j *jwtx.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		tok := bearer(c)
		if tok == "" {
			response.Fail(c, apperr.ErrUnauthorized)
			c.Abort()
			return
		}
		claims, err := j.Parse(tok)
		if err != nil || claims.Subject != jwtx.SubjectAdmin {
			response.Fail(c, apperr.ErrInvalidToken)
			c.Abort()
			return
		}
		info := &ctxkeys.AdminInfo{
			ID:       claims.UserID,
			TenantID: claims.TenantID,
			Role:     claims.Role,
		}
		ctx := ctxkeys.WithAdmin(c.Request.Context(), info)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// MemberAuth 小程序会员 JWT 鉴权
func MemberAuth(j *jwtx.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		tok := bearer(c)
		if tok == "" {
			response.Fail(c, apperr.ErrUnauthorized)
			c.Abort()
			return
		}
		claims, err := j.Parse(tok)
		if err != nil || claims.Subject != jwtx.SubjectMember {
			response.Fail(c, apperr.ErrInvalidToken)
			c.Abort()
			return
		}
		info := &ctxkeys.MemberInfo{ID: claims.UserID, OpenID: claims.OpenID, TenantID: claims.TenantID}
		ctx := ctxkeys.WithMember(c.Request.Context(), info)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
