package funds

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

type ControllerV1 struct{}

func NewV1() *ControllerV1 {
	return &ControllerV1{}
}

type DepositReq struct {
	g.Meta   `path:"/deposit" method:"post" tags:"Funds" summary:"充值"`
	Currency string `json:"currency" v:"required#请输入币种"`
	Amount   string `json:"amount" v:"required#请输入金额"`
}

type DepositRes struct {
	Success bool `json:"success"`
}

func (c *ControllerV1) Deposit(ctx context.Context, req *DepositReq) (res *DepositRes, err error) {
	uid := g.RequestFromCtx(ctx).GetCtxVar("uid").Uint64()
	
	// TODO: 实现充值逻辑
	g.Log().Infof(ctx, "User %d deposit %s %s", uid, req.Amount, req.Currency)
	
	return &DepositRes{Success: true}, nil
}

type ListAccountsReq struct {
	g.Meta `path:"/accounts" method:"get" tags:"Funds" summary:"账户列表"`
}

type ListAccountsRes struct {
	Accounts []g.Map `json:"accounts"`
}

func (c *ControllerV1) ListAccounts(ctx context.Context, req *ListAccountsReq) (res *ListAccountsRes, err error) {
	uid := g.RequestFromCtx(ctx).GetCtxVar("uid").Uint64()
	
	var accounts []g.Map
	err = g.Model("accounts").Ctx(ctx).
		Where("user_id", uid).
		Scan(&accounts)
	
	if err != nil {
		return nil, err
	}
	
	return &ListAccountsRes{Accounts: accounts}, nil
}
