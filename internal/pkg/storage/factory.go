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
	case "supabase":
		return NewSupabase(SupabaseOptions{
			ProjectURL:       cfg.Supabase.ProjectURL,
			ServiceRoleKey:   cfg.Supabase.ServiceRoleKey,
			Bucket:           cfg.Supabase.Bucket,
			PublicRead:       cfg.Supabase.PublicRead,
			BaseURL:          cfg.Supabase.BaseURL,
			SignedURLExpires: cfg.Supabase.SignedURLExpires,
			CreateBucket:     cfg.Supabase.CreateBucket,
			Upsert:           cfg.Supabase.Upsert,
		})
	default:
		return nil, fmt.Errorf("unknown storage type: %s", cfg.Type)
	}
}
