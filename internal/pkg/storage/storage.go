// Package storage 对象存储抽象（支持 local / minio / supabase），未来可扩展 oss / cos。
package storage

import (
	"context"
	"io"
)

// ObjectMeta 上传对象元数据
type ObjectMeta struct {
	// ContentType 建议透传浏览器上传的 MIME，以便下载时正确呈现
	ContentType string
	// Size 字节大小（未知时填 -1，MinIO 也可接受）
	Size int64
}

// Storage 通用对象存储接口
type Storage interface {
	// Type 返回实现类型：local / minio / supabase
	Type() string
	// Put 上传对象，返回可访问 URL。
	// key 为对象路径（不含前导斜杠），如 "tenant-1/products/202604/uuid.jpg"
	Put(ctx context.Context, key string, r io.Reader, meta ObjectMeta) (string, error)
	// URL 根据 key 获取可访问 URL（MinIO 非公开桶时返回预签名 URL）
	URL(ctx context.Context, key string) (string, error)
}
