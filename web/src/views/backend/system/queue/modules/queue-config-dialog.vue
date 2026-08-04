<!-- 消息队列 Topic 配置弹窗 -->
<template>
  <ElDialog v-model="dialogVisible" title="Topic 运行配置" width="520px" align-center :close-on-click-modal="false">
    <ElForm ref="formRef" :model="formData" :rules="rules" label-width="110px">
      <ElFormItem label="Topic">
        <ElInput :model-value="formData.topic" disabled />
      </ElFormItem>
      <ElFormItem label="显示名称">
        <ElInput v-model="formData.title" placeholder="可选" />
      </ElFormItem>
      <ElFormItem label="Worker 数" prop="workers">
        <ElInputNumber v-model="formData.workers" :min="1" :max="64" class="!w-full" />
        <div class="mt-1 text-xs text-g-400">同一进程内并行消费的 goroutine 数量</div>
      </ElFormItem>
      <ElFormItem label="最大重试" prop="maxRetry">
        <ElInputNumber v-model="formData.maxRetry" :min="0" :max="20" class="!w-full" />
      </ElFormItem>
      <ElFormItem label="重试间隔(秒)" prop="retryDelaySec">
        <ElInputNumber v-model="formData.retryDelaySec" :min="0" :max="86400" :step="5" class="!w-full" />
        <div class="mt-1 text-xs text-g-400">0 = 立即重试；需 Redis 驱动才支持延迟重试</div>
      </ElFormItem>
      <ElFormItem label="状态">
        <ElSwitch v-model="formData.status" :active-value="1" :inactive-value="0" active-text="启用" inactive-text="禁用" />
      </ElFormItem>
      <ElFormItem label="备注">
        <ElInput v-model="formData.remark" type="textarea" :rows="2" />
      </ElFormItem>
    </ElForm>
    <template #footer>
      <ElButton @click="dialogVisible = false">取消</ElButton>
      <ElButton type="primary" :loading="submitting" @click="handleSubmit">保存并生效</ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import type { FormInstance, FormRules } from 'element-plus'
  import { ElMessage } from 'element-plus'
  import { fetchQueueConfigSave, type QueueTopicStats } from '@/api/backend/system/queue'

  interface Props {
    visible: boolean
    data?: QueueTopicStats | null
  }
  interface Emits {
    (e: 'update:visible', v: boolean): void
    (e: 'success'): void
  }

  const props = defineProps<Props>()
  const emit = defineEmits<Emits>()

  const dialogVisible = computed({
    get: () => props.visible,
    set: v => emit('update:visible', v)
  })

  const formRef = ref<FormInstance>()
  const submitting = ref(false)
  const formData = reactive({
    id: 0,
    topic: '',
    title: '',
    workers: 1,
    maxRetry: 3,
    retryDelaySec: 0,
    status: 1,
    remark: ''
  })

  const rules: FormRules = {
    workers: [{ required: true, message: '请填写 Worker 数', trigger: 'blur' }]
  }

  watch(() => props.visible, (val) => {
    if (val && props.data) {
      Object.assign(formData, {
        id: props.data.configId || 0,
        topic: props.data.topic,
        title: props.data.title || props.data.topic,
        workers: props.data.workers || 1,
        maxRetry: props.data.maxRetry ?? 3,
        retryDelaySec: props.data.retryDelaySec ?? 0,
        status: props.data.status ?? 1,
        remark: props.data.remark || ''
      })
      nextTick(() => formRef.value?.clearValidate())
    }
  })

  const handleSubmit = async () => {
    if (!formRef.value) return
    await formRef.value.validate(async (valid) => {
      if (!valid) return
      submitting.value = true
      try {
        await fetchQueueConfigSave({ ...formData })
        ElMessage.success('配置已保存，Worker 已热更新')
        emit('success')
        dialogVisible.value = false
      } catch { /* */ } finally {
        submitting.value = false
      }
    })
  }
</script>
