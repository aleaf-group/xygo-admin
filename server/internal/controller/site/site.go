// +----------------------------------------------------------------------
// | XYGo Admin [ Vue3 + GoFrame 企业级中后台管理系统 ]
// +----------------------------------------------------------------------
// | Copyright (c) 2026 大连星韵网络科技有限公司 All rights reserved.
// +----------------------------------------------------------------------
// | Licensed ( https://opensource.org/licenses/MIT )
// +----------------------------------------------------------------------
// | Author: 喜羊羊 <751300685@qq.com>
// +----------------------------------------------------------------------

package site

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	api "xygo/api/site"
	"xygo/internal/library/storager"
	"xygo/internal/service"
)

// getTimezone 获取时区：优先用 sys_config 配置，兜底用 config.yaml
func getTimezone(ctx context.Context, siteTimezone string) string {
	if siteTimezone != "" {
		return siteTimezone
	}
	return g.Cfg().MustGet(ctx, "server.timezone", "Asia/Shanghai").String()
}

// ControllerV1 站点公开接口
type ControllerV1 struct{}

func NewV1() *ControllerV1 {
	return &ControllerV1{}
}

// Index 返回站点基础信息（基于 basics 分组配置）
func (c *ControllerV1) Index(ctx context.Context, req *api.IndexReq) (res *api.IndexRes, err error) {
	const group = "basics"

	cfgSvc := service.SysConfig()
	if cfgSvc == nil {
		return nil, gerror.New("SysConfig service not initialized")
	}

	items, err := cfgSvc.GetConfigByGroup(ctx, group)
	if err != nil {
		return nil, err
	}

	// open_member_center: "1" 或 缺省 → 开启，"0" → 关闭
	openMemberCenter := true
	if v, ok := items["open_member_center"]; ok && v == "0" {
		openMemberCenter = false
	}

	demoMode := g.Cfg().MustGet(ctx, "system.demoMode").Bool()

	res = &api.IndexRes{
		Group:            group,
		Items:            items,
		SiteName:         items["site_name"],
		SiteSubtitle:     items["site_subtitle"],
		Icp:              items["site_icp"],
		Timezone:         getTimezone(ctx, items["site_timezone"]),
		Description:      items["site_description"],
		ThemeColor:       items["theme_color"],
		Logo:             storager.CdnUrlByPath(ctx, items["site_logo"]),
		Closed:           items["site_closed"],
		OpenMemberCenter: openMemberCenter,
		DemoMode:         demoMode,
	}
	return
}
