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
	"context"

	api "xygo/api/admin"
	queuelogic "xygo/internal/logic/queue"
	"xygo/internal/library/queue"
)

// QueueStats 队列统计
func (c *ControllerV1) QueueStats(ctx context.Context, req *api.QueueStatsReq) (res *api.QueueStatsRes, err error) {
	return &api.QueueStatsRes{
		Driver: queue.GetDriverName(),
		Topics: queue.GetStats(ctx),
	}, nil
}

// QueueTopics 已注册 Topic 列表
func (c *ControllerV1) QueueTopics(ctx context.Context, req *api.QueueTopicsReq) (res *api.QueueTopicsRes, err error) {
	return &api.QueueTopicsRes{
		List: queue.GetRegisteredTopics(),
	}, nil
}

// QueuePushTest 测试投递消息（支持延迟）
func (c *ControllerV1) QueuePushTest(ctx context.Context, req *api.QueuePushTestReq) (res *api.QueuePushTestRes, err error) {
	if req.DelaySec > 0 {
		err = queue.DelayPush(req.Topic, req.Body, req.DelaySec)
	} else {
		err = queue.Push(req.Topic, req.Body)
	}
	return &api.QueuePushTestRes{}, err
}

// QueueConfigSave 保存 Topic 运行配置并热更新 worker
func (c *ControllerV1) QueueConfigSave(ctx context.Context, req *api.QueueConfigSaveReq) (res *api.QueueConfigSaveRes, err error) {
	err = queuelogic.SaveConfig(ctx, &req.QueueConfigSaveInp)
	return &api.QueueConfigSaveRes{}, err
}
