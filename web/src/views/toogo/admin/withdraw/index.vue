<template>
  <div class="withdraw-admin-page">
    <n-card title="提现审核">
      <template #header-extra>
        <n-space>
          <n-select v-model:value="searchParams.status" :options="statusOptions" style="width: 120px" clearable placeholder="状态" @update:value="loadData" />
          <n-date-picker v-model:value="dateRange" type="daterange" clearable @update:value="handleDateChange" />
          <n-button @click="loadData">刷新</n-button>
        </n-space>
      </template>

      <n-data-table
        :columns="columns"
        :data="list"
        :loading="loading"
        :pagination="pagination"
        @update:page="handlePageChange"
      />
    </n-card>

    <!-- 审核弹窗 -->
    <n-modal v-model:show="showAuditModal" title="提现审核" preset="card" style="width: 500px">
      <n-descriptions :column="1" label-placement="left" bordered v-if="currentRow">
        <n-descriptions-item label="订单号">{{ currentRow.orderSn }}</n-descriptions-item>
        <n-descriptions-item label="用户ID">{{ currentRow.userId }}</n-descriptions-item>
        <n-descriptions-item label="提现账户">{{ currentRow.accountType === 'balance' ? '余额' : '佣金' }}</n-descriptions-item>
        <n-descriptions-item label="提现金额">{{ currentRow.amount }} USDT</n-descriptions-item>
        <n-descriptions-item label="手续费">{{ currentRow.fee }} USDT</n-descriptions-item>
        <n-descriptions-item label="实际到账">{{ currentRow.actualAmount }} USDT</n-descriptions-item>
        <n-descriptions-item label="网络">{{ currentRow.network }}</n-descriptions-item>
        <n-descriptions-item label="钱包地址">{{ currentRow.toAddress }}</n-descriptions-item>
        <n-descriptions-item label="申请时间">{{ currentRow.createdAt }}</n-descriptions-item>
      </n-descriptions>

      <n-divider />

      <n-form :model="auditForm" label-placement="left" label-width="80">
        <n-form-item label="审核结果">
          <n-radio-group v-model:value="auditForm.status">
            <n-radio :value="2">通过</n-radio>
            <n-radio :value="4">拒绝</n-radio>
          </n-radio-group>
        </n-form-item>
        <n-form-item label="备注">
          <n-input v-model:value="auditForm.auditNote" type="textarea" :rows="2" placeholder="审核备注（拒绝时必填）" />
        </n-form-item>
      </n-form>

      <template #action>
        <n-space justify="end">
          <n-button @click="showAuditModal = false">取消</n-button>
          <n-button type="primary" @click="handleAudit" :loading="auditLoading">提交审核</n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, h } from 'vue';
import { useMessage } from 'naive-ui';
import { NButton, NTag, NSpace } from 'naive-ui';
import { ToogoFinanceApi } from '@/api/toogo';

const message = useMessage();

const list = ref<any[]>([]);
const loading = ref(false);
const showAuditModal = ref(false);
const auditLoading = ref(false);
const currentRow = ref<any>(null);
const dateRange = ref<any>(null);

const searchParams = ref({
  status: null,
  createdAt: [] as string[],
  page: 1,
  perPage: 10,
});

const pagination = ref({
  page: 1,
  pageSize: 10,
  itemCount: 0,
});

const auditForm = ref({
  status: 2,
  auditNote: '',
});

const statusOptions = [
  { label: '待审核', value: 1 },
  { label: '已通过', value: 2 },
  { label: '已完成', value: 3 },
  { label: '已拒绝', value: 4 },
];

const columns = [
  { title: 'ID', key: 'id', width: 60 },
  { title: '订单号', key: 'orderSn' },
  { title: '用户ID', key: 'userId' },
  {
    title: '账户',
    key: 'accountType',
    render: (row: any) => row.accountType === 'balance' ? '余额' : '佣金',
  },
  { title: '提现金额', key: 'amount', render: (row: any) => row.amount + ' USDT' },
  { title: '手续费', key: 'fee', render: (row: any) => row.fee + ' USDT' },
  { title: '实际到账', key: 'actualAmount', render: (row: any) => row.actualAmount + ' USDT' },
  { title: '网络', key: 'network' },
  {
    title: '状态',
    key: 'status',
    render: (row: any) => {
      const map: any = { 1: { text: '待审核', type: 'warning' }, 2: { text: '已通过', type: 'info' }, 3: { text: '已完成', type: 'success' }, 4: { text: '已拒绝', type: 'error' } };
      const status = map[row.status] || { text: '未知', type: 'default' };
      return h(NTag, { type: status.type, size: 'small' }, { default: () => status.text });
    },
  },
  { title: '申请时间', key: 'createdAt' },
  {
    title: '操作',
    key: 'actions',
    width: 100,
    render: (row: any) => h(NSpace, {}, {
      default: () => [
        row.status === 1
          ? h(NButton, { size: 'small', type: 'primary', onClick: () => openAuditModal(row) }, { default: () => '审核' })
          : h(NButton, { size: 'small', onClick: () => viewDetail(row) }, { default: () => '详情' }),
      ],
    }),
  },
];

const loadData = async () => {
  loading.value = true;
  try {
    const res = await ToogoFinanceApi.withdrawList(searchParams.value);
    list.value = res?.list || [];
    pagination.value.itemCount = res?.totalCount || 0;
  } catch (error) {
    console.error('加载失败:', error);
  } finally {
    loading.value = false;
  }
};

const handlePageChange = (page: number) => {
  searchParams.value.page = page;
  loadData();
};

const handleDateChange = (value: any) => {
  if (value) {
    searchParams.value.createdAt = [
      new Date(value[0]).toISOString().split('T')[0],
      new Date(value[1]).toISOString().split('T')[0],
    ];
  } else {
    searchParams.value.createdAt = [];
  }
  loadData();
};

const openAuditModal = (row: any) => {
  currentRow.value = row;
  auditForm.value = { status: 2, auditNote: '' };
  showAuditModal.value = true;
};

const viewDetail = (row: any) => {
  currentRow.value = row;
  auditForm.value = { status: row.status, auditNote: row.auditNote || '' };
  showAuditModal.value = true;
};

const handleAudit = async () => {
  if (auditForm.value.status === 4 && !auditForm.value.auditNote) {
    message.warning('拒绝时请填写备注');
    return;
  }

  auditLoading.value = true;
  try {
    await ToogoFinanceApi.withdrawAudit({
      id: currentRow.value.id,
      status: auditForm.value.status,
      auditNote: auditForm.value.auditNote,
    });
    message.success('审核成功');
    showAuditModal.value = false;
    loadData();
  } catch (error: any) {
    message.error(error.message || '审核失败');
  } finally {
    auditLoading.value = false;
  }
};

onMounted(() => {
  loadData();
});
</script>

<style scoped lang="less">
.withdraw-admin-page {
  padding: 16px;
}
</style>

