<template>
  <el-aside class="app-sidebar" width="220px">
    <div class="sidebar-inner">
      <el-menu
        :default-active="activeMenu"
        @select="$emit('menu', $event)"
        class="sidebar-menu"
      >
        <el-divider class="sidebar-divider" />
        <el-menu-item index="home">
          <el-icon><House /></el-icon>
          <span>首页概览</span>
        </el-menu-item>
        <el-menu-item index="trade">
          <el-icon><TrendCharts /></el-icon>
          <span>交易中心</span>
        </el-menu-item>
        <el-menu-item index="assets">
          <el-icon><Wallet /></el-icon>
          <span>资产管理</span>
        </el-menu-item>
        <el-menu-item index="orders">
          <el-icon><Document /></el-icon>
          <span>订单管理</span>
        </el-menu-item>
        <el-menu-item index="realtime">
          <el-icon><Bell /></el-icon>
          <span>实时行情</span>
        </el-menu-item>
        <el-divider class="sidebar-divider" />
        <el-menu-item index="admin">
          <el-icon><Setting /></el-icon>
          <span>后台管理</span>
        </el-menu-item>
      </el-menu>
    </div>
  </el-aside>
</template>

<script setup>
import { Bell, House, User, UserFilled, Wallet, TrendCharts, Document, Setting } from '@element-plus/icons-vue'
const props = defineProps({ activeMenu: String })
</script>

<style scoped>
.app-sidebar {
  height: 100%;
  overflow-y: auto;
  background: color-mix(in srgb, var(--bg-card), transparent 88%) !important;
  backdrop-filter: blur(14px);
  border-right: 1px solid var(--border-color) !important;
}

.sidebar-inner { padding: 16px 8px; }

.sidebar-menu {
  background: transparent !important;
  border: none !important;
  height: 100%;
}

.sidebar-divider {
  border-color: var(--border-color) !important;
  margin: 12px 0 !important;
}

.sidebar-menu :global(.el-menu-item) {
  border-radius: var(--radius-base) !important;
  margin: 4px 8px !important;
  height: 42px !important;
  color: var(--text-muted) !important;
  font-weight: 500 !important;
}

.sidebar-menu :global(.el-menu-item.is-active) {
  color: var(--primary) !important;
  font-weight: 700 !important;
  border-left: 3px solid var(--primary) !important;
}

.sidebar-menu :global(.el-menu-item:hover) {
  color: var(--text-main) !important;
  background: var(--primary-glow) !important;
}

.sidebar-menu :global(.el-menu-item) :global(.el-icon) {
  margin-right: 10px;
  color: inherit;
}

/* responsive */
@media (max-width: 768px) {
  .app-sidebar { width: 60px !important; }
  .sidebar-menu span,
  .sidebar-divider span { display: none; }
}
</style>
