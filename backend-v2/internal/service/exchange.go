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
	user *biz.UserUsecase
}

func NewExchangeService(user *biz.UserUsecase) *ExchangeService {
	return &ExchangeService{
		user: user,
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

// Implement other methods as Unimplemented or TODO
