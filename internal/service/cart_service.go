package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"

	"wechat-mall-saas/internal/model"
	apperr "wechat-mall-saas/internal/pkg/errors"
	"wechat-mall-saas/internal/repository"
)

type CartService struct {
	carts    *repository.CartRepo
	products *repository.ProductRepo
	skus     *repository.ProductSKURepo
}

func NewCartService(carts *repository.CartRepo, products *repository.ProductRepo, skus *repository.ProductSKURepo) *CartService {
	return &CartService{carts: carts, products: products, skus: skus}
}

type CartItemDTO struct {
	Key           string          `json:"key"`
	ProductID     uint64          `json:"product_id"`
	SKUID         uint64          `json:"sku_id"`
	ProductName   string          `json:"product_name"`
	CoverImage    string          `json:"cover_image"`
	Price         decimal.Decimal `json:"price"`
	Quantity      int             `json:"quantity"`
	Stock         int             `json:"stock"`
	SKUDesc       string          `json:"sku_desc"`
	IsVirtual     bool            `json:"is_virtual"`
	DeliveryTypes []string        `json:"delivery_types"`
}

type CartListInput struct {
	Keys []string
	Page int
	Size int
}

type CartListResult struct {
	List  []CartItemDTO `json:"list"`
	Total int64         `json:"total"`
	Page  int           `json:"page"`
	Size  int           `json:"size"`
}

type cartResolved struct {
	product    *model.Product
	sku        *model.ProductSKU
	price      decimal.Decimal
	stock      int
	coverImage string
	skuDesc    string
}

func cartKey(productID, skuID uint64) string {
	return fmt.Sprintf("%d-%d", productID, skuID)
}

func ParseCartKey(key string) (uint64, uint64, error) {
	parts := strings.Split(strings.TrimSpace(key), "-")
	if len(parts) != 2 {
		return 0, 0, apperr.ErrParamInvalid
	}
	productID, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil || productID == 0 {
		return 0, 0, apperr.ErrParamInvalid
	}
	skuID, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return 0, 0, apperr.ErrParamInvalid
	}
	return productID, skuID, nil
}

func normalizeCartPage(page, size int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 100 {
		size = 20
	}
	return page, size
}

func parseCartKeyPairs(keys []string) ([]repository.CartKeyPair, error) {
	if len(keys) == 0 {
		return nil, nil
	}
	pairs := make([]repository.CartKeyPair, 0, len(keys))
	seen := make(map[string]bool, len(keys))
	for _, key := range keys {
		key = strings.TrimSpace(key)
		if key == "" || seen[key] {
			continue
		}
		productID, skuID, err := ParseCartKey(key)
		if err != nil {
			return nil, err
		}
		pairs = append(pairs, repository.CartKeyPair{ProductID: productID, SKUID: skuID})
		seen[key] = true
	}
	return pairs, nil
}

func (s *CartService) List(ctx context.Context, memberID uint64, in CartListInput) (*CartListResult, error) {
	in.Page, in.Size = normalizeCartPage(in.Page, in.Size)
	keyPairs, err := parseCartKeyPairs(in.Keys)
	if err != nil {
		return nil, err
	}
	rows, total, err := s.carts.List(ctx, repository.CartListQuery{
		MemberID: memberID,
		Keys:     keyPairs,
		Page:     in.Page,
		Size:     in.Size,
	})
	if err != nil {
		return nil, err
	}
	out := make([]CartItemDTO, 0, len(rows))
	for _, row := range rows {
		resolved, err := s.resolve(ctx, row.ProductID, row.SKUID)
		if err != nil {
			return nil, err
		}
		if resolved == nil {
			continue
		}
		out = append(out, toCartItemDTO(row, resolved))
	}
	return &CartListResult{List: out, Total: total, Page: in.Page, Size: in.Size}, nil
}

func (s *CartService) Add(ctx context.Context, memberID, productID, skuID uint64, quantity int) (*CartItemDTO, error) {
	if quantity <= 0 {
		return nil, apperr.ErrParamInvalid
	}
	resolved, err := s.resolve(ctx, productID, skuID)
	if err != nil {
		return nil, err
	}
	if resolved == nil {
		return nil, apperr.New(20004, "商品不存在或已下架")
	}
	existing, err := s.carts.Find(ctx, memberID, productID, skuID)
	if err != nil {
		return nil, err
	}
	nextQuantity := quantity
	if existing != nil {
		nextQuantity += existing.Quantity
	}
	if err := ensureCartStock(resolved, nextQuantity); err != nil {
		return nil, err
	}
	item := existing
	if item == nil {
		item = &model.MemberCartItem{MemberID: memberID, ProductID: productID, SKUID: skuID}
	}
	item.Quantity = nextQuantity
	if err := s.carts.SaveQuantity(ctx, item); err != nil {
		return nil, err
	}
	dto := toCartItemDTO(*item, resolved)
	return &dto, nil
}

func (s *CartService) UpdateQuantity(ctx context.Context, memberID, productID, skuID uint64, quantity int) (*CartItemDTO, error) {
	if quantity <= 0 {
		if err := s.carts.Delete(ctx, memberID, productID, skuID); err != nil {
			return nil, err
		}
		return nil, nil
	}
	resolved, err := s.resolve(ctx, productID, skuID)
	if err != nil {
		return nil, err
	}
	if resolved == nil {
		return nil, apperr.New(20004, "商品不存在或已下架")
	}
	if err := ensureCartStock(resolved, quantity); err != nil {
		return nil, err
	}
	item, err := s.carts.Find(ctx, memberID, productID, skuID)
	if err != nil {
		return nil, err
	}
	if item == nil {
		item = &model.MemberCartItem{MemberID: memberID, ProductID: productID, SKUID: skuID}
	}
	item.Quantity = quantity
	if err := s.carts.SaveQuantity(ctx, item); err != nil {
		return nil, err
	}
	dto := toCartItemDTO(*item, resolved)
	return &dto, nil
}

func (s *CartService) Delete(ctx context.Context, memberID uint64, key string) error {
	productID, skuID, err := ParseCartKey(key)
	if err != nil {
		return err
	}
	return s.carts.Delete(ctx, memberID, productID, skuID)
}

func (s *CartService) DeleteMany(ctx context.Context, memberID uint64, keys []string) error {
	for _, key := range keys {
		if strings.TrimSpace(key) == "" {
			continue
		}
		if err := s.Delete(ctx, memberID, key); err != nil {
			return err
		}
	}
	return nil
}

func (s *CartService) Clear(ctx context.Context, memberID uint64) error {
	return s.carts.Clear(ctx, memberID)
}

func (s *CartService) resolve(ctx context.Context, productID, skuID uint64) (*cartResolved, error) {
	product, err := s.products.FindByID(ctx, productID)
	if err != nil {
		return nil, err
	}
	if product == nil || product.Status != 1 {
		return nil, nil
	}
	resolved := &cartResolved{
		product:    product,
		price:      product.Price,
		stock:      product.Stock,
		coverImage: product.CoverImage,
	}
	if product.HasSKU == 1 && skuID == 0 {
		return nil, apperr.New(20005, "请选择商品规格")
	}
	if skuID > 0 {
		sku, err := s.skus.FindByID(ctx, skuID)
		if err != nil {
			return nil, err
		}
		if sku == nil || sku.ProductID != product.ID || sku.Status != 1 {
			return nil, nil
		}
		resolved.sku = sku
		resolved.price = sku.Price
		resolved.stock = sku.Stock
		resolved.coverImage = firstNonEmpty(sku.Image, product.CoverImage)
		resolved.skuDesc = string(sku.Attributes)
	}
	return resolved, nil
}

func ensureCartStock(resolved *cartResolved, quantity int) error {
	if resolved.product.IsVirtual == 1 {
		return nil
	}
	if resolved.stock < quantity {
		return apperr.ErrStockShortage
	}
	return nil
}

func toCartItemDTO(item model.MemberCartItem, resolved *cartResolved) CartItemDTO {
	return CartItemDTO{
		Key:           cartKey(item.ProductID, item.SKUID),
		ProductID:     item.ProductID,
		SKUID:         item.SKUID,
		ProductName:   resolved.product.Name,
		CoverImage:    resolved.coverImage,
		Price:         resolved.price,
		Quantity:      item.Quantity,
		Stock:         resolved.stock,
		SKUDesc:       resolved.skuDesc,
		IsVirtual:     resolved.product.IsVirtual == 1,
		DeliveryTypes: []string(resolved.product.DeliveryType),
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
