<template>
  <div class="real-time-market">
    <el-row :gutter="20">
      <el-col :span="24">
        <el-card>
          <template #header>
            <div class="page-header">
              <h2>实时行情监控</h2>
              <div class="connection-status">
                <span class="status-label">连接状态:</span>
                <span class="status" :class="{ 'connected': isConnected }">
                  {{ isConnected ? '已连接' : '未连接' }}
                </span>
              </div>
            </div>
          </template>
          
          <el-row :gutter="20">
            <el-col :span="8">
              <h3>价格变动</h3>
              <div class="price-change-display">
                <div class="price-item" v-for="(data, symbol) in priceChanges" :key="symbol">
                  <div class="symbol">{{ symbol }}</div>
                  <div class="price">${{ formatPrice(data.price) }}</div>
                  <div class="change" :class="getChangeClass(data.change_percent)">
                    {{ formatChange(data.change_percent) }}
                  </div>
                </div>
              </div>
            </el-col>
            
            <el-col :span="8">
              <h3>最新成交</h3>
              <div class="trades-display">
                <div class="trade-item" v-for="trade in latestTrades" :key="trade.id">
                  <div class="trade-info">
                    <span class="symbol">{{ trade.symbol }}</span>
                    <span class="price">${{ formatPrice(trade.price) }}</span>
                    <span class="amount">{{ formatAmount(trade.amount) }}</span>
                  </div>
                  <div class="trade-side" :class="trade.side">
                    {{ trade.side === 'buy' ? '买入' : '卖出' }}
                  </div>
                  <div class="trade-time">
                    {{ formatTime(trade.timestamp) }}
                  </div>
                </div>
              </div>
            </el-col>
            
            <el-col :span="8">
              <h3>订单簿深度</h3>
              <div class="orderbook-display">
                <div class="orderbook-item" v-for="(data, symbol) in orderbooks" :key="symbol">
                  <div class="symbol">{{ symbol }}</div>
                  <div class="depth-info">
                    <div class="bids">
                      <div class="depth-title">买盘</div>
                      <div class="depth-row" v-for="bid in data.bids.slice(0, 5)" :key="bid[0]">
                        <span class="price">{{ formatPrice(bid[0]) }}</span>
                        <span class="amount">{{ formatAmount(bid[1]) }}</span>
                      </div>
                    </div>
                    <div class="asks">
                      <div class="depth-title">卖盘</div>
                      <div class="depth-row" v-for="ask in data.asks.slice(0, 5)" :key="ask[0]">
                        <span class="price">{{ formatPrice(ask[0]) }}</span>
                        <span class="amount">{{ formatAmount(ask[1]) }}</span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </el-col>
          </el-row>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import wsClient from '../utils/websocket.js'

const isConnected = ref(false)
const priceChanges = ref({})
const latestTrades = ref([])
const orderbooks = ref({})

// 格式化价格
const formatPrice = (price) => {
  if (!price) return '0.00'
  return parseFloat(price).toLocaleString('en-US', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 8
  })
}

// 格式化涨跌幅
const formatChange = (change) => {
  if (!change) return '0.00%'
  const value = parseFloat(change)
  return `${value >= 0 ? '+' : ''}${value.toFixed(2)}%`
}

// 格式化数量
const formatAmount = (amount) => {
  if (!amount) return '0'
  return parseFloat(amount).toFixed(4)
}

// 格式化时间
const formatTime = (timestamp) => {
  if (!timestamp) return ''
  return new Date(timestamp * 1000).toLocaleTimeString()
}

// 获取涨跌样式类
const getChangeClass = (change) => {
  if (!change) return ''
  const value = parseFloat(change)
  return value >= 0 ? 'positive' : 'negative'
}

// 处理WebSocket数据
const handleWebSocketData = (dataType, data) => {
  switch (dataType) {
    case 'price_change':
      priceChanges.value[data.symbol] = data
      break
      
    case 'trade':
      // 添加新成交记录
      const newTrade = {
        id: Date.now() + Math.random(),
        symbol: data.symbol,
        price: data.price,
        amount: data.amount,
        side: data.side,
        timestamp: data.timestamp
      }
      latestTrades.value.unshift(newTrade)
      
      // 保持最多20条记录
      if (latestTrades.value.length > 20) {
        latestTrades.value = latestTrades.value.slice(0, 20)
      }
      break
      
    case 'orderbook':
      orderbooks.value[data.symbol] = data
      break
  }
}

// 更新连接状态
const updateConnectionStatus = () => {
  isConnected.value = wsClient.isConnected
}

onMounted(() => {
  // 连接WebSocket
  wsClient.connect()
  wsClient.startHeartbeat()
  
  // 订阅主要交易对
  const symbols = ['BTC/USDT', 'ETH/USDT', 'BNB/USDT']
  symbols.forEach(symbol => {
    wsClient.subscribe(symbol, handleWebSocketData)
  })
  
  // 监听连接状态
  const checkConnection = () => {
    updateConnectionStatus()
  }
  
  // 每秒检查一次连接状态
  const connectionInterval = setInterval(checkConnection, 1000)
  
  onUnmounted(() => {
    clearInterval(connectionInterval)
    symbols.forEach(symbol => {
      wsClient.unsubscribe(symbol)
    })
  })
})
</script>

<style scoped>
.real-time-market {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.page-header h2 {
  margin: 0;
  color: #333;
}

.connection-status {
  display: flex;
  align-items: center;
  gap: 10px;
}

.status-label {
  font-size: 14px;
  color: #666;
}

.status {
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  background-color: #f56c6c;
  color: #1e293b;
  font-weight: 600;
}

.status.connected {
  background-color: #67c23a;
}

.price-change-display,
.trades-display,
.orderbook-display {
  max-height: 400px;
  overflow-y: auto;
}

.price-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px;
  border-bottom: 1px solid #f0f0f0;
  background-color: #fafafa;
  margin-bottom: 5px;
  border-radius: 4px;
}

.symbol {
  font-weight: bold;
  color: #333;
}

.price {
  font-size: 16px;
  font-weight: bold;
  color: #333;
}

.change {
  padding: 2px 6px;
  border-radius: 3px;
  font-size: 12px;
}

.change.positive {
  color: #67c23a;
  background-color: #f0f9ff;
}

.change.negative {
  color: #f56c6c;
  background-color: #fef0f0;
}

.trade-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px;
  border-bottom: 1px solid #f0f0f0;
  background-color: #fafafa;
  margin-bottom: 3px;
  border-radius: 4px;
}

.trade-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.trade-side {
  padding: 2px 6px;
  border-radius: 3px;
  font-size: 12px;
  color: #1e293b;
  font-weight: 600;
}

.trade-side.buy {
  background-color: #67c23a;
}

.trade-side.sell {
  background-color: #f56c6c;
}

.trade-time {
  font-size: 12px;
  color: #666;
}

.orderbook-item {
  margin-bottom: 20px;
  padding: 10px;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  background-color: #fafafa;
}

.depth-info {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 15px;
  margin-top: 10px;
}

.depth-title {
  font-weight: bold;
  color: #333;
  margin-bottom: 5px;
  text-align: center;
}

.depth-row {
  display: flex;
  justify-content: space-between;
  padding: 2px 0;
  font-size: 12px;
}

.bids .depth-row {
  color: #67c23a;
}

.asks .depth-row {
  color: #f56c6c;
}

.price {
  font-weight: 500;
}

.amount {
  color: #666;
}
</style> 