package member

import (
	"github.com/gin-gonic/gin"

	"wechat-mall-saas/internal/pkg/response"
	"wechat-mall-saas/internal/repository"
)

type SeckillHandler struct {
	repo *repository.SeckillRepo
}

func NewSeckillHandler(r *repository.SeckillRepo) *SeckillHandler { return &SeckillHandler{repo: r} }

// List 会员端：列出当前进行中的秒杀活动（含商品）。
func (h *SeckillHandler) List(c *gin.Context) {
	rows, err := h.repo.ListActive(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, rows)
}
