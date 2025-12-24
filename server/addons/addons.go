// Package addons
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2023 HotGo CLI
// @Author  Ms <133814250@qq.com>
// @License  https://github.com/bufanyun/hotgo/blob/master/LICENSE
package addons

// 说明：
// 本目录下的 addons（hotgo/addons）是给“插件工程(plugins)”使用的轻量入口。
// 目前主服务插件体系实际实现位于 hotgo/internal/library/addons。
// 为了保证历史插件代码（如 addons/exchange_*）可编译且不影响现有业务功能，这里提供最小兼容层。

import "context"

// Skeleton 插件骨架（最小兼容，允许插件结构体 embed）
// 这里不绑定任何主服务行为，避免对现有系统产生副作用。
type Skeleton struct{}

// Module 插件元信息（最小兼容，供插件返回描述用）
type Module struct {
	Name        string
	Title       string
	Description string
	Author      string
	Version     string
	Group       string
}

// Addon 插件接口（最小兼容）
type Addon interface {
	GetModule() Module
	Install(ctx context.Context) error
	Upgrade(ctx context.Context) error
	UnInstall(ctx context.Context) error
}

var registeredAddons []Addon

// RegisterAddon 注册插件（最小兼容）
// 当前主服务运行时不依赖这里的注册结果，因此仅保留注册列表用于调试/自检。
func RegisterAddon(a Addon) {
	if a == nil {
		return
	}
	registeredAddons = append(registeredAddons, a)
}

// GetRegisteredAddons 获取已注册插件（调试用途）
func GetRegisteredAddons() []Addon {
	return registeredAddons
}