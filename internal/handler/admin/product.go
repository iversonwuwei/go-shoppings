package admin

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"wechat-mall-saas/internal/model"
	"wechat-mall-saas/internal/pkg/response"
	"wechat-mall-saas/internal/repository"
	"wechat-mall-saas/internal/service"
)

type ProductHandler struct {
	svc *service.ProductService
}

func NewProductHandler(s *service.ProductService) *ProductHandler { return &ProductHandler{svc: s} }

func (h *ProductHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	categoryID, _ := strconv.ParseUint(c.Query("category_id"), 10, 64)
	q := repository.ProductListQuery{
		CategoryID: categoryID,
		Keyword:    c.Query("keyword"),
		Page:       page,
		Size:       size,
	}
	rows, total, err := h.svc.List(c.Request.Context(), q)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"list": rows, "total": total, "page": page, "size": size})
}

func (h *ProductHandler) Create(c *gin.Context) {
	var p model.Product
	if err := c.ShouldBindJSON(&p); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	if err := h.svc.Create(c.Request.Context(), &p); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, p)
}

func (h *ProductHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var p model.Product
	if err := c.ShouldBindJSON(&p); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	p.ID = id
	if err := h.svc.Update(c.Request.Context(), &p); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, p)
}

func (h *ProductHandler) UpdateStatus(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var body struct {
		Status int8 `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	if err := h.svc.UpdateStatus(c.Request.Context(), id, body.Status); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, nil)
}

func (h *ProductHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, nil)
}

func (h *ProductHandler) CreateSKU(c *gin.Context) {
	pid, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var sku model.ProductSKU
	if err := c.ShouldBindJSON(&sku); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	sku.ProductID = pid
	if err := h.svc.CreateSKU(c.Request.Context(), &sku); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, sku)
}

func (h *ProductHandler) ImportTemplate(c *gin.Context) {
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", `attachment; filename="product-import-template.csv"`)

	_, _ = c.Writer.Write([]byte{0xEF, 0xBB, 0xBF})
	writer := csv.NewWriter(c.Writer)
	_ = writer.Write([]string{
		"商品名称", "副标题", "分类名称", "封面图", "商品图集", "视频地址", "商品详情",
		"价格", "库存", "预警库存", "是否虚拟商品", "配送方式", "运费", "是否上架", "是否推荐", "是否热门", "排序值",
	})
	writer.Flush()
}

func (h *ProductHandler) Import(c *gin.Context) {
	const maxSize int64 = 5 << 20 // 5MB
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)

	fh, err := c.FormFile("file")
	if err != nil {
		response.FailCode(c, 20001, "请选择要导入的 CSV 文件")
		return
	}
	src, err := fh.Open()
	if err != nil {
		response.Fail(c, err)
		return
	}
	defer src.Close()

	result, err := h.svc.ImportCSV(c.Request.Context(), src)
	if err != nil {
		response.Fail(c, err)
		return
	}
	if result.Imported == 0 && len(result.Errors) > 0 {
		c.JSON(http.StatusBadRequest, response.Body{
			Code:    20001,
			Message: fmt.Sprintf("导入失败，共 %d 行有误", len(result.Errors)),
			Data:    result,
		})
		return
	}
	response.OK(c, result)
}

// ==== Category ====

type CategoryHandler struct {
	svc *service.CategoryService
}

func NewCategoryHandler(s *service.CategoryService) *CategoryHandler { return &CategoryHandler{svc: s} }

func (h *CategoryHandler) List(c *gin.Context) {
	rows, err := h.svc.List(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, rows)
}

type categoryAssetReq struct {
	CoverImage string `json:"cover_image"`
	Icon       string `json:"icon"`
}

func (h *CategoryHandler) UpdateTenantAsset(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.FailCode(c, 20001, "分类不存在")
		return
	}
	var req categoryAssetReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	if err := h.svc.UpdateTenantAsset(c.Request.Context(), id, req.CoverImage, req.Icon); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"id": id, "cover_image": req.CoverImage, "icon": req.Icon})
}

// ListAll 平台端：包含禁用
func (h *CategoryHandler) ListAll(c *gin.Context) {
	rows, err := h.svc.ListAll(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, rows)
}

func (h *CategoryHandler) Create(c *gin.Context) {
	var m model.ProductCategory
	if err := c.ShouldBindJSON(&m); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	if err := h.svc.Create(c.Request.Context(), &m); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, m)
}

func (h *CategoryHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var m model.ProductCategory
	if err := c.ShouldBindJSON(&m); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	m.ID = id
	if err := h.svc.Update(c.Request.Context(), &m); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, m)
}

func (h *CategoryHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, nil)
}
