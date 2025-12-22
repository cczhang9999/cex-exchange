package service

import (
	pb "backend-v2/api/proto"
	"backend-v2/internal/biz"
	"context"
	"fmt"

	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(NewExchangeService)

type ExchangeService struct {
	pb.UnimplementedExchangeServiceServer
	user   *biz.UserUsecase
	order  *biz.OrderUsecase
	client pb.ExchangeServiceClient
}

func NewExchangeService(user *biz.UserUsecase, order *biz.OrderUsecase, client pb.ExchangeServiceClient) *ExchangeService {
	return &ExchangeService{
		user:   user,
		order:  order,
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

func (s *ExchangeService) GetUserOrders(ctx context.Context, req *pb.GetMyOrdersRequest) (*pb.GetMyOrdersResponse, error) {
	// TODO: Extract userID from token in req.Token
	// For now using hardcoded userID 1 as in original code
	orders, err := s.order.GetUserOrders(ctx, uint64(1))
	if err != nil {
		return &pb.GetMyOrdersResponse{Success: false, Message: err.Error()}, nil
	}

	pbOrders := make([]*pb.Order, 0, len(orders))
	for _, o := range orders {
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

	return &pb.GetMyOrdersResponse{
		Success: true,
		Message: "Success",
		Orders:  pbOrders,
	}, nil
}

// Implement other methods as Unimplemented or TODO
