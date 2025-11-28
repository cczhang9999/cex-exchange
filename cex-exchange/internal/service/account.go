package service

import (
	"context"
	"strconv"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"cex-exchange/internal/dao"
	"cex-exchange/internal/model"
)

type IAccount interface {
	GetAccounts(ctx context.Context, userID uint64) ([]model.Account, error)
	Deposit(ctx context.Context, userID uint64, asset, amount string) error
	Withdraw(ctx context.Context, userID uint64, asset, amount string) error
	AddFunds(ctx context.Context, userID uint64, asset, amount, remark string) (*model.Account, error)
}

type accountImpl struct{}

var Account = &accountImpl{}

// GetAccounts 获取账户列表
func (s *accountImpl) GetAccounts(ctx context.Context, userID uint64) ([]model.Account, error) {
	return dao.Account.ListByUserID(ctx, userID)
}

// Deposit 充值
func (s *accountImpl) Deposit(ctx context.Context, userID uint64, asset, amount string) error {
	// 验证金额
	amountFloat, err := strconv.ParseFloat(amount, 64)
	if err != nil || amountFloat <= 0 {
		return gerror.New("金额格式错误")
	}
	
	// 使用事务
	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 查询或创建账户
		account, err := dao.Account.GetByUserIDAndAsset(ctx, userID, asset)
		
		var accountID uint64
		var newBalance float64
		
		if err != nil || account == nil {
			// 创建新账户
			accountID, err = dao.Account.Create(ctx, g.Map{
				"user_id": userID,
				"asset":   asset,
				"balance": amount,
				"frozen":  "0",
			})
			if err != nil {
				return err
			}
			newBalance = amountFloat
		} else {
			// 更新余额
			accountID = account.ID
			currentBalance, _ := strconv.ParseFloat(account.Balance, 64)
			newBalance = currentBalance + amountFloat
			
			err = dao.Account.UpdateBalance(ctx, tx, accountID, strconv.FormatFloat(newBalance, 'f', -1, 64))
			if err != nil {
				return err
			}
		}
		
		// 创建流水记录
		_, err = dao.AccountFlow.CreateWithTx(ctx, tx, g.Map{
			"user_id":     userID,
			"account_id":  accountID,
			"asset":       asset,
			"change_type": "deposit",
			"amount":      amount,
			"balance":     strconv.FormatFloat(newBalance, 'f', -1, 64),
			"remark":      "用户充值",
		})
		
		return err
	})
}

// Withdraw 提现
func (s *accountImpl) Withdraw(ctx context.Context, userID uint64, asset, amount string) error {
	// 验证金额
	amountFloat, err := strconv.ParseFloat(amount, 64)
	if err != nil || amountFloat <= 0 {
		return gerror.New("金额格式错误")
	}
	
	// 使用事务
	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 查询账户
		account, err := dao.Account.GetByUserIDAndAsset(ctx, userID, asset)
		if err != nil || account == nil {
			return gerror.New("账户不存在")
		}
		
		// 检查余额
		currentBalance, _ := strconv.ParseFloat(account.Balance, 64)
		if currentBalance < amountFloat {
			return gerror.New("余额不足")
		}
		
		// 更新余额
		newBalance := currentBalance - amountFloat
		err = dao.Account.UpdateBalance(ctx, tx, account.ID, strconv.FormatFloat(newBalance, 'f', -1, 64))
		if err != nil {
			return err
		}
		
		// 创建流水记录
		_, err = dao.AccountFlow.CreateWithTx(ctx, tx, g.Map{
			"user_id":     userID,
			"account_id":  account.ID,
			"asset":       asset,
			"change_type": "withdraw",
			"amount":      "-" + amount,
			"balance":     strconv.FormatFloat(newBalance, 'f', -1, 64),
			"remark":      "用户提现",
		})
		
		return err
	})
}

// AddFunds 管理员添加资金
func (s *accountImpl) AddFunds(ctx context.Context, userID uint64, asset, amount, remark string) (*model.Account, error) {
	// 验证金额
	amountFloat, err := strconv.ParseFloat(amount, 64)
	if err != nil || amountFloat <= 0 {
		return nil, gerror.New("金额格式错误")
	}
	
	// 验证用户是否存在
	exists, err := dao.User.Exists(ctx, "")
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, gerror.New("用户不存在")
	}
	
	var resultAccount *model.Account
	
	// 使用事务
	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 查询或创建账户
		account, err := dao.Account.GetByUserIDAndAsset(ctx, userID, asset)
		
		var accountID uint64
		var newBalance float64
		
		if err != nil || account == nil {
			// 创建新账户
			accountID, err = dao.Account.Create(ctx, g.Map{
				"user_id": userID,
				"asset":   asset,
				"balance": amount,
				"frozen":  "0",
			})
			if err != nil {
				return err
			}
			newBalance = amountFloat
		} else {
			// 更新余额
			accountID = account.ID
			currentBalance, _ := strconv.ParseFloat(account.Balance, 64)
			newBalance = currentBalance + amountFloat
			
			err = dao.Account.UpdateBalance(ctx, tx, accountID, strconv.FormatFloat(newBalance, 'f', -1, 64))
			if err != nil {
				return err
			}
		}
		
		// 创建流水记录
		_, err = dao.AccountFlow.CreateWithTx(ctx, tx, g.Map{
			"user_id":     userID,
			"account_id":  accountID,
			"asset":       asset,
			"change_type": "admin_add",
			"amount":      amount,
			"balance":     strconv.FormatFloat(newBalance, 'f', -1, 64),
			"remark":      "管理员添加资金: " + remark,
		})
		
		if err != nil {
			return err
		}
		
		// 查询更新后的账户
		resultAccount, err = dao.Account.GetByUserIDAndAsset(ctx, userID, asset)
		return err
	})
	
	if err != nil {
		return nil, err
	}
	
	return resultAccount, nil
}
