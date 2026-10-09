package biz

import (
	"context"
	"time"
)

// Trade 定义成交记录业务模型
type Trade struct {
	ID          uint64    `json:"id"`
	BuyOrderID  uint64    `json:"buy_order_id"`
	SellOrderID uint64    `json:"sell_order_id"`
	Symbol      string    `json:"symbol"`
	Price       float64   `json:"price"`
	Amount      float64   `json:"amount"`
	Side        string    `json:"side"` // 用户视角：buy/sell
	Fee         float64   `json:"fee"`
	BuyUserID   uint64    `json:"buy_user_id"`
	SellUserID  uint64    `json:"sell_user_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TradeRepo 定义成交记录存储接口
type TradeRepo interface {
	// ListByUserID 查询用户参与过的所有成交（作为买方或卖方），按 id 降序
	ListByUserID(ctx context.Context, userID uint64) ([]*Trade, error)
}

// TradeUsecase 定义成交业务逻辑
type TradeUsecase struct {
	repo TradeRepo
}

// NewTradeUsecase 创建成交业务逻辑实例
func NewTradeUsecase(repo TradeRepo) *TradeUsecase {
	return &TradeUsecase{repo: repo}
}

// GetMyTrades 查询用户的成交记录
func (uc *TradeUsecase) GetMyTrades(ctx context.Context, userID uint64) ([]*Trade, error) {
	return uc.repo.ListByUserID(ctx, userID)
}
