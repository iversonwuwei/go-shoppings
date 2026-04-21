package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// SeckillActivity 秒杀活动（时间段 + 限购数量 + 总库存）
type SeckillActivity struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID   uint64    `gorm:"not null;index" json:"tenant_id"`
	Name       string    `gorm:"size:100;not null" json:"name"`
	StartAt    time.Time `gorm:"not null" json:"start_at"`
	EndAt      time.Time `gorm:"not null" json:"end_at"`
	PerLimit   int       `gorm:"not null;default:1" json:"per_limit"`
	TotalStock int       `gorm:"not null" json:"total_stock"`
	Status     int8      `gorm:"not null;default:1" json:"status"` // 1启用 0停用
	CreatedAt  time.Time `json:"created_at"`

	Products []SeckillProduct `gorm:"foreignKey:SeckillID" json:"products,omitempty"`
}

func (SeckillActivity) TableName() string { return "seckill_activities" }

// SeckillProduct 活动内的秒杀商品/SKU
type SeckillProduct struct {
	ID           uint64          `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID     uint64          `gorm:"not null;index" json:"tenant_id"`
	SeckillID    uint64          `gorm:"column:seckill_id;not null;index" json:"seckill_id"`
	ProductID    uint64          `gorm:"not null" json:"product_id"`
	SKUID        uint64          `gorm:"column:sku_id" json:"sku_id"`
	SeckillPrice decimal.Decimal `gorm:"type:numeric(10,2);not null" json:"seckill_price"`
	Stock        int             `gorm:"not null" json:"stock"`
	SoldCount    int             `gorm:"not null;default:0" json:"sold_count"`
}

func (SeckillProduct) TableName() string { return "seckill_products" }
