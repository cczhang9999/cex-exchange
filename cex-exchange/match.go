package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type PlaceOrderRequest struct {
	Symbol string `json:"symbol" binding:"required"`
	Side   string `json:"side" binding:"required"` // buy/sell
	Type   string `json:"type" binding:"required"` // limit/market
	Price  string `json:"price"`                   // 限价单必填
	Amount string `json:"amount" binding:"required"`
}

// 下单接口
func PlaceOrder(c *gin.Context) {
	// 通过 Gin 的 Context 从 JWT 中间件设置的上下文获取用户ID
	uid := c.GetUint64("uid")
	var req PlaceOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	order := Order{
		UserID: uid,
		Symbol: req.Symbol,
		Side:   req.Side,
		Type:   req.Type,
		Price:  req.Price,
		Amount: req.Amount,
		Filled: "0",
		Status: "open",
	}
	if err := DB.Create(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "下单失败"})
		return
	}
	// 简单撮合逻辑
	MatchOrder(&order)
	c.JSON(http.StatusOK, order)
}

// 撤单接口
func CancelOrder(c *gin.Context) {
	uid := c.GetUint64("uid")
	id := c.Param("id")
	var order Order
	if err := DB.Where("id = ? AND user_id = ? AND status = 'open'", id, uid).First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "订单不存在或已成交/撤销"})
		return
	}
	order.Status = "cancelled"
	DB.Save(&order)
	c.JSON(http.StatusOK, gin.H{"message": "撤单成功"})
}

// 订单簿接口（返回前10档买卖盘）
func OrderBook(c *gin.Context) {
	symbol := c.Query("symbol")
	var bids, asks []Order
	DB.Where("symbol = ? AND side = 'buy' AND status = 'open'", symbol).Order("price desc").Limit(10).Find(&bids)
	DB.Where("symbol = ? AND side = 'sell' AND status = 'open'", symbol).Order("price asc").Limit(10).Find(&asks)
	c.JSON(http.StatusOK, gin.H{"bids": bids, "asks": asks})
}

// 成交记录接口
func TradeHistory(c *gin.Context) {
	symbol := c.Query("symbol")
	var trades []Trade
	DB.Where("symbol = ?", symbol).Order("id desc").Limit(50).Find(&trades)
	c.JSON(http.StatusOK, trades)
}

// 简单撮合引擎（仅限价撮合，实际应异步、并发、锁定处理）
func MatchOrder(order *Order) {
	if order.Type != "limit" {
		return // 仅演示限价单
	}
	var counterOrders []Order
	if order.Side == "buy" {
		DB.Where("symbol = ? AND side = 'sell' AND status = 'open' AND price <= ?", order.Symbol, order.Price).
			Order("price asc").Find(&counterOrders)
	} else {
		DB.Where("symbol = ? AND side = 'buy' AND status = 'open' AND price >= ?", order.Symbol, order.Price).
			Order("price desc").Find(&counterOrders)
	}
	for _, co := range counterOrders {
		// 假设全部可成交，实际应精确计算
		tradeAmount := co.Amount // 简化处理
		trade := Trade{
			BuyOrderID:  order.ID,
			SellOrderID: co.ID,
			Symbol:      order.Symbol,
			Price:       co.Price,
			Amount:      tradeAmount,
			BuyUserID:   order.UserID,
			SellUserID:  co.UserID,
		}
		DB.Create(&trade)
		order.Filled = tradeAmount
		order.Status = "filled"
		DB.Save(order)
		co.Filled = tradeAmount
		co.Status = "filled"
		DB.Save(&co)
		// 撮合成交后自动生成K线
		GenerateKline(order.Symbol, "1m")
		break // 仅撮合一笔
	}
}

// 我的订单列表
func ListMyOrders(c *gin.Context) {
	uid := c.GetUint64("uid")
	var orders []Order
	DB.Where("user_id = ?", uid).Order("id desc").Find(&orders)
	c.JSON(http.StatusOK, orders)
}

// 我的成交记录
func ListMyTrades(c *gin.Context) {
	uid := c.GetUint64("uid")
	var trades []Trade
	DB.Where("buy_user_id = ? OR sell_user_id = ?", uid, uid).Order("id desc").Find(&trades)
	c.JSON(http.StatusOK, trades)
}
