<template>
  <div class="toogo-admin-subscription">
    <n-card title="订阅记录" :bordered="false" class="proCard">
      <template #header-extra>
        <n-space>
          <n-input v-model:value="searchKey" placeholder="用户名" clearable style="width: 150px">
            <template #prefix>
              <n-icon><SearchOutlined /></n-icon>
            </template>
          </n-input>
          <n-select
            v-model:value="filterPlan"
            :options="planOptions"
            placeholder="套餐"
            clearable
            style="width: 120px"
          />
          <n-select
            v-model:value="filterStatus"
            :options="statusOptions"
            placeholder="状态"
            clearable
            style="width: 100px"
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
      <n-grid :cols="5" :x-gap="16" style="margin-bottom: 16px;">
        <n-gi>
          <n-card size="small">
            <n-statistic label="总订阅数" :value="stats.totalCount" />
          </n-card>
        </n-gi>
        <n-gi>
          <n-card size="small">
            <n-statistic label="活跃订阅" :value="stats.activeCount" />
          </n-card>
        </n-gi>
        <n-gi>
          <n-card size="small">
            <n-statistic label="今日新增" :value="stats.todayCount" />
          </n-card>
        </n-gi>
        <n-gi>
          <n-card size="small">
            <n-statistic label="本月收入" :value="stats.monthAmount">
              <template #suffix>USDT</template>
            </n-statistic>
          </n-card>
        </n-gi>
        <n-gi>
          <n-card size="small">
            <n-statistic label="总收入" :value="stats.totalAmount">
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
import { useMessage, NTag, NProgress } from 'naive-ui';
import { SearchOutlined } from '@vicons/antd';
import { http } from '@/utils/http/axios';

const message = useMessage();

// 状态
const loading = ref(false);
const dataList = ref<any[]>([]);
const searchKey = ref('');
const filterPlan = ref<string | null>(null);
const filterStatus = ref<number | null>(null);
const dateRange = ref<[number, number] | null>(null);

// 统计
const stats = ref({
  totalCount: 0,
  activeCount: 0,
  todayCount: 0,
  monthAmount: 0,
  totalAmount: 0,
});

// 选项
const planOptions = [
  { label: 'A套餐', value: 'A' },
  { label: 'B套餐', value: 'B' },
  { label: 'C套餐', value: 'C' },
  { label: 'D套餐', value: 'D' },
];

const statusOptions = [
  { label: '生效中', value: 1 },
  { label: '已过期', value: 0 },
  { label: '已取消', value: 2 },
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
  { title: '用户', key: 'userName', width: 120 },
  {
    title: '套餐',
    key: 'planName',
    width: 100,
    render: (row: any) => {
      const colors: Record<string, string> = {
        A: '#18a058',
        B: '#2080f0',
        C: '#f0a020',
        D: '#d03050',
      };
      return h(NTag, { color: { color: colors[row.planCode], textColor: '#fff' }, size: 'small' }, () => row.planName);
    },
  },
  {
    title: '周期',
    key: 'period',
    width: 80,
    render: (row: any) => {
      const labels: Record<string, string> = {
        daily: '日',
        monthly: '月',
        quarterly: '季',
        halfYearly: '半年',
        yearly: '年',
      };
      return labels[row.period] || row.period;
    },
  },
  {
    title: '金额',
    key: 'amount',
    width: 100,
    render: (row: any) => h('span', { style: { fontWeight: 'bold' } }, `${row.amount} USDT`),
  },
  { title: '机器人数', key: 'robotLimit', width: 80 },
  {
    title: '状态',
    key: 'status',
    width: 80,
    render: (row: any) => {
      const statuses: Record<number, { label: string; type: any }> = {
        1: { label: '生效中', type: 'success' },
        0: { label: '已过期', type: 'default' },
        2: { label: '已取消', type: 'error' },
      };
      const s = statuses[row.status] || { label: '未知', type: 'default' };
      return h(NTag, { type: s.type, size: 'small' }, () => s.label);
    },
  },
  {
    title: '剩余天数',
    key: 'remainDays',
    width: 120,
    render: (row: any) => {
      if (row.status !== 1) return '-';
      const percent = (row.remainDays / row.totalDays) * 100;
      return h(
        'div',
        { style: { display: 'flex', alignItems: 'center', gap: '8px' } },
        [
          h(NProgress, { type: 'line', percentage: percent, showIndicator: false, status: percent > 20 ? 'success' : 'warning', style: { width: '60px' } }),
          h('span', {}, `${row.remainDays}天`),
        ]
      );
    },
  },
  { title: '开始时间', key: 'startAt', width: 180 },
  { title: '到期时间', key: 'expireAt', width: 180 },
  { title: '邀请人', key: 'inviterName', width: 100 },
  { title: '创建时间', key: 'createdAt', width: 180 },
];

// 加载数据
async function loadData() {
  loading.value = true;
  try {
    const res = await http.request({
      url: '/toogo/admin/subscription/list',
      method: 'get',
      params: {
        page: pagination.value.page,
        pageSize: pagination.value.pageSize,
        keyword: searchKey.value,
        plan: filterPlan.value,
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
      url: '/toogo/admin/subscription/stats',
      method: 'get',
    });
    if (res.code === 0 && res.data) {
      stats.value = res.data;
    }
  } catch (error) {
    console.error('加载统计失败', error);
  }
}

onMounted(() => {
  loadData();
  loadStats();
});
</script>

<style scoped lang="less">
.toogo-admin-subscription {
  // 样式
}
</style>

