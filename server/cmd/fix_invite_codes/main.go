package main

import (
	"context"
	"fmt"
	"hotgo/internal/dao"
	"hotgo/internal/model/entity"
	
	_ "hotgo/internal/logic"
	_ "hotgo/internal/packed"
	
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/util/grand"
)

func main() {
	ctx := gctx.New()
	
	fmt.Println("===== 开始修正永久邀请码 =====")
	fmt.Println("正在查找不符合规范的邀请码...")
	
	// 查询所有用户
	var members []*entity.AdminMember
	err := dao.AdminMember.Ctx(ctx).Scan(&members)
	if err != nil {
		fmt.Printf("查询用户失败: %v\n", err)
		return
	}
	
	var needFixCount int
	var fixedCount int
	var failedCount int
	var fixedUsers []map[string]string
	
	// 检查每个用户的邀请码
	for _, member := range members {
		if member.InviteCode == "" {
			continue
		}
		
		// 检查邀请码格式：应该是8位（4位大写字母 + 4位数字）
		if !isValidInviteCode(member.InviteCode) {
			needFixCount++
			oldCode := member.InviteCode
			
			// 生成新的符合规范的邀请码
			newCode := generatePermanentInviteCode(ctx)
			
			// 确保新邀请码不重复
			for isCodeExist(ctx, newCode) {
				newCode = generatePermanentInviteCode(ctx)
			}
			
			// 更新数据库
			_, err := dao.AdminMember.Ctx(ctx).
				Where(dao.AdminMember.Columns().Id, member.Id).
				Data(g.Map{
					dao.AdminMember.Columns().InviteCode: newCode,
				}).
				Update()
			
			if err != nil {
				fmt.Printf("❌ 更新失败 - 用户: %s (ID: %d), 原邀请码: %s, 错误: %v\n", 
					member.Username, member.Id, oldCode, err)
				failedCount++
			} else {
				fmt.Printf("✅ 已修正 - 用户: %s (ID: %d), 原邀请码: %s -> 新邀请码: %s\n", 
					member.Username, member.Id, oldCode, newCode)
				fixedCount++
				fixedUsers = append(fixedUsers, map[string]string{
					"username": member.Username,
					"id":       fmt.Sprintf("%d", member.Id),
					"oldCode":  oldCode,
					"newCode":  newCode,
				})
			}
		}
	}
	
	// 输出统计结果
	fmt.Println("\n===== 修正完成 =====")
	fmt.Printf("总用户数: %d\n", len(members))
	fmt.Printf("需要修正: %d\n", needFixCount)
	fmt.Printf("修正成功: %d\n", fixedCount)
	fmt.Printf("修正失败: %d\n", failedCount)
	
	// 输出修正清单
	if len(fixedUsers) > 0 {
		fmt.Println("\n===== 修正清单 =====")
		for _, user := range fixedUsers {
			fmt.Printf("用户: %s (ID: %s)\n", user["username"], user["id"])
			fmt.Printf("  原邀请码: %s\n", user["oldCode"])
			fmt.Printf("  新邀请码: %s\n", user["newCode"])
			fmt.Println()
		}
	}
	
	fmt.Println("===== 处理完毕 =====")
}

// isValidInviteCode 检查邀请码格式是否符合规范
// 规范：8位，前4位是大写字母，后4位是数字（不含4）
func isValidInviteCode(code string) bool {
	if len(code) != 8 {
		return false
	}
	
	// 检查前4位是否都是大写字母
	for i := 0; i < 4; i++ {
		if code[i] < 'A' || code[i] > 'Z' {
			return false
		}
	}
	
	// 检查后4位是否都是数字且不含4
	validDigits := "012356789"
	for i := 4; i < 8; i++ {
		found := false
		for j := 0; j < len(validDigits); j++ {
			if code[i] == validDigits[j] {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	
	return true
}

// generatePermanentInviteCode 生成永久邀请码
// 格式：4位大写字母 + 4位数字（数字不含4）
func generatePermanentInviteCode(ctx context.Context) string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const digits = "012356789" // 不含4
	
	// 生成4位字母
	letter1 := letters[grand.N(0, len(letters)-1)]
	letter2 := letters[grand.N(0, len(letters)-1)]
	letter3 := letters[grand.N(0, len(letters)-1)]
	letter4 := letters[grand.N(0, len(letters)-1)]
	
	// 生成4位数字（不含4）
	digit1 := digits[grand.N(0, len(digits)-1)]
	digit2 := digits[grand.N(0, len(digits)-1)]
	digit3 := digits[grand.N(0, len(digits)-1)]
	digit4 := digits[grand.N(0, len(digits)-1)]
	
	return string([]byte{letter1, letter2, letter3, letter4, digit1, digit2, digit3, digit4})
}

// isCodeExist 检查邀请码是否已存在
func isCodeExist(ctx context.Context, code string) bool {
	count, err := dao.AdminMember.Ctx(ctx).
		Where(dao.AdminMember.Columns().InviteCode, code).
		Count()
	if err != nil {
		return true // 出错时认为已存在，避免重复
	}
	return count > 0
}

