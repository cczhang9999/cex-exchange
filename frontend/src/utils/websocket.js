// WebSocket客户端工具类
class WebSocketClient {
  constructor(url) {
    this.url = url
    this.ws = null
    this.isConnected = false
    this.reconnectAttempts = 0
    this.maxReconnectAttempts = 5
    this.reconnectInterval = 3000
    this.subscribers = new Map()
    this.messageHandlers = new Map()
  }

  // 连接WebSocket
  connect() {
    try {
      this.ws = new WebSocket(this.url)
      
      this.ws.onopen = () => {
        console.log('WebSocket连接成功')
        this.isConnected = true
        this.reconnectAttempts = 0
        
        // 重新订阅之前的频道
        this.subscribers.forEach((callback, symbol) => {
          this.subscribe(symbol, callback)
        })
      }

      this.ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data)
          this.handleMessage(data)
        } catch (error) {
          console.error('解析WebSocket消息失败:', error)
        }
      }

      this.ws.onclose = () => {
        console.log('WebSocket连接断开')
        this.isConnected = false
        this.reconnect()
      }

      this.ws.onerror = (error) => {
        console.error('WebSocket错误:', error)
        this.isConnected = false
      }
    } catch (error) {
      console.error('WebSocket连接失败:', error)
      this.reconnect()
    }
  }

  // 重连机制
  reconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++
      console.log(`尝试重连 (${this.reconnectAttempts}/${this.maxReconnectAttempts})`)
      
      setTimeout(() => {
        this.connect()
      }, this.reconnectInterval)
    } else {
      console.error('WebSocket重连失败，已达到最大重连次数')
    }
  }

  // 发送消息
  send(message) {
    if (this.isConnected && this.ws) {
      this.ws.send(JSON.stringify(message))
    } else {
      console.warn('WebSocket未连接，无法发送消息')
    }
  }

  // 订阅行情数据
  subscribe(symbol, callback) {
    this.subscribers.set(symbol, callback)
    
    if (this.isConnected) {
      this.send({
        type: 'subscribe',
        symbol: symbol
      })
    }
  }

  // 取消订阅
  unsubscribe(symbol) {
    this.subscribers.delete(symbol)
  }

  // 处理接收到的消息
  handleMessage(data) {
    const { type, symbol, data: messageData, timestamp } = data

    switch (type) {
      case 'welcome':
        console.log('WebSocket连接成功:', messageData)
        break
        
      case 'pong':
        // 心跳响应
        break
        
      case 'orderbook':
        this.notifySubscribers(symbol, 'orderbook', messageData)
        break
        
      case 'trade':
        this.notifySubscribers(symbol, 'trade', messageData)
        break
        
      case 'price_change':
        this.notifySubscribers(symbol, 'price_change', messageData)
        break
        
      case 'kline':
        this.notifySubscribers(symbol, 'kline', messageData)
        break
        
      default:
        console.log('未知消息类型:', type, data)
    }
  }

  // 通知订阅者
  notifySubscribers(symbol, dataType, data) {
    const callback = this.subscribers.get(symbol)
    if (callback) {
      callback(dataType, data)
    }
  }

  // 发送心跳
  startHeartbeat() {
    setInterval(() => {
      if (this.isConnected) {
        this.send({ type: 'ping' })
      }
    }, 30000) // 每30秒发送一次心跳
  }

  // 断开连接
  disconnect() {
    if (this.ws) {
      this.ws.close()
    }
    this.isConnected = false
  }
}

// 创建全局WebSocket实例
const wsClient = new WebSocketClient('ws://localhost:8080/ws')

// 导出WebSocket客户端
export default wsClient 