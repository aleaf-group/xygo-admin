<template>
  <div class="oss-sync-mini" v-loading="previewLoading">
    <div class="oss-sync-mini__head">
      <span class="oss-sync-mini__title">历史资源迁移</span>
      <ElButton link type="primary" size="small" :loading="previewLoading" @click="loadPreview">
        刷新
      </ElButton>
    </div>

    <template v-if="preview">
      <ElAlert
        v-if="!preview.ready"
        type="info"
        :closable="false"
        show-icon
        class="oss-sync-mini__alert"
        :title="`请先完善并保存${label}配置`"
        description="保存上方配置后点击刷新，即可将本地附件同步到此云存储。"
      />
      <p v-else class="oss-sync-mini__stats">
        待迁移 <strong>{{ preview.localTotal ?? preview.total }}</strong> 个
        <span class="oss-sync-mini__muted">（约 {{ formatBytes(preview.totalBytes) }}）</span>
        <template v-if="(preview.reuploadTotal ?? 0) > 0">
          · 可重传 <strong>{{ preview.reuploadTotal }}</strong> 个
        </template>
      </p>
    </template>

    <div v-if="preview?.ready" class="oss-sync-mini__opts">
      <ElCheckbox v-model="includeReupload" size="small">重新上传</ElCheckbox>
      <ElCheckbox v-model="deleteLocal" size="small">同步后删本地</ElCheckbox>
    </div>

    <div v-if="phase !== 'idle'" class="oss-sync-mini__progress">
      <ElProgress
        :percentage="progress"
        :stroke-width="6"
        :striped="phase === 'running'"
        :striped-flow="phase === 'running'"
        :status="progressStatus"
        :indeterminate="batchBusy && progress === 0"
      />
      <span class="oss-sync-mini__status">{{ statusText }}</span>
    </div>

    <div v-if="failures.length" class="oss-sync-mini__failures">
      <div v-for="f in failures.slice(0, 3)" :key="f.id" class="oss-sync-mini__fail-item">
        #{{ f.id }} {{ f.name }}：{{ f.error }}
      </div>
    </div>

    <div class="oss-sync-mini__actions">
      <ElButton
        type="warning"
        size="small"
        :loading="running"
        :disabled="!preview?.ready || !actionTotal"
        @click="handleSync"
      >
        同步到{{ label }}
      </ElButton>
    </div>
  </div>
</template>

<script setup lang="ts">
import {
  getOssSyncPreviewApi,
  runOssSyncApi,
  type OssSyncPreview
} from '@/api/backend/system/config'

const props = defineProps<{
  driver: string
  label: string
}>()

const preview = ref<OssSyncPreview | null>(null)
const previewLoading = ref(false)
const running = ref(false)
const phase = ref<'idle' | 'running' | 'done' | 'error'>('idle')
const batchBusy = ref(false)
const deleteLocal = ref(false)
const includeReupload = ref(false)
const progress = ref(0)
const failures = ref<Array<{ id: number; name: string; url: string; error: string }>>([])
const stats = reactive({ total: 0, processed: 0, success: 0, failed: 0 })

const actionTotal = computed(() => {
  if (!preview.value?.ready) return 0
  const local = preview.value.localTotal ?? preview.value.total ?? 0
  const reupload = includeReupload.value ? (preview.value.reuploadTotal ?? 0) : 0
  return local + reupload
})

const progressStatus = computed(() => {
  if (phase.value === 'done') return 'success'
  if (phase.value === 'error') return 'exception'
  return undefined
})

const statusText = computed(() => {
  if (phase.value === 'done') {
    return `完成：成功 ${stats.success}，失败 ${stats.failed}`
  }
  if (batchBusy.value) {
    return `上传中 ${stats.success} / ${stats.total}`
  }
  if (phase.value === 'running') {
    return `准备下一批 ${stats.success} / ${stats.total}`
  }
  return `${stats.processed} / ${stats.total}`
})

const formatBytes = (n?: number) => {
  const v = Number(n || 0)
  if (v < 1024) return `${v} B`
  if (v < 1024 * 1024) return `${(v / 1024).toFixed(1)} KB`
  if (v < 1024 * 1024 * 1024) return `${(v / 1024 / 1024).toFixed(1)} MB`
  return `${(v / 1024 / 1024 / 1024).toFixed(2)} GB`
}

const loadPreview = async () => {
  previewLoading.value = true
  try {
    preview.value = await getOssSyncPreviewApi(props.driver)
  } catch (e) {
    preview.value = null
    console.error(e)
  } finally {
    previewLoading.value = false
  }
}

const handleSync = async () => {
  if (!actionTotal.value) return
  try {
    await ElMessageBox.confirm(
      `将上传 ${actionTotal.value} 个附件到「${props.label}」。请确认上方配置已保存。`,
      '确认同步',
      { type: 'warning' }
    )
  } catch {
    return
  }

  running.value = true
  phase.value = 'running'
  failures.value = []
  Object.assign(stats, { total: actionTotal.value, processed: 0, success: 0, failed: 0 })
  progress.value = 0
  batchBusy.value = false

  await nextTick()

  const pageSize = 10
  let done = false
  let lastLocalId = 0
  let lastReuploadId = 0

  try {
    while (!done) {
      batchBusy.value = true
      await nextTick()

      const res = await runOssSyncApi({
        targetDriver: props.driver,
        pageSize,
        deleteLocal: deleteLocal.value,
        includeReupload: includeReupload.value,
        lastLocalId,
        lastReuploadId,
        dryRun: false
      })

      batchBusy.value = false
      if (res.total) stats.total = res.total
      stats.processed += res.processed
      stats.success += res.success
      stats.failed += res.failed
      if (res.failures?.length) {
        failures.value.push(...res.failures)
      }
      lastLocalId = res.lastLocalId || lastLocalId
      lastReuploadId = res.lastReuploadId || lastReuploadId
      progress.value = stats.total
        ? Math.min(100, Math.round((stats.processed / stats.total) * 100))
        : 100
      done = res.done || res.processed === 0
      await nextTick()
    }
    phase.value = 'done'
    progress.value = 100
    const msg = `已处理 ${stats.processed} 条：成功 ${stats.success}，失败 ${stats.failed}`
    if (stats.failed > 0) {
      ElMessage.warning(msg)
    } else {
      ElMessage.success(msg)
    }
    await loadPreview()
  } catch (e: any) {
    phase.value = 'error'
    ElMessage.error(e?.message || '同步失败')
  } finally {
    running.value = false
    batchBusy.value = false
  }
}

watch(
  () => props.driver,
  () => {
    phase.value = 'idle'
    progress.value = 0
    failures.value = []
    loadPreview()
  },
  { immediate: true }
)
</script>

<style scoped lang="scss">
.oss-sync-mini {
  margin-top: 20px;
  padding: 12px 14px;
  border: 1px solid var(--art-gray-200);
  border-radius: 6px;
  background: var(--art-gray-50, #fafafa);
}

.oss-sync-mini__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}

.oss-sync-mini__title {
  font-size: 13px;
  font-weight: 600;
  color: var(--art-gray-700);
}

.oss-sync-mini__alert {
  margin-bottom: 8px;
}

.oss-sync-mini__stats {
  margin: 0 0 8px;
  font-size: 13px;
  line-height: 1.5;
  color: var(--art-gray-600);
}

.oss-sync-mini__muted {
  color: var(--art-gray-400);
  font-size: 12px;
}

.oss-sync-mini__opts {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 16px;
  margin-bottom: 8px;
}

.oss-sync-mini__progress {
  margin: 8px 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.oss-sync-mini__status {
  font-size: 12px;
  color: var(--art-gray-500);
}

.oss-sync-mini__failures {
  margin-bottom: 8px;
  max-height: 72px;
  overflow-y: auto;
}

.oss-sync-mini__fail-item {
  font-size: 11px;
  color: var(--art-gray-500);
  line-height: 1.4;
}

.oss-sync-mini__actions {
  margin-top: 4px;
}

:deep(.dark) {
  .oss-sync-mini {
    background: var(--art-gray-800);
    border-color: var(--art-gray-700);
  }
}
</style>
