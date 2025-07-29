# 实时行情推送系统

## 功能概述

本系统实现了完整的实时行情推送功能，包括：

- **WebSocket服务器** - 后端实时数据推送
- **WebSocket客户端** - 前端实时数据接收
- **实时行情显示** - 价格变动、成交量、订单簿深度
- **K线图实时更新** - 基于WebSocket的K线数据推送
- **自动重连机制** - 网络断开时自动重连
- **心跳检测** - 保持连接活跃

## 系统架构

### 后端组件

1. **WebSocket服务器** (`websocket.go`)
   - 连接管理
   - 消息广播
   - 数据推送

2. **行情数据生成器**
   - 订单簿数据
   - 成交记录
   - 价格变动
   - K线数据

3. **数据推送触发器**
   - 撮合引擎触发
   - 定时推送

### 前端组件

1. **WebSocket客户端** (`utils/websocket.js`)
   - 连接管理
   - 消息处理
   - 自动重连

2. **实时行情组件** (`MarketData.vue`)
   - 价格显示
   - 涨跌幅
   - 成交量

3. **实时监控页面** (`RealTimeMarket.vue`)
   - 多交易对监控
   - 实时成交记录
   - 订单簿深度

## 数据格式

### WebSocket消息格式

```json
{
  "type": "orderbook|trade|price_change|kline",
  "symbol": "BTC/USDT",
  "data": {...},
  "timestamp": 1642234567
}
```

### 订单簿数据

```json
{
  "symbol": "BTC/USDT",
  "bids": [["45100", "0.5"], ["45050", "1.2"]],
  "asks": [["45150", "0.3"], ["45200", "0.9"]]
}
```

### 成交数据

```json
{
  "symbol": "BTC/USDT",
  "price": "45120",
  "amount": "0.1",
  "side": "buy",
  "timestamp": 1642234567
}
```

### 价格变动数据

```json
{
  "symbol": "BTC/USDT",
  "price": "45120",
  "change": "120.5",
  "change_percent": "2.34",
  "high_24h": "45200",
  "low_24h": "44900",
  "volume_24h": "1234.56"
}
```

## 使用方法

### 1. 启动后端服务

```bash
cd cex-exchange
go run main.go models.go auth.go funds.go match.go kline.go admin.go websocket.go
```

### 2. 启动前端服务

```bash
cd frontend
npm run dev
```

### 3. 访问实时行情

- **交易页面**: 在交易中心查看实时行情
- **监控页面**: 访问"实时行情"菜单查看多交易对监控

### 4. WebSocket连接

前端会自动连接到 `ws://localhost:8080/ws`

## 功能特性

### 实时数据推送

- **订单簿更新**: 买卖盘实时更新
- **成交记录**: 最新成交实时显示
- **价格变动**: 24小时涨跌幅、最高最低价
- **K线数据**: 实时K线图更新

### 连接管理

- **自动重连**: 网络断开时自动重连
- **心跳检测**: 30秒心跳保持连接
- **连接状态**: 实时显示连接状态
- **错误处理**: 完善的错误处理机制

### 性能优化

- **数据缓存**: 避免重复请求
- **消息队列**: 异步消息处理
- **连接池**: 高效连接管理
- **内存优化**: 限制数据量防止内存溢出

## 扩展功能

### 1. 添加新交易对

在后端 `startMarketDataPusher()` 函数中添加：

```go
symbols := []string{"BTC/USDT", "ETH/USDT", "BNB/USDT", "ADA/USDT"}
```

### 2. 自定义推送频率

修改推送间隔：

```go
ticker := time.NewTicker(500 * time.Millisecond) // 500ms推送一次
```

### 3. 添加更多数据类型

在 `websocket.go` 中添加新的数据类型和处理逻辑。

### 4. 实现用户订阅

支持用户自定义订阅特定交易对：

```javascript
wsClient.subscribe('BTC/USDT', handleData)
wsClient.subscribe('ETH/USDT', handleData)
```

## 故障排除

### 1. WebSocket连接失败

- 检查后端服务是否启动
- 确认端口8080是否被占用
- 检查防火墙设置

### 2. 数据不更新

- 检查数据库连接
- 确认有交易数据生成
- 查看浏览器控制台错误

### 3. 性能问题

- 减少推送频率
- 限制数据量
- 优化数据库查询

## 安全考虑

### 1. 生产环境配置

- 限制WebSocket连接来源
- 添加身份验证
- 实现速率限制

### 2. 数据验证

- 验证输入数据格式
- 防止SQL注入
- 限制消息大小

### 3. 监控告警

- 连接数监控
- 错误率统计
- 性能指标监控

## 技术栈

- **后端**: Go + Gin + GORM + WebSocket
- **前端**: Vue 3 + Element Plus + ECharts
- **数据库**: MySQL
- **通信**: WebSocket

## 开发计划

- [ ] 添加更多K线周期
- [ ] 实现技术指标计算
- [ ] 添加历史数据查询
- [ ] 实现移动端适配
- [ ] 添加推送通知
- [ ] 实现数据导出功能 