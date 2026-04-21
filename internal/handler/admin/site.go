package admin

import (
	"github.com/gin-gonic/gin"

	"wechat-mall-saas/internal/model"
	"wechat-mall-saas/internal/pkg/ctxkeys"
	"wechat-mall-saas/internal/pkg/response"
	"wechat-mall-saas/internal/repository"
)

// SiteConfigHandler 合并处理 custom_domain / white_label / private_deployment
// 各 section 的写入由前端 section 参数标识；具体字段按 section 生效。
type SiteConfigHandler struct {
	repo *repository.SiteConfigRepo
}

func NewSiteConfigHandler(r *repository.SiteConfigRepo) *SiteConfigHandler {
	return &SiteConfigHandler{repo: r}
}

func (h *SiteConfigHandler) Get(c *gin.Context) {
	ctx := c.Request.Context()
	s, err := h.repo.Get(ctx)
	if err != nil {
		response.Fail(c, err)
		return
	}
	if s == nil {
		s = &model.TenantSiteConfig{
			TenantID:       ctxkeys.GetTenant(ctx).ID,
			PrimaryColor:   "#409EFF",
			DeploymentMode: "shared",
			SSLStatus:      "none",
		}
	}
	response.OK(c, s)
}

type domainReq struct {
	CustomDomain string `json:"custom_domain"`
}

// UpdateDomain 自定义域名（由 FeatureCustomDomain gate）
func (h *SiteConfigHandler) UpdateDomain(c *gin.Context) {
	var req domainReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	ctx := c.Request.Context()
	cur, err := h.repo.Get(ctx)
	if err != nil {
		response.Fail(c, err)
		return
	}
	if cur == nil {
		cur = &model.TenantSiteConfig{PrimaryColor: "#409EFF", DeploymentMode: "shared", SSLStatus: "none"}
	}
	// 域名变更后重置校验状态
	if cur.CustomDomain != req.CustomDomain {
		cur.DomainVerified = 0
		cur.SSLStatus = "pending"
	}
	cur.CustomDomain = req.CustomDomain
	if err := h.repo.Upsert(ctx, cur); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, cur)
}

type brandReq struct {
	BrandName         string `json:"brand_name"`
	BrandLogo         string `json:"brand_logo"`
	PrimaryColor      string `json:"primary_color"`
	HidePlatformBrand int8   `json:"hide_platform_brand"`
	FooterText        string `json:"footer_text"`
}

// UpdateBrand 白标品牌（由 FeatureWhiteLabel gate）
func (h *SiteConfigHandler) UpdateBrand(c *gin.Context) {
	var req brandReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	ctx := c.Request.Context()
	cur, err := h.repo.Get(ctx)
	if err != nil {
		response.Fail(c, err)
		return
	}
	if cur == nil {
		cur = &model.TenantSiteConfig{DeploymentMode: "shared", SSLStatus: "none"}
	}
	cur.BrandName = req.BrandName
	cur.BrandLogo = req.BrandLogo
	if req.PrimaryColor != "" {
		cur.PrimaryColor = req.PrimaryColor
	}
	cur.HidePlatformBrand = req.HidePlatformBrand
	cur.FooterText = req.FooterText
	if err := h.repo.Upsert(ctx, cur); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, cur)
}

type deployReq struct {
	DeploymentMode  string `json:"deployment_mode"`
	PrivateEndpoint string `json:"private_endpoint"`
	PrivateNotes    string `json:"private_notes"`
}

// UpdateDeployment 私有部署（由 FeaturePrivateDeployment gate）
func (h *SiteConfigHandler) UpdateDeployment(c *gin.Context) {
	var req deployReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	ctx := c.Request.Context()
	cur, err := h.repo.Get(ctx)
	if err != nil {
		response.Fail(c, err)
		return
	}
	if cur == nil {
		cur = &model.TenantSiteConfig{PrimaryColor: "#409EFF", SSLStatus: "none"}
	}
	mode := req.DeploymentMode
	if mode != "private" {
		mode = "shared"
	}
	cur.DeploymentMode = mode
	cur.PrivateEndpoint = req.PrivateEndpoint
	cur.PrivateNotes = req.PrivateNotes
	if err := h.repo.Upsert(ctx, cur); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, cur)
}
