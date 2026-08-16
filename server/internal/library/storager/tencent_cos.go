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
	"net/http"
	"net/url"
	"strings"

	"github.com/gogf/gf/v2/os/gtime"
	"github.com/google/uuid"
	"github.com/tencentyun/cos-go-sdk-v5"
)

// TencentCOS 腾讯云 COS 存储驱动
type TencentCOS struct {
	config TencentConfig
	client *cos.Client
}

// NewTencentCOS 创建腾讯云 COS 驱动
func NewTencentCOS(cfg TencentConfig) (*TencentCOS, error) {
	if cfg.SecretId == "" || cfg.SecretKey == "" || cfg.Bucket == "" || cfg.Region == "" {
		return nil, fmt.Errorf("tencent COS config incomplete: secretId/secretKey/bucket/region required")
	}

	bucketURL, _ := url.Parse(fmt.Sprintf("https://%s.cos.%s.myqcloud.com", cfg.Bucket, cfg.Region))
	serviceURL, _ := url.Parse(fmt.Sprintf("https://cos.%s.myqcloud.com", cfg.Region))

	client := cos.NewClient(&cos.BaseURL{
		BucketURL:  bucketURL,
		ServiceURL: serviceURL,
	}, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  cfg.SecretId,
			SecretKey: cfg.SecretKey,
		},
	})

	return &TencentCOS{
		config: cfg,
		client: client,
	}, nil
}

func (t *TencentCOS) DriverName() string {
	return "tencent-cos"
}

// Upload 上传文件到腾讯云 COS
func (t *TencentCOS) Upload(ctx context.Context, file *UploadFile) (*UploadResult, error) {
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
		prefix := strings.TrimRight(t.config.Prefix, "/")
		objectKey = prefix + "/" + subdir + "/" + name
	}

	// 上传
	reader := bytes.NewReader(file.Data)
	_, err := t.client.Object.Put(ctx, objectKey, reader, &cos.ObjectPutOptions{
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
			ContentType: file.MimeType,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("COS upload error: %w", err)
	}

	return &UploadResult{
		RelPath: objectKey,
		FullUrl: t.GetFullUrl(objectKey),
	}, nil
}

// Delete 删除 COS 上的文件
func (t *TencentCOS) Delete(ctx context.Context, relPath string) error {
	_, err := t.client.Object.Delete(ctx, relPath)
	if err != nil {
		return fmt.Errorf("COS delete error: %w", err)
	}
	return nil
}

// GetFullUrl 获取完整 URL
func (t *TencentCOS) GetFullUrl(relPath string) string {
	if t.config.Domain != "" {
		domain := strings.TrimRight(t.config.Domain, "/")
		return domain + "/" + relPath
	}
	return fmt.Sprintf("https://%s.cos.%s.myqcloud.com/%s", t.config.Bucket, t.config.Region, relPath)
}
