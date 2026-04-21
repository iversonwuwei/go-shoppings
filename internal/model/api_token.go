package model

import "time"

// ApiToken 租户开放API凭证
type ApiToken struct {
	ID          uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID    uint64     `gorm:"not null;index" json:"tenant_id"`
	Name        string     `gorm:"size:100;not null" json:"name"`
	AppKey      string     `gorm:"size:64;not null;uniqueIndex" json:"app_key"`
	AppSecret   string     `gorm:"size:128;not null" json:"app_secret"`
	Scopes      string     `gorm:"size:500;not null;default:''" json:"scopes"`
	IPWhitelist string     `gorm:"column:ip_whitelist;size:500;not null;default:''" json:"ip_whitelist"`
	Status      int8       `gorm:"not null;default:1" json:"status"`
	LastUsedAt  *time.Time `json:"last_used_at"`
	ExpiresAt   *time.Time `json:"expires_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (ApiToken) TableName() string { return "api_tokens" }

// ApiRequestLog 开放API请求日志
type ApiRequestLog struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID   uint64    `gorm:"not null;index" json:"tenant_id"`
	TokenID    uint64    `gorm:"not null;index" json:"token_id"`
	AppKey     string    `gorm:"size:64;not null" json:"app_key"`
	Method     string    `gorm:"size:10;not null" json:"method"`
	Path       string    `gorm:"size:255;not null" json:"path"`
	StatusCode int       `gorm:"column:status_code;not null;default:0" json:"status_code"`
	IP         string    `gorm:"column:ip;size:64;not null;default:''" json:"ip"`
	CostMs     int       `gorm:"column:cost_ms;not null;default:0" json:"cost_ms"`
	CreatedAt  time.Time `json:"created_at"`
}

func (ApiRequestLog) TableName() string { return "api_request_logs" }
