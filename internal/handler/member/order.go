package member

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"wechat-mall-saas/internal/pkg/ctxkeys"
	apperr "wechat-mall-saas/internal/pkg/errors"
	"wechat-mall-saas/internal/pkg/response"
	"wechat-mall-saas/internal/repository"
	"wechat-mall-saas/internal/service"
)

type OrderHandler struct {
	svc *service.OrderService
}

func NewOrderHandler(s *service.OrderService) *OrderHandler { return &OrderHandler{svc: s} }

func (h *OrderHandler) Create(c *gin.Context) {
	m := ctxkeys.GetMember(c.Request.Context())
	if m == nil {
		response.Fail(c, apperr.ErrUnauthorized)
		return
	}
	var in service.OrderCreateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	o, err := h.svc.Create(c.Request.Context(), m.ID, &in)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, o)
}

func (h *OrderHandler) List(c *gin.Context) {
	m := ctxkeys.GetMember(c.Request.Context())
	if m == nil {
		response.Fail(c, apperr.ErrUnauthorized)
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	rows, total, err := h.svc.List(c.Request.Context(), repository.OrderListQuery{
		MemberID: m.ID, Status: c.Query("status"), Page: page, Size: size,
	})
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"list": rows, "total": total, "page": page, "size": size})
}

func (h *OrderHandler) Detail(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	o, err := h.svc.Detail(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, o)
}

func (h *OrderHandler) Cancel(c *gin.Context) {
	m := ctxkeys.GetMember(c.Request.Context())
	if m == nil {
		response.Fail(c, apperr.ErrUnauthorized)
		return
	}
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.Cancel(c.Request.Context(), id, m.ID); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, nil)
}

func (h *OrderHandler) Confirm(c *gin.Context) {
	m := ctxkeys.GetMember(c.Request.Context())
	if m == nil {
		response.Fail(c, apperr.ErrUnauthorized)
		return
	}
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.Confirm(c.Request.Context(), id, m.ID); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, nil)
}
