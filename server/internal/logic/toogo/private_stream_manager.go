package toogo

import (
	"context"
	"strings"
	"sync"
	"time"

	"hotgo/internal/dao"
	"hotgo/internal/library/exchange"
	"hotgo/internal/model/entity"
	"hotgo/utility/encrypt"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// PrivateStreamManager 管理私有WS流（按 apiConfigId 复用，避免每个机器人开一条连接）
type PrivateStreamManager struct {
	mu sync.RWMutex

	streams map[string]*privateStreamEntry // key=platform:apiConfigId

	// robotDebounce: 避免同一robot在短时间内被多次事件触发导致goroutine风暴
	robotDebounce map[int64]time.Time
}

type privateStreamEntry struct {
	platform   string
	apiConfig  *entity.TradingApiConfig
	stream     exchange.PrivateStream
	refCount   int
	symbolRefs map[string]int      // symbol -> refs
	robots     map[int64]struct{}  // robotId set
}

var (
	privateStreamManager     *PrivateStreamManager
	privateStreamManagerOnce sync.Once
)

func GetPrivateStreamManager() *PrivateStreamManager {
	privateStreamManagerOnce.Do(func() {
		privateStreamManager = &PrivateStreamManager{
			streams:       make(map[string]*privateStreamEntry),
			robotDebounce: make(map[int64]time.Time),
		}
	})
	return privateStreamManager
}

func streamKey(platform string, apiConfigId int64) string {
	return strings.ToLower(strings.TrimSpace(platform)) + ":" + g.NewVar(apiConfigId).String()
}

// Acquire 引用一个私有流（启动/复用）
func (m *PrivateStreamManager) Acquire(ctx context.Context, apiConfig *entity.TradingApiConfig, symbol string, robotId int64) error {
	if apiConfig == nil {
		return gerror.New("apiConfig is nil")
	}
	key := streamKey(apiConfig.Platform, apiConfig.Id)
	symbol = strings.ToUpper(strings.TrimSpace(symbol))

	m.mu.Lock()
	entry := m.streams[key]
	if entry != nil {
		entry.refCount++
		if symbol != "" {
			entry.symbolRefs[symbol]++
			_ = entry.stream.AddSymbol(symbol)
		}
		if robotId > 0 {
			entry.robots[robotId] = struct{}{}
		}
		m.mu.Unlock()
		return nil
	}
	m.mu.Unlock()

	// create new entry outside lock (but with a second check)
	cfg, err := buildExchangeConfigFromAPIConfig(ctx, apiConfig)
	if err != nil {
		return err
	}
	ps, err := exchange.NewPrivateStream(cfg)
	if err != nil {
		return err
	}

	// proxy dialer（复用 RobotTaskManager 的全局代理配置）
	if dialer, err := getWebSocketDialer(ctx); err == nil && dialer != nil {
		ps.SetProxyDialer(dialer)
	}
	ps.SetOnEvent(func(ev *exchange.PrivateEvent) {
		if ev != nil {
			ev.ApiConfigId = apiConfig.Id
		}
		m.onEvent(ev)
	})

	if err := ps.Start(ctx); err != nil {
		return err
	}
	if symbol != "" {
		_ = ps.AddSymbol(symbol)
	}

	m.mu.Lock()
	// double check
	if existing := m.streams[key]; existing != nil {
		// someone created in between, close this one
		m.mu.Unlock()
		ps.Stop()
		return nil
	}
	newEntry := &privateStreamEntry{
		platform:   strings.ToLower(strings.TrimSpace(apiConfig.Platform)),
		apiConfig:  apiConfig,
		stream:     ps,
		refCount:   1,
		symbolRefs: make(map[string]int),
		robots:     make(map[int64]struct{}),
	}
	if symbol != "" {
		newEntry.symbolRefs[symbol] = 1
	}
	if robotId > 0 {
		newEntry.robots[robotId] = struct{}{}
	}
	m.streams[key] = newEntry
	m.mu.Unlock()
	return nil
}

// Release 释放引用，引用归零则停止流
func (m *PrivateStreamManager) Release(platform string, apiConfigId int64, symbol string, robotId int64) {
	key := streamKey(platform, apiConfigId)
	symbol = strings.ToUpper(strings.TrimSpace(symbol))

	m.mu.Lock()
	entry := m.streams[key]
	if entry == nil {
		m.mu.Unlock()
		return
	}
	if robotId > 0 {
		delete(entry.robots, robotId)
	}
	if symbol != "" {
		if n, ok := entry.symbolRefs[symbol]; ok {
			n--
			if n <= 0 {
				delete(entry.symbolRefs, symbol)
				_ = entry.stream.RemoveSymbol(symbol)
			} else {
				entry.symbolRefs[symbol] = n
			}
		}
	}
	entry.refCount--
	if entry.refCount <= 0 {
		delete(m.streams, key)
		ps := entry.stream
		m.mu.Unlock()
		ps.Stop()
		return
	}
	m.mu.Unlock()
}

func (m *PrivateStreamManager) onEvent(ev *exchange.PrivateEvent) {
	if ev == nil {
		return
	}
	// 找到对应 stream entry，分发给关联 robot（按 apiConfigId 精准路由）
	key := streamKey(ev.Platform, ev.ApiConfigId)
	m.mu.RLock()
	entry := m.streams[key]
	if entry == nil {
		m.mu.RUnlock()
		return
	}
	// 按 symbol 过滤：事件没有 symbol（account update）则不过滤
	if ev.Symbol != "" && len(entry.symbolRefs) > 0 {
		if _, ok := entry.symbolRefs[strings.ToUpper(ev.Symbol)]; !ok {
			m.mu.RUnlock()
			return
		}
	}
	targets := make([]int64, 0, len(entry.robots))
	for rid := range entry.robots {
		targets = append(targets, rid)
	}
	m.mu.RUnlock()

	now := time.Now()
	for _, robotId := range targets {
		// debounce 200ms
		m.mu.Lock()
		last := m.robotDebounce[robotId]
		if !last.IsZero() && now.Sub(last) < 200*time.Millisecond {
			m.mu.Unlock()
			continue
		}
		m.robotDebounce[robotId] = now
		m.mu.Unlock()

		engine := GetRobotTaskManager().GetEngine(robotId)
		if engine != nil {
			go engine.syncAccountDataIfNeeded(context.Background(), "after_trade")
		}
		// 【方案A】私有WS订单事件增量落库（挂单/订单事实表）
		// 说明：落库尽量轻量，失败不阻断后续对账；若表未创建，会在执行日志里体现。
		if ev.Type == exchange.PrivateEventOrder {
			go UpsertExchangeOrdersFromPrivateEvent(context.Background(), robotId, ev)
		}
		// 触发DB对账（按robot去抖）
		GetOrderStatusSyncService().TriggerRobotSync(robotId)
	}
}

// buildExchangeConfigFromAPIConfig 构建 exchange.Config（解密字段，补代理）
func buildExchangeConfigFromAPIConfig(ctx context.Context, apiConfig *entity.TradingApiConfig) (*exchange.Config, error) {
	if apiConfig == nil {
		return nil, gerror.New("apiConfig is nil")
	}

	apiKey, err := encrypt.AesDecrypt(apiConfig.ApiKey)
	if err != nil {
		apiKey = apiConfig.ApiKey
	}
	secretKey, err := encrypt.AesDecrypt(apiConfig.SecretKey)
	if err != nil {
		secretKey = apiConfig.SecretKey
	}
	passphrase := ""
	if apiConfig.Passphrase != "" {
		p, err := encrypt.AesDecrypt(apiConfig.Passphrase)
		if err != nil {
			passphrase = apiConfig.Passphrase
		} else {
			passphrase = p
		}
	}

	// 代理配置（全局）
	var proxyCfg *exchange.ProxyConfig
	var proxyEnt *entity.TradingProxyConfig
	_ = dao.TradingProxyConfig.Ctx(ctx).
		Where(dao.TradingProxyConfig.Columns().UserId, 0).
		Where(dao.TradingProxyConfig.Columns().TenantId, 0).
		Where(dao.TradingProxyConfig.Columns().Enabled, 1).
		Scan(&proxyEnt)
	if proxyEnt != nil {
		host := proxyEnt.ProxyAddress
		port := 0
		if idx := strings.Index(proxyEnt.ProxyAddress, ":"); idx > 0 {
			host = proxyEnt.ProxyAddress[:idx]
			port = g.NewVar(proxyEnt.ProxyAddress[idx+1:]).Int()
		}
		proxyCfg = &exchange.ProxyConfig{
			Enabled:  true,
			Type:     proxyEnt.ProxyType,
			Host:     host,
			Port:     port,
			Username: proxyEnt.Username,
			Password: "",
		}
		if proxyEnt.AuthEnabled == 1 && proxyEnt.Password != "" {
			if pwd, err := encrypt.AesDecrypt(proxyEnt.Password); err == nil {
				proxyCfg.Password = pwd
			}
		}
	}

	return &exchange.Config{
		Platform:   strings.ToLower(strings.TrimSpace(apiConfig.Platform)),
		ApiKey:     apiKey,
		SecretKey:  secretKey,
		Passphrase: passphrase,
		IsTestnet:  false,
		Proxy:      proxyCfg,
	}, nil
}


