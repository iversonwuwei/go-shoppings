package admin

import (
	"github.com/gin-gonic/gin"

	"wechat-mall-saas/internal/pkg/response"
	"wechat-mall-saas/internal/repository"
)

// PlatformGlobalSettingsHandler 平台全局设置（平台名 / Logo / 平台微信支付商户号 等）
type PlatformGlobalSettingsHandler struct {
	repo *repository.PlatformSettingsRepo
}

func NewPlatformGlobalSettingsHandler(r *repository.PlatformSettingsRepo) *PlatformGlobalSettingsHandler {
	return &PlatformGlobalSettingsHandler{repo: r}
}

func (h *PlatformGlobalSettingsHandler) Get(c *gin.Context) {
	ps, err := h.repo.Get(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, ps)
}

type platformGlobalSettingsBody struct {
	PlatformName     *string `json:"platform_name"`
	PlatformLogo     *string `json:"platform_logo"`
	SupportPhone     *string `json:"support_phone"`
	SupportEmail     *string `json:"support_email"`
	WxpayAppID       *string `json:"wxpay_app_id"`
	WxpayMchID       *string `json:"wxpay_mch_id"`
	WxpayAPIv3Key    *string `json:"wxpay_apiv3_key"`
	WxpayCertSerial  *string `json:"wxpay_cert_serial"`
	WxpayNotifyURL   *string `json:"wxpay_notify_url"`
	SpAppID          *string `json:"sp_appid"`
	SpMchID          *string `json:"sp_mchid"`
	SpAPIv3Key       *string `json:"sp_apiv3_key"`
	SpCertSerial     *string `json:"sp_cert_serial"`
	PartnerNotifyURL *string `json:"partner_notify_url"`
}

func (h *PlatformGlobalSettingsHandler) Update(c *gin.Context) {
	var b platformGlobalSettingsBody
	if err := c.ShouldBindJSON(&b); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	fields := map[string]interface{}{}
	if b.PlatformName != nil {
		fields["platform_name"] = *b.PlatformName
	}
	if b.PlatformLogo != nil {
		fields["platform_logo"] = *b.PlatformLogo
	}
	if b.SupportPhone != nil {
		fields["support_phone"] = *b.SupportPhone
	}
	if b.SupportEmail != nil {
		fields["support_email"] = *b.SupportEmail
	}
	if b.WxpayAppID != nil {
		fields["wxpay_app_id"] = *b.WxpayAppID
	}
	if b.WxpayMchID != nil {
		fields["wxpay_mch_id"] = *b.WxpayMchID
	}
	if b.WxpayAPIv3Key != nil {
		fields["wxpay_apiv3_key"] = *b.WxpayAPIv3Key
	}
	if b.WxpayCertSerial != nil {
		fields["wxpay_cert_serial"] = *b.WxpayCertSerial
	}
	if b.WxpayNotifyURL != nil {
		fields["wxpay_notify_url"] = *b.WxpayNotifyURL
	}
	if b.SpAppID != nil {
		fields["sp_appid"] = *b.SpAppID
	}
	if b.SpMchID != nil {
		fields["sp_mchid"] = *b.SpMchID
	}
	if b.SpAPIv3Key != nil {
		fields["sp_apiv3_key"] = *b.SpAPIv3Key
	}
	if b.SpCertSerial != nil {
		fields["sp_cert_serial"] = *b.SpCertSerial
	}
	if b.PartnerNotifyURL != nil {
		fields["partner_notify_url"] = *b.PartnerNotifyURL
	}
	if err := h.repo.Upsert(c.Request.Context(), fields); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, nil)
}
