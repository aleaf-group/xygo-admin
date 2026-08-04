/**
 * 消息队列 API
 */
import { adminRequest } from '@/utils/http'

export interface QueueTopicStats {
  topic: string
  title?: string
  pending: number
  deadSize: number
  rate?: number
  avgTakeMs?: number
  workers?: number
  maxRetry?: number
  retryDelaySec?: number
  status?: number
  configId?: number
  remark?: string
}

/** 获取队列统计 */
export function fetchQueueStats() {
  return adminRequest.get<{ driver: string; topics: QueueTopicStats[] }>({ url: '/queue/stats' })
}

/** 获取已注册 Topic 列表 */
export function fetchQueueTopics() {
  return adminRequest.get<{ list: string[] }>({ url: '/queue/topics' })
}

/** 测试投递消息（delaySec>0 为延迟投递） */
export function fetchQueuePushTest(params: { topic: string; body: string; delaySec?: number }) {
  return adminRequest.post({ url: '/queue/pushTest', params })
}

/** 保存 Topic 运行配置 */
export function fetchQueueConfigSave(params: {
  id?: number
  topic: string
  title?: string
  workers: number
  maxRetry?: number
  retryDelaySec?: number
  status?: number
  remark?: string
  sort?: number
}) {
  return adminRequest.post({ url: '/queue/configSave', params })
}
