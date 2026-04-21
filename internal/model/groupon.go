package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// GrouponActivity 拼团活动
type GrouponActivity struct {
	ID            uint64          `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID      uint64          `gorm:"not null;index" json:"tenant_id"`
	Name          string          `gorm:"size:100;not null" json:"name"`
	ProductID     uint64          `gorm:"not null" json:"product_id"`
	SKUID         uint64          `gorm:"column:sku_id;not null;default:0" json:"sku_id"`
	GroupPrice    decimal.Decimal `gorm:"type:numeric(10,2);not null" json:"group_price"`
	OriginalPrice decimal.Decimal `gorm:"type:numeric(10,2);not null" json:"original_price"`
	RequireNum    int             `gorm:"not null;default:2" json:"require_num"`
	ExpireHours   int             `gorm:"not null;default:24" json:"expire_hours"`
	TotalStock    int             `gorm:"not null;default:0" json:"total_stock"`
	SoldCount     int             `gorm:"not null;default:0" json:"sold_count"`
	StartAt       time.Time       `gorm:"not null" json:"start_at"`
	EndAt         time.Time       `gorm:"not null" json:"end_at"`
	Status        int8            `gorm:"not null;default:1" json:"status"`
	CreatedAt     time.Time       `json:"created_at"`
}

func (GrouponActivity) TableName() string { return "groupon_activities" }

// Groupon 单次拼团单
type Groupon struct {
	ID         uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID   uint64     `gorm:"not null;index" json:"tenant_id"`
	ActivityID uint64     `gorm:"not null;index" json:"activity_id"`
	LeaderID   uint64     `gorm:"not null" json:"leader_id"`
	RequireNum int        `gorm:"not null" json:"require_num"`
	CurrentNum int        `gorm:"not null;default:1" json:"current_num"`
	Status     int8       `gorm:"not null;default:1" json:"status"` // 1进行中 2成团 3失败
	ExpiresAt  time.Time  `gorm:"not null" json:"expires_at"`
	SucceedAt  *time.Time `json:"succeed_at"`
	CreatedAt  time.Time  `json:"created_at"`
}

func (Groupon) TableName() string { return "groupons" }

// GrouponMember 参团成员
type GrouponMember struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID  uint64    `gorm:"not null;index" json:"tenant_id"`
	GrouponID uint64    `gorm:"not null;index" json:"groupon_id"`
	MemberID  uint64    `gorm:"not null;index" json:"member_id"`
	OrderID   uint64    `gorm:"not null;default:0" json:"order_id"`
	IsLeader  int8      `gorm:"not null;default:0" json:"is_leader"`
	JoinedAt  time.Time `json:"joined_at"`
}

func (GrouponMember) TableName() string { return "groupon_members" }
