package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"wechat-mall-saas/internal/handler"
	"wechat-mall-saas/internal/handler/admin"
	"wechat-mall-saas/internal/handler/member"
	"wechat-mall-saas/internal/pkg/cache"
	"wechat-mall-saas/internal/pkg/config"
	"wechat-mall-saas/internal/pkg/database"
	"wechat-mall-saas/internal/pkg/jwtx"
	"wechat-mall-saas/internal/pkg/logger"
	"wechat-mall-saas/internal/pkg/wxpay"
	"wechat-mall-saas/internal/repository"
	"wechat-mall-saas/internal/service"
)

func main() {
	cfgPath := flag.String("config", "configs/config.yaml", "config file path")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		fmt.Printf("load config failed: %v\n", err)
		os.Exit(1)
	}
	if err := logger.Init(cfg.Logging); err != nil {
		fmt.Printf("init logger failed: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	db, err := database.New(cfg.Database)
	if err != nil {
		logger.L.Fatal("connect db failed", zap.Error(err))
	}
	rdb, err := cache.New(cfg.Redis)
	if err != nil {
		logger.L.Fatal("connect redis failed", zap.Error(err))
	}
	jwtMgr := jwtx.New(cfg.JWT.Secret, cfg.JWT.ExpireHours, cfg.JWT.RefreshExpireHours)

	// ========== 装配 Repo ==========
	planRepo := repository.NewPlanRepo(db)
	planFeatureRepo := repository.NewPlanFeatureRepo(db)
	tenantRepo := repository.NewTenantRepo(db)
	adminRepo := repository.NewAdminRepo(db)
	tenantPlanLogRepo := repository.NewTenantPlanLogRepo(db)

	productRepo := repository.NewProductRepo(db)
	skuRepo := repository.NewProductSKURepo(db)
	categoryRepo := repository.NewCategoryRepo(db)

	memberRepo := repository.NewMemberRepo(db)
	memberAddressRepo := repository.NewMemberAddressRepo(db)
	memberLevelRepo := repository.NewMemberLevelRepo(db)
	pointsLogRepo := repository.NewPointsLogRepo(db)

	orderRepo := repository.NewOrderRepo(db)
	orderLogRepo := repository.NewOrderLogRepo(db)

	paymentRepo := repository.NewPaymentRepo(db)
	couponRepo := repository.NewCouponRepo(db)
	memberCouponRepo := repository.NewMemberCouponRepo(db)

	paymentConfigRepo := repository.NewPaymentConfigRepo(db)
	carrierRepo := repository.NewShippingCarrierRepo(db)
	seckillRepo := repository.NewSeckillRepo(db)
	pointsSettingsRepo := repository.NewPointsSettingsRepo(db)
	grouponRepo := repository.NewGrouponRepo(db)
	distributionRepo := repository.NewDistributionRepo(db)
	smsRepo := repository.NewSmsRepo(db)
	apiTokenRepo := repository.NewApiTokenRepo(db)
	deliveryRepo := repository.NewDeliveryRepo(db)
	siteRepo := repository.NewSiteConfigRepo(db)
	subOrderRepo := repository.NewTenantSubscriptionOrderRepo(db)
	platformSettingsRepo := repository.NewPlatformSettingsRepo(db)

	// ========== 装配 Service ==========
	tenantSvc := service.NewTenantService(tenantRepo, adminRepo, planRepo, tenantPlanLogRepo, rdb)
	authSvc := service.NewAuthService(adminRepo, memberRepo, tenantRepo, jwtMgr, rdb, cfg.App.Env)
	productSvc := service.NewProductService(productRepo, skuRepo, categoryRepo, tenantSvc)
	categorySvc := service.NewCategoryService(categoryRepo)
	orderSvc := service.NewOrderService(orderRepo, orderLogRepo, productRepo, skuRepo, tenantSvc)
	paymentSvc := service.NewPaymentService(paymentRepo, orderRepo, orderLogRepo, tenantRepo, memberRepo, pointsLogRepo, pointsSettingsRepo, tenantSvc)
	couponSvc := service.NewCouponService(couponRepo, memberCouponRepo, tenantSvc)
	memberSvc := service.NewMemberService(memberRepo, memberAddressRepo, pointsLogRepo)
	settingsSvc := service.NewSettingsService(paymentConfigRepo, carrierRepo, tenantSvc)
	platformWxpay := wxpay.NewClient(wxpay.Config{
		AppID:      cfg.WxPay.AppID,
		MchID:      cfg.WxPay.MchID,
		APIv3Key:   cfg.WxPay.APIv3Key,
		CertSerial: cfg.WxPay.CertSerial,
		NotifyURL:  cfg.WxPay.NotifyURL,
	})
	subscriptionSvc := service.NewSubscriptionService(subOrderRepo, tenantRepo, planRepo, tenantPlanLogRepo, tenantSvc, platformWxpay, platformSettingsRepo)

	// ========== 装配 Handler ==========
	deps := &handler.Deps{
		JWT:      jwtMgr,
		Tenant:   tenantSvc,
		Auth:     authSvc,
		Product:  productSvc,
		Category: categorySvc,
		Order:    orderSvc,
		Payment:  paymentSvc,
		Coupon:   couponSvc,
		Member:   memberSvc,

		PlanFeatureRepo: planFeatureRepo,

		AdminAuthH:         admin.NewAuthHandler(authSvc),
		AdminProductH:      admin.NewProductHandler(productSvc),
		AdminCategoryH:     admin.NewCategoryHandler(categorySvc),
		AdminOrderH:        admin.NewOrderHandler(orderSvc),
		AdminPlatformH:     admin.NewPlatformHandler(tenantSvc, tenantRepo, planRepo, planFeatureRepo),
		AdminSettingsH:     admin.NewSettingsHandler(settingsSvc),
		AdminMemberLvlH:    admin.NewMemberLevelHandler(memberLevelRepo),
		AdminSeckillH:      admin.NewSeckillHandler(seckillRepo),
		AdminCouponH:       admin.NewCouponHandler(couponRepo),
		AdminPointsH:       admin.NewPointsHandler(pointsSettingsRepo),
		AdminGrouponH:      admin.NewGrouponHandler(grouponRepo),
		AdminDistributionH: admin.NewDistributionHandler(distributionRepo),
		AdminDeliveryH:     admin.NewDeliveryHandler(deliveryRepo),
		AdminSiteH:         admin.NewSiteConfigHandler(siteRepo),
		AdminSubH:          admin.NewSubscriptionHandler(subscriptionSvc),
		PlatformSettingsH:  admin.NewPlatformSettingsHandler(settingsSvc),
		PlatformGlobalH:    admin.NewPlatformGlobalSettingsHandler(platformSettingsRepo),

		PlatformSmsH:        admin.NewPlatformSmsHandler(smsRepo),
		PlatformApiAccessH:  admin.NewPlatformApiAccessHandler(apiTokenRepo),
		PlatformDomainH:     admin.NewPlatformDomainHandler(siteRepo),
		PlatformDeploymentH: admin.NewPlatformDeploymentHandler(siteRepo),

		MemberAuthH:     member.NewAuthHandler(authSvc, tenantRepo),
		MemberProductH:  member.NewProductHandler(productSvc),
		MemberCategoryH: member.NewCategoryHandler(categorySvc),
		MemberOrderH:    member.NewOrderHandler(orderSvc),
		MemberCouponH:   member.NewCouponHandler(couponSvc),
		MemberAddressH:  member.NewAddressHandler(memberSvc),
		MemberPointsH:   member.NewPointsHandler(memberSvc),
		MemberMemberH:   member.NewMemberHandler(memberSvc),
		MemberSeckillH:  member.NewSeckillHandler(seckillRepo),

		PaymentH: handler.NewPaymentHandler(paymentSvc),

		RateQPS:   cfg.RateLimit.QPS,
		RateBurst: cfg.RateLimit.Burst,
	}
	if !cfg.RateLimit.Enabled {
		deps.RateQPS = 0
	}

	r := handler.New(deps)
	srv := &http.Server{
		Addr:           fmt.Sprintf("%s:%d", cfg.App.Host, cfg.App.Port),
		Handler:        r,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		logger.L.Info("http server listening", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.L.Fatal("server error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.L.Info("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.L.Error("server shutdown error", zap.Error(err))
	}
	_ = rdb.Close()
	logger.L.Info("bye")
}
