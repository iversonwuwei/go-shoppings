package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"wechat-mall-saas/internal/model"
)

type PlanRepo struct{ db *gorm.DB }

func NewPlanRepo(db *gorm.DB) *PlanRepo { return &PlanRepo{db: db} }

func (r *PlanRepo) List(ctx context.Context) ([]model.Plan, error) {
	var out []model.Plan
	err := r.db.WithContext(ctx).Order("monthly_fee ASC").Find(&out).Error
	return out, err
}

func (r *PlanRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.Plan{}, id).Error
}

func (r *PlanRepo) FindByID(ctx context.Context, id uint64) (*model.Plan, error) {
	var p model.Plan
	if err := r.db.WithContext(ctx).First(&p, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *PlanRepo) FindDefault(ctx context.Context) (*model.Plan, error) {
	var p model.Plan
	if err := r.db.WithContext(ctx).Where("is_default = 1").First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *PlanRepo) Create(ctx context.Context, p *model.Plan) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *PlanRepo) Update(ctx context.Context, p *model.Plan) error {
	return r.db.WithContext(ctx).Save(p).Error
}
