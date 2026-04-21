package member

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"wechat-mall-saas/internal/pkg/ctxkeys"
	apperr "wechat-mall-saas/internal/pkg/errors"
	"wechat-mall-saas/internal/pkg/response"
	"wechat-mall-saas/internal/service"
)

type CouponHandler struct {
	svc *service.CouponService
}

func NewCouponHandler(s *service.CouponService) *CouponHandler { return &CouponHandler{svc: s} }

func (h *CouponHandler) Available(c *gin.Context) {
	rows, err := h.svc.List(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, rows)
}

func (h *CouponHandler) Receive(c *gin.Context) {
	m := ctxkeys.GetMember(c.Request.Context())
	if m == nil {
		response.Fail(c, apperr.ErrUnauthorized)
		return
	}
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	mc, err := h.svc.Receive(c.Request.Context(), m.ID, id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, mc)
}

func (h *CouponHandler) My(c *gin.Context) {
	m := ctxkeys.GetMember(c.Request.Context())
	if m == nil {
		response.Fail(c, apperr.ErrUnauthorized)
		return
	}
	rows, err := h.svc.MyCoupons(c.Request.Context(), m.ID, c.Query("status"))
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, rows)
}
