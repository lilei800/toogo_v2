<template>
  <div>
    <n-card :bordered="false" class="proCard">
      <n-space vertical :size="16">
        <n-button type="primary" @click="showModal = true">
          <template #icon><n-icon><PlusOutlined /></n-icon></template>
          添加API配置
        </n-button>

        <n-data-table
          :columns="columns"
          :data="dataList"
          :pagination="pagination"
          :loading="loading"
        />
      </n-space>
    </n-card>

    <!-- 添加/编辑弹窗 -->
    <n-modal v-model:show="showModal" preset="card" :title="editId ? '编辑API配置' : '添加API配置'" style="width: 600px;">
      <n-form ref="formRef" :model="formValue" :rules="rules" label-placement="left" label-width="120">
        <n-form-item label="平台名称" path="platformName">
          <n-select v-model:value="formValue.platformName" :options="platformOptions" />
        </n-form-item>
        <n-form-item label="配置名称" path="name">
          <n-input v-model:value="formValue.name" placeholder="例如:我的Binance" />
        </n-form-item>
        <n-form-item label="API地址" path="apiUrl">
          <n-input v-model:value="formValue.apiUrl" />
        </n-form-item>
        <n-form-item label="API Key" path="apiKey">
          <n-input v-model:value="formValue.apiKey" type="password" show-password-on="click" />
        </n-form-item>
        <n-form-item label="Secret Key" path="secretKey">
          <n-input v-model:value="formValue.secretKey" type="password" show-password-on="click" />
        </n-form-item>
        <n-form-item label="Passphrase">
          <n-input v-model:value="formValue.passphrase" type="password" show-password-on="click" />
        </n-form-item>
        <n-form-item label="设为默认">
          <n-switch v-model:value="formValue.isDefault" />
        </n-form-item>
      </n-form>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showModal = false">取消</n-button>
          <n-button type="primary" :loading="submitting" @click="handleSubmit">确定</n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<script lang="ts" setup>
import { ref, reactive, onMounted, h } from 'vue';
import { NCard, NSpace, NButton, NIcon, NDataTable, NModal, NForm, NFormItem, NInput, NSelect, NSwitch, NTag, useMessage, useDialog } from 'naive-ui';
import { PlusOutlined } from '@vicons/antd';
import { getApiConfigList, createApiConfig, updateApiConfig, deleteApiConfig, testApiConfig } from '@/api/trading/api-config';

const message = useMessage();
const dialog = useDialog();

const dataList = ref([]);
const loading = ref(false);
const showModal = ref(false);
const submitting = ref(false);
const editId = ref(null);

const formValue = reactive({
  platformName: 'binance',
  name: '',
  apiUrl: '',
  apiKey: '',
  secretKey: '',
  passphrase: '',
  isDefault: false,
});

const rules = {
  platformName: { required: true, message: '请选择平台', trigger: 'change' },
  name: { required: true, message: '请输入配置名称', trigger: 'blur' },
  apiUrl: { required: true, message: '请输入API地址', trigger: 'blur' },
  apiKey: { required: true, message: '请输入API Key', trigger: 'blur' },
  secretKey: { required: true, message: '请输入Secret Key', trigger: 'blur' },
};

const platformOptions = [
  { label: 'Binance', value: 'binance' },
  { label: 'OKX', value: 'okx' },
  { label: 'Bitget', value: 'bitget' },
];

const pagination = reactive({ page: 1, pageSize: 10, pageCount: 1 });

const columns = [
  { title: 'ID', key: 'id', width: 80 },
  { title: '配置名称', key: 'name' },
  { title: '平台', key: 'platformName' },
  { title: '是否默认', key: 'isDefault', width: 100, render(row: any) {
    return h(NTag, { type: row.isDefault ? 'success' : 'default' }, { default: () => row.isDefault ? '默认' : '否' });
  }},
  { title: '创建时间', key: 'createdAt', width: 180 },
  { title: '操作', key: 'actions', width: 200, render(row: any) {
    return h(NSpace, {}, {
      default: () => [
        h(NButton, { size: 'small', onClick: () => handleTest(row) }, { default: () => '测试' }),
        h(NButton, { size: 'small', onClick: () => handleEdit(row) }, { default: () => '编辑' }),
        h(NButton, { size: 'small', type: 'error', onClick: () => handleDelete(row) }, { default: () => '删除' }),
      ]
    });
  }},
];

const loadData = async () => {
  loading.value = true;
  try {
    const res = await getApiConfigList({ page: pagination.page, pageSize: pagination.pageSize });
    dataList.value = res.list || [];
    pagination.pageCount = Math.ceil((res.total || 0) / pagination.pageSize);
  } catch (error: any) {
    message.error(error.message || '加载失败');
  } finally {
    loading.value = false;
  }
};

const handleSubmit = async () => {
  submitting.value = true;
  try {
    if (editId.value) {
      await updateApiConfig({ ...formValue, id: editId.value });
      message.success('更新成功');
    } else {
      await createApiConfig(formValue);
      message.success('创建成功');
    }
    showModal.value = false;
    loadData();
  } catch (error: any) {
    message.error(error.message || '操作失败');
  } finally {
    submitting.value = false;
  }
};

const handleEdit = (row: any) => {
  Object.assign(formValue, row);
  editId.value = row.id;
  showModal.value = true;
};

const handleTest = async (row: any) => {
  try {
    await testApiConfig({ id: row.id });
    message.success('连接测试成功');
  } catch (error: any) {
    message.error(error.message || '测试失败');
  }
};

const handleDelete = (row: any) => {
  dialog.warning({
    title: '确认删除',
    content: `确定要删除"${row.name}"吗？`,
    positiveText: '确定',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await deleteApiConfig({ id: row.id });
        message.success('删除成功');
        loadData();
      } catch (error: any) {
        message.error(error.message || '删除失败');
      }
    },
  });
};

onMounted(() => {
  loadData();
});
</script>

