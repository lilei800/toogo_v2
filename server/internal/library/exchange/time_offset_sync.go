package exchange

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gclient"
)

func newPublicHTTPClient(cfg *Config) *gclient.Client {
	c := gclient.New()
	c.SetTimeout(15 * time.Second)
	if cfg != nil && cfg.Proxy != nil && cfg.Proxy.Enabled {
		proxyAddr := cfg.Proxy.Host + ":" + strconv.Itoa(cfg.Proxy.Port)
		if strings.EqualFold(cfg.Proxy.Type, "socks5") {
			// GoFrame gclient 支持 socks5:// 形式的代理地址
			c.SetProxy("socks5://" + proxyAddr)
		} else {
			c.SetProxy("http://" + proxyAddr)
		}
	}
	return c
}

// SyncServerTimeOffset tries to fetch exchange server time and stores offset(server-local) in ms.
// It is used by both REST adapters and private WS login to survive local clock skew.
func SyncServerTimeOffset(ctx context.Context, cfg *Config) (offsetMs int64, ok bool) {
	if cfg == nil {
		return 0, false
	}
	platform := strings.ToLower(strings.TrimSpace(cfg.Platform))
	switch platform {
	case "okx":
		return syncOKXTimeOffset(ctx, cfg)
	default:
		return 0, false
	}
}

func syncOKXTimeOffset(ctx context.Context, cfg *Config) (int64, bool) {
	client := newPublicHTTPClient(cfg)
	resp, err := client.Get(ctx, "https://www.okx.com/api/v5/public/time")
	if err != nil {
		return 0, false
	}
	defer resp.Close()
	raw := resp.ReadAllString()

	j := gjson.New(raw)
	if j.Get("code").String() != "0" {
		return 0, false
	}
	data := j.Get("data").Array()
	if len(data) == 0 {
		return 0, false
	}
	serverMs, err2 := strconv.ParseInt(gjson.New(data[0]).Get("ts").String(), 10, 64)
	if err2 != nil || serverMs <= 0 {
		return 0, false
	}

	localMs := time.Now().UnixMilli()
	offset := serverMs - localMs
	setTimeOffsetMs(cfg, offset)
	g.Log().Infof(ctx, "[TimeSync] okx offset updated: offsetMs=%d", offset)
	return offset, true
}

// IsTimestampExpiredError is a helper for detecting "timestamp expired" across exchanges.
func IsTimestampExpiredError(err error, raw string) bool {
	msg := ""
	if err != nil {
		msg = err.Error()
	}
	combined := msg + " " + raw
	combined = strings.ToLower(combined)
	return strings.Contains(combined, "timestamp request expired") ||
		strings.Contains(combined, "请求时间戳过期") ||
		strings.Contains(combined, "timestamp expired")
}

// WrapTimestampExpired is used to tag errors; optional but helpful.
func WrapTimestampExpired(platform string, err error) error {
	if err == nil {
		return nil
	}
	return gerror.Wrapf(err, "[%s] timestamp expired", platform)
}


