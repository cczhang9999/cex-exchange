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

// upgrader upgrades HTTP connections to WebSocket connections.
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源，生产环境应该限制
	},
}

// wsMessage is the message envelope sent to / received from WebSocket clients.
// The JSON structure matches what the frontend websocket.js client expects:
//
//	{"type":"subscribe","symbol":"BTC/USDT"}
//	{"type":"orderbook","symbol":"BTC/USDT","data":{...},"timestamp":1234567890}
type wsMessage struct {
	Type      string      `json:"type"`
	Symbol    string      `json:"symbol,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

// wsClient represents a single WebSocket client connection.
type wsClient struct {
	conn    *websocket.Conn
	send    chan wsMessage
	symbols map[string]bool // subscribed symbols
	mu      sync.Mutex      // protects symbols
	done    chan struct{}   // closed when the client is shutting down
	once    sync.Once
}

// newWSClient creates a new wsClient wrapping the given connection.
func newWSClient(conn *websocket.Conn) *wsClient {
	return &wsClient{
		conn:    conn,
		send:    make(chan wsMessage, 100),
		symbols: make(map[string]bool),
		done:    make(chan struct{}),
	}
}

// subscribe adds a symbol to the client's subscription set.
func (c *wsClient) subscribe(symbol string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.symbols[symbol] = true
}

// unsubscribe removes a symbol from the client's subscription set.
func (c *wsClient) unsubscribe(symbol string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.symbols, symbol)
}

// isSubscribed returns true if the client is subscribed to the given symbol.
func (c *wsClient) isSubscribed(symbol string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.symbols[symbol]
}

// getSymbols returns a snapshot of the client's subscribed symbols.
func (c *wsClient) getSymbols() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	symbols := make([]string, 0, len(c.symbols))
	for s := range c.symbols {
		symbols = append(symbols, s)
	}
	return symbols
}

// sendMsg enqueues a message to the client. It is non-blocking and safe to call
// after the client has been marked as done (the message will be silently dropped).
func (c *wsClient) sendMsg(msg wsMessage) {
	select {
	case c.send <- msg:
	case <-c.done:
	default:
	}
}

// close marks the client as done (idempotent).
func (c *wsClient) close() {
	c.once.Do(func() {
		close(c.done)
	})
}

// wsManager manages all WebSocket clients, registrations, and market data pushing.
type wsManager struct {
	clients    map[*wsClient]bool
	register   chan *wsClient
	unregister chan *wsClient
	mu         sync.RWMutex
	svc        *service.ExchangeService
}

// newWSManager creates and starts a new wsManager.
func newWSManager(svc *service.ExchangeService) *wsManager {
	m := &wsManager{
		clients:    make(map[*wsClient]bool),
		register:   make(chan *wsClient, 100),
		unregister: make(chan *wsClient, 100),
		svc:        svc,
	}
	go m.run()
	go m.pusher()
	return m
}

// run is the main event loop for client registration/unregistration.
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

// clientCount returns the number of currently connected clients.
func (m *wsManager) clientCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.clients)
}

// pushToSubscribers sends a message to all clients subscribed to msg.Symbol.
// If msg.Symbol is empty, the message is sent to all connected clients.
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

// pusher periodically fetches and pushes market data for all subscribed symbols.
func (m *wsManager) pusher() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// Collect all subscribed symbols across all clients
		m.mu.RLock()
		clientList := make([]*wsClient, 0, len(m.clients))
		for c := range m.clients {
			clientList = append(clientList, c)
		}
		m.mu.RUnlock()

		symbols := make(map[string]bool)
		for _, c := range clientList {
			for _, s := range c.getSymbols() {
				symbols[s] = true
			}
		}

		for symbol := range symbols {
			m.pushMarketData(symbol)
		}
	}
}

// pushMarketData fetches the latest market data for a symbol and pushes it
// to all clients subscribed to that symbol.
func (m *wsManager) pushMarketData(symbol string) {
	ctx := context.Background()
	now := time.Now().Unix()

	// --- OrderBook ---
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
				"bids": bids,
				"asks": asks,
			},
			Timestamp: now,
		})
	}

	// --- Recent Trades ---
	if resp, err := m.svc.GetRecentTrades(ctx, &pb.GetRecentTradesRequest{Symbol: symbol, Limit: 10}); err == nil && resp.Success {
		for _, t := range resp.Trades {
			tradeData := map[string]interface{}{
				"symbol":    t.Symbol,
				"price":     t.Price,
				"amount":    t.Amount,
				"side":      t.Side,
				"timestamp": t.Timestamp,
			}
			m.pushToSubscribers(wsMessage{
				Type:      "trade",
				Symbol:    symbol,
				Data:      tradeData,
				Timestamp: now,
			})
		}
	}

	// --- Price Change (Ticker) ---
	if resp, err := m.svc.GetTicker(ctx, &pb.GetTickerRequest{Symbol: symbol}); err == nil && resp.Success && resp.Ticker != nil {
		t := resp.Ticker
		m.pushToSubscribers(wsMessage{
			Type:   "price_change",
			Symbol: symbol,
			Data: map[string]string{
				"symbol":         symbol,
				"price":          t.LastPrice,
				"change":         t.Change_24H,
				"change_percent": parseChangePercent(t.Change_24H),
				"high_24h":       t.High_24H,
				"low_24h":        t.Low_24H,
				"volume_24h":     t.Volume_24H,
			},
			Timestamp: now,
		})
	}

	// --- Kline (latest) ---
	if resp, err := m.svc.GetKlines(ctx, &pb.GetKlinesRequest{Symbol: symbol, Interval: "1m", Limit: 1}); err == nil && resp.Success && len(resp.Klines) > 0 {
		k := resp.Klines[0]
		// For 1m interval, close_time = open_time + 60 seconds
		closeTime := k.Timestamp + 60
		m.pushToSubscribers(wsMessage{
			Type:   "kline",
			Symbol: symbol,
			Data: map[string]interface{}{
				"symbol":     symbol,
				"interval":   "1m",
				"open":       k.Open,
				"high":       k.High,
				"low":        k.Low,
				"close":      k.Close,
				"volume":     k.Volume,
				"open_time":  k.Timestamp,
				"close_time": closeTime,
			},
			Timestamp: now,
		})
	}
}

// parseChangePercent extracts the numeric percentage from a change string
// such as "+2.5%", "-1.0%", or "2.5%".
func parseChangePercent(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "+")
	s = strings.TrimSuffix(s, "%")
	return s
}

// handleWebSocket upgrades the HTTP connection to a WebSocket connection,
// registers the client, sends a welcome message, and starts the read/write
// pumps.
func (m *wsManager) handleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "WebSocket upgrade failed"})
		return
	}

	client := newWSClient(conn)
	m.register <- client

	// Send welcome message
	client.sendMsg(wsMessage{
		Type:      "welcome",
		Data:      "WebSocket连接成功",
		Timestamp: time.Now().Unix(),
	})

	// Start write pump (reads from client.send and writes to conn)
	go m.writePump(client)

	// Read loop
	defer func() {
		client.close()
		m.unregister <- client
		conn.Close()
	}()

	for {
		var msg map[string]interface{}
		if err := conn.ReadJSON(&msg); err != nil {
			return
		}

		msgType, _ := msg["type"].(string)
		switch msgType {
		case "subscribe":
			symbol, _ := msg["symbol"].(string)
			if symbol == "" {
				continue
			}
			client.subscribe(symbol)
			// Immediately push current data for this symbol to this client
			go m.pushMarketData(symbol)

		case "unsubscribe":
			symbol, _ := msg["symbol"].(string)
			if symbol == "" {
				continue
			}
			client.unsubscribe(symbol)

		case "ping":
			client.sendMsg(wsMessage{
				Type:      "pong",
				Timestamp: time.Now().Unix(),
			})

		default:
			log.Printf("Unknown WebSocket message type: %v", msgType)
		}
	}
}

// writePump sends messages from the client's send channel to the WebSocket
// connection. It exits when the send channel is closed or the client is done.
func (m *wsManager) writePump(client *wsClient) {
	for {
		select {
		case msg, ok := <-client.send:
			if !ok {
				return
			}
			if err := client.conn.WriteJSON(msg); err != nil {
				client.close()
				return
			}
		case <-client.done:
			return
		}
	}
}

// getStatus handles the /ws/status endpoint.
func (m *wsManager) getStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"connected_clients": m.clientCount(),
			"status":            "running",
		},
	})
}
