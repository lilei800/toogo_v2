<template>
  <div
    ref="basicTableRef"
    :class="{
      'table-full-screen': isFullscreen && !getDarkTheme,
    }"
  >
    <div class="table-toolbar">
      <!--顶部左侧区域-->
      <div class="flex items-center table-toolbar-left">
        <template v-if="tableTitle">
          <div class="table-toolbar-left-title">
            {{ tableTitle }}
            <n-tooltip v-if="tableTitleTooltip" trigger="hover">
              <template #trigger>
                <n-icon class="ml-1 text-gray-400 cursor-pointer" size="18">
                  <QuestionCircleOutlined />
                </n-icon>
              </template>
              {{ tableTitleTooltip }}
            </n-tooltip>
          </div>
        </template>
        <slot name="tableTitle"></slot>
      </div>

      <div class="flex items-center table-toolbar-right">
        <!--顶部右侧区域-->
        <slot name="toolbar"></slot>

        <template v-if="!settingStore.isMobile && isTableSetting">
          <n-flex :size="[2, 0]" align="center">
            <!--表格边框-->

            <div class="flex items-center" v-if="isShowTableBordered">
              <div class="mr-2">边框</div>
              <n-switch size="small" v-model:value="bordered" />
            </div>

            <n-divider vertical v-if="isShowTableBordered" />

            <!--表格斑马纹-->

            <div class="flex items-center" v-if="isShowTableStriped">
              <div class="mr-2">斑马纹</div>
              <n-switch size="small" v-model:value="striped" />
            </div>

            <n-divider vertical v-if="isShowTableStriped" />

            <!--查询-->
            <n-tooltip trigger="hover" v-if="isShowTableQuery">
              <template #trigger>
                <n-button size="small" quaternary circle @click="foldQueryChange()">
                  <template #icon>
                    <n-icon :size="17">
                      <SearchOutlined />
                    </n-icon>
                  </template>
                </n-button>
              </template>
              <span>{{ foldQuery ? '展开查询' : '收起查询' }}</span>
            </n-tooltip>

            <!--刷新-->
            <n-tooltip trigger="hover" v-if="isShowTableRedo">
              <template #trigger>
                <n-button size="small" quaternary circle @click="reloadTable()">
                  <template #icon>
                    <n-icon :size="17">
                      <ReloadOutlined />
                    </n-icon>
                  </template>
                </n-button>
              </template>
              <span>刷新</span>
            </n-tooltip>

            <!-- 导出 -->
            <n-tooltip trigger="hover" v-if="isShowTableExport">
              <template #trigger>
                <n-button size="small" quaternary circle @click="downloadCsv()">
                  <template #icon>
                    <n-icon :size="17">
                      <DownloadOutlined />
                    </n-icon>
                  </template>
                </n-button>
              </template>
              <span>导出</span>
            </n-tooltip>

            <!-- 打印 -->
            <n-tooltip trigger="hover" v-if="isShowTablePrint">
              <template #trigger>
                <n-button size="small" quaternary circle @click="printModalRef?.openModal()">
                  <template #icon>
                    <n-icon :size="17">
                      <PrinterOutlined />
                    </n-icon>
                  </template>
                </n-button>
              </template>
              <span>打印</span>
            </n-tooltip>

            <!--密度-->
            <n-tooltip trigger="hover" v-if="isShowTableSize">
              <template #trigger>
                <n-button size="small" quaternary circle>
                  <n-dropdown
                    v-model:value="tableSize"
                    :options="densityOptions"
                    trigger="click"
                    @select="densitySelect"
                  >
                    <n-icon :size="17">
                      <ColumnHeightOutlined />
                    </n-icon>
                  </n-dropdown>
                </n-button>
              </template>
              <span>密度</span>
            </n-tooltip>

            <!--全屏-->
            <n-tooltip trigger="hover" v-if="isShowTableFullscreen">
              <template #trigger>
                <n-button size="small" quaternary circle @click="toggleTableFullScreen">
                  <template #icon>
                    <n-icon :size="17">
                      <FullscreenExitOutlined v-if="isFullscreen" />
                      <FullscreenOutlined v-else />
                    </n-icon>
                  </template>
                </n-button>
              </template>
              <span>{{ isFullscreen ? '还原' : '全屏' }}</span>
            </n-tooltip>

            <!--表格设置单独抽离成组件-->
            <ColumnSetting v-if="isShowTableSetting" @columnsChange="columnsChange" />
          </n-flex>
        </template>
      </div>
    </div>
    <div class="mb-4 table-checked-row" v-if="getCheckedRowAlert">
      <n-alert type="info" :show-icon="false" class="checked-alert">
        <n-space>
          <span type="info"
            >已选择 <n-text type="info" strong>{{ checkedRowKeys.length }}</n-text> 项</span
          >
          <n-button type="info" text @click="restCheckedRowKeys">清空</n-button>
        </n-space>
      </n-alert>
    </div>
    <div class="s-table" v-if="isShowTable">
      <n-data-table
        :id="tableId"
        ref="tableElRef"
        v-bind="getBindValues"
        v-model:checked-row-keys="checkedRowKeys"
        @update:checked-row-keys="checkedRowKeysChange"
        @update:filters="handleFiltersChange"
        @update:sorter="handleSorterChange"
        :pagination="pagination"
        @update:page="updatePage"
        @update:page-size="updatePageSize"
      >
        <template v-for="item in Object.keys($slots)" :key="item" #[item]="data">
          <slot v-bind="data || {}" :name="item"></slot>
        </template>
      </n-data-table>
    </div>
    <!-- 打印弹窗 -->
    <PrintModal ref="printModalRef" />
  </div>
</template>

<script lang="ts" setup>
  import { ref, useId, unref, toRaw, computed, onMounted, nextTick } from 'vue';

  import { createTableContext } from './hooks/useTableContext';
  import ColumnSetting from './components/settings/ColumnSetting.vue';
  import PrintModal from './components/Print/index.vue';

  import { useLoading } from './hooks/useLoading';
  import { useColumns } from './hooks/useColumns';
  import { useDataSource } from './hooks/useDataSource';
  import { usePagination } from './hooks/usePagination';
  import { useSearch } from './hooks/useSearch';

  import { basicProps } from './props';
  import { BasicTableProps } from './types/table';

  import { getViewportOffset } from '@/utils/domUtils';
  import { useWindowSizeFn } from '@/hooks/event/useWindowSizeFn';
  import { isBoolean } from '@/utils/is';
  import { useDesignSetting } from '@/hooks/setting/useDesignSetting';
  import { useProjectSettingStore } from '@/store/modules/projectSetting';
  import {
    ReloadOutlined,
    ColumnHeightOutlined,
    QuestionCircleOutlined,
    FullscreenExitOutlined,
    FullscreenOutlined,
    SearchOutlined,
    DownloadOutlined,
    PrinterOutlined,
  } from '@vicons/antd';
  import { useFullscreen } from '@vueuse/core';
  import { NSpace, NIcon, NButton, DataTableInst } from 'naive-ui';
  import { merge } from 'lodash-es';

  const props = defineProps({
    ...basicProps,
  });

  const emit = defineEmits([
    'fetch-success',
    'fetch-error',
    'checked-row-change',
    'edit-end',
    'edit-cancel',
    'edit-row-end',
    'edit-change',
    'fold-query-change',
    'columns-change',
  ]);

  const densityOptions = [
    {
      type: 'menu',
      label: '紧凑',
      key: 'small',
    },
    {
      type: 'menu',
      label: '默认',
      key: 'medium',
    },
    {
      type: 'menu',
      label: '宽松',
      key: 'large',
    },
  ];

  const tableId = `basic-table-${useId()}`;
  const foldQuery = ref(false);
  const striped = ref(false);
  const bordered = ref(false);
  const isShowTable = ref(true);
  const deviceHeight = ref<Number | String>('auto');
  const tableElRef = ref<DataTableInst>();
  const basicTableRef = ref<HTMLElement | null>(null);
  const wrapRef = ref(null);
  const printModalRef = ref<InstanceType<typeof PrintModal>>();
  const checkedRowKeys = ref<number[] | string[]>([]);
  let paginationEl: HTMLElement | null;

  const tableData = ref<Recordable[]>([]);
  const innerPropsRef = ref<Partial<BasicTableProps>>();

  const { isFullscreen, toggle } = useFullscreen(basicTableRef as any);

  const getProps = computed(() => {
    return { ...props, ...unref(innerPropsRef) } as unknown as BasicTableProps;
  });

  const tableTitle = unref(getProps).title || '';

  const tableTitleTooltip = unref(getProps).titleTooltip || '';

  const settingStore = useProjectSettingStore();

  const { getAppTheme, getDarkTheme } = useDesignSetting();

  const { getLoading, setLoading } = useLoading(getProps);

  const { getPaginationInfo, setPagination } = usePagination(getProps);

  const {
    getDataSourceRef,
    getRowKey,
    reload,
    restReload,
    setTableData,
    getDataSource,
    updateTableData,
    updateTableDataRecord,
    deleteTableDataRecord,
  } = useDataSource(
    getProps,
    {
      getPaginationInfo,
      setPagination,
      tableData,
      setLoading,
    },
    emit,
  );

  const { getPageColumns, setColumns, getColumns, getCacheColumns, setCacheColumnsField } =
    useColumns(getProps);

  const tableSize = ref(unref(getProps as any).size || 'medium');

  //返回表格 ref 实例
  function getTableRef() {
    return tableElRef.value;
  }

  //导出 Csv
  function downloadCsv(fileName?: string, keepOriginalData?: boolean) {
    const name = fileName || new Date().getTime() + '';
    const keepData = keepOriginalData || undefined;
    tableElRef.value?.downloadCsv({
      fileName: name,
      keepOriginalData: keepData,
      ...getProps.value.downloadCsvProps,
    });
  }

  //是否显示 选中行提示
  const getCheckedRowAlert = computed(() => {
    return unref(getProps as any).checkedRowAlert;
  });

  //表格全屏
  function toggleTableFullScreen() {
    toggle();
  }

  //table内部刷新
  async function reloadTable(opt?) {
    !getProps.value.isKeepRowKeys && (await restCheckedRowKeys());
    //TODO：这里获取filter和order参数
    const filters = getAllFilter();
    reload(merge(filters, opt));
  }

  //页码切换
  async function updatePage(page) {
    !getProps.value.isKeepRowKeys && (await restCheckedRowKeys());
    setPagination({ page });
    reloadTable();
  }

  //分页数量切换
  async function updatePageSize(size) {
    !getProps.value.isKeepRowKeys && (await restCheckedRowKeys());
    setPagination({ page: 1, pageSize: size });
    reloadTable();
  }

  //密度切换
  function densitySelect(e) {
    tableSize.value = e;
  }

  // 提取 tableSetting 的获取和默认值处理
  const tableSetting = computed(() => {
    const setting = getProps.value.tableSetting || {};
    return {
      bordered: setting.bordered ?? true,
      redo: setting.redo ?? true,
      export: setting.export ?? true,
      print: setting.print ?? true,
      size: setting.size ?? true,
      setting: setting.setting ?? true,
      fullscreen: setting.fullscreen ?? true,
      striped: setting.striped ?? true,
      query: setting.query ?? false,
    };
  });

  //获取表格大小
  const getTableSize = computed(() => tableSize.value);

  //获取斑马纹
  const getStriped = computed(() => striped.value);

  //表格边框
  const getBordered = computed(() => bordered.value);

  //是否显示表格边框
  const isShowTableBordered = computed(() => tableSetting.value?.bordered);

  //表格设置工具
  const isTableSetting = computed(() => getProps.value.showTableSetting);

  //是否显示刷新按钮
  const isShowTableRedo = computed(() => tableSetting.value?.redo);

  //是否显示导出按钮
  const isShowTableExport = computed(() => tableSetting.value?.export);

  // 是否显示打印按钮
  const isShowTablePrint = computed(() => tableSetting.value?.print);

  //是否显示尺寸调整按钮
  const isShowTableSize = computed(() => tableSetting.value?.size);

  //是否显示字段调整按钮
  const isShowTableSetting = computed(() => tableSetting.value?.setting);

  //是否显示表格全屏按钮
  const isShowTableFullscreen = computed(() => tableSetting.value?.fullscreen);

  //是否显示斑马纹开关
  const isShowTableStriped = computed(() => tableSetting.value?.striped);

  //是否显示查询表单 AdvancedTable 组件独有
  const isShowTableQuery = computed(() => tableSetting.value?.query);

  //计算高度
  const getDeviceHeight = computed(() => {
    const tableData = unref(getDataSourceRef);
    if (deviceHeight.value === 'auto') return 'auto';
    const maxHeight = tableData.length ? `${unref(deviceHeight)}px` : 'auto';
    return maxHeight;
  });

  //自动计算表格 scroll-x 数值
  const getColumnsScrollX = computed(() => {
    let num = 0;
    unref(getPageColumns).forEach((item) => {
      if (item.width) {
        num += item.width as number;
      }
    });
    return num;
  });

  //组装表格信息
  const getBindValues = computed(() => {
    const tableData = unref(getDataSourceRef);
    const autoScrollX = unref(getProps).autoScrollX;
    return {
      ...unref(getProps),
      loading: unref(getLoading),
      columns: toRaw(unref(getPageColumns)),
      rowKey: unref(getRowKey),
      data: tableData,
      size: unref(getTableSize),
      striped: unref(getStriped),
      bordered: unref(getBordered),
      singleLine: getBordered.value ? false : true,
      remote: true,
      'max-height': getDeviceHeight.value,
      'scroll-x': autoScrollX ? unref(getColumnsScrollX) : undefined,
    };
  });

  //columns 列变动
  function columnsChange(columns) {
    emit('columns-change', columns);
  }

  //折叠查询
  function foldQueryChange() {
    foldQuery.value = !foldQuery.value;
    emit('fold-query-change', foldQuery.value);
  }

  //选择行
  function checkedRowKeysChange(rowKeys) {
    checkedRowKeys.value = rowKeys;
    emit('checked-row-change', checkedRowKeys);
  }

  //清空行
  function restCheckedRowKeys() {
    checkedRowKeys.value = [];
    emit('checked-row-change', checkedRowKeys);
    redoHeight();
  }

  //更新选中行
  function setCheckedRowKeys(rowKeys) {
    checkedRowKeys.value = rowKeys;
    emit('checked-row-change', checkedRowKeys);
  }

  // #region anson:hack
  const handleSorterChange = (sorter) => {
    ref(unref(getProps).columns).value.forEach((column) => {
      if (column.sortOrder === undefined) return;
      if (!sorter) {
        column.sortOrder = false;
        return;
      }
      if (column.key === sorter.columnKey) {
        column.sortOrder = sorter.order;
      } else {
        column.sortOrder = false;
      }
    });
    //TODO：这里要获取所有的sorter和filters
    reloadTable();
  };

  const handleFiltersChange = (filters, sourceColumn) => {
    ref(unref(getProps).columns).value.forEach((column) => {
      /** column.sortOrder !== undefined means it is uncontrolled */
      if (column.filter === undefined) return;
      if (!filters) {
        return;
      }
      if (column.key === sourceColumn.key) {
        column.filterOptionValue = filters[sourceColumn.key];
      }
    });
    reloadTable();
  };

  const registerSearch = () => {
    ref(unref(getProps).columns).value.forEach((item) => {
      if (item.search) {
        const { renderFilterIcon, renderFilterMenu, searchVal } = useSearch(item, reloadTable);
        item.renderFilterIcon = renderFilterIcon;
        item.renderFilterMenu = renderFilterMenu;
        item.searchVal = searchVal;
        item.filter = 'default';
      }
    });
  };

  // TOOD 可根据项目情况自行调整数据格式
  function getAllFilter() {
    let filter: any = null;
    let order: any = null;
    getPageColumns.value.forEach((column: any) => {
      if (column.sortOrder) {
        if (!order) order = {};
        order[column.key] = column.sortOrder.replace('end', '');
      }
      const inlist = toRaw(column.filterOptionValue);

      if (inlist && inlist.length > 0) {
        if (!filter) filter = {};
        filter[column.key] = { _in: inlist };
      }
      if (column.search && column.searchVal) {
        let reg = '_eq';
        if (column.searchReg) {
          reg = column.searchReg;
        }
        let obj = {};
        let input = column.searchVal;
        if (column.searchFixKey) {
          input = column.searchFixKey(input);
        }
        obj[reg] = input;
        if (!filter) filter = {};
        filter[column.key] = obj;
      }
    });
    return { order, where: filter };
  }

  //重新计算表格高度
  function redoHeight() {
    computeTableHeight();
  }

  //获取分页信息
  const pagination = computed(() => {
    return {
      simple: settingStore.isMobile,
      ...toRaw(unref(getPaginationInfo)),
    };
  });

  function setProps(props: Partial<BasicTableProps>) {
    innerPropsRef.value = { ...unref(innerPropsRef), ...props };
  }

  const tableAction = {
    reload,
    restReload,
    reloadTable,
    restCheckedRowKeys,
    setCheckedRowKeys,
    redoHeight,
    setColumns,
    setLoading,
    getTableRef,
    setProps,
    getColumns,
    getPageColumns,
    getCacheColumns,
    setCacheColumnsField,
    downloadCsv,
    emit,
  };

  createTableContext({
    ...tableAction,
    wrapRef,
    getBindValues,
    isShowTable,
    checkedRowKeys,
    tableId,
  });

  const getCanResize = computed(() => {
    const { canResize } = unref(getProps);
    return canResize;
  });

  async function computeTableHeight() {
    const table: any = unref(tableElRef);
    if (!table) return;
    if (!unref(getCanResize)) return;
    const tableEl: any = table?.$el;
    await nextTick();
    const headEl = tableEl.querySelector('.n-data-table-thead');
    const { bottomIncludeBody } = getViewportOffset(headEl);
    const headerH = 64;
    let paginationH = 2;
    let marginH = 22;
    if (!isBoolean(pagination)) {
      paginationEl = tableEl.querySelector('.n-data-table__pagination') as HTMLElement;
      if (paginationEl) {
        const offsetHeight = paginationEl.offsetHeight;
        paginationH += offsetHeight || 0;
      } else {
        paginationH += 28;
      }
    }
    let height: number =
      bottomIncludeBody - (headerH + paginationH + marginH + (props.resizeHeightOffset || 0));
    const maxHeight = (props.maxHeight || 0) as number;
    height = (maxHeight && maxHeight < height ? maxHeight : height) as number;
    deviceHeight.value = height;
  }

  useWindowSizeFn(computeTableHeight, 280);

  onMounted(() => {
    registerSearch();
    nextTick(() => {
      computeTableHeight();
    });
  });

  //导出方法到外部使用
  defineExpose({
    reload,
    restReload,
    reloadTable, //anson add
    restCheckedRowKeys,
    setCheckedRowKeys,
    getDataSource,
    getColumns,
    setColumns,
    setLoading,
    getTableRef,
    setTableData,
    updateTableData,
    updateTableDataRecord,
    deleteTableDataRecord,
    redoHeight,
    downloadCsv,
  });
</script>
<style lang="less" scoped>
  .table-toolbar {
    display: flex;
    justify-content: space-between;
    padding: 0 0 16px 0;

    &-left {
      display: flex;
      align-items: center;
      justify-content: flex-start;
      flex: 2;

      &-title {
        display: flex;
        align-items: center;
        justify-content: flex-start;
        font-size: 16px;
        font-weight: 600;
      }
    }

    &-right {
      display: flex;
      justify-content: flex-end;
      flex: 1;

      &-icon {
        margin-left: 12px;
        font-size: 16px;
        cursor: pointer;
        color: var(--n-text-color);

        :hover {
          color: v-bind(getAppTheme);
        }
      }
    }
  }

  .table-toolbar-inner-popover-title {
    padding: 2px 0;
  }

  .table-full-screen {
    background: #fff;
    padding: 20px;
  }

  .checked-alert {
    :deep(.n-alert-body) {
      padding: 8px 47px 8px 15px;
    }
  }
</style>
