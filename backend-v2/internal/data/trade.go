package data

import (
	"backend-v2/internal/biz"
	"context"

	"github.com/google/wire"
	"gorm.io/gorm"
)

var TradeProviderSet = wire.NewSet(NewTradeRepo)

// Trade GORM Model
// 对应表结构（见 db_schema.sql trades 表）：
// id, buy_order_id, sell_order_id, symbol, price, amount,
// buy_user_id, sell_user_id, created_at, updated_at, deleted_at
type Trade struct {
	gorm.Model
	BuyOrderID  uint64
	SellOrderID uint64
	Symbol      string
	Price       float64
	Amount      float64
	BuyUserID   uint64
	SellUserID  uint64
}

type tradeRepo struct {
	data *Data
}

// NewTradeRepo 创建成交记录数据访问实例
func NewTradeRepo(data *Data) biz.TradeRepo {
	return &tradeRepo{data: data}
}

// ListByUserID 查询用户参与过的所有成交（作为买方或卖方），按 id 降序
func (r *tradeRepo) ListByUserID(ctx context.Context, userID uint64) ([]*biz.Trade, error) {
	var trades []Trade
	if err := r.data.db.WithContext(ctx).
		Where("buy_user_id = ? OR sell_user_id = ?", userID, userID).
		Order("id DESC").
		Find(&trades).Error; err != nil {
		return nil, err
	}

	res := make([]*biz.Trade, 0, len(trades))
	for _, t := range trades {
		side := "sell"
		if t.BuyUserID == userID {
			side = "buy"
		}
		res = append(res, &biz.Trade{
			ID:          uint64(t.ID),
			BuyOrderID:  t.BuyOrderID,
			SellOrderID: t.SellOrderID,
			Symbol:      t.Symbol,
			Price:       t.Price,
			Amount:      t.Amount,
			Side:        side,
			BuyUserID:   t.BuyUserID,
			SellUserID:  t.SellUserID,
			CreatedAt:   t.CreatedAt,
			UpdatedAt:   t.UpdatedAt,
		})
	}
	return res, nil
}
