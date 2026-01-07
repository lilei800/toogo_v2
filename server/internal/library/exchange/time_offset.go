package exchange

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// timeOffsetState stores serverTime-localTime offset for an exchange account.
// Keyed by platform+apiKey to keep offset consistent between REST client and private WS.
type timeOffsetState struct {
	offsetMs   atomic.Int64 // serverMs - localMs
	lastSyncMs atomic.Int64 // unix ms
}

var timeOffsetStore sync.Map // map[string]*timeOffsetState

func cfgOffsetKey(cfg *Config) string {
	if cfg == nil {
		return "nil"
	}
	// apiKey 为空时退化为指针地址，避免不同配置互相污染
	if cfg.ApiKey == "" {
		return fmt.Sprintf("%s:%p", cfg.Platform, cfg)
	}
	return fmt.Sprintf("%s:%s", cfg.Platform, cfg.ApiKey)
}

func getTimeOffsetState(cfg *Config) *timeOffsetState {
	key := cfgOffsetKey(cfg)
	if v, ok := timeOffsetStore.Load(key); ok {
		if st, ok2 := v.(*timeOffsetState); ok2 && st != nil {
			return st
		}
	}
	st := &timeOffsetState{}
	actual, _ := timeOffsetStore.LoadOrStore(key, st)
	if a, ok := actual.(*timeOffsetState); ok && a != nil {
		return a
	}
	return st
}

func getTimeOffsetMs(cfg *Config) int64 {
	return getTimeOffsetState(cfg).offsetMs.Load()
}

func setTimeOffsetMs(cfg *Config, offsetMs int64) {
	st := getTimeOffsetState(cfg)
	st.offsetMs.Store(offsetMs)
	st.lastSyncMs.Store(time.Now().UnixMilli())
}

func lastTimeSyncMs(cfg *Config) int64 {
	return getTimeOffsetState(cfg).lastSyncMs.Load()
}

func nowMsWithOffset(cfg *Config) int64 {
	return time.Now().UnixMilli() + getTimeOffsetMs(cfg)
}

func nowSecWithOffset(cfg *Config) int64 {
	return nowMsWithOffset(cfg) / 1000
}


