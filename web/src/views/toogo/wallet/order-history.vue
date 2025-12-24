<template>
  <div class="order-history-page">
    <!-- 页面标题和选项卡 -->
    <n-card :bordered="false" class="page-header">
      <n-tabs v-model:value="viewMode" type="segment" animated size="large" class="main-tabs">
        <n-tab-pane name="session" tab="运行区间">
          <template #tab>
            <div class="tab-item">
              <n-icon :component="ClockCircleOutlined" :size="18" />
              <span>运行区间</span>
            </div>
          </template>
        </n-tab-pane>
        <n-tab-pane name="order" tab="成交流水">
          <template #tab>
            <div class="tab-item">
              <n-icon :component="UnorderedListOutlined" :size="18" />
              <span>成交流水</span>
            </div>
          </template>
        </n-tab-pane>
      </n-tabs>
      <n-text depth="3" class="header-desc">
        {{
          viewMode === 'order'
            ? '每条记录对应交易所一笔成交；手续费/已实现盈亏以交易所成交为准'
            : '机器人运行时间区间；盈亏/手续费从交易所成交数据实时统计'
        }}
      </n-text>
    </n-card>

    <!-- 统计卡片 -->
    <n-grid
      :cols="6"
      :x-gap="12"
      :y-gap="12"
      responsive="screen"
      item-responsive
      class="stats-grid"
    >
      <n-gi span="6 m:1">
        <n-card class="stat-card" :bordered="false">
          <div class="stat-content">
            <div class="stat-icon" style="background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);">
              <n-icon :component="BarChartOutlined" :size="22" color="#fff" />
            </div>
            <div class="stat-info">
              <div class="stat-label">{{ viewMode === 'session' ? '区间数' : '成交笔数' }}</div>
              <div class="stat-value">{{ summary.totalSessions }} <span class="stat-unit">{{ viewMode === 'session' ? '段' : '笔' }}</span></div>
            </div>
          </div>
        </n-card>
      </n-gi>
      <n-gi span="6 m:1">
        <n-card class="stat-card profit-card" :bordered="false">
          <div class="stat-content">
            <div class="stat-icon" style="background: linear-gradient(135deg, #11998e 0%, #38ef7d 100%);">
              <n-icon :component="RiseOutlined" :size="22" color="#fff" />
            </div>
            <div class="stat-info">
              <div class="stat-label">总盈利</div>
              <div class="stat-value profit">
                +{{ summary.totalProfit.toFixed(6) }}
                <span class="stat-unit">USDT</span>
              </div>
            </div>
          </div>
        </n-card>
      </n-gi>
      <n-gi span="6 m:1">
        <n-card class="stat-card loss-card" :bordered="false">
          <div class="stat-content">
            <div class="stat-icon" style="background: linear-gradient(135deg, #eb3349 0%, #f45c43 100%);">
              <n-icon :component="FallOutlined" :size="22" color="#fff" />
            </div>
            <div class="stat-info">
              <div class="stat-label">总亏损</div>
              <div class="stat-value loss">
                {{ summary.totalLoss.toFixed(6) }}
                <span class="stat-unit">USDT</span>
              </div>
            </div>
          </div>
        </n-card>
      </n-gi>
      <n-gi span="6 m:1">
        <n-card class="stat-card" :bordered="false">
          <div class="stat-content">
            <div class="stat-icon" style="background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);">
              <n-icon :component="TransactionOutlined" :size="22" color="#fff" />
            </div>
            <div class="stat-info">
              <div class="stat-label">总手续费</div>
              <div class="stat-value fee">
                {{ summary.totalFee.toFixed(6) }}
                <span class="stat-unit">USDT</span>
              </div>
            </div>
          </div>
        </n-card>
      </n-gi>
      <n-gi span="6 m:1">
        <n-card class="stat-card" :bordered="false">
          <div class="stat-content">
            <div class="stat-icon" :style="{ background: summary.totalPnl >= 0 ? 'linear-gradient(135deg, #11998e 0%, #38ef7d 100%)' : 'linear-gradient(135deg, #eb3349 0%, #f45c43 100%)' }">
              <n-icon :component="DollarOutlined" :size="22" color="#fff" />
            </div>
            <div class="stat-info">
              <div class="stat-label">总盈亏</div>
              <div class="stat-value" :class="summary.totalPnl >= 0 ? 'profit' : 'loss'">
                {{ summary.totalPnl >= 0 ? '+' : '' }}{{ summary.totalPnl.toFixed(6) }}
                <span class="stat-unit">USDT</span>
              </div>
            </div>
          </div>
        </n-card>
      </n-gi>
      <n-gi span="6 m:1">
        <n-card class="stat-card highlight-card" :bordered="false">
          <div class="stat-content">
            <div class="stat-icon" :style="{ background: summary.totalNetPnl >= 0 ? 'linear-gradient(135deg, #00b09b 0%, #96c93d 100%)' : 'linear-gradient(135deg, #c31432 0%, #240b36 100%)' }">
              <n-icon :component="WalletOutlined" :size="22" color="#fff" />
            </div>
            <div class="stat-info">
              <div class="stat-label">净盈亏(扣手续费)</div>
              <div class="stat-value" :class="summary.totalNetPnl >= 0 ? 'profit' : 'loss'">
                {{ summary.totalNetPnl >= 0 ? '+' : '' }}{{ summary.totalNetPnl.toFixed(6) }}
                <span class="stat-unit">USDT</span>
              </div>
            </div>
          </div>
        </n-card>
      </n-gi>
    </n-grid>

    <!-- 筛选条件 -->
    <n-card :bordered="false" class="filter-card" size="small">
      <n-space align="center" :wrap="true">
        <n-input-group>
          <n-input-group-label style="width: 70px">机器人</n-input-group-label>
          <n-select
            v-model:value="filterForm.robotId"
            placeholder="全部"
            clearable
            filterable
            style="width: 180px"
            :options="robotOptions"
          />
        </n-input-group>
        
        <n-input-group>
          <n-input-group-label style="width: 70px">交易所</n-input-group-label>
          <n-select
            v-model:value="filterForm.exchange"
            placeholder="全部"
            clearable
            style="width: 120px"
            :options="exchangeOptions"
          />
        </n-input-group>

        <n-input-group>
          <n-input-group-label style="width: 70px">交易对</n-input-group-label>
          <n-input
            v-model:value="filterForm.symbol"
            placeholder="BTCUSDT"
            clearable
            style="width: 120px"
            @keyup.enter="handleSearch"
          />
        </n-input-group>

        <n-input-group>
          <n-input-group-label style="width: 50px">{{ viewMode === 'session' ? '状态' : '方向' }}</n-input-group-label>
          <n-select
            v-if="viewMode === 'session'"
            v-model:value="filterForm.isRunning"
            placeholder="全部"
            clearable
            style="width: 100px"
            :options="runningOptions"
          />
          <n-select
            v-else
            v-model:value="filterForm.side"
            placeholder="全部"
            clearable
            style="width: 100px"
            :options="sideOptions"
          />
        </n-input-group>

        <n-input-group>
          <n-input-group-label style="width: 80px">时间范围</n-input-group-label>
          <n-date-picker
            v-model:value="dateRange"
            type="datetimerange"
            clearable
            style="width: 320px"
            :shortcuts="dateShortcuts"
            @update:value="handleDateRangeChange"
          />
        </n-input-group>

        <n-divider vertical />
        
        <n-button type="primary" @click="handleSearch">
          <template #icon><n-icon :component="SearchOutlined" /></template>
          查询
        </n-button>
        <n-button @click="handleReset">重置</n-button>
        <n-button type="primary" @click="handleRefresh" :loading="loading" quaternary>
          <template #icon><n-icon :component="ReloadOutlined" /></template>
          刷新
        </n-button>
      </n-space>
    </n-card>

    <!-- 数据列表 -->
    <n-card :bordered="false" class="table-card">
      <template #header>
        <n-space align="center">
          <n-text strong>{{ viewMode === 'order' ? '成交流水' : '运行记录' }}</n-text>
          <n-tag size="small" :bordered="false">共 {{ totalCount }} 条</n-tag>
          <n-tag v-if="viewMode === 'session'" size="small" type="info" :bordered="false">
            总运行 {{ summary.totalRuntimeText || '--' }}
          </n-tag>
        </n-space>
      </template>
      <n-data-table
        :columns="tableColumns"
        :data="orderList"
        :loading="loading"
        :pagination="pagination"
        :row-key="(row) => row.id"
        :scroll-x="1400"
        remote
        @update:page="handlePageChange"
        @update:page-size="handlePageSizeChange"
        striped
        size="small"
        :row-class-name="rowClassName"
      />
    </n-card>

    <!-- 详情抽屉 -->
    <n-drawer v-model:show="showDetail" :width="520" placement="right">
      <n-drawer-content :title="viewMode === 'order' ? '成交订单 - 详情' : '运行区间 - 详情'" closable>
        <n-descriptions v-if="viewMode === 'session'" :column="1" bordered size="small" label-placement="left">
          <n-descriptions-item label="交易所">
            <n-tag size="small" :bordered="false">{{ currentRow?.exchange || '--' }}</n-tag>
          </n-descriptions-item>
          <n-descriptions-item label="机器人">{{ currentRow?.robotName || (currentRow?.robotId ? `#${currentRow?.robotId}` : '--') }}</n-descriptions-item>
          <n-descriptions-item label="交易对">
            <n-text strong>{{ currentRow?.symbol || '--' }}</n-text>
          </n-descriptions-item>
          <n-descriptions-item label="启动时间">{{ currentRow?.startTime || '--' }}</n-descriptions-item>
          <n-descriptions-item label="结束时间">{{ currentRow?.endTime || (currentRow?.isRunning ? '运行中' : '--') }}</n-descriptions-item>
          <n-descriptions-item label="结束原因">{{ currentRow?.endReasonText || '--' }}</n-descriptions-item>
          <n-descriptions-item label="运行时长">
            <n-text type="info">{{ currentRow?.runtimeText || '--' }}</n-text>
          </n-descriptions-item>
          <n-descriptions-item label="成交笔数">{{ currentRow?.tradeCount ?? 0 }}</n-descriptions-item>
          <n-descriptions-item label="区间盈亏(USDT)">
            <n-text :type="(currentRow?.totalPnl ?? 0) >= 0 ? 'success' : 'error'">
              {{ formatPnl(currentRow?.totalPnl) }}
            </n-text>
          </n-descriptions-item>
          <n-descriptions-item label="区间手续费(USDT)">
            <n-text type="warning">{{ formatFixed(currentRow?.totalFee, FEE_DIGITS) }}</n-text>
          </n-descriptions-item>
          <n-descriptions-item label="净盈亏(扣手续费)">
            <n-text :type="(currentRow?.netPnl ?? 0) >= 0 ? 'success' : 'error'" strong>
              {{ formatPnl(currentRow?.netPnl) }}
            </n-text>
          </n-descriptions-item>
          <n-descriptions-item label="最后同步">{{ currentRow?.syncedAt || '--' }}</n-descriptions-item>
        </n-descriptions>

        <n-descriptions v-else :column="1" bordered size="small" label-placement="left">
          <n-descriptions-item label="交易所">
            <n-tag size="small" :bordered="false">{{ currentRow?.exchange || '--' }}</n-tag>
          </n-descriptions-item>
          <n-descriptions-item label="机器人">{{ currentRow?.robotName || (currentRow?.robotId ? `#${currentRow?.robotId}` : '--') }}</n-descriptions-item>
          <n-descriptions-item label="交易对">
            <n-text strong>{{ currentRow?.symbol || '--' }}</n-text>
          </n-descriptions-item>
          <n-descriptions-item label="方向">{{ currentRow?.side || '--' }}</n-descriptions-item>
          <n-descriptions-item label="成交价">{{ formatFixed(currentRow?.price, 6) }}</n-descriptions-item>
          <n-descriptions-item label="成交量">{{ formatFixed(currentRow?.qty, 8) }}</n-descriptions-item>
          <n-descriptions-item label="手续费">{{ `${formatFixed(currentRow?.fee, FEE_DIGITS)} ${currentRow?.feeCoin || ''}`.trim() }}</n-descriptions-item>
          <n-descriptions-item label="已实现盈亏">
            <n-text :type="(currentRow?.realizedPnl ?? 0) >= 0 ? 'success' : 'error'" strong>
              {{ formatPnl8(currentRow?.realizedPnl) }}
            </n-text>
          </n-descriptions-item>
          <n-descriptions-item label="成交时间">{{ currentRow?.time || '--' }}</n-descriptions-item>
          <n-descriptions-item label="订单ID">{{ currentRow?.orderId || '--' }}</n-descriptions-item>
          <n-descriptions-item label="成交ID">{{ currentRow?.tradeId || '--' }}</n-descriptions-item>
        </n-descriptions>
      </n-drawer-content>
    </n-drawer>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, h, computed, watch } from 'vue';
import { NTag, NButton, NText, NIcon, NDrawer, NDrawerContent, NDescriptions, NDescriptionsItem, NSpace, useMessage } from 'naive-ui';
import { ReloadOutlined, SearchOutlined, BarChartOutlined, ClockCircleOutlined, RiseOutlined, FallOutlined, DollarOutlined, TransactionOutlined, WalletOutlined, UnorderedListOutlined } from '@vicons/antd';
import { ToogoWalletApi, ToogoRobotApi } from '@/api/toogo';
import type { DataTableColumns } from 'naive-ui';
import { format, subDays, startOfDay, endOfDay } from 'date-fns';

const message = useMessage();

// 数据
const loading = ref(false);
const orderList = ref<any[]>([]);
const totalCount = ref(0);
const viewMode = ref<'order' | 'session'>('session'); // 默认显示运行区间
const summary = reactive({
  totalSessions: 0,
  totalRuntime: 0,
  totalRuntimeText: '',
  totalPnl: 0,
  totalProfit: 0,  // 盈利（正数部分）
  totalLoss: 0,    // 亏损（负数部分）
  totalFee: 0,
  totalNetPnl: 0,
  totalTrades: 0,
});

// 筛选表单
const filterForm = reactive({
  exchange: '',
  robotId: null as number | null,
  symbol: '',
  isRunning: null as number | null,
  side: '', // BUY/SELL
  startTime: '',
  endTime: '',
});

const dateRange = ref<[number, number] | null>(null);

// 日期快捷选项
const dateShortcuts = {
  '今天': () => [startOfDay(new Date()).getTime(), endOfDay(new Date()).getTime()] as [number, number],
  '最近7天': () => [startOfDay(subDays(new Date(), 6)).getTime(), endOfDay(new Date()).getTime()] as [number, number],
  '最近30天': () => [startOfDay(subDays(new Date(), 29)).getTime(), endOfDay(new Date()).getTime()] as [number, number],
};

// 分页
const pagination = reactive({
  page: 1,
  pageSize: 20,
  pageCount: 0,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50, 100],
  showQuickJumper: true,
  prefix: ({ itemCount }: { itemCount: number }) => `共 ${itemCount} 条`,
});

// 选项
const exchangeOptions = [
  { label: 'Binance', value: 'binance' },
  { label: 'Bitget', value: 'bitget' },
  { label: 'OKX', value: 'okx' },
  { label: 'Gate', value: 'gate' },
];

const runningOptions = [
  { label: '运行中', value: 1 },
  { label: '已结束', value: 2 },
];

const sideOptions = [
  { label: 'BUY(买)', value: 'BUY' },
  { label: 'SELL(卖)', value: 'SELL' },
];

const robotOptions = ref<{ label: string; value: number }[]>([]);
const loadRobotOptions = async () => {
  try {
    const pageSize = 100;
    const maxPages = 10;
    const all: any[] = [];
    const seen = new Set<number>();

    for (let page = 1; page <= maxPages; page++) {
      const res: any = await ToogoRobotApi.list({ page, pageSize });
      const list: any[] = res?.list || [];
      for (const r of list) {
        const id = Number(r?.id);
        if (!id || Number.isNaN(id) || seen.has(id)) continue;
        seen.add(id);
        all.push(r);
      }
      if (list.length < pageSize) break;
    }

    robotOptions.value = all.map((r) => ({
      label: `${r.robotName || '未命名'} (#${r.id})`,
      value: Number(r.id),
    }));
  } catch (e) {
    // 机器人列表加载失败不影响页面使用
  }
};

// 详情抽屉
const showDetail = ref(false);
const currentRow = ref<any | null>(null);
const openDetail = (row: any) => {
  currentRow.value = row;
  showDetail.value = true;
};

const FEE_DIGITS = 7;

// 展示工具
const formatFixed = (val: any, digits: number) => {
  if (val === undefined || val === null || val === '') return '--';
  const n = Number(val);
  if (Number.isNaN(n)) return '--';
  return n.toFixed(digits);
};

const formatPnl = (val: any) => {
  if (val === undefined || val === null || val === '') return '--';
  const n = Number(val);
  if (Number.isNaN(n)) return '--';
  return (n >= 0 ? '+' : '') + n.toFixed(2);
};

const formatPnl8 = (val: any) => {
  if (val === undefined || val === null || val === '') return '--';
  const n = Number(val);
  if (Number.isNaN(n)) return '--';
  return (n >= 0 ? '+' : '') + n.toFixed(8);
};

// 行样式
const rowClassName = (row: any) => {
  if (row.isRunning) return 'running-row';
  return '';
};

// 表格列：运行区间
const sessionColumns: DataTableColumns<any> = [
  {
    title: '交易所',
    key: 'exchange',
    width: 90,
    fixed: 'left',
    render: (row) => h(NTag, { size: 'small', bordered: false, type: 'info' }, { default: () => row.exchange || '--' }),
  },
  {
    title: '机器人',
    key: 'robotName',
    width: 150,
    ellipsis: { tooltip: true },
    render: (row) => row.robotName || (row.robotId ? `#${row.robotId}` : '--'),
  },
  {
    title: '交易对',
    key: 'symbol',
    width: 110,
    render: (row) => h(NText, { strong: true }, { default: () => row.symbol || '--' }),
  },
  {
    title: '状态',
    key: 'isRunning',
    width: 80,
    render: (row) => {
      return row.isRunning
        ? h(NTag, { type: 'success', size: 'small', bordered: false }, { default: () => '运行中' })
        : h(NTag, { type: 'default', size: 'small', bordered: false }, { default: () => '已结束' });
    },
  },
  {
    title: '启动时间',
    key: 'startTime',
    width: 150,
    sorter: (a: any, b: any) => Date.parse(a.startTime || '') - Date.parse(b.startTime || ''),
    render: (row) => row.startTime || '--',
  },
  {
    title: '结束时间',
    key: 'endTime',
    width: 150,
    render: (row) => (row.endTime ? row.endTime : (row.isRunning ? h(NText, { type: 'success' }, { default: () => '运行中' }) : '--')),
  },
  {
    title: '运行时长',
    key: 'runtimeText',
    width: 100,
    render: (row) => h(NText, { depth: 2 }, { default: () => row.runtimeText || '--' }),
  },
  {
    title: '盈亏',
    key: 'totalPnl',
    width: 110,
    sorter: (a: any, b: any) => (Number(a.totalPnl) || 0) - (Number(b.totalPnl) || 0),
    render: (row) => {
      const val = Number(row.totalPnl) || 0;
      return h(NText, { type: val >= 0 ? 'success' : 'error' }, { default: () => formatPnl(row.totalPnl) });
    },
  },
  {
    title: '手续费',
    key: 'totalFee',
    width: 100,
    render: (row) => h(NText, { type: 'warning', depth: 2 }, { default: () => formatFixed(row.totalFee, 4) }),
  },
  {
    title: '净盈亏',
    key: 'netPnl',
    width: 110,
    sorter: (a: any, b: any) => (Number(a.netPnl) || 0) - (Number(b.netPnl) || 0),
    render: (row) => {
      const val = Number(row.netPnl) || 0;
      return h(NText, { type: val >= 0 ? 'success' : 'error', strong: true }, { default: () => formatPnl(row.netPnl) });
    },
  },
  {
    title: '成交笔数',
    key: 'tradeCount',
    width: 80,
    align: 'center',
    render: (row) => row.tradeCount ?? 0,
  },
  {
    title: '操作',
    key: 'actions',
    width: 120,
    fixed: 'right',
    render: (row) =>
      h(NSpace, { size: 4 }, {
        default: () => [
          h(
            NButton,
            { size: 'tiny', tertiary: true, onClick: () => openDetail(row) },
            { default: () => '详情' }
          ),
          h(
            NButton,
            { size: 'tiny', type: 'primary', secondary: true, loading: row._syncing, onClick: () => handleSync(row) },
            { default: () => '同步' }
          ),
        ]
      }),
  },
];

// 表格列：成交流水
const orderColumns: DataTableColumns<any> = [
  {
    title: '交易所',
    key: 'exchange',
    width: 90,
    fixed: 'left',
    render: (row) => h(NTag, { size: 'small', bordered: false, type: 'info' }, { default: () => row.exchange || '--' }),
  },
  {
    title: '机器人',
    key: 'robotName',
    width: 150,
    ellipsis: { tooltip: true },
    render: (row) => row.robotName || (row.robotId ? `#${row.robotId}` : '--'),
  },
  {
    title: '交易对',
    key: 'symbol',
    width: 110,
    render: (row) => h(NText, { strong: true }, { default: () => row.symbol || '--' }),
  },
  {
    title: '方向',
    key: 'side',
    width: 90,
    render: (row) => {
      const val = String(row.side || '').toUpperCase();
      const text = val || '--';
      return h(NTag, { size: 'small', bordered: false, type: val === 'BUY' ? 'success' : (val === 'SELL' ? 'error' : 'default') }, { default: () => text });
    },
  },
  {
    title: '成交价',
    key: 'price',
    width: 110,
    render: (row) => formatFixed(row.price, 6),
  },
  {
    title: '成交量',
    key: 'qty',
    width: 110,
    render: (row) => formatFixed(row.qty, 8),
  },
  {
    title: '手续费',
    key: 'fee',
    width: 120,
    render: (row) => h(NText, { type: 'warning', depth: 2 }, { default: () => `${formatFixed(row.fee, FEE_DIGITS)} ${row.feeCoin || ''}`.trim() }),
  },
  {
    title: '已实现盈亏',
    key: 'realizedPnl',
    width: 150,
    sorter: (a: any, b: any) => (Number(a.realizedPnl) || 0) - (Number(b.realizedPnl) || 0),
    render: (row) => {
      const val = Number(row.realizedPnl) || 0;
      return h(NText, { type: val >= 0 ? 'success' : 'error' }, { default: () => formatPnl8(row.realizedPnl) });
    },
  },
  {
    title: '时间',
    key: 'time',
    width: 160,
    render: (row) => row.time || '--',
  },
  {
    title: '订单ID',
    key: 'orderId',
    width: 160,
    ellipsis: { tooltip: true },
    render: (row) => row.orderId || '--',
  },
  {
    title: '成交ID',
    key: 'tradeId',
    width: 160,
    ellipsis: { tooltip: true },
    render: (row) => row.tradeId || '--',
  },
  {
    title: '操作',
    key: 'actions',
    width: 90,
    fixed: 'right',
    render: (row) =>
      h(
        NButton,
        { size: 'tiny', tertiary: true, onClick: () => openDetail(row) },
        { default: () => '详情' }
      ),
  },
];

const tableColumns = computed(() => (viewMode.value === 'order' ? orderColumns : sessionColumns));

// 加载：运行区间数据
const loadSessionData = async () => {
  loading.value = true;
  try {
    const params: any = {
      page: pagination.page,
      pageSize: pagination.pageSize,
    };
    
    if (filterForm.exchange) {
      params.exchange = filterForm.exchange;
    }
    if (filterForm.robotId) {
      const rid = Number(filterForm.robotId);
      if (!Number.isNaN(rid) && rid > 0) params.robotId = rid;
    }
    if (filterForm.symbol) {
      params.symbol = filterForm.symbol;
    }
    if (filterForm.isRunning) {
      params.isRunning = filterForm.isRunning;
    }
    if (filterForm.startTime) {
      params.startTime = filterForm.startTime;
    }
    if (filterForm.endTime) {
      params.endTime = filterForm.endTime;
    }
    
    const res: any = await ToogoWalletApi.runSessionSummary(params);
    orderList.value = res.list || [];
    totalCount.value = res.totalCount || 0;
    pagination.itemCount = totalCount.value;
    pagination.pageCount = Math.ceil(totalCount.value / pagination.pageSize);
    
    // 使用后端返回的汇总统计（全量统计，不受分页影响）
    const s = res.summary || {};
    summary.totalSessions = Number(s.totalSessions) || 0;
    summary.totalRuntime = Number(s.totalRuntime) || 0;
    summary.totalRuntimeText = s.totalRuntimeText || '';
    summary.totalPnl = Number(s.totalPnl) || 0;
    summary.totalProfit = Number(s.totalProfit) || 0;
    summary.totalLoss = Number(s.totalLoss) || 0;
    summary.totalFee = Number(s.totalFee) || 0;
    summary.totalNetPnl = Number(s.totalNetPnl) || 0;
    summary.totalTrades = Number(s.totalTrades) || 0;
  } catch (error: any) {
    message.error(error.message || '加载运行区间列表失败');
  } finally {
    loading.value = false;
  }
};

// 加载：成交明细数据
const loadOrderData = async () => {
  loading.value = true;
  try {
    const params: any = {
      page: pagination.page,
      pageSize: pagination.pageSize,
    };

    if (filterForm.exchange) params.exchange = filterForm.exchange;
    if (filterForm.robotId) {
      const rid = Number(filterForm.robotId);
      if (!Number.isNaN(rid) && rid > 0) params.robotId = rid;
    }
    if (filterForm.symbol) params.symbol = filterForm.symbol;
    if (filterForm.side) params.side = filterForm.side;
    if (filterForm.startTime) params.startTime = filterForm.startTime;
    if (filterForm.endTime) params.endTime = filterForm.endTime;

    const res: any = await ToogoWalletApi.tradeHistory(params);
    orderList.value = res.list || [];
    totalCount.value = res.totalCount || 0;
    pagination.itemCount = totalCount.value;
    pagination.pageCount = Math.ceil(totalCount.value / pagination.pageSize);
    
    // 使用后端返回的汇总统计（全量统计，不受分页影响）
    const s = res.summary || {};
    summary.totalSessions = Number(s.totalCount) || totalCount.value;
    summary.totalPnl = Number(s.totalPnl) || 0;
    summary.totalProfit = Number(s.totalProfit) || 0;
    summary.totalLoss = Number(s.totalLoss) || 0;
    summary.totalFee = Number(s.totalFee) || 0;
    summary.totalNetPnl = Number(s.totalNetPnl) || 0;
  } catch (error: any) {
    message.error(error.message || '加载成交流水失败');
  } finally {
    loading.value = false;
  }
};

const loadData = async () => {
  if (viewMode.value === 'order') {
    await loadOrderData();
    return;
  }
  await loadSessionData();
};

watch(viewMode, () => {
  pagination.page = 1;
  loadData();
});

// 同步单个运行区间
const handleSync = async (row: any) => {
  if (!row?.id) return;
  row._syncing = true;
  try {
    await ToogoWalletApi.syncRunSession({ sessionId: Number(row.id) });
    message.success('同步成功');
    await loadData();
  } catch (e: any) {
    message.error(e?.message || '同步失败');
  } finally {
    row._syncing = false;
  }
};

// 搜索
const handleSearch = () => {
  pagination.page = 1;
  loadData();
};

// 重置
const handleReset = () => {
  filterForm.exchange = '';
  filterForm.robotId = null;
  filterForm.symbol = '';
  filterForm.isRunning = null;
  filterForm.side = '';
  filterForm.startTime = '';
  filterForm.endTime = '';
  dateRange.value = null;
  handleSearch();
};

// 日期范围变化
const handleDateRangeChange = (value: [number, number] | null) => {
  if (value) {
    filterForm.startTime = format(new Date(value[0]), 'yyyy-MM-dd HH:mm:ss');
    filterForm.endTime = format(new Date(value[1]), 'yyyy-MM-dd HH:mm:ss');
  } else {
    filterForm.startTime = '';
    filterForm.endTime = '';
  }
};

// 分页变化
const handlePageChange = (page: number) => {
  pagination.page = page;
  loadData();
};

const handlePageSizeChange = (pageSize: number) => {
  pagination.pageSize = pageSize;
  pagination.page = 1;
  loadData();
};

// 刷新
const handleRefresh = () => {
  loadData();
};

// 初始化
onMounted(() => {
  loadRobotOptions();
  loadData();
});
</script>

<style scoped lang="scss">
.order-history-page {
  padding: 16px;
  background: #f5f7fa;
  min-height: 100%;
  
  .page-header {
    margin-bottom: 12px;
    border-radius: 8px;
    
    .main-tabs {
      margin-bottom: 12px;
      
      :deep(.n-tabs-tab) {
        padding: 10px 24px;
        font-size: 15px;
        font-weight: 500;
      }
      
      .tab-item {
        display: flex;
        align-items: center;
        gap: 6px;
      }
    }
    
    .header-desc {
      font-size: 13px;
      display: block;
      margin-top: 4px;
    }
  }
  
  .stats-grid {
    margin-bottom: 12px;
    
    .stat-card {
      border-radius: 8px;
      transition: all 0.3s ease;
      
      &:hover {
        transform: translateY(-2px);
        box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
      }
      
      &.highlight-card {
        border: 1px solid rgba(24, 160, 88, 0.3);
      }
      
      .stat-content {
        display: flex;
        align-items: center;
        gap: 12px;
        
        .stat-icon {
          width: 42px;
          height: 42px;
          border-radius: 10px;
          display: flex;
          align-items: center;
          justify-content: center;
          flex-shrink: 0;
        }
        
        .stat-info {
          flex: 1;
          min-width: 0;
          
          .stat-label {
            font-size: 12px;
            color: #666;
            margin-bottom: 2px;
            white-space: nowrap;
          }
          
          .stat-value {
            font-size: 18px;
            font-weight: 600;
            color: #333;
            white-space: nowrap;
            
            &.profit {
              color: #18a058;
            }
            
            &.loss {
              color: #d03050;
            }
            
            &.fee {
              color: #f0a020;
            }
            
            .stat-unit {
              font-size: 11px;
              font-weight: 400;
              color: #999;
              margin-left: 2px;
            }
          }
        }
      }
    }
  }
  
  .filter-card {
    margin-bottom: 12px;
    border-radius: 8px;
    
    :deep(.n-input-group-label) {
      background: #f5f7fa;
      border-color: #e0e0e0;
    }
  }
  
  .table-card {
    border-radius: 8px;
    
    :deep(.n-card-header) {
      padding: 12px 16px;
      border-bottom: 1px solid #f0f0f0;
    }
    
    :deep(.n-data-table) {
      .n-data-table-th {
        background: #fafafa;
        font-weight: 600;
        font-size: 13px;
      }
      
      .n-data-table-td {
        font-size: 13px;
      }
      
      .running-row {
        background: rgba(24, 160, 88, 0.05);
      }
    }
  }
}

// 暗黑模式适配
html[data-theme='dark'] {
  .order-history-page {
    background: #18181c;
    
    .stat-card .stat-content .stat-info {
      .stat-label {
        color: #999;
      }
      
      .stat-value {
        color: #fff;
      }
    }
    
    .filter-card :deep(.n-input-group-label) {
      background: #2a2a2e;
      border-color: #444;
    }
  }
}
</style>
