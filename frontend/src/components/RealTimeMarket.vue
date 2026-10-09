<template>
  <div class="real-time-market page-container">
    <el-card class="glass-panel mb-24">
      <template #header>
        <div class="page-header">
          <h2 class="page-title">实时行情监控</h2>
          <div class="connection-status">
            <span class="status-label">连接状态:</span>
            <span class="status-badge" :class="{ connected: isConnected }">
              <span class="live-dot" :class="{ off: !isConnected }"></span>
              {{ isConnected ? '已连接' : '未连接' }}
            </span>
          </div>
        </div>
      </template>

      <el-row :gutter="24">
        <!-- 价格变动 -->
        <el-col :xl="8" :lg="24" class="mb-20">
          <h3 class="section-title">价格变动</h3>
          <div class="price-change-display">
            <div
              class="price-item glass-card"
              v-for="(data, symbol) in priceChanges"
              :key="symbol"
            >
              <div class="symbol mono">{{ symbol }}</div>
              <div class="price mono">${{ formatPrice(data.price) }}</div>
              <div class="change badge" :class="getChangeClass(data.change_percent)">
                {{ formatChange(data.change_percent) }}
              </div>
            </div>
          </div>
        </el-col>

        <!-- 最新成交 -->
        <el-col :xl="8" :lg="24" class="mb-20">
          <h3 class="section-title">最新成交</h3>
          <div class="trades-display">
            <div
              class="trade-item glass-card"
              v-for="trade in latestTrades"
              :key="trade.id"
            >
              <div class="trade-info">
                <span class="symbol mono">{{ trade.symbol }}</span>
                <span class="price mono">${{ formatPrice(trade.price) }}</span>
                <span class="amount mono">{{ formatAmount(trade.amount) }}</span>
              </div>
              <el-tag
                class="side-tag"
                :type="trade.side === 'buy' ? 'success' : 'danger'"
                size="small"
              >
                {{ trade.side === 'buy' ? '买入' : '卖出' }}
              </el-tag>
              <div class="trade-time mono">{{ formatTime(trade.timestamp) }}</div>
            </div>
          </div>
        </el-col>

        <!-- 订单簿深度 -->
        <el-col :xl="8" :lg="24">
          <h3 class="section-title">订单簿深度</h3>
          <div class="orderbook-display">
            <div
              class="orderbook-item glass-card"
              v-for="(data, symbol) in orderbooks"
              :key="symbol"
            >
              <div class="symbol mono">{{ symbol }}</div>
              <div class="depth-info">
                <div class="bids">
                  <div class="depth-title">买盘</div>
                  <div
                    class="depth-row"
                    v-for="bid in data.bids.slice(0, 5)"
                    :key="bid[0]"
                  >
                    <span class="price mono">{{ formatPrice(bid[0]) }}</span>
                    <span class="amount mono">{{ formatAmount(bid[1]) }}</span>
                  </div>
                </div>
                <div class="asks">
                  <div class="depth-title">卖盘</div>
                  <div
                    class="depth-row"
                    v-for="ask in data.asks.slice(0, 5)"
                    :key="ask[0]"
                  >
                    <span class="price mono">{{ formatPrice(ask[0]) }}</span>
                    <span class="amount mono">{{ formatAmount(ask[1]) }}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </el-col>
      </el-row>
    </el-card>
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

let connectionInterval = null

onMounted(() => {
  // 连接WebSocket
  wsClient.connect()

  // 订阅主要交易对
  const symbols = ['BTC/USDT', 'ETH/USDT', 'BNB/USDT']
  symbols.forEach((symbol) => {
    wsClient.subscribe(symbol, handleWebSocketData)
  })

  // 每秒检查一次连接状态
  connectionInterval = setInterval(updateConnectionStatus, 1000)
})

onUnmounted(() => {
  clearInterval(connectionInterval)
  wsClient.disconnect()
})
</script>

<style scoped>
.real-time-market {
  padding: 8px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.page-title {
  margin: 0;
  font-size: 1.4rem;
  font-weight: 700;
  color: var(--text-main);
}

.connection-status {
  display: flex;
  align-items: center;
  gap: 10px;
}

.status-label {
  font-size: 0.85rem;
  color: var(--text-muted);
}

.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 0.78rem;
  font-weight: 700;
  padding: 4px 10px;
  border-radius: var(--radius-tag);
  color: var(--text-muted);
  background: color-mix(in srgb, var(--border-color-hover), transparent 85%);
}

.status-badge.connected {
  color: var(--success-emphasis);
  background: var(--success-glow);
}

.section-title {
  margin: 0 0 14px 0;
  font-size: 0.95rem;
  font-weight: 700;
  color: var(--text-muted);
}

/* Price / trade / orderbook lists */
.price-change-display,
.trades-display,
.orderbook-display {
  display: flex;
  flex-direction: column;
  gap: 10px;
  max-height: 420px;
  overflow-y: auto;
  padding: 4px;
}

.price-item,
.trade-item,
.orderbook-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 12px;
  border-radius: var(--radius-base);
  gap: 8px;
}

.price-item {
  display: grid;
  grid-template-columns: 1fr 1fr auto;
  text-align: left;
}

.symbol {
  font-weight: 700;
  color: var(--text-main);
}

.price {
  font-size: 1rem;
  font-weight: 700;
}

.amount {
  color: var(--text-muted);
}

.change {
  font-size: 0.78rem;
  font-weight: 700;
  padding: 2px 8px;
  border-radius: 12px;
}

.change.positive {
  color: var(--success-emphasis);
  background: var(--success-glow);
}

.change.negative {
  color: var(--danger-emphasis);
  background: var(--danger-glow);
}

/* Trade rows */
.trade-info {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
  min-width: 180px;
}

.trade-info .symbol { font-size: 0.9rem; }
.trade-info .price { font-size: 0.92rem; margin: 0 4px; }

.side-tag {
  font-weight: 700;
  border: none;
}

.trade-time {
  font-size: 0.75rem;
  color: var(--text-muted);
  white-space: nowrap;
}

/* Orderbook depth */
.depth-info {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 14px;
  width: 100%;
}

.depth-title {
  font-weight: 700;
  color: var(--text-main);
  margin-bottom: 4px;
  text-align: center;
  font-size: 0.8rem;
}

.depth-row {
  display: flex;
  justify-content: space-between;
  padding: 2px 0;
  font-size: 0.82rem;
}

.bids .depth-row { color: var(--success-emphasis); }
.asks .depth-row { color: var(--danger-emphasis); }

.depth-row .price { margin-right: 6px; }

/* shared */
.mb-24 { margin-bottom: 24px; }
.badge { display: inline-flex; align-items: center; }
.mono { font-family: 'Roboto Mono', 'Fira Code', monospace; }
</style>
