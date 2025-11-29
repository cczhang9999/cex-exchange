package service

import (
	"context"
	"strconv"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"cex-exchange/internal/dao"
	"cex-exchange/internal/engine"
	"cex-exchange/internal/model"
)

type IOrder interface {
	PlaceOrder(ctx context.Context, userID uint64, symbol, side, orderType, price, amount string) (uint64, error)
	ListMyOrders(ctx context.Context, userID uint64) ([]model.Order, error)
	CancelOrder(ctx context.Context, userID, orderID uint64) error
}

type orderImpl struct{}

var Order = &orderImpl{}

// PlaceOrder 下单
func (s *orderImpl) PlaceOrder(ctx context.Context, req *model.PlaceOrderReq) (int64, error) {
	uid := g.RequestFromCtx(ctx).GetCtxVar("uid").Uint64()
	// 插入订单
	result, err := g.Model("orders").Ctx(ctx).Data(g.Map{
		"user_id": uid,
		"symbol":  req.Symbol,
		"side":    req.Side,
		"type":    req.Type,
		"price":   req.Price,
		"amount":  req.Amount,
		"filled":  "0",
		"status":  "open",
	}).Insert()

	orderID, _ := result.LastInsertId()
	
	// 调用撮合引擎
	order := &engine.Order{
		ID:        orderID,
		UserID:    uid,
		Symbol:    req.Symbol,
		Side:      req.Side,
		Type:      req.Type,
		Price:     parseFloat(req.Price),
		Amount:    parseFloat(req.Amount),
		Filled:    0,
		Status:    "open",
		CreatedAt: time.Now(),
	}
	
	matchEngine := engine.GetEngine()
	_, err = matchEngine.Match(ctx, order)
	
	return orderID, err
}

// parseFloat 辅助函数：字符串转浮点数
func parseFloat(s string) float64 {
	if s == "" {
		return 0
	}
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

// ListMyOrders 获取我的订单
func (s *orderImpl) ListMyOrders(ctx context.Context, uid uint64, symbol string) ([]model.Order, error) {
	return dao.Order.ListByUserID(ctx, uid, symbol)
}

// CancelOrder 撤单
func (s *orderImpl) CancelOrder(ctx context.Context, userID, orderID uint64) error {
	// 查询订单
	order, err := dao.Order.GetByID(ctx, orderID)
	if err != nil {
		return err
	}

	// 验证订单所有权
	if order.UserID != userID {
		return gerror.New("无权操作此订单")
	}

	// 验证订单状态
	if order.Status != "open" {
		return gerror.New("订单状态不允许撤销")
	}

	// 更新订单状态
	return dao.Order.Update(ctx, orderID, g.Map{
		"status": "cancelled",
	})
}
