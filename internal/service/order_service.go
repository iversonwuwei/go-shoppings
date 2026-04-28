package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"wechat-mall-saas/internal/model"
	"wechat-mall-saas/internal/pkg/ctxkeys"
	apperr "wechat-mall-saas/internal/pkg/errors"
	"wechat-mall-saas/internal/pkg/utils"
	"wechat-mall-saas/internal/repository"
)

type OrderService struct {
	orders        *repository.OrderRepo
	logs          *repository.OrderLogRepo
	messages      *repository.OrderMessageRepo
	products      *repository.ProductRepo
	skus          *repository.ProductSKURepo
	coupons       *repository.CouponRepo
	memberCoupons *repository.MemberCouponRepo
	tenants       *TenantService
}

func NewOrderService(o *repository.OrderRepo, l *repository.OrderLogRepo, m *repository.OrderMessageRepo, p *repository.ProductRepo, s *repository.ProductSKURepo, c *repository.CouponRepo, mc *repository.MemberCouponRepo, t *TenantService) *OrderService {
	return &OrderService{orders: o, logs: l, messages: m, products: p, skus: s, coupons: c, memberCoupons: mc, tenants: t}
}

type OrderCreateItem struct {
	ProductID uint64 `json:"product_id" binding:"required"`
	SKUID     uint64 `json:"sku_id"`
	Quantity  int    `json:"quantity" binding:"required,min=1"`
}

type OrderCreateInput struct {
	Items        []OrderCreateItem `json:"items" binding:"required,min=1"`
	DeliveryType string            `json:"delivery_type" binding:"required"` // express/city/self_pickup
	Receiver     struct {
		Name     string `json:"name"`
		Phone    string `json:"phone"`
		Province string `json:"province"`
		City     string `json:"city"`
		District string `json:"district"`
		Address  string `json:"address"`
		Postcode string `json:"postcode"`
	} `json:"receiver"`
	BuyerRemark    string `json:"buyer_remark"`
	MemberCouponID uint64 `json:"member_coupon_id"`
}

type couponApplication struct {
	MemberCouponID uint64
	CouponID       uint64
	DiscountAmount decimal.Decimal
}

type orderTransitionInput struct {
	TenantID     uint64
	OrderID      uint64
	OrderNo      string
	MemberID     uint64
	AllowedFrom  []string
	ToStatus     string
	Fields       map[string]interface{}
	OperatorType string
	OperatorID   uint64
	Action       string
	Remark       string
	MessageType  string
	MessageTitle string
	MessageBody  string
}

func containsStatus(statuses []string, target string) bool {
	for _, item := range statuses {
		if item == target {
			return true
		}
	}
	return false
}

func (s *OrderService) transition(ctx context.Context, input orderTransitionInput) error {
	tenantID := input.TenantID
	if tenantID == 0 {
		if tenant := ctxkeys.GetTenant(ctx); tenant != nil {
			tenantID = tenant.ID
		}
	}
	if tenantID == 0 {
		return apperr.ErrTenantRequired
	}
	return s.orders.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var order model.Order
		query := tx.Where("tenant_id = ?", tenantID)
		if input.OrderID > 0 {
			query = query.Where("id = ?", input.OrderID)
		} else {
			query = query.Where("order_no = ?", input.OrderNo)
		}
		if input.MemberID > 0 {
			query = query.Where("member_id = ?", input.MemberID)
		}
		if err := query.First(&order).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return apperr.ErrNotFound
			}
			return err
		}
		if !containsStatus(input.AllowedFrom, order.Status) {
			return apperr.New(30010, "订单状态不可操作")
		}
		fields := map[string]interface{}{"status": input.ToStatus}
		for k, v := range input.Fields {
			fields[k] = v
		}
		if err := tx.Model(&model.Order{}).
			Where("id = ? AND tenant_id = ? AND status IN ?", order.ID, tenantID, input.AllowedFrom).
			Updates(fields).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.OrderLog{
			TenantID:     tenantID,
			OrderID:      order.ID,
			OperatorType: input.OperatorType,
			OperatorID:   input.OperatorID,
			Action:       input.Action,
			BeforeStatus: order.Status,
			AfterStatus:  input.ToStatus,
			Remark:       input.Remark,
		}).Error; err != nil {
			return err
		}
		if input.MessageTitle != "" {
			if err := tx.Create(&model.OrderMessage{
				TenantID:  tenantID,
				OrderID:   order.ID,
				OrderNo:   order.OrderNo,
				EventType: input.MessageType,
				Title:     input.MessageTitle,
				Content:   input.MessageBody,
				Status:    model.OrderMessageStatusUnread,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *OrderService) Create(ctx context.Context, memberID uint64, in *OrderCreateInput) (*model.Order, error) {
	// 月订单上限校验
	cnt, err := s.orders.CountMonth(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.tenants.CheckOrderLimit(ctx, cnt); err != nil {
		return nil, err
	}

	var (
		items       []model.OrderItem
		totalAmount = decimal.Zero
		tenantID    = ctxkeys.GetTenant(ctx).ID
		allVirtual  = true
		virtualMap  = make(map[uint64]bool)
	)

	// 预加载商品/SKU
	for _, it := range in.Items {
		p, err := s.products.FindByID(ctx, it.ProductID)
		if err != nil {
			return nil, err
		}
		if p == nil || p.Status != 1 {
			return nil, apperr.New(20004, "商品不存在或已下架")
		}
		if p.IsVirtual != 1 {
			allVirtual = false
		} else {
			virtualMap[p.ID] = true
		}
		price := p.Price
		skuDesc := ""
		if it.SKUID > 0 {
			sku, err := s.skus.FindByID(ctx, it.SKUID)
			if err != nil {
				return nil, err
			}
			if sku == nil || sku.ProductID != p.ID {
				return nil, apperr.New(20005, "SKU 不存在")
			}
			price = sku.Price
			skuDesc = string(sku.Attributes)
		}
		itemTotal := price.Mul(decimal.NewFromInt(int64(it.Quantity)))
		totalAmount = totalAmount.Add(itemTotal)
		items = append(items, model.OrderItem{
			TenantID:    tenantID,
			ProductID:   p.ID,
			SKUID:       it.SKUID,
			ProductName: p.Name,
			SKUDesc:     skuDesc,
			CoverImage:  p.CoverImage,
			Price:       price,
			Quantity:    it.Quantity,
			ItemTotal:   itemTotal,
		})
	}

	appliedCoupon, err := s.applyMemberCoupon(ctx, memberID, in.MemberCouponID, totalAmount)
	if err != nil {
		return nil, err
	}
	discountAmount := decimal.Zero
	couponID := uint64(0)
	if appliedCoupon != nil {
		discountAmount = appliedCoupon.DiscountAmount
		couponID = appliedCoupon.CouponID
	}
	actualAmount := totalAmount.Sub(discountAmount)
	if actualAmount.IsNegative() {
		actualAmount = decimal.Zero
	}

	order := &model.Order{
		TenantID:         tenantID,
		OrderNo:          utils.OrderNo("O"),
		MemberID:         memberID,
		TotalAmount:      totalAmount,
		DiscountAmount:   discountAmount,
		CouponID:         couponID,
		ActualAmount:     actualAmount,
		Status:           model.OrderStatusPendingPay,
		DeliveryType:     in.DeliveryType,
		ReceiverName:     in.Receiver.Name,
		ReceiverPhone:    in.Receiver.Phone,
		ReceiverProvince: in.Receiver.Province,
		ReceiverCity:     in.Receiver.City,
		ReceiverDistrict: in.Receiver.District,
		ReceiverAddress:  in.Receiver.Address,
		ReceiverPostcode: in.Receiver.Postcode,
		BuyerRemark:      in.BuyerRemark,
	}
	if len(items) > 0 && allVirtual {
		order.IsVirtual = 1
	}

	err = s.orders.CreateWithItems(ctx, order, items, func(tx *gorm.DB) error {
		for _, it := range in.Items {
			// 虚拟商品无库存概念
			if virtualMap[it.ProductID] {
				continue
			}
			if it.SKUID > 0 {
				if err := s.skus.DecreaseStock(ctx, tx, it.SKUID, it.Quantity); err != nil {
					return apperr.ErrStockShortage
				}
			} else {
				if err := s.products.DecreaseStock(ctx, tx, it.ProductID, it.Quantity); err != nil {
					return apperr.ErrStockShortage
				}
			}
		}
		if appliedCoupon != nil {
			if err := s.memberCoupons.MarkUsed(ctx, tx, appliedCoupon.MemberCouponID, memberID, order.ID); err != nil {
				if errors.Is(err, repository.ErrCouponUseLimitReached) {
					return apperr.New(30026, "已达到该优惠券使用次数限制")
				}
				return apperr.New(30025, "优惠券已被使用或不可用")
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	_ = s.logs.Create(ctx, &model.OrderLog{
		OrderID: order.ID, OperatorType: "member", OperatorID: memberID,
		Action: "create", AfterStatus: order.Status,
	})
	if s.messages != nil {
		_ = s.messages.Create(ctx, &model.OrderMessage{
			OrderID:   order.ID,
			OrderNo:   order.OrderNo,
			EventType: "order_created",
			Title:     "收到新订单",
			Content:   fmt.Sprintf("订单 %s 已创建，当前状态：待付款。", order.OrderNo),
		})
	}
	return order, nil
}

func (s *OrderService) applyMemberCoupon(ctx context.Context, memberID, memberCouponID uint64, totalAmount decimal.Decimal) (*couponApplication, error) {
	if memberCouponID == 0 {
		return nil, nil
	}
	if s.memberCoupons == nil || s.coupons == nil {
		return nil, apperr.New(30023, "优惠券服务不可用")
	}
	memberCoupon, err := s.memberCoupons.FindByIDForMember(ctx, memberID, memberCouponID)
	if err != nil {
		return nil, err
	}
	if memberCoupon == nil || memberCoupon.Status != "unused" {
		return nil, apperr.New(30024, "优惠券不可用")
	}
	now := time.Now()
	if now.Before(memberCoupon.ValidStartAt) || now.After(memberCoupon.ValidEndAt) {
		return nil, apperr.New(30024, "优惠券不在有效期内")
	}
	if memberCoupon.ThresholdAmount != nil && totalAmount.LessThan(*memberCoupon.ThresholdAmount) {
		return nil, apperr.New(30024, "订单金额未达到优惠券使用门槛")
	}
	coupon, err := s.coupons.FindByID(ctx, memberCoupon.CouponID)
	if err != nil {
		return nil, err
	}
	if coupon == nil || coupon.Status != 1 {
		return nil, apperr.New(30024, "优惠券不可用")
	}
	if memberCoupon.UseLimit > 0 {
		usedCount, err := s.memberCoupons.CountUsedByMemberCoupon(ctx, memberID, coupon.ID)
		if err != nil {
			return nil, err
		}
		if usedCount >= int64(memberCoupon.UseLimit) {
			return nil, apperr.New(30026, "已达到该优惠券使用次数限制")
		}
	}
	discountAmount := calcCouponDiscount(memberCoupon, totalAmount)
	return &couponApplication{MemberCouponID: memberCoupon.ID, CouponID: memberCoupon.CouponID, DiscountAmount: discountAmount}, nil
}

func calcCouponDiscount(coupon *model.MemberCoupon, totalAmount decimal.Decimal) decimal.Decimal {
	if coupon == nil || totalAmount.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero
	}
	discountAmount := decimal.Zero
	switch coupon.CouponType {
	case model.CouponTypeCash:
		discountAmount = coupon.DiscountValue
	case model.CouponTypeDiscount:
		rate := coupon.DiscountValue
		if rate.GreaterThan(decimal.NewFromInt(1)) {
			rate = rate.Div(decimal.NewFromInt(100))
		}
		if rate.LessThan(decimal.Zero) {
			rate = decimal.Zero
		}
		if rate.GreaterThan(decimal.NewFromInt(1)) {
			rate = decimal.NewFromInt(1)
		}
		discountAmount = totalAmount.Mul(decimal.NewFromInt(1).Sub(rate))
		if coupon.MaxDiscount != nil && coupon.MaxDiscount.GreaterThan(decimal.Zero) && discountAmount.GreaterThan(*coupon.MaxDiscount) {
			discountAmount = *coupon.MaxDiscount
		}
	case model.CouponTypeShipping:
		discountAmount = decimal.Zero
	}
	if discountAmount.IsNegative() {
		return decimal.Zero
	}
	if discountAmount.GreaterThan(totalAmount) {
		return totalAmount
	}
	return discountAmount.Round(2)
}

func (s *OrderService) Detail(ctx context.Context, id uint64) (*model.Order, error) {
	o, err := s.orders.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if o == nil {
		return nil, apperr.ErrNotFound
	}
	return o, nil
}

func (s *OrderService) List(ctx context.Context, q repository.OrderListQuery) ([]model.Order, int64, error) {
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.Size <= 0 || q.Size > 100 {
		q.Size = 20
	}
	return s.orders.List(ctx, q)
}

func (s *OrderService) Cancel(ctx context.Context, id, memberID uint64) error {
	now := time.Now()
	return s.transition(ctx, orderTransitionInput{
		OrderID:      id,
		MemberID:     memberID,
		AllowedFrom:  []string{model.OrderStatusPendingPay},
		ToStatus:     model.OrderStatusCancelled,
		Fields:       map[string]interface{}{"cancelled_at": now},
		OperatorType: "member",
		OperatorID:   memberID,
		Action:       "cancel",
		MessageType:  "order_cancelled",
		MessageTitle: "订单已取消",
		MessageBody:  fmt.Sprintf("订单 #%d 已被买家取消。", id),
	})
}

func (s *OrderService) Confirm(ctx context.Context, id, memberID uint64) error {
	now := time.Now()
	return s.transition(ctx, orderTransitionInput{
		OrderID:      id,
		MemberID:     memberID,
		AllowedFrom:  []string{model.OrderStatusShipped, model.OrderStatusDelivered},
		ToStatus:     model.OrderStatusCompleted,
		Fields:       map[string]interface{}{"completed_at": now},
		OperatorType: "member",
		OperatorID:   memberID,
		Action:       "confirm",
		MessageType:  "order_completed",
		MessageTitle: "订单已完成",
		MessageBody:  fmt.Sprintf("订单 #%d 已由买家确认收货。", id),
	})
}

func (s *OrderService) Prepare(ctx context.Context, id uint64, adminID uint64) error {
	now := time.Now()
	return s.transition(ctx, orderTransitionInput{
		OrderID:      id,
		AllowedFrom:  []string{model.OrderStatusPaid},
		ToStatus:     model.OrderStatusPreparing,
		Fields:       map[string]interface{}{"updated_at": now},
		OperatorType: "admin",
		OperatorID:   adminID,
		Action:       "prepare",
		MessageType:  "order_preparing",
		MessageTitle: "订单处理中",
		MessageBody:  fmt.Sprintf("订单 #%d 已开始处理，商家正在备货。", id),
	})
}

func (s *OrderService) Ship(ctx context.Context, id uint64, company, no string, adminID uint64) error {
	now := time.Now()
	return s.transition(ctx, orderTransitionInput{
		OrderID:      id,
		AllowedFrom:  []string{model.OrderStatusPaid, model.OrderStatusPreparing},
		ToStatus:     model.OrderStatusShipped,
		Fields:       map[string]interface{}{"express_company": company, "express_no": no, "shipped_at": now},
		OperatorType: "admin",
		OperatorID:   adminID,
		Action:       "ship",
		Remark:       company + " " + no,
		MessageType:  "order_shipped",
		MessageTitle: "订单已发货",
		MessageBody:  fmt.Sprintf("订单 #%d 已发货，物流：%s %s。", id, company, no),
	})
}

func (s *OrderService) MarkPaid(ctx context.Context, tenantID uint64, orderNo string, isVirtual bool) error {
	now := time.Now()
	toStatus := model.OrderStatusPaid
	messageType := "order_paid"
	messageTitle := "订单已支付"
	messageBody := fmt.Sprintf("订单 %s 已支付成功，等待卖家处理。", orderNo)
	fields := map[string]interface{}{"paid_at": now}
	if isVirtual {
		toStatus = model.OrderStatusCompleted
		messageType = "order_paid_completed"
		messageTitle = "虚拟订单已完成"
		messageBody = fmt.Sprintf("虚拟订单 %s 已支付并自动完成。", orderNo)
		fields["completed_at"] = now
	}
	return s.transition(ctx, orderTransitionInput{
		TenantID:     tenantID,
		OrderNo:      orderNo,
		AllowedFrom:  []string{model.OrderStatusPendingPay},
		ToStatus:     toStatus,
		Fields:       fields,
		OperatorType: "system",
		Action:       "pay_success",
		MessageType:  messageType,
		MessageTitle: messageTitle,
		MessageBody:  messageBody,
	})
}

func (s *OrderService) ListLogs(ctx context.Context, orderID uint64) ([]model.OrderLog, error) {
	return s.logs.ListByOrder(ctx, orderID)
}

func (s *OrderService) ListMessages(ctx context.Context, q repository.OrderMessageListQuery) ([]model.OrderMessage, int64, int64, error) {
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.Size <= 0 || q.Size > 100 {
		q.Size = 20
	}
	rows, total, err := s.messages.List(ctx, q)
	if err != nil {
		return nil, 0, 0, err
	}
	unread, err := s.messages.CountUnread(ctx)
	if err != nil {
		return nil, 0, 0, err
	}
	return rows, total, unread, nil
}

func (s *OrderService) MarkMessageRead(ctx context.Context, id uint64) error {
	return s.messages.MarkRead(ctx, id)
}

func (s *OrderService) MarkAllMessagesRead(ctx context.Context) error {
	return s.messages.MarkAllRead(ctx)
}
