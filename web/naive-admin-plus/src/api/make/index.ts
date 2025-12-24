import { Alova } from '@/utils/http/alova/index';

//列表
export function makeList(params) {
  return Alova.Get('https://api.naiveadmin.com/api/web/make/list', {
    params,
  });
}

//新增
export function makeAdd(data) {
  return Alova.Post('https://api.naiveadmin.com/api/web/make/add', data);
}

//编辑
export function makeEdit(data) {
  return Alova.Post('https://api.naiveadmin.com/api/web/make/edit', data);
}

//删除
export function makeDelete(data) {
  return Alova.Post('https://api.naiveadmin.com/api/web/make/delete', data);
}
