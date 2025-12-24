import { useMessage } from '@/hooks/web/useMessage';
import { noErrorTipCodes } from './config';
import { isUrl } from '@/utils';
import { isObject, isString, merge } from 'lodash-es';
import { setObjToUrlParams } from '@/utils/urlUtils';
import { useDialog } from '@/hooks/web/useDialog';
import { storage } from '@/utils/Storage';
import { PageEnum } from '@/enums/pageEnum';
import { ACCESS_TOKEN } from '@/enums/common';

const LoginPath = PageEnum.BASE_LOGIN;
const DATE_TIME_FORMAT = 'YYYY-MM-DD HH:mm';

export function joinTimestamp<T extends boolean>(
  join: boolean,
  restful: T,
): T extends true ? string : object;

export function joinTimestamp(join: boolean, restful = false): string | object {
  if (!join) {
    return restful ? '' : {};
  }
  const now = new Date().getTime();
  if (restful) {
    return `?_t=${now}`;
  }
  return { _t: now };
}

// 合并配置
export function mergeConfig(config, neConfig) {
  return merge(config, neConfig);
}

// 处理 Token 凭证
export function handlerToken(config, method) {
  if (method.config.meta?.withToken !== false) {
    const token = storage.get(ACCESS_TOKEN);
    if (token) {
      const Bearer = config.authenticationScheme;
      method.config.headers[config.tokenField] = Bearer ? `${Bearer} ${token}` : token;
    }
  }
}

// 处理url前缀
export function handlerUrlPrefix(config, method) {
  const isUrlStr = isUrl(method.url as string);
  if (!isUrlStr && config.urlPrefix) {
    method.url = `${config.urlPrefix}${method.url}`;
  }
}

// 处理请求参数
export function formatRequestDate(params: Recordable) {
  if (Object.prototype.toString.call(params) !== '[object Object]') {
    return;
  }

  for (const key in params) {
    if (params[key] && params[key]._isAMomentObject) {
      params[key] = params[key].format(DATE_TIME_FORMAT);
    }
    if (isString(key)) {
      const value = params[key];
      if (value) {
        try {
          params[key] = isString(value) ? value.trim() : value;
        } catch (error) {
          throw new Error(error as any);
        }
      }
    }
    if (isObject(params[key])) {
      formatRequestDate(params[key]);
    }
  }
}

// 处理时间戳参数
export function handlerJoinTime(method, joinTime) {
  if (joinTime === true) {
    const config = method.config;
    const params = config.params || {};
    const data = method.data || {};
    const formatDate = config.meta?.formatDate || false;
    const joinParamsToUrl = config.meta?.joinParamsToUrl || false;
    if (method.type === 'GET') {
      if (!isString(params)) {
        method.config.params = Object.assign(params || {}, joinTimestamp(joinTime, false));
      } else {
        method.url = method.url + params + `${joinTimestamp(joinTime, true)}`;
      }
    } else {
      if (!isString(params)) {
        formatDate && formatRequestDate(params);
        if (
          Reflect.has(method, 'data') &&
          method.data &&
          (Object.keys(method.data).length > 0 || method.data instanceof FormData)
        ) {
          method.data = data;
          config.params = params;
        } else {
          method.data = params;
          config.params = undefined;
        }
        if (joinParamsToUrl) {
          method.url = setObjToUrlParams(
            method.url as string,
            Object.assign({}, config.params, method.data),
          );
        }
      } else {
        // 兼容restful风格
        method.url = config.url + params;
        config.params = undefined;
      }
    }
  }
}

// open 一个错误提示
export function openErrorMessage(message: string) {
  const $message = useMessage();
  $message.error(message);
}

export function openDialog(config: {
  title?: string;
  content: string;
  positiveText?: string;
  fun?: () => void;
}) {
  const { title = '提示', positiveText = '确定' } = config;
  const $dialog = useDialog();
  $dialog?.warning({
    title: title,
    content: config.content,
    positiveText,
    closable: false,
    maskClosable: false,
    onPositiveClick: () => {
      config.fun?.();
    },
  });
}

// 错误处理
export function checkStatus(code: number, message: string) {
  if (code === 902) {
    openDialog({
      content: '登录身份已失效，请重新登录!',
      fun: () => {
        storage.clear();
        window.location.href = LoginPath;
      },
    });
    return;
  }
  if (!noErrorTipCodes.includes(code)) {
    openErrorMessage(message);
  }
}
