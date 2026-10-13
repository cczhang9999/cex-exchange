<template>
  <header class="page-head">
    <div class="page-head__title">
      <h1>
        <span v-if="icon" class="page-head__icon" :class="`tone-${tone}`">
          <el-icon><component :is="icon" /></el-icon>
        </span>
        {{ title }}
      </h1>
      <span v-if="subtitle" class="page-head__sub">{{ subtitle }}</span>
    </div>
    <div class="page-head__actions">
      <slot name="actions" />
    </div>
  </header>
</template>

<script setup>
defineProps({
  title: { type: String, default: '' },
  subtitle: { type: String, default: '' },
  icon: { type: [String, Object], default: '' },
  tone: { type: String, default: 'primary' },
})
</script>

<style scoped>
.page-head__icon {
  width: 36px;
  height: 36px;
  border-radius: 11px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 19px;
  color: #fff;
  box-shadow: 0 6px 16px -8px currentColor;
}
.tone-primary { background: var(--primary-grad); }
.tone-success { background: var(--success-grad); }
.tone-warning { background: var(--warning-grad); }
.tone-danger { background: var(--danger-grad); }
.tone-info { background: var(--info-grad); }
.tone-violet { background: var(--violet-grad); }
</style>
