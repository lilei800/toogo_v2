/**
 * 代理配置接口
 */
import { http } from '@/utils/http/axios';

/**
 * 获取代理配置
 */
export function getProxyConfig() {
  return http.request({
    // 由 axios 层自动拼接 /admin 前缀（urlPrefix），这里不要重复写 /admin
    url: '/trading/proxyConfig/get',
    method: 'get',
  });
}

/**
 * 保存代理配置
 */
export function saveProxyConfig(data: any) {
  return http.request({
    url: '/trading/proxyConfig/save',
    method: 'post',
    data,
  });
}

/**
 * 测试代理连接
 */
export function testProxyConfig() {
  return http.request({
    url: '/trading/proxyConfig/test',
    method: 'post',
  });
}

/**
 * 切换代理状态
 */
export function toggleProxyConfig() {
  return http.request({
    url: '/trading/proxyConfig/toggle',
    method: 'post',
  });
}
