// Package trading
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 HotGo CLI
// @Author  Ms <133814250@qq.com>
// @License  https://github.com/bufanyun/hotgo/blob/master/LICENSE

package trading

import (
	"context"
	"fmt"
	"hotgo/internal/consts"
	"hotgo/internal/dao"
	"hotgo/internal/library/contexts"
	toogoLogic "hotgo/internal/logic/toogo"
	"hotgo/internal/model/do"
	"hotgo/internal/model/entity"
	"hotgo/internal/model/input"
	"hotgo/utility/encrypt"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

type apiConfigImpl struct{}

// List 获取API配置列表
func (s *apiConfigImpl) List(ctx context.Context, in *input.TradingApiConfigListInp) (list []*input.TradingApiConfigListModel, totalCount int, err error) {
	mod := dao.TradingApiConfig.Ctx(ctx)

	// 租户隔离 - 重要！
	memberId := contexts.GetUserId(ctx)
	if memberId <= 0 {
		err = gerror.New("用户未登录")
		return
	}
	mod = mod.Where(dao.TradingApiConfig.Columns().UserId, memberId)

	// 条件筛选
	if in.Platform != "" {
		mod = mod.Where(dao.TradingApiConfig.Columns().Platform, in.Platform)
	}
	if in.Status > 0 {
		mod = mod.Where(dao.TradingApiConfig.Columns().Status, in.Status)
	}
	if in.ApiName != "" {
		mod = mod.WhereLike(dao.TradingApiConfig.Columns().ApiName, "%"+in.ApiName+"%")
	}

	// 软删除过滤
	mod = mod.WhereNull(dao.TradingApiConfig.Columns().DeletedAt)

	totalCount, err = mod.Count()
	if err != nil {
		return nil, 0, err
	}

	if totalCount == 0 {
		return
	}

	err = mod.Page(in.Page, in.PageSize).
		Order(dao.TradingApiConfig.Columns().IsDefault + " DESC").
		Order(dao.TradingApiConfig.Columns().CreatedAt + " DESC").
		Scan(&list)

	if err != nil {
		return nil, 0, err
	}

	// 脱敏处理
	for _, item := range list {
		item.ApiKey = s.MaskApiKey(item.ApiKey)
	}

	return
}

// Create 创建API配置
func (s *apiConfigImpl) Create(ctx context.Context, in *input.TradingApiConfigCreateInp) (id int64, err error) {
	memberId := contexts.GetUserId(ctx)
	if memberId <= 0 {
		return 0, gerror.New("用户未登录")
	}

	// 验证平台
	if !s.ValidatePlatform(in.Platform) {
		return 0, gerror.New("不支持的交易平台")
	}

	// 如果设为默认，先取消其他默认配置
	if in.IsDefault == 1 {
		err = s.CancelOtherDefaults(ctx, memberId, 0)
		if err != nil {
			return 0, err
		}
	}

	// 自动填充BaseUrl
	baseUrl := in.BaseUrl
	if baseUrl == "" {
		baseUrl = s.GetDefaultBaseUrl(in.Platform)
	}

	// 加密敏感信息
	encryptedApiKey, err := encrypt.AesEncrypt(in.ApiKey)
	if err != nil {
		return 0, gerror.Wrap(err, "API Key加密失败")
	}

	encryptedSecretKey, err := encrypt.AesEncrypt(in.SecretKey)
	if err != nil {
		return 0, gerror.Wrap(err, "Secret Key加密失败")
	}

	encryptedPassphrase := ""
	if in.Passphrase != "" {
		encryptedPassphrase, err = encrypt.AesEncrypt(in.Passphrase)
		if err != nil {
			return 0, gerror.Wrap(err, "Passphrase加密失败")
		}
	}

	data := &do.TradingApiConfig{
		UserId:       memberId,
		ApiName:      in.ApiName,
		Platform:     in.Platform,
		BaseUrl:      baseUrl,
		ApiKey:       encryptedApiKey,
		SecretKey:    encryptedSecretKey,
		Passphrase:   encryptedPassphrase,
		IsDefault:    in.IsDefault,
		Status:       consts.StatusEnabled,
		VerifyStatus: 0, // 未验证
		Remark:       in.Remark,
	}

	// 【PostgreSQL 兼容】InsertAndGetId() 不支持 PostgreSQL，改用事务 + LASTVAL()
	tx, err := g.DB().Begin(ctx)
	if err != nil {
		return 0, gerror.Wrap(err, "开启事务失败")
	}
	defer tx.Rollback()
	
	_, err = tx.Model("hg_trading_api_config").Ctx(ctx).Data(data).Insert()
	if err != nil {
		return 0, gerror.Wrap(err, "创建API配置失败")
	}
	
	val, err := tx.GetValue("SELECT LASTVAL()")
	if err != nil {
		return 0, gerror.Wrap(err, "获取API配置ID失败")
	}
	id = val.Int64()
	
	err = tx.Commit()
	if err != nil {
		return 0, gerror.Wrap(err, "提交事务失败")
	}
	
	return id, nil
}

// GetDefaultBaseUrl 根据平台获取默认API地址
func (s *apiConfigImpl) GetDefaultBaseUrl(platform string) string {
	switch platform {
	case "binance":
		return "https://fapi.binance.com"
	case "okx":
		return "https://www.okx.com"
	case "gate":
		return "https://api.gateio.ws"
	default:
		return ""
	}
}

// Update 更新API配置
func (s *apiConfigImpl) Update(ctx context.Context, in *input.TradingApiConfigUpdateInp) error {
	memberId := contexts.GetUserId(ctx)
	if memberId <= 0 {
		return gerror.New("用户未登录")
	}

	// 验证所属
	var config *entity.TradingApiConfig
	err := dao.TradingApiConfig.Ctx(ctx).
		Where(dao.TradingApiConfig.Columns().Id, in.Id).
		Where(dao.TradingApiConfig.Columns().UserId, memberId).
		WhereNull(dao.TradingApiConfig.Columns().DeletedAt).
		Scan(&config)

	if err != nil {
		return err
	}
	if config == nil {
		return gerror.New("配置不存在或无权限")
	}

	// 如果设为默认，先取消其他默认配置
	if in.IsDefault == 1 {
		err = s.CancelOtherDefaults(ctx, memberId, in.Id)
		if err != nil {
			return err
		}
	}

	data := g.Map{
		dao.TradingApiConfig.Columns().ApiName:   in.ApiName,
		dao.TradingApiConfig.Columns().Platform:  in.Platform,
		dao.TradingApiConfig.Columns().BaseUrl:   in.BaseUrl,
		dao.TradingApiConfig.Columns().IsDefault: in.IsDefault,
		dao.TradingApiConfig.Columns().Remark:    in.Remark,
	}

	// status=0 代表“未传/不修改”，避免把数据库 status 覆盖成 0 导致前端显示为“禁用”
	// 仅允许 1=启用 / 2=禁用
	if in.Status == consts.StatusEnabled || in.Status == consts.StatusDisable {
		data[dao.TradingApiConfig.Columns().Status] = in.Status
	}

	// 如果提供了新的密钥，则更新（加密）
	if in.ApiKey != "" {
		encryptedApiKey, err := encrypt.AesEncrypt(in.ApiKey)
		if err != nil {
			return gerror.Wrap(err, "API Key加密失败")
		}
		data[dao.TradingApiConfig.Columns().ApiKey] = encryptedApiKey
		data[dao.TradingApiConfig.Columns().VerifyStatus] = 0 // 重新验证
	}

	if in.SecretKey != "" {
		encryptedSecretKey, err := encrypt.AesEncrypt(in.SecretKey)
		if err != nil {
			return gerror.Wrap(err, "Secret Key加密失败")
		}
		data[dao.TradingApiConfig.Columns().SecretKey] = encryptedSecretKey
		data[dao.TradingApiConfig.Columns().VerifyStatus] = 0 // 重新验证
	}

	if in.Passphrase != "" {
		encryptedPassphrase, err := encrypt.AesEncrypt(in.Passphrase)
		if err != nil {
			return gerror.Wrap(err, "Passphrase加密失败")
		}
		data[dao.TradingApiConfig.Columns().Passphrase] = encryptedPassphrase
		data[dao.TradingApiConfig.Columns().VerifyStatus] = 0 // 重新验证
	}

	_, err = dao.TradingApiConfig.Ctx(ctx).
		Where(dao.TradingApiConfig.Columns().Id, in.Id).
		Data(data).
		Update()

	return err
}

// Delete 删除API配置（软删除）
func (s *apiConfigImpl) Delete(ctx context.Context, in *input.TradingApiConfigDeleteInp) error {
	memberId := contexts.GetUserId(ctx)
	if memberId <= 0 {
		return gerror.New("用户未登录")
	}

	// 检查是否有机器人正在使用
	count, err := dao.TradingRobot.Ctx(ctx).
		Where(dao.TradingRobot.Columns().ApiConfigId, in.Id).
		Where(dao.TradingRobot.Columns().UserId, memberId).
		WhereNull(dao.TradingRobot.Columns().DeletedAt).
		Count()

	if err != nil {
		return err
	}
	if count > 0 {
		return gerror.Newf("该配置正被%d个机器人使用，无法删除", count)
	}

	// 软删除
	_, err = dao.TradingApiConfig.Ctx(ctx).
		Where(dao.TradingApiConfig.Columns().Id, in.Id).
		Where(dao.TradingApiConfig.Columns().UserId, memberId).
		Data(g.Map{
			dao.TradingApiConfig.Columns().DeletedAt: gtime.Now(),
		}).
		Update()

	return err
}

// View 查看详情
func (s *apiConfigImpl) View(ctx context.Context, in *input.TradingApiConfigViewInp) (out *input.TradingApiConfigViewModel, err error) {
	memberId := contexts.GetUserId(ctx)
	if memberId <= 0 {
		return nil, gerror.New("用户未登录")
	}

	err = dao.TradingApiConfig.Ctx(ctx).
		Where(dao.TradingApiConfig.Columns().Id, in.Id).
		Where(dao.TradingApiConfig.Columns().UserId, memberId).
		WhereNull(dao.TradingApiConfig.Columns().DeletedAt).
		Scan(&out)

	if err != nil {
		return nil, err
	}
	if out == nil {
		return nil, gerror.New("配置不存在")
	}

	// 脱敏
	out.ApiKey = s.MaskApiKey(out.ApiKey)
	out.SecretKey = "********************"
	out.Passphrase = "********************"

	return
}

// Test 测试API连接
func (s *apiConfigImpl) Test(ctx context.Context, in *input.TradingApiConfigTestInp) (out *input.TradingApiConfigTestModel, err error) {
	memberId := contexts.GetUserId(ctx)
	if memberId <= 0 {
		return nil, gerror.New("用户未登录")
	}

	// 获取配置
	var config *entity.TradingApiConfig
	err = dao.TradingApiConfig.Ctx(ctx).
		Where(dao.TradingApiConfig.Columns().Id, in.Id).
		Where(dao.TradingApiConfig.Columns().UserId, memberId).
		WhereNull(dao.TradingApiConfig.Columns().DeletedAt).
		Scan(&config)

	if err != nil {
		return nil, err
	}
	if config == nil {
		return nil, gerror.New("配置不存在")
	}

	startTime := time.Now()

	// 统一平台口径
	config.Platform = strings.ToLower(strings.TrimSpace(config.Platform))
	// 清除 toogo 侧的交易所实例缓存，确保使用最新配置
	toogoLogic.GetExchangeManager().RemoveExchange(config.Id)

	// 获取交易所实例（使用 internal/library/exchange 实现，支持 gate/binance/okx/bitget）
	ex, err := toogoLogic.GetExchangeManager().GetExchangeFromConfig(ctx, config)
	if err != nil {
		latency := int(time.Since(startTime).Milliseconds())
		out = &input.TradingApiConfigTestModel{
			Success: false,
			Message: "获取交易所实例失败: " + err.Error(),
			Balance: "",
			Latency: latency,
		}
		// 更新验证状态为失败
		_, _ = dao.TradingApiConfig.Ctx(ctx).
			Where(dao.TradingApiConfig.Columns().Id, in.Id).
			Data(g.Map{
				dao.TradingApiConfig.Columns().LastVerifyTime: gtime.Now(),
				dao.TradingApiConfig.Columns().VerifyStatus:   2,
				dao.TradingApiConfig.Columns().VerifyMessage:  out.Message,
			}).
			Update()
		return out, nil
	}

	// 调用交易所API测试连接：获取余额
	bal, err := ex.GetBalance(ctx)
	latency := int(time.Since(startTime).Milliseconds())

	if err != nil {
		// Gate 特殊情况：合约账户未创建时会返回 USER_NOT_FOUND，但 API Key 本身可能是有效的
		// 提示用户先划转资金以初始化合约账户，避免“测试失败”造成误解。
		if config.Platform == "gate" && strings.Contains(err.Error(), "USER_NOT_FOUND") {
			out = &input.TradingApiConfigTestModel{
				Success: true,
				Message: "连接成功，但 Gate 合约账户未创建：请先从现货账户划转任意资金到合约账户（USDT Futures）以创建合约账户",
				Balance: "0.0000 USDT",
				Latency: latency,
			}

			// 更新验证状态为成功（但带提示信息）
			_, _ = dao.TradingApiConfig.Ctx(ctx).
				Where(dao.TradingApiConfig.Columns().Id, in.Id).
				Data(g.Map{
					dao.TradingApiConfig.Columns().LastVerifyTime: gtime.Now(),
					dao.TradingApiConfig.Columns().VerifyStatus:   1,
					dao.TradingApiConfig.Columns().VerifyMessage:  out.Message,
				}).
				Update()
			return out, nil
		}

		out = &input.TradingApiConfigTestModel{
			Success: false,
			Message: "API连接失败: " + err.Error(),
			Balance: "",
			Latency: latency,
		}
		// 更新验证状态为失败
		_, _ = dao.TradingApiConfig.Ctx(ctx).
			Where(dao.TradingApiConfig.Columns().Id, in.Id).
			Data(g.Map{
				dao.TradingApiConfig.Columns().LastVerifyTime: gtime.Now(),
				dao.TradingApiConfig.Columns().VerifyStatus:   2,
				dao.TradingApiConfig.Columns().VerifyMessage:  out.Message,
			}).
			Update()
		return out, nil
	}

	balanceStr := fmt.Sprintf("%.4f USDT", bal.AvailableBalance)

	// 连接成功
	out = &input.TradingApiConfigTestModel{
		Success: true,
		Message: "连接成功",
		Balance: balanceStr,
		Latency: latency,
	}

	// 更新验证状态为成功
	_, err = dao.TradingApiConfig.Ctx(ctx).
		Where(dao.TradingApiConfig.Columns().Id, in.Id).
		Data(g.Map{
			dao.TradingApiConfig.Columns().LastVerifyTime: gtime.Now(),
			dao.TradingApiConfig.Columns().VerifyStatus:   1,
			dao.TradingApiConfig.Columns().VerifyMessage:  out.Message,
		}).
		Update()

	g.Log().Infof(ctx, "API测试成功: platform=%s, balance=%s, latency=%dms", 
		config.Platform, balanceStr, latency)

	return
}

// SetDefault 设为默认
func (s *apiConfigImpl) SetDefault(ctx context.Context, in *input.TradingApiConfigSetDefaultInp) error {
	memberId := contexts.GetUserId(ctx)
	if memberId <= 0 {
		return gerror.New("用户未登录")
	}

	// 验证配置是否存在且属于当前用户
	var config *entity.TradingApiConfig
	err := dao.TradingApiConfig.Ctx(ctx).
		Where(dao.TradingApiConfig.Columns().Id, in.Id).
		Where(dao.TradingApiConfig.Columns().UserId, memberId).
		WhereNull(dao.TradingApiConfig.Columns().DeletedAt).
		Scan(&config)

	if err != nil {
		return err
	}
	if config == nil {
		return gerror.New("配置不存在或无权限")
	}

	// 取消其他默认配置
	err = s.CancelOtherDefaults(ctx, memberId, in.Id)
	if err != nil {
		return err
	}

	// 设置为默认
	_, err = dao.TradingApiConfig.Ctx(ctx).
		Where(dao.TradingApiConfig.Columns().Id, in.Id).
		Data(g.Map{
			dao.TradingApiConfig.Columns().IsDefault: 1,
		}).
		Update()

	return err
}

// GetPlatforms 获取支持的平台列表
func (s *apiConfigImpl) GetPlatforms(ctx context.Context) (list []*input.TradingApiConfigPlatformsModel, err error) {
	list = []*input.TradingApiConfigPlatformsModel{
		{
			Value:    "binance",
			Label:    "Binance（币安）",
			BaseUrl:  "https://api.binance.com",
			NeedPass: false,
		},
		{
			Value:    "okx",
			Label:    "OKX（欧易）",
			BaseUrl:  "https://www.okx.com",
			NeedPass: true,
		},
		{
			Value:    "gate",
			Label:    "Gate.io",
			BaseUrl:  "https://api.gateio.ws",
			NeedPass: false,
		},
	}
	return
}

// MaskApiKey API Key脱敏显示
func (s *apiConfigImpl) MaskApiKey(apiKey string) string {
	if apiKey == "" {
		return ""
	}
	if len(apiKey) <= 8 {
		return "****"
	}
	return apiKey[:4] + "****************" + apiKey[len(apiKey)-4:]
}

// ValidatePlatform 验证平台是否支持
func (s *apiConfigImpl) ValidatePlatform(platform string) bool {
	validPlatforms := []string{"binance", "okx", "gate"}
	for _, p := range validPlatforms {
		if p == platform {
			return true
		}
	}
	return false
}

// CancelOtherDefaults 取消其他默认配置
func (s *apiConfigImpl) CancelOtherDefaults(ctx context.Context, userId int64, exceptId int64) error {
	mod := dao.TradingApiConfig.Ctx(ctx).
		Where(dao.TradingApiConfig.Columns().UserId, userId).
		Where(dao.TradingApiConfig.Columns().IsDefault, 1).
		WhereNull(dao.TradingApiConfig.Columns().DeletedAt)

	if exceptId > 0 {
		mod = mod.WhereNot(dao.TradingApiConfig.Columns().Id, exceptId)
	}

	_, err := mod.Data(g.Map{
		dao.TradingApiConfig.Columns().IsDefault: 0,
	}).Update()

	return err
}

