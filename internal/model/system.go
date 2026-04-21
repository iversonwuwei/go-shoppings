package model

import "time"

type AdminActionLog struct {
	ID            uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID      uint64    `gorm:"not null;default:0;index" json:"tenant_id"`
	AdminID       uint64    `gorm:"not null;index" json:"admin_id"`
	AdminUsername string    `gorm:"size:50;not null" json:"admin_username"`
	Action        string    `gorm:"size:50;not null" json:"action"`
	TargetType    string    `gorm:"size:50" json:"target_type"`
	TargetID      uint64    `json:"target_id"`
	TargetDesc    string    `gorm:"size:200" json:"target_desc"`
	RequestMethod string    `gorm:"size:10" json:"request_method"`
	RequestPath   string    `gorm:"size:200" json:"request_path"`
	RequestBody   string    `gorm:"type:text" json:"request_body"`
	RequestIP     string    `gorm:"column:request_ip;size:50" json:"request_ip"`
	UserAgent     string    `gorm:"size:500" json:"user_agent"`
	CreatedAt     time.Time `json:"created_at"`
}

func (AdminActionLog) TableName() string { return "admin_action_logs" }

type Upload struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID     uint64    `gorm:"not null;index" json:"tenant_id"`
	FileKey      string    `gorm:"size:255;not null" json:"file_key"`
	OriginalName string    `gorm:"size:255;not null" json:"original_name"`
	FileSize     int64     `gorm:"not null" json:"file_size"`
	FileType     string    `gorm:"size:50;not null" json:"file_type"`
	FileExt      string    `gorm:"size:10;not null" json:"file_ext"`
	StorageType  string    `gorm:"size:20;not null;default:'local'" json:"storage_type"`
	StorageURL   string    `gorm:"column:storage_url;size:500;not null" json:"storage_url"`
	UploadedBy   uint64    `gorm:"index" json:"uploaded_by"`
	CreatedAt    time.Time `json:"created_at"`
}

func (Upload) TableName() string { return "uploads" }
