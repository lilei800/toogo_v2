<template>
  <div class="toogo-admin-commission">
    <n-card title="佣金记录" :bordered="false" class="proCard">
      <template #header-extra>
        <n-space>
          <n-input v-model:value="searchKey" placeholder="用户名" clearable style="width: 150px">
            <template #prefix>
              <n-icon><SearchOutlined /></n-icon>
            </template>
          </n-input>
          <n-select
            v-model:value="filterType"
            :options="typeOptions"
            placeholder="类型"
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
          <n-button @click="handleExport">导出</n-button>
        </n-space>
      </template>

      <!-- 统计卡片 -->
      <n-grid :cols="4" :x-gap="16" style="margin-bottom: 16px">
        <n-gi>
          <n-card size="small">
            <n-statistic label="今日佣金" :value="stats.todayAmount">
              <template #suffix>USDT</template>
            </n-statistic>
          </n-card>
        </n-gi>
        <n-gi>
          <n-card size="small">
            <n-statistic label="本月佣金" :value="stats.monthAmount">
              <template #suffix>USDT</template>
            </n-statistic>
          </n-card>
        </n-gi>
        <n-gi>
          <n-card size="small">
            <n-statistic label="总佣金发放" :value="stats.totalAmount">
              <template #suffix>USDT</template>
            </n-statistic>
          </n-card>
        </n-gi>
        <n-gi>
          <n-card size="small">
            <n-statistic label="待结算" :value="stats.pendingAmount">
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
  import { useMessage, NTag } from 'naive-ui';
  import { SearchOutlined } from '@vicons/antd';
  import { http } from '@/utils/http/axios';

  const message = useMessage();

  // 状态
  const loading = ref(false);
  const dataList = ref<any[]>([]);
  const searchKey = ref('');
  const filterType = ref<string | null>(null);
  const dateRange = ref<[number, number] | null>(null);

  // 统计
  const stats = ref({
    todayAmount: 0,
    monthAmount: 0,
    totalAmount: 0,
    pendingAmount: 0,
  });

  // 选项
  const typeOptions = [
    { label: '邀请奖励', value: 'invite' },
    { label: '订阅佣金', value: 'subscription' },
    { label: '算力佣金', value: 'power' },
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
    { title: '获得用户', key: 'userName', width: 120 },
    { title: '来源用户', key: 'fromUserName', width: 120 },
    {
      title: '类型',
      key: 'type',
      width: 100,
      render: (row: any) => {
        const types: Record<string, { label: string; type: any }> = {
          invite: { label: '邀请奖励', type: 'info' },
          subscription: { label: '订阅佣金', type: 'success' },
          power: { label: '算力佣金', type: 'warning' },
        };
        const t = types[row.type] || { label: row.type, type: 'default' };
        return h(NTag, { type: t.type, size: 'small' }, () => t.label);
      },
    },
    {
      title: '佣金金额',
      key: 'amount',
      width: 120,
      render: (row: any) =>
        h(
          'span',
          { style: { color: '#18a058', fontWeight: 'bold' } },
          `+${row.amount} ${row.unit || 'USDT'}`,
        ),
    },
    {
      title: '来源金额',
      key: 'sourceAmount',
      width: 120,
      render: (row: any) => `${row.sourceAmount} ${row.sourceUnit || 'USDT'}`,
    },
    {
      title: '比例',
      key: 'ratio',
      width: 80,
      render: (row: any) => `${row.ratio}%`,
    },
    {
      title: '代理等级',
      key: 'agentLevel',
      width: 80,
      render: (row: any) =>
        row.agentLevel ? h(NTag, { size: 'small' }, () => `L${row.agentLevel}`) : '-',
    },
    { title: '备注', key: 'remark', ellipsis: { tooltip: true } },
    { title: '创建时间', key: 'createdAt', width: 180 },
  ];

  // 加载数据
  async function loadData() {
    loading.value = true;
    try {
      const res = await http.request({
        url: '/toogo/admin/commission/list',
        method: 'get',
        params: {
          page: pagination.value.page,
          pageSize: pagination.value.pageSize,
          keyword: searchKey.value,
          type: filterType.value,
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
        url: '/toogo/admin/commission/stats',
        method: 'get',
      });
      if (res.code === 0 && res.data) {
        stats.value = res.data;
      }
    } catch (error) {
      console.error('加载统计失败', error);
    }
  }

  // 导出
  async function handleExport() {
    try {
      const res = await http.request({
        url: '/toogo/admin/commission/export',
        method: 'get',
        params: {
          keyword: searchKey.value,
          type: filterType.value,
          startTime: dateRange.value?.[0],
          endTime: dateRange.value?.[1],
        },
        responseType: 'blob',
      });
      const url = window.URL.createObjectURL(new Blob([res]));
      const link = document.createElement('a');
      link.href = url;
      link.setAttribute('download', `佣金记录_${new Date().toISOString().split('T')[0]}.xlsx`);
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      message.success('导出成功');
    } catch (error: any) {
      message.error(error.message || '导出失败');
    }
  }

  onMounted(() => {
    loadData();
    loadStats();
  });
</script>

<style scoped lang="less">
  .toogo-admin-commission {
    // 样式
  }
</style>
