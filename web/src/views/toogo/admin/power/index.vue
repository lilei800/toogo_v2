<template>
  <div class="toogo-admin-power">
    <n-tabs type="line" animated>
      <!-- Tab 1: 用户钱包管理 & 充值 -->
      <n-tab-pane name="wallet" tab="💰 用户钱包管理">
        <n-card :bordered="false">
          <!-- 搜索栏 -->
          <n-space style="margin-bottom: 16px">
            <n-input v-model:value="walletSearch.username" placeholder="用户名搜索" clearable style="width: 160px" />
            <n-input v-model:value="walletSearch.mobile" placeholder="手机号搜索" clearable style="width: 160px" />
            <n-button type="primary" @click="loadWalletList">
              <template #icon><n-icon><SearchOutlined /></n-icon></template>
              搜索
            </n-button>
            <n-button @click="resetWalletSearch">重置</n-button>
          </n-space>

          <!-- 用户钱包表格 -->
          <n-data-table
            :columns="walletColumns"
            :data="walletList"
            :loading="walletLoading"
            :pagination="walletPagination"
            :row-key="(row: any) => row.userId"
            striped
            @update:page="handleWalletPageChange"
          />
        </n-card>
      </n-tab-pane>

      <!-- Tab 2: 算力配置 -->
      <n-tab-pane name="config" tab="⚙️ 算力配置">
        <n-card :bordered="false">
          <n-form
            ref="formRef"
            :model="config"
            label-placement="left"
            label-width="180"
            style="max-width: 800px"
          >
            <n-divider title-placement="left">基础配置</n-divider>

            <n-form-item label="算力消耗比例">
              <n-input-number
                v-model:value="config.powerRatio"
                :min="1"
                :max="50"
                :precision="1"
                style="width: 200px"
              >
                <template #suffix>%</template>
              </n-input-number>
              <span style="margin-left: 16px; color: #999;">
                云机器人消耗算力 = 盈利金额 × {{ config.powerRatio }}%
              </span>
            </n-form-item>

            <n-form-item label="USDT/算力兑换比例">
              <n-input-number
                v-model:value="config.usdtToPowerRatio"
                :min="0.1"
                :max="10"
                :precision="2"
                style="width: 200px"
              />
              <span style="margin-left: 16px; color: #999;">
                1 USDT = {{ config.usdtToPowerRatio }} 算力
              </span>
            </n-form-item>

            <n-divider title-placement="left">VIP等级算力优惠</n-divider>

            <n-form-item v-for="(level, index) in config.vipDiscounts" :key="index" :label="`V${index + 1} 优惠比例`">
              <n-input-number
                v-model:value="config.vipDiscounts[index]"
                :min="0"
                :max="50"
                :precision="0"
                style="width: 200px"
              >
                <template #suffix>%</template>
              </n-input-number>
              <span style="margin-left: 16px; color: #999;">
                V{{ index + 1 }}用户算力消耗减少 {{ config.vipDiscounts[index] }}%
              </span>
            </n-form-item>

            <n-divider title-placement="left">推广奖励</n-divider>

            <n-form-item label="邀请注册赠送积分">
              <n-input-number
                v-model:value="config.inviteReward"
                :min="0"
                :max="1000"
                :precision="0"
                style="width: 200px"
              />
              <span style="margin-left: 16px; color: #999;">
                邀请人和被邀请人各获得 {{ config.inviteReward }} 算力
              </span>
            </n-form-item>

            <n-form-item label="订阅奖励算力比例">
              <n-input-number
                v-model:value="config.subscriptionRewardRatio"
                :min="0"
                :max="100"
                :precision="1"
                style="width: 200px"
              >
                <template #suffix>%</template>
              </n-input-number>
              <span style="margin-left: 16px; color: #999;">
                被邀请人订阅时，邀请人获得订阅额 × {{ config.subscriptionRewardRatio }}% 的算力
              </span>
            </n-form-item>

            <n-divider title-placement="left">代理商佣金</n-divider>

            <n-form-item label="订阅佣金比例">
              <n-input-number
                v-model:value="config.agentSubscriptionCommission"
                :min="0"
                :max="50"
                :precision="1"
                style="width: 200px"
              >
                <template #suffix>%</template>
              </n-input-number>
              <span style="margin-left: 16px; color: #999;">
                下级用户订阅时，代理商获得订阅额 × {{ config.agentSubscriptionCommission }}% 的佣金
              </span>
            </n-form-item>

            <n-form-item label="算力消耗佣金比例">
              <n-input-number
                v-model:value="config.agentPowerCommission"
                :min="0"
                :max="50"
                :precision="1"
                style="width: 200px"
              >
                <template #suffix>%</template>
              </n-input-number>
              <span style="margin-left: 16px; color: #999;">
                下级用户消耗算力时，代理商获得消耗额 × {{ config.agentPowerCommission }}% 的佣金
              </span>
            </n-form-item>

            <n-divider title-placement="left">提现配置</n-divider>

            <n-form-item label="提现手续费">
              <n-input-number
                v-model:value="config.withdrawFee"
                :min="0"
                :max="20"
                :precision="1"
                style="width: 200px"
              >
                <template #suffix>%</template>
              </n-input-number>
            </n-form-item>

            <n-form-item label="最低提现金额">
              <n-input-number
                v-model:value="config.minWithdraw"
                :min="1"
                :max="1000"
                :precision="0"
                style="width: 200px"
              >
                <template #suffix>USDT</template>
              </n-input-number>
            </n-form-item>

            <n-form-item>
              <n-space>
                <n-button type="primary" :loading="saving" @click="handleSave">保存配置</n-button>
                <n-button @click="loadConfig">重置</n-button>
              </n-space>
            </n-form-item>
          </n-form>
        </n-card>
      </n-tab-pane>

      <!-- Tab 3: 算力统计 -->
      <n-tab-pane name="stats" tab="📊 算力统计">
        <n-card :bordered="false">
          <n-grid :cols="4" :x-gap="16">
            <n-gi>
              <n-statistic label="今日消耗算力">
                <n-number-animation :from="0" :to="powerStats.todayUsed" :precision="2" />
              </n-statistic>
            </n-gi>
            <n-gi>
              <n-statistic label="今日充值算力">
                <n-number-animation :from="0" :to="powerStats.todayRecharge" :precision="2" />
              </n-statistic>
            </n-gi>
            <n-gi>
              <n-statistic label="今日赠送积分">
                <n-number-animation :from="0" :to="powerStats.todayGift" :precision="2" />
              </n-statistic>
            </n-gi>
            <n-gi>
              <n-statistic label="用户总算力余额">
                <n-number-animation :from="0" :to="powerStats.totalBalance" :precision="2" />
              </n-statistic>
            </n-gi>
          </n-grid>
        </n-card>
      </n-tab-pane>
    </n-tabs>

    <!-- 充值弹窗 -->
    <n-modal v-model:show="showRechargeModal" preset="card" title="💰 手动充值" style="width: 500px">
      <n-form ref="rechargeFormRef" :model="rechargeForm" :rules="rechargeRules" label-placement="left" label-width="100">
        <n-form-item label="用户">
          <n-text>{{ rechargeForm.username }} (ID: {{ rechargeForm.userId }})</n-text>
        </n-form-item>
        <n-form-item label="当前余额">
          <n-space>
            <n-tag type="primary">余额: {{ rechargeForm.currentBalance?.toFixed(2) || 0 }} U</n-tag>
            <n-tag type="success">算力: {{ rechargeForm.currentPower?.toFixed(2) || 0 }}</n-tag>
            <n-tag type="warning">积分: {{ rechargeForm.currentGiftPower?.toFixed(2) || 0 }}</n-tag>
          </n-space>
        </n-form-item>
        <n-form-item label="充值类型" path="accountType">
          <n-radio-group v-model:value="rechargeForm.accountType">
            <n-radio-button value="power">算力</n-radio-button>
            <n-radio-button value="gift_power">积分</n-radio-button>
            <n-radio-button value="balance">余额(USDT)</n-radio-button>
          </n-radio-group>
        </n-form-item>
        <n-form-item label="充值金额" path="amount">
          <n-input-number v-model:value="rechargeForm.amount" :min="0.01" :precision="2" style="width: 100%">
            <template #suffix>{{ rechargeForm.accountType === 'balance' ? 'USDT' : '算力' }}</template>
          </n-input-number>
        </n-form-item>
        <n-form-item label="备注">
          <n-input v-model:value="rechargeForm.remark" type="textarea" :rows="2" placeholder="充值原因/备注" />
        </n-form-item>
      </n-form>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showRechargeModal = false">取消</n-button>
          <n-button type="primary" :loading="rechargeLoading" @click="handleRecharge">确认充值</n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<script lang="ts" setup>
import { ref, h, onMounted } from 'vue';
import { useMessage, NButton, NSpace, NTag } from 'naive-ui';
import { SearchOutlined } from '@vicons/antd';
import { http } from '@/utils/http/axios';

const message = useMessage();

// ==================== 用户钱包管理 ====================
const walletSearch = ref({
  username: '',
  mobile: '',
});
const walletList = ref<any[]>([]);
const walletLoading = ref(false);
const walletPagination = ref({
  page: 1,
  pageSize: 15,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 15, 20, 50],
});

// 表格列定义
const walletColumns = [
  { title: 'ID', key: 'userId', width: 80 },
  { title: '用户名', key: 'username', width: 120 },
  { title: '手机号', key: 'mobile', width: 130 },
  {
    title: '余额(USDT)',
    key: 'balance',
    width: 110,
    render: (row: any) => h('span', { style: 'color: #2080f0; font-weight: 600' }, row.balance?.toFixed(2) || '0.00'),
  },
  {
    title: '算力',
    key: 'power',
    width: 100,
    render: (row: any) => h('span', { style: 'color: #18a058; font-weight: 600' }, row.power?.toFixed(2) || '0.00'),
  },
  {
    title: '积分',
    key: 'giftPower',
    width: 100,
    render: (row: any) => h('span', { style: 'color: #f0a020; font-weight: 600' }, row.giftPower?.toFixed(2) || '0.00'),
  },
  {
    title: '总算力',
    key: 'totalPower',
    width: 100,
    render: (row: any) => h(NTag, { type: 'success', size: 'small' }, () => row.totalPower?.toFixed(2) || '0.00'),
  },
  {
    title: 'VIP',
    key: 'vipLevel',
    width: 70,
    render: (row: any) => row.vipLevel > 0 ? h(NTag, { type: 'warning', size: 'small' }, () => `V${row.vipLevel}`) : '--',
  },
  {
    title: '操作',
    key: 'actions',
    width: 120,
    render: (row: any) => h(NSpace, { size: 'small' }, () => [
      h(NButton, { type: 'primary', size: 'small', onClick: () => openRechargeModal(row) }, () => '充值'),
    ]),
  },
];

// 加载用户钱包列表
async function loadWalletList() {
  walletLoading.value = true;
  try {
    const res = await http.request({
      url: '/toogo/wallet/userList',
      method: 'get',
      params: {
        username: walletSearch.value.username,
        mobile: walletSearch.value.mobile,
        page: walletPagination.value.page,
        perPage: walletPagination.value.pageSize,
      },
    });
    if (res.code === 0) {
      walletList.value = res.data?.list || [];
      walletPagination.value.itemCount = res.data?.totalCount || 0;
    }
  } catch (error) {
    console.error('加载用户钱包列表失败', error);
  } finally {
    walletLoading.value = false;
  }
}

function resetWalletSearch() {
  walletSearch.value = { username: '', mobile: '' };
  walletPagination.value.page = 1;
  loadWalletList();
}

function handleWalletPageChange(page: number) {
  walletPagination.value.page = page;
  loadWalletList();
}

// ==================== 充值功能 ====================
const showRechargeModal = ref(false);
const rechargeLoading = ref(false);
const rechargeFormRef = ref<any>(null);
const rechargeForm = ref({
  userId: 0,
  username: '',
  currentBalance: 0,
  currentPower: 0,
  currentGiftPower: 0,
  accountType: 'power',
  amount: 100,
  remark: '',
});

const rechargeRules = {
  accountType: { required: true, message: '请选择充值类型', trigger: 'change' },
  amount: { required: true, type: 'number', min: 0.01, message: '请输入充值金额', trigger: 'blur' },
};

function openRechargeModal(row: any) {
  rechargeForm.value = {
    userId: row.userId,
    username: row.username || row.mobile || `用户${row.userId}`,
    currentBalance: row.balance || 0,
    currentPower: row.power || 0,
    currentGiftPower: row.giftPower || 0,
    accountType: 'power',
    amount: 100,
    remark: '',
  };
  showRechargeModal.value = true;
}

async function handleRecharge() {
  rechargeLoading.value = true;
  try {
    const res = await http.request({
      url: '/toogo/wallet/adminRecharge',
      method: 'post',
      data: {
        userId: rechargeForm.value.userId,
        accountType: rechargeForm.value.accountType,
        amount: rechargeForm.value.amount,
        remark: rechargeForm.value.remark || '管理员手动充值',
      },
    });
    if (res.code === 0) {
      const typeMap: Record<string, string> = { power: '算力', gift_power: '积分', balance: '余额' };
      message.success(`充值成功！${typeMap[rechargeForm.value.accountType]}：${res.data?.beforeAmount?.toFixed(2)} → ${res.data?.afterAmount?.toFixed(2)}`);
      showRechargeModal.value = false;
      loadWalletList();
    } else {
      message.error(res.message || '充值失败');
    }
  } catch (error: any) {
    message.error(error.message || '充值失败');
  } finally {
    rechargeLoading.value = false;
  }
}

// ==================== 算力配置 ====================
const config = ref({
  powerRatio: 10,
  usdtToPowerRatio: 1,
  vipDiscounts: [5, 10, 15, 20, 22, 24, 26, 28, 29, 30],
  inviteReward: 30,
  subscriptionRewardRatio: 10,
  agentSubscriptionCommission: 20,
  agentPowerCommission: 10,
  withdrawFee: 2,
  minWithdraw: 10,
});

const powerStats = ref({
  todayUsed: 0,
  todayRecharge: 0,
  todayGift: 0,
  totalBalance: 0,
});

const saving = ref(false);
const formRef = ref<any>(null);

async function loadConfig() {
  try {
    const res = await http.request({
      url: '/toogo/admin/power/config',
      method: 'get',
    });
    if (res.code === 0 && res.data) {
      config.value = { ...config.value, ...res.data };
    }
  } catch (error) {
    console.error('加载配置失败', error);
  }
}

async function loadStats() {
  try {
    const res = await http.request({
      url: '/toogo/admin/power/stats',
      method: 'get',
    });
    if (res.code === 0 && res.data) {
      powerStats.value = res.data;
    }
  } catch (error) {
    console.error('加载统计失败', error);
  }
}

async function handleSave() {
  saving.value = true;
  try {
    const res = await http.request({
      url: '/toogo/admin/power/config',
      method: 'post',
      data: config.value,
    });
    if (res.code === 0) {
      message.success('配置保存成功');
    } else {
      message.error(res.message || '保存失败');
    }
  } catch (error: any) {
    message.error(error.message || '保存失败');
  } finally {
    saving.value = false;
  }
}

onMounted(() => {
  loadWalletList();
  loadConfig();
  loadStats();
});
</script>

<style scoped lang="less">
.toogo-admin-power {
  // 样式
}
</style>

