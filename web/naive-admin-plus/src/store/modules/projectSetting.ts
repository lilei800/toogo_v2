import { defineStore } from 'pinia';
import { store } from '@/store';
import projectSetting from '@/settings/projectSetting';
import type { IProjectSetting } from '/#/config';

const {
  themeType,
  themeSkin,
  headerLucency,
  sidebarLucency,
  containerLucency,
  navMode,
  navTheme,
  navWidth,
  navMinWidth,
  partionNavWidth,
  partionSubNavMinWidth,
  partionSubNavWidth,
  hidePartionSubNav,
  navTrigger,
  subNavWidth,
  isMixMenu,
  collapsedNav,
  isReload,
  headerHeight,
  fixedHeader,
  isMultiTabs,
  isMultiTabsIcon,
  isCrumbs,
  isCrumbsIcon,
  permissionMode,
  isPageAnimate,
  pageAnimateType,
  contentType,
  isI18n,
  borderRadius,
  isWatermark,
  mobileWidth,
  isMobile,
} = projectSetting;

export const useProjectSettingStore = defineStore('app-project-setting', {
  state: (): IProjectSetting => ({
    themeType,
    themeSkin,
    headerLucency,
    sidebarLucency,
    containerLucency,
    navMode,
    navTheme,
    navWidth,
    navMinWidth,
    partionNavWidth,
    partionSubNavMinWidth,
    partionSubNavWidth,
    hidePartionSubNav,
    navTrigger,
    subNavWidth,
    isMixMenu,
    collapsedNav,
    isReload,
    headerHeight,
    fixedHeader,
    isMultiTabs,
    isMultiTabsIcon,
    isCrumbs,
    isCrumbsIcon,
    permissionMode,
    isPageAnimate,
    pageAnimateType,
    contentType,
    isI18n,
    borderRadius,
    isWatermark,
    mobileWidth,
    isMobile,
  }),
  persist: {
    key: 'APP-PROJECT-SETTING',
  },
  getters: {
    getNavMode: (state) => state.navMode,
    getNavTheme: (state) => state.navTheme,
    getNavWidth: (state) => state.navWidth,
    getNavMinWidth: (state) => state.navMinWidth,
    getNavTrigger: (state) => state.navTrigger,
    getSubNavWidth: (state) => state.subNavWidth,
    getIsMixMenu: (state) => state.isMixMenu,
    getCollapsedNav: (state) => state.collapsedNav,
    getIsReload: (state) => state.isReload,
    getHeaderHeight: (state) => state.headerHeight,
    getFixedHeader: (state) => state.fixedHeader,
    getIsMultiTabs: (state) => state.isMultiTabs,
    getPermissionMode: (state) => state.permissionMode,
    getIsPageAnimate: (state) => state.isPageAnimate,
    getPageAnimateType: (state) => state.pageAnimateType,
    getContentType: (state) => state.contentType,
    getBorderRadius: (state) => state.borderRadius,
    getMobileWidth: (state) => state.mobileWidth,
    getIsMobile: (state) => state.isMobile,
  },
  actions: {
    setIsMobile(value: boolean) {
      this.isMobile = value;
    },
    setMobileWidth(value: number) {
      this.mobileWidth = value;
    },
    setHidePartionSubNav(value: boolean) {
      this.hidePartionSubNav = value;
    },
    setNavTheme(value: string) {
      this.navTheme = value;
    },
    setCollapsedNav(value: boolean) {
      this.collapsedNav = value;
    },
  },
});

// Need to be used outside the setup
export function useProjectSettingStoreWithOut() {
  return useProjectSettingStore(store);
}
