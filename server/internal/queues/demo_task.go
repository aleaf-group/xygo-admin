// +----------------------------------------------------------------------
// | XYGo Admin [ Vue3 + GoFrame 企业级中后台管理系统 ]
// +----------------------------------------------------------------------
// | Copyright (c) 2026 大连星韵网络科技有限公司 All rights reserved.
// +----------------------------------------------------------------------
// | Licensed ( https://opensource.org/licenses/MIT )
// +----------------------------------------------------------------------
// | Author: 喜羊羊 <751300685@qq.com>
// +----------------------------------------------------------------------

package queues

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/frame/g"

	"xygo/internal/library/queue"
)

const TopicDemoTask = "demo_task"

func init() {
	queue.Register(&DemoTaskConsumer{})
}

// DemoTaskConsumer 演示队列消费者（示例：新增 Topic 只需实现 Consumer 并 Register）
type DemoTaskConsumer struct{}

func (c *DemoTaskConsumer) GetTopic() string { return TopicDemoTask }

func (c *DemoTaskConsumer) Handle(ctx context.Context, msg *queue.Message) error {
	var data struct {
		Message string `json:"message"`
		Fail    bool   `json:"fail"` // 测试投递时设为 true 可触发 RetryError 重试
	}
	if err := json.Unmarshal([]byte(msg.Body), &data); err != nil {
		g.Log().Errorf(ctx, "[queue:demo_task] unmarshal failed: %v", err)
		return nil
	}
	if data.Fail {
		return queue.NewRetryError(fmt.Errorf("demo_task simulated failure"))
	}

	text := data.Message
	if text == "" {
		text = "empty message"
	}
	g.Log().Infof(ctx, "[queue:demo_task] consumed at %s: %s (retry=%d)",
		time.Now().Format("2006-01-02 15:04:05"), text, msg.Retry)
	return nil
}
