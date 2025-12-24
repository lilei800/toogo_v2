// Package exchange_okx
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 HotGo CLI
// @Author  Ms <133814250@qq.com>
// @License  https://github.com/bufanyun/hotgo/blob/master/LICENSE
// OKX交易所API插件

package exchange_okx

import (
	"context"
	"hotgo/addons"
)

func init() {
	addons.RegisterAddon(&Addon{})
}

// Addon OKX交易所插件
type Addon struct {
	addons.Skeleton
}

// GetModule 获取模块
func (addon *Addon) GetModule() addons.Module {
	return addons.Module{
		Name:        "exchange_okx",
		Title:       "OKX交易所",
		Description: "OKX交易所API对接插件",
		Author:      "HotGo",
		Version:     "1.0.0",
		Group:       "exchange",
	}
}

// Install 安装插件
func (addon *Addon) Install(ctx context.Context) error {
	// 无需特殊安装操作
	return nil
}

// Upgrade 更新插件
func (addon *Addon) Upgrade(ctx context.Context) error {
	// 无需特殊更新操作
	return nil
}

// UnInstall 卸载插件
func (addon *Addon) UnInstall(ctx context.Context) error {
	// 无需特殊卸载操作
	return nil
}

