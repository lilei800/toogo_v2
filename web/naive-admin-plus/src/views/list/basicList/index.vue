<template>
  <div>
    <n-card :bordered="false" class="pt-3 mb-3 proCard">
      <BasicForm
        @register="register"
        @submit="handleSubmit"
        @reset="handleReset"
        @advanced="formAdvanced"
      >
        <template #statusSlot="{ model, field }">
          <n-input v-model:value="model[field]" />
        </template>
      </BasicForm>
    </n-card>
    <n-card :bordered="false" class="proCard">
      <BasicTable
        ref="tableRef"
        :columns="columns"
        :request="loadDataTable"
        :row-key="(row) => row.id"
        :bordered="tableBorder"
        :actionColumn="actionColumn"
        @checked-row-change="onCheckedRow"
        @columns-change="onColumnsChange"
        :autoScrollX="true"
        :checkedRowAlert="true"
        :get-csv-cell="getCsvCell"
        :downloadCsvProps="{
          keepOriginalData: false,
        }"
        :printConfig="{
          rowKey: 'id',
          header: '我是自定义打印的顶部区域',
        }"
      >
        <template #tableTitle>
          <n-flex>
            <n-button type="primary" @click="addTable">
              <template #icon>
                <n-icon>
                  <PlusOutlined />
                </n-icon>
              </template>
              新建
            </n-button>
            <n-dropdown trigger="hover" :options="batchOptions" @select="handleBatchSelect">
              <n-button icon-placement="right">
                批量操作
                <template #icon>
                  <n-icon :size="14">
                    <ChevronDownOutline />
                  </n-icon>
                </template>
              </n-button>
            </n-dropdown>
          </n-flex>
        </template>
      </BasicTable>
    </n-card>
    <n-modal v-model:show="showModal" :show-icon="false" preset="dialog" title="新建用户">
      <div class="pt-6">
        <BasicForm @register="drawerRegister" />
      </div>

      <template #action>
        <n-space>
          <n-button @click="() => (showModal = false)">取消</n-button>
          <n-button type="info" :loading="formBtnLoading" @click="addSubmit">保存</n-button>
        </n-space>
      </template>
    </n-modal>

    <DrawerForm
      ref="editDrawerRef"
      :width="450"
      :drawerContent="{
        title: '编辑用户',
      }"
      positiveText="保存"
      :form="form"
      @submit="formSubmit"
    >
      <div class="mt-6">
        <BasicForm @register="drawerRegister" />
      </div>
    </DrawerForm>
  </div>
</template>

<script lang="ts" setup>
  import { h, reactive, ref, unref, nextTick } from 'vue';
  import { DataTableGetCsvCell, useMessage } from 'naive-ui';
  import { BasicTable, TableAction, BasicColumn } from '@/components/Table';
  import { BasicForm, useForm } from '@/components/Form/index';
  import { schemas } from './schemas';
  import { getTableList } from '@/api/table/list';
  import { columns } from './columns';
  import {
    PlusOutlined,
    EditOutlined,
    DeleteOutlined,
    VerticalAlignBottomOutlined,
  } from '@vicons/antd';
  import { ChevronDownOutline } from '@vicons/ionicons5';
  import { useRoute } from 'vue-router';
  import { renderIcon } from '@/utils';

  const route = useRoute();
  const message = useMessage();
  const tableRef = ref();
  const tableBorder = ref(false);
  const editDrawerRef = ref();

  const showModal = ref(false);
  const formBtnLoading = ref(false);
  const formParams = reactive({
    name: '',
    address: '',
    date: null,
  });

  const params = ref({
    name: '',
    id: route.params.id,
  });

  const [drawerRegister, form] = useForm({
    gridProps: { cols: 1 },
    layout: 'horizontal',
    submitButtonText: '提交预约',
    showActionButtonGroup: false,
    schemas,
  });

  const batchMsgMap = {
    edit: '批量修改',
    export: '批量导出',
    delete: '批量删除',
  };

  const batchOptions = ref([
    {
      label: '批量修改',
      key: 'edit',
      icon: renderIcon(EditOutlined),
    },
    {
      label: '批量导出',
      key: 'export',
      icon: renderIcon(VerticalAlignBottomOutlined),
    },
    {
      key: 'header-divider',
      type: 'divider',
    },
    {
      label: '批量删除',
      icon: renderIcon(DeleteOutlined),
      key: 'delete',
    },
  ]);

  const actionColumn: BasicColumn = reactive({
    width: 140,
    title: '操作',
    key: 'action',
    fixed: 'right',
    align: 'center',
    render(record) {
      return h(TableAction as any, {
        style: 'text',
        actions: [
          {
            label: '删除',
            type: 'error',
            size: 'small',
            onPositiveClick: handleDelete.bind(null, record),
            onNegativeClick: handleNegative.bind(null, record),
            isConfirm: true,
            confirmContent: '确定要删除吗？',
            positiveText: '确定',
          },
          {
            label: '修改',
            type: 'info',
            size: 'small',
            onClick: handleEdit.bind(null, record),
          },
        ],
      });
    },
  });

  const [register, {}] = useForm({
    gridProps: { cols: '1 s:1 m:2 l:3 xl:4 2xl:4' },
    labelWidth: 80,
    schemas,
  });

  function handleBatchSelect(key) {
    message.info(`您点击了，${batchMsgMap[key]} 按钮`);
  }

  function onColumnsChange(columns) {
    console.log('columns', columns);
  }

  function formAdvanced(status) {
    console.log(status);
    tableRef.value.redoHeight();
  }

  function addTable() {
    showModal.value = true;
  }

  const loadDataTable = async (res) => {
    return await getTableList({ ...formParams, ...params.value, ...res });
  };

  function onCheckedRow(rowKeys) {
    console.log(rowKeys.value);
  }

  function reloadTable() {
    tableRef.value.reload();
  }

  function handleEdit(record: Recordable) {
    console.log('点击了编辑', record);
    editDrawerRef.value.showDrawer();
    nextTick(() => {
      form.setFieldsValue({
        ...unref(record),
      });
    });
  }

  function handleDelete(record: Recordable) {
    console.log('点击了删除', record);
    window['$message'].info('点击了删除');
  }

  function handleNegative(record: Recordable) {
    console.log('点击了取消', record);
    message.info('点击了取消');
  }

  function handleSubmit(values: Recordable) {
    params.value = Object.assign(formParams, values) as any;
    reloadTable();
  }

  function handleReset(values: Recordable) {
    console.log(values);
  }

  function formSubmit(values, done) {
    console.log('表单值', values);
    // 1、这里做表单 api 提交
    // 2、提交完关闭弹窗
    done(true);
  }

  const getCsvCell: DataTableGetCsvCell = (value, _, column) => {
    if (column.key === 'action') {
      return '';
    }
    return value;
  };

  async function addSubmit() {
    const res = await form.submit();
    if (res) {
      message.success('提交成功');
      showModal.value = false;
    } else {
      message.error('请填写完整信息');
    }
  }
</script>
