package model

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type MemberLevel struct {
	ID           uint64          `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID     uint64          `gorm:"not null;index" json:"tenant_id"`
	Name         string          `gorm:"size:30;not null" json:"name"`
	Icon         string          `gorm:"size:255" json:"icon"`
	Color        string          `gorm:"size:20" json:"color"`
	MinGrowth    int             `gorm:"not null;default:0;index" json:"min_growth"`
	DiscountRate decimal.Decimal `gorm:"type:numeric(4,2);not null;default:100" json:"discount_rate"`
	PointsMult   decimal.Decimal `gorm:"type:numeric(3,2);not null;default:1" json:"points_mult"`
	Sort         int             `gorm:"not null;default:0" json:"sort"`
	CreatedAt    time.Time       `json:"created_at"`
}

func (MemberLevel) TableName() string { return "member_levels" }

type Member struct {
	ID            uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID      uint64         `gorm:"not null;index" json:"tenant_id"`
	OpenID        string         `gorm:"column:openid;size:50" json:"openid"`
	UnionID       string         `gorm:"column:unionid;size:50" json:"unionid,omitempty"`
	SessionKey    string         `gorm:"size:255" json:"-"`
	Nickname      string         `gorm:"size:50" json:"nickname"`
	Avatar        string         `gorm:"size:255" json:"avatar"`
	Gender        int8           `json:"gender"`
	Birthday      *time.Time     `gorm:"type:date" json:"birthday,omitempty"`
	Phone         string         `gorm:"size:20" json:"phone"`
	LevelID       uint64         `gorm:"index" json:"level_id"`
	LevelExpireAt *time.Time     `json:"level_expire_at,omitempty"`
	Points        int            `gorm:"not null;default:0" json:"points"`
	GrowthValue   int            `gorm:"not null;default:0" json:"growth_value"`
	ParentID      uint64         `gorm:"index" json:"parent_id"`
	Level1Count   int            `gorm:"not null;default:0" json:"level1_count"`
	Level2Count   int            `gorm:"not null;default:0" json:"level2_count"`
	Status        int8           `gorm:"not null;default:1" json:"status"`
	LastLoginAt   *time.Time     `json:"last_login_at,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Member) TableName() string { return "members" }

type MemberAddress struct {
	ID            uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID      uint64    `gorm:"not null;index" json:"tenant_id"`
	MemberID      uint64    `gorm:"not null;index" json:"member_id"`
	ReceiverName  string    `gorm:"size:50;not null" json:"receiver_name"`
	ReceiverPhone string    `gorm:"size:20;not null" json:"receiver_phone"`
	Province      string    `gorm:"size:20;not null" json:"province"`
	City          string    `gorm:"size:20;not null" json:"city"`
	District      string    `gorm:"size:20;not null" json:"district"`
	Address       string    `gorm:"size:255;not null" json:"address"`
	Postcode      string    `gorm:"size:10" json:"postcode"`
	IsDefault     int8      `gorm:"not null;default:0" json:"is_default"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (MemberAddress) TableName() string { return "member_addresses" }

type PointsLog struct {
	ID            uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID      uint64    `gorm:"not null;index" json:"tenant_id"`
	MemberID      uint64    `gorm:"not null;index" json:"member_id"`
	ChangeType    string    `gorm:"size:20;not null" json:"change_type"`
	ChangeValue   int       `gorm:"not null" json:"change_value"`
	BalanceBefore int       `gorm:"not null" json:"balance_before"`
	BalanceAfter  int       `gorm:"not null" json:"balance_after"`
	SourceID      uint64    `json:"source_id"`
	SourceDesc    string    `gorm:"size:200" json:"source_desc"`
	Remark        string    `gorm:"size:500" json:"remark"`
	OperatorID    uint64    `json:"operator_id"`
	CreatedAt     time.Time `json:"created_at"`
}

func (PointsLog) TableName() string { return "points_logs" }
