// Package toogo Toogo系统配置管理
package toogo

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"hotgo/internal/dao"
)

// ToogoConfig 系统配置服务
type ToogoConfig struct{}

var configService = &ToogoConfig{}

// GetConfig 获取配置服务单例
func GetConfig() *ToogoConfig {
	return configService
}

// ConfigItem 配置项
type ConfigItem struct {
	Id          int64  `json:"id"`
	Group       string `json:"group"`
	Key         string `json:"key"`
	Value       string `json:"value"`
	Type        string `json:"type"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Sort        int    `json:"sort"`
}

// ConfigGroup 配置分组
type ConfigGroup struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}

// 配置分组定义
var ConfigGroups = []ConfigGroup{
	{Key: "basic", Label: "基础配置"},
	{Key: "power", Label: "算力配置"},
	{Key: "commission", Label: "佣金配置"},
	{Key: "withdraw", Label: "提现配置"},
	{Key: "invite", Label: "邀请配置"},
	{Key: "robot", Label: "机器人配置"},
}

// GetGroups 获取配置分组
func (c *ToogoConfig) GetGroups(ctx context.Context) []ConfigGroup {
	return ConfigGroups
}

// GetList 获取配置列表
func (c *ToogoConfig) GetList(ctx context.Context, group string) ([]*ConfigItem, error) {
	model := dao.ToogoConfig.Ctx(ctx)
	if group != "" {
		model = model.Where("group", group)
	}

	var list []*ConfigItem
	err := model.OrderAsc("sort").OrderAsc("id").Scan(&list)
	if err != nil {
		return nil, err
	}

	return list, nil
}

// Get 获取单个配置
func (c *ToogoConfig) Get(ctx context.Context, group, key string) (*ConfigItem, error) {
	var item *ConfigItem
	err := dao.ToogoConfig.Ctx(ctx).
		Where("group", group).
		Where("key", key).
		Scan(&item)
	if err != nil {
		return nil, err
	}
	return item, nil
}

// GetValue 获取配置值
func (c *ToogoConfig) GetValue(ctx context.Context, group, key string) (string, error) {
	item, err := c.Get(ctx, group, key)
	if err != nil {
		return "", err
	}
	if item == nil {
		return "", nil
	}
	return item.Value, nil
}

// GetFloat 获取浮点型配置值
func (c *ToogoConfig) GetFloat(ctx context.Context, group, key string) (float64, error) {
	val, err := c.GetValue(ctx, group, key)
	if err != nil {
		return 0, err
	}
	return g.NewVar(val).Float64(), nil
}

// GetInt 获取整型配置值
func (c *ToogoConfig) GetInt(ctx context.Context, group, key string) (int, error) {
	val, err := c.GetValue(ctx, group, key)
	if err != nil {
		return 0, err
	}
	return g.NewVar(val).Int(), nil
}

// GetBool 获取布尔型配置值
func (c *ToogoConfig) GetBool(ctx context.Context, group, key string) (bool, error) {
	val, err := c.GetValue(ctx, group, key)
	if err != nil {
		return false, err
	}
	return g.NewVar(val).Bool(), nil
}

// Update 更新配置
func (c *ToogoConfig) Update(ctx context.Context, items []ConfigUpdateItem) error {
	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		for _, item := range items {
			_, err := dao.ToogoConfig.Ctx(ctx).
				Where("group", item.Group).
				Where("key", item.Key).
				Data(g.Map{"value": item.Value}).
				Update()
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// ConfigUpdateItem 配置更新项
type ConfigUpdateItem struct {
	Group string `json:"group"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

// ====== 常用配置快捷方法 ======

// GetPowerConsumeRate 获取算力消耗比例
func (c *ToogoConfig) GetPowerConsumeRate(ctx context.Context) (float64, error) {
	return c.GetFloat(ctx, "power", "consume_rate")
}

// GetMinConsumePower 获取最小消耗算力
func (c *ToogoConfig) GetMinConsumePower(ctx context.Context) (float64, error) {
	return c.GetFloat(ctx, "power", "min_consume")
}

// GetExchangeRate 获取USDT兑算力比率
func (c *ToogoConfig) GetExchangeRate(ctx context.Context) (float64, error) {
	return c.GetFloat(ctx, "power", "exchange_rate")
}

// GetWithdrawMinAmount 获取最低提现金额
func (c *ToogoConfig) GetWithdrawMinAmount(ctx context.Context) (float64, error) {
	return c.GetFloat(ctx, "withdraw", "min_amount")
}

// GetWithdrawFeeRate 获取提现手续费比例
func (c *ToogoConfig) GetWithdrawFeeRate(ctx context.Context) (float64, error) {
	return c.GetFloat(ctx, "withdraw", "fee_rate")
}

// GetInviteCodeExpireHours 获取邀请码有效期(小时)
func (c *ToogoConfig) GetInviteCodeExpireHours(ctx context.Context) (int, error) {
	return c.GetInt(ctx, "invite", "code_expire_hours")
}

// GetRegisterRewardPower 获取注册奖励算力
func (c *ToogoConfig) GetRegisterRewardPower(ctx context.Context) (float64, error) {
	return c.GetFloat(ctx, "invite", "register_reward")
}

// NeedInviteCode 是否需要邀请码注册
func (c *ToogoConfig) NeedInviteCode(ctx context.Context) (bool, error) {
	return c.GetBool(ctx, "invite", "need_invite_code")
}

