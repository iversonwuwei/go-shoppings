package service

import (
	"context"

	"wechat-mall-saas/internal/model"
	apperr "wechat-mall-saas/internal/pkg/errors"
	"wechat-mall-saas/internal/repository"
)

type ProductService struct {
	products   *repository.ProductRepo
	skus       *repository.ProductSKURepo
	categories *repository.CategoryRepo
	tenants    *TenantService
}

func NewProductService(p *repository.ProductRepo, s *repository.ProductSKURepo, c *repository.CategoryRepo, t *TenantService) *ProductService {
	return &ProductService{products: p, skus: s, categories: c, tenants: t}
}

func (s *ProductService) List(ctx context.Context, q repository.ProductListQuery) ([]model.Product, int64, error) {
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.Size <= 0 || q.Size > 100 {
		q.Size = 20
	}
	return s.products.List(ctx, q)
}

func (s *ProductService) Detail(ctx context.Context, id uint64) (*model.Product, []model.ProductSKU, error) {
	p, err := s.products.FindByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	if p == nil {
		return nil, nil, apperr.ErrNotFound
	}
	skus, err := s.skus.ListByProduct(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	// 浏览量 +1（异步/容错）
	_ = s.products.UpdateFields(ctx, id, map[string]interface{}{"view_count": p.ViewCount + 1})
	return p, skus, nil
}

func (s *ProductService) Create(ctx context.Context, p *model.Product) error {
	if p.CategoryID != nil && *p.CategoryID == 0 {
		p.CategoryID = nil
	}
	count, err := s.products.Count(ctx)
	if err != nil {
		return err
	}
	if err := s.tenants.CheckProductLimit(ctx, count); err != nil {
		return err
	}
	if p.HasSKU == 1 {
		if err := s.tenants.RequireFeature(ctx, FeatureMultiSKU); err != nil {
			return err
		}
	}
	if p.IsVirtual == 1 {
		if err := s.tenants.RequireFeature(ctx, FeatureVirtualProduct); err != nil {
			return err
		}
	}
	return s.products.Create(ctx, p)
}

func (s *ProductService) Update(ctx context.Context, p *model.Product) error {
	if p.CategoryID != nil && *p.CategoryID == 0 {
		p.CategoryID = nil
	}
	if p.IsVirtual == 1 {
		if err := s.tenants.RequireFeature(ctx, FeatureVirtualProduct); err != nil {
			return err
		}
	}
	return s.products.Update(ctx, p)
}

func (s *ProductService) UpdateStatus(ctx context.Context, id uint64, status int8) error {
	return s.products.UpdateFields(ctx, id, map[string]interface{}{"status": status})
}

func (s *ProductService) Delete(ctx context.Context, id uint64) error {
	return s.products.Delete(ctx, id)
}

func (s *ProductService) CreateSKU(ctx context.Context, sku *model.ProductSKU) error {
	if err := s.tenants.RequireFeature(ctx, FeatureMultiSKU); err != nil {
		return err
	}
	return s.skus.Create(ctx, sku)
}

// ===== Category =====

type CategoryService struct {
	repo *repository.CategoryRepo
}

func NewCategoryService(r *repository.CategoryRepo) *CategoryService {
	return &CategoryService{repo: r}
}

func (s *CategoryService) List(ctx context.Context) ([]model.ProductCategory, error) {
	return s.repo.List(ctx)
}

// ListAll 平台端：包含未启用
func (s *CategoryService) ListAll(ctx context.Context) ([]model.ProductCategory, error) {
	return s.repo.ListAll(ctx)
}

func (s *CategoryService) Create(ctx context.Context, c *model.ProductCategory) error {
	return s.repo.Create(ctx, c)
}

func (s *CategoryService) Update(ctx context.Context, c *model.ProductCategory) error {
	return s.repo.Update(ctx, c)
}

func (s *CategoryService) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}
