package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"wechat-mall-saas/internal/model"
)

type PaymentRepo struct{ db *gorm.DB }

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
	err := TenantDB(ctx, r.db).Where("status = 1").Order("id DESC").Find(&rows).Error
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
