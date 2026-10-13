// WebSocket 客户端工具类（生产级）
// - 支持同一交易对被多个组件同时订阅（Map<symbol, Set<callback>>）
// - 指数退避重连 + 心跳保活 + 主动销毁
class WebSocketClient {
  constructor(url) {
    this.url = url
    this.ws = null
    this.isConnected = false
    this.reconnectAttempts = 0
    this.maxReconnectAttempts = 8
    this.baseReconnectInterval = 1500
    this.subscribers = new Map() // symbol -> Set<callback>
    this.heartbeatTimer = null
    this.reconnectTimer = null
    this.manuallyClosed = false
  }

  connect() {
    if (this.ws && (this.ws.readyState === WebSocket.OPEN || this.ws.readyState === WebSocket.CONNECTING)) {
      return
    }
    this.manuallyClosed = false

    try {
      this.ws = new WebSocket(this.url)

      this.ws.onopen = () => {
        this.isConnected = true
        this.reconnectAttempts = 0
        // 重连后重新订阅所有频道
        this.subscribers.forEach((_, symbol) => {
          this.send({ type: 'subscribe', symbol })
        })
      }

      this.ws.onmessage = (event) => {
        try {
          this.handleMessage(JSON.parse(event.data))
        } catch (error) {
          console.error('解析 WebSocket 消息失败:', error)
        }
      }

      this.ws.onclose = () => {
        this.isConnected = false
        this.stopHeartbeat()
        if (!this.manuallyClosed) this.reconnect()
      }

      this.ws.onerror = () => {
        this.isConnected = false
      }
    } catch (error) {
      console.error('WebSocket 连接失败:', error)
      this.reconnect()
    }
  }

  // 指数退避重连，避免雪崩式重试
  reconnect() {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.warn('WebSocket 已达到最大重连次数，停止重连')
      return
    }
    const delay = Math.min(this.baseReconnectInterval * 2 ** this.reconnectAttempts, 20000)
    this.reconnectAttempts++
    clearTimeout(this.reconnectTimer)
    this.reconnectTimer = setTimeout(() => this.connect(), delay)
  }

  send(message) {
    if (this.isConnected && this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message))
    }
  }

  // 订阅行情数据（同一 symbol 可注册多个回调）
  subscribe(symbol, callback) {
    if (!this.subscribers.has(symbol)) {
      this.subscribers.set(symbol, new Set())
      if (this.isConnected) this.send({ type: 'subscribe', symbol })
    }
    this.subscribers.get(symbol).add(callback)
  }

  // 取消订阅；不传 callback 则移除该 symbol 的全部回调
  unsubscribe(symbol, callback) {
    const set = this.subscribers.get(symbol)
    if (!set) return
    if (callback) {
      set.delete(callback)
    } else {
      set.clear()
    }
    if (set.size === 0) this.subscribers.delete(symbol)
  }

  handleMessage(data) {
    const { type, symbol, data: messageData } = data
    switch (type) {
      case 'welcome':
      case 'pong':
        break
      case 'orderbook':
      case 'trade':
      case 'price_change':
      case 'kline':
        this.notifySubscribers(symbol, type, messageData)
        break
      default:
        break
    }
  }

  // 把信封中的 symbol 一并透传给回调，便于调用方按交易对归类
  notifySubscribers(symbol, dataType, data) {
    const set = this.subscribers.get(symbol)
    if (!set) return
    set.forEach((cb) => {
      try {
        cb(dataType, data, symbol)
      } catch (error) {
        console.error('WebSocket 订阅回调异常:', error)
      }
    })
  }

  startHeartbeat() {
    this.stopHeartbeat()
    this.heartbeatTimer = setInterval(() => {
      if (this.isConnected) this.send({ type: 'ping' })
    }, 30000)
  }

  stopHeartbeat() {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer)
      this.heartbeatTimer = null
    }
  }

  disconnect() {
    this.manuallyClosed = true
    this.stopHeartbeat()
    clearTimeout(this.reconnectTimer)
    if (this.ws) {
      try {
        this.ws.close()
      } catch (error) {
        /* noop */
      }
    }
    this.ws = null
    this.isConnected = false
  }
}

// 根据当前页面协议自动选择 ws / wss；Vite 开发服务器（5173）走本机后端 8080
const wsProtocol = typeof window !== 'undefined' && window.location.protocol === 'https:' ? 'wss' : 'ws'
const isViteDev =
  typeof window !== 'undefined' &&
  window.location.port === '5173' &&
  ['localhost', '127.0.0.1', '0.0.0.0'].includes(window.location.hostname)
const wsHost =
  isViteDev
    ? 'localhost:8080'
    : typeof window !== 'undefined'
      ? window.location.host
      : 'localhost:8080'

const wsClient = new WebSocketClient(`${wsProtocol}://${wsHost}/ws`)

export default wsClient
