/**
 * 附件 / 上传资源 URL 工具（管理后台）
 *
 * 约定：
 * - url / path：库中 object key（如 /upload/20260816/xxx.png），用于 CMS 入库
 * - cdnUrl：后端 storager.CdnUrl 拼好的可访问地址（本地相对路径或 CDN 完整 URL）
 */

import { computed, reactive, unref, watch, type MaybeRef } from 'vue'
import { adminRequest } from '@/utils/http'

/** 列表/选择器预览、复制链接用 */
export function attachmentDisplayUrl(item: { url?: string; cdnUrl?: string; accessUrl?: string }): string {
  return item.cdnUrl || item.accessUrl || item.url || ''
}

/** path 规范化（用于缓存 key） */
function normalizeMediaPath(path?: string | null): string {
  if (!path) return ''
  const raw = path.trim()
  if (!raw || /^https?:\/\//i.test(raw)) return raw
  return raw.startsWith('/') ? raw : `/${raw}`
}

/** 已入库 path 的简单预览（本地 /attachment 可走 vite 代理） */
export function storagePathPreview(path?: string | null): string {
  if (!path) return ''
  if (/^https?:\/\//i.test(path)) return path
  const normalized = normalizeMediaPath(path)
  if (normalized.startsWith('/attachment/')) return normalized
  return normalized
}

/** 展示用 URL：优先 cdnUrl，否则读缓存或本地 fallback */
export function mediaDisplayUrl(path?: string | null, cdnUrl?: string | null): string {
  if (cdnUrl) return cdnUrl
  if (!path) return ''
  if (/^https?:\/\//i.test(path)) return path
  const key = normalizeMediaPath(path)
  if (mediaPreviewCache[key]) return mediaPreviewCache[key]
  return storagePathPreview(path)
}

/** 运行时 path → 可访问 URL 缓存（由 ensureMediaPreviewUrl 填充） */
export const mediaPreviewCache = reactive<Record<string, string>>({})

/** 向后端解析 path 并写入缓存 */
export async function ensureMediaPreviewUrl(path?: string | null): Promise<string> {
  if (!path) return ''
  if (/^https?:\/\//i.test(path)) return path
  const key = normalizeMediaPath(path)
  if (mediaPreviewCache[key]) return mediaPreviewCache[key]

  try {
    const res = await adminRequest.get<{ url?: string; list?: Record<string, string> }>({
      url: '/upload/resolve-url',
      params: { path: key },
      showErrorMessage: false
    })
    const url = res?.url || res?.list?.[path] || res?.list?.[key] || storagePathPreview(path)
    mediaPreviewCache[key] = url
    if (path !== key) mediaPreviewCache[path] = url
    return url
  } catch {
    const fallback = storagePathPreview(path)
    mediaPreviewCache[key] = fallback
    return fallback
  }
}

/** 批量解析（配置页等多图场景） */
export async function ensureMediaPreviewUrls(paths: string[]): Promise<void> {
  const pending = paths
    .map((p) => p?.trim())
    .filter((p) => p && !/^https?:\/\//i.test(p))
    .filter((p) => !mediaPreviewCache[normalizeMediaPath(p)])

  if (!pending.length) return

  try {
    const res = await adminRequest.get<{ list?: Record<string, string> }>({
      url: '/upload/resolve-url',
      params: { paths: pending.join(',') },
      showErrorMessage: false
    })
    const list = res?.list || {}
    for (const [p, url] of Object.entries(list)) {
      if (!url) continue
      mediaPreviewCache[normalizeMediaPath(p)] = url
      mediaPreviewCache[p] = url
    }
  } catch {
    pending.forEach((p) => {
      mediaPreviewCache[normalizeMediaPath(p)] = storagePathPreview(p)
    })
  }
}

/** 根据 path → url 映射解析（Markdown 预览等） */
export function resolveMediaUrl(path: string, urlMap?: Record<string, string>): string {
  if (!path) return path
  if (/^https?:\/\//i.test(path)) return path
  const normalized = normalizeMediaPath(path)
  if (urlMap?.[normalized]) return urlMap[normalized]
  if (urlMap?.[path]) return urlMap[path]
  return mediaDisplayUrl(path)
}

/** 将 Markdown/HTML 预览中的相对资源路径替换为可访问 URL */
export function resolveMediaUrlsInHtml(html: string, urlMap?: Record<string, string>): string {
  if (!html || !urlMap || !Object.keys(urlMap).length) return html
  let out = html
  for (const [path, url] of Object.entries(urlMap)) {
    if (!path || !url) continue
    out = out.split(path).join(url)
    const noSlash = path.startsWith('/') ? path.slice(1) : path
    if (noSlash !== path) out = out.split(noSlash).join(url)
  }
  return out
}

/** 响应式媒体展示 URL（Logo/头像等全局展示用） */
export function useMediaDisplayUrl(source: MaybeRef<string | undefined | null>) {
  const displayUrl = computed(() => {
    const path = unref(source)
    if (!path) return ''
    return mediaDisplayUrl(path)
  })
  watch(
    () => unref(source),
    (path) => {
      if (path && !/^https?:\/\//i.test(path)) void ensureMediaPreviewUrl(path)
    },
    { immediate: true }
  )
  return displayUrl
}

/** 上传接口响应 → 写入 CMS/表单应存的 path */
export function uploadStoragePath(res?: { path?: string; url?: string } | null): string {
  if (!res) return ''
  const raw = (res.path || res.url || '').trim()
  if (!raw) return ''
  if (/^https?:\/\//i.test(raw)) {
    try {
      const u = new URL(raw)
      return u.pathname.startsWith('/') ? u.pathname : `/${u.pathname}`
    } catch {
      return raw
    }
  }
  return raw.startsWith('/') ? raw : `/${raw}`
}
