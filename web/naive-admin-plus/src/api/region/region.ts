import { Alova } from '@/utils/http/alova/index';

//获取
export function allProvinces(params?) {
  return Alova.Get('/area/getParent', {
    params,
  });
}

export function regionParent(params) {
  return Alova.Get('/area/findByParentId', {
    params,
  });
}
