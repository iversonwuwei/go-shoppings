package admin

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"wechat-mall-saas/internal/pkg/ctxkeys"
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

func (h *AfterSaleHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	orderID, _ := strconv.ParseUint(c.Query("order_id"), 10, 64)
	memberID, _ := strconv.ParseUint(c.Query("member_id"), 10, 64)
	rows, total, err := h.svc.ListForAdmin(c.Request.Context(), repository.AfterSaleListQuery{
		OrderID:  orderID,
		MemberID: memberID,
		Status:   c.Query("status"),
		Page:     page,
		Size:     size,
	})
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"list": rows, "total": total, "page": page, "size": size})
}

func (h *AfterSaleHandler) Detail(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	out, err := h.svc.Detail(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, out)
}

type afterSaleRemarkBody struct {
	Remark string `json:"remark"`
}

func (h *AfterSaleHandler) Approve(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var body afterSaleRemarkBody
	_ = c.ShouldBindJSON(&body)
	if err := h.svc.Approve(c.Request.Context(), id, currentAdminID(c), body.Remark); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, nil)
}

func (h *AfterSaleHandler) Reject(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var body afterSaleRemarkBody
	_ = c.ShouldBindJSON(&body)
	if err := h.svc.Reject(c.Request.Context(), id, currentAdminID(c), body.Remark); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, nil)
}

func (h *AfterSaleHandler) Receive(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var body afterSaleRemarkBody
	_ = c.ShouldBindJSON(&body)
	if err := h.svc.Receive(c.Request.Context(), id, currentAdminID(c), body.Remark); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, nil)
}

func (h *AfterSaleHandler) Refund(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var body afterSaleRemarkBody
	_ = c.ShouldBindJSON(&body)
	if err := h.svc.Refund(c.Request.Context(), id, currentAdminID(c), body.Remark); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, nil)
}

func currentAdminID(c *gin.Context) uint64 {
	admin := ctxkeys.GetAdmin(c.Request.Context())
	if admin == nil {
		return 0
	}
	return admin.ID
}
