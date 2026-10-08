package main

import (
	"backend-v2/internal/biz"
	"context"
	"fmt"
	"sort"
	"sync"
	"testing"
	"time"
)

// ==================== 基础断言工具 ====================

func assertEqual(t *testing.T, got, want interface{}, msg ...string) {
	t.Helper()
	if got != want {
		prefix := ""
		if len(msg) > 0 {
			prefix = msg[0] + ": "
		}
		t.Errorf("%sgot %v, want %v", prefix, got, want)
	}
}

func assertNotNil(t *testing.T, obj interface{}, msg ...string) {
	t.Helper()
	if obj == nil {
		prefix := ""
		if len(msg) > 0 {
			prefix = msg[0] + ": "
		}
		t.Errorf("%sexpected non-nil", prefix)
	}
}

func assertTrue(t *testing.T, cond bool, msg ...string) {
	t.Helper()
	if !cond {
		prefix := ""
		if len(msg) > 0 {
			prefix = msg[0] + ": "
		}
		t.Errorf("%sexpected true", prefix)
	}
}

// ==================== 示例1：基础测试函数 ====================

func TestHello(t *testing.T) {
	fmt.Println("=== 示例1：最简单的测试函数 ===")
	fmt.Println("Go 的测试函数以 Test 开头，参数为 *testing.T")
	fmt.Println("运行方式：go test -v -run TestHello ./tests/")
	fmt.Println()
}

// ==================== 示例2：Table-Driven 测试 ====================

func TestOrderSideString(t *testing.T) {
	fmt.Println("=== 示例2：Table-Driven 测试 ===")
	fmt.Println("Table-Driven 测试是 Go 中最推荐的测试模式，将测试数据与逻辑分离")
	fmt.Println()

	tests := []struct {
		name     string
		side     biz.OrderSide
		expected string
	}{
		{"买入方向", biz.OrderSideBuy, "buy"},
		{"卖出方向", biz.OrderSideSell, "sell"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(tt.side)
			assertEqual(t, got, tt.expected, "OrderSide 字符串值")
			fmt.Printf("  ✓ %s: %s == %s\n", tt.name, got, tt.expected)
		})
	}
	fmt.Println()
}

// ==================== 示例3：子测试（Subtest）===================

func TestOrderTypeString(t *testing.T) {
	fmt.Println("=== 示例3：子测试（Subtest）===")
	fmt.Println("t.Run() 创建子测试，可以独立运行和报告")
	fmt.Println()

	t.Run("Limit", func(t *testing.T) {
		assertEqual(t, string(biz.OrderTypeLimit), "limit", "OrderTypeLimit")
		fmt.Println("  ✓ Limit 类型正确")
	})

	t.Run("Market", func(t *testing.T) {
		assertEqual(t, string(biz.OrderTypeMarket), "market", "OrderTypeMarket")
		fmt.Println("  ✓ Market 类型正确")
	})
	fmt.Println()
}

// ==================== 示例4：内存 Mock 实现 ====================

type learningFakeRepo struct {
	orders map[uint64]*biz.Order
}

func newLearningFakeRepo() *learningFakeRepo {
	return &learningFakeRepo{orders: map[uint64]*biz.Order{}}
}

func (r *learningFakeRepo) Save(ctx context.Context, order *biz.Order) (*biz.Order, error) {
	if order.ID == 0 {
		order.ID = uint64(len(r.orders) + 1)
	}
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()
	r.orders[order.ID] = order
	return order, nil
}

func (r *learningFakeRepo) FindByID(ctx context.Context, id uint64) (*biz.Order, error) {
	if o, ok := r.orders[id]; ok {
		return o, nil
	}
	return nil, fmt.Errorf("order %d not found", id)
}

func (r *learningFakeRepo) FindByUserID(ctx context.Context, userID uint64) ([]*biz.Order, error) {
	var out []*biz.Order
	for _, o := range r.orders {
		if o.UserID == userID {
			out = append(out, o)
		}
	}
	return out, nil
}

func (r *learningFakeRepo) FindPage(ctx context.Context, q biz.OrderQuery) (*biz.OrderPage, error) {
	var matched []*biz.Order
	for _, o := range r.orders {
		if o.UserID != q.UserID {
			continue
		}
		if q.Symbol != "" && o.Symbol != q.Symbol {
			continue
		}
		if q.Status != "" && o.Status != q.Status {
			continue
		}
		matched = append(matched, o)
	}
	sort.Slice(matched, func(i, j int) bool { return matched[i].ID < matched[j].ID })

	page, pageSize := q.Page, q.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	start := (page - 1) * pageSize
	if start > len(matched) {
		start = len(matched)
	}
	end := start + pageSize
	if end > len(matched) {
		end = len(matched)
	}
	return &biz.OrderPage{
		Total:    int64(len(matched)),
		Page:     page,
		PageSize: pageSize,
		Orders:   matched[start:end],
	}, nil
}

func (r *learningFakeRepo) UpdateStatus(ctx context.Context, id uint64, status biz.OrderStatus) error {
	if o, ok := r.orders[id]; ok {
		o.Status = status
		o.UpdatedAt = time.Now()
	}
	return nil
}

func (r *learningFakeRepo) UpdateFilled(ctx context.Context, id uint64, filled float64) error {
	if o, ok := r.orders[id]; ok {
		o.Filled = filled
		o.UpdatedAt = time.Now()
	}
	return nil
}

func (r *learningFakeRepo) FindOpenOrders(ctx context.Context, symbol string) ([]*biz.Order, error) {
	var out []*biz.Order
	for _, o := range r.orders {
		if o.Symbol == symbol && o.Status == biz.OrderStatusOpen {
			out = append(out, o)
		}
	}
	return out, nil
}

func (r *learningFakeRepo) FindByUserIDWithUser(ctx context.Context, userID uint64) ([]*biz.Order, error) {
	return r.FindByUserID(ctx, userID)
}

func (r *learningFakeRepo) FindOrdersWithUserInfo(ctx context.Context, userID uint64) (map[uint64]*biz.Order, error) {
	out := map[uint64]*biz.Order{}
	orders, _ := r.FindByUserID(ctx, userID)
	for _, o := range orders {
		out[o.ID] = o
	}
	return out, nil
}

func (r *learningFakeRepo) FindOrdersWithUserUsingJoins(ctx context.Context, symbol string) ([]*biz.Order, error) {
	var out []*biz.Order
	for _, o := range r.orders {
		if o.Symbol == symbol {
			out = append(out, o)
		}
	}
	return out, nil
}

// ==================== 示例4：Mock/Stub 测试 ====================

func TestUsecaseCreateOrder(t *testing.T) {
	fmt.Println("=== 示例4：Mock/Stub 测试 ===")
	fmt.Println("使用内存 fakeRepo 隔离外部依赖，专注测试业务逻辑")
	fmt.Println()

	repo := newLearningFakeRepo()
	uc := biz.NewOrderUsecase(repo)
	ctx := context.Background()

	// 测试创建订单
	order, err := uc.CreateOrder(ctx, 1, "BTC/USDT", biz.OrderSideBuy, biz.OrderTypeLimit, 65000.0, 0.1)
	if err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}

	assertEqual(t, order.UserID, uint64(1), "UserID")
	assertEqual(t, order.Symbol, "BTC/USDT", "Symbol")
	assertEqual(t, order.Side, biz.OrderSideBuy, "Side")
	assertEqual(t, order.Type, biz.OrderTypeLimit, "Type")
	assertEqual(t, order.Price, 65000.0, "Price")
	assertEqual(t, order.Amount, 0.1, "Amount")
	assertEqual(t, order.Status, biz.OrderStatusOpen, "Status")
	assertNotNil(t, order.CreatedAt, "CreatedAt")
	fmt.Printf("  ✓ 创建订单成功: ID=%d, Symbol=%s, Side=%s\n", order.ID, order.Symbol, order.Side)
	fmt.Println()
}

// ==================== 示例5：Usecase GetOrder ====================

func TestUsecaseGetOrder(t *testing.T) {
	fmt.Println("=== 示例5：Usecase GetOrder ===")
	fmt.Println("测试通过 ID 查询订单")
	fmt.Println()

	repo := newLearningFakeRepo()
	uc := biz.NewOrderUsecase(repo)
	ctx := context.Background()

	// 先创建一个订单
	created, _ := uc.CreateOrder(ctx, 1, "ETH/USDT", biz.OrderSideSell, biz.OrderTypeMarket, 3000.0, 1.0)
	fmt.Printf("  预置订单: ID=%d, Symbol=%s\n", created.ID, created.Symbol)

	// 再通过 GetOrder 查询
	got, err := uc.GetOrder(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetOrder failed: %v", err)
	}

	assertEqual(t, got.ID, created.ID, "ID")
	assertEqual(t, got.Symbol, "ETH/USDT", "Symbol")
	fmt.Printf("  ✓ 查询订单成功: ID=%d, Symbol=%s\n", got.ID, got.Symbol)
	fmt.Println()
}

// ==================== 示例6：Usecase CancelOrder ====================

func TestUsecaseCancelOrder(t *testing.T) {
	fmt.Println("=== 示例6：Usecase CancelOrder ===")
	fmt.Println("测试取消订单的业务逻辑")
	fmt.Println()

	repo := newLearningFakeRepo()
	uc := biz.NewOrderUsecase(repo)
	ctx := context.Background()

	created, _ := uc.CreateOrder(ctx, 1, "BTC/USDT", biz.OrderSideBuy, biz.OrderTypeLimit, 65000.0, 0.5)
	fmt.Printf("  创建订单: ID=%d, Status=%s\n", created.ID, created.Status)

	err := uc.CancelOrder(ctx, created.ID)
	if err != nil {
		t.Fatalf("CancelOrder failed: %v", err)
	}

	// 验证状态已更新
	updated, _ := uc.GetOrder(ctx, created.ID)
	assertEqual(t, updated.Status, biz.OrderStatusCancelled, "Status after cancel")
	fmt.Printf("  ✓ 取消订单成功: ID=%d, Status=%s\n", updated.ID, updated.Status)
	fmt.Println()
}

// ==================== 示例7：Table-Driven + Subtest 综合 ====================

func TestOrderUsecaseGetUserOrders(t *testing.T) {
	fmt.Println("=== 示例7：Table-Driven + Subtest 综合 ===")
	fmt.Println("结合 Table-Driven 和 Subtest 测试 GetUserOrders")
	fmt.Println()

	tests := []struct {
		name      string
		userID    uint64
		setup     func(repo *learningFakeRepo)
		wantCount int
	}{
		{
			name:   "用户1有2个订单",
			userID: 1,
			setup: func(repo *learningFakeRepo) {
				_, _ = repo.Save(context.Background(), &biz.Order{UserID: 1, Symbol: "BTC/USDT", Side: biz.OrderSideBuy, Type: biz.OrderTypeLimit, Price: 65000, Amount: 0.1, Status: biz.OrderStatusOpen})
				_, _ = repo.Save(context.Background(), &biz.Order{UserID: 1, Symbol: "ETH/USDT", Side: biz.OrderSideSell, Type: biz.OrderTypeMarket, Price: 3000, Amount: 1.0, Status: biz.OrderStatusOpen})
			},
			wantCount: 2,
		},
		{
			name:   "用户2没有订单",
			userID: 2,
			setup: func(repo *learningFakeRepo) {
				_, _ = repo.Save(context.Background(), &biz.Order{UserID: 1, Symbol: "BTC/USDT", Side: biz.OrderSideBuy, Type: biz.OrderTypeLimit, Price: 65000, Amount: 0.1, Status: biz.OrderStatusOpen})
			},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newLearningFakeRepo()
			tt.setup(repo)
			uc := biz.NewOrderUsecase(repo)
			ctx := context.Background()

			orders, err := uc.GetUserOrders(ctx, tt.userID)
			if err != nil {
				t.Fatalf("GetUserOrders failed: %v", err)
			}
			assertEqual(t, len(orders), tt.wantCount, "订单数量")
			fmt.Printf("  ✓ 用户 %d: 获取到 %d 个订单\n", tt.userID, len(orders))
		})
	}
	fmt.Println()
}

// ==================== 示例8：Benchmark 测试 ====================

func BenchmarkOrderUsecaseCreateOrder(b *testing.B) {
	fmt.Println("=== 示例8：Benchmark 测试 ===")
	fmt.Println("Benchmark 用于性能测试，b.N 会自动调整迭代次数")
	fmt.Println()

	repo := newLearningFakeRepo()
	uc := biz.NewOrderUsecase(repo)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = uc.CreateOrder(ctx, uint64(i), "BTC/USDT", biz.OrderSideBuy, biz.OrderTypeLimit, 65000.0, 0.1)
	}
	fmt.Printf("  ✓ Benchmark 完成: %d 次迭代\n", b.N)
	fmt.Println()
}

// ==================== 示例9：并行测试 ====================

func TestOrderSideParallel(t *testing.T) {
	fmt.Println("=== 示例9：并行测试（Parallel Test）===")
	fmt.Println("t.Parallel() 让子测试并行执行，加速测试运行")
	fmt.Println()

	t.Run("Buy", func(t *testing.T) {
		t.Parallel()
		time.Sleep(100 * time.Millisecond)
		assertEqual(t, string(biz.OrderSideBuy), "buy", "Buy")
		fmt.Println("  ✓ Buy 并行完成")
	})

	t.Run("Sell", func(t *testing.T) {
		t.Parallel()
		time.Sleep(100 * time.Millisecond)
		assertEqual(t, string(biz.OrderSideSell), "sell", "Sell")
		fmt.Println("  ✓ Sell 并行完成")
	})
	fmt.Println()
}

// ==================== 示例10：Skip 和 SkipAll ====================

func TestSkipExample(t *testing.T) {
	fmt.Println("=== 示例10：Skip 测试 ===")
	fmt.Println("t.Skip() 跳过当前测试，t.SkipNow() 跳过整个测试函数")
	fmt.Println()

	if testing.Short() {
		t.Skip("短模式下跳过此测试")
	}
	fmt.Println("  ✓ 非短模式，正常执行")
	fmt.Println()
}

// ==================== 示例11：错误处理测试 ====================

func TestUsecaseGetOrderNotFound(t *testing.T) {
	fmt.Println("=== 示例11：错误处理测试 ===")
	fmt.Println("测试查询不存在的订单")
	fmt.Println()

	repo := newLearningFakeRepo()
	uc := biz.NewOrderUsecase(repo)
	ctx := context.Background()

	_, err := uc.GetOrder(ctx, 9999)
	if err == nil {
		t.Error("预期查询不存在的订单返回错误，但实际返回 nil")
	} else {
		fmt.Printf("  ✓ 正确处理错误: %v\n", err)
	}
	fmt.Println()
}

// ==================== 示例12：测试辅助函数 ====================

func setupTestEnv(t *testing.T) (*learningFakeRepo, *biz.OrderUsecase, context.Context) {
	t.Helper()
	repo := newLearningFakeRepo()
	uc := biz.NewOrderUsecase(repo)
	ctx := context.Background()
	return repo, uc, ctx
}

func TestHelperExample(t *testing.T) {
	fmt.Println("=== 示例12：测试辅助函数（Helper）===")
	fmt.Println("t.Helper() 标记辅助函数，错误报告时指向调用者而非辅助函数内部")
	fmt.Println()

	_, uc, ctx := setupTestEnv(t)

	_, _ = uc.CreateOrder(ctx, 1, "BTC/USDT", biz.OrderSideBuy, biz.OrderTypeLimit, 65000.0, 0.1)
	orders, _ := uc.GetUserOrders(ctx, 1)
	assertEqual(t, len(orders), 1, "订单数量")
	fmt.Println("  ✓ Helper 函数正常工作")
	fmt.Println()
}

// ==================== 示例13：时间相关测试 ====================

func TestOrderTimestamps(t *testing.T) {
	fmt.Println("=== 示例13：时间相关测试 ===")
	fmt.Println("验证创建订单时时间戳是否正确设置")
	fmt.Println()

	repo := newLearningFakeRepo()
	uc := biz.NewOrderUsecase(repo)
	ctx := context.Background()

	before := time.Now()
	order, err := uc.CreateOrder(ctx, 1, "BTC/USDT", biz.OrderSideBuy, biz.OrderTypeLimit, 65000.0, 0.1)
	after := time.Now()

	if err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}

	assertTrue(t, !order.CreatedAt.IsZero(), "CreatedAt 不应为空")
	assertTrue(t, !order.UpdatedAt.IsZero(), "UpdatedAt 不应为空")
	assertTrue(t, order.CreatedAt.After(before) && order.CreatedAt.Before(after), "CreatedAt 应在 before 和 after 之间")
	assertTrue(t, order.UpdatedAt.After(before) && order.UpdatedAt.Before(after), "UpdatedAt 应在 before 和 after 之间")
	fmt.Printf("  ✓ CreatedAt: %v\n", order.CreatedAt)
	fmt.Printf("  ✓ UpdatedAt: %v\n", order.UpdatedAt)
	fmt.Println()
}

// ==================== 示例14：并发安全测试 ====================

func TestConcurrentOrderAccess(t *testing.T) {
	fmt.Println("=== 示例14：并发安全测试 ===")
	fmt.Println("测试并发读写是否安全（此 fakeRepo 非线程安全，仅演示模式）")
	fmt.Println()

	repo := newLearningFakeRepo()
	uc := biz.NewOrderUsecase(repo)
	ctx := context.Background()

	const numGoroutines = 10
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			_, _ = uc.CreateOrder(ctx, uint64(id), "BTC/USDT", biz.OrderSideBuy, biz.OrderTypeLimit, 65000.0, 0.1)
		}(i)
	}

	wg.Wait()
	orders, _ := uc.GetUserOrders(ctx, 1)
	fmt.Printf("  ✓ 并发创建完成，共创建 %d 个订单（fakeRepo 非线程安全，结果可能不稳定）\n", len(orders))
	fmt.Println()
}

// ==================== 示例15：测试输出格式 ====================

func TestOutputFormat(t *testing.T) {
	fmt.Println("=== 示例15：测试输出格式 ===")
	fmt.Println("测试函数中可以自由使用 fmt.Printf 打印调试信息")
	fmt.Println("运行：go test -v -run TestOutputFormat ./tests/")
	fmt.Println()

	order := &biz.Order{
		ID:     1001,
		UserID: 2001,
		Symbol: "BTC/USDT",
		Side:   biz.OrderSideBuy,
		Type:   biz.OrderTypeLimit,
		Price:  65000.0,
		Amount: 0.5,
		Status: biz.OrderStatusOpen,
	}

	fmt.Printf("  订单详情:\n")
	fmt.Printf("    ID:     %d\n", order.ID)
	fmt.Printf("    UserID: %d\n", order.UserID)
	fmt.Printf("    Symbol: %s\n", order.Symbol)
	fmt.Printf("    Side:   %s\n", order.Side)
	fmt.Printf("    Type:   %s\n", order.Type)
	fmt.Printf("    Price:  %.2f\n", order.Price)
	fmt.Printf("    Amount: %.2f\n", order.Amount)
	fmt.Printf("    Status: %s\n", order.Status)
	fmt.Println()
}
