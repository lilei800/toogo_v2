export enum PageEnum {
  // 登录
  BASE_LOGIN = '/login',
  BASE_LOGIN_NAME = 'Login',
  //重定向
  REDIRECT = '/redirect',
  REDIRECT_NAME = 'Redirect',
  // 首页
  BASE_HOME = '/toogo',
  //首页跳转默认路由
  // 注意：组件文件名 index.vue 并不意味着 URL 必须带 /index
  // 后端动态菜单/静态路由默认是 /toogo/dashboard，因此这里统一为 /toogo/dashboard，避免登录后跳到不存在的路由
  BASE_HOME_REDIRECT = '/toogo/dashboard',
  // 错误
  ERROR_PAGE_NAME = 'ErrorPage',
}
