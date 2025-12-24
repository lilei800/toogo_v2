/**
 * 交易订单接口
 */
import { http } from '@/utils/http/axios';

/**
 * 获取订单列表
 */
export function getOrderList(params?: any) {
  return http.request({
    url: '/admin/trading/order/list',
    method: 'get',
    params,
  });
}

/**
 * 查看订单详情
 */
export function getOrderDetail(params: { id: number }) {
  return http.request({
    url: '/admin/trading/order/view',
    method: 'get',
    params,
  });
}

/**
 * 获取持仓订单
 */
export function getPositions(params?: any) {
  return http.request({
    url: '/admin/trading/order/positions',
    method: 'get',
    params,
  });
}

/**
 * 手动平仓
 */
export function manualCloseOrder(data: any) {
  return http.request({
    url: '/admin/trading/order/manual-close',
    method: 'post',
    data,
  });
}

/**
 * 获取订单统计
 */
export function getOrderStats(params?: any) {
  return http.request({
    url: '/admin/trading/order/stats',
    method: 'get',
    params,
  });
}

/**
 * 获取平仓日志列表
 */
export function getCloseLogs(params?: any) {
  return http.request({
    url: '/admin/trading/order/close-logs',
    method: 'get',
    params,
  });
}

/**
 * 查看平仓日志详情
 */
export function getCloseLogDetail(params: { id: number }) {
  return http.request({
    url: '/admin/trading/order/close-log-view',
    method: 'get',
    params,
  });
}

