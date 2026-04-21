package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// DistributionSetting 每租户 1 行
type DistributionSetting struct {
	TenantID    uint64          `gorm:"primaryKey" json:"tenant_id"`
	Enabled     int8            `gorm:"not null;default:1" json:"enabled"`
	Level1Rate  decimal.Decimal `gorm:"type:numeric(5,4);not null;default:0.10" json:"level1_rate"`
	Level2Rate  decimal.Decimal `gorm:"type:numeric(5,4);not null;default:0.05" json:"level2_rate"`
	MinWithdraw decimal.Decimal `gorm:"type:numeric(10,2);not null;default:10" json:"min_withdraw"`
	AutoBecome  int8            `gorm:"not null;default:0" json:"auto_become"`
	Remark      string          `gorm:"size:500;not null;default:''" json:"remark"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

func (DistributionSetting) TableName() string { return "distribution_settings" }

// Distributor 分销员
type Distributor struct {
	ID                uint64          `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID          uint64          `gorm:"not null;index" json:"tenant_id"`
	MemberID          uint64          `gorm:"not null" json:"member_id"`
	ParentID          uint64          `gorm:"not null;default:0" json:"parent_id"`
	GrandparentID     uint64          `gorm:"not null;default:0" json:"grandparent_id"`
	Status            int8            `gorm:"not null;default:0" json:"status"` // 0待审核 1正常 2冻结
	TotalCommission   decimal.Decimal `gorm:"type:numeric(12,2);not null;default:0" json:"total_commission"`
	PendingCommission decimal.Decimal `gorm:"type:numeric(12,2);not null;default:0" json:"pending_commission"`
	Withdrawn         decimal.Decimal `gorm:"type:numeric(12,2);not null;default:0" json:"withdrawn"`
	InviteCount       int             `gorm:"not null;default:0" json:"invite_count"`
	CreatedAt         time.Time       `json:"created_at"`
	ApprovedAt        *time.Time      `json:"approved_at"`
}

func (Distributor) TableName() string { return "distributors" }

// CommissionLog 佣金记录
type CommissionLog struct {
	ID            uint64          `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID      uint64          `gorm:"not null;index" json:"tenant_id"`
	DistributorID uint64          `gorm:"not null;index" json:"distributor_id"`
	MemberID      uint64          `gorm:"not null" json:"member_id"`
	OrderID       uint64          `gorm:"not null;index" json:"order_id"`
	OrderNo       string          `gorm:"size:64;not null" json:"order_no"`
	BuyerID       uint64          `gorm:"not null" json:"buyer_id"`
	Level         int8            `gorm:"not null" json:"level"`
	Amount        decimal.Decimal `gorm:"type:numeric(12,2);not null" json:"amount"`
	Rate          decimal.Decimal `gorm:"type:numeric(5,4);not null" json:"rate"`
	Status        int8            `gorm:"not null;default:1" json:"status"`
	SettledAt     *time.Time      `json:"settled_at"`
	CreatedAt     time.Time       `json:"created_at"`
}

func (CommissionLog) TableName() string { return "commission_logs" }
