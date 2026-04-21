package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"wechat-mall-saas/internal/model"
)

// TenantSubscriptionOrderRepo 租户订阅订单仓储（平台侧表，不附加 tenant_id 过滤）
type TenantSubscriptionOrderRepo struct{ db *gorm.DB }

func NewTenantSubscriptionOrderRepo(db *gorm.DB) *TenantSubscriptionOrderRepo {
	return &TenantSubscriptionOrderRepo{db: db}
}

func (r *TenantSubscriptionOrderRepo) Create(ctx context.Context, o *model.TenantSubscriptionOrder) error {
	return r.db.WithContext(ctx).Create(o).Error
}

func (r *TenantSubscriptionOrderRepo) FindByOrderNo(ctx context.Context, orderNo string) (*model.TenantSubscriptionOrder, error) {
	var o model.TenantSubscriptionOrder
	err := r.db.WithContext(ctx).Where("order_no = ?", orderNo).First(&o).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &o, err
}

func (r *TenantSubscriptionOrderRepo) FindByID(ctx context.Context, id uint64) (*model.TenantSubscriptionOrder, error) {
	var o model.TenantSubscriptionOrder
	err := r.db.WithContext(ctx).First(&o, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &o, err
}

func (r *TenantSubscriptionOrderRepo) MarkPaid(ctx context.Context, orderNo, txnID string, paidAt time.Time, expireAfter time.Time) error {
	return r.db.WithContext(ctx).
		Model(&model.TenantSubscriptionOrder{}).
		Where("order_no = ? AND status = ?", orderNo, 0).
		Updates(map[string]interface{}{
			"status":             1,
			"pay_transaction_id": txnID,
			"pay_at":             paidAt,
			"expire_after":       expireAfter,
			"updated_at":         time.Now(),
		}).Error
}

func (r *TenantSubscriptionOrderRepo) ListByTenant(ctx context.Context, tenantID uint64, page, pageSize int) ([]model.TenantSubscriptionOrder, int64, error) {
	var rows []model.TenantSubscriptionOrder
	var total int64
	q := r.db.WithContext(ctx).Model(&model.TenantSubscriptionOrder{}).Where("tenant_id = ?", tenantID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&rows).Error
	return rows, total, err
}
