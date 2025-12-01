# gRPC 后台实现文档

本文档介绍如何使用 gRPC 和 protobuf 实现的交易所后台服务。

## 📋 目录

- [项目结构](#项目结构)
- [环境准备](#环境准备)
- [快速开始](#快速开始)
- [API 文档](#api-文档)
- [客户端示例](#客户端示例)
- [测试工具](#测试工具)

## 🗂️ 项目结构

```
cex-exchange/
├── api/
│   └── proto/
│       ├── exchange.proto          # protobuf 定义文件
│       ├── exchange.pb.go          # 生成的 Go 代码（消息）
│       └── exchange_grpc.pb.go     # 生成的 Go 代码（服务）
├── internal/
│   └── grpc/
│       └── server.go               # gRPC 服务器实现
├── cmd/
│   └── grpc/
│       └── main.go                 # gRPC 服务器启动程序
├── examples/
│   └── grpc_client.go              # 客户端示例代码
└── generate_proto.sh               # protobuf 代码生成脚本
```

## 🛠️ 环境准备

### 1. 安装 protoc 编译器

**macOS:**
```bash
brew install protobuf
```

**Linux:**
```bash
# Ubuntu/Debian
sudo apt-get install protobuf-compiler

# CentOS/RHEL
sudo yum install protobuf-compiler
```

### 2. 安装 Go 插件

```bash
# 安装 protoc-gen-go（生成消息代码）
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

# 安装 protoc-gen-go-grpc（生成服务代码）
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# 确保 $GOPATH/bin 在 PATH 中
export PATH="$PATH:$(go env GOPATH)/bin"
```

### 3. 安装依赖

```bash
cd cex-exchange
go mod tidy
```

## 🚀 快速开始

### 步骤 1: 生成 protobuf 代码

```bash
# 执行代码生成脚本
./generate_proto.sh
```

这将生成以下文件：
- `api/proto/exchange.pb.go` - protobuf 消息定义
- `api/proto/exchange_grpc.pb.go` - gRPC 服务定义

### 步骤 2: 更新配置文件

在 `manifest/config/config.yaml` 中添加 gRPC 配置：

```yaml
grpc:
  port: 50051
```

### 步骤 3: 启动 gRPC 服务器

```bash
# 方式 1: 直接运行
go run cmd/grpc/main.go

# 方式 2: 编译后运行
go build -o grpc-server cmd/grpc/main.go
./grpc-server
```

服务器将在 `localhost:50051` 上启动。

### 步骤 4: 运行客户端示例

```bash
# 在另一个终端运行客户端
go run examples/grpc_client.go
```

## 📚 API 文档

### 认证服务

#### 1. 用户注册
```protobuf
rpc Register(RegisterRequest) returns (RegisterResponse);
```

**请求参数:**
- `username`: 用户名
- `password`: 密码
- `email`: 邮箱

**响应:**
- `success`: 是否成功
- `message`: 消息
- `user_id`: 用户 ID

#### 2. 用户登录
```protobuf
rpc Login(LoginRequest) returns (LoginResponse);
```

**请求参数:**
- `username`: 用户名
- `password`: 密码

**响应:**
- `success`: 是否成功
- `message`: 消息
- `token`: 认证 token
- `user_id`: 用户 ID

### 订单服务

#### 3. 下单
```protobuf
rpc PlaceOrder(PlaceOrderRequest) returns (PlaceOrderResponse);
```

**请求参数:**
- `symbol`: 交易对（如 "BTC/USDT"）
- `side`: 方向（"buy" 或 "sell"）
- `type`: 类型（"limit" 或 "market"）
- `price`: 价格
- `amount`: 数量
- `token`: 认证 token

**响应:**
- `success`: 是否成功
- `message`: 消息
- `order_id`: 订单 ID

#### 4. 撤单
```protobuf
rpc CancelOrder(CancelOrderRequest) returns (CancelOrderResponse);
```

#### 5. 获取我的订单
```protobuf
rpc GetMyOrders(GetMyOrdersRequest) returns (GetMyOrdersResponse);
```

#### 6. 获取订单簿
```protobuf
rpc GetOrderBook(GetOrderBookRequest) returns (GetOrderBookResponse);
```

### 账户服务

#### 7. 获取余额
```protobuf
rpc GetBalance(GetBalanceRequest) returns (GetBalanceResponse);
```

#### 8. 充值
```protobuf
rpc Deposit(DepositRequest) returns (DepositResponse);
```

#### 9. 提现
```protobuf
rpc Withdraw(WithdrawRequest) returns (WithdrawResponse);
```

### 市场数据服务

#### 10. 获取行情
```protobuf
rpc GetTicker(GetTickerRequest) returns (GetTickerResponse);
```

#### 11. 获取 K 线
```protobuf
rpc GetKlines(GetKlinesRequest) returns (GetKlinesResponse);
```

#### 12. 获取最近成交
```protobuf
rpc GetRecentTrades(GetRecentTradesRequest) returns (GetRecentTradesResponse);
```

### 实时数据流

#### 13. 订阅订单簿（服务端流式）
```protobuf
rpc SubscribeOrderBook(SubscribeOrderBookRequest) returns (stream OrderBookUpdate);
```

#### 14. 订阅成交（服务端流式）
```protobuf
rpc SubscribeTrades(SubscribeTradesRequest) returns (stream TradeUpdate);
```

## 💻 客户端示例

### Go 客户端

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
        Username: "testuser",
        Password: "password123",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    token := loginResp.Token
    
    // 下单
    orderResp, err := client.PlaceOrder(ctx, &pb.PlaceOrderRequest{
        Symbol: "BTC/USDT",
        Side:   "buy",
        Type:   "limit",
        Price:  "50000",
        Amount: "0.1",
        Token:  token,
    })
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("订单 ID: %d", orderResp.OrderId)
}
```

### Python 客户端

首先生成 Python 代码：

```bash
# 安装 grpcio-tools
pip install grpcio-tools

# 生成 Python 代码
python -m grpc_tools.protoc -I. \
    --python_out=. \
    --grpc_python_out=. \
    api/proto/exchange.proto
```

然后使用：

```python
import grpc
import exchange_pb2
import exchange_pb2_grpc

# 连接服务器
channel = grpc.insecure_channel('localhost:50051')
client = exchange_pb2_grpc.ExchangeServiceStub(channel)

# 登录
login_resp = client.Login(exchange_pb2.LoginRequest(
    username='testuser',
    password='password123'
))

token = login_resp.token

# 下单
order_resp = client.PlaceOrder(exchange_pb2.PlaceOrderRequest(
    symbol='BTC/USDT',
    side='buy',
    type='limit',
    price='50000',
    amount='0.1',
    token=token
))

print(f"订单 ID: {order_resp.order_id}")
```

### JavaScript/Node.js 客户端

```bash
# 安装依赖
npm install @grpc/grpc-js @grpc/proto-loader
```

```javascript
const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');

// 加载 proto 文件
const packageDefinition = protoLoader.loadSync('api/proto/exchange.proto');
const proto = grpc.loadPackageDefinition(packageDefinition).exchange;

// 创建客户端
const client = new proto.ExchangeService(
    'localhost:50051',
    grpc.credentials.createInsecure()
);

// 登录
client.Login({
    username: 'testuser',
    password: 'password123'
}, (err, response) => {
    if (err) {
        console.error(err);
        return;
    }
    
    const token = response.token;
    
    // 下单
    client.PlaceOrder({
        symbol: 'BTC/USDT',
        side: 'buy',
        type: 'limit',
        price: '50000',
        amount: '0.1',
        token: token
    }, (err, response) => {
        if (err) {
            console.error(err);
            return;
        }
        console.log('订单 ID:', response.order_id);
    });
});
```

## 🧪 测试工具

### 使用 grpcurl

安装 grpcurl：
```bash
brew install grpcurl
```

列出所有服务：
```bash
grpcurl -plaintext localhost:50051 list
```

列出服务的方法：
```bash
grpcurl -plaintext localhost:50051 list exchange.ExchangeService
```

调用方法：
```bash
# 登录
grpcurl -plaintext -d '{
  "username": "testuser",
  "password": "password123"
}' localhost:50051 exchange.ExchangeService/Login

# 下单
grpcurl -plaintext -d '{
  "symbol": "BTC/USDT",
  "side": "buy",
  "type": "limit",
  "price": "50000",
  "amount": "0.1",
  "token": "YOUR_TOKEN_HERE"
}' localhost:50051 exchange.ExchangeService/PlaceOrder
```

### 使用 BloomRPC

1. 下载并安装 [BloomRPC](https://github.com/bloomrpc/bloomrpc)
2. 导入 `api/proto/exchange.proto` 文件
3. 设置服务器地址为 `localhost:50051`
4. 在 GUI 中测试各个接口

## 🔧 高级配置

### TLS/SSL 加密

生成证书：
```bash
# 生成私钥
openssl genrsa -out server.key 2048

# 生成证书
openssl req -new -x509 -key server.key -out server.crt -days 365
```

修改服务器代码：
```go
creds, err := credentials.NewServerTLSFromFile("server.crt", "server.key")
if err != nil {
    log.Fatal(err)
}

server := grpc.NewServer(grpc.Creds(creds))
```

### 拦截器（中间件）

添加日志拦截器：
```go
func loggingInterceptor(ctx context.Context, req interface{}, 
    info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
    
    start := time.Now()
    resp, err := handler(ctx, req)
    duration := time.Since(start)
    
    log.Printf("Method: %s, Duration: %v, Error: %v", 
        info.FullMethod, duration, err)
    
    return resp, err
}

server := grpc.NewServer(
    grpc.UnaryInterceptor(loggingInterceptor),
)
```

## 📊 性能优化

### 1. 连接池

客户端使用连接池：
```go
var (
    connPool []*grpc.ClientConn
    poolSize = 10
)

func initPool() {
    for i := 0; i < poolSize; i++ {
        conn, _ := grpc.Dial("localhost:50051", 
            grpc.WithTransportCredentials(insecure.NewCredentials()))
        connPool = append(connPool, conn)
    }
}
```

### 2. 消息压缩

启用 gzip 压缩：
```go
// 服务端
server := grpc.NewServer(
    grpc.RPCCompressor(grpc.NewGZIPCompressor()),
)

// 客户端
conn, _ := grpc.Dial("localhost:50051",
    grpc.WithDefaultCallOptions(grpc.UseCompressor(gzip.Name)),
)
```

### 3. 流式处理

对于大量数据，使用流式 RPC：
```go
stream, err := client.SubscribeOrderBook(ctx, &pb.SubscribeOrderBookRequest{
    Symbol: "BTC/USDT",
})

for {
    update, err := stream.Recv()
    if err == io.EOF {
        break
    }
    // 处理更新
}
```

## 🐛 故障排查

### 常见问题

1. **protoc 命令未找到**
   - 确保已安装 protobuf 编译器
   - 检查 PATH 环境变量

2. **protoc-gen-go 未找到**
   - 运行 `go install google.golang.org/protobuf/cmd/protoc-gen-go@latest`
   - 确保 `$GOPATH/bin` 在 PATH 中

3. **连接被拒绝**
   - 检查服务器是否已启动
   - 确认端口号正确（默认 50051）

4. **认证失败**
   - 检查 token 是否有效
   - 确认 token 未过期

## 📝 最佳实践

1. **错误处理**: 始终检查错误并返回有意义的错误消息
2. **超时设置**: 为每个请求设置合理的超时时间
3. **重试机制**: 对于网络错误，实现指数退避重试
4. **监控**: 添加 metrics 和 tracing
5. **版本控制**: 使用 protobuf 的版本控制功能

## 🔗 相关资源

- [gRPC 官方文档](https://grpc.io/docs/)
- [Protocol Buffers 文档](https://developers.google.com/protocol-buffers)
- [GoFrame 文档](https://goframe.org/)

## 📄 许可证

MIT License
