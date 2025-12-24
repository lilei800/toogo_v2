<template>
  <div class="toogo-admin-robot">
    <!-- 统计概览 -->
    <n-grid :cols="5" :x-gap="16" class="stat-cards">
      <n-gi>
        <n-card class="stat-card">
          <n-statistic label="机器人总数" :value="stats.total" />
        </n-card>
      </n-gi>
      <n-gi>
        <n-card class="stat-card running">
          <n-statistic label="运行中" :value="stats.running">
            <template #suffix>
              <n-icon color="#18a058"><PlayCircleOutlined /></n-icon>
            </template>
          </n-statistic>
        </n-card>
      </n-gi>
      <n-gi>
        <n-card class="stat-card stopped">
          <n-statistic label="已停止" :value="stats.stopped">
            <template #suffix>
              <n-icon color="#d03050"><PauseCircleOutlined /></n-icon>
            </template>
          </n-statistic>
        </n-card>
      </n-gi>
      <n-gi>
        <n-card class="stat-card profit">
          <n-statistic label="总盈利(USDT)" :value="formatNumber(stats.totalProfit)">
            <template #prefix>
              <span :style="{ color: stats.totalProfit >= 0 ? '#18a058' : '#d03050' }">
                {{ stats.totalProfit >= 0 ? '+' : '' }}
              </span>
            </template>
          </n-statistic>
        </n-card>
      </n-gi>
      <n-gi>
        <n-card class="stat-card power">
          <n-statistic label="今日消耗算力" :value="formatNumber(stats.todayPower)" />
        </n-card>
      </n-gi>
    </n-grid>

    <!-- 机器人列表 -->
    <n-card title="机器人监控" :bordered="false" class="proCard" style="margin-top: 16px;">
      <template #header-extra>
        <n-space>
          <n-input v-model:value="searchKey" placeholder="搜索机器人/用户" clearable style="width: 200px">
            <template #prefix>
              <n-icon><SearchOutlined /></n-icon>
            </template>
          </n-input>
          <n-select
            v-model:value="filterExchange"
            :options="exchangeOptions"
            placeholder="交易所"
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
          <n-button type="primary" @click="loadRobots">刷新</n-button>
        </n-space>
      </template>

      <n-data-table
        :columns="columns"
        :data="filteredRobots"
        :loading="loading"
        :pagination="pagination"
        :row-key="(row) => row.id"
        :row-class-name="getRowClassName"
      />
    </n-card>

    <!-- 机器人详情弹窗 -->
    <n-drawer v-model:show="showDetail" :width="700" placement="right">
      <n-drawer-content :title="`机器人详情 - ${detailRobot?.name}`">
        <template v-if="detailRobot">
          <n-descriptions label-placement="left" bordered :column="2">
            <n-descriptions-item label="ID">{{ detailRobot.id }}</n-descriptions-item>
            <n-descriptions-item label="所属用户">{{ detailRobot.userName }}</n-descriptions-item>
            <n-descriptions-item label="交易所">{{ detailRobot.exchange }}</n-descriptions-item>
            <n-descriptions-item label="交易对">{{ detailRobot.symbol }}</n-descriptions-item>
            <n-descriptions-item label="状态">
              <n-tag :type="detailRobot.status === 'running' ? 'success' : 'warning'">
                {{ detailRobot.status === 'running' ? '运行中' : '已停止' }}
              </n-tag>
            </n-descriptions-item>
            <n-descriptions-item label="创建时间">{{ detailRobot.createdAt }}</n-descriptions-item>
          </n-descriptions>

          <n-divider>交易参数</n-divider>
          <n-descriptions label-placement="left" bordered :column="2">
            <n-descriptions-item label="杠杆倍数">{{ detailRobot.leverage }}x</n-descriptions-item>
            <n-descriptions-item label="保证金模式">{{ detailRobot.marginMode }}</n-descriptions-item>
            <n-descriptions-item label="最大盈利">{{ detailRobot.maxProfit }} USDT</n-descriptions-item>
            <n-descriptions-item label="最大亏损">{{ detailRobot.maxLoss }} USDT</n-descriptions-item>
            <n-descriptions-item label="止损比例">{{ detailRobot.stopLoss }}%</n-descriptions-item>
            <n-descriptions-item label="止盈回撤">{{ detailRobot.trailingStop }}%</n-descriptions-item>
          </n-descriptions>

          <n-divider>运行数据</n-divider>
          <n-descriptions label-placement="left" bordered :column="2">
            <n-descriptions-item label="累计盈亏">
              <span :style="{ color: detailRobot.totalPnl >= 0 ? '#18a058' : '#d03050' }">
                {{ detailRobot.totalPnl >= 0 ? '+' : '' }}{{ formatNumber(detailRobot.totalPnl) }} USDT
              </span>
            </n-descriptions-item>
            <n-descriptions-item label="消耗算力">{{ formatNumber(detailRobot.powerUsed) }}</n-descriptions-item>
            <n-descriptions-item label="订单数">{{ detailRobot.orderCount }}</n-descriptions-item>
            <n-descriptions-item label="胜率">{{ detailRobot.winRate }}%</n-descriptions-item>
          </n-descriptions>

          <n-divider>最近订单</n-divider>
          <n-data-table
            :columns="orderColumns"
            :data="detailRobot.recentOrders || []"
            size="small"
            :pagination="{ pageSize: 5 }"
          />
        </template>

        <template #footer>
          <n-space>
            <n-popconfirm @positive-click="handleForceStop(detailRobot)" v-if="detailRobot?.status === 'running'">
              <template #trigger>
                <n-button type="error">强制停止</n-button>
              </template>
              确定强制停止该机器人吗？
            </n-popconfirm>
          </n-space>
        </template>
      </n-drawer-content>
    </n-drawer>
  </div>
</template>

<script lang="ts" setup>
import { ref, computed, onMounted, onUnmounted, h } from 'vue';
import { useMessage, NButton, NTag, NSpace, NIcon, NProgress } from 'naive-ui';
import {
  PlayCircleOutlined,
  PauseCircleOutlined,
  SearchOutlined,
  EyeOutlined,
  StopOutlined,
} from '@vicons/antd';
import { http } from '@/utils/http/axios';

const message = useMessage();

// 统计数据
const stats = ref({
  total: 0,
  running: 0,
  stopped: 0,
  totalProfit: 0,
  todayPower: 0,
});

// 状态
const loading = ref(false);
const robots = ref<any[]>([]);
const searchKey = ref('');
const filterExchange = ref<string | null>(null);
const filterStatus = ref<string | null>(null);
const showDetail = ref(false);
const detailRobot = ref<any>(null);
let refreshTimer: any = null;

// 选项
const exchangeOptions = [
  { label: 'Binance', value: 'binance' },
  { label: 'OKX', value: 'okx' },
  { label: 'Bitget', value: 'bitget' },
  { label: 'Gate.io', value: 'gateio' },
];

const statusOptions = [
  { label: '运行中', value: 'running' },
  { label: '已停止', value: 'stopped' },
];

// 分页
const pagination = ref({
  page: 1,
  pageSize: 20,
  showSizePicker: true,
  pageSizes: [10, 20, 50, 100],
});

// 过滤后的机器人列表
const filteredRobots = computed(() => {
  return robots.value.filter((robot) => {
    const matchKey =
      !searchKey.value ||
      robot.name.includes(searchKey.value) ||
      robot.userName.includes(searchKey.value);
    const matchExchange = !filterExchange.value || robot.exchange === filterExchange.value;
    const matchStatus = !filterStatus.value || robot.status === filterStatus.value;
    return matchKey && matchExchange && matchStatus;
  });
});

// 格式化数字
function formatNumber(num: number) {
  return (num || 0).toFixed(2);
}

// 行样式
function getRowClassName(row: any) {
  if (row.status === 'running') {
    return row.realTimePnl >= 0 ? 'row-profit' : 'row-loss';
  }
  return '';
}

// 表格列
const columns = [
  { title: 'ID', key: 'id', width: 60 },
  { title: '机器人', key: 'name', width: 120 },
  { title: '用户', key: 'userName', width: 100 },
  {
    title: '交易所',
    key: 'exchange',
    width: 100,
    render: (row: any) => {
      const colors: Record<string, string> = {
        binance: '#F0B90B',
        okx: '#121212',
        bitget: '#00F0FF',
        gateio: '#17E6A1',
      };
      return h(NTag, { color: { color: colors[row.exchange], textColor: '#fff' }, size: 'small' }, () =>
        row.exchange.toUpperCase()
      );
    },
  },
  { title: '交易对', key: 'symbol', width: 100 },
  { title: '杠杆', key: 'leverage', width: 60, render: (row: any) => `${row.leverage}x` },
  {
    title: '状态',
    key: 'status',
    width: 80,
    render: (row: any) => {
      const isRunning = row.status === 'running';
      return h(NTag, { type: isRunning ? 'success' : 'warning', size: 'small' }, () =>
        isRunning ? '运行中' : '已停止'
      );
    },
  },
  {
    title: '实时盈亏',
    key: 'realTimePnl',
    width: 120,
    render: (row: any) => {
      if (row.status !== 'running') return '-';
      const pnl = row.realTimePnl || 0;
      return h('span', { style: { color: pnl >= 0 ? '#18a058' : '#d03050', fontWeight: 'bold' } }, [
        pnl >= 0 ? '+' : '',
        formatNumber(pnl),
        ' USDT',
      ]);
    },
  },
  {
    title: '累计盈亏',
    key: 'totalPnl',
    width: 120,
    render: (row: any) => {
      const pnl = row.totalPnl || 0;
      return h('span', { style: { color: pnl >= 0 ? '#18a058' : '#d03050' } }, [
        pnl >= 0 ? '+' : '',
        formatNumber(pnl),
        ' USDT',
      ]);
    },
  },
  {
    title: '算力消耗',
    key: 'powerUsed',
    width: 100,
    render: (row: any) => formatNumber(row.powerUsed || 0),
  },
  {
    title: '运行时长',
    key: 'runTime',
    width: 100,
  },
  {
    title: '操作',
    key: 'actions',
    width: 120,
    render: (row: any) => {
      return h(NSpace, {}, () => [
        h(
          NButton,
          { size: 'small', quaternary: true, type: 'info', onClick: () => handleViewDetail(row) },
          { default: () => '详情', icon: () => h(NIcon, null, () => h(EyeOutlined)) }
        ),
        row.status === 'running' &&
          h(
            NButton,
            { size: 'small', quaternary: true, type: 'error', onClick: () => handleForceStop(row) },
            { default: () => '停止', icon: () => h(NIcon, null, () => h(StopOutlined)) }
          ),
      ]);
    },
  },
];

// 订单表格列
const orderColumns = [
  { title: '时间', key: 'time', width: 150 },
  { title: '方向', key: 'side', width: 60, render: (row: any) => h(NTag, { type: row.side === 'long' ? 'success' : 'error', size: 'small' }, () => row.side === 'long' ? '做多' : '做空') },
  { title: '价格', key: 'price', width: 100 },
  { title: '数量', key: 'amount', width: 80 },
  { title: '盈亏', key: 'pnl', width: 100, render: (row: any) => h('span', { style: { color: row.pnl >= 0 ? '#18a058' : '#d03050' } }, `${row.pnl >= 0 ? '+' : ''}${formatNumber(row.pnl)}`) },
];

// 加载统计数据
async function loadStats() {
  try {
    const res = await http.request({
      url: '/toogo/admin/robot/stats',
      method: 'get',
    });
    if (res.code === 0 && res.data) {
      stats.value = res.data;
    }
  } catch (error) {
    console.error('加载统计失败', error);
  }
}

// 加载机器人列表
async function loadRobots() {
  loading.value = true;
  try {
    const res = await http.request({
      url: '/toogo/admin/robot/list',
      method: 'get',
    });
    if (res.code === 0) {
      robots.value = res.data?.list || [];
    }
  } catch (error) {
    console.error('加载机器人失败', error);
  } finally {
    loading.value = false;
  }
}

// 查看详情
async function handleViewDetail(robot: any) {
  try {
    const res = await http.request({
      url: '/toogo/admin/robot/detail',
      method: 'get',
      params: { id: robot.id },
    });
    if (res.code === 0) {
      detailRobot.value = res.data;
      showDetail.value = true;
    }
  } catch (error) {
    console.error('加载详情失败', error);
  }
}

// 强制停止
async function handleForceStop(robot: any) {
  try {
    const res = await http.request({
      url: '/toogo/admin/robot/stop',
      method: 'post',
      data: { id: robot.id },
    });
    if (res.code === 0) {
      message.success('机器人已停止');
      loadRobots();
      if (showDetail.value) {
        handleViewDetail(robot);
      }
    } else {
      message.error(res.message || '停止失败');
    }
  } catch (error: any) {
    message.error(error.message || '停止失败');
  }
}

onMounted(() => {
  loadStats();
  loadRobots();
  // 每30秒刷新一次
  refreshTimer = setInterval(() => {
    loadStats();
    loadRobots();
  }, 30000);
});

onUnmounted(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer);
  }
});
</script>

<style scoped lang="less">
.toogo-admin-robot {
  .stat-cards {
    .stat-card {
      text-align: center;
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

