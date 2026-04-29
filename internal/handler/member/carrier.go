package member

import (
	"github.com/gin-gonic/gin"

	"wechat-mall-saas/internal/pkg/response"
	"wechat-mall-saas/internal/service"
)

type CarrierHandler struct {
	svc *service.SettingsService
}

func NewCarrierHandler(svc *service.SettingsService) *CarrierHandler {
	return &CarrierHandler{svc: svc}
}

func (h *CarrierHandler) List(c *gin.Context) {
	rows, err := h.svc.ListCarrierOptionsForTenant(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, rows)
}

func (h *CarrierHandler) Match(c *gin.Context) {
	out, err := h.svc.MatchCarrierByTrackingNo(c.Request.Context(), c.Query("tracking_no"))
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, out)
}
