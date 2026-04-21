package service

import (
	"context"
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
	orders   *repository.OrderRepo
	logs     *repository.OrderLogRepo
	products *repository.ProductRepo
	skus     *repository.ProductSKURepo
	tenants  *TenantService
}

func NewOrderService(o *repository.OrderRepo, l *repository.OrderLogRepo, p *repository.ProductRepo, s *repository.ProductSKURepo, t *TenantService) *OrderService {
	return &OrderService{orders: o, logs: l, products: p, skus: s, tenants: t}
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
	BuyerRemark string `json:"buyer_remark"`
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

	order := &model.Order{
		TenantID:         tenantID,
		OrderNo:          utils.OrderNo("O"),
		MemberID:         memberID,
		TotalAmount:      totalAmount,
		ActualAmount:     totalAmount,
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
		return nil
	})
	if err != nil {
		return nil, err
	}
	_ = s.logs.Create(ctx, &model.OrderLog{
		OrderID: order.ID, OperatorType: "member", OperatorID: memberID,
		Action: "create", AfterStatus: order.Status,
	})
	return order, nil
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
	if err := s.orders.UpdateStatus(ctx, id, model.OrderStatusPendingPay, model.OrderStatusCancelled,
		map[string]interface{}{"cancelled_at": now}); err != nil {
		return err
	}
	_ = s.logs.Create(ctx, &model.OrderLog{
		OrderID: id, OperatorType: "member", OperatorID: memberID,
		Action: "cancel", BeforeStatus: model.OrderStatusPendingPay, AfterStatus: model.OrderStatusCancelled,
	})
	return nil
}

func (s *OrderService) Confirm(ctx context.Context, id, memberID uint64) error {
	now := time.Now()
	if err := s.orders.UpdateStatus(ctx, id, model.OrderStatusDelivered, model.OrderStatusCompleted,
		map[string]interface{}{"completed_at": now}); err != nil {
		return err
	}
	_ = s.logs.Create(ctx, &model.OrderLog{
		OrderID: id, OperatorType: "member", OperatorID: memberID,
		Action: "confirm", BeforeStatus: model.OrderStatusDelivered, AfterStatus: model.OrderStatusCompleted,
	})
	return nil
}

func (s *OrderService) Ship(ctx context.Context, id uint64, company, no string, adminID uint64) error {
	now := time.Now()
	err := s.orders.DB().WithContext(ctx).Model(&model.Order{}).
		Where("id = ? AND tenant_id = ? AND is_virtual = 0 AND status IN ?", id, ctxkeys.GetTenant(ctx).ID,
			[]string{model.OrderStatusPaid, model.OrderStatusPreparing}).
		Updates(map[string]interface{}{
			"status": model.OrderStatusShipped, "express_company": company,
			"express_no": no, "shipped_at": now,
		}).Error
	if err != nil {
		return err
	}
	_ = s.logs.Create(ctx, &model.OrderLog{
		OrderID: id, OperatorType: "admin", OperatorID: adminID,
		Action: "ship", AfterStatus: model.OrderStatusShipped,
		Remark: company + " " + no,
	})
	return nil
}
