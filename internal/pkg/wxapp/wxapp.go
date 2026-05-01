// Package wxapp 微信小程序 SDK：登录（code2session）与手机号解密
package wxapp

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type Client struct {
	AppID     string
	AppSecret string
	http      *http.Client

	mu                   sync.Mutex
	accessToken          string
	accessTokenExpiresAt time.Time
}

func NewClient(appID, appSecret string) *Client {
	return &Client{AppID: appID, AppSecret: appSecret, http: &http.Client{Timeout: 10 * time.Second}}
}

type Code2SessionResp struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid,omitempty"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

type accessTokenResp struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
}

type UnlimitedQRCodeRequest struct {
	Scene      string
	Page       string
	CheckPath  bool
	EnvVersion string
	Width      int
}

type wechatAPIErrorResp struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func (c *Client) Code2Session(jsCode string) (*Code2SessionResp, error) {
	u := fmt.Sprintf("https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		url.QueryEscape(c.AppID), url.QueryEscape(c.AppSecret), url.QueryEscape(jsCode))
	resp, err := c.http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var r Code2SessionResp
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, err
	}
	if r.ErrCode != 0 {
		return nil, fmt.Errorf("wx: %d %s", r.ErrCode, r.ErrMsg)
	}
	return &r, nil
}

func (client *Client) AccessToken(ctx context.Context) (string, error) {
	if client == nil || strings.TrimSpace(client.AppID) == "" || strings.TrimSpace(client.AppSecret) == "" {
		return "", errors.New("平台微信小程序 AppID/AppSecret 未配置，无法生成小程序码")
	}
	now := time.Now()
	client.mu.Lock()
	if client.accessToken != "" && now.Before(client.accessTokenExpiresAt.Add(-time.Minute)) {
		token := client.accessToken
		client.mu.Unlock()
		return token, nil
	}
	client.mu.Unlock()

	apiURL := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s",
		url.QueryEscape(client.AppID), url.QueryEscape(client.AppSecret))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return "", err
	}
	resp, err := client.httpClient().Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("wx access token http %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var out accessTokenResp
	if err := json.Unmarshal(body, &out); err != nil {
		return "", err
	}
	if out.ErrCode != 0 || strings.TrimSpace(out.AccessToken) == "" {
		return "", fmt.Errorf("wx access token: %d %s", out.ErrCode, out.ErrMsg)
	}
	expiresIn := out.ExpiresIn
	if expiresIn <= 0 {
		expiresIn = 7200
	}
	client.mu.Lock()
	client.accessToken = out.AccessToken
	client.accessTokenExpiresAt = now.Add(time.Duration(expiresIn) * time.Second)
	client.mu.Unlock()
	return out.AccessToken, nil
}

func (client *Client) GetUnlimitedQRCode(ctx context.Context, input UnlimitedQRCodeRequest) ([]byte, error) {
	scene := strings.TrimSpace(input.Scene)
	if scene == "" {
		return nil, errors.New("小程序码 scene 不能为空")
	}
	page := strings.TrimSpace(input.Page)
	if page == "" {
		page = "pages/home/index"
	}
	envVersion := strings.TrimSpace(input.EnvVersion)
	if envVersion == "" {
		envVersion = "release"
	}
	width := input.Width
	if width <= 0 {
		width = 360
	}
	token, err := client.AccessToken(ctx)
	if err != nil {
		return nil, err
	}
	payload := map[string]interface{}{
		"scene":       scene,
		"page":        page,
		"check_path":  input.CheckPath,
		"env_version": envVersion,
		"width":       width,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	apiURL := "https://api.weixin.qq.com/wxa/getwxacodeunlimit?access_token=" + url.QueryEscape(token)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	imageOrError, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("wx getwxacodeunlimit http %d: %s", resp.StatusCode, strings.TrimSpace(string(imageOrError)))
	}
	if json.Valid(imageOrError) {
		var apiErr wechatAPIErrorResp
		if err := json.Unmarshal(imageOrError, &apiErr); err == nil && apiErr.ErrCode != 0 {
			return nil, fmt.Errorf("wx getwxacodeunlimit: %d %s (page=%s env_version=%s check_path=%t)", apiErr.ErrCode, apiErr.ErrMsg, page, envVersion, input.CheckPath)
		}
		return nil, fmt.Errorf("wx getwxacodeunlimit returned json: %s", strings.TrimSpace(string(imageOrError)))
	}
	if len(imageOrError) == 0 {
		return nil, errors.New("wx getwxacodeunlimit returned empty image")
	}
	return imageOrError, nil
}

func (client *Client) httpClient() *http.Client {
	if client != nil && client.http != nil {
		return client.http
	}
	return http.DefaultClient
}

type PhoneInfo struct {
	PhoneNumber     string `json:"phoneNumber"`
	PurePhoneNumber string `json:"purePhoneNumber"`
	CountryCode     string `json:"countryCode"`
}

// DecryptPhone AES-128-CBC 解密加密数据
func DecryptPhone(sessionKey, encryptedData, iv string) (*PhoneInfo, error) {
	key, err := base64.StdEncoding.DecodeString(sessionKey)
	if err != nil {
		return nil, err
	}
	ivBytes, err := base64.StdEncoding.DecodeString(iv)
	if err != nil {
		return nil, err
	}
	data, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(data)%block.BlockSize() != 0 {
		return nil, errors.New("invalid encrypted data length")
	}
	mode := cipher.NewCBCDecrypter(block, ivBytes)
	plain := make([]byte, len(data))
	mode.CryptBlocks(plain, data)
	plain = pkcs7Unpad(plain)
	plain = bytes.TrimRight(plain, "\x00")
	var info PhoneInfo
	if err := json.Unmarshal(plain, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

func pkcs7Unpad(b []byte) []byte {
	if len(b) == 0 {
		return b
	}
	n := int(b[len(b)-1])
	if n > len(b) || n > aes.BlockSize {
		return b
	}
	return b[:len(b)-n]
}
