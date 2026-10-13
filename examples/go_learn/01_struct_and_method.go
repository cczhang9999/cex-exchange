package main

import (
	"fmt"
)

// 示例 1: 基础数据结构 - 订单结构体
// 这是交易所中最常见的订单数据结构
type Order struct {
	ID     int64
	UserID uint64
	Symbol string // 交易对，如 "BTC/USDT"
	Side   string // "buy" 或 "sell"
	Type   string // "limit" 或 "market"
	Price  float64
	Amount float64
	Filled float64
	Status string // "open", "partially_filled", "filled", "cancelled"
}

// Order 的方法 - 计算剩余未成交数量
func (o *Order) Remaining() float64 {
	return o.Amount - o.Filled
}

// Order 的方法 - 判断订单是否完全成交
func (o *Order) IsFilled() bool {
	return o.Filled >= o.Amount
}

func main() {
	fmt.Println("=== Go 学习示例 1: 结构体和方法 ===")

	// 创建一个订单
	order := Order{
		ID:     1001,
		UserID: 12345,
		Symbol: "BTC/USDT",
		Side:   "buy",
		Type:   "limit",
		Price:  50000.0,
		Amount: 0.5,
		Filled: 0.2,
		Status: "partially_filled",
	}

	fmt.Printf("订单信息: %+v\n", order)
	fmt.Printf("剩余未成交: %f\n", order.Remaining())
	fmt.Printf("是否完全成交: %v\n", order.IsFilled())
}
