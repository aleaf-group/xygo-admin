<!-- +----------------------------------------------------------------------
  | XYGo Admin [ Vue3 + GoFrame 企业级中后台管理系统 ]
  +----------------------------------------------------------------------
  | Copyright (c) 2026 大连星韵网络科技有限公司 All rights reserved.
  +----------------------------------------------------------------------
  | Licensed ( https://opensource.org/licenses/MIT )
  +----------------------------------------------------------------------
  | Author: 喜羊羊 <751300685@qq.com>
  +---------------------------------------------------------------------- -->
<!-- 图片上传组件 -->
<template>
  <div class="art-image-upload" :class="{ 'is-compact': compact }">
    <!-- 单图上传 -->
    <div v-if="!multiple" class="single-upload">
      <div v-if="imageUrl" class="image-preview">
        <ElImage 
          ref="imageRef"
          :src="singleDisplayUrl" 
          fit="contain" 
          class="preview-image" 
          :preview-src-list="[singleDisplayUrl]"
          :initial-index="0"
          preview-teleported
        />
        <div class="image-actions">
          <div class="action-btn" @click="handlePreview">
            <ArtSvgIcon icon="ri:eye-line" :size="16" />
          </div>
          <div class="action-btn" @click="handleRemove">
            <ArtSvgIcon icon="ri:delete-bin-line" :size="16" />
          </div>
        </div>
      </div>
      <ElUpload
        v-else
        :http-request="customUpload"
        :show-file-list="false"
        :before-upload="beforeUpload"
        class="upload-trigger"
      >
        <div class="upload-placeholder">
          <ArtSvgIcon icon="ri:image-add-line" :size="32" />
          <div class="upload-text">点击上传图片</div>
          <div class="upload-hint">{{ acceptHint }}</div>
        </div>
      </ElUpload>
    </div>

    <!-- 多图上传 -->
    <div v-else class="multiple-upload">
      <VueDraggable 
        v-model="imageList" 
        class="image-list"
        :animation="200"
        @end="handleDragEnd"
      >
        <div
          v-for="(element, index) in imageList"
          :key="element.id"
          class="image-item"
        >
          <ElImage :src="displayUrlFor(element.url)" fit="cover" class="preview-image" :preview-src-list="imageList.map(img => displayUrlFor(img.url))" />
          <div class="image-actions">
            <div class="action-btn" @click="handlePreviewMultiple(index)">
              <ArtSvgIcon icon="ri:eye-line" :size="14" />
            </div>
            <div class="action-btn" @click="handleRemoveMultiple(index)">
              <ArtSvgIcon icon="ri:delete-bin-line" :size="14" />
            </div>
          </div>
        </div>
      </VueDraggable>
      
      <ElUpload
        v-if="!limit || imageList.length < limit"
        :http-request="customUploadMultiple"
        :show-file-list="false"
        :before-upload="beforeUpload"
        class="upload-trigger-small"
      >
        <div class="upload-placeholder-small">
          <ArtSvgIcon icon="ri:add-line" :size="24" />
        </div>
      </ElUpload>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { VueDraggable } from 'vue-draggable-plus'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { UploadProps, UploadRequestOptions } from 'element-plus'
import { uploadImageApi } from '@/api/backend/common/upload'
import {
  ensureMediaPreviewUrl,
  mediaDisplayUrl,
  mediaPreviewCache,
  uploadStoragePath
} from '@/utils/media-url'

defineOptions({ name: 'ArtImageUpload' })

interface Props {
  modelValue?: string | string[]
  multiple?: boolean
  limit?: number
  accept?: string
  maxSize?: number // MB
  compact?: boolean // 紧凑模式（用于数组编辑器等场景）
}

const props = withDefaults(defineProps<Props>(), {
  multiple: false,
  limit: 9,
  accept: 'image/jpeg,image/jpg,image/png,image/gif,image/webp',
  maxSize: 5,
  compact: false
})

const emit = defineEmits<{
  'update:modelValue': [value: string | string[]]
  'change': [value: string | string[]]
}>()


// 单图模式（v-model 存 object key）
const imageUrl = computed({
  get: () => {
    if (props.multiple) return ''
    return (props.modelValue as string) || ''
  },
  set: (val) => {
    emit('update:modelValue', val || '')
    emit('change', val || '')
  }
})

const singleDisplayUrl = computed(() => mediaDisplayUrl(imageUrl.value))

watch(
  () => imageUrl.value,
  (path) => {
    if (path && !/^https?:\/\//i.test(path)) ensureMediaPreviewUrl(path)
  },
  { immediate: true }
)

const displayUrlFor = (path: string) => mediaDisplayUrl(path)

watch(
  () => props.modelValue,
  (val) => {
    if (!props.multiple) return
    const paths: string[] = []
    if (Array.isArray(val)) paths.push(...val.filter(Boolean))
    else if (typeof val === 'string' && val) paths.push(...val.split(',').map((s) => s.trim()).filter(Boolean))
    paths.forEach((p) => {
      if (!/^https?:\/\//i.test(p)) ensureMediaPreviewUrl(p)
    })
  },
  { immediate: true, deep: true }
)

// 多图模式
const imageList = computed({
  get: () => {
    if (!props.multiple) return []
    const val = props.modelValue
    if (Array.isArray(val)) {
      return val.map((url, index) => ({ url, id: index }))
    }
    if (typeof val === 'string' && val) {
      return val.split(',').map((url, index) => ({ url: url.trim(), id: index }))
    }
    return []
  },
  set: (val) => {
    const urls = val.map(item => item.url)
    const value = urls.join(',')
    emit('update:modelValue', value)
    emit('change', value)
  }
})

// 文件类型提示
const acceptHint = computed(() => {
  const types = props.accept.split(',').map(t => t.split('/')[1]?.toUpperCase() || t)
  return `支持 ${types.join('、')} 格式，不超过 ${props.maxSize}MB`
})

// 上传前校验
const beforeUpload: UploadProps['beforeUpload'] = (file) => {
  const isValidType = props.accept.split(',').some(type => file.type === type.trim())
  if (!isValidType) {
    ElMessage.error(`请上传正确格式的图片文件`)
    return false
  }

  const isValidSize = file.size / 1024 / 1024 < props.maxSize
  if (!isValidSize) {
    ElMessage.error(`图片大小不能超过 ${props.maxSize}MB`)
    return false
  }

  return true
}

// 自定义上传 - 单图
const customUpload = async (options: UploadRequestOptions) => {
  try {
    // request.post 成功时直接返回 data.data，失败时抛出异常
    const data = await uploadImageApi(options.file as File)
    
    if (data?.url || data?.path) {
      const path = uploadStoragePath(data)
      imageUrl.value = path
      if (data.url) mediaPreviewCache[path] = data.url
      ElMessage.success('上传成功')
      options.onSuccess(data)
    } else {
      throw new Error('上传响应数据格式错误')
    }
  } catch (error: any) {
    console.error('上传失败:', error)
    // 错误已经在拦截器中显示了，这里不再重复显示
    options.onError(error as any)
  }
}

// 自定义上传 - 多图
const customUploadMultiple = async (options: UploadRequestOptions) => {
  try {
    // request.post 成功时直接返回 data.data，失败时抛出异常
    const data = await uploadImageApi(options.file as File)
    
    if (data?.url || data?.path) {
      const path = uploadStoragePath(data)
      if (data.url) mediaPreviewCache[path] = data.url
      const newList = [...imageList.value, { url: path, id: Date.now() }]
      imageList.value = newList
      ElMessage.success('上传成功')
      options.onSuccess(data)
    } else {
      throw new Error('上传响应数据格式错误')
    }
  } catch (error: any) {
    console.error('上传失败:', error)
    // 错误已经在拦截器中显示了，这里不再重复显示
    options.onError(error as any)
  }
}

// 图片ref
const imageRef = ref()

// 预览图片
const handlePreview = () => {
  // 手动触发ElImage的预览功能
  if (imageRef.value && imageRef.value.$el) {
    const imgElement = imageRef.value.$el.querySelector('img')
    if (imgElement) {
      imgElement.click()
    }
  }
}

// 删除图片
const handleRemove = () => {
  imageUrl.value = ''
  emit('update:modelValue', '')
  emit('change', '')
}

// 多图预览
const handlePreviewMultiple = (index: number) => {
  // ElImage 的 preview-src-list 会自动处理预览
}

// 多图删除
const handleRemoveMultiple = (index: number) => {
  const newList = imageList.value.filter((_, i) => i !== index)
  imageList.value = newList
  const urls = newList.map(img => img.url)
  emit('update:modelValue', urls)
  emit('change', urls)
}

// 拖拽结束
const handleDragEnd = () => {
  // imageList 的 computed setter 会自动触发更新
}
</script>

<style scoped lang="scss">
.art-image-upload {
  width: 100%;
}

// 单图上传
.single-upload {
  width: 100%;
  max-width: 200px;
}

.image-preview {
  position: relative;
  width: 100%;
  aspect-ratio: 4 / 3;
  border-radius: 8px;
  overflow: hidden;
  border: 1px solid var(--art-gray-300);

  .preview-image {
    width: 100%;
    height: 100%;
    object-fit: contain;
  }

  .image-actions {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 12px;
    background: rgba(0, 0, 0, 0.5);
    opacity: 0;
    transition: opacity 0.3s;

    .action-btn {
      width: 36px;
      height: 36px;
      display: flex;
      align-items: center;
      justify-content: center;
      background: rgba(255, 255, 255, 0.9);
      border-radius: 50%;
      cursor: pointer;
      color: var(--art-gray-700);
      transition: all 0.2s;

      &:hover {
        background: #fff;
        transform: scale(1.1);
      }
    }
  }

  &:hover .image-actions {
    opacity: 1;
  }
}

.upload-trigger {
  width: 100%;

  :deep(.el-upload) {
    width: 100%;
  }
}

.upload-placeholder {
  width: 100%;
  aspect-ratio: 4 / 3;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  border: 2px dashed var(--art-gray-300);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.3s;
  color: var(--art-gray-500);

  &:hover {
    border-color: var(--theme-color);
    background: var(--theme-color-alpha-5);
    color: var(--theme-color);
  }

  .upload-text {
    margin-top: 12px;
    font-size: 14px;
    font-weight: 500;
  }

  .upload-hint {
    margin-top: 4px;
    font-size: 12px;
    color: var(--art-gray-400);
  }
}

// 多图上传
.multiple-upload {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.image-list {
  display: contents;
}

.image-item {
  position: relative;
  width: 120px;
  height: 120px;
  border-radius: 8px;
  overflow: hidden;
  border: 1px solid var(--art-gray-300);
  cursor: move;

  .preview-image {
    width: 100%;
    height: 100%;
  }

  .image-actions {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    background: rgba(0, 0, 0, 0.6);
    opacity: 0;
    transition: opacity 0.3s;

    .action-btn {
      width: 28px;
      height: 28px;
      display: flex;
      align-items: center;
      justify-content: center;
      background: rgba(255, 255, 255, 0.9);
      border-radius: 50%;
      cursor: pointer;
      color: var(--art-gray-700);
      transition: all 0.2s;

      &:hover {
        background: #fff;
        transform: scale(1.1);
      }
    }
  }

  &:hover .image-actions {
    opacity: 1;
  }
}

.upload-trigger-small {
  :deep(.el-upload) {
    width: 120px;
    height: 120px;
  }
}

.upload-placeholder-small {
  width: 120px;
  height: 120px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 2px dashed var(--art-gray-300);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.3s;
  color: var(--art-gray-400);

  &:hover {
    border-color: var(--theme-color);
    background: var(--theme-color-alpha-5);
    color: var(--theme-color);
  }
}

// 紧凑模式：适用于数组编辑器等场景
.art-image-upload.is-compact {
  .single-upload {
    max-width: 100% !important;
  }

  .image-preview {
    aspect-ratio: auto !important;
    height: 32px !important;
    min-height: 32px !important;
    max-height: 32px !important;
    width: 100% !important;
    border-radius: 4px !important;

    .preview-image {
      height: 100% !important;
      width: auto !important;
      object-fit: contain !important;
    }

    .image-actions {
      .action-btn {
        width: 24px !important;
        height: 24px !important;

        .art-svg-icon {
          font-size: 12px !important;
        }
      }
    }
  }

  .upload-trigger {
    :deep(.el-upload) {
      width: 100% !important;
    }
  }

  .upload-placeholder {
    aspect-ratio: auto !important;
    height: 32px !important;
    min-height: 32px !important;
    max-height: 32px !important;
    flex-direction: row !important;
    gap: 6px !important;
    padding: 0 12px !important;
    margin: 0 !important;
    border-width: 1px !important;
    border-radius: 4px !important;

    .art-svg-icon {
      font-size: 16px !important;
    }

    .upload-text {
      font-size: 12px !important;
      margin: 0 !important;
    }

    .upload-hint {
      display: none !important;
    }
  }

  .multiple-upload {
    gap: 8px !important;
  }

  .image-item {
    width: auto !important;
    height: 32px !important;
    min-height: 32px !important;
    max-height: 32px !important;
    border-radius: 4px !important;

    .preview-image {
      height: 100% !important;
      width: auto !important;
    }

    .image-actions {
      .action-btn {
        width: 20px !important;
        height: 20px !important;

        .art-svg-icon {
          font-size: 10px !important;
        }
      }
    }
  }

  .upload-trigger-small {
    :deep(.el-upload) {
      width: 32px !important;
      height: 32px !important;
    }
  }

  .upload-placeholder-small {
    width: 32px !important;
    height: 32px !important;
    min-height: 32px !important;
    max-height: 32px !important;
    border-width: 1px !important;
    border-radius: 4px !important;

    .art-svg-icon {
      font-size: 16px !important;
    }
  }
}
</style>
