package model

import (
	"time"

	"github.com/shopspring/decimal"
)

const (
	AfterSaleTypeRefund       = "refund"
	AfterSaleTypeReturnRefund = "return_refund"
)

const (
	AfterSaleStatusPending   = "pending"
	AfterSaleStatusApproved  = "approved"
	AfterSaleStatusRejected  = "rejected"
	AfterSaleStatusReturning = "returning"
	AfterSaleStatusReceived  = "received"
	AfterSaleStatusRefunded  = "refunded"
	AfterSaleStatusCancelled = "cancelled"
)

const (
	AfterSaleReasonTypeAll          = "all"
	AfterSaleReasonTypeRefund       = AfterSaleTypeRefund
	AfterSaleReasonTypeReturnRefund = AfterSaleTypeReturnRefund
)

type AfterSaleOrder struct {
	ID                   uint64          `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID             uint64          `gorm:"not null;index" json:"tenant_id"`
	AfterSaleNo          string          `gorm:"size:32;not null;uniqueIndex" json:"after_sale_no"`
	OrderID              uint64          `gorm:"not null;index" json:"order_id"`
	OrderNo              string          `gorm:"size:32;not null;index" json:"order_no"`
	OrderItemID          uint64          `gorm:"not null;default:0;index" json:"order_item_id"`
	MemberID             uint64          `gorm:"not null;index" json:"member_id"`
	Type                 string          `gorm:"size:20;not null" json:"type"`
	Status               string          `gorm:"size:20;not null;default:'pending';index" json:"status"`
	Amount               decimal.Decimal `gorm:"type:numeric(10,2);not null" json:"amount"`
	Reason               string          `gorm:"size:120;not null" json:"reason"`
	Description          string          `gorm:"size:500" json:"description"`
	Images               JSONB           `gorm:"type:jsonb;not null;default:'[]'" json:"images"`
	OrderStatusBefore    string          `gorm:"size:20;not null" json:"order_status_before"`
	AuditRemark          string          `gorm:"size:500" json:"audit_remark"`
	RefundRemark         string          `gorm:"size:500" json:"refund_remark"`
	ReturnExpressCode    string          `gorm:"size:30" json:"return_express_code"`
	ReturnExpressCompany string          `gorm:"size:80" json:"return_express_company"`
	ReturnExpressNo      string          `gorm:"size:80" json:"return_express_no"`
	RefundNo             string          `gorm:"size:64" json:"refund_no"`
	AppliedAt            time.Time       `gorm:"not null" json:"applied_at"`
	AuditedAt            *time.Time      `json:"audited_at,omitempty"`
	ReturnedAt           *time.Time      `json:"returned_at,omitempty"`
	ReceivedAt           *time.Time      `json:"received_at,omitempty"`
	RefundedAt           *time.Time      `json:"refunded_at,omitempty"`
	CancelledAt          *time.Time      `json:"cancelled_at,omitempty"`
	CreatedAt            time.Time       `json:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at"`
}

func (AfterSaleOrder) TableName() string { return "after_sale_orders" }

type AfterSaleReason struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Code      string    `gorm:"size:40;not null;uniqueIndex" json:"code"`
	Label     string    `gorm:"size:80;not null" json:"label"`
	Type      string    `gorm:"size:20;not null;default:'all';index" json:"type"`
	SortOrder int       `gorm:"not null;default:0;index" json:"sort_order"`
	Enabled   int8      `gorm:"not null;default:1;index" json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (AfterSaleReason) TableName() string { return "after_sale_reasons" }
