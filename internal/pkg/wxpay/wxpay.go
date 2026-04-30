// Package wxpay 微信支付 v3 客户端（JSAPI 下单 / 回调解密骨架）
//
// 生产环境请补充：证书加载、签名（RSA-SHA256）、AES-256-GCM 回调解密等。
// 此处给出可替换的接口占位与数据结构，确保业务层可独立开发与测试。
package wxpay

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/json"
	"errors"
)

type Config struct {
	AppID      string
	MchID      string
	APIv3Key   string
	CertSerial string
	NotifyURL  string
}

type Client struct {
	cfg Config
}

func NewClient(cfg Config) *Client { return &Client{cfg: cfg} }

// Configured 返回是否已具备最小可用配置。
func (c *Client) Configured() bool {
	if c == nil {
		return false
	}
	return c.cfg.AppID != "" && c.cfg.MchID != ""
}

type JSAPIOrderReq struct {
	Description string
	OutTradeNo  string
	TotalFen    int64
	OpenID      string
}

type JSAPIPayParams struct {
	AppID     string `json:"appId"`
	TimeStamp string `json:"timeStamp"`
	NonceStr  string `json:"nonceStr"`
	Package   string `json:"package"` // prepay_id=xxx
	SignType  string `json:"signType"`
	PaySign   string `json:"paySign"`
}

// PlaceJSAPIOrder 下单（占位实现）——调用方应在生产中实现真实签名与 HTTP 请求。
func (c *Client) PlaceJSAPIOrder(ctx context.Context, req JSAPIOrderReq) (*JSAPIPayParams, error) {
	if c.cfg.AppID == "" || c.cfg.MchID == "" {
		return nil, errors.New("wxpay: appid/mchid not configured")
	}
	// TODO: 实现 https://api.mch.weixin.qq.com/v3/pay/transactions/jsapi
	return &JSAPIPayParams{
		AppID:     c.cfg.AppID,
		TimeStamp: "0",
		NonceStr:  "placeholder",
		Package:   "prepay_id=PLACEHOLDER",
		SignType:  "RSA",
		PaySign:   "PLACEHOLDER",
	}, nil
}

// CallbackResource 回调报文 resource 字段结构
type CallbackResource struct {
	OriginalType   string `json:"original_type"`
	Algorithm      string `json:"algorithm"`
	Ciphertext     string `json:"ciphertext"`
	AssociatedData string `json:"associated_data"`
	Nonce          string `json:"nonce"`
}

// TransactionInfo 解密后的交易信息
type TransactionInfo struct {
	OutTradeNo    string `json:"out_trade_no"`
	TransactionID string `json:"transaction_id"`
	TradeState    string `json:"trade_state"`
	SuccessTime   string `json:"success_time"`
	Payer         struct {
		OpenID string `json:"openid"`
	} `json:"payer"`
	Amount struct {
		Total    int64  `json:"total"`
		Currency string `json:"currency"`
	} `json:"amount"`
}

// DecryptCallback AES-256-GCM 解密
func (c *Client) DecryptCallback(res CallbackResource) (*TransactionInfo, error) {
	if c.cfg.APIv3Key == "" {
		return nil, errors.New("wxpay: apiv3 key not set")
	}
	ciphertext, err := base64Decode(res.Ciphertext)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher([]byte(c.cfg.APIv3Key))
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	plain, err := gcm.Open(nil, []byte(res.Nonce), ciphertext, []byte(res.AssociatedData))
	if err != nil {
		return nil, err
	}
	var t TransactionInfo
	if err := json.Unmarshal(plain, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

func base64Decode(s string) ([]byte, error) {
	return base64Std.DecodeString(s)
}
