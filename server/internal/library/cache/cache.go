// Package cache
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2023 HotGo CLI
// @Author  Ms <133814250@qq.com>
// @License  https://github.com/bufanyun/hotgo/blob/master/LICENSE
package cache

import (
	"context"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcache"
	"github.com/gogf/gf/v2/os/gfile"
	"hotgo/internal/library/cache/file"
)

// cache 缓存驱动
var cache *gcache.Cache

// Instance 缓存实例
func Instance() *gcache.Cache {
	if cache == nil {
		panic("cache uninitialized.")
	}
	return cache
}

// SetAdapter 设置缓存适配器
func SetAdapter(ctx context.Context) {
	var adapter gcache.Adapter

	switch g.Cfg().MustGet(ctx, "cache.adapter").String() {
	case "redis":
		adapter = gcache.NewAdapterRedis(g.Redis())
	case "file":
		fileDir := g.Cfg().MustGet(ctx, "cache.fileDir").String()
		if fileDir == "" {
			// 缓存目录未配置不应导致进程退出，回退到内存缓存
			g.Log().Warningf(ctx, "file path not configured for file caching, falling back to memory cache")
			adapter = gcache.NewAdapterMemory()
		} else {
			if !gfile.Exists(fileDir) {
				if err := gfile.Mkdir(fileDir); err != nil {
					// 缓存目录创建失败不应导致进程退出，回退到内存缓存
					g.Log().Warningf(ctx, "failed to create cache directory (falling back to memory cache), err:%+v", err)
					adapter = gcache.NewAdapterMemory()
				} else {
					adapter = file.NewAdapterFile(fileDir)
				}
			} else {
				adapter = file.NewAdapterFile(fileDir)
			}
		}
	default:
		adapter = gcache.NewAdapterMemory()
	}

	// 数据库缓存，默认和通用缓冲驱动一致，如果你不想使用默认的，可以自行调整
	g.DB().GetCache().SetAdapter(adapter)

	// 通用缓存
	cache = gcache.New()
	cache.SetAdapter(adapter)
}
