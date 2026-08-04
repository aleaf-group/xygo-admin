package adminin

import (
	"xygo/internal/model/input/form"
)

// UserListInp 管理员列表查询入参
type UserListInp struct {
	form.PageReq
	Username string `p:"username" json:"username" dc:"按用户名模糊搜索"`
	Status   int    `p:"status"   d:"-1" json:"status"   dc:"状态过滤:1启用,0禁用,-1全部"`
}

// UserListItem 管理员列表项
type UserListItem struct {
	Id         uint     `json:"id"         dc:"管理员ID"`
	Username   string   `json:"username"   dc:"登录账号"`
	Nickname   string   `json:"nickname"   dc:"昵称"`
	Mobile     string   `json:"mobile"     dc:"手机号"`
	Email      string   `json:"email"      dc:"邮箱"`
	Gender     string   `json:"gender"     dc:"性别"`
	Status     int      `json:"status"     dc:"状态"`
	Avatar     string   `json:"avatar"     dc:"头像URL"`
	IsSuper    int      `json:"isSuper"    dc:"是否超管:0否,1是"`
	CreateTime int      `json:"create_time"  dc:"创建时间"`
	UpdateTime int      `json:"update_time"  dc:"更新时间"`
	Roles      []string `json:"roles"      dc:"角色标识列表"`
	RoleNames  []string `json:"roleNames"  dc:"角色名称列表"`
	IsOnline   bool     `json:"isOnline"   dc:"是否在线（WebSocket 连接）"`
}

// UserListModel 管理员列表响应模型
type UserListModel struct {
	List []UserListItem `json:"list" dc:"数据列表"`
	form.PageRes
}

// UserSaveInp 用户新增/编辑入参
type UserSaveInp struct {
	Id       uint64   `p:"id"       json:"id"       dc:"用户ID（为空表示新增）"`
	Username string   `p:"username" v:"required#用户名不能为空" json:"username" dc:"用户名"`
	Nickname string   `p:"nickname" json:"nickname" dc:"昵称"`
	Avatar   string   `p:"avatar"   json:"avatar"   dc:"头像URL"`
	Password string   `p:"password" json:"password" dc:"密码（新增必填，编辑时为空则不修改）"`
	Mobile   string   `p:"mobile"   json:"mobile"   dc:"手机号"`
	Email    string   `p:"email"    json:"email"    dc:"邮箱"`
	Gender   string   `p:"gender"   d:"0" json:"gender"   dc:"性别:0未知,1男,2女"`
	DeptId   uint64   `p:"deptId"   d:"0" json:"deptId"   dc:"部门ID"`
	Status   int      `p:"status"   d:"1" json:"status"   dc:"状态:0禁用,1启用"`
	RoleIds  []uint64 `p:"roleIds"  json:"roleIds"  dc:"角色ID列表"`
	PostIds  []uint64 `p:"postIds"  json:"postIds"  dc:"岗位ID列表"`
}

// UserDetailModel 用户详情出参（未脱敏，编辑用）
type UserDetailModel struct {
	Id       uint     `json:"id"`
	Username string   `json:"username"`
	Nickname string   `json:"nickname"`
	Mobile   string   `json:"mobile"`
	Email    string   `json:"email"`
	Gender   string   `json:"gender"`
	Avatar   string   `json:"avatar"`
	DeptId   uint64   `json:"deptId"`
	Status   int      `json:"status"`
	IsSuper  int      `json:"isSuper"`
	RoleIds  []uint64 `json:"roleIds"`
	PostIds  []uint64 `json:"postIds"`
}

// UserDeleteInp 用户删除入参
type UserDeleteInp struct {
	Id uint64 `p:"id" v:"required#用户ID不能为空" json:"id" dc:"用户ID"`
}
