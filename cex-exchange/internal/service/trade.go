package service

import (
	"context"

	"cex-exchange/internal/dao"
	"cex-exchange/internal/model"
)

type ITrade interface {
	ListMyTrades(ctx context.Context, userID uint64) ([]model.Trade, error)
}

type tradeImpl struct{}

var Trade = &tradeImpl{}

// ListMyTrades 获取用户的成交记录
func (s *tradeImpl) ListMyTrades(ctx context.Context, userID uint64) ([]model.Trade, error) {
	return dao.Trade.ListByUserID(ctx, userID)
}

func (s *tradeImpl) GetTrades(ctx context.Context, symbol string) ([]model.Trade, error) {

	return dao.Trade.ListBySymbol(ctx, symbol, 100)
}

func (s *tradeImpl) GetOrderBook(ctx context.Context, uid uint64, symbol string) (model.PlaceOrderBookRes, error) {
	orderBooks, err := dao.Order.ListByUserID(ctx, uid, symbol)
	if err != nil {
		return model.PlaceOrderBookRes{}, err
	}
	var bids []model.Order
	var asks []model.Order
	for _, orderBook := range orderBooks {
		if orderBook.Side == "buy" {
			bids = append(bids, orderBook)
		} else {
			asks = append(asks, orderBook)
		}
	}
	return model.PlaceOrderBookRes{
		OrderBids: bids,
		OrderAsks: asks,
	}, err
}
