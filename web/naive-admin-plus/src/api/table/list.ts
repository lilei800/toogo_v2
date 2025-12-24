import { Alova } from '@/utils/http/alova/index';

export interface TableItem {
  id: number;
  name: string;
  avatar: string;
  sex: string;
  status: string;
  createDate: string;
  email: string;
  city: string;
}

export const sexMap = {
  male: '男',
  female: '女',
  unknown: '未知',
};

export const statusMap = {
  close: '已取消',
  refuse: '已拒绝',
  pass: '已通过',
};

// 获取table
export function getTableList(params: any) {
  return Alova.Get('/table/list', { params });
}

//获取table select
export function getTableSelectList(params: any) {
  return Alova.Get('/table/select', { params });
}
