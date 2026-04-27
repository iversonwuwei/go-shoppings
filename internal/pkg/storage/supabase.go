package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultSignedURLExpires = 7 * 24 * 60 * 60

// SupabaseStorage 基于 Supabase Storage REST API 的实现。
type SupabaseStorage struct {
	httpClient       *http.Client
	projectURL       string
	serviceRoleKey   string
	bucket           string
	publicRead       bool
	baseURL          string
	signedURLExpires int
	createBucket     bool
	upsert           bool
}

type SupabaseOptions struct {
	ProjectURL       string
	ServiceRoleKey   string
	Bucket           string
	PublicRead       bool
	BaseURL          string
	SignedURLExpires int
	CreateBucket     bool
	Upsert           bool
}

func NewSupabase(opt SupabaseOptions) (*SupabaseStorage, error) {
	projectURL := strings.TrimRight(strings.TrimSpace(opt.ProjectURL), "/")
	serviceRoleKey := strings.TrimSpace(opt.ServiceRoleKey)
	bucket := strings.Trim(strings.TrimSpace(opt.Bucket), "/")
	if projectURL == "" || serviceRoleKey == "" || bucket == "" {
		return nil, fmt.Errorf("supabase: project_url / service_role_key / bucket 必填")
	}
	if _, err := url.ParseRequestURI(projectURL); err != nil {
		return nil, fmt.Errorf("supabase project_url invalid: %w", err)
	}

	signedURLExpires := opt.SignedURLExpires
	if signedURLExpires <= 0 {
		signedURLExpires = defaultSignedURLExpires
	}

	storage := &SupabaseStorage{
		httpClient:       &http.Client{Timeout: 30 * time.Second},
		projectURL:       projectURL,
		serviceRoleKey:   serviceRoleKey,
		bucket:           bucket,
		publicRead:       opt.PublicRead,
		baseURL:          strings.TrimRight(strings.TrimSpace(opt.BaseURL), "/"),
		signedURLExpires: signedURLExpires,
		createBucket:     opt.CreateBucket,
		upsert:           opt.Upsert,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := storage.ensureBucket(ctx); err != nil {
		return nil, err
	}

	return storage, nil
}

func (s *SupabaseStorage) Type() string { return "supabase" }

func (s *SupabaseStorage) Put(ctx context.Context, key string, reader io.Reader, meta ObjectMeta) (string, error) {
	endpoint := fmt.Sprintf("%s/storage/v1/object/%s/%s", s.projectURL, url.PathEscape(s.bucket), escapeObjectPath(key))
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, reader)
	if err != nil {
		return "", err
	}
	s.authorize(request)
	request.Header.Set("Content-Type", contentTypeOrDefault(meta.ContentType))
	request.Header.Set("x-upsert", fmt.Sprintf("%t", s.upsert))
	if meta.Size >= 0 {
		request.ContentLength = meta.Size
	}

	response, err := s.httpClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("supabase upload: %w", err)
	}
	defer response.Body.Close()
	if !statusAllowed(response.StatusCode, http.StatusOK, http.StatusCreated) {
		return "", fmt.Errorf("supabase upload: status %d: %s", response.StatusCode, readErrorBody(response.Body))
	}

	return s.URL(ctx, key)
}

func (s *SupabaseStorage) URL(ctx context.Context, key string) (string, error) {
	if s.publicRead {
		if s.baseURL != "" {
			return s.baseURL + "/" + escapeObjectPath(key), nil
		}
		return fmt.Sprintf("%s/storage/v1/object/public/%s/%s", s.projectURL, url.PathEscape(s.bucket), escapeObjectPath(key)), nil
	}

	endpoint := fmt.Sprintf("%s/storage/v1/object/sign/%s/%s", s.projectURL, url.PathEscape(s.bucket), escapeObjectPath(key))
	body, err := json.Marshal(map[string]int{"expiresIn": s.signedURLExpires})
	if err != nil {
		return "", err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	s.authorize(request)
	request.Header.Set("Content-Type", "application/json")

	response, err := s.httpClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("supabase sign url: %w", err)
	}
	defer response.Body.Close()
	if !statusAllowed(response.StatusCode, http.StatusOK, http.StatusCreated) {
		return "", fmt.Errorf("supabase sign url: status %d: %s", response.StatusCode, readErrorBody(response.Body))
	}

	var payload struct {
		SignedURL    string `json:"signedURL"`
		SignedURLAlt string `json:"signedUrl"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return "", fmt.Errorf("supabase sign url decode: %w", err)
	}
	signedURL := payload.SignedURL
	if signedURL == "" {
		signedURL = payload.SignedURLAlt
	}
	if signedURL == "" {
		return "", fmt.Errorf("supabase sign url: empty response")
	}
	return s.completeSignedURL(signedURL), nil
}

func (s *SupabaseStorage) ensureBucket(ctx context.Context) error {
	endpoint := fmt.Sprintf("%s/storage/v1/bucket/%s", s.projectURL, url.PathEscape(s.bucket))
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	s.authorize(request)

	response, err := s.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("supabase bucket check: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusOK {
		return nil
	}
	errorBody := readErrorBody(response.Body)
	if !isSupabaseNotFound(response.StatusCode, errorBody) {
		return fmt.Errorf("supabase bucket check: status %d: %s", response.StatusCode, errorBody)
	}
	if !s.createBucket {
		return fmt.Errorf("supabase bucket %q not found", s.bucket)
	}

	return s.createStorageBucket(ctx)
}

func (s *SupabaseStorage) createStorageBucket(ctx context.Context) error {
	body, err := json.Marshal(map[string]any{
		"id":     s.bucket,
		"name":   s.bucket,
		"public": s.publicRead,
	})
	if err != nil {
		return err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, s.projectURL+"/storage/v1/bucket", bytes.NewReader(body))
	if err != nil {
		return err
	}
	s.authorize(request)
	request.Header.Set("Content-Type", "application/json")

	response, err := s.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("supabase bucket create: %w", err)
	}
	defer response.Body.Close()
	if statusAllowed(response.StatusCode, http.StatusOK, http.StatusCreated, http.StatusConflict) {
		return nil
	}
	return fmt.Errorf("supabase bucket create: status %d: %s", response.StatusCode, readErrorBody(response.Body))
}

func (s *SupabaseStorage) authorize(request *http.Request) {
	request.Header.Set("Authorization", "Bearer "+s.serviceRoleKey)
	request.Header.Set("apikey", s.serviceRoleKey)
}

func (s *SupabaseStorage) completeSignedURL(signedURL string) string {
	if strings.HasPrefix(signedURL, "http://") || strings.HasPrefix(signedURL, "https://") {
		return signedURL
	}
	if strings.HasPrefix(signedURL, "/storage/v1/") {
		return s.projectURL + signedURL
	}
	if strings.HasPrefix(signedURL, "/") {
		return s.projectURL + "/storage/v1" + signedURL
	}
	return s.projectURL + "/storage/v1/" + signedURL
}

func escapeObjectPath(key string) string {
	segments := strings.Split(strings.TrimLeft(key, "/"), "/")
	for index, segment := range segments {
		segments[index] = url.PathEscape(segment)
	}
	return strings.Join(segments, "/")
}

func contentTypeOrDefault(contentType string) string {
	contentType = strings.TrimSpace(contentType)
	if contentType == "" {
		return "application/octet-stream"
	}
	return contentType
}

func statusAllowed(statusCode int, allowed ...int) bool {
	for _, candidate := range allowed {
		if statusCode == candidate {
			return true
		}
	}
	return false
}

func isSupabaseNotFound(statusCode int, body string) bool {
	if statusCode == http.StatusNotFound {
		return true
	}
	return statusCode == http.StatusBadRequest && strings.Contains(body, `"statusCode":"404"`)
}

func readErrorBody(reader io.Reader) string {
	body, err := io.ReadAll(io.LimitReader(reader, 4096))
	if err != nil {
		return err.Error()
	}
	message := strings.TrimSpace(string(body))
	if message == "" {
		return http.StatusText(http.StatusInternalServerError)
	}
	return message
}
