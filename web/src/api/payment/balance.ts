/**
 * USDT余额API
 */
import { http } from '@/utils/http/axios';

/**
 * 查看余额
 */
export function getBalance() {
  return http.request({
    url: '/payment/balance/view',
    method: 'get',
  });
}

/**
 * 资金流水列表
 */
export function balanceLogList(params: {
  page: number;
  pageSize: number;
  type?: number;
  orderSn?: string;
}) {
  return http.request({
    url: '/payment/balance/logs',
    method: 'get',
    params,
  });
}
