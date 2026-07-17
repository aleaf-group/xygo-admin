<!-- +----------------------------------------------------------------------
  | XYGo Admin [ Vue3 + GoFrame 企业级中后台管理系统 ]
  +----------------------------------------------------------------------
  | Copyright (c) 2026 大连星韵网络科技有限公司 All rights reserved.
  +----------------------------------------------------------------------
  | Licensed ( https://opensource.org/licenses/MIT )
  +----------------------------------------------------------------------
  | Author: 喜羊羊 <751300685@qq.com>
  +---------------------------------------------------------------------- -->
<!-- 聊天页（全屏版，使用真实数据 — 对接 chatStore + WebSocket） -->
<template>
  <div class="page-content flex !p-0 max-md:flex-col" :style="{ height: containerMinHeight }">
    <!-- 左侧会话列表 -->
    <div class="box-border w-90 h-full border-r border-g-300 max-md:w-full max-md:h-42 max-md:border-r-0 flex flex-col">
      <!-- 当前用户信息 + 搜索 -->
      <div class="p-5 pb-3 max-md:!hidden">
        <div class="flex-c gap-3 mb-3">
          <ElAvatar :size="42" :src="avatarUrl(myAvatar, myName)" />
          <div class="flex-1 min-w-0">
            <div class="text-base font-medium truncate">{{ myName }}</div>
            <div class="mt-0.5 text-xs text-g-500">在线</div>
          </div>
          <ElTooltip content="发起聊天" placement="bottom">
            <ArtIconButton icon="ri:chat-new-line" circle class="size-9 text-g-600" @click="showContacts = true" />
          </ElTooltip>
        </div>
        <ElInput v-model="searchQuery" placeholder="搜索会话" prefix-icon="Search" clearable />
      </div>

      <!-- 会话列表 -->
      <ElScrollbar class="flex-1">
        <div v-if="chatStore.loading" class="p-8 text-center text-g-500 text-sm">加载中...</div>
        <div v-else-if="filteredSessions.length === 0" class="p-8 text-center text-g-400 text-sm">暂无会话</div>
        <div
          v-for="session in filteredSessions"
          :key="session.id"
          class="flex-c p-3 mx-2 c-p rounded-lg tad-200 hover:bg-active-color/30 mb-0.5"
          :class="{ 'bg-active-color': chatStore.currentSessionId === session.id }"
          @click="selectSession(session)"
          @contextmenu.prevent="onSessionContextMenu(session, $event)"
        >
          <div class="relative mr-3 flex-shrink-0">
            <ElAvatar :size="40" :src="avatarUrl(session.avatar, session.name)" />
            <div v-if="session.type === 2" class="absolute -top-0.5 -right-0.5 size-4 rounded-full bg-theme/100 text-white text-[9px] flex items-center justify-center">
              <ArtSvgIcon icon="ri:group-line" class="text-[9px]" />
            </div>
            <span
              v-else-if="isSessionOnline(session)"
              class="absolute -bottom-0.5 -right-0.5 h-2.5 w-2.5 rounded-full border-2 border-white bg-green-500"
            ></span>
          </div>
          <div class="flex-1 min-w-0">
            <div class="flex-cb mb-1">
              <span class="text-sm font-medium truncate">{{ session.name }}</span>
              <span class="text-[11px] text-g-500 flex-shrink-0 ml-2">{{ formatTime(session.lastMessageTime) }}</span>
            </div>
            <div class="flex-cb">
              <span class="overflow-hidden text-xs text-g-500 text-ellipsis whitespace-nowrap flex-1">{{ getSessionSubtitle(session) || session.lastMessage || '暂无消息' }}</span>
              <ElBadge v-if="session.unreadCount > 0" :value="session.unreadCount" :max="99" class="ml-2 flex-shrink-0" />
            </div>
          </div>
        </div>
      </ElScrollbar>
    </div>

    <!-- 右侧聊天区域 -->
    <div class="box-border flex-1 h-full max-md:h-[calc(70%-30px)] flex flex-col">
      <!-- 无选中状态 -->
      <div v-if="!chatStore.currentSessionId" class="flex-1 flex items-center justify-center">
        <div class="text-center text-g-400">
          <ArtSvgIcon icon="ri:chat-3-line" class="text-5xl mb-3" />
          <p class="text-sm">选择一个会话开始聊天</p>
        </div>
      </div>

      <!-- 聊天面板 -->
      <template v-else>
        <!-- 顶栏 -->
        <div class="flex-cb pt-4 px-5 pb-0 mb-4">
          <div>
            <span class="text-base font-medium">{{ chatStore.currentSession?.name }}</span>
            <div v-if="chatStore.currentSession?.type === 2" class="text-xs text-g-500 mt-1">
              {{ chatStore.currentSession?.memberCount || 0 }} 人
            </div>
          </div>
          <div class="flex-c gap-2">
            <ElTooltip content="删除会话" placement="bottom">
              <ArtIconButton icon="ri:delete-bin-line" circle class="size-9 text-g-500" @click="handleDeleteSession" />
            </ElTooltip>
          </div>
        </div>

        <!-- 消息区域 -->
        <div
          ref="messageContainer"
          class="flex-1 py-5 px-5 overflow-y-auto border-t-d [&::-webkit-scrollbar]:!w-1"
        >
          <!-- 加载更多 -->
          <div v-if="chatStore.hasMore" class="text-center mb-4">
            <ElButton size="small" text :loading="chatStore.messagesLoading" @click="loadMore">加载更多</ElButton>
          </div>

          <template v-for="msg in chatStore.messages" :key="msg.id">
            <!-- 系统消息 -->
            <div v-if="msg.type === 3" class="text-center text-xs text-g-400 my-4">{{ msg.content }}</div>
            <!-- 普通消息 -->
            <div
              v-else
              :class="['flex gap-2.5 items-start w-full mb-6', isMyMessage(msg) ? 'flex-row-reverse' : 'flex-row']"
            >
              <ElAvatar :size="34" :src="avatarUrl(msg.senderAvatar, msg.senderName)" class="flex-shrink-0" />
              <div class="flex flex-col max-w-[65%]" :class="isMyMessage(msg) ? 'items-end' : 'items-start'">
                <div class="flex gap-2 mb-1 text-[11px] items-center" :class="isMyMessage(msg) ? 'flex-row-reverse' : 'flex-row'">
                  <span class="font-medium text-g-700">{{ msg.senderName }}</span>
                  <span v-if="getSenderDept(msg)" class="text-g-400 text-[10px]">{{ getSenderDept(msg) }}</span>
                  <span class="text-g-400">{{ formatMsgTime(msg.createdAt) }}</span>
                </div>
                <!-- 图片消息 -->
                <ElImage
                  v-if="msg.type === 2"
                  :src="msg.content"
                  :preview-src-list="[msg.content]"
                  fit="cover"
                  class="max-w-60 rounded-md cursor-pointer"
                />
                <!-- 文字消息 -->
                <div
                  v-else
                  class="py-2.5 px-3.5 text-sm leading-relaxed rounded-lg whitespace-pre-wrap break-words"
                  :class="isMyMessage(msg) ? '!bg-theme/15' : '!bg-active-color'"
                >{{ msg.content }}</div>
              </div>
            </div>
          </template>
        </div>

        <!-- 输入区域 -->
        <div class="p-4 border-t-d">
          <ElInput
            v-model="messageText"
            type="textarea"
            :rows="3"
            placeholder="输入消息，Enter 发送"
            resize="none"
            @keydown.enter.exact.prevent="handleSend"
          />
          <div class="flex-cb mt-3">
            <div class="flex-c gap-4">
              <ElUpload
                :show-file-list="false"
                accept="image/*"
                :http-request="handleImageUpload"
              >
                <ArtSvgIcon icon="ri:image-line" class="c-p text-g-500 text-lg hover:text-theme tad-200" />
              </ElUpload>
            </div>
            <ElButton type="primary" @click="handleSend" :disabled="!messageText.trim()" v-ripple class="min-w-20">发送</ElButton>
          </div>
        </div>
      </template>
    </div>

    <!-- 联系人选择对话框（左侧部门树 + 右侧联系人列表，对齐抽屉样式） -->
    <ElDialog v-model="showContacts" title="发起聊天" width="640px" @open="onContactDialogOpen">
      <ElInput v-model="contactSearch" placeholder="搜索联系人" prefix-icon="Search" clearable class="mb-3" />
      <div class="flex border border-g-200 rounded-lg overflow-hidden" style="height: 420px">
        <!-- 左侧组织架构树 -->
        <div class="w-[180px] shrink-0 border-r border-g-200 bg-g-50/40 p-2 overflow-y-auto">
          <div class="mb-2 px-1 text-xs font-medium text-g-500">组织架构</div>
          <ElTree
            :data="deptTreeWithAll"
            node-key="id"
            default-expand-all
            highlight-current
            :expand-on-click-node="false"
            :current-node-key="selectedDeptId"
            @node-click="onDeptNodeClick"
          >
            <template #default="{ data }">
              <span class="text-sm truncate">{{ data.name }}</span>
            </template>
          </ElTree>
        </div>

        <!-- 右侧联系人列表 -->
        <div class="flex-1 flex flex-col min-w-0">
          <div class="px-3 py-2 text-xs text-g-500 border-b border-g-100">
            {{ currentDeptLabel }} · 共 {{ deptFilteredContacts.length }} 人
          </div>
          <ElScrollbar class="flex-1">
            <div
              v-for="contact in deptFilteredContacts"
              :key="contact.id"
              class="flex items-center gap-3 px-3 py-2.5 c-p hover:bg-active-color/30 tad-200"
              @click="startChatWith(contact)"
            >
              <div class="relative flex-shrink-0">
                <ElAvatar :size="36" :src="avatarUrl(contact.avatar, contact.realName || contact.username)" />
                <span
                  v-if="contact.isOnline"
                  class="absolute -bottom-0.5 -right-0.5 h-2.5 w-2.5 rounded-full border-2 border-white bg-green-500"
                ></span>
              </div>
              <div class="flex-1 min-w-0">
                <div class="flex items-center gap-2">
                  <span class="text-sm font-medium truncate">{{ contact.realName || contact.username }}</span>
                  <span v-if="contact.postName" class="text-[10px] text-theme/80 bg-theme/8 px-1.5 py-0.5 rounded flex-shrink-0">{{ contact.postName }}</span>
                </div>
                <div class="text-xs text-g-500 mt-0.5 truncate">
                  {{ contact.deptName || '未分配部门' }} · {{ contact.isOnline ? '在线' : '离线' }}
                </div>
              </div>
            </div>
            <div v-if="deptFilteredContacts.length === 0" class="py-10 text-center text-g-400 text-sm">无匹配联系人</div>
          </ElScrollbar>
        </div>
      </div>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { useChatStore } from '@/store/modules/chat'
  import { useUserStore } from '@/store/modules/user'
  import { useAutoLayoutHeight } from '@/hooks/core/useLayoutHeight'
  import { fetchGetDeptList } from '@/api/backend/system/dept'
  import type { ChatSessionItem, ChatContactItem } from '@/api/backend/system/chat'

  defineOptions({ name: 'TemplateChat' })

  const { containerMinHeight } = useAutoLayoutHeight()
  const chatStore = useChatStore()
  const userStore = useUserStore()

  /** 统一头像：有头像显示头像，无头像用 DiceBear API 生成字母头像 */
  const avatarUrl = (avatar: string | undefined, name: string | undefined) => {
    if (avatar) return avatar
    const seed = encodeURIComponent(name || '?')
    return `https://api.dicebear.com/7.x/initials/svg?seed=${seed}&backgroundColor=5a8dee,10b981,ff6b6b,ffab00,03c3ec&fontSize=40`
  }

  const searchQuery = ref('')
  const contactSearch = ref('')
  const messageText = ref('')
  const messageContainer = ref<HTMLElement | null>(null)
  const showContacts = ref(false)

  // 当前用户信息
  const myName = computed(() => userStore.getUserInfo?.realName || userStore.getUserInfo?.username || '我')
  const myAvatar = computed(() => userStore.getUserInfo?.avatar || '')
  const myUserId = computed(() => userStore.getUserInfo?.id || 0)

  // 过滤会话列表
  const filteredSessions = computed(() => {
    const q = searchQuery.value.trim().toLowerCase()
    if (!q) return chatStore.sessions
    return chatStore.sessions.filter(s => s.name.toLowerCase().includes(q))
  })

  // ===== 部门树 + 联系人过滤 =====
  const deptTree = ref<any[]>([])
  const selectedDeptId = ref<number>(0)

  const deptTreeWithAll = computed(() => {
    return [{ id: 0, name: '全部部门', children: deptTree.value || [] }]
  })

  const currentDeptLabel = computed(() => {
    const targetId = selectedDeptId.value
    if (targetId === 0) return '全部部门'
    const walk = (nodes: any[] = []): string => {
      for (const node of nodes) {
        if (Number(node?.id) === targetId) return String(node?.name || '全部部门')
        if (node?.children?.length) {
          const found = walk(node.children)
          if (found) return found
        }
      }
      return ''
    }
    return walk(deptTree.value) || '全部部门'
  })

  // 收集一个部门节点及其所有子部门的 ID
  const collectDeptIds = (nodes: any[] = []): number[] => {
    const ids: number[] = []
    for (const node of nodes) {
      ids.push(Number(node.id))
      if (node.children?.length) ids.push(...collectDeptIds(node.children))
    }
    return ids
  }

  const getSubDeptIds = (deptId: number): Set<number> => {
    if (deptId === 0) return new Set() // 全部
    const walk = (nodes: any[]): any => {
      for (const node of nodes) {
        if (Number(node.id) === deptId) return node
        if (node.children?.length) {
          const found = walk(node.children)
          if (found) return found
        }
      }
      return null
    }
    const target = walk(deptTree.value)
    if (!target) return new Set([deptId])
    return new Set(collectDeptIds([target]))
  }

  const onDeptNodeClick = (node: any) => {
    selectedDeptId.value = Number(node?.id || 0)
  }

  // 按部门+搜索过滤联系人
  const deptFilteredContacts = computed(() => {
    let list = chatStore.contacts
    // 搜索过滤
    const q = contactSearch.value.trim().toLowerCase()
    if (q) {
      list = list.filter(c =>
        (c.realName || '').toLowerCase().includes(q) ||
        (c.username || '').toLowerCase().includes(q) ||
        (c.deptName || '').toLowerCase().includes(q)
      )
    }
    // 部门过滤
    if (selectedDeptId.value !== 0) {
      const deptIds = getSubDeptIds(selectedDeptId.value)
      list = list.filter(c => deptIds.has(c.deptId))
    }
    return list
  })

  const loadDeptTree = async () => {
    try {
      deptTree.value = await fetchGetDeptList({ status: 1 })
    } catch {
      deptTree.value = []
    }
  }

  const onContactDialogOpen = async () => {
    await Promise.all([chatStore.loadContacts(), loadDeptTree()])
  }

  // 判断是否是自己发的消息
  const isMyMessage = (msg: any) => {
    return msg.senderId === 0 || msg.senderId === myUserId.value
  }

  // 判断单聊会话对方是否在线
  const isSessionOnline = (session: ChatSessionItem) => {
    if (session.type !== 1) return false
    const contact = chatStore.contacts.find(c => (c.realName || c.username) === session.name)
    return contact?.isOnline || false
  }

  // 联系人信息映射（用于在消息中显示部门岗位）
  const contactMap = computed(() => {
    const map = new Map<number, ChatContactItem>()
    chatStore.contacts.forEach(c => map.set(c.id, c))
    return map
  })

  // 获取消息发送者的部门信息
  const getSenderDept = (msg: any) => {
    if (isMyMessage(msg)) return ''
    const contact = contactMap.value.get(msg.senderId)
    if (!contact) return ''
    const parts = []
    if (contact.deptName) parts.push(contact.deptName)
    if (contact.postName) parts.push(contact.postName)
    return parts.join(' · ')
  }

  // 获取会话副标题（单聊显示对方部门岗位）
  const getSessionSubtitle = (session: ChatSessionItem) => {
    if (session.type !== 1) return '' // 群聊不显示
    // 从联系人中找到对方
    const contact = chatStore.contacts.find(c => {
      const name = c.realName || c.username
      return name === session.name
    })
    if (!contact) return ''
    const parts = []
    if (contact.deptName) parts.push(contact.deptName)
    if (contact.postName) parts.push(contact.postName)
    return parts.join(' · ')
  }

  // 格式化会话时间
  const formatTime = (ts: number) => {
    if (!ts) return ''
    const date = new Date(ts * 1000)
    const now = new Date()
    const diff = now.getTime() - date.getTime()
    if (diff < 60000) return '刚刚'
    if (diff < 3600000) return Math.floor(diff / 60000) + '分钟前'
    if (date.toDateString() === now.toDateString()) {
      return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
    }
    const yesterday = new Date(now)
    yesterday.setDate(yesterday.getDate() - 1)
    if (date.toDateString() === yesterday.toDateString()) return '昨天'
    return `${date.getMonth() + 1}/${date.getDate()}`
  }

  // 格式化消息时间
  const formatMsgTime = (ts: number) => {
    if (!ts) return ''
    const date = new Date(ts * 1000)
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  }

  // 选择会话
  const selectSession = async (session: ChatSessionItem) => {
    await chatStore.enterSession(session.id)
    scrollToBottom()
  }

  // 会话右键菜单（删除）
  const onSessionContextMenu = (session: ChatSessionItem, _e: MouseEvent) => {
    // 简单处理：暂不做右键菜单，用顶栏删除按钮
  }

  // 删除当前会话
  const handleDeleteSession = async () => {
    if (!chatStore.currentSessionId) return
    try {
      await ElMessageBox.confirm('确定删除该会话吗？', '提示', { type: 'warning' })
      await chatStore.deleteSession(chatStore.currentSessionId)
      ElMessage.success('已删除')
    } catch { /* cancel */ }
  }

  // 发送消息
  const handleSend = async () => {
    const text = messageText.value.trim()
    if (!text || !chatStore.currentSessionId) return
    messageText.value = ''
    await chatStore.sendMessage(chatStore.currentSessionId, 1, text, myName.value, myAvatar.value)
    scrollToBottom()
  }

  // 发送图片
  const handleImageUpload = async (options: any) => {
    if (!chatStore.currentSessionId) return
    const formData = new FormData()
    formData.append('file', options.file)
    try {
      // 使用通用上传接口
      const { adminRequest } = await import('@/utils/http')
      const res: any = await adminRequest.post({ url: '/upload/file', data: formData })
      const url = res?.fullUrl || res?.url || ''
      if (url) {
        await chatStore.sendMessage(chatStore.currentSessionId, 2, url, myName.value, myAvatar.value)
        scrollToBottom()
      }
    } catch {
      ElMessage.error('图片上传失败')
    }
  }

  // 加载更多历史消息
  const loadMore = async () => {
    if (!chatStore.currentSessionId) return
    await chatStore.loadMessages(chatStore.currentSessionId, true)
  }

  // 发起单聊
  const startChatWith = async (contact: ChatContactItem) => {
    showContacts.value = false
    const sessionId = await chatStore.createSession(1, [contact.id])
    if (sessionId) {
      await chatStore.enterSession(sessionId)
      scrollToBottom()
    }
  }

  // 滚动到底部
  const scrollToBottom = () => {
    nextTick(() => {
      setTimeout(() => {
        if (messageContainer.value) {
          messageContainer.value.scrollTop = messageContainer.value.scrollHeight
        }
      }, 50)
    })
  }

  // 监听消息变化自动滚动
  watch(() => chatStore.messages.length, () => {
    scrollToBottom()
  })

  // 初始化
  onMounted(async () => {
    chatStore.initWsListener()
    await Promise.all([
      chatStore.loadSessions(),
      chatStore.loadContacts(), // 预加载联系人（用于显示部门岗位信息）
      chatStore.refreshUnreadTotal(),
    ])
  })
</script>

<style scoped>
:deep(.el-tree) {
  background: transparent;
}
:deep(.el-tree-node__content) {
  height: 32px;
  border-radius: 6px;
}
</style>
