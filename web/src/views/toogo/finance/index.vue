<template>
  <div class="toogo-finance">
    <!-- 账户概览 -->
    <n-grid :cols="4" :x-gap="16" class="account-cards">
      <n-gi>
        <n-card class="account-card balance-card">
          <div class="card-content">
            <div class="card-icon">
              <n-icon size="32"><WalletOutlined /></n-icon>
            </div>
            <div class="card-info">
              <div class="card-label">余额账户</div>
              <div class="card-value"
                >{{ formatNumber(accountInfo.balance) }} <span class="unit">USDT</span></div
              >
              <div class="card-actions">
                <n-button size="small" type="primary" @click="showRecharge = true">充值</n-button>
                <n-button size="small" @click="handleWithdraw('balance')">提现</n-button>
              </div>
            </div>
          </div>
        </n-card>
      </n-gi>
      <n-gi>
        <n-card class="account-card power-card">
          <div class="card-content">
            <div class="card-icon">
              <n-icon size="32"><ThunderboltOutlined /></n-icon>
            </div>
            <div class="card-info">
              <div class="card-label">算力账户</div>
              <div class="card-value"
                >{{ formatNumber(accountInfo.power) }} <span class="unit">算力</span></div
              >
            </div>
          </div>
        </n-card>
      </n-gi>
      <n-gi>
        <n-card class="account-card gift-card">
          <div class="card-content">
            <div class="card-icon">
              <n-icon size="32"><GiftOutlined /></n-icon>
            </div>
            <div class="card-info">
              <div class="card-label">积分账户</div>
              <div class="card-value"
                >{{ formatNumber(accountInfo.giftPower) }} <span class="unit">积分</span></div
              >
              <div class="card-desc">可抵扣订阅费用</div>
            </div>
          </div>
        </n-card>
      </n-gi>
      <n-gi>
        <n-card class="account-card commission-card">
          <div class="card-content">
            <div class="card-icon">
              <n-icon size="32"><DollarOutlined /></n-icon>
            </div>
            <div class="card-info">
              <div class="card-label">佣金账户</div>
              <div class="card-value"
                >{{ formatNumber(accountInfo.commission) }} <span class="unit">USDT</span></div
              >
              <div class="card-actions">
                <n-button size="small" @click="handleWithdraw('commission')">提现</n-button>
                <n-button size="small" type="info" @click="handleTransfer">转算力</n-button>
              </div>
            </div>
          </div>
        </n-card>
      </n-gi>
    </n-grid>

    <!-- 交易记录 -->
    <n-card title="交易记录" :bordered="false" class="proCard" style="margin-top: 16px">
      <template #header-extra>
        <n-space>
          <n-select
            v-model:value="queryType"
            :options="typeOptions"
            placeholder="交易类型"
            clearable
            style="width: 150px"
          />
          <n-date-picker
            v-model:value="dateRange"
            type="daterange"
            clearable
            style="width: 280px"
          />
          <n-button type="primary" @click="loadRecords">查询</n-button>
        </n-space>
      </template>

      <n-tabs type="line" animated v-model:value="activeTab">
        <n-tab-pane name="all" tab="全部">
          <RecordTable :data="records" :loading="loading" />
        </n-tab-pane>
        <n-tab-pane name="recharge" tab="充值">
          <RecordTable :data="rechargeRecords" :loading="loading" type="recharge" />
        </n-tab-pane>
        <n-tab-pane name="withdraw" tab="提现">
          <RecordTable :data="withdrawRecords" :loading="loading" type="withdraw" />
        </n-tab-pane>
        <n-tab-pane name="subscribe" tab="订阅消费">
          <RecordTable :data="subscribeRecords" :loading="loading" type="subscribe" />
        </n-tab-pane>
        <n-tab-pane name="transfer" tab="转账">
          <RecordTable :data="transferRecords" :loading="loading" type="transfer" />
        </n-tab-pane>
        <n-tab-pane name="power" tab="算力消耗">
          <RecordTable :data="powerRecords" :loading="loading" type="power" />
        </n-tab-pane>
      </n-tabs>
    </n-card>

    <!-- 充值弹窗 -->
    <n-modal
      v-model:show="showRecharge"
      preset="dialog"
      title="充值"
      positive-text="确认充值"
      negative-text="取消"
      @positive-click="handleRechargeSubmit"
      style="width: 500px"
    >
      <div class="recharge-modal">
        <n-form :model="rechargeForm" label-placement="left" label-width="80">
          <n-form-item label="充值金额">
            <n-input-number
              v-model:value="rechargeForm.amount"
              :min="10"
              :precision="2"
              placeholder="最低10 USDT"
              style="width: 100%"
            />
          </n-form-item>
          <n-form-item label="支付方式">
            <n-radio-group v-model:value="rechargeForm.payType">
              <n-space>
                <n-radio value="usdt_trc20">
                  <div class="pay-option">
                    <span>USDT (TRC20)</span>
                  </div>
                </n-radio>
                <n-radio value="usdt_erc20">
                  <div class="pay-option">
                    <span>USDT (ERC20)</span>
                  </div>
                </n-radio>
              </n-space>
            </n-radio-group>
          </n-form-item>
        </n-form>
        <n-alert type="info" title="充值说明">
          <ul style="margin: 0; padding-left: 16px">
            <li>最低充值金额：10 USDT</li>
            <li>充值完成后请耐心等待区块确认</li>
            <li>请勿向充值地址转入其他币种</li>
          </ul>
        </n-alert>
      </div>
    </n-modal>

    <!-- 提现弹窗 -->
    <n-modal
      v-model:show="showWithdraw"
      preset="dialog"
      title="提现"
      positive-text="确认提现"
      negative-text="取消"
      @positive-click="handleWithdrawSubmit"
      style="width: 500px"
    >
      <div class="withdraw-modal">
        <n-form :model="withdrawForm" label-placement="left" label-width="80">
          <n-form-item label="提现账户">
            <n-tag :type="withdrawForm.fromAccount === 'balance' ? 'primary' : 'success'">
              {{ withdrawForm.fromAccount === 'balance' ? '余额账户' : '佣金账户' }}
            </n-tag>
            <span style="margin-left: 10px; color: #999">
              可提现:
              {{
                formatNumber(
                  withdrawForm.fromAccount === 'balance'
                    ? accountInfo.balance
                    : accountInfo.commission,
                )
              }}
              USDT
            </span>
          </n-form-item>
          <n-form-item label="提现金额">
            <n-input-number
              v-model:value="withdrawForm.amount"
              :min="10"
              :max="
                withdrawForm.fromAccount === 'balance'
                  ? accountInfo.balance
                  : accountInfo.commission
              "
              :precision="2"
              placeholder="最低10 USDT"
              style="width: 100%"
            />
          </n-form-item>
          <n-form-item label="提现地址">
            <n-input v-model:value="withdrawForm.address" placeholder="请输入USDT钱包地址" />
          </n-form-item>
          <n-form-item label="网络类型">
            <n-select
              v-model:value="withdrawForm.network"
              :options="[
                { label: 'TRC20', value: 'trc20' },
                { label: 'ERC20', value: 'erc20' },
              ]"
              placeholder="选择网络"
            />
          </n-form-item>
        </n-form>
        <n-alert type="warning" title="提现须知">
          <ul style="margin: 0; padding-left: 16px">
            <li>最低提现金额：10 USDT</li>
            <li>提现手续费：{{ accountInfo.withdrawFee }}%</li>
            <li
              >实际到账：{{
                formatNumber(withdrawForm.amount * (1 - accountInfo.withdrawFee / 100))
              }}
              USDT</li
            >
          </ul>
        </n-alert>
      </div>
    </n-modal>

    <!-- 转账弹窗 -->
    <n-modal
      v-model:show="showTransfer"
      preset="dialog"
      title="转入算力"
      positive-text="确认转账"
      negative-text="取消"
      @positive-click="handleTransferSubmit"
      style="width: 500px"
    >
      <div class="transfer-modal">
        <n-form :model="transferForm" label-placement="left" label-width="80">
          <n-form-item label="转出账户">
            <n-radio-group v-model:value="transferForm.fromAccount">
              <n-space>
                <n-radio value="balance"
                  >余额账户 ({{ formatNumber(accountInfo.balance) }} USDT)</n-radio
                >
                <n-radio value="commission"
                  >佣金账户 ({{ formatNumber(accountInfo.commission) }} USDT)</n-radio
                >
              </n-space>
            </n-radio-group>
          </n-form-item>
          <n-form-item label="转账金额">
            <n-input-number
              v-model:value="transferForm.amount"
              :min="1"
              :precision="2"
              placeholder="请输入转账金额"
              style="width: 100%"
            />
          </n-form-item>
          <n-form-item label="获得算力">
            <n-input-number :value="transferForm.amount" disabled style="width: 100%" />
            <span style="margin-left: 8px; color: #999">1 USDT = 1 算力</span>
          </n-form-item>
        </n-form>
        <n-alert type="info">
          转入算力后用于云机器人消耗（按盈利金额比例扣除），转账无手续费。
        </n-alert>
      </div>
    </n-modal>
  </div>
</template>

<script lang="ts" setup>
  import { ref, computed, onMounted, h, defineComponent } from 'vue';
  import { useMessage } from 'naive-ui';
  import { WalletOutlined, ThunderboltOutlined, GiftOutlined, DollarOutlined } from '@vicons/antd';
  import { NTag } from 'naive-ui';
  import { http } from '@/utils/http/axios';
  import { ToogoWalletApi } from '@/api/toogo';

  const message = useMessage();

  // 账户信息
  const accountInfo = ref({
    balance: 0,
    power: 0,
    giftPower: 0,
    commission: 0,
    withdrawFee: 2, // 提现手续费%
  });

  // 交易记录
  const loading = ref(false);
  const records = ref<any[]>([]);
  const activeTab = ref('all');
  const queryType = ref(null);
  const dateRange = ref<[number, number] | null>(null);

  // 弹窗状态
  const showRecharge = ref(false);
  const showWithdraw = ref(false);
  const showTransfer = ref(false);

  // 表单数据
  const rechargeForm = ref({
    amount: 100,
    payType: 'usdt_trc20',
  });

  const withdrawForm = ref({
    fromAccount: 'balance',
    amount: 0,
    address: '',
    network: 'trc20',
  });

  const transferForm = ref({
    fromAccount: 'balance',
    amount: 0,
  });

  // 类型选项
  const typeOptions = [
    { label: '全部', value: null },
    { label: '充值', value: 'recharge' },
    { label: '提现', value: 'withdraw' },
    { label: '订阅消费', value: 'subscribe' },
    { label: '转账', value: 'transfer' },
    { label: '算力消耗', value: 'power' },
  ];

  // 过滤后的记录（根据changeType字段分类）
  const rechargeRecords = computed(() =>
    records.value.filter((r) => r.changeType === 'deposit' || r.changeType === 'admin_recharge'),
  );
  const withdrawRecords = computed(() =>
    records.value.filter((r) => r.changeType?.startsWith('withdraw')),
  );
  const subscribeRecords = computed(() =>
    records.value.filter(
      (r) => r.changeType === 'subscribe' || r.changeType === 'subscribe_deduct',
    ),
  );
  const transferRecords = computed(() =>
    records.value.filter((r) => r.changeType?.startsWith('transfer')),
  );
  const powerRecords = computed(() =>
    records.value.filter((r) => r.changeType === 'power_consume'),
  );

  // 格式化数字
  function toNumber(val: any, fallback = 0) {
    const n = Number(val);
    return Number.isFinite(n) ? n : fallback;
  }

  function formatNumber(num: any, precision = 2) {
    return toNumber(num, 0).toFixed(precision);
  }

  // 加载账户信息
  async function loadAccountInfo() {
    try {
      // 与控制台页面保持一致：接口返回已是业务数据对象（非 {code,data} 包裹）
      const walletRes: any = await ToogoWalletApi.overview();
      if (walletRes) {
        accountInfo.value = {
          ...accountInfo.value,
          balance: toNumber(walletRes.balance, accountInfo.value.balance),
          power: toNumber(walletRes.power, accountInfo.value.power),
          giftPower: toNumber(walletRes.giftPower, accountInfo.value.giftPower),
          commission: toNumber(walletRes.commission, accountInfo.value.commission),
          withdrawFee: toNumber(walletRes.withdrawFee, accountInfo.value.withdrawFee),
        };
      }
    } catch (error) {
      console.error('加载账户信息失败', error);
    }
  }

  // 加载交易记录
  async function loadRecords() {
    loading.value = true;
    try {
      const res = await http.request({
        url: '/toogo/wallet/log/list',
        method: 'get',
        params: {
          type: queryType.value,
          startTime: dateRange.value?.[0],
          endTime: dateRange.value?.[1],
        },
      });
      if (res.code === 0) {
        records.value = res.data?.list || [];
      }
    } catch (error) {
      console.error('加载记录失败', error);
    } finally {
      loading.value = false;
    }
  }

  // 提现
  function handleWithdraw(account: string) {
    withdrawForm.value.fromAccount = account;
    withdrawForm.value.amount = 0;
    showWithdraw.value = true;
  }

  // 转账
  function handleTransfer() {
    transferForm.value.amount = 0;
    showTransfer.value = true;
  }

  // 提交充值
  async function handleRechargeSubmit() {
    if (rechargeForm.value.amount < 10) {
      message.error('最低充值金额为10 USDT');
      return false;
    }
    try {
      const res = await http.request({
        url: '/toogo/finance/recharge',
        method: 'post',
        data: rechargeForm.value,
      });
      if (res.code === 0 && res.data?.payUrl) {
        window.open(res.data.payUrl, '_blank');
        message.success('请在新窗口完成支付');
      } else {
        message.error(res.message || '创建充值订单失败');
      }
    } catch (error: any) {
      message.error(error.message || '创建充值订单失败');
    }
    return false;
  }

  // 提交提现
  async function handleWithdrawSubmit() {
    if (withdrawForm.value.amount < 10) {
      message.error('最低提现金额为10 USDT');
      return false;
    }
    if (!withdrawForm.value.address) {
      message.error('请输入提现地址');
      return false;
    }
    try {
      const res = await http.request({
        url: '/toogo/finance/withdraw',
        method: 'post',
        data: withdrawForm.value,
      });
      if (res.code === 0) {
        message.success('提现申请已提交，请等待审核');
        showWithdraw.value = false;
        loadAccountInfo();
        loadRecords();
      } else {
        message.error(res.message || '提现失败');
      }
    } catch (error: any) {
      message.error(error.message || '提现失败');
    }
    return false;
  }

  // 提交转账
  async function handleTransferSubmit() {
    if (transferForm.value.amount <= 0) {
      message.error('请输入转账金额');
      return false;
    }
    try {
      const res = await http.request({
        url: '/toogo/wallet/transfer',
        method: 'post',
        data: {
          fromAccount: transferForm.value.fromAccount,
          toAccount: 'power',
          amount: transferForm.value.amount,
        },
      });
      if (res.code === 0) {
        message.success('转账成功');
        showTransfer.value = false;
        loadAccountInfo();
        loadRecords();
      } else {
        message.error(res.message || '转账失败');
      }
    } catch (error: any) {
      message.error(error.message || '转账失败');
    }
    return false;
  }

  // 交易记录表格组件
  const RecordTable = defineComponent({
    props: {
      data: { type: Array, default: () => [] },
      loading: { type: Boolean, default: false },
      type: { type: String, default: '' },
    },
    setup(props) {
      // 变动类型映射
      const getChangeTypeInfo = (changeType: string) => {
        const types: Record<string, { label: string; type: any }> = {
          deposit: { label: '充值', type: 'success' },
          admin_recharge: { label: '管理员充值', type: 'success' },
          withdraw: { label: '提现申请', type: 'warning' },
          withdraw_reject: { label: '提现拒绝退回', type: 'info' },
          withdraw_complete: { label: '提现完成', type: 'success' },
          withdraw_fail: { label: '提现失败退回', type: 'warning' },
          transfer_out: { label: '转出', type: 'warning' },
          transfer_in: { label: '转入', type: 'success' },
          power_consume: { label: '算力消耗', type: 'error' },
          invite_reward: { label: '邀请奖励', type: 'success' },
          invited_reward: { label: '注册奖励', type: 'success' },
          subscribe: { label: '订阅套餐', type: 'warning' },
          subscribe_deduct: { label: '订阅积分抵扣', type: 'info' },
        };
        return types[changeType] || { label: changeType, type: 'default' };
      };

      // 账户类型映射
      const getAccountTypeLabel = (accountType: string) => {
        const accounts: Record<string, string> = {
          balance: '余额账户',
          power: '算力账户',
          gift_power: '积分',
          commission: '佣金账户',
        };
        return accounts[accountType] || accountType;
      };

      const columns = [
        { title: '时间', key: 'createdAt', width: 180 },
        {
          title: '类型',
          key: 'changeType',
          width: 120,
          render: (row: any) => {
            const info = getChangeTypeInfo(row.changeType);
            return h(NTag, { type: info.type, size: 'small' }, () => info.label);
          },
        },
        {
          title: '变动金额',
          key: 'changeAmount',
          width: 120,
          render: (row: any) => {
            const amount = row.changeAmount || 0;
            const isPositive = amount >= 0;
            return h(
              'span',
              { style: { color: isPositive ? '#18a058' : '#d03050', fontWeight: 'bold' } },
              `${isPositive ? '+' : ''}${amount.toFixed(4)}`,
            );
          },
        },
        {
          title: '账户',
          key: 'accountType',
          width: 100,
          render: (row: any) => getAccountTypeLabel(row.accountType),
        },
        {
          title: '变动前',
          key: 'beforeAmount',
          width: 100,
          render: (row: any) => (row.beforeAmount || 0).toFixed(2),
        },
        {
          title: '变动后',
          key: 'afterAmount',
          width: 100,
          render: (row: any) => (row.afterAmount || 0).toFixed(2),
        },
        { title: '备注', key: 'remark', ellipsis: true },
      ];

      return () =>
        h('div', [
          h('n-data-table', {
            columns,
            data: props.data,
            loading: props.loading,
            pagination: { pageSize: 10 },
            size: 'small',
          }),
        ]);
    },
  });

  onMounted(() => {
    loadAccountInfo();
    loadRecords();
  });
</script>

<style scoped lang="less">
  .toogo-finance {
    .account-cards {
      margin-bottom: 16px;
    }

    .account-card {
      height: 140px;
      border-radius: 8px;

      .card-content {
        display: flex;
        align-items: center;
        gap: 16px;
      }

      .card-icon {
        width: 64px;
        height: 64px;
        border-radius: 12px;
        display: flex;
        align-items: center;
        justify-content: center;
        color: #fff;
      }

      .card-info {
        flex: 1;
      }

      .card-label {
        font-size: 14px;
        color: #666;
        margin-bottom: 4px;
      }

      .card-value {
        font-size: 24px;
        font-weight: 600;
        color: #333;

        .unit {
          font-size: 14px;
          font-weight: normal;
          color: #999;
        }
      }

      .card-desc {
        font-size: 12px;
        color: #999;
        margin-top: 4px;
      }

      .card-actions {
        margin-top: 8px;
        display: flex;
        gap: 8px;
      }

      &.balance-card .card-icon {
        background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
      }

      &.power-card .card-icon {
        background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
      }

      &.gift-card .card-icon {
        background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
      }

      &.commission-card .card-icon {
        background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%);
      }
    }

    .recharge-modal,
    .withdraw-modal,
    .transfer-modal {
      .pay-option {
        display: flex;
        align-items: center;
        gap: 8px;
      }
    }
  }
</style>
