package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"wechat-mall-saas/internal/model"
)

type AfterSaleRepo struct{ db *gorm.DB }

func NewAfterSaleRepo(db *gorm.DB) *AfterSaleRepo { return &AfterSaleRepo{db: db} }

func (r *AfterSaleRepo) DB() *gorm.DB { return r.db }

type AfterSaleListQuery struct {
	MemberID uint64
	OrderID  uint64
	Status   string
	Page     int
	Size     int
}

func (r *AfterSaleRepo) Create(ctx context.Context, row *model.AfterSaleOrder) error {
	row.TenantID = EnsureTenant(ctx)
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *AfterSaleRepo) FindByID(ctx context.Context, id uint64) (*model.AfterSaleOrder, error) {
	var row model.AfterSaleOrder
	if err := TenantDB(ctx, r.db).First(&row, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *AfterSaleRepo) FindActiveByOrderID(ctx context.Context, orderID uint64) (*model.AfterSaleOrder, error) {
	var row model.AfterSaleOrder
	err := TenantDB(ctx, r.db).
		Where("order_id = ? AND status NOT IN ?", orderID, []string{model.AfterSaleStatusRejected, model.AfterSaleStatusRefunded, model.AfterSaleStatusCancelled}).
		Order("id DESC").
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *AfterSaleRepo) List(ctx context.Context, q AfterSaleListQuery) ([]model.AfterSaleOrder, int64, error) {
	tx := TenantDB(ctx, r.db).Model(&model.AfterSaleOrder{})
	if q.MemberID > 0 {
		tx = tx.Where("member_id = ?", q.MemberID)
	}
	if q.OrderID > 0 {
		tx = tx.Where("order_id = ?", q.OrderID)
	}
	if q.Status != "" {
		tx = tx.Where("status = ?", q.Status)
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.AfterSaleOrder
	if err := tx.Order("id DESC").Offset((q.Page - 1) * q.Size).Limit(q.Size).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

type AfterSaleReasonRepo struct{ db *gorm.DB }

func NewAfterSaleReasonRepo(db *gorm.DB) *AfterSaleReasonRepo { return &AfterSaleReasonRepo{db: db} }

func (r *AfterSaleReasonRepo) DB() *gorm.DB { return r.db }

func (r *AfterSaleReasonRepo) ListAll(ctx context.Context) ([]model.AfterSaleReason, error) {
	var rows []model.AfterSaleReason
	err := r.db.WithContext(ctx).Order("sort_order ASC, id ASC").Find(&rows).Error
	return rows, err
}

func (r *AfterSaleReasonRepo) ListEnabled(ctx context.Context, reasonType string) ([]model.AfterSaleReason, error) {
	tx := r.db.WithContext(ctx).Where("enabled = 1")
	if reasonType != "" {
		tx = tx.Where("type IN ?", []string{model.AfterSaleReasonTypeAll, reasonType})
	}
	var rows []model.AfterSaleReason
	err := tx.Order("sort_order ASC, id ASC").Find(&rows).Error
	return rows, err
}

func (r *AfterSaleReasonRepo) FindByID(ctx context.Context, id uint64) (*model.AfterSaleReason, error) {
	var row model.AfterSaleReason
	if err := r.db.WithContext(ctx).First(&row, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *AfterSaleReasonRepo) Create(ctx context.Context, row *model.AfterSaleReason) error {
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *AfterSaleReasonRepo) Update(ctx context.Context, id uint64, fields map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&model.AfterSaleReason{}).Where("id = ?", id).Updates(fields).Error
}

func (r *AfterSaleReasonRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.AfterSaleReason{}).Error
}

func (r *AfterSaleReasonRepo) CountAfterSaleUsage(ctx context.Context, label string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.AfterSaleOrder{}).Where("reason = ?", label).Count(&count).Error
	return count, err
}
