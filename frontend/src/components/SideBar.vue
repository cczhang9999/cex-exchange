<template>
  <aside class="app-sidebar">
    <div class="sidebar-inner">
      <div class="side-section">导航</div>
      <el-menu :default-active="activeMenu" @select="$emit('menu', $event)" class="sidebar-menu">
        <el-menu-item v-for="item in primaryItems" :key="item.key" :index="item.key">
          <el-icon><component :is="item.icon" /></el-icon>
          <span>{{ item.label }}</span>
        </el-menu-item>
      </el-menu>

      <div class="side-section">管理</div>
      <el-menu :default-active="activeMenu" @select="$emit('menu', $event)" class="sidebar-menu">
        <el-menu-item v-for="item in adminItems" :key="item.key" :index="item.key">
          <el-icon><component :is="item.icon" /></el-icon>
          <span>{{ item.label }}</span>
        </el-menu-item>
      </el-menu>

      <div class="side-foot">
        <div class="side-foot-card">
          <div class="side-foot-title">
            <span class="live-dot"></span>
            系统运行正常
          </div>
          <div class="side-foot-desc">撮合引擎 · 行情推送 · 资产服务</div>
        </div>
      </div>
    </div>
  </aside>
</template>

<script setup>
defineProps({ activeMenu: { type: String, default: 'home' } })
defineEmits(['menu'])

const primaryItems = [
  { key: 'home', label: '首页概览', icon: 'House' },
  { key: 'trade', label: '交易中心', icon: 'TrendCharts' },
  { key: 'assets', label: '资产管理', icon: 'Wallet' },
  { key: 'orders', label: '订单管理', icon: 'Tickets' },
  { key: 'realtime', label: '实时行情', icon: 'DataLine' },
]

const adminItems = [{ key: 'admin', label: '后台管理', icon: 'Setting' }]
</script>

<style scoped>
.app-sidebar {
  height: 100%;
  overflow-y: auto;
  background: color-mix(in srgb, var(--surface-1), transparent 55%) !important;
  backdrop-filter: blur(16px) saturate(140%);
  border-right: 1px solid var(--border-color) !important;
}

.sidebar-inner {
  padding: 16px 10px 20px;
  display: flex;
  flex-direction: column;
  min-height: 100%;
}

.side-section {
  font-size: 0.66rem;
  font-weight: 700;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--text-faint);
  padding: 12px 14px 6px;
}

.sidebar-menu { background: transparent !important; border: none !important; }

.sidebar-menu :deep(.el-menu-item) {
  border-radius: var(--radius-base) !important;
  margin: 2px 4px !important;
  height: 42px !important;
  line-height: 42px !important;
  color: var(--text-muted) !important;
  font-weight: 550 !important;
  font-size: 0.87rem;
}
.sidebar-menu :deep(.el-menu-item.is-active) {
  color: var(--primary) !important;
  background: var(--primary-glow) !important;
  font-weight: 700 !important;
}
.sidebar-menu :deep(.el-menu-item:hover) {
  color: var(--text-main) !important;
  background: var(--surface-sunken) !important;
}
.sidebar-menu :deep(.el-menu-item.is-active:hover) {
  color: var(--primary) !important;
  background: var(--primary-glow) !important;
}
.sidebar-menu :deep(.el-menu-item .el-icon) { margin-right: 10px; color: inherit; font-size: 17px; }

.side-foot { margin-top: auto; padding: 14px 4px 0; }
.side-foot-card {
  padding: 12px 14px;
  border-radius: var(--radius-base);
  background: var(--surface-sunken);
  border: 1px solid var(--border-subtle);
}
.side-foot-title {
  display: flex;
  align-items: center;
  gap: 7px;
  font-size: 0.78rem;
  font-weight: 700;
  color: var(--text-body);
}
.side-foot-desc { font-size: 0.68rem; color: var(--text-secondary); margin-top: 4px; line-height: 1.5; }

@media (max-width: 768px) {
  .app-sidebar { width: 62px !important; }
  .side-section, .sidebar-menu :deep(.el-menu-item span), .side-foot { display: none; }
  .sidebar-menu :deep(.el-menu-item) { justify-content: center; }
  .sidebar-menu :deep(.el-menu-item .el-icon) { margin-right: 0; }
}
</style>
