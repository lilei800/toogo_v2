<template>
  <div class="toogo-crud">
    <!-- 搜索表单 -->
    <n-card v-if="showSearch" class="search-card" size="small">
      <n-form
        ref="searchFormRef"
        :model="searchModel"
        :label-width="80"
        label-placement="left"
        inline
      >
        <slot name="search" :model="searchModel">
          <n-form-item
            v-for="item in searchSchema"
            :key="item.field"
            :label="item.label"
            :path="item.field"
          >
            <component
              :is="getComponent(item.component)"
              v-model:value="searchModel[item.field]"
              v-bind="item.componentProps"
              :placeholder="item.placeholder || `请输入${item.label}`"
              clearable
              style="width: 180px"
            />
          </n-form-item>
        </slot>
        <n-form-item>
          <n-space>
            <n-button type="primary" @click="handleSearch">
              <template #icon
                ><n-icon><SearchOutline /></n-icon
              ></template>
              搜索
            </n-button>
            <n-button @click="handleReset">
              <template #icon
                ><n-icon><RefreshOutline /></n-icon
              ></template>
              重置
            </n-button>
          </n-space>
        </n-form-item>
      </n-form>
    </n-card>

    <!-- 工具栏 -->
    <n-card class="table-card" size="small">
      <template #header>
        <div class="table-header">
          <div class="table-title">{{ title }}</div>
          <div class="table-actions">
            <slot name="toolbar">
              <n-space>
                <n-button v-if="showCreate" type="primary" @click="handleCreate">
                  <template #icon
                    ><n-icon><AddOutline /></n-icon
                  ></template>
                  新增
                </n-button>
                <n-button v-if="showExport" @click="handleExport">
                  <template #icon
                    ><n-icon><DownloadOutline /></n-icon
                  ></template>
                  导出
                </n-button>
                <n-button @click="handleRefresh">
                  <template #icon
                    ><n-icon><RefreshOutline /></n-icon
                  ></template>
                  刷新
                </n-button>
              </n-space>
            </slot>
          </div>
        </div>
      </template>

      <!-- 数据表格 -->
      <n-data-table
        ref="tableRef"
        :columns="tableColumns"
        :data="tableData"
        :loading="loading"
        :pagination="pagination"
        :row-key="rowKey"
        :scroll-x="scrollX"
        :bordered="bordered"
        :single-line="singleLine"
        @update:page="handlePageChange"
        @update:page-size="handlePageSizeChange"
      />
    </n-card>

    <!-- 编辑弹窗 -->
    <n-modal
      v-model:show="showModal"
      :title="modalTitle"
      preset="card"
      :style="{ width: modalWidth }"
      :mask-closable="false"
    >
      <n-form
        ref="formRef"
        :model="formModel"
        :rules="formRules"
        :label-width="100"
        label-placement="left"
      >
        <slot name="form" :model="formModel" :is-edit="isEdit">
          <n-form-item
            v-for="item in formSchema"
            :key="item.field"
            :label="item.label"
            :path="item.field"
          >
            <component
              :is="getComponent(item.component)"
              v-model:value="formModel[item.field]"
              v-bind="item.componentProps"
              :placeholder="item.placeholder || `请输入${item.label}`"
              :disabled="item.disabled || (isEdit && item.disabledOnEdit)"
              style="width: 100%"
            />
          </n-form-item>
        </slot>
      </n-form>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showModal = false">取消</n-button>
          <n-button type="primary" :loading="submitLoading" @click="handleSubmit"> 确定 </n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
  import { ref, reactive, computed, onMounted, h } from 'vue';
  import {
    NInput,
    NSelect,
    NDatePicker,
    NInputNumber,
    NSwitch,
    useMessage,
    useDialog,
  } from 'naive-ui';
  import {
    SearchOutline,
    RefreshOutline,
    AddOutline,
    DownloadOutline,
    CreateOutline,
    TrashOutline,
  } from '@vicons/ionicons5';

  // Props定义
  interface SchemaItem {
    field: string;
    label: string;
    component?: string;
    componentProps?: Record<string, any>;
    placeholder?: string;
    disabled?: boolean;
    disabledOnEdit?: boolean;
  }

  interface Props {
    title?: string;
    api: {
      list: (params: any) => Promise<any>;
      create?: (data: any) => Promise<any>;
      update?: (data: any) => Promise<any>;
      delete?: (data: any) => Promise<any>;
      export?: (params: any) => Promise<any>;
    };
    columns: any[];
    searchSchema?: SchemaItem[];
    formSchema?: SchemaItem[];
    formRules?: Record<string, any>;
    rowKey?: string;
    showSearch?: boolean;
    showCreate?: boolean;
    showExport?: boolean;
    modalWidth?: string;
    scrollX?: number;
    bordered?: boolean;
    singleLine?: boolean;
    defaultPageSize?: number;
    beforeSubmit?: (data: any, isEdit: boolean) => any;
    afterSubmit?: (res: any, isEdit: boolean) => void;
  }

  const props = withDefaults(defineProps<Props>(), {
    title: '数据列表',
    rowKey: 'id',
    showSearch: true,
    showCreate: true,
    showExport: false,
    modalWidth: '600px',
    scrollX: 1200,
    bordered: false,
    singleLine: false,
    defaultPageSize: 10,
  });

  const emit = defineEmits(['create', 'edit', 'delete', 'refresh']);

  const message = useMessage();
  const dialog = useDialog();

  // 状态
  const loading = ref(false);
  const submitLoading = ref(false);
  const showModal = ref(false);
  const isEdit = ref(false);
  const tableData = ref<any[]>([]);
  const searchModel = reactive<Record<string, any>>({});
  const formModel = reactive<Record<string, any>>({});
  const searchFormRef = ref();
  const formRef = ref();
  const tableRef = ref();

  // 分页
  const pagination = reactive({
    page: 1,
    pageSize: props.defaultPageSize,
    itemCount: 0,
    showSizePicker: true,
    pageSizes: [10, 20, 50, 100],
    prefix: ({ itemCount }: { itemCount: number }) => `共 ${itemCount} 条`,
  });

  // 计算属性
  const modalTitle = computed(() => (isEdit.value ? '编辑' : '新增'));

  const tableColumns = computed(() => {
    const cols = [...props.columns];
    // 添加操作列
    if (props.api.update || props.api.delete) {
      cols.push({
        title: '操作',
        key: 'actions',
        width: 150,
        fixed: 'right',
        render: (row: any) => {
          return h('div', { class: 'action-buttons' }, [
            props.api.update &&
              h(
                'n-button',
                {
                  size: 'small',
                  quaternary: true,
                  type: 'primary',
                  onClick: () => handleEdit(row),
                },
                { default: () => '编辑' },
              ),
            props.api.delete &&
              h(
                'n-button',
                {
                  size: 'small',
                  quaternary: true,
                  type: 'error',
                  onClick: () => handleDelete(row),
                },
                { default: () => '删除' },
              ),
          ]);
        },
      });
    }
    return cols;
  });

  // 获取组件
  const getComponent = (name?: string) => {
    const componentMap: Record<string, any> = {
      input: NInput,
      select: NSelect,
      datePicker: NDatePicker,
      inputNumber: NInputNumber,
      switch: NSwitch,
    };
    return componentMap[name || 'input'] || NInput;
  };

  // 初始化搜索模型
  const initSearchModel = () => {
    props.searchSchema?.forEach((item) => {
      searchModel[item.field] = null;
    });
  };

  // 初始化表单模型
  const initFormModel = () => {
    props.formSchema?.forEach((item) => {
      formModel[item.field] = null;
    });
  };

  // 加载数据
  const loadData = async () => {
    loading.value = true;
    try {
      const params = {
        page: pagination.page,
        pageSize: pagination.pageSize,
        ...searchModel,
      };
      const res = await props.api.list(params);
      tableData.value = res.list || [];
      pagination.itemCount = res.totalCount || 0;
    } catch (error: any) {
      message.error(error.message || '加载失败');
    } finally {
      loading.value = false;
    }
  };

  // 搜索
  const handleSearch = () => {
    pagination.page = 1;
    loadData();
  };

  // 重置
  const handleReset = () => {
    initSearchModel();
    pagination.page = 1;
    loadData();
  };

  // 刷新
  const handleRefresh = () => {
    loadData();
    emit('refresh');
  };

  // 新增
  const handleCreate = () => {
    isEdit.value = false;
    initFormModel();
    showModal.value = true;
    emit('create');
  };

  // 编辑
  const handleEdit = (row: any) => {
    isEdit.value = true;
    Object.keys(formModel).forEach((key) => {
      formModel[key] = row[key];
    });
    formModel.id = row.id;
    showModal.value = true;
    emit('edit', row);
  };

  // 删除
  const handleDelete = (row: any) => {
    dialog.warning({
      title: '确认删除',
      content: '确定要删除这条记录吗？',
      positiveText: '确定',
      negativeText: '取消',
      onPositiveClick: async () => {
        try {
          await props.api.delete?.({ id: row.id });
          message.success('删除成功');
          loadData();
          emit('delete', row);
        } catch (error: any) {
          message.error(error.message || '删除失败');
        }
      },
    });
  };

  // 提交
  const handleSubmit = async () => {
    try {
      await formRef.value?.validate();
      submitLoading.value = true;

      let data = { ...formModel };
      if (props.beforeSubmit) {
        data = props.beforeSubmit(data, isEdit.value);
      }

      const res = isEdit.value ? await props.api.update?.(data) : await props.api.create?.(data);

      message.success(isEdit.value ? '更新成功' : '创建成功');
      showModal.value = false;
      loadData();

      if (props.afterSubmit) {
        props.afterSubmit(res, isEdit.value);
      }
    } catch (error: any) {
      if (error.message) {
        message.error(error.message);
      }
    } finally {
      submitLoading.value = false;
    }
  };

  // 导出
  const handleExport = async () => {
    if (!props.api.export) return;
    try {
      await props.api.export(searchModel);
      message.success('导出成功');
    } catch (error: any) {
      message.error(error.message || '导出失败');
    }
  };

  // 分页变化
  const handlePageChange = (page: number) => {
    pagination.page = page;
    loadData();
  };

  const handlePageSizeChange = (pageSize: number) => {
    pagination.pageSize = pageSize;
    pagination.page = 1;
    loadData();
  };

  // 暴露方法
  defineExpose({
    loadData,
    handleSearch,
    handleReset,
    handleRefresh,
  });

  // 初始化
  onMounted(() => {
    initSearchModel();
    initFormModel();
    loadData();
  });
</script>

<style scoped lang="less">
  .toogo-crud {
    .search-card {
      margin-bottom: 16px;
    }

    .table-card {
      .table-header {
        display: flex;
        justify-content: space-between;
        align-items: center;

        .table-title {
          font-size: 16px;
          font-weight: 600;
        }
      }
    }

    .action-buttons {
      display: flex;
      gap: 8px;
    }
  }
</style>
