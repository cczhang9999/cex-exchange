package dao

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"

	"cex-exchange/internal/model"
)

// TradeDao 成交记录数据访问对象
type TradeDao struct {
	table string
}

var Trade = &TradeDao{
	table: "trades",
}

// Model 获取模型
func (d *TradeDao) Model(ctx context.Context) *gdb.Model {
	return g.Model(d.table).Ctx(ctx)
}

// Create 创建成交记录
func (d *TradeDao) Create(ctx context.Context, data g.Map) (uint64, error) {
	result, err := d.Model(ctx).Data(data).Insert()
	if err != nil {
		return 0, err
	}
	id, _ := result.LastInsertId()
	return uint64(id), nil
}

// ListBySymbol 根据交易对获取成交记录
func (d *TradeDao) ListBySymbol(ctx context.Context, symbol string, limit int) ([]model.Trade, error) {
	var trades []model.Trade
	err := d.Model(ctx).
		Where("symbol", symbol).
		Order("id desc").
		Limit(limit).
		Scan(&trades)
	
	if err != nil {
		return nil, err
	}
	return trades, nil
}

// ListByUserID 获取用户成交记录
func (d *TradeDao) ListByUserID(ctx context.Context, userID uint64) ([]model.Trade, error) {
	var trades []model.Trade
	err := d.Model(ctx).
		WhereOr("buy_user_id", userID).
		WhereOr("sell_user_id", userID).
		Order("id desc").
		Scan(&trades)
	
	if err != nil {
		return nil, err
	}
	return trades, nil
}
