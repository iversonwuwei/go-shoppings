package aiimage

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"wechat-mall-saas/internal/pkg/config"
)

const defaultMaxBytes int64 = 10 << 20

type Client struct {
	enabled bool
	baseURL string
	http    *http.Client
	maxSize int64
}

type GenerateRequest struct {
	Prompt      string `json:"prompt"`
	Usage       string `json:"usage,omitempty"`
	Width       int    `json:"width,omitempty"`
	Height      int    `json:"height,omitempty"`
	AspectRatio string `json:"aspect_ratio,omitempty"`
}

type GenerateResult struct {
	Image         []byte
	ContentType   string
	Model         string
	RevisedPrompt string
}

type serviceResponse struct {
	ImageBase64   string `json:"image_base64"`
	ImageURL      string `json:"image_url"`
	ContentType   string `json:"content_type"`
	Model         string `json:"model"`
	RevisedPrompt string `json:"revised_prompt"`
}

func New(cfg config.AIImageConfig) *Client {
	maxSize := int64(cfg.MaxBytes)
	if maxSize <= 0 {
		maxSize = defaultMaxBytes
	}
	timeout := time.Duration(cfg.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 120 * time.Second
	}
	return &Client{
		enabled: cfg.Enabled,
		baseURL: strings.TrimRight(strings.TrimSpace(cfg.ServiceURL), "/"),
		http:    &http.Client{Timeout: timeout},
		maxSize: maxSize,
	}
}

func (c *Client) Enabled() bool {
	return c != nil && c.enabled && c.baseURL != ""
}

func (c *Client) Generate(ctx context.Context, req GenerateRequest) (*GenerateResult, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("AI image service is not configured")
	}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/images/generate", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("AI image service returned %d: %s", resp.StatusCode, strings.TrimSpace(string(msg)))
	}

	var payload serviceResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, c.maxSize+4096)).Decode(&payload); err != nil {
		return nil, err
	}
	image, contentType, err := c.decodeImage(ctx, payload)
	if err != nil {
		return nil, err
	}
	return &GenerateResult{
		Image:         image,
		ContentType:   contentType,
		Model:         payload.Model,
		RevisedPrompt: payload.RevisedPrompt,
	}, nil
}

func (c *Client) decodeImage(ctx context.Context, payload serviceResponse) ([]byte, string, error) {
	if payload.ImageBase64 != "" {
		imageBase64, contentType := splitDataURI(payload.ImageBase64)
		image, err := base64.StdEncoding.DecodeString(imageBase64)
		if err != nil {
			return nil, "", err
		}
		if int64(len(image)) > c.maxSize {
			return nil, "", fmt.Errorf("generated image exceeds %d bytes", c.maxSize)
		}
		if contentType == "" {
			contentType = payload.ContentType
		}
		return image, normalizeContentType(contentType, image), nil
	}
	if payload.ImageURL == "" {
		return nil, "", fmt.Errorf("AI image service did not return image_base64 or image_url")
	}
	return c.downloadImage(ctx, payload.ImageURL, payload.ContentType)
}

func (c *Client) downloadImage(ctx context.Context, imageURL, fallbackContentType string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return nil, "", err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", fmt.Errorf("download generated image returned %d", resp.StatusCode)
	}
	image, err := io.ReadAll(io.LimitReader(resp.Body, c.maxSize+1))
	if err != nil {
		return nil, "", err
	}
	if int64(len(image)) > c.maxSize {
		return nil, "", fmt.Errorf("generated image exceeds %d bytes", c.maxSize)
	}
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = fallbackContentType
	}
	return image, normalizeContentType(contentType, image), nil
}

func splitDataURI(value string) (string, string) {
	if !strings.HasPrefix(value, "data:") {
		return value, ""
	}
	parts := strings.SplitN(value, ",", 2)
	if len(parts) != 2 {
		return value, ""
	}
	meta := strings.TrimPrefix(parts[0], "data:")
	contentType := strings.TrimSuffix(meta, ";base64")
	return parts[1], contentType
}

func normalizeContentType(contentType string, image []byte) string {
	contentType = strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
	switch contentType {
	case "image/jpeg", "image/png", "image/webp", "image/gif":
		return contentType
	}
	detected := http.DetectContentType(image)
	switch detected {
	case "image/jpeg", "image/png", "image/webp", "image/gif":
		return detected
	default:
		return "image/jpeg"
	}
}
