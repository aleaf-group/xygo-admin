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

// ===================== 文件上传 =====================

type UploadFileReq struct {
	g.Meta `path:"/admin/upload/file" method:"post" mime:"multipart/form-data" tags:"AdminUpload" summary:"上传文件"`
	adminin.UploadFileInp
}

type UploadFileRes struct {
	*adminin.UploadFileModel
}

type ResolveMediaUrlReq struct {
	g.Meta `path:"/admin/upload/resolve-url" method:"get" tags:"AdminUpload" summary:"解析资源 path 为可访问 URL"`
	adminin.ResolveMediaUrlInp
}

type ResolveMediaUrlRes struct {
	*adminin.ResolveMediaUrlModel
}
