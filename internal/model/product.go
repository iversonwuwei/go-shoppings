package model

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type ProductCategory struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID   uint64    `gorm:"not null;index" json:"tenant_id"`
	ParentID   uint64    `gorm:"not null;default:0;index" json:"parent_id"`
	Name       string    `gorm:"size:50;not null" json:"name"`
	Icon       string    `gorm:"size:255" json:"icon"`
	CoverImage string    `gorm:"size:255" json:"cover_image"`
	Sort       int       `gorm:"not null;default:0" json:"sort"`
	Status     int8      `gorm:"not null;default:1" json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (ProductCategory) TableName() string { return "product_categories" }

type TenantCategoryAsset struct {
	TenantID   uint64    `gorm:"primaryKey" json:"tenant_id"`
	CategoryID uint64    `gorm:"primaryKey" json:"category_id"`
	Icon       string    `gorm:"size:255" json:"icon"`
	CoverImage string    `gorm:"size:255" json:"cover_image"`
	Sort       *int      `gorm:"column:sort" json:"sort,omitempty"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (TenantCategoryAsset) TableName() string { return "tenant_category_assets" }

type Product struct {
	ID             uint64           `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID       uint64           `gorm:"not null;index" json:"tenant_id"`
	CategoryID     *uint64          `gorm:"index" json:"category_id,omitempty"`
	Name           string           `gorm:"size:200;not null" json:"name"`
	Subtitle       string           `gorm:"size:500" json:"subtitle"`
	CoverImage     string           `gorm:"size:255;not null" json:"cover_image"`
	Images         JSONB            `gorm:"type:jsonb;not null;default:'[]'" json:"images"`
	VideoURL       string           `gorm:"column:video_url;size:500" json:"video_url"`
	Description    string           `gorm:"type:text" json:"description"`
	DetailImages   JSONB            `gorm:"type:jsonb;not null;default:'[]'" json:"detail_images"`
	Price          decimal.Decimal  `gorm:"type:numeric(10,2);not null;default:0" json:"price"`
	CostPrice      *decimal.Decimal `gorm:"type:numeric(10,2)" json:"cost_price,omitempty"`
	Stock          int              `gorm:"not null;default:0" json:"stock"`
	StockWarning   int              `gorm:"not null;default:10" json:"stock_warning"`
	HasSKU         int8             `gorm:"column:has_sku;not null;default:0" json:"has_sku"`
	IsVirtual      int8             `gorm:"column:is_virtual;not null;default:0" json:"is_virtual"`
	DeliveryType   JSONB            `gorm:"type:jsonb;not null;default:'[]'" json:"delivery_type"`
	DeliveryFee    decimal.Decimal  `gorm:"type:numeric(10,2);not null;default:0" json:"delivery_fee"`
	Status         int8             `gorm:"not null;default:1" json:"status"` // 1上架 0下架
	IsRecommend    int8             `gorm:"not null;default:0" json:"is_recommend"`
	IsHot          int8             `gorm:"not null;default:0" json:"is_hot"`
	SEOTitle       string           `gorm:"column:seo_title;size:200" json:"seo_title"`
	SEOKeywords    string           `gorm:"column:seo_keywords;size:500" json:"seo_keywords"`
	SEODescription string           `gorm:"column:seo_description;size:500" json:"seo_description"`
	Sort           int              `gorm:"not null;default:0" json:"sort"`
	SoldCount      int              `gorm:"not null;default:0" json:"sold_count"`
	ViewCount      int              `gorm:"not null;default:0" json:"view_count"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
	DeletedAt      gorm.DeletedAt   `gorm:"index" json:"-"`
}

func (Product) TableName() string { return "products" }

type ProductSKU struct {
	ID         uint64           `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID   uint64           `gorm:"not null;index" json:"tenant_id"`
	ProductID  uint64           `gorm:"not null;index" json:"product_id"`
	SKUCode    string           `gorm:"column:sku_code;size:50;not null" json:"sku_code"`
	Attributes JSONRaw          `gorm:"type:jsonb;not null;default:'{}'" json:"attributes"`
	Price      decimal.Decimal  `gorm:"type:numeric(10,2);not null" json:"price"`
	CostPrice  *decimal.Decimal `gorm:"type:numeric(10,2)" json:"cost_price,omitempty"`
	Stock      int              `gorm:"not null;default:0" json:"stock"`
	Image      string           `gorm:"size:255" json:"image"`
	Weight     *decimal.Decimal `gorm:"type:numeric(10,2)" json:"weight,omitempty"`
	Volume     *decimal.Decimal `gorm:"type:numeric(10,2)" json:"volume,omitempty"`
	Status     int8             `gorm:"not null;default:1" json:"status"`
	CreatedAt  time.Time        `json:"created_at"`
	UpdatedAt  time.Time        `json:"updated_at"`
}

func (ProductSKU) TableName() string { return "product_skus" }

type ProductAttribute struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID  uint64    `gorm:"not null;index" json:"tenant_id"`
	ProductID uint64    `gorm:"not null;index" json:"product_id"`
	Name      string    `gorm:"size:30;not null" json:"name"`
	Sort      int       `gorm:"not null;default:0" json:"sort"`
	CreatedAt time.Time `json:"created_at"`
}

func (ProductAttribute) TableName() string { return "product_attributes" }

type ProductAttributeValue struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID    uint64    `gorm:"not null;index" json:"tenant_id"`
	AttributeID uint64    `gorm:"not null;index" json:"attribute_id"`
	Value       string    `gorm:"size:50;not null" json:"value"`
	Sort        int       `gorm:"not null;default:0" json:"sort"`
	Image       string    `gorm:"size:255" json:"image"`
	CreatedAt   time.Time `json:"created_at"`
}

func (ProductAttributeValue) TableName() string { return "product_attribute_values" }
