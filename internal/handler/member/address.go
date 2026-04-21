package member

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"wechat-mall-saas/internal/model"
	"wechat-mall-saas/internal/pkg/ctxkeys"
	apperr "wechat-mall-saas/internal/pkg/errors"
	"wechat-mall-saas/internal/pkg/response"
	"wechat-mall-saas/internal/service"
)

type AddressHandler struct {
	svc *service.MemberService
}

func NewAddressHandler(s *service.MemberService) *AddressHandler { return &AddressHandler{svc: s} }

func (h *AddressHandler) List(c *gin.Context) {
	m := ctxkeys.GetMember(c.Request.Context())
	if m == nil {
		response.Fail(c, apperr.ErrUnauthorized)
		return
	}
	rows, err := h.svc.Addresses(c.Request.Context(), m.ID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, rows)
}

func (h *AddressHandler) Create(c *gin.Context) {
	m := ctxkeys.GetMember(c.Request.Context())
	if m == nil {
		response.Fail(c, apperr.ErrUnauthorized)
		return
	}
	var a model.MemberAddress
	if err := c.ShouldBindJSON(&a); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	a.MemberID = m.ID
	if err := h.svc.CreateAddress(c.Request.Context(), &a); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, a)
}

type PointsHandler struct {
	svc *service.MemberService
}

func NewPointsHandler(s *service.MemberService) *PointsHandler { return &PointsHandler{svc: s} }

func (h *PointsHandler) Logs(c *gin.Context) {
	m := ctxkeys.GetMember(c.Request.Context())
	if m == nil {
		response.Fail(c, apperr.ErrUnauthorized)
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	rows, total, err := h.svc.PointsLogs(c.Request.Context(), m.ID, page, size)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"list": rows, "total": total, "page": page, "size": size})
}
