// +----------------------------------------------------------------------
// | XYGo Admin [ Vue3 + GoFrame 企业级中后台管理系统 ]
// +----------------------------------------------------------------------
// | Copyright (c) 2026 大连星韵网络科技有限公司 All rights reserved.
// +----------------------------------------------------------------------
// | Licensed ( https://opensource.org/licenses/MIT )
// +----------------------------------------------------------------------
// | Author: 喜羊羊 <751300685@qq.com>
// +----------------------------------------------------------------------

package storager

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/google/uuid"
)

// AliyunOSS 阿里云 OSS 存储驱动
type AliyunOSS struct {
	config AliyunConfig
	client *oss.Client
	bucket *oss.Bucket
}

// NewAliyunOSS 创建阿里云 OSS 驱动
func NewAliyunOSS(cfg AliyunConfig) (*AliyunOSS, error) {
	if cfg.Endpoint == "" || cfg.AccessKeyId == "" || cfg.AccessKeySecret == "" || cfg.Bucket == "" {
		return nil, fmt.Errorf("aliyun OSS config incomplete: endpoint/accessKeyId/accessKeySecret/bucket required")
	}

	client, err := oss.New(cfg.Endpoint, cfg.AccessKeyId, cfg.AccessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("create OSS client error: %w", err)
	}

	bucket, err := client.Bucket(cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("get OSS bucket error: %w", err)
	}

	return &AliyunOSS{
		config: cfg,
		client: client,
		bucket: bucket,
	}, nil
}

func (a *AliyunOSS) DriverName() string {
	return "aliyun-oss"
}

// Upload 上传文件到阿里云 OSS
func (a *AliyunOSS) Upload(_ context.Context, file *UploadFile) (*UploadResult, error) {
	var objectKey string
	if k := strings.TrimLeft(strings.TrimSpace(file.ObjectKey), "/"); k != "" {
		objectKey = k
	} else {
		ext := file.Ext
		if ext == "" {
			ext = "bin"
		}
		subdir := gtime.Now().Format("Ymd")
		name := uuid.New().String() + "." + ext
		prefix := strings.TrimRight(a.config.Prefix, "/")
		objectKey = prefix + "/" + subdir + "/" + name
	}

	// 上传
	reader := bytes.NewReader(file.Data)
	err := a.bucket.PutObject(objectKey, reader, oss.ContentType(file.MimeType))
	if err != nil {
		return nil, fmt.Errorf("OSS upload error: %w", err)
	}

	return &UploadResult{
		RelPath: objectKey,
		FullUrl: a.GetFullUrl(objectKey),
	}, nil
}

// Delete 删除 OSS 上的文件
func (a *AliyunOSS) Delete(_ context.Context, relPath string) error {
	err := a.bucket.DeleteObject(relPath)
	if err != nil {
		return fmt.Errorf("OSS delete error: %w", err)
	}
	return nil
}

// GetFullUrl 获取完整 URL
func (a *AliyunOSS) GetFullUrl(relPath string) string {
	if a.config.Domain != "" {
		domain := strings.TrimRight(a.config.Domain, "/")
		return domain + "/" + relPath
	}
	// 默认使用 bucket.endpoint 的 URL
	endpoint := strings.TrimPrefix(a.config.Endpoint, "https://")
	endpoint = strings.TrimPrefix(endpoint, "http://")
	return fmt.Sprintf("https://%s.%s/%s", a.config.Bucket, endpoint, relPath)
}
