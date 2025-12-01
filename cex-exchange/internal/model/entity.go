package model

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// User 用户表
type User struct {
	ID        uint64      `json:"id"`
	Username  string      `json:"username"`
	Password  string      `json:"-"` // 不返回给前端
	Email     string      `json:"email"`
	CreatedAt *gtime.Time `json:"created_at"`
	UpdatedAt *gtime.Time `json:"updated_at"`
}

// Account 账户表
type Account struct {
	ID        uint64      `json:"id"`
	UserID    uint64      `json:"user_id"`
	Asset     string      `json:"asset"`
	Balance   string      `json:"balance"`
	Frozen    string      `json:"frozen"`
	CreatedAt *gtime.Time `json:"created_at"`
	UpdatedAt *gtime.Time `json:"updated_at"`
}

// Order 订单表
type Order struct {
	ID        uint64      `json:"id"`
	UserID    uint64      `json:"user_id"`
	Symbol    string      `json:"symbol"`
	Side      string      `json:"side"`   // buy/sell
	Type      string      `json:"type"`   // limit/market
	Price     string      `json:"price"`
	Amount    string      `json:"amount"`
	Filled    string      `json:"filled"`
	Status    string      `json:"status"` // open/filled/cancelled
	CreatedAt *gtime.Time `json:"created_at"`
	UpdatedAt *gtime.Time `json:"updated_at"`
}

// Trade 成交记录表
type Trade struct {
	ID          uint64      `json:"id"`
	BuyOrderID  uint64      `json:"buy_order_id"`
	SellOrderID uint64      `json:"sell_order_id"`
	Symbol      string      `json:"symbol"`
	Price       string      `json:"price"`
	Amount      string      `json:"amount"`
	BuyUserID   uint64      `json:"buy_user_id"`
	SellUserID  uint64      `json:"sell_user_id"`
	CreatedAt   *gtime.Time `json:"created_at"`
}

// AccountFlow 资金流水表
type AccountFlow struct {
	ID         uint64      `json:"id"`
	UserID     uint64      `json:"user_id"`
	AccountID  uint64      `json:"account_id"`
	Asset      string      `json:"asset"`
	ChangeType string      `json:"change_type"` // deposit/withdraw/trade/admin_add
	Amount     string      `json:"amount"`
	Balance    string      `json:"balance"`
	Remark     string      `json:"remark"`
	CreatedAt  *gtime.Time `json:"created_at"`
}

// Kline K线数据表
type Kline struct {
	ID        uint64      `json:"id"`
	Symbol    string      `json:"symbol"`
	Period    string      `json:"period"` // 1m/5m/15m/1h/4h/1d
	OpenTime  int64       `json:"open_time"`
	CloseTime int64       `json:"close_time"`
	Open      string      `json:"open"`
	High      string      `json:"high"`
	Low       string      `json:"low"`
	Close     string      `json:"close"`
	Volume    string      `json:"volume"`
	Timestamp int64       `json:"timestamp"`
	CreatedAt *gtime.Time `json:"created_at"`
	UpdatedAt *gtime.Time `json:"updated_at"`
}

// OrderBook 订单簿
type OrderBook struct {
	Symbol string       `json:"symbol"`
	Bids   []PriceLevel `json:"bids"` // 买单
	Asks   []PriceLevel `json:"asks"` // 卖单
}

// PriceLevel 价格档位
type PriceLevel struct {
	Price  string `json:"price"`
	Amount string `json:"amount"`
}

// Ticker 行情
type Ticker struct {
	Symbol     string `json:"symbol"`
	LastPrice  string `json:"last_price"`
	High24h    string `json:"high_24h"`
	Low24h     string `json:"low_24h"`
	Volume24h  string `json:"volume_24h"`
	Change24h  string `json:"change_24h"`
}

// TradeRecord 成交记录
type TradeRecord struct {
	ID        uint64 `json:"id"`
	Symbol    string `json:"symbol"`
	Price     string `json:"price"`
	Amount    string `json:"amount"`
	Side      string `json:"side"` // buy/sell
	Timestamp int64  `json:"timestamp"`
}
