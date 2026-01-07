<template>
  <div class="toogo-admin-symbol">
    <n-card title="交易对管理" :bordered="false" class="proCard">
      <template #header-extra>
        <n-space>
          <n-button type="primary" @click="handleSyncSymbols">
            <template #icon>
              <n-icon><SyncOutlined /></n-icon>
            </template>
            同步交易对
          </n-button>
          <n-button @click="showAddModal = true">
            <template #icon>
              <n-icon><PlusOutlined /></n-icon>
            </template>
            手动添加
          </n-button>
        </n-space>
      </template>

      <!-- 筛选 -->
      <n-space style="margin-bottom: 16px">
        <n-select
          v-model:value="filterExchange"
          :options="exchangeOptions"
          placeholder="交易所"
          clearable
          style="width: 120px"
        />
        <n-input v-model:value="searchKey" placeholder="搜索交易对" clearable style="width: 200px">
          <template #prefix>
            <n-icon><SearchOutlined /></n-icon>
          </template>
        </n-input>
        <n-select
          v-model:value="filterStatus"
          :options="[
            { label: '已启用', value: 1 },
            { label: '已禁用', value: 0 },
          ]"
          placeholder="状态"
          clearable
          style="width: 100px"
        />
      </n-space>

      <n-data-table
        :columns="columns"
        :data="filteredSymbols"
        :loading="loading"
        :pagination="pagination"
        :row-key="(row) => row.id"
      />
    </n-card>

    <!-- 添加交易对弹窗 -->
    <n-modal
      v-model:show="showAddModal"
      preset="dialog"
      title="添加交易对"
      positive-text="添加"
      negative-text="取消"
      @positive-click="handleAddSubmit"
      style="width: 500px"
    >
      <n-form :model="addForm" label-placement="left" label-width="80">
        <n-form-item label="交易所">
          <n-select
            v-model:value="addForm.exchange"
            :options="exchangeOptions"
            placeholder="选择交易所"
          />
        </n-form-item>
        <n-form-item label="交易对">
          <n-input v-model:value="addForm.symbol" placeholder="例如：BTCUSDT" />
        </n-form-item>
        <n-form-item label="基础币">
          <n-input v-model:value="addForm.baseCoin" placeholder="例如：BTC" />
        </n-form-item>
        <n-form-item label="计价币">
          <n-input v-model:value="addForm.quoteCoin" placeholder="例如：USDT" />
        </n-form-item>
        <n-form-item label="最小数量">
          <n-input-number v-model:value="addForm.minQty" :precision="8" style="width: 100%" />
        </n-form-item>
        <n-form-item label="价格精度">
          <n-input-number
            v-model:value="addForm.pricePrecision"
            :min="0"
            :max="10"
            style="width: 100%"
          />
        </n-form-item>
        <n-form-item label="数量精度">
          <n-input-number
            v-model:value="addForm.qtyPrecision"
            :min="0"
            :max="10"
            style="width: 100%"
          />
        </n-form-item>
      </n-form>
    </n-modal>

    <!-- 编辑交易对弹窗 -->
    <n-modal
      v-model:show="showEditModal"
      preset="dialog"
      title="编辑交易对"
      positive-text="保存"
      negative-text="取消"
      @positive-click="handleEditSubmit"
      style="width: 500px"
    >
      <n-form v-if="editingSymbol" :model="editingSymbol" label-placement="left" label-width="80">
        <n-form-item label="交易所">
          <n-input :value="editingSymbol.exchange" disabled />
        </n-form-item>
        <n-form-item label="交易对">
          <n-input :value="editingSymbol.symbol" disabled />
        </n-form-item>
        <n-form-item label="最小数量">
          <n-input-number v-model:value="editingSymbol.minQty" :precision="8" style="width: 100%" />
        </n-form-item>
        <n-form-item label="最大杠杆">
          <n-input-number
            v-model:value="editingSymbol.maxLeverage"
            :min="1"
            :max="200"
            style="width: 100%"
          />
        </n-form-item>
        <n-form-item label="交易手续费">
          <n-input-number v-model:value="editingSymbol.tradeFee" :precision="4" style="width: 100%">
            <template #suffix>%</template>
          </n-input-number>
        </n-form-item>
        <n-form-item label="状态">
          <n-switch v-model:value="editingSymbol.status" :checked-value="1" :unchecked-value="0" />
        </n-form-item>
        <n-form-item label="热门">
          <n-switch v-model:value="editingSymbol.isHot" />
        </n-form-item>
        <n-form-item label="排序">
          <n-input-number v-model:value="editingSymbol.sort" :min="0" style="width: 100%" />
        </n-form-item>
      </n-form>
    </n-modal>
  </div>
</template>

<script lang="ts" setup>
  import { ref, computed, onMounted, h } from 'vue';
  import {
    useMessage,
    useDialog,
    NButton,
    NTag,
    NSpace,
    NIcon,
    NSwitch,
    NPopconfirm,
  } from 'naive-ui';
  import {
    SyncOutlined,
    PlusOutlined,
    SearchOutlined,
    EditOutlined,
    DeleteOutlined,
  } from '@vicons/antd';
  import { http } from '@/utils/http/axios';

  const message = useMessage();
  const dialog = useDialog();

  // 状态
  const loading = ref(false);
  const syncing = ref(false);
  const symbols = ref<any[]>([]);
  const searchKey = ref('');
  const filterExchange = ref<string | null>(null);
  const filterStatus = ref<number | null>(null);
  const showAddModal = ref(false);
  const showEditModal = ref(false);
  const editingSymbol = ref<any>(null);

  // 添加表单
  const addForm = ref({
    exchange: '',
    symbol: '',
    baseCoin: '',
    quoteCoin: '',
    minQty: 0.001,
    pricePrecision: 2,
    qtyPrecision: 3,
  });

  // 选项
  const exchangeOptions = [
    { label: 'Binance', value: 'binance' },
    { label: 'OKX', value: 'okx' },
    { label: 'Gate.io', value: 'gateio' },
  ];

  // 分页
  const pagination = ref({
    page: 1,
    pageSize: 20,
    showSizePicker: true,
    pageSizes: [10, 20, 50, 100],
  });

  // 过滤后的列表
  const filteredSymbols = computed(() => {
    return symbols.value.filter((s) => {
      const matchKey =
        !searchKey.value || s.symbol.toLowerCase().includes(searchKey.value.toLowerCase());
      const matchExchange = !filterExchange.value || s.exchange === filterExchange.value;
      const matchStatus = filterStatus.value === null || s.status === filterStatus.value;
      return matchKey && matchExchange && matchStatus;
    });
  });

  // 表格列
  const columns = [
    { title: 'ID', key: 'id', width: 60 },
    {
      title: '交易所',
      key: 'exchange',
      width: 100,
      render: (row: any) => {
        const colors: Record<string, string> = {
          binance: '#F0B90B',
          okx: '#121212',
          gateio: '#17E6A1',
        };
        return h(
          NTag,
          { color: { color: colors[row.exchange] || '#666', textColor: '#fff' }, size: 'small' },
          () => row.exchange.toUpperCase(),
        );
      },
    },
    { title: '交易对', key: 'symbol', width: 120 },
    { title: '基础币', key: 'baseCoin', width: 80 },
    { title: '计价币', key: 'quoteCoin', width: 80 },
    { title: '最小数量', key: 'minQty', width: 100 },
    { title: '价格精度', key: 'pricePrecision', width: 80 },
    { title: '数量精度', key: 'qtyPrecision', width: 80 },
    { title: '最大杠杆', key: 'maxLeverage', width: 80 },
    {
      title: '热门',
      key: 'isHot',
      width: 60,
      render: (row: any) =>
        row.isHot ? h(NTag, { type: 'warning', size: 'small' }, () => '热门') : '-',
    },
    {
      title: '状态',
      key: 'status',
      width: 80,
      render: (row: any) =>
        h(NTag, { type: row.status === 1 ? 'success' : 'error', size: 'small' }, () =>
          row.status === 1 ? '启用' : '禁用',
        ),
    },
    {
      title: '操作',
      key: 'actions',
      width: 150,
      render: (row: any) => {
        return h(NSpace, {}, () => [
          h(
            NButton,
            { size: 'small', quaternary: true, type: 'info', onClick: () => handleEdit(row) },
            { default: () => '编辑', icon: () => h(NIcon, null, () => h(EditOutlined)) },
          ),
          h(
            NPopconfirm,
            { onPositiveClick: () => handleDelete(row) },
            {
              trigger: () =>
                h(
                  NButton,
                  { size: 'small', quaternary: true, type: 'error' },
                  { default: () => '删除', icon: () => h(NIcon, null, () => h(DeleteOutlined)) },
                ),
              default: () => '确定删除此交易对吗？',
            },
          ),
        ]);
      },
    },
  ];

  // 加载交易对列表
  async function loadSymbols() {
    loading.value = true;
    try {
      const res = await http.request({
        url: '/toogo/admin/symbol/list',
        method: 'get',
      });
      if (res.code === 0) {
        symbols.value = res.data?.list || [];
      }
    } catch (error) {
      console.error('加载交易对失败', error);
    } finally {
      loading.value = false;
    }
  }

  // 同步交易对
  async function handleSyncSymbols() {
    dialog.warning({
      title: '同步交易对',
      content: '将从各交易所同步最新的交易对信息，可能需要一些时间，确定继续吗？',
      positiveText: '确定',
      negativeText: '取消',
      onPositiveClick: async () => {
        syncing.value = true;
        try {
          const res = await http.request({
            url: '/toogo/admin/symbol/sync',
            method: 'post',
          });
          if (res.code === 0) {
            message.success(`同步完成，共更新 ${res.data?.count || 0} 个交易对`);
            loadSymbols();
          } else {
            message.error(res.message || '同步失败');
          }
        } catch (error: any) {
          message.error(error.message || '同步失败');
        } finally {
          syncing.value = false;
        }
      },
    });
  }

  // 编辑交易对
  function handleEdit(row: any) {
    editingSymbol.value = { ...row };
    showEditModal.value = true;
  }

  // 删除交易对
  async function handleDelete(row: any) {
    try {
      const res = await http.request({
        url: '/toogo/admin/symbol/delete',
        method: 'post',
        data: { id: row.id },
      });
      if (res.code === 0) {
        message.success('删除成功');
        loadSymbols();
      } else {
        message.error(res.message || '删除失败');
      }
    } catch (error: any) {
      message.error(error.message || '删除失败');
    }
  }

  // 提交添加
  async function handleAddSubmit() {
    try {
      const res = await http.request({
        url: '/toogo/admin/symbol/create',
        method: 'post',
        data: addForm.value,
      });
      if (res.code === 0) {
        message.success('添加成功');
        showAddModal.value = false;
        loadSymbols();
      } else {
        message.error(res.message || '添加失败');
      }
    } catch (error: any) {
      message.error(error.message || '添加失败');
    }
    return false;
  }

  // 提交编辑
  async function handleEditSubmit() {
    try {
      const res = await http.request({
        url: '/toogo/admin/symbol/update',
        method: 'post',
        data: editingSymbol.value,
      });
      if (res.code === 0) {
        message.success('保存成功');
        showEditModal.value = false;
        loadSymbols();
      } else {
        message.error(res.message || '保存失败');
      }
    } catch (error: any) {
      message.error(error.message || '保存失败');
    }
    return false;
  }

  onMounted(() => {
    loadSymbols();
  });
</script>

<style scoped lang="less">
  .toogo-admin-symbol {
    // 样式
  }
</style>
