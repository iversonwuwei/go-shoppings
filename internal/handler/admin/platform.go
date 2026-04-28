package admin

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"wechat-mall-saas/internal/model"
	"wechat-mall-saas/internal/pkg/response"
	"wechat-mall-saas/internal/repository"
	"wechat-mall-saas/internal/service"
)

type PlatformHandler struct {
	tenant      *service.TenantService
	tenantRepo  *repository.TenantRepo
	planRepo    *repository.PlanRepo
	featureRepo *repository.PlanFeatureRepo
}

func NewPlatformHandler(
	t *service.TenantService,
	tr *repository.TenantRepo,
	pr *repository.PlanRepo,
	fr *repository.PlanFeatureRepo,
) *PlatformHandler {
	return &PlatformHandler{tenant: t, tenantRepo: tr, planRepo: pr, featureRepo: fr}
}

type auditReq struct {
	Approve bool   `json:"approve"`
	Reason  string `json:"reason"`
}

func (h *PlatformHandler) ListTenants(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	keyword := c.Query("keyword")
	var statusPtr *int8
	if s := c.Query("status"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			st := int8(v)
			statusPtr = &st
		}
	}
	rows, total, err := h.tenantRepo.List(c.Request.Context(), statusPtr, keyword, page, size)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"list": rows, "total": total, "page": page, "size": size})
}

func (h *PlatformHandler) AuditTenant(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var body auditReq
	if err := c.ShouldBindJSON(&body); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	if err := h.tenant.Audit(c.Request.Context(), id, body.Approve, body.Reason); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, nil)
}

// ========== 租户运营管理：封禁 / 套餐 / 附加功能 ==========

type updateTenantStatusReq struct {
	Status int8   `json:"status"` // 1=正常 3=封禁
	Reason string `json:"reason"`
}

func (h *PlatformHandler) UpdateTenantStatus(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var body updateTenantStatusReq
	if err := c.ShouldBindJSON(&body); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	if err := h.tenant.SetStatus(c.Request.Context(), id, body.Status, body.Reason); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, nil)
}

type updateTenantPlanReq struct {
	PlanID       uint64       `json:"plan_id"`
	PlanExpireAt *requestTime `json:"plan_expire_at,omitempty"`
}

func (h *PlatformHandler) UpdateTenantPlan(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var body updateTenantPlanReq
	if err := c.ShouldBindJSON(&body); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	if err := h.tenant.SetPlan(c.Request.Context(), id, body.PlanID, requestTimePtr(body.PlanExpireAt)); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, nil)
}

type updateTenantFeaturesReq struct {
	ExtraFeatures []string `json:"extra_features"`
}

func (h *PlatformHandler) UpdateTenantFeatures(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var body updateTenantFeaturesReq
	if err := c.ShouldBindJSON(&body); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	if err := h.tenant.SetExtraFeatures(c.Request.Context(), id, body.ExtraFeatures); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, nil)
}

// 以下 stub 保持接口占位，平台运营后续可扩展
func (h *PlatformHandler) Dashboard(c *gin.Context) {
	response.OK(c, gin.H{"tenants_total": 0, "revenue_total": 0})
}

// ========== 套餐管理 ==========

func (h *PlatformHandler) ListPlans(c *gin.Context) {
	rows, err := h.planRepo.List(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, rows)
}

func (h *PlatformHandler) CreatePlan(c *gin.Context) {
	var body model.Plan
	if err := c.ShouldBindJSON(&body); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	body.ID = 0
	if err := h.planRepo.Create(c.Request.Context(), &body); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, body)
}

func (h *PlatformHandler) UpdatePlan(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	cur, err := h.planRepo.FindByID(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	if cur == nil {
		response.FailCode(c, 40004, "套餐不存在")
		return
	}
	var body model.Plan
	if err := c.ShouldBindJSON(&body); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	body.ID = cur.ID
	body.CreatedAt = cur.CreatedAt
	if err := h.planRepo.Update(c.Request.Context(), &body); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, body)
}

func (h *PlatformHandler) DeletePlan(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.planRepo.Delete(c.Request.Context(), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, nil)
}

// ========== 功能目录管理 ==========

func (h *PlatformHandler) ListFeatures(c *gin.Context) {
	rows, err := h.featureRepo.List(c.Request.Context(), false)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, rows)
}

func (h *PlatformHandler) CreateFeature(c *gin.Context) {
	var body model.PlanFeature
	if err := c.ShouldBindJSON(&body); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	body.ID = 0
	if err := h.featureRepo.Create(c.Request.Context(), &body); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, body)
}

func (h *PlatformHandler) UpdateFeature(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	cur, err := h.featureRepo.FindByID(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	if cur == nil {
		response.FailCode(c, 40004, "功能不存在")
		return
	}
	var body model.PlanFeature
	if err := c.ShouldBindJSON(&body); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	body.ID = cur.ID
	body.CreatedAt = cur.CreatedAt
	if err := h.featureRepo.Update(c.Request.Context(), &body); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, body)
}

func (h *PlatformHandler) DeleteFeature(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.featureRepo.Delete(c.Request.Context(), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, nil)
}
