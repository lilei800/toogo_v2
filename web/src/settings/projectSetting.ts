import { IProjectSetting } from '/#/config';

const setting: IProjectSetting = {
  // 主题类型 pure 纯色 skin 皮肤
  themeType: 'pure',
  // 主题皮肤
  themeSkin: 'blue-sky',
  // 顶部透明
  headerLucency: false,
  // 侧边栏透明
  sidebarLucency: false,
  // 容器透明
  containerLucency: false,
  // 导航模式 vertical 左侧菜单模式 horizontal 顶部菜单模式 vertical-mix 左侧混合菜单 vertical-sub 左侧分栏
  // 如果修改默认值为其他模式 需要把下面 isMixMenu 设置为 false
  navMode: 'vertical',
  // 导航风格 dark 暗色侧边栏 light 白色侧边栏 header-dark 暗色顶栏
  navTheme: 'light',
  //触发移动端侧边栏的宽度
  mobileWidth: 800,
  // 左侧主导航宽度
  navWidth: 200,
  // 左侧主导航收起宽度
  navMinWidth: 64,
  // 导航触发器 all 同时显示底部和右侧 right 显示右侧 footer 显示底部 none 隐藏触发器
  navTrigger: 'all',
  // 分栏主导航宽度
  partionNavWidth: 90,
  // 分栏子导航宽度
  partionSubNavWidth: 180,
  // 分栏子导航收起宽度
  partionSubNavMinWidth: 64,
  // 隐藏分栏
  hidePartionSubNav: false,
  // 左侧主子导航宽度
  subNavWidth: 124,
  // 左侧混合菜单
  isMixMenu: false,
  // 折叠左侧导航
  collapsedNav: false,
  // 显示重载按钮
  isReload: true,
  // 顶部高度
  headerHeight: 57,
  // 固定顶部
  fixedHeader: true,
  // 显示多页签
  isMultiTabs: true,
  // 页签显示菜单图标
  isMultiTabsIcon: true,
  // 显示面包屑
  isCrumbs: true,
  // 显示面包屑图标
  isCrumbsIcon: true,
  // 菜单权限模式 FIXED 前端固定路由  BACK 动态获取
  permissionMode: 'BACK',
  // 是否开启路由动画
  isPageAnimate: true,
  // 路由动画类型 默认消退
  pageAnimateType: 'fade',
  // 内容区域宽度模式 full 流式 fixed 固定宽度
  contentType: 'full',
  // 语言选择器
  isI18n: true,
  // 圆角
  borderRadius: 12,
  // 水印
  isWatermark: false,
};
export default setting;
