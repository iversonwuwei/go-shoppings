package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"wechat-mall-saas/internal/pkg/config"
	"wechat-mall-saas/internal/pkg/logger"
)

// worker 作为定时任务骨架：后续可接入订单自动取消、订单自动确认收货、套餐到期检测等
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

	logger.L.Info("worker started (stub) — add cron jobs here")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.L.Info("worker stopped")
}
