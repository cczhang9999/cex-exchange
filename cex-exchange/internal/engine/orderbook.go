package engine

import (
	"container/heap"
	"sync"
)

// OrderBook 订单簿
type OrderBook struct {
	Symbol    string
	BuyOrders *OrderHeap // 买单堆（价格从高到低）
	SellOrders *OrderHeap // 卖单堆（价格从低到高）
	OrderIndex map[int64]*Order // 订单索引
	mu         sync.RWMutex
}

// NewOrderBook 创建订单簿
func NewOrderBook(symbol string) *OrderBook {
	buyHeap := &OrderHeap{isBuy: true}
	sellHeap := &OrderHeap{isBuy: false}
	heap.Init(buyHeap)
	heap.Init(sellHeap)
	
	return &OrderBook{
		Symbol:     symbol,
		BuyOrders:  buyHeap,
		SellOrders: sellHeap,
		OrderIndex: make(map[int64]*Order),
	}
}

// AddOrder 添加订单到订单簿
func (ob *OrderBook) AddOrder(order *Order) {
	ob.mu.Lock()
	defer ob.mu.Unlock()
	
	ob.OrderIndex[order.ID] = order
	
	if order.Side == "buy" {
		heap.Push(ob.BuyOrders, order)
	} else {
		heap.Push(ob.SellOrders, order)
	}
}

// RemoveOrder 从订单簿移除订单
func (ob *OrderBook) RemoveOrder(orderID int64) {
	ob.mu.Lock()
	defer ob.mu.Unlock()
	
	delete(ob.OrderIndex, orderID)
}

// GetBestBid 获取最优买价
func (ob *OrderBook) GetBestBid() *Order {
	ob.mu.RLock()
	defer ob.mu.RUnlock()
	
	if ob.BuyOrders.Len() == 0 {
		return nil
	}
	return ob.BuyOrders.orders[0]
}

// GetBestAsk 获取最优卖价
func (ob *OrderBook) GetBestAsk() *Order {
	ob.mu.RLock()
	defer ob.mu.RUnlock()
	
	if ob.SellOrders.Len() == 0 {
		return nil
	}
	return ob.SellOrders.orders[0]
}

// OrderHeap 订单堆
type OrderHeap struct {
	orders []*Order
	isBuy  bool // true=买单（价格从高到低），false=卖单（价格从低到高）
}

func (h OrderHeap) Len() int { return len(h.orders) }

func (h OrderHeap) Less(i, j int) bool {
	// 价格优先
	if h.orders[i].Price != h.orders[j].Price {
		if h.isBuy {
			return h.orders[i].Price > h.orders[j].Price // 买单：高价优先
		}
		return h.orders[i].Price < h.orders[j].Price // 卖单：低价优先
	}
	// 时间优先（ID越小越早）
	return h.orders[i].ID < h.orders[j].ID
}

func (h OrderHeap) Swap(i, j int) {
	h.orders[i], h.orders[j] = h.orders[j], h.orders[i]
}

func (h *OrderHeap) Push(x interface{}) {
	h.orders = append(h.orders, x.(*Order))
}

func (h *OrderHeap) Pop() interface{} {
	old := h.orders
	n := len(old)
	order := old[n-1]
	h.orders = old[0 : n-1]
	return order
}
