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
 * useWebSocket - WebSocket 组合式 Hook
 *
 * 提供在 Vue 组件中方便使用 WebSocket 的接口
 * - 自动在组件销毁时清理事件监听
 * - 封装标签组订阅/退订
 */
import { onUnmounted } from 'vue'
import { storeToRefs } from 'pinia'
import { useWebSocketStore } from '@/store/modules/websocket'

export function useWebSocket() {
  const wsStore = useWebSocketStore()
  const { connected, connecting, serverMonitorData } = storeToRefs(wsStore)

  // 本组件注册的清理函数
  const cleanups: Array<() => void> = []
  // 本组件加入的标签组
  const joinedTags: string[] = []

  /**
   * 监听某个 WebSocket 事件
   * 组件销毁时自动移除
   */
  function onEvent(event: string, listener: (data: any) => void) {
    const cleanup = wsStore.on(event, listener)
    cleanups.push(cleanup)
    return cleanup
  }

  /**
   * 加入标签组（组件销毁时自动退出）
   */
  function joinTag(tag: string) {
    wsStore.joinTag(tag)
    joinedTags.push(tag)
  }

  /**
   * 手动离开标签组
   */
  function quitTag(tag: string) {
    wsStore.quitTag(tag)
    const idx = joinedTags.indexOf(tag)
    if (idx !== -1) {
      joinedTags.splice(idx, 1)
    }
  }

  /**
   * 发送消息
   */
  function send(event: string, data?: Record<string, any>) {
    wsStore.send(event, data)
  }

  // 组件销毁时自动清理
  onUnmounted(() => {
    // 清理事件监听
    cleanups.forEach((fn) => fn())
    cleanups.length = 0

    // 退出标签组
    joinedTags.forEach((tag) => {
      wsStore.quitTag(tag)
    })
    joinedTags.length = 0
  })

  return {
    // 响应式状态
    connected,
    connecting,
    serverMonitorData,
    // 方法
    onEvent,
    joinTag,
    quitTag,
    send
  }
}
