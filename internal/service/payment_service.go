package service

import (
	"context"
	"time"

	"github.com/shopspring/decimal"

	"wechat-mall-saas/internal/model"
	"wechat-mall-saas/internal/pkg/ctxkeys"
	apperr "wechat-mall-saas/internal/pkg/errors"
	"wechat-mall-saas/internal/pkg/utils"
	"wechat-mall-saas/internal/pkg/wxpay"
	"wechat-mall-saas/internal/repository"
)

type PaymentService struct {
	payments   *repository.PaymentRepo
	orders     *repository.OrderRepo
	logs       *repository.OrderLogRepo
	tenants    *repository.TenantRepo
	members    *repository.MemberRepo
	pointsLogs *repository.PointsLogRepo
	pointsCfg  *repository.PointsSettingsRepo
	tenantSvc  *TenantService
}

func NewPaymentService(
	p *repository.PaymentRepo, o *repository.OrderRepo, l *repository.OrderLogRepo,
	t *repository.TenantRepo, m *repository.MemberRepo, pl *repository.PointsLogRepo,
	ps *repository.PointsSettingsRepo, ts *TenantService,
) *PaymentService {
	return &PaymentService{payments: p, orders: o, logs: l, tenants: t, members: m, pointsLogs: pl, pointsCfg: ps, tenantSvc: ts}
}

type CreatePaymentResult struct {
	PaymentNo string                `json:"payment_no"`
	PayParams *wxpay.JSAPIPayParams `json:"pay_params,omitempty"`
}

// Create 创建支付单并向微信统一下单，返回小程序支付参数
func (s *PaymentService) Create(ctx context.Context, memberID uint64, openID, orderNo, notifyURL string) (*CreatePaymentResult, error) {
	order, err := s.orders.FindByNo(ctx, orderNo)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, apperr.ErrNotFound
	}
	if order.MemberID != memberID {
		return nil, apperr.ErrForbidden
	}
	if order.Status != model.OrderStatusPendingPay {
		return nil, apperr.New(30010, "订单状态不可支付")
	}
	t, _ := s.tenants.FindByID(ctx, ctxkeys.GetTenant(ctx).ID)
	if t == nil {
		return nil, apperr.ErrTenantInvalid
	}

	p := &model.Payment{
		PaymentNo: utils.OrderNo("P"),
		OrderNo:   order.OrderNo,
		MemberID:  memberID,
		Amount:    order.ActualAmount,
		Status:    model.PaymentStatusPending,
	}
	exp := time.Now().Add(30 * time.Minute)
	p.ExpireAt = &exp
	if err := s.payments.Create(ctx, p); err != nil {
		return nil, err
	}

	wx := wxpay.NewClient(wxpay.Config{
		AppID:      t.WechatAppID,
		MchID:      t.WechatMchID,
		APIv3Key:   t.WechatAPIv3Key,
		CertSerial: t.WechatCertSerial,
		NotifyURL:  notifyURL,
	})
	params, err := wx.PlaceJSAPIOrder(ctx, wxpay.JSAPIOrderReq{
		Description: "订单 " + order.OrderNo,
		OutTradeNo:  p.PaymentNo,
		TotalFen:    order.ActualAmount.Mul(decimal.NewFromInt(100)).IntPart(),
		OpenID:      openID,
	})
	if err != nil {
		return nil, apperr.ErrWechatPay
	}
	return &CreatePaymentResult{PaymentNo: p.PaymentNo, PayParams: params}, nil
}

// HandleCallback 处理微信支付回调（已验签/解密后的 TransactionInfo）
func (s *PaymentService) HandleCallback(ctx context.Context, info *wxpay.TransactionInfo) error {
	if info == nil || info.TradeState != "SUCCESS" {
		return apperr.ErrWechatPay
	}
	p, err := s.payments.FindByNoRaw(ctx, info.OutTradeNo)
	if err != nil {
		return err
	}
	if p == nil {
		return apperr.ErrNotFound
	}
	if p.Status == model.PaymentStatusPaid {
		return nil // 幂等
	}
	now := time.Now()
	if err := s.payments.UpdateFields(ctx, p.ID, map[string]interface{}{
		"status":                model.PaymentStatusPaid,
		"wechat_transaction_id": info.TransactionID,
		"wechat_payer_openid":   info.Payer.OpenID,
		"wechat_paid_at":        now,
	}); err != nil {
		return err
	}
	// 更新订单状态为已支付（幂等：仅从 pending_pay 迁移）
	// 虚拟商品订单直接置为已完成，无需发货/确认收货。
	_ = s.orders.DB().WithContext(ctx).Model(&model.Order{}).
		Where("order_no = ? AND tenant_id = ? AND status = ? AND is_virtual = 0", p.OrderNo, p.TenantID, model.OrderStatusPendingPay).
		Updates(map[string]interface{}{"status": model.OrderStatusPaid, "paid_at": now}).Error
	_ = s.orders.DB().WithContext(ctx).Model(&model.Order{}).
		Where("order_no = ? AND tenant_id = ? AND status = ? AND is_virtual = 1", p.OrderNo, p.TenantID, model.OrderStatusPendingPay).
		Updates(map[string]interface{}{"status": model.OrderStatusCompleted, "paid_at": now, "completed_at": now}).Error

	// 发放积分（仅当租户套餐启用 points 功能且设置启用）
	s.grantPoints(ctx, p)
	return nil
}

// grantPoints 根据 points_settings 规则给会员发放积分。
// - 检查租户是否开通 points 功能
// - 检查设置 enabled=1 且订单金额 >= min_amount
// - 计算 points = floor(amount * earn_rate * level.PointsMult)
// - 写 points_logs + member.points 累加
// 失败不影响支付回调主流程（仅打 best-effort）。
func (s *PaymentService) grantPoints(ctx context.Context, p *model.Payment) {
	if s.tenantSvc == nil || s.pointsCfg == nil || s.members == nil || s.pointsLogs == nil {
		return
	}
	if !s.tenantSvc.HasFeature(ctx, p.TenantID, FeaturePoints) {
		return
	}
	var ps model.PointsSetting
	if err := s.orders.DB().WithContext(ctx).
		Where("tenant_id = ?", p.TenantID).First(&ps).Error; err != nil {
		return
	}
	if ps.Enabled != 1 {
		return
	}
	var order model.Order
	if err := s.orders.DB().WithContext(ctx).
		Where("order_no = ? AND tenant_id = ?", p.OrderNo, p.TenantID).First(&order).Error; err != nil {
		return
	}
	if order.ActualAmount.LessThan(ps.MinAmount) {
		return
	}
	var member model.Member
	if err := s.orders.DB().WithContext(ctx).
		Where("id = ? AND tenant_id = ?", p.MemberID, p.TenantID).First(&member).Error; err != nil {
		return
	}
	// 等级倍率
	mult := decimal.NewFromInt(1)
	if member.LevelID > 0 {
		var lvl model.MemberLevel
		if err := s.orders.DB().WithContext(ctx).
			Where("id = ? AND tenant_id = ?", member.LevelID, p.TenantID).First(&lvl).Error; err == nil {
			if lvl.PointsMult.GreaterThan(decimal.Zero) {
				mult = lvl.PointsMult
			}
		}
	}
	earn := order.ActualAmount.Mul(ps.EarnRate).Mul(mult).IntPart()
	if earn <= 0 {
		return
	}
	before := member.Points
	after := before + int(earn)
	// 更新会员积分
	_ = s.orders.DB().WithContext(ctx).Model(&model.Member{}).
		Where("id = ? AND tenant_id = ?", member.ID, p.TenantID).
		Update("points", after).Error
	// 写积分日志
	_ = s.orders.DB().WithContext(ctx).Create(&model.PointsLog{
		TenantID:      p.TenantID,
		MemberID:      member.ID,
		ChangeType:    "earn",
		ChangeValue:   int(earn),
		BalanceBefore: before,
		BalanceAfter:  after,
		SourceID:      order.ID,
		SourceDesc:    "订单 " + order.OrderNo,
	}).Error
}
