package service

import (
	pb "backend-v2/api/proto"
	"backend-v2/internal/biz"
	"context"

	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(NewExchangeService)

type ExchangeService struct {
	pb.UnimplementedExchangeServiceServer
	user   *biz.UserUsecase
	client pb.ExchangeServiceClient
}

func NewExchangeService(user *biz.UserUsecase, client pb.ExchangeServiceClient) *ExchangeService {
	return &ExchangeService{
		user:   user,
		client: client,
	}
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
	return s.client.GetOrderBook(ctx, req)
}

func (s *ExchangeService) GetRecentTrades(ctx context.Context, req *pb.GetRecentTradesRequest) (*pb.GetRecentTradesResponse, error) {
	return s.client.GetRecentTrades(ctx, req)
}

// Implement other methods as Unimplemented or TODO
