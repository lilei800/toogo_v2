// Package middleware
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2025 HotGo CLI
// @Author  Ms <133814250@qq.com>
// @License  https://github.com/bufanyun/hotgo/blob/master/LICENSE
package middleware

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/i18n/gi18n"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/net/gtrace"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/text/gstr"
	"go.opentelemetry.io/otel/attribute"
	"hotgo/internal/consts"
	"hotgo/internal/library/addons"
	"hotgo/internal/library/contexts"
	"hotgo/internal/library/response"
	"hotgo/internal/library/token"
	"hotgo/internal/model"
	"hotgo/internal/service"
	"hotgo/utility/simple"
	"net/http"
	"strings"
)

type sMiddleware struct {
	LoginUrl         string // 鐧诲綍璺敱鍦板潃
	DemoWhiteList    g.Map  // 婕旂ず妯″紡鏀捐鐨勮矾鐢辩櫧鍚嶅崟
	NotRecordRequest g.Map  // 涓嶈褰曡姹傛暟鎹殑璺敱锛堝綋鍓嶈姹傛暟鎹繃澶ф椂浼氬奖鍝嶅搷搴旀晥鐜囷紝鍙互灏嗚矾寰勬斁鍒拌閫夐」涓敼鍠勶級
}

func init() {
	service.RegisterMiddleware(NewMiddleware())
}

func NewMiddleware() *sMiddleware {
	return &sMiddleware{
		LoginUrl: "/common",
		DemoWhiteList: g.Map{
			"/admin/site/accountLogin": struct{}{}, // 璐﹀彿鐧诲綍
			"/admin/site/mobileLogin":  struct{}{}, // 鎵嬫満鍙风櫥褰?
			"/admin/genCodes/preview":  struct{}{}, // 棰勮浠ｇ爜
		},
		NotRecordRequest: g.Map{
			"/admin/upload/file":       struct{}{}, // 涓婁紶鏂囦欢
			"/admin/upload/uploadPart": struct{}{}, // 涓婁紶鍒嗙墖
		},
	}
}

// Ctx 鍒濆鍖栬姹備笂涓嬫枃
func (s *sMiddleware) Ctx(r *ghttp.Request) {
	// 鍥介檯鍖?
	r.SetCtx(gi18n.WithLanguage(r.Context(), simple.GetHeaderLocale(r.Context())))

	// 閾捐矾杩借釜
	if g.Cfg().MustGet(r.Context(), "jaeger.switch").Bool() {
		ctx, span := gtrace.NewSpan(r.Context(), "middleware.ctx")
		span.SetAttributes(attribute.KeyValue{
			Key:   "traceID",
			Value: attribute.StringValue(gctx.CtxId(ctx)),
		})
		span.End()
		r.SetCtx(ctx)
	}

	data := make(g.Map)
	if _, ok := s.NotRecordRequest[r.URL.Path]; ok {
		data["request.body"] = gjson.New(nil)
	} else {
		data["request.body"] = gjson.New(r.GetBodyString())
	}

	contexts.Init(r, &model.Context{
		Data:   data,
		Module: getModule(r.URL.Path),
	})

	if len(r.Cookie.GetSessionId()) == 0 {
		r.Cookie.SetSessionId(gctx.CtxId(r.Context()))
	}

	r.SetCtx(r.GetNeverDoneCtx())
	r.Middleware.Next()
}

func getModule(path string) (module string) {
	slice := strings.Split(path, "/")
	if len(slice) < 2 {
		module = consts.AppDefault
		return
	}

	if slice[1] == "" {
		module = consts.AppDefault
		return
	}
	return slice[1]
}

// CORS allows Cross-origin resource sharing.
func (s *sMiddleware) CORS(r *ghttp.Request) {
	r.Response.CORSDefault()
	r.Middleware.Next()
}

// DemoLimit 婕旂ず绯荤粺鎿嶄綔闄愬埗
func (s *sMiddleware) DemoLimit(r *ghttp.Request) {
	if !simple.IsDemo(r.Context()) {
		r.Middleware.Next()
		return
	}

	if r.Method == http.MethodPost {
		if _, ok := s.DemoWhiteList[r.URL.Path]; ok {
			r.Middleware.Next()
			return
		}
		response.JsonExit(r, gcode.CodeNotSupported.Code(), "Demo system: operation is not allowed!")
		return
	}

	r.Middleware.Next()
}

// Addon 鎻掍欢涓棿浠?
func (s *sMiddleware) Addon(r *ghttp.Request) {
	var ctx = r.Context()

	if contexts.Get(ctx).Module == "" {
		g.Log().Warning(ctx, "application module is not initialized.")
		return
	}

	// 鏇挎崲鎺夊簲鐢ㄦā鍧楀墠缂€
	path := gstr.Replace(r.URL.Path, "/"+contexts.Get(ctx).Module+"/", "", 1)
	ss := gstr.Explode("/", path)
	if len(ss) == 0 {
		g.Log().Warning(ctx, "addon was not recognized.")
		return
	}

	module := addons.GetModule(ss[0])
	if module == nil {
		g.Log().Warningf(ctx, "addon module = nil, name:%v", ss[0])
		return
	}

	sk := module.GetSkeleton()
	if sk == nil {
		g.Log().Warningf(ctx, "addon skeleton = nil, name:%v", ss[0])
		return
	}

	contexts.SetAddonName(ctx, sk.Name)
	r.Middleware.Next()
}

// DeliverUserContext 灏嗙敤鎴蜂俊鎭紶閫掑埌涓婁笅鏂囦腑
func (s *sMiddleware) DeliverUserContext(r *ghttp.Request) (err error) {
	user, err := token.ParseLoginUser(r)
	if err != nil {
		return
	}

	switch user.App {
	case consts.AppAdmin:
		if err = service.AdminSite().BindUserContext(r.Context(), user); err != nil {
			return
		}
	default:
		contexts.SetUser(r.Context(), user)
	}
	return
}

// IsExceptAuth 鏄惁鏄笉闇€瑕侀獙璇佹潈闄愮殑璺敱鍦板潃
func (s *sMiddleware) IsExceptAuth(ctx context.Context, appName, path string) bool {
	pathList := g.Cfg().MustGet(ctx, fmt.Sprintf("router.%v.exceptAuth", appName)).Strings()
	np := normalizeExceptPath(path)
	for i := 0; i < len(pathList); i++ {
		if normalizeExceptPath(pathList[i]) == np {
			return true
		}
	}
	return false
}

// IsExceptLogin 鏄惁鏄笉闇€瑕佺櫥褰曠殑璺敱鍦板潃
func (s *sMiddleware) IsExceptLogin(ctx context.Context, appName, path string) bool {
	pathList := g.Cfg().MustGet(ctx, fmt.Sprintf("router.%v.exceptLogin", appName)).Strings()
	np := normalizeExceptPath(path)
	for i := 0; i < len(pathList); i++ {
		if normalizeExceptPath(pathList[i]) == np {
			return true
		}
	}
	return false
}

// normalizeExceptPath makes exceptAuth/exceptLogin matching robust:
// - trims spaces
// - ensures leading "/"
// - removes trailing "/" (except root)
func normalizeExceptPath(p string) string {
	// Trim common whitespace first.
	p = strings.TrimSpace(p)
	if p == "" {
		return ""
	}
	// Strip BOM/zero-width chars that sometimes sneak in via Windows editors.
	p = strings.TrimPrefix(p, "\ufeff")
	p = strings.Trim(p, "\u200b\u200c\u200d")
	p = strings.TrimSpace(p)
	if p == "" {
		return ""
	}
	// Ensure leading "/".
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	// Remove trailing "/" for stable matching, but keep "/" itself.
	if len(p) > 1 {
		p = strings.TrimRight(p, "/")
	}
	return p
}
