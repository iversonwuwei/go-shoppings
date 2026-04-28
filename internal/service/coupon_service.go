package service

import (
	"context"
	"time"

	"gorm.io/gorm"

	"wechat-mall-saas/internal/model"
	"wechat-mall-saas/internal/pkg/ctxkeys"
	apperr "wechat-mall-saas/internal/pkg/errors"
	"wechat-mall-saas/internal/repository"
)

type CouponService struct {
	coupons       *repository.CouponRepo
	memberCoupons *repository.MemberCouponRepo
	tenants       *TenantService
}

func NewCouponService(c *repository.CouponRepo, m *repository.MemberCouponRepo, t *TenantService) *CouponService {
	return &CouponService{coupons: c, memberCoupons: m, tenants: t}
}

func (s *CouponService) List(ctx context.Context) ([]model.Coupon, error) {
	if err := s.tenants.RequireFeature(ctx, FeatureCoupon); err != nil {
		return nil, err
	}
	return s.coupons.List(ctx)
}

func (s *CouponService) Create(ctx context.Context, c *model.Coupon) error {
	if err := s.tenants.RequireFeature(ctx, FeatureCoupon); err != nil {
		return err
	}
	if c.ReceiveLimitType == "" {
		c.ReceiveLimitType = model.CouponReceiveLimitLimited
	}
	if c.ReceiveLimitType == model.CouponReceiveLimitUnlimited {
		c.TotalCount = 0
		c.RemainCount = 0
	} else {
		c.RemainCount = c.TotalCount
	}
	return s.coupons.Create(ctx, c)
}

func (s *CouponService) Receive(ctx context.Context, memberID, couponID uint64) (*model.MemberCoupon, error) {
	if err := s.tenants.RequireFeature(ctx, FeatureCoupon); err != nil {
		return nil, err
	}
	var mc *model.MemberCoupon
	err := s.coupons.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		c, err := s.coupons.FindByIDForUpdate(ctx, tx, couponID)
		if err != nil {
			return err
		}
		if c == nil || c.Status != 1 {
			return apperr.ErrNotFound
		}
		now := time.Now()
		if c.ReceiveStartAt != nil && now.Before(*c.ReceiveStartAt) {
			return apperr.New(30019, "优惠券尚未开始领取")
		}
		if c.ReceiveEndAt != nil && now.After(*c.ReceiveEndAt) {
			return apperr.New(30020, "优惠券已过领取期")
		}
		if c.PerLimit > 0 {
			received, err := s.memberCoupons.CountByMemberCouponTx(ctx, tx, memberID, c.ID)
			if err != nil {
				return err
			}
			if received >= int64(c.PerLimit) {
				return apperr.New(30022, "已达到该优惠券领取次数限制")
			}
		}
		validStart := now
		validEnd := now.AddDate(0, 0, 30)
		if c.ValidStartAt != nil {
			validStart = *c.ValidStartAt
		}
		if c.ValidEndAt != nil {
			validEnd = *c.ValidEndAt
		} else if c.ValidDays > 0 {
			validEnd = now.AddDate(0, 0, c.ValidDays)
		}

		mc = &model.MemberCoupon{
			TenantID:        ctxkeys.GetTenant(ctx).ID,
			MemberID:        memberID,
			CouponID:        c.ID,
			CouponName:      c.Name,
			CouponType:      c.Type,
			ThresholdAmount: c.ThresholdAmount,
			DiscountValue:   c.DiscountValue,
			MaxDiscount:     c.MaxDiscount,
			UseLimit:        c.UseLimit,
			ReceivedAt:      now,
			ValidStartAt:    validStart,
			ValidEndAt:      validEnd,
			Status:          "unused",
		}
		if c.ReceiveLimitType != model.CouponReceiveLimitUnlimited {
			if err := s.coupons.DecreaseRemain(ctx, tx, c.ID); err != nil {
				return apperr.New(30021, "优惠券已领完")
			}
		}
		return tx.Create(mc).Error
	})
	if err != nil {
		return nil, err
	}
	return mc, nil
}

func (s *CouponService) MyCoupons(ctx context.Context, memberID uint64, status string) ([]model.MemberCoupon, error) {
	return s.memberCoupons.ListByMember(ctx, memberID, status)
}
