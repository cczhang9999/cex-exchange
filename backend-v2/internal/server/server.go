package server

import (
	pb "backend-v2/api/proto"
	"backend-v2/internal/conf"
	"backend-v2/internal/pkg/response"
	"backend-v2/internal/service"
	"errors"
	"fmt"
	"net"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"gorm.io/gorm"
	"google.golang.org/grpc"
)

var ProviderSet = wire.NewSet(NewGRPCServer, NewHTTPServer)

func NewGRPCServer(bc *conf.Bootstrap, s *service.ExchangeService) *grpc.Server {
	opts := []grpc.ServerOption{}
	srv := grpc.NewServer(opts...)
	pb.RegisterExchangeServiceServer(srv, s)
	return srv
}

func NewHTTPServer(bc *conf.Bootstrap, s *service.ExchangeService) *gin.Engine {
	r := gin.Default()

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
		resp, err := s.GetOrderBook(c.Request.Context(), &pb.GetOrderBookRequest{
			Symbol: symbol,
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

	r.GET("/api/my_orders", func(c *gin.Context) {

		resp, err := s.GetUserOrders(c.Request.Context(), &pb.GetMyOrdersRequest{})
		if err != nil {
			response.Error(c, 500, "失败: "+err.Error())
			return
		}
		response.Success(c, resp.Orders)
	})

	// 查询用户账户列表，可选参数 asset 过滤币种
	r.GET("/api/accounts", func(c *gin.Context) {
		asset := c.Query("asset")

		// TODO: Extract userID from token
		// For now using hardcoded userID 1 as in original code
		accounts, err := s.GetUserAccounts(c.Request.Context(), uint64(1), asset)
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
	return r
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
