// Package trading
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 HotGo CLI
// @Author  Ms <133814250@qq.com>
// @License  https://github.com/bufanyun/hotgo/blob/master/LICENSE

package trading

import (
	"context"
	"fmt"
	"hotgo/internal/dao"
	"hotgo/internal/model/entity"
	"hotgo/internal/model/input"
	"hotgo/utility/encrypt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"golang.org/x/net/proxy"
)

type proxyConfigImpl struct{}

// Get 获取代理配置（全局配置，user_id=0表示全局）
func (s *proxyConfigImpl) Get(ctx context.Context, in *input.TradingProxyConfigGetInp) (out *input.TradingProxyConfigModel, err error) {
	var config *entity.TradingProxyConfig
	err = dao.TradingProxyConfig.Ctx(ctx).
		Where(dao.TradingProxyConfig.Columns().UserId, 0).   // user_id=0 表示全局配置
		Where(dao.TradingProxyConfig.Columns().TenantId, 0). // tenant_id=0 表示全局配置
		Scan(&config)

	if err != nil {
		return nil, err
	}

	// 如果没有配置，返回默认值
	if config == nil {
		out = &input.TradingProxyConfigModel{
			Enabled:      0,
			ProxyType:    "http",            // HTTP代理已验证可用，使用HTTP作为默认
			ProxyAddress: "127.0.0.1:33210", // HTTP代理端口（已验证可用）
			AuthEnabled:  0,
		}
		return
	}

	// 转换为输出模型
	out = &input.TradingProxyConfigModel{
		Id:           config.Id,
		Enabled:      config.Enabled,
		ProxyType:    config.ProxyType,
		ProxyAddress: config.ProxyAddress,
		AuthEnabled:  config.AuthEnabled,
		Username:     config.Username,
		LastTestTime: config.LastTestTime,
		TestStatus:   config.TestStatus,
		TestMessage:  config.TestMessage,
		CreatedAt:    config.CreatedAt,
		UpdatedAt:    config.UpdatedAt,
	}

	return
}

// Save 保存代理配置（全局配置，user_id=0表示全局）
func (s *proxyConfigImpl) Save(ctx context.Context, in *input.TradingProxyConfigSaveInp) error {
	// 加密密码
	encryptedPassword := ""
	if in.Password != "" {
		var err error
		encryptedPassword, err = encrypt.AesEncrypt(in.Password)
		if err != nil {
			return gerror.Wrap(err, "密码加密失败")
		}
	}

	// 检查是否已存在全局配置（user_id=0 且 tenant_id=0 表示全局配置）
	var existConfig *entity.TradingProxyConfig
	err := dao.TradingProxyConfig.Ctx(ctx).
		Where(dao.TradingProxyConfig.Columns().UserId, 0).   // user_id=0 表示全局配置
		Where(dao.TradingProxyConfig.Columns().TenantId, 0). // tenant_id=0 表示全局配置
		Scan(&existConfig)

	if err != nil {
		return err
	}

	data := g.Map{
		dao.TradingProxyConfig.Columns().Enabled:      in.Enabled,
		dao.TradingProxyConfig.Columns().ProxyType:    in.ProxyType,
		dao.TradingProxyConfig.Columns().ProxyAddress: in.ProxyAddress,
		dao.TradingProxyConfig.Columns().AuthEnabled:  in.AuthEnabled,
		dao.TradingProxyConfig.Columns().Username:     in.Username,
	}

	// 只有提供了密码才更新
	if in.Password != "" {
		data[dao.TradingProxyConfig.Columns().Password] = encryptedPassword
	}

	if existConfig == nil {
		// 创建新配置（全局配置，user_id=0）
		data[dao.TradingProxyConfig.Columns().UserId] = 0
		data[dao.TradingProxyConfig.Columns().TenantId] = 0
		_, err = dao.TradingProxyConfig.Ctx(ctx).Data(data).Insert()
		if err != nil {
			g.Log().Errorf(ctx, "[ProxyConfig] 创建全局代理配置失败: %v, data: %+v", err, data)
			return gerror.Wrap(err, "创建代理配置失败")
		}
		g.Log().Infof(ctx, "[ProxyConfig] 创建全局代理配置成功: proxyType=%s, proxyAddress=%s", in.ProxyType, in.ProxyAddress)
	} else {
		// 更新现有配置
		_, err = dao.TradingProxyConfig.Ctx(ctx).
			Where(dao.TradingProxyConfig.Columns().UserId, 0).
			Where(dao.TradingProxyConfig.Columns().TenantId, 0).
			Data(data).
			Update()
		if err != nil {
			g.Log().Errorf(ctx, "[ProxyConfig] 更新全局代理配置失败: %v, data: %+v", err, data)
			return gerror.Wrap(err, "更新代理配置失败")
		}
		g.Log().Infof(ctx, "[ProxyConfig] 更新全局代理配置成功: proxyType=%s, proxyAddress=%s", in.ProxyType, in.ProxyAddress)
	}

	return nil
}

// Test 测试代理连接
func (s *proxyConfigImpl) Test(ctx context.Context, in *input.TradingProxyConfigTestInp) (out *input.TradingProxyConfigTestModel, err error) {
	startTime := time.Now()

	// 测试代理连接
	success, message, externalIP := s.testProxyConnection(in.ProxyType, in.ProxyAddress, in.AuthEnabled, in.Username, in.Password)

	latency := int(time.Since(startTime).Milliseconds())

	out = &input.TradingProxyConfigTestModel{
		Success:    success,
		Message:    message,
		ExternalIP: externalIP,
		Latency:    latency,
	}

	// 更新全局配置的测试结果
	testStatus := 2 // 失败
	if success {
		testStatus = 1 // 成功
	}

	_, _ = dao.TradingProxyConfig.Ctx(ctx).
		Where(dao.TradingProxyConfig.Columns().UserId, 0).   // user_id=0 表示全局配置
		Where(dao.TradingProxyConfig.Columns().TenantId, 0). // tenant_id=0 表示全局配置
		Data(g.Map{
			dao.TradingProxyConfig.Columns().LastTestTime: gtime.Now(),
			dao.TradingProxyConfig.Columns().TestStatus:   testStatus,
			dao.TradingProxyConfig.Columns().TestMessage:  message,
		}).
		Update()

	return
}

// Toggle 切换启用状态（全局配置）
func (s *proxyConfigImpl) Toggle(ctx context.Context, in *input.TradingProxyConfigToggleInp) error {
	// 检查全局配置是否存在
	count, err := dao.TradingProxyConfig.Ctx(ctx).
		Where(dao.TradingProxyConfig.Columns().UserId, 0).   // user_id=0 表示全局配置
		Where(dao.TradingProxyConfig.Columns().TenantId, 0). // tenant_id=0 表示全局配置
		Count()

	if err != nil {
		return err
	}

	if count == 0 {
		return gerror.New("请先保存代理配置")
	}

	// 更新启用状态
	_, err = dao.TradingProxyConfig.Ctx(ctx).
		Where(dao.TradingProxyConfig.Columns().UserId, 0).
		Where(dao.TradingProxyConfig.Columns().TenantId, 0).
		Data(g.Map{
			dao.TradingProxyConfig.Columns().Enabled: in.Enabled,
		}).
		Update()

	return err
}

// testProxyConnection 测试代理连接
func (s *proxyConfigImpl) testProxyConnection(proxyType, proxyAddress string, authEnabled int, username, password string) (bool, string, string) {
	// 构建代理URL
	var proxyURL *url.URL
	var err error

	if authEnabled == 1 && username != "" {
		// 需要认证
		if password != "" {
			proxyURL, err = url.Parse(fmt.Sprintf("%s://%s:%s@%s", proxyType, username, password, proxyAddress))
		} else {
			proxyURL, err = url.Parse(fmt.Sprintf("%s://%s@%s", proxyType, username, proxyAddress))
		}
	} else {
		// 不需要认证
		proxyURL, err = url.Parse(fmt.Sprintf("%s://%s", proxyType, proxyAddress))
	}

	if err != nil {
		return false, "代理地址格式错误: " + err.Error(), ""
	}

	// 测试连接
	var externalIP string
	var testErr error

	if proxyType == "socks5" {
		externalIP, testErr = s.testSocks5Proxy(proxyURL)
	} else {
		externalIP, testErr = s.testHTTPProxy(proxyURL)
	}

	if testErr != nil {
		return false, "连接失败: " + testErr.Error(), ""
	}

	return true, "连接成功", externalIP
}

// testSocks5Proxy 测试SOCKS5代理
func (s *proxyConfigImpl) testSocks5Proxy(proxyURL *url.URL) (string, error) {
	// 创建SOCKS5拨号器
	var auth *proxy.Auth
	if proxyURL.User != nil {
		password, _ := proxyURL.User.Password()
		auth = &proxy.Auth{
			User:     proxyURL.User.Username(),
			Password: password,
		}
	}

	dialer, err := proxy.SOCKS5("tcp", proxyURL.Host, auth, proxy.Direct)
	if err != nil {
		return "", gerror.Wrapf(err, "创建SOCKS5拨号器失败: %s", proxyURL.Host)
	}

	// 创建HTTP客户端
	httpClient := &http.Client{
		Transport: &http.Transport{
			Dial: dialer.Dial,
		},
		Timeout: 15 * time.Second,
	}

	// 测试连接（获取外网IP）
	resp, err := httpClient.Get("https://api.ipify.org?format=text")
	if err != nil {
		// 检查是否是协议版本错误，可能是代理类型配置错误
		errMsg := err.Error()
		if strings.Contains(errMsg, "unexpected protocol version") {
			return "", gerror.Newf("SOCKS5协议错误: 代理服务器 %s 可能不是SOCKS5代理。\n"+
				"可能的原因：\n"+
				"1. 该端口实际上是HTTP代理（请尝试使用HTTP代理类型）\n"+
				"2. 代理服务器未运行或端口错误\n"+
				"3. 代理服务器不支持SOCKS5协议\n\n"+
				"建议：如果HTTP代理（127.0.0.1:33210）可用，请使用HTTP代理类型", proxyURL.Host)
		}
		if strings.Contains(errMsg, "connection refused") {
			return "", gerror.Newf("连接被拒绝: 无法连接到代理服务器 %s\n"+
				"请检查：\n"+
				"1. 代理服务器是否正在运行\n"+
				"2. 端口号是否正确\n"+
				"3. 防火墙是否允许连接", proxyURL.Host)
		}
		if strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "deadline exceeded") {
			return "", gerror.Newf("连接超时: 代理服务器 %s 无响应\n"+
				"请检查：\n"+
				"1. 代理地址和端口是否正确\n"+
				"2. 网络连接是否正常\n"+
				"3. 代理服务器是否正常运行", proxyURL.Host)
		}
		return "", gerror.Wrapf(err, "通过SOCKS5代理连接失败: %s\n错误详情: %s", proxyURL.Host, errMsg)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", gerror.Newf("HTTP错误: %d", resp.StatusCode)
	}

	// 读取外网IP
	var ip string
	_, err = fmt.Fscanf(resp.Body, "%s", &ip)
	if err != nil {
		return "", gerror.Wrap(err, "读取响应失败")
	}

	return ip, nil
}

// testHTTPProxy 测试HTTP代理
func (s *proxyConfigImpl) testHTTPProxy(proxyURL *url.URL) (string, error) {
	// 创建HTTP客户端
	httpClient := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
		Timeout: 10 * time.Second,
	}

	// 测试连接（获取外网IP）
	resp, err := httpClient.Get("https://api.ipify.org?format=text")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 读取外网IP
	var ip string
	_, err = fmt.Fscanf(resp.Body, "%s", &ip)
	if err != nil {
		return "", err
	}

	return ip, nil
}

// GetProxyDialer 获取代理拨号器（供其他模块使用，使用全局配置）
func (s *proxyConfigImpl) GetProxyDialer(ctx context.Context) (proxy.Dialer, error) {
	// 获取全局代理配置
	var config *entity.TradingProxyConfig
	err := dao.TradingProxyConfig.Ctx(ctx).
		Where(dao.TradingProxyConfig.Columns().UserId, 0).   // user_id=0 表示全局配置
		Where(dao.TradingProxyConfig.Columns().TenantId, 0). // tenant_id=0 表示全局配置
		Where(dao.TradingProxyConfig.Columns().Enabled, 1).
		Scan(&config)

	if err != nil {
		return nil, err
	}

	// 如果没有启用代理，返回直连
	if config == nil {
		return proxy.Direct, nil
	}

	// 如果是HTTP代理，不支持Dialer方式
	if config.ProxyType == "http" {
		return nil, gerror.New("HTTP代理不支持Dialer方式，请使用GetProxyTransport")
	}

	// 解密密码
	var auth *proxy.Auth
	if config.AuthEnabled == 1 && config.Username != "" {
		password := ""
		if config.Password != "" {
			password, err = encrypt.AesDecrypt(config.Password)
			if err != nil {
				return nil, gerror.Wrap(err, "密码解密失败")
			}
		}

		auth = &proxy.Auth{
			User:     config.Username,
			Password: password,
		}
	}

	// 创建SOCKS5拨号器
	dialer, err := proxy.SOCKS5("tcp", config.ProxyAddress, auth, proxy.Direct)
	if err != nil {
		return nil, err
	}

	return dialer, nil
}

// GetWebSocketDialer 获取WebSocket代理拨号器（支持HTTP和SOCKS5代理）
func (s *proxyConfigImpl) GetWebSocketDialer(ctx context.Context) (func(network, addr string) (net.Conn, error), error) {
	// 获取全局代理配置
	var config *entity.TradingProxyConfig
	err := dao.TradingProxyConfig.Ctx(ctx).
		Where(dao.TradingProxyConfig.Columns().UserId, 0).   // user_id=0 表示全局配置
		Where(dao.TradingProxyConfig.Columns().TenantId, 0). // tenant_id=0 表示全局配置
		Where(dao.TradingProxyConfig.Columns().Enabled, 1).
		Scan(&config)

	if err != nil {
		return nil, err
	}

	// 如果没有启用代理，返回nil（使用默认拨号器）
	if config == nil {
		return nil, nil
	}

	// 解密密码
	password := ""
	if config.AuthEnabled == 1 && config.Password != "" {
		password, err = encrypt.AesDecrypt(config.Password)
		if err != nil {
			return nil, gerror.Wrap(err, "密码解密失败")
		}
	}

	// 根据代理类型创建拨号器
	if config.ProxyType == "http" || config.ProxyType == "https" {
		// HTTP代理：使用CONNECT方法建立隧道
		return func(network, addr string) (net.Conn, error) {
			// 连接到代理服务器
			proxyConn, err := net.DialTimeout("tcp", config.ProxyAddress, 30*time.Second)
			if err != nil {
				return nil, fmt.Errorf("连接HTTP代理失败: %v", err)
			}

			// 构建CONNECT请求
			connectReq := fmt.Sprintf("CONNECT %s HTTP/1.1\r\nHost: %s\r\n", addr, addr)

			// 如果需要认证
			if config.AuthEnabled == 1 && config.Username != "" {
				auth := config.Username
				if password != "" {
					auth += ":" + password
				}
				encoded := base64Encode([]byte(auth))
				connectReq += fmt.Sprintf("Proxy-Authorization: Basic %s\r\n", encoded)
			}

			connectReq += "\r\n"

			// 发送CONNECT请求
			_, err = proxyConn.Write([]byte(connectReq))
			if err != nil {
				proxyConn.Close()
				return nil, fmt.Errorf("发送CONNECT请求失败: %v", err)
			}

			// 读取响应
			response := make([]byte, 1024)
			n, err := proxyConn.Read(response)
			if err != nil {
				proxyConn.Close()
				return nil, fmt.Errorf("读取代理响应失败: %v", err)
			}

			// 检查响应状态
			respStr := string(response[:n])
			if !strings.Contains(respStr, "200") {
				proxyConn.Close()
				return nil, fmt.Errorf("HTTP代理CONNECT失败: %s", respStr)
			}

			return proxyConn, nil
		}, nil
	} else if config.ProxyType == "socks5" {
		// SOCKS5代理
		var auth *proxy.Auth
		if config.AuthEnabled == 1 && config.Username != "" {
			auth = &proxy.Auth{
				User:     config.Username,
				Password: password,
			}
		}

		dialer, err := proxy.SOCKS5("tcp", config.ProxyAddress, auth, proxy.Direct)
		if err != nil {
			return nil, err
		}

		return dialer.Dial, nil
	}

	return nil, gerror.Newf("不支持的代理类型: %s", config.ProxyType)
}

// base64Encode 简单的base64编码
func base64Encode(data []byte) string {
	const base64Chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	var result strings.Builder

	for i := 0; i < len(data); i += 3 {
		var n uint32
		var padding int

		n = uint32(data[i]) << 16
		if i+1 < len(data) {
			n |= uint32(data[i+1]) << 8
		} else {
			padding++
		}
		if i+2 < len(data) {
			n |= uint32(data[i+2])
		} else {
			padding++
		}

		result.WriteByte(base64Chars[(n>>18)&0x3F])
		result.WriteByte(base64Chars[(n>>12)&0x3F])
		if padding < 2 {
			result.WriteByte(base64Chars[(n>>6)&0x3F])
		} else {
			result.WriteByte('=')
		}
		if padding < 1 {
			result.WriteByte(base64Chars[n&0x3F])
		} else {
			result.WriteByte('=')
		}
	}

	return result.String()
}

// GetProxyTransport 获取代理传输层（供HTTP客户端使用，使用全局配置）
func (s *proxyConfigImpl) GetProxyTransport(ctx context.Context) (*http.Transport, error) {
	// 获取全局代理配置
	var config *entity.TradingProxyConfig
	err := dao.TradingProxyConfig.Ctx(ctx).
		Where(dao.TradingProxyConfig.Columns().UserId, 0).   // user_id=0 表示全局配置
		Where(dao.TradingProxyConfig.Columns().TenantId, 0). // tenant_id=0 表示全局配置
		Where(dao.TradingProxyConfig.Columns().Enabled, 1).
		Scan(&config)

	if err != nil {
		return nil, err
	}

	// 如果没有启用代理，返回默认传输层
	if config == nil {
		return &http.Transport{}, nil
	}

	// 构建代理URL
	var proxyURL *url.URL
	if config.AuthEnabled == 1 && config.Username != "" {
		password := ""
		if config.Password != "" {
			password, err = encrypt.AesDecrypt(config.Password)
			if err != nil {
				return nil, gerror.Wrap(err, "密码解密失败")
			}
		}

		if password != "" {
			proxyURL, _ = url.Parse(fmt.Sprintf("%s://%s:%s@%s", config.ProxyType, config.Username, password, config.ProxyAddress))
		} else {
			proxyURL, _ = url.Parse(fmt.Sprintf("%s://%s@%s", config.ProxyType, config.Username, config.ProxyAddress))
		}
	} else {
		proxyURL, _ = url.Parse(fmt.Sprintf("%s://%s", config.ProxyType, config.ProxyAddress))
	}

	// 创建传输层
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	}

	return transport, nil
}
