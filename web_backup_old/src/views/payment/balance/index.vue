<template>
  <div>
    <n-space vertical :size="16">
      <!-- 余额卡片 -->
      <n-card :bordered="false">
        <n-grid :cols="4" :x-gap="12">
          <n-gi>
            <n-statistic label="💰 可用余额" :value="balanceInfo.availableBalance || 0">
              <template #suffix>USDT</template>
            </n-statistic>
          </n-gi>
          <n-gi>
            <n-statistic label="🔒 冻结余额" :value="balanceInfo.frozenBalance || 0">
              <template #suffix>USDT</template>
            </n-statistic>
          </n-gi>
          <n-gi>
            <n-statistic label="📊 总余额" :value="(balanceInfo.availableBalance || 0) + (balanceInfo.frozenBalance || 0)">
              <template #suffix>USDT</template>
            </n-statistic>
          </n-gi>
          <n-gi>
            <n-space vertical>
              <n-button type="primary" block @click="$router.push('/payment/deposit')">
                充值
              </n-button>
              <n-button block @click="$router.push('/payment/withdraw')">
                提现
              </n-button>
            </n-space>
          </n-gi>
        </n-grid>
      </n-card>

      <!-- 资金流水 -->
      <n-card title="💳 资金流水" :bordered="false">
        <n-data-table
          :columns="columns"
          :data="dataList"
          :pagination="pagination"
          :loading="loading"
          :scroll-x="1200"
        />
      </n-card>
    </n-space>
  </div>
</template>

<script lang="ts" setup>
import { ref, reactive, onMounted, h } from 'vue';
import { NSpace, NCard, NGrid, NGi, NStatistic, NButton, NDataTable, NTag, NText, useMessage } from 'naive-ui';
import { getBalance, balanceLogList } from '@/api/payment/balance';

const message = useMessage();

const balanceInfo = ref<any>({});
const dataList = ref([]);
const loading = ref(false);
const pagination = reactive({ page: 1, pageSize: 10, pageCount: 1 });

const columns = [
  { title: 'ID', key: 'id', width: 80 },
  { title: '订单号', key: 'orderSn', width: 180 },
  { title: '类型', key: 'type', width: 100, render(row: any) {
    const types: any = { 1: 'success', 2: 'error', 3: 'info', 4: 'warning' };
    const texts: any = { 1: '充值', 2: '提现', 3: '交易', 4: '手续费' };
    return h(NTag, { type: types[row.type] }, { default: () => texts[row.type] });
  }},
  { title: '金额', key: 'amount', width: 150, render(row: any) {
    const isPositive = row.type === 1;
    return h(NText, { type: isPositive ? 'success' : 'error', strong: true }, {
      default: () => `${isPositive ? '+' : '-'}${row.amount} USDT`
    });
  }},
  { title: '余额', key: 'balance', width: 150 },
  { title: '备注', key: 'remark', ellipsis: { tooltip: true } },
  { title: '时间', key: 'createdAt', width: 180 },
];

const loadBalance = async () => {
  try {
    const res = await getBalance();
    balanceInfo.value = res;
  } catch (error: any) {
    message.error(error.message || '加载余额失败');
  }
};

const loadData = async () => {
  loading.value = true;
  try {
    const res = await balanceLogList({ page: pagination.page, pageSize: pagination.pageSize });
    dataList.value = res.list || [];
    pagination.pageCount = Math.ceil((res.total || 0) / pagination.pageSize);
  } catch (error: any) {
    message.error(error.message || '加载流水失败');
  } finally {
    loading.value = false;
  }
};

onMounted(() => {
  loadBalance();
  loadData();
});
</script>

