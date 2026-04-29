package model

import "time"

type MemberCartItem struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID  uint64    `gorm:"not null;uniqueIndex:uniq_member_cart_item;index" json:"tenant_id"`
	MemberID  uint64    `gorm:"not null;uniqueIndex:uniq_member_cart_item;index" json:"member_id"`
	ProductID uint64    `gorm:"not null;uniqueIndex:uniq_member_cart_item" json:"product_id"`
	SKUID     uint64    `gorm:"column:sku_id;not null;default:0;uniqueIndex:uniq_member_cart_item" json:"sku_id"`
	Quantity  int       `gorm:"not null;default:1" json:"quantity"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (MemberCartItem) TableName() string { return "member_cart_items" }
