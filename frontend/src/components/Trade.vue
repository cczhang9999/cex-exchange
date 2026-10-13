<template>
  <el-row :gutter="20" class="trade-layout" id="trade-root">
    <el-col :span="8" class="trade-sidebar">
      <!-- 实时行情显示 -->
      <MarketData :symbol="orderForm.symbol" />

      <el-card class="glass-panel trade-card">
        <template #header>
          <div class="flex-between">
            <h3 class="card-title">Place Order</h3>
            <el-tag size="small" effect="dark" type="info" class="glass-tag">Spot</el-tag>
          </div>
        </template>
        <el-form :model="orderForm" label-position="top" class="trade-form">
          <el-form-item label="Pair">
            <el-select v-model="orderForm.symbol" class="custom-select w-100" popper-class="custom-dropdown">
              <el-option v-for="symbol in availableSymbols" :key="symbol" :label="symbol" :value="symbol" />
            </el-select>
          </el-form-item>

          <el-form-item label="Side" class="mb-4">
            <div class="trade-type-switch">
              <div
                class="switch-item"
                :class="{ active: orderForm.side === 'buy', 'buy-active': orderForm.side === 'buy' }"
                @click="orderForm.side = 'buy'"
              >
                <el-icon><ArrowUp /></el-icon>
                <span>Buy</span>
              </div>
              <div
                class="switch-item"
                :class="{ active: orderForm.side === 'sell', 'sell-active': orderForm.side === 'sell' }"
                @click="orderForm.side = 'sell'"
              >
                <el-icon><ArrowDown /></el-icon>
                <span>Sell</span>
              </div>
            </div>
          </el-form-item>

          <el-form-item label="Type">
            <el-select v-model="orderForm.type" class="custom-select w-100" popper-class="custom-dropdown">
              <el-option label="Limit" value="limit" />
              <el-option label="Market" value="market" />
              <el-option label="Stop-Limit" value="stop_limit" />
              <el-option label="Stop-Market" value="stop_market" />
            </el-select>
          </el-form-item>

          <!-- 触发价格 (止损止盈订单) -->
          <el-form-item label="Stop Price" v-if="orderForm.type === 'stop_limit' || orderForm.type === 'stop_market'">
            <el-input v-model="orderForm.stop_price" placeholder="0.00" class="custom-input">
              <template #suffix>USDT</template>
            </el-input>
            <div class="stop-hint">
              <span v-if="orderForm.side === 'buy'">当价格 ≥ {{ orderForm.stop_price || '?' }} 时触发买入</span>
              <span v-else>当价格 ≤ {{ orderForm.stop_price || '?' }} 时触发卖出</span>
            </div>
          </el-form-item>

          <el-form-item label="Price" v-if="orderForm.type === 'limit' || orderForm.type === 'stop_limit'">
            <el-input v-model="orderForm.price" placeholder="0.00" class="custom-input">
              <template #suffix>USDT</template>
            </el-input>
          </el-form-item>

          <el-form-item label="Amount">
            <el-input v-model="orderForm.amount" placeholder="0.00" class="custom-input">
              <template #suffix>{{ orderForm.symbol.split('/')[0] }}</template>
            </el-input>
          </el-form-item>

          <!-- 交易额估算 -->
          <div v-if="(orderForm.type === 'limit' || orderForm.type === 'stop_limit') && orderForm.price && orderForm.amount" class="trade-total mb-4">
            <span>Total</span>
            <span class="total-value mono">{{ (parseFloat(orderForm.price) * parseFloat(orderForm.amount)).toFixed(2) }} USDT</span>
          </div>

          <!-- 触发价格说明 (止损订单) -->
          <div v-if="orderForm.type === 'stop_market' && orderForm.stop_price && orderForm.amount" class="trade-total mb-4">
            <span>Trigger Total</span>
            <span class="total-value mono">{{ (parseFloat(orderForm.stop_price) * parseFloat(orderForm.amount)).toFixed(2) }} USDT</span>
          </div>

          <el-form-item class="mt-6">
            <el-button
              :class="['trade-btn', orderForm.side === 'buy' ? 'btn-buy' : 'btn-sell']"
              @click="placeOrder"
            >
              {{ orderForm.side === 'buy' ? 'Buy' : 'Sell' }} {{ orderForm.symbol.split('/')[0] }}
            </el-button>
          </el-form-item>
        </el-form>
      </el-card>

      <!-- K线图 -->
      <el-card class="glass-panel chart-card">
        <template #header>
          <div class="flex-between">
            <div class="flex-center gap-8">
              <h3 class="card-title">{{ orderForm.symbol }}</h3>
              <span class="badge badge-primary">K线</span>
            </div>
            <div class="flex-center time-intervals">
              <el-button 
                v-for="period in klinePeriods" 
                :key="period.value"
                size="small" 
                text 
                :class="{ active: currentKlinePeriod === period.value }"
                @click="switchKlinePeriod(period.value)"
              >
                {{ period.label }}
              </el-button>
            </div>
          </div>
        </template>
        <div class="ohlc-row">
          <span v-for="item in ohlcItems" :key="item.label" class="ohlc-item">
            <span class="ohlc-label">{{ item.label }}</span>
            <span class="ohlc-value mono" :class="item.cls">{{ item.value }}</span>
          </span>
        </div>
        <div id="kline" class="kline-chart"></div>
      </el-card>
    </el-col>

    <el-col :span="16" class="trade-main">
      <el-card class="glass-panel mb-20">
        <template #header>
          <div class="flex-between">
            <h3 class="card-title">Order Book</h3>
            <el-tag size="small" type="info" class="glass-tag">{{ orderForm.symbol }}</el-tag>
          </div>
        </template>
        <el-row :gutter="20">
          <!-- 买单（Bids） -->
          <el-col :span="12">
            <div class="book-header book-header-buy">
              <el-icon class="text-up"><ArrowUp /></el-icon>
              <span class="text-up">Bids (Buy)</span>
            </div>
            <el-table
              :data="orderbook?.bids || []"
              size="small"
              height="300"
              :show-header="true"
              class="order-book-table"
              @row-click="handleBidClick"
            >
              <el-table-column label="Price (USDT)" width="120" align="left">
                <template #default="{ row }">
                  <span class="text-up price-text clickable mono">{{ formatPrice(row.price) }}</span>
                </template>
              </el-table-column>
              <el-table-column label="Amount" width="100" align="right">
                <template #default="{ row }">
                  <span class="amount-text mono">{{ formatAmount(row.amount) }}</span>
                </template>
              </el-table-column>
              <el-table-column label="Total (USDT)" align="right">
                <template #default="{ row }">
                  <span class="total-text mono">{{ formatTotal(row.price, row.amount) }}</span>
                </template>
              </el-table-column>
            </el-table>
          </el-col>

          <!-- 卖单（Asks） -->
          <el-col :span="12">
            <div class="book-header book-header-sell">
              <el-icon class="text-down"><ArrowDown /></el-icon>
              <span class="text-down">Asks (Sell)</span>
            </div>
            <el-table
              :data="orderbook?.asks || []"
              size="small"
              height="300"
              :show-header="true"
              class="order-book-table"
              @row-click="handleAskClick"
            >
              <el-table-column label="Price (USDT)" width="120" align="left">
                <template #default="{ row }">
                  <span class="text-down price-text clickable mono">{{ formatPrice(row.price) }}</span>
                </template>
              </el-table-column>
              <el-table-column label="Amount" width="100" align="right">
                <template #default="{ row }">
                  <span class="amount-text mono">{{ formatAmount(row.amount) }}</span>
                </template>
              </el-table-column>
              <el-table-column label="Total (USDT)" align="right">
                <template #default="{ row }">
                  <span class="total-text mono">{{ formatTotal(row.price, row.amount) }}</span>
                </template>
              </el-table-column>
            </el-table>
          </el-col>
        </el-row>
      </el-card>

      <el-card class="glass-panel">
        <template #header>
          <div class="flex-between">
            <h3 class="card-title">Recent Trades</h3>
            <el-tag size="small" type="success" class="glass-tag">{{ trades?.length || 0 }} trades</el-tag>
          </div>
        </template>
        <el-table :data="trades || []" size="small" height="200" class="trades-table">
          <el-table-column label="Time" width="100">
            <template #default="scope">
              <span class="text-muted time-text mono">{{ tradeTime(scope.row) }}</span>
            </template>
          </el-table-column>
          <el-table-column label="Side" width="80" align="center">
            <template #default="{ row }">
              <el-tag
                :type="row.side === 'buy' ? 'danger' : 'success'"
                size="small"
                effect="dark"
              >
                {{ row.side === 'buy' ? '买入' : '卖出' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="Price (USDT)" width="120" align="right">
            <template #default="{ row }">
              <span :class="row.side === 'buy' ? 'text-up' : 'text-down'" class="price-text mono">
                {{ formatPrice(row.price) }}
              </span>
            </template>
          </el-table-column>
          <el-table-column label="Amount" width="100" align="right">
            <template #default="{ row }">
              <span class="amount-text mono">{{ formatAmount(row.amount) }}</span>
            </template>
          </el-table-column>
          <el-table-column label="Total (USDT)" align="right">
            <template #default="{ row }">
              <span class="total-text mono">{{ formatTotal(row.price, row.amount) }}</span>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </el-col>
  </el-row>
</template>

<script setup>
import { ref, computed, onMounted, watch, nextTick, onUnmounted } from 'vue'
import { placeOrder as apiPlaceOrder, getOrderbook, getTrades, getKline } from '../api/api.js'
import * as echarts from 'echarts'
import wsClient from '../utils/websocket.js'
import MarketData from './MarketData.vue'
import { ElMessage } from 'element-plus'
import { ArrowUp, ArrowDown } from '@element-plus/icons-vue'
import { getChartTheme, tooltipStyle, crosshairStyle } from '../utils/chartTheme'
import { isDark } from '../utils/theme'

const props = defineProps({
  initialSymbol: { type: String, default: 'BTC/USDT' },
})

const availableSymbols = ref(['BTC/USDT', 'ETH/USDT', 'BNB/USDT', 'SOL/USDT', 'XRP/USDT'])
const orderForm = ref({ symbol: props.initialSymbol || 'BTC/USDT', side: 'buy', type: 'limit', price: '', amount: '', stop_price: '' })
const orderbook = ref({ bids: [], asks: [] })
const trades = ref([])
const priceChange = ref({})
let klineChart = null
let currentSymbol = orderForm.value.symbol
const klineData = ref([])          // [time, open, close, low, high, volume]
const lastCandle = ref(null)

// 最新一根 K 线的 OHLC 摘要
const ohlcItems = computed(() => {
  const c = lastCandle.value
  const empty = '--'
  if (!c) {
    return [
      { label: '开', value: empty, cls: '' },
      { label: '高', value: empty, cls: '' },
      { label: '低', value: empty, cls: '' },
      { label: '收', value: empty, cls: '' },
      { label: '量', value: empty, cls: '' },
    ]
  }
  const rising = c[2] >= c[1]
  const dirCls = rising ? 'text-up' : 'text-down'
  return [
    { label: '开', value: Number(c[1]).toFixed(2), cls: dirCls },
    { label: '高', value: Number(c[4]).toFixed(2), cls: 'text-up' },
    { label: '低', value: Number(c[3]).toFixed(2), cls: 'text-down' },
    { label: '收', value: Number(c[2]).toFixed(2), cls: dirCls },
    { label: '量', value: Number(c[5]).toFixed(2), cls: 'text-muted' },
  ]
})

// K线周期配置
const klinePeriods = [
  { label: '1m', value: '1m', interval: 60000 },
  { label: '5m', value: '5m', interval: 300000 },
  { label: '15m', value: '15m', interval: 900000 },
  { label: '30m', value: '30m', interval: 1800000 },
  { label: '1h', value: '1h', interval: 3600000 },
  { label: '4h', value: '4h', interval: 14400000 },
  { label: '1d', value: '1d', interval: 86400000 },
  { label: '1w', value: '1w', interval: 604800000 }
]
const currentKlinePeriod = ref('1m')

// 格式化时间显示
const formatTime = (timestamp) => {
  if (!timestamp) return ''
  const date = new Date(timestamp)
  return date.toLocaleTimeString('zh-CN', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}

// 成交记录时间：REST 返回 timestamp（秒），WebSocket 推送为 created_at（ISO 字符串）
const tradeTime = (row) => {
  if (!row) return ''
  if (row.created_at) return formatTime(row.created_at)
  const ts = Number(row.timestamp)
  return !Number.isNaN(ts) && ts > 0 ? formatTime(ts * 1000) : ''
}

// ────────────────────────── K线图 ──────────────────────────
const KLINE_MAX = 200

const toMs = (t) => {
  const n = Number(t)
  if (Number.isNaN(n) || n === 0) return Date.now()
  return n > 1e12 ? n : n * 1000
}

const fmtAxisTime = (ts, period) => {
  const d = new Date(ts)
  if (period === '1d' || period === '1w') return `${d.getMonth() + 1}/${d.getDate()}`
  if (period === '4h' || period === '1h') {
    return `${d.getMonth() + 1}/${d.getDate()} ${String(d.getHours()).padStart(2, '0')}:00`
  }
  return `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
}

// 后端无 K 线数据时本地生成，保证图表始终可用
const buildFallbackKline = (period, basePrice) => {
  const periodMs = klinePeriods.find((p) => p.value === period)?.interval || 60000
  const count = 90
  const base = Number(basePrice) > 0 ? Number(basePrice) : 45000
  const out = []
  let price = base * 0.97
  const start = Date.now() - count * periodMs
  for (let i = 0; i < count; i++) {
    const open = price
    const close = Math.max(open * (1 + (Math.random() - 0.48) * 0.006), 0.0001)
    const high = Math.max(open, close) * (1 + Math.random() * 0.0018)
    const low = Math.min(open, close) * (1 - Math.random() * 0.0018)
    const vol = 60 + Math.random() * 260
    out.push([fmtAxisTime(start + i * periodMs, period), open, close, low, high, vol])
    price = close
  }
  return out
}

// 构建（主题感知的）K 线配置
const buildKlineOption = (data, symbol) => {
  const t = getChartTheme()
  const times = data.map((d) => d[0])
  const ohlc = data.map((d) => [d[1], d[2], d[3], d[4]])
  const vols = data.map((d) => ({
    value: d[5],
    itemStyle: { color: d[2] >= d[1] ? t.volumeUp : t.volumeDown },
  }))

  const row = (label, val, color) =>
    `<div style="display:flex;justify-content:space-between;gap:20px">
       <span style="opacity:.65">${label}</span>
       <span style="font-weight:600;${color ? `color:${color}` : ''}">${Number(val).toFixed(2)}</span>
     </div>`

  return {
    backgroundColor: 'transparent',
    animation: true,
    animationDuration: 260,
    grid: [
      { left: 10, right: 64, top: 14, height: '58%' },
      { left: 10, right: 64, top: '75%', height: '16%' },
    ],
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'cross', ...crosshairStyle(t) },
      ...tooltipStyle(t),
      formatter: (params) => {
        const k = params.find((p) => p.seriesType === 'candlestick')
        if (!k) return ''
        const raw = Array.isArray(k.data) ? k.data : k.data?.value || []
        const arr = raw.length >= 5 ? raw.slice(-4) : raw
        const [o, c, l, h] = arr
        const rising = c >= o
        const col = rising ? t.upColor : t.downColor
        const pct = o ? (((c - o) / o) * 100).toFixed(2) : '0.00'
        const vol = params.find((p) => p.seriesType === 'bar')
        return `<div style="min-width:158px">
          <div style="font-weight:700;margin-bottom:6px">${k.axisValue}</div>
          ${row('开', o)}${row('高', h, t.upColor)}${row('低', l, t.downColor)}${row('收', c, col)}
          <div style="display:flex;justify-content:space-between;gap:20px;margin-top:4px">
            <span style="opacity:.65">涨跌</span>
            <span style="font-weight:700;color:${col}">${pct}%</span>
          </div>
          ${vol ? row('量', vol.value) : ''}
        </div>`
      },
    },
    axisPointer: { link: [{ xAxisIndex: 'all' }] },
    xAxis: [
      {
        type: 'category',
        data: times,
        gridIndex: 0,
        boundaryGap: false,
        axisLine: { lineStyle: { color: t.lineColor } },
        axisTick: { show: false },
        axisLabel: { color: t.textColor, fontSize: 10, hideOverlap: true },
        splitLine: { show: false },
      },
      {
        type: 'category',
        data: times,
        gridIndex: 1,
        axisLine: { lineStyle: { color: t.lineColor } },
        axisTick: { show: false },
        axisLabel: { show: false },
        splitLine: { show: false },
      },
    ],
    yAxis: [
      {
        scale: true,
        gridIndex: 0,
        position: 'right',
        axisLine: { show: false },
        axisTick: { show: false },
        splitLine: { lineStyle: { color: t.splitColor } },
        axisLabel: {
          color: t.textColor,
          fontSize: 10,
          formatter: (v) => (v >= 1000 ? (v / 1000).toFixed(1) + 'k' : v.toFixed(2)),
        },
      },
      {
        scale: true,
        gridIndex: 1,
        position: 'right',
        axisLine: { show: false },
        axisTick: { show: false },
        splitLine: { show: false },
        axisLabel: { show: false },
      },
    ],
    dataZoom: [
      { type: 'inside', xAxisIndex: [0, 1], start: 45, end: 100 },
      {
        type: 'slider',
        xAxisIndex: [0, 1],
        height: 16,
        bottom: 2,
        borderColor: 'transparent',
        backgroundColor: 'transparent',
        fillerColor: 'rgba(59,130,246,0.16)',
        handleStyle: { color: t.primary },
        moveHandleStyle: { color: t.primary },
        dataBackground: { lineStyle: { color: t.lineColor }, areaStyle: { color: t.splitColor } },
        selectedDataBackground: {
          lineStyle: { color: t.primary },
          areaStyle: { color: 'rgba(59,130,246,0.2)' },
        },
        textStyle: { color: t.textColor, fontSize: 10 },
      },
    ],
    series: [
      {
        name: symbol,
        type: 'candlestick',
        data: ohlc,
        itemStyle: {
          color: t.upColor,
          color0: t.downColor,
          borderColor: t.upColor,
          borderColor0: t.downColor,
        },
      },
      { name: '成交量', type: 'bar', xAxisIndex: 1, yAxisIndex: 1, data: vols, barWidth: '62%' },
    ],
  }
}

const resizeChart = () => {
  if (klineChart) klineChart.resize()
}

const renderChart = () => {
  if (!klineChart) return
  klineChart.setOption(buildKlineOption(klineData.value, orderForm.value.symbol), true)
}

const initKlineChart = () => {
  const dom = document.getElementById('kline')
  if (!dom) return
  if (klineChart) klineChart.dispose()
  klineChart = echarts.init(dom)
  window.addEventListener('resize', resizeChart)
  loadKlineData()
}

// 实时推送时合并最新一根 K 线（节流渲染，避免高频重绘）
let renderTimer = null
const scheduleRender = () => {
  if (renderTimer) return
  renderTimer = setTimeout(() => {
    renderTimer = null
    renderChart()
  }, 500)
}

const updateKlineWithRealData = (k) => {
  if (!k || !klineChart) return
  const close = Number(k.close)
  if (Number.isNaN(close)) return
  const label = fmtAxisTime(toMs(k.close_time ?? k.time ?? Date.now()), currentKlinePeriod.value)
  const candle = [
    label,
    Number(k.open) || close,
    close,
    Number(k.low) || close,
    Number(k.high) || close,
    Number(k.volume) || 0,
  ]
  const arr = klineData.value
  if (arr.length && arr[arr.length - 1][0] === label) {
    arr[arr.length - 1] = candle
  } else {
    arr.push(candle)
    if (arr.length > KLINE_MAX) arr.shift()
  }
  lastCandle.value = candle
  scheduleRender()
}

// 处理WebSocket实时数据
const handleWebSocketData = (dataType, data) => {
  switch (dataType) {
    case 'orderbook':
      orderbook.value = {
        bids: (data.bids || []).map(item => ({ price: item[0], amount: item[1] })),
        asks: (data.asks || []).map(item => ({ price: item[0], amount: item[1] }))
      }
      break

    case 'trade':
      // 添加新成交记录到列表顶部
      const newTrade = {
        created_at: new Date(data.timestamp * 1000).toISOString(),
        price: data.price,
        amount: data.amount,
        side: data.side
      }
      trades.value.unshift(newTrade)

      // 保持最多50条记录
      if (trades.value.length > 50) {
        trades.value = trades.value.slice(0, 50)
      }
      break

    case 'price_change':
      priceChange.value = data
      break

    case 'kline':
      // 更新K线图数据
      updateKlineWithRealData(data)
      break
  }
}

const fetchOrderbook = async () => {
  try {
    const { data } = await getOrderbook(orderForm.value.symbol)
    // 后端异常时返回 { code: 500, data: null } 且 HTTP 仍为 200，
    // 因此必须校验 code / data，否则会把 orderbook 置为 null 导致模板渲染崩溃
    const book = data?.data
    if (data?.code !== 0 || !book) {
      console.warn('订单簿数据不可用:', data?.message || 'empty data')
      orderbook.value = { bids: [], asks: [] }
      return
    }
    orderbook.value = {
      bids: Array.isArray(book.bids) ? book.bids : [],
      asks: Array.isArray(book.asks) ? book.asks : []
    }
  } catch (error) {
    console.error('获取订单簿失败:', error)
    orderbook.value = { bids: [], asks: [] }
  }
}

const fetchTrades = async () => {
  try {
    const { data } = await getTrades(orderForm.value.symbol)
    const list = data?.data
    if (data?.code !== 0 || !Array.isArray(list)) {
      console.warn('成交记录数据不可用:', data?.message || 'empty data')
      trades.value = []
      return
    }
    trades.value = list
  } catch (error) {
    console.error('获取交易记录失败:', error)
    trades.value = []
  }
}

// K线周期切换
const switchKlinePeriod = (period) => {
  currentKlinePeriod.value = period
  loadKlineData()
}

// 加载K线数据（优先后端接口，失败回退本地生成）
const loadKlineData = async () => {
  const symbol = orderForm.value.symbol
  const period = currentKlinePeriod.value
  let data = []

  try {
    const res = await getKline({ symbol, period, limit: 120 })
    const list = res?.data?.data
    if (res?.data?.code === 0 && Array.isArray(list) && list.length) {
      data = list
        .map((k) => [
          fmtAxisTime(toMs(k.time ?? k.open_time ?? k.timestamp ?? k.ts), period),
          Number(k.open),
          Number(k.close),
          Number(k.low),
          Number(k.high),
          Number(k.volume),
        ])
        .filter((r) => r.slice(1).every((n) => !Number.isNaN(n)))
    }
  } catch (error) {
    /* 后端未提供 K 线接口时回退到本地生成 */
  }

  if (!data.length) {
    const base = orderbook.value?.bids?.[0]?.price || trades.value?.[0]?.price || 45000
    data = buildFallbackKline(period, base)
  }

  klineData.value = data
  lastCandle.value = data[data.length - 1] || null
  renderChart()
}

const placeOrder = async () => {
  // 验证输入
  if (!orderForm.value.amount || parseFloat(orderForm.value.amount) <= 0) {
    ElMessage.warning('请输入数量')
    return
  }

  if ((orderForm.value.type === 'limit' || orderForm.value.type === 'stop_limit') && 
      (!orderForm.value.price || parseFloat(orderForm.value.price) <= 0)) {
    ElMessage.warning('请输入价格')
    return
  }

  // 止损止盈订单验证
  if ((orderForm.value.type === 'stop_limit' || orderForm.value.type === 'stop_market') && 
      (!orderForm.value.stop_price || parseFloat(orderForm.value.stop_price) <= 0)) {
    ElMessage.warning('请输入触发价格')
    return
  }

  // 市价单和止损市价单不需要价格验证
  if (orderForm.value.type === 'market' || orderForm.value.type === 'stop_market') {
    // 这些类型不需要价格，但需要其他验证
  }

  // 构建下单请求
  const orderPayload = {
    symbol: orderForm.value.symbol,
    side: orderForm.value.side,
    type: orderForm.value.type,
    amount: orderForm.value.amount
  }

  // 只有限价单和止损限价单需要价格
  if (orderForm.value.type === 'limit' || orderForm.value.type === 'stop_limit') {
    orderPayload.price = orderForm.value.price
  }

  // 止损订单需要触发价格
  if (orderForm.value.type === 'stop_limit' || orderForm.value.type === 'stop_market') {
    orderPayload.stop_price = orderForm.value.stop_price
  }

  try {
    const response = await apiPlaceOrder(orderPayload)

    console.log('下单响应:', response)
    console.log('响应数据:', response.data)

    // 检查响应的 code 字段
    if (response.data.code == 1) {
      // 后端返回了业务错误
      console.log('错误信息:', response.data.message)
      ElMessage.error(response.data.message || '下单失败')
      return
    }

    fetchOrderbook()
    fetchTrades()
    ElMessage.success('下单成功')
  } catch (error) {
    console.error('下单失败:', error)
    console.log('错误响应:', error.response)

    // 优先获取后端返回的 message
    let message = '下单失败'
    if (error.response?.data) {
      message = error.response.data.message || error.response.data.error || message
    }

    ElMessage.error(message)
  }
}

// 格式化价格（保留2位小数）
const formatPrice = (price) => {
  if (!price || price === '' || isNaN(price)) return '--'
  const num = parseFloat(price)
  return isNaN(num) ? '--' : num.toFixed(2)
}

// 格式化数量（保留4位小数）
const formatAmount = (amount) => {
  if (!amount || amount === '' || isNaN(amount)) return '--'
  const num = parseFloat(amount)
  return isNaN(num) ? '--' : num.toFixed(4)
}

// 格式化总额
const formatTotal = (price, amount) => {
  if (!price || !amount || isNaN(price) || isNaN(amount)) return '--'
  const total = parseFloat(price) * parseFloat(amount)
  return isNaN(total) ? '--' : total.toFixed(2)
}

// 点击买单价格，自动填充到表单
const handleBidClick = (row) => {
  if (orderForm.value.type === 'limit' || orderForm.value.type === 'stop_limit') {
    orderForm.value.price = row.price
    orderForm.value.side = 'sell' // 点击买单，说明用户想卖
    ElMessage.info(`已选择卖出价格: ${row.price}`)
  }
  // 止损止盈单的触发价格
  if (orderForm.value.type === 'stop_limit' || orderForm.value.type === 'stop_market') {
    orderForm.value.stop_price = row.price
    ElMessage.info(`已设置触发价格: ${row.price}`)
  }
}

// 点击卖单价格，自动填充到表单
const handleAskClick = (row) => {
  if (orderForm.value.type === 'limit' || orderForm.value.type === 'stop_limit') {
    orderForm.value.price = row.price
    orderForm.value.side = 'buy' // 点击卖单，说明用户想买
    ElMessage.info(`已选择买入价格: ${row.price}`)
  }
  // 止损止盈单的触发价格
  if (orderForm.value.type === 'stop_limit' || orderForm.value.type === 'stop_market') {
    orderForm.value.stop_price = row.price
    ElMessage.info(`已设置触发价格: ${row.price}`)
  }
}

onMounted(async () => {
  // 连接WebSocket
  wsClient.connect()
  wsClient.startHeartbeat()

  await fetchOrderbook()
  await fetchTrades()

  // 等待DOM渲染完成后初始化K线图
  await nextTick()
  initKlineChart()

  // 订阅实时行情数据
  wsClient.subscribe(currentSymbol, handleWebSocketData)
})

watch(() => orderForm.value.symbol, (newSymbol) => {
  // 取消之前的订阅
  wsClient.unsubscribe(currentSymbol, handleWebSocketData)

  // 更新当前交易对
  currentSymbol = newSymbol

  // 重新获取数据
  fetchOrderbook()
  fetchTrades()
  loadKlineData()

  // 订阅新的交易对
  wsClient.subscribe(currentSymbol, handleWebSocketData)
})

// 主题切换时重绘图表，保持配色与界面一致
watch(isDark, async () => {
  await nextTick()
  renderChart()
})

// 组件卸载时销毁图表
onUnmounted(() => {
  window.removeEventListener('resize', resizeChart)
  clearTimeout(renderTimer)
  if (klineChart) {
    klineChart.dispose()
    klineChart = null
  }

  // 取消WebSocket订阅
  wsClient.unsubscribe(currentSymbol, handleWebSocketData)
})
</script>

<style scoped>
.trade-layout {
  height: 100%;
}

.trade-sidebar {
  height: 100%;
  overflow-y: auto;
  padding-bottom: 24px;
}

.trade-main {
  height: 100%;
  overflow-y: auto;
  padding-bottom: 24px;
}

/* Hide scrollbar for webkit, keep it available for firefox */
.trade-sidebar::-webkit-scrollbar,
.trade-main::-webkit-scrollbar {
  display: none;
}
.trade-sidebar,
.trade-main {
  -ms-overflow-style: none;
  scrollbar-width: none;
}

.trade-card {
  border: 1px solid color-mix(in srgb, var(--border-color-hover), transparent 60%);
}

.trade-form :global(.el-form-item__label) {
  color: var(--text-muted);
  padding-bottom: 6px;
}

.trade-form :global(.el-form-item) {
  margin-bottom: 18px;
}

/* 买卖切换开关 */
.trade-type-switch {
  display: flex;
  background: color-mix(in srgb, var(--border-color-hover), transparent 88%);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-base);
  padding: 4px;
  margin-bottom: 6px;
  width: 100%;
  gap: 6px;
}

.switch-item {
  flex: 1;
  text-align: center;
  padding: 10px 0;
  cursor: pointer;
  border-radius: calc(var(--radius-base) - 4px);
  color: var(--text-muted);
  font-weight: 700;
  transition: var(--transition-snap);
  display: inline-flex;
  justify-content: center;
  align-items: center;
  gap: 6px;
}

.switch-item:hover {
  color: var(--text-main);
  background: color-mix(in srgb, var(--border-color-hover), transparent 85%);
}

.switch-item.active.buy-active {
  background: var(--up);
  color: #fff;
  box-shadow: 0 4px 12px var(--up-glow);
}

.switch-item.active.sell-active {
  background: var(--down);
  color: #fff;
  box-shadow: 0 4px 12px var(--down-glow);
}

/* 输入框样式优化 */
.custom-input :global(.el-input__wrapper),
.custom-select :global(.el-select__wrapper) {
  background-color: color-mix(in srgb, var(--bg-elevated), transparent 75%) !important;
  box-shadow: 0 0 0 1px var(--border-color) inset !important;
  border: 1px solid transparent;
  border-radius: var(--radius-input) !important;
  transition: var(--transition-snap);
}

.custom-input :global(.el-input__wrapper):hover,
.custom-input :global(.el-input__wrapper.is-focus),
.custom-select :global(.el-select__wrapper):hover,
.custom-select :global(.el-select__wrapper.is-focus) {
  border-color: var(--primary);
  box-shadow: 0 0 0 1px var(--primary) inset !important;
  background-color: color-mix(in srgb, var(--bg-elevated), transparent 60%) !important;
}

.custom-input :deep(.el-input__inner) {
  color: var(--text-main);
  font-weight: 500;
}

.custom-input :deep(.el-input__suffix-inner) {
  color: var(--text-muted);
}

/* 交易按钮 */
.trade-btn {
  width: 100% !important;
  height: 48px !important;
  font-size: 16px !important;
  font-weight: 700 !important;
  border: none !important;
  border-radius: var(--radius-base) !important;
  transition: var(--transition-snap) !important;
  display: block !important;
}

.btn-buy {
  background: var(--up-grad) !important;
  color: #fff !important;
  box-shadow: 0 4px 16px var(--up-glow);
}

.btn-buy:hover {
  transform: translateY(-2px) !important;
  box-shadow: 0 8px 20px var(--up-glow);
}

.btn-sell {
  background: var(--down-grad) !important;
  color: #fff !important;
  box-shadow: 0 4px 16px var(--down-glow);
}

.btn-sell:hover {
  transform: translateY(-2px) !important;
  box-shadow: 0 8px 20px var(--down-glow);
}

/* 交易额估算 */
.trade-total {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 14px;
  border-radius: var(--radius-base);
  background: color-mix(in srgb, var(--bg-elevated), transparent 78%);
  border: 1px solid var(--border-color);
  color: var(--text-muted);
  font-size: 13px;
}

.total-value {
  color: var(--text-main);
  font-weight: 700;
}

/* 订单簿样式 */
.book-header {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 10px 12px;
  font-weight: 700;
  font-size: 13px;
  border-bottom: 1px solid var(--border-color);
  margin-bottom: 8px;
  background: color-mix(in srgb, var(--bg-elevated), transparent 85%);
  border-radius: var(--radius-base) var(--radius-base) 0 0;
}

.book-header-sell {
  border-top: 2px solid var(--down-glow);
}
.book-header-buy {
  border-top: 2px solid var(--up-glow);
}

.order-book-table,
.trades-table {
  background: transparent !important;
}

.order-book-table :deep(tr), .trades-table :deep(tr),
.order-book-table :deep(th), .trades-table :deep(th) {
  background: transparent !important;
}

.order-book-table :deep(th) {
  background: color-mix(in srgb, var(--bg-elevated), transparent 80%) !important;
  font-weight: 700 !important;
  color: var(--text-muted) !important;
  font-size: 0.75rem !important;
  padding: 8px 0 !important;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.order-book-table :deep(td), .trades-table :deep(td) {
  border-bottom: none !important;
  padding: 7px 0 !important;
}

.order-book-table :deep(tbody tr:hover) {
  background: color-mix(in srgb, var(--primary), transparent 94%) !important;
  cursor: pointer;
}

.order-book-table :deep(.el-table__row):hover {
  background: color-mix(in srgb, var(--primary), transparent 94%) !important;
}

.price-text {
  font-family: 'Roboto Mono', 'Fira Code', monospace;
  font-weight: 600;
  font-size: 13px;
}

.price-text.clickable {
  cursor: pointer;
  transition: var(--transition-snap);
}

.price-text.clickable:hover {
  opacity: 0.8;
  text-decoration: underline;
}

.amount-text {
  color: var(--text-muted);
  font-family: 'Roboto Mono', 'Fira Code', monospace;
  font-size: 12px;
}

.total-text {
  color: var(--text-main);
  font-family: 'Roboto Mono', 'Fira Code', monospace;
  font-size: 12px;
  font-weight: 500;
}

.text-up { color: var(--up); }
.text-down { color: var(--down); }
.text-muted { color: var(--text-muted); }

.time-text {
  font-size: 11px;
  font-family: 'Roboto Mono', 'Fira Code', monospace;
}

.glass-tag {
  background: color-mix(in srgb, var(--bg-elevated), transparent 75%);
  border: 1px solid var(--border-color);
  color: var(--text-muted);
}

.time-intervals .el-button {
  color: var(--text-muted);
}

.time-intervals .el-button:hover,
.time-intervals .el-button.active {
  color: var(--primary);
  background: color-mix(in srgb, var(--bg-elevated), transparent 78%);
}

#kline {
  width: 100%;
  height: 340px;
}

/* K线图卡片 */
.chart-card :deep(.el-card__body) { padding: 12px 14px 14px !important; }

.ohlc-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px 18px;
  padding: 4px 6px 10px;
  border-bottom: 1px solid var(--border-subtle);
  margin-bottom: 6px;
}
.ohlc-item { display: inline-flex; align-items: center; gap: 6px; }
.ohlc-label { font-size: 0.72rem; color: var(--text-secondary); font-weight: 600; }
.ohlc-value { font-size: 0.8rem; font-weight: 650; color: var(--text-main); }

/* 止损止盈提示 */
.stop-hint {
  font-size: 11px;
  color: var(--text-muted);
  margin-top: 4px;
  padding-left: 4px;
}
</style>
