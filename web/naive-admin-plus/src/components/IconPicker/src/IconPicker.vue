<template>
  <n-input-group>
    <n-button>
      <template #icon v-if="iconValue">
        <Icon :icon="getIconValue" :key="getIconValue" />
      </template>
    </n-button>
    <n-select
      placeholder="请选择图标"
      :style="getWidth"
      :consistent-menu-width="false"
      v-model:value="getIconValue"
      class="i-select"
    >
      <template #empty>
        <div class="i-picker" :style="getContentWidth">
          <div class="i-picker-search" v-if="isShowSearch">
            <n-input placeholder="可输入单词搜索图标哦~" v-model:value="keyword" clearable>
              <template #suffix>
                <n-icon :component="SearchOutlined" />
              </template>
            </n-input>
          </div>

          <n-tabs
            type="line"
            v-model:value="currentTab"
            class="i-picker-tab"
            @before-leave="beforeLeave"
            @update:value="tabsChange"
          >
            <template #suffix>
              <n-tooltip trigger="hover">
                <template #trigger>
                  <n-icon
                    :component="SearchOutlined"
                    :size="22"
                    class="cursor-pointer"
                    @click="isShowSearch = !isShowSearch"
                    :depth="3"
                  />
                </template>
                {{ isShowSearch ? '隐藏' : '显示' }}搜索
              </n-tooltip>
            </template>
            <n-tab-pane name="antd" tab="antd" display-directive="show:lazy">
              <n-spin :show="loading" description="加载中...">
                <IconItem :list="antdIcons" v-model:value="iconValue" />
                <n-empty v-if="!antdIcons.length" />
              </n-spin>
            </n-tab-pane>
            <n-tab-pane name="ionicons5" tab="ionicons5" display-directive="show:lazy">
              <n-spin :show="loading" description="加载中...">
                <IconItem :list="ioniconsIcons" v-model:value="iconValue" />
                <n-empty v-if="!ioniconsIcons.length" />
              </n-spin>
            </n-tab-pane>
          </n-tabs>
          <n-divider class="mt-4 mb-4" />
          <div class="flex justify-center mt-6 picker-pagination">
            <n-pagination
              show-size-picker
              :page-sizes="[60, 120, 240]"
              v-model:page="currentPage"
              :pageSize="pageSize"
              :show-total="(total) => `共 ${total} 个`"
              :item-count="currentTab === 'antd' ? antdGetTotal : ioniconsGetTotal"
              size="small"
              @update:page="handlePageChange"
              @update:page-size="handlePageSizeChange"
            />
          </div>
        </div>
      </template>
    </n-select>
  </n-input-group>
</template>

<script lang="ts" setup>
  import { SearchOutlined } from '@vicons/antd';
  import { computed, ref, watch, onMounted } from 'vue';
  import { basicProps } from './props';
  import IconItem from './components/IconItem.vue';
  import { usePagination } from '@/hooks/web/usePagination';
  import { Icon } from '@/components/Icon';
  import { debounce } from 'lodash-es';

  const currentTab = ref('antd');
  const isShowSearch = ref<boolean>(false);
  const loading = ref<boolean>(false);
  const keyword = ref<string>('');
  const iconValue = ref();
  const currentPage = ref<number>(1);
  const pageSize = ref<number>(60);

  // 存储已加载的图标名称
  const iconCache = ref<Record<string, string[]>>({
    antd: [],
    ionicons5: [],
  });

  // 动态加载图标名称
  const loadIconNames = async (type: 'antd' | 'ionicons5') => {
    // 如果已经加载过，直接返回
    if (iconCache.value[type].length > 0) {
      return iconCache.value[type];
    }

    loading.value = true;
    try {
      if (type === 'antd') {
        const module = await import('@vicons/antd');
        const icons = Object.keys(module).map((item) => `antd:${item}`);
        iconCache.value.antd = icons;
        return icons;
      } else {
        const module = await import('@vicons/ionicons5');
        const icons = Object.keys(module).map((item) => `ionicons5:${item}`);
        iconCache.value.ionicons5 = icons;
        return icons;
      }
    } finally {
      loading.value = false;
    }
  };

  // 过滤函数
  function filter(item: string, keyword: string) {
    return item.toLowerCase().indexOf(keyword.toLowerCase()) !== -1;
  }

  const getFilteredIcons = computed(() => {
    const source = iconCache.value[currentTab.value];
    if (!source || source.length === 0) return [];

    return source.filter((item) => {
      return filter(item, keyword.value);
    });
  });

  const antdIconsList = ref<string[]>([]);
  const ionicons5List = ref<string[]>([]);

  watch(getFilteredIcons, (newIcons) => {
    if (currentTab.value === 'antd') {
      antdIconsList.value = newIcons;
    } else {
      ionicons5List.value = newIcons;
    }
  });

  // 防抖处理搜索关键词
  const debouncedSearch = debounce(() => {
    if (iconCache.value[currentTab.value].length === 0) {
      loadIconNames(currentTab.value as 'antd' | 'ionicons5');
    }
  }, 300);

  // 监听关键词变化，实时过滤图标
  watch(keyword, debouncedSearch);

  const {
    getPaginationList: antdIcons,
    getTotal: antdGetTotal,
    setCurrentPage: antdSetCurrentPage,
    setPageSize: antdSetPageSize,
  } = usePagination(antdIconsList, pageSize.value);

  const {
    getPaginationList: ioniconsIcons,
    getTotal: ioniconsGetTotal,
    setCurrentPage: ioniconsSetCurrentPage,
    setPageSize: ioniconsSetPageSize,
  } = usePagination(ionicons5List, pageSize.value);

  function handlePageChange(page: number) {
    currentPage.value = page;
    currentTab.value === 'antd' ? antdSetCurrentPage(page) : ioniconsSetCurrentPage(page);
  }

  function handlePageSizeChange(size: number) {
    pageSize.value = size;
    currentTab.value === 'antd' ? antdSetPageSize(size) : ioniconsSetPageSize(size);
  }

  const getIconValue = computed(() => {
    return iconValue.value;
  });

  const props = defineProps({
    ...basicProps,
  });

  const getWidth = computed(() => {
    return {
      width: props.width,
    };
  });

  const getContentWidth = computed(() => {
    return {
      width: props.contentWidth,
    };
  });

  function beforeLeave() {
    loading.value = true;
    return true;
  }

  async function tabsChange() {
    loading.value = true;
    try {
      currentPage.value = 1;
      await loadIconNames(currentTab.value as 'antd' | 'ionicons5');
      antdSetCurrentPage(1);
      ioniconsSetCurrentPage(1);
      loading.value = false;
    } finally {
      loading.value = false;
    }
  }

  // 初始化加载第一个标签页
  onMounted(async () => {
    await loadIconNames('antd');
  });
</script>

<style lang="less" scoped>
  .i-picker {
    padding: 0 0 10px;
    &-search {
      margin-bottom: 10px;
    }
  }
</style>
