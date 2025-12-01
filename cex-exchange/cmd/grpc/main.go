package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/gogf/gf/v2/frame/g"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "cex-exchange/api/proto"
	grpcServer "cex-exchange/internal/grpc"
)

func main() {
	ctx := g.Ctx()

	// 读取配置
	port := g.Cfg().MustGet(ctx, "grpc.port", 50051).Int()
	
	// 创建 gRPC 服务器
	server := grpc.NewServer(
		grpc.MaxRecvMsgSize(10 * 1024 * 1024), // 10MB
		grpc.MaxSendMsgSize(10 * 1024 * 1024), // 10MB
	)

	// 注册服务
	exchangeServer := grpcServer.NewExchangeServer()
	pb.RegisterExchangeServiceServer(server, exchangeServer)

	// 注册反射服务（用于 grpcurl 等工具）
	reflection.Register(server)

	// 监听端口
	address := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		g.Log().Fatalf(ctx, "Failed to listen: %v", err)
	}

	g.Log().Infof(ctx, "🚀 gRPC server starting on %s", address)

	// 优雅关闭
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		g.Log().Info(ctx, "Shutting down gRPC server...")
		server.GracefulStop()
	}()

	// 启动服务器
	if err := server.Serve(listener); err != nil {
		g.Log().Fatalf(ctx, "Failed to serve: %v", err)
	}
}
