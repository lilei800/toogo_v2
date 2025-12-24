// Toogo API 接口
import { http } from '@/utils/http/axios';

// ========== 钱包相关 ==========
export const ToogoWalletApi = {
  // 钱包概览
  overview: (params?: any) => http.request({ url: '/toogo/wallet/overview', method: 'get', params }),
  // 钱包流水
  logList: (params?: any) => http.request({ url: '/toogo/wallet/log/list', method: 'get', params }),
  // 账户互转
  transfer: (data: any) => http.request({ url: '/toogo/wallet/transfer', method: 'post', data }),
  // 历史交易订单列表
  orderHistory: (params?: any) => http.request({ url: '/toogo/wallet/order/history', method: 'get', params }),
  // 成交流水列表（交易所成交明细）
  tradeHistory: (params?: any) => http.request({ url: '/toogo/wallet/trade/history', method: 'get', params }),
  // 运行区间盈亏汇总（按机器人运行区间统计交易所数据）
  runSessionSummary: (params?: any) => http.request({ url: '/toogo/wallet/run-session/summary', method: 'get', params }),
  // 同步运行区间盈亏数据（从交易所成交记录拉取并汇总到区间）
  syncRunSession: (data: { sessionId: number }) =>
    http.request({ url: '/toogo/wallet/run-session/sync', method: 'post', data }),
};

// ========== 订阅相关 ==========
export const ToogoSubscriptionApi = {
  // 套餐列表
  planList: (params?: any) => http.request({ url: '/toogo/plan/list', method: 'get', params }),
  // 编辑套餐
  planEdit: (data: any) => http.request({ url: '/toogo/plan/edit', method: 'post', data }),
  // 删除套餐
  planDelete: (data: any) => http.request({ url: '/toogo/plan/delete', method: 'post', data }),
  // 订阅套餐
  subscribe: (data: any) => http.request({ url: '/toogo/subscription/subscribe', method: 'post', data }),
  // 订阅记录
  subscriptionList: (params?: any) => http.request({ url: '/toogo/subscription/list', method: 'get', params }),
  // 我的订阅
  mySubscription: () => http.request({ url: '/toogo/subscription/my', method: 'get' }),
};

// ========== 用户相关 ==========
export const ToogoUserApi = {
  // 用户信息
  info: (params?: any) => http.request({ url: '/toogo/user/info', method: 'get', params }),
  // 用户列表
  list: (params?: any) => http.request({ url: '/toogo/user/list', method: 'get', params }),
  // 刷新邀请码
  refreshInviteCode: () => http.request({ url: '/toogo/user/refresh-invite-code', method: 'post' }),
  // 团队列表
  teamList: (params?: any) => http.request({ url: '/toogo/user/team/list', method: 'get', params }),
  // 团队统计
  teamStat: () => http.request({ url: '/toogo/user/team/stat', method: 'get' }),
  // VIP等级列表
  vipLevelList: (params?: any) => http.request({ url: '/toogo/vip-level/list', method: 'get', params }),
  // 编辑VIP等级
  vipLevelEdit: (data: any) => http.request({ url: '/toogo/vip-level/edit', method: 'post', data }),
  // 检查VIP升级
  checkVipUpgrade: () => http.request({ url: '/toogo/user/check-vip-upgrade', method: 'post' }),
  // 管理员手动充值算力
  adminRechargePower: (data: { userId: number; amount: number; remark?: string }) =>
    http.request({ url: '/toogo/admin/recharge-power', method: 'post', data }),
  // 管理员手动充值余额
  adminRechargeBalance: (data: { userId: number; amount: number; remark?: string }) =>
    http.request({ url: '/toogo/admin/recharge-balance', method: 'post', data }),
  // 管理员手动充值积分
  adminRechargePoints: (data: { userId: number; amount: number; remark?: string }) =>
    http.request({ url: '/toogo/admin/recharge-points', method: 'post', data }),
};

// ========== 佣金相关 ==========
export const ToogoCommissionApi = {
  // 佣金记录
  logList: (params?: any) => http.request({ url: '/toogo/commission/log/list', method: 'get', params }),
  // 佣金统计
  stat: () => http.request({ url: '/toogo/commission/stat', method: 'get' }),
  // 代理商等级列表
  agentLevelList: (params?: any) => http.request({ url: '/toogo/agent-level/list', method: 'get', params }),
  // 编辑代理商等级
  agentLevelEdit: (data: any) => http.request({ url: '/toogo/agent-level/edit', method: 'post', data }),
  // 申请成为代理商
  applyAgent: (data: { remark?: string; subscribeRate: number }) => http.request({ url: '/toogo/agent/apply', method: 'post', data }),
  // 代下级提交代理申请（仅直属下级）
  applyAgentForSub: (data: { subUserId: number; remark?: string; subscribeRate: number }) =>
    http.request({ url: '/toogo/agent/applyForSub', method: 'post', data }),
  // 获取代理信息
  getAgentInfo: () => http.request({ url: '/toogo/agent/info', method: 'get' }),
  // 设置下级代理佣金比例
  setSubAgentRate: (data: { subUserId: number; subscribeRate: number }) =>
    http.request({ url: '/toogo/agent/setSubRate', method: 'post', data }),
  // 审批代理商申请（管理员）
  approveAgent: (data: {
    memberId: number;
    approved: boolean;
    subscribeRate?: number;
    agentUnlockLevel?: number;
    rejectReason?: string;
  }) => http.request({ url: '/toogo/agent/approve', method: 'post', data }),
  // 更新代理商信息（管理员）
  updateAgent: (data: {
    memberId: number;
    isAgent?: number;
    agentStatus?: number;
    agentUnlockLevel?: number;
    subscribeRate?: number;
  }) => http.request({ url: '/toogo/agent/update', method: 'post', data }),
};

// ========== 策略相关 ==========
export const ToogoStrategyApi = {
  // 策略模板组列表
  groupList: (params?: any) => http.request({ url: '/strategy/group/list', method: 'get', params }),
  // 创建策略模板组
  groupCreate: (data: any) => http.request({ url: '/strategy/group/create', method: 'post', data }),
  // 更新策略模板组
  groupUpdate: (data: any) => http.request({ url: '/strategy/group/update', method: 'post', data }),
  // 删除策略模板组
  groupDelete: (data: any) => http.request({ url: '/strategy/group/delete', method: 'post', data }),
  // 初始化策略组
  groupInit: (data: any) => http.request({ url: '/strategy/group/initStrategies', method: 'post', data }),
  // 从官方复制策略模板
  copyFromOfficial: (data: any) => http.request({ url: '/strategy/group/copyFromOfficial', method: 'post', data }),
  // 策略模板列表
  templateList: (params?: any) => http.request({ url: '/strategy/template/list', method: 'get', params }),
  // 创建策略模板
  templateCreate: (data: any) => http.request({ url: '/strategy/template/create', method: 'post', data }),
  // 更新策略模板
  templateUpdate: (data: any) => http.request({ url: '/strategy/template/update', method: 'post', data }),
  // 删除策略模板
  templateDelete: (data: any) => http.request({ url: '/strategy/template/delete', method: 'post', data }),
  // 应用策略模板到机器人
  applyTemplate: (data: any) => http.request({ url: '/strategy/template/apply', method: 'post', data }),
  // 编辑策略模板(兼容)
  templateEdit: (data: any) => http.request({ url: '/toogo/strategy/template/edit', method: 'post', data }),
  // 获取策略
  getStrategy: (params?: any) => http.request({ url: '/toogo/strategy/get', method: 'get', params }),
  // 算力消耗记录
  powerConsumeList: (params?: any) => http.request({ url: '/toogo/power-consume/list', method: 'get', params }),
  // 算力消耗统计
  powerConsumeStat: () => http.request({ url: '/toogo/power-consume/stat', method: 'get' }),
};

// ========== 波动率配置相关（支持每个货币对独立配置） ==========
export const ToogoVolatilityConfigApi = {
  // 波动率配置列表（支持按货币对筛选）
  list: (params?: any) => http.request({ url: '/volatility/config/list', method: 'get', params }),
  // 创建波动率配置（支持全局配置和货币对特定配置）
  // 设置 isTransformResponse: false 以便在 vue 中自定义错误处理
  create: (data: any) => http.request({ url: '/volatility/config/create', method: 'post', data }, { isTransformResponse: false }),
  // 更新波动率配置
  // 设置 isTransformResponse: false 以便在 vue 中自定义错误处理
  update: (data: any) => http.request({ url: '/volatility/config/update', method: 'post', data }, { isTransformResponse: false }),
  // 删除波动率配置
  delete: (data: any) => http.request({ url: '/volatility/config/delete', method: 'post', data }),
  // 批量编辑波动率配置（为多个货币对批量设置）
  batchEdit: (data: any) => http.request({ url: '/volatility/config/batch-edit', method: 'post', data }),
  // 获取所有已配置的交易对列表
  getSymbols: () => http.request({ url: '/volatility/config/symbols', method: 'get' }),
};

// ========== 充值提现相关 ==========
export const ToogoFinanceApi = {
  // 创建充值订单
  createDeposit: (data: any) => http.request({ url: '/toogo/finance/deposit/create', method: 'post', data }),
  // 充值记录
  depositList: (params?: any) => http.request({ url: '/toogo/finance/deposit/list', method: 'get', params }),
  // 创建提现申请
  createWithdraw: (data: any) => http.request({ url: '/toogo/finance/withdraw/create', method: 'post', data }),
  // 提现记录
  withdrawList: (params?: any) => http.request({ url: '/toogo/finance/withdraw/list', method: 'get', params }),
  // 提现审核
  withdrawAudit: (data: any) => http.request({ url: '/toogo/finance/withdraw/audit', method: 'post', data }),
};

// ========== 机器人相关 ==========
export const ToogoRobotApi = {
  // 机器人列表
  list: (params?: any) => http.request({ url: '/trading/robot/list', method: 'get', params }),
  // 启动机器人
  start: (data: any) => http.request({ url: '/trading/robot/start', method: 'post', data }),
  // 停止机器人
  stop: (data: any) => http.request({ url: '/trading/robot/stop', method: 'post', data }),
  // 暂停机器人
  pause: (data: any) => http.request({ url: '/trading/robot/pause', method: 'post', data }),
  // 创建机器人
  create: (data: any) => http.request({ url: '/trading/robot/create', method: 'post', data }),
  // 更新机器人
  update: (data: any) => http.request({ url: '/trading/robot/update', method: 'post', data }),
  // 删除机器人
  delete: (data: any) => http.request({ url: '/trading/robot/delete', method: 'post', data }),
  // 机器人详情
  detail: (params?: any) => http.request({ url: '/trading/robot/view', method: 'get', params }),
  // 机器人统计
  stats: (params?: any) => http.request({ url: '/trading/robot/stats', method: 'get', params }),
  // 获取实时持仓
  positions: (params?: any) => http.request({ url: '/trading/robot/positions', method: 'get', params }),
  // 获取订单列表
  orders: (params?: any) => http.request({ url: '/trading/robot/orders', method: 'get', params }),
  // 获取订单历史
  orderHistory: (params?: any) => http.request({ url: '/trading/robot/orderHistory', method: 'get', params }),
  // 手动平仓
  closePosition: (data: any) => http.request({ url: '/trading/robot/closePosition', method: 'post', data }),
  // 撤销挂单
  cancelOrder: (data: { robotId: number; orderId: string }) => http.request({ url: '/trading/robot/cancelOrder', method: 'post', data }),
  // 设置止盈回撤开关
  setTakeProfitSwitch: (data: { robotId: number; positionSide: string; enabled: boolean }) => http.request({ url: '/trading/robot/setTakeProfitSwitch', method: 'post', data }),
  // 获取方向预警日志
  signalLogs: (params: { robotId: number; limit?: number }) => http.request({ url: '/trading/robot/signalLogs', method: 'get', params }),
  // 交易执行日志
  executionLogs: (params: { robotId: number; limit?: number }) => http.request({ url: '/trading/robot/executionLogs', method: 'get', params }),
  // 获取风险偏好配置
  getRiskConfig: (params: { robotId: number }) => http.request({ url: '/trading/robot/riskConfig', method: 'get', params }),
  // 保存风险偏好配置
  saveRiskConfig: (data: { robotId: number; config: any }) => http.request({ url: '/trading/robot/riskConfig/save', method: 'post', data }),
  // 重新加载策略配置（运行中生效）
  reloadStrategy: (params: { id: number }) => http.request({ url: '/trading/robot/reloadStrategy', method: 'post', params }),
};

// ========== 交易所相关 ==========
export const ToogoExchangeApi = {
  // 获取行情
  ticker: (params?: any) => http.request({ url: '/trading/monitor/ticker', method: 'get', params }),
  // 获取K线
  klines: (params?: any) => http.request({ url: '/trading/monitor/kline', method: 'get', params }),
  // 获取市场状态
  marketState: (params?: any) => http.request({ url: '/trading/monitor/marketState', method: 'get', params }),
  // 获取交易对列表
  symbols: (params?: any) => http.request({ url: '/trading/apiConfig/platforms', method: 'get', params }),
  // 获取机器人实时分析数据
  robotAnalysis: (params: { robotId: number }) => http.request({ url: '/trading/monitor/robotAnalysis', method: 'get', params }),
  // 批量获取机器人实时分析数据
  batchRobotAnalysis: (params: { robotIds: string }) => http.request({ url: '/trading/monitor/batchRobotAnalysis', method: 'get', params }),
};

// ========== API密钥相关 ==========
export const ToogoApiKeyApi = {
  // API密钥列表
  list: (params?: any) => http.request({ url: '/trading/apiConfig/list', method: 'get', params }),
  // 添加API密钥
  add: (data: any) => http.request({ url: '/trading/apiConfig/create', method: 'post', data }),
  // 更新API密钥
  update: (data: any) => http.request({ url: '/trading/apiConfig/update', method: 'post', data }),
  // 删除API密钥
  delete: (data: any) => http.request({ url: '/trading/apiConfig/delete', method: 'post', data }),
  // 设为默认
  setDefault: (data: any) => http.request({ url: '/trading/apiConfig/setDefault', method: 'post', data }),
  // 测试API密钥
  test: (data: any) => http.request({ url: '/trading/apiConfig/test', method: 'post', data }),
  // 获取支持的平台
  platforms: () => http.request({ url: '/trading/apiConfig/platforms', method: 'get' }),
};

// ========== API配置相关 (兼容旧代码) ==========
export const ToogoApiConfigApi = {
  // API配置列表
  list: (params?: any) => http.request({ url: '/trading/apiConfig/list', method: 'get', params }),
  // 创建API配置
  create: (data: any) => http.request({ url: '/trading/apiConfig/create', method: 'post', data }),
  // 更新API配置
  update: (data: any) => http.request({ url: '/trading/apiConfig/update', method: 'post', data }),
  // 删除API配置
  delete: (data: any) => http.request({ url: '/trading/apiConfig/delete', method: 'post', data }),
  // 查看API配置详情
  detail: (params?: any) => http.request({ url: '/trading/apiConfig/view', method: 'get', params }),
  // 测试API连接
  test: (data: any) => http.request({ url: '/trading/apiConfig/test', method: 'post', data }),
  // 设为默认配置
  setDefault: (data: any) => http.request({ url: '/trading/apiConfig/setDefault', method: 'post', data }),
  // 获取支持的平台列表
  platforms: (params?: any) => http.request({ url: '/trading/apiConfig/platforms', method: 'get', params }),
};
