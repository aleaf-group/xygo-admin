package entity

// SysQueueConfig is the golang structure for table xy_sys_queue_config.
type SysQueueConfig struct {
	Id            uint64 `json:"id"            orm:"id"              description:"ID"`
	Topic         string `json:"topic"         orm:"topic"           description:"Topic 标识"`
	Title         string `json:"title"         orm:"title"           description:"显示名称"`
	Workers       int    `json:"workers"       orm:"workers"         description:"并行 Worker 数"`
	MaxRetry      int    `json:"maxRetry"      orm:"max_retry"       description:"最大重试次数"`
	RetryDelaySec int    `json:"retryDelaySec" orm:"retry_delay_sec" description:"重试间隔秒"`
	Status        int    `json:"status"        orm:"status"          description:"状态:0禁用,1启用"`
	Remark        string `json:"remark"        orm:"remark"          description:"备注"`
	Sort          int    `json:"sort"          orm:"sort"            description:"排序"`
	CreatedAt     uint   `json:"createdAt"     orm:"created_at"      description:"创建时间"`
	UpdatedAt     uint   `json:"updatedAt"     orm:"updated_at"      description:"更新时间"`
}
