package adminin

import "xygo/internal/model/input/form"

// ===================== OSS 本地附件迁移 =====================

// OssSyncPreviewModel 待迁移预览
type OssSyncPreviewModel struct {
	TargetDriver  string `json:"targetDriver" dc:"目标云驱动（local 表示当前为本地）"`
	TargetLabel   string `json:"targetLabel" dc:"目标驱动中文名"`
	Ready         bool   `json:"ready" dc:"是否已配置云存储目标（可执行迁移）"`
	LocalTotal    int    `json:"localTotal" dc:"本地存储待迁移数"`
	ReuploadTotal int    `json:"reuploadTotal" dc:"可重新上传数（已标记云存储但本地文件仍在）"`
	Total         int    `json:"total" dc:"本地待迁移数（同 localTotal，兼容）"`
	TotalBytes    uint64 `json:"totalBytes" dc:"本地待迁移总体积（字节）"`
}

// OssSyncPreviewInp 预览入参
type OssSyncPreviewInp struct {
	TargetDriver string `json:"targetDriver" form:"targetDriver" v:"required#请指定目标云存储" dc:"qiniu|aliyun-oss|tencent-cos"`
}

// OssSyncRunInp 分批同步入参
type OssSyncRunInp struct {
	TargetDriver    string `json:"targetDriver" v:"required#请指定目标云存储" dc:"qiniu|aliyun-oss|tencent-cos"`
	PageSize        int    `json:"pageSize" d:"20" dc:"每批数量"`
	DeleteLocal     bool   `json:"deleteLocal" dc:"同步成功后删除本地文件"`
	IncludeReupload bool   `json:"includeReupload" dc:"含已同步记录：本地文件仍在时重新上传到云"`
	LastLocalId     uint64 `json:"lastLocalId" d:"0" dc:"本地待迁移游标"`
	LastReuploadId  uint64 `json:"lastReuploadId" d:"0" dc:"重新上传游标"`
	DryRun          bool   `json:"dryRun" dc:"仅预览不上传"`
}

// OssSyncFailureItem 单条失败
type OssSyncFailureItem struct {
	Id    uint64 `json:"id"`
	Name  string `json:"name"`
	Url   string `json:"url"`
	Error string `json:"error"`
}

// OssSyncRunModel 分批同步结果
type OssSyncRunModel struct {
	Total          int                  `json:"total" dc:"待处理总数"`
	Processed      int                  `json:"processed" dc:"本批处理数"`
	Success        int                  `json:"success" dc:"本批成功数"`
	Failed         int                  `json:"failed" dc:"本批失败数"`
	Skipped        int                  `json:"skipped" dc:"本批跳过数"`
	Done           bool                 `json:"done" dc:"是否已全部处理"`
	LastLocalId    uint64               `json:"lastLocalId" dc:"本地待迁移游标"`
	LastReuploadId uint64               `json:"lastReuploadId" dc:"重新上传游标"`
	Failures       []OssSyncFailureItem `json:"failures"`
}

// OssSyncListInp 内部列表用
type OssSyncListInp struct {
	form.PageReq
}
