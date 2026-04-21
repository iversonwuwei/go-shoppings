package middleware

import (
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	apperr "wechat-mall-saas/internal/pkg/errors"
	"wechat-mall-saas/internal/pkg/response"
)

// IPRateLimit 简单基于 IP 的令牌桶限流
func IPRateLimit(qps, burst int) gin.HandlerFunc {
	if qps <= 0 {
		return func(c *gin.Context) { c.Next() }
	}
	var (
		mu       sync.Mutex
		limiters = make(map[string]*rate.Limiter)
	)
	get := func(ip string) *rate.Limiter {
		mu.Lock()
		defer mu.Unlock()
		if l, ok := limiters[ip]; ok {
			return l
		}
		l := rate.NewLimiter(rate.Limit(qps), burst)
		limiters[ip] = l
		return l
	}
	return func(c *gin.Context) {
		if !get(c.ClientIP()).Allow() {
			response.Fail(c, apperr.New(42900, "请求过于频繁"))
			c.Abort()
			return
		}
		c.Next()
	}
}
