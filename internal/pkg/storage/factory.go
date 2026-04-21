package storage

import (
	"fmt"

	"wechat-mall-saas/internal/pkg/config"
)

// New 根据配置创建 Storage 实例
func New(cfg config.StorageConfig) (Storage, error) {
	switch cfg.Type {
	case "", "local":
		return NewLocal(cfg.Local.Path, cfg.Local.BaseURL), nil
	case "minio":
		return NewMinio(MinioOptions{
			Endpoint:   cfg.Minio.Endpoint,
			AccessKey:  cfg.Minio.AccessKey,
			SecretKey:  cfg.Minio.SecretKey,
			Bucket:     cfg.Minio.Bucket,
			UseSSL:     cfg.Minio.UseSSL,
			Region:     cfg.Minio.Region,
			BaseURL:    cfg.Minio.BaseURL,
			PublicRead: cfg.Minio.PublicRead,
		})
	default:
		return nil, fmt.Errorf("unknown storage type: %s", cfg.Type)
	}
}
