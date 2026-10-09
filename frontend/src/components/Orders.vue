<template>
  <div class="orders-container">
    <!-- 订单筛选器 -->
    <el-card class="glass-panel content-card mb-20">
      <template #header>
        <div class="flex-between">
          <h3 class="page-title-sm">我的订单</h3>
          <el-tag type="info" effect="dark" class="glass-tag">Open Orders</el-tag>
        </div>
      </template>
      
      <!-- 筛选栏 -->
      <div class="filter-bar mb-16">
        <el-input
          v-model="orderFilters.symbol"
          placeholder="交易对"
          class="filter-input"
          clearable
          @keyup.enter="applyOrderFilters"
        >
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
        
        <el-select v-model="orderFilters.side" placeholder="方向" class="filter-select" clearable>
          <el-option label="全部" value="" />
          <el-option label="买入" value="buy" />
          <el-option label="卖出" value="sell" />
        </el-select>
        
        <el-select v-model="orderFilters.type" placeholder="类型" class="filter-select" clearable>
          <el-option label="全部" value="" />
          <el-option label="限价单" value="limit" />
          <el-option label="市价单" value="market" />
          <el-option label="止损限价" value="stop_limit" />
          <el-option label="止损市价" value="stop_market" />
        </el-select>
        
        <el-select v-model="orderFilters.status" placeholder="状态" class="filter-select" clearable>
          <el-option label="全部" value="" />
          <el-option label="进行中" value="open" />
          <el-option label="部分成交" value="partially_filled" />
          <el-option label="已成交" value="filled" />
          <el-option label="已取消" value="cancelled" />
        </el-select>
        
        <el-date-picker
          v-model="orderFilters.dateRange"
          type="daterange"
          range-separator="至"
          start-placeholder="开始日期"
          end-placeholder="结束日期"
          class="filter-date"
          @change="applyOrderFilters"
        />
        
        <el-button type="primary" @click="applyOrderFilters">
          <el-icon><Search /></el-icon> 筛选
        </el-button>
        <el-button @click="resetOrderFilters">重置</el-button>
      </div>

      <el-table 
        :data="filteredOrders" 
        class="orders-table w-100"
        v-loading="ordersLoading"
      >
        <el-table-column prop="symbol" label="交易对" width="120" />
        <el-table-column prop="side" label="方向" width="80">
          <template #default="{ row }">
            <span :class="row.side === 'buy' ? 'text-success' : 'text-danger'">
              {{ row.side === 'buy' ? '买入' : '卖出' }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="type" label="类型" width="100">
          <template #default="{ row }">
            <el-tag size="small" effect="plain" class="glass-tag">{{ getTypeText(row.type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="price" label="价格" width="120">
          <template #default="{ row }">
            <span class="mono">{{ formatPrice(row.price) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="amount" label="数量" width="100">
          <template #default="{ row }">
            <span class="mono">{{ formatAmount(row.amount) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="filled" label="已成交" width="100">
          <template #default="{ row }">
            <span class="mono">{{ formatAmount(row.filled) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)" effect="dark" size="small" class="status-tag">
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="触发价格" width="100" v-if="hasStopOrders">
          <template #default="{ row }">
            <span class="mono text-muted" v-if="row.stop_price">{{ formatPrice(row.stop_price) }}</span>
            <span class="text-muted" v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="时间" width="160">
          <template #default="{ row }">
            <span class="text-muted mono">{{ formatDate(row.created_at) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="110" fixed="right">
          <template #default="{ row }">
            <el-button
              v-if="row.status === 'open' || row.status === 'partially_filled'"
              type="danger"
              size="small"
              plain
              @click="handleCancelOrder(row)"
            >
              撤单
            </el-button>
            <el-button
              v-else
              type="info"
              size="small"
              plain
              disabled
            >
              -
            </el-button>
          </template>
        </el-table-column>
      </el-table>
      
      <!-- 分页 -->
      <el-pagination
        v-model:current-page="orderPage"
        :page-size="orderPageSize"
        :total="filteredOrders.length"
        layout="prev, pager, next"
        class="pagination-center mt-16"
        @current-change="handleOrderPageChange"
      />
    </el-card>

    <!-- 成交记录 -->
    <el-card class="glass-panel content-card">
      <template #header>
        <div class="flex-between">
          <h3 class="page-title-sm">我的成交</h3>
          <el-tag type="success" effect="dark" class="glass-tag">Trade History</el-tag>
        </div>
      </template>
      
      <!-- 成交筛选栏 -->
      <div class="filter-bar mb-16">
        <el-input
          v-model="tradeFilters.symbol"
          placeholder="交易对"
          class="filter-input"
          clearable
          @keyup.enter="applyTradeFilters"
        >
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
        
        <el-select v-model="tradeFilters.side" placeholder="方向" class="filter-select" clearable>
          <el-option label="全部" value="" />
          <el-option label="买入" value="buy" />
          <el-option label="卖出" value="sell" />
        </el-select>
        
        <el-date-picker
          v-model="tradeFilters.dateRange"
          type="daterange"
          range-separator="至"
          start-placeholder="开始日期"
          end-placeholder="结束日期"
          class="filter-date"
          @change="applyTradeFilters"
        />
        
        <el-button type="primary" @click="applyTradeFilters">
          <el-icon><Search /></el-icon> 筛选
        </el-button>
        <el-button @click="resetTradeFilters">重置</el-button>
      </div>
      
      <el-table :data="filteredTrades" class="orders-table w-100" v-loading="tradesLoading">
        <el-table-column prop="symbol" label="交易对" width="120" />
        <el-table-column prop="price" label="价格" width="120">
          <template #default="{ row }">
            <span class="mono">{{ formatPrice(row.price) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="amount" label="数量" width="100">
          <template #default="{ row }">
            <span class="mono">{{ formatAmount(row.amount) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="side" label="方向" width="80">
          <template #default="{ row }">
            <span :class="row.side === 'buy' ? 'text-success' : 'text-danger'">
              {{ row.side === 'buy' ? '买入' : '卖出' }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="fee" label="手续费" width="100">
          <template #default="{ row }">
            <span class="mono text-muted">{{ row.fee || '0' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="时间" width="160">
           <template #default="{ row }">
             <span class="text-muted mono">{{ formatDate(row.created_at) }}</span>
           </template>
        </el-table-column>
      </el-table>
      
      <!-- 分页 -->
      <el-pagination
        v-model:current-page="tradePage"
        :page-size="tradePageSize"
        :total="filteredTrades.length"
        layout="prev, pager, next"
        class="pagination-center mt-16"
        @current-change="handleTradePageChange"
      />
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { getMyOrders, getMyTrades, cancelOrder as cancelOrderApi } from '../api/api'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search } from '@element-plus/icons-vue'

const orders = ref([])
const trades = ref([])
const ordersLoading = ref(false)
const tradesLoading = ref(false)

// 订单分页
const orderPage = ref(1)
const orderPageSize = ref(20)

// 成交分页
const tradePage = ref(1)
const tradePageSize = ref(20)

// 订单筛选条件
const orderFilters = ref({
  symbol: '',
  side: '',
  type: '',
  status: '',
  dateRange: []
})

// 成交筛选条件
const tradeFilters = ref({
  symbol: '',
  side: '',
  dateRange: []
})

// 是否有止损止盈订单
const hasStopOrders = computed(() => {
  return orders.value.some(o => o.type === 'stop_limit' || o.type === 'stop_market')
})

// 过滤后的订单（前端筛选）
const filteredOrders = computed(() => {
  let result = [...orders.value]
  
  // 交易对筛选
  if (orderFilters.value.symbol) {
    result = result.filter(o => 
      o.symbol.toLowerCase().includes(orderFilters.value.symbol.toLowerCase())
    )
  }
  
  // 方向筛选
  if (orderFilters.value.side) {
    result = result.filter(o => o.side === orderFilters.value.side)
  }
  
  // 类型筛选
  if (orderFilters.value.type) {
    result = result.filter(o => o.type === orderFilters.value.type)
  }
  
  // 状态筛选
  if (orderFilters.value.status) {
    result = result.filter(o => o.status === orderFilters.value.status)
  }
  
  // 日期范围筛选
  if (orderFilters.value.dateRange && orderFilters.value.dateRange.length === 2) {
    const [start, end] = orderFilters.value.dateRange
    result = result.filter(o => {
      const orderTime = new Date(o.created_at).getTime()
      return orderTime >= start.getTime() && orderTime <= end.getTime() + 86400000
    })
  }
  
  return result
})

// 过滤后的成交（前端筛选）
const filteredTrades = computed(() => {
  let result = [...trades.value]
  
  // 交易对筛选
  if (tradeFilters.value.symbol) {
    result = result.filter(t => 
      t.symbol.toLowerCase().includes(tradeFilters.value.symbol.toLowerCase())
    )
  }
  
  // 方向筛选
  if (tradeFilters.value.side) {
    result = result.filter(t => t.side === tradeFilters.value.side)
  }
  
  // 日期范围筛选
  if (tradeFilters.value.dateRange && tradeFilters.value.dateRange.length === 2) {
    const [start, end] = tradeFilters.value.dateRange
    result = result.filter(t => {
      const tradeTime = new Date(t.created_at).getTime()
      return tradeTime >= start.getTime() && tradeTime <= end.getTime() + 86400000
    })
  }
  
  return result
})

// 格式化时间显示
const formatDate = (timestamp) => {
  if (!timestamp) return ''
  let timestampMs = timestamp
  if (String(timestamp).length <= 10) {
    timestampMs = timestamp * 1000
  }
  const date = new Date(timestampMs)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false
  })
}

// 格式化价格
const formatPrice = (price) => {
  if (!price || price === '' || isNaN(price)) return '--'
  return parseFloat(price).toFixed(2)
}

// 格式化数量
const formatAmount = (amount) => {
  if (!amount || amount === '' || isNaN(amount)) return '--'
  return parseFloat(amount).toFixed(4)
}

// 获取订单类型文本
const getTypeText = (type) => {
  const typeMap = {
    'limit': '限价单',
    'market': '市价单',
    'stop_limit': '止损限价',
    'stop_market': '止损市价'
  }
  return typeMap[type] || type
}

// 获取订单状态类型
const getStatusType = (status) => {
  const statusMap = {
    'open': 'warning',
    'partially_filled': 'primary',
    'filled': 'success',
    'cancelled': 'info'
  }
  return statusMap[status] || 'info'
}

// 获取订单状态文本
const getStatusText = (status) => {
  const statusMap = {
    'open': '进行中',
    'partially_filled': '部分成交',
    'filled': '已成交',
    'cancelled': '已取消'
  }
  return statusMap[status] || status
}

// 应用订单筛选
const applyOrderFilters = () => {
  orderPage.value = 1
}

// 重置订单筛选
const resetOrderFilters = () => {
  orderFilters.value = {
    symbol: '',
    side: '',
    type: '',
    status: '',
    dateRange: []
  }
  orderPage.value = 1
}

// 应用成交筛选
const applyTradeFilters = () => {
  tradePage.value = 1
}

// 重置成交筛选
const resetTradeFilters = () => {
  tradeFilters.value = {
    symbol: '',
    side: '',
    dateRange: []
  }
  tradePage.value = 1
}

// 取消订单
const handleCancelOrder = async (row) => {
  try {
    await ElMessageBox.confirm(
      `确定要撤消订单 #${row.id} 吗？\n交易对: ${row.symbol}\n方向: ${row.side === 'buy' ? '买入' : '卖出'}\n数量: ${row.amount}`,
      '确认撤单',
      { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
    )
    
    await cancelOrderApi(row.id)
    ElMessage.success('撤单成功')
    fetchOrders()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('撤单失败')
    }
  }
}

// 订单分页
const handleOrderPageChange = (page) => {
  orderPage.value = page
}

// 成交分页
const handleTradePageChange = (page) => {
  tradePage.value = page
}

const fetchOrders = async () => {
  ordersLoading.value = true
  try {
    const { data } = await getMyOrders()
    orders.value = data.data || []
  } catch (error) {
    console.error('获取订单失败:', error)
    ElMessage.error('获取订单失败')
  } finally {
    ordersLoading.value = false
  }
}

const fetchTrades = async () => {
  tradesLoading.value = true
  try {
    const { data } = await getMyTrades()
    trades.value = data.data || []
  } catch (error) {
    console.error('获取成交失败:', error)
    ElMessage.error('获取成交失败')
  } finally {
    tradesLoading.value = false
  }
}

onMounted(() => {
  fetchOrders()
  fetchTrades()
})
</script>

<style scoped>
.orders-container {
  padding: 8px;
  max-width: 1400px;
  margin: 0 auto;
}

.page-title-sm {
  margin: 0;
  font-size: 1.15rem;
  font-weight: 600;
  color: var(--text-main);
}

.orders-table {
  border-radius: var(--radius-base);
}

.orders-table :global(.el-table__row):hover {
  background: var(--primary-glow) !important;
}

.status-tag {
  font-weight: 700;
}

/* 筛选栏样式 */
.filter-bar {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  align-items: center;
  margin-bottom: 16px;
  padding: 12px;
  background: color-mix(in srgb, var(--bg-elevated), transparent 85%);
  border-radius: var(--radius-base);
}

.filter-input {
  width: 140px;
}

.filter-select {
  width: 120px;
}

.filter-date {
  width: 260px;
}

/* shared utilities (scoped fallback) */
.mb-20 { margin-bottom: 20px; }
.mb-16 { margin-bottom: 16px; }
.mt-16 { margin-top: 16px; }
.flex-between { display: flex; justify-content: space-between; align-items: center; }
.text-success { color: var(--success); font-weight: 600; }
.text-danger { color: var(--danger); font-weight: 600; }
.text-muted { color: var(--text-muted); }
.mono { font-family: 'Roboto Mono', 'Fira Code', monospace; }
.content-card { background: var(--bg-card); border-radius: var(--radius-card); }
.pagination-center {
  display: flex;
  justify-content: center;
  margin-top: 16px;
}
</style>
