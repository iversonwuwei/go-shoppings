package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"wechat-mall-saas/internal/pkg/cache"
	"wechat-mall-saas/internal/pkg/config"
	"wechat-mall-saas/internal/pkg/database"
	"wechat-mall-saas/internal/pkg/logger"
	"wechat-mall-saas/internal/pkg/wxpay"
	"wechat-mall-saas/internal/repository"
	"wechat-mall-saas/internal/service"
)

// worker 定时任务：目前主要任务是租户套餐到期扫描（Active→Overdue→Banned）
func main() {
	cfgPath := flag.String("config", "configs/config.yaml", "config file path")
	once := flag.Bool("once", false, "run subscription scan once and exit")
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

	tenantRepo := repository.NewTenantRepo(db)
	adminRepo := repository.NewAdminRepo(db)
	planRepo := repository.NewPlanRepo(db)
	tenantPlanLogRepo := repository.NewTenantPlanLogRepo(db)
	subOrderRepo := repository.NewTenantSubscriptionOrderRepo(db)

	tenantSvc := service.NewTenantService(tenantRepo, adminRepo, planRepo, tenantPlanLogRepo, rdb)
	wp := wxpay.NewClient(wxpay.Config{AppID: cfg.WxPay.AppID, MchID: cfg.WxPay.MchID})
	subSvc := service.NewSubscriptionService(subOrderRepo, tenantRepo, planRepo, tenantPlanLogRepo, tenantSvc, wp)

	runOnce := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		overdue, banned, err := subSvc.ScanAndTransition(ctx)
		if err != nil {
			logger.L.Error("subscription scan failed", zap.Error(err))
			return
		}
		logger.L.Info("subscription scan done",
			zap.Int("overdue_transitions", overdue),
			zap.Int("banned_transitions", banned),
		)
	}

	if *once {
		runOnce()
		return
	}

	logger.L.Info("worker started — subscription expiry scan every 6h")
	runOnce()
	ticker := time.NewTicker(6 * time.Hour)
	defer ticker.Stop()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case <-ticker.C:
			runOnce()
		case <-quit:
			logger.L.Info("worker stopped")
			return
		}
	}
}
