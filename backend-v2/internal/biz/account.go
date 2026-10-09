package biz

import (
	"context"
	"errors"
	"strconv"
	"time"
)

// Account 定义用户账户业务模型
type Account struct {
	ID        uint64    `json:"id"`
	UserID    uint64    `json:"user_id"`
	Asset     string    `json:"asset"`
	Balance   float64   `json:"balance"`
	Frozen    float64   `json:"frozen"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AccountRepo 定义账户存储接口
type AccountRepo interface {
	// FindByUserID 查询用户的全部账户
	FindByUserID(ctx context.Context, userID uint64) ([]*Account, error)
	// FindByUserIDAndAsset 查询用户指定币种的账户
	FindByUserIDAndAsset(ctx context.Context, userID uint64, asset string) (*Account, error)
	// FindAll 查询全部账户（管理员用）
	FindAll(ctx context.Context) ([]*Account, error)
	// Deposit 充值：增加用户某币种可用余额，并记录资金流水
	Deposit(ctx context.Context, userID uint64, asset string, amount float64) error
	// Withdraw 提现：扣减用户某币种可用余额（校验充足），并记录资金流水
	Withdraw(ctx context.Context, userID uint64, asset string, amount float64, address string) error
	// AdjustBalance 管理员调账：直接增减用户某币种余额，并记录资金流水
	AdjustBalance(ctx context.Context, userID uint64, asset string, delta float64, changeType, remark string) error
}

// AccountUsecase 定义账户业务逻辑
type AccountUsecase struct {
	repo AccountRepo
}

// NewAccountUsecase 创建账户业务逻辑实例
func NewAccountUsecase(repo AccountRepo) *AccountUsecase {
	return &AccountUsecase{
		repo: repo,
	}
}

// GetAccounts 查询用户账户列表
func (uc *AccountUsecase) GetAccounts(ctx context.Context, userID uint64) ([]*Account, error) {
	return uc.repo.FindByUserID(ctx, userID)
}

// GetAccount 查询用户指定币种的账户
func (uc *AccountUsecase) GetAccount(ctx context.Context, userID uint64, asset string) (*Account, error) {
	return uc.repo.FindByUserIDAndAsset(ctx, userID, asset)
}

// GetAllAccounts 查询全部账户（管理员用）
func (uc *AccountUsecase) GetAllAccounts(ctx context.Context) ([]*Account, error) {
	return uc.repo.FindAll(ctx)
}

// Deposit 充值
func (uc *AccountUsecase) Deposit(ctx context.Context, userID uint64, asset, amount string) error {
	amt, err := parseAmount(amount)
	if err != nil {
		return err
	}
	if amt <= 0 {
		return errors.New("充值金额必须大于 0")
	}
	return uc.repo.Deposit(ctx, userID, asset, amt)
}

// Withdraw 提现
func (uc *AccountUsecase) Withdraw(ctx context.Context, userID uint64, asset, amount, address string) error {
	amt, err := parseAmount(amount)
	if err != nil {
		return err
	}
	if amt <= 0 {
		return errors.New("提现金额必须大于 0")
	}
	return uc.repo.Withdraw(ctx, userID, asset, amt, address)
}

// AdjustBalance 管理员调账
func (uc *AccountUsecase) AdjustBalance(ctx context.Context, userID uint64, asset, amount string) error {
	amt, err := parseAmount(amount)
	if err != nil {
		return err
	}
	if amt <= 0 {
		return errors.New("金额必须大于 0")
	}
	return uc.repo.AdjustBalance(ctx, userID, asset, amt, "admin_add", "管理员添加资金")
}

// parseAmount 将字符串金额解析为 float64
func parseAmount(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}
