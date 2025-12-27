package main

import (
	"backend-v2/internal/biz"
	"backend-v2/internal/conf"
	"backend-v2/internal/data"
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"
)

// TestConcurrentOrderQueries 测试并发查询订单
func TestConcurrentOrderQueries(t *testing.T) {
	// 1. 初始化配置
	configPath := "../configs/config.yaml"
	if path := os.Getenv("CONFIG_PATH"); path != "" {
		configPath = path
	}
	bc, err := conf.Load(configPath)
	if err != nil {
		t.Fatalf("failed to load config from %s: %v", configPath, err)
	}

	// 2. 初始化数据层
	db := data.NewDB(bc)
	rdb := data.NewRedis(bc)
	d, cleanup, err := data.NewData(db, rdb)
	if err != nil {
		t.Fatalf("failed to init data: %v", err)
	}
	defer cleanup()

	repo := data.NewOrderRepo(d)
	ctx := context.Background()

	// 3. 并发查询测试
	const numGoroutines = 10
	var wg sync.WaitGroup

	startTime := time.Now()

	// 启动多个 goroutine 并发查询订单
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(userID uint64) {
			defer wg.Done()

			// 查询订单信息
			orders, err := repo.FindOrdersWithUserInfo(ctx, userID)
			if err != nil {
				t.Errorf("Goroutine %d: FindOrdersWithUserInfo failed: %v", userID, err)
				return
			}

			fmt.Printf("Goroutine %d: 获取到 %d 个订单\n", userID, len(orders))
			for id, order := range orders {
				fmt.Printf("  订单ID: %d, 详情: %+v\n", id, order)
			}
		}(uint64(i + 1)) // 使用不同的用户ID
	}

	wg.Wait()

	elapsed := time.Since(startTime)
	fmt.Printf("✅ 并发查询完成，总耗时: %v\n", elapsed)
}

// TestConcurrentOrderOperations 测试并发订单操作
func TestConcurrentOrderOperations(t *testing.T) {
	// 1. 初始化配置
	configPath := "../configs/config.yaml"
	if path := os.Getenv("CONFIG_PATH"); path != "" {
		configPath = path
	}
	bc, err := conf.Load(configPath)
	if err != nil {
		t.Fatalf("failed to load config from %s: %v", configPath, err)
	}

	// 2. 初始化数据层
	db := data.NewDB(bc)
	rdb := data.NewRedis(bc)
	d, cleanup, err := data.NewData(db, rdb)
	if err != nil {
		t.Fatalf("failed to init data: %v", err)
	}
	defer cleanup()

	orderRepo := data.NewOrderRepo(d)
	userRepo := data.NewUserRepo(d)
	ctx := context.Background()

	// 3. 查找一个现有的测试用户（因为UserRepo没有Create方法）
	user, err := userRepo.FindByUsername(ctx, "admin")
	if err != nil || user == nil {
		// 如果没有找到admin用户，尝试查找其他用户
		users, err := userRepo.FindUserList(ctx)
		if err != nil || len(users) == 0 {
			t.Skip("没有找到测试用户，跳过此测试")
			return
		}
		user = users[0]
		fmt.Printf("使用现有用户进行测试: %+v\n", user)
	} else {
		fmt.Printf("使用用户进行测试: %+v\n", user)
	}

	// 4. 并发创建订单测试
	const numOrders = 5
	var wg sync.WaitGroup
	results := make(chan *biz.Order, numOrders)

	startTime := time.Now()

	for i := 0; i < numOrders; i++ {
		wg.Add(1)
		go func(orderNum int) {
			defer wg.Done()

			// 创建订单
			order := &biz.Order{
				UserID: user.ID,
				Symbol: "BTC/USDT",
				Side:   biz.OrderSideBuy,
				Type:   biz.OrderTypeLimit,
				Price:  50000.0 + float64(orderNum),
				Amount: 0.1,
				Filled: 0,
				Status: biz.OrderStatusOpen,
			}

			createdOrder, err := orderRepo.Save(ctx, order)
			if err != nil {
				t.Errorf("Goroutine %d: 创建订单失败: %v", orderNum, err)
				return
			}

			fmt.Printf("Goroutine %d: 创建订单成功, ID: %d\n", orderNum, createdOrder.ID)
			results <- createdOrder
		}(i)
	}

	// 在另一个 goroutine 中等待所有操作完成
	go func() {
		wg.Wait()
		close(results)
	}()

	// 收集结果
	createdOrders := make([]*biz.Order, 0, numOrders)
	for order := range results {
		createdOrders = append(createdOrders, order)
	}

	elapsed := time.Since(startTime)
	fmt.Printf("✅ 并发创建订单完成，共创建 %d 个订单，总耗时: %v\n", len(createdOrders), elapsed)

	// 5. 并发查询刚刚创建的订单
	fmt.Println("开始并发查询刚刚创建的订单...")

	var queryWg sync.WaitGroup
	queryResults := make(chan map[uint64]*biz.Order, numOrders)

	for _, order := range createdOrders {
		queryWg.Add(1)
		go func(userID uint64) {
			defer queryWg.Done()

			// 查询用户的所有订单信息
			orders, err := orderRepo.FindOrdersWithUserInfo(ctx, userID)
			if err != nil {
				t.Errorf("查询用户 %d 的订单失败: %v", userID, err)
				return
			}

			fmt.Printf("查询用户 %d 的订单，共 %d 个\n", userID, len(orders))
			queryResults <- orders
		}(order.UserID)
	}

	// 等待查询完成
	go func() {
		queryWg.Wait()
		close(queryResults)
	}()

	// 收集查询结果
	totalQueriedOrders := 0
	for orders := range queryResults {
		totalQueriedOrders += len(orders)
	}

	fmt.Printf("✅ 并发查询完成，共查询到 %d 个订单\n", totalQueriedOrders)
}

// TestConcurrentOrderUpdates 测试并发订单更新
func TestConcurrentOrderUpdates(t *testing.T) {
	// 1. 初始化配置
	configPath := "../configs/config.yaml"
	if path := os.Getenv("CONFIG_PATH"); path != "" {
		configPath = path
	}
	bc, err := conf.Load(configPath)
	if err != nil {
		t.Fatalf("failed to load config from %s: %v", configPath, err)
	}

	// 2. 初始化数据层
	db := data.NewDB(bc)
	rdb := data.NewRedis(bc)
	d, cleanup, err := data.NewData(db, rdb)
	if err != nil {
		t.Fatalf("failed to init data: %v", err)
	}
	defer cleanup()

	orderRepo := data.NewOrderRepo(d)
	userRepo := data.NewUserRepo(d)
	ctx := context.Background()

	// 3. 查找一个现有的测试用户
	user, err := userRepo.FindByUsername(ctx, "admin")
	if err != nil || user == nil {
		// 如果没有找到admin用户，尝试查找其他用户
		users, err := userRepo.FindUserList(ctx)
		if err != nil || len(users) == 0 {
			t.Skip("没有找到测试用户，跳过此测试")
			return
		}
		user = users[0]
		fmt.Printf("使用现有用户进行测试: %+v\n", user)
	} else {
		fmt.Printf("使用用户进行测试: %+v\n", user)
	}

	// 4. 创建几个测试订单
	orders := make([]*biz.Order, 3)
	for i := range orders {
		order := &biz.Order{
			UserID: user.ID,
			Symbol: "ETH/USDT",
			Side:   biz.OrderSideSell,
			Type:   biz.OrderTypeLimit,
			Price:  3000.0 + float64(i),
			Amount: 1.0,
			Filled: 0,
			Status: biz.OrderStatusOpen,
		}
		createdOrder, err := orderRepo.Save(ctx, order)
		if err != nil {
			t.Fatalf("创建订单失败: %v", err)
		}
		orders[i] = createdOrder
		fmt.Printf("创建订单: %+v\n", createdOrder)
	}

	// 5. 并发更新订单状态
	var wg sync.WaitGroup
	statuses := []biz.OrderStatus{
		biz.OrderStatusPartiallyFilled,
		biz.OrderStatusFilled,
		biz.OrderStatusCancelled,
	}

	startTime := time.Now()

	for i, order := range orders {
		wg.Add(1)
		go func(orderID uint64, status biz.OrderStatus) {
			defer wg.Done()

			// 更新订单状态
			err := orderRepo.UpdateStatus(ctx, orderID, status)
			if err != nil {
				t.Errorf("更新订单 %d 状态失败: %v", orderID, err)
				return
			}

			fmt.Printf("订单 %d 状态已更新为: %s\n", orderID, status)
		}(order.ID, statuses[i])
	}

	wg.Wait()

	elapsed := time.Since(startTime)
	fmt.Printf("✅ 并发更新订单状态完成，总耗时: %v\n", elapsed)

	// 6. 验证更新结果
	fmt.Println("验证更新结果...")
	for _, order := range orders {
		updatedOrder, err := orderRepo.FindByID(ctx, order.ID)
		if err != nil {
			t.Errorf("查询订单 %d 失败: %v", order.ID, err)
			continue
		}
		fmt.Printf("订单 %d 状态: %s\n", updatedOrder.ID, updatedOrder.Status)
	}
}
