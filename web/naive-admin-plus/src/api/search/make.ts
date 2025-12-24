import { Alova } from '@/utils/http/alova/index';

export interface MakeItem {
  id: number;
  doctor: string;
  avatar: string;
  subject: string;
  date: string;
}

//获取预约列表
export function makeList(params?) {
  return Alova.Get<{ list: MakeItem[] }>('/make/list', {
    params,
  });
}
