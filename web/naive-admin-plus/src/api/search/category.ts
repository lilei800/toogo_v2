import { Alova } from '@/utils/http/alova/index';

export interface CategoryItem {
  id: number;
  name: string;
}

//获取分类
export function categoryList(params?) {
  return Alova.Get<{ list: CategoryItem[] }>('/category/list', {
    params,
  });
}
