package model

import "time"

// SmsSetting 短信网关配置：tenant_id=0 表示平台自身配置，tenant_id>0 表示租户配置。
type SmsSetting struct {
	TenantID     uint64    `gorm:"primaryKey;autoIncrement:false" json:"tenant_id"`
	Enabled      int8      `gorm:"not null;default:0" json:"enabled"`
	Provider     string    `gorm:"size:32;not null;default:'aliyun'" json:"provider"`
	AccessKey    string    `gorm:"size:128;not null;default:''" json:"access_key"`
	AccessSecret string    `gorm:"size:256;not null;default:''" json:"access_secret"`
	SignName     string    `gorm:"size:64;not null;default:''" json:"sign_name"`
	Region       string    `gorm:"size:32;not null;default:''" json:"region"`
	Remark       string    `gorm:"size:500;not null;default:''" json:"remark"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (SmsSetting) TableName() string { return "sms_settings" }

// SmsTemplate 短信模板
type SmsTemplate struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID   uint64    `gorm:"not null;index" json:"tenant_id"`
	Code       string    `gorm:"size:64;not null" json:"code"`
	Name       string    `gorm:"size:100;not null" json:"name"`
	TemplateID string    `gorm:"size:64;not null;default:''" json:"template_id"`
	Content    string    `gorm:"size:500;not null;default:''" json:"content"`
	Enabled    int8      `gorm:"not null;default:1" json:"enabled"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (SmsTemplate) TableName() string { return "sms_templates" }

// SmsLog 短信发送日志
type SmsLog struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID  uint64    `gorm:"not null;index" json:"tenant_id"`
	Phone     string    `gorm:"size:20;not null" json:"phone"`
	Code      string    `gorm:"size:64;not null" json:"code"`
	Content   string    `gorm:"size:500;not null;default:''" json:"content"`
	Status    int8      `gorm:"not null;default:1" json:"status"` // 1成功 2失败
	Error     string    `gorm:"size:500;not null;default:''" json:"error"`
	BizID     string    `gorm:"column:biz_id;size:64;not null;default:''" json:"biz_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (SmsLog) TableName() string { return "sms_logs" }
