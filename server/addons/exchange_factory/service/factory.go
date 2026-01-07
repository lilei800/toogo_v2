// Package service
// 交易所工厂：用于 trading 模块按 API 配置创建交易所实例（Binance/OKX）
package service

import (
	"context"
	"net"
	"net/http"
	"net/url"

	"hotgo/addons/exchange"
	binanceService "hotgo/addons/exchange_binance/service"
	okxService "hotgo/addons/exchange_okx/service"
	"hotgo/internal/dao"
	"hotgo/internal/model/entity"
	"hotgo/utility/encrypt"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"golang.org/x/net/proxy"
)

// ExchangeFactory 交易所工厂
type ExchangeFactory struct{}

// NewExchangeFactory 创建交易所工厂
func NewExchangeFactory() *ExchangeFactory {
	return &ExchangeFactory{}
}

// CreateExchange 创建交易所实例
func (f *ExchangeFactory) CreateExchange(ctx context.Context, apiConfigId int64, userId int64, tenantId int64) (exchange.IExchange, error) {
	// 获取API配置
	var apiConfig *entity.TradingApiConfig
	err := dao.TradingApiConfig.Ctx(ctx).
		Where(dao.TradingApiConfig.Columns().Id, apiConfigId).
		Where(dao.TradingApiConfig.Columns().UserId, userId).
		Where(dao.TradingApiConfig.Columns().TenantId, tenantId).
		WhereNull(dao.TradingApiConfig.Columns().DeletedAt).
		Scan(&apiConfig)
	if err != nil {
		return nil, err
	}
	if apiConfig == nil {
		return nil, gerror.New("API配置不存在或无权限")
	}

	// 检查状态（允许 status=0 的情况，视为未设置，兼容旧数据）
	if apiConfig.Status == 2 {
		return nil, gerror.New("API配置已禁用")
	}
	// 如果status=0（未设置），自动更新为正常状态
	if apiConfig.Status == 0 {
		_, _ = dao.TradingApiConfig.Ctx(ctx).
			Where(dao.TradingApiConfig.Columns().Id, apiConfigId).
			Data(g.Map{dao.TradingApiConfig.Columns().Status: 1}).
			Update()
		g.Log().Infof(ctx, "自动修复API配置状态: apiConfigId=%d", apiConfigId)
	}

	// 解密API密钥
	apiKey, err := encrypt.AesDecrypt(apiConfig.ApiKey)
	if err != nil {
		g.Log().Errorf(ctx, "[ExchangeFactory] 解密API Key失败: apiConfigId=%d, error=%v", apiConfigId, err)
		return nil, gerror.Wrap(err, "解密API Key失败，请检查密钥是否正确")
	}
	if apiKey == "" {
		g.Log().Errorf(ctx, "[ExchangeFactory] API Key解密后为空: apiConfigId=%d", apiConfigId)
		return nil, gerror.New("API Key解密后为空")
	}

	secretKey, err := encrypt.AesDecrypt(apiConfig.SecretKey)
	if err != nil {
		g.Log().Errorf(ctx, "[ExchangeFactory] 解密Secret Key失败: apiConfigId=%d, error=%v", apiConfigId, err)
		return nil, gerror.Wrap(err, "解密Secret Key失败，请检查密钥是否正确")
	}
	if secretKey == "" {
		g.Log().Errorf(ctx, "[ExchangeFactory] Secret Key解密后为空: apiConfigId=%d", apiConfigId)
		return nil, gerror.New("Secret Key解密后为空")
	}

	var passphrase string
	if apiConfig.Passphrase != "" {
		passphrase, err = encrypt.AesDecrypt(apiConfig.Passphrase)
		if err != nil {
			g.Log().Errorf(ctx, "[ExchangeFactory] 解密Passphrase失败: apiConfigId=%d, error=%v", apiConfigId, err)
			return nil, gerror.Wrap(err, "解密Passphrase失败，请检查密钥是否正确")
		}
	}

	g.Log().Infof(ctx, "[ExchangeFactory] API密钥解密成功: platform=%s, apiConfigId=%d, apiKey长度=%d",
		apiConfig.Platform, apiConfigId, len(apiKey))

	// 获取全局代理配置（不再使用用户ID，使用全局配置）
	proxyTransport, err := f.getProxyTransport(ctx)
	if err != nil {
		g.Log().Warningf(ctx, "[ExchangeFactory] 获取代理配置失败: %v，将使用直连", err)
		proxyTransport = nil
	} else if proxyTransport != nil {
		g.Log().Infof(ctx, "[ExchangeFactory] 已配置全局代理，将通过代理连接%s API", apiConfig.Platform)
	} else {
		g.Log().Infof(ctx, "[ExchangeFactory] 未配置代理，将直连%s API（如连接失败，请配置代理）", apiConfig.Platform)
	}

	// 根据平台创建对应的交易所实例
	switch apiConfig.Platform {
	case "binance":
		g.Log().Infof(ctx, "[ExchangeFactory] 创建Binance交易所实例: apiConfigId=%d", apiConfigId)
		return binanceService.NewBinanceExchange(apiKey, secretKey, proxyTransport), nil
	case "okx":
		g.Log().Infof(ctx, "[ExchangeFactory] 创建OKX交易所实例: apiConfigId=%d", apiConfigId)
		return okxService.NewOKXExchange(apiKey, secretKey, passphrase, proxyTransport), nil
	default:
		return nil, gerror.Newf("不支持的交易所平台: %s", apiConfig.Platform)
	}
}

// getProxyTransport 获取代理传输配置（使用全局配置，user_id=0）
func (f *ExchangeFactory) getProxyTransport(ctx context.Context) (*http.Transport, error) {
	// 获取全局代理配置（user_id=0 且 tenant_id=0 表示全局配置）
	var proxyConfig *entity.TradingProxyConfig
	err := dao.TradingProxyConfig.Ctx(ctx).
		Where(dao.TradingProxyConfig.Columns().UserId, 0).   // user_id=0 表示全局配置
		Where(dao.TradingProxyConfig.Columns().TenantId, 0). // tenant_id=0 表示全局配置
		Scan(&proxyConfig)
	if err != nil {
		g.Log().Errorf(ctx, "[ExchangeFactory] 查询代理配置失败: %v", err)
		return nil, err
	}

	if proxyConfig == nil {
		g.Log().Warning(ctx, "[ExchangeFactory] 未找到全局代理配置（user_id=0, tenant_id=0）")
		return nil, nil
	}

	if proxyConfig.Enabled != 1 {
		g.Log().Infof(ctx, "[ExchangeFactory] 代理配置已禁用: enabled=%d", proxyConfig.Enabled)
		return nil, nil
	}

	g.Log().Infof(ctx, "[ExchangeFactory] 找到全局代理配置: type=%s, address=%s, authEnabled=%d",
		proxyConfig.ProxyType, proxyConfig.ProxyAddress, proxyConfig.AuthEnabled)

	// 创建Transport
	transport := &http.Transport{}

	// 根据代理类型配置
	if proxyConfig.ProxyType == "socks5" {
		// SOCKS5代理需要使用golang.org/x/net/proxy
		var auth *proxy.Auth
		if proxyConfig.AuthEnabled == 1 && proxyConfig.Username != "" {
			password := proxyConfig.Password
			if password != "" {
				password, err = encrypt.AesDecrypt(password)
				if err != nil {
					return nil, gerror.Wrap(err, "解密代理密码失败")
				}
			}
			auth = &proxy.Auth{
				User:     proxyConfig.Username,
				Password: password,
			}
		}

		// 创建SOCKS5拨号器
		dialer, err := proxy.SOCKS5("tcp", proxyConfig.ProxyAddress, auth, proxy.Direct)
		if err != nil {
			return nil, gerror.Wrapf(err, "创建SOCKS5代理失败: %s", proxyConfig.ProxyAddress)
		}

		// 设置DialContext
		transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			g.Log().Debugf(ctx, "[ExchangeFactory] 通过SOCKS5代理连接: %s -> %s", proxyConfig.ProxyAddress, addr)
			return dialer.Dial(network, addr)
		}

		g.Log().Infof(ctx, "[ExchangeFactory] ✅ 已配置SOCKS5代理: %s (认证: %v)", proxyConfig.ProxyAddress, auth != nil)
	} else {
		// HTTP/HTTPS代理
		var proxyURL *url.URL
		if proxyConfig.AuthEnabled == 1 && proxyConfig.Username != "" {
			// 解密密码
			password := proxyConfig.Password
			if password != "" {
				password, err = encrypt.AesDecrypt(password)
				if err != nil {
					return nil, gerror.Wrap(err, "解密代理密码失败")
				}
			}

			proxyURL, err = url.Parse(proxyConfig.ProxyType + "://" + proxyConfig.Username + ":" + password + "@" + proxyConfig.ProxyAddress)
		} else {
			proxyURL, err = url.Parse(proxyConfig.ProxyType + "://" + proxyConfig.ProxyAddress)
		}

		if err != nil {
			return nil, gerror.Wrap(err, "解析代理地址失败")
		}

		transport.Proxy = http.ProxyURL(proxyURL)
		g.Log().Infof(ctx, "[ExchangeFactory] ✅ 已配置HTTP代理: %s", proxyURL.String())
	}

	return transport, nil
}


