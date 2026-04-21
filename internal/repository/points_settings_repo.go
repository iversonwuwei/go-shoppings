package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"wechat-mall-saas/internal/model"
)

type PointsSettingsRepo struct{ db *gorm.DB }

func NewPointsSettingsRepo(db *gorm.DB) *PointsSettingsRepo {
	return &PointsSettingsRepo{db: db}
}

// Get 返回当前租户的积分规则；若不存在返回 nil。
func (r *PointsSettingsRepo) Get(ctx context.Context) (*model.PointsSetting, error) {
	tid := EnsureTenant(ctx)
	if tid == 0 {
		return nil, nil
	}
	var ps model.PointsSetting
	if err := r.db.WithContext(ctx).Where("tenant_id = ?", tid).First(&ps).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &ps, nil
}

// Upsert 更新或插入当前租户的积分规则。
func (r *PointsSettingsRepo) Upsert(ctx context.Context, s *model.PointsSetting) error {
	s.TenantID = EnsureTenant(ctx)
	s.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(s).Error
}
