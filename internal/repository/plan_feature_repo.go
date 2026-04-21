package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"wechat-mall-saas/internal/model"
)

type PlanFeatureRepo struct{ db *gorm.DB }

func NewPlanFeatureRepo(db *gorm.DB) *PlanFeatureRepo { return &PlanFeatureRepo{db: db} }

func (r *PlanFeatureRepo) List(ctx context.Context, enabledOnly bool) ([]model.PlanFeature, error) {
	var out []model.PlanFeature
	q := r.db.WithContext(ctx)
	if enabledOnly {
		q = q.Where("status = 1")
	}
	err := q.Order("sort ASC, id ASC").Find(&out).Error
	return out, err
}

func (r *PlanFeatureRepo) FindByID(ctx context.Context, id uint64) (*model.PlanFeature, error) {
	var p model.PlanFeature
	if err := r.db.WithContext(ctx).First(&p, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *PlanFeatureRepo) Create(ctx context.Context, p *model.PlanFeature) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *PlanFeatureRepo) Update(ctx context.Context, p *model.PlanFeature) error {
	return r.db.WithContext(ctx).Save(p).Error
}

func (r *PlanFeatureRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.PlanFeature{}, id).Error
}
