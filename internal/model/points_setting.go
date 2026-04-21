package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// PointsSetting 每租户 1 行的积分规则
type PointsSetting struct {
	TenantID   uint64          `gorm:"primaryKey" json:"tenant_id"`
	Enabled    int8            `gorm:"not null;default:1" json:"enabled"`
	EarnRate   decimal.Decimal `gorm:"type:numeric(10,4);not null;default:1" json:"earn_rate"`
	MinAmount  decimal.Decimal `gorm:"type:numeric(10,2);not null;default:0" json:"min_amount"`
	RedeemRate int             `gorm:"not null;default:100" json:"redeem_rate"`
	Remark     string          `gorm:"size:500;not null;default:''" json:"remark"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

func (PointsSetting) TableName() string { return "points_settings" }
