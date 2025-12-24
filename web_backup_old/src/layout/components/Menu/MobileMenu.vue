<template>
  <n-menu
    :options="menus"
    :inverted="getInverted"
    :mode="getMode"
    :collapsed="false"
    :collapsed-width="getNavMinWidth"
    :collapsed-icon-size="20"
    :indent="props.indent"
    :expanded-keys="getOpenKeys"
    :value="getSelectedKeys"
    @update:value="clickMenuItem"
    @update:expanded-keys="menuExpanded"
  />
</template>

<script lang="ts" setup>
  import { ref, onMounted, computed, watch, unref, inject } from 'vue';
  import { useRoute, useRouter } from 'vue-router';
  import { useAsyncRouteStore } from '@/store/modules/asyncRoute';
  import { generatorMenu } from '@/utils';
  import { useMenu } from './hooks/useMenu';

  const emits = defineEmits(['update:showMobileSider']);

  const props = defineProps({
    mode: {
      // 菜单模式
      type: String as PropType<'vertical' | 'horizontal'>,
      default: 'vertical',
    },
    hideSubMenu: {
      type: Boolean,
      default: false,
    },
    indent: {
      type: Number,
      default: 24,
    },
    showMobileSider: {
      type: Boolean,
      default: false,
    },
  });

  const getNavMinWidth = computed(() => {
    return unref(inject('navMinWidth') as number);
  });

  const { replaceI18nTitle, getMenuInverted } = useMenu();

  const getInverted = computed(() => {
    return getMenuInverted.value;
  });

  // 当前路由
  const currentRoute = useRoute();
  const router = useRouter();
  const asyncRouteStore = useAsyncRouteStore();
  const menus = ref<any[]>([]);
  const selectedKeys = ref<string>(currentRoute.name as string);

  // 获取当前打开的子菜单
  const matched = currentRoute.matched;

  const openKeys = ref(matched && matched.length ? matched.map((item) => item.name) : []);

  const getOpenKeys = computed(() => {
    return openKeys.value as string[];
  });

  const getMode = computed(() => {
    return props.mode;
  });

  const getSelectedKeys = computed(() => {
    return unref(selectedKeys);
  });

  // 跟随页面路由变化，切换菜单选中状态
  watch(
    () => currentRoute.fullPath,
    () => {
      updateMenu();
      const matched = currentRoute.matched;
      openKeys.value = matched.map((item) => item.name);
      const activeMenu: string = (currentRoute.meta?.activeMenu as string) || '';
      selectedKeys.value = activeMenu ? (activeMenu as string) : (currentRoute.name as string);
      emits('update:showMobileSider', false);
    },
  );

  function updateMenu() {
    menus.value = generatorMenu(replaceI18nTitle(asyncRouteStore.getMenus));
  }

  // 点击菜单
  function clickMenuItem(key: string) {
    if (/http(s)?:/.test(key)) {
      window.open(key);
    } else {
      if (key === currentRoute.name) return;
      router.push({ name: key });
    }
  }

  //展开菜单
  function menuExpanded(keys: string[]) {
    if (!keys) return;
    const latestOpenKey = keys.find((key) => openKeys.value.indexOf(key) === -1);
    const isExistChildren = findChildrenLen(latestOpenKey as string);
    openKeys.value = isExistChildren ? (latestOpenKey ? [latestOpenKey] : []) : keys;
  }

  //查找是否存在子路由
  function findChildrenLen(key: string) {
    if (!key) return false;
    const subRouteChildren: string[] = [];
    for (const { children, key } of unref(menus)) {
      if (children && children.length) {
        subRouteChildren.push(key as string);
      }
    }
    return subRouteChildren.includes(key);
  }

  onMounted(() => {
    updateMenu();
  });
</script>

<style lang="less" scoped>
  .layout-sider :deep(.n-layout-sider .n-layout-toggle-button) {
    right: 12px;
  }
</style>
