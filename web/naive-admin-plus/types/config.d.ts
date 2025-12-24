export type ThemeType = 'pure' | 'skin';

export type ThemeSkin =
  | 'blue-sky'
  | 'blue-christmas'
  | 'blue-lattice'
  | 'pink-romantic'
  | 'green-mountain'
  | 'paint-splashing'
  | 'orange-bubble'
  | 'star-jellyfish'
  | 'star-neon';
export interface IProjectSetting {
  // 主题类型 pure 纯色 skin 皮肤
  themeType: ThemeType;
  // 主题皮肤
  themeSkin: ThemeSkin;
  // 顶部透明
  headerLucency: boolean;
  // 侧边栏透明
  sidebarLucency: boolean;
  // 容器透明
  containerLucency: boolean;
  // 导航模式 vertical 左侧菜单模式 horizontal 顶部菜单模式 vertical-mix 左侧混合菜单
  navMode: string;
  // 导航风格 dark 暗色侧边栏 light 白色侧边栏 header-dark 暗色顶栏
  navTheme: string;
  // 触发移动端侧边栏的宽度
  mobileWidth: number;
  // 左侧主导航宽度
  navWidth: number;
  // 左侧主导航收起宽度
  navMinWidth: number;
  // 分栏主导航宽度
  partionNavWidth: number;
  // 分栏子导航宽度
  partionSubNavWidth: number;
  // 分栏子导航收起宽度
  partionSubNavMinWidth: number;
  // 隐藏分栏
  hidePartionSubNav: boolean;
  // 导航触发器 all 同时显示底部和右侧 right 显示右侧 footer 显示底部 none 隐藏触发器
  navTrigger: string;
  // 折叠左侧导航
  collapsedNav: boolean;
  // 左侧主子导航宽度
  subNavWidth: number;
  // 左侧混合菜单
  isMixMenu: boolean;
  // 显示重载按钮
  isReload: boolean;
  // 是否已移动端模式处理页面
  isMobile: boolean;
  // 顶部高度
  headerHeight: number;
  // 固定顶部
  fixedHeader: boolean;
  // 显示多页签
  isMultiTabs: boolean;
  // 显示面包屑
  isCrumbs: boolean;
  // 显示面包屑图标
  isCrumbsIcon: boolean;
  // 页签显示菜单图标
  isMultiTabsIcon: boolean;
  // 菜单权限模式 FIXED 前端固定路由  BACK 动态获取
  permissionMode: string;
  // 是否开启路由动画
  isPageAnimate: boolean;
  // 路由动画类型 默认消退
  pageAnimateType: string;
  //内容区域宽度模式 full 流式 fixed 固定宽度
  contentType: string;
  // 语言选择器
  isI18n: boolean;
  // 圆角
  borderRadius: number;
  // 水印
  isWatermark: boolean;
}
export interface GlobEnvConfig {
  // 标题
  VITE_GLOB_APP_TITLE: string;
  // 接口地址
  VITE_GLOB_API_URL: string;
  // 接口前缀
  VITE_GLOB_API_URL_PREFIX?: string;
  // 项目简称
  VITE_GLOB_APP_SHORT_NAME: string;
  // 上传地址
  VITE_GLOB_UPLOAD_URL?: string;
  // 文件前缀地址
  VITE_GLOB_FILE_URL?: string;
  // 生产环境开启mock
  VITE_USE_MOCK: string;
  // 开启 mock 请求打印日志
  VITE_LOGGER_MOCK: string;
}

export interface GlobConfig {
  // 标题
  title: string;
  // 接口地址
  apiUrl: string;
  // 文件上传地址
  uploadUrl?: string;
  // api 接口前缀
  urlPrefix?: string;
  // 项目简称
  shortName: string;
  // 文件访问地址
  fileUrl: string | undefined;
}

export interface LocalConfig {
  // 生产环境开启 mock
  useMock: boolean;
  // 打印 mock 请求信息
  loggerMock: boolean;
}

// 多语言
export type LocaleType = 'zh_CN' | 'en';

export interface LocaleSetting {
  showPicker: boolean;
  // 当前语言
  locale: LocaleType;
  // 默认语言
  fallback: LocaleType;
  // 可用的 语言
  availableLocales: LocaleType[];
}
