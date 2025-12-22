package data

import (
	"backend-v2/internal/biz"
	"context"

	"github.com/google/wire"
	"gorm.io/gorm"
)

var OrderProviderSet = wire.NewSet(NewOrderRepo)

// Order GORM Model
type Order struct {
	gorm.Model
	UserID uint64
	Symbol string
	Side   string
	Type   string
	Price  float64
	Amount float64
	Filled float64
	Status string
}

type orderRepo struct {
	data *Data
}

func NewOrderRepo(data *Data) biz.OrderRepo {
	return &orderRepo{
		data: data,
	}
}

func (r *orderRepo) Save(ctx context.Context, order *biz.Order) (*biz.Order, error) {
	o := &Order{
		UserID: order.UserID,
		Symbol: order.Symbol,
		Side:   string(order.Side),
		Type:   string(order.Type),
		Price:  order.Price,
		Amount: order.Amount,
		Filled: order.Filled,
		Status: string(order.Status),
	}

	if err := r.data.db.WithContext(ctx).Create(o).Error; err != nil {
		return nil, err
	}

	order.ID = uint64(o.ID)
	return order, nil
}

func (r *orderRepo) FindByID(ctx context.Context, id uint64) (*biz.Order, error) {
	var order Order
	if err := r.data.db.WithContext(ctx).Where("id = ?", id).First(&order).Error; err != nil {
		return nil, err
	}

	return &biz.Order{
		ID:        uint64(order.ID),
		UserID:    order.UserID,
		Symbol:    order.Symbol,
		Side:      biz.OrderSide(order.Side),
		Type:      biz.OrderType(order.Type),
		Price:     order.Price,
		Amount:    order.Amount,
		Filled:    order.Filled,
		Status:    biz.OrderStatus(order.Status),
		CreatedAt: order.CreatedAt,
		UpdatedAt: order.UpdatedAt,
	}, nil
}

func (r *orderRepo) FindByUserID(ctx context.Context, userID uint64) ([]*biz.Order, error) {
	var orders []Order
	if err := r.data.db.WithContext(ctx).Where("user_id = ?", userID).Find(&orders).Error; err != nil {
		return nil, err
	}

	res := make([]*biz.Order, 0, len(orders))
	for _, o := range orders {
		res = append(res, &biz.Order{
			ID:        uint64(o.ID),
			UserID:    o.UserID,
			Symbol:    o.Symbol,
			Side:      biz.OrderSide(o.Side),
			Type:      biz.OrderType(o.Type),
			Price:     o.Price,
			Amount:    o.Amount,
			Filled:    o.Filled,
			Status:    biz.OrderStatus(o.Status),
			CreatedAt: o.CreatedAt,
			UpdatedAt: o.UpdatedAt,
		})
	}

	return res, nil
}

func (r *orderRepo) UpdateStatus(ctx context.Context, id uint64, status biz.OrderStatus) error {
	return r.data.db.WithContext(ctx).Model(&Order{}).Where("id = ?", id).Update("status", string(status)).Error
}

func (r *orderRepo) UpdateFilled(ctx context.Context, id uint64, filled float64) error {
	return r.data.db.WithContext(ctx).Model(&Order{}).Where("id = ?", id).Update("filled", filled).Error
}

func (r *orderRepo) FindOpenOrders(ctx context.Context, symbol string) ([]*biz.Order, error) {
	var orders []Order
	if err := r.data.db.WithContext(ctx).Where("symbol = ? AND status IN ?", symbol, []string{"open", "partially_filled"}).Find(&orders).Error; err != nil {
		return nil, err
	}

	res := make([]*biz.Order, 0, len(orders))
	for _, o := range orders {
		res = append(res, &biz.Order{
			ID:        uint64(o.ID),
			UserID:    o.UserID,
			Symbol:    o.Symbol,
			Side:      biz.OrderSide(o.Side),
			Type:      biz.OrderType(o.Type),
			Price:     o.Price,
			Amount:    o.Amount,
			Filled:    o.Filled,
			Status:    biz.OrderStatus(o.Status),
			CreatedAt: o.CreatedAt,
			UpdatedAt: o.UpdatedAt,
		})
	}

	return res, nil
}
