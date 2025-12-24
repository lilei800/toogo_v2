/**
 * 交易机器人接口
 */
import { http } from '@/utils/http/axios';

/**
 * 获取机器人列表
 */
export function getRobotList(params?: any) {
  return http.request({
    url: '/admin/trading/robot/list',
    method: 'get',
    params,
  });
}

/**
 * 创建机器人
 */
export function createRobot(data: any) {
  return http.request({
    url: '/admin/trading/robot/create',
    method: 'post',
    data,
  });
}

/**
 * 更新机器人
 */
export function updateRobot(data: any) {
  return http.request({
    url: '/admin/trading/robot/update',
    method: 'post',
    data,
  });
}

/**
 * 删除机器人
 */
export function deleteRobot(params: { id: number }) {
  return http.request({
    url: '/admin/trading/robot/delete',
    method: 'post',
    params,
  });
}

/**
 * 获取机器人详情
 */
export function getRobotDetail(params: { id: number }) {
  return http.request({
    url: '/admin/trading/robot/view',
    method: 'get',
    params,
  });
}

/**
 * 启动机器人
 */
export function startRobot(params: { id: number }) {
  return http.request({
    url: '/admin/trading/robot/start',
    method: 'post',
    params,
  });
}

/**
 * 暂停机器人
 */
export function pauseRobot(params: { id: number }) {
  return http.request({
    url: '/admin/trading/robot/pause',
    method: 'post',
    params,
  });
}

/**
 * 停止机器人
 */
export function stopRobot(params: { id: number }) {
  return http.request({
    url: '/admin/trading/robot/stop',
    method: 'post',
    params,
  });
}

/**
 * 获取机器人统计
 */
export function getRobotStats(params: { id: number }) {
  return http.request({
    url: '/admin/trading/robot/stats',
    method: 'get',
    params,
  });
}

/**
 * 推荐策略
 */
export function recommendStrategy(data: any) {
  return http.request({
    url: '/admin/trading/robot/recommend-strategy',
    method: 'post',
    data,
  });
}

/**
 * 重新加载机器人策略配置（运行中生效）
 */
export function reloadStrategy(params: { id: number }) {
  return http.request({
    url: '/admin/trading/robot/reload-strategy',
    method: 'post',
    params,
  });
}
