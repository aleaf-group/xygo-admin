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
	cronlogic "xygo/internal/logic/cron"
	"xygo/internal/model/input/form"
)

// ==================== 定时任务 ====================

func (c *ControllerV1) CronList(ctx context.Context, req *api.CronListReq) (res *api.CronListRes, err error) {
	result, err := cronlogic.List(ctx, &req.CronListInp)
	if err != nil {
		return nil, err
	}
	return &api.CronListRes{CronListModel: result}, nil
}

func (c *ControllerV1) CronSave(ctx context.Context, req *api.CronSaveReq) (res *api.CronSaveRes, err error) {
	id, err := cronlogic.Save(ctx, &req.CronSaveInp)
	if err != nil {
		return nil, err
	}
	return &api.CronSaveRes{Id: id}, nil
}

func (c *ControllerV1) CronDelete(ctx context.Context, req *api.CronDeleteReq) (res *api.CronDeleteRes, err error) {
	err = cronlogic.Delete(ctx, req.Id)
	return &api.CronDeleteRes{}, err
}

func (c *ControllerV1) CronStatus(ctx context.Context, req *api.CronStatusReq) (res *api.CronStatusRes, err error) {
	err = cronlogic.UpdateStatus(ctx, &req.CronStatusInp)
	return &api.CronStatusRes{}, err
}

func (c *ControllerV1) CronOnlineExec(ctx context.Context, req *api.CronOnlineExecReq) (res *api.CronOnlineExecRes, err error) {
	output, err := cronlogic.OnlineExec(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &api.CronOnlineExecRes{Output: output}, nil
}

func (c *ControllerV1) CronRegisteredTasks(ctx context.Context, req *api.CronRegisteredTasksReq) (res *api.CronRegisteredTasksRes, err error) {
	return &api.CronRegisteredTasksRes{List: cronlogic.RegisteredTasks()}, nil
}

// ==================== 分组 ====================

func (c *ControllerV1) CronGroupList(ctx context.Context, req *api.CronGroupListReq) (res *api.CronGroupListRes, err error) {
	result, err := cronlogic.GroupList(ctx, &req.CronGroupListInp)
	if err != nil {
		return nil, err
	}
	return &api.CronGroupListRes{CronGroupListModel: result}, nil
}

func (c *ControllerV1) CronGroupSave(ctx context.Context, req *api.CronGroupSaveReq) (res *api.CronGroupSaveRes, err error) {
	id, err := cronlogic.GroupSave(ctx, &req.CronGroupSaveInp)
	if err != nil {
		return nil, err
	}
	return &api.CronGroupSaveRes{Id: id}, nil
}

func (c *ControllerV1) CronGroupDelete(ctx context.Context, req *api.CronGroupDeleteReq) (res *api.CronGroupDeleteRes, err error) {
	err = cronlogic.GroupDelete(ctx, req.Id)
	return &api.CronGroupDeleteRes{}, err
}

func (c *ControllerV1) CronGroupSelect(ctx context.Context, req *api.CronGroupSelectReq) (res *api.CronGroupSelectRes, err error) {
	list, err := cronlogic.GroupSelect(ctx)
	if err != nil {
		return nil, err
	}
	return &api.CronGroupSelectRes{List: list}, nil
}

// ==================== 执行日志 ====================

func (c *ControllerV1) CronLogList(ctx context.Context, req *api.CronLogListReq) (res *api.CronLogListRes, err error) {
	result, err := cronlogic.LogList(ctx, &req.CronLogListInp)
	if err != nil {
		return nil, err
	}
	res = &api.CronLogListRes{CronLogListModel: result}
	// 补全分页信息
	if result != nil {
		result.PageRes = form.PageRes{
			Page:     req.Page,
			PageSize: req.PageSize,
			Total:    result.Total,
		}
	}
	return
}

func (c *ControllerV1) CronLogClear(ctx context.Context, req *api.CronLogClearReq) (res *api.CronLogClearRes, err error) {
	err = cronlogic.LogClear(ctx, &req.CronLogClearInp)
	return &api.CronLogClearRes{}, err
}
