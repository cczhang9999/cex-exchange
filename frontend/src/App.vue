<script setup>
// 全局应用逻辑可以在这里添加
</script>

<template>
  <div id="app">
    <router-view />
  </div>
</template>

<style>
/* 全局样式重置 */
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

html, body {
  height: 100%;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', 'Oxygen',
    'Ubuntu', 'Cantarell', 'Fira Sans', 'Droid Sans', 'Helvetica Neue',
    sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  background-color: #f0f2f5;
  overflow: hidden; /* 防止页面滚动 */
}

#app {
  height: 100vh;
  width: 100%;
  overflow: hidden; /* 确保应用占满全屏 */
}

/* 滚动条样式优化 */
::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 3px;
}

::-webkit-scrollbar-thumb {
  background: #c1c1c1;
  border-radius: 3px;
}

::-webkit-scrollbar-thumb:hover {
  background: #a8a8a8;
}

/* Element Plus 组件样式优化 */
.el-container {
  height: 100% !important;
}

.el-main {
  padding: 20px !important;
  overflow-y: auto !important;
  height: 100% !important;
}

.el-card {
  border-radius: 8px;
  border: none;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.el-button {
  border-radius: 6px;
  font-weight: 500;
}

.el-input {
  border-radius: 6px;
}

.el-table {
  border-radius: 8px;
  overflow: hidden;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .el-aside {
    width: 100% !important;
    position: fixed;
    top: 60px;
    left: -100%;
    height: calc(100vh - 60px);
    z-index: 1000;
    transition: left 0.3s;
  }
  
  .el-aside.show {
    left: 0;
  }
  
  .el-main {
    padding: 10px !important;
  }
  
  .el-header {
    padding: 0 15px !important;
  }
}

/* 动画效果 */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

/* 加载状态 */
.loading-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 200px;
}

/* 错误状态 */
.error-container {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  height: 200px;
  color: #f56c6c;
}

/* 成功状态 */
.success-container {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  height: 200px;
  color: #67c23a;
}

/* 工具类 */
.text-center {
  text-align: center;
}

.text-right {
  text-align: right;
}

.text-left {
  text-align: left;
}

.flex {
  display: flex;
}

.flex-center {
  display: flex;
  justify-content: center;
  align-items: center;
}

.flex-between {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.flex-column {
  display: flex;
  flex-direction: column;
}

.mt-10 {
  margin-top: 10px;
}

.mb-10 {
  margin-bottom: 10px;
}

.ml-10 {
  margin-left: 10px;
}

.mr-10 {
  margin-right: 10px;
}

.p-10 {
  padding: 10px;
}

.p-20 {
  padding: 20px;
}

/* 主题色彩 */
:root {
  --primary-color: #409eff;
  --success-color: #67c23a;
  --warning-color: #e6a23c;
  --danger-color: #f56c6c;
  --info-color: #909399;
  --text-primary: #303133;
  --text-regular: #606266;
  --text-secondary: #909399;
  --border-color: #dcdfe6;
  --background-color: #f0f2f5;
}

/* 暗色主题支持 */
@media (prefers-color-scheme: dark) {
  :root {
    --text-primary: #ffffff;
    --text-regular: #e5eaf3;
    --text-secondary: #a3a6ad;
    --border-color: #4c4d4f;
    --background-color: #141414;
  }
  
  body {
    background-color: var(--background-color);
    color: var(--text-primary);
  }
}
</style>
