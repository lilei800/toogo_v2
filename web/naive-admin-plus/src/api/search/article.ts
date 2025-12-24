import { Alova } from '@/utils/http/alova/index';

export interface ArticleItem {
  id: string;
  title: string;
  tags: string[];
  summary: string;
  avatar: string;
  cover: string;
  author: string;
  collection: number;
  like: number;
  comment: number;
  date: string;
}

//获取文章
export function articleList(params?) {
  return Alova.Get<{ list: ArticleItem[] }>('/article/list', {
    params,
  });
}
