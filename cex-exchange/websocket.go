package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// WebSocket连接管理器
type WebSocketManager struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan interface{}
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	mutex      sync.RWMutex
}

// 行情数据类型
type MarketData struct {
	Type      string      `json:"type"`
	Symbol    string      `json:"symbol"`
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp"`
}

// 订单簿数据
type OrderBookData struct {
	Symbol string     `json:"symbol"`
	Bids   [][]string `json:"bids"` // [price, amount]
	Asks   [][]string `json:"asks"` // [price, amount]
}

// 最新成交数据
type TradeData struct {
	Symbol    string `json:"symbol"`
	Price     string `json:"price"`
	Amount    string `json:"amount"`
	Side      string `json:"side"` // buy/sell
	Timestamp int64  `json:"timestamp"`
}

// K线数据
type KlineData struct {
	Symbol    string `json:"symbol"`
	Interval  string `json:"interval"`
	Open      string `json:"open"`
	High      string `json:"high"`
	Low       string `json:"low"`
	Close     string `json:"close"`
	Volume    string `json:"volume"`
	OpenTime  int64  `json:"open_time"`
	CloseTime int64  `json:"close_time"`
}

// 价格变动数据
type PriceChangeData struct {
	Symbol        string `json:"symbol"`
	Price         string `json:"price"`
	Change        string `json:"change"`
	ChangePercent string `json:"change_percent"`
	High24h       string `json:"high_24h"`
	Low24h        string `json:"low_24h"`
	Volume24h     string `json:"volume_24h"`
}

var (
	wsManager = &WebSocketManager{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan interface{}, 100),
		register:   make(chan *websocket.Conn, 100),
		unregister: make(chan *websocket.Conn, 100),
	}
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // 允许所有来源，生产环境应该限制
		},
	}
)

// 启动WebSocket管理器
func (manager *WebSocketManager) start() {
	for {
		select {
		case client := <-manager.register:
			manager.mutex.Lock()
			manager.clients[client] = true
			manager.mutex.Unlock()
			log.Printf("WebSocket客户端连接: %s", client.RemoteAddr())

		case client := <-manager.unregister:
			manager.mutex.Lock()
			if _, ok := manager.clients[client]; ok {
				delete(manager.clients, client)
				client.Close()
			}
			manager.mutex.Unlock()
			log.Printf("WebSocket客户端断开: %s", client.RemoteAddr())

		case message := <-manager.broadcast:
			manager.mutex.RLock()
			for client := range manager.clients {
				err := client.WriteJSON(message)
				if err != nil {
					log.Printf("发送消息失败: %v", err)
					client.Close()
					delete(manager.clients, client)
				}
			}
			manager.mutex.RUnlock()
		}
	}
}

// WebSocket连接处理
func handleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket升级失败: %v", err)
		return
	}

	wsManager.register <- conn

	// 发送欢迎消息
	welcomeMsg := MarketData{
		Type:      "welcome",
		Symbol:    "",
		Data:      "WebSocket连接成功",
		Timestamp: time.Now().Unix(),
	}
	conn.WriteJSON(welcomeMsg)

	// 处理客户端消息
	for {
		var msg map[string]interface{}
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("读取消息失败: %v", err)
			wsManager.unregister <- conn
			break
		}

		// 处理订阅请求
		if msgType, ok := msg["type"].(string); ok {
			switch msgType {
			case "subscribe":
				if symbol, ok := msg["symbol"].(string); ok {
					// 发送当前行情数据
					sendMarketData(symbol)
				}
			case "ping":
				// 响应pong
				pongMsg := MarketData{
					Type:      "pong",
					Symbol:    "",
					Data:      "",
					Timestamp: time.Now().Unix(),
				}
				conn.WriteJSON(pongMsg)
			}
		}
	}
}

// 发送行情数据
func sendMarketData(symbol string) {
	// 发送订单簿数据
	orderbook := getOrderBookData(symbol)
	orderbookMsg := MarketData{
		Type:      "orderbook",
		Symbol:    symbol,
		Data:      orderbook,
		Timestamp: time.Now().Unix(),
	}
	wsManager.broadcast <- orderbookMsg

	// 发送最新成交数据
	trades := getLatestTrades(symbol)
	for _, trade := range trades {
		tradeMsg := MarketData{
			Type:      "trade",
			Symbol:    symbol,
			Data:      trade,
			Timestamp: time.Now().Unix(),
		}
		wsManager.broadcast <- tradeMsg
	}

	// 发送价格变动数据
	priceChange := getPriceChangeData(symbol)
	priceMsg := MarketData{
		Type:      "price_change",
		Symbol:    symbol,
		Data:      priceChange,
		Timestamp: time.Now().Unix(),
	}
	wsManager.broadcast <- priceMsg
}

// 获取订单簿数据
func getOrderBookData(symbol string) OrderBookData {
	var bids, asks []Order
	DB.Where("symbol = ? AND side = 'buy' AND status = 'open'", symbol).
		Order("price desc").Limit(10).Find(&bids)
	DB.Where("symbol = ? AND side = 'sell' AND status = 'open'", symbol).
		Order("price asc").Limit(10).Find(&asks)

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

// 获取最新成交数据
func getLatestTrades(symbol string) []TradeData {
	var trades []Trade
	DB.Where("symbol = ?", symbol).Order("id desc").Limit(10).Find(&trades)

	tradeData := make([]TradeData, 0)
	for _, trade := range trades {
		side := "buy"
		if trade.BuyOrderID > trade.SellOrderID {
			side = "sell"
		}
		tradeData = append(tradeData, TradeData{
			Symbol:    trade.Symbol,
			Price:     trade.Price,
			Amount:    trade.Amount,
			Side:      side,
			Timestamp: trade.CreatedAt.Unix(),
		})
	}

	return tradeData
}

// 获取价格变动数据
func getPriceChangeData(symbol string) PriceChangeData {
	// 获取最新K线数据
	var kline Kline
	DB.Where("symbol = ?", symbol).Order("close_time desc").First(&kline)

	// 获取24小时前的K线数据
	yesterday := time.Now().Add(-24 * time.Hour).Unix()
	var yesterdayKline Kline
	DB.Where("symbol = ? AND close_time >= ?", symbol, yesterday).Order("close_time asc").First(&yesterdayKline)

	// 计算价格变动
	currentPrice, _ := strconv.ParseFloat(kline.Close, 64)
	openPrice, _ := strconv.ParseFloat(yesterdayKline.Open, 64)
	change := currentPrice - openPrice
	changePercent := (change / openPrice) * 100

	// 获取24小时最高最低价
	var highLowKline Kline
	DB.Where("symbol = ? AND close_time >= ?", symbol, yesterday).
		Select("MAX(CAST(high AS DECIMAL(20,8))) as high, MIN(CAST(low AS DECIMAL(20,8))) as low").
		First(&highLowKline)

	// 计算24小时成交量
	var volume24h float64
	DB.Model(&Kline{}).Where("symbol = ? AND close_time >= ?", symbol, yesterday).
		Select("SUM(CAST(volume AS DECIMAL(20,8)))").Scan(&volume24h)

	return PriceChangeData{
		Symbol:        symbol,
		Price:         kline.Close,
		Change:        fmt.Sprintf("%.8f", change),
		ChangePercent: fmt.Sprintf("%.2f", changePercent),
		High24h:       highLowKline.High,
		Low24h:        highLowKline.Low,
		Volume24h:     fmt.Sprintf("%.8f", volume24h),
	}
}

// 推送K线数据
func pushKlineData(symbol, interval string) {
	var kline Kline
	DB.Where("symbol = ? AND `interval` = ?", symbol, interval).
		Order("close_time desc").First(&kline)

	if kline.ID > 0 {
		klineData := KlineData{
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

		klineMsg := MarketData{
			Type:      "kline",
			Symbol:    symbol,
			Data:      klineData,
			Timestamp: time.Now().Unix(),
		}
		wsManager.broadcast <- klineMsg
	}
}

// 启动行情数据推送
func startMarketDataPusher() {
	ticker := time.NewTicker(1 * time.Second) // 每秒推送一次
	defer ticker.Stop()

	for range ticker.C {
		// 推送主要交易对的行情数据
		symbols := []string{"BTC/USDT", "ETH/USDT", "BNB/USDT"}
		for _, symbol := range symbols {
			sendMarketData(symbol)
		}
	}
}
