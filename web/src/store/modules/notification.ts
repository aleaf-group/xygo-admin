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
 * 通知消息 Store
 * 管理站内通知、公告、私信的状态
 */
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import {
  fetchNoticePull,
  fetchNoticeUnreadCount,
  fetchNoticeRead,
  fetchNoticeReadAll,
  type MessageItem,
  type UnreadCountItem,
} from '@/api/backend/system/notice'
import { useWebSocketStore } from '@/store/modules/websocket'

export const useNotificationStore = defineStore('notification', () => {
  // 消息列表
  const messages = ref<MessageItem[]>([])
  // 未读数
  const unreadList = ref<UnreadCountItem[]>([])
  const totalUnread = ref(0)

  // 按类型分组
  const notifyMessages = computed(() => messages.value.filter(m => m.type === 1))
  const announceMessages = computed(() => messages.value.filter(m => m.type === 2))
  const letterMessages = computed(() => messages.value.filter(m => m.type === 3))

  // 各类型未读数
  const notifyUnread = computed(() => unreadList.value.find(u => u.type === 1)?.count || 0)
  const announceUnread = computed(() => unreadList.value.find(u => u.type === 2)?.count || 0)
  const letterUnread = computed(() => unreadList.value.find(u => u.type === 3)?.count || 0)

  // 拉取消息
  const pullMessages = async () => {
    try {
      const res = await fetchNoticePull()
      const data = res as any
      messages.value = data?.list || []
      unreadList.value = data?.unread || []
      totalUnread.value = unreadList.value.reduce((s, u) => s + u.count, 0)
    } catch { /* ignore */ }
  }

  // 刷新未读数
  const refreshUnread = async () => {
    try {
      const res = await fetchNoticeUnreadCount()
      const data = res as any
      unreadList.value = data?.list || []
      totalUnread.value = data?.total || 0
    } catch { /* ignore */ }
  }

  // 标记已读
  const markRead = async (id: number) => {
    try {
      await fetchNoticeRead(id)
      const msg = messages.value.find(m => m.id === id)
      if (msg) msg.isRead = true
      await refreshUnread()
    } catch { /* ignore */ }
  }

  // 全部已读
  const markAllRead = async (type?: number) => {
    try {
      await fetchNoticeReadAll(type)
      messages.value.forEach(m => {
        if (!type || m.type === type) m.isRead = true
      })
      await refreshUnread()
    } catch { /* ignore */ }
  }

  // 监听 WebSocket 推送的新通知
  const initWsListener = () => {
    const wsStore = useWebSocketStore()
    wsStore.on('notice', (_data: any) => {
      // 收到新通知，重新拉取消息
      pullMessages()
    })
  }

  return {
    messages,
    unreadList,
    totalUnread,
    notifyMessages,
    announceMessages,
    letterMessages,
    notifyUnread,
    announceUnread,
    letterUnread,
    pullMessages,
    refreshUnread,
    markRead,
    markAllRead,
    initWsListener,
  }
})
