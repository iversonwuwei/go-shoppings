package member

import (
	"strconv"

	"github.com/gin-gonic/gin"

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
	onShelf := int8(1)
	q := repository.ProductListQuery{
		CategoryID: categoryID,
		Keyword:    c.Query("keyword"),
		Page:       page,
		Size:       size,
		Status:     &onShelf,
	}
	rows, total, err := h.svc.List(c.Request.Context(), q)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"list": rows, "total": total, "page": page, "size": size})
}

func (h *ProductHandler) Detail(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	p, skus, err := h.svc.Detail(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"product": p, "skus": skus})
}

func (h *ProductHandler) Hot(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	on := int8(1)
	hot := int8(1)
	rows, total, err := h.svc.List(c.Request.Context(), repository.ProductListQuery{
		Page: page, Size: size, Status: &on, IsHot: &hot,
		OrderBy: "sold_count DESC, id DESC",
	})
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"list": rows, "total": total})
}

func (h *ProductHandler) Recommend(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	on := int8(1)
	rc := int8(1)
	rows, total, err := h.svc.List(c.Request.Context(), repository.ProductListQuery{
		Page: page, Size: size, Status: &on, IsRecommend: &rc,
	})
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"list": rows, "total": total})
}

type CategoryHandler struct {
	svc *service.CategoryService
}

func NewCategoryHandler(s *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{svc: s}
}

func (h *CategoryHandler) List(c *gin.Context) {
	rows, err := h.svc.List(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, rows)
}
