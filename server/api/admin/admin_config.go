// +----------------------------------------------------------------------
// | XYGo Admin [ Vue3 + GoFrame 企业级中后台管理系统 ]
// +----------------------------------------------------------------------
// | Copyright (c) 2026 大连星韵网络科技有限公司 All rights reserved.
// +----------------------------------------------------------------------
// | Licensed ( https://opensource.org/licenses/MIT )
// +----------------------------------------------------------------------
// | Author: 喜羊羊 <751300685@qq.com>
// +----------------------------------------------------------------------

package admin

import (
	"github.com/gogf/gf/v2/frame/g"

	"xygo/internal/model/input/adminin"
)

// ===================== 配置 Schema =====================

type ConfigSchemaReq struct {
	g.Meta `path:"/admin/config/schema" method:"get" tags:"AdminConfig" summary:"系统配置schema"`
}

type ConfigSchemaRes struct {
	List []adminin.ConfigSchemaItem `json:"list"`
}

// ===================== 配置列表 =====================

type ConfigListReq struct {
	g.Meta `path:"/admin/config/list" method:"get" tags:"AdminConfig" summary:"获取配置列表"`
	adminin.ConfigListInp
}

type ConfigListRes struct {
	adminin.ConfigListModel
}

// ===================== 保存配置 =====================

type ConfigSaveReq struct {
	g.Meta `path:"/admin/config/save" method:"post" tags:"AdminConfig" summary:"保存配置"`
	adminin.ConfigSaveInp
}

type ConfigSaveRes struct{}

// ===================== 创建配置 =====================

type ConfigCreateReq struct {
	g.Meta `path:"/admin/config/create" method:"post" tags:"AdminConfig" summary:"创建配置项"`
	adminin.ConfigCreateInp
}

type ConfigCreateRes struct{}

// ===================== 删除配置 =====================

type ConfigDeleteReq struct {
	g.Meta `path:"/admin/config/delete" method:"post" tags:"AdminConfig" summary:"删除配置项"`
	adminin.ConfigDeleteInp
}

type ConfigDeleteRes struct{}

// ===================== OSS 本地附件迁移 =====================

type OssSyncPreviewReq struct {
	g.Meta `path:"/admin/config/oss-sync/preview" method:"get" tags:"AdminConfig" summary:"预览待迁移本地附件"`
	adminin.OssSyncPreviewInp
}

type OssSyncPreviewRes struct {
	adminin.OssSyncPreviewModel
}

type OssSyncRunReq struct {
	g.Meta `path:"/admin/config/oss-sync/run" method:"post" tags:"AdminConfig" summary:"分批同步本地附件到云存储"`
	adminin.OssSyncRunInp
}

type OssSyncRunRes struct {
	adminin.OssSyncRunModel
}
