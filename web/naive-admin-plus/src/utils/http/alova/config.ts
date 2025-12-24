export const defaultMockConfig = {
  // mock接口响应延迟，单位毫秒
  delay: 1000,
};

// 请求配置
export const defaultConfig: AlovaService.Config = {
  // 自定义配置 给 alova 封装的实例使用
  codeField: 'code',
  messageField: 'message',
  dataField: 'result',
  successCode: 200,
  tokenField: 'Authorization', // Token 字段名
  authenticationScheme: '', // Token 认证方案
  joinTime: true, // 是否需要加入时间戳
  baseURL: '',
  // alova 配置
  alovaConfig: {
    cacheLogger: process.env.NODE_ENV === 'development', // 在开发环境开启缓存命中日志
    cacheFor: null, // 关闭全局请求缓存
    timeout: 15 * 1000, // 超时时间
  },
};

// 错误信息 可根据项目情况 按需修改或使用
export const errorConfig = {
  // 错误时是否显示 message
  isShowMessage: true,
  default: '请求失败',
  500: '服务器错误',
  701: '请求失败',
  720: '请求超时',
  730: '参数错误',
  900: '暂无权限',
  901: '缺少凭证',
  902: '凭证校验失败',
  903: '账号异常',
  912: '需要登录',
  415: '参数格式错误',
};

export const noErrorTipCodes = [912]; // 不需要错误提示的 code
