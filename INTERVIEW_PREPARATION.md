# 面试准备文档 - 加密货币交易所项目

> 基于你的 CEX Exchange 项目，全面准备 Go 后端面试

## 📋 目录

1. [项目介绍话术](#项目介绍话术)
2. [技术栈深度问答](#技术栈深度问答)
3. [系统设计题](#系统设计题)
4. [Go 语言核心知识](#go-语言核心知识)
5. [数据库与缓存](#数据库与缓存)
6. [分布式系统](#分布式系统)
7. [算法与数据结构](#算法与数据结构)
8. [行为面试](#行为面试)

---

## 🎯 项目介绍话术

### 1 分钟版本（电梯演讲）

> "我开发了一个**加密货币交易所系统**，采用前后端分离架构。后端使用 **Go + GoFrame** 框架，实现了用户认证、订单管理、资金管理和实时撮合引擎。技术栈包括 **MySQL** 做持久化、**Redis** 做缓存、**gRPC** 做微服务通信。
> 
> 项目的核心亮点是**撮合引擎**，使用价格-时间优先算法，通过 Go 的 goroutine 实现异步撮合，保证高并发下的性能。同时使用**数据库事务 + 行锁**保证资金安全，防止超卖。
> 
> 前端用 **Vue 3** 实现了交易界面、订单簿、K 线图等功能，通过 WebSocket 实现实时数据推送。"

### 3 分钟版本（详细介绍）

**背景**：
- 这是一个模拟真实交易所的全栈项目，支持限价单、市价单交易
- 目标是学习高并发系统设计和金融交易系统的核心逻辑

**技术架构**：
```
前端: Vue 3 + Element Plus + ECharts
后端: Go + GoFrame + gRPC
数据库: MySQL 8.0
缓存: Redis 7
消息队列: (可扩展 Kafka)
```

**核心功能**：
1. **用户系统**: JWT 认证、bcrypt 密码加密
2. **订单系统**: 下单、撤单、订单查询，支持限价单和市价单
3. **撮合引擎**: 价格-时间优先算法，内存撮合 + 异步持久化
4. **资金系统**: 充值、提现、余额查询，事务保证一致性
5. **市场数据**: 实时行情、K 线、成交记录
6. **gRPC 服务**: 完整的 protobuf 定义，支持流式推送

**技术难点**：
1. **并发控制**: 使用数据库行锁 + Redis 预检查，防止余额不足
2. **撮合性能**: 内存撮合引擎，使用 Go 的 channel 和 goroutine
3. **数据一致性**: 事务处理，确保订单和资金的强一致性
4. **缓存策略**: 订单缓存、余额缓存，缓存失效机制

**项目成果**：
- 代码量: 5000+ 行 Go 代码
- 接口数: 20+ RESTful API + 14 gRPC 方法
- 性能: 单机支持 1000+ TPS（可通过分布式扩展）

---

## 💡 技术栈深度问答

### Go 语言相关

#### Q1: 为什么选择 Go 语言？

**回答**：
1. **高并发**: goroutine 轻量级，适合交易所的高并发场景
2. **性能**: 编译型语言，性能接近 C/C++
3. **简洁**: 语法简单，开发效率高
4. **生态**: gRPC、微服务支持好
5. **部署**: 单一二进制文件，部署简单

**项目中的体现**：
```go
// 撮合引擎使用 goroutine 异步处理
go func() {
    matchEngine := engine.GetEngine()
    matchEngine.Match(ctx, order)
}()
```

#### Q2: 你的项目中如何处理并发？

**回答**：
1. **数据库层面**: 使用行锁（`FOR UPDATE`）防止并发修改
2. **缓存层面**: Redis 预检查，减轻数据库压力
3. **应用层面**: 事务保证原子性

**代码示例**：
```go
// 下单时的并发控制
err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
    // 1. 查询账户并加锁
    err := tx.Model("accounts").
        Where("user_id", uid).
        Where("asset", freezeAsset).
        LockUpdate().  // FOR UPDATE 排他锁
        Scan(&account)
    
    // 2. 检查余额
    if availableBalance.LessThan(freezeAmount) {
        return gerror.New("余额不足")
    }
    
    // 3. 扣减余额（原子操作）
    _, err = tx.Model("accounts").
        Where("id", account.ID).
        Data(g.Map{
            "balance": gdb.Raw("balance - " + freezeAmount),
            "frozen":  gdb.Raw("frozen + " + freezeAmount),
        }).Update()
    
    return nil
})
```

#### Q3: goroutine 泄漏如何避免？

**回答**：
1. **使用 context 控制生命周期**
2. **确保 channel 关闭**
3. **使用 sync.WaitGroup 等待完成**

**项目中的实践**：
```go
// gRPC 流式推送，使用 context 控制
func (s *ExchangeServer) SubscribeOrderBook(req *pb.SubscribeOrderBookRequest, 
    stream pb.ExchangeService_SubscribeOrderBookServer) error {
    
    ctx := stream.Context()
    
    go func() {
        ticker := time.NewTicker(1 * time.Second)
        defer ticker.Stop()  // 确保 ticker 停止
        
        for {
            select {
            case <-ctx.Done():  // context 取消时退出
                return
            case <-ticker.C:
                // 推送数据
            }
        }
    }()
}
```

### 数据库相关

#### Q4: 如何保证订单系统的数据一致性？

**回答**：
使用**数据库事务 + 行锁**：

1. **事务隔离**: 使用 `REPEATABLE READ` 隔离级别
2. **行锁**: `SELECT ... FOR UPDATE` 锁定账户记录
3. **原子操作**: 余额扣减使用 SQL 表达式，避免读-改-写竞态

**完整流程**：
```
开始事务
  ├─ 查询账户并加锁 (FOR UPDATE)
  ├─ 检查余额是否充足
  ├─ 扣减余额 + 增加冻结 (原子操作)
  ├─ 插入订单记录
  └─ 提交事务
```

#### Q5: 数据库索引如何设计？

**回答**：
```sql
-- 用户表
CREATE INDEX idx_username ON users(username);

-- 订单表
CREATE INDEX idx_user_symbol ON orders(user_id, symbol);
CREATE INDEX idx_status ON orders(status);
CREATE INDEX idx_created_at ON orders(created_at);

-- 账户表
CREATE UNIQUE INDEX idx_user_asset ON accounts(user_id, asset);

-- 成交表
CREATE INDEX idx_symbol_time ON trades(symbol, created_at);
```

**设计原则**：
1. **高频查询字段**: user_id, symbol, status
2. **唯一约束**: user_id + asset（防止重复账户）
3. **范围查询**: created_at（时间范围查询）
4. **避免过度索引**: 影响写入性能

### Redis 相关

#### Q6: Redis 在项目中的使用场景？

**回答**：
1. **订单缓存**: 用户订单列表，减少数据库查询
2. **余额缓存**: 快速预检查，避免无效请求
3. **JWT Token**: 存储用户会话（可选）
4. **行情数据**: 实时价格、订单簿（未来扩展）

**代码示例**：
```go
// 订单缓存
func (s *orderImpl) ListMyOrders(ctx context.Context, uid uint64, symbol string) ([]model.Order, error) {
    cacheKey := fmt.Sprintf("user:orders:%d:%s", uid, symbol)
    
    // 1. 尝试从缓存获取
    cachedData, err := Redis.Get(ctx, cacheKey)
    if err == nil && cachedData != "" {
        var orders []model.Order
        json.Unmarshal([]byte(cachedData), &orders)
        return orders, nil
    }
    
    // 2. 缓存未命中，查询数据库
    orders, err := dao.Order.ListByUserID(ctx, uid, symbol)
    
    // 3. 写入缓存
    data, _ := json.Marshal(orders)
    Redis.SetEX(ctx, cacheKey, string(data), 20) // 20秒过期
    
    return orders, nil
}
```

#### Q7: 如何保证缓存和数据库的一致性？

**回答**：
使用 **Cache Aside 模式**：

1. **读取**: 先读缓存，未命中再读数据库，然后写入缓存
2. **更新**: 先更新数据库，再删除缓存（而不是更新缓存）

**为什么删除而不是更新？**
- 避免并发更新导致的数据不一致
- 懒加载，下次读取时再写入最新数据

**代码示例**：
```go
// 下单后清除缓存
func (s *orderImpl) PlaceOrder(ctx context.Context, req *model.PlaceOrderReq) (int64, error) {
    // 1. 数据库操作
    orderID, err := /* 插入订单 */
    
    // 2. 清除缓存
    s.clearUserOrderCache(ctx, uid, req.Symbol)
    
    return orderID, err
}

func (s *orderImpl) clearUserOrderCache(ctx context.Context, uid uint64, symbol string) {
    Redis.Del(ctx, fmt.Sprintf("user:orders:%d:", uid))
    Redis.Del(ctx, fmt.Sprintf("user:orders:%d:%s", uid, symbol))
}
```

#### Q8: 缓存穿透、击穿、雪崩如何解决？

**回答**：

**1. 缓存穿透**（查询不存在的数据）
- **问题**: 恶意查询不存在的 key，每次都打到数据库
- **解决**: 
  - 布隆过滤器（Bloom Filter）
  - 缓存空值，设置短过期时间

**2. 缓存击穿**（热点 key 过期）
- **问题**: 热点 key 过期瞬间，大量请求打到数据库
- **解决**:
  - 互斥锁（singleflight）
  - 热点数据永不过期

**3. 缓存雪崩**（大量 key 同时过期）
- **问题**: 大量 key 同时过期，数据库压力骤增
- **解决**:
  - 过期时间加随机值
  - 多级缓存
  - 限流降级

**代码示例**（互斥锁防击穿）：
```go
import "golang.org/x/sync/singleflight"

var sg singleflight.Group

func GetOrderBook(symbol string) (*OrderBook, error) {
    key := "orderbook:" + symbol
    
    // singleflight 确保同一时间只有一个请求查询数据库
    v, err, _ := sg.Do(key, func() (interface{}, error) {
        // 查询缓存
        cached := redis.Get(key)
        if cached != nil {
            return cached, nil
        }
        
        // 查询数据库
        data := db.Query(symbol)
        
        // 写入缓存
        redis.Set(key, data, time.Minute)
        
        return data, nil
    })
    
    return v.(*OrderBook), err
}
```

### gRPC 相关

#### Q9: 为什么使用 gRPC？

**回答**：
1. **性能**: HTTP/2 + Protobuf 二进制序列化，比 JSON 快
2. **类型安全**: Protobuf 强类型，编译时检查
3. **跨语言**: 一份 .proto 生成多种语言代码
4. **流式传输**: 支持服务端流、客户端流、双向流
5. **生态**: Google 支持，与 K8s、Istio 集成好

**性能对比**：
```
JSON (REST):  100 MB/s
Protobuf:     500 MB/s  (5x faster)
```

#### Q10: gRPC 和 REST 如何选择？

**回答**：

| 场景 | 选择 | 原因 |
|------|------|------|
| 微服务内部通信 | gRPC | 性能高、类型安全 |
| 对外 API | REST | 浏览器支持、易调试 |
| 实时推送 | gRPC | 流式传输 |
| 移动端 | gRPC | 省流量、省电 |

**项目中的实践**：
- **对外**: REST API（前端调用）
- **内部**: gRPC（未来微服务拆分）
- **实时**: gRPC 流式推送订单簿

---

## 🏗️ 系统设计题

### 设计题 1: 如何设计一个高并发的交易所系统？

**回答框架**（SNAKE 原则）：
1. **Scenario（场景）**: 明确需求
2. **Necessary（必要）**: 核心功能
3. **Application（应用）**: 架构设计
4. **Kilobit（数据）**: 数据量估算
5. **Evolve（演进）**: 优化方案

#### 1. Scenario - 场景分析

**功能需求**：
- 用户注册、登录
- 下单、撤单、查询订单
- 充值、提现、查询余额
- 实时行情、K 线、成交记录

**非功能需求**：
- **高并发**: 10万 QPS
- **低延迟**: 下单 < 100ms
- **高可用**: 99.99% 可用性
- **数据一致性**: 强一致性（资金安全）

#### 2. Necessary - 核心功能

**最小可行产品（MVP）**：
1. 用户认证
2. 下单/撤单
3. 撮合引擎
4. 资金管理

#### 3. Application - 架构设计

```
┌─────────────────────────────────────────────┐
│              负载均衡 (Nginx/LVS)            │
└──────────────┬──────────────────────────────┘
               │
    ┌──────────┴──────────┐
    │                     │
┌───▼────┐           ┌───▼────┐
│ API    │           │ API    │
│ Gateway│           │ Gateway│
└───┬────┘           └───┬────┘
    │                     │
    └──────────┬──────────┘
               │
    ┌──────────┴──────────────────────────┐
    │                                     │
┌───▼────────┐  ┌──────────┐  ┌─────────▼──┐
│ 订单服务    │  │ 用户服务  │  │ 资金服务    │
│ (gRPC)     │  │ (gRPC)   │  │ (gRPC)     │
└───┬────────┘  └────┬─────┘  └─────┬──────┘
    │                │                │
    └────────────┬───┴────────────────┘
                 │
    ┌────────────┴────────────┐
    │                         │
┌───▼────┐              ┌────▼─────┐
│ MySQL  │              │  Redis   │
│ (主从)  │              │ (集群)   │
└────────┘              └──────────┘
```

**核心组件**：

1. **API Gateway**: 
   - 限流、熔断
   - 认证、鉴权
   - 路由转发

2. **订单服务**:
   - 下单、撤单
   - 订单查询
   - 调用撮合引擎

3. **撮合引擎**:
   - 内存撮合（价格-时间优先）
   - 异步持久化
   - 高性能（单机 10万 TPS）

4. **资金服务**:
   - 充值、提现
   - 余额查询
   - 资金流水

5. **数据层**:
   - MySQL: 持久化存储
   - Redis: 缓存 + 消息队列

#### 4. Kilobit - 数据量估算

**假设**：
- 日活用户: 100万
- 每用户日均订单: 10 笔
- 日订单量: 1000万

**存储估算**：
```
订单表:
  - 每条记录: 200 bytes
  - 日增量: 1000万 * 200B = 2GB
  - 年增量: 2GB * 365 = 730GB

成交表:
  - 撮合率: 50%
  - 日增量: 500万 * 200B = 1GB
  - 年增量: 365GB
```

**QPS 估算**：
```
峰值 QPS = 日订单量 / (24 * 3600) * 峰值倍数
        = 10,000,000 / 86400 * 10
        = 1157 QPS

撮合引擎 TPS = 1157 * 2 (买卖双方) = 2314 TPS
```

#### 5. Evolve - 优化方案

**性能优化**：

1. **数据库优化**
   - **分库分表**: 按交易对分表，按日期分表
   - **读写分离**: 主库写，从库读
   - **索引优化**: 覆盖索引、联合索引

```sql
-- 订单表分表（按交易对）
CREATE TABLE orders_btc_usdt (...);
CREATE TABLE orders_eth_usdt (...);

-- 按日期分表
CREATE TABLE orders_btc_usdt_20231201 (...);
CREATE TABLE orders_btc_usdt_20231202 (...);
```

2. **缓存优化**
   - **多级缓存**: 本地缓存 + Redis
   - **缓存预热**: 启动时加载热点数据
   - **缓存更新**: 订阅 binlog 更新缓存

3. **撮合引擎优化**
   - **内存撮合**: 订单簿放内存，跳表实现
   - **批量持久化**: 异步批量写入数据库
   - **分布式撮合**: 按交易对分片

```go
// 跳表实现订单簿
type OrderBook struct {
    bids *skiplist.SkipList  // 买单（降序）
    asks *skiplist.SkipList  // 卖单（升序）
}

// 撮合逻辑
func (ob *OrderBook) Match(order *Order) []*Trade {
    if order.Side == "buy" {
        // 买单与卖单撮合
        return ob.matchWithAsks(order)
    } else {
        // 卖单与买单撮合
        return ob.matchWithBids(order)
    }
}
```

4. **消息队列**
   - **异步处理**: 订单入队，异步撮合
   - **削峰填谷**: 应对流量突增
   - **解耦服务**: 订单服务 → MQ → 撮合引擎

```
下单请求 → API → 订单服务 → Kafka → 撮合引擎 → 成交通知
```

5. **限流降级**
   - **令牌桶**: 限制 QPS
   - **熔断**: 下游服务异常时熔断
   - **降级**: 非核心功能降级

```go
// 令牌桶限流
import "golang.org/x/time/rate"

limiter := rate.NewLimiter(1000, 2000)  // 1000 QPS, 桶容量 2000

func PlaceOrder(order *Order) error {
    if !limiter.Allow() {
        return errors.New("too many requests")
    }
    // 处理订单
}
```

**高可用方案**：

1. **服务高可用**
   - **多副本**: 每个服务 3+ 实例
   - **健康检查**: K8s liveness/readiness
   - **自动重启**: 服务异常自动重启

2. **数据库高可用**
   - **主从复制**: 1 主 2 从
   - **自动故障转移**: MHA/Orchestrator
   - **备份**: 每日全量 + 实时 binlog

3. **Redis 高可用**
   - **哨兵模式**: 主从 + 哨兵
   - **集群模式**: 分片 + 副本
   - **持久化**: AOF + RDB

**监控告警**：

```
监控指标:
  - QPS、延迟、错误率
  - CPU、内存、磁盘
  - 数据库连接数、慢查询
  - Redis 命中率、内存使用

告警:
  - 错误率 > 1%
  - 延迟 > 200ms
  - CPU > 80%
  - 磁盘 > 85%
```

### 设计题 2: 如何防止超卖？

**回答**：

**方案 1: 数据库行锁**
```go
// 悲观锁
db.Transaction(func(tx *gorm.DB) error {
    var account Account
    tx.Clauses(clause.Locking{Strength: "UPDATE"}).
        Where("user_id = ?", uid).
        First(&account)
    
    if account.Balance < amount {
        return errors.New("余额不足")
    }
    
    tx.Model(&account).Update("balance", gorm.Expr("balance - ?", amount))
    return nil
})
```

**方案 2: 乐观锁（版本号）**
```go
// 使用版本号
type Account struct {
    ID      uint64
    Balance decimal.Decimal
    Version int  // 版本号
}

// 更新时检查版本号
result := db.Model(&Account{}).
    Where("id = ? AND version = ?", id, oldVersion).
    Updates(map[string]interface{}{
        "balance": newBalance,
        "version": oldVersion + 1,
    })

if result.RowsAffected == 0 {
    return errors.New("并发冲突，请重试")
}
```

**方案 3: Redis 原子操作**
```go
// Lua 脚本保证原子性
script := `
local balance = redis.call('GET', KEYS[1])
if tonumber(balance) >= tonumber(ARGV[1]) then
    redis.call('DECRBY', KEYS[1], ARGV[1])
    return 1
else
    return 0
end
`

result := redis.Eval(script, []string{key}, amount)
if result == 0 {
    return errors.New("余额不足")
}
```

**方案对比**：

| 方案 | 优点 | 缺点 | 适用场景 |
|------|------|------|----------|
| 数据库行锁 | 强一致性 | 性能较低 | 金融系统 |
| 乐观锁 | 性能高 | 冲突时需重试 | 冲突少的场景 |
| Redis 原子操作 | 性能最高 | 需保证 Redis 可靠性 | 高并发场景 |

**项目中的选择**: 数据库行锁 + Redis 预检查
- Redis 快速过滤无效请求
- 数据库行锁保证最终一致性

---

## 🔧 Go 语言核心知识

### 1. Goroutine 和 Channel

#### Q: Goroutine 的调度模型？

**回答**: GMP 模型

- **G (Goroutine)**: 用户态线程
- **M (Machine)**: 内核线程
- **P (Processor)**: 逻辑处理器，维护 G 的队列

```
┌─────────────────────────────────┐
│         全局队列 (Global Queue)  │
└─────────────────────────────────┘
              │
    ┌─────────┼─────────┐
    │         │         │
┌───▼───┐ ┌──▼────┐ ┌──▼────┐
│  P1   │ │  P2   │ │  P3   │
│ ┌───┐ │ │ ┌───┐ │ │ ┌───┐ │
│ │ G │ │ │ │ G │ │ │ │ G │ │
│ └───┘ │ │ └───┘ │ │ └───┘ │
└───┬───┘ └───┬───┘ └───┬───┘
    │         │         │
┌───▼───┐ ┌──▼────┐ ┌──▼────┐
│  M1   │ │  M2   │ │  M3   │
└───────┘ └───────┘ └───────┘
```

**调度策略**：
1. **Work Stealing**: P 的本地队列空时，从其他 P 偷取 G
2. **Hand Off**: M 阻塞时，将 P 转移给其他 M
3. **抢占式调度**: 防止 G 长时间占用 CPU

#### Q: Channel 的底层实现？

**回答**: 

```go
type hchan struct {
    qcount   uint           // 队列中的元素个数
    dataqsiz uint           // 环形队列的大小
    buf      unsafe.Pointer // 环形队列指针
    elemsize uint16         // 元素大小
    closed   uint32         // 是否关闭
    sendx    uint           // 发送索引
    recvx    uint           // 接收索引
    recvq    waitq          // 接收等待队列
    sendq    waitq          // 发送等待队列
    lock     mutex          // 互斥锁
}
```

**发送流程**：
1. 加锁
2. 如果有接收者在等待，直接发送
3. 如果缓冲区未满，放入缓冲区
4. 否则，发送者进入等待队列
5. 解锁

**关闭 Channel**：
- 关闭后不能再发送，否则 panic
- 可以继续接收，直到缓冲区为空
- 接收空 channel 返回零值

```go
ch := make(chan int, 3)
ch <- 1
ch <- 2
close(ch)

// 可以继续接收
v1 := <-ch  // 1
v2 := <-ch  // 2
v3 := <-ch  // 0 (零值)

// 判断 channel 是否关闭
v, ok := <-ch
if !ok {
    fmt.Println("channel closed")
}
```

### 2. 内存管理和 GC

#### Q: Go 的内存分配策略？

**回答**: TCMalloc 思想

**三级分配**：
1. **微对象** (< 16B): 使用 mcache 的 tiny 分配器
2. **小对象** (16B - 32KB): 使用 mcache 的 mspan
3. **大对象** (> 32KB): 直接从 mheap 分配

```
┌──────────────────────────────────┐
│           mheap (堆)              │
│  ┌────────────────────────────┐  │
│  │      mcentral (中心缓存)    │  │
│  └────────────────────────────┘  │
└──────────────────────────────────┘
              │
    ┌─────────┼─────────┐
    │         │         │
┌───▼───┐ ┌──▼────┐ ┌──▼────┐
│mcache │ │mcache │ │mcache │
│  P1   │ │  P2   │ │  P3   │
└───────┘ └───────┘ └───────┘
```

**优点**：
- 减少锁竞争（每个 P 有自己的 mcache）
- 减少内存碎片
- 提高分配效率

#### Q: Go 的 GC 算法？

**回答**: 三色标记法 + 并发 GC

**三色标记**：
- **白色**: 未被访问的对象（待回收）
- **灰色**: 已被访问，但其引用的对象未被访问
- **黑色**: 已被访问，且其引用的对象也已被访问

**GC 流程**：
1. **Mark Setup**: STW，启动写屏障
2. **Marking**: 并发标记
3. **Mark Termination**: STW，完成标记
4. **Sweeping**: 并发清扫

**写屏障**: 防止并发标记时漏标

```go
// 写屏障伪代码
func writePointer(slot *unsafe.Pointer, ptr unsafe.Pointer) {
    shade(ptr)  // 标记为灰色
    *slot = ptr
}
```

**GC 触发条件**：
1. 内存增长达到阈值（默认 100%）
2. 定时触发（2 分钟）
3. 手动触发 `runtime.GC()`

**优化建议**：
```go
// 1. 减少内存分配
var bufPool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 4096)
    },
}

buf := bufPool.Get().([]byte)
defer bufPool.Put(buf)

// 2. 使用指针接收者
type User struct {
    Name string
    Age  int
}

func (u *User) SetName(name string) {  // 指针接收者
    u.Name = name
}

// 3. 预分配切片容量
users := make([]User, 0, 1000)  // 预分配容量
```

### 3. 并发模式

#### Q: 常见的并发模式？

**1. Worker Pool**
```go
func workerPool(jobs <-chan Job, results chan<- Result) {
    const numWorkers = 10
    
    var wg sync.WaitGroup
    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for job := range jobs {
                result := process(job)
                results <- result
            }
        }()
    }
    
    wg.Wait()
    close(results)
}
```

**2. Pipeline**
```go
func pipeline() {
    // 生成器
    gen := func(nums ...int) <-chan int {
        out := make(chan int)
        go func() {
            for _, n := range nums {
                out <- n
            }
            close(out)
        }()
        return out
    }
    
    // 平方
    sq := func(in <-chan int) <-chan int {
        out := make(chan int)
        go func() {
            for n := range in {
                out <- n * n
            }
            close(out)
        }()
        return out
    }
    
    // 使用
    for n := range sq(gen(1, 2, 3, 4)) {
        fmt.Println(n)  // 1, 4, 9, 16
    }
}
```

**3. Fan-Out/Fan-In**
```go
func fanOut(in <-chan int, n int) []<-chan int {
    outs := make([]<-chan int, n)
    for i := 0; i < n; i++ {
        outs[i] = worker(in)
    }
    return outs
}

func fanIn(ins ...<-chan int) <-chan int {
    out := make(chan int)
    var wg sync.WaitGroup
    
    for _, in := range ins {
        wg.Add(1)
        go func(ch <-chan int) {
            defer wg.Done()
            for v := range ch {
                out <- v
            }
        }(in)
    }
    
    go func() {
        wg.Wait()
        close(out)
    }()
    
    return out
}
```

---

## 💾 数据库与缓存

### MySQL 优化

#### Q: 慢查询如何优化？

**排查步骤**：
1. **开启慢查询日志**
```sql
SET GLOBAL slow_query_log = 'ON';
SET GLOBAL long_query_time = 1;  -- 超过1秒的查询
```

2. **使用 EXPLAIN 分析**
```sql
EXPLAIN SELECT * FROM orders WHERE user_id = 123 AND status = 'open';
```

**关键指标**：
- `type`: 访问类型（ALL < index < range < ref < const）
- `key`: 使用的索引
- `rows`: 扫描的行数
- `Extra`: 额外信息（Using filesort, Using temporary 需优化）

**优化方案**：

1. **添加索引**
```sql
-- 联合索引（最左前缀原则）
CREATE INDEX idx_user_status ON orders(user_id, status);
```

2. **避免 SELECT ***
```sql
-- 不好
SELECT * FROM orders WHERE id = 1;

-- 好
SELECT id, user_id, symbol, price FROM orders WHERE id = 1;
```

3. **分页优化**
```sql
-- 不好（深分页）
SELECT * FROM orders ORDER BY id LIMIT 1000000, 10;

-- 好（使用子查询）
SELECT * FROM orders 
WHERE id > (SELECT id FROM orders ORDER BY id LIMIT 1000000, 1)
ORDER BY id LIMIT 10;
```

4. **避免函数操作索引列**
```sql
-- 不好（索引失效）
SELECT * FROM orders WHERE DATE(created_at) = '2023-12-01';

-- 好
SELECT * FROM orders 
WHERE created_at >= '2023-12-01 00:00:00' 
  AND created_at < '2023-12-02 00:00:00';
```

#### Q: 事务隔离级别？

**四种隔离级别**：

| 隔离级别 | 脏读 | 不可重复读 | 幻读 |
|---------|------|-----------|------|
| READ UNCOMMITTED | ✓ | ✓ | ✓ |
| READ COMMITTED | ✗ | ✓ | ✓ |
| REPEATABLE READ | ✗ | ✗ | ✓ |
| SERIALIZABLE | ✗ | ✗ | ✗ |

**MySQL 默认**: REPEATABLE READ

**MVCC 实现**：
- 每行记录有隐藏列：事务 ID、回滚指针
- 读取时根据事务 ID 判断可见性
- 通过 undo log 构建历史版本

```sql
-- 查看隔离级别
SELECT @@transaction_isolation;

-- 设置隔离级别
SET SESSION TRANSACTION ISOLATION LEVEL READ COMMITTED;
```

**项目中的选择**: REPEATABLE READ
- 防止不可重复读
- 配合行锁防止幻读

### Redis 深入

#### Q: Redis 的数据结构？

**1. String**
```bash
SET key value
GET key
INCR counter  # 原子递增
```

**应用**: 缓存、计数器、分布式锁

**2. Hash**
```bash
HSET user:1 name "Alice"
HSET user:1 age 25
HGETALL user:1
```

**应用**: 对象缓存

**3. List**
```bash
LPUSH queue task1
RPOP queue
```

**应用**: 消息队列、最新列表

**4. Set**
```bash
SADD tags:1 "go" "redis"
SISMEMBER tags:1 "go"
```

**应用**: 标签、好友关系

**5. Sorted Set**
```bash
ZADD leaderboard 100 "user1"
ZADD leaderboard 200 "user2"
ZRANGE leaderboard 0 -1 WITHSCORES
```

**应用**: 排行榜、延时队列

**底层实现**：
- **String**: SDS（Simple Dynamic String）
- **Hash**: 哈希表 / ziplist
- **List**: quicklist（linkedlist + ziplist）
- **Set**: 哈希表 / intset
- **Sorted Set**: skiplist + 哈希表

#### Q: Redis 持久化？

**1. RDB（快照）**
```bash
# 配置
save 900 1      # 900秒内至少1个key变化
save 300 10     # 300秒内至少10个key变化
save 60 10000   # 60秒内至少10000个key变化
```

**优点**: 文件小，恢复快
**缺点**: 可能丢失最后一次快照后的数据

**2. AOF（追加日志）**
```bash
# 配置
appendonly yes
appendfsync everysec  # 每秒同步
```

**优点**: 数据安全性高
**缺点**: 文件大，恢复慢

**AOF 重写**：
```bash
# 手动触发
BGREWRITEAOF

# 自动触发
auto-aof-rewrite-percentage 100
auto-aof-rewrite-min-size 64mb
```

**混合持久化**（Redis 4.0+）：
- RDB 快照 + AOF 增量
- 兼顾性能和安全

**项目中的选择**: AOF + 每秒同步
- 数据重要，不能丢失
- 每秒同步，性能和安全平衡

---

## 🌐 分布式系统

### 分布式锁

#### Q: 如何实现分布式锁？

**方案 1: Redis SETNX**
```go
// 加锁
func Lock(key string, value string, expiration time.Duration) bool {
    return redis.SetNX(key, value, expiration).Val()
}

// 解锁（Lua 脚本保证原子性）
func Unlock(key string, value string) bool {
    script := `
    if redis.call("get", KEYS[1]) == ARGV[1] then
        return redis.call("del", KEYS[1])
    else
        return 0
    end
    `
    return redis.Eval(script, []string{key}, value).Val() == 1
}

// 使用
lockKey := "lock:order:123"
lockValue := uuid.New().String()

if Lock(lockKey, lockValue, 10*time.Second) {
    defer Unlock(lockKey, lockValue)
    // 业务逻辑
}
```

**问题**: 
- 锁过期时间难以设置
- 主从切换时可能丢失锁

**方案 2: Redlock（多 Redis 实例）**
```go
import "github.com/go-redsync/redsync/v4"

// 创建 Redlock
pools := []redis.Pool{pool1, pool2, pool3, pool4, pool5}
rs := redsync.New(pools)

// 加锁
mutex := rs.NewMutex("lock:order:123",
    redsync.WithExpiry(10*time.Second),
    redsync.WithTries(3),
)

if err := mutex.Lock(); err != nil {
    return err
}
defer mutex.Unlock()

// 业务逻辑
```

**方案 3: etcd / ZooKeeper**
```go
import "go.etcd.io/etcd/client/v3/concurrency"

// 创建 session
session, _ := concurrency.NewSession(client)
defer session.Close()

// 创建锁
mutex := concurrency.NewMutex(session, "/lock/order/123")

// 加锁
if err := mutex.Lock(context.TODO()); err != nil {
    return err
}
defer mutex.Unlock(context.TODO())

// 业务逻辑
```

**对比**：

| 方案 | 优点 | 缺点 | 适用场景 |
|------|------|------|----------|
| Redis SETNX | 简单、性能高 | 可靠性一般 | 允许偶尔失败 |
| Redlock | 可靠性较高 | 复杂、性能较低 | 重要业务 |
| etcd/ZK | 可靠性最高 | 性能最低 | 强一致性要求 |

### 分布式事务

#### Q: 如何保证分布式事务？

**方案 1: 2PC（两阶段提交）**
```
准备阶段:
  协调者 → 参与者: 准备提交
  参与者 → 协调者: 准备完成

提交阶段:
  协调者 → 参与者: 提交
  参与者 → 协调者: 提交完成
```

**缺点**: 
- 同步阻塞
- 单点故障
- 数据不一致（网络分区）

**方案 2: TCC（Try-Confirm-Cancel）**
```go
// Try: 预留资源
func TryDeduct(userID int64, amount decimal.Decimal) error {
    // 冻结余额
    return db.Exec("UPDATE accounts SET frozen = frozen + ? WHERE user_id = ?", 
        amount, userID)
}

// Confirm: 确认提交
func ConfirmDeduct(userID int64, amount decimal.Decimal) error {
    // 扣减冻结金额
    return db.Exec("UPDATE accounts SET balance = balance - ?, frozen = frozen - ? WHERE user_id = ?",
        amount, amount, userID)
}

// Cancel: 取消回滚
func CancelDeduct(userID int64, amount decimal.Decimal) error {
    // 解冻余额
    return db.Exec("UPDATE accounts SET frozen = frozen - ? WHERE user_id = ?",
        amount, userID)
}
```

**方案 3: SAGA（长事务）**
```
正向操作: T1 → T2 → T3 → T4
补偿操作: C4 → C3 → C2 → C1

失败时执行补偿操作
```

**方案 4: 本地消息表**
```go
// 1. 本地事务：业务操作 + 插入消息表
db.Transaction(func(tx *gorm.DB) error {
    // 业务操作
    tx.Create(&order)
    
    // 插入消息表
    tx.Create(&Message{
        Topic: "order.created",
        Body:  orderJSON,
        Status: "pending",
    })
    
    return nil
})

// 2. 定时任务：扫描消息表，发送到 MQ
go func() {
    ticker := time.NewTicker(1 * time.Second)
    for range ticker.C {
        var messages []Message
        db.Where("status = ?", "pending").Find(&messages)
        
        for _, msg := range messages {
            if kafka.Send(msg.Topic, msg.Body) {
                db.Model(&msg).Update("status", "sent")
            }
        }
    }
}()

// 3. 消费者：处理消息（幂等）
func HandleOrderCreated(msg *Message) error {
    // 幂等处理
    if exists(msg.ID) {
        return nil
    }
    
    // 业务逻辑
    processOrder(msg.Body)
    
    // 记录已处理
    markProcessed(msg.ID)
    
    return nil
}
```

**项目中的选择**: 本地消息表 + 最终一致性
- 简单可靠
- 性能好
- 适合大多数场景

---

## 📊 算法与数据结构

### 常见算法题

#### 1. 两数之和
```go
func twoSum(nums []int, target int) []int {
    m := make(map[int]int)
    for i, num := range nums {
        if j, ok := m[target-num]; ok {
            return []int{j, i}
        }
        m[num] = i
    }
    return nil
}
```

#### 2. 反转链表
```go
func reverseList(head *ListNode) *ListNode {
    var prev *ListNode
    curr := head
    
    for curr != nil {
        next := curr.Next
        curr.Next = prev
        prev = curr
        curr = next
    }
    
    return prev
}
```

#### 3. 二叉树层序遍历
```go
func levelOrder(root *TreeNode) [][]int {
    if root == nil {
        return nil
    }
    
    var result [][]int
    queue := []*TreeNode{root}
    
    for len(queue) > 0 {
        size := len(queue)
        level := make([]int, 0, size)
        
        for i := 0; i < size; i++ {
            node := queue[0]
            queue = queue[1:]
            
            level = append(level, node.Val)
            
            if node.Left != nil {
                queue = append(queue, node.Left)
            }
            if node.Right != nil {
                queue = append(queue, node.Right)
            }
        }
        
        result = append(result, level)
    }
    
    return result
}
```

#### 4. LRU 缓存
```go
type LRUCache struct {
    capacity int
    cache    map[int]*Node
    head     *Node
    tail     *Node
}

type Node struct {
    key   int
    value int
    prev  *Node
    next  *Node
}

func Constructor(capacity int) LRUCache {
    head := &Node{}
    tail := &Node{}
    head.next = tail
    tail.prev = head
    
    return LRUCache{
        capacity: capacity,
        cache:    make(map[int]*Node),
        head:     head,
        tail:     tail,
    }
}

func (this *LRUCache) Get(key int) int {
    if node, ok := this.cache[key]; ok {
        this.moveToHead(node)
        return node.value
    }
    return -1
}

func (this *LRUCache) Put(key int, value int) {
    if node, ok := this.cache[key]; ok {
        node.value = value
        this.moveToHead(node)
        return
    }
    
    node := &Node{key: key, value: value}
    this.cache[key] = node
    this.addToHead(node)
    
    if len(this.cache) > this.capacity {
        removed := this.removeTail()
        delete(this.cache, removed.key)
    }
}

func (this *LRUCache) moveToHead(node *Node) {
    this.removeNode(node)
    this.addToHead(node)
}

func (this *LRUCache) removeNode(node *Node) {
    node.prev.next = node.next
    node.next.prev = node.prev
}

func (this *LRUCache) addToHead(node *Node) {
    node.next = this.head.next
    node.prev = this.head
    this.head.next.prev = node
    this.head.next = node
}

func (this *LRUCache) removeTail() *Node {
    node := this.tail.prev
    this.removeNode(node)
    return node
}
```

### 数据结构

#### 跳表（Skip List）
```go
type SkipList struct {
    head  *Node
    level int
}

type Node struct {
    value int
    next  []*Node  // 多层指针
}

const MaxLevel = 16

func NewSkipList() *SkipList {
    return &SkipList{
        head:  &Node{next: make([]*Node, MaxLevel)},
        level: 1,
    }
}

func (sl *SkipList) Insert(value int) {
    update := make([]*Node, MaxLevel)
    curr := sl.head
    
    // 从最高层开始查找插入位置
    for i := sl.level - 1; i >= 0; i-- {
        for curr.next[i] != nil && curr.next[i].value < value {
            curr = curr.next[i]
        }
        update[i] = curr
    }
    
    // 随机层数
    level := randomLevel()
    if level > sl.level {
        for i := sl.level; i < level; i++ {
            update[i] = sl.head
        }
        sl.level = level
    }
    
    // 插入节点
    newNode := &Node{
        value: value,
        next:  make([]*Node, level),
    }
    
    for i := 0; i < level; i++ {
        newNode.next[i] = update[i].next[i]
        update[i].next[i] = newNode
    }
}

func randomLevel() int {
    level := 1
    for rand.Float64() < 0.5 && level < MaxLevel {
        level++
    }
    return level
}
```

**应用**: Redis Sorted Set、订单簿

---

## 🎭 行为面试

### STAR 法则

**S (Situation)**: 情境  
**T (Task)**: 任务  
**A (Action)**: 行动  
**R (Result)**: 结果

### 示例问题

#### Q: 描述一个你解决的技术难题

**回答**（基于你的项目）：

**S**: 在开发交易所系统时，发现高并发下单时，偶尔会出现余额扣减错误，导致用户余额为负数。

**T**: 需要保证在高并发场景下，余额扣减的准确性和一致性，防止超卖。

**A**: 
1. **分析问题**: 通过日志发现是并发读-改-写导致的竞态条件
2. **设计方案**: 
   - 数据库层面：使用 `SELECT ... FOR UPDATE` 行锁
   - 应用层面：使用事务保证原子性
   - 缓存层面：Redis 预检查，快速过滤无效请求
3. **实施优化**:
   ```go
   // 1. Redis 预检查
   cachedBalance := redis.Get("balance:" + uid)
   if cachedBalance < amount {
       return errors.New("余额不足")
   }
   
   // 2. 数据库事务 + 行锁
   db.Transaction(func(tx *gorm.DB) error {
       var account Account
       tx.Clauses(clause.Locking{Strength: "UPDATE"}).
           Where("user_id = ?", uid).
           First(&account)
       
       if account.Balance < amount {
           return errors.New("余额不足")
       }
       
       tx.Model(&account).Update("balance", 
           gorm.Expr("balance - ?", amount))
       
       return nil
   })
   ```
4. **压力测试**: 使用 JMeter 模拟 1000 并发请求

**R**: 
- 彻底解决了余额扣减错误问题
- 性能提升 30%（Redis 预检查减少无效数据库访问）
- 通过了 1000 并发的压力测试

#### Q: 如何与团队协作？

**回答**：

1. **代码规范**: 
   - 统一使用 gofmt 格式化
   - Code Review 机制
   - Git 分支管理（feature/bugfix/hotfix）

2. **文档**:
   - API 文档（Swagger）
   - 架构文档（README）
   - 数据库设计文档

3. **沟通**:
   - 每日站会同步进度
   - 技术方案评审
   - 问题及时沟通

4. **工具**:
   - Git: 版本控制
   - JIRA: 任务管理
   - Confluence: 知识库

---

## 📚 学习资源推荐

### 书籍
1. **Go 语言**:
   - 《Go 程序设计语言》
   - 《Go 语言高级编程》
   - 《Go 语言并发之道》

2. **系统设计**:
   - 《设计数据密集型应用》
   - 《微服务设计》
   - 《高性能 MySQL》

3. **算法**:
   - 《算法导论》
   - 《剑指 Offer》
   - 《LeetCode 101》

### 网站
- [LeetCode](https://leetcode.cn/)
- [牛客网](https://www.nowcoder.com/)
- [Go 语言中文网](https://studygolang.com/)
- [系统设计面试](https://github.com/donnemartin/system-design-primer)

### 视频
- [MIT 6.824 分布式系统](https://pdos.csail.mit.edu/6.824/)
- [极客时间 - Go 语言核心 36 讲](https://time.geekbang.org/column/intro/112)

---

## ✅ 面试前检查清单

### 1 周前
- [ ] 熟悉项目代码，能流畅讲解
- [ ] 准备 3-5 个技术难点故事
- [ ] 复习 Go 基础知识
- [ ] 刷 20-30 道 LeetCode 中等题

### 3 天前
- [ ] 复习数据库、Redis 原理
- [ ] 复习系统设计常见题
- [ ] 准备自我介绍（1 分钟、3 分钟）
- [ ] 准备常见问题答案

### 1 天前
- [ ] 模拟面试（讲解项目）
- [ ] 复习笔记
- [ ] 准备问面试官的问题
- [ ] 调整心态，充足睡眠

### 面试当天
- [ ] 提前 10 分钟到达
- [ ] 准备纸笔（画图）
- [ ] 保持自信，清晰表达
- [ ] 遇到不会的题，说出思路

---

## 💪 最后的建议

1. **项目经验最重要**: 深入理解你的项目，能讲出亮点和难点
2. **基础要扎实**: Go 语言、数据结构、算法、网络、数据库
3. **系统设计要会**: 高并发、分布式、缓存、消息队列
4. **算法要刷**: LeetCode 中等题为主，掌握常见套路
5. **心态要好**: 面试是双向选择，保持自信

**记住**: 你的交易所项目已经涵盖了很多面试考点，好好准备，一定能拿到 offer！

---

**祝你面试顺利！加油！** 🎉
