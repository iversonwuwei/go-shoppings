package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"wechat-mall-saas/internal/model"
)

type MemberLevelRepo struct{ db *gorm.DB }

func NewMemberLevelRepo(db *gorm.DB) *MemberLevelRepo { return &MemberLevelRepo{db: db} }

// List 返回当前租户的会员等级，按 min_growth ASC 排序（低级在前）。
func (r *MemberLevelRepo) List(ctx context.Context) ([]model.MemberLevel, error) {
	var out []model.MemberLevel
	err := TenantDB(ctx, r.db).Order("min_growth ASC, sort ASC, id ASC").Find(&out).Error
	return out, err
}

func (r *MemberLevelRepo) FindByID(ctx context.Context, id uint64) (*model.MemberLevel, error) {
	var m model.MemberLevel
	if err := TenantDB(ctx, r.db).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func (r *MemberLevelRepo) Create(ctx context.Context, m *model.MemberLevel) error {
	m.TenantID = EnsureTenant(ctx)
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *MemberLevelRepo) Update(ctx context.Context, m *model.MemberLevel) error {
	// 限定 tenant_id，防止跨租户更新
	return TenantDB(ctx, r.db).Save(m).Error
}

func (r *MemberLevelRepo) Delete(ctx context.Context, id uint64) error {
	return TenantDB(ctx, r.db).Delete(&model.MemberLevel{}, id).Error
}

// FindByGrowth 返回 min_growth <= growth 的最高等级（用于会员自动升级）。
func (r *MemberLevelRepo) FindByGrowth(ctx context.Context, growth int) (*model.MemberLevel, error) {
	var m model.MemberLevel
	err := TenantDB(ctx, r.db).
		Where("min_growth <= ?", growth).
		Order("min_growth DESC").
		First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}
