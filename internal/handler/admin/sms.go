package admin

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"wechat-mall-saas/internal/model"
	"wechat-mall-saas/internal/pkg/ctxkeys"
	apperr "wechat-mall-saas/internal/pkg/errors"
	"wechat-mall-saas/internal/pkg/response"
	"wechat-mall-saas/internal/repository"
)

type SmsHandler struct {
	repo *repository.SmsRepo
}

func NewSmsHandler(r *repository.SmsRepo) *SmsHandler { return &SmsHandler{repo: r} }

type smsSettingsReq struct {
	Enabled      int8   `json:"enabled"`
	Provider     string `json:"provider"`
	AccessKey    string `json:"access_key"`
	AccessSecret string `json:"access_secret"`
	SignName     string `json:"sign_name"`
	Region       string `json:"region"`
	Remark       string `json:"remark"`
}

func (h *SmsHandler) GetSettings(c *gin.Context) {
	ctx := c.Request.Context()
	s, err := h.repo.GetSettings(ctx)
	if err != nil {
		response.Fail(c, err)
		return
	}
	if s == nil {
		s = &model.SmsSetting{
			TenantID: ctxkeys.GetTenant(ctx).ID,
			Enabled:  0,
			Provider: "aliyun",
		}
	}
	// 为安全，回显时隐藏 secret 的一部分
	if s.AccessSecret != "" {
		s.AccessSecret = "********"
	}
	response.OK(c, s)
}

func (h *SmsHandler) UpdateSettings(c *gin.Context) {
	var req smsSettingsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	ctx := c.Request.Context()
	// 若前端回传的是 "********" 代表未修改，保留原值
	if req.AccessSecret == "********" {
		old, _ := h.repo.GetSettings(ctx)
		if old != nil {
			req.AccessSecret = old.AccessSecret
		} else {
			req.AccessSecret = ""
		}
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
		Region:       req.Region,
		Remark:       req.Remark,
	}
	if err := h.repo.UpsertSettings(ctx, s); err != nil {
		response.Fail(c, err)
		return
	}
	// 响应时仍然打码
	s.AccessSecret = "********"
	response.OK(c, s)
}

// --------- Templates ---------

type smsTemplateReq struct {
	Code       string `json:"code" binding:"required,max=64"`
	Name       string `json:"name" binding:"required,max=100"`
	TemplateID string `json:"template_id"`
	Content    string `json:"content"`
	Enabled    int8   `json:"enabled"`
}

func (h *SmsHandler) ListTemplates(c *gin.Context) {
	rows, err := h.repo.ListTemplates(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, rows)
}

func (h *SmsHandler) CreateTemplate(c *gin.Context) {
	var req smsTemplateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	t := &model.SmsTemplate{
		Code:       req.Code,
		Name:       req.Name,
		TemplateID: req.TemplateID,
		Content:    req.Content,
		Enabled:    defaultCouponStatus(req.Enabled),
	}
	if err := h.repo.CreateTemplate(c.Request.Context(), t); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, t)
}

func (h *SmsHandler) UpdateTemplate(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.Fail(c, apperr.ErrParamInvalid)
		return
	}
	var req smsTemplateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	fields := map[string]interface{}{
		"code":        req.Code,
		"name":        req.Name,
		"template_id": req.TemplateID,
		"content":     req.Content,
		"enabled":     defaultCouponStatus(req.Enabled),
	}
	if err := h.repo.UpdateTemplate(c.Request.Context(), id, fields); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"id": id})
}

func (h *SmsHandler) DeleteTemplate(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.Fail(c, apperr.ErrParamInvalid)
		return
	}
	if err := h.repo.DeleteTemplate(c.Request.Context(), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, nil)
}

// --------- Logs ---------

func (h *SmsHandler) ListLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	phone := c.Query("phone")
	rows, total, err := h.repo.ListLogs(c.Request.Context(), phone, page, size)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"list": rows, "total": total, "page": page, "size": size})
}
