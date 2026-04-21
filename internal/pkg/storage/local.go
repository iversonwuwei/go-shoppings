package storage

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// LocalStorage 本地文件系统存储
type LocalStorage struct {
	basePath string
	baseURL  string
}

func NewLocal(basePath, baseURL string) *LocalStorage {
	return &LocalStorage{basePath: basePath, baseURL: strings.TrimRight(baseURL, "/")}
}

func (s *LocalStorage) Type() string { return "local" }

func (s *LocalStorage) Put(_ context.Context, key string, r io.Reader, _ ObjectMeta) (string, error) {
	abs := filepath.Join(s.basePath, filepath.FromSlash(key))
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return "", err
	}
	f, err := os.Create(abs)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := io.Copy(f, r); err != nil {
		_ = os.Remove(abs)
		return "", err
	}
	return s.baseURL + "/" + key, nil
}

func (s *LocalStorage) URL(_ context.Context, key string) (string, error) {
	return s.baseURL + "/" + key, nil
}

// BasePath 返回本地存储根路径，供 gin.Static 使用
func (s *LocalStorage) BasePath() string { return s.basePath }
