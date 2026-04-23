package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"wechat-mall-saas/internal/model"
)

type OrderRepo struct{ db *gorm.DB }

func NewOrderRepo(db *gorm.DB) *OrderRepo { return &OrderRepo{db: db} }

func (r *OrderRepo) DB() *gorm.DB { return r.db }

func (r *OrderRepo) CreateWithItems(ctx context.Context, order *model.Order, items []model.OrderItem,
	stockFn func(tx *gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		order.TenantID = EnsureTenant(ctx)
		if err := tx.Create(order).Error; err != nil {
			return err
		}
		for i := range items {
			items[i].TenantID = order.TenantID
			items[i].OrderID = order.ID
		}
		if len(items) > 0 {
			if err := tx.Create(&items).Error; err != nil {
				return err
			}
		}
		if stockFn != nil {
			if err := stockFn(tx); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *OrderRepo) FindByID(ctx context.Context, id uint64) (*model.Order, error) {
	var o model.Order
	if err := TenantDB(ctx, r.db).Preload("Items").First(&o, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &o, nil
}

func (r *OrderRepo) FindByNo(ctx context.Context, no string) (*model.Order, error) {
	var o model.Order
	if err := TenantDB(ctx, r.db).Where("order_no = ?", no).Preload("Items").First(&o).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &o, nil
}

type OrderListQuery struct {
	MemberID uint64
	Status   string
	Page     int
	Size     int
}

func (r *OrderRepo) List(ctx context.Context, q OrderListQuery) ([]model.Order, int64, error) {
	tx := TenantDB(ctx, r.db).Model(&model.Order{})
	if q.MemberID > 0 {
		tx = tx.Where("member_id = ?", q.MemberID)
	}
	if q.Status != "" {
		tx = tx.Where("status = ?", q.Status)
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.Order
	if err := tx.Order("id DESC").Offset((q.Page - 1) * q.Size).Limit(q.Size).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *OrderRepo) UpdateStatus(ctx context.Context, id uint64, from, to string, setAt map[string]interface{}) error {
	fields := map[string]interface{}{"status": to}
	for k, v := range setAt {
		fields[k] = v
	}
	res := TenantDB(ctx, r.db).Model(&model.Order{}).Where("id = ? AND status = ?", id, from).Updates(fields)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("order status mismatch")
	}
	return nil
}

func (r *OrderRepo) CountMonth(ctx context.Context) (int64, error) {
	var n int64
	start := time.Now().AddDate(0, 0, -time.Now().Day()+1)
	err := TenantDB(ctx, r.db).Model(&model.Order{}).Where("created_at >= ?", start).Count(&n).Error
	return n, err
}

type OrderLogRepo struct{ db *gorm.DB }

func NewOrderLogRepo(db *gorm.DB) *OrderLogRepo { return &OrderLogRepo{db: db} }

func (r *OrderLogRepo) Create(ctx context.Context, l *model.OrderLog) error {
	if l.TenantID == 0 {
		l.TenantID = EnsureTenant(ctx)
	}
	return r.db.WithContext(ctx).Create(l).Error
}

func (r *OrderLogRepo) ListByOrder(ctx context.Context, orderID uint64) ([]model.OrderLog, error) {
	var rows []model.OrderLog
	err := TenantDB(ctx, r.db).Where("order_id = ?", orderID).Order("id ASC").Find(&rows).Error
	return rows, err
}

type OrderMessageRepo struct{ db *gorm.DB }

func NewOrderMessageRepo(db *gorm.DB) *OrderMessageRepo { return &OrderMessageRepo{db: db} }

type OrderMessageListQuery struct {
	Status string
	Page   int
	Size   int
}

func (r *OrderMessageRepo) Create(ctx context.Context, m *model.OrderMessage) error {
	if m.TenantID == 0 {
		m.TenantID = EnsureTenant(ctx)
	}
	if m.Status == "" {
		m.Status = model.OrderMessageStatusUnread
	}
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *OrderMessageRepo) List(ctx context.Context, q OrderMessageListQuery) ([]model.OrderMessage, int64, error) {
	tx := TenantDB(ctx, r.db).Model(&model.OrderMessage{})
	if q.Status != "" {
		tx = tx.Where("status = ?", q.Status)
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.OrderMessage
	if err := tx.Order("id DESC").Offset((q.Page - 1) * q.Size).Limit(q.Size).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *OrderMessageRepo) CountUnread(ctx context.Context) (int64, error) {
	var total int64
	err := TenantDB(ctx, r.db).Model(&model.OrderMessage{}).Where("status = ?", model.OrderMessageStatusUnread).Count(&total).Error
	return total, err
}

func (r *OrderMessageRepo) MarkRead(ctx context.Context, id uint64) error {
	now := time.Now()
	res := TenantDB(ctx, r.db).Model(&model.OrderMessage{}).
		Where("id = ? AND status = ?", id, model.OrderMessageStatusUnread).
		Updates(map[string]interface{}{"status": model.OrderMessageStatusRead, "read_at": now})
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (r *OrderMessageRepo) MarkAllRead(ctx context.Context) error {
	now := time.Now()
	return TenantDB(ctx, r.db).Model(&model.OrderMessage{}).
		Where("status = ?", model.OrderMessageStatusUnread).
		Updates(map[string]interface{}{"status": model.OrderMessageStatusRead, "read_at": now}).Error
}
