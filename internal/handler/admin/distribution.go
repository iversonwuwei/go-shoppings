package admin

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"wechat-mall-saas/internal/model"
	"wechat-mall-saas/internal/pkg/ctxkeys"
	apperr "wechat-mall-saas/internal/pkg/errors"
	"wechat-mall-saas/internal/pkg/response"
	"wechat-mall-saas/internal/repository"
)

type DistributionHandler struct {
	repo *repository.DistributionRepo
}

func NewDistributionHandler(r *repository.DistributionRepo) *DistributionHandler {
	return &DistributionHandler{repo: r}
}

type distributionSettingsReq struct {
	Enabled     int8            `json:"enabled"`
	Level1Rate  decimal.Decimal `json:"level1_rate"`
	Level2Rate  decimal.Decimal `json:"level2_rate"`
	MinWithdraw decimal.Decimal `json:"min_withdraw"`
	AutoBecome  int8            `json:"auto_become"`
	Remark      string          `json:"remark"`
}

func (h *DistributionHandler) GetSettings(c *gin.Context) {
	ctx := c.Request.Context()
	s, err := h.repo.GetSettings(ctx)
	if err != nil {
		response.Fail(c, err)
		return
	}
	if s == nil {
		s = &model.DistributionSetting{
			TenantID:    ctxkeys.GetTenant(ctx).ID,
			Enabled:     1,
			Level1Rate:  decimal.NewFromFloat(0.10),
			Level2Rate:  decimal.NewFromFloat(0.05),
			MinWithdraw: decimal.NewFromInt(10),
			AutoBecome:  0,
		}
	}
	response.OK(c, s)
}

func (h *DistributionHandler) UpdateSettings(c *gin.Context) {
	var req distributionSettingsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	s := &model.DistributionSetting{
		Enabled:     defaultCouponStatus(req.Enabled),
		Level1Rate:  req.Level1Rate,
		Level2Rate:  req.Level2Rate,
		MinWithdraw: req.MinWithdraw,
		AutoBecome:  req.AutoBecome,
		Remark:      req.Remark,
	}
	if err := h.repo.UpsertSettings(c.Request.Context(), s); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, s)
}

// ListDistributors ?status=-1|0|1|2
func (h *DistributionHandler) ListDistributors(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	var status int8 = -1
	if v := c.Query("status"); v != "" {
		if x, err := strconv.Atoi(v); err == nil {
			status = int8(x)
		}
	}
	rows, total, err := h.repo.ListDistributors(c.Request.Context(), status, page, size)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"list": rows, "total": total, "page": page, "size": size})
}

type auditDistributorReq struct {
	Status int8 `json:"status" binding:"required"` // 1 正常 2 冻结
}

func (h *DistributionHandler) AuditDistributor(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.Fail(c, apperr.ErrParamInvalid)
		return
	}
	var req auditDistributorReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	fields := map[string]interface{}{"status": req.Status}
	if req.Status == 1 {
		fields["approved_at"] = time.Now()
	}
	if err := h.repo.UpdateDistributor(c.Request.Context(), id, fields); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"id": id})
}

func (h *DistributionHandler) ListCommissions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	distributorID, _ := strconv.ParseUint(c.Query("distributor_id"), 10, 64)
	rows, total, err := h.repo.ListCommissions(c.Request.Context(), distributorID, page, size)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"list": rows, "total": total, "page": page, "size": size})
}
