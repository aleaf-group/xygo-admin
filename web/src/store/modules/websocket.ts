// +----------------------------------------------------------------------
// | XYGo Admin [ Vue3 + GoFrame 企业级中后台管理系统 ]
// +----------------------------------------------------------------------
// | Copyright (c) 2026 大连星韵网络科技有限公司 All rights reserved.
// +----------------------------------------------------------------------
// | Licensed ( https://opensource.org/licenses/MIT )
// +----------------------------------------------------------------------
// | Author: 喜羊羊 <751300685@qq.com>
// +----------------------------------------------------------------------

/**
 * WebSocket 状态管理
 *
 * 基于已有的 WebSocketClient 封装，提供响应式状态管理
 * - 连接/断开管理
 * - 事件监听与分发
 * - 监控数据推送接收
 * - 踢人通知处理
 */
import { defineStore } from 'pinia'
import { ref, shallowRef } from 'vue'
import { ElNotification } from 'element-plus'
import WebSocketClient from '@/utils/socket'
import { useUserStore } from '@/store/modules/user'

// WebSocket 消息类型
export interface WsMessage {
  event: string
  data: any
  code: number
  errorMsg?: string
  timestamp: number
}

// 事件监听器类型
type EventListener = (data: any) => void

export const useWebSocketStore = defineStore('websocket', () => {
  // 连接状态
  const connected = ref(false)
  const connecting = ref(false)

  // WebSocket 客户端实例
  const client = shallowRef<WebSocketClient | null>(null)

  // 事件监听器 Map
  const listeners = new Map<string, Set<EventListener>>()

  // 最新的监控数据（用于服务器监控页面）
  const serverMonitorData = shallowRef<any>(null)

  /**
   * 构建 WebSocket URL
   */
  function buildWsUrl(): string {
    const userStore = useUserStore()
    const token = userStore.accessToken

    // 开发环境通过 vite proxy 代理
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = window.location.host
    return `${protocol}//${host}/socket/?token=${encodeURIComponent(token)}`
  }

  /**
   * 连接 WebSocket
   */
  function connect() {
    const userStore = useUserStore()
    if (!userStore.isLogin || !userStore.accessToken) {
      return
    }

    if (connected.value || connecting.value) {
      return
    }

    connecting.value = true

    const url = buildWsUrl()

    client.value = WebSocketClient.getInstance({
      url,
      messageHandler: handleMessage,
      reconnectInterval: 5000,
      heartbeatInterval: 10000,
      pingInterval: 30000,
      maxReconnectAttempts: 20
    })

    client.value.init()
  }

  /**
   * 断开连接
   */
  function disconnect() {
    if (client.value) {
      client.value.close(true)
    }
    WebSocketClient.destroyInstance()
    client.value = null
    connected.value = false
    connecting.value = false
    serverMonitorData.value = null
    listeners.clear()
  }

  /**
   * 处理收到的消息
   */
  function handleMessage(event: MessageEvent) {
    try {
      const msg: WsMessage = JSON.parse(event.data)

      // 连接成功
      if (msg.event === 'connected') {
        connected.value = true
        connecting.value = false
        return
      }

      // pong 心跳响应
      if (msg.event === 'pong') {
        return
      }

      // 被踢下线
      if (msg.event === 'kicked') {
        handleKicked()
        return
      }

      // 监控数据
      if (msg.event === 'monitor/server') {
        serverMonitorData.value = msg.data
      }

      // 系统告警（队列积压等）
      if (msg.event === 'system/alert') {
        const payload = msg.data || {}
        ElNotification({
          title: payload.title || '系统告警',
          message: payload.content || '收到一条系统告警',
          type: 'warning',
          duration: 0, // 不自动关闭，需手动点
          position: 'top-right',
        })
      }

      // 分发给事件监听器
      const eventListeners = listeners.get(msg.event)
      if (eventListeners) {
        eventListeners.forEach((listener) => {
          try {
            listener(msg.data)
          } catch (e) {
            console.error(`WebSocket listener error for event "${msg.event}":`, e)
          }
        })
      }
    } catch (e) {
      // 非 JSON 消息忽略
    }
  }

  /**
   * 处理被踢下线
   */
  function handleKicked() {
    disconnect()
    const userStore = useUserStore()
    userStore.logOut({ callApi: false, redirect: true })
  }

  /**
   * 发送事件消息
   */
  function send(event: string, data?: Record<string, any>) {
    if (client.value) {
      client.value.send(JSON.stringify({ event, data }))
    }
  }

  /**
   * 加入标签组
   */
  function joinTag(tag: string) {
    send('join', { tag })
  }

  /**
   * 离开标签组
   */
  function quitTag(tag: string) {
    send('quit', { tag })
  }

  /**
   * 注册事件监听器
   * @returns 取消监听的函数
   */
  function on(event: string, listener: EventListener): () => void {
    if (!listeners.has(event)) {
      listeners.set(event, new Set())
    }
    listeners.get(event)!.add(listener)

    return () => {
      const set = listeners.get(event)
      if (set) {
        set.delete(listener)
        if (set.size === 0) {
          listeners.delete(event)
        }
      }
    }
  }

  /**
   * 移除事件监听器
   */
  function off(event: string, listener?: EventListener) {
    if (!listener) {
      listeners.delete(event)
    } else {
      const set = listeners.get(event)
      if (set) {
        set.delete(listener)
      }
    }
  }

  return {
    // 状态
    connected,
    connecting,
    serverMonitorData,
    // 方法
    connect,
    disconnect,
    send,
    joinTag,
    quitTag,
    on,
    off
  }
})
