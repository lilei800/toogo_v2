import { Alova } from '@/utils/http/alova/index';

//获取分类
export function getClassifyList(params?) {
  return Alova.Get('/classifyList', {
    params,
  });
}
