package admin

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"wechat-mall-saas/internal/pkg/ctxkeys"
	apperr "wechat-mall-saas/internal/pkg/errors"
	"wechat-mall-saas/internal/pkg/response"
	"wechat-mall-saas/internal/service"
)

type SubscriptionHandler struct {
	sub *service.SubscriptionService
}

func NewSubscriptionHandler(s *service.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{sub: s}
}

type createSubOrderReq struct {
	PlanID       uint64 `json:"plan_id"`       // 0 表示续订当前套餐
	BillingCycle string `json:"billing_cycle"` // monthly / yearly
	OpenID       string `json:"openid"`        // JSAPI 支付需要
}

// Create 创建订阅订单并返回微信支付参数
func (h *SubscriptionHandler) Create(c *gin.Context) {
	t := ctxkeys.GetTenant(c.Request.Context())
	if t == nil {
		response.FailCode(c, 10004, "缺少租户标识")
		return
	}
	var req createSubOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	adminInfo := ctxkeys.GetAdmin(c.Request.Context())
	if adminInfo == nil {
		response.Fail(c, apperr.ErrUnauthorized)
		return
	}
	order, pay, err := h.sub.CreateOrder(c.Request.Context(), t.ID, req.PlanID, req.BillingCycle, req.OpenID, adminInfo.ID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"order": order, "pay": pay})
}

// List 商户后台：订阅订单列表
func (h *SubscriptionHandler) List(c *gin.Context) {
	t := ctxkeys.GetTenant(c.Request.Context())
	if t == nil {
		response.FailCode(c, 10004, "缺少租户标识")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	rows, total, err := h.sub.ListOrders(c.Request.Context(), t.ID, page, size)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"list": rows, "total": total, "page": page, "page_size": size})
}

// WxpayCallback 微信支付回调（公开路由，平台统一商户号）
// 目前为简化版：接收 { order_no, transaction_id, paid_at } JSON，生产需替换为真实微信 v3 解密。
func (h *SubscriptionHandler) WxpayCallback(c *gin.Context) {
	var body struct {
		OrderNo       string `json:"order_no" binding:"required"`
		TransactionID string `json:"transaction_id"`
		PaidAt        string `json:"paid_at"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "FAIL", "message": err.Error()})
		return
	}
	paidAt := time.Now()
	if body.PaidAt != "" {
		if t, err := time.Parse(time.RFC3339, body.PaidAt); err == nil {
			paidAt = t
		}
	}
	if err := h.sub.OnPaySuccess(c.Request.Context(), body.OrderNo, body.TransactionID, paidAt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "FAIL", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "SUCCESS", "message": "OK"})
}
