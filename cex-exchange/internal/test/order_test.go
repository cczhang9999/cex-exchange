package test

import (
	"context"
	"fmt"
	"github.com/goccy/go-json"
	"testing"
	"time"

	"cex-exchange/internal/model"
	"cex-exchange/internal/service"
	"github.com/gogf/gf/v2/os/gtime"
)

// TestListMyOrders 测试ListMyOrders功能
func TestListMyOrders1(t *testing.T) {
	// 1. 初始化上下文和Redis连接
	ctx := context.Background()

	// 初始化Redis连接
	service.Redis.Init()

	// 测试用户ID
	userID := uint64(1)
	symbol := "BTC/USDT"

	fmt.Printf("=== 开始测试 ListMyOrders ===\n")
	fmt.Printf("用户ID: %d, 交易对: %s\n", userID, symbol)

	// 2. 清除之前的测试缓存
	cacheKey := fmt.Sprintf("user:orders:%d:%s", userID, symbol)
	err := service.Redis.Del(ctx, cacheKey)
	if err != nil {
		fmt.Printf("清除缓存失败: %v\n", err)
	}

	// 3. 第一次调用 - 应该缓存未命中，从数据库查询
	fmt.Printf("\n--- 第一次调用 (缓存未命中) ---\n")
	startTime := time.Now()
	orders1, err := service.Order.ListMyOrders(ctx, userID, symbol)
	if err != nil {
		t.Fatalf("ListMyOrders 失败: %v", err)
	}
	duration1 := time.Since(startTime)

	fmt.Printf("查询耗时: %v\n", duration1)
	fmt.Printf("返回订单数量: %d\n", len(orders1))

	// 打印订单详情
	for i, order := range orders1 {
		fmt.Printf("订单[%d]: ID=%d, Symbol=%s, Side=%s, Price=%s, Amount=%s, Status=%s, CreatedAt=%s\n",
			i, order.ID, order.Symbol, order.Side, order.Price, order.Amount, order.Status, order.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	// 4. 验证缓存是否已设置
	cachedData, err := service.Redis.Get(ctx, cacheKey)
	if err != nil {
		fmt.Printf("读取缓存失败: %v\n", err)
	} else if cachedData != "" {
		fmt.Printf("✅ 缓存已设置，数据长度: %d 字节\n", len(cachedData))
	} else {
		fmt.Printf("⚠️ 缓存未设置\n")
	}

	// 5. 第二次调用 - 应该缓存命中
	fmt.Printf("\n--- 第二次调用 (缓存命中) ---\n")
	startTime = time.Now()
	orders2, err := service.Order.ListMyOrders(ctx, userID, symbol)
	if err != nil {
		t.Fatalf("ListMyOrders 失败: %v", err)
	}
	duration2 := time.Since(startTime)

	fmt.Printf("查询耗时: %v\n", duration2)
	fmt.Printf("返回订单数量: %d\n", len(orders2))

	// 6. 性能对比
	if duration2 < duration1 {
		fmt.Printf("✅ 缓存生效！第二次比第一次快 %v\n", duration1-duration2)
	} else {
		fmt.Printf("⚠️ 缓存可能未生效\n")
	}

	// 7. 验证数据一致性
	if len(orders1) == len(orders2) {
		fmt.Printf("✅ 数据一致性检查通过\n")
	} else {
		fmt.Printf("❌ 数据不一致: 第一次%d个，第二次%d个\n", len(orders1), len(orders2))
	}

	// 8. 测试不同交易对
	fmt.Printf("\n--- 测试不同交易对 ---\n")
	ethSymbol := "ETH/USDT"
	ethOrders, err := service.Order.ListMyOrders(ctx, userID, ethSymbol)
	if err != nil {
		t.Fatalf("ListMyOrders ETH 失败: %v", err)
	}
	fmt.Printf("ETH/USDT 订单数量: %d\n", len(ethOrders))

	// 9. 测试所有交易对（symbol为空）
	fmt.Printf("\n--- 测试所有交易对 ---\n")
	allOrders, err := service.Order.ListMyOrders(ctx, userID, "")
	if err != nil {
		t.Fatalf("ListMyOrders 所有交易对失败: %v", err)
	}
	fmt.Printf("所有交易对订单数量: %d\n", len(allOrders))

	fmt.Printf("\n=== 测试完成 ===\n")
}

// TestListMyOrdersWithMockData 使用模拟数据测试
func TestListMyOrdersWithMockData(t *testing.T) {
	ctx := context.Background()
	service.Redis.Init()

	userID := uint64(1)
	symbol := "BTC/USDT"

	fmt.Printf("=== 使用模拟数据测试 ===\n")

	// 创建模拟订单数据
	mockOrders := []model.Order{
		{
			ID:        1001,
			UserID:    userID,
			Symbol:    symbol,
			Side:      "buy",
			Type:      "limit",
			Price:     "50000.00",
			Amount:    "0.1",
			Filled:    "0.05",
			Status:    "partially_filled",
			CreatedAt: gtime.New(time.Now().Add(-2 * time.Hour)),
			UpdatedAt: gtime.New(time.Now().Add(-1 * time.Hour)),
		},
		{
			ID:        1002,
			UserID:    userID,
			Symbol:    symbol,
			Side:      "sell",
			Type:      "limit",
			Price:     "51000.00",
			Amount:    "0.2",
			Filled:    "0.0",
			Status:    "open",
			CreatedAt: gtime.New(time.Now().Add(-1 * time.Hour)),
			UpdatedAt: gtime.New(time.Now()),
		},
	}

	// 将模拟数据存入缓存
	cacheKey := fmt.Sprintf("user:orders:%d:%s", userID, symbol)
	mockData, _ := json.Marshal(mockOrders)
	err := service.Redis.SetEX(ctx, cacheKey, string(mockData), 60)
	if err != nil {
		t.Fatalf("设置模拟缓存失败: %v", err)
	}

	fmt.Printf("✅ 模拟数据已存入缓存\n")

	// 调用ListMyOrders，应该返回缓存数据
	orders, err := service.Order.ListMyOrders(ctx, userID, symbol)
	if err != nil {
		t.Fatalf("ListMyOrders 失败: %v", err)
	}

	fmt.Printf("返回订单数量: %d\n", len(orders))
	for i, order := range orders {
		fmt.Printf("订单[%d]: ID=%d, Side=%s, Price=%s, Amount=%s, Status=%s\n",
			i, order.ID, order.Side, order.Price, order.Amount, order.Status)
	}

	// 验证返回的数据与模拟数据一致
	if len(orders) == len(mockOrders) {
		fmt.Printf("✅ 缓存数据返回正确\n")
	} else {
		fmt.Printf("❌ 数据不匹配: 期望%d个，实际%d个\n", len(mockOrders), len(orders))
	}
}
