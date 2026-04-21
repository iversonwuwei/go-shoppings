package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"wechat-mall-saas/internal/model"
	apperr "wechat-mall-saas/internal/pkg/errors"
	"wechat-mall-saas/internal/repository"
)

// SettingsService 管理商户收款（商户提交+平台审核）与物流承运商（平台统一维护）
type SettingsService struct {
	payCfg  *repository.PaymentConfigRepo
	carrier *repository.ShippingCarrierRepo
	tenant  *TenantService
}

func NewSettingsService(p *repository.PaymentConfigRepo, c *repository.ShippingCarrierRepo, t *TenantService) *SettingsService {
	return &SettingsService{payCfg: p, carrier: c, tenant: t}
}

// ========== 商户侧：收款配置 ==========

func (s *SettingsService) ListPaymentConfigs(ctx context.Context) ([]model.TenantPaymentConfig, error) {
	tid := repository.EnsureTenant(ctx)
	if tid == 0 {
		return nil, apperr.New(40001, "tenant required")
	}
	return s.payCfg.ListByTenant(ctx, tid)
}

type PaymentConfigInput struct {
	Provider      string `json:"provider"`
	MchID         string `json:"mch_id"`
	AppID         string `json:"app_id"`
	APIV3Key      string `json:"api_v3_key"`
	CertSerialNo  string `json:"cert_serial_no"`
	PrivateKeyPEM string `json:"private_key_pem"`
	CertPEM       string `json:"cert_pem"`
	NotifyURL     string `json:"notify_url"`
}

func (s *SettingsService) SubmitPaymentConfig(ctx context.Context, in PaymentConfigInput) (*model.TenantPaymentConfig, error) {
	tid := repository.EnsureTenant(ctx)
	if tid == 0 {
		return nil, apperr.New(40001, "tenant required")
	}
	if in.Provider == "" {
		in.Provider = "wechat"
	}
	if in.MchID == "" || in.APIV3Key == "" {
		return nil, apperr.New(20001, "mch_id 与 api_v3_key 必填")
	}
	m := &model.TenantPaymentConfig{
		TenantID:      tid,
		Provider:      in.Provider,
		MchID:         in.MchID,
		AppID:         in.AppID,
		APIV3Key:      in.APIV3Key,
		CertSerialNo:  in.CertSerialNo,
		PrivateKeyPEM: in.PrivateKeyPEM,
		CertPEM:       in.CertPEM,
		NotifyURL:     in.NotifyURL,
	}
	if err := s.payCfg.Upsert(ctx, m); err != nil {
		return nil, err
	}
	return s.payCfg.FindByTenantProvider(ctx, tid, in.Provider)
}

// ========== 平台侧：收款审核 ==========

func (s *SettingsService) ListPaymentAudit(ctx context.Context, status *int8, page, size int) ([]model.TenantPaymentConfig, int64, error) {
	return s.payCfg.ListForAudit(ctx, status, page, size)
}

func (s *SettingsService) AuditPayment(ctx context.Context, id uint64, approve bool, remark string) error {
	return s.payCfg.Audit(ctx, id, approve, remark)
}

// ========== 商户侧：物流承运商（只读） ==========

// ListCarriersForTenant 商户端仅能看到平台启用的承运商列表；不返回敏感密钥字段。
func (s *SettingsService) ListCarriersForTenant(ctx context.Context) ([]model.ShippingCarrier, error) {
	if repository.EnsureTenant(ctx) == 0 {
		return nil, apperr.New(40001, "tenant required")
	}
	rows, err := s.carrier.ListEnabled(ctx)
	if err != nil {
		return nil, err
	}
	for i := range rows {
		rows[i].APIKey = ""
		rows[i].APISecret = ""
	}
	return rows, nil
}

// ========== 平台侧：物流承运商管理 ==========

type CarrierInput struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	APIProvider string `json:"api_provider"`
	APICustomer string `json:"api_customer"`
	APIKey      string `json:"api_key"`
	APISecret   string `json:"api_secret"`
	Priority    int    `json:"priority"`
	Enabled     *int8  `json:"enabled,omitempty"`
}

func (s *SettingsService) ListAllCarriers(ctx context.Context) ([]model.ShippingCarrier, error) {
	return s.carrier.ListAll(ctx)
}

func (s *SettingsService) CreateCarrier(ctx context.Context, in CarrierInput) (*model.ShippingCarrier, error) {
	if in.Code == "" || in.Name == "" {
		return nil, apperr.New(20001, "code 与 name 必填")
	}
	m := &model.ShippingCarrier{
		Code:        in.Code,
		Name:        in.Name,
		APIProvider: in.APIProvider,
		APICustomer: in.APICustomer,
		APIKey:      in.APIKey,
		APISecret:   in.APISecret,
		Priority:    in.Priority,
	}
	if in.Enabled != nil {
		m.Enabled = *in.Enabled
	} else {
		m.Enabled = 1
	}
	if err := s.carrier.Create(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *SettingsService) UpdateCarrier(ctx context.Context, id uint64, in CarrierInput) (*model.ShippingCarrier, error) {
	fields := map[string]interface{}{
		"code":         in.Code,
		"name":         in.Name,
		"api_provider": in.APIProvider,
		"api_customer": in.APICustomer,
		"priority":     in.Priority,
	}
	if in.APIKey != "" {
		fields["api_key"] = in.APIKey
	}
	if in.APISecret != "" {
		fields["api_secret"] = in.APISecret
	}
	if in.Enabled != nil {
		fields["enabled"] = *in.Enabled
	}
	if err := s.carrier.Update(ctx, id, fields); err != nil {
		return nil, err
	}
	return s.carrier.FindByID(ctx, id)
}

func (s *SettingsService) SetCarrierEnabled(ctx context.Context, id uint64, enabled bool) error {
	v := int8(0)
	if enabled {
		v = 1
	}
	return s.carrier.Update(ctx, id, map[string]interface{}{"enabled": v})
}

func (s *SettingsService) DeleteCarrier(ctx context.Context, id uint64) error {
	return s.carrier.Delete(ctx, id)
}

// ========== 第三方物流查询（商户+平台共用） ==========

type TrackNode struct {
	Time    time.Time `json:"time"`
	Context string    `json:"context"`
	Status  string    `json:"status"`
}

type TrackResult struct {
	CarrierCode string      `json:"carrier_code"`
	CarrierName string      `json:"carrier_name"`
	Provider    string      `json:"api_provider"`
	TrackingNo  string      `json:"tracking_no"`
	Status      string      `json:"status"`
	Nodes       []TrackNode `json:"nodes"`
}

// QueryTrack 物流轨迹查询。
// 真实 Provider 需接入 kuaidi100 / 阿里云快递 / 顺丰等官方接口；当前返回占位数据以保证流程跑通。
func (s *SettingsService) QueryTrack(ctx context.Context, carrierCode, trackingNo string) (*TrackResult, error) {
	if carrierCode == "" || trackingNo == "" {
		return nil, apperr.New(20001, "carrier_code 与 tracking_no 必填")
	}
	c, err := s.carrier.FindEnabledByCode(ctx, carrierCode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.New(40404, "承运商不存在或未启用")
		}
		return nil, err
	}
	now := time.Now()
	return &TrackResult{
		CarrierCode: c.Code,
		CarrierName: c.Name,
		Provider:    c.APIProvider,
		TrackingNo:  trackingNo,
		Status:      "transit",
		Nodes: []TrackNode{
			{Time: now.Add(-6 * time.Hour), Context: "已揽件", Status: "collected"},
			{Time: now.Add(-2 * time.Hour), Context: fmt.Sprintf("%s 运输中", c.Name), Status: "transit"},
		},
	}, nil
}
