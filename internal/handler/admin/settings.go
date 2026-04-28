package admin

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"wechat-mall-saas/internal/pkg/response"
	"wechat-mall-saas/internal/service"
)

// SettingsHandler 商户侧 设置（收款 + 物流只读）
type SettingsHandler struct {
	svc *service.SettingsService
}

func NewSettingsHandler(s *service.SettingsService) *SettingsHandler { return &SettingsHandler{svc: s} }

// ===== 收款配置 =====

func (h *SettingsHandler) ListPayment(c *gin.Context) {
	rows, err := h.svc.ListPaymentConfigs(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, rows)
}

func (h *SettingsHandler) SubmitPayment(c *gin.Context) {
	var in service.PaymentConfigInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	out, err := h.svc.SubmitPaymentConfig(c.Request.Context(), in)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, out)
}

// ===== 物流承运商（只读） =====

func (h *SettingsHandler) ListCarriers(c *gin.Context) {
	rows, err := h.svc.ListCarriersForTenant(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, rows)
}

// QueryTrack 物流轨迹查询
func (h *SettingsHandler) QueryTrack(c *gin.Context) {
	code := c.Query("carrier_code")
	no := c.Query("tracking_no")
	out, err := h.svc.QueryTrack(c.Request.Context(), code, no)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, out)
}

// ===== 平台端：收款审核 + 物流承运商管理 =====

type PlatformSettingsHandler struct {
	svc *service.SettingsService
}

func NewPlatformSettingsHandler(s *service.SettingsService) *PlatformSettingsHandler {
	return &PlatformSettingsHandler{svc: s}
}

func (h *PlatformSettingsHandler) ListPaymentAudit(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	var statusPtr *int8
	if s := c.Query("status"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			st := int8(v)
			statusPtr = &st
		}
	}
	rows, total, err := h.svc.ListPaymentAudit(c.Request.Context(), statusPtr, page, size)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"list": rows, "total": total, "page": page, "size": size})
}

type auditBody struct {
	Approve bool   `json:"approve"`
	Remark  string `json:"remark"`
}

func (h *PlatformSettingsHandler) AuditPayment(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var body auditBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	if err := h.svc.AuditPayment(c.Request.Context(), id, body.Approve, body.Remark); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, nil)
}

// -------- 物流承运商管理 --------

func (h *PlatformSettingsHandler) ListCarriers(c *gin.Context) {
	rows, err := h.svc.ListAllCarriers(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, rows)
}

func (h *PlatformSettingsHandler) CreateCarrier(c *gin.Context) {
	var in service.CarrierInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	out, err := h.svc.CreateCarrier(c.Request.Context(), in)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, out)
}

func (h *PlatformSettingsHandler) UpdateCarrier(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var in service.CarrierInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	out, err := h.svc.UpdateCarrier(c.Request.Context(), id, in)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, out)
}

func (h *PlatformSettingsHandler) ToggleCarrier(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var body struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	if err := h.svc.SetCarrierEnabled(c.Request.Context(), id, body.Enabled); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, nil)
}

func (h *PlatformSettingsHandler) DeleteCarrier(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.DeleteCarrier(c.Request.Context(), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, nil)
}

// -------- 售后原因管理 --------

func (h *PlatformSettingsHandler) ListAfterSaleReasons(c *gin.Context) {
	rows, err := h.svc.ListAfterSaleReasons(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, rows)
}

func (h *PlatformSettingsHandler) CreateAfterSaleReason(c *gin.Context) {
	var in service.AfterSaleReasonInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	out, err := h.svc.CreateAfterSaleReason(c.Request.Context(), in)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, out)
}

func (h *PlatformSettingsHandler) UpdateAfterSaleReason(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var in service.AfterSaleReasonInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	out, err := h.svc.UpdateAfterSaleReason(c.Request.Context(), id, in)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, out)
}

func (h *PlatformSettingsHandler) ToggleAfterSaleReason(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var body struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	if err := h.svc.SetAfterSaleReasonEnabled(c.Request.Context(), id, body.Enabled); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, nil)
}

func (h *PlatformSettingsHandler) DeleteAfterSaleReason(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.DeleteAfterSaleReason(c.Request.Context(), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, nil)
}
