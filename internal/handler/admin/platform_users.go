package admin

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"wechat-mall-saas/internal/model"
	"wechat-mall-saas/internal/pkg/ctxkeys"
	apperr "wechat-mall-saas/internal/pkg/errors"
	"wechat-mall-saas/internal/pkg/response"
	"wechat-mall-saas/internal/pkg/utils"
	"wechat-mall-saas/internal/repository"
)

// 平台角色常量（也用于前端）
//   super    —— 超级管理员（全部权限）
//   operator —— 运营人员（日常运营：租户审核/套餐/域名等）
//   finance  —— 财务（只读 + 财务相关审核）
//   support  —— 客服（只读）
const (
	PlatformRoleSuper    = "super"
	PlatformRoleOperator = "operator"
	PlatformRoleFinance  = "finance"
	PlatformRoleSupport  = "support"
)

type PlatformUserHandler struct {
	admins *repository.AdminRepo
}

func NewPlatformUserHandler(a *repository.AdminRepo) *PlatformUserHandler {
	return &PlatformUserHandler{admins: a}
}

// List 分页查询平台管理员列表
func (h *PlatformUserHandler) List(c *gin.Context) {
	keyword := c.Query("keyword")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	rows, total, err := h.admins.ListPlatformAdmins(c.Request.Context(), keyword, page, size)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"list": rows, "total": total, "page": page, "page_size": size})
}

type platformUserBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
	RealName string `json:"real_name"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	Status   *int8  `json:"status"`
}

func validRole(r string) bool {
	switch r {
	case PlatformRoleSuper, PlatformRoleOperator, PlatformRoleFinance, PlatformRoleSupport:
		return true
	}
	return false
}

// Create 创建平台管理员
func (h *PlatformUserHandler) Create(c *gin.Context) {
	var b platformUserBody
	if err := c.ShouldBindJSON(&b); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	if b.Username == "" || b.Password == "" {
		response.FailCode(c, 20001, "用户名和密码必填")
		return
	}
	if !validRole(b.Role) {
		response.FailCode(c, 20001, "无效角色")
		return
	}
	if exist, _ := h.admins.FindByUsername(c.Request.Context(), b.Username); exist != nil {
		response.FailCode(c, 20001, "用户名已存在")
		return
	}
	hash, err := utils.HashPassword(b.Password)
	if err != nil {
		response.Fail(c, err)
		return
	}
	status := int8(1)
	if b.Status != nil {
		status = *b.Status
	}
	a := &model.Admin{
		Username: b.Username,
		Password: hash,
		RealName: b.RealName,
		Phone:    b.Phone,
		Email:    b.Email,
		Role:     b.Role,
		TenantID: 0,
		Status:   status,
	}
	if err := h.admins.Create(c.Request.Context(), a); err != nil {
		response.Fail(c, err)
		return
	}
	a.Password = ""
	response.OK(c, a)
}

// Update 更新平台管理员信息（不含密码）
func (h *PlatformUserHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.FailCode(c, 20001, "id 无效")
		return
	}
	var b platformUserBody
	if err := c.ShouldBindJSON(&b); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	fields := map[string]interface{}{}
	if b.RealName != "" {
		fields["real_name"] = b.RealName
	}
	if b.Phone != "" {
		fields["phone"] = b.Phone
	}
	if b.Email != "" {
		fields["email"] = b.Email
	}
	if b.Role != "" {
		if !validRole(b.Role) {
			response.FailCode(c, 20001, "无效角色")
			return
		}
		// 仅 super 可赋予 super 角色，由路由层控制；此处再拦一次防止直接传
		if b.Role == PlatformRoleSuper {
			a := ctxkeys.GetAdmin(c.Request.Context())
			if a == nil || a.Role != PlatformRoleSuper {
				response.FailCode(c, 10003, "仅超级管理员可授予超级管理员角色")
				return
			}
		}
		fields["role"] = b.Role
	}
	if b.Status != nil {
		fields["status"] = *b.Status
	}
	if len(fields) == 0 {
		response.OK(c, nil)
		return
	}
	if err := h.admins.UpdateFields(c.Request.Context(), id, fields); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, nil)
}

type resetPasswordBody struct {
	Password string `json:"password"`
}

// ResetPassword 管理员重置他人密码
func (h *PlatformUserHandler) ResetPassword(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var b resetPasswordBody
	if err := c.ShouldBindJSON(&b); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	if len(b.Password) < 6 {
		response.FailCode(c, 20001, "密码至少 6 位")
		return
	}
	hash, err := utils.HashPassword(b.Password)
	if err != nil {
		response.Fail(c, err)
		return
	}
	if err := h.admins.UpdatePassword(c.Request.Context(), id, hash); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, nil)
}

// Delete 删除平台管理员
func (h *PlatformUserHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.FailCode(c, 20001, "id 无效")
		return
	}
	// 禁止自删
	me := ctxkeys.GetAdmin(c.Request.Context())
	if me != nil && me.ID == id {
		response.FailCode(c, 10003, "不能删除自己")
		return
	}
	// 查询目标：不允许删除仅存的最后一个 super
	target, err := h.admins.FindByID(c.Request.Context(), id)
	if err != nil || target == nil {
		response.Fail(c, apperr.ErrNotFound)
		return
	}
	if err := h.admins.Delete(c.Request.Context(), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, nil)
}

// Me 当前登录管理员信息（含角色），前端用于渲染菜单权限
func (h *PlatformUserHandler) Me(c *gin.Context) {
	me := ctxkeys.GetAdmin(c.Request.Context())
	if me == nil {
		response.Fail(c, apperr.ErrUnauthorized)
		return
	}
	a, err := h.admins.FindByID(c.Request.Context(), me.ID)
	if err != nil || a == nil {
		response.Fail(c, apperr.ErrUnauthorized)
		return
	}
	a.Password = ""
	response.OK(c, a)
}
