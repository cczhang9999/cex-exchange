package data

import (
	"backend-v2/internal/biz"
	"context"
	"errors"
	"fmt"

	"github.com/google/wire"
	"gorm.io/gorm"
)

var AccountProviderSet = wire.NewSet(NewAccountRepo)

// Account GORM Model
// 对应表结构：
// id bigint unsigned AUTO_INCREMENT 账户ID
// user_id bigint unsigned 用户ID
// asset varchar(32) 币种
// balance decimal(32,16) 可用余额
// frozen decimal(32,16) 冻结余额
// created_at / updated_at / deleted_at
type Account struct {
	gorm.Model
	UserID  uint64
	Asset   string
	Balance float64
	Frozen  float64
}

type accountRepo struct {
	data *Data
}

func NewAccountRepo(data *Data) biz.AccountRepo {
	return &accountRepo{
		data: data,
	}
}

// FindByUserID 查询用户的全部账户
func (r *accountRepo) FindByUserID(ctx context.Context, userID uint64) ([]*biz.Account, error) {
	var accounts []Account
	if err := r.data.db.WithContext(ctx).Where("user_id = ?", userID).Find(&accounts).Error; err != nil {
		return nil, err
	}

	res := make([]*biz.Account, 0, len(accounts))
	for _, a := range accounts {
		res = append(res, &biz.Account{
			ID:        uint64(a.ID),
			UserID:    a.UserID,
			Asset:     a.Asset,
			Balance:   a.Balance,
			Frozen:    a.Frozen,
			CreatedAt: a.CreatedAt,
			UpdatedAt: a.UpdatedAt,
		})
	}
	return res, nil
}

// FindByUserIDAndAsset 查询用户指定币种的账户
func (r *accountRepo) FindByUserIDAndAsset(ctx context.Context, userID uint64, asset string) (*biz.Account, error) {
	var account Account
	if err := r.data.db.WithContext(ctx).Where("user_id = ? AND asset = ?", userID, asset).First(&account).Error; err != nil {
		return nil, err
	}

	return &biz.Account{
		ID:        uint64(account.ID),
		UserID:    account.UserID,
		Asset:     account.Asset,
		Balance:   account.Balance,
		Frozen:    account.Frozen,
		CreatedAt: account.CreatedAt,
		UpdatedAt: account.UpdatedAt,
	}, nil
}

// FindAll 查询全部账户（管理员用）
func (r *accountRepo) FindAll(ctx context.Context) ([]*biz.Account, error) {
	var accounts []Account
	if err := r.data.db.WithContext(ctx).Find(&accounts).Error; err != nil {
		return nil, err
	}

	res := make([]*biz.Account, 0, len(accounts))
	for _, a := range accounts {
		res = append(res, &biz.Account{
			ID:        uint64(a.ID),
			UserID:    a.UserID,
			Asset:     a.Asset,
			Balance:   a.Balance,
			Frozen:    a.Frozen,
			CreatedAt: a.CreatedAt,
			UpdatedAt: a.UpdatedAt,
		})
	}
	return res, nil
}

// upsertAccount 在事务内查找或创建用户的币种账户，返回账户 ID
func (r *accountRepo) upsertAccount(ctx context.Context, tx *gorm.DB, userID uint64, asset string) (uint64, error) {
	var account Account
	if err := tx.Where("user_id = ? AND asset = ?", userID, asset).
		FirstOrCreate(&account, Account{
			UserID:  userID,
			Asset:   asset,
			Balance: 0,
			Frozen:  0,
		}).Error; err != nil {
		return 0, err
	}
	return uint64(account.ID), nil
}

// makeFlow 构建账户流水记录
func makeFlow(userID, accountID uint64, asset, changeType string, amount, balance float64, refID *uint64, remark *string) *AccountFlow {
	return &AccountFlow{
		UserID:     userID,
		AccountID:  accountID,
		Asset:      asset,
		ChangeType: changeType,
		Amount:     amount,
		Balance:    balance,
		RefID:      refID,
		Remark:     remark,
	}
}

// Deposit 充值：事务内查找/创建账户，增加余额，记录流水
func (r *accountRepo) Deposit(ctx context.Context, userID uint64, asset string, amount float64) error {
	return r.data.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		accountID, err := r.upsertAccount(ctx, tx, userID, asset)
		if err != nil {
			return err
		}

		var account Account
		if err := tx.Where("id = ?", accountID).First(&account).Error; err != nil {
			return err
		}

		newBalance := account.Balance + amount
		if err := tx.Model(&Account{}).Where("id = ?", accountID).
			Update("balance", newBalance).Error; err != nil {
			return err
		}

		remark := "充值"
		refID := &userID
		flow := makeFlow(userID, accountID, asset, "deposit", amount, newBalance, refID, &remark)
		return tx.Create(flow).Error
	})
}

// Withdraw 提现：校验余额充足，扣减可用余额，记录流水
func (r *accountRepo) Withdraw(ctx context.Context, userID uint64, asset string, amount float64, address string) error {
	return r.data.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var account Account
		if err := tx.Where("user_id = ? AND asset = ?", userID, asset).First(&account).Error; err != nil {
			return err
		}
		if account.ID == 0 {
			return errors.New("账户不存在，请先充值")
		}

		if account.Balance < amount {
			return fmt.Errorf("余额不足，当前可用: %.8f %s, 需要: %.8f %s", account.Balance, asset, amount, asset)
		}

		newBalance := account.Balance - amount
		if err := tx.Model(&Account{}).Where("id = ?", account.ID).
			Update("balance", newBalance).Error; err != nil {
			return err
		}

		remark := "提现"
		if address != "" {
			remark = fmt.Sprintf("提现到 %s", address)
		}
		refID := &userID
		flow := makeFlow(userID, uint64(account.ID), asset, "withdraw", -amount, newBalance, refID, &remark)
		if err := tx.Create(flow).Error; err != nil {
			return err
		}
		return nil
	})
}

// AdjustBalance 管理员调账：增减用户币种余额，记录流水
func (r *accountRepo) AdjustBalance(ctx context.Context, userID uint64, asset string, delta float64, changeType, remark string) error {
	return r.data.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		accountID, err := r.upsertAccount(ctx, tx, userID, asset)
		if err != nil {
			return err
		}

		var account Account
		if err := tx.Where("id = ?", accountID).First(&account).Error; err != nil {
			return err
		}

		newBalance := account.Balance + delta
		if err := tx.Model(&Account{}).Where("id = ?", accountID).
			Update("balance", newBalance).Error; err != nil {
			return err
		}

		rem := remark
		refID := &userID
		flow := makeFlow(userID, accountID, asset, changeType, delta, newBalance, refID, &rem)
		return tx.Create(flow).Error
	})
}
