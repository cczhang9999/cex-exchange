package data

import (
	"backend-v2/internal/biz"
	"context"

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
