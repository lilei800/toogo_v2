// Package toogo
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 Toogo.Ai
// @Author  Toogo Team
package toogo

import (
	"context"
	"fmt"
	"hotgo/internal/dao"
	"hotgo/internal/library/contexts"
	exlib "hotgo/internal/library/exchange"
	"hotgo/internal/model/entity"
	"hotgo/internal/model/input/toogoin"
	"hotgo/internal/service"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcache"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/grand"
	"golang.org/x/time/rate"
)

type sToogoWallet struct{}

// orderHistoryTradeCache 用于订单历史查询的短TTL缓存，避免同一页/频繁刷新重复打交易所成交接口。
// 只缓存很短时间（秒级），主要用于“并发/短时间重复请求”去重，降低服务器与交易所API压力。
var orderHistoryTradeCache = gcache.New()

// tradeHistoryLimiter 对“交易明细页的成交拉取”做低优先级限流（按 apiConfigId）。
// 目的：避免用户频繁刷新交易明细时占满交易所限频额度，从而影响“机器人持仓刷新”等实时性更高的接口。
// 策略：不等待（Allow 即放行，否则直接放弃本次拉取，靠缓存/DB兜底）。
var (
	tradeHistoryLimiterMu sync.Mutex
	tradeHistoryLimiter   = map[int64]*rate.Limiter{}
)

func getTradeHistoryLimiter(apiId int64) *rate.Limiter {
	tradeHistoryLimiterMu.Lock()
	defer tradeHistoryLimiterMu.Unlock()
	if lim := tradeHistoryLimiter[apiId]; lim != nil {
		return lim
	}
	// 每个 apiConfigId：约 1 req/s，burst=1（尽量温和，不抢占额度）
	lim := rate.NewLimiter(rate.Every(1*time.Second), 1)
	tradeHistoryLimiter[apiId] = lim
	return lim
}

func NewToogoWallet() *sToogoWallet {
	return &sToogoWallet{}
}

func init() {
	service.RegisterToogoWallet(NewToogoWallet())
}

// GetOrCreate 获取或创建用户钱包
func (s *sToogoWallet) GetOrCreate(ctx context.Context, userId int64) (wallet *entity.ToogoWallet, err error) {
	err = dao.ToogoWallet.Ctx(ctx).Where(dao.ToogoWallet.Columns().UserId, userId).Scan(&wallet)
	if err != nil {
		return nil, gerror.Wrap(err, "获取钱包信息失败")
	}

	if wallet == nil {
		// 创建新钱包
		wallet = &entity.ToogoWallet{
			UserId:    userId,
			CreatedAt: gtime.Now(),
			UpdatedAt: gtime.Now(),
		}
		_, err = dao.ToogoWallet.Ctx(ctx).Data(wallet).Insert()
		if err != nil {
			return nil, gerror.Wrap(err, "创建钱包失败")
		}
	}
	return
}

// GetOverview 获取钱包概览
func (s *sToogoWallet) GetOverview(ctx context.Context, in *toogoin.WalletOverviewInp) (res *toogoin.WalletOverviewModel, err error) {
	wallet, err := s.GetOrCreate(ctx, in.UserId)
	if err != nil {
		return nil, err
	}

	res = &toogoin.WalletOverviewModel{
		Balance:           wallet.Balance,
		FrozenBalance:     wallet.FrozenBalance,
		Power:             wallet.Power,
		FrozenPower:       wallet.FrozenPower,
		GiftPower:         wallet.GiftPower,
		Commission:        wallet.Commission,
		FrozenCommission:  wallet.FrozenCommission,
		TotalDeposit:      wallet.TotalDeposit,
		TotalWithdraw:     wallet.TotalWithdraw,
		TotalPowerConsume: wallet.TotalPowerConsume,
		TotalCommission:   wallet.TotalCommission,
		TotalPower:        wallet.Power + wallet.GiftPower,
	}
	return
}

// ChangeBalance 变更账户余额 (核心方法)
func (s *sToogoWallet) ChangeBalance(ctx context.Context, in *toogoin.ChangeBalanceInp) (err error) {
	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 获取钱包
		wallet, err := s.GetOrCreate(ctx, in.UserId)
		if err != nil {
			return err
		}

		var (
			beforeAmount float64
			afterAmount  float64
			updateField  string
		)

		// 根据账户类型处理
		switch in.AccountType {
		case "balance":
			beforeAmount = wallet.Balance
			afterAmount = beforeAmount + in.Amount
			if afterAmount < 0 {
				return gerror.New("余额不足")
			}
			updateField = dao.ToogoWallet.Columns().Balance

		case "power":
			beforeAmount = wallet.Power
			afterAmount = beforeAmount + in.Amount
			// 【优化】允许算力为负数，不检查算力是否足够
			// 平仓后即使算力不足，也会扣除算力（允许负算力）
			updateField = dao.ToogoWallet.Columns().Power

		case "gift_power":
			beforeAmount = wallet.GiftPower
			afterAmount = beforeAmount + in.Amount
			if afterAmount < 0 {
				return gerror.New("积分不足")
			}
			updateField = dao.ToogoWallet.Columns().GiftPower

		case "commission":
			beforeAmount = wallet.Commission
			afterAmount = beforeAmount + in.Amount
			if afterAmount < 0 {
				return gerror.New("佣金不足")
			}
			updateField = dao.ToogoWallet.Columns().Commission

		default:
			return gerror.Newf("不支持的账户类型: %s", in.AccountType)
		}

		// 更新钱包余额
		_, err = dao.ToogoWallet.Ctx(ctx).
			Where(dao.ToogoWallet.Columns().UserId, in.UserId).
			Data(g.Map{updateField: afterAmount}).
			Update()
		if err != nil {
			return gerror.Wrap(err, "更新钱包余额失败")
		}

		// 记录流水
		logData := &entity.ToogoWalletLog{
			UserId:       in.UserId,
			AccountType:  in.AccountType,
			ChangeType:   in.ChangeType,
			ChangeAmount: in.Amount,
			BeforeAmount: beforeAmount,
			AfterAmount:  afterAmount,
			RelatedId:    in.RelatedId,
			RelatedType:  in.RelatedType,
			OrderSn:      in.OrderSn,
			Remark:       in.Remark,
			CreatedAt:    gtime.Now(),
		}
		_, err = dao.ToogoWalletLog.Ctx(ctx).Data(logData).Insert()
		if err != nil {
			return gerror.Wrap(err, "记录账户流水失败")
		}

		return nil
	})
}

// Transfer 账户互转 (余额/佣金 -> 算力)
func (s *sToogoWallet) Transfer(ctx context.Context, in *toogoin.TransferInp) (res *toogoin.TransferModel, err error) {
	// 默认转入算力账户
	if in.ToAccount == "" {
		in.ToAccount = "power"
	}
	// 验证转账方向
	if in.ToAccount != "power" {
		return nil, gerror.New("只能转入算力账户")
	}
	if in.FromAccount != "balance" && in.FromAccount != "commission" {
		return nil, gerror.New("只能从余额账户或佣金账户转出")
	}

	// 获取兑换比率 (1 USDT = 1 算力)
	rate := 1.0
	powerAmount := in.Amount * rate

	// 生成订单号
	orderSn := fmt.Sprintf("TF%s%s", gtime.Now().Format("YmdHis"), grand.S(6))

	err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 扣减来源账户
		err = s.ChangeBalance(ctx, &toogoin.ChangeBalanceInp{
			UserId:      in.UserId,
			AccountType: in.FromAccount,
			ChangeType:  "transfer_out",
			Amount:      -in.Amount,
			OrderSn:     orderSn,
			Remark:      "账户互转-转出",
		})
		if err != nil {
			return err
		}

		// 增加目标账户
		err = s.ChangeBalance(ctx, &toogoin.ChangeBalanceInp{
			UserId:      in.UserId,
			AccountType: in.ToAccount,
			ChangeType:  "transfer_in",
			Amount:      powerAmount,
			OrderSn:     orderSn,
			Remark:      "账户互转-转入",
		})
		if err != nil {
			return err
		}

		// 记录互转订单
		transferData := &entity.ToogoTransfer{
			UserId:      in.UserId,
			OrderSn:     orderSn,
			FromAccount: in.FromAccount,
			ToAccount:   in.ToAccount,
			Amount:      in.Amount,
			PowerAmount: powerAmount,
			Rate:        rate,
			Status:      2, // 已完成
			CreatedAt:   gtime.Now(),
		}
		_, err = dao.ToogoTransfer.Ctx(ctx).Data(transferData).Insert()
		return err
	})

	if err != nil {
		return nil, err
	}

	res = &toogoin.TransferModel{
		OrderSn:     orderSn,
		Amount:      in.Amount,
		PowerAmount: powerAmount,
	}
	return
}

// ConsumePower 消耗算力（已禁用）
// 说明：按产品需求不再对“盈利订单”扣除算力；保留空实现以兼容历史调用。
func (s *sToogoWallet) ConsumePower(ctx context.Context, userId int64, robotId int64, orderId int64, orderSn string, profitAmount float64) error {
	g.Log().Infof(ctx, "[ConsumePower] 已禁用：跳过算力扣除 userId=%d, robotId=%d, orderId=%d, profit=%.4f",
		userId, robotId, orderId, profitAmount)
		return nil
}

// WalletLogList 钱包流水列表
func (s *sToogoWallet) WalletLogList(ctx context.Context, in *toogoin.WalletLogListInp) (list []*toogoin.WalletLogListModel, totalCount int, err error) {
	mod := dao.ToogoWalletLog.Ctx(ctx)
	cols := dao.ToogoWalletLog.Columns()

	if in.UserId > 0 {
		mod = mod.Where(cols.UserId, in.UserId)
	}
	if in.AccountType != "" {
		mod = mod.Where(cols.AccountType, in.AccountType)
	}
	if in.ChangeType != "" {
		mod = mod.Where(cols.ChangeType, in.ChangeType)
	}
	if len(in.CreatedAt) == 2 {
		mod = mod.WhereBetween(cols.CreatedAt, in.CreatedAt[0], in.CreatedAt[1])
	}

	err = mod.OrderDesc(cols.Id).Page(in.Page, in.PerPage).ScanAndCount(&list, &totalCount, true)
	if err != nil {
		err = gerror.Wrap(err, "获取钱包流水列表失败")
	}
	return
}

// AdminRecharge 管理员手动充值（算力/积分/余额）
func (s *sToogoWallet) AdminRecharge(ctx context.Context, userId int64, accountType string, amount float64, remark string) (beforeAmount, afterAmount float64, err error) {
	// 获取钱包
	wallet, err := s.GetOrCreate(ctx, userId)
	if err != nil {
		return 0, 0, err
	}

	// 获取当前余额
	switch accountType {
	case "power":
		beforeAmount = wallet.Power
	case "gift_power":
		beforeAmount = wallet.GiftPower
	case "balance":
		beforeAmount = wallet.Balance
	default:
		return 0, 0, gerror.Newf("不支持的账户类型: %s", accountType)
	}

	// 生成订单号
	orderSn := fmt.Sprintf("AR%s%s", gtime.Now().Format("YmdHis"), grand.S(6))

	// 如果备注为空，使用默认备注
	if remark == "" {
		remark = "管理员手动充值"
	}

	// 调用变更余额方法
	err = s.ChangeBalance(ctx, &toogoin.ChangeBalanceInp{
		UserId:      userId,
		AccountType: accountType,
		ChangeType:  "admin_recharge",
		Amount:      amount,
		OrderSn:     orderSn,
		Remark:      remark,
	})
	if err != nil {
		return 0, 0, err
	}

	afterAmount = beforeAmount + amount

	g.Log().Infof(ctx, "[AdminRecharge] 管理员充值成功: userId=%d, accountType=%s, amount=%.4f, before=%.4f, after=%.4f, orderSn=%s",
		userId, accountType, amount, beforeAmount, afterAmount, orderSn)

	return beforeAmount, afterAmount, nil
}

// UserWalletList 用户钱包列表（管理员用）
func (s *sToogoWallet) UserWalletList(ctx context.Context, username, mobile string, page, perPage int) (list []*toogoin.UserWalletListModel, totalCount int, err error) {
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 {
		perPage = 20
	}

	// 构建基础查询条件
	baseModel := g.DB().Model("hg_toogo_wallet w").
		LeftJoin("hg_admin_member m", "w.user_id = m.id")

	// 搜索条件
	if username != "" {
		baseModel = baseModel.WhereLike("m.username", "%"+username+"%")
	}
	if mobile != "" {
		baseModel = baseModel.WhereLike("m.mobile", "%"+mobile+"%")
	}

	// 查询总数（使用克隆的模型）
	totalCount, err = baseModel.Clone().Count()
	if err != nil {
		return nil, 0, gerror.Wrap(err, "查询用户钱包总数失败")
	}

	// 查询列表
	var records []struct {
		UserId     int64   `orm:"user_id"`
		Balance    float64 `orm:"balance"`
		Power      float64 `orm:"power"`
		GiftPower  float64 `orm:"gift_power"`
		Commission float64 `orm:"commission"`
		Username   string  `orm:"username"`
		Mobile     string  `orm:"mobile"`
		CreatedAt  string  `orm:"created_at"`
	}

	err = baseModel.Clone().
		Fields("w.user_id, w.balance, w.power, w.gift_power, w.commission, w.created_at, m.username, m.mobile").
		OrderDesc("w.id").
		Page(page, perPage).
		Scan(&records)
	if err != nil {
		return nil, 0, gerror.Wrap(err, "查询用户钱包列表失败")
	}

	// 获取VIP等级
	userIds := make([]int64, 0, len(records))
	for _, r := range records {
		userIds = append(userIds, r.UserId)
	}

	vipMap := make(map[int64]int)
	if len(userIds) > 0 {
		var toogoUsers []struct {
			MemberId int64 `json:"member_id"`
			VipLevel int   `json:"vip_level"`
		}
		_ = dao.ToogoUser.Ctx(ctx).WhereIn(dao.ToogoUser.Columns().MemberId, userIds).
			Fields("member_id, vip_level").Scan(&toogoUsers)
		for _, u := range toogoUsers {
			vipMap[u.MemberId] = u.VipLevel
		}
	}

	// 转换结果
	list = make([]*toogoin.UserWalletListModel, 0, len(records))
	for _, r := range records {
		item := &toogoin.UserWalletListModel{
			UserId:     r.UserId,
			Username:   r.Username,
			Mobile:     r.Mobile,
			Balance:    r.Balance,
			Power:      r.Power,
			GiftPower:  r.GiftPower,
			TotalPower: r.Power + r.GiftPower,
			Commission: r.Commission,
			VipLevel:   vipMap[r.UserId],
			CreatedAt:  r.CreatedAt,
		}
		list = append(list, item)
	}

	return list, totalCount, nil
}

// OrderHistoryList 历史交易订单列表
func (s *sToogoWallet) OrderHistoryList(ctx context.Context, in *toogoin.OrderHistoryListInp) (list []*toogoin.OrderHistoryModel, totalCount int, err error) {
	memberId := contexts.GetUserId(ctx)
	if memberId <= 0 {
		return nil, 0, gerror.New("用户未登录")
	}

	// 查询用户范围：
	// - 普通用户：只能查自己的
	// - 管理端：可按 userId 筛选；不传则查全部用户
	effectiveUserId := int64(0)
	if contexts.IsUserDept(ctx) {
		effectiveUserId = memberId
	} else if in.UserId > 0 {
		effectiveUserId = in.UserId
	}

	// 构建查询（查询所有状态的订单，包括持仓中和已平仓）
	mod := dao.TradingOrder.Ctx(ctx).As("o")
	if effectiveUserId > 0 {
		mod = mod.Where("o.user_id", effectiveUserId)
	}

	// 条件筛选
	if in.RobotId > 0 {
		mod = mod.Where("o.robot_id", in.RobotId)
	}
	if in.Exchange != "" {
		// dao.TradingOrder.Columns() 可能未包含新字段（历史原因），这里直接使用列名字符串
		mod = mod.Where("o.exchange", in.Exchange)
	}
	if in.Symbol != "" {
		mod = mod.Where("o.symbol", in.Symbol)
	}
	if in.Direction != "" {
		mod = mod.Where("o.direction", in.Direction)
	}
	if in.Status > 0 {
		mod = mod.Where("o.status", in.Status)
	}
	// 时间筛选/排序：
	// - 已平仓(status=2)：用 close_time（用户在“成交明细/平仓记录”里通常关心的是平仓发生的时间）
	// - 其他状态：close_time 可能为空，使用 created_at
	timeCol := "o.created_at"
	orderCol := "o.created_at"
	if in.Status == 2 {
		// close_time 可能为空或被历史脏数据写成 Year=2006 的占位值；
		// 为了让“最新已平仓订单”一定出现在第一页，按“有效 close_time，否则 updated_at”排序/筛选。
		// 注意：这里使用 SQL 表达式是安全的（无用户输入拼接）。
		// 兼容 MySQL/PG：用 CASE + EXTRACT 替代 IF/YEAR
		timeCol = "CASE WHEN o.close_time IS NULL OR EXTRACT(YEAR FROM o.close_time)=2006 THEN o.updated_at ELSE o.close_time END"
		orderCol = timeCol
	}
	if in.StartTime != "" {
		mod = mod.WhereGTE(timeCol, in.StartTime)
	}
	if in.EndTime != "" {
		mod = mod.WhereLTE(timeCol, in.EndTime)
	}

	// 获取总数
	totalCount, err = mod.Count()
	if err != nil {
		return nil, 0, err
	}
	if totalCount == 0 {
		return []*toogoin.OrderHistoryModel{}, 0, nil
	}

	// 【修复】分页查询：显式指定字段，确保时间字段和市场状态、风险偏好被正确获取
	type orderRow struct {
		Id              int64       `orm:"id"`
		UserId          int64       `orm:"user_id"`
		RobotId         int64       `orm:"robot_id"`
		Exchange        string      `orm:"exchange"`
		OrderSn         string      `orm:"order_sn"`
		ExchangeOrderId string      `orm:"exchange_order_id"`
		CloseOrderId    string      `orm:"close_order_id"`
		Symbol          string      `orm:"symbol"`
		Direction       string      `orm:"direction"`
		OpenPrice       float64     `orm:"open_price"`
		ClosePrice      float64     `orm:"close_price"`
		Quantity        float64     `orm:"quantity"`
		OpenFee         float64     `orm:"open_fee"`
		OpenFeeCoin     string      `orm:"open_fee_coin"`
		CloseFee        float64     `orm:"close_fee"`
		CloseFeeCoin    string      `orm:"close_fee_coin"`
		RealizedProfit  float64     `orm:"realized_profit"`
		Status          int         `orm:"status"`
		CloseReason     string      `orm:"close_reason"`
		OpenTime        *gtime.Time `orm:"open_time"`
		CloseTime       *gtime.Time `orm:"close_time"`
		CreatedAt       *gtime.Time `orm:"created_at"`
		UpdatedAt       *gtime.Time `orm:"updated_at"`
	}

	var orders []*orderRow
	err = mod.
		Fields("o.id", "o.user_id", "o.robot_id", "o.exchange", "o.order_sn", "o.exchange_order_id", "o.symbol", "o.direction",
			"o.open_price", "o.close_price", "o.quantity", "o.realized_profit",
			"o.status", "o.close_reason",
			"o.open_fee", "o.open_fee_coin", "o.close_fee", "o.close_fee_coin",
			"o.close_order_id",
			"o.open_time", "o.close_time", "o.created_at", "o.updated_at").
		OrderDesc(orderCol).
		Page(in.Page, in.PerPage).
		Scan(&orders)
	if err != nil {
		// 兼容：部分环境可能尚未执行迁移脚本，缺少 exchange/open_fee/close_fee 字段；此时回退重试
		msg := strings.ToLower(err.Error())
		if strings.Contains(msg, "unknown column") &&
			(strings.Contains(msg, "exchange") || strings.Contains(msg, "open_fee") || strings.Contains(msg, "close_fee")) {
			err2 := mod.
				Fields("o.id", "o.user_id", "o.robot_id", "o.order_sn", "o.exchange_order_id", "o.symbol", "o.direction",
					"o.open_price", "o.close_price", "o.quantity", "o.leverage", "o.margin", "o.realized_profit",
					"o.unrealized_profit", "o.highest_profit", "o.stop_loss_percent", "o.auto_start_retreat_percent",
					"o.profit_retreat_percent", "o.hold_duration", "o.status", "o.close_reason",
					"o.market_state", "o.risk_level",
					"o.open_time", "o.close_time", "o.created_at", "o.updated_at").
				OrderDesc(orderCol).
				Page(in.Page, in.PerPage).
				Scan(&orders)
			if err2 == nil {
				err = nil
			} else {
				return nil, 0, err
			}
		} else {
			return nil, 0, err
		}
	}

	// 获取机器人名称映射
	robotIds := make([]int64, 0, len(orders))
	userIds := make([]int64, 0, len(orders))
	for _, order := range orders {
		if order.RobotId > 0 {
			robotIds = append(robotIds, order.RobotId)
		}
		if order.UserId > 0 {
			userIds = append(userIds, order.UserId)
		}
	}
	robotMap := make(map[int64]string)
	if len(robotIds) > 0 {
		var robots []*entity.TradingRobot
		_ = dao.TradingRobot.Ctx(ctx).
			Fields(dao.TradingRobot.Columns().Id, dao.TradingRobot.Columns().RobotName).
			WhereIn(dao.TradingRobot.Columns().Id, robotIds).
			Scan(&robots)
		for _, robot := range robots {
			robotMap[robot.Id] = robot.RobotName
		}
	}

	// 获取用户名称映射（用户端可能不需要，但管理端用于“区分用户”展示）
	usernameMap := make(map[int64]string)
	if len(userIds) > 0 {
		var members []struct {
			Id       int64  `orm:"id"`
			Username string `orm:"username"`
		}
		_ = dao.AdminMember.Ctx(ctx).
			Fields(dao.AdminMember.Columns().Id, dao.AdminMember.Columns().Username).
			WhereIn(dao.AdminMember.Columns().Id, userIds).
			Scan(&members)
		for _, m := range members {
			usernameMap[m.Id] = m.Username
		}
	}

	// ===== 从平台成交记录(GetTradeHistory)补齐财务字段：数量/手续费/已实现盈亏/开平仓时间 =====
	// 设计目标：
	// - UI 展示以平台成交(fill)为准，避免依赖本地表的历史回填是否成功
	// - 按 (apiConfigId + symbol) 分组：每组只请求一次成交历史，按 orderId 聚合后回填到订单行
	type tradeHistoryProvider interface {
		GetTradeHistory(ctx context.Context, symbol string, limit int) ([]*exlib.Trade, error)
	}
	// robotId -> apiConfigId
	robotApiConfigId := make(map[int64]int64)
	if len(robotIds) > 0 {
		var robotRows []struct {
			Id          int64 `orm:"id"`
			ApiConfigId int64 `orm:"api_config_id"`
		}
		_ = dao.TradingRobot.Ctx(ctx).
			Fields("id", "api_config_id").
			WhereIn("id", robotIds).
			Scan(&robotRows)
		for _, r := range robotRows {
			robotApiConfigId[r.Id] = r.ApiConfigId
		}
	}
	// apiConfigId -> apiConfig entity
	apiConfigIds := make([]int64, 0, len(robotApiConfigId))
	seenApi := make(map[int64]bool)
	for _, v := range robotApiConfigId {
		if v <= 0 || seenApi[v] {
			continue
		}
		seenApi[v] = true
		apiConfigIds = append(apiConfigIds, v)
	}
	apiConfigs := make(map[int64]*entity.TradingApiConfig)
	if len(apiConfigIds) > 0 {
		var cfgs []*entity.TradingApiConfig
		_ = dao.TradingApiConfig.Ctx(ctx).WhereIn("id", apiConfigIds).Scan(&cfgs)
		for _, c := range cfgs {
			if c != nil {
				apiConfigs[c.Id] = c
			}
		}
	}
	// (apiConfigId|symbol) -> needOrderIds set
	//
	// 【资源优化】只对“本地缺失关键字段”的订单才去拉成交历史：
	// - 已实现盈亏/手续费/时间等一旦在平仓时落库，就不需要在列表查询时再打交易所
	// - 显著降低 GetTradeHistory 调用频率与数据量
	needOrderIds := make(map[string]map[string]bool)

	// 已平仓订单的“盈亏/手续费”是交易明细页最重要的信息：
	// - 若 DB 已有数据（平仓时落库），优先使用 DB
	// - 若缺失/不可信，则用 close_order_id / exchange_order_id 精准聚合成交记录补齐
	isRecentClose := func(t *gtime.Time, within time.Duration) bool {
		if t == nil || t.IsZero() || t.Year() == 2006 {
			return false
		}
		return time.Since(t.Time) <= within
	}
	const recentCloseWindow = 14 * 24 * time.Hour

	for _, o := range orders {
		if o == nil || o.RobotId <= 0 || strings.TrimSpace(o.Symbol) == "" {
			continue
		}
		apiId := robotApiConfigId[o.RobotId]
		if apiId <= 0 {
			continue
		}
		symbol := strings.TrimSpace(o.Symbol)
		groupKey := fmt.Sprintf("%d|%s", apiId, symbol)

		// 是否需要补全“开仓成交口径”
		needOpenAgg := false
		if (o.OpenPrice <= 0 || o.Quantity <= 0) && strings.TrimSpace(o.ExchangeOrderId) != "" {
			needOpenAgg = true
		}
		// 开仓手续费：如果本地为0且币种为USDT(或空)，也视为缺失需要补齐
		// 注：手续费=0 有可能是真实值（返佣/免手续费），这里仅在“本地字段明确缺失”的情况下补齐；
		// 当前 orderRow 里 open_fee/open_fee_coin 已存在才会参与判断。
		if !needOpenAgg && (o.OpenFee == 0) && strings.TrimSpace(o.OpenFeeCoin) == "" && strings.TrimSpace(o.ExchangeOrderId) != "" {
			needOpenAgg = true
		}

		// 是否需要补全“平仓成交口径”
		needCloseAgg := false
		if o.Status == 2 && strings.TrimSpace(o.CloseOrderId) != "" {
			// 平仓价/平仓时间/手续费：任一缺失则补齐
			if o.ClosePrice <= 0 || (o.CloseTime == nil || o.CloseTime.IsZero() || o.CloseTime.Year() == 2006) {
				needCloseAgg = true
			}
			// 平仓手续费：缺失则补齐（手续费对账很关键）
			if !needCloseAgg && (o.CloseFee == 0) && strings.TrimSpace(o.CloseFeeCoin) == "" {
				needCloseAgg = true
			}
			// 已实现盈亏：RealizedProfit=0 可能真实为0，也可能是未回填。
			// 为了让交易明细“盈亏/手续费”稳定可用，这里对“近期已平仓订单”更积极地补齐：
			// - 若 close_time 不可信（缺失/占位），已经会触发 needCloseAgg
			// - 若 close_time 可信且在近期窗口内，但 realized_profit=0，仍然尝试补齐一次
			if !needCloseAgg && o.RealizedProfit == 0 && isRecentClose(o.CloseTime, recentCloseWindow) {
				needCloseAgg = true
			}
		}

		if !needOpenAgg && !needCloseAgg {
			continue
		}

		set := needOrderIds[groupKey]
		if set == nil {
			set = make(map[string]bool)
			needOrderIds[groupKey] = set
		}
		if needOpenAgg {
			set[strings.TrimSpace(o.ExchangeOrderId)] = true
		}
		if needCloseAgg {
			set[strings.TrimSpace(o.CloseOrderId)] = true
		}
	}
	// groupKey -> (orderId -> agg)
	groupAgg := make(map[string]map[string]*tradeAggByOrderId)
	for groupKey, idSet := range needOrderIds {
		if len(idSet) == 0 {
			continue
		}
		parts := strings.SplitN(groupKey, "|", 2)
		if len(parts) != 2 {
			continue
		}
		apiId := int64(0)
		_, _ = fmt.Sscanf(parts[0], "%d", &apiId)
		symbol := parts[1]
		if apiId <= 0 || symbol == "" {
			continue
		}

		// 优先使用运行中引擎的 exchange（避免重复创建/解密），否则用 ExchangeManager 缓存实例
		var ex exlib.Exchange
		for _, o := range orders {
			if o != nil && robotApiConfigId[o.RobotId] == apiId {
				if eng := GetRobotTaskManager().GetEngine(o.RobotId); eng != nil && eng.Exchange != nil {
					ex = eng.Exchange
					break
				}
			}
		}
		if ex == nil {
			cfg := apiConfigs[apiId]
			if cfg == nil {
				continue
			}
			ex2, exErr := GetExchangeManager().GetExchangeFromConfig(ctx, cfg)
			if exErr != nil {
				continue
			}
			ex = ex2
		}

		p, okP := ex.(tradeHistoryProvider)
		if !okP {
			continue
		}

		// ===== 成交历史拉取策略（以“盈亏/手续费”为核心，兼顾资源）=====
		// 1) 先用较小 limit 获取，配合短TTL缓存避免重复打API
		// 2) 如果命中率太低（找不到需要的 orderId），再用更大 limit 重试一次
		fetchTrades := func(limit int) ([]*exlib.Trade, error) {
			if limit <= 0 {
				limit = 400
			}
			// 低优先级限流：不等待，避免影响持仓刷新等更关键接口
			if apiId > 0 {
				if !getTradeHistoryLimiter(apiId).Allow() {
					return nil, gerror.New("trade history rate limited (low priority)")
				}
			}
			cacheKey := fmt.Sprintf("toogo:orderHistory:trades:%d:%s:%d", apiId, symbol, limit)
			if v, _ := orderHistoryTradeCache.Get(ctx, cacheKey); v != nil && !v.IsEmpty() {
				if cached, ok := v.Val().([]*exlib.Trade); ok {
					return cached, nil
				}
			}
			trades, err := p.GetTradeHistory(ctx, symbol, limit)
			if err == nil && len(trades) > 0 {
				// 10秒足够覆盖“分页/刷新/并发请求”场景
				_ = orderHistoryTradeCache.Set(ctx, cacheKey, trades, 10*time.Second)
			}
			return trades, err
		}

		buildAgg := func(trades []*exlib.Trade) map[string]*tradeAggByOrderId {
			aggMap := make(map[string]*tradeAggByOrderId)
			for _, t := range trades {
				if t == nil {
					continue
				}
				oid := strings.TrimSpace(t.OrderId)
				if oid == "" || !idSet[oid] {
					continue
				}
				a := aggMap[oid]
				if a == nil {
					a = &tradeAggByOrderId{}
					aggMap[oid] = a
				}
				a.add(t)
			}
			for _, a := range aggMap {
				if a != nil {
					a.finalize()
				}
			}
			return aggMap
		}

		// 初始拉取
		trades, thErr := fetchTrades(400)
		if thErr != nil || len(trades) == 0 {
			continue
		}
		aggMap := buildAgg(trades)

		// 命中率检查：如果需要的订单ID大部分都没命中，说明 limit 不够/成交太多，重试一次更大limit
		if len(idSet) > 0 {
			hit := 0
			for oid := range idSet {
				if _, ok := aggMap[oid]; ok {
					hit++
				}
			}
			hitRate := float64(hit) / float64(len(idSet))
			if hitRate < 0.30 {
				trades2, thErr2 := fetchTrades(1200)
				if thErr2 == nil && len(trades2) > 0 {
					aggMap = buildAgg(trades2)
				}
			}
		}
		groupAgg[groupKey] = aggMap
	}

	// 转换结果
	list = make([]*toogoin.OrderHistoryModel, 0, len(orders))
	for _, order := range orders {
		// ===== 时间有效性判断：过滤历史占位值（2006-01-02 15:04:05）与 zero =====
		isValidTime := func(t *gtime.Time) bool {
			return t != nil && !t.IsZero() && t.Year() != 2006
		}
		isValidTs := func(ts int64) bool { return ts > 0 && ts < 4102444800000 } // <2100-01-01
		formatTs := func(ts int64) string {
			if !isValidTs(ts) {
				return ""
			}
			// 交易所通常毫秒
			if ts > 1e12 {
				return gtime.NewFromTimeStamp(ts / 1000).Format("Y-m-d H:i:s")
			}
			// 秒级
			return gtime.NewFromTimeStamp(ts).Format("Y-m-d H:i:s")
		}

		// ===== 标准化：平仓原因（兼容历史写入的中文/非标准值）=====
		normalizeCloseReason := func(raw string) (code string, text string) {
			r := strings.TrimSpace(strings.ToLower(raw))
			if r == "" {
				return "", ""
			}
			// 英文标准值
			switch r {
			case "stop_loss":
				return "stop_loss", "止损"
			case "take_profit":
				return "take_profit", "止盈"
			case "manual":
				return "manual", "手动"
			case "external_manual", "user_manual", "manual_external":
				return "external_manual", "外部手动"
			case "liquidation", "forced_liquidation", "force_liquidation":
				return "liquidation", "强平"
			case "timeout":
				return "timeout", "超时"
			case "unknown":
				return "unknown", "未知"
			}
			// 兼容历史/日志型写法（包含关键词即可）
			if strings.Contains(r, "stop_loss") || strings.Contains(raw, "止损") {
				return "stop_loss", "止损"
			}
			if strings.Contains(r, "take_profit") || strings.Contains(raw, "止盈") {
				return "take_profit", "止盈"
			}
			if strings.Contains(r, "manual") || strings.Contains(raw, "手动") {
				return "manual", "手动"
			}
			if strings.Contains(r, "external") || strings.Contains(raw, "外部") || strings.Contains(raw, "交易所") {
				return "external_manual", "外部手动"
			}
			if strings.Contains(r, "liquid") || strings.Contains(raw, "强平") || strings.Contains(raw, "爆仓") {
				return "liquidation", "强平"
			}
			if strings.Contains(r, "timeout") || strings.Contains(raw, "超时") {
				return "timeout", "超时"
			}
			// 兜底
			return "unknown", "未知"
		}

		// 方向文本
		directionText := "多"
		if order.Direction == "short" {
			directionText = "空"
		}

		// 状态文本
		statusText := "持仓中"
		if order.Status == 2 {
			statusText = "已平仓"
		} else if order.Status == 3 {
			statusText = "已取消"
		}

		// 平仓原因文本（标准化后输出，避免前端显示 --）
		_, closeReasonText := normalizeCloseReason(order.CloseReason)
		// 已平仓但原因缺失：给用户一个确定的兜底（避免 "--"）
		if order.Status == 2 && closeReasonText == "" {
			closeReasonText = "未知"
		}

		// ===== 优先使用平台成交汇总 =====
		apiId := robotApiConfigId[order.RobotId]
		groupKey := ""
		if apiId > 0 && strings.TrimSpace(order.Symbol) != "" {
			groupKey = fmt.Sprintf("%d|%s", apiId, strings.TrimSpace(order.Symbol))
		}
		var openAgg, closeAgg *tradeAggByOrderId
		if groupKey != "" {
			if m := groupAgg[groupKey]; m != nil {
				if strings.TrimSpace(order.ExchangeOrderId) != "" {
					openAgg = m[strings.TrimSpace(order.ExchangeOrderId)]
				}
				if strings.TrimSpace(order.CloseOrderId) != "" {
					closeAgg = m[strings.TrimSpace(order.CloseOrderId)]
				}
			}
		}

		// 数量：仅来自平台成交汇总（开仓优先；若开仓未命中则退化使用平仓汇总）
		var quantity *float64
		if openAgg != nil && openAgg.SumQty > 0 {
			v := openAgg.SumQty
			quantity = &v
		} else if closeAgg != nil && closeAgg.SumQty > 0 {
			v := closeAgg.SumQty
			quantity = &v
		} else if order.Quantity > 0 {
			// 兜底：使用本地订单表记录（部分环境可能未能拉到成交汇总）
			v := order.Quantity
			quantity = &v
		}

		// 开/平仓均价：优先成交汇总
		var openPrice *float64
		if openAgg != nil && openAgg.AvgPrice > 0 {
			v := openAgg.AvgPrice
			openPrice = &v
		} else if order.OpenPrice > 0 {
			v := order.OpenPrice
			openPrice = &v
		}
		var closePrice *float64
		if closeAgg != nil && closeAgg.AvgPrice > 0 {
			v := closeAgg.AvgPrice
			closePrice = &v
		} else if order.ClosePrice > 0 {
			v := order.ClosePrice
			closePrice = &v
		}

		// 手续费：仅来自平台成交汇总；且仅当手续费币种为 USDT（或为空）时才返回“(USDT)”口径
		isUSDT := func(coin string) bool {
			c := strings.TrimSpace(strings.ToUpper(coin))
			return c == "" || c == "USDT"
		}
		var openFee *float64
		if openAgg != nil && openAgg.HasAnyRecord && isUSDT(openAgg.FeeCoin) {
			// 注意：手续费可能为 0（返佣/免手续费/极小额四舍五入），0 也是有效值，不能当成“缺失”
			v := openAgg.Commission
			openFee = &v
		} else if isUSDT(order.OpenFeeCoin) {
			// 兜底：使用本地订单表手续费（可能存在小数位较多，前端负责格式化展示）
			v := order.OpenFee
			openFee = &v
		}
		var closeFee *float64
		if closeAgg != nil && closeAgg.HasAnyRecord && isUSDT(closeAgg.FeeCoin) {
			v := closeAgg.Commission
			closeFee = &v
		} else if isUSDT(order.CloseFeeCoin) {
			v := order.CloseFee
			closeFee = &v
		}
		var feeTotal *float64
		if openFee != nil || closeFee != nil {
			sum := 0.0
			if openFee != nil {
				sum += *openFee
			}
			if closeFee != nil {
				sum += *closeFee
			}
			feeTotal = &sum
		}

		// 已实现盈亏：优先平台成交汇总（平仓成交汇总优先；0 也是有效值），否则回退本地订单表
		var realizedProfit *float64
		if closeAgg != nil && closeAgg.HasAnyRecord {
			// 注意：0 也是有效值（盈亏为 0 的平仓）
			v := closeAgg.RealizedPnl
			realizedProfit = &v
		} else {
			// 兜底：使用本地订单表字段（部分环境/部分订单无法拉取成交历史）
			v := order.RealizedProfit
			realizedProfit = &v
		}
		var profitAmount *float64
		var lossAmount *float64
		if realizedProfit != nil {
			if *realizedProfit > 0 {
				v := *realizedProfit
				profitAmount = &v
			} else if *realizedProfit < 0 {
				v := math.Abs(*realizedProfit)
				lossAmount = &v
			} else {
				// 0 也输出为 0，避免前端展示为 "--"
				z := 0.0
				profitAmount = &z
				lossAmount = &z
			}
		}

		item := &toogoin.OrderHistoryModel{
			Id:              order.Id,
			UserId:          order.UserId,
			Username:        usernameMap[order.UserId],
			Exchange:        order.Exchange,
			RobotId:         order.RobotId,
			RobotName:       robotMap[order.RobotId],
			OrderSn:         order.OrderSn,
			ExchangeOrderId: order.ExchangeOrderId,
			CloseOrderId:    order.CloseOrderId,
			Symbol:          order.Symbol,
			Direction:       order.Direction,
			DirectionText:   directionText,
			OpenPrice:       openPrice,
			ClosePrice:      closePrice,
			Quantity:        quantity,
			OpenFee:         openFee,
			CloseFee:        closeFee,
			FeeTotal:        feeTotal,
			ProfitAmount:    profitAmount,
			LossAmount:      lossAmount,
			RealizedProfit:  realizedProfit,
			OpenTime:        "",
			CloseTime:       "",
			Status:          order.Status,
			StatusText:      statusText,
			CloseReasonText: closeReasonText,
			CreatedAt:       "",
		}

		// ===== 开仓时间展示（优先平台成交时间，其次 open_time/created_at/updated_at；全部过滤 2006 占位值）=====
		if openAgg != nil && isValidTs(openAgg.MinTs) {
			item.OpenTime = formatTs(openAgg.MinTs)
		}
		if item.OpenTime == "" {
			switch {
			case isValidTime(order.OpenTime):
				item.OpenTime = order.OpenTime.Format("Y-m-d H:i:s")
			case isValidTime(order.CreatedAt):
				item.OpenTime = order.CreatedAt.Format("Y-m-d H:i:s")
			case isValidTime(order.UpdatedAt):
				item.OpenTime = order.UpdatedAt.Format("Y-m-d H:i:s")
			default:
				item.OpenTime = ""
			}
		}

		// 平仓时间：
		// - 已平仓(status=2)：优先 close_time
		// - 若 close_time 是历史占位值（Year=2006）或缺失，则用 updated_at 兜底展示（避免展示“2006-01-02 15:04:05”误导用户）
		if order.Status == 2 {
			// 优先平台成交时间（更贴近真实平仓发生时刻）
			if closeAgg != nil && isValidTs(closeAgg.MaxTs) {
				item.CloseTime = formatTs(closeAgg.MaxTs)
			}
			if item.CloseTime == "" {
				switch {
				case isValidTime(order.CloseTime):
					item.CloseTime = order.CloseTime.Format("Y-m-d H:i:s")
				case isValidTime(order.UpdatedAt):
					item.CloseTime = order.UpdatedAt.Format("Y-m-d H:i:s")
				case isValidTime(order.CreatedAt):
					item.CloseTime = order.CreatedAt.Format("Y-m-d H:i:s")
				default:
					item.CloseTime = ""
				}
			}
		} else {
			item.CloseTime = "" // 非已平仓不展示
		}

		// 创建时间：必须存在
		if isValidTime(order.CreatedAt) {
			item.CreatedAt = order.CreatedAt.Format("Y-m-d H:i:s")
		} else if isValidTime(order.UpdatedAt) {
			// 兼容极端脏数据：created_at 也被写成占位值
			item.CreatedAt = order.UpdatedAt.Format("Y-m-d H:i:s")
		} else {
			// 【修复】如果 CreatedAt 也为空，记录警告并使用当前时间（不应该发生）
			g.Log().Warningf(ctx, "[OrderHistoryList] 订单 CreatedAt 为空: orderId=%d", order.Id)
			item.CreatedAt = gtime.Now().Format("Y-m-d H:i:s")
		}

		list = append(list, item)
	}

	return list, totalCount, nil
}

// TradeHistoryList 成交流水列表（每条记录对应交易所一笔成交）
func (s *sToogoWallet) TradeHistoryList(ctx context.Context, in *toogoin.TradeHistoryListInp) (list []*toogoin.TradeFillModel, totalCount int, summary *toogoin.TradeFillSummary, err error) {
	memberId := contexts.GetUserId(ctx)
	if memberId <= 0 {
		return nil, 0, nil, gerror.New("用户未登录")
	}

	// 查询用户范围：
	// - 普通用户：只能查自己的
	// - 管理端：可按 userId 筛选；不传则查全部用户
	effectiveUserId := int64(0)
	if contexts.IsUserDept(ctx) {
		effectiveUserId = memberId
	} else if in.UserId > 0 {
		effectiveUserId = in.UserId
	}

	mod := dao.TradingTradeFill.Ctx(ctx).As("f")
	if effectiveUserId > 0 {
		mod = mod.Where("f.user_id", effectiveUserId)
	}
	if in.RobotId > 0 {
		mod = mod.Where("f.robot_id", in.RobotId)
	}
	if in.SessionId > 0 {
		mod = mod.Where("f.session_id", in.SessionId)
	}
	if in.ApiConfigId > 0 {
		mod = mod.Where("f.api_config_id", in.ApiConfigId)
	}
	if in.Exchange != "" {
		mod = mod.Where("f.exchange", in.Exchange)
	}
	if in.Symbol != "" {
		mod = mod.Where("f.symbol", in.Symbol)
	}
	if strings.TrimSpace(in.OrderId) != "" {
		mod = mod.Where("f.order_id", strings.TrimSpace(in.OrderId))
	}
	if strings.TrimSpace(in.TradeId) != "" {
		mod = mod.Where("f.trade_id", strings.TrimSpace(in.TradeId))
	}
	if strings.TrimSpace(in.Side) != "" {
		mod = mod.Where("f.side", strings.ToUpper(strings.TrimSpace(in.Side)))
	}

	// 时间过滤：使用 ts(毫秒) 作为主时间口径
	parseMs := func(v string) int64 {
		v = strings.TrimSpace(v)
		if v == "" {
			return 0
		}
		t, e := gtime.StrToTime(v)
		if e != nil || t == nil || t.IsZero() {
			return 0
		}
		return t.TimestampMilli()
	}
	startMs := parseMs(in.StartTime)
	endMs := parseMs(in.EndTime)
	if startMs > 0 {
		mod = mod.WhereGTE("f.ts", startMs)
	}
	if endMs > 0 {
		mod = mod.WhereLTE("f.ts", endMs)
	}

	// 初始化汇总统计
	summary = &toogoin.TradeFillSummary{}

	totalCount, err = mod.Count()
	if err != nil {
		return nil, 0, nil, err
	}
	if totalCount == 0 {
		return []*toogoin.TradeFillModel{}, 0, summary, nil
	}

	// 【新增】查询汇总统计（全量统计，不受分页影响）
	type aggRow struct {
		TotalPnl    float64 `orm:"total_pnl"`
		TotalProfit float64 `orm:"total_profit"`
		TotalLoss   float64 `orm:"total_loss"`
		TotalFee    float64 `orm:"total_fee"`
	}
	var agg aggRow
	aggErr := mod.Clone().
		Fields(
			"COALESCE(SUM(realized_pnl), 0) AS total_pnl",
			"COALESCE(SUM(CASE WHEN realized_pnl > 0 THEN realized_pnl ELSE 0 END), 0) AS total_profit",
			"COALESCE(SUM(CASE WHEN realized_pnl < 0 THEN realized_pnl ELSE 0 END), 0) AS total_loss",
			"COALESCE(SUM(fee), 0) AS total_fee",
		).
		Scan(&agg)
	if aggErr == nil {
		summary.TotalCount = totalCount
		summary.TotalPnl = agg.TotalPnl
		summary.TotalProfit = agg.TotalProfit
		summary.TotalLoss = agg.TotalLoss
		summary.TotalFee = agg.TotalFee
		summary.TotalNetPnl = agg.TotalPnl - agg.TotalFee
	}

	type row struct {
		Id            int64   `orm:"id"`
		ApiConfigId   int64   `orm:"api_config_id"`
		Exchange      string  `orm:"exchange"`
		UserId        int64   `orm:"user_id"`
		RobotId       int64   `orm:"robot_id"`
		SessionId     *int64  `orm:"session_id"`
		Symbol        string  `orm:"symbol"`
		OrderId       string  `orm:"order_id"`
		ClientOrderId string  `orm:"client_order_id"`
		TradeId       string  `orm:"trade_id"`
		Side          string  `orm:"side"`
		Qty           float64 `orm:"qty"`
		Price         float64 `orm:"price"`
		Fee           float64 `orm:"fee"`
		FeeCoin       string  `orm:"fee_coin"`
		RealizedPnl   float64 `orm:"realized_pnl"`
		Ts            int64   `orm:"ts"`
	}
	var rows []*row
	err = mod.
		Fields(
			"f.id", "f.api_config_id", "f.exchange",
			"f.user_id", "f.robot_id", "f.session_id",
			"f.symbol", "f.order_id", "f.client_order_id", "f.trade_id",
			"f.side", "f.qty", "f.price", "f.fee", "f.fee_coin", "f.realized_pnl",
			"f.ts",
		).
		OrderDesc("f.ts").
		OrderDesc("f.id").
		Page(in.Page, in.PerPage).
		Scan(&rows)
	if err != nil {
		return nil, 0, nil, err
	}

	// robotName 映射
	robotIds := make([]int64, 0, len(rows))
	userIds := make([]int64, 0, len(rows))
	for _, r := range rows {
		if r == nil {
			continue
		}
		if r.RobotId > 0 {
			robotIds = append(robotIds, r.RobotId)
		}
		if r.UserId > 0 {
			userIds = append(userIds, r.UserId)
		}
	}
	robotMap := make(map[int64]string)
	if len(robotIds) > 0 {
		var robots []*entity.TradingRobot
		_ = dao.TradingRobot.Ctx(ctx).
			Fields(dao.TradingRobot.Columns().Id, dao.TradingRobot.Columns().RobotName).
			WhereIn(dao.TradingRobot.Columns().Id, robotIds).
			Scan(&robots)
		for _, rb := range robots {
			robotMap[rb.Id] = rb.RobotName
		}
	}
	usernameMap := make(map[int64]string)
	if len(userIds) > 0 {
		var members []struct {
			Id       int64  `orm:"id"`
			Username string `orm:"username"`
		}
		_ = dao.AdminMember.Ctx(ctx).
			Fields(dao.AdminMember.Columns().Id, dao.AdminMember.Columns().Username).
			WhereIn(dao.AdminMember.Columns().Id, userIds).
			Scan(&members)
		for _, m := range members {
			usernameMap[m.Id] = m.Username
		}
	}

	formatTs := func(ts int64) string {
		if ts <= 0 || ts > 4102444800000 {
			return ""
		}
		if ts > 1e12 {
			return gtime.NewFromTimeStamp(ts / 1000).Format("Y-m-d H:i:s")
		}
		return gtime.NewFromTimeStamp(ts).Format("Y-m-d H:i:s")
	}

	list = make([]*toogoin.TradeFillModel, 0, len(rows))
	for _, r := range rows {
		if r == nil {
			continue
		}
		item := &toogoin.TradeFillModel{
			Id:            r.Id,
			ApiConfigId:   r.ApiConfigId,
			Exchange:      r.Exchange,
			UserId:        r.UserId,
			Username:      usernameMap[r.UserId],
			RobotId:       r.RobotId,
			RobotName:     robotMap[r.RobotId],
			SessionId:     r.SessionId,
			Symbol:        r.Symbol,
			OrderId:       r.OrderId,
			ClientOrderId: r.ClientOrderId,
			TradeId:       r.TradeId,
			Side:          r.Side,
			Qty:           r.Qty,
			Price:         r.Price,
			Fee:           r.Fee,
			FeeCoin:       r.FeeCoin,
			RealizedPnl:   r.RealizedPnl,
			Ts:            r.Ts,
			Time:          formatTs(r.Ts),
		}
		list = append(list, item)
	}

	return list, totalCount, summary, nil
}

// RunSessionSummaryList 运行区间盈亏汇总列表
func (s *sToogoWallet) RunSessionSummaryList(ctx context.Context, in *toogoin.RunSessionSummaryListInp) (list []*toogoin.RunSessionSummaryModel, totalCount int, summary *toogoin.RunSessionTotalSummary, err error) {
	memberId := contexts.GetUserId(ctx)
	if memberId <= 0 {
		return nil, 0, nil, gerror.New("用户未登录")
	}

	// 【兼容补齐】如果机器人处于“运行中”，但由于历史原因没有写入 run_session（例如：建表/上线之前已运行），
	// 则自动补插一条“end_time=NULL”的运行区间记录，确保页面能看到“运行中”的区间。
	//
	// 说明：只在需要包含“运行中”的查询中执行（isRunning=0/1）。
	if in.IsRunning == 0 || in.IsRunning == 1 {
		type rbRow struct {
			Id        int64       `orm:"id"`
			Exchange  string      `orm:"exchange"`
			Symbol    string      `orm:"symbol"`
			StartTime *gtime.Time `orm:"start_time"`
		}
		var robots []*rbRow
		rbMod := dao.TradingRobot.Ctx(ctx).
			Fields("id", "exchange", "symbol", "start_time").
			Where(dao.TradingRobot.Columns().UserId, memberId).
			Where(dao.TradingRobot.Columns().Status, 2). // 2=运行中
			WhereNull(dao.TradingRobot.Columns().DeletedAt)
		if in.RobotId > 0 {
			rbMod = rbMod.Where(dao.TradingRobot.Columns().Id, in.RobotId)
		}
		if in.Exchange != "" {
			rbMod = rbMod.Where(dao.TradingRobot.Columns().Exchange, in.Exchange)
		}
		if in.Symbol != "" {
			rbMod = rbMod.Where(dao.TradingRobot.Columns().Symbol, in.Symbol)
		}
		_ = rbMod.Scan(&robots)

		if len(robots) > 0 {
			ids := make([]int64, 0, len(robots))
			for _, r := range robots {
				if r != nil && r.Id > 0 {
					ids = append(ids, r.Id)
				}
			}
			if len(ids) > 0 {
				type sessRow struct {
					RobotId int64 `orm:"robot_id"`
				}
				var sess []*sessRow
				_ = dao.TradingRobotRunSession.Ctx(ctx).
					Fields(dao.TradingRobotRunSession.Columns().RobotId).
					Where(dao.TradingRobotRunSession.Columns().UserId, memberId).
					WhereIn(dao.TradingRobotRunSession.Columns().RobotId, ids).
					WhereNull(dao.TradingRobotRunSession.Columns().EndTime).
					Scan(&sess)
				exists := make(map[int64]bool, len(sess))
				for _, s := range sess {
					if s != nil && s.RobotId > 0 {
						exists[s.RobotId] = true
					}
				}

				now := gtime.Now()
				data := make([]g.Map, 0)
				for _, r := range robots {
					if r == nil || r.Id <= 0 {
						continue
					}
					if exists[r.Id] {
						continue
					}
					st := now
					if r.StartTime != nil && !r.StartTime.IsZero() && r.StartTime.Year() != 2006 {
						st = r.StartTime
					}
					data = append(data, g.Map{
						"robot_id":   r.Id,
						"user_id":    memberId,
						"exchange":   r.Exchange,
						"symbol":     r.Symbol,
						"start_time": st,
					})
				}
				if len(data) > 0 {
					_, _ = dao.TradingRobotRunSession.Ctx(ctx).Data(data).Insert()
				}
			}
		}
	}

	// 构建查询
	mod := dao.TradingRobotRunSession.Ctx(ctx).
		Where(dao.TradingRobotRunSession.Columns().UserId, memberId)

	// 条件筛选
	if in.RobotId > 0 {
		mod = mod.Where(dao.TradingRobotRunSession.Columns().RobotId, in.RobotId)
	}
	if in.Exchange != "" {
		mod = mod.Where(dao.TradingRobotRunSession.Columns().Exchange, in.Exchange)
	}
	if in.Symbol != "" {
		mod = mod.Where(dao.TradingRobotRunSession.Columns().Symbol, in.Symbol)
	}
	if in.IsRunning == 1 {
		mod = mod.WhereNull(dao.TradingRobotRunSession.Columns().EndTime)
	} else if in.IsRunning == 2 {
		mod = mod.WhereNotNull(dao.TradingRobotRunSession.Columns().EndTime)
	}
	if in.StartTime != "" {
		mod = mod.WhereGTE(dao.TradingRobotRunSession.Columns().StartTime, in.StartTime)
	}
	if in.EndTime != "" {
		mod = mod.WhereLTE(dao.TradingRobotRunSession.Columns().StartTime, in.EndTime)
	}

	// 获取总数
	totalCount, err = mod.Count()
	if err != nil {
		return nil, 0, nil, err
	}

	if totalCount == 0 {
		return []*toogoin.RunSessionSummaryModel{}, 0, &toogoin.RunSessionTotalSummary{}, nil
	}

	// 查询列表
	var sessions []*entity.TradingRobotRunSession
	err = mod.OrderDesc(dao.TradingRobotRunSession.Columns().Id).
		Page(in.Page, in.PerPage).
		Scan(&sessions)
	if err != nil {
		return nil, 0, nil, err
	}

	// 获取机器人名称映射
	robotIds := make([]int64, 0, len(sessions))
	for _, sess := range sessions {
		if sess.RobotId > 0 {
			robotIds = append(robotIds, sess.RobotId)
		}
	}
	robotMap := make(map[int64]string)
	if len(robotIds) > 0 {
		var robots []*entity.TradingRobot
		_ = dao.TradingRobot.Ctx(ctx).
			Fields(dao.TradingRobot.Columns().Id, dao.TradingRobot.Columns().RobotName).
			WhereIn(dao.TradingRobot.Columns().Id, robotIds).
			Scan(&robots)
		for _, robot := range robots {
			robotMap[robot.Id] = robot.RobotName
		}
	}

	// 转换结果
	list = make([]*toogoin.RunSessionSummaryModel, 0, len(sessions))
	for _, sess := range sessions {
		// 结束原因文本
		endReasonText := ""
		switch sess.EndReason {
		case "pause":
			endReasonText = "暂停"
		case "stop":
			endReasonText = "停止"
		case "auto_stop":
			endReasonText = "自动停止"
		case "error":
			endReasonText = "异常"
		default:
			if sess.EndTime == nil {
				endReasonText = "运行中"
			}
		}

		// 【核心】运行中的区间：实时计算从启动时间到现在的统计数据
		isRunning := sess.EndTime == nil
		var totalPnl, totalFee *float64
		var tradeCount, runtimeSeconds int

		if isRunning && sess.StartTime != nil && !sess.StartTime.IsZero() {
			// 计算实时运行时长（秒）
			runtimeSeconds = int(gtime.Now().Sub(sess.StartTime).Seconds())

			// 从成交流水表统计该区间的盈亏/手续费/成交笔数
			// 时间范围：start_time 到现在
			startMs := sess.StartTime.UnixMilli()
			nowMs := gtime.Now().UnixMilli()

			type aggRow struct {
				TotalPnl   float64 `orm:"total_pnl"`
				TotalFee   float64 `orm:"total_fee"`
				TradeCount int     `orm:"trade_count"`
			}
			var agg aggRow
			err := dao.TradingTradeFill.Ctx(ctx).
				Fields("COALESCE(SUM(realized_pnl), 0) AS total_pnl, COALESCE(SUM(fee), 0) AS total_fee, COUNT(*) AS trade_count").
				Where("robot_id", sess.RobotId).
				Where("user_id", memberId).
				WhereGTE("ts", startMs).
				WhereLTE("ts", nowMs).
				Scan(&agg)
			if err == nil {
				totalPnl = &agg.TotalPnl
				totalFee = &agg.TotalFee
				tradeCount = agg.TradeCount
			}
		} else {
			// 已结束的区间：使用数据库中已保存的统计数据
			totalPnl = sess.TotalPnl
			totalFee = sess.TotalFee
			tradeCount = sess.TradeCount
			runtimeSeconds = sess.RuntimeSeconds
		}

		// 运行时长文本
		runtimeText := formatDuration(runtimeSeconds)

		// 净盈亏
		var netPnl *float64
		if totalPnl != nil && totalFee != nil {
			v := *totalPnl - *totalFee
			netPnl = &v
		} else if totalPnl != nil {
			netPnl = totalPnl
		}

		item := &toogoin.RunSessionSummaryModel{
			Id:             sess.Id,
			RobotId:        sess.RobotId,
			RobotName:      robotMap[sess.RobotId],
			Exchange:       sess.Exchange,
			Symbol:         sess.Symbol,
			StartTime:      "",
			EndTime:        "",
			EndReason:      sess.EndReason,
			EndReasonText:  endReasonText,
			RuntimeSeconds: runtimeSeconds,
			RuntimeText:    runtimeText,
			TotalPnl:       totalPnl,
			TotalFee:       totalFee,
			NetPnl:         netPnl,
			TradeCount:     tradeCount,
			SyncedAt:       "",
			IsRunning:      isRunning,
		}

		if sess.StartTime != nil {
			item.StartTime = sess.StartTime.Format("Y-m-d H:i:s")
		}
		if sess.EndTime != nil {
			item.EndTime = sess.EndTime.Format("Y-m-d H:i:s")
		}
		if sess.SyncedAt != nil {
			item.SyncedAt = sess.SyncedAt.Format("Y-m-d H:i:s")
		}

		list = append(list, item)
	}

	// 【核心修改】全量汇总统计（不受分页影响）
	// 对于运行区间，需要分别统计"已结束区间"（从DB）和"运行中区间"（实时计算）
	summary = &toogoin.RunSessionTotalSummary{
		TotalSessions: totalCount,
	}

	// 1) 查询已结束区间的汇总（从DB直接聚合）
	type endedAgg struct {
		TotalRuntime int     `orm:"total_runtime"`
		TotalPnl     float64 `orm:"total_pnl"`
		TotalFee     float64 `orm:"total_fee"`
		TotalTrades  int     `orm:"total_trades"`
		TotalProfit  float64 `orm:"total_profit"`
		TotalLoss    float64 `orm:"total_loss"`
	}
	var ea endedAgg
	endedMod := dao.TradingRobotRunSession.Ctx(ctx).
		Where(dao.TradingRobotRunSession.Columns().UserId, memberId).
		WhereNotNull(dao.TradingRobotRunSession.Columns().EndTime)
	if in.RobotId > 0 {
		endedMod = endedMod.Where(dao.TradingRobotRunSession.Columns().RobotId, in.RobotId)
	}
	if in.Exchange != "" {
		endedMod = endedMod.Where(dao.TradingRobotRunSession.Columns().Exchange, in.Exchange)
	}
	if in.Symbol != "" {
		endedMod = endedMod.Where(dao.TradingRobotRunSession.Columns().Symbol, in.Symbol)
	}
	if in.StartTime != "" {
		endedMod = endedMod.WhereGTE(dao.TradingRobotRunSession.Columns().StartTime, in.StartTime)
	}
	if in.EndTime != "" {
		endedMod = endedMod.WhereLTE(dao.TradingRobotRunSession.Columns().StartTime, in.EndTime)
	}
	_ = endedMod.Fields(
		"COALESCE(SUM(runtime_seconds), 0) AS total_runtime",
		"COALESCE(SUM(total_pnl), 0) AS total_pnl",
		"COALESCE(SUM(total_fee), 0) AS total_fee",
		"COALESCE(SUM(trade_count), 0) AS total_trades",
		"COALESCE(SUM(CASE WHEN total_pnl > 0 THEN total_pnl ELSE 0 END), 0) AS total_profit",
		"COALESCE(SUM(CASE WHEN total_pnl < 0 THEN total_pnl ELSE 0 END), 0) AS total_loss",
	).Scan(&ea)

	summary.TotalRuntime = ea.TotalRuntime
	summary.TotalPnl = ea.TotalPnl
	summary.TotalFee = ea.TotalFee
	summary.TotalTrades = ea.TotalTrades
	summary.TotalProfit = ea.TotalProfit
	summary.TotalLoss = ea.TotalLoss

	// 2) 如果查询包含运行中的区间（IsRunning=0 或 1），需要实时计算运行中区间的统计
	if in.IsRunning == 0 || in.IsRunning == 1 {
		// 查询所有运行中区间
		var runningSessions []*entity.TradingRobotRunSession
		runningMod := dao.TradingRobotRunSession.Ctx(ctx).
			Where(dao.TradingRobotRunSession.Columns().UserId, memberId).
			WhereNull(dao.TradingRobotRunSession.Columns().EndTime)
		if in.RobotId > 0 {
			runningMod = runningMod.Where(dao.TradingRobotRunSession.Columns().RobotId, in.RobotId)
		}
		if in.Exchange != "" {
			runningMod = runningMod.Where(dao.TradingRobotRunSession.Columns().Exchange, in.Exchange)
		}
		if in.Symbol != "" {
			runningMod = runningMod.Where(dao.TradingRobotRunSession.Columns().Symbol, in.Symbol)
		}
		if in.StartTime != "" {
			runningMod = runningMod.WhereGTE(dao.TradingRobotRunSession.Columns().StartTime, in.StartTime)
		}
		if in.EndTime != "" {
			runningMod = runningMod.WhereLTE(dao.TradingRobotRunSession.Columns().StartTime, in.EndTime)
		}
		_ = runningMod.Scan(&runningSessions)

		// 实时计算每个运行中区间的统计
		for _, rs := range runningSessions {
			if rs == nil || rs.StartTime == nil || rs.StartTime.IsZero() {
				continue
			}
			// 运行时长
			rt := int(gtime.Now().Sub(rs.StartTime).Seconds())
			summary.TotalRuntime += rt

			// 从成交流水表统计
			startMs := rs.StartTime.UnixMilli()
			nowMs := gtime.Now().UnixMilli()
			type fillAgg struct {
				TotalPnl    float64 `orm:"total_pnl"`
				TotalProfit float64 `orm:"total_profit"`
				TotalLoss   float64 `orm:"total_loss"`
				TotalFee    float64 `orm:"total_fee"`
				TradeCount  int     `orm:"trade_count"`
			}
			var fa fillAgg
			_ = dao.TradingTradeFill.Ctx(ctx).
				Fields(
					"COALESCE(SUM(realized_pnl), 0) AS total_pnl",
					"COALESCE(SUM(CASE WHEN realized_pnl > 0 THEN realized_pnl ELSE 0 END), 0) AS total_profit",
					"COALESCE(SUM(CASE WHEN realized_pnl < 0 THEN realized_pnl ELSE 0 END), 0) AS total_loss",
					"COALESCE(SUM(fee), 0) AS total_fee",
					"COUNT(*) AS trade_count",
				).
				Where("robot_id", rs.RobotId).
				Where("user_id", memberId).
				WhereGTE("ts", startMs).
				WhereLTE("ts", nowMs).
				Scan(&fa)

			summary.TotalPnl += fa.TotalPnl
			summary.TotalProfit += fa.TotalProfit
			summary.TotalLoss += fa.TotalLoss
			summary.TotalFee += fa.TotalFee
			summary.TotalTrades += fa.TradeCount
		}
	}

	summary.TotalNetPnl = summary.TotalPnl - summary.TotalFee
	summary.TotalRuntimeText = formatDuration(summary.TotalRuntime)

	return list, totalCount, summary, nil
}

// formatDuration 格式化时长
func formatDuration(seconds int) string {
	if seconds <= 0 {
		return "0秒"
	}
	days := seconds / 86400
	hours := (seconds % 86400) / 3600
	minutes := (seconds % 3600) / 60
	secs := seconds % 60

	if days > 0 {
		return fmt.Sprintf("%d天%d时%d分", days, hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%d时%d分%d秒", hours, minutes, secs)
	}
	if minutes > 0 {
		return fmt.Sprintf("%d分%d秒", minutes, secs)
	}
	return fmt.Sprintf("%d秒", secs)
}

// SyncRunSession 同步运行区间盈亏数据
func (s *sToogoWallet) SyncRunSession(ctx context.Context, sessionId int64) (totalPnl, totalFee float64, tradeCount int, err error) {
	memberId := contexts.GetUserId(ctx)
	if memberId <= 0 {
		return 0, 0, 0, gerror.New("用户未登录")
	}

	// 查询区间记录
	var session *entity.TradingRobotRunSession
	err = dao.TradingRobotRunSession.Ctx(ctx).
		Where(dao.TradingRobotRunSession.Columns().Id, sessionId).
		Where(dao.TradingRobotRunSession.Columns().UserId, memberId).
		Scan(&session)
	if err != nil {
		return 0, 0, 0, gerror.Wrap(err, "查询区间记录失败")
	}
	if session == nil {
		return 0, 0, 0, gerror.New("区间记录不存在")
	}

	// 查询机器人信息
	var robot *entity.TradingRobot
	err = dao.TradingRobot.Ctx(ctx).Where(dao.TradingRobot.Columns().Id, session.RobotId).Scan(&robot)
	if err != nil || robot == nil {
		return 0, 0, 0, gerror.New("机器人不存在")
	}

	// 获取API配置
	var apiConfig *entity.TradingApiConfig
	err = dao.TradingApiConfig.Ctx(ctx).Where(dao.TradingApiConfig.Columns().Id, robot.ApiConfigId).Scan(&apiConfig)
	if err != nil || apiConfig == nil {
		return 0, 0, 0, gerror.New("API配置不存在")
	}

	// 创建交易所实例
	ex, err := GetExchangeManager().GetExchangeFromConfig(ctx, apiConfig)
	if err != nil {
		return 0, 0, 0, gerror.Wrap(err, "创建交易所实例失败")
	}

	// 获取成交历史
	type tradeHistoryProvider interface {
		GetTradeHistory(ctx context.Context, symbol string, limit int) ([]*exlib.Trade, error)
	}
	p, ok := ex.(tradeHistoryProvider)
	if !ok {
		return 0, 0, 0, gerror.New("交易所不支持成交历史查询")
	}

	trades, err := p.GetTradeHistory(ctx, session.Symbol, 1000)
	if err != nil {
		return 0, 0, 0, gerror.Wrap(err, "获取成交历史失败")
	}

	// 【新增】同步时批量落库成交流水（幂等去重：uk_api_exchange_trade）
	// 说明：运行区间同步属于“后台同步/对账”环节，成交数据应优先落库，页面查询只读DB。
	sid := sessionId
	_, _, _ = upsertTradeFillsFromTrades(ctx, robot.ApiConfigId, robot.Exchange, session.Symbol, trades, &sid)

	// 过滤区间内的成交
	startMs := int64(0)
	if session.StartTime != nil {
		startMs = session.StartTime.UnixMilli()
	}
	endMs := int64(0)
	if session.EndTime != nil {
		endMs = session.EndTime.UnixMilli()
	} else {
		endMs = gtime.Now().UnixMilli() // 仍在运行，用当前时间
	}

	for _, t := range trades {
		if t == nil || t.Time <= 0 {
			continue
		}
		// 过滤时间范围
		if t.Time < startMs || t.Time > endMs {
			continue
		}
		tradeCount++
		totalPnl += t.RealizedPnl
		totalFee += math.Abs(t.Commission)
	}

	// 更新区间记录
	_, err = dao.TradingRobotRunSession.Ctx(ctx).
		Where(dao.TradingRobotRunSession.Columns().Id, sessionId).
		Data(g.Map{
			"total_pnl":   totalPnl,
			"total_fee":   totalFee,
			"trade_count": tradeCount,
			"synced_at":   gtime.Now(),
		}).
		Update()
	if err != nil {
		g.Log().Warningf(ctx, "更新区间盈亏失败: sessionId=%d, err=%v", sessionId, err)
	}

	return totalPnl, totalFee, tradeCount, nil
}
