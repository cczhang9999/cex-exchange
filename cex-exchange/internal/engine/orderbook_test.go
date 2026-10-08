package engine

import (
	"testing"
	"time"
)

// newTestOrder 构造测试订单：只填断言需要的字段，避免每个用例重复十几行结构体。
func newTestOrder(id int64, side string, price, amount float64) *Order {
	return &Order{
		ID:        id,
		UserID:    uint64(id),
		Symbol:    "BTC/USDT",
		Side:      side,
		Type:      "limit",
		Price:     price,
		Amount:    amount,
		Status:    "open",
		CreatedAt: time.Unix(id, 0),
	}
}

// TestOrderBookBestPrice 表驱动用例：价格优先 + 时间优先。
func TestOrderBookBestPrice(t *testing.T) {
	tests := []struct {
		name    string
		orders  []*Order
		wantBid int64 // 期望的最优买单 ID，0 表示没有
		wantAsk int64 // 期望的最优卖单 ID，0 表示没有
	}{
		{
			name:   "空订单簿没有最优价",
			orders: nil,
		},
		{
			name: "买单取最高价",
			orders: []*Order{
				newTestOrder(1, "buy", 100, 1),
				newTestOrder(2, "buy", 105, 1),
				newTestOrder(3, "buy", 99, 1),
			},
			wantBid: 2,
		},
		{
			name: "卖单取最低价",
			orders: []*Order{
				newTestOrder(4, "sell", 110, 1),
				newTestOrder(5, "sell", 108, 1),
				newTestOrder(6, "sell", 120, 1),
			},
			wantAsk: 5,
		},
		{
			name: "同价时间优先（ID 小者优先）",
			orders: []*Order{
				newTestOrder(8, "buy", 100, 1),
				newTestOrder(7, "buy", 100, 1),
			},
			wantBid: 7,
		},
		{
			name: "买卖单互不干扰",
			orders: []*Order{
				newTestOrder(9, "sell", 108, 1),
				newTestOrder(10, "buy", 100, 1),
			},
			wantBid: 10,
			wantAsk: 9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ob := NewOrderBook("BTC/USDT")
			for _, o := range tt.orders {
				ob.AddOrder(o)
			}

			gotBid := int64(0)
			if best := ob.GetBestBid(); best != nil {
				gotBid = best.ID
			}
			if gotBid != tt.wantBid {
				t.Errorf("GetBestBid() = 订单 %d, 期望 %d", gotBid, tt.wantBid)
			}

			gotAsk := int64(0)
			if best := ob.GetBestAsk(); best != nil {
				gotAsk = best.ID
			}
			if gotAsk != tt.wantAsk {
				t.Errorf("GetBestAsk() = 订单 %d, 期望 %d", gotAsk, tt.wantAsk)
			}
		})
	}
}

// TestOrderRemainingAndIsFilled 纯函数用例：一个用例一个断言，失败信息带输入。
func TestOrderRemainingAndIsFilled(t *testing.T) {
	tests := []struct {
		name         string
		amount       float64
		filled       float64
		wantRemain   float64
		wantIsFilled bool
	}{
		{name: "完全未成交", amount: 2, filled: 0, wantRemain: 2, wantIsFilled: false},
		{name: "部分成交", amount: 2, filled: 0.5, wantRemain: 1.5, wantIsFilled: false},
		{name: "完全成交", amount: 2, filled: 2, wantRemain: 0, wantIsFilled: true},
		{name: "超量成交按已成交处理", amount: 2, filled: 2.5, wantRemain: -0.5, wantIsFilled: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &Order{Amount: tt.amount, Filled: tt.filled}

			if got := o.Remaining(); got != tt.wantRemain {
				t.Errorf("Remaining(amount=%v, filled=%v) = %v, 期望 %v",
					tt.amount, tt.filled, got, tt.wantRemain)
			}
			if got := o.IsFilled(); got != tt.wantIsFilled {
				t.Errorf("IsFilled(amount=%v, filled=%v) = %v, 期望 %v",
					tt.amount, tt.filled, got, tt.wantIsFilled)
			}
		})
	}
}

// TestOrderBookRemoveOrder 验证移出索引后的可见行为（当前实现不移除堆内元素）。
func TestOrderBookRemoveOrder(t *testing.T) {
	ob := NewOrderBook("BTC/USDT")
	o := newTestOrder(1, "buy", 100, 1)
	ob.AddOrder(o)

	ob.RemoveOrder(o.ID)

	if _, ok := ob.OrderIndex[o.ID]; ok {
		t.Errorf("RemoveOrder 后 OrderIndex 中仍存在订单 %d", o.ID)
	}
}
