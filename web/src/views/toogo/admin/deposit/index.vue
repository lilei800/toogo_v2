<template>
  <div class="toogo-admin-deposit">
    <n-card title="充值管理" :bordered="false" class="proCard">
      <template #header-extra>
        <n-space>
          <n-input v-model:value="searchKey" placeholder="用户名/订单号" clearable style="width: 200px">
            <template #prefix>
              <n-icon><SearchOutlined /></n-icon>
            </template>
          </n-input>
          <n-select
            v-model:value="filterStatus"
            :options="statusOptions"
            placeholder="状态"
            clearable
            style="width: 120px"
          />
          <n-date-picker
            v-model:value="dateRange"
            type="daterange"
            clearable
            style="width: 280px"
          />
          <n-button type="primary" @click="loadData">查询</n-button>
        </n-space>
      </template>

      <!-- 统计卡片 -->
      <n-grid :cols="4" :x-gap="16" style="margin-bottom: 16px;">
        <n-gi>
          <n-card size="small">
            <n-statistic label="今日充值" :value="stats.todayAmount">
              <template #suffix>USDT</template>
            </n-statistic>
          </n-card>
        </n-gi>
        <n-gi>
          <n-card size="small">
            <n-statistic label="今日订单数" :value="stats.todayCount" />
          </n-card>
        </n-gi>
        <n-gi>
          <n-card size="small">
            <n-statistic label="本月充值" :value="stats.monthAmount">
              <template #suffix>USDT</template>
            </n-statistic>
          </n-card>
        </n-gi>
        <n-gi>
          <n-card size="small">
            <n-statistic label="总充值" :value="stats.totalAmount">
              <template #suffix>USDT</template>
            </n-statistic>
          </n-card>
        </n-gi>
      </n-grid>

      <n-data-table
        :columns="columns"
        :data="dataList"
        :loading="loading"
        :pagination="pagination"
        :row-key="(row) => row.id"
      />
    </n-card>
  </div>
</template>

<script lang="ts" setup>
import { ref, onMounted, h } from 'vue';
import { useMessage, NButton, NTag, NSpace, NIcon } from 'naive-ui';
import { SearchOutlined, CheckOutlined, CloseOutlined } from '@vicons/antd';
import { http } from '@/utils/http/axios';

const message = useMessage();

// 状态
const loading = ref(false);
const dataList = ref<any[]>([]);
const searchKey = ref('');
const filterStatus = ref<number | null>(null);
const dateRange = ref<[number, number] | null>(null);

// 统计
const stats = ref({
  todayAmount: 0,
  todayCount: 0,
  monthAmount: 0,
  totalAmount: 0,
});

// 选项
const statusOptions = [
  { label: '待支付', value: 0 },
  { label: '已完成', value: 1 },
  { label: '已取消', value: 2 },
  { label: '已过期', value: 3 },
];

// 分页
const pagination = ref({
  page: 1,
  pageSize: 20,
  showSizePicker: true,
  pageSizes: [10, 20, 50],
  itemCount: 0,
  onChange: (page: number) => {
    pagination.value.page = page;
    loadData();
  },
  onUpdatePageSize: (pageSize: number) => {
    pagination.value.pageSize = pageSize;
    pagination.value.page = 1;
    loadData();
  },
});

// 表格列
const columns = [
  { title: 'ID', key: 'id', width: 60 },
  { title: '订单号', key: 'orderNo', width: 180 },
  { title: '用户', key: 'userName', width: 120 },
  {
    title: '充值金额',
    key: 'amount',
    width: 120,
    render: (row: any) => h('span', { style: { color: '#18a058', fontWeight: 'bold' } }, `${row.amount} USDT`),
  },
  {
    title: '支付方式',
    key: 'payType',
    width: 120,
    render: (row: any) => {
      const types: Record<string, string> = {
        usdt_trc20: 'USDT (TRC20)',
        usdt_erc20: 'USDT (ERC20)',
      };
      return types[row.payType] || row.payType;
    },
  },
  {
    title: '状态',
    key: 'status',
    width: 100,
    render: (row: any) => {
      const statuses: Record<number, { label: string; type: any }> = {
        0: { label: '待支付', type: 'warning' },
        1: { label: '已完成', type: 'success' },
        2: { label: '已取消', type: 'error' },
        3: { label: '已过期', type: 'default' },
      };
      const s = statuses[row.status] || { label: '未知', type: 'default' };
      return h(NTag, { type: s.type, size: 'small' }, () => s.label);
    },
  },
  { title: '交易哈希', key: 'txHash', width: 180, ellipsis: { tooltip: true } },
  { title: '创建时间', key: 'createdAt', width: 180 },
  { title: '完成时间', key: 'paidAt', width: 180 },
  {
    title: '操作',
    key: 'actions',
    width: 150,
    render: (row: any) => {
      if (row.status !== 0) return '-';
      return h(NSpace, {}, () => [
        h(
          NButton,
          { size: 'small', type: 'success', onClick: () => handleConfirm(row) },
          { default: () => '确认到账', icon: () => h(NIcon, null, () => h(CheckOutlined)) }
        ),
        h(
          NButton,
          { size: 'small', type: 'error', onClick: () => handleCancel(row) },
          { default: () => '取消', icon: () => h(NIcon, null, () => h(CloseOutlined)) }
        ),
      ]);
    },
  },
];

// 加载数据
async function loadData() {
  loading.value = true;
  try {
    const res = await http.request({
      url: '/toogo/admin/deposit/list',
      method: 'get',
      params: {
        page: pagination.value.page,
        pageSize: pagination.value.pageSize,
        keyword: searchKey.value,
        status: filterStatus.value,
        startTime: dateRange.value?.[0],
        endTime: dateRange.value?.[1],
      },
    });
    if (res.code === 0) {
      dataList.value = res.data?.list || [];
      pagination.value.itemCount = res.data?.total || 0;
    }
  } catch (error) {
    console.error('加载数据失败', error);
  } finally {
    loading.value = false;
  }
}

// 加载统计
async function loadStats() {
  try {
    const res = await http.request({
      url: '/toogo/admin/deposit/stats',
      method: 'get',
    });
    if (res.code === 0 && res.data) {
      stats.value = res.data;
    }
  } catch (error) {
    console.error('加载统计失败', error);
  }
}

// 确认到账
async function handleConfirm(row: any) {
  try {
    const res = await http.request({
      url: '/toogo/admin/deposit/confirm',
      method: 'post',
      data: { id: row.id },
    });
    if (res.code === 0) {
      message.success('已确认到账');
      loadData();
      loadStats();
    } else {
      message.error(res.message || '操作失败');
    }
  } catch (error: any) {
    message.error(error.message || '操作失败');
  }
}

// 取消订单
async function handleCancel(row: any) {
  try {
    const res = await http.request({
      url: '/toogo/admin/deposit/cancel',
      method: 'post',
      data: { id: row.id },
    });
    if (res.code === 0) {
      message.success('已取消');
      loadData();
    } else {
      message.error(res.message || '操作失败');
    }
  } catch (error: any) {
    message.error(error.message || '操作失败');
  }
}

onMounted(() => {
  loadData();
  loadStats();
});
</script>

<style scoped lang="less">
.toogo-admin-deposit {
  // 样式
}
</style>

