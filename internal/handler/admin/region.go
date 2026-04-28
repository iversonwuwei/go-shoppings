package admin

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"wechat-mall-saas/internal/model"
	apperr "wechat-mall-saas/internal/pkg/errors"
	"wechat-mall-saas/internal/pkg/response"
	"wechat-mall-saas/internal/repository"
)

type RegionHandler struct {
	repo *repository.RegionRepo
}

func NewRegionHandler(r *repository.RegionRepo) *RegionHandler {
	return &RegionHandler{repo: r}
}

type regionReq struct {
	ParentID uint64 `json:"parent_id"`
	Code     string `json:"code"`
	Name     string `json:"name" binding:"required,max=50"`
	Sort     int    `json:"sort"`
	Enabled  *int8  `json:"enabled"`
}

func (h *RegionHandler) PublicTree(c *gin.Context) {
	rows, err := h.repo.Tree(c.Request.Context(), false)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, rows)
}

func (h *RegionHandler) List(c *gin.Context) {
	rows, err := h.repo.List(c.Request.Context(), true)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, rows)
}

func (h *RegionHandler) Tree(c *gin.Context) {
	rows, err := h.repo.Tree(c.Request.Context(), true)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, rows)
}

func (h *RegionHandler) Create(c *gin.Context) {
	var req regionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	level, err := h.levelForParent(c, req.ParentID, 0)
	if err != nil {
		response.Fail(c, err)
		return
	}
	enabled := int8(1)
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	row := &model.Region{
		ParentID: req.ParentID,
		Code:     strings.TrimSpace(req.Code),
		Name:     strings.TrimSpace(req.Name),
		Level:    level,
		Sort:     req.Sort,
		Enabled:  enabled,
	}
	if row.Name == "" {
		response.Fail(c, apperr.ErrParamInvalid)
		return
	}
	if err := h.repo.Create(c.Request.Context(), row); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, row)
}

func (h *RegionHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.Fail(c, apperr.ErrParamInvalid)
		return
	}
	cur, err := h.repo.FindByID(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	if cur == nil {
		response.Fail(c, apperr.ErrNotFound)
		return
	}
	var req regionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	level, err := h.levelForParent(c, req.ParentID, id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	cur.ParentID = req.ParentID
	cur.Code = strings.TrimSpace(req.Code)
	cur.Name = strings.TrimSpace(req.Name)
	cur.Level = level
	cur.Sort = req.Sort
	if req.Enabled != nil {
		cur.Enabled = *req.Enabled
	}
	if cur.Name == "" {
		response.Fail(c, apperr.ErrParamInvalid)
		return
	}
	if err := h.repo.Update(c.Request.Context(), cur); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, cur)
}

func (h *RegionHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.Fail(c, apperr.ErrParamInvalid)
		return
	}
	count, err := h.repo.CountChildren(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	if count > 0 {
		response.Fail(c, apperr.New(20001, "请先删除下级城市信息"))
		return
	}
	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, nil)
}

func (h *RegionHandler) levelForParent(c *gin.Context, parentID, selfID uint64) (int8, error) {
	if parentID == 0 {
		return 1, nil
	}
	if parentID == selfID {
		return 0, apperr.New(20001, "父级不能选择自身")
	}
	parent, err := h.repo.FindByID(c.Request.Context(), parentID)
	if err != nil {
		return 0, err
	}
	if parent == nil {
		return 0, apperr.New(20001, "父级城市不存在")
	}
	if parent.Level >= 3 {
		return 0, apperr.New(20001, "区县下不能继续添加下级")
	}
	ancestor := parent
	for ancestor != nil && ancestor.ParentID != 0 {
		if ancestor.ParentID == selfID {
			return 0, apperr.New(20001, "父级不能选择自己的下级")
		}
		next, err := h.repo.FindByID(c.Request.Context(), ancestor.ParentID)
		if err != nil {
			return 0, err
		}
		ancestor = next
	}
	return parent.Level + 1, nil
}
