package server

import (
	pb "backend-v2/api/proto"
	"backend-v2/internal/conf"
	"backend-v2/internal/pkg/response"
	"backend-v2/internal/server/middleware"
	"backend-v2/internal/service"
	"fmt"
	"net"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
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

	// Protected routes
	auth := r.Group("/api", middleware.AuthMiddleware(bc.Auth.JwtSecret))
	{
		auth.GET("/me", func(c *gin.Context) {
			uid, _ := c.Get("uid")
			response.Success(c, gin.H{
				"user_id": uid,
			})
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
