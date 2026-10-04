package biz

import (
	"context"
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
