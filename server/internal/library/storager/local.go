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
	"context"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/google/uuid"
)

// Local 本地文件存储驱动
type Local struct {
	config LocalConfig
}

// NewLocal 创建本地存储驱动
func NewLocal(cfg LocalConfig) *Local {
	if cfg.BasePath == "" {
		cfg.BasePath = "resource/public/attachment/upload"
	}
	if cfg.UrlPrefix == "" {
		cfg.UrlPrefix = "/attachment/upload"
	}
	return &Local{config: cfg}
}

func (l *Local) DriverName() string {
	return "local"
}

// Upload 上传文件到本地磁盘
func (l *Local) Upload(_ context.Context, file *UploadFile) (*UploadResult, error) {
	ext := file.Ext
	if ext == "" {
		ext = "bin"
	}

	var relPath string
	var fullPath string
	if k := strings.TrimLeft(strings.TrimSpace(file.ObjectKey), "/"); k != "" {
		relPath = "/" + k
		fullPath = LocalPhysicalPath(relPath)
	} else {
		subdir := gtime.Now().Format("Ymd")
		name := uuid.New().String() + "." + ext
		relPath = strings.TrimRight(l.config.UrlPrefix, "/") + "/" + subdir + "/" + name
		relPath = strings.ReplaceAll(relPath, `\`, `/`)
		if !strings.HasPrefix(relPath, "/") {
			relPath = "/" + relPath
		}
		fullPath = gfile.Join(l.config.BasePath, subdir, name)
	}
	if fullPath == "" {
		return nil, fmt.Errorf("invalid object key: %s", file.ObjectKey)
	}
	if err := gfile.Mkdir(gfile.Dir(fullPath)); err != nil {
		return nil, err
	}
	if err := gfile.PutBytes(fullPath, file.Data); err != nil {
		return nil, err
	}

	return &UploadResult{
		RelPath: relPath,
		FullUrl: relPath, // 本地模式 URL 就是相对路径，由静态文件服务提供
	}, nil
}

// Delete 删除本地文件
func (l *Local) Delete(_ context.Context, relPath string) error {
	// relPath 如 /attachment/upload/20260210/xxx.jpg
	// 转换为物理路径：resource/public + relPath
	physicalPath := gfile.Join("resource/public", strings.TrimLeft(relPath, "/"))
	if gfile.Exists(physicalPath) {
		return gfile.Remove(physicalPath)
	}
	return nil
}

// GetFullUrl 本地模式直接返回相对路径
func (l *Local) GetFullUrl(relPath string) string {
	return relPath
}
