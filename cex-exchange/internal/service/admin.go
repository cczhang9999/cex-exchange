package service

import (
	"context"

	"cex-exchange/internal/dao"
	"cex-exchange/internal/model"
)

type IAdmin interface {
	ListUsers(ctx context.Context, page, limit int) ([]model.User, int, error)
	ListOrders(ctx context.Context, page, limit int) ([]model.Order, int, error)
	ListAccounts(ctx context.Context, page, limit int) ([]model.Account, int, error)
}

type adminImpl struct{}

var Admin = &adminImpl{}

// ListUsers 获取用户列表
func (s *adminImpl) ListUsers(ctx context.Context, page, limit int) ([]model.User, int, error) {
	return dao.User.List(ctx, page, limit)
}

// ListOrders 获取订单列表
func (s *adminImpl) ListOrders(ctx context.Context, page, limit int) ([]model.Order, int, error) {
	return dao.Order.List(ctx, page, limit)
}

// ListAccounts 获取账户列表
func (s *adminImpl) ListAccounts(ctx context.Context, page, limit int) ([]model.Account, int, error) {
	return dao.Account.List(ctx, page, limit)
}
