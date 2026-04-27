package database

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"wechat-mall-saas/internal/pkg/config"
)

func New(cfg config.DatabaseConfig) (*gorm.DB, error) {
	dsn, err := buildDSN(cfg)
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: cfg.PreferSimpleProtocol,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	return db, nil
}

func buildDSN(cfg config.DatabaseConfig) (string, error) {
	if dsn := strings.TrimSpace(cfg.DSN); dsn != "" {
		return dsn, nil
	}

	if cfg.Port == 0 {
		cfg.Port = 5432
	}
	if strings.TrimSpace(cfg.SSLMode) == "" {
		cfg.SSLMode = "require"
	}
	if strings.TrimSpace(cfg.Host) == "" || strings.TrimSpace(cfg.User) == "" || strings.TrimSpace(cfg.Name) == "" {
		return "", fmt.Errorf("database: set database.dsn or SUPABASE_DB_DSN/DATABASE_URL")
	}

	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Shanghai",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode), nil
}
