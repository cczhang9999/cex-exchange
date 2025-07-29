<template>
  <div class="market-data">
    <el-card>
      <template #header>
        <div class="market-header">
          <span class="symbol">{{ symbol }}</span>
          <span class="status" :class="{ 'connected': isConnected }">
            {{ isConnected ? '实时' : '离线' }}
          </span>
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
            <span class="label">24h涨跌:</span>
            <span class="value" :class="getChangeClass(priceChange.change_percent)">
              {{ formatChange(priceChange.change_percent) }}
            </span>
          </div>
          <div class="detail-item">
            <span class="label">24h最高:</span>
            <span class="value">${{ formatPrice(priceChange.high_24h) }}</span>
          </div>
          <div class="detail-item">
            <span class="label">24h最低:</span>
            <span class="value">${{ formatPrice(priceChange.low_24h) }}</span>
          </div>
          <div class="detail-item">
            <span class="label">24h成交量:</span>
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
  font-size: 18px;
  font-weight: bold;
  color: #333;
}

.status {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  background-color: #f56c6c;
  color: white;
}

.status.connected {
  background-color: #67c23a;
}

.price-info {
  padding: 10px 0;
}

.current-price {
  display: flex;
  align-items: baseline;
  margin-bottom: 15px;
}

.price {
  font-size: 24px;
  font-weight: bold;
  color: #333;
  margin-right: 10px;
}

.change {
  font-size: 14px;
  padding: 2px 6px;
  border-radius: 3px;
}

.change.positive {
  color: #67c23a;
  background-color: #f0f9ff;
}

.change.negative {
  color: #f56c6c;
  background-color: #fef0f0;
}

.price-details {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
}

.detail-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 5px 0;
  border-bottom: 1px solid #f0f0f0;
}

.detail-item:last-child {
  border-bottom: none;
}

.label {
  color: #666;
  font-size: 12px;
}

.value {
  font-size: 12px;
  font-weight: 500;
}

.value.positive {
  color: #67c23a;
}

.value.negative {
  color: #f56c6c;
}
</style> 