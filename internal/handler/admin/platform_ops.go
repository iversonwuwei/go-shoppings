package admin

import (
	"crypto/rand"
	"encoding/hex"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"wechat-mall-saas/internal/model"
	apperr "wechat-mall-saas/internal/pkg/errors"
	"wechat-mall-saas/internal/pkg/response"
	"wechat-mall-saas/internal/repository"
)

// ==================== 平台 SMS ====================

type PlatformSmsHandler struct {
	repo *repository.SmsRepo
}

func NewPlatformSmsHandler(r *repository.SmsRepo) *PlatformSmsHandler {
	return &PlatformSmsHandler{repo: r}
}

type platformSmsSettingsReq struct {
	Enabled      int8   `json:"enabled"`
	Provider     string `json:"provider"`
	AccessKey    string `json:"access_key"`
	AccessSecret string `json:"access_secret"`
	SignName     string `json:"sign_name"`
}

func (h *PlatformSmsHandler) GetSettings(c *gin.Context) {
	ctx := c.Request.Context()
	s, err := h.repo.PlatformGetGlobalSettings(ctx)
	if err != nil {
		response.Fail(c, err)
		return
	}
	if s == nil {
		s = &model.SmsSetting{TenantID: 0, Provider: "aliyun"}
	}
	if s.AccessSecret != "" {
		s.AccessSecret = "********"
	}
	response.OK(c, s)
}

func (h *PlatformSmsHandler) UpdateSettings(c *gin.Context) {
	var req platformSmsSettingsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	ctx := c.Request.Context()
	old, _ := h.repo.PlatformGetGlobalSettings(ctx)
	if req.AccessSecret == "********" {
		if old != nil {
			req.AccessSecret = old.AccessSecret
		} else {
			req.AccessSecret = ""
		}
	}
	if req.Enabled == 1 && (req.AccessKey == "" || req.AccessSecret == "") {
		response.FailCode(c, 20001, "请填写阿里云 AccessKey 和 AccessSecret")
		return
	}
	req.SignName = strings.TrimSpace(req.SignName)
	if req.Enabled == 1 && req.SignName == "" {
		response.FailCode(c, 20001, "请填写阿里云短信签名名称")
		return
	}
	provider := req.Provider
	if provider == "" {
		provider = "aliyun"
	}
	s := &model.SmsSetting{
		Enabled:      req.Enabled,
		Provider:     provider,
		AccessKey:    req.AccessKey,
		AccessSecret: req.AccessSecret,
		SignName:     req.SignName,
	}
	if old != nil {
		s.Region = old.Region
		s.Remark = old.Remark
	}
	if err := h.repo.PlatformUpsertGlobalSettings(ctx, s); err != nil {
		response.Fail(c, err)
		return
	}
	s.AccessSecret = "********"
	response.OK(c, s)
}

func (h *PlatformSmsHandler) ListTemplates(c *gin.Context) {
	rows, err := h.repo.PlatformListTemplates(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, rows)
}

type platformSmsTemplateReq struct {
	Code       string `json:"code" binding:"required,max=64"`
	Name       string `json:"name" binding:"required,max=100"`
	TemplateID string `json:"template_id"`
	Content    string `json:"content"`
	Enabled    int8   `json:"enabled"`
}

func platformSmsTemplateName(code string) (string, bool) {
	switch code {
	case "apply":
		return "入驻申请验证码", true
	case "login":
		return "平台短信登录", true
	case "reset_password":
		return "找回密码验证码", true
	default:
		return "", false
	}
}

func (h *PlatformSmsHandler) CreateTemplate(c *gin.Context) {
	var req platformSmsTemplateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	name, ok := platformSmsTemplateName(req.Code)
	if !ok {
		response.FailCode(c, 20001, "不支持的短信业务用途")
		return
	}
	if req.Name != "" {
		name = req.Name
	}
	t := &model.SmsTemplate{
		Code:       req.Code,
		Name:       name,
		TemplateID: req.TemplateID,
		Content:    req.Content,
		Enabled:    couponStatus(req.Enabled),
	}
	if err := h.repo.PlatformCreateTemplate(c.Request.Context(), t); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, t)
}

// 平台更新模板状态 / 内容（审核）
func (h *PlatformSmsHandler) UpdateTemplate(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.Fail(c, apperr.ErrParamInvalid)
		return
	}
	var req platformSmsTemplateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	name, ok := platformSmsTemplateName(req.Code)
	if !ok {
		response.FailCode(c, 20001, "不支持的短信业务用途")
		return
	}
	if req.Name != "" {
		name = req.Name
	}
	fields := map[string]interface{}{
		"code":        req.Code,
		"name":        name,
		"template_id": req.TemplateID,
		"content":     req.Content,
		"enabled":     couponStatus(req.Enabled),
	}
	if err := h.repo.PlatformUpdateTemplateAny(c.Request.Context(), id, fields); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"id": id})
}

func (h *PlatformSmsHandler) DeleteTemplate(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.Fail(c, apperr.ErrParamInvalid)
		return
	}
	if err := h.repo.PlatformDeleteTemplateAny(c.Request.Context(), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, nil)
}

func (h *PlatformSmsHandler) ListLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	phone := c.Query("phone")
	rows, total, err := h.repo.PlatformListLogs(c.Request.Context(), phone, page, size)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"list": rows, "total": total, "page": page, "size": size})
}

// ==================== 平台 API Access ====================

type PlatformApiAccessHandler struct {
	repo *repository.ApiTokenRepo
}

func NewPlatformApiAccessHandler(r *repository.ApiTokenRepo) *PlatformApiAccessHandler {
	return &PlatformApiAccessHandler{repo: r}
}

type platformApiTokenReq struct {
	TenantID    uint64 `json:"tenant_id"`
	Name        string `json:"name" binding:"required,max=100"`
	Scopes      string `json:"scopes"`
	IPWhitelist string `json:"ip_whitelist"`
	Status      int8   `json:"status"`
	ExpiresAt   string `json:"expires_at"`
}

func platRandHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func platParseExpires(s string) *time.Time {
	if s == "" {
		return nil
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return &t
	}
	if t, err := time.Parse("2006-01-02 15:04:05", s); err == nil {
		return &t
	}
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return &t
	}
	return nil
}

func platMaskSecret(t *model.ApiToken) {
	if t == nil || t.AppSecret == "" {
		return
	}
	s := t.AppSecret
	if len(s) > 8 {
		t.AppSecret = s[:4] + "****" + s[len(s)-4:]
	} else {
		t.AppSecret = "****"
	}
}

func (h *PlatformApiAccessHandler) List(c *gin.Context) {
	tid, _ := strconv.ParseUint(c.Query("tenant_id"), 10, 64)
	rows, err := h.repo.PlatformList(c.Request.Context(), tid)
	if err != nil {
		response.Fail(c, err)
		return
	}
	for i := range rows {
		platMaskSecret(&rows[i])
	}
	response.OK(c, rows)
}

func (h *PlatformApiAccessHandler) Create(c *gin.Context) {
	var req platformApiTokenReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	if req.TenantID == 0 {
		response.FailCode(c, 20001, "tenant_id is required")
		return
	}
	t := &model.ApiToken{
		Name:        req.Name,
		AppKey:      "ak_" + platRandHex(12),
		AppSecret:   platRandHex(32),
		Scopes:      req.Scopes,
		IPWhitelist: req.IPWhitelist,
		Status:      defaultCouponStatus(req.Status),
		ExpiresAt:   platParseExpires(req.ExpiresAt),
	}
	if err := h.repo.PlatformCreate(c.Request.Context(), req.TenantID, t); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, t)
}

func (h *PlatformApiAccessHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.Fail(c, apperr.ErrParamInvalid)
		return
	}
	var req platformApiTokenReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	fields := map[string]interface{}{
		"name":         req.Name,
		"scopes":       req.Scopes,
		"ip_whitelist": req.IPWhitelist,
		"status":       defaultCouponStatus(req.Status),
	}
	if exp := platParseExpires(req.ExpiresAt); exp != nil {
		fields["expires_at"] = exp
	} else if req.ExpiresAt == "" {
		fields["expires_at"] = nil
	}
	if err := h.repo.PlatformUpdate(c.Request.Context(), id, fields); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"id": id})
}

func (h *PlatformApiAccessHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.Fail(c, apperr.ErrParamInvalid)
		return
	}
	if err := h.repo.PlatformDelete(c.Request.Context(), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, nil)
}

func (h *PlatformApiAccessHandler) Regenerate(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.Fail(c, apperr.ErrParamInvalid)
		return
	}
	secret := platRandHex(32)
	if err := h.repo.PlatformUpdate(c.Request.Context(), id, map[string]interface{}{
		"app_secret": secret,
	}); err != nil {
		response.Fail(c, err)
		return
	}
	t, err := h.repo.PlatformFind(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, t)
}

func (h *PlatformApiAccessHandler) ListLogs(c *gin.Context) {
	tid, _ := strconv.ParseUint(c.Query("tenant_id"), 10, 64)
	tokenID, _ := strconv.ParseUint(c.Query("token_id"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	rows, total, err := h.repo.PlatformListLogs(c.Request.Context(), tid, tokenID, page, size)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"list": rows, "total": total, "page": page, "size": size})
}

// ==================== 平台 自定义域名 ====================

type PlatformDomainHandler struct {
	repo *repository.SiteConfigRepo
}

func NewPlatformDomainHandler(r *repository.SiteConfigRepo) *PlatformDomainHandler {
	return &PlatformDomainHandler{repo: r}
}

func (h *PlatformDomainHandler) List(c *gin.Context) {
	rows, err := h.repo.PlatformListWithDomain(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, rows)
}

// Verify 审核通过：domain_verified=1，ssl_status=active
func (h *PlatformDomainHandler) Verify(c *gin.Context) {
	tid, _ := strconv.ParseUint(c.Param("tid"), 10, 64)
	if tid == 0 {
		response.Fail(c, apperr.ErrParamInvalid)
		return
	}
	fields := map[string]interface{}{
		"domain_verified": 1,
		"ssl_status":      "active",
	}
	if err := h.repo.PlatformUpdateByTenantID(c.Request.Context(), tid, fields); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, nil)
}

// Reject 审核拒绝：清空域名
func (h *PlatformDomainHandler) Reject(c *gin.Context) {
	tid, _ := strconv.ParseUint(c.Param("tid"), 10, 64)
	if tid == 0 {
		response.Fail(c, apperr.ErrParamInvalid)
		return
	}
	fields := map[string]interface{}{
		"custom_domain":   "",
		"domain_verified": 0,
		"ssl_status":      "none",
	}
	if err := h.repo.PlatformUpdateByTenantID(c.Request.Context(), tid, fields); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, nil)
}

// ==================== 平台 私有部署 ====================

type PlatformDeploymentHandler struct {
	repo *repository.SiteConfigRepo
}

func NewPlatformDeploymentHandler(r *repository.SiteConfigRepo) *PlatformDeploymentHandler {
	return &PlatformDeploymentHandler{repo: r}
}

func (h *PlatformDeploymentHandler) List(c *gin.Context) {
	mode := c.Query("mode")
	rows, err := h.repo.PlatformListDeployments(c.Request.Context(), mode)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, rows)
}

type platformDeploymentReq struct {
	TenantID        uint64 `json:"tenant_id" binding:"required"`
	DeploymentMode  string `json:"deployment_mode"`
	PrivateEndpoint string `json:"private_endpoint"`
	PrivateNotes    string `json:"private_notes"`
}

// Update 平台运维更新租户的部署模式 / 端点 / 备注
func (h *PlatformDeploymentHandler) Update(c *gin.Context) {
	var req platformDeploymentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	mode := req.DeploymentMode
	if mode != "private" && mode != "shared" {
		mode = "shared"
	}
	fields := map[string]interface{}{
		"deployment_mode":  mode,
		"private_endpoint": req.PrivateEndpoint,
		"private_notes":    req.PrivateNotes,
	}
	// 若该租户还没有站点配置行，则创建
	cur, err := h.repo.PlatformFindByTenantID(c.Request.Context(), req.TenantID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	if cur == nil {
		s := &model.TenantSiteConfig{
			TenantID:        req.TenantID,
			DeploymentMode:  mode,
			PrivateEndpoint: req.PrivateEndpoint,
			PrivateNotes:    req.PrivateNotes,
			PrimaryColor:    "#409EFF",
			SSLStatus:       "none",
			UpdatedAt:       time.Now(),
		}
		if err := h.repo.PlatformInsert(c.Request.Context(), s); err != nil {
			response.Fail(c, err)
			return
		}
		response.OK(c, nil)
		return
	}
	if err := h.repo.PlatformUpdateByTenantID(c.Request.Context(), req.TenantID, fields); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, nil)
}

// ==================== 平台 商城快捷入口 ====================

type PlatformStorefrontHandler struct {
	repo *repository.SiteConfigRepo
}

func NewPlatformStorefrontHandler(r *repository.SiteConfigRepo) *PlatformStorefrontHandler {
	return &PlatformStorefrontHandler{repo: r}
}

func (h *PlatformStorefrontHandler) GetQuickEntries(c *gin.Context) {
	tid, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if tid == 0 {
		response.Fail(c, apperr.ErrParamInvalid)
		return
	}
	cur, err := h.repo.PlatformFindByTenantID(c.Request.Context(), tid)
	if err != nil {
		response.Fail(c, err)
		return
	}
	cur = normalizeSiteConfig(tid, cur)
	response.OK(c, gin.H{"storefront_quick_entries": decodeQuickEntries(cur.StorefrontQuickEntries)})
}

type platformQuickEntriesReq struct {
	StorefrontQuickEntries []storefrontQuickEntry `json:"storefront_quick_entries"`
}

func (h *PlatformStorefrontHandler) UpdateQuickEntries(c *gin.Context) {
	tid, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if tid == 0 {
		response.Fail(c, apperr.ErrParamInvalid)
		return
	}
	var req platformQuickEntriesReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	ctx := c.Request.Context()
	cur, err := h.repo.PlatformFindByTenantID(ctx, tid)
	if err != nil {
		response.Fail(c, err)
		return
	}
	if cur == nil {
		cur = defaultSiteConfig(tid)
	}
	cur.StorefrontQuickEntries = encodeJSON(req.StorefrontQuickEntries)
	if cur.TenantID == 0 {
		cur.TenantID = tid
	}
	if existing, err := h.repo.PlatformFindByTenantID(ctx, tid); err != nil {
		response.Fail(c, err)
		return
	} else if existing == nil {
		cur.UpdatedAt = time.Now()
		if err := h.repo.PlatformInsert(ctx, cur); err != nil {
			response.Fail(c, err)
			return
		}
	} else if err := h.repo.PlatformUpdateByTenantID(ctx, tid, map[string]interface{}{
		"storefront_quick_entries": cur.StorefrontQuickEntries,
	}); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"storefront_quick_entries": decodeQuickEntries(cur.StorefrontQuickEntries)})
}
