// Package queue 消息队列配置加载与 Bootstrap
package queue

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"xygo/internal/dao"
	queueLib "xygo/internal/library/queue"
	"xygo/internal/model/entity"
	"xygo/internal/model/input/adminin"
)

// Bootstrap 同步 Topic 配置并启动消费者
func Bootstrap(ctx context.Context) error {
	if err := SyncRegisteredTopics(ctx); err != nil {
		return err
	}
	configs, err := LoadTopicConfigs(ctx)
	if err != nil {
		return err
	}
	queueLib.ApplyTopicConfigs(configs)
	queueLib.StartConsumers(ctx)
	return nil
}

// Reload 重新加载配置并重启消费者
func Reload(ctx context.Context) error {
	if err := SyncRegisteredTopics(ctx); err != nil {
		return err
	}
	configs, err := LoadTopicConfigs(ctx)
	if err != nil {
		return err
	}
	queueLib.ApplyTopicConfigs(configs)
	return queueLib.ReloadConsumers(ctx)
}

// SyncRegisteredTopics 为代码注册的 Topic 补齐数据库配置
func SyncRegisteredTopics(ctx context.Context) error {
	now := uint(time.Now().Unix())
	for i, topic := range queueLib.GetRegisteredTopics() {
		count, err := dao.SysQueueConfig.Ctx(ctx).Where("topic", topic).Count()
		if err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		_, err = dao.SysQueueConfig.Ctx(ctx).Data(g.Map{
			"topic":           topic,
			"title":           topic,
			"workers":         1,
			"max_retry":       3,
			"retry_delay_sec": 0,
			"status":          1,
			"sort":            i + 1,
			"created_at":      now,
			"updated_at":      now,
		}).Insert()
		if err != nil {
			return err
		}
	}
	return nil
}

// LoadTopicConfigs 从数据库加载全部 Topic 配置
func LoadTopicConfigs(ctx context.Context) (map[string]queueLib.TopicConfig, error) {
	var rows []entity.SysQueueConfig
	if err := dao.SysQueueConfig.Ctx(ctx).OrderAsc("sort").OrderAsc("id").Scan(&rows); err != nil {
		return nil, err
	}
	out := make(map[string]queueLib.TopicConfig, len(rows))
	for _, row := range rows {
		out[row.Topic] = entityToTopicConfig(row)
	}
	return out, nil
}

// SaveConfig 保存 Topic 配置并热更新消费者
func SaveConfig(ctx context.Context, in *adminin.QueueConfigSaveInp) error {
	registered := queueLib.GetRegisteredTopics()
	found := false
	for _, t := range registered {
		if t == in.Topic {
			found = true
			break
		}
	}
	if !found {
		return gerror.Newf("Topic '%s' 未在代码中注册", in.Topic)
	}

	now := uint(time.Now().Unix())
	data := g.Map{
		"topic":           in.Topic,
		"title":           in.Title,
		"workers":         in.Workers,
		"max_retry":       in.MaxRetry,
		"retry_delay_sec": in.RetryDelaySec,
		"status":          in.Status,
		"remark":          in.Remark,
		"sort":            in.Sort,
		"updated_at":      now,
	}

	if in.Id > 0 {
		_, err := dao.SysQueueConfig.Ctx(ctx).Where("id", in.Id).Data(data).Update()
		if err != nil {
			return err
		}
	} else {
		var existing entity.SysQueueConfig
		_ = dao.SysQueueConfig.Ctx(ctx).Where("topic", in.Topic).Scan(&existing)
		if existing.Id > 0 {
			_, err := dao.SysQueueConfig.Ctx(ctx).Where("id", existing.Id).Data(data).Update()
			if err != nil {
				return err
			}
		} else {
			data["created_at"] = now
			_, err := dao.SysQueueConfig.Ctx(ctx).Data(data).Insert()
			if err != nil {
				return err
			}
		}
	}
	return Reload(ctx)
}

func entityToTopicConfig(row entity.SysQueueConfig) queueLib.TopicConfig {
	workers := row.Workers
	if workers < 1 {
		workers = 1
	}
	return queueLib.TopicConfig{
		Id:            row.Id,
		Topic:         row.Topic,
		Title:         row.Title,
		Workers:       workers,
		MaxRetry:      row.MaxRetry,
		RetryDelaySec: row.RetryDelaySec,
		Status:        row.Status,
		Remark:        row.Remark,
		Sort:          row.Sort,
	}
}
