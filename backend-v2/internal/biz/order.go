package biz

import (
	"context"
	"github.com/google/wire"
	"time"
)

// 更新ProviderSet，包含OrderUsecase
var ProviderSet = wire.NewSet(NewUserUsecase, NewOrderUsecase)

// 定义订单方向
type OrderSide string

const (
	OrderSideBuy  OrderSide = "buy"
	OrderSideSell OrderSide = "sell"
)

// 定义订单类型
type OrderType string

const (
	OrderTypeLimit  OrderType = "limit"
	OrderTypeMarket OrderType = "market"
)

// 定义订单状态
type OrderStatus string

const (
	OrderStatusOpen            OrderStatus = "open"
	OrderStatusPartiallyFilled OrderStatus = "partially_filled"
	OrderStatusFilled          OrderStatus = "filled"
	OrderStatusCancelled       OrderStatus = "cancelled"
)

// Order 定义订单业务模型
type Order struct {
	ID        uint64      `json:"id"`
	UserID    uint64      `json:"user_id"`
	Symbol    string      `json:"symbol"`
	Side      OrderSide   `json:"side"`
	Type      OrderType   `json:"type"`
	Price     float64     `json:"price,omitempty"`
	Amount    float64     `json:"amount"`
	Filled    float64     `json:"filled"`
	Status    OrderStatus `json:"status"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
	DeletedAt *time.Time  `json:"deleted_at,omitempty"`
}

// OrderRepo 定义订单存储接口
type OrderRepo interface {
	Save(ctx context.Context, order *Order) (*Order, error)
	FindByID(ctx context.Context, id uint64) (*Order, error)
	FindByUserID(ctx context.Context, userID uint64) ([]*Order, error)
	UpdateStatus(ctx context.Context, id uint64, status OrderStatus) error
	UpdateFilled(ctx context.Context, id uint64, filled float64) error
	FindOpenOrders(ctx context.Context, symbol string) ([]*Order, error)
	// 关联查询方法
	FindByUserIDWithUser(ctx context.Context, userID uint64) ([]*Order, error)
	FindOrdersWithUserInfo(ctx context.Context, userID uint64) (map[uint64]*Order, error)
	FindOrdersWithUserUsingJoins(ctx context.Context, symbol string) ([]*Order, error)
}

// OrderUsecase 定义订单业务逻辑
type OrderUsecase struct {
	repo OrderRepo
}

// NewOrderUsecase 创建订单业务逻辑实例
func NewOrderUsecase(repo OrderRepo) *OrderUsecase {
	return &OrderUsecase{
		repo: repo,
	}
}

// CreateOrder 创建订单
func (uc *OrderUsecase) CreateOrder(ctx context.Context, userID uint64, symbol string, side OrderSide, orderType OrderType, price, amount float64) (*Order, error) {
	order := &Order{
		UserID:    userID,
		Symbol:    symbol,
		Side:      side,
		Type:      orderType,
		Price:     price,
		Amount:    amount,
		Filled:    0,
		Status:    OrderStatusOpen,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	return uc.repo.Save(ctx, order)
}

// GetOrder 获取订单详情
func (uc *OrderUsecase) GetOrder(ctx context.Context, id uint64) (*Order, error) {
	return uc.repo.FindByID(ctx, id)
}

// GetUserOrders 获取用户订单列表
func (uc *OrderUsecase) GetUserOrders(ctx context.Context, userID uint64) ([]*Order, error) {
	return uc.repo.FindByUserID(ctx, userID)
}

// GetUserOrdersWithUser 获取用户订单列表及用户信息
func (uc *OrderUsecase) GetUserOrdersWithUser(ctx context.Context, userID uint64) ([]*Order, error) {
	return uc.repo.FindByUserIDWithUser(ctx, userID)
}

// GetOrdersWithUserInfo 获取订单列表及用户信息
func (uc *OrderUsecase) GetOrdersWithUserInfo(ctx context.Context, userID uint64) (map[uint64]*Order, error) {
	return uc.repo.FindOrdersWithUserInfo(ctx, userID)
}

// GetOrdersWithUserByJoins 获取订单列表及用户信息（使用Joins）
func (uc *OrderUsecase) GetOrdersWithUserByJoins(ctx context.Context, symbol string) ([]*Order, error) {
	return uc.repo.FindOrdersWithUserUsingJoins(ctx, symbol)
}

// CancelOrder 取消订单
func (uc *OrderUsecase) CancelOrder(ctx context.Context, id uint64) error {
	return uc.repo.UpdateStatus(ctx, id, OrderStatusCancelled)
}
