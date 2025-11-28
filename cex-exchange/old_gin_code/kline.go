package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type KlineQuery struct {
	Symbol   string `form:"symbol" binding:"required"`
	Interval string `form:"interval" binding:"required"`
	Limit    int    `form:"limit"`
}

// K线查询接口
func GetKlines(c *gin.Context) {
	var q KlineQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	if q.Limit == 0 {
		q.Limit = 100
	}
	var klines []Kline
	DB.Where("symbol = ? AND `interval` = ?", q.Symbol, q.Interval).Order("open_time desc").Limit(q.Limit).Find(&klines)
	c.JSON(http.StatusOK, klines)
}

// 简单K线生成逻辑（实际应由撮合或定时任务驱动）
func GenerateKline(symbol, interval string) {
	// 以当前时间为K线结束，1分钟K线为例
	now := time.Now().Unix()
	openTime := now - 60
	closeTime := now
	var trades []Trade
	DB.Where("symbol = ? AND created_at >= ? AND created_at < ?", symbol, time.Unix(openTime, 0), time.Unix(closeTime, 0)).Find(&trades)
	if len(trades) == 0 {
		return
	}
	open, _ := strconv.ParseFloat(trades[0].Price, 64)
	high, low := open, open
	close := open
	volume := 0.0
	for _, t := range trades {
		p, _ := strconv.ParseFloat(t.Price, 64)
		if p > high {
			high = p
		}
		if p < low {
			low = p
		}
		close = p
		v, _ := strconv.ParseFloat(t.Amount, 64)
		volume += v
	}
	kline := Kline{
		Symbol:    symbol,
		Interval:  interval,
		Open:      trades[0].Price,
		High:      strconv.FormatFloat(high, 'f', -1, 64),
		Low:       strconv.FormatFloat(low, 'f', -1, 64),
		Close:     strconv.FormatFloat(close, 'f', -1, 64),
		Volume:    strconv.FormatFloat(volume, 'f', -1, 64),
		OpenTime:  openTime,
		CloseTime: closeTime,
	}
	DB.Create(&kline)
}
