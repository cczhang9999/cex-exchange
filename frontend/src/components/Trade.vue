<template>
  <el-row :gutter="20" style="height: 100%;">
    <el-col :span="8" class="scrollable-col">
      <!-- 实时行情显示 -->
      <MarketData :symbol="orderForm.symbol" />
      
      <el-card class="glass-panel mb-20 trade-card">
        <template #header>
          <div class="flex-between">
            <h3 class="card-title">Place Order</h3>
            <el-tag size="small" effect="dark" type="info" class="glass-tag">Spot</el-tag>
          </div>
        </template>
        <el-form :model="orderForm" label-position="top" class="trade-form">
          <el-form-item label="Pair">
            <el-select v-model="orderForm.symbol" class="custom-select" style="width: 100%" popper-class="custom-dropdown">
              <el-option v-for="symbol in availableSymbols" :key="symbol" :label="symbol" :value="symbol" />
            </el-select>
          </el-form-item>
          
          <el-form-item class="mb-4">
            <div class="trade-type-switch">
              <div 
                class="switch-item" 
                :class="{ active: orderForm.side === 'buy', 'buy-active': orderForm.side === 'buy' }"
                @click="orderForm.side = 'buy'"
              >
                Buy
              </div>
              <div 
                class="switch-item" 
                :class="{ active: orderForm.side === 'sell', 'sell-active': orderForm.side === 'sell' }"
                @click="orderForm.side = 'sell'"
              >
                Sell
              </div>
            </div>
          </el-form-item>

          <el-form-item label="Type">
            <el-select v-model="orderForm.type" style="width: 100%" class="custom-select" popper-class="custom-dropdown">
              <el-option label="Limit" value="limit" />
              <el-option label="Market" value="market" />
            </el-select>
          </el-form-item>
          
          <el-form-item label="Price" v-if="orderForm.type === 'limit'">
            <el-input v-model="orderForm.price" placeholder="0.00" class="custom-input">
              <template #suffix>USDT</template>
            </el-input>
          </el-form-item>
          
          <el-form-item label="Amount">
            <el-input v-model="orderForm.amount" placeholder="0.00" class="custom-input">
              <template #suffix>{{ orderForm.symbol.split('/')[0] }}</template>
            </el-input>
          </el-form-item>
          
          <!-- 交易额估算 (仅限价单显示) -->
          <div v-if="orderForm.type === 'limit' && orderForm.price && orderForm.amount" class="trade-total mb-4">
            <span>Total</span>
            <span class="total-value">{{ (parseFloat(orderForm.price) * parseFloat(orderForm.amount)).toFixed(2) }} USDT</span>
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
              <el-button size="small" text :class="{ active: true }">1m</el-button>
              <el-button size="small" text>15m</el-button>
              <el-button size="small" text>1h</el-button>
            </div>
          </div>
        </template>
        <div id="kline" style="height: 300px;"></div>
      </el-card>
    </el-col>
    <el-col :span="16" class="scrollable-col">
      <el-card class="glass-panel mb-20">
        <template #header>
          <h3 class="card-title">Order Book</h3>
        </template>
        <el-row :gutter="20">
          <el-col :span="12">
            <div class="book-header text-success">Bids (Buy)</div>
            <el-table :data="orderbook.bids" size="small" height="300" :show-header="false" class="order-book-table">
              <el-table-column prop="price" label="Price" align="left">
                <template #default="{ row }">
                  <span class="text-success price-text">{{ row.price }}</span>
                </template>
              </el-table-column>
              <el-table-column prop="amount" label="Amount" align="right">
                <template #default="{ row }">
                  <span class="amount-text">{{ row.amount }}</span>
                </template>
              </el-table-column>
            </el-table>
          </el-col>
          <el-col :span="12">
            <div class="book-header text-danger">Asks (Sell)</div>
            <el-table :data="orderbook.asks" size="small" height="300" :show-header="false" class="order-book-table">
              <el-table-column prop="price" label="Price" align="left">
                <template #default="{ row }">
                  <span class="text-danger price-text">{{ row.price }}</span>
                </template>
              </el-table-column>
              <el-table-column prop="amount" label="Amount" align="right">
                <template #default="{ row }">
                  <span class="amount-text">{{ row.amount }}</span>
                </template>
              </el-table-column>
            </el-table>
          </el-col>
        </el-row>
      </el-card>
      
      <el-card class="glass-panel">
        <template #header>
          <h3 class="card-title">Recent Trades</h3>
        </template>
        <el-table :data="trades" size="small" height="200" class="trades-table">
          <el-table-column prop="created_at" label="Time" width="120">
            <template #default="scope">
              <span class="text-muted">{{ formatTime(scope.row.created_at) }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="price" label="Price">
            <template #default="{ row }">
              <span :class="row.side === 'buy' ? 'text-success' : 'text-danger'" class="price-text">{{ row.price }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="amount" label="Amount" align="right">
            <template #default="{ row }">
              <span class="amount-text">{{ row.amount }}</span>
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

const availableSymbols = ref(['BTC/USDT', 'ETH/USDT', 'BNB/USDT', 'SOL/USDT', 'XRP/USDT'])
const orderForm = ref({ symbol: 'BTC/USDT', side: 'buy', type: 'limit', price: '', amount: '' })
const orderbook = ref({ bids: [], asks: [] })
const trades = ref([])
const priceChange = ref({})
let klineChart = null
let currentSymbol = 'BTC/USDT'

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
        color: '#e0e0e0'
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
        bids: data.bids.map(item => ({ price: item[0], amount: item[1] })),
        asks: data.asks.map(item => ({ price: item[0], amount: item[1] }))
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

const placeOrder = async () => {
  // 验证输入
  if (!orderForm.value.amount || parseFloat(orderForm.value.amount) <= 0) {
    ElMessage.warning('请输入数量')
    return
  }
  
  if (orderForm.value.type === 'limit' && (!orderForm.value.price || parseFloat(orderForm.value.price) <= 0)) {
    ElMessage.warning('请输入价格')
    return
  }

  try {
    await apiPlaceOrder(orderForm.value)
    fetchOrderbook()
    fetchTrades()
    ElMessage.success('下单成功')
  } catch (error) {
    console.error('下单失败:', error)
    ElMessage.error('下单失败')
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
.mb-20 {
  margin-bottom: 20px;
}
.mb-4 {
  margin-bottom: 16px;
}
.mt-6 {
  margin-top: 24px;
}

.flex-between {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.flex-center {
  display: flex;
  align-items: center;
  gap: 8px;
}

.card-title {
  margin: 0;
  font-weight: 600;
  color: var(--text-main);
  font-size: 16px;
}

/* 交易面板样式 */
.trade-card {
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.trade-form :deep(.el-form-item__label) {
  color: var(--text-muted);
  padding-bottom: 4px;
}

/* 买卖切换开关 */
.trade-type-switch {
  display: flex;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 4px;
  padding: 4px;
  margin-bottom: 20px;
  width: 100%;
  box-sizing: border-box;
  gap: 8px;
}

.switch-item {
  flex: 1;
  text-align: center;
  padding: 8px 0;
  cursor: pointer;
  border-radius: 4px;
  color: var(--text-muted);
  font-weight: 600;
  transition: all 0.3s;
  display: flex;
  justify-content: center;
  align-items: center;
}

.switch-item:hover {
  color: var(--text-main);
}

.switch-item.active.buy-active {
  background: var(--success);
  color: white;
}

.switch-item.active.sell-active {
  background: var(--danger);
  color: white;
}

/* 输入框样式优化 */
.custom-input :deep(.el-input__wrapper),
.custom-select :deep(.el-select__wrapper) {
  background-color: rgba(255, 255, 255, 0.05);
  box-shadow: none !important;
  border: 1px solid rgba(255, 255, 255, 0.1);
  transition: all 0.3s;
}

.custom-input :deep(.el-input__wrapper:hover),
.custom-input :deep(.el-input__wrapper.is-focus),
.custom-select :deep(.el-select__wrapper:hover),
.custom-select :deep(.el-select__wrapper.is-focus) {
  border-color: var(--primary);
  background-color: rgba(255, 255, 255, 0.08);
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
  height: 44px !important;
  font-size: 16px !important;
  font-weight: 600 !important;
  border: none !important;
  border-radius: 4px !important;
  transition: all 0.2s !important;
  display: block !important;
  opacity: 1 !important;
  visibility: visible !important;
}

.btn-buy {
  background: var(--success) !important;
  color: white !important;
}

.btn-buy:hover {
  background: #0abb75 !important; /* 稍微亮一点的绿色 */
  transform: translateY(-1px) !important;
}

.btn-sell {
  background: var(--danger) !important;
  color: white !important;
}

.btn-sell:hover {
  background: #f75569 !important; /* 稍微亮一点的红色 */
  transform: translateY(-1px) !important;
}

/* 交易额估算 */
.trade-total {
  display: flex;
  justify-content: space-between;
  color: var(--text-muted);
  font-size: 12px;
}

.total-value {
  color: var(--text-main);
  font-weight: 500;
}

/* 订单簿样式 */
.book-header {
  padding: 8px 12px;
  font-weight: 600;
  font-size: 12px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  margin-bottom: 4px;
}

.order-book-table, .trades-table {
  background: transparent !important;
}

.order-book-table :deep(tr), .trades-table :deep(tr),
.order-book-table :deep(th), .trades-table :deep(th) {
  background: transparent !important;
}

.order-book-table :deep(td), .trades-table :deep(td) {
  border-bottom: none !important;
  padding: 4px 0 !important;
}

.price-text {
  font-family: 'Roboto Mono', monospace;
  font-weight: 500;
}

.amount-text {
  color: var(--text-muted);
  font-family: 'Roboto Mono', monospace;
}

.text-success { color: var(--success); }
.text-danger { color: var(--danger); }
.text-muted { color: var(--text-muted); }

.glass-tag {
  background: rgba(255, 255, 255, 0.1);
  border: none;
  color: var(--text-muted);
}

.time-intervals .el-button {
  color: var(--text-muted);
}

.time-intervals .el-button:hover,
.time-intervals .el-button.active {
  color: var(--primary);
  background: rgba(255, 255, 255, 0.05);
}

#kline { 
  width: 100%; 
  height: 300px;
}

.scrollable-col {
  height: 100%;
  overflow-y: auto;
  padding-bottom: 20px; /* Add some padding at the bottom */
}

/* Hide scrollbar for Chrome, Safari and Opera */
.scrollable-col::-webkit-scrollbar {
  display: none;
}

/* Hide scrollbar for IE, Edge and Firefox */
.scrollable-col {
  -ms-overflow-style: none;  /* IE and Edge */
  scrollbar-width: none;  /* Firefox */
}
</style>
 