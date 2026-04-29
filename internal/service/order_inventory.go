package service

import (
	"context"

	"gorm.io/gorm"

	"wechat-mall-saas/internal/model"
)

type orderStockItem struct {
	ProductID uint64
	SKUID     uint64
	Quantity  int
	IsVirtual int8
}

func restoreOrderStockTx(ctx context.Context, tx *gorm.DB, tenantID, orderID uint64) error {
	var order model.Order
	if err := tx.WithContext(ctx).
		Select("id", "is_virtual").
		Where("tenant_id = ? AND id = ?", tenantID, orderID).
		First(&order).Error; err != nil {
		return err
	}
	if order.IsVirtual == 1 {
		return nil
	}

	var items []orderStockItem
	if err := tx.WithContext(ctx).
		Table("order_items AS oi").
		Select("oi.product_id, oi.sku_id, oi.quantity, COALESCE(p.is_virtual, 0) AS is_virtual").
		Joins("LEFT JOIN products AS p ON p.tenant_id = oi.tenant_id AND p.id = oi.product_id").
		Where("oi.tenant_id = ? AND oi.order_id = ?", tenantID, orderID).
		Scan(&items).Error; err != nil {
		return err
	}

	for _, item := range items {
		if item.Quantity <= 0 || item.IsVirtual == 1 {
			continue
		}
		if item.SKUID > 0 {
			if err := tx.WithContext(ctx).
				Model(&model.ProductSKU{}).
				Where("tenant_id = ? AND id = ?", tenantID, item.SKUID).
				UpdateColumn("stock", gorm.Expr("stock + ?", item.Quantity)).Error; err != nil {
				return err
			}
			continue
		}
		if err := tx.WithContext(ctx).
			Model(&model.Product{}).
			Where("tenant_id = ? AND id = ?", tenantID, item.ProductID).
			UpdateColumn("stock", gorm.Expr("stock + ?", item.Quantity)).Error; err != nil {
			return err
		}
	}
	return nil
}
