package service

import (
	pb "backend-v2/api/proto"
	"backend-v2/internal/biz"
	"backend-v2/internal/conf"
	"backend-v2/internal/pkg/jwt"
	"backend-v2/internal/server/middleware"
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(NewExchangeService)

type ExchangeService struct {
	pb.UnimplementedExchangeServiceServer
	user        *biz.UserUsecase
	order       *biz.OrderUsecase
	account     *biz.AccountUsecase
	accountFlow *biz.AccountFlowUsecase
	kline       *biz.KlineUsecase
	trade       *biz.TradeUsecase
	client      pb.ExchangeServiceClient
	jwtSecret   string
}

func NewExchangeService(user *biz.UserUsecase, order *biz.OrderUsecase, account *biz.AccountUsecase, accountFlow *biz.AccountFlowUsecase, kline *biz.KlineUsecase, client pb.ExchangeServiceClient, conf *conf.Bootstrap, trade *biz.TradeUsecase) *ExchangeService {
	jwtSecret := ""
	if conf != nil && conf.Auth != nil {
		jwtSecret = conf.Auth.JwtSecret
	}
	return &ExchangeService{
		user:        user,
		order:       order,
		account:     account,
		accountFlow: accountFlow,
		kline:       kline,
		trade:       trade,
		client:      client,
		jwtSecret:   jwtSecret,
	}
}

// extractUserIDFromToken 解析 token，返回 user_id；token 为空或解析失败时返回 0 和错误
func (s *ExchangeService) extractUserIDFromToken(token string) (uint64, error) {
	if token == "" {
		return 0, fmt.Errorf("token is required")
	}
	claims, err := jwt.ParseToken(token, s.jwtSecret)
	if err != nil {
		return 0, fmt.Errorf("invalid token: %w", err)
	}
	return claims.UserID, nil
}

// requireUserID 获取当前请求的认证用户 ID。
// 优先从 context 中读取（由 gRPC auth 拦截器设置），其次从请求中的 token 字段解析。
func (s *ExchangeService) requireUserID(ctx context.Context, token string) (uint64, error) {
	// 1. 优先从 context 获取（gRPC 拦截器已完成认证）
	if uid, ok := middleware.FromContext(ctx); ok {
		return uid, nil
	}
	// 2. 回退到解析请求中的 token 字段
	return s.extractUserIDFromToken(token)
}

func (s *ExchangeService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	token, id, err := s.user.Login(ctx, req.Username, req.Password)
	if err != nil {
		return &pb.LoginResponse{Success: false, Message: err.Error()}, nil
	}
	return &pb.LoginResponse{
		Success: true,
		Message: "Login successful",
		Token:   token,
		UserId:  id,
	}, nil
}

func (s *ExchangeService) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	u, err := s.user.Register(ctx, req.Username, req.Password, req.Email, req.Phone)
	if err != nil {
		return &pb.RegisterResponse{Success: false, Message: err.Error()}, nil
	}
	return &pb.RegisterResponse{
		Success: true,
		Message: "Register successful",
		UserId:  u.ID,
	}, nil
}

func (s *ExchangeService) GetOrderBook(ctx context.Context, req *pb.GetOrderBookRequest) (*pb.GetOrderBookResponse, error) {
	if s.client == nil {
		// 后端 cex-exchange 服务未配置时，返回空盘
		return &pb.GetOrderBookResponse{
			Success: true,
			Message: "Success",
			Bids:    []*pb.PriceLevel{},
			Asks:    []*pb.PriceLevel{},
		}, nil
	}
	return s.client.GetOrderBook(ctx, req)
}

func (s *ExchangeService) GetRecentTrades(ctx context.Context, req *pb.GetRecentTradesRequest) (*pb.GetRecentTradesResponse, error) {
	if s.client == nil {
		return &pb.GetRecentTradesResponse{
			Success: true,
			Message: "Success",
			Trades:  []*pb.Trade{},
		}, nil
	}
	return s.client.GetRecentTrades(ctx, req)
}

func (s *ExchangeService) GetKlines(ctx context.Context, req *pb.GetKlinesRequest) (*pb.GetKlinesResponse, error) {
	klines, err := s.kline.GetKlines(ctx, req.Symbol, req.Interval, req.StartTime, req.EndTime, req.Limit)
	if err != nil {
		return &pb.GetKlinesResponse{Success: false, Message: err.Error()}, nil
	}

	pbKlines := make([]*pb.Kline, 0, len(klines))
	for _, k := range klines {
		pbKlines = append(pbKlines, &pb.Kline{
			Timestamp: k.OpenTime.Unix(),
			Open:      fmt.Sprintf("%.8f", k.Open),
			High:      fmt.Sprintf("%.8f", k.High),
			Low:       fmt.Sprintf("%.8f", k.Low),
			Close:     fmt.Sprintf("%.8f", k.Close),
			Volume:    fmt.Sprintf("%.8f", k.Volume),
		})
	}

	return &pb.GetKlinesResponse{
		Success: true,
		Message: "Success",
		Klines:  pbKlines,
	}, nil
}

// PlaceOrder 下单 (gRPC 接口)
// 从 token 中提取用户 ID，调用 biz 层创建订单，返回订单 ID
func (s *ExchangeService) PlaceOrder(ctx context.Context, req *pb.PlaceOrderRequest) (*pb.PlaceOrderResponse, error) {
	// 1. 身份认证：优先从 context 获取，回退到请求 token 字段
	userID, err := s.requireUserID(ctx, req.Token)
	if err != nil {
		return &pb.PlaceOrderResponse{Success: false, Message: "认证失败: " + err.Error()}, nil
	}

	// 2. 校验参数
	if req.Symbol == "" {
		return &pb.PlaceOrderResponse{Success: false, Message: "symbol is required"}, nil
	}
	if req.Side == "" {
		return &pb.PlaceOrderResponse{Success: false, Message: "side is required"}, nil
	}
	if req.Type == "" {
		return &pb.PlaceOrderResponse{Success: false, Message: "type is required"}, nil
	}

	// 3. 解析字符串类型的价格、数量 (避免浮点精度问题)
	price, err := strconv.ParseFloat(req.Price, 64)
	if err != nil {
		return &pb.PlaceOrderResponse{Success: false, Message: "invalid price: " + err.Error()}, nil
	}
	amount, err := strconv.ParseFloat(req.Amount, 64)
	if err != nil {
		return &pb.PlaceOrderResponse{Success: false, Message: "invalid amount: " + err.Error()}, nil
	}

	// 4. 转换为 biz 类型并创建订单
	side := biz.OrderSide(req.Side)
	orderType := biz.OrderType(req.Type)
	order, err := s.order.CreateOrder(ctx, userID, req.Symbol, side, orderType, price, amount)
	if err != nil {
		return &pb.PlaceOrderResponse{Success: false, Message: "下单失败: " + err.Error()}, nil
	}

	log.Printf("PlaceOrder: user_id=%d, symbol=%s, side=%s, type=%s, price=%s, amount=%s -> order_id=%d",
		userID, req.Symbol, req.Side, req.Type, req.Price, req.Amount, order.ID)

	return &pb.PlaceOrderResponse{
		Success: true,
		Message: "Order placed successfully",
		OrderId: int64(order.ID),
	}, nil
}

// CancelOrder 取消订单 (gRPC 接口)
// 从 token 中提取用户 ID，验证订单归属后取消
func (s *ExchangeService) CancelOrder(ctx context.Context, req *pb.CancelOrderRequest) (*pb.CancelOrderResponse, error) {
	// 1. 身份认证
	userID, err := s.requireUserID(ctx, req.Token)
	if err != nil {
		return &pb.CancelOrderResponse{Success: false, Message: "认证失败: " + err.Error()}, nil
	}

	// 2. 查询订单，验证归属
	order, err := s.order.GetOrder(ctx, req.OrderId)
	if err != nil || order == nil {
		return &pb.CancelOrderResponse{Success: false, Message: "订单不存在"}, nil
	}
	if order.UserID != userID {
		return &pb.CancelOrderResponse{Success: false, Message: "无权取消此订单"}, nil
	}

	// 3. 取消订单
	if err := s.order.CancelOrder(ctx, req.OrderId); err != nil {
		return &pb.CancelOrderResponse{Success: false, Message: "取消失败: " + err.Error()}, nil
	}

	log.Printf("CancelOrder: user_id=%d, order_id=%d", userID, req.OrderId)
	return &pb.CancelOrderResponse{
		Success: true,
		Message: "Order cancelled successfully",
	}, nil
}

// GetMyOrders 查询用户订单 (gRPC 接口)
// 从 token 中提取用户 ID，根据 symbol 可选过滤返回用户订单
func (s *ExchangeService) GetMyOrders(ctx context.Context, req *pb.GetMyOrdersRequest) (*pb.GetMyOrdersResponse, error) {
	// 1. 身份认证
	userID, err := s.requireUserID(ctx, req.Token)
	if err != nil {
		return &pb.GetMyOrdersResponse{Success: false, Message: "认证失败: " + err.Error()}, nil
	}

	// 2. 查询用户订单
	orders, err := s.order.GetUserOrders(ctx, userID)
	if err != nil {
		return &pb.GetMyOrdersResponse{Success: false, Message: "查询失败: " + err.Error()}, nil
	}

	// 3. 过滤 symbol（如果指定）
	pbOrders := make([]*pb.Order, 0, len(orders))
	for _, o := range orders {
		if req.Symbol != "" && o.Symbol != req.Symbol {
			continue
		}
		pbOrders = append(pbOrders, &pb.Order{
			Id:        o.ID,
			UserId:    o.UserID,
			Symbol:    o.Symbol,
			Side:      string(o.Side),
			Type:      string(o.Type),
			Price:     fmt.Sprintf("%.8f", o.Price),
			Amount:    fmt.Sprintf("%.8f", o.Amount),
			Filled:    fmt.Sprintf("%.8f", o.Filled),
			Status:    string(o.Status),
			CreatedAt: o.CreatedAt.Unix(),
			UpdatedAt: o.UpdatedAt.Unix(),
		})
	}

	log.Printf("GetMyOrders: user_id=%d, symbol=%s, count=%d", userID, req.Symbol, len(pbOrders))
	return &pb.GetMyOrdersResponse{
		Success: true,
		Message: "Success",
		Orders:  pbOrders,
	}, nil
}

// GetUserOrders 查询用户订单（HTTP 兼容方法，token 为空时使用默认用户 ID）
// 保留此方法以兼容 HTTP 路由中的调用。gRPC 接口请使用 GetMyOrders。
func (s *ExchangeService) GetUserOrders(ctx context.Context, req *pb.GetMyOrdersRequest) (*pb.GetMyOrdersResponse, error) {
	// 如果 token 为空，回退到默认用户 ID 1（向后兼容 HTTP 端）
	userID := uint64(1)
	if req.Token != "" {
		if uid, err := s.extractUserIDFromToken(req.Token); err == nil {
			userID = uid
		}
	}

	orders, err := s.order.GetUserOrders(ctx, userID)
	if err != nil {
		return &pb.GetMyOrdersResponse{Success: false, Message: err.Error()}, nil
	}

	pbOrders := make([]*pb.Order, 0, len(orders))
	for _, o := range orders {
		if req.Symbol != "" && o.Symbol != req.Symbol {
			continue
		}
		pbOrders = append(pbOrders, &pb.Order{
			Id:        o.ID,
			UserId:    o.UserID,
			Symbol:    o.Symbol,
			Side:      string(o.Side),
			Type:      string(o.Type),
			Price:     fmt.Sprintf("%.8f", o.Price),
			Amount:    fmt.Sprintf("%.8f", o.Amount),
			Filled:    fmt.Sprintf("%.8f", o.Filled),
			Status:    string(o.Status),
			CreatedAt: o.CreatedAt.Unix(),
			UpdatedAt: o.UpdatedAt.Unix(),
		})
	}

	log.Print("GetUserOrders: ", pbOrders)
	return &pb.GetMyOrdersResponse{
		Success: true,
		Message: "Success",
		Orders:  pbOrders,
	}, nil
}

// GetBalance 查询用户资产余额 (gRPC 接口)
// 从 token 中提取用户 ID，返回用户所有账户余额
func (s *ExchangeService) GetBalance(ctx context.Context, req *pb.GetBalanceRequest) (*pb.GetBalanceResponse, error) {
	// 1. 身份认证
	userID, err := s.requireUserID(ctx, req.Token)
	if err != nil {
		return &pb.GetBalanceResponse{Success: false, Message: "认证失败: " + err.Error()}, nil
	}

	// 2. 查询用户账户
	accounts, err := s.account.GetAccounts(ctx, userID)
	if err != nil {
		return &pb.GetBalanceResponse{Success: false, Message: "查询失败: " + err.Error()}, nil
	}

	// 3. 转换为 pb 类型
	pbBalances := make([]*pb.Balance, 0, len(accounts))
	for _, a := range accounts {
		pbBalances = append(pbBalances, &pb.Balance{
			Asset:   a.Asset,
			Balance: fmt.Sprintf("%.8f", a.Balance),
			Frozen:  fmt.Sprintf("%.8f", a.Frozen),
		})
	}

	log.Printf("GetBalance: user_id=%d, count=%d", userID, len(pbBalances))
	return &pb.GetBalanceResponse{
		Success:  true,
		Message:  "Success",
		Balances: pbBalances,
	}, nil
}

// GetUserAccounts 查询用户账户列表，asset 为空时查询全部币种
func (s *ExchangeService) GetUserAccounts(ctx context.Context, userID uint64, asset string) ([]*biz.Account, error) {
	if asset != "" {
		acc, err := s.account.GetAccount(ctx, userID, asset)
		if err != nil {
			return nil, err
		}
		return []*biz.Account{acc}, nil
	}
	return s.account.GetAccounts(ctx, userID)
}

// GetUserAccountFlows 分页查询用户资金流水
func (s *ExchangeService) GetUserAccountFlows(ctx context.Context, q biz.AccountFlowQuery) (*biz.AccountFlowPage, error) {
	return s.accountFlow.GetAccountFlows(ctx, q)
}

// GetUserOrderPage 分页查询用户订单
func (s *ExchangeService) GetUserOrderPage(ctx context.Context, q biz.OrderQuery) (*biz.OrderPage, error) {
	return s.order.GetOrdersPage(ctx, q)
}

// GetMyTrades 查询用户成交记录（HTTP 兼容方法）
func (s *ExchangeService) GetMyTrades(ctx context.Context, userID uint64) ([]*biz.Trade, error) {
	if s.trade == nil {
		return nil, fmt.Errorf("trade service unavailable")
	}
	return s.trade.GetMyTrades(ctx, userID)
}

// ==================== 管理员接口（HTTP 兼容方法） ====================

// ListUsers 获取用户列表（不返回密码）
func (s *ExchangeService) ListUsers(ctx context.Context) ([]*biz.User, error) {
	if s.user == nil {
		return nil, fmt.Errorf("user service unavailable")
	}
	users, err := s.user.ListUsers(ctx)
	if err != nil {
		return nil, err
	}
	for i := range users {
		users[i].Password = ""
	}
	return users, nil
}

// ListAllAccounts 获取全部账户（管理员）
func (s *ExchangeService) ListAllAccounts(ctx context.Context) ([]*biz.Account, error) {
	if s.account == nil {
		return nil, fmt.Errorf("account service unavailable")
	}
	return s.account.GetAllAccounts(ctx)
}

// AdminAddFunds 管理员为用户添加资金
func (s *ExchangeService) AdminAddFunds(ctx context.Context, userID uint64, asset, amount string) error {
	if s.account == nil {
		return fmt.Errorf("account service unavailable")
	}
	return s.account.AdjustBalance(ctx, userID, asset, amount)
}

// AdminAdjustBalance 管理员调账（同 AddFunds）
func (s *ExchangeService) AdminAdjustBalance(ctx context.Context, userID uint64, asset, amount string) error {
	return s.AdminAddFunds(ctx, userID, asset, amount)
}

// AdminBlockUser 管理员封禁/解封用户
func (s *ExchangeService) AdminBlockUser(ctx context.Context, userID uint64, block bool) error {
	if s.user == nil {
		return fmt.Errorf("user service unavailable")
	}
	return s.user.BlockUser(ctx, userID, block)
}

// ==================== 以下为补全的 gRPC 接口 ====================

// GetTicker 获取行情 (gRPC 接口)
// 优先委托给后端 cex-exchange gRPC 服务获取行情，失败时返回本地模拟数据。
func (s *ExchangeService) GetTicker(ctx context.Context, req *pb.GetTickerRequest) (*pb.GetTickerResponse, error) {
	if s.client != nil {
		resp, err := s.client.GetTicker(ctx, req)
		if err == nil {
			return resp, nil
		}
		log.Printf("GetTicker: delegate to client failed, using local fallback: %v", err)
	}
	// 本地模拟行情
	ticker := &pb.Ticker{
		Symbol:     req.Symbol,
		LastPrice:  "50000.00",
		High_24H:   "52000.00",
		Low_24H:    "48000.00",
		Volume_24H: "1234.56",
		Change_24H: "+2.5%",
	}
	return &pb.GetTickerResponse{
		Success: true,
		Message: "Success",
		Ticker:  ticker,
	}, nil
}

// Deposit 充值 (gRPC 接口)
func (s *ExchangeService) Deposit(ctx context.Context, req *pb.DepositRequest) (*pb.DepositResponse, error) {
	userID, err := s.requireUserID(ctx, req.Token)
	if err != nil {
		return &pb.DepositResponse{Success: false, Message: "认证失败: " + err.Error()}, nil
	}
	if req.Asset == "" {
		return &pb.DepositResponse{Success: false, Message: "asset is required"}, nil
	}
	if err := s.account.Deposit(ctx, userID, req.Asset, req.Amount); err != nil {
		return &pb.DepositResponse{Success: false, Message: "充值失败: " + err.Error()}, nil
	}
	log.Printf("Deposit: user_id=%d, asset=%s, amount=%s", userID, req.Asset, req.Amount)
	return &pb.DepositResponse{Success: true, Message: "充值成功"}, nil
}

// Withdraw 提现 (gRPC 接口)
func (s *ExchangeService) Withdraw(ctx context.Context, req *pb.WithdrawRequest) (*pb.WithdrawResponse, error) {
	userID, err := s.requireUserID(ctx, req.Token)
	if err != nil {
		return &pb.WithdrawResponse{Success: false, Message: "认证失败: " + err.Error()}, nil
	}
	if req.Asset == "" {
		return &pb.WithdrawResponse{Success: false, Message: "asset is required"}, nil
	}
	if err := s.account.Withdraw(ctx, userID, req.Asset, req.Amount, req.Address); err != nil {
		return &pb.WithdrawResponse{Success: false, Message: "提现失败: " + err.Error()}, nil
	}
	log.Printf("Withdraw: user_id=%d, asset=%s, amount=%s, address=%s", userID, req.Asset, req.Amount, req.Address)
	return &pb.WithdrawResponse{Success: true, Message: "提现成功"}, nil
}

// SubscribeOrderBook 订阅订单簿更新（服务端流式）(gRPC 接口)
// 当 client 为空时，每秒推送一次本地获取到的订单簿；当 client 可用时通过 GetOrderBook
// 拉取（委托至后端 cex-exchange 服务）。
func (s *ExchangeService) SubscribeOrderBook(req *pb.SubscribeOrderBookRequest, stream pb.ExchangeService_SubscribeOrderBookServer) error {
	ctx := stream.Context()
	symbol := req.Symbol
	if symbol == "" {
		symbol = "BTC/USDT"
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			resp, err := s.GetOrderBook(ctx, &pb.GetOrderBookRequest{Symbol: symbol, Depth: 20})
			if err != nil || !resp.Success {
				continue
			}
			update := &pb.OrderBookUpdate{
				Symbol:    symbol,
				Bids:      resp.Bids,
				Asks:      resp.Asks,
				Timestamp: time.Now().Unix(),
			}
			if err := stream.Send(update); err != nil {
				return err
			}
		}
	}
}

// SubscribeTrades 订阅成交更新（服务端流式）(gRPC 接口)
// 每隔 500ms 通过 GetRecentTrades 拉取最新成交并推送到客户端。
func (s *ExchangeService) SubscribeTrades(req *pb.SubscribeTradesRequest, stream pb.ExchangeService_SubscribeTradesServer) error {
	ctx := stream.Context()
	symbol := req.Symbol
	if symbol == "" {
		symbol = "BTC/USDT"
	}

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			resp, err := s.GetRecentTrades(ctx, &pb.GetRecentTradesRequest{Symbol: symbol, Limit: 1})
			if err != nil || len(resp.Trades) == 0 {
				continue
			}
			t := resp.Trades[0]
			update := &pb.TradeUpdate{
				Symbol:    t.Symbol,
				Price:     t.Price,
				Amount:    t.Amount,
				Side:      t.Side,
				Timestamp: time.Now().Unix(),
			}
			if err := stream.Send(update); err != nil {
				return err
			}
		}
	}
}
