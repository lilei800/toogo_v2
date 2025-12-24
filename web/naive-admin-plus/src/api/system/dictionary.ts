import { Alova } from '@/utils/http/alova/index';

/**
 * @description: 获取字典列表
 */
export function getDictionary(params?) {
  return Alova.Get('/dictionary/list', {
    params,
  });
}

/**
 * @description: 获取字典详情
 */
export function getDictionaryInfo(params) {
  return Alova.Get('/dictionary/info', {
    params,
  });
}
