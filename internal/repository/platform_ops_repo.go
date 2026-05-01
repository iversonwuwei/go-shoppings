package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"wechat-mall-saas/internal/model"
)

// =================== 平台 SMS ===================
// 全局网关配置：固定使用 tenant_id = 0 的单行存储，仅平台管理员读写

func (r *SmsRepo) PlatformGetGlobalSettings(ctx context.Context) (*model.SmsSetting, error) {
	var s model.SmsSetting
	if err := r.db.WithContext(ctx).Where("tenant_id = 0").First(&s).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *SmsRepo) PlatformUpsertGlobalSettings(ctx context.Context, s *model.SmsSetting) error {
	now := time.Now()
	row := &model.SmsSetting{
		TenantID:     0,
		Enabled:      s.Enabled,
		Provider:     s.Provider,
		AccessKey:    s.AccessKey,
		AccessSecret: s.AccessSecret,
		SignName:     s.SignName,
		Region:       s.Region,
		Remark:       s.Remark,
		UpdatedAt:    now,
	}
	if err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "tenant_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"enabled", "provider", "access_key", "access_secret", "sign_name", "region", "remark", "updated_at"}),
	}).Create(row).Error; err != nil {
		return err
	}
	s.TenantID = 0
	s.UpdatedAt = now
	return nil
}

func (r *SmsRepo) PlatformListTemplates(ctx context.Context) ([]model.SmsTemplate, error) {
	var rows []model.SmsTemplate
	err := r.db.WithContext(ctx).Where("tenant_id = 0").Order("id DESC").Find(&rows).Error
	return rows, err
}

func (r *SmsRepo) PlatformCreateTemplate(ctx context.Context, t *model.SmsTemplate) error {
	t.TenantID = 0
	var existing model.SmsTemplate
	err := r.db.WithContext(ctx).Where("tenant_id = 0 AND code = ?", t.Code).Order("id DESC").First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return r.db.WithContext(ctx).Create(t).Error
	}
	if err != nil {
		return err
	}
	fields := map[string]interface{}{
		"name":        t.Name,
		"template_id": t.TemplateID,
		"content":     t.Content,
		"enabled":     t.Enabled,
		"updated_at":  time.Now(),
	}
	if err := r.db.WithContext(ctx).Model(&model.SmsTemplate{}).Where("id = ? AND tenant_id = 0", existing.ID).Updates(fields).Error; err != nil {
		return err
	}
	t.ID = existing.ID
	t.CreatedAt = existing.CreatedAt
	t.UpdatedAt = time.Now()
	return nil
}

func (r *SmsRepo) PlatformUpdateTemplateAny(ctx context.Context, id uint64, fields map[string]interface{}) error {
	fields["updated_at"] = time.Now()
	return r.db.WithContext(ctx).Model(&model.SmsTemplate{}).
		Where("id = ? AND tenant_id = 0", id).Updates(fields).Error
}

func (r *SmsRepo) PlatformDeleteTemplateAny(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Where("tenant_id = 0").Delete(&model.SmsTemplate{}, id).Error
}

// 平台：仅列平台自身短信日志，可按 phone 过滤。
func (r *SmsRepo) PlatformListLogs(ctx context.Context, phone string, page, size int) ([]model.SmsLog, int64, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 200 {
		size = 20
	}
	q := r.db.WithContext(ctx).Model(&model.SmsLog{}).Where("tenant_id = 0")
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

// =================== 平台 API Access ===================

func (r *ApiTokenRepo) PlatformList(ctx context.Context, tenantID uint64) ([]model.ApiToken, error) {
	q := r.db.WithContext(ctx).Model(&model.ApiToken{})
	if tenantID > 0 {
		q = q.Where("tenant_id = ?", tenantID)
	}
	var rows []model.ApiToken
	err := q.Order("id DESC").Find(&rows).Error
	return rows, err
}

func (r *ApiTokenRepo) PlatformCreate(ctx context.Context, tenantID uint64, t *model.ApiToken) error {
	t.TenantID = tenantID
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *ApiTokenRepo) PlatformUpdate(ctx context.Context, id uint64, fields map[string]interface{}) error {
	fields["updated_at"] = time.Now()
	return r.db.WithContext(ctx).Model(&model.ApiToken{}).
		Where("id = ?", id).Updates(fields).Error
}

func (r *ApiTokenRepo) PlatformDelete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.ApiToken{}, id).Error
}

func (r *ApiTokenRepo) PlatformFind(ctx context.Context, id uint64) (*model.ApiToken, error) {
	var t model.ApiToken
	if err := r.db.WithContext(ctx).First(&t, id).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *ApiTokenRepo) PlatformListLogs(ctx context.Context, tenantID, tokenID uint64, page, size int) ([]model.ApiRequestLog, int64, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 200 {
		size = 20
	}
	q := r.db.WithContext(ctx).Model(&model.ApiRequestLog{})
	if tenantID > 0 {
		q = q.Where("tenant_id = ?", tenantID)
	}
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

// =================== 平台 站点配置（域名 / 部署） ===================

// ListWithDomain 列出所有绑定自定义域名的租户站点
func (r *SiteConfigRepo) PlatformListWithDomain(ctx context.Context) ([]model.TenantSiteConfig, error) {
	var rows []model.TenantSiteConfig
	err := r.db.WithContext(ctx).
		Where("custom_domain <> ''").
		Order("updated_at DESC").
		Find(&rows).Error
	return rows, err
}

// ListByDeploymentMode mode=""时返回全部
func (r *SiteConfigRepo) PlatformListDeployments(ctx context.Context, mode string) ([]model.TenantSiteConfig, error) {
	q := r.db.WithContext(ctx).Model(&model.TenantSiteConfig{})
	if mode != "" {
		q = q.Where("deployment_mode = ?", mode)
	}
	var rows []model.TenantSiteConfig
	err := q.Order("tenant_id ASC").Find(&rows).Error
	return rows, err
}

func (r *SiteConfigRepo) PlatformFindByTenantID(ctx context.Context, tid uint64) (*model.TenantSiteConfig, error) {
	var s model.TenantSiteConfig
	if err := r.db.WithContext(ctx).Where("tenant_id = ?", tid).First(&s).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *SiteConfigRepo) PlatformUpdateByTenantID(ctx context.Context, tid uint64, fields map[string]interface{}) error {
	fields["updated_at"] = time.Now()
	return r.db.WithContext(ctx).Model(&model.TenantSiteConfig{}).
		Where("tenant_id = ?", tid).Updates(fields).Error
}

// PlatformInsert 平台为指定租户直接创建站点配置行
func (r *SiteConfigRepo) PlatformInsert(ctx context.Context, s *model.TenantSiteConfig) error {
	if s.UpdatedAt.IsZero() {
		s.UpdatedAt = time.Now()
	}
	return r.db.WithContext(ctx).Create(s).Error
}
