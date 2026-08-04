// Package packed 资源嵌入占位（仓库内仅保留此空壳）。
//
// 请勿将 gf pack 生成的大文件提交到 git。
//   - 日常开发：server/ gf run + web/ pnpm dev（dist 可仅 .gitkeep 占位）
//   - 单体部署：cd web && pnpm build → server/resource/public/dist，再启动后端
//   - 发版打包：gf build 临时 pack（产物进 Release，勿 commit 本文件）
package packed
