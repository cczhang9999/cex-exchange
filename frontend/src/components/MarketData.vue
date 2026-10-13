<template>
  <div class="market-data">
    <el-card class="glass-panel market-card">
      <template #header>
        <div class="market-header">
          <div class="flex-center gap-8">
            <span class="symbol mono">{{ symbol }}</span>
            <span class="live-tag" :class="{ connected: isConnected }">
              <span class="live-dot" :class="{ off: !isConnected }"></span>
              {{ isConnected ? 'Live' : 'Offline' }}
            </span>
          </div>
          <Sparkline
            v-if="history.length > 1"
            :data="history"
            :width="80"
            :height="26"
            trend="auto"
          />
        </div>
      </template>

      <div class="price-info">
        <div class="current-price">
          <span class="price mono">${{ fmt.price(priceChange.price, 2) }}</span>
          <span class="change" :class="dirClass">
            <el-icon><component :is="up ? 'Top' : 'Bottom'" /></el-icon>
            {{ fmt.percent(priceChange.change_percent) }}
          </span>
        </div>

        <div class="price-details">
          <div class="detail-item">
            <span class="label">24h 最高</span>
            <span class="value mono">{{ fmt.price(priceChange.high_24h, 2) }}</span>
          </div>
          <div class="detail-item">
            <span class="label">24h 最低</span>
            <span class="value mono">{{ fmt.price(priceChange.low_24h, 2) }}</span>
          </div>
          <div class="detail-item">
            <span class="label">24h 成交量</span>
            <span class="value mono">{{ fmt.compact(priceChange.volume_24h) }}</span>
          </div>
          <div class="detail-item">
            <span class="label">24h 涨跌额</span>
            <span class="value mono" :class="dirClass">
              {{ deltaAbs }}
            </span>
          </div>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import wsClient from '../utils/websocket.js'
import Sparkline from './ui/Sparkline.vue'
import { fmt } from '../composables/useMarket'

const props = defineProps({
  symbol: { type: String, default: 'BTC/USDT' },
})

const priceChange = ref({})
const isConnected = ref(false)
const history = ref([])
const HISTORY_MAX = 60

const up = computed(() => Number(priceChange.value?.change_percent ?? 0) >= 0)
const dirClass = computed(() => (up.value ? 'positive' : 'negative'))

const deltaAbs = computed(() => {
  const p = Number(priceChange.value?.price)
  const c = Number(priceChange.value?.change_percent)
  if (Number.isNaN(p) || Number.isNaN(c)) return '--'
  const prev = p / (1 + c / 100)
  const d = p - prev
  return `${d >= 0 ? '+' : ''}${d.toFixed(2)}`
})

const pushHistory = (price) => {
  const n = Number(price)
  if (Number.isNaN(n)) return
  history.value.push(n)
  if (history.value.length > HISTORY_MAX) history.value.shift()
}

const handleWebSocketData = (dataType, data) => {
  if (dataType === 'price_change') {
    priceChange.value = data
    pushHistory(data.price)
  } else if (dataType === 'trade') {
    pushHistory(data.price)
  }
}

let connectionInterval = null

const subscribe = () => wsClient.subscribe(props.symbol, handleWebSocketData)
const unsubscribe = () => wsClient.unsubscribe(props.symbol, handleWebSocketData)

onMounted(() => {
  wsClient.connect()
  subscribe()
  isConnected.value = wsClient.isConnected
  connectionInterval = setInterval(() => {
    isConnected.value = wsClient.isConnected
  }, 1000)
})

watch(
  () => props.symbol,
  () => {
    unsubscribe()
    priceChange.value = {}
    history.value = []
    subscribe()
  }
)

onUnmounted(() => {
  clearInterval(connectionInterval)
  unsubscribe()
})
</script>

<style scoped>
.market-data { margin-bottom: 18px; }

.market-card :deep(.el-card__header) {
  background: transparent !important;
  border-bottom: 1px solid var(--border-color) !important;
  padding: 12px 18px !important;
}
.market-card :deep(.el-card__body) { padding: 16px 18px !important; }

.market-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
}

.symbol {
  font-size: 18px;
  font-weight: 700;
  color: var(--text-main);
  letter-spacing: 0.3px;
}

.live-tag {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 0.72rem;
  font-weight: 700;
  padding: 3px 9px;
  border-radius: var(--radius-pill);
  color: var(--text-muted);
  background: var(--surface-sunken);
}
.live-tag.connected { color: var(--success-emphasis); background: var(--success-glow); }

.price-info { padding: 6px 0 0; }

.current-price {
  display: flex;
  align-items: baseline;
  gap: 10px;
  margin-bottom: 18px;
  flex-wrap: wrap;
}

.price {
  font-size: 28px;
  font-weight: 800;
  color: var(--text-main);
  letter-spacing: -0.02em;
  line-height: 1.1;
}

.change {
  font-size: 0.8rem;
  padding: 3px 9px;
  border-radius: var(--radius-pill);
  font-weight: 700;
  display: inline-flex;
  align-items: center;
  gap: 3px;
}
.change.positive { color: var(--up); background: var(--up-glow); }
.change.negative { color: var(--down); background: var(--down-glow); }

.price-details {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
}

.detail-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
  padding: 9px 12px;
  border-radius: var(--radius-base);
  background: var(--surface-sunken);
  border: 1px solid var(--border-subtle);
}

.label { color: var(--text-muted); font-size: 0.75rem; white-space: nowrap; }
.value { font-size: 0.85rem; font-weight: 650; color: var(--text-main); }
.value.positive { color: var(--up); }
.value.negative { color: var(--down); }
</style>
