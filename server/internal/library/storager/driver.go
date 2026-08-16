// +----------------------------------------------------------------------
// | XYGo Admin [ Vue3 + GoFrame 企业级中后台管理系统 ]
// +----------------------------------------------------------------------
// | Copyright (c) 2026 大连星韵网络科技有限公司 All rights reserved.
// +----------------------------------------------------------------------
// | Licensed ( https://opensource.org/licenses/MIT )
// +----------------------------------------------------------------------
// | Author: 喜羊羊 <751300685@qq.com>
// +----------------------------------------------------------------------

// Package storager 文件存储驱动抽象层
// 支持 local（本地）、aliyun-oss（阿里云）、tencent-cos（腾讯云）
package storager

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/gogf/gf/v2/frame/g"

	"xygo/internal/service"
)

// UploadFile 上传文件信息
type UploadFile struct {
	Data      []byte // 文件内容
	Filename  string // 原始文件名
	Ext       string // 扩展名（不含点，小写）
	MimeType  string // MIME 类型
	Size      int64  // 文件大小
	ObjectKey string // 指定 object key（迁移历史文件时使用，非空则跳过 UUID 生成）
}

// UploadResult 上传结果
type UploadResult struct {
	RelPath string // 相对路径（存入数据库）
	FullUrl string // 完整访问 URL
}

// Storager 存储驱动接口
type Storager interface {
	// Upload 上传文件，返回相对路径和完整 URL
	Upload(ctx context.Context, file *UploadFile) (*UploadResult, error)
	// Delete 删除文件（传入相对路径）
	Delete(ctx context.Context, relPath string) error
	// GetFullUrl 根据相对路径获取完整访问 URL
	GetFullUrl(relPath string) string
	// DriverName 返回驱动名称
	DriverName() string
}

// StorageConfig 存储配置（从 config.yaml 读取）
type StorageConfig struct {
	Driver      string        `json:"driver"`
	MaxSize     string        `json:"maxSize"`
	AllowedExts string        `json:"allowedExts"`
	Local       LocalConfig   `json:"local"`
	Aliyun      AliyunConfig  `json:"aliyun"`
	Tencent     TencentConfig `json:"tencent"`
	Qiniu       QiniuConfig   `json:"qiniu"`
}

// LocalConfig 本地存储配置
type LocalConfig struct {
	BasePath  string `json:"basePath"`
	UrlPrefix string `json:"urlPrefix"`
}

// AliyunConfig 阿里云 OSS 配置
type AliyunConfig struct {
	Endpoint        string `json:"endpoint"`
	AccessKeyId     string `json:"accessKeyId"`
	AccessKeySecret string `json:"accessKeySecret"`
	Bucket          string `json:"bucket"`
	Prefix          string `json:"prefix"`
	Domain          string `json:"domain"` // 自定义域名（可选）
}

// TencentConfig 腾讯云 COS 配置
type TencentConfig struct {
	SecretId  string `json:"secretId"`
	SecretKey string `json:"secretKey"`
	Bucket    string `json:"bucket"`
	Region    string `json:"region"`
	Prefix    string `json:"prefix"`
	Domain    string `json:"domain"` // 自定义域名（可选）
}

// 全局存储驱动实例
var (
	instance Storager
	once     sync.Once
	config   *StorageConfig
)

// Instance 获取全局存储驱动实例（单例）
func Instance(ctx ...context.Context) Storager {
	once.Do(func() {
		c := context.Background()
		if len(ctx) > 0 && ctx[0] != nil {
			c = ctx[0]
		}
		cfg := GetConfig(c)
		var err error
		instance, err = NewDriver(cfg)
		if err != nil {
			g.Log().Errorf(c, "[Storager] init driver error: %v, fallback to local", err)
			instance = NewLocal(cfg.Local)
		}
		g.Log().Infof(c, "[Storager] initialized with driver: %s", instance.DriverName())
	})
	return instance
}

// ResetInstance 重置单例（用于切换驱动时）
func ResetInstance() {
	once = sync.Once{}
	instance = nil
	config = nil
}

// OnConfigChanged 当 oss 分组配置保存后调用，重置驱动单例与配置缓存
func OnConfigChanged(group string) {
	if strings.EqualFold(group, "oss") {
		ResetInstance()
	}
}

// GetConfig 从数据库 sys_config 表（oss 分组）读取存储配置
// 复用已有的 service.SysConfig().GetConfigByGroup 通用方法
func GetConfig(ctx context.Context) *StorageConfig {
	if config != nil {
		return config
	}

	config = defaultConfig()

	// 通过通用配置服务读取 oss 分组
	cfg, err := service.SysConfig().GetConfigByGroup(ctx, "oss")
	if err != nil {
		g.Log().Warningf(ctx, "[Storager] read oss config error: %v, using defaults", err)
		return config
	}

	// 解析驱动类型（兼容中英文值）
	if v := cfg["oss_driver"]; v != "" {
		config.Driver = normalizeDriverName(v)
	}
	// 兼容 upload_driver key
	if config.Driver == "" || config.Driver == "local" {
		if v := cfg["upload_driver"]; v != "" {
			config.Driver = normalizeDriverName(v)
		}
	}
	// 通用配置
	if v := cfg["upload_max_size"]; v != "" {
		config.MaxSize = v
	}
	if v := cfg["upload_allowed_suffixes"]; v != "" {
		config.AllowedExts = v
	}

	// 本地配置（从 config.yaml 补充，因为路径是文件系统相关的）
	yamlLocal, _ := g.Cfg().Get(ctx, "storage.local")
	if !yamlLocal.IsNil() {
		_ = yamlLocal.Scan(&config.Local)
	}
	if config.Local.BasePath == "" {
		config.Local.BasePath = "resource/public/attachment/upload"
	}
	if config.Local.UrlPrefix == "" {
		config.Local.UrlPrefix = "/attachment/upload"
	}

	// 阿里云 OSS
	config.Aliyun.Endpoint = cfg["oss_aliyun_endpoint"]
	config.Aliyun.AccessKeyId = cfg["oss_aliyun_access_key_id"]
	config.Aliyun.AccessKeySecret = cfg["oss_aliyun_access_key_secret"]
	config.Aliyun.Bucket = cfg["oss_aliyun_bucket"]
	config.Aliyun.Domain = cfg["oss_aliyun_domain"]
	config.Aliyun.Prefix = cfg["oss_aliyun_root"]
	if config.Aliyun.Prefix == "" {
		config.Aliyun.Prefix = "upload/"
	}

	// 腾讯云 COS
	config.Tencent.SecretId = cfg["oss_cos_secret_id"]
	config.Tencent.SecretKey = cfg["oss_cos_secret_key"]
	config.Tencent.Bucket = cfg["oss_cos_bucket"]
	config.Tencent.Region = cfg["oss_cos_region"]
	config.Tencent.Domain = cfg["oss_cos_domain"]
	config.Tencent.Prefix = cfg["oss_cos_root"]
	if config.Tencent.Prefix == "" {
		config.Tencent.Prefix = "upload/"
	}

	// 七牛云
	config.Qiniu.AccessKey = cfg["oss_qiniu_access_key"]
	config.Qiniu.SecretKey = cfg["oss_qiniu_secret_key"]
	config.Qiniu.Bucket = cfg["oss_qiniu_bucket"]
	config.Qiniu.Domain = cfg["oss_qiniu_domain"]
	config.Qiniu.Prefix = cfg["oss_qiniu_root"]
	config.Qiniu.Zone = cfg["oss_qiniu_zone"]
	if config.Qiniu.Prefix == "" {
		config.Qiniu.Prefix = "upload/"
	}

	g.Log().Infof(ctx, "[Storager] loaded config from database, driver: %s", config.Driver)
	return config
}

func defaultConfig() *StorageConfig {
	return &StorageConfig{
		Driver:      "local",
		MaxSize:     "10mb",
		AllowedExts: "jpg,jpeg,png,gif,webp,mp4,pdf,doc,docx,xls,xlsx,zip,rar",
		Local: LocalConfig{
			BasePath:  "resource/public/attachment/upload",
			UrlPrefix: "/attachment/upload",
		},
	}
}

// normalizeDriverName 统一驱动名称
// 数据库 select options: local / oss / cos / qiniu
func normalizeDriverName(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	switch v {
	case "local", "":
		return "local"
	case "oss", "aliyun-oss", "aliyun":
		return "aliyun-oss"
	case "cos", "tencent-cos", "tencent":
		return "tencent-cos"
	case "qiniu", "qn":
		return "qiniu"
	default:
		return v
	}
}

// NewDriver 根据配置创建存储驱动
func NewDriver(cfg *StorageConfig) (Storager, error) {
	switch strings.ToLower(cfg.Driver) {
	case "local", "":
		return NewLocal(cfg.Local), nil
	case "aliyun-oss", "aliyun", "oss":
		return NewAliyunOSS(cfg.Aliyun)
	case "tencent-cos", "tencent", "cos":
		return NewTencentCOS(cfg.Tencent)
	case "qiniu", "qn":
		return NewQiniu(cfg.Qiniu)
	default:
		return nil, fmt.Errorf("unsupported storage driver: %s", cfg.Driver)
	}
}
