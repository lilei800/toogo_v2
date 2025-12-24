/**
 * Toogo 机器人相关 API
 */
import { http } from '@/utils/http/axios';

/**
 * 获取机器人列表
 */
export function getRobotList(params?: any) {
  return http.request({
    url: '/toogo/robot/list',
    method: 'get',
    params,
  });
}

/**
 * 创建机器人
 */
export function createRobot(data: any) {
  return http.request({
    url: '/toogo/robot/create',
    method: 'post',
    data,
  });
}

/**
 * 更新机器人
 */
export function updateRobot(data: any) {
  return http.request({
    url: '/toogo/robot/update',
    method: 'post',
    data,
  });
}

/**
 * 删除机器人
 */
export function deleteRobot(params: { id: number }) {
  return http.request({
    url: '/toogo/robot/delete',
    method: 'post',
    params,
  });
}

/**
 * 启动机器人
 */
export function startRobot(params: { id: number }) {
  return http.request({
    url: '/toogo/robot/start',
    method: 'post',
    params,
  });
}

/**
 * 停止机器人
 */
export function stopRobot(params: { id: number }) {
  return http.request({
    url: '/toogo/robot/stop',
    method: 'post',
    params,
  });
}

/**
 * 暂停机器人
 */
export function pauseRobot(params: { id: number }) {
  return http.request({
    url: '/toogo/robot/pause',
    method: 'post',
    params,
  });
}

/**
 * 获取机器人引擎状态
 */
export function getRobotEngineStatus(params: { robotId: number }) {
  return http.request({
    url: '/toogo/robot/engine-status',
    method: 'get',
    params,
  });
}

/**
 * 获取机器人信号日志
 */
export function getRobotSignalLogs(params: { robotId: number; limit?: number }) {
  return http.request({
    url: '/toogo/robot/signal-logs',
    method: 'get',
    params,
  });
}

/**
 * 关闭持仓
 */
export function closePosition(data: { robotId: number; positionSide?: string }) {
  return http.request({
    url: '/toogo/robot/close-position',
    method: 'post',
    data,
  });
}

/**
 * 获取机器人详情
 */
export function getRobotDetail(params: { id: number }) {
  return http.request({
    url: '/toogo/robot/detail',
    method: 'get',
    params,
  });
}

/**
 * 获取机器人统计
 */
export function getRobotStats(params: { id: number }) {
  return http.request({
    url: '/toogo/robot/stats',
    method: 'get',
    params,
  });
}

/**
 * 重新加载机器人策略配置
 */
export function reloadStrategy(params: { id: number }) {
  return http.request({
    url: '/toogo/robot/reload-strategy',
    method: 'post',
    params,
  });
}

