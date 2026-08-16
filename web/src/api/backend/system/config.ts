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
 * 系统配置管理 API
 */
import { adminRequest } from '@/utils/http'

/**
 * 配置分组相关接口
 */

// 配置分组项
export interface ConfigGroupItem {
  group: string
  groupName: string
  icon: string
  description: string
  sort: number
}

// 获取配置分组列表
export const getConfigGroupListApi = () => {
  return adminRequest.get<{ list: ConfigGroupItem[] }>({
    url: '/config/group/list'
  })
}

// 保存配置分组（添加或编辑）
export const saveConfigGroupApi = (data: {
  group: string
  groupName: string
  icon?: string
  description?: string
  sort?: number
  isEdit: boolean
}) => {
  return adminRequest.post({
    url: '/config/group/save',
    data
  })
}

// 删除配置分组
export const deleteConfigGroupApi = (group: string) => {
  return adminRequest.post({
    url: '/config/group/delete',
    data: { group }
  })
}

/**
 * 配置项相关接口
 */

// 配置项
export interface ConfigItem {
  key: string
  value: string
}

// 配置项 Schema
export interface ConfigSchemaItem {
  id: number
  group: string
  groupName: string
  name: string
  key: string
  value: string
  type: string
  options: any
  rules: any
  sort: number
  remark: string
  allowDel: number  // 是否允许删除：0-否，1-是
}

// 获取配置 Schema
export const getConfigSchemaApi = () => {
  return adminRequest.get<{ list: ConfigSchemaItem[] }>({
    url: '/config/schema'
  })
}

// 获取指定分组的配置列表
export const getConfigListApi = (group: string) => {
  return adminRequest.get<{
    group: string
    items: Record<string, string>
    list: ConfigItem[]
  }>({
    url: '/config/list',
    params: { group }
  })
}

// 保存分组配置
export const saveConfigApi = (data: {
  group: string
  items: ConfigItem[]
}) => {
  return adminRequest.post({
    url: '/config/save',
    data
  })
}

// 创建配置项
export const createConfigItemApi = (data: {
  group: string
  groupName: string
  name: string
  key: string
  type: string
  value?: string
  options?: string
  rules?: string
  sort?: number
  remark?: string
}) => {
  return adminRequest.post({
    url: '/config/create',
    data
  })
}

// 删除配置项
export const deleteConfigItemApi = (key: string) => {
  return adminRequest.post({
    url: '/config/delete',
    data: { key }
  })
}

/** OSS 本地附件迁移 */
export interface OssSyncPreview {
  targetDriver: string
  targetLabel: string
  ready: boolean
  localTotal: number
  reuploadTotal: number
  total: number
  totalBytes: number
}

export interface OssSyncRunResult {
  total: number
  processed: number
  success: number
  failed: number
  skipped: number
  lastLocalId: number
  lastReuploadId: number
  done: boolean
  failures: Array<{ id: number; name: string; url: string; error: string }>
}

export const getOssSyncPreviewApi = (targetDriver: string) => {
  return adminRequest.get<OssSyncPreview>({
    url: '/config/oss-sync/preview',
    params: { targetDriver }
  })
}

export const runOssSyncApi = (data: {
  targetDriver: string
  pageSize?: number
  deleteLocal?: boolean
  includeReupload?: boolean
  lastLocalId?: number
  lastReuploadId?: number
  dryRun?: boolean
}) => {
  return adminRequest.post<OssSyncRunResult>({
    url: '/config/oss-sync/run',
    data,
    timeout: 120000
  })
}
