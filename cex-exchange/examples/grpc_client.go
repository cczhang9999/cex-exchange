package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "cex-exchange/api/proto"
)

func main() {
	// 连接到 gRPC 服务器
	conn, err := grpc.Dial("localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewExchangeServiceClient(conn)
	ctx := context.Background()

	fmt.Println("🚀 gRPC 客户端示例")
	fmt.Println("==================")

	// 1. 用户注册
	fmt.Println("\n1️⃣ 测试用户注册...")
	registerResp, err := client.Register(ctx, &pb.RegisterRequest{
		Username: "testuser",
		Password: "password123",
		Email:    "test@example.com",
	})
	if err != nil {
		log.Printf("注册失败: %v", err)
	} else {
		fmt.Printf("✅ 注册成功: %+v\n", registerResp)
	}

	// 2. 用户登录
	fmt.Println("\n2️⃣ 测试用户登录...")
	loginResp, err := client.Login(ctx, &pb.LoginRequest{
		Username: "testuser",
		Password: "password123",
	})
	if err != nil {
		log.Fatalf("登录失败: %v", err)
	}
	fmt.Printf("✅ 登录成功: token=%s, user_id=%d\n", loginResp.Token, loginResp.UserId)

	token := loginResp.Token

	// 3. 获取账户余额
	fmt.Println("\n3️⃣ 测试获取账户余额...")
	balanceResp, err := client.GetBalance(ctx, &pb.GetBalanceRequest{
		Token: token,
	})
	if err != nil {
		log.Printf("获取余额失败: %v", err)
	} else {
		fmt.Printf("✅ 账户余额:\n")
		for _, balance := range balanceResp.Balances {
			fmt.Printf("   %s: 可用=%s, 冻结=%s\n", balance.Asset, balance.Balance, balance.Frozen)
		}
	}

	// 4. 充值
	fmt.Println("\n4️⃣ 测试充值...")
	depositResp, err := client.Deposit(ctx, &pb.DepositRequest{
		Asset:  "USDT",
		Amount: "10000",
		Token:  token,
	})
	if err != nil {
		log.Printf("充值失败: %v", err)
	} else {
		fmt.Printf("✅ 充值成功: %+v\n", depositResp)
	}

	// 5. 下单
	fmt.Println("\n5️⃣ 测试下单...")
	placeOrderResp, err := client.PlaceOrder(ctx, &pb.PlaceOrderRequest{
		Symbol: "BTC/USDT",
		Side:   "buy",
		Type:   "limit",
		Price:  "50000",
		Amount: "0.1",
		Token:  token,
	})
	if err != nil {
		log.Printf("下单失败: %v", err)
	} else {
		fmt.Printf("✅ 下单成功: order_id=%d\n", placeOrderResp.OrderId)
	}

	// 6. 获取我的订单
	fmt.Println("\n6️⃣ 测试获取我的订单...")
	myOrdersResp, err := client.GetMyOrders(ctx, &pb.GetMyOrdersRequest{
		Symbol: "BTC/USDT",
		Token:  token,
	})
	if err != nil {
		log.Printf("获取订单失败: %v", err)
	} else {
		fmt.Printf("✅ 我的订单 (%d 条):\n", len(myOrdersResp.Orders))
		for _, order := range myOrdersResp.Orders {
			fmt.Printf("   订单 #%d: %s %s %s @ %s, 数量=%s, 状态=%s\n",
				order.Id, order.Symbol, order.Side, order.Type, order.Price, order.Amount, order.Status)
		}
	}

	// 7. 获取订单簿
	fmt.Println("\n7️⃣ 测试获取订单簿...")
	orderBookResp, err := client.GetOrderBook(ctx, &pb.GetOrderBookRequest{
		Symbol: "BTC/USDT",
		Depth:  10,
	})
	if err != nil {
		log.Printf("获取订单簿失败: %v", err)
	} else {
		fmt.Printf("✅ 订单簿:\n")
		fmt.Printf("   买单 (%d 档):\n", len(orderBookResp.Bids))
		for i, bid := range orderBookResp.Bids {
			if i < 5 {
				fmt.Printf("      %s @ %s\n", bid.Amount, bid.Price)
			}
		}
		fmt.Printf("   卖单 (%d 档):\n", len(orderBookResp.Asks))
		for i, ask := range orderBookResp.Asks {
			if i < 5 {
				fmt.Printf("      %s @ %s\n", ask.Amount, ask.Price)
			}
		}
	}

	// 8. 获取行情
	fmt.Println("\n8️⃣ 测试获取行情...")
	tickerResp, err := client.GetTicker(ctx, &pb.GetTickerRequest{
		Symbol: "BTC/USDT",
	})
	if err != nil {
		log.Printf("获取行情失败: %v", err)
	} else {
		ticker := tickerResp.Ticker
		fmt.Printf("✅ 行情: 最新价=%s, 24h高=%s, 24h低=%s, 24h量=%s, 24h涨跌=%s\n",
			ticker.LastPrice, ticker.High_24H, ticker.Low_24H, ticker.Volume_24H, ticker.Change_24H)
	}

	// 9. 获取最近成交
	fmt.Println("\n9️⃣ 测试获取最近成交...")
	tradesResp, err := client.GetRecentTrades(ctx, &pb.GetRecentTradesRequest{
		Symbol: "BTC/USDT",
		Limit:  10,
	})
	if err != nil {
		log.Printf("获取成交失败: %v", err)
	} else {
		fmt.Printf("✅ 最近成交 (%d 条):\n", len(tradesResp.Trades))
		for i, trade := range tradesResp.Trades {
			if i < 5 {
				fmt.Printf("   #%d: %s %s @ %s, 时间=%s\n",
					trade.Id, trade.Side, trade.Amount, trade.Price,
					time.Unix(trade.Timestamp, 0).Format("15:04:05"))
			}
		}
	}

	// 10. 订阅订单簿更新（流式）
	fmt.Println("\n🔟 测试订阅订单簿更新（5秒）...")
	streamCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	stream, err := client.SubscribeOrderBook(streamCtx, &pb.SubscribeOrderBookRequest{
		Symbol: "BTC/USDT",
	})
	if err != nil {
		log.Printf("订阅订单簿失败: %v", err)
	} else {
		count := 0
		for {
			update, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Printf("接收更新失败: %v", err)
				break
			}
			count++
			fmt.Printf("   📊 订单簿更新 #%d: 买单=%d档, 卖单=%d档, 时间=%s\n",
				count, len(update.Bids), len(update.Asks),
				time.Unix(update.Timestamp, 0).Format("15:04:05"))
		}
	}

	// 11. 订阅成交更新（流式）
	fmt.Println("\n1️⃣1️⃣ 测试订阅成交更新（5秒）...")
	streamCtx2, cancel2 := context.WithTimeout(ctx, 5*time.Second)
	defer cancel2()

	tradeStream, err := client.SubscribeTrades(streamCtx2, &pb.SubscribeTradesRequest{
		Symbol: "BTC/USDT",
	})
	if err != nil {
		log.Printf("订阅成交失败: %v", err)
	} else {
		count := 0
		for {
			update, err := tradeStream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Printf("接收更新失败: %v", err)
				break
			}
			count++
			fmt.Printf("   💰 成交更新 #%d: %s %s @ %s, 时间=%s\n",
				count, update.Side, update.Amount, update.Price,
				time.Unix(update.Timestamp, 0).Format("15:04:05"))
		}
	}

	fmt.Println("\n✅ 所有测试完成！")
}
