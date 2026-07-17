<!-- +----------------------------------------------------------------------
  | XYGo Admin [ Vue3 + GoFrame 企业级中后台管理系统 ]
  +----------------------------------------------------------------------
  | Copyright (c) 2026 大连星韵网络科技有限公司 All rights reserved.
  +----------------------------------------------------------------------
  | Licensed ( https://opensource.org/licenses/MIT )
  +----------------------------------------------------------------------
  | Author: 喜羊羊 <751300685@qq.com>
  +---------------------------------------------------------------------- -->
<!-- 管理员聊天窗口 - 支持单聊/群聊/文字/图片 -->
<template>
  <div>
    <ElDrawer v-model="isDrawerVisible" :size="isMobile ? '100%' : '480px'" :with-header="false" @open="onDrawerOpen" @close="onDrawerClosed">
      <!-- ===================== 视图1：会话列表 ===================== -->
      <div v-if="view === 'sessions'" class="flex h-full flex-col">
        <!-- 顶部 -->
        <div class="mb-4 flex-cb">
          <span class="text-base font-semibold">消息</span>
          <div class="flex-c gap-2">
            <ElTooltip content="新建聊天" placement="bottom">
              <button
                type="button"
                class="c-p rounded p-1 text-lg text-g-600 transition-colors hover:bg-g-100 hover:text-theme"
                @click.stop="openContactPicker"
              >
                <ArtSvgIcon icon="ri:edit-box-line" />
              </button>
            </ElTooltip>
            <ElIcon class="c-p" :size="20" @click="closeChat">
              <Close />
            </ElIcon>
          </div>
        </div>

        <!-- 搜索 -->
        <ElInput v-model="searchText" placeholder="搜索会话" :prefix-icon="Search" clearable class="mb-3" size="default" />

        <!-- 会话列表 -->
        <div class="flex-1 overflow-y-auto [&::-webkit-scrollbar]:!w-1">
          <div v-if="chatStore.loading" class="flex-cc py-10">
            <ElIcon class="is-loading text-2xl text-g-400"><Loading /></ElIcon>
          </div>
          <template v-else-if="filteredSessions.length > 0">
            <div
              v-for="session in filteredSessions"
              :key="session.id"
              class="group flex cursor-pointer items-center gap-3 rounded-lg px-3 py-3 transition-colors hover:bg-g-100/80"
              @click="openChat(session.id)"
              @contextmenu.prevent="onSessionContextMenu($event, session)"
            >
              <!-- 头像 -->
              <div class="relative shrink-0">
                <ElAvatar :size="42" :src="session.avatar || defaultAvatar(session.name)" />
                <span
                  v-if="session.type === 2"
                  class="absolute -bottom-0.5 -right-0.5 flex h-4 w-4 items-center justify-center rounded-full bg-theme text-[10px] text-white"
                >
                  <ArtSvgIcon icon="ri:group-line" class="text-[10px]" />
                </span>
              </div>
              <!-- 内容 -->
              <div class="min-w-0 flex-1">
                <div class="flex-cb">
                  <span class="truncate text-sm font-medium">{{ getSessionDisplayName(session) }}</span>
                  <span class="shrink-0 text-xs text-g-500">{{ formatTime(session.lastMessageTime) }}</span>
                </div>
                <div class="mt-1 flex-cb">
                  <span class="truncate text-xs text-g-500">
                    {{ session.type === 1 ? getSessionOrgInfo(session) : (session.lastMessage || '暂无消息') }}
                  </span>
                  <span
                    v-if="session.unreadCount > 0"
                    class="ml-1 inline-flex h-4.5 min-w-4.5 shrink-0 items-center justify-center rounded-full bg-danger/100 px-1 text-[10px] text-white"
                  >
                    {{ session.unreadCount > 99 ? '99+' : session.unreadCount }}
                  </span>
                </div>
                <div v-if="session.type === 1" class="mt-0.5 truncate text-xs text-g-400">
                  {{ session.lastMessage || '暂无消息' }}
                </div>
              </div>
            </div>
          </template>
          <div v-else class="flex-cc flex-col gap-2 py-16 text-g-400">
            <ArtSvgIcon icon="ri:chat-3-line" class="text-4xl" />
            <span class="text-sm">暂无会话</span>
            <ElButton type="primary" size="small" plain @click="openContactPicker" v-ripple>开始聊天</ElButton>
          </div>
        </div>
      </div>

      <!-- ===================== 视图2：聊天对话 ===================== -->
      <div v-else-if="view === 'chat'" class="flex h-full flex-col">
        <!-- 顶部 -->
        <div class="mb-3 flex-cb">
          <div class="flex-c gap-2">
            <ElIcon class="c-p" :size="20" @click="backToSessions">
              <ArrowLeft />
            </ElIcon>
            <div>
              <span class="text-base font-medium">{{ chatStore.currentSession ? getSessionDisplayName(chatStore.currentSession) : '聊天' }}</span>
              <div class="mt-0.5 text-xs text-g-500">
                <template v-if="chatStore.currentSession?.type === 2">
                  {{ chatStore.currentSession?.memberCount || 0 }} 人
                </template>
                <template v-else>
                  <span class="flex-c gap-1">
                    <span class="inline-block h-1.5 w-1.5 rounded-full bg-success/100"></span>
                    在线
                  </span>
                </template>
              </div>
            </div>
          </div>
          <ElIcon class="c-p" :size="20" @click="closeChat">
            <Close />
          </ElIcon>
        </div>

        <!-- 消息区域 -->
        <div
          ref="messageContainer"
          class="flex-1 overflow-y-auto border-t-d px-3 py-4 [&::-webkit-scrollbar]:!w-1"
          @scroll="onMessageScroll"
        >
          <!-- 加载更多 -->
          <div v-if="chatStore.hasMore" class="mb-4 flex-cc">
            <ElButton size="small" text :loading="chatStore.messagesLoading" @click="loadMoreMessages">
              加载更早的消息
            </ElButton>
          </div>

          <template v-for="(msg, index) in chatStore.messages" :key="msg.id">
            <!-- 时间分割线 -->
            <div v-if="shouldShowTime(index)" class="my-4 flex-cc">
              <span class="rounded-full bg-g-200/60 px-3 py-0.5 text-xs text-g-500">
                {{ formatFullTime(msg.createdAt) }}
              </span>
            </div>

            <!-- 系统消息 -->
            <div v-if="msg.type === 3" class="my-3 flex-cc">
              <span class="rounded-full bg-g-200/60 px-3 py-0.5 text-xs text-g-500">{{ msg.content }}</span>
            </div>

            <!-- 普通消息 -->
            <div
              v-else
              :class="['mb-4 flex w-full items-start gap-2', isMine(msg) ? 'flex-row-reverse' : 'flex-row']"
            >
              <ElAvatar :size="32" :src="getAvatar(msg)" class="shrink-0" />
              <div :class="['flex max-w-[70%] flex-col', isMine(msg) ? 'items-end' : 'items-start']">
                <div :class="['mb-1 flex gap-2 text-xs', isMine(msg) ? 'flex-row-reverse' : 'flex-row']">
                  <span class="font-medium text-g-700">{{ msg.senderName }}</span>
                </div>
                <!-- 文字消息 -->
                <div
                  v-if="msg.type === 1"
                  :class="[
                    'whitespace-pre-wrap break-all rounded-lg px-3.5 py-2.5 text-sm leading-relaxed',
                    isMine(msg) ? 'bg-theme/15 text-g-900' : 'bg-g-100 text-g-900'
                  ]"
                >{{ msg.content }}</div>
                <!-- 图片消息 -->
                <div v-else-if="msg.type === 2" class="max-w-60">
                  <ElImage
                    :src="msg.content"
                    fit="cover"
                    class="!max-h-60 cursor-pointer rounded-lg"
                    :preview-src-list="[msg.content]"
                    preview-teleported
                  />
                </div>
              </div>
            </div>
          </template>
        </div>

        <!-- 输入区域 -->
        <div class="border-t-d px-3 pt-3">
          <!-- 工具栏 -->
          <div class="mb-2 flex-c gap-1">
            <ElTooltip content="发送图片" placement="top">
              <span class="c-p rounded p-1.5 text-g-500 transition-colors hover:bg-g-100 hover:text-theme" @click="triggerImageUpload">
                <ArtSvgIcon icon="ri:image-line" class="text-base" />
              </span>
            </ElTooltip>
            <input
              ref="imageInput"
              type="file"
              accept="image/*"
              class="hidden"
              @change="handleImageUpload"
            />
          </div>
          <!-- 输入框 -->
          <div class="flex items-end gap-2">
            <ElInput
              v-model="messageText"
              type="textarea"
              :rows="2"
              placeholder="输入消息，Enter 发送"
              resize="none"
              class="flex-1"
              @keydown.enter.exact.prevent="handleSend"
            />
            <ElButton type="primary" :loading="sending" @click="handleSend" v-ripple class="mb-0.5 min-w-16">
              发送
            </ElButton>
          </div>
        </div>
      </div>

      <!-- ===================== 视图3：联系人选择器 ===================== -->
      <div v-else-if="view === 'contacts'" class="flex h-full flex-col">
        <!-- 顶部 -->
        <div class="mb-4 flex-cb">
          <div class="flex-c gap-2">
            <ElIcon class="c-p" :size="20" @click="view = 'sessions'">
              <ArrowLeft />
            </ElIcon>
            <span class="text-base font-semibold">选择联系人</span>
          </div>
          <ElIcon class="c-p" :size="20" @click="closeChat">
            <Close />
          </ElIcon>
        </div>

        <!-- 搜索 -->
        <ElInput v-model="contactSearch" placeholder="搜索管理员" :prefix-icon="Search" clearable class="mb-3" />
        <div class="mb-3 flex-1 overflow-hidden rounded-lg border border-g-200">
          <!-- 左：组织结构 -->
          <div class="h-full flex">
            <div class="h-full w-[170px] shrink-0 border-r border-g-200 bg-g-50/40 p-2">
              <div class="mb-2 px-1 text-xs font-medium text-g-500">组织结构</div>
              <ElScrollbar height="100%">
                <ElTree
                  :data="deptTreeWithAll"
                  node-key="id"
                  default-expand-all
                  highlight-current
                  :expand-on-click-node="false"
                  :current-node-key="Number(selectedDeptId || 0)"
                  :props="{ label: 'name', children: 'children' }"
                  @node-click="onDeptNodeClick"
                />
              </ElScrollbar>
            </div>

            <!-- 右：人员 -->
            <div class="min-w-0 flex-1 p-2">
              <div class="mb-2 text-xs text-g-500">
                {{ currentDeptLabel }} · 共 {{ filteredContacts.length }} 人
              </div>

              <!-- 已选中标签 -->
              <div v-if="selectedContacts.length > 0" class="mb-2 flex flex-wrap gap-1.5">
                <ElTag
                  v-for="c in selectedContacts"
                  :key="c.id"
                  closable
                  size="small"
                  @close="toggleContact(c)"
                >
                  {{ c.username }}
                </ElTag>
              </div>

              <!-- 联系人列表 -->
              <div class="h-[calc(100%-40px)] overflow-y-auto [&::-webkit-scrollbar]:!w-1">
                <div
                  v-for="contact in filteredContacts"
                  :key="contact.id"
                  class="flex cursor-pointer items-center gap-3 rounded-lg px-3 py-2.5 transition-colors hover:bg-g-100/80"
                  @click="toggleContact(contact)"
                >
                  <ElCheckbox
                    :model-value="isContactSelected(contact.id)"
                    @click.stop
                    @change="toggleContact(contact)"
                  />
                  <div class="relative">
                    <ElAvatar :size="36" :src="contact.avatar || defaultAvatar(contact.username)" />
                    <span
                      v-if="contact.isOnline"
                      class="absolute -bottom-0.5 -right-0.5 h-2.5 w-2.5 rounded-full border-2 border-white bg-success/100"
                    ></span>
                  </div>
                  <div class="min-w-0 flex-1">
                    <div class="truncate text-sm font-medium">{{ contact.username }}</div>
                    <div class="truncate text-xs text-g-500">
                      {{ contact.realName ? `${contact.realName} · ` : '' }}{{ contact.deptName || '未分配部门' }} · {{ contact.isOnline ? '在线' : '离线' }}
                    </div>
                  </div>
                </div>
                <div v-if="filteredContacts.length === 0" class="flex-cc py-10 text-sm text-g-400">
                  无匹配联系人
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 群名输入（多选时显示） -->
        <div v-if="selectedContacts.length > 1" class="border-t-d px-1 pt-3">
          <ElInput v-model="groupName" placeholder="输入群名称（可选）" class="mb-3" />
        </div>

        <!-- 确认按钮 -->
        <div class="border-t-d px-1 pt-3">
          <ElButton
            type="primary"
            class="w-full"
            :disabled="selectedContacts.length === 0"
            :loading="creating"
            @click="confirmCreateSession"
            v-ripple
          >
            {{ selectedContacts.length > 1 ? `创建群聊 (${selectedContacts.length}人)` : selectedContacts.length === 1 ? '开始聊天' : '请选择联系人' }}
          </ElButton>
        </div>
      </div>
    </ElDrawer>
  </div>
</template>

<script setup lang="ts">
  import { Close, ArrowLeft, Search, Loading } from '@element-plus/icons-vue'
  import { mittBus } from '@/utils/sys'
  import { useChatStore } from '@/store/modules/chat'
  import { useUserStore } from '@/store/modules/user'
  import { uploadImageApi } from '@/api/backend/common/upload'
  import { fetchGetDeptList } from '@/api/backend/system/dept'
  import type { ChatContactItem, ChatSessionItem } from '@/api/backend/system/chat'
  import { ElMessage, ElMessageBox } from 'element-plus'

  defineOptions({ name: 'ArtChatWindow' })

  // ==================== 基础状态 ====================
  const MOBILE_BREAKPOINT = 640
  const { width } = useWindowSize()
  const isMobile = computed(() => width.value < MOBILE_BREAKPOINT)

  const chatStore = useChatStore()
  const userStore = useUserStore()
  const currentUserId = computed(() => userStore.getUserInfo?.id || 0)
  const currentUserName = computed(() => (userStore.getUserInfo as any)?.username || '')
  const currentUserAvatar = computed(() => (userStore.getUserInfo as any)?.avatar || '')

  /** 生成默认头像 URL（与后端风格一致） */
  const defaultAvatar = (name: string) => {
    return `https://ui-avatars.com/api/?background=random&name=${encodeURIComponent(name || '?')}`
  }

  const isDrawerVisible = ref(false)
  const view = ref<'sessions' | 'chat' | 'contacts'>('sessions')

  /** 判断消息是否是自己发的（senderId=0 表示本地刚发送的） */
  const isMine = (msg: any) => msg.senderId === currentUserId.value || msg.senderId === 0

  /** 获取消息头像（无头像时用统一的 API 生成） */
  const getAvatar = (msg: any) => msg.senderAvatar || defaultAvatar(msg.senderName || '?')

  // ==================== 会话列表 ====================
  const searchText = ref('')
  const filteredSessions = computed(() => {
    if (!searchText.value) return chatStore.sessions
    const keyword = searchText.value.toLowerCase()
    return chatStore.sessions.filter(s =>
      s.name?.toLowerCase().includes(keyword) ||
      s.lastMessage?.toLowerCase().includes(keyword)
    )
  })

  const contactMapByUsername = computed(() => {
    const map = new Map<string, ChatContactItem>()
    for (const c of chatStore.contacts) {
      if (c.username) map.set(c.username, c)
      if (c.realName) map.set(c.realName, c)
    }
    return map
  })

  const getSessionDisplayName = (session: ChatSessionItem) => {
    const base = session.name || '未命名'
    if (session.type !== 1) return base
    const contact = contactMapByUsername.value.get(base)
    if (!contact?.realName || contact.realName === base) return base
    return `${base}（${contact.realName}）`
  }

  const getSessionOrgInfo = (session: ChatSessionItem) => {
    if (session.type !== 1) return ''
    const base = session.name || ''
    const contact = contactMapByUsername.value.get(base)
    const dept = contact?.deptName || '未分配部门'
    const realName = contact?.realName && contact.realName !== base ? contact.realName : ''
    return realName ? `${dept} · ${realName}` : dept
  }

  // ==================== 聊天对话 ====================
  const messageContainer = ref<HTMLElement | null>(null)
  const messageText = ref('')
  const sending = ref(false)
  const imageInput = ref<HTMLInputElement | null>(null)

  const scrollToBottom = (smooth = true) => {
    nextTick(() => {
      setTimeout(() => {
        if (messageContainer.value) {
          messageContainer.value.scrollTo({
            top: messageContainer.value.scrollHeight,
            behavior: smooth ? 'smooth' : 'instant'
          })
        }
      }, 50)
    })
  }

  const handleSend = async () => {
    const text = messageText.value.trim()
    if (!text || sending.value) return
    try {
      sending.value = true
      await chatStore.sendMessage(
        chatStore.currentSessionId, 1, text,
        currentUserName.value, currentUserAvatar.value || defaultAvatar(currentUserName.value)
      )
      messageText.value = ''
      scrollToBottom()
    } catch {
      ElMessage.error('发送失败')
    } finally {
      sending.value = false
    }
  }

  const triggerImageUpload = () => {
    imageInput.value?.click()
  }

  const handleImageUpload = async (e: Event) => {
    const target = e.target as HTMLInputElement
    const file = target.files?.[0]
    if (!file) return
    target.value = '' // 清空，允许重复选择同一文件

    try {
      sending.value = true
      const res = await uploadImageApi(file) as any
      const url = res?.url
      if (url) {
        await chatStore.sendMessage(
          chatStore.currentSessionId, 2, url,
          currentUserName.value, currentUserAvatar.value || defaultAvatar(currentUserName.value)
        )
        scrollToBottom()
      }
    } catch {
      ElMessage.error('图片上传失败')
    } finally {
      sending.value = false
    }
  }

  const loadMoreMessages = () => {
    chatStore.loadMessages(chatStore.currentSessionId, true)
  }

  const onMessageScroll = () => {
    // 预留：滚动到顶部自动加载更多
  }

  const shouldShowTime = (index: number): boolean => {
    if (index === 0) return true
    const msgs = chatStore.messages
    const current = msgs[index].createdAt
    const prev = msgs[index - 1].createdAt
    return current - prev > 300 // 5分钟间隔显示时间
  }

  // 监听消息变化自动滚动
  watch(() => chatStore.messages.length, (newLen, oldLen) => {
    if (newLen > oldLen) {
      scrollToBottom()
    }
  })

  // ==================== 联系人选择器 ====================
  const contactSearch = ref('')
  const selectedContacts = ref<ChatContactItem[]>([])
  const groupName = ref('')
  const selectedDeptId = ref<number | string | undefined>(0)
  const creating = ref(false)
  const deptTree = ref<any[]>([])

  const deptTreeWithAll = computed(() => {
    return [{ id: 0, name: '全部部门', children: deptTree.value || [] }]
  })

  const currentDeptLabel = computed(() => {
    const targetId = Number(selectedDeptId.value || 0)
    if (targetId === 0) return '全部部门'
    const walk = (nodes: any[] = []): string => {
      for (const node of nodes) {
        if (Number(node?.id) === targetId) return String(node?.name || '全部部门')
        if (Array.isArray(node?.children) && node.children.length > 0) {
          const childName = walk(node.children)
          if (childName) return childName
        }
      }
      return ''
    }
    return walk(deptTree.value) || '全部部门'
  })

  const collectDeptIds = (nodes: any[] = []): number[] => {
    const ids: number[] = []
    for (const node of nodes) {
      if (Number(node?.id) > 0) ids.push(Number(node.id))
      if (Array.isArray(node?.children) && node.children.length > 0) {
        ids.push(...collectDeptIds(node.children))
      }
    }
    return ids
  }

  const getDeptScope = (deptId: number): Set<number> => {
    const walk = (nodes: any[]): any | null => {
      for (const node of nodes) {
        if (Number(node?.id) === Number(deptId)) return node
        if (Array.isArray(node?.children) && node.children.length > 0) {
          const found = walk(node.children)
          if (found) return found
        }
      }
      return null
    }
    const target = walk(deptTree.value)
    if (!target) return new Set([Number(deptId)])
    return new Set(collectDeptIds([target]))
  }

  const loadDeptTree = async () => {
    try {
      deptTree.value = await fetchGetDeptList({ status: 1 })
    } catch {
      deptTree.value = []
    }
  }

  const onDeptNodeClick = (node: any) => {
    selectedDeptId.value = Number(node?.id || 0)
  }

  const filteredContacts = computed(() => {
    const keyword = contactSearch.value.toLowerCase()
    const deptId = selectedDeptId.value === undefined || selectedDeptId.value === null || selectedDeptId.value === ''
      ? 0
      : Number(selectedDeptId.value)
    const deptScope = deptId > 0 ? getDeptScope(deptId) : null
    return chatStore.contacts.filter(c => {
      const byKeyword = !keyword || c.username?.toLowerCase().includes(keyword) || c.realName?.toLowerCase().includes(keyword)
      const byDept = !deptScope || deptScope.has(Number(c.deptId || 0))
      return byKeyword && byDept
    })
  })

  const isContactSelected = (id: number) => {
    return selectedContacts.value.some(c => c.id === id)
  }

  const toggleContact = (contact: ChatContactItem) => {
    const idx = selectedContacts.value.findIndex(c => c.id === contact.id)
    if (idx >= 0) {
      selectedContacts.value.splice(idx, 1)
    } else {
      selectedContacts.value.push(contact)
    }
  }

  const confirmCreateSession = async () => {
    if (selectedContacts.value.length === 0) return
    try {
      creating.value = true
      const type = selectedContacts.value.length === 1 ? 1 : 2
      const userIds = selectedContacts.value.map(c => c.id)
      const name = type === 2 ? (groupName.value || undefined) : undefined
      const sessionId = await chatStore.createSession(type, userIds, name)
      if (sessionId) {
        await chatStore.enterSession(sessionId)
        view.value = 'chat'
        scrollToBottom(false)
      }
    } catch {
      ElMessage.error('创建失败')
    } finally {
      creating.value = false
    }
  }

  // ==================== 视图切换 ====================
  const openContactPicker = async () => {
    selectedContacts.value = []
    contactSearch.value = ''
    groupName.value = ''
    selectedDeptId.value = 0
    view.value = 'contacts'
    await Promise.all([chatStore.loadContacts(), loadDeptTree()])
  }

  const openChat = async (sessionId: number) => {
    await chatStore.enterSession(sessionId)
    view.value = 'chat'
    scrollToBottom(false)
  }

  const backToSessions = () => {
    chatStore.leaveSession()
    view.value = 'sessions'
  }

  const onSessionContextMenu = (_e: MouseEvent, session: ChatSessionItem) => {
    ElMessageBox.confirm('确定要删除该会话吗？', '提示', {
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      type: 'warning',
    }).then(() => {
      chatStore.deleteSession(session.id)
    }).catch(() => {})
  }

  // ==================== 抽屉控制 ====================
  const onDrawerOpen = async () => {
    view.value = 'sessions'
    searchText.value = ''
    await Promise.all([chatStore.loadSessions(), chatStore.loadContacts()])
  }

  const onDrawerClosed = () => {
    // 包括遮罩点击、ESC、右上角关闭在内，统一重置会话态
    chatStore.leaveSession()
    view.value = 'sessions'
  }

  const openDrawer = () => {
    isDrawerVisible.value = true
  }

  const closeChat = () => {
    isDrawerVisible.value = false
    // 关闭抽屉后退出当前会话，后续新消息才能正确计入未读徽章
    chatStore.leaveSession()
    view.value = 'sessions'
  }

  // ==================== 工具函数 ====================
  const formatTime = (timestamp: number): string => {
    if (!timestamp) return ''
    const date = new Date(timestamp * 1000)
    const now = new Date()
    const diffDays = Math.floor((now.getTime() - date.getTime()) / 86400000)

    if (diffDays === 0) {
      return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
    } else if (diffDays === 1) {
      return '昨天'
    } else if (diffDays < 7) {
      const days = ['周日', '周一', '周二', '周三', '周四', '周五', '周六']
      return days[date.getDay()]
    } else {
      return `${date.getMonth() + 1}/${date.getDate()}`
    }
  }

  const formatFullTime = (timestamp: number): string => {
    if (!timestamp) return ''
    const date = new Date(timestamp * 1000)
    const now = new Date()
    const diffDays = Math.floor((now.getTime() - date.getTime()) / 86400000)

    const time = date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
    if (diffDays === 0) {
      return time
    } else if (diffDays === 1) {
      return `昨天 ${time}`
    } else {
      return `${date.getMonth() + 1}/${date.getDate()} ${time}`
    }
  }

  // ==================== 生命周期 ====================
  onMounted(() => {
    mittBus.on('openChat', openDrawer)
  })

  onUnmounted(() => {
    mittBus.off('openChat', openDrawer)
  })
</script>
