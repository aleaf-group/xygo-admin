package user

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/gogf/gf/v2/crypto/gmd5"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/grand"

	"xygo/internal/consts"
	"xygo/internal/dao"
	"xygo/internal/library/contexts"
	"xygo/internal/library/dbdialect"
	"xygo/internal/library/hgorm/handler"
	"xygo/internal/library/security"
	"xygo/internal/model/entity"
	"xygo/internal/model/input/adminin"
	"xygo/internal/service"
	"xygo/internal/websocket"
)

type sAdminUser struct{}

func init() {
	service.RegisterAdminUser(New())
}

// New 构造用户服务
func New() *sAdminUser {
	return &sAdminUser{}
}

// List 获取用户列表
func (s *sAdminUser) List(ctx context.Context, in *adminin.UserListInp) (list []adminin.UserListItem, total int, err error) {
	// 基础查询 + 数据权限过滤
	model := dao.AdminUser.Ctx(ctx).Handler(handler.FilterAuth)

	// 过滤：用户名（模糊）
	if in.Username != "" {
		model = model.WhereLike("username", "%"+in.Username+"%")
	}

	// 过滤：状态
	if in.Status == 0 || in.Status == 1 {
		model = model.Where("status", in.Status)
	}

	// 统计总数（会自动应用数据权限过滤）
	count, err := model.Clone().Count()
	if err != nil {
		return nil, 0, err
	}

	// 分页参数兜底
	if in.Page <= 0 {
		in.Page = 1
	}
	if in.PageSize <= 0 {
		in.PageSize = 20
	}

	var records []adminin.UserListItem
	err = model.
		Fields("id, username, nickname, mobile, email, gender, status, avatar, is_super, create_time, update_time").
		Page(in.Page, in.PageSize).
		Scan(&records)
	if err != nil {
		return nil, 0, err
	}

	// 收集用户ID
	userIds := make([]uint, 0, len(records))
	for _, r := range records {
		userIds = append(userIds, r.Id)
	}

	// 查询用户角色（key 与 name）
	roleMap := make(map[uint][]string)
	roleNameMap := make(map[uint][]string)
	if len(userIds) > 0 {
		var roleRows []struct {
			UserId   uint   `json:"userId"`
			RoleKey  string `json:"roleKey"`
			RoleName string `json:"roleName"`
		}
		if err = dao.AdminUserRole.Ctx(ctx).
			As("ur").
			LeftJoin(dao.AdminRole.Table()+" r", "ur.role_id = r.id").
			Fields("ur.user_id as userId, r."+dbdialect.Get().QuoteIdentifier("key")+" as roleKey, r."+dbdialect.Get().QuoteIdentifier("name")+" as roleName").
			WhereIn("ur.user_id", userIds).
			Scan(&roleRows); err != nil {
			return nil, 0, err
		}
		for _, row := range roleRows {
			if row.RoleKey != "" {
				roleMap[row.UserId] = append(roleMap[row.UserId], row.RoleKey)
			}
			if row.RoleName != "" {
				roleNameMap[row.UserId] = append(roleNameMap[row.UserId], row.RoleName)
			}
		}
	}

	// 填充角色与头像兜底 + 敏感字段脱敏
	for i := range records {
		records[i].Roles = roleMap[records[i].Id]
		records[i].RoleNames = roleNameMap[records[i].Id]
		if records[i].Roles == nil {
			records[i].Roles = []string{}
		}
		if records[i].RoleNames == nil {
			records[i].RoleNames = []string{}
		}
		if records[i].Avatar == "" {
			// 生成首字母头像
			name := records[i].Username
			if name == "" {
				name = fmt.Sprintf("U%d", records[i].Id)
			}
			records[i].Avatar = fmt.Sprintf("https://ui-avatars.com/api/?background=random&name=%s", url.QueryEscape(name))
		}
		// ✨ 敏感字段脱敏（列表不返回明文）
		if records[i].Mobile != "" {
			records[i].Mobile = security.MaskMobile(records[i].Mobile)
		}
		if records[i].Email != "" {
			records[i].Email = security.MaskEmail(records[i].Email)
		}
		records[i].IsOnline = websocket.IsUserOnline("admin", uint64(records[i].Id))
	}

	return records, count, nil
}

// Detail 获取用户详情
func (s *sAdminUser) Detail(ctx context.Context, id uint64) (*adminin.UserListItem, error) {
	var user *entity.AdminUser

	if err := dao.AdminUser.Ctx(ctx).
		Where("id", id).
		Scan(&user); err != nil {
		return nil, err
	}

	if user == nil {
		return nil, gerror.NewCode(consts.CodeDataNotFound, "用户不存在")
	}

	return &adminin.UserListItem{
		Id:         uint(user.Id),
		Username:   user.Username,
		Nickname:   user.Nickname,
		Mobile:     user.Mobile,
		Email:      user.Email,
		Gender:     strconv.Itoa(user.Gender),
		Status:     user.Status,
		Avatar:     user.Avatar,
		IsSuper:    user.IsSuper,
		CreateTime: int(user.CreateTime),
		UpdateTime: int(user.UpdateTime),
	}, nil
}

// DetailForEdit 获取用户详情（未脱敏，编辑用，含角色和岗位ID）
func (s *sAdminUser) DetailForEdit(ctx context.Context, id uint64) (*adminin.UserDetailModel, error) {
	var user *entity.AdminUser
	if err := dao.AdminUser.Ctx(ctx).Where("id", id).Scan(&user); err != nil {
		return nil, err
	}
	if user == nil {
		return nil, gerror.NewCode(consts.CodeDataNotFound, "用户不存在")
	}

	result := &adminin.UserDetailModel{
		Id:       uint(user.Id),
		Username: user.Username,
		Nickname: user.Nickname,
		Mobile:   user.Mobile,
		Email:    user.Email,
		Gender:   strconv.Itoa(user.Gender),
		Avatar:   user.Avatar,
		DeptId:   user.DeptId,
		Status:   user.Status,
		IsSuper:  user.IsSuper,
	}

	// 查询角色ID
	var roleRows []struct {
		RoleId uint64 `json:"roleId"`
	}
	_ = dao.AdminUserRole.Ctx(ctx).
		Where("user_id", id).
		Fields("role_id as roleId").
		Scan(&roleRows)
	result.RoleIds = make([]uint64, 0, len(roleRows))
	for _, r := range roleRows {
		result.RoleIds = append(result.RoleIds, r.RoleId)
	}

	// 查询岗位ID
	var postRows []struct {
		PostId uint64 `json:"postId"`
	}
	_ = dao.AdminUserPost.Ctx(ctx).
		Where("user_id", id).
		Fields("post_id as postId").
		Scan(&postRows)
	result.PostIds = make([]uint64, 0, len(postRows))
	for _, r := range postRows {
		result.PostIds = append(result.PostIds, r.PostId)
	}

	return result, nil
}

// Save 保存用户（新增/编辑）
func (s *sAdminUser) Save(ctx context.Context, in *adminin.UserSaveInp) (uint, error) {
	now := uint(time.Now().Unix())
	operatorId := uint64(0)
	if u := contexts.GetUser(ctx); u != nil {
		operatorId = u.Id
	}

	// 1. 校验用户名唯一性
	m := dao.AdminUser.Ctx(ctx).Where("username", in.Username)
	if in.Id > 0 {
		m = m.WhereNot("id", in.Id)
	}
	count, err := m.Count()
	if err != nil {
		return 0, err
	}
	if count > 0 {
		return 0, gerror.New("用户名已存在")
	}

	// 2. 构建数据
	gender, _ := strconv.Atoi(in.Gender)
	data := g.Map{
		"username":    in.Username,
		"nickname":    in.Nickname,
		"avatar":      in.Avatar,
		"mobile":      in.Mobile,
		"email":       in.Email,
		"gender":      gender,
		"dept_id":     in.DeptId,
		"status":      in.Status,
		"update_time": now,
		"updated_by":  operatorId,
	}

	var userId uint
	if in.Id == 0 {
		// ======== 新增 ========
		if in.Password == "" {
			return 0, gerror.New("新增用户必须设置密码")
		}
		salt := grand.S(6) // 生成6位随机盐
		data["salt"] = salt
		data["password"] = gmd5.MustEncryptString(in.Password + salt)
		data["create_time"] = now
		data["created_by"] = operatorId

		result, err := dao.AdminUser.Ctx(ctx).Data(data).Insert()
		if err != nil {
			return 0, err
		}
		id, _ := result.LastInsertId()
		userId = uint(id)
	} else {
		// ======== 编辑 ========
		if in.Password != "" {
			// 修改密码时重新生成盐
			salt := grand.S(6)
			data["salt"] = salt
			data["password"] = gmd5.MustEncryptString(in.Password + salt)
		}
		_, err := dao.AdminUser.Ctx(ctx).Where("id", in.Id).Data(data).Update()
		if err != nil {
			return 0, err
		}
		userId = uint(in.Id)
	}

	// 3. 更新角色关联
	if in.RoleIds != nil {
		_ = syncUserRoles(ctx, uint64(userId), in.RoleIds)
	}

	// 4. 更新岗位关联
	if in.PostIds != nil {
		_ = syncUserPosts(ctx, uint64(userId), in.PostIds)
	}

	return userId, nil
}

// Delete 删除用户
func (s *sAdminUser) Delete(ctx context.Context, id uint64) error {
	// 1. 检查用户是否存在
	var user *entity.AdminUser
	if err := dao.AdminUser.Ctx(ctx).Where("id", id).Scan(&user); err != nil {
		return err
	}
	if user == nil {
		return gerror.New("用户不存在")
	}

	// 2. 超管不能删除
	if user.IsSuper == 1 {
		return gerror.New("超级管理员不能删除")
	}

	// 3. 删除用户
	if _, err := dao.AdminUser.Ctx(ctx).Where("id", id).Delete(); err != nil {
		return err
	}

	// 4. 清理关联数据
	_, _ = dao.AdminUserRole.Ctx(ctx).Where("user_id", id).Delete()
	_, _ = dao.AdminUserPost.Ctx(ctx).Where("user_id", id).Delete()

	return nil
}

// ==================== 内部辅助 ====================

// syncUserRoles 同步用户角色关联
func syncUserRoles(ctx context.Context, userId uint64, roleIds []uint64) error {
	_, _ = dao.AdminUserRole.Ctx(ctx).Where("user_id", userId).Delete()
	for _, roleId := range roleIds {
		if roleId == 0 {
			continue
		}
		_, _ = dao.AdminUserRole.Ctx(ctx).Data(g.Map{
			"user_id": userId,
			"role_id": roleId,
		}).Insert()
	}
	return nil
}

// syncUserPosts 同步用户岗位关联
func syncUserPosts(ctx context.Context, userId uint64, postIds []uint64) error {
	_, _ = dao.AdminUserPost.Ctx(ctx).Where("user_id", userId).Delete()
	for _, postId := range postIds {
		if postId == 0 {
			continue
		}
		_, _ = dao.AdminUserPost.Ctx(ctx).Data(g.Map{
			"user_id": userId,
			"post_id": postId,
		}).Insert()
	}
	return nil
}
