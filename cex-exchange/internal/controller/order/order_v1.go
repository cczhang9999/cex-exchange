package order

import (
	"context"
	"log"

	"cex-exchange/internal/model"
	"cex-exchange/internal/service"

	"github.com/gogf/gf/v2/frame/g"
)

type ControllerV1 struct{}

func NewV1() *ControllerV1 {
	return &ControllerV1{}
}

type PlaceOrderRes struct {
	OrderID int64 `json:"orderId"`
}

func (c *ControllerV1) PlaceOrder(ctx context.Context, req *model.PlaceOrderReq) (res *PlaceOrderRes, err error) {
	orderId,err:=service.Order.PlaceOrder(ctx, req)
	return &PlaceOrderRes{orderId},err	

}

type ListMyOrdersReq struct {
	g.Meta `path:"/my_orders" method:"get" tags:"Order" summary:"我的订单"`
}

// ListMyOrders 获取当前用户的订单列表
func (c *ControllerV1) ListMyOrders(ctx context.Context, req *ListMyOrdersReq) (res []model.Order, err error) {
	// 从请求上下文中获取用户ID
	uid := g.RequestFromCtx(ctx).GetCtxVar("uid").Uint64()
	// 调用 service 层获取订单列表
	return service.Order.ListMyOrders(ctx, uid, "")
}

type ListMyTradesReq struct {
	g.Meta `path:"/my_trades" method:"get" tags:"Trade" summary:"我的成交记录"`
}

// ListMyTrades 获取当前用户的成交记录
func (c *ControllerV1) ListMyTrades(ctx context.Context, req *ListMyTradesReq) (res []model.Trade, err error) {
	// 从请求上下文中获取用户ID
	uid := g.RequestFromCtx(ctx).GetCtxVar("uid").Uint64()
	log.Println(uid)
	// 调用 service 层获取成交记录
	// service.Trade.ListMyTrades 会查询 trades 表，并通过 buy_user_id 或 sell_user_id 匹配
	return service.Trade.ListMyTrades(ctx, uid)
}
