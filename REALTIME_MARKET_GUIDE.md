# 实时行情监控系统 - 使用指南

## 🎯 功能概述

本系统实现了完整的实时行情推送功能，包括：

- ✅ **WebSocket服务器** - 后端实时数据推送
- ✅ **WebSocket客户端** - 前端实时数据接收
- ✅ **实时行情显示** - 价格变动、成交量、订单簿深度
- ✅ **K线图实时更新** - 基于WebSocket的K线数据推送
- ✅ **自动重连机制** - 网络断开时自动重连
- ✅ **心跳检测** - 保持连接活跃

## 📁 项目结构

### 后端文件
```
cex-exchange/internal/
├── websocket/
│   ├── manager.go          # WebSocket连接管理器
│   ├── data.go             # 数据查询函数
│   └── ...
├── controller/websocket/
│   └── websocket.go        # WebSocket HTTP控制器
└── cmd/
    └── cmd.go              # 路由配置和初始化
```

### 前端文件
```
frontend/src/
├── utils/
│   └── websocket.js        # WebSocket客户端工具类
├── components/
│   ├── MarketData.vue      # 单个交易对行情组件
│   └── RealTimeMarket.vue  # 实时行情监控页面
└── router.js               # 路由配置
```

## 🚀 快速开始

### 1. 启动后端服务

```bash
cd /Users/zhanjun/IdeaProjects/cex-exchange/cex-exchange
go run main.go
```

后端将在以下端口启动：
- HTTP服务: `http://localhost:8080`
- WebSocket服务: `ws://localhost:8080/ws`

### 2. 启动前端服务

```bash
cd /Users/zhanjun/IdeaProjects/cex-exchange/frontend
npm run dev
```

前端将在 `http://localhost:5173` 启动

### 3. 访问实时行情页面

在浏览器中访问：
- **实时行情监控页面**: http://localhost:5173/market
- **交易页面（含行情组件）**: http://localhost:5173/trade

## 📊 WebSocket API

### 连接端点
```
ws://localhost:8080/ws
```

### 消息格式

#### 1. 订阅行情数据
```json
{
  "type": "subscribe",
  "symbol": "BTC/USDT"
}
```

#### 2. 心跳检测
```json
{
  "type": "ping"
}
```

响应：
```json
{
  "type": "pong",
  "symbol": "",
  "data": "",
  "timestamp": 1638360000
}
```

### 推送数据类型

#### 1. 订单簿数据 (orderbook)
```json
{
  "type": "orderbook",
  "symbol": "BTC/USDT",
  "data": {
    "symbol": "BTC/USDT",
    "bids": [["45100", "0.5"], ["45050", "1.2"]],
    "asks": [["45150", "0.3"], ["45200", "0.9"]]
  },
  "timestamp": 1638360000
}
```

#### 2. 成交数据 (trade)
```json
{
  "type": "trade",
  "symbol": "BTC/USDT",
  "data": {
    "symbol": "BTC/USDT",
    "price": "45120",
    "amount": "0.1",
    "side": "buy",
    "timestamp": 1638360000
  },
  "timestamp": 1638360000
}
```

#### 3. 价格变动数据 (price_change)
```json
{
  "type": "price_change",
  "symbol": "BTC/USDT",
  "data": {
    "symbol": "BTC/USDT",
    "price": "45120",
    "change": "120.5",
    "change_percent": "2.34",
    "high_24h": "45200",
    "low_24h": "44900",
    "volume_24h": "1234.56"
  },
  "timestamp": 1638360000
}
```

#### 4. K线数据 (kline)
```json
{
  "type": "kline",
  "symbol": "BTC/USDT",
  "data": {
    "symbol": "BTC/USDT",
    "interval": "1m",
    "open": "45000",
    "high": "45200",
    "low": "44900",
    "close": "45120",
    "volume": "123.45",
    "open_time": 1638360000,
    "close_time": 1638360060
  },
  "timestamp": 1638360000
}
```

## 🔧 配置说明

### 后端配置

#### 1. 修改推送频率
在 `internal/websocket/manager.go` 中：
```go
func StartMarketDataPusher() {
    ticker := time.NewTicker(1 * time.Second) // 修改这里，如 500ms
    // ...
}
```

#### 2. 添加交易对
在 `internal/websocket/manager.go` 中：
```go
func StartMarketDataPusher() {
    // ...
    symbols := []string{"BTC/USDT", "ETH/USDT", "BNB/USDT", "ADA/USDT"} // 添加新交易对
    // ...
}
```

#### 3. 限制连接来源（生产环境）
在 `internal/websocket/manager.go` 中：
```go
upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        // 只允许特定域名
        origin := r.Header.Get("Origin")
        return origin == "https://yourdomain.com"
    },
}
```

### 前端配置

#### 1. 修改WebSocket服务器地址
在 `frontend/src/utils/websocket.js` 中：
```javascript
const wsClient = new WebSocketClient('ws://your-server:8080/ws')
```

#### 2. 修改重连参数
在 `frontend/src/utils/websocket.js` 中：
```javascript
constructor(url) {
    this.maxReconnectAttempts = 5      // 最大重连次数
    this.reconnectInterval = 3000       // 重连间隔（毫秒）
}
```

#### 3. 修改心跳间隔
在 `frontend/src/utils/websocket.js` 中：
```javascript
startHeartbeat() {
    setInterval(() => {
        if (this.isConnected) {
            this.send({ type: 'ping' })
        }
    }, 30000) // 修改心跳间隔（毫秒）
}
```

## 🎨 前端组件使用

### 1. 在页面中使用MarketData组件
```vue
<template>
  <div>
    <MarketData symbol="BTC/USDT" />
  </div>
</template>

<script setup>
import MarketData from './components/MarketData.vue'
</script>
```

### 2. 在页面中使用RealTimeMarket组件
```vue
<template>
  <div>
    <RealTimeMarket />
  </div>
</template>

<script setup>
import RealTimeMarket from './components/RealTimeMarket.vue'
</script>
```

### 3. 自定义WebSocket订阅
```javascript
import wsClient from '@/utils/websocket.js'

// 连接WebSocket
wsClient.connect()

// 订阅行情数据
wsClient.subscribe('BTC/USDT', (dataType, data) => {
  if (dataType === 'price_change') {
    console.log('价格变动:', data)
  } else if (dataType === 'trade') {
    console.log('新成交:', data)
  } else if (dataType === 'orderbook') {
    console.log('订单簿更新:', data)
  }
})

// 取消订阅
wsClient.unsubscribe('BTC/USDT')

// 断开连接
wsClient.disconnect()
```

## 🔍 调试和监控

### 1. 查看WebSocket状态
访问: `http://localhost:8080/ws/status`

响应示例:
```json
{
  "code": 200,
  "data": {
    "connected_clients": 5,
    "status": "running"
  }
}
```

### 2. 浏览器控制台调试
打开浏览器开发者工具，在Console中可以看到：
- WebSocket连接状态
- 接收到的消息
- 错误信息

### 3. 后端日志
后端会输出以下日志：
```
WebSocket管理器已启动
行情数据推送已启动
WebSocket客户端连接: 127.0.0.1:xxxxx
WebSocket客户端断开: 127.0.0.1:xxxxx
```

## ⚠️ 常见问题

### 1. WebSocket连接失败
**问题**: 前端无法连接到WebSocket服务器

**解决方案**:
- 确认后端服务已启动
- 检查端口8080是否被占用
- 检查防火墙设置
- 确认WebSocket URL正确（ws://localhost:8080/ws）

### 2. 数据不更新
**问题**: 连接成功但没有数据推送

**解决方案**:
- 检查数据库中是否有数据
- 确认交易对名称正确
- 查看后端日志是否有错误
- 检查浏览器控制台是否有错误

### 3. 连接频繁断开
**问题**: WebSocket连接不稳定

**解决方案**:
- 检查网络连接
- 增加心跳间隔时间
- 检查服务器资源使用情况
- 查看后端日志中的错误信息

### 4. 性能问题
**问题**: 页面卡顿或响应慢

**解决方案**:
- 减少推送频率（修改ticker间隔）
- 限制订单簿深度（修改Limit值）
- 限制成交记录数量
- 优化数据库查询

## 📈 性能优化建议

### 1. 后端优化
- 使用Redis缓存订单簿数据
- 实现订阅机制，只推送用户关注的交易对
- 使用消息队列处理高并发
- 实现数据压缩

### 2. 前端优化
- 使用虚拟滚动显示大量数据
- 实现数据节流和防抖
- 只订阅当前页面需要的交易对
- 离开页面时断开连接

### 3. 网络优化
- 使用CDN加速静态资源
- 启用Gzip压缩
- 使用WebSocket压缩扩展
- 实现断线重连指数退避

## 🔐 安全建议

### 1. 生产环境配置
```go
// 限制连接来源
upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        origin := r.Header.Get("Origin")
        allowedOrigins := []string{
            "https://yourdomain.com",
            "https://www.yourdomain.com",
        }
        for _, allowed := range allowedOrigins {
            if origin == allowed {
                return true
            }
        }
        return false
    },
}
```

### 2. 添加身份验证
```go
// 在HandleConnection中验证token
func (m *Manager) HandleConnection(ctx context.Context, conn *websocket.Conn) {
    // 读取第一条消息验证token
    var authMsg map[string]interface{}
    conn.ReadJSON(&authMsg)
    
    token, ok := authMsg["token"].(string)
    if !ok || !validateToken(token) {
        conn.Close()
        return
    }
    
    // 继续处理...
}
```

### 3. 实现速率限制
```go
// 限制每个IP的连接数和消息频率
type RateLimiter struct {
    connections map[string]int
    messages    map[string]int
    mutex       sync.RWMutex
}
```

## 📝 开发计划

- [ ] 添加更多K线周期（5m, 15m, 1h, 1d）
- [ ] 实现技术指标计算（MA, MACD, RSI）
- [ ] 添加历史数据查询API
- [ ] 实现移动端适配
- [ ] 添加推送通知功能
- [ ] 实现数据导出功能
- [ ] 添加更多图表类型
- [ ] 实现自定义监控面板

## 📚 相关文档

- [GoFrame官方文档](https://goframe.org/)
- [Gorilla WebSocket文档](https://github.com/gorilla/websocket)
- [Vue 3官方文档](https://vuejs.org/)
- [Element Plus文档](https://element-plus.org/)

## 🤝 贡献指南

欢迎提交Issue和Pull Request！

## 📄 许可证

MIT License
