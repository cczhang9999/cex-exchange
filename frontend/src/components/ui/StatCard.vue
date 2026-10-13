<template>
  <div class="stat-card rise-in" :style="{ '--accent': accent }">
    <div class="stat-card__top">
      <div class="stat-card__label">{{ label }}</div>
      <div v-if="icon" class="stat-card__icon" :class="`ic-${tone}`">
        <el-icon><component :is="icon" /></el-icon>
      </div>
    </div>

    <div class="stat-card__value" :class="{ 'is-loading': loading }">
      <span v-if="loading" class="skeleton" style="display:inline-block;width:60%;height:26px;border-radius:8px;"></span>
      <template v-else>{{ value }}</template>
    </div>

    <div v-if="sub || delta !== null" class="stat-card__meta">
      <span
        v-if="delta !== null && delta !== undefined"
        class="badge"
        :class="delta >= 0 ? 'badge-up' : 'badge-down'"
      >
        <el-icon><component :is="delta >= 0 ? 'Top' : 'Bottom'" /></el-icon>
        {{ Math.abs(delta).toFixed(2) }}%
      </span>
      <span v-if="sub" class="text-muted text-sm">{{ sub }}</span>
    </div>

    <div v-if="spark && spark.length > 1" class="stat-card__spark">
      <Sparkline :data="spark" :width="220" :height="40" />
    </div>
  </div>
</template>

<script setup>
import Sparkline from './Sparkline.vue'

defineProps({
  label: { type: String, default: '' },
  value: { type: [String, Number], default: '—' },
  sub: { type: String, default: '' },
  delta: { type: Number, default: null },
  icon: { type: [String, Object], default: '' },
  // tone 决定图标底色
  tone: { type: String, default: 'primary' },
  accent: { type: String, default: 'var(--primary)' },
  spark: { type: Array, default: () => [] },
  loading: { type: Boolean, default: false },
})
</script>

<style scoped>
.stat-card {
  display: flex;
  flex-direction: column;
}
.stat-card__value.is-loading { min-height: 34px; display: flex; align-items: center; }
.stat-card__icon { color: #fff; }

.ic-primary { background: var(--primary-grad); }
.ic-success { background: var(--success-grad); }
.ic-danger { background: var(--danger-grad); }
.ic-warning { background: var(--warning-grad); }
.ic-info { background: var(--info-grad); }
.ic-violet { background: var(--violet-grad); }
</style>
