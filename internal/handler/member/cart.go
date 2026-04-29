package member

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"wechat-mall-saas/internal/pkg/ctxkeys"
	apperr "wechat-mall-saas/internal/pkg/errors"
	"wechat-mall-saas/internal/pkg/response"
	"wechat-mall-saas/internal/service"
)

type CartHandler struct {
	svc *service.CartService
}

func NewCartHandler(s *service.CartService) *CartHandler { return &CartHandler{svc: s} }

type cartItemReq struct {
	ProductID uint64 `json:"product_id" binding:"required"`
	SKUID     uint64 `json:"sku_id"`
	Quantity  int    `json:"quantity" binding:"required,min=1"`
}

type cartQuantityReq struct {
	Quantity int `json:"quantity" binding:"required"`
}

type cartDeleteReq struct {
	Keys []string `json:"keys" binding:"required,min=1"`
}

func currentMemberID(c *gin.Context) (uint64, bool) {
	m := ctxkeys.GetMember(c.Request.Context())
	if m == nil || m.ID == 0 {
		response.Fail(c, apperr.ErrUnauthorized)
		return 0, false
	}
	return m.ID, true
}

func (h *CartHandler) List(c *gin.Context) {
	memberID, ok := currentMemberID(c)
	if !ok {
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	keys := splitKeys(c.Query("keys"))
	result, err := h.svc.List(c.Request.Context(), memberID, service.CartListInput{Keys: keys, Page: page, Size: size})
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, result)
}

func (h *CartHandler) Add(c *gin.Context) {
	memberID, ok := currentMemberID(c)
	if !ok {
		return
	}
	var req cartItemReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	item, err := h.svc.Add(c.Request.Context(), memberID, req.ProductID, req.SKUID, req.Quantity)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, item)
}

func (h *CartHandler) UpdateQuantity(c *gin.Context) {
	memberID, ok := currentMemberID(c)
	if !ok {
		return
	}
	productID, skuID, err := service.ParseCartKey(c.Param("key"))
	if err != nil {
		response.Fail(c, err)
		return
	}
	var req cartQuantityReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	item, err := h.svc.UpdateQuantity(c.Request.Context(), memberID, productID, skuID, req.Quantity)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, item)
}

func (h *CartHandler) Delete(c *gin.Context) {
	memberID, ok := currentMemberID(c)
	if !ok {
		return
	}
	if err := h.svc.Delete(c.Request.Context(), memberID, c.Param("key")); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, nil)
}

func (h *CartHandler) DeleteMany(c *gin.Context) {
	memberID, ok := currentMemberID(c)
	if !ok {
		return
	}
	var req cartDeleteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	if err := h.svc.DeleteMany(c.Request.Context(), memberID, req.Keys); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, nil)
}

func (h *CartHandler) Clear(c *gin.Context) {
	memberID, ok := currentMemberID(c)
	if !ok {
		return
	}
	if err := h.svc.Clear(c.Request.Context(), memberID); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, nil)
}

func splitKeys(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	keys := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			keys = append(keys, part)
		}
	}
	return keys
}
