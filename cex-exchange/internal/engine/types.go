package engine

import "time"

// Order 内存中的订单结构
type Order struct {
	ID        int64     // 订单ID
	UserID    uint64    // 用户ID
	Symbol    string    // 交易对
	Side      string    // buy/sell
	Type      string    // limit/market
	Price     float64   // 价格
	Amount    float64   // 数量
	Filled    float64   // 已成交数量
	Status    string    // open/partially_filled/filled/cancelled
	CreatedAt time.Time // 创建时间
}

// Trade 成交记录
type Trade struct {
	BuyOrderID  int64   // 买单ID
	SellOrderID int64   // 卖单ID
	Symbol      string  // 交易对
	Price       float64 // 成交价格
	Amount      float64 // 成交数量
	BuyUserID   uint64  // 买方用户ID
	SellUserID  uint64  // 卖方用户ID
}

// Remaining 获取剩余未成交数量
func (o *Order) Remaining() float64 {
	return o.Amount - o.Filled
}

// IsFilled 是否完全成交
func (o *Order) IsFilled() bool {
	return o.Filled >= o.Amount
}
