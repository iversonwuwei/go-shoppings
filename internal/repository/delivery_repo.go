package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"wechat-mall-saas/internal/model"
)

type DeliveryRepo struct{ db *gorm.DB }

func NewDeliveryRepo(db *gorm.DB) *DeliveryRepo { return &DeliveryRepo{db: db} }

func (r *DeliveryRepo) Get(ctx context.Context) (*model.DeliverySetting, error) {
	tid := EnsureTenant(ctx)
	if tid == 0 {
		return nil, nil
	}
	var s model.DeliverySetting
	if err := r.db.WithContext(ctx).Where("tenant_id = ?", tid).First(&s).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *DeliveryRepo) Upsert(ctx context.Context, s *model.DeliverySetting) error {
	s.TenantID = EnsureTenant(ctx)
	s.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(s).Error
}

type SiteConfigRepo struct{ db *gorm.DB }

func NewSiteConfigRepo(db *gorm.DB) *SiteConfigRepo { return &SiteConfigRepo{db: db} }

func (r *SiteConfigRepo) Get(ctx context.Context) (*model.TenantSiteConfig, error) {
	tid := EnsureTenant(ctx)
	if tid == 0 {
		return nil, nil
	}
	var s model.TenantSiteConfig
	if err := r.db.WithContext(ctx).Where("tenant_id = ?", tid).First(&s).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *SiteConfigRepo) Upsert(ctx context.Context, s *model.TenantSiteConfig) error {
	s.TenantID = EnsureTenant(ctx)
	s.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(s).Error
}
