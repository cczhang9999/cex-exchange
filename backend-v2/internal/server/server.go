package server

import (
	pb "backend-v2/api/proto"
	"backend-v2/internal/conf"
	"backend-v2/internal/pkg/response"
	"backend-v2/internal/server/middleware"
	"backend-v2/internal/service"
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)
const (
	targetAddr = "localhost:50051" // cex-exchange gRPC default port
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

	// OrderBook endpoint (Dynamic)
	r.GET("/api/orderbook", func(c *gin.Context) {
		symbol := c.Query("symbol")

		log.Println("symbol: ", symbol)
		// 建立临时连接 (为了稳定性和简单性)
		conn, err := grpc.NewClient(targetAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			response.Error(c, 500, "无法连接 gRPC 服务器: "+err.Error())
			return
		}
		defer conn.Close()

		client := pb.NewExchangeServiceClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		resp, err := client.GetOrderBook(ctx, &pb.GetOrderBookRequest{
			Symbol: symbol,
		})
		if err != nil {
			response.Error(c, 500, "gRPC 调用失败: "+err.Error())
			return
		}

		response.Success(c, resp)
	})

		// OrderBook endpoint (Dynamic)
	r.GET("/api/trades", func(c *gin.Context) {
		symbol := c.Query("symbol")
		limitStr := c.DefaultQuery("limit", "20")
		var limit int32
		fmt.Sscanf(limitStr, "%d", &limit)
		log.Println("symbol: ", symbol)
		// 建立临时连接 (为了稳定性和简单性)
		conn, err := grpc.NewClient(targetAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			response.Error(c, 500, "无法连接 gRPC 服务器: "+err.Error())
			return
		}
		defer conn.Close()

		client := pb.NewExchangeServiceClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

	resp, err := client.GetRecentTrades(ctx, &pb.GetRecentTradesRequest{
				Symbol: symbol,
				Limit:  limit,
			})
		if err != nil {
			response.Error(c, 500, "gRPC 调用失败: "+err.Error())
			return
		}

		response.Success(c, resp.Trades)
	})

	// Protected routes
	auth := r.Group("/api", middleware.AuthMiddleware(bc.Auth.JwtSecret))
	{
		auth.GET("/trades11", func(c *gin.Context) {
			symbol := c.Query("symbol")
			limitStr := c.DefaultQuery("limit", "20")
			var limit int32
			fmt.Sscanf(limitStr, "%d", &limit)

			conn, err := grpc.NewClient(targetAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				response.Error(c, 500, "无法连接 gRPC 服务器: "+err.Error())
				return
			}
			defer conn.Close()

			client := pb.NewExchangeServiceClient(conn)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			resp, err := client.GetRecentTrades(ctx, &pb.GetRecentTradesRequest{
				Symbol: symbol,
				Limit:  limit,
			})
			if err != nil {
				response.Error(c, 500, "gRPC 调用失败: "+err.Error())
				return
			}
			response.Success(c, resp)
		})
	}
	

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
