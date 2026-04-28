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

type AfterSaleHandler struct {
	svc *service.AfterSaleService
}

func NewAfterSaleHandler(svc *service.AfterSaleService) *AfterSaleHandler {
	return &AfterSaleHandler{svc: svc}
}

func (h *AfterSaleHandler) Reasons(c *gin.Context) {
	rows, err := h.svc.ListReasons(c.Request.Context(), c.Query("type"))
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, rows)
}

func (h *AfterSaleHandler) Apply(c *gin.Context) {
	member := ctxkeys.GetMember(c.Request.Context())
	if member == nil {
		response.Fail(c, apperr.ErrUnauthorized)
		return
	}
	orderID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var input service.AfterSaleApplyInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	out, err := h.svc.Apply(c.Request.Context(), member.ID, orderID, input)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, out)
}

func (h *AfterSaleHandler) List(c *gin.Context) {
	member := ctxkeys.GetMember(c.Request.Context())
	if member == nil {
		response.Fail(c, apperr.ErrUnauthorized)
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	orderID, _ := strconv.ParseUint(c.Query("order_id"), 10, 64)
	rows, total, err := h.svc.ListForMember(c.Request.Context(), member.ID, repository.AfterSaleListQuery{
		OrderID: orderID,
		Status:  c.Query("status"),
		Page:    page,
		Size:    size,
	})
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"list": rows, "total": total, "page": page, "size": size})
}

func (h *AfterSaleHandler) Detail(c *gin.Context) {
	member := ctxkeys.GetMember(c.Request.Context())
	if member == nil {
		response.Fail(c, apperr.ErrUnauthorized)
		return
	}
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	out, err := h.svc.DetailForMember(c.Request.Context(), id, member.ID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, out)
}

func (h *AfterSaleHandler) Cancel(c *gin.Context) {
	member := ctxkeys.GetMember(c.Request.Context())
	if member == nil {
		response.Fail(c, apperr.ErrUnauthorized)
		return
	}
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.Cancel(c.Request.Context(), id, member.ID); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, nil)
}

func (h *AfterSaleHandler) SubmitReturn(c *gin.Context) {
	member := ctxkeys.GetMember(c.Request.Context())
	if member == nil {
		response.Fail(c, apperr.ErrUnauthorized)
		return
	}
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var input service.AfterSaleReturnInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	if err := h.svc.SubmitReturn(c.Request.Context(), id, member.ID, input); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, nil)
}
