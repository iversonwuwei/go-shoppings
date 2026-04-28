package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"wechat-mall-saas/internal/model"
)

type RegionRepo struct{ db *gorm.DB }

func NewRegionRepo(db *gorm.DB) *RegionRepo { return &RegionRepo{db: db} }

func (r *RegionRepo) List(ctx context.Context, includeDisabled bool) ([]model.Region, error) {
	var rows []model.Region
	tx := r.db.WithContext(ctx).Model(&model.Region{})
	if !includeDisabled {
		tx = tx.Where("enabled = 1")
	}
	err := tx.Order("level ASC, sort DESC, id ASC").Find(&rows).Error
	return rows, err
}

func (r *RegionRepo) Tree(ctx context.Context, includeDisabled bool) ([]model.Region, error) {
	rows, err := r.List(ctx, includeDisabled)
	if err != nil {
		return nil, err
	}
	return BuildRegionTree(rows), nil
}

func BuildRegionTree(rows []model.Region) []model.Region {
	byParent := make(map[uint64][]model.Region)
	for _, row := range rows {
		row.Children = nil
		byParent[row.ParentID] = append(byParent[row.ParentID], row)
	}
	var build func(uint64) []model.Region
	build = func(parentID uint64) []model.Region {
		items := byParent[parentID]
		for i := range items {
			items[i].Children = build(items[i].ID)
		}
		return items
	}
	return build(0)
}

func (r *RegionRepo) FindByID(ctx context.Context, id uint64) (*model.Region, error) {
	var row model.Region
	if err := r.db.WithContext(ctx).First(&row, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *RegionRepo) Create(ctx context.Context, row *model.Region) error {
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *RegionRepo) Update(ctx context.Context, row *model.Region) error {
	return r.db.WithContext(ctx).Save(row).Error
}

func (r *RegionRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.Region{}, id).Error
}

func (r *RegionRepo) CountChildren(ctx context.Context, id uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Region{}).Where("parent_id = ?", id).Count(&count).Error
	return count, err
}
