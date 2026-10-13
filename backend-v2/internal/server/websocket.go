package server

import (
	"context"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	pb "backend-v2/api/proto"
	"backend-v2/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// upgrader 将 HTTP 连接升级为 WebSocket 连接的升级器
// 配置 CheckOrigin 返回 true 允许所有来源，生产环境建议限制来源以防止跨站 WebSocket 攻击
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源，生产环境应该限制
	},
}

// wsMessage 是与 WebSocket 客户端收发消息的信封结构
// JSON 结构与前端 websocket.js 客户端期望的格式匹配:
//
//	发送订阅: {"type":"subscribe","symbol":"BTC/USDT"}
//	接收订单簿: {"type":"orderbook","symbol":"BTC/USDT","data":{...},"timestamp":1234567890}
type wsMessage struct {
	Type      string      `json:"type"`             // 消息类型: subscribe, unsubscribe, ping, orderbook, trade, price_change, kline, welcome, pong
	Symbol    string      `json:"symbol,omitempty"` // 交易对符号，如 "BTC/USDT"，可选
	Data      interface{} `json:"data,omitempty"`   // 消息数据负载，可根据 type 不同而类型不同
	Timestamp int64       `json:"timestamp"`        // Unix 时间戳（秒）
}

// wsClient 代表一个 WebSocket 客户端连接
// 每个客户端维护自己的订阅 symbol 集合和消息发送通道
type wsClient struct {
	conn    *websocket.Conn // 底层的 WebSocket 连接
	send    chan wsMessage  // 消息发送通道，带缓冲避免阻塞
	symbols map[string]bool // 该客户端订阅的交易对集合
	mu      sync.Mutex      // 保护 symbols 字段的互斥锁
	done    chan struct{}   // 关闭通道，当客户端关闭时会被关闭
	once    sync.Once       // 确保 done 通道只被关闭一次
}

// newWSClient 创建一个新的 wsClient，包装给定的 WebSocket 连接
// 初始化时创建带 100 缓冲的消息通道和空的订阅集合
func newWSClient(conn *websocket.Conn) *wsClient {
	return &wsClient{
		conn:    conn,
		send:    make(chan wsMessage, 100), // 缓冲 100 条消息，防止快速发送时阻塞
		symbols: make(map[string]bool),
		done:    make(chan struct{}),
	}
}

// subscribe 将一个交易对添加到客户端的订阅集合中
// 使用互斥锁保护 symbols 映射，确保并发安全
func (c *wsClient) subscribe(symbol string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.symbols[symbol] = true
}

// unsubscribe 从客户端的订阅集合中移除一个交易对
// 使用互斥锁保护 symbols 映射，确保并发安全
func (c *wsClient) unsubscribe(symbol string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.symbols, symbol)
}

// isSubscribed 检查客户端是否订阅了给定的交易对
// 使用互斥锁保护 symbols 映射，确保并发安全
func (c *wsClient) isSubscribed(symbol string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.symbols[symbol]
}

// getSymbols 返回客户端订阅的交易对集合的快照
// 复制一份 symbols 避免返回后被外部修改
// 使用互斥锁保护 symbols 映射，确保并发安全
func (c *wsClient) getSymbols() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	symbols := make([]string, 0, len(c.symbols))
	for s := range c.symbols {
		symbols = append(symbols, s)
	}
	return symbols
}

// sendMsg 将消息加入发送队列。它是非阻塞的，并且在客户端标记为完成 后调用也是安全的
// （消息会被静默丢弃）。使用 select 实现非阻塞发送
func (c *wsClient) sendMsg(msg wsMessage) {
	select {
	case c.send <- msg:
	case <-c.done: // 客户端已关闭，不再发送
	default: // 通道满时静默丢弃，避免阻塞
	}
}

// close 将客户端标记为完成（幂等操作）
// 使用 sync.Once 确保 done 通道只被关闭一次
func (c *wsClient) close() {
	c.once.Do(func() {
		close(c.done)
	})
}

// wsManager 管理所有 WebSocket 客户端、注册/注销以及市场数据推送
// 使用两个通道处理客户端的注册和注销，避免直接操作共享map
type wsManager struct {
	clients    map[*wsClient]bool       // 所有已连接客户端的集合
	register   chan *wsClient           // 客户端注册通道
	unregister chan *wsClient           // 客户端注销通道
	mu         sync.RWMutex             // 保护 clients 字段的读写锁
	svc        *service.ExchangeService // 交易所服务，用于获取市场数据
}

// newWSManager 创建并启动一个新的 wsManager
// 内部启动两个 goroutine：run() 处理客户端管理，pusher() 定期推送市场数据
func newWSManager(svc *service.ExchangeService) *wsManager {
	m := &wsManager{
		clients:    make(map[*wsClient]bool),
		register:   make(chan *wsClient, 100), // 缓冲通道避免阻塞
		unregister: make(chan *wsClient, 100),
		svc:        svc,
	}
	go m.run()    // 启动客户端管理循环
	go m.pusher() // 启动市场数据推送循环
	return m
}

// run 是客户端注册/注销的主事件循环
// 通过监听 register 和 unregister 通道来处理客户端的连接和断开
func (m *wsManager) run() {
	for {
		select {
		case client := <-m.register:
			m.mu.Lock()
			m.clients[client] = true
			count := len(m.clients)
			m.mu.Unlock()
			log.Printf("WebSocket client connected: %s, total: %d", client.conn.RemoteAddr(), count)

		case client := <-m.unregister:
			m.mu.Lock()
			if _, ok := m.clients[client]; ok {
				delete(m.clients, client)
			}
			count := len(m.clients)
			m.mu.Unlock()
			log.Printf("WebSocket client disconnected: %s, total: %d", client.conn.RemoteAddr(), count)
		}
	}
}

// clientCount 返回当前连接的客户端数量
// 使用读锁，允许并发读取
func (m *wsManager) clientCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.clients)
}

// pushToSubscribers 向订阅了 msg.Symbol 的所有客户端发送消息
// 如果 msg.Symbol 为空，则向所有已连接的客户端发送消息（广播）
// 消息发送是非阻塞的，如果客户端的发送通道满，消息会被静默丢弃
func (m *wsManager) pushToSubscribers(msg wsMessage) {
	m.mu.RLock()
	clients := make([]*wsClient, 0, len(m.clients))
	for c := range m.clients {
		if msg.Symbol == "" || c.isSubscribed(msg.Symbol) {
			clients = append(clients, c)
		}
	}
	m.mu.RUnlock()

	for _, c := range clients {
		c.sendMsg(msg)
	}
}

// pusher 定期获取并推送所有已订阅交易组的市场数据
// 每秒执行一次，遍历所有客户端收集订阅的symbol，然后调用pushMarketData推送数据
// 注意：每个客户端收到的数据可能不同，取决于其订阅的交易对
func (m *wsManager) pusher() {
	ticker := time.NewTicker(1 * time.Second) // 每秒触发一次
	defer ticker.Stop()

	for range ticker.C {
		// 收集所有客户端
		m.mu.RLock()
		clientList := make([]*wsClient, 0, len(m.clients))
		for c := range m.clients {
			clientList = append(clientList, c)
		}
		m.mu.RUnlock()

		// 收集所有订阅的交易对（去重）
		symbols := make(map[string]bool)
		for _, c := range clientList {
			for _, s := range c.getSymbols() {
				symbols[s] = true
			}
		}

		// 为每个交易对推送市场数据
		for symbol := range symbols {
			m.pushMarketData(symbol)
		}
	}
}

// pushMarketData 获取指定交易对的最新市场数据并推送给订阅了该交易对的所有客户端
// 推送的数据类型包括：订单簿( orderbook)、最近成交(trade)、价格变动(price_change)、K线(kline)
// 每种数据类型都会构造成 wsMessage 通过 pushToSubscribers 发送
func (m *wsManager) pushMarketData(symbol string) {
	ctx := context.Background()
	now := time.Now().Unix()

	// --- 订单簿 (OrderBook) ---
	// 获取深度为20的订单簿数据，构造 bids(买盘) 和 asks(卖盘) 数组
	if resp, err := m.svc.GetOrderBook(ctx, &pb.GetOrderBookRequest{Symbol: symbol, Depth: 20}); err == nil && resp.Success {
		bids := make([][]string, 0, len(resp.Bids))
		for _, b := range resp.Bids {
			bids = append(bids, []string{b.Price, b.Amount})
		}
		asks := make([][]string, 0, len(resp.Asks))
		for _, a := range resp.Asks {
			asks = append(asks, []string{a.Price, a.Amount})
		}
		m.pushToSubscribers(wsMessage{
			Type:   "orderbook",
			Symbol: symbol,
			Data: map[string]interface{}{
				"bids": bids, // 买盘队列，按价格降序
				"asks": asks, // 卖盘队列，按价格升序
			},
			Timestamp: now,
		})
	}

	// --- 最近成交 (Recent Trades) ---
	// 获取最近10条成交记录，逐条推送给订阅者
	if resp, err := m.svc.GetRecentTrades(ctx, &pb.GetRecentTradesRequest{Symbol: symbol, Limit: 10}); err == nil && resp.Success {
		for _, t := range resp.Trades {
			tradeData := map[string]interface{}{
				"symbol":    t.Symbol,    // 交易对
				"price":     t.Price,     // 成交价格
				"amount":    t.Amount,    // 成交数量
				"side":      t.Side,      // 成交方向: buy 或 sell
				"timestamp": t.Timestamp, // 成交时间戳
			}
			m.pushToSubscribers(wsMessage{
				Type:      "trade",
				Symbol:    symbol,
				Data:      tradeData,
				Timestamp: now,
			})
		}
	}

	// --- 价格变动 (Ticker / Price Change) ---
	// 获取24小时价格变动信息，包括最新价、涨跌幅、最高价、最低价、24小时成交量
	if resp, err := m.svc.GetTicker(ctx, &pb.GetTickerRequest{Symbol: symbol}); err == nil && resp.Success && resp.Ticker != nil {
		t := resp.Ticker
		m.pushToSubscribers(wsMessage{
			Type:   "price_change",
			Symbol: symbol,
			Data: map[string]string{
				"symbol":         symbol,
				"price":          t.LastPrice,                      // 最新价格
				"change":         t.Change_24H,                     // 24小时涨跌额
				"change_percent": parseChangePercent(t.Change_24H), // 24小时涨跌幅百分比（去掉+号和%号）
				"high_24h":       t.High_24H,                       // 24小时最高价
				"low_24h":        t.Low_24H,                        // 24小时最低价
				"volume_24h":     t.Volume_24H,                     // 24小时成交量
			},
			Timestamp: now,
		})
	}

	// --- K线 (Kline / Candlestick) ---
	// 获取最新一根1分钟K线数据
	// 注意：K线的 close_time = open_time + 60秒（对于1分钟K线）
	if resp, err := m.svc.GetKlines(ctx, &pb.GetKlinesRequest{Symbol: symbol, Interval: "1m", Limit: 1}); err == nil && resp.Success && len(resp.Klines) > 0 {
		k := resp.Klines[0]
		closeTime := k.Timestamp + 60 // 1分钟K线的收盘时间
		m.pushToSubscribers(wsMessage{
			Type:   "kline",
			Symbol: symbol,
			Data: map[string]interface{}{
				"symbol":     symbol,
				"interval":   "1m",        // K线周期，目前固定为1分钟
				"open":       k.Open,      // 开盘价
				"high":       k.High,      // 最高价
				"low":        k.Low,       // 最低价
				"close":      k.Close,     // 收盘价
				"volume":     k.Volume,    // 成交量
				"open_time":  k.Timestamp, // 开盘时间
				"close_time": closeTime,   // 收盘时间
			},
			Timestamp: now,
		})
	}
}

// parseChangePercent 从涨跌字符串中提取数值百分比
// 支持格式: "+2.5%", "-1.0%", "2.5%"（去掉+号和%号后返回 "2.5" 或 "-1.0"）
func parseChangePercent(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "+") // 去掉可能的前缀 +
	s = strings.TrimSuffix(s, "%") // 去掉可能的后缀 %
	return s
}

// handleWebSocket 处理 WebSocket 连接升级、客户端注册、欢迎消息发送以及读写循环启动
// 流程: 1) HTTP -> WebSocket 升级 2) 创建客户端并注册到 manager 3) 发送欢迎消息 4) 启动写入泵 5) 进入读取循环
// 读取循环处理三种消息类型: subscribe(订阅), unsubscribe(取消订阅), ping(心跳)
// 当客户端断开连接时（读取出错或主动关闭），会执行 defer 中的清理逻辑：关闭客户端、注销、从 manager 移除连接
func (m *wsManager) handleWebSocket(c *gin.Context) {
	// 将 HTTP 连接升级为 WebSocket 连接
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "WebSocket upgrade failed"})
		return
	}

	// 创建客户端并注册到 manager
	client := newWSClient(conn)
	m.register <- client

	// 发送欢迎消息，告知客户端连接成功
	client.sendMsg(wsMessage{
		Type:      "welcome",
		Data:      "WebSocket连接成功",
		Timestamp: time.Now().Unix(),
	})

	// 启动写入泵 goroutine，从 client.send 通道读取消息并写入 WebSocket 连接
	go m.writePump(client)

	// 读取循环：持续读取客户端发送的消息并处理
	defer func() {
		client.close()         // 标记客户端为关闭状态
		m.unregister <- client // 从 manager 注销客户端
		conn.Close()           // 关闭 WebSocket 连接
	}()

	for {
		var msg map[string]interface{}
		// 读取客户端发送的 JSON 消息
		if err := conn.ReadJSON(&msg); err != nil {
			return // 读取失败（客户端断开），退出读取循环，触发 defer 清理
		}

		// 解析消息类型
		msgType, _ := msg["type"].(string)
		switch msgType {
		case "subscribe":
			// 订阅交易对：将 symbol 添加到客户端订阅集合，并立即推送该交易对的当前市场数据
			symbol, _ := msg["symbol"].(string)
			if symbol == "" {
				continue
			}
			client.subscribe(symbol)
			// 异步推送当前市场数据，让客户端立即收到数据
			go m.pushMarketData(symbol)

		case "unsubscribe":
			// 取消订阅：从客户端订阅集合中移除 symbol
			symbol, _ := msg["symbol"].(string)
			if symbol == "" {
				continue
			}
			client.unsubscribe(symbol)

		case "ping":
			// 心跳：响应 pong，保持连接活跃
			client.sendMsg(wsMessage{
				Type:      "pong",
				Timestamp: time.Now().Unix(),
			})

		default:
			log.Printf("Unknown WebSocket message type: %v", msgType)
		}
	}
}

// writePump 从客户端的发送通道读取消息并写入 WebSocket 连接
// 这是一个独立的 goroutine，与读取循环并行运行
// 当发送通道关闭（done 通道被关闭）时，writePump 会优雅退出
func (m *wsManager) writePump(client *wsClient) {
	for {
		select {
		case msg, ok := <-client.send:
			if !ok {
				// 发送通道已关闭，说明客户端不再需要发送消息
				return
			}
			// 将消息序列化为 JSON 并写入 WebSocket 连接
			if err := client.conn.WriteJSON(msg); err != nil {
				client.close() // 写入失败，关闭客户端
				return
			}
		case <-client.done:
			// 客户端已被标记为关闭
			return
		}
	}
}

// getStatus 处理 /ws/status 端点，返回当前 WebSocket 服务状态
// 返回信息包括：已连接的客户端数量和服务运行状态
func (m *wsManager) getStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"connected_clients": m.clientCount(), // 当前连接的客户端数量
			"status":            "running",       // 服务状态
		},
	})
}
