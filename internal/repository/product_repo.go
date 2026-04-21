package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"wechat-mall-saas/internal/model"
)

type ProductRepo struct{ db *gorm.DB }

func NewProductRepo(db *gorm.DB) *ProductRepo { return &ProductRepo{db: db} }

type ProductListQuery struct {
	CategoryID  uint64
	Keyword     string
	Status      *int8
	IsRecommend *int8
	IsHot       *int8
	Page        int
	Size        int
	OrderBy     string
}

func (r *ProductRepo) List(ctx context.Context, q ProductListQuery) ([]model.Product, int64, error) {
	tx := TenantDB(ctx, r.db).Model(&model.Product{})
	if q.CategoryID > 0 {
		tx = tx.Where("category_id = ?", q.CategoryID)
	}
	if q.Status != nil {
		tx = tx.Where("status = ?", *q.Status)
	}
	if q.IsRecommend != nil {
		tx = tx.Where("is_recommend = ?", *q.IsRecommend)
	}
	if q.IsHot != nil {
		tx = tx.Where("is_hot = ?", *q.IsHot)
	}
	if q.Keyword != "" {
		tx = tx.Where("name ILIKE ?", "%"+q.Keyword+"%")
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	order := "sort DESC, id DESC"
	if q.OrderBy != "" {
		order = q.OrderBy
	}
	var rows []model.Product
	if err := tx.Order(order).Offset((q.Page - 1) * q.Size).Limit(q.Size).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *ProductRepo) FindByID(ctx context.Context, id uint64) (*model.Product, error) {
	var p model.Product
	if err := TenantDB(ctx, r.db).First(&p, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *ProductRepo) Create(ctx context.Context, p *model.Product) error {
	p.TenantID = EnsureTenant(ctx)
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *ProductRepo) Update(ctx context.Context, p *model.Product) error {
	return TenantDB(ctx, r.db).Save(p).Error
}

func (r *ProductRepo) UpdateFields(ctx context.Context, id uint64, fields map[string]interface{}) error {
	return TenantDB(ctx, r.db).Model(&model.Product{}).Where("id = ?", id).Updates(fields).Error
}

func (r *ProductRepo) Delete(ctx context.Context, id uint64) error {
	return TenantDB(ctx, r.db).Delete(&model.Product{}, id).Error
}

func (r *ProductRepo) Count(ctx context.Context) (int64, error) {
	var n int64
	err := TenantDB(ctx, r.db).Model(&model.Product{}).Count(&n).Error
	return n, err
}

// DecreaseStock 扣减库存（乐观锁：stock >= qty）
func (r *ProductRepo) DecreaseStock(ctx context.Context, tx *gorm.DB, productID uint64, qty int) error {
	res := tx.Model(&model.Product{}).
		Where("id = ? AND tenant_id = ? AND stock >= ?", productID, EnsureTenant(ctx), qty).
		UpdateColumn("stock", gorm.Expr("stock - ?", qty))
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("stock shortage")
	}
	return nil
}

type ProductSKURepo struct{ db *gorm.DB }

func NewProductSKURepo(db *gorm.DB) *ProductSKURepo { return &ProductSKURepo{db: db} }

func (r *ProductSKURepo) ListByProduct(ctx context.Context, productID uint64) ([]model.ProductSKU, error) {
	var rows []model.ProductSKU
	err := TenantDB(ctx, r.db).Where("product_id = ?", productID).Find(&rows).Error
	return rows, err
}

func (r *ProductSKURepo) FindByID(ctx context.Context, id uint64) (*model.ProductSKU, error) {
	var s model.ProductSKU
	if err := TenantDB(ctx, r.db).First(&s, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *ProductSKURepo) Create(ctx context.Context, s *model.ProductSKU) error {
	s.TenantID = EnsureTenant(ctx)
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *ProductSKURepo) Update(ctx context.Context, s *model.ProductSKU) error {
	return TenantDB(ctx, r.db).Save(s).Error
}

func (r *ProductSKURepo) DecreaseStock(ctx context.Context, tx *gorm.DB, skuID uint64, qty int) error {
	res := tx.Model(&model.ProductSKU{}).
		Where("id = ? AND tenant_id = ? AND stock >= ?", skuID, EnsureTenant(ctx), qty).
		UpdateColumn("stock", gorm.Expr("stock - ?", qty))
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("sku stock shortage")
	}
	return nil
}

type CategoryRepo struct{ db *gorm.DB }

func NewCategoryRepo(db *gorm.DB) *CategoryRepo { return &CategoryRepo{db: db} }

func (r *CategoryRepo) List(ctx context.Context) ([]model.ProductCategory, error) {
	var rows []model.ProductCategory
	// 分类改为由平台统一管理（tenant_id = 0），所有租户共享
	err := r.db.WithContext(ctx).Where("tenant_id = 0 AND status = 1").Order("sort DESC, id ASC").Find(&rows).Error
	return rows, err
}

// ListAll 平台端列表（含未启用）
func (r *CategoryRepo) ListAll(ctx context.Context) ([]model.ProductCategory, error) {
	var rows []model.ProductCategory
	err := r.db.WithContext(ctx).Where("tenant_id = 0").Order("sort DESC, id ASC").Find(&rows).Error
	return rows, err
}

func (r *CategoryRepo) FindByID(ctx context.Context, id uint64) (*model.ProductCategory, error) {
	var c model.ProductCategory
	if err := r.db.WithContext(ctx).Where("tenant_id = 0").First(&c, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func (r *CategoryRepo) Create(ctx context.Context, c *model.ProductCategory) error {
	c.TenantID = 0
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *CategoryRepo) Update(ctx context.Context, c *model.ProductCategory) error {
	c.TenantID = 0
	return r.db.WithContext(ctx).Where("tenant_id = 0").Save(c).Error
}

func (r *CategoryRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Where("tenant_id = 0").Delete(&model.ProductCategory{}, id).Error
}
