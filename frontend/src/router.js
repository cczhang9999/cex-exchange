import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  { path: '/', component: () => import('./components/Home.vue') },
  { path: '/login', component: () => import('./components/Login.vue') },
  { path: '/register', component: () => import('./components/Register.vue') },
  { path: '/assets', component: () => import('./components/Assets.vue') },
  { path: '/trade', component: () => import('./components/Trade.vue') },
  { path: '/orders', component: () => import('./components/Orders.vue') },
  { path: '/admin', component: () => import('./components/Admin.vue') },
  { path: '/market', component: () => import('./components/RealTimeMarket.vue') },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

export default router 