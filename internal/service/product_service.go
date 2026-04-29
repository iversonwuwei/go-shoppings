package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"

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

type ProductImportError struct {
	Row     int    `json:"row"`
	Message string `json:"message"`
}

type ProductImportResult struct {
	Imported int                  `json:"imported"`
	Errors   []ProductImportError `json:"errors"`
}

type InventoryAdjustInput struct {
	ChangeType   string `json:"change_type"`
	Quantity     int    `json:"quantity"`
	StockWarning *int   `json:"stock_warning"`
	Remark       string `json:"remark"`
}

type InventoryAdjustResult struct {
	Product       *model.Product `json:"product"`
	ChangeType    string         `json:"change_type"`
	Quantity      int            `json:"quantity"`
	BeforeStock   int            `json:"before_stock"`
	AfterStock    int            `json:"after_stock"`
	BeforeWarning int            `json:"before_warning"`
	AfterWarning  int            `json:"after_warning"`
	Remark        string         `json:"remark"`
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
	if repository.EnsureTenant(ctx) == 0 {
		return apperr.ErrTenantRequired
	}
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
	p.Stock = 0
	if p.IsVirtual == 1 {
		p.StockWarning = 0
	} else {
		p.StockWarning = 10
	}
	return s.products.Create(ctx, p)
}

func (s *ProductService) Update(ctx context.Context, p *model.Product) error {
	if repository.EnsureTenant(ctx) == 0 {
		return apperr.ErrTenantRequired
	}
	if p.CategoryID != nil && *p.CategoryID == 0 {
		p.CategoryID = nil
	}
	if p.IsVirtual == 1 {
		if err := s.tenants.RequireFeature(ctx, FeatureVirtualProduct); err != nil {
			return err
		}
	}
	if p.ID == 0 {
		return apperr.ErrParamInvalid
	}
	current, err := s.products.FindByID(ctx, p.ID)
	if err != nil {
		return err
	}
	if current == nil {
		return apperr.ErrNotFound
	}
	fields := map[string]interface{}{
		"category_id":     p.CategoryID,
		"name":            p.Name,
		"subtitle":        p.Subtitle,
		"cover_image":     p.CoverImage,
		"images":          p.Images,
		"video_url":       p.VideoURL,
		"description":     p.Description,
		"detail_images":   p.DetailImages,
		"price":           p.Price,
		"cost_price":      p.CostPrice,
		"has_sku":         p.HasSKU,
		"is_virtual":      p.IsVirtual,
		"delivery_type":   p.DeliveryType,
		"delivery_fee":    p.DeliveryFee,
		"status":          p.Status,
		"is_recommend":    p.IsRecommend,
		"is_hot":          p.IsHot,
		"seo_title":       p.SEOTitle,
		"seo_keywords":    p.SEOKeywords,
		"seo_description": p.SEODescription,
		"sort":            p.Sort,
		"updated_at":      time.Now(),
	}
	return s.products.UpdateFields(ctx, p.ID, fields)
}

func (s *ProductService) ListInventory(ctx context.Context, q repository.ProductListQuery) ([]model.Product, int64, error) {
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.Size <= 0 || q.Size > 100 {
		q.Size = 20
	}
	q.OrderBy = "stock ASC, id DESC"
	return s.products.List(ctx, q)
}

func (s *ProductService) AdjustInventory(ctx context.Context, id uint64, input InventoryAdjustInput) (*InventoryAdjustResult, error) {
	if repository.EnsureTenant(ctx) == 0 {
		return nil, apperr.ErrTenantRequired
	}
	if id == 0 {
		return nil, apperr.ErrParamInvalid
	}
	p, err := s.products.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, apperr.ErrNotFound
	}
	if p.IsVirtual == 1 {
		return nil, apperr.New(20001, "虚拟商品无需库存管理")
	}

	changeType := strings.ToLower(strings.TrimSpace(input.ChangeType))
	if changeType == "" {
		changeType = "set"
	}
	beforeStock := p.Stock
	afterStock := beforeStock
	switch changeType {
	case "in", "increase":
		if input.Quantity <= 0 {
			return nil, apperr.New(20001, "入库数量必须大于 0")
		}
		afterStock = beforeStock + input.Quantity
		changeType = "in"
	case "out", "decrease":
		if input.Quantity <= 0 {
			return nil, apperr.New(20001, "出库数量必须大于 0")
		}
		if input.Quantity > beforeStock {
			return nil, apperr.ErrStockShortage
		}
		afterStock = beforeStock - input.Quantity
		changeType = "out"
	case "set", "check":
		if input.Quantity < 0 {
			return nil, apperr.New(20001, "盘点库存不能小于 0")
		}
		afterStock = input.Quantity
		changeType = "set"
	case "warning":
		changeType = "warning"
	default:
		return nil, apperr.New(20001, "库存调整类型不正确")
	}

	beforeWarning := p.StockWarning
	afterWarning := beforeWarning
	fields := map[string]interface{}{
		"updated_at": time.Now(),
	}
	if changeType != "warning" {
		fields["stock"] = afterStock
	}
	if input.StockWarning != nil {
		if *input.StockWarning < 0 {
			return nil, apperr.New(20001, "预警库存不能小于 0")
		}
		afterWarning = *input.StockWarning
		fields["stock_warning"] = afterWarning
	}
	if len(fields) == 1 {
		return nil, apperr.New(20001, "没有可保存的库存变更")
	}
	if err := s.products.UpdateFields(ctx, id, fields); err != nil {
		return nil, err
	}
	p.Stock = afterStock
	p.StockWarning = afterWarning
	p.UpdatedAt = time.Now()
	return &InventoryAdjustResult{
		Product:       p,
		ChangeType:    changeType,
		Quantity:      input.Quantity,
		BeforeStock:   beforeStock,
		AfterStock:    afterStock,
		BeforeWarning: beforeWarning,
		AfterWarning:  afterWarning,
		Remark:        strings.TrimSpace(input.Remark),
	}, nil
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

func (s *ProductService) ImportCSV(ctx context.Context, file io.Reader) (*ProductImportResult, error) {
	if repository.EnsureTenant(ctx) == 0 {
		return nil, apperr.ErrTenantRequired
	}
	raw, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	raw = bytes.TrimPrefix(raw, []byte{0xEF, 0xBB, 0xBF})

	reader := csv.NewReader(bytes.NewReader(raw))
	reader.FieldsPerRecord = -1
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, apperr.New(20001, "导入文件格式错误，请使用模板填写后上传")
	}
	if len(rows) < 2 {
		return nil, apperr.New(20001, "导入文件为空，请至少填写一行商品数据")
	}
	if len(rows) > 1001 {
		return nil, apperr.New(20001, "单次最多导入 1000 条商品")
	}

	headerIndex := make(map[string]int, len(rows[0]))
	for idx, title := range rows[0] {
		headerIndex[normalizeImportHeader(title)] = idx
	}

	categories, err := s.categories.List(ctx)
	if err != nil {
		return nil, err
	}
	categoryMap := make(map[string]uint64, len(categories))
	for _, item := range categories {
		categoryMap[strings.ToLower(strings.TrimSpace(item.Name))] = item.ID
	}

	result := &ProductImportResult{Errors: make([]ProductImportError, 0)}
	for idx, row := range rows[1:] {
		line := idx + 2
		if importRowEmpty(row) {
			continue
		}
		product, rowErr := buildImportProduct(row, headerIndex, categoryMap)
		if rowErr != nil {
			result.Errors = append(result.Errors, ProductImportError{Row: line, Message: rowErr.Error()})
			continue
		}
		if err := s.Create(ctx, product); err != nil {
			result.Errors = append(result.Errors, ProductImportError{Row: line, Message: err.Error()})
			continue
		}
		result.Imported++
	}
	return result, nil
}

func buildImportProduct(row []string, headerIndex map[string]int, categoryMap map[string]uint64) (*model.Product, error) {
	name := importCell(row, headerIndex, "商品名称", "name")
	if name == "" {
		return nil, fmt.Errorf("商品名称不能为空")
	}

	price, err := importDecimal(importCell(row, headerIndex, "价格", "price"), true, "价格")
	if err != nil {
		return nil, err
	}
	deliveryFee, err := importDecimal(importCell(row, headerIndex, "运费", "delivery_fee"), false, "运费")
	if err != nil {
		return nil, err
	}

	sort, err := importInt(importCell(row, headerIndex, "排序值", "sort"), 0, "排序值")
	if err != nil {
		return nil, err
	}

	isVirtual, err := importBoolToInt(importCell(row, headerIndex, "是否虚拟商品", "is_virtual"), 0, "是否虚拟商品")
	if err != nil {
		return nil, err
	}
	status, err := importBoolToInt(importCell(row, headerIndex, "是否上架", "status"), 1, "是否上架")
	if err != nil {
		return nil, err
	}
	isRecommend, err := importBoolToInt(importCell(row, headerIndex, "是否推荐", "is_recommend"), 0, "是否推荐")
	if err != nil {
		return nil, err
	}
	isHot, err := importBoolToInt(importCell(row, headerIndex, "是否热门", "is_hot"), 0, "是否热门")
	if err != nil {
		return nil, err
	}

	var categoryID *uint64
	if categoryName := strings.TrimSpace(importCell(row, headerIndex, "分类名称", "category_name")); categoryName != "" {
		id, ok := categoryMap[strings.ToLower(categoryName)]
		if !ok {
			return nil, fmt.Errorf("分类不存在：%s", categoryName)
		}
		categoryID = &id
	}

	deliveryTypes := importList(importCell(row, headerIndex, "配送方式", "delivery_type"))
	stockWarning := 10
	if isVirtual == 1 {
		deliveryTypes = nil
		stockWarning = 0
		deliveryFee = decimal.Zero
	} else if len(deliveryTypes) == 0 {
		deliveryTypes = []string{"express"}
	}

	return &model.Product{
		CategoryID:   categoryID,
		Name:         name,
		Subtitle:     importCell(row, headerIndex, "副标题", "subtitle"),
		CoverImage:   importCell(row, headerIndex, "封面图", "cover_image"),
		Images:       model.JSONB(importList(importCell(row, headerIndex, "商品图集", "images"))),
		VideoURL:     importCell(row, headerIndex, "视频地址", "video_url"),
		Description:  importCell(row, headerIndex, "商品详情", "description"),
		DetailImages: model.JSONB(importList(importCell(row, headerIndex, "详情图片", "宣传图片", "detail_images"))),
		Price:        price,
		Stock:        0,
		StockWarning: stockWarning,
		IsVirtual:    int8(isVirtual),
		DeliveryType: model.JSONB(deliveryTypes),
		DeliveryFee:  deliveryFee,
		Status:       int8(status),
		IsRecommend:  int8(isRecommend),
		IsHot:        int8(isHot),
		Sort:         sort,
	}, nil
}

func normalizeImportHeader(v string) string {
	return strings.ToLower(strings.TrimSpace(strings.TrimPrefix(v, "\uFEFF")))
}

func importCell(row []string, headerIndex map[string]int, keys ...string) string {
	for _, key := range keys {
		if idx, ok := headerIndex[normalizeImportHeader(key)]; ok && idx < len(row) {
			return strings.TrimSpace(row[idx])
		}
	}
	return ""
}

func importRowEmpty(row []string) bool {
	for _, item := range row {
		if strings.TrimSpace(item) != "" {
			return false
		}
	}
	return true
}

func importInt(raw string, defaultValue int, field string) (int, error) {
	if strings.TrimSpace(raw) == "" {
		return defaultValue, nil
	}
	n, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return 0, fmt.Errorf("%s格式错误", field)
	}
	if n < 0 {
		return 0, fmt.Errorf("%s不能小于 0", field)
	}
	return n, nil
}

func importDecimal(raw string, required bool, field string) (decimal.Decimal, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		if required {
			return decimal.Zero, fmt.Errorf("%s不能为空", field)
		}
		return decimal.Zero, nil
	}
	v, err := decimal.NewFromString(raw)
	if err != nil {
		return decimal.Zero, fmt.Errorf("%s格式错误", field)
	}
	if v.IsNegative() {
		return decimal.Zero, fmt.Errorf("%s不能小于 0", field)
	}
	return v, nil
}

func importBoolToInt(raw string, defaultValue int, field string) (int, error) {
	raw = strings.ToLower(strings.TrimSpace(raw))
	if raw == "" {
		return defaultValue, nil
	}
	switch raw {
	case "1", "true", "yes", "y", "是", "上架", "推荐", "热门":
		return 1, nil
	case "0", "false", "no", "n", "否", "下架", "普通":
		return 0, nil
	default:
		return 0, fmt.Errorf("%s仅支持 1/0/是/否", field)
	}
}

func importList(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	replacer := strings.NewReplacer("，", "|", "、", "|", ";", "|", "；", "|", ",", "|")
	parts := strings.Split(replacer.Replace(raw), "|")
	result := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, item := range parts {
		key := strings.TrimSpace(item)
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, key)
	}
	return result
}

// ===== Category =====

type CategoryService struct {
	repo  *repository.CategoryRepo
	asset *repository.TenantCategoryAssetRepo
}

func NewCategoryService(r *repository.CategoryRepo, a *repository.TenantCategoryAssetRepo) *CategoryService {
	return &CategoryService{repo: r, asset: a}
}

func (s *CategoryService) List(ctx context.Context) ([]model.ProductCategory, error) {
	rows, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	tid := repository.EnsureTenant(ctx)
	if tid == 0 || s.asset == nil {
		return rows, nil
	}
	assets, err := s.asset.ListByTenant(ctx, tid)
	if err != nil {
		return nil, err
	}
	assetMap := make(map[uint64]model.TenantCategoryAsset, len(assets))
	for _, item := range assets {
		assetMap[item.CategoryID] = item
	}
	for i := range rows {
		if asset, ok := assetMap[rows[i].ID]; ok {
			if asset.Icon != "" {
				rows[i].Icon = asset.Icon
			}
			if asset.CoverImage != "" {
				rows[i].CoverImage = asset.CoverImage
			}
			if asset.Sort != nil {
				rows[i].Sort = *asset.Sort
			}
		}
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].Sort == rows[j].Sort {
			return rows[i].ID < rows[j].ID
		}
		return rows[i].Sort > rows[j].Sort
	})
	return rows, nil
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

func (s *CategoryService) UpdateTenantAsset(ctx context.Context, categoryID uint64, coverImage, icon string, sortValue *int) error {
	tid := repository.EnsureTenant(ctx)
	if tid == 0 {
		return apperr.ErrTenantRequired
	}
	if sortValue != nil && *sortValue < 0 {
		return apperr.New(20001, "排序值不能小于 0")
	}
	existing, err := s.repo.FindByID(ctx, categoryID)
	if err != nil {
		return err
	}
	if existing == nil {
		return apperr.ErrNotFound
	}
	if s.asset == nil {
		return apperr.ErrInternal
	}
	if err := s.asset.UpsertMedia(ctx, &model.TenantCategoryAsset{
		TenantID:   tid,
		CategoryID: categoryID,
		CoverImage: strings.TrimSpace(coverImage),
		Icon:       strings.TrimSpace(icon),
	}); err != nil {
		return err
	}
	if sortValue != nil {
		return s.asset.UpsertSort(ctx, tid, categoryID, *sortValue)
	}
	return nil
}
