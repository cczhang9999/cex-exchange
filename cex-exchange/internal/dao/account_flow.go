package dao

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"

	"cex-exchange/internal/model"
)

// AccountFlowDao 资金流水数据访问对象
type AccountFlowDao struct {
	table string
}

var AccountFlow = &AccountFlowDao{
	table: "account_flows",
}

// Model 获取模型
func (d *AccountFlowDao) Model(ctx context.Context) *gdb.Model {
	return g.Model(d.table).Ctx(ctx)
}

// Create 创建流水记录
func (d *AccountFlowDao) Create(ctx context.Context, data g.Map) (uint64, error) {
	result, err := d.Model(ctx).Data(data).Insert()
	if err != nil {
		return 0, err
	}
	id, _ := result.LastInsertId()
	return uint64(id), nil
}

// CreateWithTx 在事务中创建流水记录
func (d *AccountFlowDao) CreateWithTx(ctx context.Context, tx gdb.TX, data g.Map) (uint64, error) {
	result, err := tx.Model(d.table).Data(data).Insert()
	if err != nil {
		return 0, err
	}
	id, _ := result.LastInsertId()
	return uint64(id), nil
}

// ListByUserID 获取用户流水记录
func (d *AccountFlowDao) ListByUserID(ctx context.Context, userID uint64, page, limit int) ([]model.AccountFlow, int, error) {
	var flows []model.AccountFlow
	
	total, err := d.Model(ctx).Where("user_id", userID).Count()
	if err != nil {
		return nil, 0, err
	}
	
	err = d.Model(ctx).
		Where("user_id", userID).
		Page(page, limit).
		Order("id desc").
		Scan(&flows)
	
	if err != nil {
		return nil, 0, err
	}
	
	return flows, total, nil
}
