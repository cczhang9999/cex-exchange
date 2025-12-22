package data

import (
	pb "backend-v2/api/proto"
	"backend-v2/internal/conf"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// NewExchangeClient creates a gRPC client connection to cex-exchange service
func NewExchangeClient(bc *conf.Bootstrap) (pb.ExchangeServiceClient, func(), error) {
	if bc.Client == nil || bc.Client.Exchange == nil {
		return nil, nil, fmt.Errorf("exchange client configuration is missing")
	}

	addr := bc.Client.Exchange.Addr

	// 建立到 cex-exchange 的连接
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to cex-exchange at %s: %w", addr, err)
	}

	client := pb.NewExchangeServiceClient(conn)

	// 返回清理函数用于关闭连接
	cleanup := func() {
		conn.Close()
	}

	return client, cleanup, nil
}
