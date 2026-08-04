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

	"xygo/internal/library/queue"
	"xygo/internal/model/input/adminin"
)

// QueueStatsReq 队列统计
type QueueStatsReq struct {
	g.Meta `path:"/admin/queue/stats" method:"get" tags:"Queue" summary:"队列统计"`
}

type QueueStatsRes struct {
	Driver string             `json:"driver"`
	Topics []queue.TopicStats `json:"topics"`
}

// QueueTopicsReq 已注册的 Topic 列表
type QueueTopicsReq struct {
	g.Meta `path:"/admin/queue/topics" method:"get" tags:"Queue" summary:"已注册Topic列表"`
}

type QueueTopicsRes struct {
	List []string `json:"list"`
}

// QueuePushTestReq 测试投递消息
type QueuePushTestReq struct {
	g.Meta   `path:"/admin/queue/pushTest" method:"post" tags:"Queue" summary:"测试投递消息"`
	Topic    string `json:"topic"    p:"topic"    v:"required#Topic不能为空" dc:"队列Topic"`
	Body     string `json:"body"     p:"body"     v:"required#消息内容不能为空" dc:"消息内容(JSON)"`
	DelaySec int64  `json:"delaySec" p:"delaySec" d:"0" dc:"延迟秒数（0=即时投递）"`
}

type QueuePushTestRes struct{}

// QueueConfigSaveReq 保存 Topic 运行配置
type QueueConfigSaveReq struct {
	g.Meta `path:"/admin/queue/configSave" method:"post" tags:"Queue" summary:"保存Topic运行配置"`
	adminin.QueueConfigSaveInp
}

type QueueConfigSaveRes struct{}
