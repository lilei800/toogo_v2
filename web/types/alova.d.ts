// Alova 请求封装类型定义
declare namespace AlovaService {
  // 配置
  interface Config {
    urlPrefix?: string; // 接口地址前缀
    codeField: string; // 状态码字段名
    dataField: string; // 接口返回数据字段名
    messageField: string; // 错误信息字段名
    successCode: number; // 成功状态码
    tokenField: string; // Token 字段名
    authenticationScheme?: string; // 认证方式
    guestField?: string; // 游客凭证字段名
    joinTime?: boolean; // 是否需要加入时间戳
    baseURL: string; // 接口请求地址
    alovaConfig: {
      cacheLogger?: boolean; // 在开发环境开启缓存命中日志
      cacheFor?: null; // 关闭全局请求缓存
      timeout?: timeout; // 超时时间
    };
  }
}
