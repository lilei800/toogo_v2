// Package crons Toogo定时任务
package crons

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"hotgo/internal/dao"
	"hotgo/internal/library/cron"
)

func init() {
	cron.Register(ToogoPowerSettlementTask)
	cron.Register(ToogoInviteCodeCleanupTask)
	cron.Register(ToogoVipLevelCheckTask)
}

// ToogoPowerSettlementTask 算力结算定时任务
var ToogoPowerSettlementTask = &cToogoPowerSettlement{name: "ToogoPowerSettlement"}

type cToogoPowerSettlement struct {
	name string
}

func (c *cToogoPowerSettlement) GetName() string {
	return c.name
}

// Execute 执行任务
func (c *cToogoPowerSettlement) Execute(ctx context.Context, parser *cron.Parser) (err error) {
	parser.Logger.Debug(ctx, "[Cron] ToogoPowerSettlement: 执行算力结算")
	// 算力结算主要在平仓时实时完成
	// 这里可以做一些补偿性结算或统计
	return
}

// ToogoInviteCodeCleanupTask 清理过期邀请码
var ToogoInviteCodeCleanupTask = &cToogoInviteCodeCleanup{name: "ToogoInviteCodeCleanup"}

type cToogoInviteCodeCleanup struct {
	name string
}

func (c *cToogoInviteCodeCleanup) GetName() string {
	return c.name
}

// Execute 清理24小时前的邀请码
func (c *cToogoInviteCodeCleanup) Execute(ctx context.Context, parser *cron.Parser) (err error) {
	parser.Logger.Debug(ctx, "[Cron] ToogoInviteCodeCleanup: 清理过期邀请码")
	
	// 邀请码24小时有效，过期后自动刷新
	expireTime := time.Now().Add(-24 * time.Hour)
	
	result, err := dao.ToogoUser.Ctx(ctx).
		WhereLT("invite_code_expire", expireTime).
		Data(g.Map{
			"invite_code":        "", // 清空邀请码，下次请求时重新生成
			"invite_code_expire": nil,
		}).
		Update()
	
	if err != nil {
		parser.Logger.Warning(ctx, "[Cron] 清理过期邀请码失败:", err)
		return err
	}
	
	affected, _ := result.RowsAffected()
	if affected > 0 {
		parser.Logger.Debugf(ctx, "[Cron] 已清理%d个过期邀请码", affected)
	}
	return
}

// ToogoVipLevelCheckTask VIP等级检查升级
var ToogoVipLevelCheckTask = &cToogoVipLevelCheck{name: "ToogoVipLevelCheck"}

type cToogoVipLevelCheck struct {
	name string
}

func (c *cToogoVipLevelCheck) GetName() string {
	return c.name
}

// Execute 检查用户VIP等级是否可升级
func (c *cToogoVipLevelCheck) Execute(ctx context.Context, parser *cron.Parser) (err error) {
	parser.Logger.Debug(ctx, "[Cron] ToogoVipLevelCheck: 检查VIP等级")
	
	// 获取所有VIP等级配置
	var levels []*struct {
		Level                int     `json:"level"`
		InviteCount          int     `json:"inviteCount"`
		PersonalPowerConsume float64 `json:"personalPowerConsume"`
		TeamPowerConsume     float64 `json:"teamPowerConsume"`
	}
	
	err = dao.ToogoVipLevel.Ctx(ctx).OrderAsc("level").Scan(&levels)
	if err != nil || len(levels) == 0 {
		return
	}

	// 查询所有用户的统计数据
	type UserStat struct {
		MemberId             int64   `json:"memberId"`
		VipLevel             int     `json:"vipLevel"`
		InviteCount          int     `json:"inviteCount"`
		TotalPowerConsume    float64 `json:"totalPowerConsume"`
		TeamTotalPowerConsume float64 `json:"teamTotalPowerConsume"`
	}
	
	var users []*UserStat
	err = dao.ToogoUser.Ctx(ctx).
		Fields("member_id, vip_level, invite_count, total_power_consume, team_total_power_consume").
		Scan(&users)
	if err != nil {
		return
	}

	now := time.Now()
	for _, user := range users {
		// 找到用户应该达到的等级
		targetLevel := 0
		for _, lvl := range levels {
			if user.InviteCount >= lvl.InviteCount &&
				user.TotalPowerConsume >= lvl.PersonalPowerConsume &&
				user.TeamTotalPowerConsume >= lvl.TeamPowerConsume {
				targetLevel = lvl.Level
			}
		}

		// 如果需要升级
		if targetLevel > user.VipLevel {
			_, _ = dao.ToogoUser.Ctx(ctx).
				Where("member_id", user.MemberId).
				Data(g.Map{
					"vip_level":  targetLevel,
					"updated_at": now,
				}).
				Update()
			
			parser.Logger.Infof(ctx, "[VIP] 用户升级: memberId=%d, %d -> %d", 
				user.MemberId, user.VipLevel, targetLevel)
		}
	}
	return
}
