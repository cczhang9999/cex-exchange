package main

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "cex-exchange/api/proto" // 使用 cex-exchange 模块下的 proto 路径

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	grpcAddr = "localhost:9000"
	httpAddr = "http://localhost:8080"
)

func main() {
	fmt.Println("🚀 开始测试 Exchange Backend...")
	fmt.Println("==============================")

	// 1. 测试 gRPC 接口
	testGRPC()

	fmt.Println("\n------------------------------")

	// 2. 测试 HTTP 接口
	//testHTTP()
}

func testGRPC() {
	fmt.Println("\n📡 [gRPC 测试]")

	// 连接服务器
	conn, err := grpc.Dial(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("无法连接 gRPC 服务器: %v", err)
	}
	defer conn.Close()

	client := pb.NewExchangeServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	username := "ggg"

	// B. 测试登录
	fmt.Print("2. 登录 (gRPC)... ")
	loginResp, err := client.Login(ctx, &pb.LoginRequest{
		Username: username,
		Password: "1",
	})
	if err != nil {
		fmt.Printf("❌ 失败: %v\n", err)
	} else if !loginResp.Success {
		fmt.Printf("❌ 业务失败: %s\n", loginResp.Message)
	} else {
		fmt.Printf("✅ 成功! Token: %s\n", loginResp.Token)
	}

	fmt.Printf("%v\n", loginResp)
}

// func testHTTP() {
// 	fmt.Println("\n🌐 [HTTP 测试]")

// 	// A. 测试注册
// 	fmt.Print("1. 注册 (HTTP)... ")
// 	regData := map[string]string{
// 		"username": "http_user_" + fmt.Sprint(time.Now().Unix()),
// 		"password": "password123",
// 		"email":    "http@test.com",
// 		"phone":    "13900139000",
// 	}
// 	jsonData, _ := json.Marshal(regData)
// 	resp, err := http.Post(httpAddr+"/api/register", "application/json", bytes.NewBuffer(jsonData))
// 	if err != nil {
// 		fmt.Printf("❌ 失败: %v\n", err)
// 	} else {
// 		defer resp.Body.Close()
// 		body, _ := io.ReadAll(resp.Body)
// 		if resp.StatusCode == 200 {
// 			fmt.Printf("✅ 成功! 响应: %s\n", string(body))
// 		} else {
// 			fmt.Printf("❌ 失败! 状态码: %d, 响应: %s\n", resp.StatusCode, string(body))
// 		}
// 	}

// 	// B. 测试登录
// 	fmt.Print("2. 登录 (HTTP)... ")
// 	loginData := map[string]string{
// 		"username": "testuser",
// 		"password": "password123",
// 	}
// 	jsonData, _ = json.Marshal(loginData)
// 	resp, err = http.Post(httpAddr+"/api/login", "application/json", bytes.NewBuffer(jsonData))
// 	if err != nil {
// 		fmt.Printf("❌ 失败: %v\n", err)
// 	} else {
// 		defer resp.Body.Close()
// 		body, _ := io.ReadAll(resp.Body)
// 		if resp.StatusCode == 200 {
// 			fmt.Printf("✅ 成功! 响应: %s\n", string(body))
// 		} else {
// 			fmt.Printf("❌ 失败! 状态码: %d, 响应: %s\n", resp.StatusCode, string(body))
// 		}
// 	}
// }
