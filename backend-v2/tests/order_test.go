package main

import (
	pb "backend-v2/api/proto"
	"backend-v2/internal/biz"
	"backend-v2/internal/conf"
	"backend-v2/internal/data"
	"backend-v2/internal/service"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"os"
	"testing"
)

// TestGetUserOrdersViaService 直接调用 service 层的 GetUserOrders 方法（单元测试）
func TestGetUserOrdersViaService(t *testing.T) {
	// 1. 初始化配置
	configPath := "../configs/config.yaml"
	if path := os.Getenv("CONFIG_PATH"); path != "" {
		configPath = path
	}
	bc, err := conf.Load(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// 2. 初始化数据层
	db := data.NewDB(bc)
	rdb := data.NewRedis(bc)
	d, cleanup, err := data.NewData(db, rdb)
	if err != nil {
		t.Fatalf("failed to init data: %v", err)
	}
	defer cleanup()

	// 3. 创建 repo 和 usecase
	orderRepo := data.NewOrderRepo(d)
	userRepo := data.NewUserRepo(d)

	orderUsecase := biz.NewOrderUsecase(orderRepo)
	userUsecase := biz.NewUserUsecase(userRepo, bc)

	// 4. 创建 exchange client（可以传 nil 如果不需要调用外部服务）
	var exchangeClient pb.ExchangeServiceClient
	if bc.Client != nil && bc.Client.Exchange != nil {
		conn, err := grpc.Dial(bc.Client.Exchange.Addr,
			grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			t.Logf("警告: 无法连接到 exchange 服务: %v", err)
		} else {
			defer conn.Close()
			exchangeClient = pb.NewExchangeServiceClient(conn)
		}
	}

	// 5. 创建 ExchangeService 实例
	exchangeService := service.NewExchangeService(userUsecase, orderUsecase, exchangeClient)

	// 6. 调用 GetUserOrders 方法
	ctx := context.Background()
	resp, err := exchangeService.GetUserOrders(ctx, &pb.GetMyOrdersRequest{
		Token: "mock-token",
	})
	if err != nil {
		t.Fatalf("GetUserOrders 失败: %v", err)
	}

	// 7. 验证结果
	if !resp.Success {
		t.Fatalf("GetUserOrders 返回失败: %s", resp.Message)
	}

	fmt.Printf("✅ 成功获取用户订单，共 %d 条\n", len(resp.Orders))
	for _, order := range resp.Orders {
		fmt.Printf("订单详情: ID=%d, 用户ID=%d, 交易对=%s, 方向=%s, 类型=%s, 价格=%s, 数量=%s, 已成交=%s, 状态=%s\n",
			order.Id, order.UserId, order.Symbol, order.Side, order.Type,
			order.Price, order.Amount, order.Filled, order.Status)
	}
}

// TestCreateOrder 原有的测试（修复方法名）
func TestCreateOrder(t *testing.T) {
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
	// 3. 调用并验证
	ctx := context.Background()

	// 修复：使用正确的方法名 FindByUserID
	orderList, err := repo.FindByUserID(ctx, 1)
	if err != nil {
		t.Fatalf("FindByUserID failed: %v", err)
	}
	for _, o := range orderList {
		fmt.Printf("Order: %+v\n", o)
	}

	orderMyList, err := repo.FindByUserID(ctx, 1)
	if err != nil {
		t.Fatalf("FindByUserID failed: %v", err)
	}
	for _, o := range orderMyList {
		fmt.Printf("Order: %+v\n", o)
	}
}
