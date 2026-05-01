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
	PlatformName               *string `json:"platform_name"`
	PlatformLogo               *string `json:"platform_logo"`
	SupportPhone               *string `json:"support_phone"`
	SupportEmail               *string `json:"support_email"`
	PrivacyPolicyTitle         *string `json:"privacy_policy_title"`
	PrivacyPolicyEffectiveDate *string `json:"privacy_policy_effective_date"`
	PrivacyPolicyContent       *string `json:"privacy_policy_content"`
	PrivacyPolicyContactPhone  *string `json:"privacy_policy_contact_phone"`
	PrivacyPolicyContactEmail  *string `json:"privacy_policy_contact_email"`
	WxpayAppID                 *string `json:"wxpay_app_id"`
	WxpayMchID                 *string `json:"wxpay_mch_id"`
	WxpayAPIv3Key              *string `json:"wxpay_apiv3_key"`
	WxpayCertSerial            *string `json:"wxpay_cert_serial"`
	WxpayNotifyURL             *string `json:"wxpay_notify_url"`
	SpAppID                    *string `json:"sp_appid"`
	SpMchID                    *string `json:"sp_mchid"`
	SpAPIv3Key                 *string `json:"sp_apiv3_key"`
	SpCertSerial               *string `json:"sp_cert_serial"`
	PartnerNotifyURL           *string `json:"partner_notify_url"`
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
	if b.PrivacyPolicyTitle != nil {
		fields["privacy_policy_title"] = *b.PrivacyPolicyTitle
	}
	if b.PrivacyPolicyEffectiveDate != nil {
		fields["privacy_policy_effective_date"] = *b.PrivacyPolicyEffectiveDate
	}
	if b.PrivacyPolicyContent != nil {
		fields["privacy_policy_content"] = *b.PrivacyPolicyContent
	}
	if b.PrivacyPolicyContactPhone != nil {
		fields["privacy_policy_contact_phone"] = *b.PrivacyPolicyContactPhone
	}
	if b.PrivacyPolicyContactEmail != nil {
		fields["privacy_policy_contact_email"] = *b.PrivacyPolicyContactEmail
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

type platformPrivacyPolicyBody struct {
	PrivacyPolicyTitle         string `json:"privacy_policy_title"`
	PrivacyPolicyEffectiveDate string `json:"privacy_policy_effective_date"`
	PrivacyPolicyContent       string `json:"privacy_policy_content"`
	PrivacyPolicyContactPhone  string `json:"privacy_policy_contact_phone"`
	PrivacyPolicyContactEmail  string `json:"privacy_policy_contact_email"`
}

func privacyPolicyPayload(title, effectiveDate, content, contactPhone, contactEmail string) gin.H {
	return gin.H{
		"privacy_policy_title":          title,
		"privacy_policy_effective_date": effectiveDate,
		"privacy_policy_content":        content,
		"privacy_policy_contact_phone":  contactPhone,
		"privacy_policy_contact_email":  contactEmail,
	}
}

func (h *PlatformGlobalSettingsHandler) GetPrivacy(c *gin.Context) {
	ps, err := h.repo.Get(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, privacyPolicyPayload(
		ps.PrivacyPolicyTitle,
		ps.PrivacyPolicyEffectiveDate,
		ps.PrivacyPolicyContent,
		ps.PrivacyPolicyContactPhone,
		ps.PrivacyPolicyContactEmail,
	))
}

func (h *PlatformGlobalSettingsHandler) UpdatePrivacy(c *gin.Context) {
	var b platformPrivacyPolicyBody
	if err := c.ShouldBindJSON(&b); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	fields := map[string]interface{}{
		"privacy_policy_title":          b.PrivacyPolicyTitle,
		"privacy_policy_effective_date": b.PrivacyPolicyEffectiveDate,
		"privacy_policy_content":        b.PrivacyPolicyContent,
		"privacy_policy_contact_phone":  b.PrivacyPolicyContactPhone,
		"privacy_policy_contact_email":  b.PrivacyPolicyContactEmail,
	}
	if err := h.repo.Upsert(c.Request.Context(), fields); err != nil {
		response.Fail(c, err)
		return
	}
	ps, err := h.repo.Get(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, privacyPolicyPayload(
		ps.PrivacyPolicyTitle,
		ps.PrivacyPolicyEffectiveDate,
		ps.PrivacyPolicyContent,
		ps.PrivacyPolicyContactPhone,
		ps.PrivacyPolicyContactEmail,
	))
}

func defaultPrivacyPolicyContent() string {
	return "1. 我们如何收集和使用个人信息\n" +
		"为保障小程序基础服务，我们会在你使用注册登录、下单支付、售后服务、地址管理等功能时，按最小必要原则收集对应信息，并仅用于实现该功能。\n\n" +
		"2. 我们如何共享、转让、公开披露个人信息\n" +
		"除法律法规要求或经你单独同意外，我们不会向无关第三方共享你的个人信息。\n\n" +
		"3. 我们如何存储个人信息\n" +
		"我们仅在实现服务目的所必需期限内保存你的个人信息，并采取访问控制、传输加密等安全措施。\n\n" +
		"4. 你如何管理个人信息\n" +
		"你可以通过小程序中的个人中心功能访问、更正、删除部分信息，或通过下方联系方式申请处理。\n\n" +
		"5. 未成年人保护\n" +
		"若你是未满 14 周岁的未成年人，请在监护人指导下使用本服务。\n\n" +
		"6. 协议更新\n" +
		"当隐私协议发生重大变更时，我们将通过页面公告等方式提示你。"
}

func (h *PlatformGlobalSettingsHandler) PublicPrivacy(c *gin.Context) {
	ps, err := h.repo.Get(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	title := ps.PrivacyPolicyTitle
	if title == "" {
		title = "微信小程序隐私保护指引"
	}
	effectiveDate := ps.PrivacyPolicyEffectiveDate
	if effectiveDate == "" {
		effectiveDate = "2026-05-01"
	}
	content := ps.PrivacyPolicyContent
	if content == "" {
		content = defaultPrivacyPolicyContent()
	}
	response.OK(c, gin.H{
		"title":          title,
		"effective_date": effectiveDate,
		"content":        content,
		"contact_phone":  ps.PrivacyPolicyContactPhone,
		"contact_email":  ps.PrivacyPolicyContactEmail,
		"updated_at":     ps.UpdatedAt,
	})
}
