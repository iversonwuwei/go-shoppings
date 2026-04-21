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

type MemberLevelHandler struct {
	repo *repository.MemberLevelRepo
}

func NewMemberLevelHandler(r *repository.MemberLevelRepo) *MemberLevelHandler {
	return &MemberLevelHandler{repo: r}
}

type memberLevelReq struct {
	Name         string          `json:"name" binding:"required,max=30"`
	Icon         string          `json:"icon"`
	Color        string          `json:"color"`
	MinGrowth    int             `json:"min_growth"`
	DiscountRate decimal.Decimal `json:"discount_rate"`
	PointsMult   decimal.Decimal `json:"points_mult"`
	Sort         int             `json:"sort"`
}

func (h *MemberLevelHandler) List(c *gin.Context) {
	rows, err := h.repo.List(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, rows)
}

func (h *MemberLevelHandler) Create(c *gin.Context) {
	var req memberLevelReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	m := &model.MemberLevel{
		Name:         req.Name,
		Icon:         req.Icon,
		Color:        req.Color,
		MinGrowth:    req.MinGrowth,
		DiscountRate: nonZeroDecimal(req.DiscountRate, decimal.NewFromInt(100)),
		PointsMult:   nonZeroDecimal(req.PointsMult, decimal.NewFromInt(1)),
		Sort:         req.Sort,
	}
	if err := h.repo.Create(c.Request.Context(), m); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, m)
}

func (h *MemberLevelHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, apperr.ErrParamInvalid)
		return
	}
	var req memberLevelReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	exist, err := h.repo.FindByID(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	if exist == nil {
		response.Fail(c, apperr.ErrNotFound)
		return
	}
	exist.Name = req.Name
	exist.Icon = req.Icon
	exist.Color = req.Color
	exist.MinGrowth = req.MinGrowth
	exist.DiscountRate = nonZeroDecimal(req.DiscountRate, decimal.NewFromInt(100))
	exist.PointsMult = nonZeroDecimal(req.PointsMult, decimal.NewFromInt(1))
	exist.Sort = req.Sort
	if err := h.repo.Update(c.Request.Context(), exist); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, exist)
}

func (h *MemberLevelHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, apperr.ErrParamInvalid)
		return
	}
	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"id": id})
}

func nonZeroDecimal(v, def decimal.Decimal) decimal.Decimal {
	if v.IsZero() {
		return def
	}
	return v
}
