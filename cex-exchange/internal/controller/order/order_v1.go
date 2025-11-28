package order

import (
	"context"

	"cex-exchange/internal/model"
	"cex-exchange/internal/service"

	"github.com/gogf/gf/v2/frame/g"
)

type ControllerV1 struct{}

func NewV1() *ControllerV1 {
	return &ControllerV1{}
}

type PlaceOrderReq struct {
	g.Meta `path:"/order" method:"post" tags:"Order" summary:"下单"`
	Symbol string `json:"symbol" v:"required#请输入交易对"`
	Side   string `json:"side" v:"required|in:buy,sell#请选择方向|方向只能是buy或sell"`
	Type   string `json:"type" v:"required|in:limit,market#请选择类型|类型只能是limit或market"`
	Price  string `json:"price"`
	Amount string `json:"amount" v:"required#请输入数量"`
}

type PlaceOrderRes struct {
	OrderID uint64 `json:"order_id"`
}

func (c *ControllerV1) PlaceOrder(ctx context.Context, req *PlaceOrderReq) (res *PlaceOrderRes, err error) {
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

	if err != nil {
		return nil, err
	}

	orderID, _ := result.LastInsertId()

	// TODO: 调用撮合引擎

	return &PlaceOrderRes{OrderID: uint64(orderID)}, nil
}

type ListMyOrdersReq struct {
	g.Meta `path:"/my_orders" method:"get" tags:"Order" summary:"我的订单"`
}

// ListMyOrders 获取当前用户的订单列表
func (c *ControllerV1) ListMyOrders(ctx context.Context, req *ListMyOrdersReq) (res []model.Order, err error) {
	// 从请求上下文中获取用户ID
	uid := g.RequestFromCtx(ctx).GetCtxVar("uid").Uint64()
	// 调用 service 层获取订单列表
	return service.Order.ListMyOrders(ctx, uid)
}

type ListMyTradesReq struct {
	g.Meta `path:"/my_trades" method:"get" tags:"Trade" summary:"我的成交记录"`
}

// ListMyTrades 获取当前用户的成交记录
func (c *ControllerV1) ListMyTrades(ctx context.Context, req *ListMyTradesReq) (res []model.Trade, err error) {
	// 从请求上下文中获取用户ID
	uid := g.RequestFromCtx(ctx).GetCtxVar("uid").Uint64()
	// 调用 service 层获取成交记录
	// service.Trade.ListMyTrades 会查询 trades 表，并通过 buy_user_id 或 sell_user_id 匹配
	return service.Trade.ListMyTrades(ctx, uid)
}
