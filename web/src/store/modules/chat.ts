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
 * 聊天 Store
 * 管理会话列表、消息、联系人、未读数等状态
 */
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import {
  fetchChatSessions,
  fetchChatMessages,
  fetchChatSend,
  fetchChatRead,
  fetchChatContacts,
  fetchChatSessionCreate,
  fetchChatSessionDelete,
  fetchChatUnreadTotal,
  fetchChatGroupUpdate,
  type ChatSessionItem,
  type ChatMessageItem,
  type ChatContactItem,
} from '@/api/backend/system/chat'
import { useWebSocketStore } from '@/store/modules/websocket'
import { useUserStore } from '@/store/modules/user'

export const useChatStore = defineStore('chat', () => {
  // ==================== 状态 ====================

  /** 会话列表 */
  const sessions = ref<ChatSessionItem[]>([])
  /** 当前会话ID */
  const currentSessionId = ref<number>(0)
  /** 当前会话的消息列表 */
  const messages = ref<ChatMessageItem[]>([])
  /** 是否还有更多历史消息 */
  const hasMore = ref(false)
  /** 联系人列表 */
  const contacts = ref<ChatContactItem[]>([])
  /** 聊天未读总数 */
  const totalUnread = ref(0)
  /** 加载状态 */
  const loading = ref(false)
  /** 消息加载状态 */
  const messagesLoading = ref(false)
  /** WebSocket 监听是否已初始化 */
  const wsListenerInited = ref(false)
  /** 已处理过的消息ID（防重复） */
  const handledMsgIds = new Set<number>()
  const userStore = useUserStore()

  // ==================== 计算属性 ====================

  /** 当前会话信息 */
  const currentSession = computed(() => {
    return sessions.value.find(s => s.id === currentSessionId.value) || null
  })

  const rememberMsgId = (id: number) => {
    if (!id) return
    handledMsgIds.add(id)
    // 防止集合无限增长
    if (handledMsgIds.size > 2000) {
      const first = handledMsgIds.values().next()
      if (!first.done) handledMsgIds.delete(first.value)
    }
  }

  const appendMessageIfNotExists = (msg: ChatMessageItem): boolean => {
    if (msg.id && handledMsgIds.has(msg.id)) return false
    messages.value.push(msg)
    rememberMsgId(msg.id)
    return true
  }

  // ==================== 会话操作 ====================

  /** 加载会话列表 */
  const loadSessions = async () => {
    try {
      loading.value = true
      const res = await fetchChatSessions()
      const data = res as any
      sessions.value = data?.list || []
    } catch { /* ignore */ } finally {
      loading.value = false
    }
  }

  /** 创建会话 */
  const createSession = async (type: number, userIds: number[], name?: string): Promise<number> => {
    try {
      const res = await fetchChatSessionCreate({ type, userIds, name })
      const data = res as any
      const sessionId = data?.sessionId || 0
      if (sessionId) {
        await loadSessions()
      }
      return sessionId
    } catch {
      return 0
    }
  }

  /** 删除会话 */
  const deleteSession = async (sessionId: number) => {
    try {
      await fetchChatSessionDelete(sessionId)
      sessions.value = sessions.value.filter(s => s.id !== sessionId)
      if (currentSessionId.value === sessionId) {
        currentSessionId.value = 0
        messages.value = []
      }
    } catch { /* ignore */ }
  }

  // ==================== 消息操作 ====================

  /** 加载消息列表 */
  const loadMessages = async (sessionId: number, loadMore = false) => {
    if (!sessionId) return
    try {
      messagesLoading.value = true
      const lastMsgId = loadMore && messages.value.length > 0 ? messages.value[0].id : undefined
      const res = await fetchChatMessages({ sessionId, lastMsgId, pageSize: 30 })
      const data = res as any
      const list: ChatMessageItem[] = data?.list || []
      hasMore.value = data?.hasMore || false

      if (loadMore) {
        messages.value = [...list, ...messages.value]
      } else {
        messages.value = list
      }
      list.forEach((m) => rememberMsgId(m.id))
    } catch { /* ignore */ } finally {
      messagesLoading.value = false
    }
  }

  /** 进入会话 */
  const enterSession = async (sessionId: number) => {
    currentSessionId.value = sessionId
    messages.value = []
    hasMore.value = false
    await loadMessages(sessionId)
    await markRead(sessionId)
  }

  /** 离开会话 */
  const leaveSession = () => {
    currentSessionId.value = 0
    messages.value = []
    hasMore.value = false
  }

  /** 发送消息 */
  const sendMessage = async (sessionId: number, type: number, content: string, senderName: string, senderAvatar: string) => {
    const res = await fetchChatSend({ sessionId, type, content })
    const data = res as any
    const now = Math.floor(Date.now() / 1000)
    const tempId = Date.now() + Math.floor(Math.random() * 1000)

    // 发送成功后本地追加消息（后端不再推给发送者自己）
    const msg: ChatMessageItem = {
      id: data?.id || tempId, // 兜底用毫秒级临时ID，避免同秒冲突
      sessionId,
      senderId: 0, // 0 = 自己发的
      senderName,
      senderAvatar,
      type,
      content,
      createdAt: data?.createdAt || now,
    }
    appendMessageIfNotExists(msg)

    // 更新会话列表中的最后消息
    const session = sessions.value.find(s => s.id === sessionId)
    if (session) {
      session.lastMessage = type === 2 ? '[图片]' : content
      session.lastMessageTime = msg.createdAt
      // 排到最前面
      const idx = sessions.value.indexOf(session)
      if (idx > 0) {
        sessions.value.splice(idx, 1)
        sessions.value.unshift(session)
      }
    }

    // 不做整段刷新，走 WebSocket + 本地增量，避免闪烁与重复请求
  }

  /** 标记已读 */
  const markRead = async (sessionId: number) => {
    try {
      await fetchChatRead(sessionId)
      const session = sessions.value.find(s => s.id === sessionId)
      if (session && session.unreadCount > 0) {
        session.unreadCount = 0
      }
      // 从服务端同步准确的未读总数
      await refreshUnreadTotal()
    } catch { /* ignore */ }
  }

  // ==================== 联系人操作 ====================

  /** 加载联系人 */
  const loadContacts = async () => {
    try {
      const res = await fetchChatContacts()
      const data = res as any
      contacts.value = data?.list || []
    } catch { /* ignore */ }
  }

  // ==================== 群聊操作 ====================

  /** 更新群聊 */
  const updateGroup = async (sessionId: number, params: { name?: string; addUsers?: number[]; delUsers?: number[] }) => {
    try {
      await fetchChatGroupUpdate({ sessionId, ...params })
      await loadSessions()
    } catch { /* ignore */ }
  }

  // ==================== 未读数 ====================

  /** 刷新未读总数 */
  const refreshUnreadTotal = async () => {
    try {
      const res = await fetchChatUnreadTotal()
      const data = res as any
      totalUnread.value = data?.total || 0
    } catch { /* ignore */ }
  }

  // ==================== WebSocket 监听 ====================

  /** 初始化 WebSocket 监听（防止重复注册） */
  const initWsListener = () => {
    if (wsListenerInited.value) return
    wsListenerInited.value = true
    const wsStore = useWebSocketStore()

    // 监听新消息
    wsStore.on('chat/message', (data: any) => {
      const msgId = Number(data.id || 0)
      const sessionId = Number(data.sessionId || 0)
      const senderId = Number(data.senderId || 0)
      const msg: ChatMessageItem = {
        id: msgId,
        sessionId,
        senderId,
        senderName: data.senderName,
        senderAvatar: data.senderAvatar,
        type: Number(data.type || 1),
        content: data.content,
        createdAt: Number(data.createdAt || Math.floor(Date.now() / 1000)),
      }

      // 防重复：同一条消息ID只处理一次
      if (msg.id && handledMsgIds.has(msg.id)) return

      // 调试日志：可直观看到每条消息是否被重复处理
      console.debug('[chat][ws][message]', {
        id: msg.id,
        sessionId: msg.sessionId,
        senderId: msg.senderId,
        currentSessionId: currentSessionId.value,
        currentUserId: userStore.getUserInfo?.userId
      })

      // 如果是当前正在查看的会话，追加消息并标记已读
      if (msg.sessionId === currentSessionId.value) {
        const appended = appendMessageIfNotExists(msg)
        if (appended) {
          markRead(msg.sessionId)
        }
      } else {
        // 列表态也记录已处理ID，避免重复推送导致徽章暴涨
        rememberMsgId(msg.id)
      }

      // 更新会话列表中的最后消息和未读数
      updateSessionFromMessage(msg)
    })

    // 监听已读回执
    wsStore.on('chat/read', (_data: any) => {
      // 暂时不做特殊处理，后续可以显示已读状态
    })
  }

  /** 根据新消息更新会话列表 */
  const updateSessionFromMessage = (msg: ChatMessageItem) => {
    const session = sessions.value.find(s => s.id === msg.sessionId)
    if (session) {
      session.lastMessage = msg.type === 2 ? '[图片]' : msg.content
      session.lastMessageTime = msg.createdAt
      // 不是当前会话才增加未读数
      if (msg.sessionId !== currentSessionId.value) {
        session.unreadCount++
        totalUnread.value++
      }
      // 把这个会话排到最前面
      const idx = sessions.value.indexOf(session)
      if (idx > 0) {
        sessions.value.splice(idx, 1)
        sessions.value.unshift(session)
      }
    } else {
      // 新会话，重新加载列表
      loadSessions()
      refreshUnreadTotal()
    }
  }

  return {
    // 状态
    sessions,
    currentSessionId,
    currentSession,
    messages,
    hasMore,
    contacts,
    totalUnread,
    loading,
    messagesLoading,
    // 会话操作
    loadSessions,
    createSession,
    deleteSession,
    enterSession,
    leaveSession,
    // 消息操作
    loadMessages,
    sendMessage,
    markRead,
    // 联系人
    loadContacts,
    // 群聊
    updateGroup,
    // 未读数
    refreshUnreadTotal,
    // WebSocket
    initWsListener,
  }
})
