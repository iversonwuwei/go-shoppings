package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"wechat-mall-saas/internal/model"
	"wechat-mall-saas/internal/pkg/ctxkeys"
	"wechat-mall-saas/internal/pkg/response"
	"wechat-mall-saas/internal/repository"
)

type PointsHandler struct {
	repo *repository.PointsSettingsRepo
}

func NewPointsHandler(r *repository.PointsSettingsRepo) *PointsHandler {
	return &PointsHandler{repo: r}
}

type pointsSettingsReq struct {
	Enabled    int8            `json:"enabled"`
	EarnRate   decimal.Decimal `json:"earn_rate"`
	MinAmount  decimal.Decimal `json:"min_amount"`
	RedeemRate int             `json:"redeem_rate"`
	Remark     string          `json:"remark"`
}

// Get 获取当前租户的积分规则，不存在则返回默认值。
func (h *PointsHandler) Get(c *gin.Context) {
	ctx := c.Request.Context()
	ps, err := h.repo.Get(ctx)
	if err != nil {
		response.Fail(c, err)
		return
	}
	if ps == nil {
		ps = &model.PointsSetting{
			TenantID:   ctxkeys.GetTenant(ctx).ID,
			Enabled:    1,
			EarnRate:   decimal.NewFromInt(1),
			MinAmount:  decimal.Zero,
			RedeemRate: 100,
		}
	}
	response.OK(c, ps)
}

// Update 保存积分规则（upsert）。
func (h *PointsHandler) Update(c *gin.Context) {
	var req pointsSettingsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	if req.RedeemRate <= 0 {
		req.RedeemRate = 100
	}
	ps := &model.PointsSetting{
		Enabled:    defaultCouponStatus(req.Enabled), // 复用：0→1
		EarnRate:   req.EarnRate,
		MinAmount:  req.MinAmount,
		RedeemRate: req.RedeemRate,
		Remark:     req.Remark,
	}
	if err := h.repo.Upsert(c.Request.Context(), ps); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, ps)
}
