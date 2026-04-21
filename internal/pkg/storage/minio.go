package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinioStorage 基于 MinIO / 任意 S3 兼容对象存储的实现
type MinioStorage struct {
	cli        *minio.Client
	bucket     string
	baseURL    string // 对外访问前缀；空则用 endpoint
	publicRead bool
	endpoint   string
	useSSL     bool
}

// MinioOptions 初始化参数
type MinioOptions struct {
	Endpoint   string
	AccessKey  string
	SecretKey  string
	Bucket     string
	UseSSL     bool
	Region     string
	BaseURL    string
	PublicRead bool
}

// NewMinio 创建 MinIO 存储；会自动确保 bucket 存在
func NewMinio(opt MinioOptions) (*MinioStorage, error) {
	if opt.Endpoint == "" || opt.AccessKey == "" || opt.SecretKey == "" || opt.Bucket == "" {
		return nil, fmt.Errorf("minio: endpoint / access_key / secret_key / bucket 必填")
	}
	cli, err := minio.New(opt.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(opt.AccessKey, opt.SecretKey, ""),
		Secure: opt.UseSSL,
		Region: opt.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("minio new client: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	exists, err := cli.BucketExists(ctx, opt.Bucket)
	if err != nil {
		return nil, fmt.Errorf("minio bucket exists: %w", err)
	}
	if !exists {
		if err := cli.MakeBucket(ctx, opt.Bucket, minio.MakeBucketOptions{Region: opt.Region}); err != nil {
			return nil, fmt.Errorf("minio make bucket: %w", err)
		}
		if opt.PublicRead {
			policy := fmt.Sprintf(`{
				"Version":"2012-10-17",
				"Statement":[{
					"Effect":"Allow",
					"Principal":{"AWS":["*"]},
					"Action":["s3:GetObject"],
					"Resource":["arn:aws:s3:::%s/*"]
				}]
			}`, opt.Bucket)
			_ = cli.SetBucketPolicy(ctx, opt.Bucket, policy)
		}
	}
	return &MinioStorage{
		cli:        cli,
		bucket:     opt.Bucket,
		baseURL:    strings.TrimRight(opt.BaseURL, "/"),
		publicRead: opt.PublicRead,
		endpoint:   opt.Endpoint,
		useSSL:     opt.UseSSL,
	}, nil
}

func (s *MinioStorage) Type() string { return "minio" }

func (s *MinioStorage) Put(ctx context.Context, key string, r io.Reader, meta ObjectMeta) (string, error) {
	_, err := s.cli.PutObject(ctx, s.bucket, key, r, meta.Size, minio.PutObjectOptions{
		ContentType: meta.ContentType,
	})
	if err != nil {
		return "", fmt.Errorf("minio put: %w", err)
	}
	return s.URL(ctx, key)
}

func (s *MinioStorage) URL(ctx context.Context, key string) (string, error) {
	if s.publicRead {
		if s.baseURL != "" {
			return s.baseURL + "/" + key, nil
		}
		scheme := "http"
		if s.useSSL {
			scheme = "https"
		}
		return fmt.Sprintf("%s://%s/%s/%s", scheme, s.endpoint, s.bucket, key), nil
	}
	// 预签名 7 天
	u, err := s.cli.PresignedGetObject(ctx, s.bucket, key, 7*24*time.Hour, url.Values{})
	if err != nil {
		return "", fmt.Errorf("minio presign: %w", err)
	}
	return u.String(), nil
}
