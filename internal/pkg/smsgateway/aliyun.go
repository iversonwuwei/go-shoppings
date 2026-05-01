package smsgateway

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
)

const aliyunEndpoint = "dysmsapi.aliyuncs.com"

type Client struct {
	httpClient *http.Client
}

type SendRequest struct {
	AccessKeyID     string
	AccessKeySecret string
	Endpoint        string
	PhoneNumbers    string
	SignName        string
	TemplateCode    string
	TemplateParam   string
	OutID           string
}

type SendResponse struct {
	Code      string `json:"Code"`
	Message   string `json:"Message"`
	BizID     string `json:"BizId"`
	RequestID string `json:"RequestId"`
}

func NewClient() *Client {
	return &Client{httpClient: &http.Client{Timeout: 10 * time.Second}}
}

func (c *Client) SendAliyun(ctx context.Context, input SendRequest) (*SendResponse, error) {
	if c == nil {
		c = NewClient()
	}
	endpoint := normalizeEndpoint(input.Endpoint)
	if strings.TrimSpace(input.AccessKeyID) == "" || strings.TrimSpace(input.AccessKeySecret) == "" {
		return nil, fmt.Errorf("aliyun sms access key is required")
	}
	if strings.TrimSpace(input.PhoneNumbers) == "" || strings.TrimSpace(input.SignName) == "" || strings.TrimSpace(input.TemplateCode) == "" {
		return nil, fmt.Errorf("aliyun sms phone, sign name and template code are required")
	}

	query := map[string]string{
		"PhoneNumbers": input.PhoneNumbers,
		"SignName":     input.SignName,
		"TemplateCode": input.TemplateCode,
	}
	if strings.TrimSpace(input.TemplateParam) != "" {
		query["TemplateParam"] = input.TemplateParam
	}
	if strings.TrimSpace(input.OutID) != "" {
		query["OutId"] = input.OutID
	}

	payloadHash := sha256Hex(nil)
	nonce := randomHex(16)
	acsDate := time.Now().UTC().Format("2006-01-02T15:04:05Z")
	headers := map[string]string{
		"host":                 endpoint,
		"x-acs-action":         "SendSms",
		"x-acs-content-sha256": payloadHash,
		"x-acs-date":           acsDate,
		"x-acs-signature-nonce": nonce,
		"x-acs-version":        "2017-05-25",
	}
	signedHeaders := "host;x-acs-action;x-acs-content-sha256;x-acs-date;x-acs-signature-nonce;x-acs-version"
	canonicalQuery := canonicalQuery(query)
	canonicalHeaders := canonicalHeaders(headers, strings.Split(signedHeaders, ";"))
	canonicalRequest := strings.Join([]string{
		http.MethodPost,
		"/",
		canonicalQuery,
		canonicalHeaders,
		signedHeaders,
		payloadHash,
	}, "\n")
	stringToSign := "ACS3-HMAC-SHA256\n" + sha256Hex([]byte(canonicalRequest))
	signature := hmacSHA256Hex([]byte(input.AccessKeySecret), []byte(stringToSign))
	authorization := fmt.Sprintf("ACS3-HMAC-SHA256 Credential=%s,SignedHeaders=%s,Signature=%s", input.AccessKeyID, signedHeaders, signature)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://"+endpoint+"/?"+canonicalQuery, nil)
	if err != nil {
		return nil, err
	}
	req.Host = endpoint
	for name, value := range headers {
		if name == "host" {
			continue
		}
		req.Header.Set(name, value)
	}
	req.Header.Set("Authorization", authorization)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("aliyun sms http %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var out SendResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	if out.Code != "OK" {
		return &out, fmt.Errorf("aliyun sms %s: %s", out.Code, out.Message)
	}
	return &out, nil
}

func normalizeEndpoint(value string) string {
	endpoint := strings.TrimSpace(value)
	if endpoint == "" || strings.EqualFold(endpoint, "cn-hangzhou") {
		return aliyunEndpoint
	}
	endpoint = strings.TrimPrefix(endpoint, "https://")
	endpoint = strings.TrimPrefix(endpoint, "http://")
	if idx := strings.Index(endpoint, "/"); idx >= 0 {
		endpoint = endpoint[:idx]
	}
	if endpoint == "" {
		return aliyunEndpoint
	}
	return endpoint
}

func canonicalQuery(values map[string]string) string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, percentEncode(key)+"="+percentEncode(values[key]))
	}
	return strings.Join(parts, "&")
}

func canonicalHeaders(values map[string]string, names []string) string {
	var builder strings.Builder
	for _, name := range names {
		builder.WriteString(strings.ToLower(name))
		builder.WriteByte(':')
		builder.WriteString(strings.TrimSpace(values[strings.ToLower(name)]))
		builder.WriteByte('\n')
	}
	return builder.String()
}

func percentEncode(value string) string {
	var builder strings.Builder
	for _, b := range []byte(value) {
		if (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z') || (b >= '0' && b <= '9') || b == '-' || b == '_' || b == '.' || b == '~' {
			builder.WriteByte(b)
			continue
		}
		builder.WriteString(fmt.Sprintf("%%%02X", b))
	}
	return builder.String()
}

func sha256Hex(payload []byte) string {
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}

func hmacSHA256Hex(secret, payload []byte) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

func randomHex(size int) string {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(buf)
}