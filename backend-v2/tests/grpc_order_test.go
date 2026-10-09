package main

import (
	"context"
	"net"
	"testing"
	"time"

	pb "backend-v2/api/proto"
	"backend-v2/internal/biz"
	"backend-v2/internal/conf"
	"backend-v2/internal/pkg/jwt"
	"backend-v2/internal/server/middleware"
	"backend-v2/internal/service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"
)

// grpcTestConfig 创建一个用于测试的 *conf.Bootstrap，包含 JWT secret
func grpcTestConfig(jwtSecret string) *conf.Bootstrap {
	return &conf.Bootstrap{
		Auth: &conf.Auth{
			JwtSecret: jwtSecret,
			JwtExpiry: "24h",
		},
		Server: &conf.Server{
			Grpc: &conf.ServerGRPC{Addr: ":0"},
		},
	}
}

// startTestGRPCServer 启动一个内存 gRPC 服务器，返回监听器 + 客户端
// 使用 bufconn 实现无网络的 gRPC 本地测试（真正的远程调用路径，但无需占用端口）
func startTestGRPCServer(t *testing.T, svc *service.ExchangeService, jwtSecret string) (pb.ExchangeServiceClient, func()) {
	t.Helper()

	// 创建 gRPC server，带上 auth 拦截器
	lis := bufconn.Listen(1024 * 1024)
	grpcOpts := []grpc.ServerOption{}
	if jwtSecret != "" {
		grpcOpts = append(grpcOpts, grpc.ChainUnaryInterceptor(
			middleware.GrpcAuthInterceptor(jwtSecret),
		))
	}
	grpcServer := grpc.NewServer(grpcOpts...)
	pb.RegisterExchangeServiceServer(grpcServer, svc)

	// 用 goroutine 启动 server
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			t.Logf("grpc server stopped: %v", err)
		}
	}()

	// 创建 bufconn client
	dialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}
	conn, err := grpc.DialContext(context.Background(), "bufnet",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("failed to dial bufnet: %v", err)
	}

	client := pb.NewExchangeServiceClient(conn)

	cleanup := func() {
		conn.Close()
		grpcServer.Stop()
	}
	return client, cleanup
}

// genTestToken 生成一个用于测试的 JWT token
func genTestToken(t *testing.T, secret string, userID uint64) string {
	t.Helper()
	token, err := jwt.GenerateToken(userID, secret, 24*time.Hour)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}
	return token
}

// TestGRPCCPlaceOrderViaRemoteCall 测试通过 gRPC 远程调用下单接口
// 流程：启动 gRPC 服务器 → gRPC 客户端远程调用 PlaceOrder → 验证订单已创建
func TestGRPCCPlaceOrderViaRemoteCall(t *testing.T) {
	const jwtSecret = "test-secret-key"

	// 1. 准备 fake repo + usecase
	orderRepo := newFakeOrderRepo()
	orderUc := biz.NewOrderUsecase(orderRepo)

	// 2. 创建 service（传入配置以支持 JWT 解析）
	cfg := grpcTestConfig(jwtSecret)
	svc := service.NewExchangeService(nil, orderUc, nil, nil, nil, nil, cfg, nil)

	// 3. 启动 gRPC 服务器 + 客户端
	client, cleanup := startTestGRPCServer(t, svc, jwtSecret)
	defer cleanup()

	// 4. 生成 token
	token := genTestToken(t, jwtSecret, 1)

	// 5. 通过 gRPC 远程调用 PlaceOrder
	// token 通过 gRPC metadata 发送（标准 gRPC 认证方式），同时也放入请求体的 Token 字段作为备用
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)

	resp, err := client.PlaceOrder(ctx, &pb.PlaceOrderRequest{
		Symbol: "BTC/USDT",
		Side:   "buy",
		Type:   "limit",
		Price:  "50000.00",
		Amount: "0.5",
		Token:  token,
	})
	if err != nil {
		t.Fatalf("gRPC PlaceOrder 调用失败: %v", err)
	}
	if !resp.Success {
		t.Fatalf("PlaceOrder 返回失败: %s", resp.Message)
	}
	if resp.OrderId <= 0 {
		t.Fatalf("PlaceOrder 返回的 order_id 应大于 0，got %d", resp.OrderId)
	}
	t.Logf("✅ PlaceOrder 成功，order_id=%d, message=%s", resp.OrderId, resp.Message)

	// 6. 验证订单已写入 repo
	orders, _ := orderRepo.FindByUserID(context.Background(), 1)
	if len(orders) != 1 {
		t.Fatalf("期望 1 笔订单，实际 %d", len(orders))
	}
	if orders[0].Symbol != "BTC/USDT" || orders[0].Side != biz.OrderSideBuy {
		t.Fatalf("订单数据不匹配: %+v", orders[0])
	}
	t.Logf("📝 订单已保存到 repo: ID=%d, Symbol=%s, Price=%.2f, Amount=%.2f",
		orders[0].ID, orders[0].Symbol, orders[0].Price, orders[0].Amount)
}

// TestGRPCCCancelOrderViaRemoteCall 测试通过 gRPC 远程调用取消订单接口
func TestGRPCCancelOrderViaRemoteCall(t *testing.T) {
	const jwtSecret = "test-secret-key"

	orderRepo := newFakeOrderRepo()
	orderUc := biz.NewOrderUsecase(orderRepo)
	cfg := grpcTestConfig(jwtSecret)
	svc := service.NewExchangeService(nil, orderUc, nil, nil, nil, nil, cfg, nil)

	client, cleanup := startTestGRPCServer(t, svc, jwtSecret)
	defer cleanup()

	token := genTestToken(t, jwtSecret, 1)

	// 先创建一个订单
	ctx := context.Background()
	order, _ := orderUc.CreateOrder(ctx, 1, "ETH/USDT", biz.OrderSideSell, biz.OrderTypeLimit, 3000.0, 1.0)

	// 远程调用 CancelOrder（token 通过 metadata 发送）
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
	resp, err := client.CancelOrder(ctx, &pb.CancelOrderRequest{
		OrderId: order.ID,
		Token:   token,
	})
	if err != nil {
		t.Fatalf("gRPC CancelOrder 调用失败: %v", err)
	}
	if !resp.Success {
		t.Fatalf("CancelOrder 返回失败: %s", resp.Message)
	}
	t.Logf("✅ CancelOrder 成功: %s", resp.Message)

	// 验证状态已更新为 cancelled
	updated, _ := orderUc.GetOrder(ctx, order.ID)
	if updated.Status != biz.OrderStatusCancelled {
		t.Fatalf("期望订单状态为 cancelled，实际 %s", updated.Status)
	}
}

// TestGRPCGetMyOrdersViaRemoteCall 测试通过 gRPC 远程调用查询订单接口
func TestGRPCGetMyOrdersViaRemoteCall(t *testing.T) {
	const jwtSecret = "test-secret-key"

	orderRepo := newFakeOrderRepo()
	orderUc := biz.NewOrderUsecase(orderRepo)
	cfg := grpcTestConfig(jwtSecret)
	svc := service.NewExchangeService(nil, orderUc, nil, nil, nil, nil, cfg, nil)

	client, cleanup := startTestGRPCServer(t, svc, jwtSecret)
	defer cleanup()

	token := genTestToken(t, jwtSecret, 1)

	// 预置 3 笔订单
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		_, _ = orderUc.CreateOrder(ctx, 1, "BTC/USDT", biz.OrderSideBuy, biz.OrderTypeLimit,
			50000.0+float64(i), 0.1)
	}

	// 远程调用 GetMyOrders（token 通过 metadata 发送）
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
	resp, err := client.GetMyOrders(ctx, &pb.GetMyOrdersRequest{
		Token: token,
	})
	if err != nil {
		t.Fatalf("gRPC GetMyOrders 调用失败: %v", err)
	}
	if !resp.Success {
		t.Fatalf("GetMyOrders 返回失败: %s", resp.Message)
	}
	if len(resp.Orders) != 3 {
		t.Fatalf("期望返回 3 笔订单，实际 %d", len(resp.Orders))
	}
	t.Logf("✅ GetMyOrders 成功，返回 %d 笔订单", len(resp.Orders))
	for _, o := range resp.Orders {
		t.Logf("  订单: ID=%d, Symbol=%s, Side=%s, Price=%s, Amount=%s, Status=%s",
			o.Id, o.Symbol, o.Side, o.Price, o.Amount, o.Status)
	}
}

// TestGRPCOrderAuthFailure 测试未认证（无 token / 无效 token）时 gRPC 返回认证错误
func TestGRPCOrderAuthFailure(t *testing.T) {
	const jwtSecret = "test-secret-key"

	orderRepo := newFakeOrderRepo()
	orderUc := biz.NewOrderUsecase(orderRepo)
	cfg := grpcTestConfig(jwtSecret)
	svc := service.NewExchangeService(nil, orderUc, nil, nil, nil, nil, cfg, nil)

	client, cleanup := startTestGRPCServer(t, svc, jwtSecret)
	defer cleanup()

	ctx := context.Background()

	// 1. 无 token
	_, err := client.GetMyOrders(ctx, &pb.GetMyOrdersRequest{})
	if err == nil {
		t.Fatal("期望无 token 时返回错误，却成功了")
	}
	t.Logf("✅ 无 token 被正确拒绝: %v", err)

	// 2. 无效 token（通过 metadata 发送）
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer invalid-token")
	_, err = client.PlaceOrder(ctx, &pb.PlaceOrderRequest{
		Symbol: "BTC/USDT",
		Side:   "buy",
		Type:   "limit",
		Price:  "1",
		Amount: "1",
		Token:  "invalid-token",
	})
	if err == nil {
		t.Fatal("期望无效 token 时返回错误，却成功了")
	}
	t.Logf("✅ 无效 token 被正确拒绝: %v", err)

	// 3. 错误 secret 签名的 token（通过 metadata 发送）
	wrongToken, _ := jwt.GenerateToken(1, "wrong-secret", 24*time.Hour)
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+wrongToken)
	_, err = client.GetMyOrders(ctx, &pb.GetMyOrdersRequest{
		Token: wrongToken,
	})
	if err == nil {
		t.Fatal("期望错误签名的 token 被拒绝，却成功了")
	}
	t.Logf("✅ 错误签名的 token 被正确拒绝: %v", err)
}

// TestGRPCGetMyOrdersWithSymbolFilter 测试通过 gRPC 远程调用带 symbol 过滤的订单查询
func TestGRPCGetMyOrdersWithSymbolFilter(t *testing.T) {
	const jwtSecret = "test-secret-key"

	orderRepo := newFakeOrderRepo()
	orderUc := biz.NewOrderUsecase(orderRepo)
	cfg := grpcTestConfig(jwtSecret)
	svc := service.NewExchangeService(nil, orderUc, nil, nil, nil, nil, cfg, nil)

	client, cleanup := startTestGRPCServer(t, svc, jwtSecret)
	defer cleanup()

	token := genTestToken(t, jwtSecret, 1)
	ctx := context.Background()

	// 预置 BTC/USDT 和 ETH/USDT 的订单
	_, _ = orderUc.CreateOrder(ctx, 1, "BTC/USDT", biz.OrderSideBuy, biz.OrderTypeLimit, 50000, 0.1)
	_, _ = orderUc.CreateOrder(ctx, 1, "BTC/USDT", biz.OrderSideSell, biz.OrderTypeLimit, 51000, 0.2)
	_, _ = orderUc.CreateOrder(ctx, 1, "ETH/USDT", biz.OrderSideBuy, biz.OrderTypeLimit, 3000, 1.0)

	// 按 symbol 过滤（token 通过 metadata 发送）
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
	resp, err := client.GetMyOrders(ctx, &pb.GetMyOrdersRequest{
		Symbol: "BTC/USDT",
		Token:  token,
	})
	if err != nil {
		t.Fatalf("gRPC GetMyOrders 调用失败: %v", err)
	}
	if !resp.Success {
		t.Fatalf("GetMyOrders 返回失败: %s", resp.Message)
	}
	if len(resp.Orders) != 2 {
		t.Fatalf("期望返回 2 笔 BTC/USDT 订单，实际 %d", len(resp.Orders))
	}
	for _, o := range resp.Orders {
		if o.Symbol != "BTC/USDT" {
			t.Fatalf("过滤失败，返回了非 BTC/USDT 订单: %s", o.Symbol)
		}
	}
	t.Logf("✅ symbol 过滤生效，返回 %d 笔 BTC/USDT 订单", len(resp.Orders))
}
