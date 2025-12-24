<template>
  <pro-layout v-bind="layoutConfig">
    <template #header-right>
      <RightItem />
    </template>
    <!-- 顶部菜单 -->
    <template #header-menu>
      <TopMenu v-if="['vertical-mix', 'horizontal'].includes(getNavMode)" responsive />
    </template>

    <MainView />

    <!-- 系统配置入口 -->
    <n-tooltip trigger="hover">
      <template #trigger>
        <n-el tag="div" class="system-setting">
          <div class="shadow system-setting-btn" @click="openSetting">
            <n-icon :size="20" color="var(--primary-color)">
              <ToggleSharp />
            </n-icon>

            <n-icon :size="20" color="var(--primary-color)" class="-mt-1 opacity-50">
              <ToggleSharp />
            </n-icon>
          </div>
        </n-el>
      </template>
      系统配置
    </n-tooltip>

    <!--项目配置-->
    <ProjectSetting ref="drawerSetting" />

    <!--全屏水印-->
    <n-watermark
      v-if="settingStore.isWatermark"
      content="NaiveAdmin"
      cross
      fullscreen
      :font-size="16"
      :line-height="16"
      :width="384"
      :height="384"
      :x-offset="12"
      :y-offset="60"
      :rotate="-15"
    />
  </pro-layout>
</template>

<script lang="ts" setup>
  import { ref, reactive, unref, computed, watch, provide, onMounted } from 'vue';
  import { MainView } from './components/Main';
  import { RightItem } from './components/Header';
  import { proLayout } from './components/ProLayout/index';
  import { TopMenu } from './components/Menu';
  import { ToggleSharp } from '@vicons/ionicons5';
  import { useProjectSetting } from '@/hooks/setting/useProjectSetting';
  import ProjectSetting from './components/ProjectSetting/index.vue';

  import { useProjectSettingStore } from '@/store/modules/projectSetting';

  import { useFullscreen } from '@vueuse/core';

  const { getNavMode, getNavTheme, getNavWidth, getSubNavWidth, getNavMinWidth, getCollapsedNav } =
    useProjectSetting();

  const settingStore = useProjectSettingStore();

  const drawerSetting = ref();
  const collapsed = ref<boolean>(getCollapsedNav.value);
  const adminBodyRef = ref<HTMLElement | null>(null);
  const layoutConfig = reactive({
    title: 'Toogo.Ai',
  });

  const { isFullscreen } = useFullscreen(adminBodyRef);

  provide('isPageFullScreen', isFullscreen);
  provide('collapsed', collapsed);
  provide('openSetting', openSetting);
  provide('navTheme', getNavTheme);
  provide('navWidth', getNavWidth);
  provide('navMinWidth', getNavMinWidth);

  watch(
    () => collapsed.value,
    (newValue) => {
      settingStore.setCollapsedNav(newValue);
    },
    { immediate: true },
  );

  watch(
    () => settingStore.collapsedNav,
    (newValue) => {
      collapsed.value = newValue;
    },
    { immediate: true },
  );

  //打开设置
  function openSetting() {
    const { openDrawer } = drawerSetting.value;
    openDrawer();
  }

  const leftWidth = computed(() => {
    return unref(getNavWidth) + unref(getSubNavWidth) + 'px';
  });

  provide('leftWidth', leftWidth);

  //判断是否触发移动端模式
  const watchMobileMode = () => {
    if (document.body.clientWidth <= settingStore.mobileWidth) {
      settingStore.setIsMobile(true);
    } else {
      settingStore.setIsMobile(false);
    }
  };

  const watchViewportWidth = () => {
    const Width = document.body.clientWidth;
    if (Width <= 950) {
      settingStore.setCollapsedNav(true);
    } else settingStore.setCollapsedNav(false);

    watchMobileMode();
  };

  onMounted(() => {
    watchMobileMode();
    window.addEventListener('resize', watchViewportWidth);
  });
</script>

<style lang="less" scoped>
  .system-setting {
    position: fixed;
    z-index: 10;
    right: 26px;
    bottom: 86px;

    &-btn {
      background: var(--card-color);
      display: inline-flex;
      align-items: center;
      justify-content: center;
      flex-direction: column;
      cursor: pointer;
      border-radius: 50%;
      width: 52px;
      height: 52px;
      color: var(--card-color);
      box-shadow: 0 1px 2px -2px rgba(0, 0, 0, 0.08), 0 3px 6px 0 rgba(0, 0, 0, 0.06),
        0 5px 12px 4px rgba(0, 0, 0, 0.04);
    }
  }
</style>
