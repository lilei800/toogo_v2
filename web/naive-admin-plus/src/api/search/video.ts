import { Alova } from '@/utils/http/alova/index';

export interface VideoItem {
  id: string;
  title: string;
  summary: string;
  avatargroup: any[];
  cover: string;
  viewingtimes: number;
  date: string;
}

//获取视频
export function videoList(params?) {
  return Alova.Get<{ list: VideoItem[] }>('/video/list', {
    params,
  });
}
