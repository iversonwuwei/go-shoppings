package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"wechat-mall-saas/internal/model"
	"wechat-mall-saas/internal/pkg/ctxkeys"
	"wechat-mall-saas/internal/pkg/response"
	"wechat-mall-saas/internal/repository"
)

type DeliveryHandler struct {
	repo *repository.DeliveryRepo
}

func NewDeliveryHandler(r *repository.DeliveryRepo) *DeliveryHandler {
	return &DeliveryHandler{repo: r}
}

type deliveryReq struct {
	ExpressEnabled    int8            `json:"express_enabled"`
	ExpressFreeAmount decimal.Decimal `json:"express_free_amount"`
	ExpressDefaultFee decimal.Decimal `json:"express_default_fee"`
	CityEnabled       int8            `json:"city_enabled"`
	CityRadiusKm      decimal.Decimal `json:"city_radius_km"`
	CityBaseFee       decimal.Decimal `json:"city_base_fee"`
	CityPerKmFee      decimal.Decimal `json:"city_per_km_fee"`
	CityMinOrder      decimal.Decimal `json:"city_min_order"`
	PickupEnabled     int8            `json:"pickup_enabled"`
	PickupAddress     string          `json:"pickup_address"`
	PickupHours       string          `json:"pickup_hours"`
	PickupPhone       string          `json:"pickup_phone"`
	Remark            string          `json:"remark"`
}

func (h *DeliveryHandler) Get(c *gin.Context) {
	ctx := c.Request.Context()
	s, err := h.repo.Get(ctx)
	if err != nil {
		response.Fail(c, err)
		return
	}
	if s == nil {
		s = &model.DeliverySetting{
			TenantID:       ctxkeys.GetTenant(ctx).ID,
			ExpressEnabled: 1,
			CityRadiusKm:   decimal.NewFromInt(5),
			CityBaseFee:    decimal.NewFromInt(5),
			CityPerKmFee:   decimal.NewFromInt(1),
		}
	}
	response.OK(c, s)
}

func (h *DeliveryHandler) Update(c *gin.Context) {
	var req deliveryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	s := &model.DeliverySetting{
		ExpressEnabled:    req.ExpressEnabled,
		ExpressFreeAmount: req.ExpressFreeAmount,
		ExpressDefaultFee: req.ExpressDefaultFee,
		CityEnabled:       req.CityEnabled,
		CityRadiusKm:      req.CityRadiusKm,
		CityBaseFee:       req.CityBaseFee,
		CityPerKmFee:      req.CityPerKmFee,
		CityMinOrder:      req.CityMinOrder,
		PickupEnabled:     req.PickupEnabled,
		PickupAddress:     req.PickupAddress,
		PickupHours:       req.PickupHours,
		PickupPhone:       req.PickupPhone,
		Remark:            req.Remark,
	}
	if err := h.repo.Upsert(c.Request.Context(), s); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, s)
}
