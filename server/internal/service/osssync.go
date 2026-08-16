// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"
	"xygo/internal/model/input/adminin"
)

type (
	IOssSync interface {
		Preview(ctx context.Context, in *adminin.OssSyncPreviewInp) (*adminin.OssSyncPreviewModel, error)
		Run(ctx context.Context, in *adminin.OssSyncRunInp) (*adminin.OssSyncRunModel, error)
	}
)

var (
	localOssSync IOssSync
)

func OssSync() IOssSync {
	if localOssSync == nil {
		panic("implement not found for interface IOssSync, forgot register?")
	}
	return localOssSync
}

func RegisterOssSync(i IOssSync) {
	localOssSync = i
}
