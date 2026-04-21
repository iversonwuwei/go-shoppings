package service

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"

	"wechat-mall-saas/internal/model"
	apperr "wechat-mall-saas/internal/pkg/errors"
	"wechat-mall-saas/internal/pkg/wxpay"
	"wechat-mall-saas/internal/repository"
)

// 宽限期（天）：逾期 N 天转欠费，M 天封禁
const (
	TrialDays        = 7
	GraceOverdueDays = 3
	GraceBannedDays  = 5
)

// SubscriptionService 管理租户订阅订单 / 支付 / 到期联动
type SubscriptionService struct {
	orders       *repository.TenantSubscriptionOrderRepo
	tenants      *repository.TenantRepo
	plans        *repository.PlanRepo
	planLogs     *repository.TenantPlanLogRepo
	tenant       *TenantService
	wxpay        *wxpay.Client
	platSettings *repository.PlatformSettingsRepo
}

func NewSubscriptionService(
	orders *repository.TenantSubscriptionOrderRepo,
	tenants *repository.TenantRepo,
	plans *repository.PlanRepo,
	planLogs *repository.TenantPlanLogRepo,
	tenant *TenantService,
	wp *wxpay.Client,
	platSettings *repository.PlatformSettingsRepo,
) *SubscriptionService {
	return &SubscriptionService{
		orders: orders, tenants: tenants, plans: plans,
		planLogs: planLogs, tenant: tenant, wxpay: wp,
		platSettings: platSettings,
	}
}

// resolveWxpayClient 运行时解析：优先使用平台设置里的配置，未填写则回退到启动时注入的 s.wxpay
func (s *SubscriptionService) resolveWxpayClient(ctx context.Context) *wxpay.Client {
	if s.platSettings == nil {
		return s.wxpay
	}
	ps, err := s.platSettings.Get(ctx)
	if err != nil || ps == nil || ps.WxpayAppID == "" || ps.WxpayMchID == "" {
		return s.wxpay
	}
	return wxpay.NewClient(wxpay.Config{
		AppID:      ps.WxpayAppID,
		MchID:      ps.WxpayMchID,
		APIv3Key:   ps.WxpayAPIv3Key,
		CertSerial: ps.WxpayCertSerial,
		NotifyURL:  ps.WxpayNotifyURL,
	})
}

// IsInTrial 判断租户当前是否处于试用期（仅审核通过后计算）。
func (s *SubscriptionService) IsInTrial(t *model.Tenant) bool {
	if t == nil || t.Status != TenantStatusActive {
		return false
	}
	// 试用期 = plan_expire_at 尚未到期 且 从未付费（无任何已付订单）
	// 简化判断：plan_expire_at - created_at 如果接近 7 天内且无成功订单 => 试用期。
	// 我们用显式"有无成功订单"判定，避免时间漂移误判。
	return true // 详细判断由调用方结合已支付订单
}

// CreateOrder 创建订阅订单并返回 JSAPI 支付参数
// openID 可为空（若平台 AppID 配置为 H5/Native，传空即可；JSAPI 下单需要 OpenID）
func (s *SubscriptionService) CreateOrder(ctx context.Context, tenantID, planID uint64, billingCycle, openID string) (*model.TenantSubscriptionOrder, *wxpay.JSAPIPayParams, error) {
	if tenantID == 0 {
		return nil, nil, apperr.ErrTenantRequired
	}
	if billingCycle != "monthly" && billingCycle != "yearly" {
		return nil, nil, apperr.New(20001, "计费周期必须是 monthly 或 yearly")
	}
	t, err := s.tenants.FindByID(ctx, tenantID)
	if err != nil {
		return nil, nil, err
	}
	if t == nil {
		return nil, nil, apperr.ErrTenantInvalid
	}
	if t.Status == TenantStatusPending {
		return nil, nil, apperr.New(30020, "租户尚未通过审核，无法创建订阅订单")
	}
	if t.Status == TenantStatusBanned {
		return nil, nil, apperr.New(30021, "租户已封禁，请联系平台")
	}
	// 试用期内禁止切换套餐
	if planID != 0 && planID != t.PlanID {
		hasPaid, err := s.hasAnyPaidOrder(ctx, tenantID)
		if err != nil {
			return nil, nil, err
		}
		if !hasPaid {
			return nil, nil, apperr.New(30022, "试用期内不能切换套餐，请先完成首次付费")
		}
	}
	targetPlanID := planID
	if targetPlanID == 0 {
		targetPlanID = t.PlanID
	}
	p, err := s.plans.FindByID(ctx, targetPlanID)
	if err != nil {
		return nil, nil, err
	}
	if p == nil {
		return nil, nil, apperr.New(20001, "套餐不存在")
	}
	amount := p.MonthlyFee
	if billingCycle == "yearly" {
		amount = p.YearlyFee
	}
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, nil, apperr.New(20001, "套餐价格无效，请联系平台")
	}

	orderNo := fmt.Sprintf("TSUB%d%06d", time.Now().UnixNano()/1e6, tenantID%1000000)
	order := &model.TenantSubscriptionOrder{
		TenantID:     tenantID,
		PlanID:       targetPlanID,
		BillingCycle: billingCycle,
		Amount:       amount,
		Status:       0,
		OrderNo:      orderNo,
		ExpireBefore: &t.PlanExpireAt,
	}
	if err := s.orders.Create(ctx, order); err != nil {
		return nil, nil, err
	}

	// 发起微信支付下单（平台统一商户号）
	if s.wxpay == nil {
		return order, nil, apperr.New(40002, "平台未配置微信支付")
	}
	client := s.resolveWxpayClient(ctx)
	totalFen := amount.Mul(decimal.NewFromInt(100)).IntPart()
	pay, err := client.PlaceJSAPIOrder(ctx, wxpay.JSAPIOrderReq{
		Description: fmt.Sprintf("%s 订阅-%s", p.Name, billingCycleLabel(billingCycle)),
		OutTradeNo:  orderNo,
		TotalFen:    totalFen,
		OpenID:      openID,
	})
	return order, pay, err
}

func billingCycleLabel(c string) string {
	if c == "monthly" {
		return "按月"
	}
	return "按年"
}

func (s *SubscriptionService) hasAnyPaidOrder(ctx context.Context, tenantID uint64) (bool, error) {
	rows, _, err := s.orders.ListByTenant(ctx, tenantID, 1, 1000)
	if err != nil {
		return false, err
	}
	for _, o := range rows {
		if o.Status == 1 {
			return true, nil
		}
	}
	return false, nil
}

// OnPaySuccess 支付成功回调：延长 plan_expire_at、切换套餐、写日志
func (s *SubscriptionService) OnPaySuccess(ctx context.Context, orderNo, transactionID string, paidAt time.Time) error {
	o, err := s.orders.FindByOrderNo(ctx, orderNo)
	if err != nil {
		return err
	}
	if o == nil {
		return apperr.ErrNotFound
	}
	if o.Status == 1 {
		return nil // 幂等
	}
	t, err := s.tenants.FindByID(ctx, o.TenantID)
	if err != nil {
		return err
	}
	if t == nil {
		return apperr.ErrTenantInvalid
	}
	// 计算新到期：max(now, 当前到期) + 周期
	base := t.PlanExpireAt
	now := time.Now()
	if base.Before(now) {
		base = now
	}
	var newExpire time.Time
	if o.BillingCycle == "monthly" {
		newExpire = base.AddDate(0, 1, 0)
	} else {
		newExpire = base.AddDate(1, 0, 0)
	}
	// 订单落账
	if err := s.orders.MarkPaid(ctx, orderNo, transactionID, paidAt, newExpire); err != nil {
		return err
	}
	// 租户：更新 plan_id / plan_expire_at / billing_cycle / 状态恢复 Active
	fields := map[string]interface{}{
		"plan_id":        o.PlanID,
		"plan_expire_at": newExpire,
		"billing_cycle":  o.BillingCycle,
	}
	if t.Status == TenantStatusOverdue || t.Status == TenantStatusBanned {
		fields["status"] = TenantStatusActive
	}
	if err := s.tenants.UpdateFields(ctx, t.ID, fields); err != nil {
		return err
	}
	// 变更日志
	changeType := "renew"
	if o.PlanID != t.PlanID {
		changeType = "upgrade"
	}
	_ = s.planLogs.Create(ctx, &model.TenantPlanLog{
		TenantID:    t.ID,
		OldPlanID:   t.PlanID,
		NewPlanID:   o.PlanID,
		ChangeType:  changeType,
		EffectiveAt: now,
		ExpireAt:    newExpire,
		Amount:      o.Amount,
	})
	s.tenant.Invalidate(ctx, t.ID)
	return nil
}

// ListOrders 商户后台：按租户分页列出订阅订单
func (s *SubscriptionService) ListOrders(ctx context.Context, tenantID uint64, page, pageSize int) ([]model.TenantSubscriptionOrder, int64, error) {
	return s.orders.ListByTenant(ctx, tenantID, page, pageSize)
}

// ScanAndTransition 扫描到期租户，按宽限期切换状态：
//   - Active  且  plan_expire_at < now - 3d  => Overdue
//   - Overdue 且  plan_expire_at < now - 5d  => Banned
func (s *SubscriptionService) ScanAndTransition(ctx context.Context) (overdue, banned int, err error) {
	now := time.Now()
	overdueCutoff := now.AddDate(0, 0, -GraceOverdueDays)
	bannedCutoff := now.AddDate(0, 0, -GraceBannedDays)

	// Active -> Overdue
	list, err := s.tenants.ScanByExpireCutoff(ctx, TenantStatusActive, overdueCutoff)
	if err != nil {
		return 0, 0, err
	}
	for _, t := range list {
		_ = s.tenants.UpdateFields(ctx, t.ID, map[string]interface{}{"status": TenantStatusOverdue})
		s.tenant.Invalidate(ctx, t.ID)
		overdue++
	}
	// Overdue -> Banned
	list2, err := s.tenants.ScanByExpireCutoff(ctx, TenantStatusOverdue, bannedCutoff)
	if err != nil {
		return overdue, 0, err
	}
	for _, t := range list2 {
		_ = s.tenants.UpdateFields(ctx, t.ID, map[string]interface{}{"status": TenantStatusBanned})
		s.tenant.Invalidate(ctx, t.ID)
		banned++
	}
	return overdue, banned, nil
}
