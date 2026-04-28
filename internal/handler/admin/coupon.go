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

type CouponHandler struct {
	repo *repository.CouponRepo
}

func NewCouponHandler(r *repository.CouponRepo) *CouponHandler { return &CouponHandler{repo: r} }

type couponReq struct {
	Name             string           `json:"name" binding:"required,max=50"`
	Type             string           `json:"type" binding:"required,oneof=cash discount shipping"`
	ThresholdAmount  *decimal.Decimal `json:"threshold_amount"`
	DiscountValue    decimal.Decimal  `json:"discount_value" binding:"required"`
	MaxDiscount      *decimal.Decimal `json:"max_discount"`
	ReceiveLimitType string           `json:"receive_limit_type"`
	TotalCount       int              `json:"total_count" binding:"min=0"`
	PerLimit         int              `json:"per_limit" binding:"min=0"`
	UseLimit         int              `json:"use_limit" binding:"min=0"`
	ReceiveStartAt   *requestTime     `json:"receive_start_at"`
	ReceiveEndAt     *requestTime     `json:"receive_end_at"`
	ValidStartAt     *requestTime     `json:"valid_start_at"`
	ValidEndAt       *requestTime     `json:"valid_end_at"`
	ValidDays        int              `json:"valid_days"`
	ApplicableType   string           `json:"applicable_type"`
	Status           int8             `json:"status"`
}

func (h *CouponHandler) List(c *gin.Context) {
	rows, err := h.repo.ListAll(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, rows)
}

func (h *CouponHandler) Create(c *gin.Context) {
	var req couponReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	if req.ApplicableType == "" {
		req.ApplicableType = "all"
	}
	receiveLimitType := defaultCouponReceiveLimitType(req.ReceiveLimitType)
	totalCount, remainCount := couponCounts(receiveLimitType, req.TotalCount, req.TotalCount)
	coupon := &model.Coupon{
		Name: req.Name, Type: req.Type,
		ThresholdAmount: req.ThresholdAmount, DiscountValue: req.DiscountValue, MaxDiscount: req.MaxDiscount,
		ReceiveLimitType: receiveLimitType,
		TotalCount:       totalCount,
		RemainCount:      remainCount,
		PerLimit:         req.PerLimit,
		UseLimit:         req.UseLimit,
		ReceiveStartAt:   requestTimePtr(req.ReceiveStartAt), ReceiveEndAt: requestTimePtr(req.ReceiveEndAt),
		ValidStartAt: requestTimePtr(req.ValidStartAt), ValidEndAt: requestTimePtr(req.ValidEndAt), ValidDays: req.ValidDays,
		ApplicableType: req.ApplicableType,
		Status:         couponStatus(req.Status),
	}
	if err := h.repo.Create(c.Request.Context(), coupon); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, coupon)
}

func (h *CouponHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, apperr.ErrParamInvalid)
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
	var req couponReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	receiveLimitType := defaultCouponReceiveLimitType(req.ReceiveLimitType)
	fields := map[string]interface{}{
		"name":               req.Name,
		"type":               req.Type,
		"threshold_amount":   req.ThresholdAmount,
		"discount_value":     req.DiscountValue,
		"max_discount":       req.MaxDiscount,
		"receive_limit_type": receiveLimitType,
		"per_limit":          req.PerLimit,
		"use_limit":          req.UseLimit,
		"receive_start_at":   requestTimePtr(req.ReceiveStartAt),
		"receive_end_at":     requestTimePtr(req.ReceiveEndAt),
		"valid_start_at":     requestTimePtr(req.ValidStartAt),
		"valid_end_at":       requestTimePtr(req.ValidEndAt),
		"valid_days":         req.ValidDays,
		"applicable_type":    defaultStr(req.ApplicableType, "all"),
		"status":             couponStatus(req.Status),
	}
	// total_count 调整时同步 remain_count（已发放部分不退还）；不限总发放时库存字段归零。
	if receiveLimitType == model.CouponReceiveLimitUnlimited {
		fields["total_count"] = 0
		fields["remain_count"] = 0
	} else if req.TotalCount != exist.TotalCount || exist.ReceiveLimitType == model.CouponReceiveLimitUnlimited {
		delta := req.TotalCount - exist.TotalCount
		fields["total_count"] = req.TotalCount
		if exist.ReceiveLimitType == model.CouponReceiveLimitUnlimited {
			fields["remain_count"] = req.TotalCount
		} else {
			fields["remain_count"] = max0(exist.RemainCount + delta)
		}
	}
	if err := h.repo.Update(c.Request.Context(), id, fields); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"id": id})
}

func (h *CouponHandler) Delete(c *gin.Context) {
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

func couponStatus(v int8) int8 {
	if v == 1 {
		return 1
	}
	return 0
}

func defaultCouponStatus(v int8) int8 {
	if v == 0 {
		return 1
	}
	return v
}

func defaultStr(v, d string) string {
	if v == "" {
		return d
	}
	return v
}

func max0(n int) int {
	if n < 0 {
		return 0
	}
	return n
}

func defaultCouponReceiveLimitType(v string) string {
	if v == model.CouponReceiveLimitUnlimited {
		return model.CouponReceiveLimitUnlimited
	}
	return model.CouponReceiveLimitLimited
}

func couponCounts(receiveLimitType string, totalCount, remainCount int) (int, int) {
	if receiveLimitType == model.CouponReceiveLimitUnlimited {
		return 0, 0
	}
	return max0(totalCount), max0(remainCount)
}
