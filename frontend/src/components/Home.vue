<template>
  <!-- 交易中心全屏模式 -->
  <div v-if="currentContent === 'trade'" class="trade-fullscreen">
    <el-button class="back-home-btn glass-panel" @click="showContent('home')">
      <el-icon><ArrowLeft /></el-icon>
      <span>返回首页</span>
    </el-button>
    <Trade :initial-symbol="tradeSymbol" />
  </div>

  <!-- 主布局模式 -->
  <el-container v-else class="app-layout">
    <el-header class="app-header-slot">
      <Header :active="activeMenu" @nav="showContent" />
    </el-header>

    <el-container class="app-body">
      <el-aside width="232px" class="app-aside">
        <SideBar :activeMenu="activeMenu" @menu="handleMenuSelect" />
      </el-aside>

      <el-main class="app-main">
        <!-- 首页概览 -->
        <div v-if="currentContent === 'home'" class="dashboard">
          <!-- 实时行情滚动条 -->
          <div class="ticker-bar glass-panel">
            <div class="ticker-label">
              <span class="live-dot" :class="{ off: !isConnected }"></span>
              <span>{{ isConnected ? '实时行情' : '离线' }}</span>
            </div>
            <div class="ticker-track">
              <button
                v-for="row in marketRows"
                :key="row.symbol"
                class="ticker-chip"
                @click="goTrade(row.symbol)"
              >
                <span class="ticker-symbol">{{ row.symbol }}</span>
                <span class="ticker-price mono">{{ fmt.price(row.price, 2) }}</span>
                <span class="ticker-change" :class="row.changePercent >= 0 ? 'text-up' : 'text-down'">
                  {{ fmt.percent(row.changePercent) }}
                </span>
              </button>
            </div>
          </div>

          <!-- KPI 概览 -->
          <div class="kpi-grid stagger">
            <StatCard
              label="BTC / USDT"
              :value="'$' + fmt.price(btc.price, 2)"
              :delta="btc.changePercent"
              sub="24h 涨跌"
              icon="TrendCharts"
              tone="warning"
              :spark="btc.history"
            />
            <StatCard
              label="24h 总成交额"
              :value="'$' + fmt.compact(totalVolume)"
              icon="DataLine"
              tone="primary"
              sub="全市场"
              :spark="volumeSpark"
            />
            <StatCard
              label="我的资产估值"
              :value="assetsLoading ? '—' : '$' + fmt.money(assetValue)"
              icon="Wallet"
              tone="success"
              :sub="assetsError ? '未登录 / 无数据' : `${assetCount} 个币种`"
              :loading="assetsLoading"
              :spark="assetSpark"
            />
            <StatCard
              label="进行中订单"
              :value="ordersLoading ? '—' : String(openOrderCount)"
              icon="Tickets"
              tone="violet"
              :sub="ordersError ? '未登录 / 无数据' : `共 ${orderCount} 笔`"
              :loading="ordersLoading"
              :spark="orderSpark"
            />
          </div>

          <!-- 市场行情 -->
          <section class="glass-panel panel">
            <div class="panel-head">
              <div class="flex-center gap-10">
                <h3 class="card-title">市场行情</h3>
                <span class="badge badge-primary">Live</span>
              </div>
              <div class="flex-center gap-10">
                <el-radio-group v-model="marketFilter" size="small">
                  <el-radio-button label="all">全部</el-radio-button>
                  <el-radio-button label="up">上涨</el-radio-button>
                  <el-radio-button label="down">下跌</el-radio-button>
                </el-radio-group>
              </div>
            </div>

            <el-table :data="filteredMarketRows" class="market-table" style="width:100%">
              <el-table-column label="交易对" min-width="150">
                <template #default="{ row }">
                  <div class="pair-cell">
                    <span class="pair-avatar" :style="{ background: pairColor(row.symbol) }">
                      {{ row.symbol.split('/')[0].slice(0, 1) }}
                    </span>
                    <div class="pair-meta">
                      <span class="pair-name">{{ row.symbol.split('/')[0] }}</span>
                      <span class="pair-quote">/ {{ row.symbol.split('/')[1] }}</span>
                    </div>
                  </div>
                </template>
              </el-table-column>

              <el-table-column label="最新价" min-width="130" align="right">
                <template #default="{ row }">
                  <span class="mono font-semibold" :class="row.changePercent >= 0 ? 'text-up' : 'text-down'">
                    {{ fmt.price(row.price, 2) }}
                  </span>
                </template>
              </el-table-column>

              <el-table-column label="24h 涨跌" width="120" align="right">
                <template #default="{ row }">
                  <span class="badge" :class="row.changePercent >= 0 ? 'badge-up' : 'badge-down'">
                    {{ fmt.percent(row.changePercent) }}
                  </span>
                </template>
              </el-table-column>

              <el-table-column label="24h 最高" min-width="110" align="right" class-name="hide-sm">
                <template #default="{ row }">
                  <span class="mono text-muted">{{ fmt.price(row.high, 2) }}</span>
                </template>
              </el-table-column>

              <el-table-column label="24h 最低" min-width="110" align="right" class-name="hide-sm">
                <template #default="{ row }">
                  <span class="mono text-muted">{{ fmt.price(row.low, 2) }}</span>
                </template>
              </el-table-column>

              <el-table-column label="24h 成交量" min-width="120" align="right" class-name="hide-sm">
                <template #default="{ row }">
                  <span class="mono text-muted">{{ fmt.compact(row.volume) }}</span>
                </template>
              </el-table-column>

              <el-table-column label="走势" width="130" align="center" class-name="hide-sm">
                <template #default="{ row }">
                  <Sparkline :data="row.history" :width="110" :height="32" />
                </template>
              </el-table-column>

              <el-table-column label="操作" width="96" align="right">
                <template #default="{ row }">
                  <el-button type="primary" size="small" @click="goTrade(row.symbol)">交易</el-button>
                </template>
              </el-table-column>
            </el-table>
          </section>

          <!-- 快捷操作 + 公告 -->
          <div class="bottom-grid">
            <section class="glass-panel panel">
              <div class="panel-head">
                <h3 class="card-title">快捷操作</h3>
              </div>
              <div class="quick-grid">
                <button
                  v-for="action in quickActions"
                  :key="action.key"
                  class="quick-card"
                  @click="showContent(action.key)"
                >
                  <span class="quick-icon" :class="`tone-${action.tone}`">
                    <el-icon><component :is="action.icon" /></el-icon>
                  </span>
                  <span class="quick-text">
                    <span class="quick-title">{{ action.title }}</span>
                    <span class="quick-desc">{{ action.desc }}</span>
                  </span>
                  <el-icon class="quick-arrow"><ArrowRight /></el-icon>
                </button>
              </div>
            </section>

            <section class="glass-panel panel">
              <div class="panel-head">
                <h3 class="card-title">公告</h3>
                <span class="text-muted text-sm">{{ announcements.length }} 条</span>
              </div>
              <ul class="notice-list">
                <li v-for="(item, i) in announcements" :key="i" class="notice-row">
                  <span class="notice-dot" :style="{ background: item.color }"></span>
                  <span class="notice-text">{{ item.text }}</span>
                  <span class="notice-time">{{ item.time }}</span>
                </li>
              </ul>
            </section>
          </div>
        </div>

        <!-- 用户登录 -->
        <div v-else-if="currentContent === 'login'"><Login /></div>
        <!-- 用户注册 -->
        <div v-else-if="currentContent === 'register'"><Register /></div>
        <!-- 资产管理 -->
        <div v-else-if="currentContent === 'assets'"><Assets /></div>
        <!-- 订单管理 -->
        <div v-else-if="currentContent === 'orders'"><Orders /></div>
        <!-- 后台管理 -->
        <div v-else-if="currentContent === 'admin'"><Admin /></div>
        <!-- 实时行情监控 -->
        <div v-else-if="currentContent === 'realtime'"><RealTimeMarket /></div>
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ArrowLeft, ArrowRight } from '@element-plus/icons-vue'
import Trade from './Trade.vue'
import Header from './Header.vue'
import SideBar from './SideBar.vue'
import Assets from './Assets.vue'
import Admin from './Admin.vue'
import RealTimeMarket from './RealTimeMarket.vue'
import Login from './Login.vue'
import Register from './Register.vue'
import Orders from './Orders.vue'
import StatCard from './ui/StatCard.vue'
import Sparkline from './ui/Sparkline.vue'
import { useMarket, fmt } from '../composables/useMarket'
import { getAccounts, getMyOrders } from '../api/api'

// 当前显示的内容
const currentContent = ref('home')
const activeMenu = ref('home')
const tradeSymbol = ref('BTC/USDT')

// 实时行情
const { rows, isConnected } = useMarket()
const marketRows = computed(() => Object.values(rows))
const btc = computed(() => rows['BTC/USDT'] || { price: 0, changePercent: 0, history: [] })

const totalVolume = computed(() =>
  marketRows.value.reduce((sum, r) => sum + (Number(r.volume) || 0), 0)
)
const volumeSpark = computed(() => marketRows.value[0]?.history?.map((v) => v * 1000) || [])

// 资产 / 订单概览（未登录时优雅降级）
const assetsLoading = ref(true)
const assetsError = ref(false)
const assetRows = ref([])
const ordersLoading = ref(true)
const ordersError = ref(false)
const orderRows = ref([])

const assetValue = computed(() =>
  assetRows.value.reduce((sum, a) => {
    const total = (parseFloat(a.balance) || 0) + (parseFloat(a.frozen) || 0)
    return sum + total
  }, 0)
)
const assetCount = computed(() => assetRows.value.length)
const assetSpark = computed(() => {
  const base = assetValue.value || 12000
  return Array.from({ length: 24 }, (_, i) => base * (1 + Math.sin(i / 3) * 0.03 + i * 0.001))
})

const openOrderCount = computed(
  () => orderRows.value.filter((o) => o.status === 'open' || o.status === 'partially_filled').length
)
const orderCount = computed(() => orderRows.value.length)
const orderSpark = computed(() => {
  const base = Math.max(orderCount.value, 4)
  return Array.from({ length: 24 }, (_, i) => base + Math.sin(i / 2) * 2)
})

const loadAssets = async () => {
  assetsLoading.value = true
  assetsError.value = false
  try {
    const { data } = await getAccounts()
    assetRows.value = data?.data || []
  } catch {
    assetsError.value = true
    assetRows.value = []
  } finally {
    assetsLoading.value = false
  }
}

const loadOrders = async () => {
  ordersLoading.value = true
  ordersError.value = false
  try {
    const { data } = await getMyOrders()
    orderRows.value = data?.data || []
  } catch {
    ordersError.value = true
    orderRows.value = []
  } finally {
    ordersLoading.value = false
  }
}

// 市场筛选
const marketFilter = ref('all')
const filteredMarketRows = computed(() => {
  if (marketFilter.value === 'up') return marketRows.value.filter((r) => r.changePercent >= 0)
  if (marketFilter.value === 'down') return marketRows.value.filter((r) => r.changePercent < 0)
  return marketRows.value
})

// 交易对头像配色
const PAIR_COLORS = {
  BTC: 'linear-gradient(135deg,#f7931a,#f2a900)',
  ETH: 'linear-gradient(135deg,#627eea,#8a92b2)',
  BNB: 'linear-gradient(135deg,#f3ba2f,#e8a90c)',
  SOL: 'linear-gradient(135deg,#9945ff,#14f195)',
  XRP: 'linear-gradient(135deg,#23292f,#4b5563)',
}
const pairColor = (symbol) => PAIR_COLORS[symbol.split('/')[0]] || 'linear-gradient(135deg,#3b82f6,#2563eb)'

// 快捷操作
const quickActions = [
  { key: 'trade', title: '开始交易', desc: '现货 · 限价 / 市价 / 止损', icon: 'TrendCharts', tone: 'primary' },
  { key: 'assets', title: '资产管理', desc: '充值 · 提现 · 地址簿', icon: 'Wallet', tone: 'success' },
  { key: 'orders', title: '订单管理', desc: '挂单 · 成交记录', icon: 'Tickets', tone: 'warning' },
  { key: 'realtime', title: '实时行情', desc: '盘口深度 · 逐笔成交', icon: 'DataLine', tone: 'violet' },
]

// 公告
const announcements = ref([
  { text: '系统维护：1 月 15 日 02:00–04:00 (UTC)', color: 'var(--warning)', time: '2h' },
  { text: '新币上线：ETH/USDT 现已开放交易', color: 'var(--success)', time: '6h' },
  { text: '零手续费交易活动进行中', color: 'var(--primary)', time: '1d' },
  { text: '安全提醒：请为账户开启双重验证 (2FA)', color: 'var(--danger)', time: '2d' },
])

const handleMenuSelect = (index) => {
  currentContent.value = index
  activeMenu.value = index
}

const showContent = (content) => {
  currentContent.value = content
  activeMenu.value = content
}

const goTrade = (symbol) => {
  tradeSymbol.value = symbol
  showContent('trade')
}

onMounted(() => {
  loadAssets()
  loadOrders()
})
</script>

<style scoped>
.app-layout { height: 100vh; width: 100%; overflow: hidden; }

.app-header-slot {
  height: var(--header-h) !important;
  padding: 0 !important;
  background: transparent;
  border: none;
  box-shadow: none;
}

.app-body { height: calc(100vh - var(--header-h)); width: 100%; overflow: hidden; }
.app-aside { background: transparent !important; border: none !important; overflow: visible; }

.app-main {
  background: transparent !important;
  padding: 20px 24px 32px !important;
  overflow-y: auto !important;
  width: 100%;
}

.trade-fullscreen {
  position: fixed;
  inset: 0;
  height: 100vh;
  background: var(--bg-app);
  overflow: hidden;
}

.back-home-btn {
  position: absolute;
  top: 18px;
  left: 18px;
  z-index: 20;
  gap: 6px;
  height: 38px;
  padding: 0 16px;
  border: 1px solid var(--border-color) !important;
  background: var(--surface-1) !important;
  color: var(--text-main) !important;
  font-weight: 600;
}

/* ===== Dashboard ===== */
.dashboard { display: flex; flex-direction: column; gap: 18px; max-width: var(--content-max); margin: 0 auto; }

/* Ticker */
.ticker-bar {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 10px 16px;
  overflow: hidden;
}
.ticker-label {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.78rem;
  font-weight: 700;
  color: var(--text-muted);
  white-space: nowrap;
  padding-right: 16px;
  border-right: 1px solid var(--border-color);
}
.ticker-track {
  display: flex;
  align-items: center;
  gap: 10px;
  overflow-x: auto;
  scrollbar-width: none;
  padding: 2px 0;
}
.ticker-track::-webkit-scrollbar { display: none; }
.ticker-chip {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  padding: 6px 14px;
  border-radius: var(--radius-pill);
  border: 1px solid var(--border-subtle);
  background: var(--surface-3);
  cursor: pointer;
  white-space: nowrap;
  transition: var(--transition-snap);
  font-size: 0.8rem;
}
.ticker-chip:hover { background: var(--surface-2); border-color: var(--border-color-hover); transform: translateY(-1px); }
.ticker-symbol { font-weight: 700; color: var(--text-main); }
.ticker-price { color: var(--text-body); font-weight: 600; }
.ticker-change { font-weight: 700; font-size: 0.76rem; }

/* KPI */
.kpi-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(230px, 1fr));
  gap: 16px;
}

/* Panels */
.panel { padding: 18px 20px 20px; }
.panel-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 14px;
  flex-wrap: wrap;
}

/* Market table */
.pair-cell { display: flex; align-items: center; gap: 10px; }
.pair-avatar {
  width: 32px;
  height: 32px;
  border-radius: 10px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-weight: 800;
  font-size: 0.85rem;
  flex-shrink: 0;
  box-shadow: var(--shadow-xs);
}
.pair-meta { display: flex; flex-direction: column; line-height: 1.25; }
.pair-name { font-weight: 700; color: var(--text-main); font-size: 0.9rem; }
.pair-quote { font-size: 0.72rem; color: var(--text-secondary); }
.market-table :deep(.el-table__row) { cursor: default; }

/* Bottom grid */
.bottom-grid {
  display: grid;
  grid-template-columns: 1.15fr 1fr;
  gap: 18px;
}

.quick-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; }
.quick-card {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px;
  border-radius: var(--radius-base);
  border: 1px solid var(--border-color);
  background: var(--surface-3);
  cursor: pointer;
  text-align: left;
  transition: var(--transition-ease);
  font-family: inherit;
}
.quick-card:hover {
  background: var(--surface-2);
  border-color: var(--border-color-hover);
  transform: translateY(-2px);
  box-shadow: var(--shadow-sm);
}
.quick-icon {
  width: 40px;
  height: 40px;
  border-radius: 12px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 19px;
  flex-shrink: 0;
}
.quick-text { display: flex; flex-direction: column; gap: 1px; flex: 1; min-width: 0; }
.quick-title { font-weight: 650; color: var(--text-main); font-size: 0.88rem; }
.quick-desc { font-size: 0.72rem; color: var(--text-muted); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.quick-arrow { color: var(--text-faint); font-size: 14px; }

.tone-primary { background: var(--primary-grad); }
.tone-success { background: var(--success-grad); }
.tone-warning { background: var(--warning-grad); }
.tone-violet { background: var(--violet-grad); }

/* Announcements */
.notice-list { list-style: none; margin: 0; padding: 0; display: flex; flex-direction: column; }
.notice-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 11px 4px;
  border-bottom: 1px solid var(--border-subtle);
  transition: var(--transition-snap);
}
.notice-row:last-child { border-bottom: none; }
.notice-row:hover { background: var(--surface-sunken); border-radius: var(--radius-xs); }
.notice-dot { width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0; box-shadow: 0 0 0 3px color-mix(in srgb, currentColor, transparent 85%); }
.notice-text { flex: 1; font-size: 0.85rem; color: var(--text-body); }
.notice-time { font-size: 0.72rem; color: var(--text-secondary); font-variant-numeric: tabular-nums; }

.text-up { color: var(--up); }
.text-down { color: var(--down); }

@media (max-width: 1100px) {
  .bottom-grid { grid-template-columns: 1fr; }
}
@media (max-width: 768px) {
  .app-main { padding: 14px !important; }
  .quick-grid { grid-template-columns: 1fr; }
  .market-table :deep(.hide-sm) { display: none; }
}
</style>
