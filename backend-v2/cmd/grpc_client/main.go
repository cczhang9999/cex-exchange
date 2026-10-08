// Package main implements a standalone gRPC client for the ExchangeService.
//
// Demonstrates how to remotely call order-related gRPC interfaces:
//   - PlaceOrder        : 下单
//   - CancelOrder       : 取消订单
//   - GetMyOrders       : 查询用户订单
//   - GetBalance        : 查询账户余额
//
// Usage:
//
//	go run cmd/grpc_client/main.go \
//	    --server localhost:9000 \
//	    --token <jwt-token>
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	pb "backend-v2/api/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func main() {
	serverAddr := flag.String("server", "localhost:9000", "gRPC 服务器地址")
	token := flag.String("token", "", "JWT token")
	symbol := flag.String("symbol", "BTC/USDT", "交易对")
	side := flag.String("side", "buy", "买卖方向")
	orderType := flag.String("type", "limit", "订单类型")
	price := flag.String("price", "50000.00", "价格")
	amount := flag.String("amount", "0.5", "数量")
	flag.Parse()

	if *token == "" {
		fmt.Fprintln(os.Stderr, "错误: --token 参数必填")
		os.Exit(1)
	}

	// 1. 建立 gRPC 连接
	conn, err := grpc.NewClient(*serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("❌ 无法连接 gRPC 服务器 %s: %v", *serverAddr, err)
	}
	defer conn.Close()

	client := pb.NewExchangeServiceClient(conn)

	// 2. 准备带 token 的 context (通过 gRPC metadata 发送)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+*token)

	fmt.Printf("🔗 正在连接 gRPC 服务器: %s\n", *serverAddr)
	fmt.Println()

	// === 1. PlaceOrder - 下单 ===
	fmt.Println("=== PlaceOrder (下单) ===")
	placeResp, err := client.PlaceOrder(ctx, &pb.PlaceOrderRequest{
		Symbol: *symbol,
		Side:   *side,
		Type:   *orderType,
		Price:  *price,
		Amount: *amount,
		Token:  *token,
	})
	if err != nil {
		log.Fatalf("❌ PlaceOrder 失败: %v", err)
	}
	if placeResp.Success {
		fmt.Printf("✅ 下单成功: order_id=%d, message=%s\n", placeResp.OrderId, placeResp.Message)
	} else {
		fmt.Printf("⚠️ 下单失败: %s\n", placeResp.Message)
	}
	fmt.Println()

	// === 2. GetMyOrders - 查询订单 ===
	fmt.Println("=== GetMyOrders (查询订单) ===")
	ordersResp, err := client.GetMyOrders(ctx, &pb.GetMyOrdersRequest{
		Symbol: *symbol,
		Token:  *token,
	})
	if err != nil {
		log.Fatalf("❌ GetMyOrders 失败: %v", err)
	}
	if ordersResp.Success {
		fmt.Printf("✅ 查询成功，共 %d 笔订单:\n", len(ordersResp.Orders))
		for _, o := range ordersResp.Orders {
			fmt.Printf("  - ID=%d, Symbol=%s, Side=%s, Type=%s, Price=%s, Amount=%s, Filled=%s, Status=%s\n",
				o.Id, o.Symbol, o.Side, o.Type, o.Price, o.Amount, o.Filled, o.Status)
		}
	} else {
		fmt.Printf("⚠️ 查询失败: %s\n", ordersResp.Message)
	}
	fmt.Println()

	// === 3. CancelOrder - 取消订单 (如果有订单) ===
	if len(ordersResp.Orders) > 0 {
		fmt.Println("=== CancelOrder (取消订单) ===")
		orderID := ordersResp.Orders[0].Id
		cancelResp, err := client.CancelOrder(ctx, &pb.CancelOrderRequest{
			OrderId: orderID,
			Token:   *token,
		})
		if err != nil {
			log.Fatalf("❌ CancelOrder 失败: %v", err)
		}
		if cancelResp.Success {
			fmt.Printf("✅ 取消成功: order_id=%d, message=%s\n", orderID, cancelResp.Message)
		} else {
			fmt.Printf("⚠️ 取消失败: %s\n", cancelResp.Message)
		}
		fmt.Println()
	}

	// === 4. GetBalance - 查询余额 ===
	fmt.Println("=== GetBalance (查询余额) ===")
	balanceResp, err := client.GetBalance(ctx, &pb.GetBalanceRequest{
		Token: *token,
	})
	if err != nil {
		log.Fatalf("❌ GetBalance 失败: %v", err)
	}
	if balanceResp.Success {
		fmt.Printf("✅ 查询成功，共 %d 个账户:\n", len(balanceResp.Balances))
		for _, b := range balanceResp.Balances {
			fmt.Printf("  - Asset=%s, Balance=%s, Frozen=%s\n", b.Asset, b.Balance, b.Frozen)
		}
	} else {
		fmt.Printf("⚠️ 查询失败: %s\n", balanceResp.Message)
	}
}
