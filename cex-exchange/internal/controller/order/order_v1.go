package order

import (
	"context"

	"cex-exchange/internal/model"

	"github.com/gogf/gf/v2/frame/g"
)

type ControllerV1 struct{}

func NewV1() *ControllerV1 {
	return &ControllerV1{}
}

type PlaceOrderReq struct {
	g.Meta `path:"/order" method:"post" tags:"Order" summary:"下单"`
	Symbol string `json:"symbol" v:"required#请输入交易对"`
	Side   string `json:"side" v:"required|in:buy,sell#请选择方向|方向只能是buy或sell"`
	Type   string `json:"type" v:"required|in:limit,market#请选择类型|类型只能是limit或market"`
	Price  string `json:"price"`
	Amount string `json:"amount" v:"required#请输入数量"`
}

type PlaceOrderRes struct {
	OrderID uint64 `json:"order_id"`
}

func (c *ControllerV1) PlaceOrder(ctx context.Context, req *PlaceOrderReq) (res *PlaceOrderRes, err error) {
	uid := g.RequestFromCtx(ctx).GetCtxVar("uid").Uint64()

	// 插入订单
	result, err := g.Model("orders").Ctx(ctx).Data(g.Map{
		"user_id": uid,
		"symbol":  req.Symbol,
		"side":    req.Side,
		"type":    req.Type,
		"price":   req.Price,
		"amount":  req.Amount,
		"filled":  "0",
		"status":  "open",
	}).Insert()

	if err != nil {
		return nil, err
	}

	orderID, _ := result.LastInsertId()

	// TODO: 调用撮合引擎

	return &PlaceOrderRes{OrderID: uint64(orderID)}, nil
}

type ListMyOrdersReq struct {
	g.Meta `path:"/my_orders" method:"get" tags:"Order" summary:"我的订单"`
}

type ListMyOrdersRes struct {
	Orders []*model.Order `json:"orders"`
}

// ListMyOrders 获取当前用户的订单列表
func (c *ControllerV1) ListMyOrders(ctx context.Context, req *ListMyOrdersReq) (res *ListMyOrdersRes, err error) {
	// 从请求上下文中获取用户ID
	// g.RequestFromCtx(ctx) 从 context 中获取 GoFrame 的 Request 对象
	// GetCtxVar("uid") 获取中间件设置的上下文变量 "uid"
	// Uint64() 将值转换为 uint64 类型
	uid := g.RequestFromCtx(ctx).GetCtxVar("uid").Uint64()

	// 定义订单切片，用于存储查询结果
	var orders []*model.Order
	
	// g 是 GoFrame 框架的全局对象，提供了各种便捷方法
	// g.Model("orders") 创建一个数据库模型，对应 orders 表
	// Ctx(ctx) 设置上下文，用于日志追踪和超时控制
	// Where("user_id", uid) 添加 WHERE 条件：user_id = uid
	// Order("id desc") 按 id 降序排序（最新的订单在前）
	// Scan(&orders) 将查询结果扫描到 orders 切片中
	err = g.Model("orders").Ctx(ctx).
		Where("user_id", uid).
		Order("id desc").
		Scan(&orders)

	// 如果查询出错，返回错误
	if err != nil {
		return nil, err
	}
	
	// 返回订单列表响应
	// GoFrame 会自动将返回值包装成 {code: 0, message: "success", data: {...}} 格式
	return &ListMyOrdersRes{Orders: orders}, nil
}
