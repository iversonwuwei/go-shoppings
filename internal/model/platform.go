package model

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// ========== Platform ==========

type Plan struct {
	ID           uint64          `gorm:"primaryKey;autoIncrement" json:"id"`
	Name         string          `gorm:"size:50;not null" json:"name"`
	Code         string          `gorm:"size:30;not null;uniqueIndex" json:"code"`
	MonthlyFee   decimal.Decimal `gorm:"type:numeric(10,2);not null" json:"monthly_fee"`
	YearlyFee    decimal.Decimal `gorm:"type:numeric(10,2);not null" json:"yearly_fee"`
	ProductLimit int             `gorm:"not null;default:0" json:"product_limit"`
	OrderLimit   int             `gorm:"not null;default:0" json:"order_limit"`
	UserLimit    int             `gorm:"not null;default:0" json:"user_limit"`
	Features     JSONB           `gorm:"type:jsonb;not null;default:'[]'" json:"features"`
	IsDefault    int8            `gorm:"not null;default:0" json:"is_default"`
	Status       int8            `gorm:"not null;default:1" json:"status"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

func (Plan) TableName() string { return "plans" }

type Tenant struct {
	ID               uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Code             string    `gorm:"size:30;not null;uniqueIndex" json:"code"`
	CompanyName      string    `gorm:"size:100;not null" json:"company_name"`
	ContactName      string    `gorm:"size:50;not null" json:"contact_name"`
	ContactPhone     string    `gorm:"size:20;not null" json:"contact_phone"`
	ContactEmail     string    `gorm:"size:100;not null" json:"contact_email"`
	WechatAppID      string    `gorm:"column:wechat_appid;size:50" json:"wechat_appid,omitempty"`
	WechatSecret     string    `gorm:"size:255" json:"-"`
	WechatMchID      string    `gorm:"column:wechat_mchid;size:30" json:"wechat_mchid,omitempty"`
	WechatAPIv3Key   string    `gorm:"column:wechat_apiv3_key;size:255" json:"-"`
	WechatCertSerial string    `gorm:"size:100" json:"wechat_cert_serial,omitempty"`
	PlanID           uint64    `gorm:"not null" json:"plan_id"`
	PlanExpireAt     time.Time `gorm:"not null" json:"plan_expire_at"`
	BillingCycle     string    `gorm:"size:10;not null;default:'yearly'" json:"billing_cycle"` // monthly / yearly
	BrandName        string    `gorm:"size:50" json:"brand_name"`
	BrandLogo        string    `gorm:"size:255" json:"brand_logo"`
	BrandTheme       string    `gorm:"size:20;default:'#1989fa'" json:"brand_theme"`
	BrandDomain      string    `gorm:"size:100" json:"brand_domain"`
	Status           int8      `gorm:"not null;default:0" json:"status"` // 0待审核 1正常 2欠费 3封禁
	RejectReason     string    `gorm:"size:255" json:"reject_reason,omitempty"`
	ExtraFeatures    JSONB     `gorm:"type:jsonb;not null;default:'[]'" json:"extra_features"` // 平台单独授予的功能（与 plan.Features 取并集）
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func (Tenant) TableName() string { return "tenants" }

type Admin struct {
	ID          uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Username    string     `gorm:"size:50;not null;uniqueIndex" json:"username"`
	Password    string     `gorm:"size:255;not null" json:"-"`
	RealName    string     `gorm:"size:50" json:"real_name"`
	Phone       string     `gorm:"size:20" json:"phone"`
	Email       string     `gorm:"size:100" json:"email"`
	Role        string     `gorm:"size:20;not null;default:'admin'" json:"role"` // super / admin / tenant
	TenantID    uint64     `gorm:"default:0" json:"tenant_id"`
	Status      int8       `gorm:"not null;default:1" json:"status"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	LastLoginIP string     `gorm:"size:50" json:"last_login_ip,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (Admin) TableName() string { return "admins" }

type TenantPlanLog struct {
	ID          uint64          `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID    uint64          `gorm:"not null;index" json:"tenant_id"`
	OldPlanID   uint64          `json:"old_plan_id"`
	NewPlanID   uint64          `gorm:"not null" json:"new_plan_id"`
	ChangeType  string          `gorm:"size:20;not null" json:"change_type"`
	EffectiveAt time.Time       `gorm:"not null" json:"effective_at"`
	ExpireAt    time.Time       `gorm:"not null" json:"expire_at"`
	Amount      decimal.Decimal `gorm:"type:numeric(10,2);not null" json:"amount"`
	CreatedAt   time.Time       `json:"created_at"`
}

func (TenantPlanLog) TableName() string { return "tenant_plan_logs" }

// TenantSubscriptionOrder 租户订阅订单（向平台统一商户号支付）
type TenantSubscriptionOrder struct {
	ID                     uint64          `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID               uint64          `gorm:"not null;index" json:"tenant_id"`
	PlanID                 uint64          `gorm:"not null" json:"plan_id"`
	BillingCycle           string          `gorm:"size:10;not null" json:"billing_cycle"` // monthly / yearly
	Amount                 decimal.Decimal `gorm:"type:numeric(10,2);not null" json:"amount"`
	Status                 int8            `gorm:"not null;default:0" json:"status"` // 0待支付 1已支付 2已取消 3已退款
	OrderNo                string          `gorm:"size:64;not null;uniqueIndex" json:"order_no"`
	CreatedByAdminID       uint64          `gorm:"not null;default:0;index" json:"created_by_admin_id"`
	CreatedByAdminUsername string          `gorm:"size:50;not null;default:''" json:"created_by_admin_username"`
	PayTransactionID       string          `gorm:"size:64;not null;default:''" json:"pay_transaction_id"`
	PayAt                  *time.Time      `json:"pay_at,omitempty"`
	ExpireBefore           *time.Time      `json:"expire_before,omitempty"`
	ExpireAfter            *time.Time      `json:"expire_after,omitempty"`
	CreatedAt              time.Time       `json:"created_at"`
	UpdatedAt              time.Time       `json:"updated_at"`
}

func (TenantSubscriptionOrder) TableName() string { return "tenant_subscription_orders" }

// PlatformSettings 平台全局设置（单行，id 固定为 1）
type PlatformSettings struct {
	ID               uint64    `gorm:"primaryKey" json:"id"`
	PlatformName     string    `json:"platform_name"`
	PlatformLogo     string    `json:"platform_logo"`
	SupportPhone     string    `json:"support_phone"`
	SupportEmail     string    `json:"support_email"`
	WxpayAppID       string    `gorm:"column:wxpay_app_id" json:"wxpay_app_id"`
	WxpayMchID       string    `gorm:"column:wxpay_mch_id" json:"wxpay_mch_id"`
	WxpayAPIv3Key    string    `gorm:"column:wxpay_apiv3_key" json:"wxpay_apiv3_key"`
	WxpayCertSerial  string    `gorm:"column:wxpay_cert_serial" json:"wxpay_cert_serial"`
	WxpayNotifyURL   string    `gorm:"column:wxpay_notify_url" json:"wxpay_notify_url"`
	SpAppID          string    `gorm:"column:sp_appid" json:"sp_appid"`
	SpMchID          string    `gorm:"column:sp_mchid" json:"sp_mchid"`
	SpAPIv3Key       string    `gorm:"column:sp_apiv3_key" json:"sp_apiv3_key"`
	SpCertSerial     string    `gorm:"column:sp_cert_serial" json:"sp_cert_serial"`
	PartnerNotifyURL string    `gorm:"column:partner_notify_url" json:"partner_notify_url"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func (PlatformSettings) TableName() string { return "platform_settings" }

// PlanFeature 平台统一维护的套餐功能目录（plans.features 存储的是 code 列表）
type PlanFeature struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Code        string    `gorm:"size:40;not null;uniqueIndex" json:"code"`
	Name        string    `gorm:"size:50;not null" json:"name"`
	Description string    `gorm:"size:255;not null;default:''" json:"description"`
	GroupName   string    `gorm:"column:group_name;size:30;not null;default:''" json:"group_name"`
	Sort        int       `gorm:"not null;default:0" json:"sort"`
	Status      int8      `gorm:"not null;default:1" json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (PlanFeature) TableName() string { return "plan_features" }

// Region 平台统一维护的省/市/区数据，供小程序地址选择使用。
type Region struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ParentID  uint64    `gorm:"not null;default:0;index" json:"parent_id"`
	Code      string    `gorm:"size:32;not null;default:'';index" json:"code"`
	Name      string    `gorm:"size:50;not null" json:"name"`
	Level     int8      `gorm:"not null;default:1;index" json:"level"`
	Sort      int       `gorm:"not null;default:0" json:"sort"`
	Enabled   int8      `gorm:"not null;default:1;index" json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Children  []Region  `gorm:"-" json:"children,omitempty"`
}

func (Region) TableName() string { return "regions" }

// 公共删除时间字段
type SoftDelete struct {
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
