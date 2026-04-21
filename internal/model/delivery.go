package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// DeliverySetting 租户配送设置（快递/同城/自提合一行）
type DeliverySetting struct {
	TenantID uint64 `gorm:"primaryKey" json:"tenant_id"`

	ExpressEnabled    int8            `gorm:"not null;default:1" json:"express_enabled"`
	ExpressFreeAmount decimal.Decimal `gorm:"type:numeric(10,2);not null;default:0" json:"express_free_amount"`
	ExpressDefaultFee decimal.Decimal `gorm:"type:numeric(10,2);not null;default:0" json:"express_default_fee"`

	CityEnabled  int8            `gorm:"not null;default:0" json:"city_enabled"`
	CityRadiusKm decimal.Decimal `gorm:"type:numeric(6,2);not null;default:5" json:"city_radius_km"`
	CityBaseFee  decimal.Decimal `gorm:"type:numeric(10,2);not null;default:5" json:"city_base_fee"`
	CityPerKmFee decimal.Decimal `gorm:"type:numeric(10,2);not null;default:1" json:"city_per_km_fee"`
	CityMinOrder decimal.Decimal `gorm:"type:numeric(10,2);not null;default:0" json:"city_min_order"`

	PickupEnabled int8   `gorm:"not null;default:0" json:"pickup_enabled"`
	PickupAddress string `gorm:"size:255;not null;default:''" json:"pickup_address"`
	PickupHours   string `gorm:"size:100;not null;default:''" json:"pickup_hours"`
	PickupPhone   string `gorm:"size:30;not null;default:''" json:"pickup_phone"`

	Remark    string    `gorm:"size:500;not null;default:''" json:"remark"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (DeliverySetting) TableName() string { return "delivery_settings" }

// TenantSiteConfig 租户站点品牌/域名/部署配置合一行
type TenantSiteConfig struct {
	TenantID uint64 `gorm:"primaryKey" json:"tenant_id"`

	CustomDomain   string `gorm:"size:128;not null;default:''" json:"custom_domain"`
	DomainVerified int8   `gorm:"not null;default:0" json:"domain_verified"`
	SSLStatus      string `gorm:"column:ssl_status;size:20;not null;default:'none'" json:"ssl_status"`

	BrandName         string `gorm:"size:100;not null;default:''" json:"brand_name"`
	BrandLogo         string `gorm:"size:500;not null;default:''" json:"brand_logo"`
	PrimaryColor      string `gorm:"size:16;not null;default:'#409EFF'" json:"primary_color"`
	HidePlatformBrand int8   `gorm:"not null;default:0" json:"hide_platform_brand"`
	FooterText        string `gorm:"size:255;not null;default:''" json:"footer_text"`

	DeploymentMode  string `gorm:"size:16;not null;default:'shared'" json:"deployment_mode"`
	PrivateEndpoint string `gorm:"size:255;not null;default:''" json:"private_endpoint"`
	PrivateNotes    string `gorm:"size:500;not null;default:''" json:"private_notes"`

	UpdatedAt time.Time `json:"updated_at"`
}

func (TenantSiteConfig) TableName() string { return "tenant_site_configs" }
