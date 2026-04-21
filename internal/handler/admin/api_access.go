package admin

import (
	"crypto/rand"
	"encoding/hex"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"wechat-mall-saas/internal/model"
	apperr "wechat-mall-saas/internal/pkg/errors"
	"wechat-mall-saas/internal/pkg/response"
	"wechat-mall-saas/internal/repository"
)

type ApiAccessHandler struct {
	repo *repository.ApiTokenRepo
}

func NewApiAccessHandler(r *repository.ApiTokenRepo) *ApiAccessHandler {
	return &ApiAccessHandler{repo: r}
}

type apiTokenReq struct {
	Name        string `json:"name" binding:"required,max=100"`
	Scopes      string `json:"scopes"`
	IPWhitelist string `json:"ip_whitelist"`
	Status      int8   `json:"status"`
	ExpiresAt   string `json:"expires_at"` // RFC3339，可空
}

func randHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func parseExpires(s string) *time.Time {
	if s == "" {
		return nil
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return &t
	}
	if t, err := time.Parse("2006-01-02 15:04:05", s); err == nil {
		return &t
	}
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return &t
	}
	return nil
}

func maskSecret(t *model.ApiToken) {
	if t == nil || t.AppSecret == "" {
		return
	}
	s := t.AppSecret
	if len(s) > 8 {
		t.AppSecret = s[:4] + "****" + s[len(s)-4:]
	} else {
		t.AppSecret = "****"
	}
}

func (h *ApiAccessHandler) List(c *gin.Context) {
	rows, err := h.repo.List(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	for i := range rows {
		maskSecret(&rows[i])
	}
	response.OK(c, rows)
}

func (h *ApiAccessHandler) Create(c *gin.Context) {
	var req apiTokenReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	t := &model.ApiToken{
		Name:        req.Name,
		AppKey:      "ak_" + randHex(12),
		AppSecret:   randHex(32),
		Scopes:      req.Scopes,
		IPWhitelist: req.IPWhitelist,
		Status:      defaultCouponStatus(req.Status),
		ExpiresAt:   parseExpires(req.ExpiresAt),
	}
	if err := h.repo.Create(c.Request.Context(), t); err != nil {
		response.Fail(c, err)
		return
	}
	// 创建时返回完整 secret（仅此一次）
	response.OK(c, t)
}

func (h *ApiAccessHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.Fail(c, apperr.ErrParamInvalid)
		return
	}
	var req apiTokenReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	fields := map[string]interface{}{
		"name":         req.Name,
		"scopes":       req.Scopes,
		"ip_whitelist": req.IPWhitelist,
		"status":       defaultCouponStatus(req.Status),
	}
	if exp := parseExpires(req.ExpiresAt); exp != nil {
		fields["expires_at"] = exp
	} else if req.ExpiresAt == "" {
		fields["expires_at"] = nil
	}
	if err := h.repo.Update(c.Request.Context(), id, fields); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"id": id})
}

func (h *ApiAccessHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.Fail(c, apperr.ErrParamInvalid)
		return
	}
	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, nil)
}

// Regenerate 重新生成 AppSecret
func (h *ApiAccessHandler) Regenerate(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.Fail(c, apperr.ErrParamInvalid)
		return
	}
	secret := randHex(32)
	if err := h.repo.Update(c.Request.Context(), id, map[string]interface{}{
		"app_secret": secret,
	}); err != nil {
		response.Fail(c, err)
		return
	}
	t, err := h.repo.Find(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	// 重置后返回完整 secret（仅此一次）
	response.OK(c, t)
}

func (h *ApiAccessHandler) ListLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	tokenID, _ := strconv.ParseUint(c.Query("token_id"), 10, 64)
	rows, total, err := h.repo.ListLogs(c.Request.Context(), tokenID, page, size)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{"list": rows, "total": total, "page": page, "size": size})
}
