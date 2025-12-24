<template>
  <div class="global-engine-page">
    <!-- 统计概览 -->
    <n-grid :cols="5" :x-gap="16" class="stat-cards">
      <n-gi>
        <n-card class="stat-card">
          <n-statistic label="运行状态">
            <template #prefix>
              <div :class="['status-dot', engineRunning ? 'running' : 'stopped']"></div>
            </template>
            {{ engineRunning ? '运行中' : '已停止' }}
          </n-statistic>
        </n-card>
      </n-gi>
      <n-gi>
        <n-card class="stat-card">
          <n-statistic label="活跃机器人" :value="engineDetail?.robotTaskManager?.activeRobots || 0">
            <template #suffix>
              <n-icon color="#18a058"><RobotOutlined /></n-icon>
            </template>
          </n-statistic>
        </n-card>
      </n-gi>
      <n-gi>
        <n-card class="stat-card">
          <n-statistic label="行情订阅" :value="engineDetail?.marketDataService?.subscriptions || 0">
            <template #suffix>
              <n-icon color="#2080f0"><LineChartOutlined /></n-icon>
            </template>
          </n-statistic>
        </n-card>
      </n-gi>
      <n-gi>
        <n-card class="stat-card">
          <n-statistic label="市场分析" :value="engineDetail?.marketAnalyzer?.analysisCount || 0">
            <template #suffix>
              <n-icon color="#f0a020"><AreaChartOutlined /></n-icon>
            </template>
          </n-statistic>
        </n-card>
      </n-gi>
      <n-gi>
        <n-card class="stat-card">
          <n-statistic label="整体胜率">
            <template #prefix>
              <n-icon color="#13c2c2"><TrophyOutlined /></n-icon>
            </template>
            {{ (engineDetail?.tradeStatistics?.winRate || 0).toFixed(1) }}%
          </n-statistic>
        </n-card>
      </n-gi>
    </n-grid>

    <!-- 行情数据服务 -->
    <n-card title="行情数据服务" :bordered="false" class="proCard" style="margin-top: 16px;">
      <template #header-extra>
        <n-space>
          <n-tag :type="engineDetail?.marketDataService?.running ? 'success' : 'error'" size="small">
            {{ engineDetail?.marketDataService?.running ? '运行中' : '已停止' }}
          </n-tag>
          <n-button @click="refreshData" :loading="loading">
            <template #icon><n-icon><ReloadOutlined /></n-icon></template>
            刷新
          </n-button>
          <n-button type="info" @click="handleRestart" :loading="actionLoading">
            <template #icon><n-icon><SyncOutlined /></n-icon></template>
            重启引擎
          </n-button>
        </n-space>
      </template>

      <n-data-table
        :columns="tickerColumns"
        :data="engineDetail?.marketDataService?.tickerList || []"
        :loading="loading"
        :max-height="280"
        :bordered="false"
        size="small"
        :pagination="false"
      />
    </n-card>

    <!-- 市场分析引擎 -->
    <n-card title="市场分析引擎" :bordered="false" class="proCard" style="margin-top: 16px;">
      <template #header-extra>
        <n-tag :type="engineDetail?.marketAnalyzer?.running ? 'success' : 'error'" size="small">
          {{ engineDetail?.marketAnalyzer?.running ? '运行中' : '已停止' }}
        </n-tag>
      </template>
      <n-data-table
        :columns="analysisColumns"
        :data="engineDetail?.marketAnalyzer?.analysisList || []"
        :max-height="280"
        :bordered="false"
        size="small"
        :pagination="false"
      />
    </n-card>

    <!-- 机器人任务管理器 -->
    <n-card title="机器人任务管理器" :bordered="false" class="proCard" style="margin-top: 16px;">
      <template #header-extra>
        <n-space>
          <n-tag :type="engineDetail?.robotTaskManager?.running ? 'success' : 'error'" size="small">
            {{ engineDetail?.robotTaskManager?.running ? '运行中' : '已停止' }}
          </n-tag>
          <span style="color: #999; font-size: 12px;">
            活跃: {{ engineDetail?.robotTaskManager?.activeRobots || 0 }} 台
          </span>
        </n-space>
      </template>
      <n-data-table
        :columns="robotColumns"
        :data="engineDetail?.robotTaskManager?.robotList || []"
        :max-height="300"
        :bordered="false"
        size="small"
        :pagination="false"
        :row-class-name="getRobotRowClass"
      />
    </n-card>

    <!-- 统计服务 -->
    <n-grid :cols="2" :x-gap="16" style="margin-top: 16px;">
      <!-- 预警日志服务 -->
      <n-gi>
        <n-card title="预警日志服务" :bordered="false" class="proCard">
          <template #header-extra>
            <n-tag :type="engineDetail?.alertLogger?.running ? 'success' : 'error'" size="small">
              {{ engineDetail?.alertLogger?.running ? '运行中' : '已停止' }}
            </n-tag>
          </template>
          <n-grid :cols="3" :x-gap="16">
            <n-gi>
              <n-statistic label="市场状态日志" :value="engineDetail?.alertLogger?.marketStateLogs || 0" />
            </n-gi>
            <n-gi>
              <n-statistic label="风险偏好日志" :value="engineDetail?.alertLogger?.riskPreferenceLogs || 0" />
            </n-gi>
            <n-gi>
              <n-statistic label="方向日志" :value="engineDetail?.alertLogger?.directionLogs || 0" />
            </n-gi>
          </n-grid>
        </n-card>
      </n-gi>

      <!-- 交易统计服务 -->
      <n-gi>
        <n-card title="交易统计服务" :bordered="false" class="proCard">
          <template #header-extra>
            <n-tag :type="engineDetail?.tradeStatistics?.running ? 'success' : 'error'" size="small">
              {{ engineDetail?.tradeStatistics?.running ? '运行中' : '已停止' }}
            </n-tag>
          </template>
          <n-grid :cols="4" :x-gap="16">
            <n-gi>
              <n-statistic label="总交易数" :value="engineDetail?.tradeStatistics?.totalTrades || 0" />
            </n-gi>
            <n-gi>
              <n-statistic label="今日交易" :value="engineDetail?.tradeStatistics?.todayTrades || 0" />
            </n-gi>
            <n-gi>
              <n-statistic label="总盈亏 (USDT)">
                <span :style="{ color: (engineDetail?.tradeStatistics?.totalProfit || 0) >= 0 ? '#18a058' : '#d03050' }">
                  {{ (engineDetail?.tradeStatistics?.totalProfit || 0).toFixed(2) }}
                </span>
              </n-statistic>
            </n-gi>
            <n-gi>
              <n-statistic label="今日盈亏 (USDT)">
                <span :style="{ color: (engineDetail?.tradeStatistics?.todayProfit || 0) >= 0 ? '#18a058' : '#d03050' }">
                  {{ (engineDetail?.tradeStatistics?.todayProfit || 0).toFixed(2) }}
                </span>
              </n-statistic>
            </n-gi>
          </n-grid>
        </n-card>
      </n-gi>
    </n-grid>
  </div>
</template>

<script lang="ts" setup>
import { ref, h, onMounted, onUnmounted } from 'vue';
import { NTag, NProgress, NTooltip, useMessage, useDialog } from 'naive-ui';
import {
  ReloadOutlined,
  SyncOutlined,
  RobotOutlined,
  LineChartOutlined,
  AreaChartOutlined,
  TrophyOutlined,
} from '@vicons/antd';
import {
  getGlobalEngineDetail,
  restartGlobalEngine,
} from '@/api/trading/alert';

const message = useMessage();
const dialog = useDialog();

const loading = ref(false);
const actionLoading = ref(false);
const engineRunning = ref(true);
const engineDetail = ref<any>(null);

// 自动刷新定时器
let refreshTimer: any = null;

// 行情订阅表格列
const tickerColumns = [
  {
    title: '交易对',
    key: 'symbol',
    width: 120,
    render: (row: any) => h('span', { style: { fontWeight: 'bold' } }, row.symbol),
  },
  {
    title: '交易所',
    key: 'platform',
    width: 100,
    render: (row: any) => h(NTag, { size: 'small', type: 'info' }, () => row.platform?.toUpperCase()),
  },
  {
    title: '最新价',
    key: 'lastPrice',
    width: 120,
    render: (row: any) => row.lastPrice?.toFixed(2) || '-',
  },
  {
    title: '24H涨跌',
    key: 'change24h',
    width: 100,
    render: (row: any) => {
      const val = row.change24h || 0;
      const color = val >= 0 ? '#18a058' : '#d03050';
      return h('span', { style: { color } }, `${val >= 0 ? '+' : ''}${val.toFixed(2)}%`);
    },
  },
  {
    title: '数据状态',
    key: 'dataFresh',
    width: 100,
    render: (row: any) =>
      h(NTag, { size: 'small', type: row.dataFresh ? 'success' : 'warning' }, () =>
        row.dataFresh ? '正常' : '过期'
      ),
  },
  {
    title: '引用数',
    key: 'refCount',
    width: 80,
  },
  {
    title: '最后更新',
    key: 'lastUpdate',
    width: 150,
  },
];

// 市场分析表格列
const analysisColumns = [
  {
    title: '交易对',
    key: 'symbol',
    width: 100,
  },
  {
    title: '交易所',
    key: 'platform',
    width: 80,
    render: (row: any) => h(NTag, { size: 'small', type: 'info' }, () => row.platform?.toUpperCase()),
  },
  {
    title: '市场状态',
    key: 'marketState',
    width: 100,
    render: (row: any) => {
      const stateMap: any = {
        trend: { text: '趋势', type: 'success' },
        volatile: { text: '震荡', type: 'warning' },
        high_vol: { text: '高波动', type: 'error' },
        low_vol: { text: '低波动', type: 'info' },
      };
      const state = stateMap[row.marketState] || { text: row.marketState || '-', type: 'default' };
      return h(NTag, { size: 'small', type: state.type }, () => state.text);
    },
  },
  {
    title: '趋势强度',
    key: 'trendStrength',
    width: 120,
    render: (row: any) => {
      const val = row.trendStrength || 0;
      const percent = Math.min(100, Math.abs(val) * 100);
      return h(NProgress, {
        type: 'line',
        percentage: percent,
        status: val > 0 ? 'success' : val < 0 ? 'error' : 'default',
        showIndicator: false,
        height: 6,
      });
    },
  },
];

// 机器人表格列
const robotColumns = [
  { title: 'ID', key: 'robotId', width: 50 },
  { title: '名称', key: 'robotName', width: 100, ellipsis: { tooltip: true } },
  {
    title: '交易对',
    key: 'symbol',
    width: 90,
  },
  {
    title: '行情',
    key: 'connected',
    width: 60,
    render: (row: any) =>
      h(NTag, { size: 'small', type: row.connected ? 'success' : 'error' }, () =>
        row.connected ? '√' : '×'
      ),
  },
  {
    title: 'API',
    key: 'apiConnected',
    width: 60,
    render: (row: any) => {
      if (row.apiConnected) {
        return h(NTag, { size: 'small', type: 'success' }, () => '√');
      }
      return h(
        NTooltip,
        {},
        {
          trigger: () => h(NTag, { size: 'small', type: 'error' }, () => '×'),
          default: () => row.apiError || 'API连接失败',
        }
      );
    },
  },
  {
    title: '账户余额',
    key: 'totalBalance',
    width: 100,
    render: (row: any) => {
      if (!row.apiConnected) return h('span', { style: { color: '#999' } }, '-');
      return `${(row.totalBalance || 0).toFixed(2)}`;
    },
  },
  {
    title: '持仓',
    key: 'hasPosition',
    width: 80,
    render: (row: any) => {
      if (!row.hasPosition) return h('span', { style: { color: '#999' } }, '无');
      const type = row.positionSide === 'LONG' ? 'success' : 'error';
      const text = row.positionSide === 'LONG' ? '做多' : '做空';
      return h(NTag, { size: 'small', type }, () => text);
    },
  },
  {
    title: '持仓数量',
    key: 'positionAmt',
    width: 90,
    render: (row: any) => {
      if (!row.hasPosition) return '-';
      return row.positionAmt?.toFixed(4) || '-';
    },
  },
  {
    title: '未实现盈亏',
    key: 'unrealizedPnl',
    width: 100,
    render: (row: any) => {
      if (!row.hasPosition) return '-';
      const val = row.unrealizedPnl || 0;
      const color = val >= 0 ? '#18a058' : '#d03050';
      return h('span', { style: { color } }, `${val >= 0 ? '+' : ''}${val.toFixed(2)}`);
    },
  },
  {
    title: '信号',
    key: 'directionSignal',
    width: 60,
    render: (row: any) => {
      const sig = row.directionSignal;
      if (sig === 'LONG') return h(NTag, { size: 'small', type: 'success' }, () => '↑多');
      if (sig === 'SHORT') return h(NTag, { size: 'small', type: 'error' }, () => '↓空');
      return h('span', { style: { color: '#999' } }, '-');
    },
  },
  {
    title: '胜算',
    key: 'winProbability',
    width: 60,
    render: (row: any) => `${(row.winProbability || 0).toFixed(0)}%`,
  },
  {
    title: '上次评估',
    key: 'lastRiskEval',
    width: 140,
    render: (row: any) => row.lastRiskEval || '-',
  },
];

// 机器人行样式
function getRobotRowClass(row: any) {
  if (row.hasPosition) {
    return row.unrealizedPnl >= 0 ? 'row-profit' : 'row-loss';
  }
  return '';
}

// 刷新数据
const refreshData = async () => {
  loading.value = true;
  try {
    const res = await getGlobalEngineDetail();
    engineDetail.value = res;
    engineRunning.value = res?.running ?? true;
  } catch (error: any) {
    console.error('获取引擎详情失败:', error);
  } finally {
    loading.value = false;
  }
};

// 重启引擎
const handleRestart = () => {
  dialog.warning({
    title: '确认重启',
    content: '重启全局引擎会短暂中断所有交易，确定要重启吗？',
    positiveText: '确定重启',
    negativeText: '取消',
    onPositiveClick: async () => {
      actionLoading.value = true;
      try {
        const res = await restartGlobalEngine();
        if (res?.success) {
          message.success(res.message || '重启成功');
          setTimeout(refreshData, 1000);
        } else {
          message.error(res?.message || '重启失败');
        }
      } catch (error: any) {
        message.error(error.message || '重启失败');
      } finally {
        actionLoading.value = false;
      }
    },
  });
};

onMounted(() => {
  refreshData();
  // 每5秒自动刷新
  refreshTimer = setInterval(refreshData, 5000);
});

onUnmounted(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer);
  }
});
</script>

<style scoped lang="less">
.global-engine-page {
  .stat-cards {
    .stat-card {
      text-align: center;
    }
  }

  .status-dot {
    display: inline-block;
    width: 10px;
    height: 10px;
    border-radius: 50%;
    margin-right: 6px;
    vertical-align: middle;
  }

  .status-dot.running {
    background-color: #18a058;
    animation: pulse 1.5s ease-in-out infinite;
  }

  .status-dot.stopped {
    background-color: #d03050;
  }

  @keyframes pulse {
    0% {
      box-shadow: 0 0 0 0 rgba(24, 160, 88, 0.4);
    }
    70% {
      box-shadow: 0 0 0 8px rgba(24, 160, 88, 0);
    }
    100% {
      box-shadow: 0 0 0 0 rgba(24, 160, 88, 0);
    }
  }

  :deep(.row-profit) {
    background-color: rgba(24, 160, 88, 0.05);
  }

  :deep(.row-loss) {
    background-color: rgba(208, 48, 80, 0.05);
  }
}
</style>
