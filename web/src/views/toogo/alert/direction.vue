<template>
  <div class="alert-page">
    <n-card title="方向预警日志" :bordered="false">
      <!-- 搜索表单 -->
      <n-space vertical>
        <n-form inline :model="searchForm" label-placement="left">
          <n-form-item label="机器人ID">
            <n-input-number 
              v-model:value="searchForm.robotId" 
              placeholder="机器人ID" 
              clearable 
              style="width: 150px"
            />
          </n-form-item>
          <n-form-item label="交易对">
            <n-input v-model:value="searchForm.symbol" placeholder="如: BTCUSDT" clearable />
          </n-form-item>
          <n-form-item label="信号类型">
            <n-select
              v-model:value="searchForm.signalType"
              :options="signalTypeOptions"
              placeholder="请选择信号类型"
              clearable
              style="width: 120px"
            />
          </n-form-item>
          <n-form-item>
            <n-button type="primary" @click="handleSearch">查询</n-button>
            <n-button style="margin-left: 10px" @click="handleReset">重置</n-button>
          </n-form-item>
        </n-form>

        <!-- 数据表格 -->
        <n-data-table
          :columns="columns"
          :data="tableData"
          :loading="loading"
          :pagination="pagination"
          :row-key="(row: any) => row.id"
          @update:page="handlePageChange"
          @update:page-size="handlePageSizeChange"
        />
      </n-space>
    </n-card>
  </div>
</template>

<script lang="ts" setup>
import { ref, reactive, onMounted, h } from 'vue';
import { NTag, NSpace } from 'naive-ui';
import { getDirectionLogList } from '@/api/trading/alert';

const loading = ref(false);
const tableData = ref<any[]>([]);

const searchForm = reactive({
  robotId: null as number | null,
  symbol: '',
  signalType: null as string | null,
});

const pagination = reactive({
  page: 1,
  pageSize: 20,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50, 100],
});

const signalTypeOptions = [
  { label: '做多', value: 'long' },
  { label: '做空', value: 'short' },
];

const getSignalTypeTag = (type: string) => {
  const t = type?.toLowerCase();
  if (t === 'long') return { type: 'success' as const, text: '做多' };
  if (t === 'short') return { type: 'error' as const, text: '做空' };
  return { type: 'default' as const, text: type || '-' };
};

const getMarketStateTag = (state: string) => {
  const map: Record<string, { type: 'success' | 'warning' | 'error' | 'info' | 'default'; text: string }> = {
    trend: { type: 'success', text: '趋势' },
    volatile: { type: 'warning', text: '震荡' },
    high_vol: { type: 'error', text: '高波动' },
    low_vol: { type: 'info', text: '低波动' },
  };
  return map[state?.toLowerCase()] || { type: 'default', text: state || '-' };
};

const getRiskPrefTag = (pref: string) => {
  const map: Record<string, { type: 'success' | 'warning' | 'error' | 'default'; text: string }> = {
    aggressive: { type: 'error', text: '激进' },
    balanced: { type: 'warning', text: '平衡' },
    conservative: { type: 'success', text: '保守' },
  };
  return map[pref?.toLowerCase()] || { type: 'default', text: pref || '-' };
};

const columns = [
  { title: 'ID', key: 'id', width: 80 },
  { title: '机器人', key: 'robotId', width: 80 },
  { title: '交易对', key: 'symbol', width: 120 },
  {
    title: '信号类型',
    key: 'signalType',
    width: 100,
    render(row: any) {
      const tag = getSignalTypeTag(row.signalType);
      return h(NTag, { type: tag.type, size: 'small' }, { default: () => tag.text });
    },
  },
  {
    title: '市场/风险',
    key: 'marketRisk',
    width: 140,
    render(row: any) {
      const marketTag = getMarketStateTag(row.marketState);
      const riskTag = getRiskPrefTag(row.riskPreference);
      return h(NSpace, { size: 'small' }, {
        default: () => [
          h(NTag, { type: marketTag.type, size: 'small', bordered: false }, { default: () => marketTag.text }),
          h(NTag, { type: riskTag.type, size: 'small', bordered: false }, { default: () => riskTag.text }),
        ],
      });
    },
  },
  {
    title: '信号强度',
    key: 'signalStrength',
    width: 100,
    render(row: any) {
      const strength = row.signalStrength || 0;
      const color = strength >= 70 ? '#18a058' : strength >= 50 ? '#f0a020' : '#d03050';
      return h('span', { style: { color, fontWeight: 'bold' } }, `${strength.toFixed(1)}%`);
    },
  },
  {
    title: '价格信息',
    key: 'priceInfo',
    width: 200,
    render(row: any) {
      return h('div', { style: { fontSize: '12px', lineHeight: '1.6' } }, [
        h('div', `当前: ${row.currentPrice?.toFixed(2) || '-'}`),
        h('div', { style: { color: '#18a058' } }, `窗口低: ${row.windowMinPrice?.toFixed(2) || '-'}`),
        h('div', { style: { color: '#d03050' } }, `窗口高: ${row.windowMaxPrice?.toFixed(2) || '-'}`),
        row.threshold ? h('div', { style: { color: '#666' } }, `阈值: ${row.threshold.toFixed(2)}`) : null,
      ]);
    },
  },
  {
    title: '执行状态',
    key: 'executed',
    width: 100,
    render(row: any) {
      const executed = row.executed === 1;
      return h(NTag, { 
        type: executed ? 'success' : 'default', 
        size: 'small' 
      }, { 
        default: () => executed ? '已执行' : '未执行' 
      });
    },
  },
  { 
    title: '原因', 
    key: 'reason', 
    ellipsis: { tooltip: true },
    minWidth: 150 
  },
  { title: '时间', key: 'createdAt', width: 180 },
];

const fetchData = async () => {
  loading.value = true;
  try {
    const res = await getDirectionLogList({
      ...searchForm,
      page: pagination.page,
      perPage: pagination.pageSize,
    });
    tableData.value = res.list || [];
    pagination.itemCount = res.total || 0;
  } catch (error) {
    console.error(error);
  } finally {
    loading.value = false;
  }
};

const handleSearch = () => {
  pagination.page = 1;
  fetchData();
};

const handleReset = () => {
  searchForm.robotId = null;
  searchForm.symbol = '';
  searchForm.signalType = null;
  handleSearch();
};

const handlePageChange = (page: number) => {
  pagination.page = page;
  fetchData();
};

const handlePageSizeChange = (pageSize: number) => {
  pagination.pageSize = pageSize;
  pagination.page = 1;
  fetchData();
};

onMounted(() => {
  fetchData();
});
</script>

<style scoped>
.alert-page {
  padding: 16px;
}
</style>

