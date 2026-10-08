<template>
  <el-header class="app-header">
    <div class="header-left">
      <el-icon class="logo-icon"><TrendCharts /></el-icon>
      <span class="brand-text">CEX Exchange</span>
    </div>

    <div class="header-nav">
      <el-button link class="nav-link" @click="$emit('nav', 'home')">首页概览</el-button>
      <el-button link class="nav-link" @click="$emit('nav', 'trade')">交易中心</el-button>
      <el-button link class="nav-link" @click="$emit('nav', 'assets')">资产管理</el-button>
      <el-button link class="nav-link" @click="$emit('nav', 'orders')">订单管理</el-button>
      <el-button link class="nav-link" @click="$emit('nav', 'admin')">后台管理</el-button>
    </div>

    <div class="header-actions">
      <el-divider direction="vertical" class="header-divider" />
      <el-tooltip content="切换主题" placement="bottom">
        <el-button
          link
          class="theme-toggle"
          @click="toggleTheme"
          :icon="isDark ? Moon : Sunny"
        />
      </el-tooltip>
      <el-divider direction="vertical" class="header-divider" />
      <el-icon class="user-icon"><User /></el-icon>
      <span class="username">{{ userStore.username || '未登录' }}</span>
      <el-button type="danger" size="small" class="logout-btn" @click="handleLogout">登出</el-button>
    </div>
  </el-header>
</template>

<script setup>
import { Sunny, Moon, User, TrendCharts } from '@element-plus/icons-vue'
import { useUserStore } from '../stores'
import { useRouter } from 'vue-router'
import { ElMessageBox, ElMessage } from 'element-plus'
import { isDark, toggleTheme } from '../utils/theme'

const userStore = useUserStore()
const router = useRouter()

const handleLogout = async () => {
  try {
    await ElMessageBox.confirm('确定要退出登录吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    })

    userStore.logout()
    ElMessage.success('已退出登录')
    router.push('/login')
  } catch {
    // 用户取消操作
  }
}
</script>

<style scoped>
.app-header {
  height: 64px !important;
  padding: 0 24px !important;
  display: flex !important;
  align-items: center !important;
  justify-content: space-between !important;
  background: color-mix(in srgb, var(--bg-card), transparent 70%) !important;
  backdrop-filter: blur(14px);
  border-bottom: 1px solid var(--border-color) !important;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 10px;
  font-weight: 800;
  font-size: 1.15rem;
  color: var(--text-main);
}

.logo-icon {
  font-size: 22px;
  color: var(--primary);
}

.header-nav {
  display: flex;
  align-items: center;
  gap: 4px;
}

.nav-link {
  color: var(--text-muted) !important;
  font-weight: 500 !important;
  border-radius: var(--radius-base) !important;
  padding: 6px 12px !important;
  transition: var(--transition-snap);
}

.nav-link:hover {
  color: var(--text-main) !important;
  background: var(--bg-card-hover) !important;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.header-divider {
  border-left-color: var(--border-color) !important;
  height: 32px !important;
  margin: 0 !important;
}

.user-icon {
  font-size: 18px;
  color: var(--primary);
}

.username {
  font-weight: 600;
  color: var(--text-main);
  min-width: 80px;
}

.logout-btn {
  border-radius: var(--radius-base) !important;
  font-weight: 500 !important;
  border: 1px solid var(--danger-glow);
  background: transparent;
  color: var(--danger) !important;
}

.logout-btn:hover {
  background: var(--danger-glow);
  transform: translateY(-1px);
}

.theme-toggle {
  color: var(--text-muted) !important;
  font-size: 18px;
}

.theme-toggle:hover {
  color: var(--text-main) !important;
}

/* responsive */
@media (max-width: 768px) {
  .header-nav { display: none; }
}
</style>
