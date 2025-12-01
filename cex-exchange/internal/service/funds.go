package service

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/shopspring/decimal"

	"cex-exchange/internal/dao"
	"cex-exchange/internal/model"
)

type IFunds interface {
	GetBalance(ctx context.Context, userID uint64) ([]model.Account, error)
	Deposit(ctx context.Context, userID uint64, asset, amount string) error
	Withdraw(ctx context.Context, userID uint64, asset, amount, address string) error
}

type fundsImpl struct{}

var Funds = &fundsImpl{}

// GetBalance 获取账户余额
func (s *fundsImpl) GetBalance(ctx context.Context, userID uint64) ([]model.Account, error) {
	return dao.Account.ListByUserID(ctx, userID)
}

// Deposit 充值
func (s *fundsImpl) Deposit(ctx context.Context, userID uint64, asset, amount string) error {
	// 解析金额
	depositAmount, err := decimal.NewFromString(amount)
	if err != nil || depositAmount.LessThanOrEqual(decimal.Zero) {
		return gerror.New("无效的充值金额")
	}

	// 数据库事务
	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 查询账户
		var account model.Account
		err := tx.Model("accounts").
			Ctx(ctx).
			Where("user_id", userID).
			Where("asset", asset).
			LockUpdate().
			Scan(&account)

		if err != nil {
			return err
		}

		if account.ID == 0 {
			// 账户不存在，创建新账户
			_, err = tx.Model("accounts").Ctx(ctx).Data(g.Map{
				"user_id": userID,
				"asset":   asset,
				"balance": amount,
				"frozen":  "0",
			}).Insert()
			return err
		}

		// 账户存在，增加余额
		_, err = tx.Model("accounts").
			Ctx(ctx).
			Where("id", account.ID).
			Data(g.Map{
				"balance": gdb.Raw(fmt.Sprintf("balance + %s", depositAmount.String())),
			}).Update()

		return err
	})

	if err != nil {
		return err
	}

	// 清除缓存
	cacheKey := fmt.Sprintf("user:balance:%d:%s", userID, asset)
	Redis.Del(ctx, cacheKey)

	return nil
}

// Withdraw 提现
func (s *fundsImpl) Withdraw(ctx context.Context, userID uint64, asset, amount, address string) error {
	// 解析金额
	withdrawAmount, err := decimal.NewFromString(amount)
	if err != nil || withdrawAmount.LessThanOrEqual(decimal.Zero) {
		return gerror.New("无效的提现金额")
	}

	// 数据库事务
	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 查询账户并加锁
		var account model.Account
		err := tx.Model("accounts").
			Ctx(ctx).
			Where("user_id", userID).
			Where("asset", asset).
			LockUpdate().
			Scan(&account)

		if err != nil {
			return err
		}

		if account.ID == 0 {
			return gerror.Newf("您还没有 %s 账户", asset)
		}

		// 检查余额
		balance, _ := decimal.NewFromString(account.Balance)
		if balance.LessThan(withdrawAmount) {
			return gerror.Newf("余额不足，当前可用: %s %s, 需要: %s %s",
				balance.String(), asset, withdrawAmount.String(), asset)
		}

		// 扣减余额
		_, err = tx.Model("accounts").
			Ctx(ctx).
			Where("id", account.ID).
			Data(g.Map{
				"balance": gdb.Raw(fmt.Sprintf("balance - %s", withdrawAmount.String())),
			}).Update()

		if err != nil {
			return err
		}

		// 记录提现记录（可选）
		// TODO: 实现提现记录表

		return nil
	})

	if err != nil {
		return err
	}

	// 清除缓存
	cacheKey := fmt.Sprintf("user:balance:%d:%s", userID, asset)
	Redis.Del(ctx, cacheKey)

	g.Log().Infof(ctx, "用户 %d 提现 %s %s 到地址 %s", userID, amount, asset, address)

	return nil
}
