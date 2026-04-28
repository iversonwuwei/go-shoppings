package repository

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"gorm.io/gorm"

	"wechat-mall-saas/internal/model"
)

type MemberRepo struct{ db *gorm.DB }

func NewMemberRepo(db *gorm.DB) *MemberRepo { return &MemberRepo{db: db} }

type MemberListQuery struct {
	Keyword string
	Status  *int8
	LevelID *uint64
	Page    int
	Size    int
}

func (r *MemberRepo) FindByOpenID(ctx context.Context, openID string) (*model.Member, error) {
	var m model.Member
	if err := TenantDB(ctx, r.db).Where("openid = ?", openID).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func (r *MemberRepo) FindByID(ctx context.Context, id uint64) (*model.Member, error) {
	var m model.Member
	if err := TenantDB(ctx, r.db).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func (r *MemberRepo) FindByPhone(ctx context.Context, phone string) (*model.Member, error) {
	var m model.Member
	if err := TenantDB(ctx, r.db).Where("phone = ?", phone).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func (r *MemberRepo) List(ctx context.Context, q MemberListQuery) ([]model.Member, int64, error) {
	tx := TenantDB(ctx, r.db).Model(&model.Member{})
	if q.Status != nil {
		tx = tx.Where("status = ?", *q.Status)
	}
	if q.LevelID != nil {
		tx = tx.Where("level_id = ?", *q.LevelID)
	}
	keyword := strings.TrimSpace(q.Keyword)
	if keyword != "" {
		lowerLike := "%" + strings.ToLower(keyword) + "%"
		like := "%" + keyword + "%"
		cond := "(LOWER(COALESCE(nickname, '')) LIKE ? OR phone LIKE ? OR openid LIKE ?"
		args := []interface{}{lowerLike, like, like}
		if id, err := strconv.ParseUint(keyword, 10, 64); err == nil && id > 0 {
			cond += " OR id = ?"
			args = append(args, id)
		}
		cond += ")"
		tx = tx.Where(cond, args...)
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.Member
	if err := tx.Order("id DESC").Offset((q.Page - 1) * q.Size).Limit(q.Size).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *MemberRepo) Create(ctx context.Context, m *model.Member) error {
	m.TenantID = EnsureTenant(ctx)
	if strings.TrimSpace(m.Phone) == "" {
		return r.db.WithContext(ctx).Omit("phone").Create(m).Error
	}
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *MemberRepo) Update(ctx context.Context, m *model.Member) error {
	return TenantDB(ctx, r.db).Save(m).Error
}

func (r *MemberRepo) UpdateFields(ctx context.Context, id uint64, fields map[string]interface{}) error {
	return TenantDB(ctx, r.db).Model(&model.Member{}).Where("id = ?", id).Updates(fields).Error
}

func (r *MemberRepo) Count(ctx context.Context) (int64, error) {
	var n int64
	err := TenantDB(ctx, r.db).Model(&model.Member{}).Count(&n).Error
	return n, err
}

type MemberAddressRepo struct{ db *gorm.DB }

func NewMemberAddressRepo(db *gorm.DB) *MemberAddressRepo { return &MemberAddressRepo{db: db} }

func (r *MemberAddressRepo) ListByMember(ctx context.Context, memberID uint64) ([]model.MemberAddress, error) {
	var rows []model.MemberAddress
	err := TenantDB(ctx, r.db).Where("member_id = ?", memberID).Order("is_default DESC, id DESC").Find(&rows).Error
	return rows, err
}

func (r *MemberAddressRepo) Create(ctx context.Context, a *model.MemberAddress) error {
	a.TenantID = EnsureTenant(ctx)
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *MemberAddressRepo) FindByID(ctx context.Context, id uint64) (*model.MemberAddress, error) {
	var a model.MemberAddress
	if err := TenantDB(ctx, r.db).First(&a, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}

type PointsLogRepo struct{ db *gorm.DB }

func NewPointsLogRepo(db *gorm.DB) *PointsLogRepo { return &PointsLogRepo{db: db} }

func (r *PointsLogRepo) Create(ctx context.Context, l *model.PointsLog) error {
	l.TenantID = EnsureTenant(ctx)
	return r.db.WithContext(ctx).Create(l).Error
}

func (r *PointsLogRepo) ListByMember(ctx context.Context, memberID uint64, page, size int) ([]model.PointsLog, int64, error) {
	tx := TenantDB(ctx, r.db).Model(&model.PointsLog{}).Where("member_id = ?", memberID)
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.PointsLog
	if err := tx.Order("id DESC").Offset((page - 1) * size).Limit(size).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}
