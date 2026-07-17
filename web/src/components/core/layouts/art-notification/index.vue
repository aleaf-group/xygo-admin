<!-- +----------------------------------------------------------------------
  | XYGo Admin [ Vue3 + GoFrame 企业级中后台管理系统 ]
  +----------------------------------------------------------------------
  | Copyright (c) 2026 大连星韵网络科技有限公司 All rights reserved.
  +----------------------------------------------------------------------
  | Licensed ( https://opensource.org/licenses/MIT )
  +----------------------------------------------------------------------
  | Author: 喜羊羊 <751300685@qq.com>
  +---------------------------------------------------------------------- -->
<!-- 通知组件（对接真实数据） -->
<template>
  <div
    class="art-notification-panel art-card-sm !shadow-xl"
    :style="{ transform: show ? 'scaleY(1)' : 'scaleY(0.9)', opacity: show ? 1 : 0 }"
    v-show="visible"
    @click.stop
  >
    <div class="flex-cb px-3.5 mt-3.5">
      <span class="text-base font-medium text-g-800">{{ $t('notice.title') }}</span>
      <span class="text-xs text-g-800 px-1.5 py-1 c-p select-none rounded hover:bg-g-200" @click="handleReadAll">
        {{ $t('notice.btnRead') }}
      </span>
    </div>

    <ul class="box-border flex items-end w-full h-12.5 px-3.5 border-b-d">
      <li
        v-for="(tab, index) in tabs"
        :key="index"
        class="h-12 leading-12 mr-5 overflow-hidden text-[13px] text-g-700 c-p select-none"
        :class="{ 'bar-active': activeTab === index }"
        @click="activeTab = index"
      >
        {{ tab.label }} ({{ tab.count }})
      </li>
    </ul>

    <div class="w-full h-[calc(100%-95px)]">
      <div class="h-[calc(100%-60px)] overflow-y-scroll scrollbar-thin">
        <ul>
          <li
            v-for="item in currentList"
            :key="item.id"
            class="box-border flex-c px-3.5 py-3.5 c-p last:border-b-0 hover:bg-g-200/60"
            :class="{ 'opacity-50': item.isRead }"
            @click="handleClick(item)"
          >
            <div class="size-9 leading-9 text-center rounded-lg flex-cc" :class="getTagStyle(item.tag || item.type)">
              <ArtSvgIcon class="text-lg !bg-transparent" :icon="getTagIcon(item.tag || item.type)" />
            </div>
            <div class="w-[calc(100%-45px)] ml-3.5">
              <h4 class="text-sm font-normal leading-5.5 text-g-900">
                {{ item.title }}
                <span v-if="!item.isRead" class="inline-block w-1.5 h-1.5 rounded-full bg-red-500 ml-1 align-middle"></span>
              </h4>
              <p class="mt-1.5 text-xs text-g-500">{{ formatTime(item.createdAt) }}</p>
            </div>
          </li>
        </ul>

        <!-- 空状态 -->
        <div v-if="currentList.length === 0" class="relative top-25 h-full text-g-500 text-center !bg-transparent">
          <ArtSvgIcon icon="system-uicons:inbox" class="text-5xl" />
          <p class="mt-3.5 text-xs !bg-transparent">暂无{{ tabs[activeTab]?.label }}</p>
        </div>
      </div>

      <div class="relative box-border w-full px-3.5">
        <ElButton class="w-full mt-3" @click="handleViewAll" v-ripple>
          {{ $t('notice.viewAll') }}
        </ElButton>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { useNotificationStore } from '@/store/modules/notification'
  import { ADMIN_BASE_PATH } from '@/router/routesAlias'

  defineOptions({ name: 'ArtNotification' })

  const props = defineProps<{ value: boolean }>()
  const emit = defineEmits<{ 'update:value': [value: boolean] }>()
  const router = useRouter()

  const show = ref(false)
  const visible = ref(false)
  const activeTab = ref(0)

  const store = useNotificationStore()

  const tabs = computed(() => [
    { label: '通知', count: store.notifyUnread, list: store.notifyMessages },
    { label: '公告', count: store.announceUnread, list: store.announceMessages },
    { label: '私信', count: store.letterUnread, list: store.letterMessages },
  ])

  const currentList = computed(() => tabs.value[activeTab.value]?.list || [])

  const getTagIcon = (tag: any): string => {
    const map: Record<string, string> = {
      info: 'ri:information-line', success: 'ri:checkbox-circle-line',
      warning: 'ri:error-warning-line', danger: 'ri:close-circle-line',
      '1': 'ri:notification-3-line', '2': 'ri:volume-down-line', '3': 'ri:mail-line',
    }
    return map[String(tag)] || 'ri:notification-3-line'
  }

  const getTagStyle = (tag: any): string => {
    const map: Record<string, string> = {
      info: 'bg-info/12 text-info', success: 'bg-success/12 text-success',
      warning: 'bg-warning/12 text-warning', danger: 'bg-danger/12 text-danger',
      '1': 'bg-theme/12 text-theme', '2': 'bg-success/12 text-success', '3': 'bg-warning/12 text-warning',
    }
    return map[String(tag)] || 'bg-theme/12 text-theme'
  }

  const formatTime = (ts: number): string => {
    if (!ts) return ''
    return new Date(ts * 1000).toLocaleString('zh-CN')
  }

  const handleClick = (item: any) => {
    if (!item.isRead) {
      store.markRead(item.id)
    }
  }

  const handleReadAll = () => {
    const typeMap = [1, 2, 3]
    store.markAllRead(typeMap[activeTab.value])
  }

  const handleViewAll = () => {
    emit('update:value', false)
    router.push(`${ADMIN_BASE_PATH}/system/notice`)
  }

  // 动画控制
  watch(() => props.value, (val) => {
    if (val) {
      visible.value = true
      store.pullMessages() // 每次打开刷新
      setTimeout(() => { show.value = true }, 5)
    } else {
      show.value = false
      setTimeout(() => { visible.value = false }, 350)
    }
  })
</script>

<style scoped>
  @reference '@styles/core/tailwind.css';

  .art-notification-panel {
    @apply absolute top-14.5 right-5 w-90 h-125 overflow-hidden
    transition-all duration-300 origin-top will-change-[top,left]
    max-[640px]:top-[65px] max-[640px]:right-0 max-[640px]:w-full max-[640px]:h-[80vh];
  }

  .bar-active {
    color: var(--theme-color) !important;
    border-bottom: 2px solid var(--theme-color);
  }

  .scrollbar-thin::-webkit-scrollbar { width: 5px !important; }
  .dark .scrollbar-thin::-webkit-scrollbar-track { background-color: var(--default-box-color); }
  .dark .scrollbar-thin::-webkit-scrollbar-thumb { background-color: #222 !important; }
</style>
