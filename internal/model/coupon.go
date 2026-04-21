package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// 优惠券类型
const (
	CouponTypeCash     = "cash"     // 现金券
	CouponTypeDiscount = "discount" // 折扣券
	CouponTypeShipping = "shipping" // 免邮券
)

type Coupon struct {
	ID              uint64           `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID        uint64           `gorm:"not null;index" json:"tenant_id"`
	Name            string           `gorm:"size:50;not null" json:"name"`
	Type            string           `gorm:"size:20;not null" json:"type"`
	ThresholdAmount *decimal.Decimal `gorm:"type:numeric(10,2)" json:"threshold_amount,omitempty"`
	DiscountValue   decimal.Decimal  `gorm:"type:numeric(10,2);not null" json:"discount_value"`
	MaxDiscount     *decimal.Decimal `gorm:"type:numeric(10,2)" json:"max_discount,omitempty"`
	TotalCount      int              `gorm:"not null;default:0" json:"total_count"`
	RemainCount     int              `gorm:"not null;default:0" json:"remain_count"`
	PerLimit        int              `gorm:"not null;default:1" json:"per_limit"`
	ReceiveStartAt  *time.Time       `json:"receive_start_at,omitempty"`
	ReceiveEndAt    *time.Time       `json:"receive_end_at,omitempty"`
	ValidStartAt    *time.Time       `json:"valid_start_at,omitempty"`
	ValidEndAt      *time.Time       `json:"valid_end_at,omitempty"`
	ValidDays       int              `json:"valid_days"`
	ApplicableType  string           `gorm:"size:20;not null;default:'all'" json:"applicable_type"`
	ApplicableIDs   JSONRaw          `gorm:"type:jsonb" json:"applicable_ids,omitempty"`
	MemberLevels    JSONRaw          `gorm:"type:jsonb" json:"member_levels,omitempty"`
	Status          int8             `gorm:"not null;default:1" json:"status"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
}

func (Coupon) TableName() string { return "coupons" }

type MemberCoupon struct {
	ID              uint64           `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID        uint64           `gorm:"not null;index" json:"tenant_id"`
	MemberID        uint64           `gorm:"not null;index" json:"member_id"`
	CouponID        uint64           `gorm:"not null" json:"coupon_id"`
	CouponName      string           `gorm:"size:50;not null" json:"coupon_name"`
	CouponType      string           `gorm:"size:20;not null" json:"coupon_type"`
	ThresholdAmount *decimal.Decimal `gorm:"type:numeric(10,2)" json:"threshold_amount,omitempty"`
	DiscountValue   decimal.Decimal  `gorm:"type:numeric(10,2);not null" json:"discount_value"`
	MaxDiscount     *decimal.Decimal `gorm:"type:numeric(10,2)" json:"max_discount,omitempty"`
	ReceivedAt      time.Time        `gorm:"not null" json:"received_at"`
	ValidStartAt    time.Time        `gorm:"not null" json:"valid_start_at"`
	ValidEndAt      time.Time        `gorm:"not null" json:"valid_end_at"`
	UsedAt          *time.Time       `json:"used_at,omitempty"`
	UsedOrderID     uint64           `json:"used_order_id"`
	Status          string           `gorm:"size:20;not null;default:'unused'" json:"status"`
	CreatedAt       time.Time        `json:"created_at"`
}

func (MemberCoupon) TableName() string { return "member_coupons" }
