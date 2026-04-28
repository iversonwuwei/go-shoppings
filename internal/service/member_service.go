package service

import (
	"context"
	"time"

	"wechat-mall-saas/internal/model"
	apperr "wechat-mall-saas/internal/pkg/errors"
	"wechat-mall-saas/internal/repository"
)

type MemberService struct {
	members   *repository.MemberRepo
	addresses *repository.MemberAddressRepo
	points    *repository.PointsLogRepo
	levels    *repository.MemberLevelRepo
}

func NewMemberService(m *repository.MemberRepo, a *repository.MemberAddressRepo, p *repository.PointsLogRepo, l *repository.MemberLevelRepo) *MemberService {
	return &MemberService{members: m, addresses: a, points: p, levels: l}
}

type AdminMemberItem struct {
	ID            uint64     `json:"id"`
	TenantID      uint64     `json:"tenant_id"`
	OpenID        string     `json:"openid"`
	UnionID       string     `json:"unionid,omitempty"`
	Nickname      string     `json:"nickname"`
	Avatar        string     `json:"avatar"`
	Gender        int8       `json:"gender"`
	Birthday      *time.Time `json:"birthday,omitempty"`
	Phone         string     `json:"phone"`
	LevelID       *uint64    `json:"level_id,omitempty"`
	LevelName     string     `json:"level_name,omitempty"`
	LevelColor    string     `json:"level_color,omitempty"`
	LevelExpireAt *time.Time `json:"level_expire_at,omitempty"`
	Points        int        `json:"points"`
	GrowthValue   int        `json:"growth_value"`
	ParentID      uint64     `json:"parent_id"`
	Level1Count   int        `json:"level1_count"`
	Level2Count   int        `json:"level2_count"`
	Status        int8       `json:"status"`
	LastLoginAt   *time.Time `json:"last_login_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type AdminMemberDetail struct {
	Member     AdminMemberItem       `json:"member"`
	Addresses  []model.MemberAddress `json:"addresses"`
	PointsLogs []model.PointsLog     `json:"points_logs"`
}

func (s *MemberService) Profile(ctx context.Context, id uint64) (*model.Member, error) {
	m, err := s.members.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, apperr.ErrNotFound
	}
	return m, nil
}

func (s *MemberService) UpdateProfile(ctx context.Context, id uint64, fields map[string]interface{}) error {
	allowed := map[string]bool{"nickname": true, "avatar": true, "gender": true, "birthday": true}
	clean := map[string]interface{}{}
	for k, v := range fields {
		if allowed[k] {
			clean[k] = v
		}
	}
	return s.members.UpdateFields(ctx, id, clean)
}

func (s *MemberService) Addresses(ctx context.Context, memberID uint64) ([]model.MemberAddress, error) {
	return s.addresses.ListByMember(ctx, memberID)
}

func (s *MemberService) CreateAddress(ctx context.Context, a *model.MemberAddress) error {
	return s.addresses.Create(ctx, a)
}

func (s *MemberService) PointsLogs(ctx context.Context, memberID uint64, page, size int) ([]model.PointsLog, int64, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 100 {
		size = 20
	}
	return s.points.ListByMember(ctx, memberID, page, size)
}

func (s *MemberService) AdminList(ctx context.Context, q repository.MemberListQuery) ([]AdminMemberItem, int64, error) {
	if repository.EnsureTenant(ctx) == 0 {
		return nil, 0, apperr.ErrTenantRequired
	}
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.Size <= 0 || q.Size > 100 {
		q.Size = 20
	}
	rows, total, err := s.members.List(ctx, q)
	if err != nil {
		return nil, 0, err
	}
	levelMap, err := s.memberLevelMap(ctx)
	if err != nil {
		return nil, 0, err
	}
	out := make([]AdminMemberItem, 0, len(rows))
	for _, row := range rows {
		out = append(out, adminMemberItem(row, levelMap))
	}
	return out, total, nil
}

func (s *MemberService) AdminDetail(ctx context.Context, id uint64) (*AdminMemberDetail, error) {
	m, err := s.members.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, apperr.ErrNotFound
	}
	levelMap, err := s.memberLevelMap(ctx)
	if err != nil {
		return nil, err
	}
	addresses, err := s.addresses.ListByMember(ctx, id)
	if err != nil {
		return nil, err
	}
	points, _, err := s.points.ListByMember(ctx, id, 1, 20)
	if err != nil {
		return nil, err
	}
	return &AdminMemberDetail{
		Member:     adminMemberItem(*m, levelMap),
		Addresses:  addresses,
		PointsLogs: points,
	}, nil
}

func (s *MemberService) UpdateMemberStatus(ctx context.Context, id uint64, status int8) error {
	if status != 0 && status != 1 {
		return apperr.ErrParamInvalid
	}
	m, err := s.members.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if m == nil {
		return apperr.ErrNotFound
	}
	return s.members.UpdateFields(ctx, id, map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	})
}

func (s *MemberService) UpdateMemberLevel(ctx context.Context, id uint64, levelID *uint64, expireAt *time.Time) error {
	m, err := s.members.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if m == nil {
		return apperr.ErrNotFound
	}
	fields := map[string]interface{}{
		"level_id":        nil,
		"level_expire_at": nil,
		"updated_at":      time.Now(),
	}
	if levelID != nil && *levelID > 0 {
		level, err := s.levels.FindByID(ctx, *levelID)
		if err != nil {
			return err
		}
		if level == nil {
			return apperr.ErrNotFound
		}
		fields["level_id"] = *levelID
		fields["level_expire_at"] = expireAt
	}
	return s.members.UpdateFields(ctx, id, fields)
}

func (s *MemberService) IsActive(ctx context.Context, id uint64) (bool, error) {
	m, err := s.members.FindByID(ctx, id)
	if err != nil {
		return false, err
	}
	return m != nil && m.Status == 1, nil
}

func (s *MemberService) memberLevelMap(ctx context.Context) (map[uint64]model.MemberLevel, error) {
	out := map[uint64]model.MemberLevel{}
	if s.levels == nil {
		return out, nil
	}
	levels, err := s.levels.List(ctx)
	if err != nil {
		return nil, err
	}
	for _, level := range levels {
		out[level.ID] = level
	}
	return out, nil
}

func adminMemberItem(m model.Member, levels map[uint64]model.MemberLevel) AdminMemberItem {
	item := AdminMemberItem{
		ID:            m.ID,
		TenantID:      m.TenantID,
		OpenID:        m.OpenID,
		UnionID:       m.UnionID,
		Nickname:      m.Nickname,
		Avatar:        m.Avatar,
		Gender:        m.Gender,
		Birthday:      m.Birthday,
		Phone:         m.Phone,
		LevelID:       m.LevelID,
		LevelExpireAt: m.LevelExpireAt,
		Points:        m.Points,
		GrowthValue:   m.GrowthValue,
		ParentID:      m.ParentID,
		Level1Count:   m.Level1Count,
		Level2Count:   m.Level2Count,
		Status:        m.Status,
		LastLoginAt:   m.LastLoginAt,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
	if m.LevelID != nil {
		if level, ok := levels[*m.LevelID]; ok {
			item.LevelName = level.Name
			item.LevelColor = level.Color
		}
	}
	return item
}
