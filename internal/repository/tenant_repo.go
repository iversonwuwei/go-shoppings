package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"wechat-mall-saas/internal/model"
)

type TenantRepo struct{ db *gorm.DB }

func NewTenantRepo(db *gorm.DB) *TenantRepo { return &TenantRepo{db: db} }

func (r *TenantRepo) FindByID(ctx context.Context, id uint64) (*model.Tenant, error) {
	var t model.Tenant
	if err := r.db.WithContext(ctx).First(&t, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

func (r *TenantRepo) FindByCode(ctx context.Context, code string) (*model.Tenant, error) {
	var t model.Tenant
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&t).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

func (r *TenantRepo) Create(ctx context.Context, t *model.Tenant) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *TenantRepo) Update(ctx context.Context, t *model.Tenant) error {
	return r.db.WithContext(ctx).Save(t).Error
}

func (r *TenantRepo) UpdateFields(ctx context.Context, id uint64, fields map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&model.Tenant{}).Where("id = ?", id).Updates(fields).Error
}

// ScanExpired 扫描需要进入欠费/封禁状态的租户。
// - 状态=Active 且 plan_expire_at < now-overdueAfter => 返回（进入 Overdue）
// - 状态=Overdue 且 plan_expire_at < now-bannedAfter => 返回（进入 Banned）
// 调用方根据 target 状态分开批处理。
func (r *TenantRepo) ScanByExpireCutoff(ctx context.Context, status int8, cutoff time.Time) ([]model.Tenant, error) {
	var rows []model.Tenant
	err := r.db.WithContext(ctx).
		Where("status = ? AND plan_expire_at < ?", status, cutoff).
		Find(&rows).Error
	return rows, err
}

func (r *TenantRepo) List(ctx context.Context, status *int8, keyword string, page, size int) ([]model.Tenant, int64, error) {
	var (
		total int64
		rows  []model.Tenant
	)
	tx := r.db.WithContext(ctx).Model(&model.Tenant{})
	if status != nil {
		tx = tx.Where("status = ?", *status)
	}
	if keyword != "" {
		kw := "%" + keyword + "%"
		tx = tx.Where("company_name ILIKE ? OR code ILIKE ? OR contact_phone ILIKE ?", kw, kw, kw)
	}
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := tx.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

type AdminRepo struct{ db *gorm.DB }

func NewAdminRepo(db *gorm.DB) *AdminRepo { return &AdminRepo{db: db} }

func (r *AdminRepo) FindByUsername(ctx context.Context, username string) (*model.Admin, error) {
	var a model.Admin
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&a).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}

func (r *AdminRepo) FindByID(ctx context.Context, id uint64) (*model.Admin, error) {
	var a model.Admin
	if err := r.db.WithContext(ctx).First(&a, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}

func (r *AdminRepo) Create(ctx context.Context, a *model.Admin) error {
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *AdminRepo) UpdateLogin(ctx context.Context, id uint64, ip string) error {
	return r.db.WithContext(ctx).Model(&model.Admin{}).Where("id = ?", id).
		Updates(map[string]interface{}{"last_login_at": gorm.Expr("CURRENT_TIMESTAMP"), "last_login_ip": ip}).Error
}

// FindByPhone 在指定 tenant 作用域内按手机号查找管理员。tenantID=0 表示平台管理员。
func (r *AdminRepo) FindByPhone(ctx context.Context, tenantID uint64, phone string) (*model.Admin, error) {
	if phone == "" {
		return nil, nil
	}
	var a model.Admin
	if err := r.db.WithContext(ctx).Where("phone = ? AND tenant_id = ?", phone, tenantID).First(&a).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}

// UpdatePassword 更新管理员密码（已加密）
func (r *AdminRepo) UpdatePassword(ctx context.Context, id uint64, hash string) error {
	return r.db.WithContext(ctx).Model(&model.Admin{}).Where("id = ?", id).
		Update("password", hash).Error
}

// UpdateStatusByTenant 根据租户 ID 批量更新管理员状态（用于租户审核联动）
func (r *AdminRepo) UpdateStatusByTenant(ctx context.Context, tenantID uint64, status int8) error {
	return r.db.WithContext(ctx).Model(&model.Admin{}).Where("tenant_id = ?", tenantID).
		Update("status", status).Error
}

type TenantPlanLogRepo struct{ db *gorm.DB }

func NewTenantPlanLogRepo(db *gorm.DB) *TenantPlanLogRepo { return &TenantPlanLogRepo{db: db} }

func (r *TenantPlanLogRepo) Create(ctx context.Context, l *model.TenantPlanLog) error {
	return r.db.WithContext(ctx).Create(l).Error
}
