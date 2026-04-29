package repository

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	"wechat-mall-saas/internal/model"
)

type CartRepo struct{ db *gorm.DB }

func NewCartRepo(db *gorm.DB) *CartRepo { return &CartRepo{db: db} }

type CartKeyPair struct {
	ProductID uint64
	SKUID     uint64
}

type CartListQuery struct {
	MemberID uint64
	Keys     []CartKeyPair
	Page     int
	Size     int
}

func (r *CartRepo) List(ctx context.Context, q CartListQuery) ([]model.MemberCartItem, int64, error) {
	var rows []model.MemberCartItem
	tx := TenantDB(ctx, r.db).
		Model(&model.MemberCartItem{}).
		Where("member_id = ?", q.MemberID)
	if len(q.Keys) > 0 {
		parts := make([]string, 0, len(q.Keys))
		args := make([]interface{}, 0, len(q.Keys)*2)
		for _, key := range q.Keys {
			parts = append(parts, "(product_id = ? AND sku_id = ?)")
			args = append(args, key.ProductID, key.SKUID)
		}
		tx = tx.Where(strings.Join(parts, " OR "), args...)
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := tx.
		Order("updated_at DESC, id DESC").
		Offset((q.Page - 1) * q.Size).
		Limit(q.Size).
		Find(&rows).Error
	return rows, total, err
}

func (r *CartRepo) Find(ctx context.Context, memberID, productID, skuID uint64) (*model.MemberCartItem, error) {
	var item model.MemberCartItem
	err := TenantDB(ctx, r.db).
		Where("member_id = ? AND product_id = ? AND sku_id = ?", memberID, productID, skuID).
		First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *CartRepo) SaveQuantity(ctx context.Context, item *model.MemberCartItem) error {
	item.TenantID = EnsureTenant(ctx)
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *CartRepo) Delete(ctx context.Context, memberID, productID, skuID uint64) error {
	return TenantDB(ctx, r.db).
		Where("member_id = ? AND product_id = ? AND sku_id = ?", memberID, productID, skuID).
		Delete(&model.MemberCartItem{}).Error
}

func (r *CartRepo) Clear(ctx context.Context, memberID uint64) error {
	return TenantDB(ctx, r.db).
		Where("member_id = ?", memberID).
		Delete(&model.MemberCartItem{}).Error
}
