<script setup>
// 全局应用逻辑可以在这里添加
</script>

<template>
  <div id="app">
    <router-view />
  </div>
</template>

<style>
/* Layout */
#app {
  width: 100vw;
  height: 100vh;
  overflow: hidden;
}

/* Global Transitions */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
