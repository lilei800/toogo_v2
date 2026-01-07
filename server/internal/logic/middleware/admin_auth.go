package middleware

import (
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/text/gstr"
	"hotgo/internal/consts"
	"hotgo/internal/library/contexts"
	"hotgo/internal/library/response"
	"hotgo/internal/service"
	"hotgo/utility/simple"
	"sync"
	"strings"
	"time"
)

// adminAuthHitLogLimiter 用于限制 AdminAuth hit 日志频率，避免终端洪流影响行情WS/实时报价。
// key: stripped path -> time.Time(last log)
var adminAuthHitLogLimiter sync.Map

// AdminAuth admin auth middleware
func (s *sMiddleware) AdminAuth(r *ghttp.Request) {
	ctx := r.Context()

	// Strip router prefix (/admin) to match config exceptAuth/permissions paths.
	path := gstr.Replace(r.URL.Path, simple.RouterPrefix(ctx, consts.AppAdmin), "", 1)
	// Extra fallback: in case prefix config is missing/mismatched.
	path = gstr.Replace(path, "/"+consts.AppAdmin, "", 1)

	// 注意：这里是“每个请求都会经过”的中间件，打 WARN 会造成终端洪流，影响 WS 行情回调和 UI 实时报价。
	// 默认改为 debug + 30s 节流（需要排查路由问题时再开 debug 即可）。
	if last, ok := adminAuthHitLogLimiter.Load(path); ok {
		if t, ok2 := last.(time.Time); ok2 && time.Since(t) < 30*time.Second {
			// skip
		} else {
			adminAuthHitLogLimiter.Store(path, time.Now())
			g.Log().Debugf(ctx, "AdminAuth hit raw=%s stripped=%s", r.URL.Path, path)
		}
	} else {
		adminAuthHitLogLimiter.Store(path, time.Now())
		g.Log().Debugf(ctx, "AdminAuth hit raw=%s stripped=%s", r.URL.Path, path)
	}

	// Except login
	if s.IsExceptLogin(ctx, consts.AppAdmin, path) {
		r.Middleware.Next()
		return
	}

	// Bind user into context
	if err := s.DeliverUserContext(r); err != nil {
		response.JsonExit(r, gcode.CodeNotAuthorized.Code(), err.Error())
		return
	} // Hard allowlist: required for frontend bootstrap (user info & menus).
	np := normalizeExceptPath(path)
	// Also match raw URL path (in case router prefix stripping fails).
	rawp := normalizeExceptPath(r.URL.Path)
	if np == "/member/info" || np == "/role/dynamic" ||
		strings.HasSuffix(rawp, "/member/info") || strings.HasSuffix(rawp, "/role/dynamic") {
		r.Middleware.Next()
		return
	}

	// Except auth
	if s.IsExceptAuth(ctx, consts.AppAdmin, path) {
		r.Middleware.Next()
		return
	}

	// Verify permission
	if !service.AdminRole().Verify(ctx, path, r.Method) {
		g.Log().Debugf(ctx, "AdminAuth fail path:%+v, GetRoleKey:%+v, r.Method:%+v", path, contexts.GetRoleKey(ctx), r.Method)
		response.JsonExit(r, gcode.CodeSecurityReason.Code(), "No permission")
		return
	}

	r.Middleware.Next()
}
