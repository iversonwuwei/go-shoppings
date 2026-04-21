package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"wechat-mall-saas/internal/model"
)

// ========== Payment Config (租户提交 + 平台审核) ==========

type PaymentConfigRepo struct{ db *gorm.DB }

func NewPaymentConfigRepo(db *gorm.DB) *PaymentConfigRepo { return &PaymentConfigRepo{db: db} }

func (r *PaymentConfigRepo) DB() *gorm.DB { return r.db }

func (r *PaymentConfigRepo) FindByTenantProvider(ctx context.Context, tenantID uint64, provider string) (*model.TenantPaymentConfig, error) {
	var m model.TenantPaymentConfig
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND provider = ?", tenantID, provider).
		First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *PaymentConfigRepo) ListByTenant(ctx context.Context, tenantID uint64) ([]model.TenantPaymentConfig, error) {
	var rows []model.TenantPaymentConfig
	err := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("id ASC").
		Find(&rows).Error
	return rows, err
}

func (r *PaymentConfigRepo) Upsert(ctx context.Context, m *model.TenantPaymentConfig) error {
	now := time.Now()
	m.SubmittedAt = &now
	m.AuditStatus = model.ConfigAuditPending
	m.Enabled = 0
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "tenant_id"}, {Name: "provider"}},
		DoUpdates: clause.AssignmentColumns([]string{"mch_id", "app_id", "api_v3_key", "cert_serial_no", "private_key_pem", "cert_pem", "notify_url", "submitted_at", "audit_status", "audit_remark", "enabled", "updated_at"}),
	}).Create(m).Error
}

func (r *PaymentConfigRepo) Audit(ctx context.Context, id uint64, approve bool, remark string) error {
	now := time.Now()
	fields := map[string]interface{}{
		"audit_status": map[bool]int8{true: model.ConfigAuditApproved, false: model.ConfigAuditRejected}[approve],
		"audit_remark": remark,
		"audited_at":   &now,
	}
	if approve {
		fields["enabled"] = int8(1)
	}
	return r.db.WithContext(ctx).Model(&model.TenantPaymentConfig{}).
		Where("id = ?", id).Updates(fields).Error
}

func (r *PaymentConfigRepo) ListForAudit(ctx context.Context, status *int8, page, size int) ([]model.TenantPaymentConfig, int64, error) {
	var total int64
	var rows []model.TenantPaymentConfig
	tx := r.db.WithContext(ctx).Model(&model.TenantPaymentConfig{})
	if status != nil {
		tx = tx.Where("audit_status = ?", *status)
	}
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	err := tx.Order("submitted_at DESC NULLS LAST, id DESC").
		Offset((page - 1) * size).Limit(size).Find(&rows).Error
	return rows, total, err
}

// ========== Shipping Carrier（平台统一维护，租户只读 + 轨迹查询） ==========

type ShippingCarrierRepo struct{ db *gorm.DB }

func NewShippingCarrierRepo(db *gorm.DB) *ShippingCarrierRepo { return &ShippingCarrierRepo{db: db} }

func (r *ShippingCarrierRepo) DB() *gorm.DB { return r.db }

// ListAll 平台端列出全部承运商
func (r *ShippingCarrierRepo) ListAll(ctx context.Context) ([]model.ShippingCarrier, error) {
	var rows []model.ShippingCarrier
	err := r.db.WithContext(ctx).Order("priority DESC, id ASC").Find(&rows).Error
	return rows, err
}

// ListEnabled 租户端可用的承运商（已启用）
func (r *ShippingCarrierRepo) ListEnabled(ctx context.Context) ([]model.ShippingCarrier, error) {
	var rows []model.ShippingCarrier
	err := r.db.WithContext(ctx).
		Where("enabled = 1").
		Order("priority DESC, id ASC").Find(&rows).Error
	return rows, err
}

func (r *ShippingCarrierRepo) FindByID(ctx context.Context, id uint64) (*model.ShippingCarrier, error) {
	var m model.ShippingCarrier
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

// Create 平台创建：默认启用、审核通过
func (r *ShippingCarrierRepo) Create(ctx context.Context, m *model.ShippingCarrier) error {
	now := time.Now()
	m.SubmittedAt = &now
	m.AuditedAt = &now
	m.AuditStatus = model.ConfigAuditApproved
	if m.Enabled == 0 {
		m.Enabled = 1
	}
	return r.db.WithContext(ctx).Create(m).Error
}

// Update 平台更新
func (r *ShippingCarrierRepo) Update(ctx context.Context, id uint64, fields map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&model.ShippingCarrier{}).
		Where("id = ?", id).Updates(fields).Error
}

func (r *ShippingCarrierRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&model.ShippingCarrier{}).Error
}

// FindEnabledByCode 按编码查询已启用承运商
func (r *ShippingCarrierRepo) FindEnabledByCode(ctx context.Context, code string) (*model.ShippingCarrier, error) {
	var m model.ShippingCarrier
	err := r.db.WithContext(ctx).
		Where("code = ? AND enabled = 1", code).
		First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}
