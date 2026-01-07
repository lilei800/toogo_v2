/**
 * 市场监控接口
 */
import { http } from '@/utils/http/axios';

/**
 * 获取实时行情
 */
export function getMarketTicker(params: { symbol: string; exchange: string }) {
  return http.request({
    url: '/admin/trading/monitor/ticker',
    method: 'get',
    params,
  });
}

/**
 * 获取K线数据
 */
export function getKlineData(params: {
  symbol: string;
  exchange: string;
  interval: string;
  limit?: number;
}) {
  return http.request({
    url: '/admin/trading/monitor/kline',
    method: 'get',
    params,
  });
}

/**
 * 获取监控日志
 */
export function getMonitorLogs(params?: any) {
  return http.request({
    url: '/admin/trading/monitor/logs',
    method: 'get',
    params,
  });
}
