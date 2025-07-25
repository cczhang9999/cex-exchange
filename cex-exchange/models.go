package main

import (
	"time"
)

// 用户表结构体
type User struct {
	ID        uint64    `gorm:"primaryKey;column:id" json:"id"`         // 用户ID，主键
	Username  string    `gorm:"unique;column:username" json:"username"` // 用户名，唯一
	Password  string    `gorm:"column:password_hash" json:"-"`          // 密码哈希，json不返回
	Email     string    `gorm:"unique;column:email" json:"email"`       // 邮箱，唯一
	Phone     string    `gorm:"unique;column:phone" json:"phone"`       // 手机号，唯一
	Status    int8      `gorm:"column:status" json:"status"`            // 用户状态
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`    // 创建时间
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`    // 更新时间
}

// 指定表名为users
func (User) TableName() string {
	return "users"
}

// 资金账户表结构体
type Account struct {
	ID        uint64    `gorm:"primaryKey;column:id" json:"id"`      // 账户ID，主键
	UserID    uint64    `gorm:"column:user_id" json:"user_id"`       // 所属用户ID
	Asset     string    `gorm:"column:asset" json:"asset"`           // 币种
	Balance   string    `gorm:"column:balance" json:"balance"`       // 可用余额，string防止精度丢失
	Frozen    string    `gorm:"column:frozen" json:"frozen"`         // 冻结余额
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"` // 创建时间
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"` // 更新时间
}

// 指定表名为accounts
func (Account) TableName() string { return "accounts" }

// 资金流水表结构体
type AccountFlow struct {
	ID         uint64    `gorm:"primaryKey;column:id" json:"id"`        // 流水ID，主键
	UserID     uint64    `gorm:"column:user_id" json:"user_id"`         // 用户ID
	AccountID  uint64    `gorm:"column:account_id" json:"account_id"`   // 账户ID
	Asset      string    `gorm:"column:asset" json:"asset"`             // 币种
	ChangeType string    `gorm:"column:change_type" json:"change_type"` // 变动类型
	Amount     string    `gorm:"column:amount" json:"amount"`           // 变动金额
	Balance    string    `gorm:"column:balance" json:"balance"`         // 变动后余额
	RefID      *uint64   `gorm:"column:ref_id" json:"ref_id"`           // 关联业务ID
	Remark     string    `gorm:"column:remark" json:"remark"`           // 备注
	CreatedAt  time.Time `gorm:"column:created_at" json:"created_at"`   // 创建时间
}

// 指定表名为account_flows
func (AccountFlow) TableName() string { return "account_flows" }

// 订单表结构体
type Order struct {
	ID        uint64    `gorm:"primaryKey;column:id" json:"id"`      // 订单ID，主键
	UserID    uint64    `gorm:"column:user_id" json:"user_id"`       // 下单用户ID
	Symbol    string    `gorm:"column:symbol" json:"symbol"`         // 交易对
	Side      string    `gorm:"column:side" json:"side"`             // 买卖方向
	Type      string    `gorm:"column:type" json:"type"`             // 订单类型
	Price     string    `gorm:"column:price" json:"price"`           // 委托价格
	Amount    string    `gorm:"column:amount" json:"amount"`         // 委托数量
	Filled    string    `gorm:"column:filled" json:"filled"`         // 已成交数量
	Status    string    `gorm:"column:status" json:"status"`         // 订单状态
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"` // 创建时间
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"` // 更新时间
}

// 指定表名为orders
func (Order) TableName() string { return "orders" }

// 成交表结构体
type Trade struct {
	ID          uint64    `gorm:"primaryKey;column:id" json:"id"`            // 成交ID，主键
	BuyOrderID  uint64    `gorm:"column:buy_order_id" json:"buy_order_id"`   // 买单ID
	SellOrderID uint64    `gorm:"column:sell_order_id" json:"sell_order_id"` // 卖单ID
	Symbol      string    `gorm:"column:symbol" json:"symbol"`               // 交易对
	Price       string    `gorm:"column:price" json:"price"`                 // 成交价格
	Amount      string    `gorm:"column:amount" json:"amount"`               // 成交数量
	BuyUserID   uint64    `gorm:"column:buy_user_id" json:"buy_user_id"`     // 买方用户ID
	SellUserID  uint64    `gorm:"column:sell_user_id" json:"sell_user_id"`   // 卖方用户ID
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`       // 成交时间
}

// 指定表名为trades
func (Trade) TableName() string { return "trades" }

// K线表结构体
type Kline struct {
	ID        uint64 `gorm:"primaryKey;column:id" json:"id"`      // K线ID，主键
	Symbol    string `gorm:"column:symbol" json:"symbol"`         // 交易对
	Interval  string `gorm:"column:interval" json:"interval"`     // K线周期
	Open      string `gorm:"column:open" json:"open"`             // 开盘价
	High      string `gorm:"column:high" json:"high"`             // 最高价
	Low       string `gorm:"column:low" json:"low"`               // 最低价
	Close     string `gorm:"column:close" json:"close"`           // 收盘价
	Volume    string `gorm:"column:volume" json:"volume"`         // 成交量
	OpenTime  int64  `gorm:"column:open_time" json:"open_time"`   // K线开始时间
	CloseTime int64  `gorm:"column:close_time" json:"close_time"` // K线结束时间
}

// 指定表名为klines
func (Kline) TableName() string { return "klines" }
