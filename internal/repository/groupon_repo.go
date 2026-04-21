package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"wechat-mall-saas/internal/model"
)

type GrouponRepo struct{ db *gorm.DB }

func NewGrouponRepo(db *gorm.DB) *GrouponRepo { return &GrouponRepo{db: db} }

// ListActivities 管理端：分页 + 按 start_at DESC
func (r *GrouponRepo) ListActivities(ctx context.Context, page, size int) ([]model.GrouponActivity, int64, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 200 {
		size = 20
	}
	var rows []model.GrouponActivity
	var total int64
	q := TenantDB(ctx, r.db).Model(&model.GrouponActivity{})
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Order("start_at DESC, id DESC").
		Offset((page - 1) * size).Limit(size).
		Find(&rows).Error
	return rows, total, err
}

// ListActive 会员端：当前时间内 status=1 的活动
func (r *GrouponRepo) ListActive(ctx context.Context) ([]model.GrouponActivity, error) {
	var rows []model.GrouponActivity
	err := TenantDB(ctx, r.db).
		Where("status = 1 AND start_at <= NOW() AND end_at >= NOW()").
		Order("start_at ASC").Find(&rows).Error
	return rows, err
}

func (r *GrouponRepo) FindActivity(ctx context.Context, id uint64) (*model.GrouponActivity, error) {
	var a model.GrouponActivity
	if err := TenantDB(ctx, r.db).First(&a, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}

func (r *GrouponRepo) CreateActivity(ctx context.Context, a *model.GrouponActivity) error {
	a.TenantID = EnsureTenant(ctx)
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *GrouponRepo) UpdateActivity(ctx context.Context, id uint64, fields map[string]interface{}) error {
	return TenantDB(ctx, r.db).Model(&model.GrouponActivity{}).
		Where("id = ?", id).Updates(fields).Error
}

func (r *GrouponRepo) DeleteActivity(ctx context.Context, id uint64) error {
	return TenantDB(ctx, r.db).Delete(&model.GrouponActivity{}, id).Error
}

// ListGroupons 管理端：查看某活动下的团单
func (r *GrouponRepo) ListGroupons(ctx context.Context, activityID uint64, page, size int) ([]model.Groupon, int64, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 200 {
		size = 20
	}
	var rows []model.Groupon
	var total int64
	q := TenantDB(ctx, r.db).Model(&model.Groupon{})
	if activityID > 0 {
		q = q.Where("activity_id = ?", activityID)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Order("id DESC").
		Offset((page - 1) * size).Limit(size).
		Find(&rows).Error
	return rows, total, err
}
