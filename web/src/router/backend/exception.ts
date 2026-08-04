// +----------------------------------------------------------------------
// | XYGo Admin [ Vue3 + GoFrame 企业级中后台管理系统 ]
// +----------------------------------------------------------------------
// | Copyright (c) 2026 大连星韵网络科技有限公司 All rights reserved.
// +----------------------------------------------------------------------
// | Licensed ( https://opensource.org/licenses/MIT )
// +----------------------------------------------------------------------
// | Author: 喜羊羊 <751300685@qq.com>
// +----------------------------------------------------------------------

import { AppRouteRecord } from '@/types/router'

export const exceptionRoutes: AppRouteRecord = {
  path: '/exception',
  name: 'Exception',
  component: '/index/index',
  meta: {
    title: 'menus.exception.title',
    icon: 'ri:error-warning-line'
  },
  children: [
    {
      path: '403',
      name: 'AdminException403',
      component: '/common/exception/403',
      meta: {
        title: 'menus.exception.forbidden',
        keepAlive: true,
        isHideTab: true,
        isFullPage: true
      }
    },
    {
      path: '404',
      name: 'AdminException404',
      component: '/common/exception/404',
      meta: {
        title: 'menus.exception.notFound',
        keepAlive: true,
        isHideTab: true,
        isFullPage: true
      }
    },
    {
      path: '500',
      name: 'AdminException500',
      component: '/common/exception/500',
      meta: {
        title: 'menus.exception.serverError',
        keepAlive: true,
        isHideTab: true,
        isFullPage: true
      }
    }
  ]
}
