# GoFrame 集成指南

## 项目结构

```
cex-exchange/
├── main.go                 # 入口文件
├── manifest/
│   └── config/
│       └── config.yaml     # 配置文件
├── internal/
│   ├── cmd/
│   │   └── cmd.go         # 命令行入口
│   ├── controller/        # 控制器层
│   │   ├── auth/
│   │   ├── order/
│   │   └── funds/
│   ├── service/           # 业务逻辑层
│   ├── dao/               # 数据访问层
│   ├── model/             # 数据模型
│   ├── middleware/        # 中间件
│   └── packed/            # 资源打包
```

## 启动步骤

### 1. 修改配置文件

编辑 `manifest/config/config.yaml`，修改数据库和 Redis 配置：

```yaml
database:
  default:
    host: "127.0.0.1"
    user: "root"
    pass: "your_password"
    name: "cex_exchange"

redis:
  default:
    address: "127.0.0.1:6379"
    pass: "your_redis_password"
```

### 2. 运行项目

```bash
# 开发模式
go run main.go

# 编译运行
go build -o cex-exchange
./cex-exchange
```

### 3. 访问 API

服务默认运行在 `http://localhost:8080`

#### 注册
```bash
curl -X POST http://localhost:8080/api/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"123456","email":"test@example.com"}'
```

#### 登录
```bash
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"123456"}'
```

#### 下单（需要 token）
```bash
curl -X POST http://localhost:8080/api/order \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"symbol":"BTC/USDT","side":"buy","type":"limit","price":"45000","amount":"0.1"}'
```

## GoFrame 核心特性

### 1. 规范路由

使用结构体标签定义路由：

```go
type RegisterReq struct {
    g.Meta   `path:"/register" method:"post" tags:"Auth" summary:"用户注册"`
    Username string `json:"username" v:"required|length:3,20"`
}
```

### 2. 自动参数验证

GoFrame 会自动验证请求参数：

```go
Username string `json:"username" v:"required|length:3,20#请输入用户名|用户名长度为3-20个字符"`
```

### 3. 统一响应格式

通过 `ResponseHandler` 中间件，所有响应都会被包装成：

```json
{
  "code": 0,
  "message": "success",
  "data": {...}
}
```

### 4. 数据库操作

```go
// 查询
g.Model("users").Where("id", 1).One()

// 插入
g.Model("users").Data(g.Map{"username": "test"}).Insert()

// 更新
g.Model("users").Where("id", 1).Update(g.Map{"status": 1})
```

### 5. Redis 操作

```go
// 设置
g.Redis().Set(ctx, "key", "value")

// 获取
g.Redis().Get(ctx, "key")
```

## 下一步

1. 实现 Service 层业务逻辑
2. 使用 DAO 层封装数据库操作
3. 添加 WebSocket 支持
4. 实现撮合引擎
5. 添加单元测试

## 参考资料

- [GoFrame 官方文档](https://goframe.org)
- [GoFrame GitHub](https://github.com/gogf/gf)
