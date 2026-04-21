package member

import (
	"github.com/gin-gonic/gin"

	"wechat-mall-saas/internal/pkg/ctxkeys"
	apperr "wechat-mall-saas/internal/pkg/errors"
	"wechat-mall-saas/internal/pkg/response"
	"wechat-mall-saas/internal/service"
)

type MemberHandler struct {
	svc *service.MemberService
}

func NewMemberHandler(s *service.MemberService) *MemberHandler { return &MemberHandler{svc: s} }

func (h *MemberHandler) Profile(c *gin.Context) {
	m := ctxkeys.GetMember(c.Request.Context())
	if m == nil {
		response.Fail(c, apperr.ErrUnauthorized)
		return
	}
	p, err := h.svc.Profile(c.Request.Context(), m.ID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, p)
}

func (h *MemberHandler) UpdateProfile(c *gin.Context) {
	m := ctxkeys.GetMember(c.Request.Context())
	if m == nil {
		response.Fail(c, apperr.ErrUnauthorized)
		return
	}
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	if err := h.svc.UpdateProfile(c.Request.Context(), m.ID, body); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, nil)
}
