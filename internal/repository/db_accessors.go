package repository

import "gorm.io/gorm"

// DB 暴露底层 *gorm.DB 以便 Service 层开启事务或执行原生语句
func (r *CouponRepo) DB() *gorm.DB       { return r.db }
func (r *MemberCouponRepo) DB() *gorm.DB { return r.db }
func (r *ProductRepo) DB() *gorm.DB      { return r.db }
func (r *ProductSKURepo) DB() *gorm.DB   { return r.db }
func (r *MemberRepo) DB() *gorm.DB       { return r.db }
func (r *PaymentRepo) DB() *gorm.DB      { return r.db }
