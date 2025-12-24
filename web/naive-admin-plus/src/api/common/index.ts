import { Alova } from '@/utils/http/alova/index';

// 图片上传
export function upload(data) {
  return Alova.Post('/upload', data);
}

// 状态列表
export function stateList(params?) {
  return Alova.Get('/state_list', {
    params,
  });
}
