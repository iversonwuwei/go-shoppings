package admin

import (
	"strconv"

	"github.com/gin-gonic/gin"

	apperr "wechat-mall-saas/internal/pkg/errors"
	"wechat-mall-saas/internal/pkg/response"
	"wechat-mall-saas/internal/repository"
	"wechat-mall-saas/internal/service"
)

type MemberHandler struct {
	svc *service.MemberService
}

func NewMemberHandler(s *service.MemberService) *MemberHandler { return &MemberHandler{svc: s} }

func (h *MemberHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	var status *int8
	if raw := c.Query("status"); raw != "" {
		v, err := strconv.ParseInt(raw, 10, 8)
		if err != nil {
			response.Fail(c, apperr.ErrParamInvalid)
			return
		}
		vv := int8(v)
		status = &vv
	}
	var levelID *uint64
	if raw := c.Query("level_id"); raw != "" {
		v, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			response.Fail(c, apperr.ErrParamInvalid)
			return
		}
		levelID = &v
	}
	rows, total, err := h.svc.AdminList(c.Request.Context(), repository.MemberListQuery{
		Keyword: c.Query("keyword"),
		Status:  status,
		LevelID: levelID,
		Page:    page,
		Size:    size,
	})
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"list": rows, "total": total, "page": page, "size": size})
}

func (h *MemberHandler) Detail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, apperr.ErrParamInvalid)
		return
	}
	detail, err := h.svc.AdminDetail(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, detail)
}

func (h *MemberHandler) UpdateStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, apperr.ErrParamInvalid)
		return
	}
	var body struct {
		Status *int8 `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	if err := h.svc.UpdateMemberStatus(c.Request.Context(), id, *body.Status); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, nil)
}

func (h *MemberHandler) UpdateLevel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, apperr.ErrParamInvalid)
		return
	}
	var body struct {
		LevelID       *uint64      `json:"level_id"`
		LevelExpireAt *requestTime `json:"level_expire_at"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	if err := h.svc.UpdateMemberLevel(c.Request.Context(), id, body.LevelID, requestTimePtr(body.LevelExpireAt)); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, nil)
}
