package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"wechat-mall-saas/internal/model"
	apperr "wechat-mall-saas/internal/pkg/errors"
	"wechat-mall-saas/internal/pkg/utils"
	"wechat-mall-saas/internal/repository"
)

type AfterSaleService struct {
	afterSales *repository.AfterSaleRepo
	reasons    *repository.AfterSaleReasonRepo
}

func NewAfterSaleService(afterSales *repository.AfterSaleRepo, reasons *repository.AfterSaleReasonRepo) *AfterSaleService {
	return &AfterSaleService{afterSales: afterSales, reasons: reasons}
}

type AfterSaleApplyInput struct {
	Type        string          `json:"type"`
	Reason      string          `json:"reason"`
	Description string          `json:"description"`
	Amount      decimal.Decimal `json:"amount"`
	Images      []string        `json:"images"`
}

type AfterSaleReturnInput struct {
	ReturnExpressCompany string `json:"return_express_company"`
	ReturnExpressNo      string `json:"return_express_no"`
}

func (svc *AfterSaleService) Apply(ctx context.Context, memberID, orderID uint64, input AfterSaleApplyInput) (*model.AfterSaleOrder, error) {
	tenantID := repository.EnsureTenant(ctx)
	if tenantID == 0 {
		return nil, apperr.ErrTenantRequired
	}
	if memberID == 0 {
		return nil, apperr.ErrUnauthorized
	}
	reason := strings.TrimSpace(input.Reason)
	if reason == "" {
		return nil, apperr.New(20001, "售后原因必填")
	}
	input.Description = strings.TrimSpace(input.Description)

	var created model.AfterSaleOrder
	err := svc.afterSales.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var order model.Order
		if err := tx.Where("tenant_id = ? AND id = ? AND member_id = ?", tenantID, orderID, memberID).First(&order).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return apperr.ErrNotFound
			}
			return err
		}
		if !containsStatus(afterSaleAllowedOrderStatuses(), order.Status) {
			return apperr.New(30010, "当前订单状态不可申请售后")
		}
		var activeCount int64
		if err := tx.Model(&model.AfterSaleOrder{}).
			Where("tenant_id = ? AND order_id = ? AND status NOT IN ?", tenantID, order.ID, afterSaleFinalStatuses()).
			Count(&activeCount).Error; err != nil {
			return err
		}
		if activeCount > 0 {
			return apperr.New(30011, "订单已有处理中售后")
		}

		afterSaleType := input.Type
		if afterSaleType == "" {
			afterSaleType = defaultAfterSaleType(order.Status)
		}
		if !isAfterSaleTypeAllowed(order.Status, afterSaleType) {
			return apperr.New(30010, "当前订单状态不支持该售后类型")
		}
		validReason, err := svc.isReasonEnabled(tx, reason, afterSaleType)
		if err != nil {
			return err
		}
		if !validReason {
			return apperr.New(30013, "请选择有效售后原因")
		}
		amount := input.Amount
		if amount.IsZero() {
			amount = order.ActualAmount
		}
		if !amount.Equal(order.ActualAmount) {
			return apperr.New(30012, "首版仅支持整单全额售后")
		}
		now := time.Now()
		created = model.AfterSaleOrder{
			TenantID:          tenantID,
			AfterSaleNo:       utils.OrderNo("AS"),
			OrderID:           order.ID,
			OrderNo:           order.OrderNo,
			MemberID:          memberID,
			Type:              afterSaleType,
			Status:            model.AfterSaleStatusPending,
			Amount:            amount,
			Reason:            reason,
			Description:       input.Description,
			Images:            model.JSONB(input.Images),
			OrderStatusBefore: order.Status,
			AppliedAt:         now,
		}
		if err := tx.Create(&created).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.Order{}).
			Where("tenant_id = ? AND id = ? AND status = ?", tenantID, order.ID, order.Status).
			Update("status", model.OrderStatusRefunding).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.OrderItem{}).
			Where("tenant_id = ? AND order_id = ?", tenantID, order.ID).
			Update("refund_status", "refunding").Error; err != nil {
			return err
		}
		return createAfterSaleLogAndMessage(tx, tenantID, order.ID, order.OrderNo, "member", memberID, "after_sale_apply", order.Status, model.OrderStatusRefunding, reason, "收到售后申请", fmt.Sprintf("订单 %s 已提交售后申请，原因：%s。", order.OrderNo, reason))
	})
	if err != nil {
		return nil, err
	}
	return &created, nil
}

func (svc *AfterSaleService) ListReasons(ctx context.Context, reasonType string) ([]model.AfterSaleReason, error) {
	if repository.EnsureTenant(ctx) == 0 {
		return nil, apperr.ErrTenantRequired
	}
	reasonType = normalizeAfterSaleReasonType(reasonType)
	return svc.reasons.ListEnabled(ctx, reasonType)
}

func (svc *AfterSaleService) ListForMember(ctx context.Context, memberID uint64, query repository.AfterSaleListQuery) ([]model.AfterSaleOrder, int64, error) {
	query.MemberID = memberID
	normalizeAfterSaleListQuery(&query)
	return svc.afterSales.List(ctx, query)
}

func (svc *AfterSaleService) ListForAdmin(ctx context.Context, query repository.AfterSaleListQuery) ([]model.AfterSaleOrder, int64, error) {
	normalizeAfterSaleListQuery(&query)
	return svc.afterSales.List(ctx, query)
}

func (svc *AfterSaleService) DetailForMember(ctx context.Context, id, memberID uint64) (*model.AfterSaleOrder, error) {
	row, err := svc.afterSales.FindByID(ctx, id)
	if err != nil || row == nil {
		return row, err
	}
	if row.MemberID != memberID {
		return nil, apperr.ErrForbidden
	}
	return row, nil
}

func (svc *AfterSaleService) Detail(ctx context.Context, id uint64) (*model.AfterSaleOrder, error) {
	return svc.afterSales.FindByID(ctx, id)
}

func (svc *AfterSaleService) Cancel(ctx context.Context, id, memberID uint64) error {
	return svc.finishWithRestore(ctx, id, memberID, "member", memberID, model.AfterSaleStatusCancelled, "after_sale_cancel", "售后已取消", "买家已取消售后申请。", "")
}

func (svc *AfterSaleService) Approve(ctx context.Context, id, adminID uint64, remark string) error {
	now := time.Now()
	return svc.transitionAfterSale(ctx, id, []string{model.AfterSaleStatusPending}, model.AfterSaleStatusApproved, map[string]interface{}{
		"audit_remark": remark,
		"audited_at":   now,
	}, "admin", adminID, "after_sale_approve", remark, "售后申请已通过", "商家已同意售后申请。")
}

func (svc *AfterSaleService) Reject(ctx context.Context, id, adminID uint64, remark string) error {
	return svc.finishWithRestore(ctx, id, 0, "admin", adminID, model.AfterSaleStatusRejected, "after_sale_reject", "售后申请已驳回", "商家已驳回售后申请。", remark)
}

func (svc *AfterSaleService) SubmitReturn(ctx context.Context, id, memberID uint64, input AfterSaleReturnInput) error {
	if input.ReturnExpressCompany == "" || input.ReturnExpressNo == "" {
		return apperr.New(20001, "退货物流公司和单号必填")
	}
	now := time.Now()
	return svc.transitionAfterSaleForMember(ctx, id, memberID, []string{model.AfterSaleStatusApproved}, model.AfterSaleStatusReturning, map[string]interface{}{
		"return_express_company": input.ReturnExpressCompany,
		"return_express_no":      input.ReturnExpressNo,
		"returned_at":            now,
	}, "after_sale_return", input.ReturnExpressCompany+" "+input.ReturnExpressNo, "买家已寄回商品", fmt.Sprintf("买家已提交退货物流：%s %s。", input.ReturnExpressCompany, input.ReturnExpressNo))
}

func (svc *AfterSaleService) Receive(ctx context.Context, id, adminID uint64, remark string) error {
	now := time.Now()
	return svc.transitionAfterSale(ctx, id, []string{model.AfterSaleStatusReturning}, model.AfterSaleStatusReceived, map[string]interface{}{
		"audit_remark": remark,
		"received_at":  now,
	}, "admin", adminID, "after_sale_receive", remark, "商家已收到退货", "商家已确认收到退货商品。")
}

func (svc *AfterSaleService) Refund(ctx context.Context, id, adminID uint64, remark string) error {
	tenantID := repository.EnsureTenant(ctx)
	now := time.Now()
	return svc.afterSales.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var row model.AfterSaleOrder
		if err := tx.Where("tenant_id = ? AND id = ?", tenantID, id).First(&row).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return apperr.ErrNotFound
			}
			return err
		}
		if !canMarkAfterSaleRefunded(row) {
			return apperr.New(30010, "当前售后状态不可退款完成")
		}
		if err := tx.Model(&model.AfterSaleOrder{}).Where("tenant_id = ? AND id = ? AND status = ?", tenantID, row.ID, row.Status).Updates(map[string]interface{}{
			"status":        model.AfterSaleStatusRefunded,
			"refund_remark": remark,
			"refund_no":     utils.OrderNo("R"),
			"refunded_at":   now,
		}).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.Order{}).Where("tenant_id = ? AND id = ?", tenantID, row.OrderID).Update("status", model.OrderStatusRefunded).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.OrderItem{}).Where("tenant_id = ? AND order_id = ?", tenantID, row.OrderID).Updates(map[string]interface{}{
			"refund_status": "refunded",
			"refund_amount": gorm.Expr("item_total"),
		}).Error; err != nil {
			return err
		}
		return createAfterSaleLogAndMessage(tx, tenantID, row.OrderID, row.OrderNo, "admin", adminID, "after_sale_refund", model.OrderStatusRefunding, model.OrderStatusRefunded, remark, "售后退款已完成", "商家已标记售后退款完成。")
	})
}

func (svc *AfterSaleService) finishWithRestore(ctx context.Context, id, memberID uint64, operatorType string, operatorID uint64, toStatus, action, title, body, remark string) error {
	tenantID := repository.EnsureTenant(ctx)
	now := time.Now()
	return svc.afterSales.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var row model.AfterSaleOrder
		query := tx.Where("tenant_id = ? AND id = ?", tenantID, id)
		if memberID > 0 {
			query = query.Where("member_id = ?", memberID)
		}
		if err := query.First(&row).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return apperr.ErrNotFound
			}
			return err
		}
		if !containsStatus([]string{model.AfterSaleStatusPending, model.AfterSaleStatusApproved}, row.Status) {
			return apperr.New(30010, "当前售后状态不可操作")
		}
		updates := map[string]interface{}{"status": toStatus, "cancelled_at": now}
		if toStatus == model.AfterSaleStatusRejected {
			updates = map[string]interface{}{"status": toStatus, "audit_remark": remark, "audited_at": now}
		}
		if err := tx.Model(&model.AfterSaleOrder{}).Where("tenant_id = ? AND id = ? AND status = ?", tenantID, row.ID, row.Status).Updates(updates).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.Order{}).Where("tenant_id = ? AND id = ?", tenantID, row.OrderID).Update("status", row.OrderStatusBefore).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.OrderItem{}).Where("tenant_id = ? AND order_id = ?", tenantID, row.OrderID).Updates(map[string]interface{}{"refund_status": "none", "refund_amount": nil}).Error; err != nil {
			return err
		}
		return createAfterSaleLogAndMessage(tx, tenantID, row.OrderID, row.OrderNo, operatorType, operatorID, action, model.OrderStatusRefunding, row.OrderStatusBefore, remark, title, body)
	})
}

func (svc *AfterSaleService) transitionAfterSaleForMember(ctx context.Context, id, memberID uint64, from []string, toStatus string, fields map[string]interface{}, action, remark, title, body string) error {
	return svc.transitionAfterSaleScoped(ctx, id, memberID, from, toStatus, fields, "member", memberID, action, remark, title, body)
}

func (svc *AfterSaleService) transitionAfterSale(ctx context.Context, id uint64, from []string, toStatus string, fields map[string]interface{}, operatorType string, operatorID uint64, action, remark, title, body string) error {
	return svc.transitionAfterSaleScoped(ctx, id, 0, from, toStatus, fields, operatorType, operatorID, action, remark, title, body)
}

func (svc *AfterSaleService) transitionAfterSaleScoped(ctx context.Context, id, memberID uint64, from []string, toStatus string, fields map[string]interface{}, operatorType string, operatorID uint64, action, remark, title, body string) error {
	tenantID := repository.EnsureTenant(ctx)
	return svc.afterSales.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var row model.AfterSaleOrder
		query := tx.Where("tenant_id = ? AND id = ?", tenantID, id)
		if memberID > 0 {
			query = query.Where("member_id = ?", memberID)
		}
		if err := query.First(&row).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return apperr.ErrNotFound
			}
			return err
		}
		if !containsStatus(from, row.Status) {
			return apperr.New(30010, "当前售后状态不可操作")
		}
		if row.Type == model.AfterSaleTypeRefund && toStatus == model.AfterSaleStatusReturning {
			return apperr.New(30010, "仅退款售后无需退货物流")
		}
		updates := map[string]interface{}{"status": toStatus}
		for key, value := range fields {
			updates[key] = value
		}
		if err := tx.Model(&model.AfterSaleOrder{}).Where("tenant_id = ? AND id = ? AND status = ?", tenantID, row.ID, row.Status).Updates(updates).Error; err != nil {
			return err
		}
		return createAfterSaleLogAndMessage(tx, tenantID, row.OrderID, row.OrderNo, operatorType, operatorID, action, model.OrderStatusRefunding, model.OrderStatusRefunding, remark, title, body)
	})
}

func normalizeAfterSaleListQuery(query *repository.AfterSaleListQuery) {
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.Size <= 0 || query.Size > 100 {
		query.Size = 20
	}
}

func afterSaleAllowedOrderStatuses() []string {
	return []string{model.OrderStatusPaid, model.OrderStatusPreparing, model.OrderStatusShipped, model.OrderStatusDelivered, model.OrderStatusCompleted}
}

func afterSaleFinalStatuses() []string {
	return []string{model.AfterSaleStatusRejected, model.AfterSaleStatusRefunded, model.AfterSaleStatusCancelled}
}

func defaultAfterSaleType(orderStatus string) string {
	if containsStatus([]string{model.OrderStatusPaid, model.OrderStatusPreparing}, orderStatus) {
		return model.AfterSaleTypeRefund
	}
	return model.AfterSaleTypeReturnRefund
}

func isAfterSaleTypeAllowed(orderStatus, afterSaleType string) bool {
	switch afterSaleType {
	case model.AfterSaleTypeRefund:
		return containsStatus([]string{model.OrderStatusPaid, model.OrderStatusPreparing}, orderStatus)
	case model.AfterSaleTypeReturnRefund:
		return containsStatus([]string{model.OrderStatusShipped, model.OrderStatusDelivered, model.OrderStatusCompleted}, orderStatus)
	default:
		return false
	}
}

func canMarkAfterSaleRefunded(row model.AfterSaleOrder) bool {
	if row.Type == model.AfterSaleTypeRefund {
		return row.Status == model.AfterSaleStatusApproved
	}
	return row.Type == model.AfterSaleTypeReturnRefund && row.Status == model.AfterSaleStatusReceived
}

func (svc *AfterSaleService) isReasonEnabled(tx *gorm.DB, reason, reasonType string) (bool, error) {
	var count int64
	err := tx.Model(&model.AfterSaleReason{}).
		Where("enabled = 1 AND label = ? AND type IN ?", reason, []string{model.AfterSaleReasonTypeAll, reasonType}).
		Count(&count).Error
	return count > 0, err
}

func normalizeAfterSaleReasonType(reasonType string) string {
	switch reasonType {
	case model.AfterSaleReasonTypeRefund, model.AfterSaleReasonTypeReturnRefund:
		return reasonType
	default:
		return ""
	}
}

func createAfterSaleLogAndMessage(tx *gorm.DB, tenantID, orderID uint64, orderNo, operatorType string, operatorID uint64, action, beforeStatus, afterStatus, remark, title, body string) error {
	if err := tx.Create(&model.OrderLog{
		TenantID:     tenantID,
		OrderID:      orderID,
		OperatorType: operatorType,
		OperatorID:   operatorID,
		Action:       action,
		BeforeStatus: beforeStatus,
		AfterStatus:  afterStatus,
		Remark:       remark,
	}).Error; err != nil {
		return err
	}
	if title == "" {
		return nil
	}
	return tx.Create(&model.OrderMessage{
		TenantID:  tenantID,
		OrderID:   orderID,
		OrderNo:   orderNo,
		EventType: action,
		Title:     title,
		Content:   body,
		Status:    model.OrderMessageStatusUnread,
	}).Error
}
