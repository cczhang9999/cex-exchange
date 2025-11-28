package dao

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"

	"cex-exchange/internal/model"
)

// AccountDao 账户数据访问对象
type AccountDao struct {
	table string
}

var Account = &AccountDao{
	table: "accounts",
}

// Model 获取模型
func (d *AccountDao) Model(ctx context.Context) *gdb.Model {
	return g.Model(d.table).Ctx(ctx)
}

// GetByUserIDAndAsset 根据用户ID和资产类型获取账户
func (d *AccountDao) GetByUserIDAndAsset(ctx context.Context, userID uint64, asset string) (*model.Account, error) {
	var account model.Account
	err := d.Model(ctx).
		Where("user_id", userID).
		Where("asset", asset).
		Scan(&account)
	
	if err != nil {
		return nil, err
	}
	return &account, nil
}

// ListByUserID 获取用户所有账户
func (d *AccountDao) ListByUserID(ctx context.Context, userID uint64) ([]model.Account, error) {
	var accounts []model.Account
	err := d.Model(ctx).
		Where("user_id", userID).
		Scan(&accounts)
	
	if err != nil {
		return nil, err
	}
	return accounts, nil
}

// List 获取账户列表（分页）
func (d *AccountDao) List(ctx context.Context, page, limit int) ([]model.Account, int, error) {
	var accounts []model.Account
	
	total, err := d.Model(ctx).Count()
	if err != nil {
		return nil, 0, err
	}
	
	err = d.Model(ctx).
		Page(page, limit).
		Order("id desc").
		Scan(&accounts)
	
	if err != nil {
		return nil, 0, err
	}
	
	return accounts, total, nil
}

// Create 创建账户
func (d *AccountDao) Create(ctx context.Context, data g.Map) (uint64, error) {
	result, err := d.Model(ctx).Data(data).Insert()
	if err != nil {
		return 0, err
	}
	id, _ := result.LastInsertId()
	return uint64(id), nil
}

// Update 更新账户
func (d *AccountDao) Update(ctx context.Context, id uint64, data g.Map) error {
	_, err := d.Model(ctx).Where("id", id).Update(data)
	return err
}

// UpdateBalance 更新余额（事务内）
func (d *AccountDao) UpdateBalance(ctx context.Context, tx gdb.TX, id uint64, balance string) error {
	_, err := tx.Model(d.table).Where("id", id).Update(g.Map{"balance": balance})
	return err
}
