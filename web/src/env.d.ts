// +----------------------------------------------------------------------
// | XYGo Admin [ Vue3 + GoFrame 企业级中后台管理系统 ]
// +----------------------------------------------------------------------
// | Copyright (c) 2026 大连星韵网络科技有限公司 All rights reserved.
// +----------------------------------------------------------------------
// | Licensed ( https://opensource.org/licenses/MIT )
// +----------------------------------------------------------------------
// | Author: 喜羊羊 <751300685@qq.com>
// +----------------------------------------------------------------------

/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_ADMIN_PATH: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}

declare module 'nprogress'

declare module 'crypto-js'

declare module 'vue-img-cutter'

declare module 'file-saver'

declare module '@wangeditor/editor-for-vue'

declare module 'qrcode.vue' {
  export type Level = 'L' | 'M' | 'Q' | 'H'
  export type RenderAs = 'canvas' | 'svg'
  export type GradientType = 'linear' | 'radial'
  export interface ImageSettings {
    src: string
    height: number
    width: number
    excavate: boolean
  }
  export interface QRCodeProps {
    value: string
    size?: number
    level?: Level
    background?: string
    foreground?: string
    renderAs?: RenderAs
  }
  const QrcodeVue: any
  export default QrcodeVue
}

// 全局变量声明
declare const __APP_VERSION__: string // 版本号
