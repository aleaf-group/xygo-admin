package dao

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

var SysQueueConfig = &sysQueueConfigDao{table: "xy_sys_queue_config"}

type sysQueueConfigDao struct {
	table string
}

func (d *sysQueueConfigDao) Table() string { return d.table }

func (d *sysQueueConfigDao) Ctx(ctx context.Context) *gdb.Model {
	return g.DB().Model(d.table).Safe().Ctx(ctx)
}
