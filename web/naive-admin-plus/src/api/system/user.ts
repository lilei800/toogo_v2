import { Alova } from '@/utils/http/alova/index';
export interface BasicPageParams {
  pageNumber: number;
  pageSize: number;
  total: number;
}
export interface UserInfo {
  id: string;
  avatar: string;
  birthday: Date;
  email: string;
  username: string;
  account: string;
  sex: string;
  role: string;
  permissions: { label: string; value: string }[];
  status: number;
  createDate: Date;
}

/**
 * @description: 获取用户信息
 */
export function getUserInfo() {
  return Alova.Get<UserInfo>('/admin_info');
}

/**
 * @description: 用户登录
 */
export function login(data) {
  return Alova.Post<ResponseModel<{ id: string; token: string }>>('/login', data, {
    // 返回服务端完整Json结构 包含 code message result 字段
    meta: {
      isTransformResponse: false,
    },
  });
}

/**
 * @description: 用户修改密码
 */
export function changePassword(data, uid) {
  return Alova.Post(`/user/u${uid}/changepw`, data);
}

/**
 * @description: 用户登出
 */
export function userLogout(data?) {
  return Alova.Post('/logout', data, {
    meta: {
      authRole: 'logout',
    },
  });
}

/**
 * @description: 获取用户列表
 */
export function getUserList(params) {
  return Alova.Get('/user_list', { params });
}
