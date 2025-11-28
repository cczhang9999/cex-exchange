package dao

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"

	"cex-exchange/internal/model"
)

// KlineDao K线数据访问对象
type KlineDao struct {
	table string
}

var Kline = &KlineDao{
	table: "klines",
}

// Model 获取模型
func (d *KlineDao) Model(ctx context.Context) *gdb.Model {
	return g.Model(d.table).Ctx(ctx)
}

// GetBySymbolAndPeriod 获取K线数据
func (d *KlineDao) GetBySymbolAndPeriod(ctx context.Context, symbol, period string, limit int) ([]model.Kline, error) {
	var klines []model.Kline
	err := d.Model(ctx).
		Where("symbol", symbol).
		Where("period", period).
		Order("open_time desc").
		Limit(limit).
		Scan(&klines)
	
	if err != nil {
		return nil, err
	}
	return klines, nil
}

// Create 创建K线
func (d *KlineDao) Create(ctx context.Context, data g.Map) (uint64, error) {
	result, err := d.Model(ctx).Data(data).Insert()
	if err != nil {
		return 0, err
	}
	id, _ := result.LastInsertId()
	return uint64(id), nil
}

// Update 更新K线
func (d *KlineDao) Update(ctx context.Context, id uint64, data g.Map) error {
	_, err := d.Model(ctx).Where("id", id).Update(data)
	return err
}
