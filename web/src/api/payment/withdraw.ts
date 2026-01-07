/**
 * USDT提现API
 */
import { http } from '@/utils/http/axios';

/**
 * 申请提现
 */
export function applyWithdraw(params: { amount: number; toAddress: string; network: string }) {
  return http.request({
    url: '/payment/withdraw/apply',
    method: 'post',
    params,
  });
}

/**
 * 提现订单列表
 */
export function withdrawList(params: {
  page: number;
  pageSize: number;
  status?: number;
  orderSn?: string;
  toAddress?: string;
  network?: string;
}) {
  return http.request({
    url: '/payment/withdraw/list',
    method: 'get',
    params,
  });
}

/**
 * 查看提现订单
 */
export function withdrawView(params: { id: number }) {
  return http.request({
    url: '/payment/withdraw/view',
    method: 'get',
    params,
  });
}

/**
 * 审核提现（管理员）
 */
export function auditWithdraw(params: { id: number; status: number; remark?: string }) {
  return http.request({
    url: '/payment/withdraw/audit',
    method: 'post',
    params,
  });
}

/**
 * 检查提现状态
 */
export function checkWithdraw(params: { id: number }) {
  return http.request({
    url: '/payment/withdraw/check',
    method: 'post',
    params,
  });
}

/**
 * 取消提现
 */
export function cancelWithdraw(params: { id: number }) {
  return http.request({
    url: '/payment/withdraw/cancel',
    method: 'post',
    params,
  });
}
