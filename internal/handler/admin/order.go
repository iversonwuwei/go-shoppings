package admin

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"wechat-mall-saas/internal/pkg/ctxkeys"
	"wechat-mall-saas/internal/pkg/response"
	"wechat-mall-saas/internal/repository"
	"wechat-mall-saas/internal/service"
)

type OrderHandler struct {
	svc *service.OrderService
}

func NewOrderHandler(s *service.OrderService) *OrderHandler { return &OrderHandler{svc: s} }

func (h *OrderHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	q := repository.OrderListQuery{
		Status: c.Query("status"),
		Page:   page,
		Size:   size,
	}
	rows, total, err := h.svc.List(c.Request.Context(), q)
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

func (h *OrderHandler) Logs(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	rows, err := h.svc.ListLogs(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, rows)
}

func (h *OrderHandler) Ship(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var body struct {
		Company string `json:"express_company" binding:"required"`
		No      string `json:"express_no" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	var adminID uint64
	if a := ctxkeys.GetAdmin(c.Request.Context()); a != nil {
		adminID = a.ID
	}
	if err := h.svc.Ship(c.Request.Context(), id, body.Company, body.No, adminID); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, nil)
}

func (h *OrderHandler) Messages(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	status := c.Query("status")
	rows, total, unread, err := h.svc.ListMessages(c.Request.Context(), repository.OrderMessageListQuery{
		Status: status,
		Page:   page,
		Size:   size,
	})
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"list": rows, "total": total, "unread": unread, "page": page, "size": size})
}

func (h *OrderHandler) MarkMessageRead(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.MarkMessageRead(c.Request.Context(), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, nil)
}

func (h *OrderHandler) MarkAllMessagesRead(c *gin.Context) {
	if err := h.svc.MarkAllMessagesRead(c.Request.Context()); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, nil)
}
