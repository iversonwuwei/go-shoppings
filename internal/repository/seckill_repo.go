package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"wechat-mall-saas/internal/model"
)

type SeckillRepo struct{ db *gorm.DB }

func NewSeckillRepo(db *gorm.DB) *SeckillRepo { return &SeckillRepo{db: db} }

// ListActivities 返回当前租户的所有活动（按 start_at DESC），预加载商品。
func (r *SeckillRepo) ListActivities(ctx context.Context) ([]model.SeckillActivity, error) {
	var rows []model.SeckillActivity
	err := TenantDB(ctx, r.db).
		Preload("Products").
		Order("start_at DESC, id DESC").
		Find(&rows).Error
	return rows, err
}

// ListActive 返回当前时间段内 status=1 的活动（会员端）。
func (r *SeckillRepo) ListActive(ctx context.Context) ([]model.SeckillActivity, error) {
	var rows []model.SeckillActivity
	now := time.Now()
	err := TenantDB(ctx, r.db).
		Preload("Products").
		Where("status = 1 AND start_at <= ? AND end_at >= ?", now, now).
		Order("start_at ASC").
		Find(&rows).Error
	return rows, err
}

func (r *SeckillRepo) FindActivity(ctx context.Context, id uint64) (*model.SeckillActivity, error) {
	var a model.SeckillActivity
	if err := TenantDB(ctx, r.db).Preload("Products").First(&a, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}

// CreateActivity 在一个事务中创建活动及其商品。
func (r *SeckillRepo) CreateActivity(ctx context.Context, a *model.SeckillActivity, products []model.SeckillProduct) error {
	a.TenantID = EnsureTenant(ctx)
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(a).Error; err != nil {
			return err
		}
		for i := range products {
			products[i].TenantID = a.TenantID
			products[i].SeckillID = a.ID
			products[i].SoldCount = 0
		}
		if len(products) > 0 {
			if err := tx.Create(&products).Error; err != nil {
				return err
			}
		}
		a.Products = products
		return nil
	})
}

// UpdateActivity 更新活动头并重建商品清单。
func (r *SeckillRepo) UpdateActivity(ctx context.Context, a *model.SeckillActivity, products []model.SeckillProduct) error {
	tid := EnsureTenant(ctx)
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.SeckillActivity{}).
			Where("id = ? AND tenant_id = ?", a.ID, tid).
			Updates(map[string]interface{}{
				"name":        a.Name,
				"start_at":    a.StartAt,
				"end_at":      a.EndAt,
				"per_limit":   a.PerLimit,
				"total_stock": a.TotalStock,
				"status":      a.Status,
			}).Error; err != nil {
			return err
		}
		// 简化实现：删除旧商品后重建；已售数量重置为 0
		if err := tx.Where("seckill_id = ? AND tenant_id = ?", a.ID, tid).
			Delete(&model.SeckillProduct{}).Error; err != nil {
			return err
		}
		for i := range products {
			products[i].TenantID = tid
			products[i].SeckillID = a.ID
			products[i].SoldCount = 0
		}
		if len(products) > 0 {
			if err := tx.Create(&products).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *SeckillRepo) DeleteActivity(ctx context.Context, id uint64) error {
	tid := EnsureTenant(ctx)
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("seckill_id = ? AND tenant_id = ?", id, tid).
			Delete(&model.SeckillProduct{}).Error; err != nil {
			return err
		}
		return tx.Where("id = ? AND tenant_id = ?", id, tid).
			Delete(&model.SeckillActivity{}).Error
	})
}

// ClaimStock 原子扣减秒杀库存，失败（库存不足）返回 false。
// 使用 Postgres 的条件 UPDATE，避免超卖。
func (r *SeckillRepo) ClaimStock(ctx context.Context, seckillProductID uint64, qty int) (bool, error) {
	tid := EnsureTenant(ctx)
	res := r.db.WithContext(ctx).Exec(
		`UPDATE seckill_products SET sold_count = sold_count + ?
		 WHERE id = ? AND tenant_id = ? AND stock - sold_count >= ?`,
		qty, seckillProductID, tid, qty,
	)
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}
