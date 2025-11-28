package service

import (
	"context"

	"cex-exchange/internal/dao"
	"cex-exchange/internal/model"
)

type ITrade interface {
	ListMyTrades(ctx context.Context, userID uint64) ([]model.Trade, error)
}

type tradeImpl struct{}

var Trade = &tradeImpl{}

// ListMyTrades 获取用户的成交记录
func (s *tradeImpl) ListMyTrades(ctx context.Context, userID uint64) ([]model.Trade, error) {
	return dao.Trade.ListByUserID(ctx, userID)
}
