package trade

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

type TradeReq struct {
	g.Meta `path:"/trades" method:"get" tags:"trades" summary:"我的成交记录"`
	Symbol string `json:"symbol"`
}

func (c *ControllerV1) GetTrades(ctx context.Context, req *TradeReq) (res []model.Trade, err error) {
	symbol := req.Symbol
	log.Println("symbol====", symbol)
	return service.Trade.GetTrades(ctx, symbol)
}

type OrderBookReq struct {
	g.Meta `path:"/orderbook" method:"get" tags:"orderbook" summary:"我的成交记录"`
	Symbol string `json:"symbol"`
}

func (c *ControllerV1) GetOrderBook(ctx context.Context, req *OrderBookReq) (res *model.PlaceOrderBookRes, err error) {
	// 从请求上下文中获取用户ID
	symbol := req.Symbol
	log.Println("symbol====", symbol)
	// 调用 service 层获取成交记录
	// service.Trade.ListMyTrades 会查询 trades 表，并通过 buy_user_id 或 sell_user_id 匹配
	data, err := service.Trade.GetOrderBook(ctx, 0, symbol)
	return &data, nil
}
