// Live market data composable — subscribes to the exchange WebSocket for a set of
// symbols and maintains price, 24h stats, and a short rolling history (for sparklines).
// Falls back to seeded baselines so the dashboard renders cleanly before the first tick.
import { reactive, ref, onMounted, onUnmounted } from 'vue'
import wsClient from '../utils/websocket'

export const MARKET_SYMBOLS = ['BTC/USDT', 'ETH/USDT', 'BNB/USDT', 'SOL/USDT', 'XRP/USDT']

// Baseline reference values (used only until live ticks arrive).
const SEEDS = {
  'BTC/USDT': { price: 45123.45, change: 2.34, high: 46200, low: 43800, volume: 1245000 },
  'ETH/USDT': { price: 2380.12, change: 1.18, high: 2450, low: 2310, volume: 892000 },
  'BNB/USDT': { price: 312.44, change: -0.76, high: 320, low: 306, volume: 431000 },
  'SOL/USDT': { price: 98.31, change: 4.52, high: 101, low: 92, volume: 356000 },
  'XRP/USDT': { price: 0.5421, change: -1.32, high: 0.56, low: 0.53, volume: 210000 },
}

const HISTORY_LEN = 40

function seedHistory(base) {
  // Gentle pseudo-random walk so sparklines look alive before real data.
  const out = []
  let p = base
  for (let i = 0; i < HISTORY_LEN; i++) {
    p = p * (1 + (Math.random() - 0.48) * 0.004)
    out.push(p)
  }
  return out
}

export function useMarket(symbols = MARKET_SYMBOLS) {
  const rows = reactive({})
  const isConnected = ref(false)
  const lastUpdate = ref(0)

  symbols.forEach((s) => {
    const seed = SEEDS[s] || { price: 1, change: 0, high: 1, low: 1, volume: 0 }
    rows[s] = {
      symbol: s,
      price: seed.price,
      changePercent: seed.change,
      high: seed.high,
      low: seed.low,
      volume: seed.volume,
      history: seedHistory(seed.price),
      updatedAt: 0,
    }
  })

  const handle = (dataType, data, symbol) => {
    const key = data?.symbol || symbol
    if (!rows[key]) return
    const row = rows[key]
    if (dataType === 'price_change') {
      const price = parseFloat(data.price)
      if (!Number.isNaN(price)) {
        row.price = price
        row.history.push(price)
        if (row.history.length > HISTORY_LEN) row.history.shift()
      }
      if (data.change_percent !== undefined) row.changePercent = parseFloat(data.change_percent)
      if (data.high_24h !== undefined) row.high = parseFloat(data.high_24h)
      if (data.low_24h !== undefined) row.low = parseFloat(data.low_24h)
      if (data.volume_24h !== undefined) row.volume = parseFloat(data.volume_24h)
      row.updatedAt = Date.now()
      lastUpdate.value = Date.now()
    } else if (dataType === 'trade') {
      const price = parseFloat(data.price)
      if (!Number.isNaN(price)) {
        row.price = price
        row.history.push(price)
        if (row.history.length > HISTORY_LEN) row.history.shift()
      }
      row.updatedAt = Date.now()
      lastUpdate.value = Date.now()
    }
  }

  let statusTimer = null

  onMounted(() => {
    wsClient.connect()
    symbols.forEach((s) => wsClient.subscribe(s, handle))
    statusTimer = setInterval(() => {
      isConnected.value = wsClient.isConnected
    }, 1000)
    isConnected.value = wsClient.isConnected
  })

  onUnmounted(() => {
    clearInterval(statusTimer)
    symbols.forEach((s) => wsClient.unsubscribe(s))
  })

  return { rows, isConnected, lastUpdate, symbols }
}

// Formatting helpers shared across market views.
export const fmt = {
  price(v, max = 2) {
    if (v === null || v === undefined || v === '' || Number.isNaN(Number(v))) return '--'
    return Number(v).toLocaleString('en-US', {
      minimumFractionDigits: 2,
      maximumFractionDigits: Math.max(2, max),
    })
  },
  amount(v, digits = 4) {
    if (v === null || v === undefined || v === '' || Number.isNaN(Number(v))) return '--'
    return Number(v).toFixed(digits)
  },
  percent(v) {
    if (v === null || v === undefined || Number.isNaN(Number(v))) return '0.00%'
    const n = Number(v)
    return `${n >= 0 ? '+' : ''}${n.toFixed(2)}%`
  },
  compact(v) {
    const n = Number(v)
    if (!n || Number.isNaN(n)) return '0'
    if (n >= 1e9) return (n / 1e9).toFixed(2) + 'B'
    if (n >= 1e6) return (n / 1e6).toFixed(2) + 'M'
    if (n >= 1e3) return (n / 1e3).toFixed(2) + 'K'
    return n.toFixed(2)
  },
  money(v) {
    const n = Number(v)
    if (Number.isNaN(n)) return '--'
    return n.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
  },
}
