package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"gorm.io/gorm"

	"wechat-mall-saas/internal/model"
	apperr "wechat-mall-saas/internal/pkg/errors"
	"wechat-mall-saas/internal/repository"
)

// SettingsService 管理商户收款（商户提交+平台审核）与物流承运商（平台统一维护）
type SettingsService struct {
	payCfg           *repository.PaymentConfigRepo
	carrier          *repository.ShippingCarrierRepo
	afterSaleReasons *repository.AfterSaleReasonRepo
	tenant           *TenantService
}

func NewSettingsService(p *repository.PaymentConfigRepo, c *repository.ShippingCarrierRepo, r *repository.AfterSaleReasonRepo, t *TenantService) *SettingsService {
	return &SettingsService{payCfg: p, carrier: c, afterSaleReasons: r, tenant: t}
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
	Provider              string `json:"provider"`
	MchID                 string `json:"mch_id"`
	AppID                 string `json:"app_id"`
	SpAppID               string `json:"sp_appid"`
	SpMchID               string `json:"sp_mchid"`
	SubAppID              string `json:"sub_appid"`
	SubMchID              string `json:"sub_mchid"`
	SettlementAccountName string `json:"settlement_account_name"`
	SettlementAccountNo   string `json:"settlement_account_no"`
	SettlementBankName    string `json:"settlement_bank_name"`
	SettlementRemark      string `json:"settlement_remark"`
	APIV3Key              string `json:"api_v3_key"`
	CertSerialNo          string `json:"cert_serial_no"`
	PrivateKeyPEM         string `json:"private_key_pem"`
	CertPEM               string `json:"cert_pem"`
	NotifyURL             string `json:"notify_url"`
}

func (s *SettingsService) SubmitPaymentConfig(ctx context.Context, in PaymentConfigInput) (*model.TenantPaymentConfig, error) {
	tid := repository.EnsureTenant(ctx)
	if tid == 0 {
		return nil, apperr.New(40001, "tenant required")
	}
	if in.Provider == "" {
		in.Provider = "manual_settlement"
	}
	if in.Provider == "manual_settlement" && (strings.TrimSpace(in.SettlementAccountName) == "" || strings.TrimSpace(in.SettlementAccountNo) == "") {
		return nil, apperr.New(20001, "结算户名和结算账号必填")
	}
	m := &model.TenantPaymentConfig{
		TenantID:              tid,
		Provider:              in.Provider,
		MchID:                 in.MchID,
		AppID:                 in.AppID,
		SpAppID:               in.SpAppID,
		SpMchID:               in.SpMchID,
		SubAppID:              in.SubAppID,
		SubMchID:              in.SubMchID,
		SettlementAccountName: in.SettlementAccountName,
		SettlementAccountNo:   in.SettlementAccountNo,
		SettlementBankName:    in.SettlementBankName,
		SettlementRemark:      in.SettlementRemark,
		APIV3Key:              in.APIV3Key,
		CertSerialNo:          in.CertSerialNo,
		PrivateKeyPEM:         in.PrivateKeyPEM,
		CertPEM:               in.CertPEM,
		NotifyURL:             in.NotifyURL,
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

type CarrierOption struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type CarrierMatchResult struct {
	Matched bool           `json:"matched"`
	Carrier *CarrierOption `json:"carrier,omitempty"`
	Source  string         `json:"source,omitempty"`
}

func (s *SettingsService) ListCarrierOptionsForTenant(ctx context.Context) ([]CarrierOption, error) {
	rows, err := s.ListCarriersForTenant(ctx)
	if err != nil {
		return nil, err
	}
	options := make([]CarrierOption, 0, len(rows))
	for _, row := range rows {
		options = append(options, CarrierOption{Code: row.Code, Name: row.Name})
	}
	return options, nil
}

func (s *SettingsService) MatchCarrierByTrackingNo(ctx context.Context, trackingNo string) (*CarrierMatchResult, error) {
	if repository.EnsureTenant(ctx) == 0 {
		return nil, apperr.New(40001, "tenant required")
	}
	trackingNo = normalizeTrackingNo(trackingNo)
	if trackingNo == "" {
		return nil, apperr.New(20001, "tracking_no 必填")
	}
	rows, err := s.carrier.ListEnabled(ctx)
	if err != nil {
		return nil, err
	}
	if option, ok := s.matchCarrierByKuaidi100(ctx, trackingNo, rows); ok {
		return &CarrierMatchResult{Matched: true, Carrier: option, Source: "kuaidi100"}, nil
	}
	for _, code := range inferCarrierCodes(trackingNo) {
		for _, row := range rows {
			if carrierMatchesCode(row, code) {
				option := CarrierOption{Code: row.Code, Name: row.Name}
				return &CarrierMatchResult{Matched: true, Carrier: &option, Source: "local"}, nil
			}
		}
	}
	return &CarrierMatchResult{Matched: false}, nil
}

func (s *SettingsService) matchCarrierByKuaidi100(ctx context.Context, trackingNo string, rows []model.ShippingCarrier) (*CarrierOption, bool) {
	apiKey := kuaidi100APIKey(rows)
	if apiKey == "" {
		return nil, false
	}
	codes, err := queryKuaidi100AutoNumber(ctx, apiKey, trackingNo)
	if err != nil {
		return nil, false
	}
	for _, code := range codes {
		for _, row := range rows {
			if carrierMatchesCode(row, code) {
				return &CarrierOption{Code: row.Code, Name: row.Name}, true
			}
		}
	}
	return nil, false
}

func kuaidi100APIKey(rows []model.ShippingCarrier) string {
	for _, row := range rows {
		if strings.EqualFold(strings.TrimSpace(row.APIProvider), "kuaidi100") {
			if key := strings.TrimSpace(row.APIKey); key != "" {
				return key
			}
		}
	}
	return ""
}

type kuaidi100AutoNumberResp struct {
	ComCode string `json:"comCode"`
	Auto    []struct {
		ComCode string `json:"comCode"`
	} `json:"auto"`
}

func queryKuaidi100AutoNumber(ctx context.Context, apiKey, trackingNo string) ([]string, error) {
	endpoint, err := url.Parse("https://poll.kuaidi100.com/autonumber/autoComNum")
	if err != nil {
		return nil, err
	}
	query := endpoint.Query()
	query.Set("key", apiKey)
	query.Set("num", trackingNo)
	query.Set("text", trackingNo)
	endpoint.RawQuery = query.Encode()

	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("kuaidi100 autonumber status %d", resp.StatusCode)
	}

	var payload kuaidi100AutoNumberResp
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	codes := make([]string, 0, len(payload.Auto)+1)
	if code := strings.TrimSpace(payload.ComCode); code != "" {
		codes = append(codes, code)
	}
	for _, item := range payload.Auto {
		if code := strings.TrimSpace(item.ComCode); code != "" {
			codes = append(codes, code)
		}
	}
	return uniqueStrings(codes), nil
}

func uniqueStrings(values []string) []string {
	out := make([]string, 0, len(values))
	seen := make(map[string]bool, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		out = append(out, value)
		seen[value] = true
	}
	return out
}

func normalizeTrackingNo(value string) string {
	value = strings.ToUpper(strings.TrimSpace(value))
	var b strings.Builder
	for _, r := range value {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func inferCarrierCodes(trackingNo string) []string {
	switch {
	case strings.HasPrefix(trackingNo, "SF"):
		return []string{"shunfeng"}
	case strings.HasPrefix(trackingNo, "ZTO"):
		return []string{"zhongtong"}
	case strings.HasPrefix(trackingNo, "YTO"), strings.HasPrefix(trackingNo, "YT"):
		return []string{"yuantong"}
	case strings.HasPrefix(trackingNo, "STO"):
		return []string{"shentong"}
	case strings.HasPrefix(trackingNo, "YD"):
		return []string{"yunda"}
	case strings.HasPrefix(trackingNo, "EMS"):
		return []string{"ems"}
	case strings.HasPrefix(trackingNo, "JD"):
		return []string{"jd"}
	case strings.HasPrefix(trackingNo, "JT"), strings.HasPrefix(trackingNo, "JTE"):
		return []string{"jtexpress"}
	case strings.HasPrefix(trackingNo, "DB"):
		return []string{"debangwuliu"}
	case strings.HasPrefix(trackingNo, "DN"):
		return []string{"danniao"}
	case strings.HasPrefix(trackingNo, "ANE"):
		return []string{"annengwuliu"}
	case strings.HasPrefix(trackingNo, "KYS"):
		return []string{"kuayue"}
	default:
		return nil
	}
}

func carrierMatchesCode(row model.ShippingCarrier, code string) bool {
	keys := []string{strings.ToLower(row.Code), strings.ToLower(row.Name)}
	for _, alias := range carrierAliases[code] {
		keys = append(keys, strings.ToLower(alias))
	}
	targets := append([]string{code}, carrierAliases[code]...)
	for _, key := range keys {
		for _, target := range targets {
			target = strings.ToLower(target)
			if key == target || strings.Contains(key, target) {
				return true
			}
		}
	}
	return false
}

var carrierAliases = map[string][]string{
	"shunfeng":    {"sf", "shunfeng", "顺丰", "顺丰速运"},
	"zhongtong":   {"zto", "zhongtong", "中通", "中通快递"},
	"yuantong":    {"yto", "yuantong", "圆通", "圆通速递"},
	"yunda":       {"yd", "yunda", "韵达", "韵达快递"},
	"shentong":    {"sto", "shentong", "申通", "申通快递"},
	"ems":         {"ems", "邮政", "中国邮政"},
	"jd":          {"jd", "jingdong", "京东", "京东物流"},
	"jtexpress":   {"jt", "jtexpress", "极兔", "极兔速递"},
	"debangwuliu": {"db", "debang", "德邦", "德邦快递"},
	"danniao":     {"dn", "danniao", "丹鸟", "丹鸟物流"},
	"annengwuliu": {"ane", "anneng", "安能", "安能物流"},
	"kuayue":      {"kys", "kuayue", "跨越", "跨越速运"},
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

// ========== 平台侧：售后原因管理 ==========

type AfterSaleReasonInput struct {
	Code      string `json:"code"`
	Label     string `json:"label"`
	Type      string `json:"type"`
	SortOrder int    `json:"sort_order"`
	Enabled   *int8  `json:"enabled,omitempty"`
}

func (s *SettingsService) ListAfterSaleReasons(ctx context.Context) ([]model.AfterSaleReason, error) {
	return s.afterSaleReasons.ListAll(ctx)
}

func (s *SettingsService) CreateAfterSaleReason(ctx context.Context, in AfterSaleReasonInput) (*model.AfterSaleReason, error) {
	if in.Code == "" || in.Label == "" {
		return nil, apperr.New(20001, "code 与 label 必填")
	}
	reasonType, err := normalizeReasonInputType(in.Type)
	if err != nil {
		return nil, err
	}
	row := &model.AfterSaleReason{Code: in.Code, Label: in.Label, Type: reasonType, SortOrder: in.SortOrder, Enabled: 1}
	if in.Enabled != nil {
		row.Enabled = *in.Enabled
	}
	if err := s.afterSaleReasons.Create(ctx, row); err != nil {
		return nil, err
	}
	return row, nil
}

func (s *SettingsService) UpdateAfterSaleReason(ctx context.Context, id uint64, in AfterSaleReasonInput) (*model.AfterSaleReason, error) {
	if in.Code == "" || in.Label == "" {
		return nil, apperr.New(20001, "code 与 label 必填")
	}
	reasonType, err := normalizeReasonInputType(in.Type)
	if err != nil {
		return nil, err
	}
	fields := map[string]interface{}{
		"code":       in.Code,
		"label":      in.Label,
		"type":       reasonType,
		"sort_order": in.SortOrder,
	}
	if in.Enabled != nil {
		fields["enabled"] = *in.Enabled
	}
	if err := s.afterSaleReasons.Update(ctx, id, fields); err != nil {
		return nil, err
	}
	return s.afterSaleReasons.FindByID(ctx, id)
}

func (s *SettingsService) SetAfterSaleReasonEnabled(ctx context.Context, id uint64, enabled bool) error {
	v := int8(0)
	if enabled {
		v = 1
	}
	return s.afterSaleReasons.Update(ctx, id, map[string]interface{}{"enabled": v})
}

func (s *SettingsService) DeleteAfterSaleReason(ctx context.Context, id uint64) error {
	row, err := s.afterSaleReasons.FindByID(ctx, id)
	if err != nil || row == nil {
		if err != nil {
			return err
		}
		return apperr.ErrNotFound
	}
	count, err := s.afterSaleReasons.CountAfterSaleUsage(ctx, row.Label)
	if err != nil {
		return err
	}
	if count > 0 {
		return apperr.New(30014, "售后原因已被使用，请停用代替删除")
	}
	return s.afterSaleReasons.Delete(ctx, id)
}

func normalizeReasonInputType(reasonType string) (string, error) {
	switch reasonType {
	case "", model.AfterSaleReasonTypeAll:
		return model.AfterSaleReasonTypeAll, nil
	case model.AfterSaleReasonTypeRefund, model.AfterSaleReasonTypeReturnRefund:
		return reasonType, nil
	default:
		return "", apperr.New(20001, "售后原因类型不合法")
	}
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
