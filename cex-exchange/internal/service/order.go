package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"cex-exchange/internal/dao"
	"cex-exchange/internal/engine"
	"cex-exchange/internal/model"
)

type IOrder interface {
	PlaceOrder(ctx context.Context, userID uint64, symbol, side, orderType, price, amount string) (uint64, error)
	ListMyOrders(ctx context.Context, userID uint64) ([]model.Order, error)
	CancelOrder(ctx context.Context, userID, orderID uint64) error
}

type orderImpl struct{}

var Order = &orderImpl{}

// PlaceOrder 下单
func (s *orderImpl) PlaceOrder(ctx context.Context, req *model.PlaceOrderReq) (int64, error) {
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
		return 0, err
	}

	orderID, _ := result.LastInsertId()

	// 清除用户订单缓存
	s.clearUserOrderCache(ctx, uid, req.Symbol)

	// 调用撮合引擎
	order := &engine.Order{
		ID:        orderID,
		UserID:    uid,
		Symbol:    req.Symbol,
		Side:      req.Side,
		Type:      req.Type,
		Price:     parseFloat(req.Price),
		Amount:    parseFloat(req.Amount),
		Filled:    0,
		Status:    "open",
		CreatedAt: time.Now(),
	}

	matchEngine := engine.GetEngine()
	_, err = matchEngine.Match(ctx, order)

	return orderID, err
}

// parseFloat 辅助函数：字符串转浮点数
func parseFloat(s string) float64 {
	if s == "" {
		return 0
	}
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

// ListMyOrders 获取我的订单（带缓存）
func (s *orderImpl) ListMyOrders(ctx context.Context, uid uint64, symbol string) ([]model.Order, error) {
	cacheKey := fmt.Sprintf("user:orders:%d:%s", uid, symbol)

	// 1. 尝试从缓存获取
	cachedData, err := Redis.Get(ctx, cacheKey)
	if err != nil {
		g.Log().Warningf(ctx, "Redis Get Error: %v", err)
	}
	if err == nil && cachedData != "" {
		var orders []model.Order
		if err := json.Unmarshal([]byte(cachedData), &orders); err == nil {
			g.Log().Info(ctx, "✅ Cache Hit for user:", uid)
			return orders, nil
		} else {
			g.Log().Warningf(ctx, "Redis Unmarshal Error: %v", err)
		}
	} else {
		g.Log().Info(ctx, "⚠️ Cache Miss for user:", uid)
	}

	// 2. 缓存未命中，查询数据库
	// 注意：这里使用 = 而不是 :=，因为我们要给返回值的 orders 赋值
	// 或者直接声明新变量，最后返回它
	orders, err := dao.Order.ListByUserID(ctx, uid, symbol)
	if err != nil {
		return nil, err
	}

	// 3. 写入缓存（仅当有数据时）
	if len(orders) > 0 {
		data, _ := json.Marshal(orders)
		// 设置过期时间，例如 20 秒
		err := Redis.SetEX(ctx, cacheKey, string(data), 20)
		if err != nil {
			g.Log().Errorf(ctx, "Redis SetEX Error: %v", err)
		} else {
			g.Log().Info(ctx, "✅ Cache Set Success for user:", uid)
		}
	}

	return orders, nil
}

// CancelOrder 撤单
func (s *orderImpl) CancelOrder(ctx context.Context, userID, orderID uint64) error {
	// 查询订单
	order, err := dao.Order.GetByID(ctx, orderID)
	if err != nil {
		return err
	}

	// 验证订单所有权
	if order.UserID != userID {
		return gerror.New("无权操作此订单")
	}

	// 验证订单状态
	if order.Status != "open" {
		return gerror.New("订单状态不允许撤销")
	}

	// 更新订单状态
	err = dao.Order.Update(ctx, orderID, g.Map{
		"status": "cancelled",
	})
	if err != nil {
		return err
	}

	// 清除用户订单缓存
	s.clearUserOrderCache(ctx, userID, order.Symbol)

	return nil
}

// clearUserOrderCache 清除用户订单缓存
func (s *orderImpl) clearUserOrderCache(ctx context.Context, uid uint64, symbol string) {
	// 1. 清除该用户"所有订单"的缓存 (symbol为空的情况)
	Redis.Del(ctx, fmt.Sprintf("user:orders:%d:", uid))

	// 2. 如果指定了 symbol，清除该 symbol 的缓存
	if symbol != "" {
		Redis.Del(ctx, fmt.Sprintf("user:orders:%d:%s", uid, symbol))
	}
}
