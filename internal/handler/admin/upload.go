package admin

import (
	"bytes"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"wechat-mall-saas/internal/model"
	"wechat-mall-saas/internal/pkg/aiimage"
	"wechat-mall-saas/internal/pkg/ctxkeys"
	"wechat-mall-saas/internal/pkg/response"
	"wechat-mall-saas/internal/pkg/storage"
	"wechat-mall-saas/internal/repository"
)

// UploadHandler 通用文件上传
type UploadHandler struct {
	repo    *repository.UploadRepo
	storage storage.Storage
	ai      *aiimage.Client
}

func NewUploadHandler(repo *repository.UploadRepo, s storage.Storage, ai *aiimage.Client) *UploadHandler {
	return &UploadHandler{repo: repo, storage: s, ai: ai}
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
	if tenantID == 0 {
		if m := ctxkeys.GetMember(c.Request.Context()); m != nil {
			tenantID = m.TenantID
			uploader = m.ID
		}
	}
	if tenantID == 0 {
		if t := ctxkeys.GetTenant(c.Request.Context()); t != nil {
			tenantID = t.ID
		}
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

type aiImageReq struct {
	Prompt      string `json:"prompt" binding:"required,max=1200"`
	Usage       string `json:"usage"`
	Folder      string `json:"folder"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	AspectRatio string `json:"aspect_ratio"`
}

type aiUsageSize struct {
	width       int
	height      int
	aspectRatio string
}

var aiUsageSizes = map[string]aiUsageSize{
	"common":            {width: 1024, height: 1024, aspectRatio: "1:1"},
	"product-cover":     {width: 1024, height: 1024, aspectRatio: "1:1"},
	"product-gallery":   {width: 1024, height: 1024, aspectRatio: "1:1"},
	"category-cover":    {width: 1024, height: 1024, aspectRatio: "1:1"},
	"storefront-banner": {width: 1500, height: 660, aspectRatio: "750:330"},
	"brand-logo":        {width: 1024, height: 512, aspectRatio: "2:1"},
	"platform-logo":     {width: 1024, height: 512, aspectRatio: "2:1"},
}

// AIImage 使用 AI 服务生成图片并保存到当前对象存储。
func (h *UploadHandler) AIImage(c *gin.Context) {
	if h.ai == nil || !h.ai.Enabled() {
		response.FailCode(c, 30030, "AI 图片生成服务未配置")
		return
	}
	var req aiImageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailCode(c, 20001, err.Error())
		return
	}
	prompt := strings.TrimSpace(req.Prompt)
	if prompt == "" {
		response.FailCode(c, 20001, "请输入图片生成描述")
		return
	}
	usage := sanitizeSegment(defaultString(req.Usage, "common"))
	size := aiUsageSizes[usage]
	if size.width == 0 || size.height == 0 {
		size = aiUsageSizes["common"]
	}
	if req.Width > 0 {
		size.width = req.Width
	}
	if req.Height > 0 {
		size.height = req.Height
	}
	if strings.TrimSpace(req.AspectRatio) != "" {
		size.aspectRatio = strings.TrimSpace(req.AspectRatio)
	}
	result, err := h.ai.Generate(c.Request.Context(), aiimage.GenerateRequest{
		Prompt:      prompt,
		Usage:       usage,
		Width:       size.width,
		Height:      size.height,
		AspectRatio: size.aspectRatio,
	})
	if err != nil {
		response.FailCode(c, 30031, "AI 图片生成失败: "+err.Error())
		return
	}

	ext, mime := imageExtFromContentType(result.ContentType)
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
	fileName := "ai-" + uuid.NewString() + ext
	safeFolder := sanitizeSegment(defaultString(req.Folder, "ai"))
	key := strings.Join([]string{tenantSeg, safeFolder, dateSeg, fileName}, "/")

	url, err := h.storage.Put(c.Request.Context(), key, bytes.NewReader(result.Image), storage.ObjectMeta{
		ContentType: mime,
		Size:        int64(len(result.Image)),
	})
	if err != nil {
		response.Fail(c, err)
		return
	}

	rec := &model.Upload{
		TenantID:     tenantID,
		FileKey:      key,
		OriginalName: fileName,
		FileSize:     int64(len(result.Image)),
		FileType:     mime,
		FileExt:      strings.TrimPrefix(ext, "."),
		StorageType:  h.storage.Type(),
		StorageURL:   url,
		UploadedBy:   uploader,
	}
	_ = h.repo.Create(c.Request.Context(), rec)

	response.OK(c, gin.H{
		"url":            url,
		"file_key":       key,
		"size":           len(result.Image),
		"ext":            strings.TrimPrefix(ext, "."),
		"name":           fileName,
		"model":          result.Model,
		"revised_prompt": result.RevisedPrompt,
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

func imageExtFromContentType(contentType string) (string, string) {
	switch strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0])) {
	case "image/png":
		return ".png", "image/png"
	case "image/webp":
		return ".webp", "image/webp"
	case "image/gif":
		return ".gif", "image/gif"
	default:
		return ".jpg", "image/jpeg"
	}
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
