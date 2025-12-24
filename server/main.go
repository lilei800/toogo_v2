// Package main
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2023 HotGo CLI
// @Author  Ms <133814250@qq.com>
// @License  https://github.com/bufanyun/hotgo/blob/master/LICENSE
package main

import (
	_ "hotgo/internal/packed"

	_ "hotgo/addons/modules"
	"hotgo/internal/cmd"
	"hotgo/internal/global"
	_ "hotgo/internal/logic"
	"os"
	"runtime/debug"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	_ "github.com/gogf/gf/contrib/drivers/pgsql/v2"
	_ "github.com/gogf/gf/contrib/nosql/redis/v2"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
)

func main() {
	// 添加 panic 恢复机制，记录进程退出原因
	defer func() {
		if r := recover(); r != nil {
			ctx := gctx.GetInitCtx()
			g.Log().Fatalf(ctx, "进程异常退出 (panic): %+v\n堆栈信息:\n%s", r, string(debug.Stack()))
		}
	}()

	var ctx = gctx.GetInitCtx()

	// 记录进程启动信息
	g.Log().Infof(ctx, "========== HotGo 进程启动 ==========")
	g.Log().Infof(ctx, "进程 PID: %d", os.Getpid())
	g.Log().Infof(ctx, "启动时间: %s", gctx.CtxId(ctx))

	global.Init(ctx)

	// 记录服务启动
	g.Log().Infof(ctx, "开始启动服务...")

	// 运行主服务
	cmd.Main.Run(ctx)

	// 记录进程正常退出
	g.Log().Infof(ctx, "========== HotGo 进程正常退出 ==========")
}
