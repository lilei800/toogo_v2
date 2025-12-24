import { http } from '@/utils/http/axios';

// 获取方向预警日志列表
export function getDirectionLogList(params: any) {
  return http.request({
    url: '/trading/alert/direction/list',
    method: 'get',
    params,
  });
}

// 获取机器人实时状态
export function getRobotRealtime(robotId: number) {
  return http.request({
    url: '/trading/alert/robot/realtime',
    method: 'get',
    params: { robotId },
  });
}

// 获取机器人实时状态列表
export function getRobotRealtimeList(params: any) {
  return http.request({
    url: '/trading/alert/robot/realtimeList',
    method: 'get',
    params,
  });
}

// 获取市场分析
export function getMarketAnalysis(platform: string, symbol: string) {
  return http.request({
    url: '/trading/alert/market/analysis',
    method: 'get',
    params: { platform, symbol },
  });
}

// 获取方向信号
export function getDirectionSignal(platform: string, symbol: string) {
  return http.request({
    url: '/trading/alert/direction/signal',
    method: 'get',
    params: { platform, symbol },
  });
}

// 获取风险评估
export function getRiskEvaluation(robotId: number) {
  return http.request({
    url: '/trading/alert/risk/evaluation',
    method: 'get',
    params: { robotId },
  });
}

// 获取引擎状态
export function getEngineStatus() {
  return http.request({
    url: '/trading/alert/engine/status',
    method: 'get',
  });
}

// 获取全局引擎详情
export function getGlobalEngineDetail() {
  return http.request({
    url: '/trading/alert/engine/detail',
    method: 'get',
  });
}

// 启动全局引擎
export function startGlobalEngine() {
  return http.request({
    url: '/trading/alert/engine/start',
    method: 'post',
  });
}

// 停止全局引擎
export function stopGlobalEngine() {
  return http.request({
    url: '/trading/alert/engine/stop',
    method: 'post',
  });
}

// 重启全局引擎
export function restartGlobalEngine() {
  return http.request({
    url: '/trading/alert/engine/restart',
    method: 'post',
  });
}
