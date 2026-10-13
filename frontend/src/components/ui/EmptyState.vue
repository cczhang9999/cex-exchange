<template>
  <div class="empty-state">
    <div class="empty-state__icon">
      <el-icon><component :is="icon" /></el-icon>
    </div>
    <div class="empty-state__title">{{ title }}</div>
    <div v-if="description" class="empty-state__desc">{{ description }}</div>
    <slot />
  </div>
</template>

<script setup>
defineProps({
  icon: { type: [String, Object], default: 'Document' },
  title: { type: String, default: '暂无数据' },
  description: { type: String, default: '' },
})
</script>
