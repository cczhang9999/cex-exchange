package admin

import (
	"cex-exchange/internal/model"
	"context"
	"strconv"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

type ControllerV1 struct{}

func NewV1() *ControllerV1 {
	return &ControllerV1{}
}

// ListUsersReq 用户列表请求
type ListUsersReq struct {
	g.Meta `path:"/admin/users" method:"get" tags:"Admin" summary:"用户列表"`
	Page   int `json:"page" d:"1" v:"min:1"`
	Limit  int `json:"limit" d:"20" v:"min:1|max:100"`
}

type ListUsersRes struct {
	Users []g.Map `json:"users"`
	Total int     `json:"total"`
}

func (c *ControllerV1) ListUsers(ctx context.Context, req *ListUsersReq) (res *ListUsersRes, err error) {
	// 查询总数
	total, err := g.Model("users").Ctx(ctx).Count()
	if err != nil {
		return nil, err
	}

	// 分页查询，使用 All() 返回 Result
	result, err := g.Model("users").Ctx(ctx).
		Page(req.Page, req.Limit).
		All()

	if err != nil {
		return nil, err
	}

	return &ListUsersRes{
		Users: result.List(),
		Total: total,
	}, nil
}

// ListOrdersReq 订单列表请求
type ListOrdersReq struct {
	g.Meta `path:"/admin/orders" method:"get" tags:"Admin" summary:"订单列表"`
	Page   int `json:"page" d:"1" v:"min:1"`
	Limit  int `json:"limit" d:"20" v:"min:1|max:100"`
}

type ListOrdersRes struct {
	Orders []model.Order `json:"orders"`
	Total  int           `json:"total"`
}

func (c *ControllerV1) ListOrders(ctx context.Context, req *ListOrdersReq) (res *ListOrdersRes, err error) {
	var orders []model.Order

	total, err := g.Model("orders").Ctx(ctx).Count()
	if err != nil {
		return nil, err
	}

	err = g.Model("orders").Ctx(ctx).
		Page(req.Page, req.Limit).
		Order("id desc").
		Scan(&orders)

	if err != nil {
		return nil, err
	}

	return &ListOrdersRes{
		Orders: orders,
		Total:  total,
	}, nil
}

// ListAccountsReq 账户列表请求
type ListAccountsReq struct {
	g.Meta `path:"/admin/accounts" method:"get" tags:"Admin" summary:"账户列表"`
	Page   int `json:"page" d:"1" v:"min:1"`
	Limit  int `json:"limit" d:"20" v:"min:1|max:100"`
}

type ListAccountsRes struct {
	Accounts []model.Account `json:"accounts"`
	Total    int             `json:"total"`
}

func (c *ControllerV1) ListAccounts(ctx context.Context, req *ListAccountsReq) (res *ListAccountsRes, err error) {
	var accounts []model.Account

	total, err := g.Model("accounts").Ctx(ctx).Count()
	if err != nil {
		return nil, err
	}

	err = g.Model("accounts").Ctx(ctx).
		Page(req.Page, req.Limit).
		Order("id desc").
		Scan(&accounts)

	if err != nil {
		return nil, err
	}

	return &ListAccountsRes{
		Accounts: accounts,
		Total:    total,
	}, nil
}

// AddFundsReq 添加资金请求
type AddFundsReq struct {
	g.Meta `path:"/admin/accounts/add-funds" method:"post" tags:"Admin" summary:"添加资金"`
	UserID uint64 `json:"user_id" v:"required#请输入用户ID"`
	Asset  string `json:"asset" v:"required#请输入资产类型"`
	Amount string `json:"amount" v:"required#请输入金额"`
	Remark string `json:"remark" v:"required#请输入备注"`
}

type AddFundsRes struct {
	Success bool  `json:"success"`
	Account g.Map `json:"account"`
}

func (c *ControllerV1) AddFunds(ctx context.Context, req *AddFundsReq) (res *AddFundsRes, err error) {
	// 验证金额
	amount, err := strconv.ParseFloat(req.Amount, 64)
	if err != nil || amount <= 0 {
		return nil, gerror.New("金额格式错误")
	}

	// 验证用户是否存在
	count, err := g.Model("users").Ctx(ctx).Where("id", req.UserID).Count()
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, gerror.New("用户不存在")
	}

	// 使用事务
	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 查询或创建账户
		var account g.Map
		err := tx.Model("accounts").
			Where("user_id", req.UserID).
			Where("asset", req.Asset).
			Scan(&account)

		var accountID uint64
		var newBalance float64

		if err != nil || len(account) == 0 {
			// 创建新账户
			result, err := tx.Model("accounts").Data(g.Map{
				"user_id": req.UserID,
				"asset":   req.Asset,
				"balance": req.Amount,
				"frozen":  "0",
			}).Insert()

			if err != nil {
				return err
			}

			id, _ := result.LastInsertId()
			accountID = uint64(id)
			newBalance = amount
		} else {
			// 更新余额
			accountID = account["id"].(uint64)
			currentBalance, _ := strconv.ParseFloat(account["balance"].(string), 64)
			newBalance = currentBalance + amount

			_, err = tx.Model("accounts").
				Where("id", accountID).
				Update(g.Map{"balance": strconv.FormatFloat(newBalance, 'f', -1, 64)})

			if err != nil {
				return err
			}
		}

		var accountNew = g.Map{
			"user_id":     req.UserID,
			"account_id":  accountID,
			"asset":       req.Asset,
			"change_type": "admin_add",
			"amount":      req.Amount,
			"balance":     strconv.FormatFloat(newBalance, 'f', -1, 64),
			"ref_id":      0, // 建议显式填充，表示系统/管理员操作
			"remark":      "管理员添加资金:" + req.Remark,
		}

		// 创建流水记录
		_, err = tx.Model("account_flows").Data(accountNew).Insert()

		return err
	})

	if err != nil {
		return nil, err
	}

	// 查询更新后的账户
	var account g.Map
	g.Model("accounts").Ctx(ctx).
		Where("user_id", req.UserID).
		Where("asset", req.Asset).
		Scan(&account)

	return &AddFundsRes{
		Success: true,
		Account: account,
	}, nil
}
