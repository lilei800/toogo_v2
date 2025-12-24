<template>
  <div>
    <n-card :bordered="false" :title="`机器人详情 - ${robotInfo.name}`">
      <n-space vertical :size="16">
        <!-- 统计卡片 -->
        <n-grid :cols="4" :x-gap="12">
          <n-gi>
            <n-statistic label="总盈亏" :value="robotInfo.totalPnl">
              <template #suffix>USDT</template>
            </n-statistic>
          </n-gi>
          <n-gi>
            <n-statistic label="持仓订单" :value="robotInfo.positionCount" />
          </n-gi>
          <n-gi>
            <n-statistic label="运行时长" :value="robotInfo.runtime">
              <template #suffix>秒</template>
            </n-statistic>
          </n-gi>
          <n-gi>
            <n-statistic label="状态">
              <n-tag :type="getStatusType(robotInfo.status)">
                {{ getStatusText(robotInfo.status) }}
              </n-tag>
            </n-statistic>
          </n-gi>
        </n-grid>

        <!-- 持仓订单列表 -->
        <n-card title="持仓订单" :bordered="false">
          <n-data-table
            :columns="positionColumns"
            :data="positionList"
            :pagination="false"
            :scroll-x="1200"
          />
        </n-card>

        <!-- 平仓日志 -->
        <n-card title="平仓日志" :bordered="false">
          <n-data-table
            :columns="closeLogColumns"
            :data="closeLogList"
            :pagination="closePagination"
            :scroll-x="1200"
          />
        </n-card>
      </n-space>
    </n-card>
  </div>
</template>

<script lang="ts" setup>
import { ref, reactive, onMounted, h } from 'vue';
import { useRoute } from 'vue-router';
import {
  NCard,
  NSpace,
  NGrid,
  NGi,
  NStatistic,
  NTag,
  NText,
  NDataTable,
  useMessage,
} from 'naive-ui';
import { getRobotDetail } from '@/api/trading/robot';
import { getPositions, getCloseLogs } from '@/api/trading/order';

const route = useRoute();
const message = useMessage();

const robotInfo = ref<any>({});
const positionList = ref([]);
const closeLogList = ref([]);

const closePagination = reactive({
  page: 1,
  pageSize: 10,
  pageCount: 1,
});

const positionColumns = [
  { title: 'ID', key: 'id', width: 80 },
  { title: '交易对', key: 'symbol', width: 120 },
  { title: '方向', key: 'side', width: 80, render(row: any) {
    return h(NTag, { type: row.side === 'long' ? 'success' : 'error' }, {
      default: () => row.side === 'long' ? '多' : '空'
    });
  }},
  { title: '开仓价', key: 'entryPrice', width: 120 },
  { title: '数量', key: 'amount', width: 120 },
  { title: '杠杆', key: 'leverage', width: 80 },
  { title: '未实现盈亏', key: 'unrealizedPnl', width: 150, render(row: any) {
    return h(NText, { type: row.unrealizedPnl >= 0 ? 'success' : 'error' }, {
      default: () => `${row.unrealizedPnl} USDT`
    });
  }},
  { title: '开仓时间', key: 'createdAt', width: 180 },
];

const closeLogColumns = [
  { title: 'ID', key: 'id', width: 80 },
  { title: '交易对', key: 'symbol', width: 120 },
  { title: '方向', key: 'side', width: 80 },
  { title: '开仓价', key: 'entryPrice', width: 120 },
  { title: '平仓价', key: 'exitPrice', width: 120 },
  { title: '盈亏', key: 'pnl', width: 150, render(row: any) {
    return h(NText, { type: row.pnl >= 0 ? 'success' : 'error' }, {
      default: () => `${row.pnl} USDT`
    });
  }},
  { title: '平仓原因', key: 'closeReason', width: 150 },
  { title: '平仓时间', key: 'closedAt', width: 180 },
];

// 加载机器人详情
const loadRobotDetail = async () => {
  try {
    const res = await getRobotDetail({ id: Number(route.params.id) });
    robotInfo.value = res;
  } catch (error: any) {
    message.error(error.message || '加载失败');
  }
};

// 加载持仓列表
const loadPositions = async () => {
  try {
    const res = await getPositions({ robotId: Number(route.params.id) });
    positionList.value = res.list || [];
  } catch (error: any) {
    message.error(error.message || '加载持仓失败');
  }
};

// 加载平仓日志
const loadCloseLogs = async () => {
  try {
    const res = await getCloseLogs({
      robotId: Number(route.params.id),
      page: closePagination.page,
      pageSize: closePagination.pageSize,
    });
    closeLogList.value = res.list || [];
    closePagination.pageCount = Math.ceil((res.total || 0) / closePagination.pageSize);
  } catch (error: any) {
    message.error(error.message || '加载日志失败');
  }
};

const getStatusType = (status: number) => {
  const types: any = { 1: 'default', 2: 'success', 3: 'warning', 4: 'error' };
  return types[status] || 'default';
};

const getStatusText = (status: number) => {
  const texts: any = { 1: '未启动', 2: '运行中', 3: '已暂停', 4: '已停用' };
  return texts[status] || '未知';
};

onMounted(() => {
  loadRobotDetail();
  loadPositions();
  loadCloseLogs();
});
</script>

