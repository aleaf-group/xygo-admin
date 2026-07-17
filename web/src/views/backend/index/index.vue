<!-- +----------------------------------------------------------------------
  | XYGo Admin [ Vue3 + GoFrame 企业级中后台管理系统 ]
  +----------------------------------------------------------------------
  | Copyright (c) 2026 大连星韵网络科技有限公司 All rights reserved.
  +----------------------------------------------------------------------
  | Licensed ( https://opensource.org/licenses/MIT )
  +----------------------------------------------------------------------
  | Author: 喜羊羊 <751300685@qq.com>
  +---------------------------------------------------------------------- -->
<!-- 布局容器 -->
<template>
  <div class="app-layout">
    <aside id="app-sidebar">
      <ArtSidebarMenu />
    </aside>

    <main id="app-main">
      <div id="app-header">
        <ArtHeaderBar />
      </div>
      <div id="app-content">
        <ArtPageContent />
      </div>
    </main>

    <div id="app-global">
      <ArtGlobalComponent />
    </div>
  </div>
</template>

<script setup lang="ts">
  import { onMounted, onUnmounted } from 'vue'
  import { useWebSocketStore } from '@/store/modules/websocket'

  defineOptions({ name: 'AppLayout' })

  const wsStore = useWebSocketStore()

  // Layout 挂载后自动建立 WebSocket 连接
  onMounted(() => {
    wsStore.connect()
  })

  // Layout 卸载时断开 WebSocket
  onUnmounted(() => {
    wsStore.disconnect()
  })
</script>

<style lang="scss" scoped>
  @use './style';
</style>
