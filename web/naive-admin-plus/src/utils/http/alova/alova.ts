import { createAlova } from 'alova';
import { defaultConfig } from './config';
import VueHook from 'alova/vue';
import { mockAdapter } from './adapter';
import { ResultEnum } from '@/enums/httpEnum';
import { checkStatus, handlerJoinTime, handlerToken, handlerUrlPrefix } from './helper';

export function createInstance(config: AlovaService.Config) {
  return createAlova({
    ...defaultConfig.alovaConfig,
    baseURL: config.baseURL,
    statesHook: VueHook,
    requestAdapter: mockAdapter,
    beforeRequest: (method) => {
      // 处理 Token
      handlerToken(config, method);
      // 处理请求前缀
      handlerUrlPrefix(config, method);
      // 处理时间戳参数
      handlerJoinTime(method, defaultConfig.joinTime);
    },
    responded: {
      onSuccess: async (response, method) => {
        // 是否返回原生响应头 比如：需要获取响应头时使用该属性
        if (method.meta?.isReturnNativeResponse) {
          return response;
        }
        // 不进行任何处理，直接返回
        // 用于需要直接获取 code、result、 message 这些信息时开启
        if (method.meta?.isTransformResponse === false) {
          return response.data;
        }
        // 请求成功 返回数据结构，请根据自身情况修改数配置
        const jsonResult = response.data;
        const code = jsonResult[config.codeField];
        const message = jsonResult[config.messageField] || 'Unknown error';
        const dataSource = jsonResult[config.dataField];
        if (code === ResultEnum.SUCCESS) {
          return dataSource;
        }
        // 请求失败 比如 status 200 但业务 code 返回错误
        return checkStatus(code, message);
      },
      onError(error) {
        // 请求失败
        try {
          const response = error.response;
          let code = error.code;
          let message = error.message;

          if (response) {
            const jsonResult = response.data || {};
            code = jsonResult[config.codeField] || error.code;
            message = jsonResult[config.messageField] || error.message;
          }
          checkStatus(code, message);
          // 打印错误信息
          console.error('Error occurred:', { code, message });
          return Promise.reject(message || 'Unknown error');
        } catch (err) {
          console.error('Error handling failed:', err);
          return Promise.reject('Error handling failed');
        }
      },
    },
  });
}
