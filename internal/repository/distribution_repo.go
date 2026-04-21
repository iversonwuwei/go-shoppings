package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"wechat-mall-saas/internal/model"
)

type DistributionRepo struct{ db *gorm.DB }

func NewDistributionRepo(db *gorm.DB) *DistributionRepo { return &DistributionRepo{db: db} }

// --------- Settings ---------

func (r *DistributionRepo) GetSettings(ctx context.Context) (*model.DistributionSetting, error) {
	tid := EnsureTenant(ctx)
	if tid == 0 {
		return nil, nil
	}
	var s model.DistributionSetting
	if err := r.db.WithContext(ctx).Where("tenant_id = ?", tid).First(&s).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *DistributionRepo) UpsertSettings(ctx context.Context, s *model.DistributionSetting) error {
	s.TenantID = EnsureTenant(ctx)
	s.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(s).Error
}

// --------- Distributors ---------

func (r *DistributionRepo) ListDistributors(ctx context.Context, status int8, page, size int) ([]model.Distributor, int64, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 200 {
		size = 20
	}
	q := TenantDB(ctx, r.db).Model(&model.Distributor{})
	if status >= 0 {
		q = q.Where("status = ?", status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.Distributor
	err := q.Order("id DESC").Offset((page - 1) * size).Limit(size).Find(&rows).Error
	return rows, total, err
}

func (r *DistributionRepo) FindDistributor(ctx context.Context, id uint64) (*model.Distributor, error) {
	var d model.Distributor
	if err := TenantDB(ctx, r.db).First(&d, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &d, nil
}

func (r *DistributionRepo) UpdateDistributor(ctx context.Context, id uint64, fields map[string]interface{}) error {
	return TenantDB(ctx, r.db).Model(&model.Distributor{}).
		Where("id = ?", id).Updates(fields).Error
}

// --------- Commission Logs ---------

func (r *DistributionRepo) ListCommissions(ctx context.Context, distributorID uint64, page, size int) ([]model.CommissionLog, int64, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 200 {
		size = 20
	}
	q := TenantDB(ctx, r.db).Model(&model.CommissionLog{})
	if distributorID > 0 {
		q = q.Where("distributor_id = ?", distributorID)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.CommissionLog
	err := q.Order("id DESC").Offset((page - 1) * size).Limit(size).Find(&rows).Error
	return rows, total, err
}
