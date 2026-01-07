<template>
  <div class="subscription-page">
    <!-- 账户余额和当前套餐状态 -->
    <n-grid cols="1 s:2 m:2 l:2 xl:2 2xl:2" :x-gap="12" :y-gap="12" class="mb-3">
      <!-- 账户余额 -->
      <n-gi>
        <n-card :bordered="true" size="small">
          <n-statistic label="账户余额" :value="accountBalance.toFixed(2)" suffix="USDT">
            <template #prefix>
              <n-icon :component="WalletOutlined" />
            </template>
          </n-statistic>
          <template #action>
            <n-button size="small" type="primary" @click="$router.push('/toogo/finance')">
              充值余额
            </n-button>
          </template>
        </n-card>
      </n-gi>
      <!-- 当前套餐状态 -->
      <n-gi>
        <n-card v-if="currentSubscription.hasSubscription" :bordered="true" size="small">
          <n-statistic label="当前订阅" :value="currentSubscription.planName">
            <template #suffix>
              <n-tag type="success" size="small" style="margin-left: 8px"
                >剩余 {{ currentSubscription.remainingDays }} 天</n-tag
              >
            </template>
          </n-statistic>
        </n-card>
        <n-card v-else :bordered="true" size="small">
          <n-statistic label="当前订阅" value="无订阅">
            <template #prefix>
              <n-icon :component="RobotOutlined" />
            </template>
          </n-statistic>
          <template #action>
            <n-text depth="3" style="font-size: 12px">请选择套餐开启自动交易</n-text>
          </template>
        </n-card>
      </n-gi>
    </n-grid>

    <!-- 套餐列表 -->
    <n-card title="订阅套餐" :bordered="true" size="small">
      <template #header-extra>
        <n-text depth="3" style="font-size: 13px">选择适合您的套餐，开启自动交易</n-text>
      </template>

      <n-grid cols="1 s:2 m:2 l:3 xl:4 2xl:4" :x-gap="12" :y-gap="12" responsive="screen">
        <n-gi v-for="plan in planList" :key="plan.id">
          <n-card
            class="plan-card"
            :class="{
              'plan-popular': plan.isDefault === 1,
              'plan-selected': selectedPlan?.id === plan.id,
            }"
            hoverable
            size="small"
            :bordered="false"
            @click="selectPlan(plan)"
            style="cursor: pointer"
          >
            <!-- 推荐标签 -->
            <div class="popular-badge" v-if="plan.isDefault === 1">
              <n-tag type="warning" size="small">推荐</n-tag>
            </div>

            <div class="plan-header">
              <h3 class="plan-name">{{ plan.planName }}</h3>
              <div class="plan-robots">
                <n-icon :component="RobotOutlined" size="24" color="#2563eb" />
                <span class="robot-count">{{ plan.robotLimit }}</span>
                <span class="robot-label">个机器人</span>
              </div>
            </div>

            <n-divider style="margin: 10px 0" />

            <!-- 订阅周期选择 -->
            <div class="plan-period-select" @click.stop>
              <n-text depth="3" style="font-size: 12px; display: block; margin-bottom: 6px"
                >选择订阅周期</n-text
              >
              <n-select
                :value="getPlanSelectedPeriod(plan.id)"
                :options="getAvailablePeriodOptionsForPlan(plan)"
                size="small"
                @update:value="(val) => handlePeriodChange(plan.id, val)"
                @click.stop
              />
            </div>

            <n-divider style="margin: 10px 0" />

            <!-- 价格显示和每日费用提示 -->
            <div class="plan-price-display">
              <div class="price-main">
                <n-text strong style="font-size: 22px; color: #2563eb; font-weight: 700">
                  {{ getPlanPrice(plan, getPlanSelectedPeriod(plan.id)) }}
                </n-text>
                <n-text depth="3" style="font-size: 11px; margin-left: 4px">
                  USDT / {{ getPeriodLabel(getPlanSelectedPeriod(plan.id)) }}
                </n-text>
              </div>
              <div class="daily-price-hint">
                <n-text depth="3" style="font-size: 11px">
                  约
                  {{ getDailyPricePerRobot(plan, getPlanSelectedPeriod(plan.id)).toFixed(4) }}
                  USDT/天/机器人
                </n-text>
              </div>
            </div>

            <n-divider style="margin: 10px 0" />

            <div class="plan-action" @click.stop>
              <n-space vertical :size="6">
                <n-tag
                  v-if="selectedPlan?.id === plan.id"
                  type="success"
                  size="medium"
                  style="width: 100%; justify-content: center"
                >
                  ✓ 已选择
                </n-tag>
                <n-button text block size="small" type="primary" @click="showPlanDetail(plan)">
                  查看详情
                </n-button>
              </n-space>
            </div>
          </n-card>
        </n-gi>
      </n-grid>
    </n-card>

    <!-- 订阅操作面板 -->
    <n-card v-if="selectedPlan" class="mt-3 order-panel" :bordered="true" size="small">
      <template #header>
        <n-space align="center" justify="space-between">
          <n-text strong style="font-size: 15px">订单详情</n-text>
          <n-button text type="default" @click="handleCancelSelect">
            <template #icon><n-icon :component="CloseOutlined" /></template>
            取消
          </n-button>
        </n-space>
      </template>

      <n-space vertical :size="16">
        <!-- 订单信息 -->
        <n-descriptions :column="2" label-placement="left" size="small" bordered>
          <n-descriptions-item label="套餐名称">
            <n-text strong>{{ selectedPlan.planName }}</n-text>
          </n-descriptions-item>
          <n-descriptions-item label="订阅周期">
            <n-text strong>{{ getPeriodLabel(selectedPeriod) }}</n-text>
          </n-descriptions-item>
          <n-descriptions-item label="订阅时长"> {{ getPeriodDays() }} 天 </n-descriptions-item>
          <n-descriptions-item label="赠送算力"> {{ getGiftPower() }} 算力 </n-descriptions-item>
        </n-descriptions>

        <!-- 支付方式选择 -->
        <n-card size="small" :bordered="true" title="支付方式">
          <n-radio-group v-model:value="payType" size="small">
            <n-space vertical :size="12">
              <n-radio value="balance" :disabled="!canUseBalance">
                <n-space align="center" :size="12">
                  <n-icon :component="WalletOutlined" />
                  <span>余额支付</span>
                  <n-text depth="3" style="font-size: 12px">
                    (可用余额: {{ accountBalance.toFixed(2) }} USDT)
                  </n-text>
                  <n-tag v-if="!canUseBalance" type="error" size="small">余额不足</n-tag>
                </n-space>
              </n-radio>
              <n-alert
                v-if="!canUseBalance && payType === 'balance'"
                type="warning"
                size="small"
                :show-icon="true"
              >
                余额不足，请先充值或使用其他支付方式
                <template #action>
                  <n-button
                    text
                    type="primary"
                    size="small"
                    @click="$router.push('/toogo/finance')"
                  >
                    去充值
                  </n-button>
                </template>
              </n-alert>
            </n-space>
          </n-radio-group>
        </n-card>

        <!-- 积分抵扣选项 -->
        <n-card v-if="userPoints > 0" size="small" :bordered="true" title="积分抵扣">
          <n-space vertical :size="8">
            <n-checkbox v-model:checked="usePoints">
              使用积分抵扣 (可用: {{ userPoints.toFixed(2) }} 积分)
            </n-checkbox>
            <n-text v-if="usePoints" depth="3" style="font-size: 12px">
              可抵扣 {{ Math.min(userPoints, getSelectedPrice()).toFixed(2) }} 积分
            </n-text>
          </n-space>
        </n-card>

        <!-- 费用明细 -->
        <n-card size="small" :bordered="true" title="费用明细" class="payment-summary">
          <n-space vertical :size="8">
            <n-space align="center" justify="space-between">
              <n-text depth="3">套餐价格</n-text>
              <n-text strong>{{ getSelectedPrice().toFixed(2) }} USDT</n-text>
            </n-space>
            <n-space v-if="usePoints" align="center" justify="space-between">
              <n-text depth="3">积分抵扣</n-text>
              <n-text type="success"
                >-{{ Math.min(userPoints, getSelectedPrice()).toFixed(2) }} 积分</n-text
              >
            </n-space>
            <n-divider style="margin: 8px 0" />
            <n-space align="center" justify="space-between">
              <n-text strong style="font-size: 16px">实付金额</n-text>
              <n-text type="warning" strong style="font-size: 20px">
                {{ getFinalAmount().toFixed(2) }} USDT
              </n-text>
            </n-space>
            <n-alert
              v-if="payType === 'balance' && accountBalance < getFinalAmount()"
              type="error"
              size="small"
              :show-icon="true"
            >
              余额不足，当前余额 {{ accountBalance.toFixed(2) }} USDT，需要
              {{ getFinalAmount().toFixed(2) }} USDT
              <template #action>
                <n-button text type="primary" size="small" @click="$router.push('/toogo/finance')">
                  立即充值
                </n-button>
              </template>
            </n-alert>
          </n-space>
        </n-card>

        <!-- 操作按钮 -->
        <n-space justify="end">
          <n-button @click="handleCancelSelect">取消</n-button>
          <n-button
            type="primary"
            size="large"
            @click="handleSubscribe"
            :loading="subscribeLoading"
            :disabled="payType === 'balance' && accountBalance < getFinalAmount()"
          >
            <template #icon><n-icon :component="CreditCardOutlined" /></template>
            确认订阅
          </n-button>
        </n-space>
      </n-space>
    </n-card>

    <!-- 订阅记录 -->
    <n-card
      title="订阅记录"
      :bordered="false"
      :segmented="{ content: true }"
      size="small"
      class="mt-3"
    >
      <n-data-table
        :columns="subscriptionColumns"
        :data="subscriptionList"
        :loading="loading"
        :pagination="pagination"
        size="small"
      />
    </n-card>
  </div>
</template>

<script setup lang="ts">
  import { ref, computed, onMounted, h } from 'vue';
  import { useMessage } from 'naive-ui';
  import {
    RobotOutlined,
    CheckOutlined,
    GiftOutlined,
    CheckCircleOutlined,
    CreditCardOutlined,
    WalletOutlined,
    InfoCircleOutlined,
    CloseOutlined,
    EyeOutlined,
  } from '@vicons/antd';
  import { ToogoSubscriptionApi, ToogoWalletApi } from '@/api/toogo';
  import { NTag, useDialog } from 'naive-ui';

  const message = useMessage();
  const dialog = useDialog();

  const planList = ref<any[]>([]);
  const subscriptionList = ref<any[]>([]);
  const currentSubscription = ref<any>({});
  const loading = ref(false);
  const subscribeLoading = ref(false);
  const selectedPlan = ref<any>(null);
  const selectedPeriod = ref(''); // 选择的周期
  const selectedPlanPeriod = ref<{ id: number; period: string } | null>(null); // 每个套餐选择的周期
  const planPeriodMap = ref<Map<number, string>>(new Map()); // 存储每个套餐选择的周期
  const userPoints = ref(0); // 用户积分余额
  const usePoints = ref(false); // 是否使用积分抵扣
  const accountBalance = ref(0); // 账户余额
  const payType = ref('balance'); // 支付方式：balance-余额支付
  const isFirstLoad = ref(true); // 是否是首次加载

  const pricingPeriods = [
    { key: 'daily', label: '日付', discount: '' },
    { key: 'monthly', label: '月付', discount: '' },
    { key: 'quarterly', label: '季付', discount: '9折' },
    { key: 'half_year', label: '半年付', discount: '8.5折' },
    { key: 'yearly', label: '年付', discount: '最优惠' },
  ];

  const periodOptions = [
    { label: '日付', value: 'daily' },
    { label: '月付', value: 'monthly' },
    { label: '季付', value: 'quarterly' },
    { label: '半年付', value: 'half_year' },
    { label: '年付', value: 'yearly' },
  ];

  const pagination = ref({
    page: 1,
    pageSize: 10,
    showSizePicker: true,
    pageSizes: [10, 20, 50],
  });

  const subscriptionColumns = [
    { title: '订单号', key: 'orderSn', ellipsis: { tooltip: true } },
    { title: '套餐', key: 'planName' },
    { title: '周期', key: 'periodType', width: 80 },
    {
      title: '金额',
      key: 'amount',
      width: 100,
      render: (row: any) =>
        h(
          'span',
          { style: { color: 'var(--warning-color)', fontWeight: '600' } },
          `${row.amount} USDT`,
        ),
    },
    { title: '天数', key: 'days', width: 60 },
    {
      title: '状态',
      key: 'status',
      width: 80,
      render: (row: any) => {
        const statusMap: any = {
          1: { text: '待支付', type: 'warning' },
          2: { text: '生效中', type: 'success' },
          3: { text: '已过期', type: 'default' },
          4: { text: '已取消', type: 'error' },
        };
        const status = statusMap[row.status] || { text: '未知', type: 'default' };
        return h(NTag, { type: status.type, size: 'small' }, { default: () => status.text });
      },
    },
    { title: '开始时间', key: 'startTime', width: 160 },
    { title: '到期时间', key: 'expireTime', width: 160 },
  ];

  const selectPlan = (plan: any) => {
    selectedPlan.value = plan;
    // 获取该套餐选择的周期（优先使用套餐的默认周期）
    const selectedPeriodForPlan = getPlanSelectedPeriod(plan.id);

    // 检查选择的周期是否可用
    const availablePeriods = getAvailablePeriods(plan);
    let finalPeriod = selectedPeriodForPlan;

    // 如果选择的周期不在可用列表中，优先使用套餐的默认周期，否则选择第一个可用的
    if (!availablePeriods.some((p) => p.key === finalPeriod)) {
      // 优先使用套餐的默认周期
      if (plan.defaultPeriod && availablePeriods.some((p) => p.key === plan.defaultPeriod)) {
        finalPeriod = plan.defaultPeriod;
      } else if (availablePeriods.length > 0) {
        finalPeriod = availablePeriods[0].key;
      } else {
        finalPeriod = 'monthly';
      }
    }

    // 更新周期选择
    planPeriodMap.value.set(plan.id, finalPeriod);
    selectedPeriod.value = finalPeriod;
    selectedPlanPeriod.value = { id: plan.id, period: finalPeriod };
  };

  // 取消选择套餐
  const handleCancelSelect = () => {
    selectedPlan.value = null;
    selectedPlanPeriod.value = null;
  };

  // 显示套餐详情
  const showPlanDetail = (plan: any) => {
    dialog.info({
      title: `${plan.planName} - 套餐详情`,
      content: () =>
        h('div', { style: 'padding: 16px 0;' }, [
          h('div', { style: 'margin-bottom: 16px;' }, [
            h(
              'n-text',
              { strong: true, style: 'font-size: 16px;' },
              `支持 ${plan.robotLimit} 个云机器人同时运行`,
            ),
          ]),
          h('n-divider'),
          h('div', { style: 'margin-bottom: 12px;' }, [
            h('n-text', { strong: true }, '价格方案：'),
          ]),
          h(
            'div',
            { style: 'margin-left: 16px;' },
            pricingPeriods.map((period: any) => {
              const price = getPlanPrice(plan, period.key);
              return h(
                'div',
                {
                  style:
                    'display: flex; justify-content: space-between; padding: 8px 0; border-bottom: 1px solid var(--border-color);',
                },
                [
                  h('span', period.label),
                  h('n-space', { align: 'center', size: 8 }, [
                    h('span', { style: 'font-weight: 600;' }, `${price} USDT`),
                    period.discount
                      ? h(
                          NTag,
                          { size: 'small', type: period.key === 'yearly' ? 'warning' : 'success' },
                          { default: () => period.discount },
                        )
                      : null,
                  ]),
                ],
              );
            }),
          ),
          h('n-divider', { style: 'margin: 12px 0;' }),
          h('div', { style: 'margin-bottom: 12px;' }, [
            h('n-text', { strong: true }, '套餐特性：'),
          ]),
          h('div', { style: 'margin-left: 16px;' }, [
            h(
              'div',
              {
                class: 'feature-item',
                style: 'display: flex; align-items: center; gap: 8px; padding: 6px 0;',
              },
              [
                h('n-icon', { component: CheckOutlined, color: 'var(--success-color)' }),
                h('span', `支持 ${plan.robotLimit} 个云机器人同时运行`),
              ],
            ),
            h(
              'div',
              {
                class: 'feature-item',
                style: 'display: flex; align-items: center; gap: 8px; padding: 6px 0;',
              },
              [
                h('n-icon', { component: CheckOutlined, color: 'var(--success-color)' }),
                h('span', '全自动行情分析'),
              ],
            ),
            h(
              'div',
              {
                class: 'feature-item',
                style: 'display: flex; align-items: center; gap: 8px; padding: 6px 0;',
              },
              [
                h('n-icon', { component: CheckOutlined, color: 'var(--success-color)' }),
                h('span', '全自动下单平仓'),
              ],
            ),
            h(
              'div',
              {
                class: 'feature-item',
                style: 'display: flex; align-items: center; gap: 8px; padding: 6px 0;',
              },
              [
                h('n-icon', { component: GiftOutlined, color: 'var(--warning-color)' }),
                h('span', '支持积分抵扣'),
              ],
            ),
          ]),
        ]),
      positiveText: '选择此套餐',
      onPositiveClick: () => {
        selectPlan(plan);
      },
      style: { width: '500px' },
    });
  };

  const getPlanPrice = (plan: any, period: string) => {
    const priceMap: any = {
      daily: plan.priceDaily,
      monthly: plan.priceMonthly,
      quarterly: plan.priceQuarterly,
      half_year: plan.priceHalfYear,
      yearly: plan.priceYearly,
    };
    return priceMap[period] || 0;
  };

  // 获取套餐可用的价格方案（只显示有价格的方案）
  const getAvailablePeriods = (plan: any) => {
    return pricingPeriods.filter((period) => {
      const price = getPlanPrice(plan, period.key);
      return price > 0;
    });
  };

  // 获取可用的周期选项（用于套餐卡片）
  const getAvailablePeriodOptionsForPlan = (plan: any) => {
    const availablePeriods = getAvailablePeriods(plan);
    return periodOptions.filter((option) => {
      return availablePeriods.some((period) => period.key === option.value);
    });
  };

  // 获取套餐选择的周期
  const getPlanSelectedPeriod = (planId: number) => {
    // 如果用户手动选择过，使用用户的选择
    if (planPeriodMap.value.has(planId)) {
      return planPeriodMap.value.get(planId)!;
    }

    // 使用套餐管理中设置的推荐周期
    const plan = planList.value.find((p) => p.id === planId);
    if (plan?.defaultPeriod) {
      return plan.defaultPeriod;
    }

    // 如果没有设置推荐周期，使用第一个可用的价格方案
    if (plan) {
      const availablePeriods = getAvailablePeriods(plan);
      if (availablePeriods.length > 0) {
        return availablePeriods[0].key;
      }
    }

    return '';
  };

  // 处理周期变化
  const handlePeriodChange = (planId: number, period: string) => {
    planPeriodMap.value.set(planId, period);
    // 如果当前套餐被选中，同步更新selectedPeriod
    if (selectedPlan.value?.id === planId) {
      selectedPeriod.value = period;
      selectedPlanPeriod.value = { id: planId, period };
    }
  };

  // 获取周期标签
  const getPeriodLabel = (period: string) => {
    const option = periodOptions.find((opt) => opt.value === period);
    return option?.label || period;
  };

  // 计算每个机器人每天的费用
  const getDailyPricePerRobot = (plan: any, period: string) => {
    const price = getPlanPrice(plan, period);
    if (price <= 0 || plan.robotLimit <= 0) return 0;

    const daysMap: any = {
      daily: 1,
      monthly: 30,
      quarterly: 90,
      half_year: 180,
      yearly: 365,
    };
    const days = daysMap[period] || 30;

    // 总价格 / 天数 / 机器人数量
    return price / days / plan.robotLimit;
  };

  const getSelectedPrice = () => {
    if (!selectedPlan.value) return 0;
    return getPlanPrice(selectedPlan.value, selectedPeriod.value);
  };

  const getGiftPower = () => {
    if (!selectedPlan.value) return 0;
    const giftMap: any = {
      daily: selectedPlan.value.giftPowerDaily || 0,
      monthly: selectedPlan.value.giftPowerMonthly || 0,
      quarterly: selectedPlan.value.giftPowerQuarterly || 0,
      half_year: selectedPlan.value.giftPowerHalfYear || 0,
      yearly: selectedPlan.value.giftPowerYearly || 0,
    };
    return giftMap[selectedPeriod.value] || 0;
  };

  // 获取订阅周期天数
  const getPeriodDays = () => {
    const daysMap: any = {
      daily: 1,
      monthly: 30,
      quarterly: 90,
      half_year: 180,
      yearly: 365,
    };
    return daysMap[selectedPeriod.value] || 30;
  };

  // 计算最终支付金额（考虑积分抵扣）
  const getFinalAmount = () => {
    const price = getSelectedPrice();
    if (!usePoints.value) return price;
    const pointsDeduct = Math.min(userPoints.value, price);
    return Math.max(0, price - pointsDeduct);
  };

  // 判断是否可以使用余额支付
  const canUseBalance = computed(() => {
    return accountBalance.value >= getFinalAmount();
  });

  const loadData = async () => {
    loading.value = true;
    try {
      const [planRes, subRes, mySubRes, walletRes] = await Promise.all([
        ToogoSubscriptionApi.planList({ status: 1 }),
        ToogoSubscriptionApi.subscriptionList({
          page: pagination.value.page,
          perPage: pagination.value.pageSize,
        }),
        ToogoSubscriptionApi.mySubscription(),
        ToogoWalletApi.overview(),
      ]);
      planList.value = planRes?.list || [];
      subscriptionList.value = subRes?.list || [];
      currentSubscription.value = mySubRes || {};
      userPoints.value = walletRes?.giftPower || 0; // 加载用户积分余额
      accountBalance.value = walletRes?.balance || 0; // 加载账户余额

      // 首次加载时自动选择推荐套餐（isDefault === 1）
      if (isFirstLoad.value) {
        const recommendedPlan = planList.value.find((plan) => plan.isDefault === 1);
        if (recommendedPlan) {
          selectPlan(recommendedPlan);
        }
        isFirstLoad.value = false;
      }
    } catch (error) {
      console.error('加载数据失败:', error);
    } finally {
      loading.value = false;
    }
  };

  const doSubscribe = async () => {
    if (!selectedPlan.value) {
      message.warning('请选择套餐');
      return;
    }

    // 余额支付时检查余额是否充足
    if (payType.value === 'balance') {
      const finalAmount = getFinalAmount();
      if (accountBalance.value < finalAmount) {
        message.error(
          `余额不足，当前余额 ${accountBalance.value.toFixed(2)} USDT，需要 ${finalAmount.toFixed(
            2,
          )} USDT`,
        );
        return;
      }
    }

    subscribeLoading.value = true;
    try {
      const res = await ToogoSubscriptionApi.subscribe({
        planId: selectedPlan.value.id,
        periodType: selectedPeriod.value,
        payType: payType.value,
        usePoints: usePoints.value, // 传递是否使用积分抵扣
        // 不传递 userId，让后端从上下文获取
      });

      // 显示支付详情
      let successMsg = `订阅成功！`;
      if (res?.pointsDeduct > 0) {
        successMsg += `积分抵扣 ${res.pointsDeduct.toFixed(2)} 积分，`;
      }
      if (res?.balancePaid > 0) {
        successMsg += `余额支付 ${res.balancePaid.toFixed(2)} USDT，`;
      }
      successMsg += `到期时间: ${res?.expireTime || '--'}`;

      message.success(successMsg);
      selectedPlan.value = null;
      usePoints.value = false;
      payType.value = 'balance';
      await loadData(); // 重新加载数据，更新余额和订阅信息
    } catch (error: any) {
      message.error(error.message || '订阅失败');
    } finally {
      subscribeLoading.value = false;
    }
  };

  const handleSubscribe = async () => {
    if (!selectedPlan.value) {
      message.warning('请选择套餐');
      return;
    }
    if (subscribeLoading.value) return;

    const planName = selectedPlan.value.planName || '--';
    const periodLabel = getPeriodLabel(selectedPeriod.value) || '--';
    const finalAmount = getFinalAmount();

    dialog.warning({
      title: '确认订阅',
      content: () =>
        h('div', { style: 'line-height: 1.8;' }, [
          h('div', [h('b', '套餐：'), planName]),
          h('div', [h('b', '周期：'), periodLabel]),
          h('div', [
            h('b', '支付方式：'),
            payType.value === 'balance' ? '余额支付' : payType.value,
          ]),
          usePoints.value
            ? h('div', [
                h('b', '积分抵扣：'),
                `${Math.min(userPoints.value, getSelectedPrice()).toFixed(2)} 积分`,
              ])
            : null,
          h('div', [h('b', '应付金额：'), `${finalAmount.toFixed(2)} USDT`]),
        ]),
      positiveText: '确认支付',
      negativeText: '取消',
      onPositiveClick: async () => {
        await doSubscribe();
      },
    });
  };

  onMounted(() => {
    loadData();
  });
</script>

<style lang="less" scoped>
  .subscription-page {
    padding: 8px;
  }

  .plan-card {
    position: relative;
    transition: all 0.2s ease;
    height: 100%;
    border: 1px solid #e5e7eb;
    border-radius: 6px;
    background: #fff;

    &:hover {
      border-color: #d1d5db;
      box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
    }

    &.plan-popular {
      border: 1.5px solid #fbbf24;

      &:hover {
        border-color: #f59e0b;
      }
    }

    &.plan-selected {
      border: 2px solid #2563eb;
      background: #f0f9ff;
      box-shadow: 0 0 0 2px rgba(37, 99, 235, 0.1);
    }

    .popular-badge {
      position: absolute;
      top: 8px;
      right: 8px;
      z-index: 1;
    }

    .plan-header {
      text-align: center;

      .plan-name {
        font-size: 16px;
        font-weight: 600;
        margin: 0 0 12px 0;
        color: #1f2937;
      }

      .plan-robots {
        display: flex;
        align-items: center;
        justify-content: center;
        gap: 8px;

        .robot-count {
          font-size: 28px;
          font-weight: 700;
          color: #2563eb;
        }

        .robot-label {
          font-size: 13px;
          color: #6b7280;
        }
      }
    }

    .plan-period-select {
      padding: 6px 0;
    }

    .plan-price-display {
      text-align: center;
      padding: 6px 0;

      .price-main {
        display: flex;
        align-items: baseline;
        justify-content: center;
        margin-bottom: 4px;
      }

      .daily-price-hint {
        margin-top: 4px;
      }
    }

    .plan-action {
      margin-top: 8px;
    }
  }

  .summary-item {
    display: flex;
    align-items: center;
    gap: 8px;

    .label {
      color: #6b7280;
      font-size: 13px;
    }
  }

  .mb-3 {
    margin-bottom: 12px;
  }

  .mt-3 {
    margin-top: 12px;
  }

  // 优化账户余额卡片样式
  :deep(.n-card) {
    border-radius: 6px;

    &:not(.plan-card) {
      border: 1px solid #e5e7eb;
    }
  }

  // 优化按钮样式
  :deep(.n-button) {
    border-radius: 4px;
  }

  // 优化订单详情卡片
  .order-panel {
    background: #fafafa;

    :deep(.n-card__content) {
      padding: 16px;
    }

    .n-descriptions {
      :deep(.n-descriptions-table-wrapper) {
        border-radius: 4px;
      }
    }

    .payment-summary {
      background: #fff;
    }
  }

  // 优化支付方式卡片
  :deep(.n-radio-group) {
    width: 100%;

    .n-radio {
      width: 100%;
      padding: 8px 12px;
      border: 1px solid #e5e7eb;
      border-radius: 4px;
      transition: all 0.2s;

      &:hover {
        background: #f9fafb;
        border-color: #d1d5db;
      }

      &.n-radio--checked {
        background: #eff6ff;
        border-color: #2563eb;
      }
    }
  }

  // 优化数据表格
  :deep(.n-data-table) {
    .n-data-table-th {
      background: #f9fafb;
      font-weight: 600;
    }
  }
</style>
