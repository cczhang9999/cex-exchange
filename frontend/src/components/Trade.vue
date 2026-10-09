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
      <el-card class="glass-panel">
        <template #header>
          <div class="flex-between">
            <h3 class="card-title">Chart</h3>
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
        <div id="kline"></div>
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
              <el-icon class="text-success"><ArrowUp /></el-icon>
              <span class="text-success">Bids (Buy)</span>
            </div>
            <el-table
              :data="orderbook.bids"
              size="small"
              height="300"
              :show-header="true"
              class="order-book-table"
              @row-click="handleBidClick"
            >
              <el-table-column label="Price (USDT)" width="120" align="left">
                <template #default="{ row }">
                  <span class="text-success price-text clickable mono">{{ formatPrice(row.price) }}</span>
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
              <el-icon class="text-danger"><ArrowDown /></el-icon>
              <span class="text-danger">Asks (Sell)</span>
            </div>
            <el-table
              :data="orderbook.asks"
              size="small"
              height="300"
              :show-header="true"
              class="order-book-table"
              @row-click="handleAskClick"
            >
              <el-table-column label="Price (USDT)" width="120" align="left">
                <template #default="{ row }">
                  <span class="text-danger price-text clickable mono">{{ formatPrice(row.price) }}</span>
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
            <el-tag size="small" type="success" class="glass-tag">{{ trades.length }} trades</el-tag>
          </div>
        </template>
        <el-table :data="trades" size="small" height="200" class="trades-table">
          <el-table-column label="Time" width="100">
            <template #default="scope">
              <span class="text-muted time-text mono">{{ formatTime(scope.row.created_at) }}</span>
            </template>
          </el-table-column>
          <el-table-column label="Side" width="80" align="center">
            <template #default="{ row }">
              <el-tag
                :type="row.side === 'buy' ? 'success' : 'danger'"
                size="small"
                effect="dark"
              >
                {{ row.side === 'buy' ? '买入' : '卖出' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="Price (USDT)" width="120" align="right">
            <template #default="{ row }">
              <span :class="row.side === 'buy' ? 'text-success' : 'text-danger'" class="price-text mono">
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
import { ref, onMounted, watch, nextTick, onUnmounted } from 'vue'
import { placeOrder as apiPlaceOrder, getOrderbook, getTrades } from '../api/api.js'
import * as echarts from 'echarts'
import wsClient from '../utils/websocket.js'
import MarketData from './MarketData.vue'
import { ElMessage } from 'element-plus'
import { ArrowUp, ArrowDown } from '@element-plus/icons-vue'

const availableSymbols = ref(['BTC/USDT', 'ETH/USDT', 'BNB/USDT', 'SOL/USDT', 'XRP/USDT'])
const orderForm = ref({ symbol: 'BTC/USDT', side: 'buy', type: 'limit', price: '', amount: '', stop_price: '' })
const orderbook = ref({ bids: [], asks: [] })
const trades = ref([])
const priceChange = ref({})
let klineChart = null
let currentSymbol = 'BTC/USDT'

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

// 初始化K线图
const initKlineChart = () => {
  const chartDom = document.getElementById('kline')
  if (!chartDom) return

  klineChart = echarts.init(chartDom)

  // 模拟K线数据
  const klineData = [
    ['2024-01-15 10:00', 45100, 45150, 45050, 45120, 100],
    ['2024-01-15 10:01', 45120, 45180, 45100, 45160, 150],
    ['2024-01-15 10:02', 45160, 45200, 45140, 45180, 120],
    ['2024-01-15 10:03', 45180, 45220, 45160, 45200, 180],
    ['2024-01-15 10:04', 45200, 45250, 45180, 45230, 200],
    ['2024-01-15 10:05', 45230, 45280, 45200, 45250, 160],
    ['2024-01-15 10:06', 45250, 45300, 45220, 45280, 140],
    ['2024-01-15 10:07', 45280, 45320, 45250, 45300, 170],
    ['2024-01-15 10:08', 45300, 45350, 45280, 45320, 190],
    ['2024-01-15 10:09', 45320, 45380, 45300, 45350, 220],
    ['2024-01-15 10:10', 45350, 45400, 45320, 45380, 180]
  ]

  const option = {
    backgroundColor: 'transparent',
    title: {
      text: 'BTC/USDT',
      left: 'center',
      textStyle: {
        color: 'var(--text-muted)'
      }
    },
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'cross'
      },
      backgroundColor: 'rgba(0,0,0,0.8)',
      borderColor: '#333',
      textStyle: {
        color: '#eee'
      },
      formatter: function (params) {
        const data = params[0].data
        return `Time: ${params[0].axisValue}<br/>
                Open: ${data[1]}<br/>
                Close: ${data[2]}<br/>
                Low: ${data[3]}<br/>
                High: ${data[4]}<br/>
                Vol: ${data[5]}`
      }
    },
    grid: {
      left: '10%',
      right: '10%',
      bottom: '15%',
      top: '15%'
    },
    xAxis: {
      type: 'category',
      data: klineData.map(item => item[0]),
      scale: true,
      boundaryGap: false,
      axisLine: { lineStyle: { color: '#555' } },
      splitLine: { show: false },
      axisLabel: { color: '#888' }
    },
    yAxis: {
      scale: true,
      splitArea: { show: false },
      axisLine: { lineStyle: { color: '#555' } },
      splitLine: { lineStyle: { color: '#333' } },
      axisLabel: { color: '#888' }
    },
    dataZoom: [
      {
        type: 'inside',
        start: 50,
        end: 100
      }
    ],
    series: [
      {
        name: 'BTC/USDT',
        type: 'candlestick',
        data: klineData.map(item => item.slice(1)),
        itemStyle: {
          color: '#0ecb81',
          color0: '#f6465d',
          borderColor: '#0ecb81',
          borderColor0: '#f6465d'
        }
      }
    ]
  }

  klineChart.setOption(option)
}

// 更新K线图
const updateKlineChart = () => {
  if (klineChart) {
    // 这里可以添加实时更新K线数据的逻辑
    klineChart.resize()
  }
}

// 使用实时数据更新K线图
const updateKlineWithRealData = (klineData) => {
  if (!klineChart) return

  const option = klineChart.getOption()
  const series = option.series[0]

  // 添加新的K线数据点
  const newDataPoint = [
    new Date(klineData.close_time * 1000).toLocaleString(),
    parseFloat(klineData.open),
    parseFloat(klineData.close),
    parseFloat(klineData.low),
    parseFloat(klineData.high),
    parseFloat(klineData.volume)
  ]

  // 更新数据
  series.data.push(newDataPoint)

  // 保持最多100个数据点
  if (series.data.length > 100) {
    series.data = series.data.slice(-100)
  }

  // 更新x轴数据
  option.xAxis[0].data = series.data.map(item => item[0])

  klineChart.setOption(option)
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
    orderbook.value = data.data
  } catch (error) {
    console.error('获取订单簿失败:', error)
  }
}

const fetchTrades = async () => {
  try {
    const { data } = await getTrades(orderForm.value.symbol)
    trades.value = data.data
  } catch (error) {
    console.error('获取交易记录失败:', error)
  }
}

// K线周期切换
const switchKlinePeriod = (period) => {
  currentKlinePeriod.value = period
  // 重新加载K线数据
  loadKlineData()
}

// 加载K线数据（根据当前周期）
const loadKlineData = () => {
  if (klineChart) {
    // 根据选择的周期获取不同的数据
    // 实际项目中应该调用API获取对应周期的K线数据
    // 这里使用模拟数据
    const baseTime = new Date('2024-01-15 10:00').getTime()
    const periodMs = klinePeriods.find(p => p.value === currentKlinePeriod.value)?.interval || 60000
    
    const klineData = []
    for (let i = 0; i < 50; i++) {
      const time = new Date(baseTime + i * periodMs)
      const basePrice = 45000 + Math.random() * 500
      klineData.push([
        time.toLocaleString('zh-CN', { hour: '2-digit', minute: '2-digit' }),
        basePrice,
        basePrice + Math.random() * 100,
        basePrice - Math.random() * 100,
        basePrice + (Math.random() - 0.5) * 50,
        100 + Math.random() * 200
      ])
    }
    
    klineChart.setOption({
      xAxis: { data: klineData.map(item => item[0]) },
      series: [{ data: klineData.map(item => item.slice(1)) }]
    })
  }
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
  wsClient.unsubscribe(currentSymbol)

  // 更新当前交易对
  currentSymbol = newSymbol

  // 重新获取数据
  fetchOrderbook()
  fetchTrades()
  updateKlineChart()

  // 订阅新的交易对
  wsClient.subscribe(currentSymbol, handleWebSocketData)
})

// 组件卸载时销毁图表
onUnmounted(() => {
  if (klineChart) {
    klineChart.dispose()
  }

  // 取消WebSocket订阅
  wsClient.unsubscribe(currentSymbol)
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
  background: var(--success);
  color: #1e293b;
  box-shadow: 0 4px 12px var(--success-glow);
}

.switch-item.active.sell-active {
  background: var(--danger);
  color: #fff;
  box-shadow: 0 4px 12px var(--danger-glow);
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
  background: linear-gradient(135deg, var(--success) 0%, var(--success-emphasis) 100%) !important;
  color: #1e293b !important;
  box-shadow: 0 4px 16px var(--success-glow);
}

.btn-buy:hover {
  transform: translateY(-2px) !important;
  box-shadow: 0 8px 20px var(--success-glow);
}

.btn-sell {
  background: linear-gradient(135deg, var(--danger) 0%, var(--danger-emphasis) 100%) !important;
  color: #fff !important;
  box-shadow: 0 4px 16px var(--danger-glow);
}

.btn-sell:hover {
  transform: translateY(-2px) !important;
  box-shadow: 0 8px 20px var(--danger-glow);
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
  border-top: 2px solid var(--danger-glow);
}
.book-header-buy {
  border-top: 2px solid var(--success-glow);
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

.text-success { color: var(--success); }
.text-danger { color: var(--danger); }
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
  height: 300px;
}

/* 止损止盈提示 */
.stop-hint {
  font-size: 11px;
  color: var(--text-muted);
  margin-top: 4px;
  padding-left: 4px;
}
</style>
