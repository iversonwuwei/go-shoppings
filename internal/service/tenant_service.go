package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"wechat-mall-saas/internal/model"
	"wechat-mall-saas/internal/pkg/cache"
	"wechat-mall-saas/internal/pkg/ctxkeys"
	apperr "wechat-mall-saas/internal/pkg/errors"
	"wechat-mall-saas/internal/repository"
)

// TenantService 处理租户上下文加载 / 套餐校验等横切业务
type TenantService struct {
	tenants *repository.TenantRepo
	admins  *repository.AdminRepo
	plans   *repository.PlanRepo
	logs    *repository.TenantPlanLogRepo
	cache   *cache.Client
}

func NewTenantService(t *repository.TenantRepo, a *repository.AdminRepo, p *repository.PlanRepo, l *repository.TenantPlanLogRepo, c *cache.Client) *TenantService {
	return &TenantService{tenants: t, admins: a, plans: p, logs: l, cache: c}
}

// PublicPlans 返回对外展示的套餐列表（已启用）
func (s *TenantService) PublicPlans(ctx context.Context) ([]model.Plan, error) {
	rows, err := s.plans.List(ctx)
	if err != nil {
		return nil, err
	}
	out := rows[:0]
	for _, p := range rows {
		if p.Status == 1 {
			out = append(out, p)
		}
	}
	return out, nil
}

type tenantCache struct {
	ID           uint64    `json:"id"`
	Code         string    `json:"code"`
	PlanID       uint64    `json:"plan_id"`
	Status       int8      `json:"status"`
	PlanExpireAt time.Time `json:"plan_expire_at"`
	Features     []string  `json:"features"`
}

// LoadContext 根据租户 ID 加载 TenantInfo（含 5 分钟缓存）
func (s *TenantService) LoadContext(ctx context.Context, id uint64) (*ctxkeys.TenantInfo, error) {
	key := fmt.Sprintf("tenant:ctx:%d", id)
	if s.cache != nil {
		if v, err := s.cache.Get(ctx, key).Bytes(); err == nil && len(v) > 0 {
			var tc tenantCache
			if json.Unmarshal(v, &tc) == nil {
				return tcToInfo(&tc), nil
			}
		}
	}
	t, err := s.tenants.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, apperr.ErrTenantInvalid
	}
	p, err := s.plans.FindByID(ctx, t.PlanID)
	if err != nil {
		return nil, err
	}
	features := []string{}
	if p != nil {
		features = []string(p.Features)
	}
	// 并入平台单独授予的附加功能（extra_features），去重
	if len(t.ExtraFeatures) > 0 {
		seen := make(map[string]bool, len(features))
		for _, f := range features {
			seen[f] = true
		}
		for _, f := range t.ExtraFeatures {
			if f != "" && !seen[f] {
				features = append(features, f)
				seen[f] = true
			}
		}
	}
	tc := &tenantCache{
		ID: t.ID, Code: t.Code, PlanID: t.PlanID, Status: t.Status,
		PlanExpireAt: t.PlanExpireAt, Features: features,
	}
	if s.cache != nil {
		if bs, err := json.Marshal(tc); err == nil {
			_ = s.cache.Set(ctx, key, bs, 5*time.Minute).Err()
		}
	}
	return tcToInfo(tc), nil
}

func tcToInfo(tc *tenantCache) *ctxkeys.TenantInfo {
	return &ctxkeys.TenantInfo{
		ID:       tc.ID,
		Code:     tc.Code,
		PlanID:   tc.PlanID,
		Features: tc.Features,
		Status:   tc.Status,
		Expired:  !tc.PlanExpireAt.IsZero() && tc.PlanExpireAt.Before(time.Now()),
	}
}

// Invalidate 清除租户缓存
func (s *TenantService) Invalidate(ctx context.Context, id uint64) {
	if s.cache != nil {
		_ = s.cache.Del(ctx, fmt.Sprintf("tenant:ctx:%d", id)).Err()
	}
}

// RequireFeature 校验当前租户是否开通指定功能
func (s *TenantService) RequireFeature(ctx context.Context, feat string) error {
	t := ctxkeys.GetTenant(ctx)
	if t == nil {
		return apperr.ErrTenantRequired
	}
	if t.Expired {
		return apperr.ErrPlanExpired
	}
	if !ctxkeys.HasFeature(t, feat) {
		return apperr.ErrFeatureDisabled
	}
	return nil
}

// HasFeature 用于无 tenant ctx 的场景（如支付回调），通过 tenantID 查询并判断功能是否开通。
func (s *TenantService) HasFeature(ctx context.Context, tenantID uint64, feat string) bool {
	if tenantID == 0 {
		return false
	}
	info, err := s.LoadContext(ctx, tenantID)
	if err != nil || info == nil || info.Expired {
		return false
	}
	return ctxkeys.HasFeature(info, feat)
}

// CheckProductLimit 校验商品数量上限
func (s *TenantService) CheckProductLimit(ctx context.Context, current int64) error {
	t := ctxkeys.GetTenant(ctx)
	if t == nil {
		return apperr.ErrTenantRequired
	}
	p, err := s.plans.FindByID(ctx, t.PlanID)
	if err != nil || p == nil {
		return apperr.ErrInternal
	}
	if p.ProductLimit > 0 && current >= int64(p.ProductLimit) {
		return apperr.ErrLimitExceeded
	}
	return nil
}

// CheckOrderLimit 校验月订单上限
func (s *TenantService) CheckOrderLimit(ctx context.Context, monthCount int64) error {
	t := ctxkeys.GetTenant(ctx)
	if t == nil {
		return apperr.ErrTenantRequired
	}
	p, err := s.plans.FindByID(ctx, t.PlanID)
	if err != nil || p == nil {
		return apperr.ErrInternal
	}
	if p.OrderLimit > 0 && monthCount >= int64(p.OrderLimit) {
		return apperr.ErrLimitExceeded
	}
	return nil
}

// Register 提交入驻申请（pending 状态，默认套餐 30 天试用）
func (s *TenantService) Register(ctx context.Context, in *model.Tenant) (*model.Tenant, error) {
	if in.Code == "" || in.CompanyName == "" || in.ContactPhone == "" {
		return nil, apperr.ErrParamInvalid
	}
	if exist, _ := s.tenants.FindByCode(ctx, in.Code); exist != nil {
		return nil, apperr.ErrDuplicated
	}
	if in.PlanID == 0 {
		if dp, _ := s.plans.FindDefault(ctx); dp != nil {
			in.PlanID = dp.ID
		}
	}
	// 申请时给 7 天试用期；真正的试用期起点在平台审核通过时会被重置。
	if in.PlanExpireAt.IsZero() {
		in.PlanExpireAt = time.Now().AddDate(0, 0, 7)
	}
	// 计费周期：仅允许 monthly / yearly，默认 yearly
	if in.BillingCycle != "monthly" && in.BillingCycle != "yearly" {
		in.BillingCycle = "yearly"
	}
	in.Status = TenantStatusPending
	if err := s.tenants.Create(ctx, in); err != nil {
		return nil, err
	}
	return in, nil
}

// Audit 审核租户（通过=1，拒绝=带 reason）
func (s *TenantService) Audit(ctx context.Context, id uint64, approve bool, reason string) error {
	fields := map[string]interface{}{}
	if approve {
		fields["status"] = TenantStatusActive
		// 审核通过时重置 7 天试用期的起点
		fields["plan_expire_at"] = time.Now().AddDate(0, 0, 7)
	} else {
		fields["status"] = TenantStatusBanned
		fields["reject_reason"] = reason
	}
	if err := s.tenants.UpdateFields(ctx, id, fields); err != nil {
		return err
	}
	// 联动更新该租户下管理员账号的启用状态
	if s.admins != nil {
		var adminStatus int8 = 0
		if approve {
			adminStatus = 1
		}
		_ = s.admins.UpdateStatusByTenant(ctx, id, adminStatus)
	}
	s.Invalidate(ctx, id)
	return nil
}

// SetStatus 平台管理员手动封禁 / 恢复租户
// status: 1=正常  3=封禁
func (s *TenantService) SetStatus(ctx context.Context, id uint64, status int8, reason string) error {
	t, err := s.tenants.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if t == nil {
		return apperr.ErrTenantInvalid
	}
	if status != TenantStatusActive && status != TenantStatusBanned {
		return apperr.ErrParamInvalid
	}
	fields := map[string]interface{}{"status": status}
	if status == TenantStatusBanned {
		fields["reject_reason"] = reason
	} else {
		fields["reject_reason"] = ""
	}
	if err := s.tenants.UpdateFields(ctx, id, fields); err != nil {
		return err
	}
	if s.admins != nil {
		var adminStatus int8 = 1
		if status == TenantStatusBanned {
			adminStatus = 0
		}
		_ = s.admins.UpdateStatusByTenant(ctx, id, adminStatus)
	}
	s.Invalidate(ctx, id)
	return nil
}

// SetPlan 平台管理员修改租户套餐 / 续期
func (s *TenantService) SetPlan(ctx context.Context, id, planID uint64, expireAt *time.Time) error {
	if planID == 0 {
		return apperr.ErrParamInvalid
	}
	p, err := s.plans.FindByID(ctx, planID)
	if err != nil {
		return err
	}
	if p == nil {
		return apperr.ErrParamInvalid
	}
	fields := map[string]interface{}{"plan_id": planID}
	if expireAt != nil && !expireAt.IsZero() {
		fields["plan_expire_at"] = *expireAt
	}
	if err := s.tenants.UpdateFields(ctx, id, fields); err != nil {
		return err
	}
	s.Invalidate(ctx, id)
	return nil
}

// SetExtraFeatures 平台管理员为租户单独授予 / 撤销附加功能（与套餐功能取并集）
func (s *TenantService) SetExtraFeatures(ctx context.Context, id uint64, codes []string) error {
	// 去重
	seen := map[string]bool{}
	out := make([]string, 0, len(codes))
	for _, c := range codes {
		if c == "" || seen[c] {
			continue
		}
		seen[c] = true
		out = append(out, c)
	}
	bs, err := json.Marshal(out)
	if err != nil {
		return err
	}
	if err := s.tenants.UpdateFields(ctx, id, map[string]interface{}{"extra_features": bs}); err != nil {
		return err
	}
	s.Invalidate(ctx, id)
	return nil
}
