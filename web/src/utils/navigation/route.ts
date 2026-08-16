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
 * 路由工具模块
 *
 * 提供路由处理和菜单路径相关的工具函数
 *
 * ## 主要功能
 *
 * - iframe 路由检测，判断是否为外部嵌入页面
 * - 菜单项有效性验证，过滤隐藏和无效菜单
 * - 路径标准化处理，统一路径格式
 * - 递归查找菜单树中第一个有效路径
 * - 支持多级嵌套菜单的路径解析
 *
 * ## 使用场景
 *
 * - 系统初始化时获取默认跳转路径
 * - 菜单权限过滤后获取首个可访问页面
 * - 路由重定向逻辑处理
 * - iframe 页面特殊处理
 *
 * @module utils/navigation/route
 * @author Art Design Pro Team
 */

import { AppRouteRecord } from '@/types'

// 检查是否为 iframe 路由
export function isIframe(url: string): boolean {
  return url.startsWith('/outside/iframe/')
}

/**
 * 验证菜单项是否有效
 * @param menuItem 菜单项
 * @returns 是否为有效菜单项
 */
const isValidMenuItem = (menuItem: AppRouteRecord): boolean => {
  return !!(menuItem.path && menuItem.path.trim() && !menuItem.meta?.isHide)
}

/**
 * 标准化路径格式
 * @param path 路径
 * @returns 标准化后的路径
 */
const normalizePath = (path: string): string => {
  return path.startsWith('/') ? path : `/${path}`
}

/**
 * 递归获取菜单的第一个有效路径
 * @param menuList 菜单列表
 * @returns 第一个有效路径，如果没有找到则返回空字符串
 */
export const getFirstMenuPath = (menuList: AppRouteRecord[]): string => {
  if (!Array.isArray(menuList) || menuList.length === 0) {
    return ''
  }

  for (const menuItem of menuList) {
    if (!isValidMenuItem(menuItem)) {
      continue
    }

    // 如果有子菜单，优先查找子菜单
    if (menuItem.children?.length) {
      const childPath = getFirstMenuPath(menuItem.children)
      if (childPath) {
        return childPath
      }
    }

    // 返回当前菜单项的标准化路径
    return normalizePath(menuItem.path!)
  }

  return ''
}

/** 个人中心页面 component（菜单 SSOT，不随层级变化） */
export const USER_CENTER_COMPONENT = '/system/user-center'

/** 个人中心默认 route name */
export const USER_CENTER_ROUTE_NAME = 'UserCenter'

function normalizeComponentPath(component?: string | (() => Promise<any>)): string {
  if (typeof component !== 'string' || !component.trim()) return ''
  const path = component.trim()
  if (path === 'Layout') return ''
  return path.startsWith('/') ? path : `/${path}`
}

/**
 * 在菜单树中按 component 或 route name 查找实际访问路径
 * （菜单层级调整后 path 会变，component 通常不变）
 */
export function findMenuPath(
  menuList: AppRouteRecord[],
  options: { component?: string; name?: string }
): string {
  if (!Array.isArray(menuList) || !menuList.length) return ''

  const targetComponent = options.component ? normalizeComponentPath(options.component) : ''
  const targetName = options.name?.trim() || ''

  for (const item of menuList) {
    const itemComponent = normalizeComponentPath(item.component as string)
    const itemName = String(item.name || '')
    const matched =
      (targetComponent && itemComponent === targetComponent) ||
      (targetName && itemName === targetName)

    if (matched && item.path?.trim()) {
      return normalizePath(item.path)
    }

    if (item.children?.length) {
      const childPath = findMenuPath(item.children, options)
      if (childPath) return childPath
    }
  }

  return ''
}
