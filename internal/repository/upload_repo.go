package repository

import (
	"context"

	"gorm.io/gorm"

	"wechat-mall-saas/internal/model"
)

type UploadRepo struct{ db *gorm.DB }

func NewUploadRepo(db *gorm.DB) *UploadRepo { return &UploadRepo{db: db} }

func (r *UploadRepo) Create(ctx context.Context, u *model.Upload) error {
	return r.db.WithContext(ctx).Create(u).Error
}
