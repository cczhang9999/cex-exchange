package server

import (
	pb "backend-v2/api/proto"
	"backend-v2/internal/biz"
	"backend-v2/internal/conf"
	"backend-v2/internal/pkg/jwt"
	"backend-v2/internal/pkg/response"
	"backend-v2/internal/server/middleware"
	"backend-v2/internal/service"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

var ProviderSet = wire.NewSet(NewGRPCServer, NewHTTPServer)

func NewGRPCServer(bc *conf.Bootstrap, s *service.ExchangeService) *grpc.Server {
	var opts []grpc.ServerOption

	// 如果配置了 JWT secret，则注册认证拦截器
	if bc != nil && bc.Auth != nil && bc.Auth.JwtSecret != "" {
		opts = append(opts, grpc.ChainUnaryInterceptor(middleware.GrpcAuthInterceptor(bc.Auth.JwtSecret)))
	}

	srv := grpc.NewServer(opts...)
	pb.RegisterExchangeServiceServer(srv, s)
	return srv
}

// userIDFromRequest 从 Authorization: Bearer <token> 中提取用户 ID。
// 当配置了 JWT secret 且 token 合法时返回真实用户 ID；
// 否则返回默认值 1（兼容无认证场景 / 旧代码行为）。
func userIDFromRequest(c *gin.Context, secret string) uint64 {
	if secret != "" {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
				if claims, err := jwt.ParseToken(parts[1], secret); err == nil {
					return claims.UserID
				}
			}
		}
	}
	return 1
}

func NewHTTPServer(bc *conf.Bootstrap, s *service.ExchangeService) *gin.Engine {
	r := gin.Default()

	// JWT secret（用于从 Authorization 头部提取用户 ID）
	secret := ""
	if bc != nil && bc.Auth != nil {
		secret = bc.Auth.JwtSecret
	}
	// 便于在各路由处理函数中提取当前用户 ID
	uidOf := func(c *gin.Context) uint64 { return userIDFromRequest(c, secret) }

	// Public routes
	r.POST("/api/login", func(c *gin.Context) {
		type LoginReq struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		var req LoginReq
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, 400, "Invalid request body")
			return
		}

		res, err := s.Login(c.Request.Context(), &pb.LoginRequest{
			Username: req.Username,
			Password: req.Password,
		})
		if err != nil {
			response.Error(c, 500, "Internal server error")
			return
		}

		if !res.Success {
			response.Error(c, 401, res.Message)
			return
		}

		fmt.Printf("User %s logged in successfully\n", req.Username)
		response.Success(c, gin.H{
			"token":   res.Token,
			"user_id": res.UserId,
		})
	})

	//register
	r.POST("/api/register", func(c *gin.Context) {
		type RegisterReq struct {
			Username string `json:"username"`
			Password string `json:"password"`
			Email    string `json:"email"`
			Phone    string `json:"phone"`
		}

		var req RegisterReq
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, 400, "Invalid request body")
			return
		}

		res, err := s.Register(c.Request.Context(), &pb.RegisterRequest{
			Username: req.Username,
			Password: req.Password,
			Email:    req.Email,
			Phone:    req.Phone,
		})
		if err != nil {
			response.Error(c, 500, "Internal server error")
			return
		}

		if !res.Success {
			response.Error(c, 401, res.Message)
			return
		}

		response.Success(c, gin.H{})
	})

	// OrderBook endpoint
	r.GET("/api/orderbook", func(c *gin.Context) {
		symbol := c.Query("symbol")
		depth, _ := strconv.Atoi(c.DefaultQuery("depth", "20"))

		resp, err := s.GetOrderBook(c.Request.Context(), &pb.GetOrderBookRequest{
			Symbol: symbol,
			Depth:  int32(depth),
		})
		if err != nil {
			response.Error(c, 500, "gRPC 调用失败: "+err.Error())
			return
		}
		response.Success(c, resp)
	})

	// Trades endpoint
	r.GET("/api/trades", func(c *gin.Context) {
		symbol := c.Query("symbol")
		limitStr := c.DefaultQuery("limit", "20")
		var limit int32
		fmt.Sscanf(limitStr, "%d", &limit)

		resp, err := s.GetRecentTrades(c.Request.Context(), &pb.GetRecentTradesRequest{
			Symbol: symbol,
			Limit:  limit,
		})
		if err != nil {
			response.Error(c, 500, "gRPC 调用失败: "+err.Error())
			return
		}
		response.Success(c, resp.Trades)
	})

	// 行情 Ticker endpoint
	r.GET("/api/ticker", func(c *gin.Context) {
		symbol := c.Query("symbol")
		resp, err := s.GetTicker(c.Request.Context(), &pb.GetTickerRequest{
			Symbol: symbol,
		})
		if err != nil {
			response.Error(c, 500, "查询失败: "+err.Error())
			return
		}
		response.Success(c, resp.Ticker)
	})

	// K线数据 endpoint
	r.GET("/api/klines", func(c *gin.Context) {
		symbol := c.Query("symbol")
		interval := c.DefaultQuery("interval", "1m")
		startTime, _ := strconv.ParseInt(c.Query("start_time"), 10, 64)
		endTime, _ := strconv.ParseInt(c.Query("end_time"), 10, 64)
		limitStr := c.DefaultQuery("limit", "500")
		var limit int32
		fmt.Sscanf(limitStr, "%d", &limit)

		resp, err := s.GetKlines(c.Request.Context(), &pb.GetKlinesRequest{
			Symbol:    symbol,
			Interval:  interval,
			StartTime: startTime,
			EndTime:   endTime,
			Limit:     limit,
		})
		if err != nil {
			response.Error(c, 500, "查询失败: "+err.Error())
			return
		}
		response.Success(c, resp.Klines)
	})

	// ---------- 以下为需要认证的接口（通过 Authorization: Bearer <token>） ----------

	// 查询用户订单（全部）
	r.GET("/api/my_orders", func(c *gin.Context) {
		// token 透传给 service 做统一校验/解析
		token := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
		resp, err := s.GetUserOrders(c.Request.Context(), &pb.GetMyOrdersRequest{
			Symbol: c.Query("symbol"),
			Token:  token,
		})
		if err != nil {
			response.Error(c, 500, "失败: "+err.Error())
			return
		}
		response.Success(c, resp.Orders)
	})

	// 查询用户成交记录
	r.GET("/api/my_trades", func(c *gin.Context) {
		trades, err := s.GetMyTrades(c.Request.Context(), uidOf(c))
		if err != nil {
			response.Error(c, 500, "查询失败: "+err.Error())
			return
		}
		response.Success(c, trades)
	})

	// 下单
	r.POST("/api/order", func(c *gin.Context) {
		type PlaceOrderReq struct {
			Symbol string `json:"symbol"`
			Side   string `json:"side"`
			Type   string `json:"type"`
			Price  string `json:"price"`
			Amount string `json:"amount"`
		}
		var req PlaceOrderReq
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, 400, "Invalid request body")
			return
		}
		token := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
		resp, err := s.PlaceOrder(c.Request.Context(), &pb.PlaceOrderRequest{
			Symbol: req.Symbol,
			Side:   req.Side,
			Type:   req.Type,
			Price:  req.Price,
			Amount: req.Amount,
			Token:  token,
		})
		if err != nil {
			response.Error(c, 500, "下单失败: "+err.Error())
			return
		}
		if !resp.Success {
			response.Error(c, 400, resp.Message)
			return
		}
		response.Success(c, gin.H{"order_id": resp.OrderId, "message": resp.Message})
	})

	// 撤单
	r.POST("/api/order/cancel/:id", func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			response.Error(c, 400, "无效的订单 ID")
			return
		}
		token := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
		resp, err := s.CancelOrder(c.Request.Context(), &pb.CancelOrderRequest{
			OrderId: id,
			Token:   token,
		})
		if err != nil {
			response.Error(c, 500, "撤单失败: "+err.Error())
			return
		}
		if !resp.Success {
			response.Error(c, 400, resp.Message)
			return
		}
		response.Success(c, gin.H{
			"order_id": id,
			"message":  resp.Message,
		})
	})

	// 查询用户账户列表，可选参数 asset 过滤币种
	r.GET("/api/accounts", func(c *gin.Context) {
		asset := c.Query("asset")
		accounts, err := s.GetUserAccounts(c.Request.Context(), uidOf(c), asset)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				response.Error(c, 404, "账户不存在")
				return
			}
			response.Error(c, 500, "查询失败: "+err.Error())
			return
		}
		response.Success(c, accounts)
	})

	// 查询用户资金流水列表
	r.GET("/api/account_flows", func(c *gin.Context) {
		q := biz.AccountFlowQuery{
			Asset:      c.Query("asset"),
			ChangeType: c.Query("change_type"),
			UserID:     uidOf(c),
		}

		if p, err := strconv.Atoi(c.Query("page")); err == nil && p > 0 {
			q.Page = p
		}
		if ps, err := strconv.Atoi(c.Query("page_size")); err == nil && ps > 0 {
			q.PageSize = ps
		}
		if t, ok := parseQueryTime(c.Query("start_time")); ok {
			q.StartTime = t
		}
		if t, ok := parseQueryTime(c.Query("end_time")); ok {
			q.EndTime = t
		}

		resp, err := s.GetUserAccountFlows(c.Request.Context(), q)
		if err != nil {
			response.Error(c, 500, "查询失败: "+err.Error())
			return
		}
		response.Success(c, resp)
	})

	// 查询用户订单分页列表
	r.GET("/api/orders", func(c *gin.Context) {
		q := biz.OrderQuery{
			UserID: uidOf(c),
			Symbol: c.Query("symbol"),
		}

		if status := biz.OrderStatus(c.Query("status")); status != "" {
			q.Status = status
		}
		if side := biz.OrderSide(c.Query("side")); side != "" {
			q.Side = side
		}
		if p, err := strconv.Atoi(c.Query("page")); err == nil && p > 0 {
			q.Page = p
		}
		if ps, err := strconv.Atoi(c.Query("page_size")); err == nil && ps > 0 {
			q.PageSize = ps
		}
		if t, ok := parseQueryTime(c.Query("start_time")); ok {
			q.StartTime = t
		}
		if t, ok := parseQueryTime(c.Query("end_time")); ok {
			q.EndTime = t
		}

		resp, err := s.GetUserOrderPage(c.Request.Context(), q)
		if err != nil {
			response.Error(c, 500, "查询失败: "+err.Error())
			return
		}
		response.Success(c, resp)
	})

	// 充值
	r.POST("/api/deposit", func(c *gin.Context) {
		type DepositReq struct {
			Asset  string `json:"asset"`
			Amount string `json:"amount"`
			Token  string `json:"token"`
		}
		var req DepositReq
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, 400, "Invalid request body")
			return
		}
		if req.Token == "" {
			req.Token = strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
		}
		resp, err := s.Deposit(c.Request.Context(), &pb.DepositRequest{
			Asset:  req.Asset,
			Amount: req.Amount,
			Token:  req.Token,
		})
		if err != nil {
			response.Error(c, 500, "充值失败: "+err.Error())
			return
		}
		if !resp.Success {
			response.Error(c, 400, resp.Message)
			return
		}
		response.Success(c, gin.H{"message": resp.Message})
	})

	// 提现
	r.POST("/api/withdraw", func(c *gin.Context) {
		type WithdrawReq struct {
			Asset   string `json:"asset"`
			Amount  string `json:"amount"`
			Address string `json:"address"`
			Token   string `json:"token"`
		}
		var req WithdrawReq
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, 400, "Invalid request body")
			return
		}
		if req.Token == "" {
			req.Token = strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
		}
		resp, err := s.Withdraw(c.Request.Context(), &pb.WithdrawRequest{
			Asset:   req.Asset,
			Amount:  req.Amount,
			Address: req.Address,
			Token:   req.Token,
		})
		if err != nil {
			response.Error(c, 500, "提现失败: "+err.Error())
			return
		}
		if !resp.Success {
			response.Error(c, 400, resp.Message)
			return
		}
		response.Success(c, gin.H{"message": resp.Message})
	})

	// ---------- 管理员接口 ----------
	admin := r.Group("/api/admin")
	admin.Use(func(c *gin.Context) {
		// 管理员接口通过 Authorization 头部提取用户 ID
		_ = uidOf(c)
		c.Next()
	})
	{
		// 用户列表（支持搜索与分页）
		admin.GET("/users", func(c *gin.Context) {
			users, err := s.ListUsers(c.Request.Context())
			if err != nil {
				response.Error(c, 500, "查询失败: "+err.Error())
				return
			}

			// 按用户名或邮箱搜索（不区分大小写）
			search := strings.TrimSpace(strings.ToLower(c.Query("search")))
			filtered := make([]gin.H, 0, len(users))
			for _, u := range users {
				if search != "" &&
					!strings.Contains(strings.ToLower(u.Username), search) &&
					!strings.Contains(strings.ToLower(u.Email), search) {
					continue
				}
				filtered = append(filtered, gin.H{
					"id":         u.ID,
					"username":   u.Username,
					"email":      u.Email,
					"phone":      u.Phone,
					"status":     u.Status,
					"is_blocked": u.IsBlocked(),
					"created_at": u.CreatedAt.Format("2006-01-02 15:04:05"),
					"updated_at": u.UpdatedAt.Format("2006-01-02 15:04:05"),
				})
			}

			// 分页
			page, err := strconv.Atoi(c.Query("page"))
			if err != nil || page < 1 {
				page = 1
			}
			limit, err := strconv.Atoi(c.Query("limit"))
			if err != nil || limit < 1 {
				limit = 20
			}
			total := len(filtered)
			start := (page - 1) * limit
			if start > total {
				start = total
			}
			end := start + limit
			if end > total {
				end = total
			}

			response.Success(c, gin.H{
				"users": filtered[start:end],
				"total": total,
				"page":  page,
				"limit": limit,
			})
		})

		// 订单列表（分页）
		admin.GET("/orders", func(c *gin.Context) {
			q := biz.OrderQuery{
				Symbol: c.Query("symbol"),
			}
			if status := biz.OrderStatus(c.Query("status")); status != "" {
				q.Status = status
			}
			if side := biz.OrderSide(c.Query("side")); side != "" {
				q.Side = side
			}
			if p, err := strconv.Atoi(c.Query("page")); err == nil && p > 0 {
				q.Page = p
			}
			if ps, err := strconv.Atoi(c.Query("page_size")); err == nil && ps > 0 {
				q.PageSize = ps
			}
			if t, ok := parseQueryTime(c.Query("start_time")); ok {
				q.StartTime = t
			}
			if t, ok := parseQueryTime(c.Query("end_time")); ok {
				q.EndTime = t
			}
			resp, err := s.GetUserOrderPage(c.Request.Context(), q)
			if err != nil {
				response.Error(c, 500, "查询失败: "+err.Error())
				return
			}
			response.Success(c, resp)
		})

		// 账户列表（分页）
		admin.GET("/accounts", func(c *gin.Context) {
			accounts, err := s.ListAllAccounts(c.Request.Context())
			if err != nil {
				response.Error(c, 500, "查询失败: "+err.Error())
				return
			}
			response.Success(c, gin.H{
				"accounts": accounts,
				"total":    len(accounts),
			})
		})

		// 添加资金
		admin.POST("/accounts/add-funds", func(c *gin.Context) {
			type AddFundsReq struct {
				UserID uint64 `json:"user_id"`
				Asset  string `json:"asset"`
				Amount string `json:"amount"`
				Remark string `json:"remark"`
			}
			var req AddFundsReq
			if err := c.ShouldBindJSON(&req); err != nil {
				response.Error(c, 400, "Invalid request body")
				return
			}
			if err := s.AdminAddFunds(c.Request.Context(), req.UserID, req.Asset, req.Amount); err != nil {
				response.Error(c, 400, err.Error())
				return
			}
			response.Success(c, gin.H{"message": "添加资金成功"})
		})

		// 调整资金
		admin.POST("/accounts/adjust", func(c *gin.Context) {
			type AdjustReq struct {
				UserID uint64 `json:"user_id"`
				Asset  string `json:"asset"`
				Amount string `json:"amount"`
				Remark string `json:"remark"`
				Type   string `json:"type"`
			}
			var req AdjustReq
			if err := c.ShouldBindJSON(&req); err != nil {
				response.Error(c, 400, "Invalid request body")
				return
			}
			if err := s.AdminAdjustBalance(c.Request.Context(), req.UserID, req.Asset, req.Amount); err != nil {
				response.Error(c, 400, err.Error())
				return
			}
			response.Success(c, gin.H{"message": "调账成功"})
		})

		// 管理员撤单
		admin.POST("/orders/:id/cancel", func(c *gin.Context) {
			id, err := strconv.ParseUint(c.Param("id"), 10, 64)
			if err != nil {
				response.Error(c, 400, "无效的订单 ID")
				return
			}
			resp, err := s.CancelOrder(c.Request.Context(), &pb.CancelOrderRequest{OrderId: id})
			if err != nil {
				response.Error(c, 500, "撤单失败: "+err.Error())
				return
			}
			if !resp.Success {
				response.Error(c, 400, resp.Message)
				return
			}
			response.Success(c, gin.H{"order_id": id, "message": resp.Message})
		})

		// 封禁/解封用户
		admin.POST("/users/:id/block", func(c *gin.Context) {
			id, err := strconv.ParseUint(c.Param("id"), 10, 64)
			if err != nil {
				response.Error(c, 400, "无效的用户 ID")
				return
			}
			type BlockReq struct {
				Block bool `json:"block"`
			}
			var req BlockReq
			block := true // 默认封禁
			if err := c.ShouldBindJSON(&req); err == nil {
				block = req.Block
			}
			if err := s.AdminBlockUser(c.Request.Context(), id, block); err != nil {
				response.Error(c, 400, err.Error())
				return
			}
			if block {
				response.Success(c, gin.H{"message": "用户已封禁"})
			} else {
				response.Success(c, gin.H{"message": "用户已解封"})
			}
		})
	}

	// WebSocket routes
	wsMgr := newWSManager(s)
	r.GET("/ws", wsMgr.handleWebSocket)
	r.GET("/ws/status", wsMgr.getStatus)

	return r
}

// parseQueryTime 解析查询参数时间，支持 RFC3339 或 "2006-01-02 15:04:05" 格式
func parseQueryTime(s string) (time.Time, bool) {
	if s == "" {
		return time.Time{}, false
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, true
	}
	if t, err := time.Parse("2006-01-02 15:04:05", s); err == nil {
		return t, true
	}
	return time.Time{}, false
}

type Server struct {
	Grpc *grpc.Server
	Http *gin.Engine
	Conf *conf.Server
}

func NewServer(grpc *grpc.Server, http *gin.Engine, bc *conf.Bootstrap) *Server {
	return &Server{Grpc: grpc, Http: http, Conf: bc.Server}
}

func (s *Server) Run() error {
	// Start gRPC
	lis, err := net.Listen("tcp", s.Conf.Grpc.Addr)
	if err != nil {
		return err
	}
	fmt.Printf("gRPC server listening on %s\n", s.Conf.Grpc.Addr)
	go func() {
		if err := s.Grpc.Serve(lis); err != nil {
			panic(err)
		}
	}()

	// Start HTTP
	fmt.Printf("HTTP server listening on %s\n", s.Conf.Http.Addr)
	return s.Http.Run(s.Conf.Http.Addr)
}
