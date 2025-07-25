// src/api/api.js
import axios from 'axios'

// 登录
export const login = (form) => axios.post('/api/login', form)
// 注册
export const register = (form) => axios.post('/api/register', form)
// 下单
export const placeOrder = (order) => axios.post('/api/order', order)
// 获取订单簿
export const getOrderbook = (symbol) => axios.get('/api/orderbook', { params: { symbol } })
// 获取成交记录
export const getTrades = (symbol) => axios.get('/api/trades', { params: { symbol } })
// 其他接口可按需添加... 