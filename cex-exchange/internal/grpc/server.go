package grpc

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "cex-exchange/api/proto"
	"cex-exchange/internal/model"
	"cex-exchange/internal/service"
)

// ExchangeServer 实现 gRPC 服务
type ExchangeServer struct {
	pb.UnimplementedExchangeServiceServer
}

// NewExchangeServer 创建新的 gRPC 服务器实例
func NewExchangeServer() *ExchangeServer {
	return &ExchangeServer{}
}

// ==================== 认证相关 ====================

// Login 用户登录
func (s *ExchangeServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	g.Log().Infof(ctx, "gRPC Login: username=%s", req.Username)

	// 调用现有的认证服务
	token, uid, err := service.Auth.Login(ctx, req.Username, req.Password)
	if err != nil {
		return &pb.LoginResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.LoginResponse{
		Success: true,
		Message: "登录成功",
		Token:   token,
		UserId:  uid,
	}, nil
}

// Register 用户注册
func (s *ExchangeServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	g.Log().Infof(ctx, "gRPC Register: username=%s, email=%s", req.Username, req.Email)

	// 调用现有的认证服务
	uid, err := service.Auth.Register(ctx, req.Username, req.Password, req.Email)
	if err != nil {
		return &pb.RegisterResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.RegisterResponse{
		Success: true,
		Message: "注册成功",
		UserId:  uid,
	}, nil
}

// ==================== 订单相关 ====================

// PlaceOrder 下单
func (s *ExchangeServer) PlaceOrder(ctx context.Context, req *pb.PlaceOrderRequest) (*pb.PlaceOrderResponse, error) {
	g.Log().Infof(ctx, "gRPC PlaceOrder: symbol=%s, side=%s, price=%s, amount=%s",
		req.Symbol, req.Side, req.Price, req.Amount)

	// 验证 token 并获取用户 ID
	uid, err := s.validateToken(ctx, req.Token)
	if err != nil {
		return &pb.PlaceOrderResponse{
			Success: false,
			Message: "认证失败: " + err.Error(),
		}, nil
	}

	// 设置用户 ID 到上下文
	ctx = context.WithValue(ctx, "uid", uid)
	g.RequestFromCtx(ctx).SetCtxVar("uid", uid)

	// 调用订单服务
	orderReq := &model.PlaceOrderReq{
		Symbol: req.Symbol,
		Side:   req.Side,
		Type:   req.Type,
		Price:  req.Price,
		Amount: req.Amount,
	}

	orderID, err := service.Order.PlaceOrder(ctx, orderReq)
	if err != nil {
		return &pb.PlaceOrderResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.PlaceOrderResponse{
		Success: true,
		Message: "下单成功",
		OrderId: orderID,
	}, nil
}

// CancelOrder 撤单
func (s *ExchangeServer) CancelOrder(ctx context.Context, req *pb.CancelOrderRequest) (*pb.CancelOrderResponse, error) {
	g.Log().Infof(ctx, "gRPC CancelOrder: order_id=%d", req.OrderId)

	// 验证 token
	uid, err := s.validateToken(ctx, req.Token)
	if err != nil {
		return &pb.CancelOrderResponse{
			Success: false,
			Message: "认证失败: " + err.Error(),
		}, nil
	}

	// 调用订单服务
	err = service.Order.CancelOrder(ctx, uid, req.OrderId)
	if err != nil {
		return &pb.CancelOrderResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.CancelOrderResponse{
		Success: true,
		Message: "撤单成功",
	}, nil
}

// GetMyOrders 获取我的订单
func (s *ExchangeServer) GetMyOrders(ctx context.Context, req *pb.GetMyOrdersRequest) (*pb.GetMyOrdersResponse, error) {
	g.Log().Infof(ctx, "gRPC GetMyOrders: symbol=%s", req.Symbol)

	// 验证 token
	uid, err := s.validateToken(ctx, req.Token)
	if err != nil {
		return &pb.GetMyOrdersResponse{
			Success: false,
			Message: "认证失败: " + err.Error(),
		}, nil
	}

	// 调用订单服务
	orders, err := service.Order.ListMyOrders(ctx, uid, req.Symbol)
	if err != nil {
		return &pb.GetMyOrdersResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// 转换为 protobuf 格式
	pbOrders := make([]*pb.Order, len(orders))
	for i, order := range orders {
		pbOrders[i] = &pb.Order{
			Id:        order.ID,
			UserId:    order.UserID,
			Symbol:    order.Symbol,
			Side:      order.Side,
			Type:      order.Type,
			Price:     order.Price,
			Amount:    order.Amount,
			Filled:    order.Filled,
			Status:    order.Status,
			CreatedAt: order.CreatedAt.Unix(),
			UpdatedAt: order.UpdatedAt.Unix(),
		}
	}

	return &pb.GetMyOrdersResponse{
		Success: true,
		Message: "查询成功",
		Orders:  pbOrders,
	}, nil
}

// GetOrderBook 获取订单簿
func (s *ExchangeServer) GetOrderBook(ctx context.Context, req *pb.GetOrderBookRequest) (*pb.GetOrderBookResponse, error) {
	g.Log().Infof(ctx, "gRPC GetOrderBook: symbol=%s, depth=%d", req.Symbol, req.Depth)

	// 调用市场数据服务
	orderBook, err := service.Market.GetOrderBook(ctx, req.Symbol, int(req.Depth))
	if err != nil {
		return &pb.GetOrderBookResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// 转换为 protobuf 格式
	bids := make([]*pb.PriceLevel, len(orderBook.Bids))
	for i, bid := range orderBook.Bids {
		bids[i] = &pb.PriceLevel{
			Price:  bid.Price,
			Amount: bid.Amount,
		}
	}

	asks := make([]*pb.PriceLevel, len(orderBook.Asks))
	for i, ask := range orderBook.Asks {
		asks[i] = &pb.PriceLevel{
			Price:  ask.Price,
			Amount: ask.Amount,
		}
	}

	return &pb.GetOrderBookResponse{
		Success: true,
		Message: "查询成功",
		Bids:    bids,
		Asks:    asks,
	}, nil
}

// ==================== 账户相关 ====================

// GetBalance 获取账户余额
func (s *ExchangeServer) GetBalance(ctx context.Context, req *pb.GetBalanceRequest) (*pb.GetBalanceResponse, error) {
	g.Log().Info(ctx, "gRPC GetBalance")

	// 验证 token
	uid, err := s.validateToken(ctx, req.Token)
	if err != nil {
		return &pb.GetBalanceResponse{
			Success: false,
			Message: "认证失败: " + err.Error(),
		}, nil
	}

	// 调用账户服务
	balances, err := service.Funds.GetBalance(ctx, uid)
	if err != nil {
		return &pb.GetBalanceResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// 转换为 protobuf 格式
	pbBalances := make([]*pb.Balance, len(balances))
	for i, balance := range balances {
		pbBalances[i] = &pb.Balance{
			Asset:   balance.Asset,
			Balance: balance.Balance,
			Frozen:  balance.Frozen,
		}
	}

	return &pb.GetBalanceResponse{
		Success:  true,
		Message:  "查询成功",
		Balances: pbBalances,
	}, nil
}

// Deposit 充值
func (s *ExchangeServer) Deposit(ctx context.Context, req *pb.DepositRequest) (*pb.DepositResponse, error) {
	g.Log().Infof(ctx, "gRPC Deposit: asset=%s, amount=%s", req.Asset, req.Amount)

	// 验证 token
	uid, err := s.validateToken(ctx, req.Token)
	if err != nil {
		return &pb.DepositResponse{
			Success: false,
			Message: "认证失败: " + err.Error(),
		}, nil
	}

	// 调用账户服务
	err = service.Funds.Deposit(ctx, uid, req.Asset, req.Amount)
	if err != nil {
		return &pb.DepositResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.DepositResponse{
		Success: true,
		Message: "充值成功",
	}, nil
}

// Withdraw 提现
func (s *ExchangeServer) Withdraw(ctx context.Context, req *pb.WithdrawRequest) (*pb.WithdrawResponse, error) {
	g.Log().Infof(ctx, "gRPC Withdraw: asset=%s, amount=%s, address=%s",
		req.Asset, req.Amount, req.Address)

	// 验证 token
	uid, err := s.validateToken(ctx, req.Token)
	if err != nil {
		return &pb.WithdrawResponse{
			Success: false,
			Message: "认证失败: " + err.Error(),
		}, nil
	}

	// 调用账户服务
	err = service.Funds.Withdraw(ctx, uid, req.Asset, req.Amount, req.Address)
	if err != nil {
		return &pb.WithdrawResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.WithdrawResponse{
		Success: true,
		Message: "提现成功",
	}, nil
}

// ==================== 市场数据相关 ====================

// GetTicker 获取行情
func (s *ExchangeServer) GetTicker(ctx context.Context, req *pb.GetTickerRequest) (*pb.GetTickerResponse, error) {
	g.Log().Infof(ctx, "gRPC GetTicker: symbol=%s", req.Symbol)

	// 调用市场数据服务
	ticker, err := service.Market.GetTicker(ctx, req.Symbol)
	if err != nil {
		return &pb.GetTickerResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.GetTickerResponse{
		Success: true,
		Message: "查询成功",
		Ticker: &pb.Ticker{
			Symbol:     ticker.Symbol,
			LastPrice:  ticker.LastPrice,
			High24H:    ticker.High24h,
			Low24H:     ticker.Low24h,
			Volume24H:  ticker.Volume24h,
			Change24H:  ticker.Change24h,
		},
	}, nil
}

// GetKlines 获取 K 线数据
func (s *ExchangeServer) GetKlines(ctx context.Context, req *pb.GetKlinesRequest) (*pb.GetKlinesResponse, error) {
	g.Log().Infof(ctx, "gRPC GetKlines: symbol=%s, interval=%s", req.Symbol, req.Interval)

	// 调用市场数据服务
	klines, err := service.Market.GetKlines(ctx, req.Symbol, req.Interval, req.StartTime, req.EndTime, int(req.Limit))
	if err != nil {
		return &pb.GetKlinesResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// 转换为 protobuf 格式
	pbKlines := make([]*pb.Kline, len(klines))
	for i, kline := range klines {
		pbKlines[i] = &pb.Kline{
			Timestamp: kline.Timestamp,
			Open:      kline.Open,
			High:      kline.High,
			Low:       kline.Low,
			Close:     kline.Close,
			Volume:    kline.Volume,
		}
	}

	return &pb.GetKlinesResponse{
		Success: true,
		Message: "查询成功",
		Klines:  pbKlines,
	}, nil
}

// GetRecentTrades 获取最近成交
func (s *ExchangeServer) GetRecentTrades(ctx context.Context, req *pb.GetRecentTradesRequest) (*pb.GetRecentTradesResponse, error) {
	g.Log().Infof(ctx, "gRPC GetRecentTrades: symbol=%s, limit=%d", req.Symbol, req.Limit)

	// 调用市场数据服务
	trades, err := service.Market.GetRecentTrades(ctx, req.Symbol, int(req.Limit))
	if err != nil {
		return &pb.GetRecentTradesResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// 转换为 protobuf 格式
	pbTrades := make([]*pb.Trade, len(trades))
	for i, trade := range trades {
		pbTrades[i] = &pb.Trade{
			Id:        trade.ID,
			Symbol:    trade.Symbol,
			Price:     trade.Price,
			Amount:    trade.Amount,
			Side:      trade.Side,
			Timestamp: trade.Timestamp,
		}
	}

	return &pb.GetRecentTradesResponse{
		Success: true,
		Message: "查询成功",
		Trades:  pbTrades,
	}, nil
}

// ==================== 实时数据流 ====================

// SubscribeOrderBook 订阅订单簿更新（服务端流式）
func (s *ExchangeServer) SubscribeOrderBook(req *pb.SubscribeOrderBookRequest, stream pb.ExchangeService_SubscribeOrderBookServer) error {
	ctx := stream.Context()
	g.Log().Infof(ctx, "gRPC SubscribeOrderBook: symbol=%s", req.Symbol)

	// 创建订阅通道
	updateChan := make(chan *pb.OrderBookUpdate, 100)

	// 启动订单簿更新推送
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// 获取最新订单簿
				orderBook, err := service.Market.GetOrderBook(ctx, req.Symbol, 20)
				if err != nil {
					g.Log().Errorf(ctx, "获取订单簿失败: %v", err)
					continue
				}

				// 转换为 protobuf 格式
				bids := make([]*pb.PriceLevel, len(orderBook.Bids))
				for i, bid := range orderBook.Bids {
					bids[i] = &pb.PriceLevel{
						Price:  bid.Price,
						Amount: bid.Amount,
					}
				}

				asks := make([]*pb.PriceLevel, len(orderBook.Asks))
				for i, ask := range orderBook.Asks {
					asks[i] = &pb.PriceLevel{
						Price:  ask.Price,
						Amount: ask.Amount,
					}
				}

				update := &pb.OrderBookUpdate{
					Symbol:    req.Symbol,
					Bids:      bids,
					Asks:      asks,
					Timestamp: time.Now().Unix(),
				}

				select {
				case updateChan <- update:
				default:
					// 通道满了，跳过这次更新
				}
			}
		}
	}()

	// 发送更新到客户端
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case update := <-updateChan:
			if err := stream.Send(update); err != nil {
				return err
			}
		}
	}
}

// SubscribeTrades 订阅成交更新（服务端流式）
func (s *ExchangeServer) SubscribeTrades(req *pb.SubscribeTradesRequest, stream pb.ExchangeService_SubscribeTradesServer) error {
	ctx := stream.Context()
	g.Log().Infof(ctx, "gRPC SubscribeTrades: symbol=%s", req.Symbol)

	// 创建订阅通道
	updateChan := make(chan *pb.TradeUpdate, 100)

	// 启动成交更新推送
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// 获取最新成交
				trades, err := service.Market.GetRecentTrades(ctx, req.Symbol, 1)
				if err != nil {
					continue
				}

				if len(trades) > 0 {
					trade := trades[0]
					update := &pb.TradeUpdate{
						Symbol:    trade.Symbol,
						Price:     trade.Price,
						Amount:    trade.Amount,
						Side:      trade.Side,
						Timestamp: trade.Timestamp,
					}

					select {
					case updateChan <- update:
					default:
						// 通道满了，跳过这次更新
					}
				}
			}
		}
	}()

	// 发送更新到客户端
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case update := <-updateChan:
			if err := stream.Send(update); err != nil {
				return err
			}
		}
	}
}

// ==================== 辅助方法 ====================

// validateToken 验证 token 并返回用户 ID
func (s *ExchangeServer) validateToken(ctx context.Context, token string) (uint64, error) {
	if token == "" {
		return 0, status.Error(codes.Unauthenticated, "token 不能为空")
	}

	// 调用认证服务验证 token
	uid, err := service.Auth.ValidateToken(ctx, token)
	if err != nil {
		return 0, status.Error(codes.Unauthenticated, "token 无效或已过期")
	}

	return uid, nil
}
