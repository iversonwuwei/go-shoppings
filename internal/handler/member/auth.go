package member

import (
	"github.com/gin-gonic/gin"

	"wechat-mall-saas/internal/pkg/ctxkeys"
	apperr "wechat-mall-saas/internal/pkg/errors"
	"wechat-mall-saas/internal/pkg/response"
	"wechat-mall-saas/internal/pkg/wxapp"
	"wechat-mall-saas/internal/repository"
	"wechat-mall-saas/internal/service"
)

type AuthHandler struct {
	auth    *service.AuthService
	tenants *repository.TenantRepo
}

func NewAuthHandler(a *service.AuthService, t *repository.TenantRepo) *AuthHandler {
	return &AuthHandler{auth: a, tenants: t}
}

type wxLoginReq struct {
	Code     string `json:"code" binding:"required"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Gender   int8   `json:"gender"`
}

type devLoginReq struct {
	Phone    string `json:"phone" binding:"required"`
	Nickname string `json:"nickname"`
}

func (h *AuthHandler) DevLogin(c *gin.Context) {
	var req devLoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	res, err := h.auth.MemberDevLogin(c.Request.Context(), req.Phone, req.Nickname)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, res)
}

func (h *AuthHandler) LoginByWechat(c *gin.Context) {
	t := ctxkeys.GetTenant(c.Request.Context())
	if t == nil {
		response.Fail(c, apperr.ErrTenantRequired)
		return
	}
	tenant, err := h.tenants.FindByID(c.Request.Context(), t.ID)
	if err != nil || tenant == nil {
		response.Fail(c, apperr.ErrTenantInvalid)
		return
	}
	var req wxLoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	input := service.WechatMemberLoginInput{
		Code:     req.Code,
		Nickname: req.Nickname,
		Avatar:   req.Avatar,
		Gender:   req.Gender,
	}
	if tenant.WechatAppID == "" || tenant.WechatSecret == "" {
		if !h.auth.AllowMemberDevLogin() {
			response.Fail(c, apperr.New(20010, "租户未配置微信小程序"))
			return
		}
		res, err := h.auth.MemberDevLoginByWechat(c.Request.Context(), input)
		if err != nil {
			response.Fail(c, err)
			return
		}
		response.OK(c, res)
		return
	}
	wx := wxapp.NewClient(tenant.WechatAppID, tenant.WechatSecret)
	res, err := h.auth.MemberLoginByWechat(c.Request.Context(), wx, input)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, res)
}

type bindPhoneReq struct {
	EncryptedData string `json:"encryptedData" binding:"required"`
	IV            string `json:"iv" binding:"required"`
}

func (h *AuthHandler) BindPhone(c *gin.Context) {
	m := ctxkeys.GetMember(c.Request.Context())
	if m == nil {
		response.Fail(c, apperr.ErrUnauthorized)
		return
	}
	var req bindPhoneReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	phone, err := h.auth.BindPhoneByWechat(c.Request.Context(), m.ID, req.EncryptedData, req.IV)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"phone": phone})
}
