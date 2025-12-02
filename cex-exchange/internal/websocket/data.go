package websocket

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// GetOrderBookData 获取订单簿数据
func GetOrderBookData(symbol string) OrderBookData {
	ctx := context.Background()

	// 获取买单（按价格降序）
	var bids []struct {
		Price  string
		Amount string
	}
	err := g.DB().Model("orders").
		Where("symbol = ? AND side = 'buy' AND status = 'open'", symbol).
		Order("CAST(price AS DECIMAL(20,8)) DESC").
		Limit(10).
		Fields("price, amount").
		Scan(&bids)

	if err != nil {
		g.Log().Error(ctx, "获取买单失败:", err)
	}

	// 获取卖单（按价格升序）
	var asks []struct {
		Price  string
		Amount string
	}
	err = g.DB().Model("orders").
		Where("symbol = ? AND side = 'sell' AND status = 'open'", symbol).
		Order("CAST(price AS DECIMAL(20,8)) ASC").
		Limit(10).
		Fields("price, amount").
		Scan(&asks)

	if err != nil {
		g.Log().Error(ctx, "获取卖单失败:", err)
	}

	// 转换为字符串数组格式
	bidsData := make([][]string, 0)
	asksData := make([][]string, 0)

	for _, bid := range bids {
		bidsData = append(bidsData, []string{bid.Price, bid.Amount})
	}

	for _, ask := range asks {
		asksData = append(asksData, []string{ask.Price, ask.Amount})
	}

	return OrderBookData{
		Symbol: symbol,
		Bids:   bidsData,
		Asks:   asksData,
	}
}

// GetLatestTrades 获取最新成交数据
func GetLatestTrades(symbol string) []TradeData {
	ctx := context.Background()

	var trades []struct {
		ID          int64
		Symbol      string
		Price       string
		Amount      string
		BuyOrderID  int64 `orm:"buy_order_id"`
		SellOrderID int64 `orm:"sell_order_id"`
		CreatedAt   time.Time
	}

	err := g.DB().Model("trades").
		Where("symbol = ?", symbol).
		Order("id DESC").
		Limit(10).
		Scan(&trades)

	if err != nil {
		g.Log().Error(ctx, "获取成交记录失败:", err)
		return []TradeData{}
	}

	tradeData := make([]TradeData, 0)
	for _, trade := range trades {
		side := "buy"
		if trade.BuyOrderID > trade.SellOrderID {
			side = "sell"
		}

		timestamp := time.Now().Unix()
		if !trade.CreatedAt.IsZero() {
			timestamp = trade.CreatedAt.Unix()
		}

		tradeData = append(tradeData, TradeData{
			Symbol:    trade.Symbol,
			Price:     trade.Price,
			Amount:    trade.Amount,
			Side:      side,
			Timestamp: timestamp,
		})
	}

	return tradeData
}

// GetPriceChangeData 获取价格变动数据
func GetPriceChangeData(symbol string) PriceChangeData {
	ctx := context.Background()

	// 获取最新K线数据
	var latestKline struct {
		Close string
	}
	err := g.DB().Model("klines").
		Where("symbol = ?", symbol).
		Order("close_time DESC").
		Fields("close").
		Limit(1).
		Scan(&latestKline)

	if err != nil {
		g.Log().Error(ctx, "获取最新K线失败:", err)
		return PriceChangeData{Symbol: symbol}
	}

	// 获取24小时前的K线数据
	yesterday := time.Now().Add(-24 * time.Hour).Unix()
	var yesterdayKline struct {
		Open string
	}
	err = g.DB().Model("klines").
		Where("symbol = ? AND close_time >= ?", symbol, yesterday).
		Order("close_time ASC").
		Fields("open").
		Limit(1).
		Scan(&yesterdayKline)

	if err != nil {
		g.Log().Error(ctx, "获取昨日K线失败:", err)
	}

	// 计算价格变动
	currentPrice, _ := strconv.ParseFloat(latestKline.Close, 64)
	openPrice, _ := strconv.ParseFloat(yesterdayKline.Open, 64)
	if openPrice == 0 {
		openPrice = currentPrice
	}

	change := currentPrice - openPrice
	changePercent := 0.0
	if openPrice != 0 {
		changePercent = (change / openPrice) * 100
	}

	// 获取24小时最高最低价
	var highLow struct {
		High string
		Low  string
	}
	err = g.DB().Model("klines").
		Where("symbol = ? AND close_time >= ?", symbol, yesterday).
		Fields("MAX(CAST(high AS DECIMAL(20,8))) as high, MIN(CAST(low AS DECIMAL(20,8))) as low").
		Scan(&highLow)

	if err != nil {
		g.Log().Error(ctx, "获取24小时最高最低价失败:", err)
	}

	// 计算24小时成交量
	var volume24h float64
	err = g.DB().Model("klines").
		Where("symbol = ? AND close_time >= ?", symbol, yesterday).
		Fields("SUM(CAST(volume AS DECIMAL(20,8))) as total_volume").
		Scan(&volume24h)

	if err != nil {
		g.Log().Error(ctx, "获取24小时成交量失败:", err)
	}

	return PriceChangeData{
		Symbol:        symbol,
		Price:         latestKline.Close,
		Change:        fmt.Sprintf("%.8f", change),
		ChangePercent: fmt.Sprintf("%.2f", changePercent),
		High24h:       highLow.High,
		Low24h:        highLow.Low,
		Volume24h:     fmt.Sprintf("%.8f", volume24h),
	}
}

// GetKlineData 获取K线数据
func GetKlineData(symbol, interval string) *KlineData {
	ctx := context.Background()

	var kline struct {
		Symbol    string
		Interval  string
		Open      string
		High      string
		Low       string
		Close     string
		Volume    string
		OpenTime  int64 `orm:"open_time"`
		CloseTime int64 `orm:"close_time"`
	}

	err := g.DB().Model("klines").
		Where("symbol = ? AND `interval` = ?", symbol, interval).
		Order("close_time DESC").
		Limit(1).
		Scan(&kline)

	if err != nil {
		g.Log().Error(ctx, "获取K线数据失败:", err)
		return nil
	}

	if kline.Symbol == "" {
		return nil
	}

	return &KlineData{
		Symbol:    kline.Symbol,
		Interval:  kline.Interval,
		Open:      kline.Open,
		High:      kline.High,
		Low:       kline.Low,
		Close:     kline.Close,
		Volume:    kline.Volume,
		OpenTime:  kline.OpenTime,
		CloseTime: kline.CloseTime,
	}
}
