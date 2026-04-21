package admin

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"wechat-mall-saas/internal/pkg/response"
	"wechat-mall-saas/internal/service"
)

type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(a *service.AuthService) *AuthHandler { return &AuthHandler{auth: a} }

type loginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	res, err := h.auth.AdminLogin(c.Request.Context(), req.Username, req.Password, c.ClientIP())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, res)
}

// ========== 短信验证码 / 手机号登录 / 忘记密码 ==========

type sendCodeReq struct {
	Phone   string `json:"phone" binding:"required"`
	Purpose string `json:"purpose" binding:"required"` // apply / login / reset_password
}

// SendCode 发送验证码入口（平台 / 租户 / 公共入驻 均复用）
func (h *AuthHandler) SendCode(c *gin.Context) {
	var req sendCodeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	dev, err := h.auth.SendVerifyCode(c.Request.Context(), req.Phone, req.Purpose)
	if err != nil {
		response.Fail(c, err)
		return
	}
	resp := gin.H{"sent": true}
	if dev != "" {
		resp["dev_code"] = dev
	}
	response.OK(c, resp)
}

type loginBySMSReq struct {
	Phone string `json:"phone" binding:"required"`
	Code  string `json:"code" binding:"required"`
}

// LoginBySMS 租户侧手机号 + 验证码登录，tenantID 由 X-Tenant-ID 头指定
func (h *AuthHandler) LoginBySMS(c *gin.Context) {
	var req loginBySMSReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	tid := parseTenantHeader(c)
	res, err := h.auth.AdminLoginBySMS(c.Request.Context(), tid, req.Phone, req.Code, c.ClientIP())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, res)
}

// PlatformLoginBySMS 平台侧手机号 + 验证码登录，tenantID=0
func (h *AuthHandler) PlatformLoginBySMS(c *gin.Context) {
	var req loginBySMSReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	res, err := h.auth.AdminLoginBySMS(c.Request.Context(), 0, req.Phone, req.Code, c.ClientIP())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, res)
}

type resetPwdReq struct {
	Phone       string `json:"phone" binding:"required"`
	Code        string `json:"code" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// ResetPassword 租户侧忘记密码重置，tenantID 由 X-Tenant-ID 头指定
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req resetPwdReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	tid := parseTenantHeader(c)
	if err := h.auth.ResetAdminPassword(c.Request.Context(), tid, req.Phone, req.Code, req.NewPassword); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"reset": true})
}

// PlatformResetPassword 平台侧忘记密码重置，tenantID=0
func (h *AuthHandler) PlatformResetPassword(c *gin.Context) {
	var req resetPwdReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	if err := h.auth.ResetAdminPassword(c.Request.Context(), 0, req.Phone, req.Code, req.NewPassword); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"reset": true})
}

func parseTenantHeader(c *gin.Context) uint64 {
	v := c.GetHeader("X-Tenant-ID")
	if v == "" {
		return 0
	}
	n, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return 0
	}
	return n
}
