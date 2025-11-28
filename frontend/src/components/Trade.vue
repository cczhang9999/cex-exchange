<template>
  <el-row :gutter="20" style="height: 100%;">
    <el-col :span="8">
      <!-- 实时行情显示 -->
      <MarketData :symbol="orderForm.symbol" />
      
      <el-card class="glass-panel mb-20">
        <template #header>
          <div class="flex-between">
            <h3>Place Order</h3>
            <el-tag size="small" effect="plain" class="glass-tag">Spot</el-tag>
          </div>
        </template>
        <el-form :model="orderForm" label-position="top">
          <el-form-item label="Pair">
            <el-input v-model="orderForm.symbol" prefix-icon="Search" />
          </el-form-item>
          <el-form-item label="Side">
            <el-radio-group v-model="orderForm.side" style="width: 100%">
              <el-radio-button label="buy" class="w-50">Buy</el-radio-button>
              <el-radio-button label="sell" class="w-50">Sell</el-radio-button>
            </el-radio-group>
          </el-form-item>
          <el-form-item label="Type">
            <el-select v-model="orderForm.type" style="width: 100%">
              <el-option label="Limit" value="limit" />
              <el-option label="Market" value="market" />
            </el-select>
          </el-form-item>
          <el-form-item label="Price" v-if="orderForm.type === 'limit'">
            <el-input v-model="orderForm.price" placeholder="0.00">
              <template #append>USDT</template>
            </el-input>
          </el-form-item>
          <el-form-item label="Amount">
            <el-input v-model="orderForm.amount" placeholder="0.00">
              <template #append>{{ orderForm.symbol.split('/')[0] }}</template>
            </el-input>
          </el-form-item>
          <el-form-item>
            <el-button 
              :type="orderForm.side === 'buy' ? 'success' : 'danger'" 
              @click="placeOrder" 
              style="width: 100%; height: 40px; font-size: 16px;"
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
            <h3>Chart</h3>
            <div class="flex-center">
              <el-button size="small" text bg>1m</el-button>
              <el-button size="small" text>15m</el-button>
              <el-button size="small" text>1h</el-button>
            </div>
          </div>
        </template>
        <div id="kline" style="height: 300px;"></div>
      </el-card>
    </el-col>
    <el-col :span="16">
      <el-card class="glass-panel mb-20">
        <template #header>
          <h3>Order Book</h3>
        </template>
        <el-row :gutter="20">
          <el-col :span="12">
            <div class="book-header text-success">Bids</div>
            <el-table :data="orderbook.bids" size="small" height="300" :show-header="false">
              <el-table-column prop="price" label="Price" align="left">
                <template #default="{ row }">
                  <span class="text-success">{{ row.price }}</span>
                </template>
              </el-table-column>
              <el-table-column prop="amount" label="Amount" align="right" />
            </el-table>
          </el-col>
          <el-col :span="12">
            <div class="book-header text-danger">Asks</div>
            <el-table :data="orderbook.asks" size="small" height="300" :show-header="false">
              <el-table-column prop="price" label="Price" align="left">
                <template #default="{ row }">
                  <span class="text-danger">{{ row.price }}</span>
                </template>
              </el-table-column>
              <el-table-column prop="amount" label="Amount" align="right" />
            </el-table>
          </el-col>
        </el-row>
      </el-card>
      
      <el-card class="glass-panel">
        <template #header>
          <h3>Recent Trades</h3>
        </template>
        <el-table :data="trades" size="small" height="200">
          <el-table-column prop="created_at" label="Time" width="120">
            <template #default="scope">
              <span class="text-muted">{{ formatTime(scope.row.created_at) }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="price" label="Price">
            <template #default="{ row }">
              <span :class="row.side === 'buy' ? 'text-success' : 'text-danger'">{{ row.price }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="amount" label="Amount" align="right" />
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
    title: {
      text: 'BTC/USDT',
      left: 'center',
      textStyle: {
        color: '#333'
      }
    },
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'cross'
      },
      formatter: function (params) {
        const data = params[0].data
        return `时间: ${params[0].axisValue}<br/>
                开盘: ${data[1]}<br/>
                收盘: ${data[2]}<br/>
                最低: ${data[3]}<br/>
                最高: ${data[4]}<br/>
                成交量: ${data[5]}`
      }
    },
    grid: {
      left: '10%',
      right: '10%',
      bottom: '15%'
    },
    xAxis: {
      type: 'category',
      data: klineData.map(item => item[0]),
      scale: true,
      boundaryGap: false,
      axisLine: { onZero: false },
      splitLine: { show: false },
      min: 'dataMin',
      max: 'dataMax'
    },
    yAxis: {
      scale: true,
      splitArea: {
        show: true
      }
    },
    dataZoom: [
      {
        type: 'inside',
        start: 50,
        end: 100
      },
      {
        show: true,
        type: 'slider',
        top: '90%',
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
          color: '#FD1050',
          color0: '#0CF49B',
          borderColor: '#FD1050',
          borderColor0: '#0CF49B'
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
    orderbook.value = data
  } catch (error) {
    console.error('获取订单簿失败:', error)
  }
}

const fetchTrades = async () => {
  try {
    const { data } = await getTrades(orderForm.value.symbol)
    trades.value = data
  } catch (error) {
    console.error('获取交易记录失败:', error)
  }
}

const placeOrder = async () => {
  try {
    await apiPlaceOrder(orderForm.value)
    fetchOrderbook()
    fetchTrades()
  } catch (error) {
    console.error('下单失败:', error)
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

h3 {
  margin: 0;
  font-weight: 600;
  color: var(--text-main);
}

.book-header {
  padding: 10px;
  font-weight: 600;
  border-bottom: 1px solid var(--border-color);
  background: rgba(0, 0, 0, 0.2);
}

.text-success { color: var(--success); }
.text-danger { color: var(--danger); }
.text-muted { color: var(--text-muted); }

.glass-tag {
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.2);
  color: var(--text-main);
}

.w-50 {
  width: 50%;
}

:deep(.el-radio-button__inner) {
  background: transparent;
  border: 1px solid var(--border-color);
  color: var(--text-muted);
  width: 100%;
}

:deep(.el-radio-button__original-radio:checked + .el-radio-button__inner) {
  background: var(--primary);
  border-color: var(--primary);
  color: white;
  box-shadow: none;
}

#kline { 
  width: 100%; 
  height: 300px;
}
</style> 