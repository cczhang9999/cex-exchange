package websocket

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Manager WebSocket连接管理器
type Manager struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan interface{}
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	mutex      sync.RWMutex
}

// MarketData 行情数据类型
type MarketData struct {
	Type      string      `json:"type"`
	Symbol    string      `json:"symbol"`
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp"`
}

// OrderBookData 订单簿数据
type OrderBookData struct {
	Symbol string     `json:"symbol"`
	Bids   [][]string `json:"bids"` // [price, amount]
	Asks   [][]string `json:"asks"` // [price, amount]
}

// TradeData 最新成交数据
type TradeData struct {
	Symbol    string `json:"symbol"`
	Price     string `json:"price"`
	Amount    string `json:"amount"`
	Side      string `json:"side"` // buy/sell
	Timestamp int64  `json:"timestamp"`
}

// PriceChangeData 价格变动数据
type PriceChangeData struct {
	Symbol        string `json:"symbol"`
	Price         string `json:"price"`
	Change        string `json:"change"`
	ChangePercent string `json:"change_percent"`
	High24h       string `json:"high_24h"`
	Low24h        string `json:"low_24h"`
	Volume24h     string `json:"volume_24h"`
}

// KlineData K线数据
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

var (
	// WSManager 全局WebSocket管理器
	WSManager *Manager
	upgrader  = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // 允许所有来源，生产环境应该限制
		},
	}
)

// InitManager 初始化WebSocket管理器
func InitManager() {
	WSManager = &Manager{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan interface{}, 100),
		register:   make(chan *websocket.Conn, 100),
		unregister: make(chan *websocket.Conn, 100),
	}
	go WSManager.Start()
}

// Start 启动WebSocket管理器
func (m *Manager) Start() {
	for {
		select {
		case client := <-m.register:
			m.mutex.Lock()
			m.clients[client] = true
			m.mutex.Unlock()
			log.Printf("WebSocket客户端连接: %s", client.RemoteAddr())

		case client := <-m.unregister:
			m.mutex.Lock()
			if _, ok := m.clients[client]; ok {
				delete(m.clients, client)
				client.Close()
			}
			m.mutex.Unlock()
			log.Printf("WebSocket客户端断开: %s", client.RemoteAddr())

		case message := <-m.broadcast:
			m.mutex.RLock()
			for client := range m.clients {
				err := client.WriteJSON(message)
				if err != nil {
					log.Printf("发送消息失败: %v", err)
					client.Close()
					delete(m.clients, client)
				}
			}
			m.mutex.RUnlock()
		}
	}
}

// Broadcast 广播消息
func (m *Manager) Broadcast(message interface{}) {
	m.broadcast <- message
}

// GetClientCount 获取当前连接数
func (m *Manager) GetClientCount() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return len(m.clients)
}

// HandleConnection 处理WebSocket连接
func (m *Manager) HandleConnection(ctx context.Context, conn *websocket.Conn) {
	m.register <- conn

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
			m.unregister <- conn
			break
		}

		// 处理订阅请求
		if msgType, ok := msg["type"].(string); ok {
			switch msgType {
			case "subscribe":
				if symbol, ok := msg["symbol"].(string); ok {
					// 发送当前行情数据
					go SendMarketData(symbol)
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

// SendMarketData 发送行情数据
func SendMarketData(symbol string) {
	// 发送订单簿数据
	orderbook := GetOrderBookData(symbol)
	orderbookMsg := MarketData{
		Type:      "orderbook",
		Symbol:    symbol,
		Data:      orderbook,
		Timestamp: time.Now().Unix(),
	}
	WSManager.Broadcast(orderbookMsg)

	// 发送最新成交数据
	trades := GetLatestTrades(symbol)
	for _, trade := range trades {
		tradeMsg := MarketData{
			Type:      "trade",
			Symbol:    symbol,
			Data:      trade,
			Timestamp: time.Now().Unix(),
		}
		WSManager.Broadcast(tradeMsg)
	}

	// 发送价格变动数据
	priceChange := GetPriceChangeData(symbol)
	priceMsg := MarketData{
		Type:      "price_change",
		Symbol:    symbol,
		Data:      priceChange,
		Timestamp: time.Now().Unix(),
	}
	WSManager.Broadcast(priceMsg)
}

// BroadcastTrade 广播成交信息
func BroadcastTrade(symbol, price, amount, side string) {
	trade := TradeData{
		Symbol:    symbol,
		Price:     price,
		Amount:    amount,
		Side:      side,
		Timestamp: time.Now().Unix(),
	}

	msg := MarketData{
		Type:      "trade",
		Symbol:    symbol,
		Data:      trade,
		Timestamp: time.Now().Unix(),
	}

	WSManager.Broadcast(msg)
}

// BroadcastOrderBook 广播订单簿更新
func BroadcastOrderBook(symbol string) {
	orderbook := GetOrderBookData(symbol)
	msg := MarketData{
		Type:      "orderbook",
		Symbol:    symbol,
		Data:      orderbook,
		Timestamp: time.Now().Unix(),
	}

	WSManager.Broadcast(msg)
}

// BroadcastPriceChange 广播价格变动
func BroadcastPriceChange(symbol string) {
	priceChange := GetPriceChangeData(symbol)
	msg := MarketData{
		Type:      "price_change",
		Symbol:    symbol,
		Data:      priceChange,
		Timestamp: time.Now().Unix(),
	}

	WSManager.Broadcast(msg)
}

// BroadcastKline 广播K线数据
func BroadcastKline(symbol, interval string) {
	kline := GetKlineData(symbol, interval)
	if kline != nil {
		msg := MarketData{
			Type:      "kline",
			Symbol:    symbol,
			Data:      kline,
			Timestamp: time.Now().Unix(),
		}

		WSManager.Broadcast(msg)
	}
}

// StartMarketDataPusher 启动行情数据推送
func StartMarketDataPusher() {
	ticker := time.NewTicker(1 * time.Second) // 每秒推送一次
	defer ticker.Stop()

	for range ticker.C {
		// 推送主要交易对的行情数据
		symbols := []string{"BTC/USDT", "ETH/USDT", "BNB/USDT"}
		for _, symbol := range symbols {
			go SendMarketData(symbol)
		}
	}
}

// MarshalJSON 自定义JSON序列化
func (m MarketData) MarshalJSON() ([]byte, error) {
	type Alias MarketData
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(&m),
	})
}
