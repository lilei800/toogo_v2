import { http } from '@/utils/http/axios';
import { ApiEnum } from '@/enums/apiEnum';
import { aesEcb } from '@/utils/encrypt';

export interface BasicResponseModel<T = any> {
  code: number;
  message: string;
  data: T;
}

export interface BasicPageParams {
  pageNumber: number;
  pageSize: number;
  total: number;
}

export function getConfig() {
  return http.request({
    url: ApiEnum.SiteConfig,
    method: 'get',
    headers: { hostname: location.hostname },
  });
}

/**
 * @description: 获取用户信息
 */
export function getUserInfo() {
  return http.request({
    url: ApiEnum.MemberInfo,
    method: 'get',
  });
}

export function updateMemberProfile(params) {
  return http.request({
    url: '/member/updateProfile',
    method: 'post',
    params,
  });
}

export function updateMemberPwd(params) {
  return http.request({
    url: '/member/updatePwd',
    method: 'post',
    params,
  });
}

export function updateMemberMobile(params) {
  return http.request({
    url: '/member/updateMobile',
    method: 'post',
    params,
  });
}

export function updateMemberEmail(params) {
  return http.request({
    url: '/member/updateEmail',
    method: 'post',
    params,
  });
}

export function SendBindEmail() {
  return http.request({
    url: '/ems/sendBind',
    method: 'post',
  });
}

// 注册：发送邮箱验证码
export function SendRegisterEmail(params: { email: string }) {
  return http.request({
    url: '/site/sendRegisterEmail',
    method: 'post',
    params,
  });
}

// 忘记密码：发送重置密码邮件
export function SendResetPwdEmail(params: { email: string }) {
  return http.request({
    url: '/site/sendResetPwdEmail',
    method: 'post',
    params,
  });
}

// 重置密码：通过token设置新密码
export function PasswordReset(params: { token: string; password: string }) {
  return http.request<BasicResponseModel>(
    {
      url: '/site/passwordReset',
      method: 'post',
      params,
    },
    {
      isTransformResponse: false,
    }
  );
}

export function SendBindSms() {
  return http.request({
    url: '/sms/sendBind',
    method: 'post',
  });
}

export function SendSms(params) {
  return http.request({
    url: '/sms/send',
    method: 'post',
    params,
  });
}

export function updateMemberCash(params) {
  return http.request({
    url: '/member/updateCash',
    method: 'post',
    params,
  });
}

/**
 * @description: 用户登录配置
 */
export function getLoginConfig() {
  return http.request<BasicResponseModel>({
    url: ApiEnum.SiteLoginConfig,
    method: 'get',
  });
}

/**
 * @description: 用户注册
 */
export function register(params) {
  return http.request<BasicResponseModel>(
    {
      url: ApiEnum.SiteRegister,
      method: 'POST',
      params,
    },
    {
      isTransformResponse: false,
    }
  );
}

/**
 * @description: 用户登录
 */
export function login(params) {
  // 注意：密码已经在 HotgoLoginForm.vue 中加密过了，这里不需要再加密
  return http.request<BasicResponseModel>(
    {
      url: ApiEnum.SiteAccountLogin,
      method: 'POST',
      params,
    },
    {
      isTransformResponse: false,
    }
  );
}

/**
 * @description: 手机号登录
 */
export function mobileLogin(params) {
  return http.request<BasicResponseModel>(
    {
      url: ApiEnum.SiteMobileLogin,
      method: 'POST',
      params,
    },
    {
      isTransformResponse: false,
    }
  );
}

/**
 * @description: 用户注销
 */
export function logout() {
  return http.request<BasicResponseModel>(
    {
      url: ApiEnum.SiteLogout,
      method: 'POST',
    },
    {
      isTransformResponse: false,
    }
  );
}

/**
 * @description: 用户注销（别名）
 */
export const userLogout = logout;

/**
 * @description: 用户修改密码
 */
export function changePassword(params, uid) {
  return http.request(
    {
      url: `/user/u${uid}/changepw`,
      method: 'POST',
      params,
    },
    {
      isTransformResponse: false,
    }
  );
}

/**
 * @description: 获取用户列表（兼容原版）
 */
export function getUserList(params) {
  return http.request({
    url: '/member/list',
    method: 'get',
    params,
  });
}