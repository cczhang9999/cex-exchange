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

	// 定义与User的关联关系
	User User `gorm:"foreignKey:UserID"`
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

// 添加关联查询方法 - 通过预加载获取订单及用户信息
func (r *orderRepo) FindByUserIDWithUser(ctx context.Context, userID uint64) ([]*biz.Order, error) {
	var orders []Order
	// 使用Preload进行关联查询
	if err := r.data.db.WithContext(ctx).Preload("User").Where("user_id = ?", userID).Find(&orders).Error; err != nil {
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

// 使用原生SQL进行关联查询
func (r *orderRepo) FindOrdersWithUserInfo(ctx context.Context, userID uint64) ([]*biz.Order, error) {
	var results []struct {
		ID        uint                   `json:"id"`
		UserID    uint64                 `json:"user_id"`
		Symbol    string                 `json:"symbol"`
		Side      string                 `json:"side"`
		Type      string                 `json:"type"`
		Price     float64                `json:"price"`
		Amount    float64                `json:"amount"`
		Filled    float64                `json:"filled"`
		Status    string                 `json:"status"`
		CreatedAt interface{}            `json:"created_at"`
		UpdatedAt interface{}            `json:"updated_at"`
		Username  string                 `json:"username"`
		Email     string                 `json:"email"`
		UserInfo  map[string]interface{} `json:"user_info"`
	}

	// 使用原生SQL进行关联查询
	sql := `
		SELECT o.id, o.user_id, o.symbol, o.side, o.type, o.price, o.amount, o.filled, o.status, 
		       o.created_at, o.updated_at, u.username, u.email
		FROM orders o
		LEFT JOIN users u ON o.user_id = u.id
		WHERE o.user_id = ?
	`

	if err := r.data.db.WithContext(ctx).Raw(sql, userID).Scan(&results).Error; err != nil {
		return nil, err
	}

	// 转换为业务模型
	res := make([]*biz.Order, 0, len(results))
	for _, result := range results {
		order := &biz.Order{
			ID:     uint64(result.ID),
			UserID: result.UserID,
			Symbol: result.Symbol,
			Side:   biz.OrderSide(result.Side),
			Type:   biz.OrderType(result.Type),
			Price:  result.Price,
			Amount: result.Amount,
			Filled: result.Filled,
			Status: biz.OrderStatus(result.Status),
		}
		res = append(res, order)
	}

	return res, nil
}

// 使用Joins进行关联查询
func (r *orderRepo) FindOrdersWithUserUsingJoins(ctx context.Context, symbol string) ([]*biz.Order, error) {
	var results []struct {
		ID        uint64      `json:"id"`
		UserID    uint64      `json:"user_id"`
		Symbol    string      `json:"symbol"`
		Side      string      `json:"side"`
		Type      string      `json:"type"`
		Price     float64     `json:"price"`
		Amount    float64     `json:"amount"`
		Filled    float64     `json:"filled"`
		Status    string      `json:"status"`
		CreatedAt interface{} `json:"created_at"`
		UpdatedAt interface{} `json:"updated_at"`
		Username  string      `json:"username"`
	}

	// 使用Joins进行关联查询
	if err := r.data.db.WithContext(ctx).
		Table("orders o").
		Select("o.id, o.user_id, o.symbol, o.side, o.type, o.price, o.amount, o.filled, o.status, o.created_at, o.updated_at, u.username").
		Joins("left join users u on o.user_id = u.id").
		Where("o.symbol = ?", symbol).
		Scan(&results).Error; err != nil {
		return nil, err
	}

	// 转换为业务模型
	res := make([]*biz.Order, 0, len(results))
	for _, result := range results {
		order := &biz.Order{
			ID:     uint64(result.ID),
			UserID: result.UserID,
			Symbol: result.Symbol,
			Side:   biz.OrderSide(result.Side),
			Type:   biz.OrderType(result.Type),
			Price:  result.Price,
			Amount: result.Amount,
			Filled: result.Filled,
			Status: biz.OrderStatus(result.Status),
		}
		res = append(res, order)
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
