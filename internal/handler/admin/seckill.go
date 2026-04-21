package admin

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"wechat-mall-saas/internal/model"
	apperr "wechat-mall-saas/internal/pkg/errors"
	"wechat-mall-saas/internal/pkg/response"
	"wechat-mall-saas/internal/repository"
)

type SeckillHandler struct {
	repo *repository.SeckillRepo
}

func NewSeckillHandler(r *repository.SeckillRepo) *SeckillHandler { return &SeckillHandler{repo: r} }

type seckillProductReq struct {
	ProductID    uint64          `json:"product_id" binding:"required"`
	SKUID        uint64          `json:"sku_id"`
	SeckillPrice decimal.Decimal `json:"seckill_price" binding:"required"`
	Stock        int             `json:"stock" binding:"required,min=1"`
}

type seckillReq struct {
	Name       string              `json:"name" binding:"required,max=100"`
	StartAt    time.Time           `json:"start_at" binding:"required"`
	EndAt      time.Time           `json:"end_at" binding:"required"`
	PerLimit   int                 `json:"per_limit" binding:"min=1"`
	TotalStock int                 `json:"total_stock" binding:"min=1"`
	Status     int8                `json:"status"`
	Products   []seckillProductReq `json:"products" binding:"required,min=1"`
}

func (h *SeckillHandler) List(c *gin.Context) {
	rows, err := h.repo.ListActivities(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, rows)
}

func (h *SeckillHandler) Create(c *gin.Context) {
	var req seckillReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	if !req.EndAt.After(req.StartAt) {
		response.FailCode(c, 20001, "结束时间必须晚于开始时间")
		return
	}
	a := &model.SeckillActivity{
		Name: req.Name, StartAt: req.StartAt, EndAt: req.EndAt,
		PerLimit: defaultInt(req.PerLimit, 1), TotalStock: req.TotalStock,
		Status: defaultStatus(req.Status),
	}
	products := toSeckillProducts(req.Products)
	if err := h.repo.CreateActivity(c.Request.Context(), a, products); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, a)
}

func (h *SeckillHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, apperr.ErrParamInvalid)
		return
	}
	exist, err := h.repo.FindActivity(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	if exist == nil {
		response.Fail(c, apperr.ErrNotFound)
		return
	}
	var req seckillReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	if !req.EndAt.After(req.StartAt) {
		response.FailCode(c, 20001, "结束时间必须晚于开始时间")
		return
	}
	exist.Name = req.Name
	exist.StartAt = req.StartAt
	exist.EndAt = req.EndAt
	exist.PerLimit = defaultInt(req.PerLimit, 1)
	exist.TotalStock = req.TotalStock
	exist.Status = defaultStatus(req.Status)
	if err := h.repo.UpdateActivity(c.Request.Context(), exist, toSeckillProducts(req.Products)); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"id": id})
}

func (h *SeckillHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, apperr.ErrParamInvalid)
		return
	}
	if err := h.repo.DeleteActivity(c.Request.Context(), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"id": id})
}

func toSeckillProducts(in []seckillProductReq) []model.SeckillProduct {
	out := make([]model.SeckillProduct, 0, len(in))
	for _, p := range in {
		out = append(out, model.SeckillProduct{
			ProductID:    p.ProductID,
			SKUID:        p.SKUID,
			SeckillPrice: p.SeckillPrice,
			Stock:        p.Stock,
		})
	}
	return out
}

func defaultInt(v, d int) int {
	if v <= 0 {
		return d
	}
	return v
}

func defaultStatus(v int8) int8 {
	if v == 0 {
		return 1
	}
	return v
}
