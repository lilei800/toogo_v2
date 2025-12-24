/**
 * 机器人状态监控逻辑
 */
import { ref, onMounted, onUnmounted, watch } from 'vue';
import { getRobotEngineStatus, getRobotSignalLogs, closePosition } from '@/api/toogo/robot';
import type { Robot } from './useRobotList';

export interface EngineStatus {
  robotId: number;
  symbol: string;
  platform: string;
  running: boolean;
  connected: boolean;
  lastPrice: number;
  totalBalance: number;
  availBalance: number;
  marketState: string;
  trendDirection: string;
  volatility: number;
  // winProbability 已移除（不再由 RiskManager 动态计算）
  signalDirection: string;
  signalStrength: number;
  signalConfidence: number;
  hasPosition: boolean;
  positionSide: string;
  positionAmt: number;
  entryPrice: number;
  unrealizedPnl: number;
  priceWindowData: { timestamp: number; price: number }[];
  windowMinPrice: number;
  windowMaxPrice: number;
  windowCurrentPrice: number;
  longTriggerPrice: number;
  shortTriggerPrice: number;
  signalProgress: number;
  signalReason: string;
  strategyWindow: number;
  strategyThreshold: number;
  currentMarketState: string;
  currentRiskPref: string;
}

export interface AnalysisData {
  signal: {
    direction: string;
    signalStrength: number;
    signalConfidence: number;
    currentMarketState: string;
    currentRiskPref: string;
    windowMinPrice: number;
    windowMaxPrice: number;
    currentPrice: number;
    distanceFromMin: number;
    distanceFromMax: number;
    signalThreshold: number;
    signalProgress: number;
    reason: string;
    strategyWindow: number;
    strategyThreshold: number;
  };
  config: {
    autoTradeEnabled: boolean;
    autoCloseEnabled: boolean;
    leverage: number;
    marginPercent: number;
    stopLossPercent: number;
    takeProfitPercent: number;
    totalProfit: number;
    runtimeSeconds: number;
  };
  account: {
    totalBalance: number;
    availableBalance: number;
    unrealizedPnl: number;
  };
  priceWindow: { timestamp: number; price: number }[];
  lastUpdate: number;
}

export interface TickerData {
  lastPrice: number;
  change24h: number;
}

export interface PositionData {
  symbol: string;
  positionSide: string;
  positionAmt: number;
  entryPrice: number;
  markPrice: number;
  unrealizedPnl: number;
  leverage: number;
}

export interface SignalLog {
  id: number;
  robotId: number;
  signalType: string;
  signalStrength: number;
  currentPrice: number;
  reason: string;
  createdAt: string;
}

export function useRobotStatus(robotList: { value: Robot[] }) {
  // 状态数据
  const analysisData = ref<Record<number, AnalysisData>>({});
  const tickerData = ref<Record<number, TickerData>>({});
  const positionData = ref<Record<number, PositionData[]>>({});
  const signalLogs = ref<Record<number, SignalLog[]>>({});
  
  let refreshTimer: ReturnType<typeof setInterval> | null = null;

  // 获取机器人引擎状态
  const fetchEngineStatus = async (robotId: number) => {
    try {
      const res = await getRobotEngineStatus({ robotId });
      if (res.code === 0 && res.data) {
        const status: EngineStatus = res.data;
        
        // 更新分析数据
        analysisData.value[robotId] = {
          signal: {
            direction: status.signalDirection,
            signalStrength: status.signalStrength,
            signalConfidence: status.signalConfidence,
            currentMarketState: status.currentMarketState,
            currentRiskPref: status.currentRiskPref,
            windowMinPrice: status.windowMinPrice,
            windowMaxPrice: status.windowMaxPrice,
            currentPrice: status.windowCurrentPrice,
            distanceFromMin: status.windowCurrentPrice - status.windowMinPrice,
            distanceFromMax: status.windowMaxPrice - status.windowCurrentPrice,
            signalThreshold: status.strategyThreshold,
            signalProgress: status.signalProgress,
            reason: status.signalReason,
            strategyWindow: status.strategyWindow,
            strategyThreshold: status.strategyThreshold,
          },
          config: {
            autoTradeEnabled: true,
            autoCloseEnabled: true,
            leverage: 10,
            marginPercent: 10,
            stopLossPercent: 5,
            takeProfitPercent: 30,
            totalProfit: 0,
            runtimeSeconds: 0,
          },
          account: {
            totalBalance: status.totalBalance,
            availableBalance: status.availBalance,
            unrealizedPnl: status.unrealizedPnl,
          },
          priceWindow: status.priceWindowData || [],
          lastUpdate: Date.now(),
        };
        
        // 更新行情数据
        tickerData.value[robotId] = {
          lastPrice: status.lastPrice,
          change24h: 0,
        };
        
        // 更新持仓数据
        if (status.hasPosition) {
          positionData.value[robotId] = [{
            symbol: status.symbol,
            positionSide: status.positionSide,
            positionAmt: status.positionAmt,
            entryPrice: status.entryPrice,
            markPrice: status.lastPrice,
            unrealizedPnl: status.unrealizedPnl,
            leverage: 10,
          }];
        } else {
          positionData.value[robotId] = [];
        }
      }
    } catch (error) {
      console.error(`获取机器人${robotId}状态失败:`, error);
    }
  };

  // 获取信号日志
  const fetchSignalLogs = async (robotId: number) => {
    try {
      const res = await getRobotSignalLogs({ robotId, limit: 10 });
      if (res.code === 0 && res.data) {
        signalLogs.value[robotId] = res.data;
      }
    } catch (error) {
      console.error(`获取机器人${robotId}信号日志失败:`, error);
    }
  };

  // 刷新运行中机器人的状态
  const refreshRunningRobots = async () => {
    const runningRobots = robotList.value.filter(r => r.status === 2);
    await Promise.all(runningRobots.map(async (robot) => {
      await fetchEngineStatus(robot.id);
      await fetchSignalLogs(robot.id);
    }));
  };

  // 启动定时刷新
  const startRefresh = () => {
    stopRefresh();
    refreshRunningRobots();
    refreshTimer = setInterval(refreshRunningRobots, 2000); // 每2秒刷新
  };

  // 停止定时刷新
  const stopRefresh = () => {
    if (refreshTimer) {
      clearInterval(refreshTimer);
      refreshTimer = null;
    }
  };

  // 获取连接状态
  const getConnectionStatus = (robotId: number) => {
    const data = analysisData.value[robotId];
    if (!data) {
      return { class: 'warning', text: '连接中...' };
    }
    const lastUpdate = data.lastUpdate || 0;
    const isRecent = Date.now() - lastUpdate < 10000;
    return {
      class: isRecent ? 'success' : 'warning',
      text: isRecent ? '已连接' : '连接中...',
    };
  };

  // 监听机器人列表变化
  watch(() => robotList.value, () => {
    if (robotList.value.some(r => r.status === 2)) {
      startRefresh();
    }
  }, { immediate: true });

  onMounted(() => {
    if (robotList.value.some(r => r.status === 2)) {
      startRefresh();
    }
  });

  onUnmounted(() => {
    stopRefresh();
  });

  return {
    // 状态
    analysisData,
    tickerData,
    positionData,
    signalLogs,
    
    // 方法
    fetchEngineStatus,
    fetchSignalLogs,
    refreshRunningRobots,
    getConnectionStatus,
    startRefresh,
    stopRefresh,
  };
}

// 格式化相关函数
export function formatMarketState(state: string): string {
  const map: Record<string, string> = {
    'trend': '趋势市场',
    'range': '震荡市场',    // 添加 range 映射
    'volatile': '震荡市场',
    'high_vol': '高波动',
    'low_vol': '低波动',
  };
  return map[state] || state || '--';
}

export function formatRiskPref(pref: string): string {
  const map: Record<string, string> = {
    'conservative': '保守型',
    'balanced': '平衡型',
    'aggressive': '激进型',
  };
  return map[pref] || pref || '--';
}

export function getMarketStateType(state: string): 'success' | 'warning' | 'error' | 'info' | 'default' {
  switch (state) {
    case 'trend': return 'success';
    case 'range': return 'warning';      // 添加 range 类型映射（震荡-警告色）
    case 'volatile': return 'warning';
    case 'high_vol': return 'error';
    case 'low_vol': return 'info';
    default: return 'default';
  }
}

export function formatPrice(price: number | undefined): string {
  if (price === undefined || price === null) return '--';
  return price.toFixed(2);
}

export function formatPriceUsdt(price: number | undefined): string {
  if (price === undefined || price === null) return '--';
  return price.toFixed(1) + 'U';
}

export function formatWindowTime(seconds: number | undefined): string {
  if (!seconds) return '--';
  if (seconds >= 60) {
    return `${Math.floor(seconds / 60)}分${seconds % 60}秒`;
  }
  return `${seconds}秒`;
}

export function formatRuntime(seconds: number | undefined): string {
  if (!seconds) return '0秒';
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  if (hours > 0) {
    return `${hours}小时${minutes}分`;
  }
  if (minutes > 0) {
    return `${minutes}分钟`;
  }
  return `${seconds}秒`;
}

export function formatLogTime(time: string): string {
  if (!time) return '--';
  const date = new Date(time);
  return `${date.getHours().toString().padStart(2, '0')}:${date.getMinutes().toString().padStart(2, '0')}`;
}

export function formatUpdateTime(timestamp: number | undefined): string {
  if (!timestamp) return '--';
  const date = new Date(timestamp);
  return `${date.getHours().toString().padStart(2, '0')}:${date.getMinutes().toString().padStart(2, '0')}:${date.getSeconds().toString().padStart(2, '0')}`;
}

export function formatPriceChange(change: number | undefined): string {
  if (change === undefined || change === null) return '';
  const sign = change >= 0 ? '+' : '';
  return `${sign}${(change * 100).toFixed(2)}%`;
}

export function getPriceChangeClass(change: number | undefined): string {
  if (change === undefined || change === null) return '';
  return change >= 0 ? 'up' : 'down';
}

