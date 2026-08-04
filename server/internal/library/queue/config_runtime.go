package queue

import "sync"

// TopicConfig Topic 运行配置（来自数据库或默认值）
type TopicConfig struct {
	Id            uint64
	Topic         string
	Title         string
	Workers       int
	MaxRetry      int
	RetryDelaySec int
	Status        int
	Remark        string
	Sort          int
}

var (
	topicConfigMu  sync.RWMutex
	topicConfigMap = map[string]TopicConfig{}
)

// ApplyTopicConfigs 应用 Topic 运行配置（启动或热更新前调用）
func ApplyTopicConfigs(configs map[string]TopicConfig) {
	topicConfigMu.Lock()
	defer topicConfigMu.Unlock()
	topicConfigMap = configs
}

// GetTopicConfig 获取 Topic 配置（无则返回默认值）
func GetTopicConfig(topic string) TopicConfig {
	topicConfigMu.RLock()
	defer topicConfigMu.RUnlock()
	if cfg, ok := topicConfigMap[topic]; ok {
		return cfg
	}
	return TopicConfig{
		Topic:         topic,
		Workers:       1,
		MaxRetry:      3,
		RetryDelaySec: 0,
		Status:        1,
	}
}

func defaultMaxRetry(topic string) int {
	maxRetry := GetTopicConfig(topic).MaxRetry
	if maxRetry <= 0 {
		return 3
	}
	return maxRetry
}
