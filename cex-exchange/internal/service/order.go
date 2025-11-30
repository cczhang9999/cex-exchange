package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/shopspring/decimal"

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

	// 1. 解析交易对，确定需要冻结的资产
	parts := strings.Split(req.Symbol, "/")
	if len(parts) != 2 {
		return 0, gerror.New("无效的交易对")
	}
	baseCurrency := parts[0]   // 例如 BTC
	quoteCurrency := parts[1]  // 例如 USDT

	// 2. 解析价格和数量
	price, err := decimal.NewFromString(req.Price)
	if err != nil || price.LessThanOrEqual(decimal.Zero) {
		return 0, gerror.New("无效的价格")
	}
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil || amount.LessThanOrEqual(decimal.Zero) {
		return 0, gerror.New("无效的数量")
	}

	// 3. 确定冻结资产和金额
	var freezeAsset string
	var freezeAmount decimal.Decimal
	if req.Side == "buy" {
		freezeAsset = quoteCurrency
		freezeAmount = price.Mul(amount) // 买入需要冻结 USDT
	} else {
		freezeAsset = baseCurrency
		freezeAmount = amount // 卖出需要冻结 BTC
	}

	// 4. Redis 预检查（可选，提升性能）
	cacheKey := fmt.Sprintf("user:balance:%d:%s", uid, freezeAsset)
	cachedBalance, err := Redis.Get(ctx, cacheKey)
	if err == nil && cachedBalance != "" {
		// 缓存命中，快速检查
		balance, _ := decimal.NewFromString(cachedBalance)
		if balance.LessThan(freezeAmount) {
			g.Log().Warningf(ctx, "Redis预检查: 余额不足 %s < %s", balance.String(), freezeAmount.String())
			return 0, gerror.Newf("余额不足，当前可用: %s %s, 需要: %s %s", 
				balance.String(), freezeAsset, freezeAmount.String(), freezeAsset)
		}
	}

	var orderID int64

	// 5. 数据库事务：精确检查并扣减余额
	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 5.1 查询账户并加锁
		var account model.Account
		err := tx.Model("accounts").
			Ctx(ctx).
			Where("user_id", uid).
			Where("asset", freezeAsset).
			LockUpdate(). // FOR UPDATE 排他锁
			Scan(&account)

		// GoFrame 的 Scan 在没有结果时不会返回错误，而是返回零值
		// 所以我们只需要检查 account.ID 是否为 0
		if err != nil {
			// 真正的数据库错误（连接失败、语法错误等）
			g.Log().Errorf(ctx, "查询账户没有余额: %v", err)
			return gerror.Newf("查询账户没有%s余额，请充值%s", freezeAsset, freezeAsset)
		}

		if account.ID == 0 {
			// 账户不存在
			return gerror.Newf("您还没有 %s 账户，请先充值", freezeAsset)
		}

		// 5.2 检查余额
		availableBalance, _ := decimal.NewFromString(account.Balance)
		if availableBalance.LessThan(freezeAmount) {
			return gerror.Newf("余额不足，当前可用: %s %s, 需要: %s %s", 
				availableBalance.String(), freezeAsset, freezeAmount.String(), freezeAsset)
		}

		// 5.3 扣减余额，增加冻结（原子操作）
		_, err = tx.Model("accounts").
			Ctx(ctx).
			Where("id", account.ID).
			Data(g.Map{
				"balance": gdb.Raw(fmt.Sprintf("balance - %s", freezeAmount.String())),
				"frozen":  gdb.Raw(fmt.Sprintf("frozen + %s", freezeAmount.String())),
			}).Update()

		if err != nil {
			return err
		}

		// 5.4 更新 Redis 缓存
		newBalance := availableBalance.Sub(freezeAmount)
		Redis.SetEX(ctx, cacheKey, newBalance.String(), 300) // 5分钟过期

		// 5.5 插入订单
		result, err := tx.Model("orders").Ctx(ctx).Data(g.Map{
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
			return err
		}

		orderID, _ = result.LastInsertId()
		return nil
	})

	if err != nil {
		return 0, err
	}

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
