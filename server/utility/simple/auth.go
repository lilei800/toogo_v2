// Package simple
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2025 HotGo CLI
// @Author  Ms <133814250@qq.com>
// @License  https://github.com/bufanyun/hotgo/blob/master/LICENSE
package simple

import (
	"context"
	"hotgo/internal/library/contexts"
	"hotgo/internal/model"
)

// User 获取上下文中的用户信息
func User(ctx context.Context) *model.Identity {
	return contexts.GetUser(ctx)
}

// UserId 获取上下文中的用户ID
func UserId(ctx context.Context) int64 {
	return contexts.GetUserId(ctx)
}

