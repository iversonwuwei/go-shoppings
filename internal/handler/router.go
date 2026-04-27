package handler

import (
	"strings"

	"github.com/gin-gonic/gin"

	"wechat-mall-saas/internal/handler/admin"
	"wechat-mall-saas/internal/handler/member"
	"wechat-mall-saas/internal/handler/middleware"
	"wechat-mall-saas/internal/model"
	"wechat-mall-saas/internal/pkg/jwtx"
	"wechat-mall-saas/internal/pkg/response"
	"wechat-mall-saas/internal/repository"
	"wechat-mall-saas/internal/service"
)

// Deps 注入所有 handler 需要的依赖（由 main.go 装配）
type Deps struct {
	JWT *jwtx.Manager

	Tenant   *service.TenantService
	Auth     *service.AuthService
	Product  *service.ProductService
	Category *service.CategoryService
	Order    *service.OrderService
	Payment  *service.PaymentService
	Coupon   *service.CouponService
	Member   *service.MemberService

	PlanFeatureRepo *repository.PlanFeatureRepo
	TenantRepo      *repository.TenantRepo

	AdminAuthH         *admin.AuthHandler
	AdminProductH      *admin.ProductHandler
	AdminCategoryH     *admin.CategoryHandler
	AdminOrderH        *admin.OrderHandler
	AdminPlatformH     *admin.PlatformHandler
	AdminSettingsH     *admin.SettingsHandler
	AdminMemberLvlH    *admin.MemberLevelHandler
	AdminSeckillH      *admin.SeckillHandler
	AdminCouponH       *admin.CouponHandler
	AdminPointsH       *admin.PointsHandler
	AdminGrouponH      *admin.GrouponHandler
	AdminDistributionH *admin.DistributionHandler
	AdminDeliveryH     *admin.DeliveryHandler
	AdminSiteH         *admin.SiteConfigHandler
	AdminSubH          *admin.SubscriptionHandler
	PlatformSettingsH  *admin.PlatformSettingsHandler
	PlatformGlobalH    *admin.PlatformGlobalSettingsHandler
	PlatformUsersH     *admin.PlatformUserHandler
	UploadH            *admin.UploadHandler
	StorageBasePath    string // 本地静态文件根目录

	PlatformSmsH        *admin.PlatformSmsHandler
	PlatformApiAccessH  *admin.PlatformApiAccessHandler
	PlatformDomainH     *admin.PlatformDomainHandler
	PlatformDeploymentH *admin.PlatformDeploymentHandler
	PlatformStorefrontH *admin.PlatformStorefrontHandler

	MemberAuthH     *member.AuthHandler
	MemberProductH  *member.ProductHandler
	MemberCategoryH *member.CategoryHandler
	MemberOrderH    *member.OrderHandler
	MemberCouponH   *member.CouponHandler
	MemberAddressH  *member.AddressHandler
	MemberPointsH   *member.PointsHandler
	MemberMemberH   *member.MemberHandler
	MemberSeckillH  *member.SeckillHandler

	PaymentH *PaymentHandler

	RateQPS   int
	RateBurst int
}

// New 构造 Gin 引擎并注册所有路由
func New(d *Deps) *gin.Engine {
	r := gin.New()
	r.Use(middleware.Recovery(), middleware.CORS(), middleware.Logging(), middleware.IPRateLimit(d.RateQPS, d.RateBurst))

	r.GET("/healthz", func(c *gin.Context) { response.OK(c, gin.H{"status": "ok"}) })

	// 静态文件：本地上传图片
	if d.StorageBasePath != "" {
		r.Static("/uploads", d.StorageBasePath)
	}

	v1 := r.Group("/api/v1")

	// 平台管理员
	plat := v1.Group("/platform")
	{
		plat.POST("/auth/login", d.AdminAuthH.Login)
		plat.POST("/auth/verify-code/send", d.AdminAuthH.SendCode)
		plat.POST("/auth/login-sms", d.AdminAuthH.PlatformLoginBySMS)
		plat.POST("/auth/reset-password", d.AdminAuthH.PlatformResetPassword)
		sec := plat.Group("")
		sec.Use(middleware.AdminAuth(d.JWT))
		sec.GET("/dashboard", d.AdminPlatformH.Dashboard)
		sec.GET("/tenants", d.AdminPlatformH.ListTenants)
		sec.POST("/tenants/:id/audit", d.AdminPlatformH.AuditTenant)

		sec.GET("/plans", d.AdminPlatformH.ListPlans)
		sec.POST("/plans", d.AdminPlatformH.CreatePlan)
		sec.PUT("/plans/:id", d.AdminPlatformH.UpdatePlan)
		sec.DELETE("/plans/:id", d.AdminPlatformH.DeletePlan)

		sec.GET("/features", d.AdminPlatformH.ListFeatures)
		sec.POST("/features", d.AdminPlatformH.CreateFeature)
		sec.PUT("/features/:id", d.AdminPlatformH.UpdateFeature)
		sec.DELETE("/features/:id", d.AdminPlatformH.DeleteFeature)

		sec.GET("/payment-configs", d.PlatformSettingsH.ListPaymentAudit)
		sec.POST("/payment-configs/:id/audit", d.PlatformSettingsH.AuditPayment)

		sec.GET("/carriers", d.PlatformSettingsH.ListCarriers)
		sec.POST("/carriers", d.PlatformSettingsH.CreateCarrier)
		sec.PUT("/carriers/:id", d.PlatformSettingsH.UpdateCarrier)
		sec.PATCH("/carriers/:id/enabled", d.PlatformSettingsH.ToggleCarrier)
		sec.DELETE("/carriers/:id", d.PlatformSettingsH.DeleteCarrier)

		// 平台全局设置（平台名 / Logo / 平台微信支付商户号 / 客服联系方式）
		sec.GET("/settings", d.PlatformGlobalH.Get)
		sec.PUT("/settings", d.PlatformGlobalH.Update)

		// 通用文件上传（平台，tenant_id=0）
		sec.POST("/upload/image", d.UploadH.Image)

		// 平台用户管理（仅超级管理员可管理，其他角色只能查看自己）
		sec.GET("/me", d.PlatformUsersH.Me)
		userGrp := sec.Group("/users")
		userGrp.Use(middleware.RequireRole(admin.PlatformRoleSuper))
		userGrp.GET("", d.PlatformUsersH.List)
		userGrp.POST("", d.PlatformUsersH.Create)
		userGrp.PUT("/:id", d.PlatformUsersH.Update)
		userGrp.POST("/:id/reset-password", d.PlatformUsersH.ResetPassword)
		userGrp.DELETE("/:id", d.PlatformUsersH.Delete)

		// 平台统一管理 短信通知（网关/模板/日志）
		sec.GET("/sms/settings", d.PlatformSmsH.GetSettings)
		sec.PUT("/sms/settings", d.PlatformSmsH.UpdateSettings)
		sec.GET("/sms/templates", d.PlatformSmsH.ListTemplates)
		sec.PUT("/sms/templates/:id", d.PlatformSmsH.UpdateTemplate)
		sec.DELETE("/sms/templates/:id", d.PlatformSmsH.DeleteTemplate)
		sec.GET("/sms/logs", d.PlatformSmsH.ListLogs)

		// 平台统一管理 开放API 凭据
		sec.GET("/api-tokens", d.PlatformApiAccessH.List)
		sec.POST("/api-tokens", d.PlatformApiAccessH.Create)
		sec.PUT("/api-tokens/:id", d.PlatformApiAccessH.Update)
		sec.DELETE("/api-tokens/:id", d.PlatformApiAccessH.Delete)
		sec.POST("/api-tokens/:id/regenerate", d.PlatformApiAccessH.Regenerate)
		sec.GET("/api-tokens/logs", d.PlatformApiAccessH.ListLogs)

		// 平台审核 自定义域名
		sec.GET("/domains", d.PlatformDomainH.List)
		sec.POST("/domains/:tid/verify", d.PlatformDomainH.Verify)
		sec.POST("/domains/:tid/reject", d.PlatformDomainH.Reject)

		// 平台管理 私有部署
		sec.GET("/deployments", d.PlatformDeploymentH.List)
		sec.PUT("/deployments", d.PlatformDeploymentH.Update)

		// 平台统一管理 商品分类（tenant_id=0，租户共享）
		catGrp := sec.Group("/categories")
		catGrp.Use(middleware.RequireRole(admin.PlatformRoleSuper, admin.PlatformRoleOperator))
		catGrp.GET("", d.AdminCategoryH.ListAll)
		catGrp.POST("", d.AdminCategoryH.Create)
		catGrp.PUT("/:id", d.AdminCategoryH.Update)
		catGrp.DELETE("/:id", d.AdminCategoryH.Delete)

		// 平台运营 租户管理（仅 super/operator 可手动调整）
		tenantMgr := sec.Group("/tenants")
		tenantMgr.Use(middleware.RequireRole(admin.PlatformRoleSuper, admin.PlatformRoleOperator))
		tenantMgr.PATCH("/:id/status", d.AdminPlatformH.UpdateTenantStatus)
		tenantMgr.PATCH("/:id/plan", d.AdminPlatformH.UpdateTenantPlan)
		tenantMgr.PATCH("/:id/features", d.AdminPlatformH.UpdateTenantFeatures)
		tenantMgr.GET("/:id/storefront/quick-entries", d.PlatformStorefrontH.GetQuickEntries)
		tenantMgr.PUT("/:id/storefront/quick-entries", d.PlatformStorefrontH.UpdateQuickEntries)
	}

	// 公开端（产品介绍页 / 申请入驻）
	pub := v1.Group("/public")
	{
		pub.GET("/plans", func(c *gin.Context) {
			plans, err := d.Tenant.PublicPlans(c.Request.Context())
			if err != nil {
				response.Fail(c, err)
				return
			}
			response.OK(c, plans)
		})
		pub.GET("/features", func(c *gin.Context) {
			rows, err := d.PlanFeatureRepo.List(c.Request.Context(), true)
			if err != nil {
				response.Fail(c, err)
				return
			}
			response.OK(c, rows)
		})
		pub.POST("/apply", func(c *gin.Context) {
			var body struct {
				model.Tenant
				Username   string `json:"username"`
				Password   string `json:"password"`
				VerifyCode string `json:"verify_code"`
			}
			if err := c.ShouldBindJSON(&body); err != nil {
				response.FailCode(c, 20001, err.Error())
				return
			}
			t, a, err := d.Auth.RegisterTenantWithAdmin(c.Request.Context(), d.Tenant, &body.Tenant, body.Username, body.Password, body.VerifyCode)
			if err != nil {
				response.Fail(c, err)
				return
			}
			resp := gin.H{"id": t.ID, "code": t.Code, "status": t.Status}
			if a != nil {
				resp["admin_id"] = a.ID
				resp["username"] = a.Username
			}
			response.OK(c, resp)
		})
		// 入驻申请：手机号验证码
		pub.POST("/verify-code/send", d.AdminAuthH.SendCode)
		// 按租户 code 或已审核自定义域名解析租户摘要，供前台/登录页使用（隐藏内部主键）
		pub.GET("/tenant/resolve", func(c *gin.Context) {
			code := strings.TrimSpace(c.Query("code"))
			host := strings.TrimSpace(c.Query("host"))
			if code == "" && host == "" {
				response.FailCode(c, 20001, "code 或 host 不能为空")
				return
			}
			var (
				t   *model.Tenant
				err error
			)
			if host != "" {
				t, err = d.TenantRepo.FindByCustomDomain(c.Request.Context(), host)
			} else {
				t, err = d.TenantRepo.FindByCode(c.Request.Context(), code)
			}
			if err != nil {
				response.Fail(c, err)
				return
			}
			if t == nil {
				response.FailCode(c, 40400, "租户不存在")
				return
			}
			response.OK(c, gin.H{
				"id":           t.ID,
				"code":         t.Code,
				"status":       t.Status,
				"company_name": t.CompanyName,
				"brand_name":   t.BrandName,
			})
		})
		// 微信支付回调（租户订阅付费，平台统一商户号）
		pub.POST("/subscription/callback", d.AdminSubH.WxpayCallback)
	}

	// 租户注册（兼容老地址）
	v1.POST("/tenant/register", func(c *gin.Context) {
		var body model.Tenant
		if err := c.ShouldBindJSON(&body); err != nil {
			response.FailCode(c, 20001, err.Error())
			return
		}
		t, err := d.Tenant.Register(c.Request.Context(), &body)
		if err != nil {
			response.Fail(c, err)
			return
		}
		response.OK(c, t)
	})

	// 租户后台（管理员）
	ad := v1.Group("/admin")
	ad.POST("/auth/login", d.AdminAuthH.Login)
	ad.POST("/auth/verify-code/send", d.AdminAuthH.SendCode)
	ad.POST("/auth/login-sms", d.AdminAuthH.LoginBySMS)
	ad.POST("/auth/reset-password", d.AdminAuthH.ResetPassword)
	adAuth := ad.Group("")
	adAuth.Use(middleware.Tenant(d.Tenant, true), middleware.AdminAuth(d.JWT))
	{
		adAuth.GET("/products", d.AdminProductH.List)
		adAuth.GET("/products/import-template", d.AdminProductH.ImportTemplate)
		adAuth.POST("/products/import", d.AdminProductH.Import)
		adAuth.POST("/products", d.AdminProductH.Create)
		adAuth.PUT("/products/:id", d.AdminProductH.Update)
		adAuth.PATCH("/products/:id/status", d.AdminProductH.UpdateStatus)
		adAuth.DELETE("/products/:id", d.AdminProductH.Delete)
		adAuth.POST("/products/:id/skus", middleware.RequireFeature(service.FeatureMultiSKU), d.AdminProductH.CreateSKU)
		adAuth.GET("/inventory/products", d.AdminProductH.InventoryList)
		adAuth.PATCH("/inventory/products/:id", d.AdminProductH.AdjustInventory)

		adAuth.GET("/categories", d.AdminCategoryH.List)
		adAuth.PUT("/categories/:id/media", d.AdminCategoryH.UpdateTenantAsset)
		// 分类改由平台统一管理，租户端只读

		adAuth.GET("/orders", d.AdminOrderH.List)
		adAuth.GET("/orders/:id", d.AdminOrderH.Detail)
		adAuth.GET("/orders/:id/logs", d.AdminOrderH.Logs)
		adAuth.POST("/orders/:id/ship", d.AdminOrderH.Ship)
		adAuth.GET("/order-messages", d.AdminOrderH.Messages)
		adAuth.POST("/order-messages/read-all", d.AdminOrderH.MarkAllMessagesRead)
		adAuth.POST("/order-messages/:id/read", d.AdminOrderH.MarkMessageRead)

		adAuth.GET("/settings/payment", d.AdminSettingsH.ListPayment)
		adAuth.PUT("/settings/payment", d.AdminSettingsH.SubmitPayment)
		adAuth.GET("/settings/carriers", d.AdminSettingsH.ListCarriers)
		adAuth.GET("/settings/carriers/track", d.AdminSettingsH.QueryTrack)

		// 通用文件上传（租户）
		adAuth.POST("/upload/image", d.UploadH.Image)

		// 会员等级管理（需套餐包含 member_level 功能）
		lvl := adAuth.Group("/member/levels", middleware.RequireFeature(service.FeatureMemberLevel))
		{
			lvl.GET("", d.AdminMemberLvlH.List)
			lvl.POST("", d.AdminMemberLvlH.Create)
			lvl.PUT("/:id", d.AdminMemberLvlH.Update)
			lvl.DELETE("/:id", d.AdminMemberLvlH.Delete)
		}

		// 秒杀活动管理（需套餐包含 seckill 功能）
		sk := adAuth.Group("/seckill", middleware.RequireFeature(service.FeatureSeckill))
		{
			sk.GET("/activities", d.AdminSeckillH.List)
			sk.POST("/activities", d.AdminSeckillH.Create)
			sk.PUT("/activities/:id", d.AdminSeckillH.Update)
			sk.DELETE("/activities/:id", d.AdminSeckillH.Delete)
		}

		// 优惠券管理（需套餐包含 coupon 功能）
		cp := adAuth.Group("/coupons", middleware.RequireFeature(service.FeatureCoupon))
		{
			cp.GET("", d.AdminCouponH.List)
			cp.POST("", d.AdminCouponH.Create)
			cp.PUT("/:id", d.AdminCouponH.Update)
			cp.DELETE("/:id", d.AdminCouponH.Delete)
		}

		// 积分规则（需套餐包含 points 功能）
		pts := adAuth.Group("/points", middleware.RequireFeature(service.FeaturePoints))
		{
			pts.GET("/settings", d.AdminPointsH.Get)
			pts.PUT("/settings", d.AdminPointsH.Update)
		}

		// 拼团活动（需套餐包含 group_buy 功能）
		gp := adAuth.Group("/groupon", middleware.RequireFeature(service.FeatureGroupBuy))
		{
			gp.GET("/activities", d.AdminGrouponH.List)
			gp.POST("/activities", d.AdminGrouponH.Create)
			gp.PUT("/activities/:id", d.AdminGrouponH.Update)
			gp.DELETE("/activities/:id", d.AdminGrouponH.Delete)
			gp.GET("/groupons", d.AdminGrouponH.Groupons)
		}

		// 分销管理（需套餐包含 distribution 功能）
		ds := adAuth.Group("/distribution", middleware.RequireFeature(service.FeatureDistribution))
		{
			ds.GET("/settings", d.AdminDistributionH.GetSettings)
			ds.PUT("/settings", d.AdminDistributionH.UpdateSettings)
			ds.GET("/distributors", d.AdminDistributionH.ListDistributors)
			ds.PUT("/distributors/:id/audit", d.AdminDistributionH.AuditDistributor)
			ds.GET("/commissions", d.AdminDistributionH.ListCommissions)
		}

		// 配送设置（整体读取不限制；写入按 section 分别 gate）
		adAuth.GET("/delivery/settings", d.AdminDeliveryH.Get)
		// express/city/self_pickup 任一开通即可更新（同一整行），这里将三个功能码最宽松用 express 作为写入门槛
		adAuth.PUT("/delivery/settings",
			middleware.RequireFeature(service.FeatureExpressDelivery),
			d.AdminDeliveryH.Update)

		// 站点配置（读取不限制；各 section 分别 gate）
		adAuth.GET("/site/config", d.AdminSiteH.Get)
		adAuth.PUT("/site/domain",
			middleware.RequireFeature(service.FeatureCustomDomain),
			d.AdminSiteH.UpdateDomain)
		adAuth.PUT("/site/brand",
			middleware.RequireFeature(service.FeatureWhiteLabel),
			d.AdminSiteH.UpdateBrand)
		adAuth.PUT("/site/storefront", d.AdminSiteH.UpdateStorefront)
		adAuth.PUT("/site/deployment",
			middleware.RequireFeature(service.FeaturePrivateDeployment),
			d.AdminSiteH.UpdateDeployment)

		// 订阅付费（租户向平台统一商户号付款 / 查询订单）
		adAuth.POST("/subscription/orders", d.AdminSubH.Create)
		adAuth.GET("/subscription/orders", d.AdminSubH.List)
	}

	// 小程序（会员端）
	mb := v1.Group("/member")
	mb.Use(middleware.Tenant(d.Tenant, true))
	mb.POST("/auth/dev-login", d.MemberAuthH.DevLogin)
	mb.POST("/auth/login-by-wechat", d.MemberAuthH.LoginByWechat)
	mb.GET("/products", d.MemberProductH.List)
	mb.GET("/products/hot", d.MemberProductH.Hot)
	mb.GET("/products/recommend", d.MemberProductH.Recommend)
	mb.GET("/products/:id", d.MemberProductH.Detail)
	mb.GET("/categories", d.MemberCategoryH.List)
	mb.GET("/coupons", d.MemberCouponH.Available)
	mb.GET("/storefront/config", d.AdminSiteH.GetStorefront)
	mb.GET("/seckill/activities", middleware.RequireFeature(service.FeatureSeckill), d.MemberSeckillH.List)

	mbAuth := mb.Group("")
	mbAuth.Use(middleware.MemberAuth(d.JWT))
	{
		mbAuth.POST("/auth/bind-phone", d.MemberAuthH.BindPhone)
		mbAuth.GET("/profile", d.MemberMemberH.Profile)
		mbAuth.PUT("/profile", d.MemberMemberH.UpdateProfile)

		mbAuth.GET("/addresses", d.MemberAddressH.List)
		mbAuth.POST("/addresses", d.MemberAddressH.Create)

		mbAuth.POST("/orders", d.MemberOrderH.Create)
		mbAuth.GET("/orders", d.MemberOrderH.List)
		mbAuth.GET("/orders/:id", d.MemberOrderH.Detail)
		mbAuth.POST("/orders/:id/cancel", d.MemberOrderH.Cancel)
		mbAuth.POST("/orders/:id/confirm", d.MemberOrderH.Confirm)

		mbAuth.POST("/coupons/:id/receive", d.MemberCouponH.Receive)
		mbAuth.GET("/my/coupons", d.MemberCouponH.My)

		mbAuth.GET("/points/logs", d.MemberPointsH.Logs)

		mbAuth.POST("/payments", d.PaymentH.Create)
	}

	// 支付回调
	v1.POST("/payments/callback/wechat", d.PaymentH.WechatCallback)

	return r
}
