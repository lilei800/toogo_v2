package market

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
)

// ResolveAnalysisPlatform resolves which platform should be used as the "analysis data source"
// (Klines + MarketAnalyzer) for a given execution platform.
//
// Default behavior: analysisPlatform == execPlatform.
//
// Config:
//   toogo.analysisSourceOverrides.<execPlatform>: "<analysisPlatform>"
// Example:
//   toogo:
//     analysisSourceOverrides:
//       gate: okx
func ResolveAnalysisPlatform(ctx context.Context, execPlatform string) string {
	execPlatform = normalizePlatform(execPlatform)
	if execPlatform == "" {
		return ""
	}

	// Per-platform override (string).
	v, _ := g.Cfg().Get(ctx, "toogo.analysisSourceOverrides."+execPlatform)
	if !v.IsEmpty() {
		ap := normalizePlatform(strings.TrimSpace(v.String()))
		if ap != "" {
			return ap
		}
	}
	return execPlatform
}
