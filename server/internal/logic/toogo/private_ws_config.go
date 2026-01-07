package toogo

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
)

// isPrivateWSEnabled controls whether we start private WS (orders/positions/account) for a platform.
//
// Motivation: Bitget private WS can be unreliable (silent / missing fields), and users may prefer
// polling-based reconciliation for "platform manual orders" visibility.
//
// Config:
// toogo:
//   privateWs:
//     enabled: true
//     platform:
//       bitget: false
func isPrivateWSEnabled(ctx context.Context, platform string) bool {
	platform = strings.ToLower(strings.TrimSpace(platform))
	enabledVar, _ := g.Cfg().Get(ctx, "toogo.privateWs.enabled")
	enabled := true
	if enabledVar != nil && enabledVar.String() != "" {
		enabled = enabledVar.Bool()
	}
	if !enabled {
		return false
	}
	if platform == "" {
		return true
	}
	v, err := g.Cfg().Get(ctx, "toogo.privateWs.platform."+platform)
	if err != nil {
		return true
	}
	// if not configured, keep default true
	if v == nil || strings.TrimSpace(v.String()) == "" {
		return true
	}
	return v.Bool()
}

func getPrivateWSPollIntervalSeconds(ctx context.Context, platform string, defaultSec int) int {
	if defaultSec <= 0 {
		defaultSec = 10
	}
	platform = strings.ToLower(strings.TrimSpace(platform))
	// allow global override
	if v, err := g.Cfg().Get(ctx, "toogo.privateWs.pollingIntervalSeconds"); err == nil && v != nil && strings.TrimSpace(v.String()) != "" {
		if n := v.Int(); n > 0 {
			return n
		}
	}
	// allow per-platform override
	if platform != "" {
		if v, err := g.Cfg().Get(ctx, "toogo.privateWs.pollingIntervalSecondsByPlatform."+platform); err == nil && v != nil && strings.TrimSpace(v.String()) != "" {
			if n := v.Int(); n > 0 {
				return n
			}
		}
	}
	return defaultSec
}
