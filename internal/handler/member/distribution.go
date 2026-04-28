package member

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"wechat-mall-saas/internal/model"
	"wechat-mall-saas/internal/pkg/ctxkeys"
	apperr "wechat-mall-saas/internal/pkg/errors"
	"wechat-mall-saas/internal/pkg/response"
	"wechat-mall-saas/internal/repository"
)

type DistributionHandler struct {
	distributions *repository.DistributionRepo
	members       *repository.MemberRepo
}

func NewDistributionHandler(distributions *repository.DistributionRepo, members *repository.MemberRepo) *DistributionHandler {
	return &DistributionHandler{distributions: distributions, members: members}
}

func (handler *DistributionHandler) Overview(ginContext *gin.Context) {
	memberInfo := ctxkeys.GetMember(ginContext.Request.Context())
	if memberInfo == nil {
		response.Fail(ginContext, apperr.ErrUnauthorized)
		return
	}
	currentMember, err := handler.members.FindByID(ginContext.Request.Context(), memberInfo.ID)
	if err != nil {
		response.Fail(ginContext, err)
		return
	}
	if currentMember == nil {
		response.Fail(ginContext, apperr.ErrUnauthorized)
		return
	}
	settings, err := handler.distributionSettings(ginContext)
	if err != nil {
		response.Fail(ginContext, err)
		return
	}
	distributor, err := handler.distributions.FindDistributorByMemberID(ginContext.Request.Context(), memberInfo.ID)
	if err != nil {
		response.Fail(ginContext, err)
		return
	}
	response.OK(ginContext, gin.H{
		"settings":        settings,
		"distributor":     distributor,
		"can_apply":       settings.Enabled == 1 && distributor == nil,
		"invite_code":     strconv.FormatUint(memberInfo.ID, 10),
		"bound_parent_id": currentMember.ParentID,
	})
}

func (handler *DistributionHandler) Apply(ginContext *gin.Context) {
	memberInfo := ctxkeys.GetMember(ginContext.Request.Context())
	if memberInfo == nil {
		response.Fail(ginContext, apperr.ErrUnauthorized)
		return
	}
	settings, err := handler.distributionSettings(ginContext)
	if err != nil {
		response.Fail(ginContext, err)
		return
	}
	if settings.Enabled != 1 {
		response.Fail(ginContext, apperr.New(30004, "分销功能未启用"))
		return
	}
	existing, err := handler.distributions.FindDistributorByMemberID(ginContext.Request.Context(), memberInfo.ID)
	if err != nil {
		response.Fail(ginContext, err)
		return
	}
	if existing != nil {
		response.OK(ginContext, existing)
		return
	}
	currentMember, err := handler.members.FindByID(ginContext.Request.Context(), memberInfo.ID)
	if err != nil {
		response.Fail(ginContext, err)
		return
	}
	if currentMember == nil {
		response.Fail(ginContext, apperr.ErrUnauthorized)
		return
	}
	status := int8(0)
	var approvedAt *time.Time
	if settings.AutoBecome == 1 {
		now := time.Now()
		status = 1
		approvedAt = &now
	}
	grandparentID := uint64(0)
	if currentMember.ParentID > 0 {
		parentDistributor, err := handler.distributions.FindDistributorByMemberID(ginContext.Request.Context(), currentMember.ParentID)
		if err != nil {
			response.Fail(ginContext, err)
			return
		}
		if parentDistributor != nil && parentDistributor.Status == 1 {
			grandparentID = parentDistributor.ParentID
		}
	}
	distributor := &model.Distributor{
		MemberID:      memberInfo.ID,
		ParentID:      currentMember.ParentID,
		GrandparentID: grandparentID,
		Status:        status,
		ApprovedAt:    approvedAt,
	}
	if err := handler.distributions.CreateDistributor(ginContext.Request.Context(), distributor); err != nil {
		response.Fail(ginContext, err)
		return
	}
	response.OK(ginContext, distributor)
}

type bindDistributionParentReq struct {
	InviterMemberID uint64 `json:"inviter_member_id" binding:"required"`
}

func (handler *DistributionHandler) BindParent(ginContext *gin.Context) {
	memberInfo := ctxkeys.GetMember(ginContext.Request.Context())
	if memberInfo == nil {
		response.Fail(ginContext, apperr.ErrUnauthorized)
		return
	}
	var req bindDistributionParentReq
	if err := ginContext.ShouldBindJSON(&req); err != nil {
		response.FailCode(ginContext, 20001, err.Error())
		return
	}
	if req.InviterMemberID == 0 || req.InviterMemberID == memberInfo.ID {
		response.Fail(ginContext, apperr.ErrParamInvalid)
		return
	}
	currentMember, err := handler.members.FindByID(ginContext.Request.Context(), memberInfo.ID)
	if err != nil {
		response.Fail(ginContext, err)
		return
	}
	if currentMember == nil {
		response.Fail(ginContext, apperr.ErrUnauthorized)
		return
	}
	if currentMember.ParentID > 0 {
		response.OK(ginContext, gin.H{"bound_parent_id": currentMember.ParentID})
		return
	}
	inviter, err := handler.members.FindByID(ginContext.Request.Context(), req.InviterMemberID)
	if err != nil {
		response.Fail(ginContext, err)
		return
	}
	if inviter == nil || inviter.Status != 1 {
		response.Fail(ginContext, apperr.ErrNotFound)
		return
	}
	inviterDistributor, err := handler.distributions.FindDistributorByMemberID(ginContext.Request.Context(), req.InviterMemberID)
	if err != nil {
		response.Fail(ginContext, err)
		return
	}
	if inviterDistributor == nil || inviterDistributor.Status != 1 {
		response.Fail(ginContext, apperr.New(20001, "邀请人不是有效分销员"))
		return
	}
	if err := handler.members.UpdateFields(ginContext.Request.Context(), memberInfo.ID, map[string]interface{}{
		"parent_id":  req.InviterMemberID,
		"updated_at": time.Now(),
	}); err != nil {
		response.Fail(ginContext, err)
		return
	}
	if err := handler.distributions.IncrementInviteCountByMemberID(ginContext.Request.Context(), req.InviterMemberID); err != nil {
		response.Fail(ginContext, err)
		return
	}
	response.OK(ginContext, gin.H{"bound_parent_id": req.InviterMemberID})
}

func (handler *DistributionHandler) Commissions(ginContext *gin.Context) {
	memberInfo := ctxkeys.GetMember(ginContext.Request.Context())
	if memberInfo == nil {
		response.Fail(ginContext, apperr.ErrUnauthorized)
		return
	}
	page, _ := strconv.Atoi(ginContext.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(ginContext.DefaultQuery("size", "20"))
	distributor, err := handler.distributions.FindDistributorByMemberID(ginContext.Request.Context(), memberInfo.ID)
	if err != nil {
		response.Fail(ginContext, err)
		return
	}
	if distributor == nil {
		response.OK(ginContext, gin.H{"list": []model.CommissionLog{}, "total": 0, "page": page, "size": size})
		return
	}
	rows, total, err := handler.distributions.ListCommissions(ginContext.Request.Context(), distributor.ID, page, size)
	if err != nil {
		response.Fail(ginContext, err)
		return
	}
	response.OK(ginContext, gin.H{"list": rows, "total": total, "page": page, "size": size})
}

func (handler *DistributionHandler) distributionSettings(ginContext *gin.Context) (*model.DistributionSetting, error) {
	settings, err := handler.distributions.GetSettings(ginContext.Request.Context())
	if err != nil {
		return nil, err
	}
	if settings != nil {
		return settings, nil
	}
	tenantInfo := ctxkeys.GetTenant(ginContext.Request.Context())
	tenantID := uint64(0)
	if tenantInfo != nil {
		tenantID = tenantInfo.ID
	}
	return &model.DistributionSetting{
		TenantID:    tenantID,
		Enabled:     1,
		Level1Rate:  decimal.NewFromFloat(0.10),
		Level2Rate:  decimal.NewFromFloat(0.05),
		MinWithdraw: decimal.NewFromInt(10),
		AutoBecome:  0,
	}, nil
}
