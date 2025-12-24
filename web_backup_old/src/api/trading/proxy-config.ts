/**
 * 代理配置接口
 */
import { http } from '@/utils/http/axios';

/**
 * 获取代理配置
 */
export function getProxyConfig() {
  return http.request({
    url: '/admin/trading/proxy-config/get',
    method: 'get',
  });
}

/**
 * 保存代理配置
 */
export function saveProxyConfig(data: any) {
  return http.request({
    url: '/admin/trading/proxy-config/save',
    method: 'post',
    data,
  });
}

/**
 * 测试代理连接
 */
export function testProxyConfig() {
  return http.request({
    url: '/admin/trading/proxy-config/test',
    method: 'post',
  });
}

/**
 * 切换代理状态
 */
export function toggleProxyConfig() {
  return http.request({
    url: '/admin/trading/proxy-config/toggle',
    method: 'post',
  });
}

