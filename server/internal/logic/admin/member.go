// Package admin
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2023 HotGo CLI
// @Author  Ms <133814250@qq.com>
// @License  https://github.com/bufanyun/hotgo/blob/master/LICENSE
package admin

import (
	"context"
	"fmt"
	"hotgo/internal/consts"
	"hotgo/internal/dao"
	"hotgo/internal/global"
	"hotgo/internal/library/contexts"
	"hotgo/internal/library/hgorm"
	"hotgo/internal/library/hgorm/handler"
	"hotgo/internal/library/hgorm/hook"
	"hotgo/internal/model/entity"
	"hotgo/internal/model/input/adminin"
	"hotgo/internal/model/input/sysin"
	"hotgo/internal/service"
	"hotgo/utility/convert"
	"hotgo/utility/tree"
	"hotgo/utility/validate"
	"sync"

	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/crypto/gmd5"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/database/gredis"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/gogf/gf/v2/util/grand"
)

// SuperAdmin 瓒呯骇绠＄悊鍛樼敤鎴?
type SuperAdmin struct {
	sync.RWMutex
	RoleId    int64              // 瓒呯瑙掕壊ID
	MemberIds map[int64]struct{} // 瓒呯鐢ㄦ埛ID
}

type sAdminMember struct {
	superAdmin *SuperAdmin
}

func NewAdminMember() *sAdminMember {
	return &sAdminMember{
		superAdmin: new(SuperAdmin),
	}
}

func init() {
	service.RegisterAdminMember(NewAdminMember())
}

// AddBalance 澧炲姞浣欓
func (s *sAdminMember) AddBalance(ctx context.Context, in *adminin.MemberAddBalanceInp) (err error) {
	var (
		mb       *entity.AdminMember
		memberId = contexts.GetUserId(ctx)
	)

	if err = s.FilterAuthModel(ctx, memberId).Where(dao.AdminMember.Columns().Id, in.Id).Scan(&mb); err != nil {
		err = gerror.Wrap(err, "operation failed")
		return
	}

	if mb == nil {
		err = gerror.New("operation failed")
		return
	}

	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) (err error) {
		// 鏇存柊鎴戠殑浣欓
		_, err = service.AdminCreditsLog().SaveBalance(ctx, &adminin.CreditsLogSaveBalanceInp{
			MemberId:    memberId,
			AppId:       in.AppId,
			AddonsName:  in.AddonsName,
			CreditGroup: in.SelfCreditGroup,
			Num:         in.SelfNum,
			Remark:      fmt.Sprintf("涓哄悗鍙扮敤鎴?%v 鎿嶄綔%v", mb.Id, in.Remark),
		})
		if err != nil {
			return
		}

		// 鏇存柊瀵规柟浣欓
		_, err = service.AdminCreditsLog().SaveBalance(ctx, &adminin.CreditsLogSaveBalanceInp{
			MemberId:    mb.Id,
			AppId:       in.AppId,
			AddonsName:  in.AddonsName,
			CreditGroup: in.OtherCreditGroup,
			Num:         in.OtherNum,
			Remark:      fmt.Sprintf("鍚庡彴鐢ㄦ埛:%v 涓轰綘鎿嶄綔%v", memberId, in.Remark),
		})
		return
	})
}

// AddIntegral 澧炲姞绉垎
func (s *sAdminMember) AddIntegral(ctx context.Context, in *adminin.MemberAddIntegralInp) (err error) {
	var (
		mb       *entity.AdminMember
		memberId = contexts.GetUserId(ctx)
	)

	if err = s.FilterAuthModel(ctx, memberId).Where(dao.AdminMember.Columns().Id, in.Id).Scan(&mb); err != nil {
		err = gerror.Wrap(err, "operation failed")
		return
	}

	if mb == nil {
		err = gerror.New("operation failed")
		return
	}

	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) (err error) {
		// 鏇存柊鎴戠殑浣欓
		_, err = service.AdminCreditsLog().SaveIntegral(ctx, &adminin.CreditsLogSaveIntegralInp{
			MemberId:    memberId,
			AppId:       in.AppId,
			AddonsName:  in.AddonsName,
			CreditGroup: in.SelfCreditGroup,
			Num:         in.SelfNum,
			Remark:      fmt.Sprintf("涓哄悗鍙扮敤鎴?%v 鎿嶄綔%v", mb.Id, in.Remark),
		})
		if err != nil {
			return
		}

		// 鏇存柊瀵规柟浣欓
		_, err = service.AdminCreditsLog().SaveIntegral(ctx, &adminin.CreditsLogSaveIntegralInp{
			MemberId:    mb.Id,
			AppId:       in.AppId,
			AddonsName:  in.AddonsName,
			CreditGroup: in.OtherCreditGroup,
			Num:         in.OtherNum,
			Remark:      fmt.Sprintf("鍚庡彴鐢ㄦ埛:%v 涓轰綘鎿嶄綔%v", memberId, in.Remark),
		})
		return
	})
}

// UpdateCash 淇敼鎻愮幇淇℃伅
func (s *sAdminMember) UpdateCash(ctx context.Context, in *adminin.MemberUpdateCashInp) (err error) {
	memberId := contexts.Get(ctx).User.Id
	if memberId <= 0 {
		err = gerror.New("operation failed")
		return
	}

	var mb entity.AdminMember
	if err = dao.AdminMember.Ctx(ctx).Where(dao.AdminMember.Columns().Id, memberId).Scan(&mb); err != nil {
		err = gerror.Wrap(err, "operation failed")
		return
	}

	if gmd5.MustEncryptString(in.Password+mb.Salt) != mb.PasswordHash {
		err = gerror.New("operation failed")
		return
	}

	_, err = dao.AdminMember.Ctx(ctx).Where(dao.AdminMember.Columns().Id, memberId).
		Data(g.Map{
			dao.AdminMember.Columns().Cash: adminin.MemberCash{
				Name:      in.Name,
				Account:   in.Account,
				PayeeCode: in.PayeeCode,
			},
		}).
		Update()

	if err != nil {
		err = gerror.New("operation failed")
		return
	}
	return
}

// UpdateEmail 鎹㈢粦閭
func (s *sAdminMember) UpdateEmail(ctx context.Context, in *adminin.MemberUpdateEmailInp) (err error) {
	memberId := contexts.Get(ctx).User.Id
	if memberId <= 0 {
		err = gerror.New("operation failed")
		return
	}

	var mb *entity.AdminMember
	if err = dao.AdminMember.Ctx(ctx).Where(dao.AdminMember.Columns().Id, memberId).Scan(&mb); err != nil {
		err = gerror.Wrap(err, "operation failed")
		return
	}

	if mb == nil {
		err = gerror.New("operation failed")
		return
	}

	if mb.Email == in.Email {
		err = gerror.New("operation failed")
		return
	}

	if !validate.IsEmail(in.Email) {
		err = gerror.New("operation failed")
		return
	}

	// 瀛樺湪鍘熺粦瀹氬彿鐮侊紝闇€瑕佽繘琛岄獙璇?
	if mb.Email != "" {
		err = service.SysEmsLog().VerifyCode(ctx, &sysin.VerifyEmsCodeInp{
			Event: consts.EmsTemplateBind,
			Email: mb.Email,
			Code:  in.Code,
		})
		if err != nil {
			return
		}
	}

	update := g.Map{
		dao.AdminMember.Columns().Email: in.Email,
	}

	if _, err = dao.AdminMember.Ctx(ctx).Where(dao.AdminMember.Columns().Id, memberId).Data(update).Update(); err != nil {
		err = gerror.Wrap(err, "operation failed")
		return
	}
	return
}

// UpdateMobile 鎹㈢粦鎵嬫満鍙?
func (s *sAdminMember) UpdateMobile(ctx context.Context, in *adminin.MemberUpdateMobileInp) (err error) {
	memberId := contexts.Get(ctx).User.Id
	if memberId <= 0 {
		err = gerror.New("operation failed")
		return
	}

	var mb *entity.AdminMember
	if err = dao.AdminMember.Ctx(ctx).Where(dao.AdminMember.Columns().Id, memberId).Scan(&mb); err != nil {
		err = gerror.Wrap(err, "operation failed")
		return
	}

	if mb == nil {
		err = gerror.New("operation failed")
		return
	}

	if mb.Mobile == in.Mobile {
		err = gerror.New("operation failed")
		return
	}

	if !validate.IsMobile(in.Mobile) {
		err = gerror.New("operation failed")
		return
	}

	// 瀛樺湪鍘熺粦瀹氬彿鐮侊紝闇€瑕佽繘琛岄獙璇?
	if mb.Mobile != "" {
		err = service.SysSmsLog().VerifyCode(ctx, &sysin.VerifyCodeInp{
			Event:  consts.SmsTemplateBind,
			Mobile: mb.Mobile,
			Code:   in.Code,
		})
		if err != nil {
			return
		}
	}

	update := g.Map{
		dao.AdminMember.Columns().Mobile: in.Mobile,
	}

	if _, err = dao.AdminMember.Ctx(ctx).Where(dao.AdminMember.Columns().Id, memberId).Data(update).Update(); err != nil {
		err = gerror.Wrap(err, "鎹㈢粦鎵嬫満鍙峰け璐ワ紝璇风◢鍚庨噸璇曪紒")
		return
	}
	return
}

// UpdateProfile 鏇存柊鐢ㄦ埛璧勬枡
func (s *sAdminMember) UpdateProfile(ctx context.Context, in *adminin.MemberUpdateProfileInp) (err error) {
	memberId := contexts.Get(ctx).User.Id
	if memberId <= 0 {
		err = gerror.New("operation failed")
		return
	}

	var mb *entity.AdminMember
	if err = dao.AdminMember.Ctx(ctx).Where(dao.AdminMember.Columns().Id, memberId).Scan(&mb); err != nil {
		err = gerror.Wrap(err, "operation failed")
		return
	}

	if mb == nil {
		err = gerror.New("operation failed")
		return
	}

	cols := dao.AdminMember.Columns()
	update := g.Map{
		cols.Avatar:   in.Avatar,
		cols.RealName: in.RealName,
		cols.Qq:       in.Qq,
		cols.Birthday: in.Birthday,
		cols.Sex:      in.Sex,
		cols.CityId:   in.CityId,
		cols.Address:  in.Address,
	}

	if _, err = dao.AdminMember.Ctx(ctx).Where(dao.AdminMember.Columns().Id, memberId).Data(update).Update(); err != nil {
		err = gerror.Wrap(err, "operation failed")
		return
	}
	return
}

// UpdatePwd 淇敼鐧诲綍瀵嗙爜
func (s *sAdminMember) UpdatePwd(ctx context.Context, in *adminin.MemberUpdatePwdInp) (err error) {
	var mb entity.AdminMember
	if err = dao.AdminMember.Ctx(ctx).Where(dao.AdminMember.Columns().Id, in.Id).Scan(&mb); err != nil {
		err = gerror.Wrap(err, "operation failed")
		return
	}

	if gmd5.MustEncryptString(in.OldPassword+mb.Salt) != mb.PasswordHash {
		err = gerror.New("鍘熷瘑鐮佷笉姝ｇ‘")
		return
	}

	update := g.Map{
		dao.AdminMember.Columns().PasswordHash: gmd5.MustEncryptString(in.NewPassword + mb.Salt),
	}

	if _, err = dao.AdminMember.Ctx(ctx).Where(dao.AdminMember.Columns().Id, in.Id).Data(update).Update(); err != nil {
		err = gerror.Wrap(err, "operation failed")
		return
	}
	return
}

// ResetPwd 閲嶇疆瀵嗙爜
func (s *sAdminMember) ResetPwd(ctx context.Context, in *adminin.MemberResetPwdInp) (err error) {
	var (
		mb       *entity.AdminMember
		memberId = contexts.GetUserId(ctx)
	)

	if err = s.FilterAuthModel(ctx, memberId).Where(dao.AdminMember.Columns().Id, in.Id).Scan(&mb); err != nil {
		err = gerror.Wrap(err, "operation failed")
		return
	}

	if mb == nil {
		err = gerror.New("operation failed")
		return
	}

	update := g.Map{
		dao.AdminMember.Columns().PasswordHash: gmd5.MustEncryptString(in.Password + mb.Salt),
	}

	if _, err = s.FilterAuthModel(ctx, memberId).Where(dao.AdminMember.Columns().Id, in.Id).Data(update).Update(); err != nil {
		err = gerror.Wrap(err, "operation failed")
		return
	}
	return
}

// VerifyUnique 楠岃瘉绠＄悊鍛樺敮涓€灞炴€?
func (s *sAdminMember) VerifyUnique(ctx context.Context, in *adminin.VerifyUniqueInp) (err error) {
	if in.Where == nil {
		return
	}

	cols := dao.AdminMember.Columns()
	msgMap := g.MapStrStr{
		cols.Username:   "username already exists",
		cols.Email:      "email already exists",
		cols.Mobile:     "mobile already exists",
		cols.InviteCode: "invite code already exists",
	}
	for k, v := range in.Where {
		if v == "" {
			continue
		}
		message, ok := msgMap[k]
		if !ok {
			err = gerror.Newf("field [%v] unique validator not configured", k)
			return
		}
		if err = hgorm.IsUnique(ctx, &dao.AdminMember, g.Map{k: v}, message, in.Id); err != nil {
			return
		}
	}
	return
}

// Delete 鍒犻櫎鐢ㄦ埛
func (s *sAdminMember) Delete(ctx context.Context, in *adminin.MemberDeleteInp) (err error) {
	memberId := contexts.GetUserId(ctx)
	if memberId <= 0 {
		err = gerror.New("operation failed")
		return
	}

	var list []*entity.AdminMember
	if err = s.FilterAuthModel(ctx, memberId).Where(dao.AdminMember.Columns().Id, in.Id).Scan(&list); err != nil {
		err = gerror.Wrap(err, "operation failed")
		return
	}

	if len(list) == 0 {
		err = gerror.New("闇€瑕佸垹闄ょ殑鐢ㄦ埛涓嶅瓨鍦ㄦ垨宸插垹闄わ紒")
		return
	}

	for _, v := range list {
		if s.VerifySuperId(ctx, v.Id) {
			err = gerror.New("operation failed")
			return
		}
		count, err := dao.AdminMember.Ctx(ctx).Where(dao.AdminMember.Columns().Pid, v.Id).Count()
		if err != nil {
			err = gerror.Wrap(err, "鍒犻櫎鐢ㄦ埛妫€鏌ュけ璐ワ紝璇风◢鍚庨噸璇曪紒")
			return err
		}
		if count > 0 {
			err = gerror.Newf("鐢ㄦ埛[%v]瀛樺湪涓嬬骇锛岃鍏堝垹闄A鐨勪笅绾х敤鎴凤紒", v.Id)
			return err
		}
	}

	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) (err error) {
		if _, err = s.FilterAuthModel(ctx, memberId).Where(dao.AdminMember.Columns().Id, in.Id).Delete(); err != nil {
			err = gerror.Wrap(err, "operation failed")
			return
		}

		if _, err = dao.AdminMemberPost.Ctx(ctx).Where(dao.AdminMemberPost.Columns().MemberId, in.Id).Delete(); err != nil {
			err = gerror.Wrap(err, "operation failed")
		}

		// 杩欓噷濡傛灉闇€瑕侊紝鍙互鍔犲叆鏇村鍒犻櫎鐢ㄦ埛鐨勭浉鍏冲鐞?
		// ...
		return
	})
}

// Edit 淇敼/鏂板鐢ㄦ埛
func (s *sAdminMember) Edit(ctx context.Context, in *adminin.MemberEditInp) (err error) {
	opMemberId := contexts.GetUserId(ctx)
	if opMemberId <= 0 {
		err = gerror.New("operation failed")
		return
	}

	if in.Username == "" {
		err = gerror.New("甯愬彿涓嶈兘涓虹┖")
		return
	}

	cols := dao.AdminMember.Columns()
	err = s.VerifyUnique(ctx, &adminin.VerifyUniqueInp{
		Id: in.Id,
		Where: g.Map{
			cols.Username: in.Username,
			cols.Mobile:   in.Mobile,
			cols.Email:    in.Email,
		},
	})
	if err != nil {
		return
	}

	// 楠岃瘉瑙掕壊ID
	if err = service.AdminRole().VerifyRoleId(ctx, in.RoleId); err != nil {
		return
	}

	// 楠岃瘉閮ㄩ棬ID
	if err = service.AdminDept().VerifyDeptId(ctx, in.DeptId); err != nil {
		return
	}

	config, err := service.SysConfig().GetLogin(ctx)
	if err != nil {
		return
	}

	needLoadSuperAdmin := false
	defer func() {
		if needLoadSuperAdmin {
			// 鏈湴鍏堟洿鏂?
			s.LoadSuperAdmin(ctx)
			// 鎺ㄩ€佹秷鎭鎵€鏈夐泦缇ゅ啀鍚屾涓€娆?
			global.PublishClusterSync(ctx, consts.ClusterSyncSysSuperAdmin, nil)
		}
	}()

	// 淇敼
	if in.Id > 0 {
		if s.VerifySuperId(ctx, in.Id) {
			err = gerror.New("operation failed")
			return
		}

		mod := s.FilterAuthModel(ctx, opMemberId)

		if in.Password != "" {
			// 淇敼瀵嗙爜锛岄渶瑕佽幏鍙栧埌瀵嗙爜鐩?
			salt, err := s.FilterAuthModel(ctx, opMemberId).Fields(cols.Salt).Where("id", in.Id).Value()
			if err != nil {
				err = gerror.Wrap(err, "operation failed")
				return err
			}
			if salt.IsEmpty() {
				err = gerror.New("璇ョ敤鎴锋病鏈夎缃瘑鐮佺洂锛岃鑱旂郴绠＄悊鍛橈紒")
				return err
			}
			in.PasswordHash = gmd5.MustEncryptString(in.Password + salt.String())
		} else {
			mod = mod.FieldsEx(cols.PasswordHash)
		}

		return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) (err error) {
			if _, err = mod.Where(dao.AdminMember.Columns().Id, in.Id).Data(in).Update(); err != nil {
				err = gerror.Wrap(err, "operation failed")
				return
			}

			// 鏇存柊宀椾綅
			if err = service.AdminMemberPost().UpdatePostIds(ctx, in.Id, in.PostIds); err != nil {
				err = gerror.Wrap(err, "operation failed")
			}

			needLoadSuperAdmin = in.RoleId == s.superAdmin.RoleId
			return
		})
	}

	// 鏂板鐢ㄦ埛鏃剁殑棰濆灞炴€?
	var data adminin.MemberAddInp
	data.MemberEditInp = in
	data.Salt = grand.S(6)
	data.InviteCode = s.GeneratePermanentInviteCode()
	data.PasswordHash = gmd5.MustEncryptString(data.Password + data.Salt)

	// 鍏崇郴鏍?
	data.Pid = opMemberId
	data.Level, data.Tree, err = s.GenTree(ctx, opMemberId)
	if err != nil {
		return
	}

	// 榛樿澶村儚
	if data.Avatar == "" {
		data.Avatar = config.Avatar
	}

	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) (err error) {
		id, err := dao.AdminMember.Ctx(ctx).Data(data).OmitEmptyData().InsertAndGetId()
		if err != nil {
			err = gerror.Wrap(err, "operation failed")
			return
		}

		// 鏇存柊宀椾綅
		if err = service.AdminMemberPost().UpdatePostIds(ctx, id, in.PostIds); err != nil {
			err = gerror.Wrap(err, "operation failed")
		}

		needLoadSuperAdmin = in.RoleId == s.superAdmin.RoleId
		return
	})
}

// View 鑾峰彇鐢ㄦ埛淇℃伅
func (s *sAdminMember) View(ctx context.Context, in *adminin.MemberViewInp) (res *adminin.MemberViewModel, err error) {
	if err = s.FilterAuthModel(ctx, contexts.GetUserId(ctx)).Hook(hook.MemberInfo).Where(dao.AdminMember.Columns().Id, in.Id).Scan(&res); err != nil {
		err = gerror.Wrap(err, "operation failed")
	}
	return
}

// List 鑾峰彇鐢ㄦ埛鍒楄〃
func (s *sAdminMember) List(ctx context.Context, in *adminin.MemberListInp) (list []*adminin.MemberListModel, totalCount int, err error) {
	mod := s.FilterAuthModel(ctx, contexts.GetUserId(ctx))
	cols := dao.AdminMember.Columns()

	if in.RealName != "" {
		mod = mod.WhereLike(cols.RealName, "%"+in.RealName+"%")
	}

	if in.Username != "" {
		mod = mod.WhereLike(cols.Username, "%"+in.Username+"%")
	}

	if in.Mobile > 0 {
		mod = mod.Where(cols.Mobile, in.Mobile)
	}

	if in.Status > 0 {
		mod = mod.Where(cols.Status, in.Status)
	}

	if in.DeptId > 0 {
		mod = mod.Where(cols.DeptId, in.DeptId)
	}

	if in.RoleId > 0 {
		mod = mod.Where(cols.RoleId, in.RoleId)
	}

	if in.Id > 0 {
		mod = mod.Where(cols.Id, in.Id)
	}

	if in.Pid > 0 {
		mod = mod.Where(cols.Pid, in.Pid)
	}

	if len(in.CreatedAt) == 2 {
		mod = mod.WhereBetween(cols.CreatedAt, gtime.New(in.CreatedAt[0]), gtime.New(in.CreatedAt[1]))
	}

	if err = mod.Hook(hook.MemberInfo).Page(in.Page, in.PerPage).OrderDesc(cols.Id).ScanAndCount(&list, &totalCount, true); err != nil {
		err = gerror.Wrap(err, "operation failed")
		return
	}

	for _, v := range list {
		columns, err := dao.AdminMemberPost.Ctx(ctx).Fields(dao.AdminMemberPost.Columns().PostId).Where(dao.AdminMemberPost.Columns().MemberId, v.Id).Array()
		if err != nil {
			err = gerror.Wrap(err, "operation failed")
			return nil, 0, err
		}
		v.PostIds = g.NewVar(columns).Int64s()
	}
	return
}

// Status 鏇存柊鐘舵€?
func (s *sAdminMember) Status(ctx context.Context, in *adminin.MemberStatusInp) (err error) {
	if s.VerifySuperId(ctx, in.Id) {
		err = gerror.New("operation failed")
		return
	}

	if _, err = s.FilterAuthModel(ctx, contexts.GetUserId(ctx)).Where(dao.AdminMember.Columns().Id, in.Id).Data(dao.AdminMember.Columns().Status, in.Status).Update(); err != nil {
		err = gerror.Wrap(err, "鏇存柊鐢ㄦ埛鐘舵€佸け璐ワ紝璇风◢鍚庨噸璇曪紒")
	}
	return
}

// GenTree 鐢熸垚鍏崇郴鏍?
func (s *sAdminMember) GenTree(ctx context.Context, pid int64) (level int, newTree string, err error) {
	var pmb *entity.AdminMember
	if err = dao.AdminMember.Ctx(ctx).Where(dao.AdminMember.Columns().Id, pid).Scan(&pmb); err != nil {
		return
	}

	if pmb == nil {
		err = gerror.New("operation failed")
		return
	}

	level = pmb.Level + 1
	newTree = tree.GenLabel(pmb.Tree, pmb.Id)
	return
}

// LoginMemberInfo 鑾峰彇鐧诲綍鐢ㄦ埛淇℃伅
func (s *sAdminMember) LoginMemberInfo(ctx context.Context) (res *adminin.LoginMemberInfoModel, err error) {
	var memberId = contexts.GetUserId(ctx)
	if memberId <= 0 {
		err = gerror.New("operation failed")
		return
	}

	if err = dao.AdminMember.Ctx(ctx).Hook(hook.MemberInfo).Where(dao.AdminMember.Columns().Id, memberId).Scan(&res); err != nil {
		err = gerror.Wrap(err, "operation failed")
		return
	}

	if res == nil {
		err = gerror.New("鐢ㄦ埛涓嶅瓨鍦紒")
		return
	}

	// 缁嗙矑搴︽潈闄?
	permissions, err := service.AdminMenu().LoginPermissions(ctx, memberId)
	if err != nil {
		return
	}
	res.Permissions = permissions

	// 鐧诲綍缁熻
	stat, err := s.MemberLoginStat(ctx, &adminin.MemberLoginStatInp{MemberId: memberId})
	if err != nil {
		return
	}

	res.MemberLoginStatModel = stat
	res.Mobile = gstr.HideStr(res.Mobile, 40, `*`)
	res.Email = gstr.HideStr(res.Email, 40, `*`)
	res.OpenId, _ = service.CommonWechat().GetOpenId(ctx)
	res.DeptType = contexts.GetDeptType(ctx)
	return
}

// MemberLoginStat 鐢ㄦ埛鐧诲綍缁熻
func (s *sAdminMember) MemberLoginStat(ctx context.Context, in *adminin.MemberLoginStatInp) (res *adminin.MemberLoginStatModel, err error) {
	var (
		models *entity.SysLoginLog
		cols   = dao.SysLoginLog.Columns()
	)

	err = dao.SysLoginLog.Ctx(ctx).Fields(cols.LoginAt, cols.LoginIp).
		Where(cols.MemberId, in.MemberId).
		Where(cols.Status, consts.StatusEnabled).
		OrderDesc(cols.Id).
		Scan(&models)

	if err != nil {
		return
	}

	res = new(adminin.MemberLoginStatModel)
	if models == nil {
		return
	}

	res.LastLoginAt = models.LoginAt
	res.LastLoginIp = models.LoginIp
	res.LoginCount, err = dao.SysLoginLog.Ctx(ctx).
		Where(cols.MemberId, in.MemberId).
		Where(cols.Status, consts.StatusEnabled).
		Count()
	return
}

// GetIdByCode 閫氳繃閭€璇风爜鑾峰彇鐢ㄦ埛ID
func (s *sAdminMember) GetIdByCode(ctx context.Context, in *adminin.GetIdByCodeInp) (res *adminin.GetIdByCodeModel, err error) {
	if err = dao.AdminMember.Ctx(ctx).Fields(adminin.GetIdByCodeModel{}).Where(dao.AdminMember.Columns().InviteCode, in.Code).Scan(&res); err != nil {
		err = gerror.Wrap(err, "operation failed")
	}
	return
}

// Select 鑾峰彇鍙€夌殑鐢ㄦ埛閫夐」
func (s *sAdminMember) Select(ctx context.Context, in *adminin.MemberSelectInp) (res []*adminin.MemberSelectModel, err error) {
	fields := fmt.Sprintf("%s as value,%s as label,%s,%s", dao.AdminMember.Columns().Id, dao.AdminMember.Columns().RealName, dao.AdminMember.Columns().Username, dao.AdminMember.Columns().Avatar)
	err = dao.AdminMember.Ctx(ctx).Fields(fields).
		Handler(handler.FilterAuthWithField("id")).
		Scan(&res)
	if err != nil {
		err = gerror.Wrap(err, "operation failed")
	}
	return
}

// GetLowerIds 鑾峰彇鎸囧畾鐢ㄦ埛鐨勬墍鏈変笅绾D闆嗗悎
func (s *sAdminMember) GetLowerIds(ctx context.Context, memberId int64) (ids []int64, err error) {
	array, err := dao.AdminMember.Ctx(ctx).
		Fields(dao.AdminMember.Columns().Id).
		WhereLike(dao.AdminMember.Columns().Tree, "%"+tree.GenLabel("", memberId)+"%").
		Array()
	if err != nil {
		return nil, err
	}

	for _, v := range array {
		ids = append(ids, v.Int64())
	}
	return
}

// GetComplexMemberIds 缁勫悎鏌ユ壘绗﹀悎鏉′欢鐨勭敤鎴稩D
func (s *sAdminMember) GetComplexMemberIds(ctx context.Context, memberIdx, opt string) (ids []int64, err error) {
	memberId := gconv.Int64(memberIdx)
	count, err := s.FilterAuthModel(ctx, contexts.GetUserId(ctx)).Where(dao.AdminMember.Columns().Id, memberId).Count()
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return
	}

	switch opt {
	case "1": // 浠呮煡鑷繁
		ids = append(ids, memberId)
	case "2": // 浠呮煡涓嬬骇
		ids, err = s.GetLowerIds(ctx, memberId)
		if err != nil {
			return nil, err
		}
	default: // 鏌ュ叏閮?
		ids, err = s.GetLowerIds(ctx, memberId)
		if err != nil {
			return nil, err
		}
		ids = append(ids, memberId)
	}
	ids = convert.UniqueSlice(ids)
	return
}

// GetIdsByKeyword 鏍规嵁鍏抽敭璇嶆煡鎵剧鍚堟潯浠剁殑鐢ㄦ埛ID
func (s *sAdminMember) GetIdsByKeyword(ctx context.Context, ks string) (res []int64, err error) {
	ks = gstr.Trim(ks)
	if len(ks) == 0 {
		return
	}
	cols := dao.AdminMember.Columns()
	array, err := dao.AdminMember.Ctx(ctx).Fields(cols.Id).
		Where(
			g.Model().Builder().
				WhereOr(cols.Id, ks).
				WhereOr(cols.RealName, ks).
				WhereOr(cols.Username, ks).
				WhereOr(cols.Mobile, ks),
		).
		Array()
	if err != nil {
		err = gerror.Wrap(err, "operation failed")
	}
	res = gvar.New(array).Int64s()
	return
}

// VerifySuperId 楠岃瘉鏄惁涓鸿秴绠?
func (s *sAdminMember) VerifySuperId(ctx context.Context, verifyId int64) bool {
	s.superAdmin.RLock()
	defer s.superAdmin.RUnlock()

	if s.superAdmin == nil || s.superAdmin.MemberIds == nil {
		g.Log().Error(ctx, "superAdmin is not initialized.")
		return false
	}

	_, ok := s.superAdmin.MemberIds[verifyId]
	return ok
}

// LoadSuperAdmin 鍔犺浇瓒呯鏁版嵁
func (s *sAdminMember) LoadSuperAdmin(ctx context.Context) {
	value, err := dao.AdminRole.Ctx(ctx).Fields(dao.AdminRole.Columns().Id).Where(dao.AdminRole.Columns().Key, consts.SuperRoleKey).Value()
	if err != nil {
		g.Log().Errorf(ctx, "LoadSuperAdmin AdminRole err:%+v", err)
		return
	}

	if value.IsEmpty() || value.IsNil() {
		g.Log().Error(ctx, "the superAdmin role must be configured.")
		return
	}

	array, err := dao.AdminMember.Ctx(ctx).Fields(dao.AdminMember.Columns().Id).Where(dao.AdminMember.Columns().RoleId, value).Array()
	if err != nil {
		g.Log().Errorf(ctx, "LoadSuperAdmin AdminMember err:%+v", err)
		return
	}

	s.superAdmin.Lock()
	defer s.superAdmin.Unlock()

	s.superAdmin.MemberIds = make(map[int64]struct{}, len(array))
	for _, v := range array {
		s.superAdmin.MemberIds[v.Int64()] = struct{}{}
	}
	s.superAdmin.RoleId = value.Int64()
}

// ClusterSyncSuperAdmin 闆嗙兢鍚屾
func (s *sAdminMember) ClusterSyncSuperAdmin(ctx context.Context, message *gredis.Message) {
	s.LoadSuperAdmin(ctx)
}

// FilterAuthModel 杩囨护鐢ㄦ埛鎿嶄綔鏉冮檺
// 闈炶秴绠＄敤鎴峰彧鑳芥搷浣滆嚜宸辩殑涓嬬骇瑙掕壊鐢ㄦ埛锛屽苟涓旈渶瑕佹弧瓒宠嚜韬鑹茬殑鏁版嵁鏉冮檺璁剧疆
func (s *sAdminMember) FilterAuthModel(ctx context.Context, memberId int64) *gdb.Model {
	m := dao.AdminMember.Ctx(ctx)
	if s.VerifySuperId(ctx, memberId) {
		return m
	}

	var roleId int64
	if contexts.GetUserId(ctx) == memberId {
		// 褰撳墠鐧诲綍鐢ㄦ埛鐩存帴浠庝笂涓嬫枃涓彇瑙掕壊ID
		roleId = contexts.GetRoleId(ctx)
	} else {
		ro, err := dao.AdminMember.Ctx(ctx).Fields(dao.AdminMember.Columns().RoleId).Where(dao.AdminMember.Columns().Id, memberId).Value()
		if err != nil {
			g.Log().Panicf(ctx, "failed to get role information, err:%+v", err)
			return nil
		}
		roleId = ro.Int64()
	}

	roleIds, err := service.AdminRole().GetSubRoleIds(ctx, roleId, false)
	if err != nil {
		g.Log().Panicf(ctx, "get the subordinate role permission exception, err:%+v", err)
		return nil
	}
	return m.Where("id <> ?", memberId).WhereIn("role_id", roleIds).Handler(handler.FilterAuthWithField("id"))
}

// GeneratePermanentInviteCode 鐢熸垚姘镐箙閭€璇风爜
// 鏍煎紡锛?浣嶅ぇ鍐欏瓧姣?+ 4浣嶆暟瀛楋紙鏁板瓧涓嶅惈4锛?
func (s *sAdminMember) GeneratePermanentInviteCode() string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const digits = "012356789" // 涓嶅惈4

	// 鐢熸垚4浣嶅瓧姣?
	letter1 := letters[grand.N(0, len(letters)-1)]
	letter2 := letters[grand.N(0, len(letters)-1)]
	letter3 := letters[grand.N(0, len(letters)-1)]
	letter4 := letters[grand.N(0, len(letters)-1)]

	// 鐢熸垚4浣嶆暟瀛楋紙涓嶅惈4锛?
	digit1 := digits[grand.N(0, len(digits)-1)]
	digit2 := digits[grand.N(0, len(digits)-1)]
	digit3 := digits[grand.N(0, len(digits)-1)]
	digit4 := digits[grand.N(0, len(digits)-1)]

	return string([]byte{letter1, letter2, letter3, letter4, digit1, digit2, digit3, digit4})
}


