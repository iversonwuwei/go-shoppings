package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"wechat-mall-saas/internal/model"
)

type SmsRepo struct{ db *gorm.DB }

func NewSmsRepo(db *gorm.DB) *SmsRepo { return &SmsRepo{db: db} }

// --------- Settings ---------

func (r *SmsRepo) GetSettings(ctx context.Context) (*model.SmsSetting, error) {
	tid := EnsureTenant(ctx)
	if tid == 0 {
		return nil, nil
	}
	var s model.SmsSetting
	if err := r.db.WithContext(ctx).Where("tenant_id = ?", tid).First(&s).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *SmsRepo) UpsertSettings(ctx context.Context, s *model.SmsSetting) error {
	s.TenantID = EnsureTenant(ctx)
	s.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(s).Error
}

// --------- Templates ---------

func (r *SmsRepo) ListTemplates(ctx context.Context) ([]model.SmsTemplate, error) {
	var rows []model.SmsTemplate
	err := TenantDB(ctx, r.db).Order("id DESC").Find(&rows).Error
	return rows, err
}

func (r *SmsRepo) CreateTemplate(ctx context.Context, t *model.SmsTemplate) error {
	t.TenantID = EnsureTenant(ctx)
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *SmsRepo) UpdateTemplate(ctx context.Context, id uint64, fields map[string]interface{}) error {
	fields["updated_at"] = time.Now()
	return TenantDB(ctx, r.db).Model(&model.SmsTemplate{}).
		Where("id = ?", id).Updates(fields).Error
}

func (r *SmsRepo) DeleteTemplate(ctx context.Context, id uint64) error {
	return TenantDB(ctx, r.db).Delete(&model.SmsTemplate{}, id).Error
}

// --------- Logs ---------

func (r *SmsRepo) ListLogs(ctx context.Context, phone string, page, size int) ([]model.SmsLog, int64, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 200 {
		size = 20
	}
	q := TenantDB(ctx, r.db).Model(&model.SmsLog{})
	if phone != "" {
		q = q.Where("phone = ?", phone)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.SmsLog
	err := q.Order("id DESC").Offset((page - 1) * size).Limit(size).Find(&rows).Error
	return rows, total, err
}
