<template>
  <div class="dashboard-container">
    <!-- 顶部状态栏：欢迎与机器人总览 -->
    <n-grid cols="1 s:1 m:3 l:3" :x-gap="12" :y-gap="12" responsive="screen">
      <n-gi>
        <n-card :bordered="false" class="welcome-card">
          <!-- 用户信息头部 -->
          <div class="user-info-header">
            <div class="user-main">
              <h2 class="greeting-text">{{ greeting }}</h2>
              <div class="username">{{ userInfo.username || '交易员' }}</div>
            </div>
            <div class="user-badges">
              <n-tag :type="userInfo.vipLevel > 1 ? 'warning' : 'default'" size="small" round>
                VIP {{ userInfo.vipLevel || 1 }}
              </n-tag>
              <n-tag :type="subscriptionInfo.planName ? 'info' : 'default'" size="small" round>
                {{ subscriptionInfo.planName || '无订阅' }}
              </n-tag>
              <n-tag v-if="toogoUserInfo.agentStatus === 2" type="success" size="small" round>
                <template #icon>
                  <n-icon :component="TeamOutlined" />
                </template>
                {{ toogoUserInfo.agentUnlockLevel === 1 ? '高级代理' : '代理' }}
              </n-tag>
            </div>
          </div>

          <n-divider style="margin: 12px 0;" />

          <!-- 订阅信息 -->
          <div class="subscription-info">
            <div class="subscription-item" @click="$router.push('/toogo/subscription')">
              <div class="sub-icon">
                <n-icon :component="RobotOutlined" />
          </div>
              <div class="sub-content">
                <div class="sub-plan">{{ subscriptionInfo.planName || '无订阅' }}</div>
                <div class="sub-label">{{ subscriptionInfo.planName ? '订阅有效期' : '请购买订阅' }}</div>
                <div class="sub-value" v-if="subscriptionInfo.planName">
                  <span :class="subscriptionInfo.remainingDays > 7 ? 'normal' : 'warning'">
                    {{ subscriptionInfo.remainingDays || 0 }}
                  </span>
                  <span class="unit">天</span>
          </div>
            </div>
              <n-button text type="primary" size="small">续费</n-button>
            </div>
          </div>

          <n-divider style="margin: 12px 0;" />

          <!-- 账户信息网格 -->
          <div class="account-grid">
            <div class="account-card" @click="showDepositModal = true">
              <div class="account-header">
                <n-icon :component="WalletOutlined" :size="14" class="account-icon-small balance" />
                <span class="account-title">账户余额</span>
          </div>
              <div class="account-amount">
                {{ walletData.balance?.toFixed(2) || '0.00' }}
                <span class="account-unit">USDT</span>
          </div>
            </div>

            <div class="account-card" @click="showTransferModal = true">
              <div class="account-header">
                <n-icon :component="ThunderboltOutlined" :size="14" class="account-icon-small power" />
                <span class="account-title">云算力</span>
            </div>
              <div class="account-amount">
                {{ walletData.power?.toFixed(2) || '0.00' }}
                <span class="account-unit">Power</span>
          </div>
          </div>

            <div class="account-card" @click="$router.push('/toogo/finance')">
              <div class="account-header">
                <n-icon :component="GiftOutlined" :size="14" class="account-icon-small gift" />
                <span class="account-title">积分账户</span>
              </div>
              <div class="account-amount">
                {{ walletData.giftPower?.toFixed(2) || '0.00' }}
                <span class="account-unit">积分</span>
              </div>
            </div>

            <div class="account-card" @click="$router.push('/toogo/commission')">
              <div class="account-header">
                <n-icon :component="RiseOutlined" :size="14" class="account-icon-small commission" />
                <span class="account-title">累计佣金</span>
              </div>
              <div class="account-amount">
                {{ walletData.totalCommission?.toFixed(2) || '0.00' }}
                <span class="account-unit">USDT</span>
              </div>
            </div>
          </div>
        </n-card>
      </n-gi>
      <n-gi span="2">
        <n-card :bordered="false" class="robot-overview-card">
          <template #header>
            <div style="display: flex; justify-content: space-between; align-items: center;">
              <span>机器人总览</span>
              <n-button text size="tiny" @click="refreshRobotData" :loading="robotLoading">
                <template #icon><n-icon :component="ReloadOutlined" /></template>
                刷新
              </n-button>
            </div>
          </template>
          <div class="robot-overview-content">
            <!-- 左侧：机器人动画 -->
            <div class="robot-animation-container">
              <ToogoRobot />
            </div>

            <!-- 右侧：信息展示 -->
            <div class="robot-info-panel">
              <!-- 机器人状态分布 -->
              <div class="robot-status-distribution">
                <div class="status-item running">
                  <div class="status-count">{{ activeRobotCount }}</div>
                  <div class="status-label">运行中</div>
          </div>
                <div class="status-divider"></div>
                <div class="status-item paused">
                  <div class="status-count">{{ pausedRobotCount }}</div>
                  <div class="status-label">已暂停</div>
          </div>
                <div class="status-divider"></div>
                <div class="status-item idle">
                  <div class="status-count">{{ notStartedRobotCount }}</div>
                  <div class="status-label">未启动</div>
                </div>
              </div>

              <!-- 统计数据 -->
              <div class="robot-stats-grid">
                <div class="stat-mini-item">
                  <div class="stat-label">机器人额度</div>
                  <div class="stat-value">{{ userInfo.robotLimit || 1 }}</div>
                </div>
                <div class="stat-mini-item">
                  <div class="stat-label">已占用额度</div>
                  <div class="stat-value">{{ usedQuota }}</div>
                </div>
                <div class="stat-mini-item">
                  <div class="stat-label">今日净盈亏</div>
                  <div class="stat-value" :class="todayPnl >= 0 ? 'profit' : 'loss'">
                    {{ todayPnl >= 0 ? '+' : '' }}{{ todayPnl.toFixed(2) }}
                  </div>
                </div>
                <div class="stat-mini-item">
                  <div class="stat-label">累计净盈亏</div>
                  <div class="stat-value" :class="totalPnl >= 0 ? 'profit' : 'loss'">
                    {{ totalPnl >= 0 ? '+' : '' }}{{ totalPnl.toFixed(2) }}
                  </div>
                </div>
              </div>
            </div>
          </div>
        </n-card>
      </n-gi>
    </n-grid>

    <!-- 快速开始（横向步骤流程）-->
    <n-card :bordered="false" class="quick-start-card proCard mt-3" size="small">
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center;">
          <span>快速开始</span>
          <n-text depth="3" style="font-size: 12px;">
            已完成 <span style="color: #18a058; font-weight: 600;">{{ completedStepsCount }}</span>/6 步
          </n-text>
        </div>
          </template>
      
      <div class="horizontal-steps">
        <!-- 步骤1: 添加API -->
        <div 
          class="h-step-item"
          :class="{ 
            completed: hasCompletedStep(1),
            current: getCurrentStep() === 1
          }"
          @click="$router.push('/toogo/api')"
        >
          <div class="h-step-circle">
            <n-icon v-if="hasCompletedStep(1)" :component="CheckOutlined" />
            <span v-else>1</span>
              </div>
          <div class="h-step-line"></div>
          <div class="h-step-content">
            <div class="h-step-name">添加API接口</div>
            <div class="h-step-hint">绑定交易所</div>
            </div>
          </div>

        <!-- 步骤2: 订阅 -->
        <div 
          class="h-step-item"
          :class="{ 
            completed: hasCompletedStep(2),
            current: getCurrentStep() === 2
          }"
          @click="$router.push('/toogo/subscription')"
        >
          <div class="h-step-circle">
            <n-icon v-if="hasCompletedStep(2)" :component="CheckOutlined" />
            <span v-else>2</span>
              </div>
          <div class="h-step-line"></div>
          <div class="h-step-content">
            <div class="h-step-name">订阅机器人</div>
            <div class="h-step-hint">获取配额</div>
              </div>
            </div>

        <!-- 步骤3: 创建 -->
        <div 
          class="h-step-item"
          :class="{ 
            completed: hasCompletedStep(3),
            current: getCurrentStep() === 3
          }"
          @click="$router.push('/toogo/robot/create')"
        >
          <div class="h-step-circle">
            <n-icon v-if="hasCompletedStep(3)" :component="CheckOutlined" />
            <span v-else>3</span>
          </div>
          <div class="h-step-line"></div>
          <div class="h-step-content">
            <div class="h-step-name">创建机器人</div>
            <div class="h-step-hint">开始交易</div>
          </div>
        </div>

        <!-- 步骤4: 优化 -->
        <div 
          class="h-step-item"
          :class="{ 
            completed: hasCompletedStep(4),
            current: getCurrentStep() === 4
          }"
          @click="$router.push('/toogo/strategy')"
        >
          <div class="h-step-circle">
            <n-icon v-if="hasCompletedStep(4)" :component="CheckOutlined" />
            <span v-else>4</span>
          </div>
          <div class="h-step-line"></div>
          <div class="h-step-content">
            <div class="h-step-name">优化策略</div>
            <div class="h-step-hint">提升收益</div>
          </div>
        </div>

        <!-- 步骤5: 代理 -->
        <div 
          class="h-step-item"
          :class="{ 
            completed: hasCompletedStep(5),
            current: getCurrentStep() === 5
          }"
          @click="$router.push('/toogo/team')"
        >
          <div class="h-step-circle">
            <n-icon v-if="hasCompletedStep(5)" :component="CheckOutlined" />
            <span v-else>5</span>
          </div>
          <div class="h-step-line"></div>
          <div class="h-step-content">
            <div class="h-step-name">申请代理</div>
            <div class="h-step-hint">赚取佣金</div>
          </div>
        </div>

        <!-- 步骤6: 邀请 -->
        <div 
          class="h-step-item last"
          :class="{ 
            completed: hasCompletedStep(6),
            current: getCurrentStep() === 6
          }"
          @click="copyPermanentCode"
        >
          <div class="h-step-circle">
            <n-icon v-if="hasCompletedStep(6)" :component="CheckOutlined" />
            <span v-else>6</span>
          </div>
          <div class="h-step-content">
            <div class="h-step-name">邀请用户</div>
            <div class="h-step-hint">扩大团队</div>
          </div>
        </div>
      </div>
    </n-card>


    <!-- 用户中心与运行区间 -->
    <n-grid cols="1 s:1 m:1 l:3" :x-gap="16" :y-gap="16" responsive="screen" class="mt-4">
      <!-- 我的邀请码 -->
      <n-gi>
        <n-card title="我的邀请码" :bordered="false" size="small">
          <n-space vertical :size="16">
            <!-- 永久邀请码 - 仅高级代理可见 -->
            <div class="invite-section" v-if="toogoUserInfo.agentUnlockLevel === 1">
              <div class="section-header">
                <n-icon :component="KeyOutlined" size="18" />
                <span>永久邀请码</span>
                <n-tag type="success" size="tiny" :bordered="false">永久有效</n-tag>
                <n-tag type="warning" size="tiny" :bordered="false" style="margin-left: 4px;">高级代理专属</n-tag>
              </div>
              <div class="code-display">
                <n-text code strong style="font-size: 20px; letter-spacing: 2px;">
                  {{ baseUserInfo.inviteCode || '------' }}
                </n-text>
                <n-button text type="primary" size="small" @click="copyPermanentCode">
                  <template #icon><n-icon :component="CopyOutlined" /></template>
              </n-button>
              </div>
            </div>

            <n-divider style="margin: 8px 0" v-if="toogoUserInfo.agentUnlockLevel === 1" />

            <!-- 动态邀请码 -->
            <div class="invite-section">
              <div class="section-header">
                <n-icon :component="ThunderboltOutlined" size="18" />
                <span>动态邀请码</span>
                <n-tag type="warning" size="tiny" :bordered="false">24小时</n-tag>
              </div>
              <div class="code-display">
                <n-text code strong style="font-size: 20px; letter-spacing: 2px;">
                  {{ toogoUserInfo.inviteCode || '------' }}
                </n-text>
                <n-button text type="primary" size="small" @click="copyDynamicCode">
                  <template #icon><n-icon :component="CopyOutlined" /></template>
                </n-button>
                <n-button text type="info" size="small" :loading="refreshing" @click="refreshDynamicCode">
                  <template #icon><n-icon :component="ReloadOutlined" /></template>
                </n-button>
              </div>
              <n-text depth="3" style="font-size: 11px;" v-if="toogoUserInfo.inviteCodeExpire">
                过期：{{ formatExpireTime(toogoUserInfo.inviteCodeExpire) }}
              </n-text>
            </div>
            </n-space>
        </n-card>
      </n-gi>

      <!-- 运行区间列表 -->
      <n-gi span="2">
        <n-card title="运行区间 - 最近记录" :bordered="false" size="small">
          <template #header-extra>
            <n-button type="primary" size="small" @click="$router.push('/toogo/wallet/order-history')">
              查看全部
              </n-button>
          </template>

          <div v-if="sessionLoading" class="p-8 text-center">
            <n-spin size="medium" />
          </div>
          <n-empty v-else-if="sessionList.length === 0" description="暂无运行区间记录" size="small" class="py-8" />
          
          <n-data-table
            v-else
            :columns="sessionColumns"
            :data="sessionList"
            :pagination="false"
            size="small"
            :max-height="400"
            :scroll-x="1200"
          />
        </n-card>
      </n-gi>
    </n-grid>

    <!-- 弹窗组件 -->
    <n-modal v-model:show="showDepositModal" title="USDT充值" preset="dialog" style="width: 420px">
      <n-form :model="depositForm" label-placement="left" label-width="80">
        <n-form-item label="充值金额">
          <n-input-number v-model:value="depositForm.amount" :min="10" :precision="2" placeholder="最低10 USDT" style="width: 100%">
            <template #suffix>USDT</template>
          </n-input-number>
        </n-form-item>
        <n-form-item label="网络">
          <n-select v-model:value="depositForm.network" :options="networkOptions" />
        </n-form-item>
      </n-form>
      <template #action>
        <n-space justify="end">
          <n-button @click="showDepositModal = false">取消</n-button>
          <n-button type="primary" @click="handleDeposit" :loading="depositLoading">确认充值</n-button>
        </n-space>
      </template>
    </n-modal>

    <n-modal v-model:show="showWithdrawModal" title="提现申请" preset="dialog" style="width: 480px">
      <n-form :model="withdrawForm" label-placement="left" label-width="80">
        <n-form-item label="账户">
          <n-select v-model:value="withdrawForm.accountType" :options="accountTypeOptions" />
        </n-form-item>
        <n-form-item label="提现金额">
          <n-input-number v-model:value="withdrawForm.amount" :min="10" :precision="2" style="width: 100%">
            <template #suffix>USDT</template>
          </n-input-number>
        </n-form-item>
        <n-form-item label="网络">
          <n-select v-model:value="withdrawForm.network" :options="networkOptions" />
        </n-form-item>
        <n-form-item label="钱包地址">
          <n-input v-model:value="withdrawForm.toAddress" placeholder="请输入USDT钱包地址" />
        </n-form-item>
      </n-form>
      <template #action>
        <n-space justify="end">
          <n-button @click="showWithdrawModal = false">取消</n-button>
          <n-button type="primary" @click="handleWithdraw" :loading="withdrawLoading">提交申请</n-button>
        </n-space>
      </template>
    </n-modal>

    <n-modal v-model:show="showTransferModal" title="转入算力" preset="dialog" style="width: 420px">
      <n-form :model="transferForm" label-placement="left" label-width="80">
        <n-form-item label="来源账户">
          <n-select v-model:value="transferForm.fromAccount" :options="transferAccountOptions" />
        </n-form-item>
        <n-form-item label="转入金额">
          <n-input-number v-model:value="transferForm.amount" :min="1" :precision="2" style="width: 100%" />
        </n-form-item>
        <n-form-item>
          <n-text depth="3">余额和佣金账户可以1:1转入算力，无手续费</n-text>
        </n-form-item>
      </n-form>
      <template #action>
        <n-space justify="end">
          <n-button @click="showTransferModal = false">取消</n-button>
          <n-button type="primary" @click="handleTransfer" :loading="transferLoading">确认转入</n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, onActivated, computed, h } from 'vue';
import { useRouter } from 'vue-router';
import { useMessage, NTag, NText } from 'naive-ui';
import {
  WalletOutlined, ThunderboltOutlined, RiseOutlined, GiftOutlined,
  KeyOutlined, CopyOutlined, ReloadOutlined, TeamOutlined, CheckOutlined, RobotOutlined
} from '@vicons/antd';
import { ToogoWalletApi, ToogoUserApi, ToogoSubscriptionApi, ToogoCommissionApi, ToogoRobotApi, ToogoFinanceApi } from '@/api/toogo';
import { getUserInfo } from '@/api/system/user';
import ToogoRobot from '@/components/ToogoRobot/index.vue';

const router = useRouter();
const message = useMessage();

// 定时器
let refreshTimer: any = null;

// 数据状态
const walletData = ref<any>({});
const userInfo = ref<any>({});
const baseUserInfo = ref<any>({});
const toogoUserInfo = ref<any>({});
const subscriptionInfo = ref<any>({});
const commissionStat = ref<any>({});
const robotList = ref<any[]>([]);
const sessionList = ref<any[]>([]);
const sessionLoading = ref(false);
const robotLoading = ref(false);
const refreshing = ref(false);

// 弹窗与加载状态
const showDepositModal = ref(false);
const showWithdrawModal = ref(false);
const showTransferModal = ref(false);
const depositLoading = ref(false);
const withdrawLoading = ref(false);
const transferLoading = ref(false);

const depositForm = ref({ amount: 100, currency: 'USDT', network: 'TRC20' });
const withdrawForm = ref({ accountType: 'balance', amount: 10, currency: 'USDT', network: 'TRC20', toAddress: '' });
const transferForm = ref({ fromAccount: 'balance', amount: 100 });

// 选项
const networkOptions = [{ label: 'TRC20', value: 'TRC20' }, { label: 'ERC20', value: 'ERC20' }];
const accountTypeOptions = [{ label: '余额账户', value: 'balance' }, { label: '佣金账户', value: 'commission' }];
const transferAccountOptions = [{ label: '余额账户', value: 'balance' }, { label: '佣金账户', value: 'commission' }];

// 运行区间表格列
const sessionColumns = [
  {
    title: '交易所',
    key: 'exchange',
    width: 90,
    render: (row: any) => h(NTag, { size: 'small', bordered: false, type: 'info' }, { default: () => row.exchange || '--' }),
  },
  {
    title: '机器人',
    key: 'robotName',
    width: 150,
    ellipsis: { tooltip: true },
    render: (row: any) => row.robotName || (row.robotId ? `#${row.robotId}` : '--'),
  },
  { 
    title: '交易对',
    key: 'symbol',
    width: 100,
    render: (row: any) => h(NText, { strong: true }, { default: () => row.symbol || '--' }),
  },
  {
    title: '状态',
    key: 'isRunning',
    width: 80,
    render: (row: any) => {
      return row.isRunning
        ? h(NTag, { type: 'success', size: 'small', bordered: false }, { default: () => '运行中' })
        : h(NTag, { type: 'default', size: 'small', bordered: false }, { default: () => '已结束' });
    },
  },
  { 
    title: '启动时间',
    key: 'startTime',
    width: 150,
    render: (row: any) => row.startTime || '--',
  },
  {
    title: '运行时长',
    key: 'runtimeText',
    width: 100,
    render: (row: any) => h(NText, { depth: 2 }, { default: () => row.runtimeText || '--' }),
  },
  {
    title: '盈亏',
    key: 'totalPnl',
    width: 110,
    render: (row: any) => {
      const val = Number(row.totalPnl) || 0;
      return h(NText, { type: val >= 0 ? 'success' : 'error' }, { default: () => `${val >= 0 ? '+' : ''}${val.toFixed(4)}` });
    },
  },
  {
    title: '净盈亏',
    key: 'netPnl',
    width: 110,
    render: (row: any) => {
      const val = Number(row.netPnl) || 0;
      return h(NText, { type: val >= 0 ? 'success' : 'error', strong: true }, { default: () => `${val >= 0 ? '+' : ''}${val.toFixed(4)}` });
    },
  },
  {
    title: '成交笔数',
    key: 'tradeCount',
    width: 80,
    align: 'center' as const,
    render: (row: any) => row.tradeCount ?? 0,
  },
];

// 计算属性
const greeting = computed(() => {
  const hour = new Date().getHours();
  if (hour < 6) return '凌晨好';
  if (hour < 12) return '上午好';
  if (hour < 18) return '下午好';
  return '晚上好';
});

// 机器人统计数据
const totalRobots = computed(() => robotList.value.length);
const activeRobotCount = computed(() => robotList.value.filter(r => r.status === 2).length); // 运行中
const pausedRobotCount = computed(() => robotList.value.filter(r => r.status === 3).length); // 已暂停
const notStartedRobotCount = computed(() => robotList.value.filter(r => r.status === 1).length); // 未启动
const usedQuota = computed(() => activeRobotCount.value + pausedRobotCount.value + notStartedRobotCount.value); // 已占用额度

// 盈亏数据（从交易记录获取，最精准）⭐
const todayPnl = ref(0);      // 今日净盈亏
const totalPnl = ref(0);      // 累计净盈亏
const totalTrades = ref(0);   // 总交易笔数

// 格式化过期时间
const formatExpireTime = (time: string) => {
  if (!time) return '';
  const date = new Date(time);
  const now = new Date();
  const diff = date.getTime() - now.getTime();
  
  if (diff < 0) return '已过期';
  
  const hours = Math.floor(diff / (1000 * 60 * 60));
  const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));
  
  return `剩余 ${hours}h${minutes}m`;
};

// 方法
const loadData = async () => {
  try {
    const [walletRes, baseUserRes, toogoUserRes, subRes, commRes] = await Promise.all([
      ToogoWalletApi.overview(),
      getUserInfo(),
      ToogoUserApi.info(),
      ToogoSubscriptionApi.mySubscription(),
      ToogoCommissionApi.stat(),
    ]);
    walletData.value = walletRes || {};
    baseUserInfo.value = baseUserRes || {};
    toogoUserInfo.value = toogoUserRes || {};
    userInfo.value = toogoUserRes || {};
    subscriptionInfo.value = subRes || {};
    commissionStat.value = commRes || {};
  } catch (error) {
    console.error('加载数据失败:', error);
  }
};

const loadRobotList = async () => {
  robotLoading.value = true;
  try {
    // 后端要求使用 Page 和 PageSize（大驼峰）
    const res = await ToogoRobotApi.list({ Page: 1, PageSize: 100 });
    console.log('控制台-机器人API返回:', res);
    
    // 兼容不同的返回格式
    if (res) {
      if (Array.isArray(res)) {
        robotList.value = res;
      } else if (res.list) {
        robotList.value = res.list;
      } else if (res.data && res.data.list) {
        robotList.value = res.data.list;
      } else {
        robotList.value = [];
      }
    } else {
      robotList.value = [];
    }
    
    console.log('控制台-机器人列表:', robotList.value);
    console.log('控制台-总数:', robotList.value.length);
    console.log('控制台-运行中:', robotList.value.filter(r => r.status === 2).length);
  } catch (error) {
    console.error('加载机器人失败:', error);
    robotList.value = [];
  } finally {
    robotLoading.value = false;
  }
};

const loadSessionList = async () => {
  sessionLoading.value = true;
  try {
    const res: any = await ToogoWalletApi.runSessionSummary({ Page: 1, PageSize: 10 });
    sessionList.value = res?.list || [];
  } catch (error: any) {
    console.error('加载运行区间失败:', error);
    message.error(error.message || '加载运行区间失败');
  } finally {
    sessionLoading.value = false;
  }
};

// 加载盈亏数据（基于成交流水统计，最精准）⭐
const loadPnlData = async () => {
  try {
    // 获取今天的开始和结束时间（0:00 - 24:00）
    const now = new Date();
    const todayStart = new Date(now.getFullYear(), now.getMonth(), now.getDate(), 0, 0, 0);
    const todayEnd = new Date(now.getFullYear(), now.getMonth(), now.getDate(), 23, 59, 59);

    // 并行加载今日和累计盈亏（使用成交流水API）
    const [todayRes, totalRes] = await Promise.all([
      // 今日盈亏：0:00-24:00的成交流水统计
      ToogoWalletApi.tradeHistory({
        page: 1,
        pageSize: 1, // 只需要汇总数据
        startTime: todayStart.toISOString(),
        endTime: todayEnd.toISOString(),
      }),
      // 累计盈亏：所有成交流水统计
      ToogoWalletApi.tradeHistory({
        page: 1,
        pageSize: 1, // 只需要汇总数据
      }),
    ]);

    // 今日盈亏（使用净盈亏 = 盈亏 - 手续费）
    const todaySummary = todayRes?.summary || {};
    todayPnl.value = Number(todaySummary.totalNetPnl) || 0;

    // 累计盈亏（使用净盈亏 = 盈亏 - 手续费）
    const totalSummary = totalRes?.summary || {};
    totalPnl.value = Number(totalSummary.totalNetPnl) || 0;
    totalTrades.value = Number(totalSummary.totalCount) || 0;

    console.log('📊 盈亏数据已更新（基于成交流水）:', {
      今日净盈亏: todayPnl.value.toFixed(2) + ' USDT',
      累计净盈亏: totalPnl.value.toFixed(2) + ' USDT',
      总成交笔数: totalTrades.value,
      统计时间: `${todayStart.toLocaleTimeString()} - ${todayEnd.toLocaleTimeString()}`,
    });
  } catch (error: any) {
    console.error('加载盈亏数据失败:', error);
    // 不显示错误消息，避免干扰用户
  }
};

// 判断快速开始步骤是否完成
const hasCompletedStep = (step: number) => {
  switch (step) {
    case 1: // 添加API接口
      return (userInfo.value.robotLimit || 0) > 0; // 有机器人额度说明已添加API
    case 2: // 订阅机器人
      return (subscriptionInfo.value.remainingDays || 0) > 0; // 有有效订阅
    case 3: // 创建机器人
      return totalRobots.value > 0; // 有机器人
    case 4: // 优化策略
      return activeRobotCount.value > 0; // 有运行中的机器人
    case 5: // 申请代理
      return toogoUserInfo.value.agentStatus === 2; // 代理已通过
    case 6: // 邀请用户
      return (baseUserInfo.value.inviteCount || 0) > 0; // 有邀请用户
    default:
      return false;
  }
};

// 获取当前应该进行的步骤（第一个未完成的步骤）
const getCurrentStep = () => {
  for (let i = 1; i <= 6; i++) {
    if (!hasCompletedStep(i)) {
      return i;
    }
  }
  return 6; // 全部完成，返回最后一步
};

// 已完成步骤数
const completedStepsCount = computed(() => {
  let count = 0;
  for (let i = 1; i <= 6; i++) {
    if (hasCompletedStep(i)) count++;
  }
  return count;
});

// 复制到剪贴板的通用方法
const copyToClipboard = (text: string) => {
  const input = document.createElement('input');
  input.value = text;
  input.style.position = 'fixed';
  input.style.opacity = '0';
  document.body.appendChild(input);
  input.select();
  
  try {
    const success = document.execCommand('copy');
    document.body.removeChild(input);
    return success;
  } catch (err) {
    document.body.removeChild(input);
    return false;
  }
};

// 复制永久邀请码
const copyPermanentCode = () => {
  const code = baseUserInfo.value?.inviteCode || '';
  if (!code) {
    message.error('邀请码为空');
    return;
  }
  
  if (copyToClipboard(code)) {
    message.success('永久邀请码已复制');
  } else {
    message.error('复制失败，请手动复制');
  }
};

// 复制动态邀请码
const copyDynamicCode = () => {
  const code = toogoUserInfo.value?.inviteCode || '';
  if (!code) {
    message.error('邀请码为空');
    return;
  }
  
  if (copyToClipboard(code)) {
    message.success('动态邀请码已复制');
  } else {
    message.error('复制失败，请手动复制');
  }
};

// 刷新动态邀请码
const refreshDynamicCode = async () => {
  refreshing.value = true;
  try {
    const res = await ToogoUserApi.refreshInviteCode();
    toogoUserInfo.value.inviteCode = res?.inviteCode;
    toogoUserInfo.value.inviteCodeExpire = res?.inviteCodeExpire;
    message.success('动态邀请码已刷新');
  } catch (error) {
    message.error('刷新失败');
  } finally {
    refreshing.value = false;
  }
};

// 弹窗操作逻辑
const handleDeposit = async () => {
  depositLoading.value = true;
  try {
    const res = await ToogoFinanceApi.createDeposit(depositForm.value);
    message.success(`订单已创建，请向 ${res?.toAddress} 转账`);
    showDepositModal.value = false;
  } catch (e: any) {
    message.error(e.message || '失败');
  } finally {
    depositLoading.value = false;
  }
};

const handleWithdraw = async () => {
  withdrawLoading.value = true;
  try {
    await ToogoFinanceApi.createWithdraw(withdrawForm.value);
    message.success('提现申请已提交');
    showWithdrawModal.value = false;
    loadData();
  } catch (e: any) {
    message.error(e.message || '失败');
  } finally {
    withdrawLoading.value = false;
  }
};

const handleTransfer = async () => {
  transferLoading.value = true;
  try {
    await ToogoWalletApi.transfer({
      fromAccount: transferForm.value.fromAccount,
      toAccount: 'power',
      amount: transferForm.value.amount,
    });
    message.success('转入成功');
    showTransferModal.value = false;
    loadData();
  } catch (e: any) {
    message.error(e.message || '失败');
  } finally {
    transferLoading.value = false;
  }
};

// 手动刷新机器人数据
const refreshRobotData = async () => {
  await Promise.all([loadRobotList(), loadData(), loadPnlData()]);
  message.success('数据已刷新');
};

// 启动自动刷新（每15秒刷新机器人数据）
const startAutoRefresh = () => {
  refreshTimer = setInterval(() => {
    loadRobotList();
    loadData();
    loadPnlData(); // 自动刷新精准盈亏数据
  }, 15000); // 15秒刷新一次
};

// 停止自动刷新
const stopAutoRefresh = () => {
  if (refreshTimer) {
    clearInterval(refreshTimer);
    refreshTimer = null;
  }
};

onMounted(() => {
  loadData();
  loadRobotList();
  loadSessionList();
  loadPnlData(); // 加载精准盈亏数据
  startAutoRefresh();
});

// 页面激活时（从其他页面返回）刷新数据
onActivated(() => {
  loadData();
  loadRobotList();
  loadSessionList();
  loadPnlData(); // 刷新精准盈亏数据
});

onUnmounted(() => {
  stopAutoRefresh();
});
</script>

<style lang="less" scoped>
.dashboard-container {
  padding: 8px;
  min-height: calc(100vh - 100px);

  .mt-3 { margin-top: 12px; }
  .mt-4 { margin-top: 16px; }
  .mb-2 { margin-bottom: 8px; }
  .mb-3 { margin-bottom: 12px; }

  .welcome-card {
    height: 100%;
    display: flex;
    flex-direction: column;

    :deep(.n-card__content) {
      padding: 12px 16px !important;
      flex: 1;
      display: flex;
      flex-direction: column;
}

    .user-info-header {
      display: flex;
      justify-content: space-between;
      align-items: flex-start;
      flex-wrap: wrap;
      gap: 12px;

      .user-main {
        flex: 1;
        min-width: 0;
        
        .greeting-text {
          margin: 0;
          font-size: 13px;
          font-weight: 500;
          color: #999;
          line-height: 1.2;
        }

        .username {
          font-size: 20px;
  font-weight: 700;
          color: #333;
          margin-top: 4px;
          line-height: 1.2;
        }
}

      .user-badges {
  display: flex;
        gap: 6px;
        flex-wrap: wrap;
  align-items: center;
        padding-top: 2px;
      }
}

    .subscription-info {
      .subscription-item {
        display: flex;
        align-items: center;
        gap: 12px;
        padding: 12px;
        background: #fffbeb;
        border: 1px solid #fed7aa;
        border-radius: 8px;
        cursor: pointer;
        transition: all 0.2s;

        &:hover {
          background: #fef3c7;
          border-color: #fdba74;
}

        .sub-icon {
          width: 40px;
          height: 40px;
          background: #fa8c16;
          border-radius: 6px;
          display: flex;
          align-items: center;
          justify-content: center;
          color: #fff;
          font-size: 20px;
          flex-shrink: 0;
}

        .sub-content {
          flex: 1;
          min-width: 0;
          display: flex;
          flex-direction: column;
          justify-content: center;

          .sub-plan {
            font-size: 14px;
            font-weight: 700;
            color: #92400e;
            margin-bottom: 4px;
            line-height: 1.2;
}

          .sub-label {
            font-size: 11px;
            color: #92400e;
            margin-bottom: 3px;
            line-height: 1.2;
            opacity: 0.8;
}

          .sub-value {
            font-size: 22px;
  font-weight: 700;
            color: #92400e;
            line-height: 1.1;
            display: flex;
            align-items: baseline;

            span.normal {
              color: #059669;
}

            span.warning {
              color: #dc2626;
            }

            .unit {
              font-size: 13px;
              font-weight: 500;
              margin-left: 3px;
            }
          }
        }

        :deep(.n-button) {
          align-self: center;
        }
      }
    }

    .account-grid {
      display: grid;
      grid-template-columns: repeat(2, 1fr);
      gap: 8px;
      flex: 1;

      .account-card {
        padding: 12px 8px;
        background: #fafafa;
        border: 1px solid #e5e7eb;
        border-radius: 6px;
        cursor: pointer;
        transition: all 0.2s;
        display: flex;
        flex-direction: column;

        &:hover {
          background: #fff;
          border-color: #ccc;
}

        .account-header {
  display: flex;
  align-items: center;
          gap: 6px;
          margin-bottom: 6px;

          .account-icon-small {
            width: 24px;
            height: 24px;
            padding: 0;
            border-radius: 6px;
            display: flex;
            align-items: center;
            justify-content: center;
            flex-shrink: 0;
            font-size: 14px;

            &.balance { color: #3b82f6; background: #eff6ff; }
            &.power { color: #f59e0b; background: #fffbeb; }
            &.gift { color: #06b6d4; background: #ecfeff; }
            &.commission { color: #10b981; background: #f0fdf4; }
    }

          .account-title {
      font-size: 12px;
            color: #666;
            font-weight: 500;
            line-height: 1.2;
  }
}

        .account-amount {
    font-size: 18px;
    font-weight: 700;
          color: #333;
          line-height: 1.2;
          display: flex;
          align-items: baseline;
          padding-left: 32px;

          .account-unit {
            font-size: 11px;
            font-weight: 400;
            color: #999;
            margin-left: 3px;
          }
        }
    }
  }
}

  .quick-start-card {
  :deep(.n-card__content) {
      padding: 16px !important;
}

    .horizontal-steps {
    display: flex;
    justify-content: space-between;
      align-items: flex-start;
      gap: 0;
      padding: 4px 0;

      .h-step-item {
        flex: 1;
      display: flex;
        flex-direction: column;
      align-items: center;
        position: relative;
        cursor: pointer;
        transition: all 0.2s;

        &:hover {
          .h-step-circle {
            transform: scale(1.05);
          }
    }

        .h-step-circle {
          width: 32px;
          height: 32px;
          border-radius: 50%;
          background: #f5f5f5;
          border: 2px solid #d9d9d9;
      display: flex;
          align-items: center;
          justify-content: center;
          font-size: 14px;
          font-weight: 600;
          color: #999;
          transition: all 0.2s;
          z-index: 2;
          position: relative;
        }

        .h-step-line {
          position: absolute;
          top: 16px;
          left: 50%;
          width: 100%;
          height: 2px;
          background: #e8e8e8;
          z-index: 1;
          transition: all 0.2s;
        }

        .h-step-content {
          margin-top: 8px;
          text-align: center;

          .h-step-name {
            font-size: 12px;
            font-weight: 600;
            color: #666;
            margin-bottom: 2px;
            transition: all 0.2s;
      }

          .h-step-hint {
            font-size: 10px;
            color: #999;
            transition: all 0.2s;
          }
        }

        // 已完成状态
        &.completed {
          .h-step-circle {
            background: #52c41a;
            border-color: #52c41a;
            color: #fff;
          }

          .h-step-line {
            background: #52c41a;
          }

          .h-step-name {
            color: #52c41a;
      }

          .h-step-hint {
            color: #52c41a;
  }
}

        // 当前步骤
        &.current {
          .h-step-circle {
            background: #1890ff;
            border-color: #1890ff;
            color: #fff;
          }

          .h-step-name {
            color: #1890ff;
            font-weight: 600;
          }

          .h-step-hint {
            color: #1890ff;
          }
        }

        // 最后一个步骤（没有连接线）
        &.last {
          .h-step-line {
            display: none;
          }
        }
      }
    }

    // 响应式：小屏幕改为垂直布局
    @media (max-width: 768px) {
      .horizontal-steps {
        flex-direction: column;
        gap: 10px;

        .h-step-item {
          flex-direction: row;
  align-items: center;
          justify-content: flex-start;

          .h-step-circle {
            width: 28px;
            height: 28px;
            font-size: 13px;
          }

          .h-step-line {
            display: none;
          }

          .h-step-content {
            margin-top: 0;
            margin-left: 10px;
            text-align: left;
          }
        }
      }
    }
  }

  .robot-overview-card {
    height: 100%;
    display: flex;
    flex-direction: column;

    :deep(.n-card__content) {
      padding: 16px !important;
      flex: 1;
      display: flex;
      flex-direction: column;
    }

    .robot-overview-content {
      display: flex;
      flex-direction: row;
      gap: 16px;
      align-items: stretch;
      flex: 1;

      .robot-animation-container {
        flex: 0 0 340px;
        display: flex;
  justify-content: center;
        align-items: center;
        padding: 0;
        overflow: hidden;
        background: #fafafa;
        border: 1px solid #e5e7eb;
        border-radius: 6px;
        
        :deep(.toogo-container) {
          -webkit-font-smoothing: antialiased;
          -moz-osx-font-smoothing: grayscale;
          text-rendering: optimizeLegibility;
          
          header {
            display: none;
          }

          .scene {
            height: 100%;
            width: 340px;
            margin: 0;
          }

          .chat-bubble {
            font-size: 12px;
            padding: 7px 12px;
            top: 15px;
            right: 25px;
            font-weight: 400;
            letter-spacing: 0.2px;
            line-height: 1.5;
            color: #2080f0;
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Microsoft YaHei', 'PingFang SC', sans-serif;
            -webkit-font-smoothing: antialiased;
            -moz-osx-font-smoothing: grayscale;
          }

          #toogo-robot {
            width: 300px;
            height: 300px;
            
            .logo-text {
              font-size: 14px !important;
              font-weight: 500;
              letter-spacing: 0.3px;
              -webkit-font-smoothing: antialiased;
              -moz-osx-font-smoothing: grayscale;
            }
          }
        }
      }

      .robot-info-panel {
        flex: 1;
    display: flex;
        flex-direction: column;
        gap: 12px;
        min-width: 0;
    justify-content: space-between;
      }

      .robot-status-distribution {
        display: flex;
        justify-content: space-around;
    align-items: center;
        padding: 16px 12px;
        background: #f8fafb;
        border: 1px solid #e5e7eb;
        border-radius: 6px;
        flex: none;

        .status-item {
          text-align: center;
          flex: 1;

          .status-count {
            font-size: 28px;
            font-weight: bold;
            margin-bottom: 4px;
          }

          .status-label {
            font-size: 12px;
            color: #666;
    }

          &.running {
            .status-count {
              color: #18a058;
            }
          }

          &.paused {
            .status-count {
              color: #f0a020;
            }
          }

          &.idle {
            .status-count {
              color: #999;
            }
          }
        }

        .status-divider {
          width: 1px;
          height: 40px;
          background: linear-gradient(to bottom, transparent, #d0d5dd, transparent);
        }
      }

      .robot-stats-grid {
        display: grid;
        grid-template-columns: repeat(2, 1fr);
        gap: 8px;
        flex: none;

        .stat-mini-item {
          text-align: center;
          padding: 12px 8px;
          background: #fafafa;
          border: 1px solid #e5e7eb;
          border-radius: 6px;
          transition: all 0.2s;
      display: flex;
          flex-direction: column;
          justify-content: center;

          &:hover {
            background: #f5f5f5;
          }

          .stat-label {
            font-size: 11px;
            color: #999;
            margin-bottom: 4px;
            white-space: nowrap;
          }

          .stat-value {
            font-size: 18px;
            font-weight: 700;
            color: #333;

            &.profit { color: #18a058; }
            &.loss { color: #d03050; }
            &.warning { color: #f59e0b; }
          }
        }
      }

      // 响应式设计：小屏幕改为垂直布局
      @media (max-width: 1024px) {
        flex-direction: column;

        .robot-animation-container {
          flex: 0 0 auto;
          height: 200px;
          
          :deep(.toogo-container) {
            .scene {
              width: 280px;
            }
            
            #toogo-robot {
              width: 240px;
              height: 240px;
            }
            
            .chat-bubble {
              font-size: 11px;
              padding: 6px 10px;
            }
          }
        }
      }
    }

    // 机器人总览卡片内容区域内边距调整
    :deep(.n-card__content) {
      padding: 16px 18px !important;
    }
    }

  .stat-card {
    border-radius: 12px;
    transition: transform 0.2s;
    &:hover { transform: translateY(-4px); }

    .stat-inner {
      display: flex;
      gap: 16px;
      align-items: center;
      .icon-wrap {
        width: 48px;
        height: 48px;
        border-radius: 12px;
        display: flex;
        align-items: center;
        justify-content: center;
        font-size: 24px;
        color: #fff;
      }
      .info {
        .label { color: #999; font-size: 13px; }
        .value { font-size: 20px; font-weight: bold; margin-top: 2px; }
      }
    }
    .actions {
      margin-top: 16px;
      padding-top: 12px;
      border-top: 1px solid #f0f0f0;
      display: flex;
      gap: 16px;
    }

    &.balance .icon-wrap { background: #3b82f6; }
    &.power .icon-wrap { background: #f59e0b; }
    &.gift .icon-wrap { background: #06b6d4; }
    &.pnl .icon-wrap { background: #10b981; }
  }

  .invite-section {
    .section-header {
      display: flex;
      align-items: center;
      gap: 8px;
      margin-bottom: 12px;
      font-weight: 600;
    }
    .code-display {
      display: flex;
      align-items: center;
      gap: 8px;
      padding: 12px;
      background: #f8fafc;
      border-radius: 8px;
    }
  }
}
</style>


