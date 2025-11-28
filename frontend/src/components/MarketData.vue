<template>
  <div class="market-data">
    <el-card class="glass-panel">
      <template #header>
        <div class="market-header">
          <div class="flex-center">
            <span class="symbol">{{ symbol }}</span>
            <el-tag size="small" :type="isConnected ? 'success' : 'danger'" effect="dark" class="ml-10">
              {{ isConnected ? 'Live' : 'Offline' }}
            </el-tag>
          </div>
        </div>
      </template>
      
      <div class="price-info">
        <div class="current-price">
          <span class="price">${{ formatPrice(priceChange.price) }}</span>
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
            <span class="value">${{ formatPrice(priceChange.high_24h) }}</span>
          </div>
          <div class="detail-item">
            <span class="label">24h Low</span>
            <span class="value">${{ formatPrice(priceChange.low_24h) }}</span>
          </div>
          <div class="detail-item">
            <span class="label">24h Vol</span>
            <span class="value">{{ formatVolume(priceChange.volume_24h) }}</span>
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

// 格式化成交量
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

onMounted(() => {
  // 订阅行情数据
  wsClient.subscribe(props.symbol, handleWebSocketData)
  
  // 监听连接状态
  const checkConnection = () => {
    updateConnectionStatus()
  }
  
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

.price-info {
  padding: 10px 0;
}

.current-price {
  display: flex;
  align-items: baseline;
  margin-bottom: 20px;
}

.price {
  font-size: 32px;
  font-weight: 700;
  color: var(--text-main);
  margin-right: 12px;
  text-shadow: 0 0 20px rgba(59, 130, 246, 0.1);
}

.change {
  font-size: 14px;
  padding: 4px 8px;
  border-radius: 6px;
  font-weight: 600;
}

.change.positive {
  color: var(--success);
  background-color: rgba(16, 185, 129, 0.1);
}

.change.negative {
  color: var(--danger);
  background-color: rgba(239, 68, 68, 0.1);
}

.price-details {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.detail-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 0;
  border-bottom: 1px solid var(--border-color);
}

.detail-item:last-child {
  border-bottom: none;
}

.label {
  color: var(--text-muted);
  font-size: 13px;
}

.value {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-main);
}

.value.positive {
  color: var(--success);
}

.value.negative {
  color: var(--danger);
}
</style> 