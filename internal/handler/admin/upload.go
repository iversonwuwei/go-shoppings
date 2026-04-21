package admin

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"wechat-mall-saas/internal/model"
	"wechat-mall-saas/internal/pkg/config"
	"wechat-mall-saas/internal/pkg/ctxkeys"
	"wechat-mall-saas/internal/pkg/response"
	"wechat-mall-saas/internal/repository"
)

// UploadHandler 通用文件上传（当前仅本地存储 + 图片）
type UploadHandler struct {
	repo    *repository.UploadRepo
	storage config.StorageConfig
}

func NewUploadHandler(repo *repository.UploadRepo, storage config.StorageConfig) *UploadHandler {
	return &UploadHandler{repo: repo, storage: storage}
}

// 允许的图片 MIME / 扩展名
var allowedImageExt = map[string]string{
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".gif":  "image/gif",
	".webp": "image/webp",
	".bmp":  "image/bmp",
	".svg":  "image/svg+xml",
}

// Image 上传单张图片
//
//	表单字段: file
//	可选 query: folder=products|logo|avatar|banner （仅用于路径分类）
//	返回: { url, file_key, size, ext }
func (h *UploadHandler) Image(c *gin.Context) {
	const maxSize = 10 << 20 // 10MB
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)

	fh, err := c.FormFile("file")
	if err != nil {
		response.FailCode(c, 20001, "未选择文件: "+err.Error())
		return
	}
	if fh.Size > maxSize {
		response.FailCode(c, 20001, "文件超过 10MB 限制")
		return
	}
	ext := strings.ToLower(filepath.Ext(fh.Filename))
	mime, ok := allowedImageExt[ext]
	if !ok {
		response.FailCode(c, 20001, "仅支持图片格式: jpg/png/gif/webp/bmp/svg")
		return
	}

	// 租户隔离路径；平台(tenant_id=0)统一放 platform/
	var tenantID uint64
	if a := ctxkeys.GetAdmin(c.Request.Context()); a != nil {
		tenantID = a.TenantID
	}
	folder := c.DefaultQuery("folder", "common")
	safeFolder := sanitizeSegment(folder)

	tenantSeg := "platform"
	if tenantID > 0 {
		tenantSeg = fmt.Sprintf("tenant-%d", tenantID)
	}
	dateSeg := time.Now().Format("200601")
	fileName := uuid.NewString() + ext
	relDir := filepath.ToSlash(filepath.Join(tenantSeg, safeFolder, dateSeg))
	relPath := filepath.ToSlash(filepath.Join(relDir, fileName))
	absDir := filepath.Join(h.storage.Local.Path, relDir)
	absPath := filepath.Join(h.storage.Local.Path, relPath)

	if err := os.MkdirAll(absDir, 0o755); err != nil {
		response.Fail(c, err)
		return
	}
	src, err := fh.Open()
	if err != nil {
		response.Fail(c, err)
		return
	}
	defer src.Close()
	dst, err := os.Create(absPath)
	if err != nil {
		response.Fail(c, err)
		return
	}
	written, err := io.Copy(dst, src)
	_ = dst.Close()
	if err != nil {
		_ = os.Remove(absPath)
		response.Fail(c, err)
		return
	}

	url := strings.TrimRight(h.storage.Local.BaseURL, "/") + "/" + relPath
	var uploader uint64
	if a := ctxkeys.GetAdmin(c.Request.Context()); a != nil {
		uploader = a.ID
	}
	rec := &model.Upload{
		TenantID:     tenantID,
		FileKey:      relPath,
		OriginalName: fh.Filename,
		FileSize:     written,
		FileType:     mime,
		FileExt:      strings.TrimPrefix(ext, "."),
		StorageType:  "local",
		StorageURL:   url,
		UploadedBy:   uploader,
	}
	_ = h.repo.Create(c.Request.Context(), rec) // 记录失败不影响上传成功

	response.OK(c, gin.H{
		"url":      url,
		"file_key": relPath,
		"size":     written,
		"ext":      strings.TrimPrefix(ext, "."),
		"name":     fh.Filename,
	})
}

// 防止 folder 参数中出现路径遍历
func sanitizeSegment(s string) string {
	s = strings.ReplaceAll(s, "..", "")
	s = strings.Trim(s, "/\\")
	if s == "" {
		return "common"
	}
	// 只保留字母数字-_
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			b.WriteRune(r)
		}
	}
	out := b.String()
	if out == "" {
		return "common"
	}
	return out
}
