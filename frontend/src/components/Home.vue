<template>
  <!-- 交易中心全屏模式 -->
  <div v-if="currentContent === 'trade'" class="trade-fullscreen">
    <el-button type="primary" @click="showContent('home')" class="back-home-btn">
      <el-icon><ArrowLeft /></el-icon>
      <span>返回首页</span>
    </el-button>
    <Trade />
  </div>

  <!-- 主布局模式 -->
  <el-container v-else class="app-layout">
    <el-header class="app-header-slot">
      <Header @nav="showContent" />
    </el-header>

    <el-container class="app-body">
      <el-aside width="220px" class="app-aside">
        <SideBar :activeMenu="activeMenu" @menu="handleMenuSelect" />
      </el-aside>

      <el-main class="app-main">
        <!-- 首页概览 -->
        <div v-if="currentContent === 'home'" class="dashboard fade-in">
          <div class="dashboard-grid">
            <el-card class="glass-panel dashboard-card" shadow="hover">
              <template #header>
                <div class="flex-center">
                  <el-icon class="text-primary"><TrendCharts /></el-icon>
                  <span>Market Overview</span>
                </div>
              </template>
              <div class="text-center p-20">
                <h3 class="stat-label text-primary">BTC/USDT</h3>
                <p class="stat-value text-success">$45,123.45</p>
                <p class="text-muted stat-change">+2.34%</p>
              </div>
            </el-card>

            <el-card class="glass-panel dashboard-card" shadow="hover">
              <template #header>
                <div class="flex-center">
                  <el-icon class="text-warning"><Wallet /></el-icon>
                  <span>Total Assets</span>
                </div>
              </template>
              <div class="text-center p-20">
                <h3 class="stat-label text-warning">Balance</h3>
                <p class="stat-value text-success">$12,345.67</p>
                <p class="text-muted stat-change">24h Change</p>
              </div>
            </el-card>

            <el-card class="glass-panel dashboard-card" shadow="hover">
              <template #header>
                <div class="flex-center">
                  <el-icon class="text-danger"><Document /></el-icon>
                  <span>Trading Stats</span>
                </div>
              </template>
              <div class="text-center p-20">
                <h3 class="stat-label text-danger">Today's Orders</h3>
                <p class="stat-value text-success">156</p>
                <p class="text-muted stat-change">Orders Executed</p>
              </div>
            </el-card>
          </div>

          <el-row :gutter="20" class="mt-20">
            <el-col :span="12">
              <el-card class="glass-panel h-full">
                <template #header>
                  <div class="flex-between">
                    <span>Quick Actions</span>
                  </div>
                </template>
                <el-row :gutter="15">
                  <el-col :span="12">
                    <el-button type="primary" class="action-btn w-100 mb-10" @click="showContent('trade')">
                      <el-icon class="mr-10"><TrendCharts /></el-icon>
                      Trade Now
                    </el-button>
                  </el-col>
                  <el-col :span="12">
                    <el-button type="success" class="action-btn w-100 mb-10" @click="showContent('assets')">
                      <el-icon class="mr-10"><Wallet /></el-icon>
                      Assets
                    </el-button>
                  </el-col>
                  <el-col :span="12">
                    <el-button type="warning" class="action-btn w-100 mb-10" @click="showContent('orders')">
                      <el-icon class="mr-10"><Document /></el-icon>
                      Orders
                    </el-button>
                  </el-col>
                  <el-col :span="12">
                    <el-button type="info" class="action-btn w-100 mb-10" @click="showContent('admin')">
                      <el-icon class="mr-10"><Setting /></el-icon>
                      Admin
                    </el-button>
                  </el-col>
                </el-row>
              </el-card>
            </el-col>
            <el-col :span="12">
              <el-card class="glass-panel h-full">
                <template #header>
                  <div class="flex-between">
                    <span>Announcements</span>
                  </div>
                </template>
                <el-timeline class="notice-timeline w-100">
                  <el-timeline-item v-for="(item, i) in announcements" :key="i" :type="item.type" :color="item.color" size="large">
                    <div class="notice-item">{{ item.text }}</div>
                  </el-timeline-item>
                </el-timeline>
              </el-card>
            </el-col>
          </el-row>
        </div>

        <!-- 用户登录 -->
        <div v-else-if="currentContent === 'login'">
          <Login />
        </div>

        <!-- 用户注册 -->
        <div v-else-if="currentContent === 'register'">
          <Register />
        </div>

        <!-- 资产管理 -->
        <div v-else-if="currentContent === 'assets'">
          <Assets />
        </div>

        <!-- 订单管理 -->
        <div v-else-if="currentContent === 'orders'">
          <Orders />
        </div>

        <!-- 后台管理 -->
        <div v-else-if="currentContent === 'admin'">
          <Admin />
        </div>

        <!-- 实时行情监控 -->
        <div v-else-if="currentContent === 'realtime'">
          <RealTimeMarket />
        </div>
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { ref } from 'vue'
import { ArrowLeft, House, User, UserFilled, Wallet, TrendCharts, Document, Setting, ArrowUp, ArrowDown } from '@element-plus/icons-vue'
import Trade from './Trade.vue'
import Header from './Header.vue'
import SideBar from './SideBar.vue'
import Assets from './Assets.vue'
import Admin from './Admin.vue'
import RealTimeMarket from './RealTimeMarket.vue'
import Login from './Login.vue'
import Register from './Register.vue'
import Orders from './Orders.vue'

// 当前显示的内容
const currentContent = ref('home')
const activeMenu = ref('home')

// 模拟公告
const announcements = ref([
  { text: 'System Maintenance: Jan 15, 02:00-04:00 UTC', type: 'warning', color: 'var(--warning)' },
  { text: 'New Listing: ETH/USDT is now live!', type: 'success', color: 'var(--success)' },
  { text: 'Zero Fee Trading Promotion', type: 'primary', color: 'var(--primary)' },
  { text: 'Security Alert: Enable 2FA for your account', type: 'danger', color: 'var(--danger)' },
])

// 处理菜单选择
const handleMenuSelect = (index) => {
  currentContent.value = index
  activeMenu.value = index
}

// 显示内容
const showContent = (content) => {
  currentContent.value = content
  activeMenu.value = content
}
</script>

<style scoped>
.app-layout {
  height: 100vh;
  width: 100%;
  overflow: hidden;
}

.app-header-slot {
  height: 64px !important;
  padding: 0 !important;
  background: transparent;
  border: none;
  box-shadow: none;
}

.app-body {
  height: calc(100vh - 64px);
  width: 100%;
  overflow: hidden;
}

.app-aside {
  background: transparent !important;
  border: none !important;
  overflow: visible;
}

.app-main {
  background: transparent !important;
  padding: 24px !important;
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
  top: 20px;
  left: 20px;
  z-index: 10;
  gap: 6px;
}

/* Dashboard */
.dashboard { padding: 8px 0; }

.dashboard-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(230px, 1fr));
  gap: 20px;
}

.dashboard-card {
  height: 220px;
  padding: 12px;
}

.stat-label {
  font-size: 1rem;
  font-weight: 700;
  margin-bottom: 8px;
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
  margin: 10px 0;
}

.stat-change {
  font-size: 0.85rem;
  font-weight: 600;
}

.h-full { height: 100%; }

.action-btn {
  height: 46px;
  font-weight: 600;
  border-radius: var(--radius-base);
}

.action-btn:hover {
  transform: translateY(-2px);
}

.notice-timeline :global(.el-timeline-item__dot) {
  top: 4px !important;
}

.notice-item {
  color: var(--text-muted);
  font-size: 0.9rem;
  padding: 6px 0;
}

.notice-item:hover {
  color: var(--text-main);
}

.mt-20 { margin-top: 20px; }
.mb-10 { margin-bottom: 10px; }
.mr-10 { margin-right: 10px; }
.w-100 { width: 100%; }

.text-primary { color: var(--primary); }
.text-success { color: var(--success); }
.text-warning { color: var(--warning); }
.text-danger { color: var(--danger); }
.text-muted { color: var(--text-muted); }

.fade-in {
  animation: fadeIn 0.4s ease-out;
}
@keyframes fadeIn {
  from { opacity: 0; transform: translateY(8px); }
  to { opacity: 1; transform: translateY(0); }
}

@media (max-width: 768px) {
  .app-main { padding: 16px !important; }
}
</style>
