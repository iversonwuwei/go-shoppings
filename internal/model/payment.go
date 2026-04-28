package model

import (
	"time"

	"github.com/shopspring/decimal"
)

const (
	PaymentStatusPending = "pending"
	PaymentStatusPaid    = "paid"
	PaymentStatusClosed  = "closed"
	PaymentStatusRefund  = "refunded"

	PaymentSceneMemberOrder = "member_order"
)

type Payment struct {
	ID                  uint64           `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID            uint64           `gorm:"not null;index" json:"tenant_id"`
	PaymentNo           string           `gorm:"size:32;not null;uniqueIndex" json:"payment_no"`
	OrderNo             string           `gorm:"size:32;index" json:"order_no"`
	MemberID            uint64           `gorm:"not null" json:"member_id"`
	Amount              decimal.Decimal  `gorm:"type:numeric(10,2);not null" json:"amount"`
	Status              string           `gorm:"size:20;not null;default:'pending'" json:"status"`
	PayScene            string           `gorm:"size:32;not null;default:'member_order'" json:"pay_scene"`
	SpAppID             string           `gorm:"column:sp_appid;size:64" json:"sp_appid"`
	SpMchID             string           `gorm:"column:sp_mchid;size:64" json:"sp_mchid"`
	SubAppID            string           `gorm:"column:sub_appid;size:64" json:"sub_appid"`
	SubMchID            string           `gorm:"column:sub_mchid;size:64" json:"sub_mchid"`
	SettlementTenantID  uint64           `gorm:"not null;default:0;index" json:"settlement_tenant_id"`
	WechatTradeType     string           `gorm:"size:20" json:"wechat_trade_type"`
	WechatTransactionID string           `gorm:"size:64;index" json:"wechat_transaction_id"`
	WechatPayerOpenID   string           `gorm:"column:wechat_payer_openid;size:64" json:"wechat_payer_openid"`
	WechatPaidAt        *time.Time       `json:"wechat_paid_at,omitempty"`
	RefundAmount        *decimal.Decimal `gorm:"type:numeric(10,2)" json:"refund_amount,omitempty"`
	RefundStatus        string           `gorm:"size:20;default:'none'" json:"refund_status"`
	ClosedAt            *time.Time       `json:"closed_at,omitempty"`
	CloseReason         string           `gorm:"size:200" json:"close_reason,omitempty"`
	ExpireAt            *time.Time       `json:"expire_at,omitempty"`
	CreatedAt           time.Time        `json:"created_at"`
	UpdatedAt           time.Time        `json:"updated_at"`
}

func (Payment) TableName() string { return "payments" }
