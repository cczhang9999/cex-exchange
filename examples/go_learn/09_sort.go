package main

import (
	"fmt"
	"sort"
)

// 示例 9: 排序和查找
// 订单簿(orderbook)需要按价格排序

// OrderInBook 订单簿中的订单
type OrderInBook struct {
	Price  float64
	Amount float64
	UserID uint64
}

// PriceLevel 价格档位（聚合多个相同价格的订单）
type PriceLevel struct {
	Price  float64
	Amount float64
	Orders []OrderInBook
}

func main() {
	fmt.Println("=== Go 学习示例 9: 排序和查找 ===")

	// 1. 对订单簿价格排序（买单：价格降序，卖单：价格升序）
	buyOrders := []OrderInBook{
		{Price: 50000.0, Amount: 1.5, UserID: 1},
		{Price: 50100.0, Amount: 2.0, UserID: 2},
		{Price: 49900.0, Amount: 0.5, UserID: 3},
		{Price: 50050.0, Amount: 1.0, UserID: 4},
	}

	// 买单排序：价格高的优先（降序）
	sort.Slice(buyOrders, func(i, j int) bool {
		return buyOrders[i].Price > buyOrders[j].Price
	})
	fmt.Println("买单排序（价格降序）:")
	for _, o := range buyOrders {
		fmt.Printf("  价格: %.2f, 数量: %.4f, 用户: %d\n", o.Price, o.Amount, o.UserID)
	}

	// 卖单排序：价格低的优先（升序）
	sellOrders := []OrderInBook{
		{Price: 50200.0, Amount: 1.0, UserID: 5},
		{Price: 50000.0, Amount: 0.5, UserID: 6},
		{Price: 50150.0, Amount: 2.5, UserID: 7},
	}

	sort.Slice(sellOrders, func(i, j int) bool {
		return sellOrders[i].Price < sellOrders[j].Price
	})
	fmt.Println("\n卖单排序（价格升序）:")
	for _, o := range sellOrders {
		fmt.Printf("  价格: %.2f, 数量: %.4f, 用户: %d\n", o.Price, o.Amount, o.UserID)
	}

	// 2. 聚合相同价格的订单
	priceLevels := aggregateOrders(buyOrders)
	fmt.Println("\n聚合后的买单价格档位:")
	for _, level := range priceLevels {
		fmt.Printf("  价格: %.2f, 总数量: %.4f, 订单数: %d\n",
			level.Price, level.Amount, len(level.Orders))
	}

	// 3. 二分查找（用于快速定位价格档位）
	prices := []float64{49900, 50000, 50100, 50200}
	target := 50050.0
	idx := sort.SearchFloat64s(prices, target)
	fmt.Printf("\n查找 %.2f 的位置: %d\n", target, idx)

	// 4. 自定义排序器
	type Trade struct {
		Symbol string
		Price  float64
		Time   int64 // 时间戳
	}

	trades := []Trade{
		{"BTC/USDT", 50000, 1000},
		{"BTC/USDT", 50100, 1001},
		{"BTC/USDT", 50000, 1002},
	}

	// 按价格和时间排序
	sort.Slice(trades, func(i, j int) bool {
		if trades[i].Price != trades[j].Price {
			return trades[i].Price < trades[j].Price // 价格升序
		}
		return trades[i].Time < trades[j].Time // 时间升序
	})
	fmt.Println("\n成交记录排序:")
	for _, t := range trades {
		fmt.Printf("  %s @ %.2f @ %d\n", t.Symbol, t.Price, t.Time)
	}
}

// aggregateOrders 聚合相同价格的订单
func aggregateOrders(orders []OrderInBook) []PriceLevel {
	// 按价格排序
	sort.Slice(orders, func(i, j int) bool {
		return orders[i].Price > orders[j].Price
	})

	// 聚合
	levels := []PriceLevel{}
	for _, o := range orders {
		found := false
		for i := range levels {
			if levels[i].Price == o.Price {
				levels[i].Amount += o.Amount
				levels[i].Orders = append(levels[i].Orders, o)
				found = true
				break
			}
		}
		if !found {
			levels = append(levels, PriceLevel{
				Price:  o.Price,
				Amount: o.Amount,
				Orders: []OrderInBook{o},
			})
		}
	}
	return levels
}
