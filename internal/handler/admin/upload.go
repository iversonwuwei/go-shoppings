package admin

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"wechat-mall-saas/internal/model"
	"wechat-mall-saas/internal/pkg/ctxkeys"
	"wechat-mall-saas/internal/pkg/response"
	"wechat-mall-saas/internal/pkg/storage"
	"wechat-mall-saas/internal/repository"
)

// UploadHandler 通用文件上传
type UploadHandler struct {
	repo    *repository.UploadRepo
	storage storage.Storage
}

func NewUploadHandler(repo *repository.UploadRepo, s storage.Storage) *UploadHandler {
	return &UploadHandler{repo: repo, storage: s}
}

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
//	可选 query: folder=products|logo|avatar|banner
//	返回: { url, file_key, size, ext, name }
func (h *UploadHandler) Image(c *gin.Context) {
	const maxSize int64 = 10 << 20 // 10MB
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

	var tenantID, uploader uint64
	if a := ctxkeys.GetAdmin(c.Request.Context()); a != nil {
		tenantID = a.TenantID
		uploader = a.ID
	}
	tenantSeg := "platform"
	if tenantID > 0 {
		tenantSeg = fmt.Sprintf("tenant-%d", tenantID)
	}
	dateSeg := time.Now().Format("200601")
	fileName := uuid.NewString() + ext
	safeFolder := sanitizeSegment(c.DefaultQuery("folder", "common"))
	key := strings.Join([]string{tenantSeg, safeFolder, dateSeg, fileName}, "/")

	src, err := fh.Open()
	if err != nil {
		response.Fail(c, err)
		return
	}
	defer src.Close()

	url, err := h.storage.Put(c.Request.Context(), key, src, storage.ObjectMeta{
		ContentType: mime,
		Size:        fh.Size,
	})
	if err != nil {
		response.Fail(c, err)
		return
	}

	rec := &model.Upload{
		TenantID:     tenantID,
		FileKey:      key,
		OriginalName: fh.Filename,
		FileSize:     fh.Size,
		FileType:     mime,
		FileExt:      strings.TrimPrefix(ext, "."),
		StorageType:  h.storage.Type(),
		StorageURL:   url,
		UploadedBy:   uploader,
	}
	_ = h.repo.Create(c.Request.Context(), rec)

	response.OK(c, gin.H{
		"url":      url,
		"file_key": key,
		"size":     fh.Size,
		"ext":      strings.TrimPrefix(ext, "."),
		"name":     fh.Filename,
	})
}

// sanitizeSegment 防止 folder 参数中出现路径遍历
func sanitizeSegment(s string) string {
	s = strings.ReplaceAll(s, "..", "")
	s = strings.Trim(s, "/\\")
	if s == "" {
		return "common"
	}
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
