package engine

import (
	"container/heap"
	"context"
	"sync"

	"github.com/gogf/gf/v2/frame/g"
)

// MatchingEngine 撮合引擎
type MatchingEngine struct {
	orderBooks map[string]*OrderBook
	mu         sync.RWMutex
}

var (
	globalEngine *MatchingEngine
	once         sync.Once
)

// GetEngine 获取全局撮合引擎实例
func GetEngine() *MatchingEngine {
	once.Do(func() {
		globalEngine = &MatchingEngine{
			orderBooks: make(map[string]*OrderBook),
		}
	})
	return globalEngine
}

// getOrderBook 获取或创建订单簿
func (e *MatchingEngine) getOrderBook(symbol string) *OrderBook {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	if book, exists := e.orderBooks[symbol]; exists {
		return book
	}
	
	book := NewOrderBook(symbol)
	e.orderBooks[symbol] = book
	return book
}

// Match 撮合订单
func (e *MatchingEngine) Match(ctx context.Context, order *Order) ([]*Trade, error) {
	book := e.getOrderBook(order.Symbol)
	var trades []*Trade
	
	if order.Type == "market" {
		// 市价单：持续撮合直到完全成交或对手盘清空
		trades = e.matchMarketOrder(ctx, book, order)
	} else {
		// 限价单：撮合满足价格条件的订单
		trades = e.matchLimitOrder(ctx, book, order)
	}
	
	// 如果有成交，写入数据库
	if len(trades) > 0 {
		e.saveTrades(ctx, trades)
		e.updateOrdersInDB(ctx, order, trades)
	}
	
	// 如果订单未完全成交且为限价单，加入订单簿
	if !order.IsFilled() && order.Type == "limit" {
		book.AddOrder(order)
	}
	
	return trades, nil
}

// matchMarketOrder 撮合市价单
func (e *MatchingEngine) matchMarketOrder(ctx context.Context, book *OrderBook, order *Order) []*Trade {
	var trades []*Trade
	
	// 获取对手盘
	var counterHeap *OrderHeap
	if order.Side == "buy" {
		counterHeap = book.SellOrders
	} else {
		counterHeap = book.BuyOrders
	}
	
	book.mu.Lock()
	defer book.mu.Unlock()
	
	// 持续撮合直到完全成交或对手盘清空
	for counterHeap.Len() > 0 && !order.IsFilled() {
		counterOrder := counterHeap.orders[0]
		
		// 计算成交数量
		tradeAmount := min(order.Remaining(), counterOrder.Remaining())
		
		// 创建成交记录
		trade := e.createTrade(order, counterOrder, counterOrder.Price, tradeAmount)
		trades = append(trades, trade)
		
		// 更新成交数量
		order.Filled += tradeAmount
		counterOrder.Filled += tradeAmount
		
		// 如果对手单完全成交，从堆中移除
		if counterOrder.IsFilled() {
			heap.Pop(counterHeap)
			delete(book.OrderIndex, counterOrder.ID)
			counterOrder.Status = "filled"
		} else {
			counterOrder.Status = "partially_filled"
		}
	}
	
	// 更新订单状态
	if order.IsFilled() {
		order.Status = "filled"
	} else {
		order.Status = "partially_filled"
	}
	
	return trades
}

// matchLimitOrder 撮合限价单
func (e *MatchingEngine) matchLimitOrder(ctx context.Context, book *OrderBook, order *Order) []*Trade {
	var trades []*Trade
	
	// 获取对手盘
	var counterHeap *OrderHeap
	if order.Side == "buy" {
		counterHeap = book.SellOrders
	} else {
		counterHeap = book.BuyOrders
	}
	
	book.mu.Lock()
	defer book.mu.Unlock()
	
	// 撮合满足价格条件的订单
	for counterHeap.Len() > 0 && !order.IsFilled() {
		counterOrder := counterHeap.orders[0]
		
		// 检查价格是否匹配
		if order.Side == "buy" {
			if order.Price < counterOrder.Price {
				break // 买价低于卖价，无法成交
			}
		} else {
			if order.Price > counterOrder.Price {
				break // 卖价高于买价，无法成交
			}
		}
		
		// 计算成交数量和价格（使用对手盘价格）
		tradeAmount := min(order.Remaining(), counterOrder.Remaining())
		tradePrice := counterOrder.Price
		
		// 创建成交记录
		trade := e.createTrade(order, counterOrder, tradePrice, tradeAmount)
		trades = append(trades, trade)
		
		// 更新成交数量
		order.Filled += tradeAmount
		counterOrder.Filled += tradeAmount
		
		// 如果对手单完全成交，从堆中移除
		if counterOrder.IsFilled() {
			heap.Pop(counterHeap)
			delete(book.OrderIndex, counterOrder.ID)
			counterOrder.Status = "filled"
		} else {
			counterOrder.Status = "partially_filled"
		}
	}
	
	// 更新订单状态
	if order.IsFilled() {
		order.Status = "filled"
	} else if order.Filled > 0 {
		order.Status = "partially_filled"
	} else {
		order.Status = "open"
	}
	
	return trades
}

// createTrade 创建成交记录
func (e *MatchingEngine) createTrade(order1, order2 *Order, price, amount float64) *Trade {
	var buyOrderID, sellOrderID int64
	var buyUserID, sellUserID uint64
	
	if order1.Side == "buy" {
		buyOrderID = order1.ID
		sellOrderID = order2.ID
		buyUserID = order1.UserID
		sellUserID = order2.UserID
	} else {
		buyOrderID = order2.ID
		sellOrderID = order1.ID
		buyUserID = order2.UserID
		sellUserID = order1.UserID
	}
	
	return &Trade{
		BuyOrderID:  buyOrderID,
		SellOrderID: sellOrderID,
		Symbol:      order1.Symbol,
		Price:       price,
		Amount:      amount,
		BuyUserID:   buyUserID,
		SellUserID:  sellUserID,
	}
}

// saveTrades 保存成交记录到数据库
// func (e *MatchingEngine) saveTrades(ctx context.Context, trades []*Trade) {
// 	for _, trade := range trades {
// 		_, err := g.Model("trades").Ctx(ctx).Data(g.Map{
// 			"buy_order_id":  trade.BuyOrderID,
// 			"sell_order_id": trade.SellOrderID,
// 			"symbol":        trade.Symbol,
// 			"price":         trade.Price,
// 			"amount":        trade.Amount,
// 			"buy_user_id":   trade.BuyUserID,
// 			"sell_user_id":  trade.SellUserID,
// 		}).Insert()
		
// 		if err != nil {
// 			g.Log().Errorf(ctx, "保存成交记录失败: %v", err)
// 		}
// 	}
// }

// 真正的批量插入写法
func (e *MatchingEngine) saveTrades(ctx context.Context, trades []*Trade) {
	if len(trades) == 0 {
		return
	}
	
	// 准备批量数据
	data := make([]g.Map, len(trades))
	for i, trade := range trades {
		data[i] = g.Map{
			"buy_order_id":  trade.BuyOrderID,
			"sell_order_id": trade.SellOrderID,
			"symbol":        trade.Symbol,
			"price":         trade.Price,
			"amount":        trade.Amount,
			"buy_user_id":   trade.BuyUserID,
			"sell_user_id":  trade.SellUserID,
		}
	}
	
	// 一次性批量插入
	_, err := g.Model("trades").Ctx(ctx).Data(data).Insert()
	if err != nil {
		g.Log().Errorf(ctx, "批量保存成交记录失败: %v", err)
	}
}

// updateOrdersInDB 更新订单状态到数据库
func (e *MatchingEngine) updateOrdersInDB(ctx context.Context, order *Order, trades []*Trade) {
	// 更新新订单状态
	_, err := g.Model("orders").Ctx(ctx).Where("id", order.ID).Data(g.Map{
		"filled": order.Filled,
		"status": order.Status,
	}).Update()
	
	if err != nil {
		g.Log().Errorf(ctx, "更新订单状态失败: %v", err)
	}
	
	// 更新对手单状态
	affectedOrderIDs := make(map[int64]bool)
	for _, trade := range trades {
		if order.Side == "buy" {
			affectedOrderIDs[trade.SellOrderID] = true
		} else {
			affectedOrderIDs[trade.BuyOrderID] = true
		}
	}
	
	book := e.getOrderBook(order.Symbol)
	for orderID := range affectedOrderIDs {
		if counterOrder, exists := book.OrderIndex[orderID]; exists {
			_, err := g.Model("orders").Ctx(ctx).Where("id", orderID).Data(g.Map{
				"filled": counterOrder.Filled,
				"status": counterOrder.Status,
			}).Update()
			
			if err != nil {
				g.Log().Errorf(ctx, "更新订单状态失败: %v", err)
			}
		}
	}
}

// CancelOrder 取消订单
func (e *MatchingEngine) CancelOrder(symbol string, orderID int64) {
	book := e.getOrderBook(symbol)
	book.RemoveOrder(orderID)
}

// min 辅助函数
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
