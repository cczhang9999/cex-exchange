<template>
  <div class="market-data">
    <el-card class="glass-panel market-card">
      <template #header>
        <div class="market-header">
          <div class="flex-center gap-8">
            <span class="symbol mono">{{ symbol }}</span>
            <span
              class="live-tag"
              :class="{ connected: isConnected }"
            >
              <span class="live-dot" :class="{ off: !isConnected }"></span>
              {{ isConnected ? 'Live' : 'Offline' }}
            </span>
          </div>
        </div>
      </template>

      <div class="price-info">
        <div class="current-price">
          <span class="price mono">${{ formatPrice(priceChange.price) }}</span>
          <span class="change" :class="getChangeClass(priceChange.change_percent)">
            {{ formatChange(priceChange.change_percent) }}
          </span>
        </div>

        <div class="price-details">
          <div class="detail-item">
            <span class="label">24h Change</span>
            <span class="value" :class="getChangeClass(priceChange.change_percent)">
              {{ formatChange(priceChange.change_percent) }}
            </span>
          </div>
          <div class="detail-item">
            <span class="label">24h High</span>
            <span class="value mono">${{ formatPrice(priceChange.high_24h) }}</span>
          </div>
          <div class="detail-item">
            <span class="label">24h Low</span>
            <span class="value mono">${{ formatPrice(priceChange.low_24h) }}</span>
          </div>
          <div class="detail-item">
            <span class="label">24h Vol</span>
            <span class="value mono">{{ formatVolume(priceChange.volume_24h) }}</span>
          </div>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import wsClient from '../utils/websocket.js'

const props = defineProps({
  symbol: {
    type: String,
    default: 'BTC/USDT'
  }
})

const priceChange = ref({})
const isConnected = ref(false)

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

// 格式化成产量
const formatVolume = (volume) => {
  if (!volume) return '0'
  const value = parseFloat(volume)
  if (value >= 1000000) {
    return (value / 1000000).toFixed(2) + 'M'
  } else if (value >= 1000) {
    return (value / 1000).toFixed(2) + 'K'
  }
  return value.toFixed(2)
}

// 获取涨跌样式类
const getChangeClass = (change) => {
  if (!change) return ''
  const value = parseFloat(change)
  return value >= 0 ? 'positive' : 'negative'
}

// 处理WebSocket数据
const handleWebSocketData = (dataType, data) => {
  if (dataType === 'price_change') {
    priceChange.value = data
  }
}

// 更新连接状态
const updateConnectionStatus = () => {
  isConnected.value = wsClient.isConnected
}

const checkConnection = () => {
  updateConnectionStatus()
}

onMounted(() => {
  // 订阅行情数据
  wsClient.subscribe(props.symbol, handleWebSocketData)

  // 每秒检查一次连接状态
  const connectionInterval = setInterval(checkConnection, 1000)

  onUnmounted(() => {
    clearInterval(connectionInterval)
    wsClient.unsubscribe(props.symbol)
  })
})
</script>

<style scoped>
.market-data {
  margin-bottom: 20px;
}

.market-card {
  padding: 4px;
}

.market-card :global(.el-card__header) {
  background: transparent !important;
  border-bottom: 1px solid var(--border-color) !important;
  padding: 12px 20px !important;
}

.market-card :global(.el-card__body) { padding: 16px !important; }

.market-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.symbol {
  font-size: 20px;
  font-weight: 700;
  color: var(--text-main);
  letter-spacing: 0.5px;
}

.live-tag {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 0.78rem;
  font-weight: 700;
  padding: 4px 10px;
  border-radius: var(--radius-tag);
  color: var(--text-muted);
  background: color-mix(in srgb, var(--border-color-hover), transparent 80%);
}

.live-tag.connected {
  color: var(--success-emphasis);
  background: var(--success-glow);
}

.price-info {
  padding: 10px 0;
}

.current-price {
  display: flex;
  align-items: baseline;
  gap: 10px;
  margin-bottom: 20px;
}

.price {
  font-size: 30px;
  font-weight: 800;
  color: var(--text-main);
  text-shadow: 0 0 20px rgba(59, 130, 246, 0.08);
}

.change {
  font-size: 13px;
  padding: 4px 10px;
  border-radius: var(--radius-base);
  font-weight: 700;
  display: inline-flex;
  align-items: center;
}

.change.positive {
  color: var(--success-emphasis);
  background: var(--success-glow);
}

.change.negative {
  color: var(--danger-emphasis);
  background: var(--danger-glow);
}

.price-details {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 14px;
}

.detail-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 12px;
  border-radius: var(--radius-base);
  background: color-mix(in srgb, var(--bg-elevated), transparent 85%);
  border: 1px solid var(--border-color);
}

.label {
  color: var(--text-muted);
  font-size: 0.8rem;
}

.value {
  font-size: 0.92rem;
  font-weight: 600;
  color: var(--text-main);
}

.value.positive { color: var(--success-emphasis); }
.value.negative { color: var(--danger-emphasis); }

.mono { font-family: 'Roboto Mono', 'Fira Code', monospace; }
</style>
