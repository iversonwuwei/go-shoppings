package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"wechat-mall-saas/internal/model"
)

type ApiTokenRepo struct{ db *gorm.DB }

func NewApiTokenRepo(db *gorm.DB) *ApiTokenRepo { return &ApiTokenRepo{db: db} }

func (r *ApiTokenRepo) List(ctx context.Context) ([]model.ApiToken, error) {
	var rows []model.ApiToken
	err := TenantDB(ctx, r.db).Order("id DESC").Find(&rows).Error
	return rows, err
}

func (r *ApiTokenRepo) Create(ctx context.Context, t *model.ApiToken) error {
	t.TenantID = EnsureTenant(ctx)
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *ApiTokenRepo) Update(ctx context.Context, id uint64, fields map[string]interface{}) error {
	fields["updated_at"] = time.Now()
	return TenantDB(ctx, r.db).Model(&model.ApiToken{}).
		Where("id = ?", id).Updates(fields).Error
}

func (r *ApiTokenRepo) Delete(ctx context.Context, id uint64) error {
	return TenantDB(ctx, r.db).Delete(&model.ApiToken{}, id).Error
}

func (r *ApiTokenRepo) Find(ctx context.Context, id uint64) (*model.ApiToken, error) {
	var t model.ApiToken
	err := TenantDB(ctx, r.db).First(&t, id).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *ApiTokenRepo) ListLogs(ctx context.Context, tokenID uint64, page, size int) ([]model.ApiRequestLog, int64, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 200 {
		size = 20
	}
	q := TenantDB(ctx, r.db).Model(&model.ApiRequestLog{})
	if tokenID > 0 {
		q = q.Where("token_id = ?", tokenID)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.ApiRequestLog
	err := q.Order("id DESC").Offset((page - 1) * size).Limit(size).Find(&rows).Error
	return rows, total, err
}
