package adminin

// ===================== 文件上传 =====================

// UploadFileInp 文件上传入参
type UploadFileInp struct {
	Drive string `p:"drive" json:"drive" dc:"可选，覆盖配置中的默认驱动"`
	Topic string `p:"topic" json:"topic" dc:"可选，附件分组/主题标识"`
	// 文件字段 name= file，由框架从 multipart 读取
}

// UploadFileModel 上传结果
type UploadFileModel struct {
	URL          string `json:"url" dc:"浏览器可访问地址（本地相对路径 / 云存储 CDN 完整 URL）"`
	Path         string `json:"path" dc:"存储路径 object key（写入附件表 / CMS 用）"`
	Size         int64  `json:"size" dc:"字节"`
	Mime         string `json:"mime" dc:"MIME"`
	Ext          string `json:"ext" dc:"扩展名（含 .）"`
	Drive        string `json:"drive" dc:"实际使用的驱动"`
	AttachmentId uint64 `json:"attachmentId" dc:"附件记录ID"`
}

// ResolveMediaUrlInp 将库中 path 解析为可访问 URL（预览用）
type ResolveMediaUrlInp struct {
	Path  string `p:"path" json:"path" dc:"单个 path"`
	Paths string `p:"paths" json:"paths" dc:"逗号分隔多个 path"`
}

// ResolveMediaUrlModel 解析结果
type ResolveMediaUrlModel struct {
	URL  string            `json:"url,omitempty" dc:"单个 path 的可访问 URL"`
	List map[string]string `json:"list,omitempty" dc:"path → 可访问 URL"`
}
