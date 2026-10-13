<template>
  <div class="rtm page-container">
    <PageHeader title="实时行情监控" subtitle="WebSocket 实时推送 · 价格 / 成交 / 盘口深度" icon="DataLine" tone="violet">
      <template #actions>
        <span class="status-badge" :class="{ connected: isConnected }">
          <span class="live-dot" :class="{ off: !isConnected }"></span>
          {{ isConnected ? '已连接' : '未连接' }}
        </span>
      </template>
    </PageHeader>

    <div class="rtm-grid">
      <!-- 价格变动 -->
      <section class="glass-panel panel">
        <div class="panel-head">
          <h3 class="card-title">价格变动</h3>
          <span class="badge badge-primary">{{ Object.keys(priceChanges).length }} 个交易对</span>
        </div>
        <div v-if="Object.keys(priceChanges).length" class="list">
          <div v-for="(data, symbol) in priceChanges" :key="symbol" class="price-row">
            <div class="row-left">
              <span class="pair-avatar" :style="{ background: pairColor(symbol) }">
                {{ String(symbol).split('/')[0].slice(0, 1) }}
              </span>
              <span class="symbol mono">{{ symbol }}</span>
            </div>
            <span class="price mono">${{ fmt.price(data.price, 2) }}</span>
            <span class="badge" :class="Number(data.change_percent) >= 0 ? 'badge-up' : 'badge-down'">
              {{ fmt.percent(data.change_percent) }}
            </span>
          </div>
        </div>
        <EmptyState v-else icon="Loading" title="等待行情推送" description="连接成功后，实时价格将显示在这里" />
      </section>

      <!-- 最新成交 -->
      <section class="glass-panel panel">
        <div class="panel-head">
          <h3 class="card-title">最新成交</h3>
          <span class="badge badge-ghost">{{ latestTrades.length }} 笔</span>
        </div>
        <div v-if="latestTrades.length" class="list">
          <div v-for="trade in latestTrades" :key="trade.id" class="trade-row">
            <span class="symbol mono">{{ trade.symbol }}</span>
            <span class="price mono" :class="trade.side === 'buy' ? 'text-up' : 'text-down'">
              ${{ fmt.price(trade.price, 2) }}
            </span>
            <span class="amount mono text-muted">{{ fmt.amount(trade.amount) }}</span>
            <span class="side-tag" :class="trade.side === 'buy' ? 'side-buy' : 'side-sell'">
              {{ trade.side === 'buy' ? '买入' : '卖出' }}
            </span>
            <span class="trade-time mono">{{ formatTime(trade.timestamp) }}</span>
          </div>
        </div>
        <EmptyState v-else icon="Tickets" title="暂无成交" description="最新逐笔成交记录将实时刷新" />
      </section>

      <!-- 订单簿深度 -->
      <section class="glass-panel panel">
        <div class="panel-head">
          <h3 class="card-title">订单簿深度</h3>
          <span class="badge badge-ghost">前 5 档</span>
        </div>
        <div v-if="Object.keys(orderbooks).length" class="list">
          <div v-for="(data, symbol) in orderbooks" :key="symbol" class="book-row">
            <div class="symbol mono book-symbol">{{ symbol }}</div>
            <div class="depth-info">
              <div class="depth-col">
                <div class="depth-title depth-title-bid">买盘</div>
                <div v-for="bid in (data.bids || []).slice(0, 5)" :key="'b' + bid[0]" class="depth-row bid-row">
                  <span class="mono">{{ fmt.price(bid[0], 2) }}</span>
                  <span class="mono">{{ fmt.amount(bid[1]) }}</span>
                </div>
              </div>
              <div class="depth-col">
                <div class="depth-title depth-title-ask">卖盘</div>
                <div v-for="ask in (data.asks || []).slice(0, 5)" :key="'a' + ask[0]" class="depth-row ask-row">
                  <span class="mono">{{ fmt.price(ask[0], 2) }}</span>
                  <span class="mono">{{ fmt.amount(ask[1]) }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
        <EmptyState v-else icon="Grid" title="暂无盘口数据" description="订单簿快照将随行情推送更新" />
      </section>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import wsClient from '../utils/websocket.js'
import PageHeader from './ui/PageHeader.vue'
import EmptyState from './ui/EmptyState.vue'
import { fmt } from '../composables/useMarket'

const isConnected = ref(false)
const priceChanges = ref({})
const latestTrades = ref([])
const orderbooks = ref({})

const PAIR_COLORS = {
  BTC: 'linear-gradient(135deg,#f7931a,#f2a900)',
  ETH: 'linear-gradient(135deg,#627eea,#8a92b2)',
  BNB: 'linear-gradient(135deg,#f3ba2f,#e8a90c)',
  SOL: 'linear-gradient(135deg,#9945ff,#14f195)',
  XRP: 'linear-gradient(135deg,#23292f,#4b5563)',
}
const pairColor = (symbol) => PAIR_COLORS[String(symbol).split('/')[0]] || 'linear-gradient(135deg,#3b82f6,#2563eb)'

const formatTime = (timestamp) => {
  if (!timestamp) return ''
  const ms = String(timestamp).length > 10 ? Number(timestamp) : Number(timestamp) * 1000
  return new Date(ms).toLocaleTimeString('zh-CN', { hour12: false })
}

const handleWebSocketData = (dataType, data, symbol) => {
  const key = data?.symbol || symbol
  switch (dataType) {
    case 'price_change':
      priceChanges.value[key] = data
      break
    case 'trade': {
      latestTrades.value.unshift({
        id: `${Date.now()}-${Math.random().toString(36).slice(2, 7)}`,
        symbol: key,
        price: data.price,
        amount: data.amount,
        side: data.side,
        timestamp: data.timestamp,
      })
      if (latestTrades.value.length > 20) latestTrades.value = latestTrades.value.slice(0, 20)
      break
    }
    case 'orderbook':
      orderbooks.value[key] = data
      break
  }
}

const SYMBOLS = ['BTC/USDT', 'ETH/USDT', 'BNB/USDT']
let connectionInterval = null

onMounted(() => {
  wsClient.connect()
  SYMBOLS.forEach((s) => wsClient.subscribe(s, handleWebSocketData))
  isConnected.value = wsClient.isConnected
  connectionInterval = setInterval(() => {
    isConnected.value = wsClient.isConnected
  }, 1000)
})

onUnmounted(() => {
  clearInterval(connectionInterval)
  SYMBOLS.forEach((s) => wsClient.unsubscribe(s, handleWebSocketData))
})
</script>

<style scoped>
.rtm-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: 18px;
  align-items: start;
}

.panel { padding: 18px 20px 20px; }
.panel-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 14px;
}

.list { display: flex; flex-direction: column; gap: 8px; max-height: 460px; overflow-y: auto; padding-right: 2px; }

/* 价格变动 */
.price-row {
  display: grid;
  grid-template-columns: 1fr auto auto;
  align-items: center;
  gap: 12px;
  padding: 10px 12px;
  border-radius: var(--radius-base);
  background: var(--surface-sunken);
  border: 1px solid var(--border-subtle);
  transition: var(--transition-snap);
}
.price-row:hover { border-color: var(--border-color-hover); }
.row-left { display: flex; align-items: center; gap: 9px; min-width: 0; }
.pair-avatar {
  width: 28px; height: 28px; border-radius: 9px; flex-shrink: 0;
  display: inline-flex; align-items: center; justify-content: center;
  color: #fff; font-weight: 800; font-size: 0.78rem;
}
.symbol { font-weight: 700; color: var(--text-main); font-size: 0.85rem; }
.price { font-size: 0.9rem; font-weight: 700; color: var(--text-main); }

/* 最新成交 */
.trade-row {
  display: grid;
  grid-template-columns: 1fr auto auto auto auto;
  align-items: center;
  gap: 10px;
  padding: 9px 12px;
  border-radius: var(--radius-base);
  background: var(--surface-sunken);
  border: 1px solid var(--border-subtle);
  font-size: 0.8rem;
}
.amount { font-size: 0.78rem; }
.side-tag {
  font-size: 0.68rem;
  font-weight: 700;
  padding: 2px 8px;
  border-radius: var(--radius-pill);
  white-space: nowrap;
}
.side-buy { background: var(--up-glow); color: var(--up); }
.side-sell { background: var(--down-glow); color: var(--down); }
.trade-time { font-size: 0.7rem; color: var(--text-secondary); white-space: nowrap; }

/* 订单簿 */
.book-row {
  padding: 12px;
  border-radius: var(--radius-base);
  background: var(--surface-sunken);
  border: 1px solid var(--border-subtle);
}
.book-symbol { font-size: 0.85rem; margin-bottom: 8px; }
.depth-info { display: grid; grid-template-columns: 1fr 1fr; gap: 14px; }
.depth-title { font-size: 0.72rem; font-weight: 700; margin-bottom: 5px; }
.depth-title-bid { color: var(--up); }
.depth-title-ask { color: var(--down); }
.depth-row {
  display: flex;
  justify-content: space-between;
  gap: 8px;
  padding: 2px 0;
  font-size: 0.76rem;
}
.bid-row { color: var(--up); }
.ask-row { color: var(--down); }

.text-up { color: var(--up); }
.text-down { color: var(--down); }

.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  font-size: 0.78rem;
  font-weight: 700;
  padding: 5px 12px;
  border-radius: var(--radius-pill);
  color: var(--text-muted);
  background: var(--surface-sunken);
  border: 1px solid var(--border-subtle);
}
.status-badge.connected { color: var(--success-emphasis); background: var(--success-glow); }
</style>
