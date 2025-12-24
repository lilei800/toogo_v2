/**
 * 交易所API配置接口
 */
import { http } from '@/utils/http/axios';

/**
 * 获取API配置列表
 */
export function getApiConfigList(params?: any) {
  return http.request({
    url: '/trading/api-config/list',
    method: 'get',
    params,
  });
}

/**
 * 创建API配置
 */
export function createApiConfig(data: any) {
  return http.request({
    url: '/trading/api-config/create',
    method: 'post',
    data,
  });
}

/**
 * 更新API配置
 */
export function updateApiConfig(data: any) {
  return http.request({
    url: '/trading/api-config/update',
    method: 'post',
    data,
  });
}

/**
 * 删除API配置
 */
export function deleteApiConfig(params: { id: number }) {
  return http.request({
    url: '/trading/api-config/delete',
    method: 'post',
    params,
  });
}

/**
 * 获取API配置详情
 */
export function getApiConfigDetail(params: { id: number }) {
  return http.request({
    url: '/trading/api-config/view',
    method: 'get',
    params,
  });
}

/**
 * 测试API连接
 */
export function testApiConfig(params: { id: number }) {
  return http.request({
    url: '/trading/api-config/test',
    method: 'post',
    params,
  });
}

/**
 * 设为默认API配置
 */
export function setDefaultApiConfig(params: { id: number }) {
  return http.request({
    url: '/trading/api-config/set-default',
    method: 'post',
    params,
  });
}

/**
 * 获取支持的平台列表
 */
export function getPlatforms() {
  return http.request({
    url: '/trading/api-config/platforms',
    method: 'get',
  });
}

