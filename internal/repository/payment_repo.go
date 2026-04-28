package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"wechat-mall-saas/internal/model"
)

type PaymentRepo struct{ db *gorm.DB }

var (
	ErrCouponUnavailable     = errors.New("coupon unavailable")
	ErrCouponAlreadyUsed     = errors.New("coupon already used")
	ErrCouponUseLimitReached = errors.New("coupon use limit reached")
)

func NewPaymentRepo(db *gorm.DB) *PaymentRepo { return &PaymentRepo{db: db} }

func (r *PaymentRepo) Create(ctx context.Context, p *model.Payment) error {
	p.TenantID = EnsureTenant(ctx)
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *PaymentRepo) FindByNo(ctx context.Context, no string) (*model.Payment, error) {
	var p model.Payment
	if err := TenantDB(ctx, r.db).Where("payment_no = ?", no).First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

// FindByNoRaw 不受租户限制，供支付回调使用（回调外部请求，无 tenant ctx）
func (r *PaymentRepo) FindByNoRaw(ctx context.Context, no string) (*model.Payment, error) {
	var p model.Payment
	if err := r.db.WithContext(ctx).Where("payment_no = ?", no).First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *PaymentRepo) UpdateFields(ctx context.Context, id uint64, fields map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&model.Payment{}).Where("id = ?", id).Updates(fields).Error
}

type CouponRepo struct{ db *gorm.DB }

func NewCouponRepo(db *gorm.DB) *CouponRepo { return &CouponRepo{db: db} }

func (r *CouponRepo) List(ctx context.Context) ([]model.Coupon, error) {
	var rows []model.Coupon
	now := time.Now()
	err := TenantDB(ctx, r.db).
		Where("status = 1").
		Where("receive_start_at IS NULL OR receive_start_at <= ?", now).
		Where("receive_end_at IS NULL OR receive_end_at >= ?", now).
		Where("receive_limit_type = ? OR remain_count > 0", model.CouponReceiveLimitUnlimited).
		Order("id DESC").Find(&rows).Error
	return rows, err
}

func (r *CouponRepo) FindByID(ctx context.Context, id uint64) (*model.Coupon, error) {
	var c model.Coupon
	if err := TenantDB(ctx, r.db).First(&c, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func (r *CouponRepo) FindByIDForUpdate(ctx context.Context, tx *gorm.DB, id uint64) (*model.Coupon, error) {
	var c model.Coupon
	if err := tx.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND tenant_id = ?", id, EnsureTenant(ctx)).First(&c).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func (r *CouponRepo) Create(ctx context.Context, c *model.Coupon) error {
	c.TenantID = EnsureTenant(ctx)
	return r.db.WithContext(ctx).Create(c).Error
}

// ListAll 租户后台用：返回当前租户全部优惠券（包含已停用）。
func (r *CouponRepo) ListAll(ctx context.Context) ([]model.Coupon, error) {
	var rows []model.Coupon
	err := TenantDB(ctx, r.db).Order("id DESC").Find(&rows).Error
	return rows, err
}

// Update 更新除 remain_count 以外的字段。total_count 变更时会同步刷新 remain_count。
func (r *CouponRepo) Update(ctx context.Context, id uint64, fields map[string]interface{}) error {
	return TenantDB(ctx, r.db).Model(&model.Coupon{}).
		Where("id = ?", id).Updates(fields).Error
}

func (r *CouponRepo) Delete(ctx context.Context, id uint64) error {
	return TenantDB(ctx, r.db).Where("id = ?", id).Delete(&model.Coupon{}).Error
}

func (r *CouponRepo) DecreaseRemain(ctx context.Context, tx *gorm.DB, id uint64) error {
	res := tx.Model(&model.Coupon{}).
		Where("id = ? AND tenant_id = ? AND remain_count > 0", id, EnsureTenant(ctx)).
		UpdateColumn("remain_count", gorm.Expr("remain_count - 1"))
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("coupon out of stock")
	}
	return nil
}

type MemberCouponRepo struct{ db *gorm.DB }

func NewMemberCouponRepo(db *gorm.DB) *MemberCouponRepo { return &MemberCouponRepo{db: db} }

func (r *MemberCouponRepo) Create(ctx context.Context, mc *model.MemberCoupon) error {
	mc.TenantID = EnsureTenant(ctx)
	return r.db.WithContext(ctx).Create(mc).Error
}

func (r *MemberCouponRepo) ListByMember(ctx context.Context, memberID uint64, status string) ([]model.MemberCoupon, error) {
	tx := TenantDB(ctx, r.db).Where("member_id = ?", memberID)
	if status != "" {
		tx = tx.Where("status = ?", status)
	}
	var rows []model.MemberCoupon
	err := tx.Order("id DESC").Find(&rows).Error
	return rows, err
}

func (r *MemberCouponRepo) UpdateFields(ctx context.Context, memberID, id uint64, fields map[string]interface{}) error {
	res := TenantDB(ctx, r.db).Model(&model.MemberCoupon{}).
		Where("id = ? AND member_id = ?", id, memberID).
		Updates(fields)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrCouponUnavailable
	}
	return nil
}

func (r *MemberCouponRepo) CountByMemberCoupon(ctx context.Context, memberID, couponID uint64) (int64, error) {
	return r.CountByMemberCouponTx(ctx, r.db, memberID, couponID)
}

func (r *MemberCouponRepo) CountByMemberCouponTx(ctx context.Context, tx *gorm.DB, memberID, couponID uint64) (int64, error) {
	var total int64
	err := tx.WithContext(ctx).Model(&model.MemberCoupon{}).
		Where("tenant_id = ? AND member_id = ? AND coupon_id = ?", EnsureTenant(ctx), memberID, couponID).
		Count(&total).Error
	return total, err
}

func (r *MemberCouponRepo) CountUsedByMemberCoupon(ctx context.Context, memberID, couponID uint64) (int64, error) {
	return r.CountUsedByMemberCouponTx(ctx, r.db, memberID, couponID)
}

func (r *MemberCouponRepo) CountUsedByMemberCouponTx(ctx context.Context, tx *gorm.DB, memberID, couponID uint64) (int64, error) {
	var total int64
	err := tx.WithContext(ctx).Model(&model.MemberCoupon{}).
		Where("tenant_id = ? AND member_id = ? AND coupon_id = ? AND status = ?", EnsureTenant(ctx), memberID, couponID, "used").
		Count(&total).Error
	return total, err
}

func (r *MemberCouponRepo) FindByIDForMember(ctx context.Context, memberID, id uint64) (*model.MemberCoupon, error) {
	var row model.MemberCoupon
	if err := TenantDB(ctx, r.db).Where("id = ? AND member_id = ?", id, memberID).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *MemberCouponRepo) MarkUsed(ctx context.Context, tx *gorm.DB, id, memberID, orderID uint64) error {
	var memberCoupon model.MemberCoupon
	if err := tx.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND tenant_id = ? AND member_id = ?", id, EnsureTenant(ctx), memberID).
		First(&memberCoupon).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCouponUnavailable
		}
		return err
	}
	if memberCoupon.Status != "unused" {
		return ErrCouponAlreadyUsed
	}
	var coupon model.Coupon
	if err := tx.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND tenant_id = ?", memberCoupon.CouponID, EnsureTenant(ctx)).
		First(&coupon).Error; err != nil {
		return err
	}
	if coupon.Status != 1 {
		return ErrCouponUnavailable
	}
	if memberCoupon.UseLimit > 0 {
		usedCount, err := r.CountUsedByMemberCouponTx(ctx, tx, memberID, memberCoupon.CouponID)
		if err != nil {
			return err
		}
		if usedCount >= int64(memberCoupon.UseLimit) {
			return ErrCouponUseLimitReached
		}
	}
	res := tx.Model(&model.MemberCoupon{}).
		Where("id = ? AND tenant_id = ? AND member_id = ? AND status = ?", id, EnsureTenant(ctx), memberID, "unused").
		Updates(map[string]interface{}{
			"status":        "used",
			"used_at":       time.Now(),
			"used_order_id": orderID,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrCouponAlreadyUsed
	}
	return nil
}
