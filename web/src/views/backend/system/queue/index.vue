<!-- 消息队列管理页面 -->
<template>
  <div class="queue-page art-full-height">
    <ElCard shadow="never">
      <!-- 工具栏 -->
      <div class="mb-4 flex-cb">
        <div class="flex-c gap-3">
          <span class="text-base font-semibold">消息队列</span>
          <ElTag type="info" size="small">驱动：{{ stats.driver || '-' }}</ElTag>
          <ElTag type="success" size="small">Topic：{{ stats.topics?.length || 0 }} 个</ElTag>
        </div>
        <div class="flex-c gap-2">
          <ElButton @click="showPushDialog = true">测试投递</ElButton>
          <ElButton :icon="Refresh" circle @click="loadStats" />
        </div>
      </div>

      <!-- 队列统计表格 -->
      <ElTable :data="stats.topics || []" v-loading="loading" border stripe>
        <ElTableColumn prop="topic" label="Topic" min-width="140">
          <template #default="{ row }">
            <code class="text-sm text-blue-600">{{ row.topic }}</code>
            <div v-if="row.title && row.title !== row.topic" class="text-xs text-g-400">{{ row.title }}</div>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="workers" label="Worker" width="80" align="center" />
        <ElTableColumn prop="maxRetry" label="重试" width="70" align="center" />
        <ElTableColumn prop="retryDelaySec" label="重试间隔" width="90" align="center">
          <template #default="{ row }">
            <span class="text-sm">{{ row.retryDelaySec ? row.retryDelaySec + 's' : '-' }}</span>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="pending" label="待消费" width="90" align="center">
          <template #default="{ row }">
            <ElTag :type="row.pending > 0 ? 'warning' : 'success'" size="small">{{ row.pending }}</ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="deadSize" label="死信" width="80" align="center">
          <template #default="{ row }">
            <ElTag :type="row.deadSize > 0 ? 'danger' : 'info'" size="small">{{ row.deadSize }}</ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="rate" label="速率(条/分)" width="110" align="center">
          <template #default="{ row }">
            <span class="text-sm">{{ row.rate ? row.rate.toFixed(0) : '0' }}</span>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="avgTakeMs" label="平均耗时" width="100" align="center">
          <template #default="{ row }">
            <span class="text-sm">{{ row.rate > 0 ? row.avgTakeMs.toFixed(1) + 'ms' : '-' }}</span>
          </template>
        </ElTableColumn>
        <ElTableColumn label="状态" width="80" align="center">
          <template #default="{ row }">
            <span v-if="row.status === 1" class="flex-c justify-center gap-1">
              <span class="inline-block h-2 w-2 rounded-full bg-success/100"></span>
              <span class="text-xs text-g-600">运行</span>
            </span>
            <ElTag v-else type="info" size="small">停用</ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn label="操作" width="160" align="right" fixed="right">
          <template #default="{ row }">
            <ElButton v-if="hasAuth('edit')" size="small" text type="primary" @click="openConfig(row)">配置</ElButton>
            <ElButton size="small" text type="primary" @click="pushToTopic(row.topic)">投递测试</ElButton>
          </template>
        </ElTableColumn>
      </ElTable>

      <!-- 说明 -->
      <div class="mt-4 rounded-lg bg-g-50 p-4">
        <div class="mb-2 text-sm font-medium text-g-700">使用说明</div>
        <ul class="space-y-1 text-xs text-g-500">
          <li>消息队列支持 <strong>Redis</strong> 和 <strong>Disk</strong> 双驱动，可在配置文件中切换</li>
          <li><strong>Worker 数</strong>：同一进程内并行消费的 goroutine 数量（类似 Supervisor numprocs）</li>
          <li>消费失败返回 <code>NewRetryError</code> 时按上方配置重试，超限进入<strong>死信队列</strong>（topic:dead）</li>
          <li>新增消费者：在 <code>server/internal/queues/</code> 实现 Consumer 接口并 Register（示例见 <code>demo_task.go</code>），重启后自动出现在列表</li>
          <li>生产者投递：<code>queue.Push("topic_name", data)</code></li>
        </ul>
      </div>
    </ElCard>

    <QueueConfigDialog v-model:visible="showConfigDialog" :data="currentTopic" @success="loadStats" />

    <!-- 测试投递弹窗 -->
    <ElDialog v-model="showPushDialog" title="测试投递消息" width="500px" align-center>
      <ElForm label-width="80px">
        <ElFormItem label="Topic">
          <ElSelect v-model="pushForm.topic" placeholder="选择 Topic" class="!w-full">
            <ElOption v-for="t in registeredTopics" :key="t" :label="t" :value="t" />
          </ElSelect>
        </ElFormItem>
        <ElFormItem label="延迟（秒）">
          <ElInputNumber v-model="pushForm.delaySec" :min="0" :max="86400" :step="10" class="!w-full" />
          <div class="mt-1 text-xs text-g-400">0 = 即时投递，填 300 = 5分钟后消费</div>
        </ElFormItem>
        <ElFormItem label="消息内容">
          <ElInput v-model="pushForm.body" type="textarea" :rows="5" placeholder='JSON 格式，如: {"key":"value"}' />
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="showPushDialog = false">取消</ElButton>
        <ElButton type="primary" :loading="pushing" @click="handlePush">投递</ElButton>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { Refresh } from '@element-plus/icons-vue'
  import { ElMessage } from 'element-plus'
  import { useAuth } from '@/hooks/core/useAuth'
  import { fetchQueueStats, fetchQueueTopics, fetchQueuePushTest, type QueueTopicStats } from '@/api/backend/system/queue'
  import QueueConfigDialog from './modules/queue-config-dialog.vue'

  defineOptions({ name: 'QueueManage' })

  const { hasAuth } = useAuth()
  const loading = ref(false)
  const stats = ref<{ driver: string; topics: QueueTopicStats[] }>({ driver: '', topics: [] })
  const registeredTopics = ref<string[]>([])
  const showPushDialog = ref(false)
  const showConfigDialog = ref(false)
  const currentTopic = ref<QueueTopicStats | null>(null)
  const pushing = ref(false)
  const pushForm = reactive({ topic: '', body: '', delaySec: 0 })

  const loadStats = async () => {
    loading.value = true
    try {
      const res = await fetchQueueStats() as any
      stats.value = { driver: res?.driver || '', topics: res?.topics || [] }
    } catch { /* */ } finally {
      loading.value = false
    }
  }

  const loadTopics = async () => {
    try {
      const res = await fetchQueueTopics() as any
      registeredTopics.value = res?.list || []
    } catch { /* */ }
  }

  const openConfig = (row: QueueTopicStats) => {
    currentTopic.value = row
    showConfigDialog.value = true
  }

  const topicDefaultBody: Record<string, string> = {
    login_log: JSON.stringify({ username: 'test', ip: '127.0.0.1', location: '测试', user_agent: 'QueueTest', browser: 'Test', os: 'Test', status: 1, message: '队列测试登录', created_at: Math.floor(Date.now() / 1000) }, null, 2),
    operation_log: JSON.stringify({ user_id: 1, username: 'admin', module: 'test', title: '队列测试', method: 'POST', url: '/test', ip: '127.0.0.1', location: '测试', user_agent: 'QueueTest', request_body: '{}', response_body: '{}', error_message: '', status: 1, elapsed: 10, created_at: Math.floor(Date.now() / 1000) }, null, 2),
    notice_push: JSON.stringify({ userIds: [1], event: 'notice', payload: { title: '队列测试通知', content: '这是一条来自队列投递测试的通知' } }, null, 2),
    demo_task: JSON.stringify({ message: 'Hello from demo_task queue', fail: false }, null, 2),
  }

  const pushToTopic = (topic: string) => {
    pushForm.topic = topic
    pushForm.body = topicDefaultBody[topic] || '{}'
    pushForm.delaySec = 0
    showPushDialog.value = true
  }

  const handlePush = async () => {
    if (!pushForm.topic) { ElMessage.warning('请选择 Topic'); return }
    if (!pushForm.body) { ElMessage.warning('请输入消息内容'); return }
    pushing.value = true
    try {
      await fetchQueuePushTest({ topic: pushForm.topic, body: pushForm.body, delaySec: pushForm.delaySec })
      ElMessage.success('投递成功')
      showPushDialog.value = false
      loadStats()
    } catch { /* */ } finally {
      pushing.value = false
    }
  }

  onMounted(() => {
    loadStats()
    loadTopics()
  })
</script>
