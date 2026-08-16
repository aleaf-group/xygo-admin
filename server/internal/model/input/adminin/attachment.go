// +----------------------------------------------------------------------
// | XYGo Admin [ Vue3 + GoFrame 企业级中后台管理系统 ]
// +----------------------------------------------------------------------
// | Copyright (c) 2026 大连星韵网络科技有限公司 All rights reserved.
// +----------------------------------------------------------------------
// | Licensed ( https://opensource.org/licenses/MIT )
// +----------------------------------------------------------------------
// | Author: 喜羊羊 <751300685@qq.com>
// +----------------------------------------------------------------------

package adminin

import "xygo/internal/model/input/form"

// ===================== 附件列表 =====================

// AttachmentListInp 附件列表入参
type AttachmentListInp struct {
	form.PageReq
	Topic   string `json:"topic" form:"topic" dc:"可选，按分组/主题过滤"`
	Storage string `json:"storage" form:"storage" dc:"可选，按存储驱动过滤"`
	Name    string `json:"name" form:"name" dc:"可选，按文件名模糊搜索"`
}

// AttachmentListItem 附件列表项
type AttachmentListItem struct {
	Id         uint64 `json:"id" dc:"附件ID"`
	Topic      string `json:"topic" dc:"分组/主题"`
	UserId     uint64 `json:"userId" dc:"上传用户ID"`
	Url       string `json:"url" dc:"存储路径（object key，入库/CMS 用）"`
	CdnUrl    string `json:"cdnUrl" dc:"浏览器可访问地址（本地相对路径 / 云存储 CDN 完整 URL）"`
	Name       string `json:"name" dc:"文件名"`
	Size       uint64 `json:"size" dc:"文件大小（字节）"`
	Mimetype   string `json:"mimetype" dc:"MIME类型"`
	Storage    string `json:"storage" dc:"存储驱动"`
	Sha1       string `json:"sha1" dc:"文件SHA1"`
	Quote      uint   `json:"quote" dc:"引用次数"`
	Width      uint   `json:"width" dc:"图片宽度"`
	Height     uint   `json:"height" dc:"图片高度"`
	CreateTime uint   `json:"createTime" dc:"创建时间"`
	UpdateTime uint   `json:"updateTime" dc:"更新时间"`
}

// AttachmentListModel 附件列表响应
type AttachmentListModel struct {
	List []AttachmentListItem `json:"list" dc:"附件列表"`
	form.PageRes
}

// ===================== 删除附件 =====================

// AttachmentDeleteInp 删除附件入参
type AttachmentDeleteInp struct {
	Id uint64 `json:"id" v:"required#附件ID不能为空" dc:"附件ID"`
}
