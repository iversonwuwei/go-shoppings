package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"wechat-mall-saas/internal/pkg/ctxkeys"
	apperr "wechat-mall-saas/internal/pkg/errors"
	"wechat-mall-saas/internal/pkg/response"
	"wechat-mall-saas/internal/service"
)

type PaymentHandler struct {
	svc *service.PaymentService
}

func NewPaymentHandler(s *service.PaymentService) *PaymentHandler { return &PaymentHandler{svc: s} }

type createPayReq struct {
	OrderNo string `json:"order_no" binding:"required"`
}

func (h *PaymentHandler) Create(c *gin.Context) {
	m := ctxkeys.GetMember(c.Request.Context())
	if m == nil {
		response.Fail(c, apperr.ErrUnauthorized)
		return
	}
	var req createPayReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	scheme := "https"
	if c.Request.TLS == nil {
		scheme = "http"
	}
	notifyURL := scheme + "://" + c.Request.Host + "/api/v1/payments/callback/wechat"
	res, err := h.svc.Create(c.Request.Context(), m.ID, m.OpenID, req.OrderNo, notifyURL)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, res)
}

// WechatCallback 占位实现：生产需完成验签 + 密钥解密 + 调用 PaymentService.HandleCallback
func (h *PaymentHandler) WechatCallback(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"code": "SUCCESS", "message": "SUCCESS"})
}
