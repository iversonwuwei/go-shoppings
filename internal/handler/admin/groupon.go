package admin

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"wechat-mall-saas/internal/model"
	apperr "wechat-mall-saas/internal/pkg/errors"
	"wechat-mall-saas/internal/pkg/response"
	"wechat-mall-saas/internal/repository"
)

type GrouponHandler struct {
	repo *repository.GrouponRepo
}

func NewGrouponHandler(r *repository.GrouponRepo) *GrouponHandler { return &GrouponHandler{repo: r} }

type grouponReq struct {
	Name          string          `json:"name" binding:"required,max=100"`
	ProductID     uint64          `json:"product_id" binding:"required"`
	SKUID         uint64          `json:"sku_id"`
	GroupPrice    decimal.Decimal `json:"group_price" binding:"required"`
	OriginalPrice decimal.Decimal `json:"original_price" binding:"required"`
	RequireNum    int             `json:"require_num" binding:"required,min=2"`
	ExpireHours   int             `json:"expire_hours" binding:"required,min=1"`
	TotalStock    int             `json:"total_stock" binding:"min=0"`
	StartAt       requestTime     `json:"start_at" binding:"required"`
	EndAt         requestTime     `json:"end_at" binding:"required"`
	Status        int8            `json:"status"`
}

func (h *GrouponHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	rows, total, err := h.repo.ListActivities(c.Request.Context(), page, size)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"list": rows, "total": total, "page": page, "size": size})
}

func (h *GrouponHandler) Create(c *gin.Context) {
	var req grouponReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	if req.StartAt.IsZero() || req.EndAt.IsZero() {
		response.FailCode(c, 20001, "请选择活动时间")
		return
	}
	if !req.EndAt.After(req.StartAt.Time) {
		response.FailCode(c, 20001, "结束时间必须晚于开始时间")
		return
	}
	a := &model.GrouponActivity{
		Name:          req.Name,
		ProductID:     req.ProductID,
		SKUID:         req.SKUID,
		GroupPrice:    req.GroupPrice,
		OriginalPrice: req.OriginalPrice,
		RequireNum:    req.RequireNum,
		ExpireHours:   req.ExpireHours,
		TotalStock:    req.TotalStock,
		StartAt:       req.StartAt.Time,
		EndAt:         req.EndAt.Time,
		Status:        defaultCouponStatus(req.Status),
	}
	if err := h.repo.CreateActivity(c.Request.Context(), a); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, a)
}

func (h *GrouponHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.Fail(c, apperr.ErrParamInvalid)
		return
	}
	var req grouponReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	if req.StartAt.IsZero() || req.EndAt.IsZero() {
		response.FailCode(c, 20001, "请选择活动时间")
		return
	}
	if !req.EndAt.After(req.StartAt.Time) {
		response.FailCode(c, 20001, "结束时间必须晚于开始时间")
		return
	}
	fields := map[string]interface{}{
		"name":           req.Name,
		"product_id":     req.ProductID,
		"sku_id":         req.SKUID,
		"group_price":    req.GroupPrice,
		"original_price": req.OriginalPrice,
		"require_num":    req.RequireNum,
		"expire_hours":   req.ExpireHours,
		"total_stock":    req.TotalStock,
		"start_at":       req.StartAt.Time,
		"end_at":         req.EndAt.Time,
		"status":         defaultCouponStatus(req.Status),
	}
	if err := h.repo.UpdateActivity(c.Request.Context(), id, fields); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"id": id})
}

func (h *GrouponHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.Fail(c, apperr.ErrParamInvalid)
		return
	}
	if err := h.repo.DeleteActivity(c.Request.Context(), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, nil)
}

// Groupons 列出某活动下的团单（activity_id=0 表示列所有）
func (h *GrouponHandler) Groupons(c *gin.Context) {
	activityID, _ := strconv.ParseUint(c.Query("activity_id"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	rows, total, err := h.repo.ListGroupons(c.Request.Context(), activityID, page, size)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"list": rows, "total": total, "page": page, "size": size})
}
