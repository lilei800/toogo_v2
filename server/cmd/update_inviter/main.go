package main

import (
	"context"
	"fmt"
	"hotgo/internal/dao"
	"hotgo/internal/model/entity"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	
	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
)

func main() {
	ctx := gctx.New()

	fmt.Println("========================================")
	fmt.Println("更新用户上级关系工具")
	fmt.Println("========================================")
	
	// 目标用户名和新上级邀请码
	targetUsername := "dong"
	newInviterCode := "RSOW2235"
	
	// 步骤1: 查找新上级用户（先从toogo_user查找，再从admin_member查找永久邀请码）
	var newInviter *entity.ToogoUser
	err := dao.ToogoUser.Ctx(ctx).Where(dao.ToogoUser.Columns().InviteCode, newInviterCode).Scan(&newInviter)
	if err != nil {
		fmt.Printf("❌ 查询新上级用户失败: %v\n", err)
		return
	}
	
	// 如果在toogo_user中没找到，尝试在admin_member的永久邀请码中查找
	if newInviter == nil {
		var inviterMember *entity.AdminMember
		err = dao.AdminMember.Ctx(ctx).Where("invite_code", newInviterCode).Scan(&inviterMember)
		if err != nil {
			fmt.Printf("❌ 查询永久邀请码失败: %v\n", err)
			return
		}
		if inviterMember == nil {
			fmt.Printf("❌ 未找到邀请码为 %s 的用户（在临时邀请码和永久邀请码中均未找到）\n", newInviterCode)
			return
		}
		// 根据member_id获取toogo_user信息
		err = dao.ToogoUser.Ctx(ctx).Where(dao.ToogoUser.Columns().MemberId, inviterMember.Id).Scan(&newInviter)
		if err != nil || newInviter == nil {
			fmt.Printf("❌ 未找到会员ID为 %d 的Toogo用户信息\n", inviterMember.Id)
			return
		}
		fmt.Printf("✅ 在永久邀请码中找到用户\n")
	}
	
	// 获取新上级的用户名
	var newInviterMember *entity.AdminMember
	dao.AdminMember.Ctx(ctx).Where("id", newInviter.MemberId).Scan(&newInviterMember)
	newInviterUsername := "未知"
	if newInviterMember != nil {
		newInviterUsername = newInviterMember.Username
	}
	
	fmt.Printf("\n✅ 找到新上级用户:\n")
	fmt.Printf("   用户名: %s\n", newInviterUsername)
	fmt.Printf("   会员ID: %d\n", newInviter.MemberId)
	fmt.Printf("   邀请码: %s\n", newInviter.InviteCode)
	fmt.Printf("   当前邀请人数: %d\n", newInviter.InviteCount)
	
	// 步骤2: 查找目标用户
	var targetMember *entity.AdminMember
	err = dao.AdminMember.Ctx(ctx).Where("username", targetUsername).Scan(&targetMember)
	if err != nil || targetMember == nil {
		fmt.Printf("❌ 未找到用户名为 %s 的用户\n", targetUsername)
		return
	}
	
	var targetUser *entity.ToogoUser
	err = dao.ToogoUser.Ctx(ctx).Where(dao.ToogoUser.Columns().MemberId, targetMember.Id).Scan(&targetUser)
	if err != nil || targetUser == nil {
		fmt.Printf("❌ 未找到用户 %s 的Toogo信息\n", targetUsername)
		return
	}
	
	fmt.Printf("\n✅ 找到目标用户:\n")
	fmt.Printf("   用户名: %s\n", targetUsername)
	fmt.Printf("   会员ID: %d\n", targetUser.MemberId)
	fmt.Printf("   当前上级ID: %d\n", targetUser.InviterId)
	
	// 获取旧上级信息
	oldInviterUsername := "无"
	if targetUser.InviterId > 0 {
		var oldInviter *entity.ToogoUser
		dao.ToogoUser.Ctx(ctx).Where(dao.ToogoUser.Columns().MemberId, targetUser.InviterId).Scan(&oldInviter)
		if oldInviter != nil {
			var oldInviterMember *entity.AdminMember
			dao.AdminMember.Ctx(ctx).Where("id", oldInviter.MemberId).Scan(&oldInviterMember)
			if oldInviterMember != nil {
				oldInviterUsername = oldInviterMember.Username
			}
		}
	}
	fmt.Printf("   当前上级: %s\n", oldInviterUsername)
	
	// 步骤3: 执行更新操作（在事务中）
	fmt.Printf("\n🔄 开始更新上级关系...\n")
	
	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 如果有旧上级，减少旧上级的邀请计数
		if targetUser.InviterId > 0 {
			_, err := dao.ToogoUser.Ctx(ctx).
				Where(dao.ToogoUser.Columns().MemberId, targetUser.InviterId).
				Data(g.Map{
					dao.ToogoUser.Columns().InviteCount: gdb.Raw("GREATEST(0, invite_count - 1)"),
				}).
				Update()
			if err != nil {
				return fmt.Errorf("减少旧上级邀请计数失败: %v", err)
			}
			fmt.Printf("   ✓ 已减少旧上级 %s 的邀请计数\n", oldInviterUsername)
		}
		
		// 更新目标用户的inviter_id
		_, err := dao.ToogoUser.Ctx(ctx).
			Where(dao.ToogoUser.Columns().MemberId, targetUser.MemberId).
			Data(g.Map{
				dao.ToogoUser.Columns().InviterId: newInviter.MemberId,
			}).
			Update()
		if err != nil {
			return fmt.Errorf("更新用户上级失败: %v", err)
		}
		fmt.Printf("   ✓ 已更新 %s 的上级为 %s\n", targetUsername, newInviterUsername)
		
		// 增加新上级的邀请计数
		_, err = dao.ToogoUser.Ctx(ctx).
			Where(dao.ToogoUser.Columns().MemberId, newInviter.MemberId).
			Data(g.Map{
				dao.ToogoUser.Columns().InviteCount: gdb.Raw("invite_count + 1"),
			}).
			Update()
		if err != nil {
			return fmt.Errorf("增加新上级邀请计数失败: %v", err)
		}
		fmt.Printf("   ✓ 已增加新上级 %s 的邀请计数\n", newInviterUsername)
		
		return nil
	})
	
	if err != nil {
		fmt.Printf("\n❌ 更新失败: %v\n", err)
		return
	}
	
	// 步骤4: 验证更新结果
	fmt.Printf("\n✅ 更新成功！正在验证...\n\n")
	
	var updatedUser *entity.ToogoUser
	dao.ToogoUser.Ctx(ctx).Where(dao.ToogoUser.Columns().MemberId, targetUser.MemberId).Scan(&updatedUser)
	
	var verifyInviter *entity.ToogoUser
	dao.ToogoUser.Ctx(ctx).Where(dao.ToogoUser.Columns().MemberId, updatedUser.InviterId).Scan(&verifyInviter)
	
	verifyInviterUsername := "未知"
	if verifyInviter != nil {
		var verifyMember *entity.AdminMember
		dao.AdminMember.Ctx(ctx).Where("id", verifyInviter.MemberId).Scan(&verifyMember)
		if verifyMember != nil {
			verifyInviterUsername = verifyMember.Username
		}
	}
	
	fmt.Println("========================================")
	fmt.Println("📊 最终结果:")
	fmt.Println("========================================")
	fmt.Printf("用户: %s\n", targetUsername)
	fmt.Printf("新上级: %s (会员ID: %d)\n", verifyInviterUsername, updatedUser.InviterId)
	fmt.Printf("上级邀请码: %s\n", verifyInviter.InviteCode)
	fmt.Printf("上级邀请人数: %d\n", verifyInviter.InviteCount)
	fmt.Println("========================================")
	fmt.Println("✅ 操作完成！")
}

