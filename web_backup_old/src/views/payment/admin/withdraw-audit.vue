<template>
  <div>
    <n-card :bordered="false" class="proCard">
      <n-space vertical :size="16">
        <!-- 统计卡片 -->
        <n-grid :cols="4" :x-gap="12">
          <n-gi>
            <n-statistic label="待审核" :value="stats.pending || 0" />
          </n-gi>
          <n-gi>
            <n-statistic label="处理中" :value="stats.processing || 0" />
          </n-gi>
          <n-gi>
            <n-statistic label="今日已完成" :value="stats.completed || 0" />
          </n-gi>
          <n-gi>
            <n-statistic label="今日已拒绝" :value="stats.rejected || 0" />
          </n-gi>
        </n-grid>

        <!-- 筛选 -->
        <n-space>
          <n-select v-model:value="filters.status" :options="statusOptions" style="width: 150px;" @update:value="loadData" />
          <n-date-picker v-model:value="filters.dateRange" type="daterange" clearable @update:value="loadData" />
          <n-input v-model:value="filters.orderSn" placeholder="订单号" clearable @update:value="loadData" style="width: 200px;" />
          <n-button @click="loadData">
            <template #icon><n-icon><SearchOutlined /></n-icon></template>
            搜索
          </n-button>
        </n-space>

        <!-- 数据表格 -->
        <n-data-table
          :columns="columns"
          :data="dataList"
          :pagination="pagination"
          :loading="loading"
          :row-key="(row: any) => row.id"
          :scroll-x="1400"
        />
      </n-space>
    </n-card>

    <!-- 审核弹窗 -->
    <n-modal v-model:show="showAuditModal" preset="card" title="审核提现" style="width: 500px;">
      <n-form ref="auditFormRef" :model="auditForm">
        <n-form-item label="审核结果">
          <n-radio-group v-model:value="auditForm.status">
            <n-space>
              <n-radio :value="3">通过</n-radio>
              <n-radio :value="4">拒绝</n-radio>
            </n-space>
          </n-radio-group>
        </n-form-item>
        <n-form-item label="备注">
          <n-input v-model:value="auditForm.remark" type="textarea" :rows="3" placeholder="请输入审核备注" />
        </n-form-item>
      </n-form>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showAuditModal = false">取消</n-button>
          <n-button type="primary" :loading="auditing" @click="handleAuditSubmit">确定</n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<script lang="ts" setup>
import { ref, reactive, onMounted, h } from 'vue';
import { NCard, NSpace, NGrid, NGi, NStatistic, NSelect, NDatePicker, NInput, NButton, NIcon, NDataTable, NTag, NText, NModal, NForm, NFormItem, NRadioGroup, NRadio, useMessage } from 'naive-ui';
import { SearchOutlined } from '@vicons/antd';
import { withdrawList, auditWithdraw } from '@/api/payment/withdraw';

const message = useMessage();

const stats = ref<any>({});
const dataList = ref([]);
const loading = ref(false);
const pagination = reactive({ page: 1, pageSize: 10, pageCount: 1 });
const showAuditModal = ref(false);
const auditing = ref(false);

const filters = reactive({
  status: null,
  dateRange: null,
  orderSn: '',
});

const auditForm = reactive({
  id: null,
  status: 3,
  remark: '',
});

const statusOptions = [
  { label: '全部', value: null },
  { label: '待审核', value: 1 },
  { label: '处理中', value: 2 },
  { label: '已完成', value: 3 },
  { label: '已拒绝', value: 4 },
];

const columns = [
  { title: 'ID', key: 'id', width: 80 },
  { title: '订单号', key: 'orderSn', width: 180 },
  { title: '用户', key: 'userName', width: 120 },
  { title: '金额', key: 'amount', width: 120 },
  { title: '手续费', key: 'fee', width: 100 },
  { title: '实际到账', key: 'actualAmount', width: 120, render(row: any) {
    return h(NText, { type: 'success', strong: true }, { default: () => `${row.actualAmount} USDT` });
  }},
  { title: '提现地址', key: 'toAddress', width: 200, ellipsis: { tooltip: true } },
  { title: '状态', key: 'status', width: 100, render(row: any) {
    const types: any = { 1: 'warning', 2: 'info', 3: 'success', 4: 'error' };
    const texts: any = { 1: '待审核', 2: '处理中', 3: '已完成', 4: '已拒绝' };
    return h(NTag, { type: types[row.status] }, { default: () => texts[row.status] });
  }},
  { title: '创建时间', key: 'createdAt', width: 180 },
  { title: '操作', key: 'actions', width: 120, fixed: 'right', render(row: any) {
    return row.status === 1 ? h(NButton, { size: 'small', type: 'primary', onClick: () => handleAudit(row) }, { default: () => '审核' }) : null;
  }},
];

const loadData = async () => {
  loading.value = true;
  try {
    const res = await withdrawList({
      page: pagination.page,
      pageSize: pagination.pageSize,
      ...filters,
    });
    dataList.value = res.list || [];
    pagination.pageCount = Math.ceil((res.total || 0) / pagination.pageSize);
    stats.value = res.stats || {};
  } catch (error: any) {
    message.error(error.message || '加载失败');
  } finally {
    loading.value = false;
  }
};

const handleAudit = (row: any) => {
  auditForm.id = row.id;
  auditForm.status = 3;
  auditForm.remark = '';
  showAuditModal.value = true;
};

const handleAuditSubmit = async () => {
  auditing.value = true;
  try {
    await auditWithdraw(auditForm);
    message.success('审核成功');
    showAuditModal.value = false;
    loadData();
  } catch (error: any) {
    message.error(error.message || '审核失败');
  } finally {
    auditing.value = false;
  }
};

onMounted(() => {
  loadData();
});
</script>

<style scoped>
.proCard {
  min-height: calc(100vh - 200px);
}
</style>

