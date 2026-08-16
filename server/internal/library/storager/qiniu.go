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

	"github.com/gogf/gf/v2/os/gtime"
	"github.com/google/uuid"
	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
)

// QiniuConfig 七牛云配置
type QiniuConfig struct {
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
	Bucket    string `json:"bucket"`
	Domain    string `json:"domain"` // CDN 域名，如 https://cdn.example.com
	Prefix    string `json:"prefix"` // 对象前缀
	Zone      string `json:"zone"`   // 存储区域: huadong / huabei / huanan / beimei / xinjiapo
}

// Qiniu 七牛云存储驱动
type Qiniu struct {
	config QiniuConfig
	mac    *qbox.Mac
}

// NewQiniu 创建七牛云驱动
func NewQiniu(cfg QiniuConfig) (*Qiniu, error) {
	if cfg.AccessKey == "" || cfg.SecretKey == "" || cfg.Bucket == "" || cfg.Domain == "" {
		return nil, fmt.Errorf("qiniu config incomplete: accessKey/secretKey/bucket/domain required")
	}
	mac := qbox.NewMac(cfg.AccessKey, cfg.SecretKey)
	return &Qiniu{config: cfg, mac: mac}, nil
}

func (q *Qiniu) DriverName() string {
	return "qiniu"
}

// getZone 根据配置获取存储区域
func (q *Qiniu) getZone() *storage.Zone {
	switch strings.ToLower(q.config.Zone) {
	case "huadong", "z0":
		return &storage.ZoneHuadong
	case "huabei", "z1":
		return &storage.ZoneHuabei
	case "huanan", "z2":
		return &storage.ZoneHuanan
	case "beimei", "na0":
		return &storage.ZoneBeimei
	case "xinjiapo", "as0":
		return &storage.ZoneXinjiapo
	default:
		return &storage.ZoneHuadong // 默认华东
	}
}

// Upload 上传文件到七牛云
func (q *Qiniu) Upload(_ context.Context, file *UploadFile) (*UploadResult, error) {
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
		prefix := strings.TrimRight(q.config.Prefix, "/")
		objectKey = prefix + "/" + subdir + "/" + name
	}

	// 生成上传凭证
	putPolicy := storage.PutPolicy{
		Scope: q.config.Bucket,
	}
	upToken := putPolicy.UploadToken(q.mac)

	// 配置上传
	cfg := storage.Config{
		Zone:          q.getZone(),
		UseHTTPS:      true,
		UseCdnDomains: true,
	}
	formUploader := storage.NewFormUploader(&cfg)
	ret := storage.PutRet{}

	reader := bytes.NewReader(file.Data)
	dataLen := int64(len(file.Data))

	err := formUploader.Put(context.Background(), &ret, upToken, objectKey, reader, dataLen, &storage.PutExtra{
		MimeType: file.MimeType,
	})
	if err != nil {
		return nil, fmt.Errorf("qiniu upload error: %w", err)
	}

	return &UploadResult{
		RelPath: ret.Key,
		FullUrl: q.GetFullUrl(ret.Key),
	}, nil
}

// Delete 删除七牛云上的文件
func (q *Qiniu) Delete(_ context.Context, relPath string) error {
	cfg := storage.Config{
		Zone:     q.getZone(),
		UseHTTPS: true,
	}
	bucketManager := storage.NewBucketManager(q.mac, &cfg)
	err := bucketManager.Delete(q.config.Bucket, relPath)
	if err != nil {
		return fmt.Errorf("qiniu delete error: %w", err)
	}
	return nil
}

// GetFullUrl 获取完整 URL
func (q *Qiniu) GetFullUrl(relPath string) string {
	domain := strings.TrimRight(q.config.Domain, "/")
	return domain + "/" + relPath
}
