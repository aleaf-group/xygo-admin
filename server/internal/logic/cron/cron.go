// +----------------------------------------------------------------------
// | XYGo Admin [ Vue3 + GoFrame 企业级中后台管理系统 ]
// +----------------------------------------------------------------------
// | Copyright (c) 2026 大连星韵网络科技有限公司 All rights reserved.
// +----------------------------------------------------------------------
// | Licensed ( https://opensource.org/licenses/MIT )
// +----------------------------------------------------------------------
// | Author: 喜羊羊 <751300685@qq.com>
// +----------------------------------------------------------------------

// Package cron 定时任务业务逻辑
package cron

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"xygo/internal/dao"
	cronlib "xygo/internal/library/cron"
	"xygo/internal/model/entity"
	"xygo/internal/model/input/adminin"
	"xygo/internal/model/input/form"
)

// ==================== 定时任务 CRUD ====================

// List 任务列表
func List(ctx context.Context, in *adminin.CronListInp) (*adminin.CronListModel, error) {
	m := dao.SysCron.Ctx(ctx)
	if in.GroupId > 0 {
		m = m.Where("group_id", in.GroupId)
	}
	if in.Status == 0 || in.Status == 1 {
		m = m.Where("status", in.Status)
	}
	if in.Name != "" {
		m = m.WhereLike("name", "%"+in.Name+"%")
	}

	count, err := m.Clone().Count()
	if err != nil {
		return nil, err
	}

	if in.Page <= 0 {
		in.Page = 1
	}
	if in.PageSize <= 0 {
		in.PageSize = 20
	}

	var items []entity.SysCron
	err = m.OrderAsc("sort").OrderDesc("id").Page(in.Page, in.PageSize).Scan(&items)
	if err != nil {
		return nil, err
	}

	// 获取分组名称映射
	groupMap := make(map[uint64]string)
	if len(items) > 0 {
		var groups []entity.SysCronGroup
		_ = dao.SysCronGroup.Ctx(ctx).Scan(&groups)
		for _, g := range groups {
			groupMap[g.Id] = g.Name
		}
	}

	list := make([]adminin.CronListItem, 0, len(items))
	for _, item := range items {
		list = append(list, adminin.CronListItem{
			Id:        item.Id,
			GroupId:   item.GroupId,
			GroupName: groupMap[item.GroupId],
			Title:     item.Title,
			Name:      item.Name,
			Params:    item.Params,
			Pattern:   item.Pattern,
			Policy:    item.Policy,
			Count:     item.Count,
			Sort:      item.Sort,
			Remark:    item.Remark,
			Status:    item.Status,
			CreatedAt: uint(item.CreatedAt),
			UpdatedAt: uint(item.UpdatedAt),
		})
	}

	return &adminin.CronListModel{
		List: list,
		PageRes: form.PageRes{
			Page:     in.Page,
			PageSize: in.PageSize,
			Total:    count,
		},
	}, nil
}

// Save 保存任务
func Save(ctx context.Context, in *adminin.CronSaveInp) (uint64, error) {
	now := uint(time.Now().Unix())

	// 检查 name 唯一
	m := dao.SysCron.Ctx(ctx).Where("name", in.Name)
	if in.Id > 0 {
		m = m.WhereNot("id", in.Id)
	}
	count, err := m.Count()
	if err != nil {
		return 0, err
	}
	if count > 0 {
		return 0, gerror.New("任务标识已存在")
	}

	// 检查任务是否已注册
	if cronlib.GetTask(in.Name) == nil {
		return 0, gerror.Newf("任务标识 '%s' 未在代码中注册，请先在 internal/crons/ 中实现并注册", in.Name)
	}

	data := g.Map{
		"group_id":   in.GroupId,
		"title":      in.Title,
		"name":       in.Name,
		"params":     in.Params,
		"pattern":    in.Pattern,
		"policy":     in.Policy,
		"count":      in.Count,
		"sort":       in.Sort,
		"remark":     in.Remark,
		"status":     in.Status,
		"updated_at": now,
	}

	var id uint64
	if in.Id == 0 {
		data["created_at"] = now
		result, err := dao.SysCron.Ctx(ctx).Data(data).Insert()
		if err != nil {
			return 0, err
		}
		lastId, _ := result.LastInsertId()
		id = uint64(lastId)
	} else {
		_, err := dao.SysCron.Ctx(ctx).Where("id", in.Id).Data(data).Update()
		if err != nil {
			return 0, err
		}
		id = in.Id
	}

	// 同步调度状态
	syncJobSchedule(ctx, id)
	return id, nil
}

// Delete 删除任务
func Delete(ctx context.Context, id uint64) error {
	var item entity.SysCron
	err := dao.SysCron.Ctx(ctx).Where("id", id).Scan(&item)
	if err != nil {
		return err
	}
	if item.Id == 0 {
		return gerror.New("任务不存在")
	}

	// 先停止调度
	cronlib.StopJob(item.Id, item.Name)

	// 删除任务
	_, err = dao.SysCron.Ctx(ctx).Where("id", id).Delete()
	return err
}

// UpdateStatus 更新任务状态
func UpdateStatus(ctx context.Context, in *adminin.CronStatusInp) error {
	_, err := dao.SysCron.Ctx(ctx).Where("id", in.Id).Data(g.Map{
		"status":     in.Status,
		"updated_at": uint(time.Now().Unix()),
	}).Update()
	if err != nil {
		return err
	}
	syncJobSchedule(ctx, in.Id)
	return nil
}

// OnlineExec 在线执行一次
func OnlineExec(ctx context.Context, id uint64) (string, error) {
	var item entity.SysCron
	err := dao.SysCron.Ctx(ctx).Where("id", id).Scan(&item)
	if err != nil {
		return "", err
	}
	if item.Id == 0 {
		return "", gerror.New("任务不存在")
	}

	return cronlib.OnlineExec(&cronlib.CronJob{
		Id:     item.Id,
		Name:   item.Name,
		Title:  item.Title,
		Params: item.Params,
	})
}

// StartAll 启动所有已启用的任务（服务启动时调用）
func StartAll(ctx context.Context) {
	// 注入日志回调
	cronlib.SetLogCallback(recordLog)

	var items []entity.SysCron
	err := dao.SysCron.Ctx(ctx).Where("status", 1).Scan(&items)
	if err != nil {
		g.Log().Errorf(ctx, "[cron] load cron jobs failed: %v", err)
		return
	}

	for _, item := range items {
		job := &cronlib.CronJob{
			Id:      item.Id,
			Name:    item.Name,
			Title:   item.Title,
			Pattern: item.Pattern,
			Params:  item.Params,
			Policy:  item.Policy,
			Count:   item.Count,
		}
		if err := cronlib.StartJob(job); err != nil {
			g.Log().Warningf(ctx, "[cron] start job '%s' failed: %v", item.Name, err)
		}
	}

	g.Log().Infof(ctx, "[cron] %d jobs started", len(items))
}

// syncJobSchedule 同步单个任务的调度状态
func syncJobSchedule(ctx context.Context, id uint64) {
	var item entity.SysCron
	_ = dao.SysCron.Ctx(ctx).Where("id", id).Scan(&item)
	if item.Id == 0 {
		return
	}
	if item.Status == 1 {
		_ = cronlib.StartJob(&cronlib.CronJob{
			Id:      item.Id,
			Name:    item.Name,
			Title:   item.Title,
			Pattern: item.Pattern,
			Params:  item.Params,
			Policy:  item.Policy,
			Count:   item.Count,
		})
	} else {
		cronlib.StopJob(item.Id, item.Name)
	}
}

// ==================== 执行日志 ====================

// recordLog 记录执行日志（由 cronlib 回调）
func recordLog(ctx context.Context, cronId uint64, name, title, params string, status int, output, errMsg string, takeMs int) {
	_, _ = dao.SysCronLog.Ctx(ctx).Data(g.Map{
		"cron_id":    cronId,
		"name":       name,
		"title":      title,
		"params":     params,
		"status":     status,
		"output":     output,
		"err_msg":    errMsg,
		"take_ms":    takeMs,
		"created_at": uint(time.Now().Unix()),
	}).Insert()
}

// LogList 执行日志列表
func LogList(ctx context.Context, in *adminin.CronLogListInp) (*adminin.CronLogListModel, error) {
	m := dao.SysCronLog.Ctx(ctx)
	if in.CronId > 0 {
		m = m.Where("cron_id", in.CronId)
	}
	if in.Status == 1 || in.Status == 2 {
		m = m.Where("status", in.Status)
	}

	count2, err := m.Clone().Count()
	if err != nil {
		return nil, err
	}

	if in.Page <= 0 {
		in.Page = 1
	}
	if in.PageSize <= 0 {
		in.PageSize = 20
	}

	var items []entity.SysCronLog
	err = m.OrderDesc("id").Page(in.Page, in.PageSize).Scan(&items)
	if err != nil {
		return nil, err
	}

	list := make([]adminin.CronLogListItem, 0, len(items))
	for _, item := range items {
		list = append(list, adminin.CronLogListItem{
			Id:        item.Id,
			CronId:    item.CronId,
			Name:      item.Name,
			Title:     item.Title,
			Params:    item.Params,
			Status:    item.Status,
			Output:    item.Output,
			ErrMsg:    item.ErrMsg,
			TakeMs:    item.TakeMs,
			CreatedAt: uint(item.CreatedAt),
		})
	}

	return &adminin.CronLogListModel{
		List: list,
		PageRes: form.PageRes{
			Page:     in.Page,
			PageSize: in.PageSize,
			Total:    count2,
		},
	}, nil
}

// LogClear 清空执行日志
func LogClear(ctx context.Context, in *adminin.CronLogClearInp) error {
	m := dao.SysCronLog.Ctx(ctx)
	if in.CronId > 0 {
		m = m.Where("cron_id", in.CronId)
	} else {
		// GoFrame 不允许无 WHERE 条件的 Delete，清空全部时补一个恒真条件
		m = m.Where("1=1")
	}
	_, err := m.Delete()
	return err
}

// ==================== 分组 ====================

// GroupList 分组列表
func GroupList(ctx context.Context, in *adminin.CronGroupListInp) (*adminin.CronGroupListModel, error) {
	m := dao.SysCronGroup.Ctx(ctx)
	if in.Status == 0 || in.Status == 1 {
		m = m.Where("status", in.Status)
	}

	total, _ := m.Clone().Count()

	if in.Page <= 0 {
		in.Page = 1
	}
	if in.PageSize <= 0 {
		in.PageSize = 50
	}

	var items []entity.SysCronGroup
	_ = m.OrderAsc("sort").OrderDesc("id").Page(in.Page, in.PageSize).Scan(&items)

	list := make([]adminin.CronGroupListItem, 0, len(items))
	for _, item := range items {
		list = append(list, adminin.CronGroupListItem{
			Id:        item.Id,
			Name:      item.Name,
			Sort:      item.Sort,
			Remark:    item.Remark,
			Status:    item.Status,
			CreatedAt: uint(item.CreatedAt),
			UpdatedAt: uint(item.UpdatedAt),
		})
	}

	return &adminin.CronGroupListModel{
		List: list,
		PageRes: form.PageRes{
			Page:     in.Page,
			PageSize: in.PageSize,
			Total:    total,
		},
	}, nil
}

// GroupSave 保存分组
func GroupSave(ctx context.Context, in *adminin.CronGroupSaveInp) (uint64, error) {
	now := uint(time.Now().Unix())
	data := g.Map{
		"name":       in.Name,
		"sort":       in.Sort,
		"remark":     in.Remark,
		"status":     in.Status,
		"updated_at": now,
	}

	if in.Id == 0 {
		data["created_at"] = now
		result, err := dao.SysCronGroup.Ctx(ctx).Data(data).Insert()
		if err != nil {
			return 0, err
		}
		lastId, _ := result.LastInsertId()
		return uint64(lastId), nil
	}

	_, err := dao.SysCronGroup.Ctx(ctx).Where("id", in.Id).Data(data).Update()
	return in.Id, err
}

// GroupDelete 删除分组
func GroupDelete(ctx context.Context, id uint64) error {
	// 检查是否有任务引用
	count, err := dao.SysCron.Ctx(ctx).Where("group_id", id).Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return gerror.New("该分组下还有任务，无法删除")
	}
	_, err = dao.SysCronGroup.Ctx(ctx).Where("id", id).Delete()
	return err
}

// GroupSelect 分组下拉选项
func GroupSelect(ctx context.Context) ([]adminin.CronGroupSelectItem, error) {
	var items []entity.SysCronGroup
	err := dao.SysCronGroup.Ctx(ctx).Where("status", 1).OrderAsc("sort").Scan(&items)
	if err != nil {
		return nil, err
	}
	list := make([]adminin.CronGroupSelectItem, 0, len(items))
	for _, item := range items {
		list = append(list, adminin.CronGroupSelectItem{
			Id:   item.Id,
			Name: item.Name,
		})
	}
	return list, nil
}

// RegisteredTasks 获取所有已注册的任务标识（前端下拉选）
func RegisteredTasks() []string {
	return cronlib.GetAllTasks()
}
