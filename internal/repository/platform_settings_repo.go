package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"wechat-mall-saas/internal/model"
)

type PlatformSettingsRepo struct {
	db *gorm.DB
}

func NewPlatformSettingsRepo(db *gorm.DB) *PlatformSettingsRepo {
	return &PlatformSettingsRepo{db: db}
}

// Get 取唯一一行（id=1），不存在则返回空结构
func (r *PlatformSettingsRepo) Get(ctx context.Context) (*model.PlatformSettings, error) {
	var s model.PlatformSettings
	err := r.db.WithContext(ctx).Where("id = ?", 1).Take(&s).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &model.PlatformSettings{ID: 1}, nil
		}
		return nil, err
	}
	return &s, nil
}

// Upsert 按字段 map 覆盖更新（id=1）；如不存在则创建
func (r *PlatformSettingsRepo) Upsert(ctx context.Context, fields map[string]interface{}) error {
	fields["updated_at"] = time.Now()
	db := r.db.WithContext(ctx)
	// 存在则更新
	res := db.Model(&model.PlatformSettings{}).Where("id = ?", 1).Updates(fields)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		fields["id"] = 1
		return db.Model(&model.PlatformSettings{}).Create(fields).Error
	}
	return nil
}
