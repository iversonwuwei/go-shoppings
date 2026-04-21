package model

import "time"

// 审核状态
const (
	ConfigAuditPending  int8 = 0 // 待审核
	ConfigAuditApproved int8 = 1 // 通过
	ConfigAuditRejected int8 = 2 // 拒绝
)

// TenantPaymentConfig 商户收款配置（微信支付 / 支付宝等）
// 每个租户每种 provider 一条记录
type TenantPaymentConfig struct {
	ID            uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID      uint64     `gorm:"not null;uniqueIndex:idx_tpc_tenant_provider,priority:1" json:"tenant_id"`
	Provider      string     `gorm:"size:20;not null;uniqueIndex:idx_tpc_tenant_provider,priority:2" json:"provider"` // wechat / alipay
	MchID         string     `gorm:"size:64" json:"mch_id"`
	AppID         string     `gorm:"size:64" json:"app_id"`
	APIV3Key      string     `gorm:"column:api_v3_key;size:128" json:"api_v3_key"`
	CertSerialNo  string     `gorm:"size:64" json:"cert_serial_no"`
	PrivateKeyPEM string     `gorm:"column:private_key_pem;type:text" json:"private_key_pem"`
	CertPEM       string     `gorm:"column:cert_pem;type:text" json:"cert_pem,omitempty"`
	NotifyURL     string     `gorm:"size:255" json:"notify_url"`
	Enabled       int8       `gorm:"not null;default:0" json:"enabled"`      // 0 关闭 1 启用（必须审核通过后才允许启用）
	AuditStatus   int8       `gorm:"not null;default:0" json:"audit_status"` // 0/1/2
	AuditRemark   string     `gorm:"size:500" json:"audit_remark"`
	SubmittedAt   *time.Time `json:"submitted_at,omitempty"`
	AuditedAt     *time.Time `json:"audited_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

func (TenantPaymentConfig) TableName() string { return "tenant_payment_configs" }

// ShippingCarrier 商户物流承运商配置（含第三方物流查询接口凭证）
type ShippingCarrier struct {
	ID          uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID    uint64     `gorm:"not null;index" json:"tenant_id"`
	Code        string     `gorm:"size:30;not null" json:"code"` // 承运商编码 sf/yto/zto/jd/ems 等
	Name        string     `gorm:"size:50;not null" json:"name"`
	APIProvider string     `gorm:"column:api_provider;size:30" json:"api_provider"` // kuaidi100 / aliyun / none
	APICustomer string     `gorm:"column:api_customer;size:128" json:"api_customer"`
	APIKey      string     `gorm:"column:api_key;size:256" json:"api_key"`
	APISecret   string     `gorm:"column:api_secret;size:256" json:"api_secret,omitempty"`
	Priority    int        `gorm:"not null;default:0" json:"priority"`
	Enabled     int8       `gorm:"not null;default:0" json:"enabled"`
	AuditStatus int8       `gorm:"not null;default:0" json:"audit_status"`
	AuditRemark string     `gorm:"size:500" json:"audit_remark"`
	SubmittedAt *time.Time `json:"submitted_at,omitempty"`
	AuditedAt   *time.Time `json:"audited_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (ShippingCarrier) TableName() string { return "shipping_carriers" }
