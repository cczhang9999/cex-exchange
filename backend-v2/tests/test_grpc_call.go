package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	pb "backend-v2/api/proto"
	"backend-v2/internal/conf"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	fmt.Println("🔗 [backend-v2] 正在尝试调用 [cex-exchange] gRPC 服务...")

	// 加载配置
	configPath := "../configs/config.yaml"
	if path := os.Getenv("CONFIG_PATH"); path != "" {
		configPath = path
	}

	bc, err := conf.Load(configPath)
	if err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}

	targetAddr := bc.Client.Exchange.Addr
	fmt.Printf("📍 目标地址: %s\n", targetAddr)

	// 1. 建立连接
	conn, err := grpc.Dial(targetAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("无法连接 gRPC 服务器: %v", err)
	}
	defer conn.Close()

	client := pb.NewExchangeServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 2. 调用测试接口 (例如 GetTicker)
	fmt.Printf("📡 调用 GetTicker (symbol: BTC/USDT)... ")
	resp, err := client.GetOrderBook(ctx, &pb.GetOrderBookRequest{
		Symbol: "BTC/USDT",
	})

	if err != nil {
		fmt.Printf("❌ 失败: %v\n", err)
		fmt.Println("\n💡 提示: 请确保 cex-exchange 的 gRPC 服务已启动。")
		fmt.Println("可以使用以下命令启动: cd cex-exchange && go run cmd/grpc/main.go")
		return
	}

	if resp.Success {
		fmt.Printf("✅ 成功!\n")
		fmt.Printf("%+v\n", resp)
	} else {
		fmt.Printf("⚠️ 业务逻辑返回失败: %s\n", resp.Message)
	}
}
