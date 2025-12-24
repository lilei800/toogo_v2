import { defineStore } from 'pinia';
import { createStorage } from '@/utils/Storage';
import { store } from '@/store';
import { ACCESS_TOKEN, CURRENT_USER, IS_LOCKSCREEN } from '@/enums/common';
import { ResultEnum } from '@/enums/httpEnum';

const Storage = createStorage({ storage: localStorage });
import { getUserInfo, login, userLogout, getLoginConfig, getConfig } from '@/api/system/user';
import { storage } from '@/utils/Storage';
import { isWechatBrowser } from '@/utils/is';

export interface IUserState {
  token: string;
  username: string;
  welcome: string;
  avatar: string;
  permissions: any[];
  info: any;
  loginConfig: any;
  config: any;
}

export const useUserStore = defineStore({
  id: 'app-user',
  state: (): IUserState => ({
    token: Storage.get(ACCESS_TOKEN, ''),
    username: '',
    welcome: '',
    avatar: '',
    permissions: [],
    info: Storage.get(CURRENT_USER, {}),
    loginConfig: {},
    config: {},
  }),
  getters: {
    getToken(): string {
      return this.token;
    },
    getAvatar(): string {
      return this.avatar;
    },
    getNickname(): string {
      return this.username;
    },
    getPermissions(): [any][] {
      return this.permissions;
    },
    getUserInfo(): object {
      return this.info;
    },
    getLoginConfig(): any {
      return this.loginConfig;
    },
    getConfig(): any {
      return this.config;
    },
  },
  actions: {
    setToken(token: string) {
      this.token = token;
    },
    setAvatar(avatar: string) {
      this.avatar = avatar;
    },
    setPermissions(permissions) {
      this.permissions = permissions;
    },
    setUserInfo(info) {
      this.info = info;
    },
    setConfig(config) {
        this.config = config;
    },
    // 登录
    async login(userInfo) {
      try {
        const response = await login(userInfo);
        const { data: result, code } = response;
        if (code === ResultEnum.SUCCESS) {
          const ex = 7 * 24 * 60 * 60 * 1000;
          storage.set(ACCESS_TOKEN, result.token, ex);
          storage.set(CURRENT_USER, result, ex);
          storage.set(IS_LOCKSCREEN, false);
          this.setToken(result.token);
          this.setUserInfo(result);
        }
        return Promise.resolve(response);
      } catch (e) {
        return Promise.reject(e);
      }
    },

    // 获取用户信息
    GetInfo() {
      return new Promise((resolve, reject) => {
        getUserInfo()
          .then((res) => {
            try {
              if (res) {
                if (res.permissions && res.permissions.length) {
                  this.setPermissions(res.permissions);
                }
                this.setUserInfo(res);
                this.setAvatar(res.avatar);
                resolve(res);
              } else {
                reject('Get User Info Failed');
              }
            } catch (error) {}
          })
          .catch((error) => {
            reject(error);
          });
      });
    },

    // 登出
    async logout() {
      return new Promise((resolve) => {
        // 无论API调用成功与否，都执行退出操作
        userLogout()
          .then(() => {
            // API调用成功，正常退出
            this.setPermissions([]);
            this.setUserInfo({} as IUserState);
            storage.remove(ACCESS_TOKEN);
            storage.remove(CURRENT_USER);
            return resolve(true);
          })
          .catch((error) => {
            // API调用失败（如token已过期），仍然执行退出操作
            console.warn('退出登录API调用失败，但继续执行退出操作:', error);
            this.setPermissions([]);
            this.setUserInfo({} as IUserState);
            storage.remove(ACCESS_TOKEN);
            storage.remove(CURRENT_USER);
            return resolve(true);
          });
      });
    },

    // 加载登录配置
    async LoadLoginConfig() {
      try {
        const res = await getLoginConfig();
        this.loginConfig = res;
        return Promise.resolve(res);
      } catch (error) {
        return Promise.reject(error);
      }
    },
    // 获取基础配置
    GetConfig() {
      return new Promise((resolve, reject) => {
        getConfig()
          .then((res) => {
            this.setConfig(res);
            resolve(res);
          })
          .catch((error) => {
            reject(error);
          });
      });
    },
    // 是否允许获取微信openid
    allowWxOpenId(): boolean {
      if (!isWechatBrowser()) {
        return false;
      }
      if (this.loginConfig && this.loginConfig.loginAutoOpenId !== 1) {
        return false;
      }
      return !this.info || !this.info.openId;
    },
  },
});

// Need to be used outside the setup
export function useUserStoreWidthOut() {
  return useUserStore(store);
}
