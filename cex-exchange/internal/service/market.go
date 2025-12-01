package service

import (
	"context"

	"cex-exchange/internal/model"
)

type IMarket interface {
	GetOrderBook(ctx context.Context, symbol string, depth int) (*model.OrderBook, error)
	GetTicker(ctx context.Context, symbol string) (*model.Ticker, error)
	GetKlines(ctx context.Context, symbol, interval string, startTime, endTime int64, limit int) ([]model.Kline, error)
	GetRecentTrades(ctx context.Context, symbol string, limit int) ([]model.TradeRecord, error)
}

type marketImpl struct{}

var Market = &marketImpl{}

// GetOrderBook 获取订单簿
func (s *marketImpl) GetOrderBook(ctx context.Context, symbol string, depth int) (*model.OrderBook, error) {
	// TODO: 实现从撮合引擎或数据库获取订单簿
	// 这里返回模拟数据
	return &model.OrderBook{
		Symbol: symbol,
		Bids: []model.PriceLevel{
			{Price: "49900", Amount: "1.5"},
			{Price: "49800", Amount: "2.3"},
			{Price: "49700", Amount: "3.1"},
		},
		Asks: []model.PriceLevel{
			{Price: "50100", Amount: "1.2"},
			{Price: "50200", Amount: "2.1"},
			{Price: "50300", Amount: "3.5"},
		},
	}, nil
}

// GetTicker 获取行情
func (s *marketImpl) GetTicker(ctx context.Context, symbol string) (*model.Ticker, error) {
	// TODO: 实现从缓存或数据库获取行情
	// 这里返回模拟数据
	return &model.Ticker{
		Symbol:     symbol,
		LastPrice:  "50000",
		High24h:    "52000",
		Low24h:     "48000",
		Volume24h:  "1234.56",
		Change24h:  "+2.5%",
	}, nil
}

// GetKlines 获取 K 线数据
func (s *marketImpl) GetKlines(ctx context.Context, symbol, interval string, startTime, endTime int64, limit int) ([]model.Kline, error) {
	// TODO: 实现从数据库获取 K 线数据
	// 这里返回模拟数据
	return []model.Kline{
		{
			Timestamp: 1638360000,
			Open:      "49000",
			High:      "50500",
			Low:       "48500",
			Close:     "50000",
			Volume:    "123.45",
		},
		{
			Timestamp: 1638363600,
			Open:      "50000",
			High:      "51000",
			Low:       "49500",
			Close:     "50500",
			Volume:    "234.56",
		},
	}, nil
}

// GetRecentTrades 获取最近成交
func (s *marketImpl) GetRecentTrades(ctx context.Context, symbol string, limit int) ([]model.TradeRecord, error) {
	// TODO: 实现从数据库获取最近成交
	// 这里返回模拟数据
	return []model.TradeRecord{
		{
			ID:        1,
			Symbol:    symbol,
			Price:     "50000",
			Amount:    "0.5",
			Side:      "buy",
			Timestamp: 1638360000,
		},
		{
			ID:        2,
			Symbol:    symbol,
			Price:     "50100",
			Amount:    "0.3",
			Side:      "sell",
			Timestamp: 1638360060,
		},
	}, nil
}
