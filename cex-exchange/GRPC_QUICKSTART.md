# gRPC 后台实现 - 快速开始指南

## 📋 概述

本项目实现了一个完整的加密货币交易所 gRPC 后台服务，包含以下功能：

- ✅ 用户认证（注册、登录）
- ✅ 订单管理（下单、撤单、查询）
- ✅ 账户管理（余额查询、充值、提现）
- ✅ 市场数据（行情、K线、成交记录）
- ✅ 实时数据流（订单簿推送、成交推送）

## 🚀 快速开始

### 步骤 1: 安装 protoc 编译器

**macOS:**
```bash
brew install protobuf
```

**验证安装:**
```bash
protoc --version
# 应该输出类似: libprotoc 3.21.12
```

### 步骤 2: 安装 Go 插件

```bash
# 安装 protoc-gen-go（生成消息代码）
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

# 安装 protoc-gen-go-grpc（生成服务代码）
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# 确保 $GOPATH/bin 在 PATH 中
export PATH="$PATH:$(go env GOPATH)/bin"

# 验证安装
which protoc-gen-go
which protoc-gen-go-grpc
```

### 步骤 3: 生成 protobuf 代码

```bash
cd /Users/zhanjun/IdeaProjects/cex-exchange/cex-exchange

# 运行代码生成脚本
./generate_proto.sh
```

成功后会生成：
- `api/proto/exchange.pb.go` - protobuf 消息定义
- `api/proto/exchange_grpc.pb.go` - gRPC 服务定义

### 步骤 4: 更新配置文件

在 `manifest/config/config.yaml` 中添加 gRPC 配置：

```yaml
# gRPC 配置
grpc:
  port: 50051

# JWT 配置（如果还没有）
jwt:
  secret: "your-secret-key-change-this-in-production"
  expires: 86400  # 24小时
```

### 步骤 5: 启动 gRPC 服务器

```bash
# 方式 1: 直接运行
go run cmd/grpc/main.go

# 方式 2: 编译后运行
go build -o grpc-server cmd/grpc/main.go
./grpc-server
```

你应该看到：
```
🚀 gRPC server starting on :50051
```

### 步骤 6: 测试 gRPC 服务

#### 使用 Go 客户端测试

```bash
# 在另一个终端运行客户端示例
go run examples/grpc_client.go
```

#### 使用 grpcurl 测试

```bash
# 安装 grpcurl
brew install grpcurl

# 列出所有服务
grpcurl -plaintext localhost:50051 list

# 列出服务的方法
grpcurl -plaintext localhost:50051 list exchange.ExchangeService

# 调用登录接口
grpcurl -plaintext -d '{
  "username": "testuser",
  "password": "password123"
}' localhost:50051 exchange.ExchangeService/Login
```

## 📁 项目结构

```
cex-exchange/
├── api/
│   └── proto/
│       ├── exchange.proto          # protobuf 定义文件
│       ├── exchange.pb.go          # 生成的消息代码
│       └── exchange_grpc.pb.go     # 生成的服务代码
├── internal/
│   ├── grpc/
│   │   └── server.go               # gRPC 服务器实现
│   ├── service/
│   │   ├── auth.go                 # 认证服务
│   │   ├── order.go                # 订单服务
│   │   ├── funds.go                # 资金服务
│   │   └── market.go               # 市场数据服务
│   ├── dao/                        # 数据访问层
│   └── model/                      # 数据模型
├── cmd/
│   └── grpc/
│       └── main.go                 # gRPC 服务器启动程序
├── examples/
│   └── grpc_client.go              # 客户端示例代码
├── generate_proto.sh               # protobuf 代码生成脚本
├── GRPC_README.md                  # 详细文档
└── GRPC_QUICKSTART.md              # 本文件
```

## 🔧 API 使用示例

### 1. 用户注册

```bash
grpcurl -plaintext -d '{
  "username": "alice",
  "password": "password123",
  "email": "alice@example.com"
}' localhost:50051 exchange.ExchangeService/Register
```

### 2. 用户登录

```bash
grpcurl -plaintext -d '{
  "username": "alice",
  "password": "password123"
}' localhost:50051 exchange.ExchangeService/Login
```

响应示例：
```json
{
  "success": true,
  "message": "登录成功",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "userId": "1"
}
```

### 3. 获取账户余额

```bash
grpcurl -plaintext -d '{
  "token": "YOUR_TOKEN_HERE"
}' localhost:50051 exchange.ExchangeService/GetBalance
```

### 4. 充值

```bash
grpcurl -plaintext -d '{
  "asset": "USDT",
  "amount": "10000",
  "token": "YOUR_TOKEN_HERE"
}' localhost:50051 exchange.ExchangeService/Deposit
```

### 5. 下单

```bash
grpcurl -plaintext -d '{
  "symbol": "BTC/USDT",
  "side": "buy",
  "type": "limit",
  "price": "50000",
  "amount": "0.1",
  "token": "YOUR_TOKEN_HERE"
}' localhost:50051 exchange.ExchangeService/PlaceOrder
```

### 6. 获取我的订单

```bash
grpcurl -plaintext -d '{
  "symbol": "BTC/USDT",
  "token": "YOUR_TOKEN_HERE"
}' localhost:50051 exchange.ExchangeService/GetMyOrders
```

### 7. 获取订单簿

```bash
grpcurl -plaintext -d '{
  "symbol": "BTC/USDT",
  "depth": 10
}' localhost:50051 exchange.ExchangeService/GetOrderBook
```

### 8. 获取行情

```bash
grpcurl -plaintext -d '{
  "symbol": "BTC/USDT"
}' localhost:50051 exchange.ExchangeService/GetTicker
```

## 💡 Go 客户端代码示例

```go
package main

import (
    "context"
    "log"
    
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    
    pb "cex-exchange/api/proto"
)

func main() {
    // 连接服务器
    conn, err := grpc.Dial("localhost:50051", 
        grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()
    
    client := pb.NewExchangeServiceClient(conn)
    ctx := context.Background()
    
    // 登录
    loginResp, err := client.Login(ctx, &pb.LoginRequest{
        Username: "alice",
        Password: "password123",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("登录成功! Token: %s, UserID: %d", 
        loginResp.Token, loginResp.UserId)
    
    // 获取余额
    balanceResp, err := client.GetBalance(ctx, &pb.GetBalanceRequest{
        Token: loginResp.Token,
    })
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("账户余额:")
    for _, balance := range balanceResp.Balances {
        log.Printf("  %s: 可用=%s, 冻结=%s", 
            balance.Asset, balance.Balance, balance.Frozen)
    }
}
```

## 🐛 常见问题

### 1. protoc 命令未找到

**解决方案:**
```bash
brew install protobuf
```

### 2. protoc-gen-go 未找到

**解决方案:**
```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
export PATH="$PATH:$(go env GOPATH)/bin"
```

### 3. 连接被拒绝

**检查项:**
- gRPC 服务器是否已启动
- 端口号是否正确（默认 50051）
- 防火墙是否阻止连接

### 4. token 无效

**检查项:**
- token 是否正确复制
- token 是否已过期
- JWT secret 配置是否正确

## 📊 性能测试

使用 ghz 进行压力测试：

```bash
# 安装 ghz
brew install ghz

# 测试登录接口
ghz --insecure \
  --proto api/proto/exchange.proto \
  --call exchange.ExchangeService/Login \
  -d '{"username":"alice","password":"password123"}' \
  -n 10000 \
  -c 100 \
  localhost:50051
```

## 🔐 生产环境部署

### 1. 启用 TLS

生成证书：
```bash
openssl genrsa -out server.key 2048
openssl req -new -x509 -key server.key -out server.crt -days 365
```

修改服务器代码启用 TLS（参考 GRPC_README.md）

### 2. 使用环境变量

```bash
export GRPC_PORT=50051
export JWT_SECRET="your-production-secret"
export DB_HOST="your-db-host"
```

### 3. Docker 部署

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o grpc-server cmd/grpc/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/grpc-server .
COPY manifest/config/config.yaml manifest/config/
EXPOSE 50051
CMD ["./grpc-server"]
```

## 📚 更多资源

- [完整 API 文档](./GRPC_README.md)
- [gRPC 官方文档](https://grpc.io/docs/)
- [Protocol Buffers 文档](https://developers.google.com/protocol-buffers)

## ✅ 下一步

1. 阅读完整的 [GRPC_README.md](./GRPC_README.md)
2. 查看 [examples/grpc_client.go](./examples/grpc_client.go) 了解更多示例
3. 根据业务需求扩展 protobuf 定义
4. 实现更多的 gRPC 服务方法
5. 添加监控和日志
6. 进行性能优化

## 🎉 完成！

现在你已经成功设置并运行了 gRPC 后台服务！
