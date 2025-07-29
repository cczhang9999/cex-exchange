<template>
  <el-row :gutter="20" style="height: 100%;">
    <el-col :span="8">
      <!-- 实时行情显示 -->
      <MarketData :symbol="orderForm.symbol" />
      
      <el-card>
        <h3>下单</h3>
        <el-form :model="orderForm" label-width="80px">
          <el-form-item label="交易对">
            <el-input v-model="orderForm.symbol" />
          </el-form-item>
          <el-form-item label="方向">
            <el-select v-model="orderForm.side">
              <el-option label="买入" value="buy" />
              <el-option label="卖出" value="sell" />
            </el-select>
          </el-form-item>
          <el-form-item label="类型">
            <el-select v-model="orderForm.type">
              <el-option label="限价单" value="limit" />
              <el-option label="市价单" value="market" />
            </el-select>
          </el-form-item>
          <el-form-item label="价格" v-if="orderForm.type === 'limit'">
            <el-input v-model="orderForm.price" />
          </el-form-item>
          <el-form-item label="数量">
            <el-input v-model="orderForm.amount" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="placeOrder">下单</el-button>
          </el-form-item>
        </el-form>
      </el-card>
      
      <!-- K线图 -->
      <el-card style="margin-top: 20px;">
        <h3>K线图</h3>
        <div id="kline" style="height: 300px;"></div>
      </el-card>
    </el-col>
    <el-col :span="16">
      <el-card style="margin-bottom: 20px;">
        <h3>订单簿</h3>
        <el-row :gutter="10">
          <el-col :span="12">
            <div style="font-weight:bold; color:#67C23A; margin-bottom: 5px;">买盘</div>
            <el-table :data="orderbook.bids" size="small" height="300">
              <el-table-column prop="price" label="买价" />
              <el-table-column prop="amount" label="数量" />
            </el-table>
          </el-col>
          <el-col :span="12">
            <div style="font-weight:bold; color:#F56C6C; margin-bottom: 5px;">卖盘</div>
            <el-table :data="orderbook.asks" size="small" height="300">
              <el-table-column prop="price" label="卖价" />
              <el-table-column prop="amount" label="数量" />
            </el-table>
          </el-col>
        </el-row>
      </el-card>
      
      <el-card>
        <h3>成交记录</h3>
        <el-table :data="trades" size="small" height="200">
          <el-table-column prop="created_at" label="时间" width="120">
            <template #default="scope">
              {{ formatTime(scope.row.created_at) }}
            </template>
          </el-table-column>
          <el-table-column prop="price" label="价格" width="100" />
          <el-table-column prop="amount" label="数量" width="100" />
          <el-table-column prop="side" label="方向" width="80">
            <template #default="scope">
              <el-tag :type="scope.row.side === 'buy' ? 'success' : 'danger'" size="small">
                {{ scope.row.side === 'buy' ? '买入' : '卖出' }}
              </el-tag>
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
    // 使用模拟数据
    orderbook.value = {
      bids: [
        { price: '45,100', amount: '0.5' },
        { price: '45,050', amount: '1.2' },
        { price: '45,000', amount: '0.8' },
        { price: '44,950', amount: '0.3' },
        { price: '44,900', amount: '1.5' }
      ],
      asks: [
        { price: '45,150', amount: '0.3' },
        { price: '45,200', amount: '0.9' },
        { price: '45,250', amount: '1.1' },
        { price: '45,300', amount: '0.7' },
        { price: '45,350', amount: '0.4' }
      ]
    }
  }
}

const fetchTrades = async () => {
  try {
    const { data } = await getTrades(orderForm.value.symbol)
    trades.value = data
  } catch (error) {
    console.error('获取交易记录失败:', error)
    // 使用模拟数据，包含 created_at 字段
    const now = new Date()
    trades.value = [
      { created_at: new Date(now.getTime() - 1000 * 60 * 5).toISOString(), price: '45,120', amount: '0.1', side: 'buy' },
      { created_at: new Date(now.getTime() - 1000 * 60 * 4).toISOString(), price: '45,110', amount: '0.2', side: 'sell' },
      { created_at: new Date(now.getTime() - 1000 * 60 * 3).toISOString(), price: '45,130', amount: '0.05', side: 'buy' },
      { created_at: new Date(now.getTime() - 1000 * 60 * 2).toISOString(), price: '45,125', amount: '0.15', side: 'sell' },
      { created_at: new Date(now.getTime() - 1000 * 60 * 1).toISOString(), price: '45,115', amount: '0.08', side: 'buy' },
      { created_at: new Date(now.getTime() - 1000 * 30).toISOString(), price: '45,135', amount: '0.12', side: 'sell' },
      { created_at: new Date(now.getTime() - 1000 * 20).toISOString(), price: '45,105', amount: '0.25', side: 'buy' },
      { created_at: new Date(now.getTime() - 1000 * 10).toISOString(), price: '45,140', amount: '0.06', side: 'sell' },
      { created_at: new Date(now.getTime() - 1000 * 5).toISOString(), price: '45,118', amount: '0.18', side: 'buy' },
      { created_at: new Date(now.getTime() - 1000 * 1).toISOString(), price: '45,128', amount: '0.09', side: 'sell' }
    ]
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
#kline { 
  width: 100%; 
  height: 300px;
}
</style> 