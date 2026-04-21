// Package wxapp 微信小程序 SDK：登录（code2session）与手机号解密
package wxapp

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	AppID     string
	AppSecret string
	http      *http.Client
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
