package service

import (
	"context"

	"wechat-mall-saas/internal/model"
	apperr "wechat-mall-saas/internal/pkg/errors"
	"wechat-mall-saas/internal/repository"
)

type MemberService struct {
	members   *repository.MemberRepo
	addresses *repository.MemberAddressRepo
	points    *repository.PointsLogRepo
}

func NewMemberService(m *repository.MemberRepo, a *repository.MemberAddressRepo, p *repository.PointsLogRepo) *MemberService {
	return &MemberService{members: m, addresses: a, points: p}
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
