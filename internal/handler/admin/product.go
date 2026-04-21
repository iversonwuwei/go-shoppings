package admin

import (
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
