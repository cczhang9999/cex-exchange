package dao

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"

	"cex-exchange/internal/model"
)

// OrderDao 订单数据访问对象
type OrderDao struct {
	table string
}

var Order = &OrderDao{
	table: "orders",
}

// Model 获取模型
func (d *OrderDao) Model(ctx context.Context) *gdb.Model {
	return g.Model(d.table).Ctx(ctx)
}

// Create 创建订单
func (d *OrderDao) Create(ctx context.Context, data g.Map) (uint64, error) {
	result, err := d.Model(ctx).Data(data).Insert()
	if err != nil {
		return 0, err
	}
	id, _ := result.LastInsertId()
	return uint64(id), nil
}

// GetByID 根据ID获取订单
func (d *OrderDao) GetByID(ctx context.Context, id uint64) (*model.Order, error) {
	var order model.Order
	err := d.Model(ctx).Where("id", id).Scan(&order)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// ListByUserID 获取用户订单列表
func (d *OrderDao) ListByUserID(ctx context.Context, userID uint64) ([]model.Order, error) {
	var orders []model.Order
	err := d.Model(ctx).
		Where("user_id", userID).
		Order("id desc").
		Scan(&orders)
	
	if err != nil {
		return nil, err
	}
	return orders, nil
}

// List 获取订单列表（分页）
func (d *OrderDao) List(ctx context.Context, page, limit int) ([]model.Order, int, error) {
	var orders []model.Order
	
	total, err := d.Model(ctx).Count()
	if err != nil {
		return nil, 0, err
	}
	
	err = d.Model(ctx).
		Page(page, limit).
		Order("id desc").
		Scan(&orders)
	
	if err != nil {
		return nil, 0, err
	}
	
	return orders, total, nil
}

// Update 更新订单
func (d *OrderDao) Update(ctx context.Context, id uint64, data g.Map) error {
	_, err := d.Model(ctx).Where("id", id).Update(data)
	return err
}

// GetOpenOrders 获取未成交订单
func (d *OrderDao) GetOpenOrders(ctx context.Context, symbol, side string, price string) ([]model.Order, error) {
	var orders []model.Order
	
	m := d.Model(ctx).
		Where("symbol", symbol).
		Where("side", side).
		Where("status", "open")
	
	if side == "sell" {
		m = m.Where("price <=", price).Order("price asc")
	} else {
		m = m.Where("price >=", price).Order("price desc")
	}
	
	err := m.Scan(&orders)
	if err != nil {
		return nil, err
	}
	
	return orders, nil
}
