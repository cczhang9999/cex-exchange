<template>
  <header class="app-header">
    <div class="header-left" @click="$emit('nav', 'home')">
      <span class="logo-mark">
        <el-icon><TrendCharts /></el-icon>
      </span>
      <span class="brand">
        <span class="brand-name">CEX Exchange</span>
        <span class="brand-sub">专业数字资产交易平台</span>
      </span>
    </div>

    <nav class="header-nav">
      <button
        v-for="item in navItems"
        :key="item.key"
        class="nav-link"
        :class="{ active: active === item.key }"
        @click="$emit('nav', item.key)"
      >
        {{ item.label }}
      </button>
    </nav>

    <div class="header-actions">
      <el-tooltip :content="isDark ? '切换到亮色' : '切换到暗色'" placement="bottom">
        <button class="icon-btn" @click="toggleTheme">
          <el-icon><component :is="isDark ? Sunny : Moon" /></el-icon>
        </button>
      </el-tooltip>

      <div class="user-chip">
        <span class="avatar">
          <el-icon><UserFilled /></el-icon>
        </span>
        <span class="username">{{ userStore.username || '未登录' }}</span>
      </div>

      <el-button v-if="userStore.username" size="small" class="logout-btn" @click="handleLogout">
        登出
      </el-button>
      <el-button v-else size="small" type="primary" @click="$emit('nav', 'login')">登录</el-button>
    </div>
  </header>
</template>

<script setup>
import { Sunny, Moon, UserFilled, TrendCharts } from '@element-plus/icons-vue'
import { useUserStore } from '../stores'
import { useRouter } from 'vue-router'
import { ElMessageBox, ElMessage } from 'element-plus'
import { isDark, toggleTheme } from '../utils/theme'

defineProps({
  active: { type: String, default: 'home' },
})
defineEmits(['nav'])

const userStore = useUserStore()
const router = useRouter()

const navItems = [
  { key: 'home', label: '首页概览' },
  { key: 'trade', label: '交易中心' },
  { key: 'assets', label: '资产管理' },
  { key: 'orders', label: '订单管理' },
  { key: 'realtime', label: '实时行情' },
  { key: 'admin', label: '后台管理' },
]

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
    /* 用户取消 */
  }
}
</script>

<style scoped>
.app-header {
  height: var(--header-h) !important;
  padding: 0 22px !important;
  display: flex !important;
  align-items: center !important;
  justify-content: space-between !important;
  gap: 16px;
  background: color-mix(in srgb, var(--surface-1), transparent 25%) !important;
  backdrop-filter: blur(18px) saturate(150%);
  -webkit-backdrop-filter: blur(18px) saturate(150%);
  border-bottom: 1px solid var(--border-color) !important;
  box-shadow: 0 1px 0 rgba(15, 23, 42, 0.02);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
  user-select: none;
  flex-shrink: 0;
}

.logo-mark {
  width: 34px;
  height: 34px;
  border-radius: 10px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 19px;
  color: #fff;
  background: var(--primary-grad);
  box-shadow: 0 6px 16px -6px var(--primary-glow);
}

.brand { display: flex; flex-direction: column; line-height: 1.15; }
.brand-name { font-weight: 750; font-size: 1.02rem; color: var(--text-main); letter-spacing: -0.02em; }
.brand-sub { font-size: 0.66rem; color: var(--text-secondary); font-weight: 500; }

.header-nav { display: flex; align-items: center; gap: 2px; }
.nav-link {
  border: none;
  background: transparent;
  color: var(--text-muted);
  font-weight: 550;
  font-size: 0.86rem;
  padding: 7px 13px;
  border-radius: var(--radius-input);
  cursor: pointer;
  transition: var(--transition-snap);
  font-family: inherit;
  white-space: nowrap;
}
.nav-link:hover { color: var(--text-main); background: var(--surface-sunken); }
.nav-link.active {
  color: var(--primary);
  background: var(--primary-glow);
  font-weight: 700;
}

.header-actions { display: flex; align-items: center; gap: 10px; flex-shrink: 0; }

.icon-btn {
  width: 34px;
  height: 34px;
  border-radius: var(--radius-input);
  border: 1px solid var(--border-color);
  background: var(--surface-2);
  color: var(--text-muted);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  font-size: 16px;
  transition: var(--transition-snap);
}
.icon-btn:hover { color: var(--primary); border-color: var(--border-color-hover); background: var(--surface-1); }

.user-chip {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 12px 4px 4px;
  border-radius: var(--radius-pill);
  background: var(--surface-sunken);
  border: 1px solid var(--border-subtle);
}
.avatar {
  width: 26px;
  height: 26px;
  border-radius: 50%;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: var(--primary-grad);
  color: #fff;
  font-size: 14px;
}
.username { font-weight: 600; color: var(--text-main); font-size: 0.82rem; max-width: 120px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

.logout-btn {
  border-radius: var(--radius-input) !important;
  font-weight: 550 !important;
  border: 1px solid var(--border-color) !important;
  background: transparent !important;
  color: var(--text-muted) !important;
}
.logout-btn:hover { color: var(--danger) !important; border-color: var(--danger) !important; background: var(--danger-glow) !important; }

@media (max-width: 1100px) {
  .header-nav { display: none; }
  .brand-sub { display: none; }
}
@media (max-width: 640px) {
  .username { display: none; }
}
</style>
