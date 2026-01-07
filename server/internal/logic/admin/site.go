// Package admin
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2023 HotGo CLI
// @Author  Ms <133814250@qq.com>
// @License  https://github.com/bufanyun/hotgo/blob/master/LICENSE
package admin

import (
	"context"
	"hotgo/internal/consts"
	"hotgo/internal/dao"
	"hotgo/internal/library/contexts"
	"hotgo/internal/library/token"
	"hotgo/internal/model"
	"hotgo/internal/model/entity"
	"hotgo/internal/model/input/adminin"
	"hotgo/internal/model/input/toogoin"
	"hotgo/internal/model/input/sysin"
	"hotgo/internal/service"
	"hotgo/utility/simple"

	"github.com/gogf/gf/v2/crypto/gmd5"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/grand"
)

type sAdminSite struct{}

func NewAdminSite() *sAdminSite {
	return &sAdminSite{}
}

func init() {
	service.RegisterAdminSite(NewAdminSite())
}

// resolveInviterIdByCode 统一邀请码入口：
// - 先按基础邀请码 admin_member.invite_code 查
// - 再按 Toogo 动态邀请码 toogo_user.invite_code 查（需 status=1 且未过期）
func (s *sAdminSite) resolveInviterIdByCode(ctx context.Context, code string) (inviterId int64, err error) {
	code = gstr.Trim(code)
	if code == "" {
		return 0, nil
	}

	// 1) 基础邀请码
	pmb, err := service.AdminMember().GetIdByCode(ctx, &adminin.GetIdByCodeInp{Code: code})
	if err != nil {
		return 0, err
	}
	if pmb != nil && pmb.Id > 0 {
		return pmb.Id, nil
	}

	// 2) Toogo 动态邀请码
	var tu *entity.ToogoUser
	err = dao.ToogoUser.Ctx(ctx).
		Where(dao.ToogoUser.Columns().InviteCode, code).
		Where(dao.ToogoUser.Columns().Status, 1).
		Scan(&tu)
	if err != nil {
		return 0, gerror.Wrap(err, "查询邀请码失败")
	}
	if tu == nil || tu.MemberId <= 0 {
		return 0, nil
	}
	if tu.InviteCodeExpire != nil && tu.InviteCodeExpire.Before(gtime.Now()) {
		return 0, nil
	}
	return tu.MemberId, nil
}

// Register 账号注册
func (s *sAdminSite) Register(ctx context.Context, in *adminin.RegisterInp) (err error) {
	config, err := service.SysConfig().GetLogin(ctx)
	if err != nil {
		return
	}

	if config.ForceInvite == 1 && in.InviteCode == "" {
		err = gerror.New("请填写邀请码")
		return
	}

	var data adminin.MemberAddInp

	// 默认上级
	data.Pid = 1

	// 存在邀请人
	if in.InviteCode != "" {
		inviterId, err := s.resolveInviterIdByCode(ctx, in.InviteCode)
		if err != nil {
			return err
		}

		if inviterId <= 0 {
			err = gerror.New("邀请人信息不存在")
			return err
		}

		data.Pid = inviterId
	}

	if config.RegisterSwitch != 1 {
		err = gerror.New("管理员未开放注册")
		return
	}

	if config.RoleId < 1 {
		err = gerror.New("管理员未配置默认角色")
		return
	}

	if config.DeptId < 1 {
		err = gerror.New("管理员未配置默认部门")
		return
	}

	// 验证唯一性（仅验证Email）
	err = service.AdminMember().VerifyUnique(ctx, &adminin.VerifyUniqueInp{
		Where: g.Map{
			dao.AdminMember.Columns().Email: in.Email,
		},
	})
	if err != nil {
		return
	}

	data.MemberEditInp = &adminin.MemberEditInp{
		Id:       0,
		RoleId:   config.RoleId,
		PostIds:  config.PostIds,
		DeptId:   config.DeptId,
		Username: in.Username,
		Password: in.Password,
		RealName: "",
		Avatar:   config.Avatar,
		Sex:      3, // 保密
		Email:    in.Email,
		Status:   consts.StatusEnabled,
	}
	data.Salt = grand.S(6)
	data.InviteCode = grand.S(12)
	data.PasswordHash = gmd5.MustEncryptString(data.Password + data.Salt)
	data.Level, data.Tree, err = service.AdminMember().GenTree(ctx, data.Pid)
	if err != nil {
		return
	}

	// 提交注册信息
	var newMemberId int64
	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) (err error) {
		// PostgreSQL 兼容：部分驱动不支持 LastInsertId，InsertAndGetId 可能失败
		// 采用事务内 Insert + LASTVAL() 获取自增ID（同项目 trading/api_config/robot 的做法）
		_, err = tx.Model(dao.AdminMember.Table()).Ctx(ctx).Data(data).OmitEmptyData().Insert()
		if err != nil {
			err = gerror.Wrap(err, consts.ErrorORM)
			return
		}
		val, e := tx.GetValue("SELECT LASTVAL()")
		if e != nil {
			err = gerror.Wrap(e, "获取用户ID失败")
			return
		}
		newMemberId = val.Int64()
		if newMemberId <= 0 {
			err = gerror.New("获取用户ID失败")
			return
		}

		// 更新岗位
		if err = service.AdminMemberPost().UpdatePostIds(ctx, newMemberId, config.PostIds); err != nil {
			err = gerror.Wrap(err, consts.ErrorORM)
		}
		return
	})
	if err != nil {
		return err
	}

	// 绑定 Toogo 邀请关系（不影响基础注册成功）
	if in.InviteCode != "" && newMemberId > 0 {
		if bindErr := service.ToogoUser().RegisterWithInvite(ctx, &toogoin.RegisterWithInviteInp{
			MemberId:    newMemberId,
			InviteCode:  in.InviteCode,
		}); bindErr != nil {
			g.Log().Warningf(ctx, "绑定Toogo邀请关系失败: %v", bindErr)
		}
	}
	return nil
}

// AccountLogin 账号登录
func (s *sAdminSite) AccountLogin(ctx context.Context, in *adminin.AccountLoginInp) (res *adminin.LoginModel, err error) {
	defer func() {
		service.SysLoginLog().Push(ctx, &sysin.LoginLogPushInp{Response: res, Err: err})
	}()

	var mb *entity.AdminMember
	if err = dao.AdminMember.Ctx(ctx).Where("username", in.Username).Scan(&mb); err != nil {
		err = gerror.Wrap(err, consts.ErrorORM)
		return
	}

	if mb == nil {
		err = gerror.New("用户名或密码错误")
		return
	}

	res = new(adminin.LoginModel)
	res.Id = mb.Id
	res.Username = mb.Username
	if mb.Salt == "" {
		err = gerror.New("用户信息错误")
		return
	}

	if err = simple.CheckPassword(in.Password, mb.Salt, mb.PasswordHash); err != nil {
		return
	}

	if mb.Status != consts.StatusEnabled {
		err = gerror.New("账号已被禁用")
		return
	}

	res, err = s.handleLogin(ctx, mb)
	return
}

// MobileLogin 手机号登录
func (s *sAdminSite) MobileLogin(ctx context.Context, in *adminin.MobileLoginInp) (res *adminin.LoginModel, err error) {
	defer func() {
		service.SysLoginLog().Push(ctx, &sysin.LoginLogPushInp{Response: res, Err: err})
	}()

	var mb *entity.AdminMember
	if err = dao.AdminMember.Ctx(ctx).Where("mobile ", in.Mobile).Scan(&mb); err != nil {
		err = gerror.Wrap(err, consts.ErrorORM)
		return
	}

	if mb == nil {
		err = gerror.New("账号不存在")
		return
	}

	res = new(adminin.LoginModel)
	res.Id = mb.Id
	res.Username = mb.Username

	err = service.SysSmsLog().VerifyCode(ctx, &sysin.VerifyCodeInp{
		Event:  consts.SmsTemplateLogin,
		Mobile: in.Mobile,
		Code:   in.Code,
	})

	if err != nil {
		return
	}

	if mb.Status != consts.StatusEnabled {
		err = gerror.New("账号已被禁用")
		return
	}

	res, err = s.handleLogin(ctx, mb)
	return
}

// handleLogin .
func (s *sAdminSite) handleLogin(ctx context.Context, mb *entity.AdminMember) (res *adminin.LoginModel, err error) {
	role, dept, err := s.getLoginRoleAndDept(ctx, mb.RoleId, mb.DeptId)
	if err != nil {
		return nil, err
	}

	user := &model.Identity{
		Id:       mb.Id,
		Pid:      mb.Pid,
		DeptId:   dept.Id,
		DeptType: dept.Type,
		RoleId:   role.Id,
		RoleKey:  role.Key,
		Username: mb.Username,
		RealName: mb.RealName,
		Avatar:   mb.Avatar,
		Email:    mb.Email,
		Mobile:   mb.Mobile,
		App:      consts.AppAdmin,
		LoginAt:  gtime.Now(),
	}

	lt, expires, err := token.Login(ctx, user)
	if err != nil {
		return nil, err
	}

	res = &adminin.LoginModel{
		Username: user.Username,
		Id:       user.Id,
		Token:    lt,
		Expires:  expires,
	}
	return
}

// getLoginRoleAndDept 获取登录的角色和部门信息
func (s *sAdminSite) getLoginRoleAndDept(ctx context.Context, roleId, deptId int64) (role *entity.AdminRole, dept *entity.AdminDept, err error) {
	if err = dao.AdminRole.Ctx(ctx).Fields(dao.AdminRole.Columns().Id, dao.AdminRole.Columns().Key, dao.AdminRole.Columns().Status).Where(dao.AdminRole.Columns().Id, roleId).Scan(&role); err != nil {
		err = gerror.Wrap(err, consts.ErrorORM)
		return
	}

	if role == nil {
		err = gerror.New("角色不存在或已被删除")
		return
	}

	if role.Status != consts.StatusEnabled {
		err = gerror.New("角色已被禁用，如有疑问请联系管理员")
		return
	}

	if err = dao.AdminDept.Ctx(ctx).Fields(dao.AdminDept.Columns().Id, dao.AdminDept.Columns().Type, dao.AdminDept.Columns().Status).Where(dao.AdminDept.Columns().Id, deptId).Scan(&dept); err != nil {
		err = gerror.Wrap(err, "获取部门信息失败，请稍后重试！")
		return
	}

	if dept == nil {
		err = gerror.New("部门不存在或已被删除")
		return
	}

	if dept.Status != consts.StatusEnabled {
		err = gerror.New("部门已被禁用，如有疑问请联系管理员")
		return
	}
	return
}

// BindUserContext 绑定用户上下文
func (s *sAdminSite) BindUserContext(ctx context.Context, claims *model.Identity) (err error) {
	//// 如果不想每次访问都重新加载用户信息，可以放开注释。但在本次登录未失效前，用户信息不会刷新
	// contexts.SetUser(ctx, claims)
	// return

	var mb *entity.AdminMember
	if err = dao.AdminMember.Ctx(ctx).Where(dao.AdminMember.Columns().Id, claims.Id).Scan(&mb); err != nil {
		err = gerror.Wrap(err, "获取用户信息失败，请稍后重试！")
		return
	}

	if mb == nil {
		err = gerror.Wrap(err, "账号不存在或已被删除！")
		return
	}

	if mb.Status != consts.StatusEnabled {
		err = gerror.New("账号已被禁用，如有疑问请联系管理员")
		return
	}

	role, dept, err := s.getLoginRoleAndDept(ctx, mb.RoleId, mb.DeptId)
	if err != nil {
		return err
	}

	user := &model.Identity{
		Id:       mb.Id,
		Pid:      mb.Pid,
		DeptId:   dept.Id,
		DeptType: dept.Type,
		RoleId:   mb.RoleId,
		RoleKey:  role.Key,
		Username: mb.Username,
		RealName: mb.RealName,
		Avatar:   mb.Avatar,
		Email:    mb.Email,
		Mobile:   mb.Mobile,
		App:      claims.App,
		LoginAt:  claims.LoginAt,
	}

	contexts.SetUser(ctx, user)
	return
}
