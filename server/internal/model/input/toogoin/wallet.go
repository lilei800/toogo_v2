// Package toogoin
// @Link  https://github.com/bufanyun/hotgo
// @Copyright  Copyright (c) 2024 Toogo.Ai
// @Author  Toogo Team
package toogoin

import (
	"hotgo/internal/model/entity"
	"hotgo/internal/model/input/form"
)

// WalletGetInp 获取钱包信息输入
type WalletGetInp struct {
	UserId int64 `json:"userId" description:"用户ID"`
}

// WalletGetModel 获取钱包信息返回
type WalletGetModel struct {
	*entity.ToogoWallet
}

// WalletOverviewInp 钱包概览输入
type WalletOverviewInp struct {
	UserId int64 `json:"userId" description:"用户ID"`
}

// WalletOverviewModel 钱包概览返回
type WalletOverviewModel struct {
	Balance           float64 `json:"balance" description:"余额(USDT)"`
	FrozenBalance     float64 `json:"frozenBalance" description:"冻结余额"`
	Power             float64 `json:"power" description:"算力余额"`
	FrozenPower       float64 `json:"frozenPower" description:"冻结算力"`
	GiftPower         float64 `json:"giftPower" description:"积分余额"`
	Commission        float64 `json:"commission" description:"佣金余额(USDT)"`
	FrozenCommission  float64 `json:"frozenCommission" description:"冻结佣金"`
	TotalDeposit      float64 `json:"totalDeposit" description:"累计充值"`
	TotalWithdraw     float64 `json:"totalWithdraw" description:"累计提现"`
	TotalPowerConsume float64 `json:"totalPowerConsume" description:"累计消耗算力"`
	TotalCommission   float64 `json:"totalCommission" description:"累计获得佣金"`
	TotalPower        float64 `json:"totalPower" description:"总可用算力(算力+赠送)"`
}

// WalletLogListInp 钱包流水列表输入
type WalletLogListInp struct {
	form.PageReq
	UserId      int64    `json:"userId" description:"用户ID"`
	AccountType string   `json:"accountType" description:"账户类型"`
	ChangeType  string   `json:"changeType" description:"变动类型"`
	CreatedAt   []string `json:"createdAt" description:"创建时间"`
}

// WalletLogListModel 钱包流水列表返回
type WalletLogListModel struct {
	*entity.ToogoWalletLog
}

// TransferInp 账户互转输入
type TransferInp struct {
	UserId      int64   `json:"userId" description:"用户ID，用户端不传则自动获取当前登录用户"`
	FromAccount string  `json:"fromAccount" v:"required|in:balance,commission" description:"转出账户"`
	ToAccount   string  `json:"toAccount" d:"power" description:"转入账户，默认为算力账户"`
	Amount      float64 `json:"amount" v:"required|min:0.01" description:"转账金额(USDT)"`
}

// TransferModel 账户互转返回
type TransferModel struct {
	OrderSn     string  `json:"orderSn" description:"订单号"`
	Amount      float64 `json:"amount" description:"转账金额"`
	PowerAmount float64 `json:"powerAmount" description:"获得算力"`
}

// DepositCreateInp 创建充值订单输入
type DepositCreateInp struct {
	UserId  int64   `json:"userId" v:"required" description:"用户ID"`
	Amount  float64 `json:"amount" v:"required|min:1" description:"充值金额(USDT)"`
	Network string  `json:"network" v:"required|in:TRC20,ERC20,BEP20" description:"网络"`
}

// DepositCreateModel 创建充值订单返回
type DepositCreateModel struct {
	OrderSn    string  `json:"orderSn" description:"订单号"`
	Amount     float64 `json:"amount" description:"充值金额"`
	ToAddress  string  `json:"toAddress" description:"充值地址"`
	Network    string  `json:"network" description:"网络"`
	ExpireTime string  `json:"expireTime" description:"过期时间"`
}

// WithdrawCreateInp 创建提现订单输入
type WithdrawCreateInp struct {
	UserId      int64   `json:"userId" v:"required" description:"用户ID"`
	AccountType string  `json:"accountType" v:"required|in:balance,commission" description:"账户类型"`
	Amount      float64 `json:"amount" v:"required|min:10" description:"提现金额(USDT)"`
	ToAddress   string  `json:"toAddress" v:"required" description:"提现地址"`
	Network     string  `json:"network" v:"required|in:TRC20,ERC20,BEP20" description:"网络"`
}

// WithdrawCreateModel 创建提现订单返回
type WithdrawCreateModel struct {
	OrderSn    string  `json:"orderSn" description:"订单号"`
	Amount     float64 `json:"amount" description:"提现金额"`
	Fee        float64 `json:"fee" description:"手续费"`
	RealAmount float64 `json:"realAmount" description:"实际到账金额"`
}

// ChangeBalanceInp 变更余额输入
type ChangeBalanceInp struct {
	UserId      int64   `json:"userId" v:"required" description:"用户ID"`
	AccountType string  `json:"accountType" v:"required" description:"账户类型"`
	ChangeType  string  `json:"changeType" v:"required" description:"变动类型"`
	Amount      float64 `json:"amount" v:"required" description:"变动金额"`
	RelatedId   int64   `json:"relatedId" description:"关联ID"`
	RelatedType string  `json:"relatedType" description:"关联类型"`
	OrderSn     string  `json:"orderSn" description:"关联订单号"`
	Remark      string  `json:"remark" description:"备注"`
}

// AdminRechargePowerInp 管理员手动充值算力输入
type AdminRechargePowerInp struct {
	UserId int64   `json:"userId" v:"required" description:"目标用户ID"`
	Amount float64 `json:"amount" v:"required|min:0.01" description:"充值算力数量"`
	Remark string  `json:"remark" description:"充值备注"`
}

// AdminRechargePowerModel 管理员手动充值算力返回
type AdminRechargePowerModel struct {
	UserId     int64   `json:"userId" description:"用户ID"`
	Amount     float64 `json:"amount" description:"充值算力数量"`
	NewBalance float64 `json:"newBalance" description:"充值后算力余额"`
}

// AdminRechargeBalanceInp 管理员手动充值余额输入
type AdminRechargeBalanceInp struct {
	UserId int64   `json:"userId" v:"required" description:"目标用户ID"`
	Amount float64 `json:"amount" v:"required|min:0.01" description:"充值余额金额"`
	Remark string  `json:"remark" description:"充值备注"`
}

// AdminRechargeBalanceModel 管理员手动充值余额返回
type AdminRechargeBalanceModel struct {
	UserId     int64   `json:"userId" description:"用户ID"`
	Amount     float64 `json:"amount" description:"充值余额金额"`
	NewBalance float64 `json:"newBalance" description:"充值后余额"`
}

// AdminRechargePointsInp 管理员手动充值积分输入
type AdminRechargePointsInp struct {
	UserId int64   `json:"userId" v:"required" description:"目标用户ID"`
	Amount float64 `json:"amount" v:"required|min:0.01" description:"充值积分数量"`
	Remark string  `json:"remark" description:"充值备注"`
}

// AdminRechargePointsModel 管理员手动充值积分返回
type AdminRechargePointsModel struct {
	UserId     int64   `json:"userId" description:"用户ID"`
	Amount     float64 `json:"amount" description:"充值积分数量"`
	NewBalance float64 `json:"newBalance" description:"充值后积分余额"`
}

// UserWalletListModel 用户钱包列表返回
type UserWalletListModel struct {
	UserId     int64   `json:"userId" description:"用户ID"`
	Username   string  `json:"username" description:"用户名"`
	Mobile     string  `json:"mobile" description:"手机号"`
	Balance    float64 `json:"balance" description:"余额(USDT)"`
	Power      float64 `json:"power" description:"算力余额"`
	GiftPower  float64 `json:"giftPower" description:"积分"`
	TotalPower float64 `json:"totalPower" description:"总可用算力"`
	Commission float64 `json:"commission" description:"佣金余额"`
	VipLevel   int     `json:"vipLevel" description:"VIP等级"`
	CreatedAt  string  `json:"createdAt" description:"注册时间"`
}

// OrderHistoryListInp 历史交易订单列表输入
type OrderHistoryListInp struct {
	form.PageReq
	UserId    int64  `json:"userId" description:"用户ID（管理员可选，不传则默认查询全部用户；普通用户忽略该字段）"`
	RobotId   int64  `json:"robotId" description:"机器人ID（可选）"`
	Exchange  string `json:"exchange" description:"交易所（可选，如 binance/bitget/okx/gate）"`
	Symbol    string `json:"symbol" description:"交易对（可选）"`
	Direction string `json:"direction" description:"方向：long/short（可选）"`
	Status    int    `json:"status" description:"状态：1=持仓中,2=已平仓（可选）"`
	StartTime string `json:"startTime" description:"开始时间（可选）"`
	EndTime   string `json:"endTime" description:"结束时间（可选）"`
}

// OrderHistoryModel 历史交易订单模型
type OrderHistoryModel struct {
	Id              int64    `json:"id" description:"订单ID"`
	UserId          int64    `json:"userId" description:"用户ID"`
	Username        string   `json:"username" description:"用户名（可能为空）"`
	Exchange        string   `json:"exchange" description:"交易所"`
	RobotId         int64    `json:"robotId" description:"机器人ID"`
	RobotName       string   `json:"robotName" description:"机器人名称"`
	OrderSn         string   `json:"orderSn" description:"订单号"`
	ExchangeOrderId string   `json:"exchangeOrderId" description:"交易所订单ID（开仓订单ID）"`
	CloseOrderId    string   `json:"closeOrderId" description:"平仓订单ID（交易所订单ID，可能为空）"`
	Symbol          string   `json:"symbol" description:"交易对"`
	Direction       string   `json:"direction" description:"方向：long/short"`
	DirectionText   string   `json:"directionText" description:"方向文本：多/空"`
	OpenPrice       *float64 `json:"openPrice,omitempty" description:"开仓价格（优先平台成交汇总）"`
	ClosePrice      *float64 `json:"closePrice,omitempty" description:"平仓价格（优先平台成交汇总）"`
	Quantity        *float64 `json:"quantity,omitempty" description:"数量（优先平台成交汇总）"`
	OpenFee         *float64 `json:"openFee,omitempty" description:"开仓手续费（交易所成交汇总）"`
	CloseFee        *float64 `json:"closeFee,omitempty" description:"平仓手续费（交易所成交汇总）"`
	FeeTotal        *float64 `json:"feeTotal,omitempty" description:"手续费合计（openFee+closeFee，USDT口径；若非USDT则不返回）"`
	ProfitAmount    *float64 `json:"profitAmount,omitempty" description:"盈利金额(>=0)：max(realizedProfit,0)"`
	LossAmount      *float64 `json:"lossAmount,omitempty" description:"亏损金额(>=0)：abs(min(realizedProfit,0))"`
	RealizedProfit  *float64 `json:"realizedProfit,omitempty" description:"已实现盈亏（优先平台成交汇总）"`
	OpenTime        string   `json:"openTime" description:"开仓时间"`
	CloseTime       string   `json:"closeTime" description:"平仓时间"`
	Status          int      `json:"status" description:"状态：1=持仓中,2=已平仓,3=已取消"`
	StatusText      string   `json:"statusText" description:"状态文本"`
	CloseReasonText string   `json:"closeReasonText" description:"平仓原因文本"`
	CreatedAt       string   `json:"createdAt" description:"创建时间（可选）"`
}

// ========== 成交流水（交易所成交明细） ==========

// TradeHistoryListInp 成交流水列表输入（每条记录对应交易所一笔成交）
type TradeHistoryListInp struct {
	form.PageReq
	UserId      int64  `json:"userId" description:"用户ID（管理员可选，不传则默认查询全部用户；普通用户忽略该字段）"`
	RobotId     int64  `json:"robotId" description:"机器人ID（可选）"`
	SessionId   int64  `json:"sessionId" description:"运行区间ID（可选）"`
	ApiConfigId int64  `json:"apiConfigId" description:"API配置ID（可选）"`
	Exchange    string `json:"exchange" description:"交易所（可选，如 binance/bitget/okx/gate）"`
	Symbol      string `json:"symbol" description:"交易对（可选）"`
	OrderId     string `json:"orderId" description:"交易所订单ID（可选）"`
	TradeId     string `json:"tradeId" description:"成交ID（可选）"`
	Side        string `json:"side" description:"方向：BUY/SELL 或 OPEN/CLOSE（可选）"`
	StartTime   string `json:"startTime" description:"开始时间（可选）"`
	EndTime     string `json:"endTime" description:"结束时间（可选）"`
}

// TradeFillModel 成交流水模型（用于交易明细页展示）
type TradeFillModel struct {
	Id            int64   `json:"id" description:"流水ID"`
	ApiConfigId   int64   `json:"apiConfigId" description:"API配置ID"`
	Exchange      string  `json:"exchange" description:"交易所"`
	UserId        int64   `json:"userId" description:"用户ID"`
	Username      string  `json:"username" description:"用户名（可能为空）"`
	RobotId       int64   `json:"robotId" description:"机器人ID"`
	RobotName     string  `json:"robotName" description:"机器人名称"`
	SessionId     *int64  `json:"sessionId,omitempty" description:"运行区间ID(可选)"`
	Symbol        string  `json:"symbol" description:"交易对"`
	OrderId       string  `json:"orderId" description:"交易所订单ID"`
	ClientOrderId string  `json:"clientOrderId" description:"客户端订单ID(可选)"`
	TradeId       string  `json:"tradeId" description:"成交ID"`
	Side          string  `json:"side" description:"方向"`
	Qty           float64 `json:"qty" description:"成交数量"`
	Price         float64 `json:"price" description:"成交价格"`
	Fee           float64 `json:"fee" description:"手续费(正数)"`
	FeeCoin       string  `json:"feeCoin" description:"手续费币种"`
	RealizedPnl   float64 `json:"realizedPnl" description:"已实现盈亏"`
	Ts            int64   `json:"ts" description:"成交时间戳(毫秒)"`
	Time          string  `json:"time" description:"成交时间(格式化)"`
}

// TradeFillSummary 成交流水汇总统计
type TradeFillSummary struct {
	TotalCount  int     `json:"totalCount" description:"总成交笔数"`
	TotalPnl    float64 `json:"totalPnl" description:"总盈亏(USDT)"`
	TotalProfit float64 `json:"totalProfit" description:"总盈利(正数部分)"`
	TotalLoss   float64 `json:"totalLoss" description:"总亏损(负数部分)"`
	TotalFee    float64 `json:"totalFee" description:"总手续费(USDT)"`
	TotalNetPnl float64 `json:"totalNetPnl" description:"总净盈亏(扣手续费)"`
}

// ========== 运行区间盈亏汇总 ==========

// RunSessionSummaryListInp 运行区间汇总列表输入
type RunSessionSummaryListInp struct {
	form.PageReq
	RobotId  int64  `json:"robotId" description:"机器人ID（可选）"`
	Exchange string `json:"exchange" description:"交易所（可选）"`
	Symbol   string `json:"symbol" description:"交易对（可选）"`
	// IsRunning 区间状态筛选：0=全部，1=运行中（end_time is null），2=已结束（end_time not null）
	IsRunning int    `json:"isRunning" description:"区间状态：0=全部,1=运行中,2=已结束（可选）"`
	StartTime string `json:"startTime" description:"开始时间（可选）"`
	EndTime   string `json:"endTime" description:"结束时间（可选）"`
}

// RunSessionSummaryModel 运行区间汇总模型
type RunSessionSummaryModel struct {
	Id             int64    `json:"id" description:"区间ID"`
	RobotId        int64    `json:"robotId" description:"机器人ID"`
	RobotName      string   `json:"robotName" description:"机器人名称"`
	Exchange       string   `json:"exchange" description:"交易所"`
	Symbol         string   `json:"symbol" description:"交易对"`
	StartTime      string   `json:"startTime" description:"启动时间"`
	EndTime        string   `json:"endTime" description:"结束时间"`
	EndReason      string   `json:"endReason" description:"结束原因"`
	EndReasonText  string   `json:"endReasonText" description:"结束原因文本"`
	RuntimeSeconds int      `json:"runtimeSeconds" description:"运行时长(秒)"`
	RuntimeText    string   `json:"runtimeText" description:"运行时长文本"`
	TotalPnl       *float64 `json:"totalPnl" description:"区间总盈亏(USDT)"`
	TotalFee       *float64 `json:"totalFee" description:"区间总手续费(USDT)"`
	NetPnl         *float64 `json:"netPnl" description:"净盈亏(扣手续费)"`
	TradeCount     int      `json:"tradeCount" description:"成交笔数"`
	SyncedAt       string   `json:"syncedAt" description:"最后同步时间"`
	IsRunning      bool     `json:"isRunning" description:"是否仍在运行"`
}

// RunSessionTotalSummary 运行区间汇总统计
type RunSessionTotalSummary struct {
	TotalSessions    int     `json:"totalSessions" description:"总区间数"`
	TotalRuntime     int     `json:"totalRuntime" description:"总运行时长(秒)"`
	TotalRuntimeText string  `json:"totalRuntimeText" description:"总运行时长文本"`
	TotalPnl         float64 `json:"totalPnl" description:"总盈亏(USDT)"`
	TotalProfit      float64 `json:"totalProfit" description:"总盈利(正数部分)"`
	TotalLoss        float64 `json:"totalLoss" description:"总亏损(负数部分)"`
	TotalFee         float64 `json:"totalFee" description:"总手续费(USDT)"`
	TotalNetPnl      float64 `json:"totalNetPnl" description:"总净盈亏(USDT)"`
	TotalTrades      int     `json:"totalTrades" description:"总成交笔数"`
}
