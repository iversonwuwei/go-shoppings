package admin

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/skip2/go-qrcode"

	"wechat-mall-saas/internal/model"
	"wechat-mall-saas/internal/pkg/ctxkeys"
	apperr "wechat-mall-saas/internal/pkg/errors"
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

type storefrontQuickEntry struct {
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	Path     string `json:"path"`
}

type storefrontServiceCard struct {
	Title string `json:"title"`
	Desc  string `json:"desc"`
}

type storefrontBanner struct {
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	Image    string `json:"image"`
	Path     string `json:"path"`
}

type storefrontPromoCard struct {
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	Tag      string `json:"tag"`
	Path     string `json:"path"`
}

type storefrontMemberEntry struct {
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	Path     string `json:"path"`
}

type siteConfigDTO struct {
	TenantID uint64 `json:"tenant_id"`

	CustomDomain   string `json:"custom_domain"`
	DomainVerified int8   `json:"domain_verified"`
	SSLStatus      string `json:"ssl_status"`

	BrandName         string `json:"brand_name"`
	BrandLogo         string `json:"brand_logo"`
	PrimaryColor      string `json:"primary_color"`
	HidePlatformBrand int8   `json:"hide_platform_brand"`
	FooterText        string `json:"footer_text"`

	DeploymentMode  string `json:"deployment_mode"`
	PrivateEndpoint string `json:"private_endpoint"`
	PrivateNotes    string `json:"private_notes"`

	StorefrontNotice            string                  `json:"storefront_notice"`
	StorefrontHeroTitle         string                  `json:"storefront_hero_title"`
	StorefrontHeroSubtitle      string                  `json:"storefront_hero_subtitle"`
	StorefrontSearchPlaceholder string                  `json:"storefront_search_placeholder"`
	StorefrontCategoryTitle     string                  `json:"storefront_category_title"`
	StorefrontCouponTitle       string                  `json:"storefront_coupon_title"`
	StorefrontHotTitle          string                  `json:"storefront_hot_title"`
	StorefrontRecommendTitle    string                  `json:"storefront_recommend_title"`
	StorefrontQuickEntries      []storefrontQuickEntry  `json:"storefront_quick_entries"`
	StorefrontServiceCards      []storefrontServiceCard `json:"storefront_service_cards"`
	StorefrontBanners           []storefrontBanner      `json:"storefront_banners"`
	StorefrontPromoCards        []storefrontPromoCard   `json:"storefront_promo_cards"`
	StorefrontMemberEntries     []storefrontMemberEntry `json:"storefront_member_entries"`
	StorefrontHomeSections      []string                `json:"storefront_home_sections"`
	StorefrontProfileSections   []string                `json:"storefront_profile_sections"`
	StorefrontSearchKeywords    []string                `json:"storefront_search_keywords"`
}

func defaultSiteConfig(tid uint64) *model.TenantSiteConfig {
	return &model.TenantSiteConfig{
		TenantID:                    tid,
		PrimaryColor:                "#FF6B4A",
		DeploymentMode:              "shared",
		SSLStatus:                   "none",
		StorefrontNotice:            "新人福利已开启，领券后下单更划算",
		StorefrontHeroTitle:         "限时福利 · 热卖推荐 · 会员优选",
		StorefrontHeroSubtitle:      "支持优惠券、热卖推荐、会员积分、订单查看与地址管理，适合作为微信商城 SaaS 租户首页模板。",
		StorefrontSearchPlaceholder: "搜索水果、零食、会员权益",
		StorefrontCategoryTitle:     "热门分类",
		StorefrontCouponTitle:       "限时优惠券",
		StorefrontHotTitle:          "爆款热卖",
		StorefrontRecommendTitle:    "猜你喜欢",
		StorefrontQuickEntries:      `[{"title":"全部商品","subtitle":"热销排行","path":"/catalog"},{"title":"领券中心","subtitle":"新人福利","path":"/coupons"},{"title":"购物车","subtitle":"一键结算","path":"/cart"},{"title":"会员中心","subtitle":"积分权益","path":"/profile"}]`,
		StorefrontServiceCards:      `[{"title":"新人专享","desc":"满 99 减 10 / 周末折扣券已开放"},{"title":"会员体验","desc":"登录后可直接领券、下单、查看订单"}]`,
		StorefrontBanners:           `[{"title":"新人首单礼","subtitle":"登录领取专享优惠券包","image":"https://images.unsplash.com/photo-1542838132-92c53300491e?auto=format&fit=crop&w=1200&q=80","path":"/coupons"},{"title":"当季热卖","subtitle":"热销水果 / 零食 / 组合装每日更新","image":"https://images.unsplash.com/photo-1516684732162-798a0062be99?auto=format&fit=crop&w=1200&q=80","path":"/catalog?source=hot"}]`,
		StorefrontPromoCards:        `[{"title":"限时秒杀","subtitle":"每日 10 点 / 20 点上新","tag":"今日必抢","path":"/catalog?source=hot"},{"title":"会员专区","subtitle":"积分权益 / 福利券 / 订单服务","tag":"会员优先","path":"/profile"}]`,
		StorefrontMemberEntries:     `[{"title":"我的订单","subtitle":"待支付 / 售后进度","path":"/orders"},{"title":"优惠券","subtitle":"查看已领取福利","path":"/coupons"},{"title":"收货地址","subtitle":"管理常用地址","path":"/addresses"},{"title":"购物车","subtitle":"快捷去结算","path":"/cart"}]`,
		StorefrontHomeSections:      `["banners","quick_entries","promo_cards","service_cards","categories","coupons","hot","recommend"]`,
		StorefrontProfileSections:   `["member_entries","member_info","addresses","points"]`,
		StorefrontSearchKeywords:    `["水果礼盒","零食组合","限时秒杀","新人优惠券"]`,
	}
}

func normalizeSiteConfig(tid uint64, s *model.TenantSiteConfig) *model.TenantSiteConfig {
	if s == nil {
		return defaultSiteConfig(tid)
	}
	d := defaultSiteConfig(tid)
	if s.TenantID == 0 {
		s.TenantID = tid
	}
	if s.PrimaryColor == "" {
		s.PrimaryColor = d.PrimaryColor
	}
	if s.DeploymentMode == "" {
		s.DeploymentMode = d.DeploymentMode
	}
	if s.SSLStatus == "" {
		s.SSLStatus = d.SSLStatus
	}
	if s.StorefrontNotice == "" {
		s.StorefrontNotice = d.StorefrontNotice
	}
	if s.StorefrontHeroTitle == "" {
		s.StorefrontHeroTitle = d.StorefrontHeroTitle
	}
	if s.StorefrontHeroSubtitle == "" {
		s.StorefrontHeroSubtitle = d.StorefrontHeroSubtitle
	}
	if s.StorefrontSearchPlaceholder == "" {
		s.StorefrontSearchPlaceholder = d.StorefrontSearchPlaceholder
	}
	if s.StorefrontCategoryTitle == "" {
		s.StorefrontCategoryTitle = d.StorefrontCategoryTitle
	}
	if s.StorefrontCouponTitle == "" {
		s.StorefrontCouponTitle = d.StorefrontCouponTitle
	}
	if s.StorefrontHotTitle == "" {
		s.StorefrontHotTitle = d.StorefrontHotTitle
	}
	if s.StorefrontRecommendTitle == "" {
		s.StorefrontRecommendTitle = d.StorefrontRecommendTitle
	}
	if s.StorefrontQuickEntries == "" {
		s.StorefrontQuickEntries = d.StorefrontQuickEntries
	}
	if s.StorefrontServiceCards == "" {
		s.StorefrontServiceCards = d.StorefrontServiceCards
	}
	if s.StorefrontBanners == "" {
		s.StorefrontBanners = d.StorefrontBanners
	}
	if s.StorefrontPromoCards == "" {
		s.StorefrontPromoCards = d.StorefrontPromoCards
	}
	if s.StorefrontMemberEntries == "" {
		s.StorefrontMemberEntries = d.StorefrontMemberEntries
	}
	if s.StorefrontHomeSections == "" {
		s.StorefrontHomeSections = d.StorefrontHomeSections
	}
	if s.StorefrontProfileSections == "" {
		s.StorefrontProfileSections = d.StorefrontProfileSections
	}
	if s.StorefrontSearchKeywords == "" {
		s.StorefrontSearchKeywords = d.StorefrontSearchKeywords
	}
	return s
}

func decodeQuickEntries(raw string) []storefrontQuickEntry {
	rows := make([]storefrontQuickEntry, 0)
	if raw == "" {
		return rows
	}
	_ = json.Unmarshal([]byte(raw), &rows)
	return rows
}

func decodeServiceCards(raw string) []storefrontServiceCard {
	rows := make([]storefrontServiceCard, 0)
	if raw == "" {
		return rows
	}
	_ = json.Unmarshal([]byte(raw), &rows)
	return rows
}

func decodeBanners(raw string) []storefrontBanner {
	rows := make([]storefrontBanner, 0)
	if raw == "" {
		return rows
	}
	_ = json.Unmarshal([]byte(raw), &rows)
	return rows
}

func decodePromoCards(raw string) []storefrontPromoCard {
	rows := make([]storefrontPromoCard, 0)
	if raw == "" {
		return rows
	}
	_ = json.Unmarshal([]byte(raw), &rows)
	return rows
}

func decodeMemberEntries(raw string) []storefrontMemberEntry {
	rows := make([]storefrontMemberEntry, 0)
	if raw == "" {
		return rows
	}
	_ = json.Unmarshal([]byte(raw), &rows)
	return rows
}

func decodeStringList(raw string) []string {
	rows := make([]string, 0)
	if raw == "" {
		return rows
	}
	_ = json.Unmarshal([]byte(raw), &rows)
	return rows
}

func encodeJSON(v interface{}) string {
	bs, err := json.Marshal(v)
	if err != nil {
		return "[]"
	}
	return string(bs)
}

func toSiteConfigDTO(s *model.TenantSiteConfig) *siteConfigDTO {
	return &siteConfigDTO{
		TenantID: s.TenantID,

		CustomDomain:   s.CustomDomain,
		DomainVerified: s.DomainVerified,
		SSLStatus:      s.SSLStatus,

		BrandName:         s.BrandName,
		BrandLogo:         s.BrandLogo,
		PrimaryColor:      s.PrimaryColor,
		HidePlatformBrand: s.HidePlatformBrand,
		FooterText:        s.FooterText,

		DeploymentMode:  s.DeploymentMode,
		PrivateEndpoint: s.PrivateEndpoint,
		PrivateNotes:    s.PrivateNotes,

		StorefrontNotice:            s.StorefrontNotice,
		StorefrontHeroTitle:         s.StorefrontHeroTitle,
		StorefrontHeroSubtitle:      s.StorefrontHeroSubtitle,
		StorefrontSearchPlaceholder: s.StorefrontSearchPlaceholder,
		StorefrontCategoryTitle:     s.StorefrontCategoryTitle,
		StorefrontCouponTitle:       s.StorefrontCouponTitle,
		StorefrontHotTitle:          s.StorefrontHotTitle,
		StorefrontRecommendTitle:    s.StorefrontRecommendTitle,
		StorefrontQuickEntries:      decodeQuickEntries(s.StorefrontQuickEntries),
		StorefrontServiceCards:      decodeServiceCards(s.StorefrontServiceCards),
		StorefrontBanners:           decodeBanners(s.StorefrontBanners),
		StorefrontPromoCards:        decodePromoCards(s.StorefrontPromoCards),
		StorefrontMemberEntries:     decodeMemberEntries(s.StorefrontMemberEntries),
		StorefrontHomeSections:      decodeStringList(s.StorefrontHomeSections),
		StorefrontProfileSections:   decodeStringList(s.StorefrontProfileSections),
		StorefrontSearchKeywords:    decodeStringList(s.StorefrontSearchKeywords),
	}
}

func (h *SiteConfigHandler) Get(c *gin.Context) {
	ctx := c.Request.Context()
	s, err := h.repo.Get(ctx)
	if err != nil {
		response.Fail(c, err)
		return
	}
	s = normalizeSiteConfig(ctxkeys.GetTenant(ctx).ID, s)
	response.OK(c, toSiteConfigDTO(s))
}

func (h *SiteConfigHandler) GetStorefront(c *gin.Context) {
	ctx := c.Request.Context()
	s, err := h.repo.Get(ctx)
	if err != nil {
		response.Fail(c, err)
		return
	}
	s = normalizeSiteConfig(ctxkeys.GetTenant(ctx).ID, s)
	response.OK(c, toSiteConfigDTO(s))
}

func (h *SiteConfigHandler) MiniQRCode(c *gin.Context) {
	ctx := c.Request.Context()
	tenant := ctxkeys.GetTenant(ctx)
	code := ""
	if tenant != nil {
		code = strings.TrimSpace(tenant.Code)
	}
	if tenant == nil || tenant.ID == 0 || code == "" {
		response.Fail(c, apperr.ErrTenantRequired)
		return
	}
	admin := ctxkeys.GetAdmin(ctx)
	if admin == nil || admin.TenantID == 0 || admin.TenantID != tenant.ID {
		response.Fail(c, apperr.ErrForbidden)
		return
	}
	page := "pages/home/index"
	scene := fmt.Sprintf("t=%s", code)
	query := fmt.Sprintf("tenantCode=%s", url.QueryEscape(code))
	path := fmt.Sprintf("%s?%s", page, query)
	payload := fmt.Sprintf("go-shoppings-miniprogram://%s", path)
	png, err := qrcode.Encode(payload, qrcode.Medium, 360)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{
		"tenant_id":      tenant.ID,
		"tenant_code":    code,
		"page":           page,
		"scene":          scene,
		"query":          query,
		"path":           path,
		"qr_payload":     payload,
		"image_data_url": "data:image/png;base64," + base64.StdEncoding.EncodeToString(png),
		"simulated":      true,
	})
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
	cur = normalizeSiteConfig(ctxkeys.GetTenant(ctx).ID, cur)
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
	response.OK(c, toSiteConfigDTO(cur))
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
	cur = normalizeSiteConfig(ctxkeys.GetTenant(ctx).ID, cur)
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
	response.OK(c, toSiteConfigDTO(cur))
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
	cur = normalizeSiteConfig(ctxkeys.GetTenant(ctx).ID, cur)
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
	response.OK(c, toSiteConfigDTO(cur))
}

type storefrontReq struct {
	PrimaryColor                string                  `json:"primary_color"`
	StorefrontNotice            string                  `json:"storefront_notice"`
	StorefrontHeroTitle         string                  `json:"storefront_hero_title"`
	StorefrontHeroSubtitle      string                  `json:"storefront_hero_subtitle"`
	StorefrontSearchPlaceholder string                  `json:"storefront_search_placeholder"`
	StorefrontCategoryTitle     string                  `json:"storefront_category_title"`
	StorefrontCouponTitle       string                  `json:"storefront_coupon_title"`
	StorefrontHotTitle          string                  `json:"storefront_hot_title"`
	StorefrontRecommendTitle    string                  `json:"storefront_recommend_title"`
	StorefrontQuickEntries      []storefrontQuickEntry  `json:"storefront_quick_entries"`
	StorefrontServiceCards      []storefrontServiceCard `json:"storefront_service_cards"`
	StorefrontBanners           []storefrontBanner      `json:"storefront_banners"`
	StorefrontPromoCards        []storefrontPromoCard   `json:"storefront_promo_cards"`
	StorefrontMemberEntries     []storefrontMemberEntry `json:"storefront_member_entries"`
	StorefrontHomeSections      []string                `json:"storefront_home_sections"`
	StorefrontProfileSections   []string                `json:"storefront_profile_sections"`
	StorefrontSearchKeywords    []string                `json:"storefront_search_keywords"`
}

func (h *SiteConfigHandler) UpdateStorefront(c *gin.Context) {
	var req storefrontReq
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
	cur = normalizeSiteConfig(ctxkeys.GetTenant(ctx).ID, cur)
	if color := strings.TrimSpace(req.PrimaryColor); color != "" {
		cur.PrimaryColor = color
	}
	cur.StorefrontNotice = req.StorefrontNotice
	cur.StorefrontHeroTitle = req.StorefrontHeroTitle
	cur.StorefrontHeroSubtitle = req.StorefrontHeroSubtitle
	cur.StorefrontSearchPlaceholder = req.StorefrontSearchPlaceholder
	cur.StorefrontCategoryTitle = req.StorefrontCategoryTitle
	cur.StorefrontCouponTitle = req.StorefrontCouponTitle
	cur.StorefrontHotTitle = req.StorefrontHotTitle
	cur.StorefrontRecommendTitle = req.StorefrontRecommendTitle
	if req.StorefrontQuickEntries != nil {
		cur.StorefrontQuickEntries = encodeJSON(req.StorefrontQuickEntries)
	}
	cur.StorefrontServiceCards = encodeJSON(req.StorefrontServiceCards)
	cur.StorefrontBanners = encodeJSON(req.StorefrontBanners)
	cur.StorefrontPromoCards = encodeJSON(req.StorefrontPromoCards)
	cur.StorefrontMemberEntries = encodeJSON(req.StorefrontMemberEntries)
	cur.StorefrontHomeSections = encodeJSON(req.StorefrontHomeSections)
	cur.StorefrontProfileSections = encodeJSON(req.StorefrontProfileSections)
	cur.StorefrontSearchKeywords = encodeJSON(req.StorefrontSearchKeywords)
	if err := h.repo.Upsert(ctx, cur); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, toSiteConfigDTO(cur))
}
