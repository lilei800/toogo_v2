/**
 * USDT充值API
 */
import { http } from '@/utils/http/axios';

/**
 * 创建充值订单
 */
export function createDeposit(params: { amount: number; network: string }) {
  return http.request({
    url: '/payment/deposit/create',
    method: 'post',
    params,
  });
}

/**
 * 充值订单列表
 */
export function depositList(params: {
  page: number;
  pageSize: number;
  status?: number;
  orderSn?: string;
  network?: string;
}) {
  return http.request({
    url: '/payment/deposit/list',
    method: 'get',
    params,
  });
}

/**
 * 查看充值订单
 */
export function depositView(params: { id: number }) {
  return http.request({
    url: '/payment/deposit/view',
    method: 'get',
    params,
  });
}

/**
 * 检查充值状态
 */
export function checkDeposit(params: { id: number }) {
  return http.request({
    url: '/payment/deposit/check',
    method: 'post',
    params,
  });
}

/**
 * 取消充值订单
 */
export function cancelDeposit(params: { id: number }) {
  return http.request({
    url: '/payment/deposit/cancel',
    method: 'post',
    params,
  });
}
