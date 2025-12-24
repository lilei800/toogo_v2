<script lang="ts" setup>
  import { watch } from 'vue';
  import { Sunny, SunnyOutline } from '@vicons/ionicons5';
  import { useDesignSettingStore } from '@/store/modules/designSetting';
  import { cloneDeep } from 'lodash-es';
  import { useProjectSettingStore } from '@/store/modules/projectSetting';
  import useViewTransition from '@/hooks/web/useViewTransition';

  const designStore = useDesignSettingStore();
  const settingStore = useProjectSettingStore();

  watch(
    () => designStore.darkTheme,
    (to) => {
      settingStore.navTheme = to ? 'header-dark' : 'light';
    },
  );

  async function toggleDarkness({ clientX: x, clientY: y }: MouseEvent) {
    const isDark = cloneDeep(designStore.darkTheme);
    const htmlRoot = document.getElementById('htmlRoot');

    const { startViewTransition } = useViewTransition(() => {
      if (!htmlRoot) return;
      if (!isDark) htmlRoot.setAttribute('data-theme', 'dark');
      else htmlRoot.setAttribute('data-theme', 'light');
      designStore.setDarkTheme(!isDark);
    });

    startViewTransition()?.ready.then(() => {
      const endRadius = Math.hypot(Math.max(x, innerWidth - x), Math.max(y, innerHeight - y));
      const clipPath = [`circle(0px at ${x}px ${y}px)`, `circle(${endRadius}px at ${x}px ${y}px)`];
      document.documentElement.animate(
        { clipPath: isDark ? clipPath.reverse() : clipPath },
        {
          duration: 300,
          easing: 'ease-out',
          pseudoElement: isDark ? '::view-transition-old(root)' : '::view-transition-new(root)',
        },
      );
    });
  }
</script>

<template>
  <n-tooltip trigger="hover">
    <template #trigger>
      <n-button strong circle secondary type="tertiary" @click="toggleDarkness">
        <n-icon :size="18" v-if="designStore.getDarkTheme" :component="Sunny" />
        <n-icon :size="18" v-else :component="SunnyOutline" />
      </n-button>
    </template>
    {{ designStore.getDarkTheme ? '切换亮色模式' : '切换暗色模式' }}
  </n-tooltip>
</template>
