package adminin

// QueueConfigSaveInp 保存队列 Topic 配置
type QueueConfigSaveInp struct {
	Id            uint64 `p:"id"            json:"id"            dc:"配置ID"`
	Topic         string `p:"topic"         json:"topic"         v:"required#Topic不能为空" dc:"Topic 标识"`
	Title         string `p:"title"         json:"title"         dc:"显示名称"`
	Workers       int    `p:"workers"       json:"workers"       v:"required|min:1|max:64#Worker数必填|Worker至少为1|Worker最多64" dc:"并行 Worker 数"`
	MaxRetry      int    `p:"maxRetry"      json:"maxRetry"      v:"min:0|max:20#重试次数0-20" d:"3" dc:"最大重试次数"`
	RetryDelaySec int    `p:"retryDelaySec" json:"retryDelaySec" v:"min:0|max:86400#重试间隔0-86400秒" d:"0" dc:"重试间隔秒"`
	Status        int    `p:"status"        json:"status"        v:"in:0,1#状态无效" d:"1" dc:"状态:0禁用,1启用"`
	Remark        string `p:"remark"        json:"remark"        dc:"备注"`
	Sort          int    `p:"sort"          json:"sort"          d:"0" dc:"排序"`
}
