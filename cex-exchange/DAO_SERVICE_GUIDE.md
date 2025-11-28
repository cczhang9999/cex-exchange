# DAO 和 Service 层使用指南

## 📁 项目分层架构

```
internal/
├── model/          # 数据模型（Entity）
│   └── entity.go
├── dao/            # 数据访问层（Data Access Object）
│   ├── user.go
│   ├── order.go
│   ├── account.go
│   ├── trade.go
│   ├── account_flow.go
│   └── kline.go
├── service/        # 业务逻辑层（Service）
│   ├── auth.go
│   ├── order.go
│   ├── account.go
│   └── admin.go
└── controller/     # 控制器层（Controller）
    ├── auth/
    ├── order/
    ├── funds/
    └── admin/
```

## 🎯 分层职责

### 1. Model 层（数据模型）
定义数据库表对应的结构体，包含字段映射和 JSON 序列化规则。

```go
type User struct {
    ID        uint64      `json:"id"`
    Username  string      `json:"username"`
    Password  string      `json:"-"` // 不返回给前端
    Email     string      `json:"email"`
    CreatedAt *gtime.Time `json:"created_at"`
}
```

### 2. DAO 层（数据访问）
封装所有数据库操作，提供 CRUD 方法。

**特点：**
- 只负责数据库操作
- 不包含业务逻辑
- 可复用的数据访问方法

**示例：**
```go
// 获取用户
user, err := dao.User.GetByID(ctx, 1)

// 创建订单
orderID, err := dao.Order.Create(ctx, g.Map{
    "user_id": 1,
    "symbol": "BTC/USDT",
    "side": "buy",
})

// 分页查询
users, total, err := dao.User.List(ctx, page, limit)
```

### 3. Service 层（业务逻辑）
封装业务逻辑，协调多个 DAO 操作，处理事务。

**特点：**
- 包含业务规则和验证
- 协调多个 DAO 操作
- 处理事务
- 可被多个 Controller 调用

**示例：**
```go
// 用户注册（包含密码加密、生成Token）
token, err := service.Auth.Register(ctx, "username", "password", "email")

// 充值（包含余额更新、流水记录）
err := service.Account.Deposit(ctx, userID, "USDT", "1000")

// 下单（包含订单创建、撮合引擎调用）
orderID, err := service.Order.PlaceOrder(ctx, userID, "BTC/USDT", "buy", "limit", "45000", "0.1")
```

### 4. Controller 层（接口层）
处理 HTTP 请求，调用 Service，返回响应。

**特点：**
- 参数验证（通过 GoFrame 标签）
- 调用 Service 层
- 返回统一格式响应

## 📝 使用示例

### 示例 1：用户注册流程

**Controller 层：**
```go
func (c *ControllerV1) Register(ctx context.Context, req *RegisterReq) (res *RegisterRes, error) {
    // 调用 Service 层
    token, err := service.Auth.Register(ctx, req.Username, req.Password, req.Email)
    if err != nil {
        return nil, err
    }
    
    return &RegisterRes{Token: token}, nil
}
```

**Service 层：**
```go
func (s *authImpl) Register(ctx context.Context, username, password, email string) (string, error) {
    // 1. 检查用户名是否存在（调用 DAO）
    exists, err := dao.User.Exists(ctx, username)
    if exists {
        return "", gerror.New("用户名已存在")
    }
    
    // 2. 密码加密（业务逻辑）
    hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    
    // 3. 创建用户（调用 DAO）
    uid, err := dao.User.Create(ctx, g.Map{
        "username": username,
        "password": string(hashedPassword),
        "email": email,
    })
    
    // 4. 生成 Token（业务逻辑）
    token, err := s.GenerateToken(ctx, uid)
    
    return token, nil
}
```

**DAO 层：**
```go
func (d *UserDao) Create(ctx context.Context, data g.Map) (uint64, error) {
    result, err := d.Model(ctx).Data(data).Insert()
    if err != nil {
        return 0, err
    }
    id, _ := result.LastInsertId()
    return uint64(id), nil
}
```

### 示例 2：充值流程（带事务）

**Service 层：**
```go
func (s *accountImpl) Deposit(ctx context.Context, userID uint64, asset, amount string) error {
    // 使用事务
    return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
        // 1. 查询或创建账户
        account, err := dao.Account.GetByUserIDAndAsset(ctx, userID, asset)
        
        // 2. 更新余额
        newBalance := currentBalance + amountFloat
        err = dao.Account.UpdateBalance(ctx, tx, account.ID, newBalance)
        
        // 3. 创建流水记录
        _, err = dao.AccountFlow.CreateWithTx(ctx, tx, flowData)
        
        return err
    })
}
```

## 🔧 DAO 常用方法

### User DAO
```go
dao.User.GetByID(ctx, id)              // 根据ID获取
dao.User.GetByUsername(ctx, username)  // 根据用户名获取
dao.User.Create(ctx, data)             // 创建用户
dao.User.List(ctx, page, limit)        // 分页列表
dao.User.Exists(ctx, username)         // 检查是否存在
```

### Order DAO
```go
dao.Order.Create(ctx, data)                    // 创建订单
dao.Order.GetByID(ctx, id)                     // 获取订单
dao.Order.ListByUserID(ctx, userID)            // 用户订单列表
dao.Order.List(ctx, page, limit)               // 分页列表
dao.Order.Update(ctx, id, data)                // 更新订单
dao.Order.GetOpenOrders(ctx, symbol, side, price) // 获取未成交订单
```

### Account DAO
```go
dao.Account.GetByUserIDAndAsset(ctx, userID, asset) // 获取账户
dao.Account.ListByUserID(ctx, userID)               // 用户所有账户
dao.Account.Create(ctx, data)                       // 创建账户
dao.Account.Update(ctx, id, data)                   // 更新账户
dao.Account.UpdateBalance(ctx, tx, id, balance)     // 更新余额（事务内）
```

## 💡 最佳实践

### 1. Controller 只做参数验证和调用 Service
```go
// ✅ 正确
func (c *ControllerV1) PlaceOrder(ctx context.Context, req *PlaceOrderReq) (*PlaceOrderRes, error) {
    orderID, err := service.Order.PlaceOrder(ctx, uid, req.Symbol, req.Side, req.Type, req.Price, req.Amount)
    return &PlaceOrderRes{OrderID: orderID}, err
}

// ❌ 错误：Controller 直接操作 DAO
func (c *ControllerV1) PlaceOrder(ctx context.Context, req *PlaceOrderReq) (*PlaceOrderRes, error) {
    orderID, err := dao.Order.Create(ctx, data) // 不应该在 Controller 直接调用 DAO
    return &PlaceOrderRes{OrderID: orderID}, err
}
```

### 2. Service 处理业务逻辑和事务
```go
// ✅ 正确：Service 处理事务
func (s *accountImpl) Deposit(ctx context.Context, userID uint64, asset, amount string) error {
    return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
        // 多个 DAO 操作在一个事务中
        dao.Account.UpdateBalance(ctx, tx, id, balance)
        dao.AccountFlow.CreateWithTx(ctx, tx, data)
        return nil
    })
}
```

### 3. DAO 只做数据库操作
```go
// ✅ 正确：DAO 只负责数据访问
func (d *UserDao) GetByID(ctx context.Context, id uint64) (*model.User, error) {
    var user model.User
    err := d.Model(ctx).Where("id", id).Scan(&user)
    return &user, err
}

// ❌ 错误：DAO 包含业务逻辑
func (d *UserDao) GetByID(ctx context.Context, id uint64) (*model.User, error) {
    var user model.User
    err := d.Model(ctx).Where("id", id).Scan(&user)
    
    // 不应该在 DAO 中验证密码
    if bcrypt.CompareHashAndPassword(...) != nil {
        return nil, gerror.New("密码错误")
    }
    
    return &user, err
}
```

## 🎓 总结

| 层级 | 职责 | 示例 |
|------|------|------|
| **Controller** | 接收请求、参数验证、调用 Service | 处理 HTTP 请求 |
| **Service** | 业务逻辑、事务处理、协调 DAO | 注册用户、充值、下单 |
| **DAO** | 数据库操作、CRUD | 查询用户、创建订单 |
| **Model** | 数据结构定义 | User, Order, Account |

这种分层架构的优势：
- ✅ **职责清晰**：每层只做自己的事
- ✅ **易于测试**：可以单独测试每一层
- ✅ **代码复用**：Service 可被多个 Controller 调用
- ✅ **易于维护**：修改业务逻辑只需改 Service
- ✅ **事务管理**：Service 层统一处理事务
