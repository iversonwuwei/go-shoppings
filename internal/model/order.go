package model

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// 订单状态
const (
	OrderStatusPendingPay = "pending_pay"
	OrderStatusPaid       = "paid"
	OrderStatusPreparing  = "preparing"
	OrderStatusShipped    = "shipped"
	OrderStatusDelivered  = "delivered"
	OrderStatusCompleted  = "completed"
	OrderStatusCancelled  = "cancelled"
	OrderStatusRefunding  = "refunding"
	OrderStatusRefunded   = "refunded"
)

const (
	OrderMessageStatusUnread = "unread"
	OrderMessageStatusRead   = "read"
)

type Order struct {
	ID                 uint64           `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID           uint64           `gorm:"not null;index" json:"tenant_id"`
	OrderNo            string           `gorm:"size:32;not null;uniqueIndex" json:"order_no"`
	MemberID           uint64           `gorm:"not null;index" json:"member_id"`
	TotalAmount        decimal.Decimal  `gorm:"type:numeric(10,2);not null" json:"total_amount"`
	DeliveryFee        decimal.Decimal  `gorm:"type:numeric(10,2);not null;default:0" json:"delivery_fee"`
	DiscountAmount     decimal.Decimal  `gorm:"type:numeric(10,2);not null;default:0" json:"discount_amount"`
	CouponID           uint64           `json:"coupon_id"`
	PointsDiscount     decimal.Decimal  `gorm:"type:numeric(10,2);not null;default:0" json:"points_discount"`
	ActualAmount       decimal.Decimal  `gorm:"type:numeric(10,2);not null" json:"actual_amount"`
	CostAmount         *decimal.Decimal `gorm:"type:numeric(10,2)" json:"cost_amount,omitempty"`
	Status             string           `gorm:"size:20;not null;default:'pending_pay';index" json:"status"`
	ReceiverName       string           `gorm:"size:50" json:"receiver_name"`
	ReceiverPhone      string           `gorm:"size:20" json:"receiver_phone"`
	ReceiverProvince   string           `gorm:"size:20" json:"receiver_province"`
	ReceiverCity       string           `gorm:"size:20" json:"receiver_city"`
	ReceiverDistrict   string           `gorm:"size:20" json:"receiver_district"`
	ReceiverAddress    string           `gorm:"size:255" json:"receiver_address"`
	ReceiverPostcode   string           `gorm:"size:10" json:"receiver_postcode"`
	DeliveryType       string           `gorm:"size:20;not null" json:"delivery_type"`
	IsVirtual          int8             `gorm:"column:is_virtual;not null;default:0" json:"is_virtual"`
	ExpressCompany     string           `gorm:"size:30" json:"express_company"`
	ExpressNo          string           `gorm:"size:50" json:"express_no"`
	SelfPickupCode     string           `gorm:"size:20" json:"self_pickup_code"`
	SelfPickupAddress  string           `gorm:"size:255" json:"self_pickup_address"`
	BuyerRemark        string           `gorm:"size:500" json:"buyer_remark"`
	DistributionStatus string           `gorm:"size:20;default:'pending'" json:"distribution_status"`
	PaidAt             *time.Time       `json:"paid_at,omitempty"`
	ShippedAt          *time.Time       `json:"shipped_at,omitempty"`
	DeliveredAt        *time.Time       `json:"delivered_at,omitempty"`
	CompletedAt        *time.Time       `json:"completed_at,omitempty"`
	CancelledAt        *time.Time       `json:"cancelled_at,omitempty"`
	ExpiredAt          *time.Time       `json:"expired_at,omitempty"`
	CreatedAt          time.Time        `json:"created_at"`
	UpdatedAt          time.Time        `json:"updated_at"`
	DeletedAt          gorm.DeletedAt   `gorm:"index" json:"-"`

	Items []OrderItem `gorm:"foreignKey:OrderID" json:"items,omitempty"`
}

func (Order) TableName() string { return "orders" }

type OrderItem struct {
	ID           uint64           `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID     uint64           `gorm:"not null;index" json:"tenant_id"`
	OrderID      uint64           `gorm:"not null;index" json:"order_id"`
	ProductID    uint64           `gorm:"not null" json:"product_id"`
	SKUID        uint64           `gorm:"column:sku_id" json:"sku_id"`
	ProductName  string           `gorm:"size:200;not null" json:"product_name"`
	SKUDesc      string           `gorm:"size:200" json:"sku_desc"`
	CoverImage   string           `gorm:"size:255;not null" json:"cover_image"`
	Price        decimal.Decimal  `gorm:"type:numeric(10,2);not null" json:"price"`
	Quantity     int              `gorm:"not null;default:1" json:"quantity"`
	ItemTotal    decimal.Decimal  `gorm:"type:numeric(10,2);not null" json:"item_total"`
	RefundStatus string           `gorm:"size:20;default:'none'" json:"refund_status"`
	RefundAmount *decimal.Decimal `gorm:"type:numeric(10,2)" json:"refund_amount,omitempty"`
	CreatedAt    time.Time        `json:"created_at"`
}

func (OrderItem) TableName() string { return "order_items" }

type OrderLog struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID     uint64    `gorm:"not null;index" json:"tenant_id"`
	OrderID      uint64    `gorm:"not null;index" json:"order_id"`
	OperatorType string    `gorm:"size:20;not null" json:"operator_type"`
	OperatorID   uint64    `json:"operator_id"`
	Action       string    `gorm:"size:50;not null" json:"action"`
	BeforeStatus string    `gorm:"size:20" json:"before_status"`
	AfterStatus  string    `gorm:"size:20" json:"after_status"`
	Remark       string    `gorm:"size:500" json:"remark"`
	CreatedAt    time.Time `json:"created_at"`
}

func (OrderLog) TableName() string { return "order_logs" }

type OrderMessage struct {
	ID        uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID  uint64     `gorm:"not null;index" json:"tenant_id"`
	OrderID   uint64     `gorm:"not null;index" json:"order_id"`
	OrderNo   string     `gorm:"size:32;not null;index" json:"order_no"`
	EventType string     `gorm:"size:40;not null" json:"event_type"`
	Title     string     `gorm:"size:120;not null" json:"title"`
	Content   string     `gorm:"size:500;not null" json:"content"`
	Status    string     `gorm:"size:20;not null;default:'unread';index" json:"status"`
	ReadAt    *time.Time `json:"read_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

func (OrderMessage) TableName() string { return "order_messages" }
