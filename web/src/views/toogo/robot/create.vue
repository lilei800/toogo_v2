<template>
  <div class="robot-create-page">
    <n-card title="创建云机器人" :bordered="false">
      <template #header-extra>
        <n-button @click="$router.back()">返回列表</n-button>
      </template>

      <n-form ref="formRef" :model="formData" :rules="rules" label-placement="left" label-width="120">
        <!-- 基础设置 -->
        <n-card title="基础设置" size="small" :bordered="true" style="margin-bottom: 16px">
          <n-grid :cols="3" :x-gap="24">
            <n-gi>
              <n-form-item label="机器人名称" path="robotName">
                <n-input v-model:value="formData.robotName" placeholder="请输入机器人名称" />
              </n-form-item>
            </n-gi>
            <n-gi>
              <n-form-item label="API配置" path="apiConfigId">
                <n-select
                  v-model:value="formData.apiConfigId"
                  :options="apiConfigOptions"
                  placeholder="请选择API配置"
                  @update:value="onApiConfigChange"
                />
              </n-form-item>
            </n-gi>
            <n-gi>
              <n-form-item label="交易平台">
                <n-input :value="selectedPlatformLabel" disabled placeholder="自动跟随API配置" />
              </n-form-item>
            </n-gi>
            <n-gi>
              <n-form-item label="交易对" path="symbol">
                <n-select
                  v-model:value="formData.symbol"
                  :options="tradingPairOptions"
                  filterable
                  placeholder="选择交易对"
                />
              </n-form-item>
            </n-gi>
            <n-gi>
              <n-form-item label="最大盈利目标" path="maxProfitTarget">
                <n-input-number
                  v-model:value="formData.maxProfitTarget"
                  :min="0"
                  :precision="2"
                  style="width: 100%"
                  placeholder="达到后自动停止"
                >
                  <template #suffix>USDT</template>
                </n-input-number>
              </n-form-item>
            </n-gi>
            <n-gi>
              <n-form-item label="最大亏损额" path="maxLossAmount">
                <n-input-number
                  v-model:value="formData.maxLossAmount"
                  :min="0"
                  :precision="2"
                  style="width: 100%"
                  placeholder="达到后自动停止"
                >
                  <template #suffix>USDT</template>
                </n-input-number>
              </n-form-item>
            </n-gi>
          </n-grid>
        </n-card>

        <!-- 策略模板选择 -->
        <n-card title="策略模板" size="small" :bordered="true" style="margin-bottom: 16px">

          <n-tabs type="segment" animated v-model:value="strategyTabName">
            <!-- 我的策略模板 -->
            <n-tab-pane name="my" tab="📋 我的策略">
              <div style="padding: 12px 0">
                <n-spin :show="loadingMyStrategies">
                  <div v-if="myStrategyGroups.length > 0">
                    <n-radio-group v-model:value="selectedStrategyGroupId" name="myStrategy">
                      <n-grid :cols="2" :x-gap="12" :y-gap="12">
                        <n-gi v-for="group in myStrategyGroups" :key="group.id">
                          <n-card 
                            hoverable 
                            size="small" 
                            :class="{ 'strategy-card-selected': selectedStrategyGroupId === group.id }"
                            @click="selectedStrategyGroupId = group.id; formData.strategySource = 'my'; loadStrategyTemplate(group.id)"
                          >
                            <n-radio :value="group.id" style="width: 100%">
                              <n-space vertical :size="4">
                                <n-space align="center">
                                  <n-text strong>{{ group.groupName }}</n-text>
                                  <n-tag v-if="group.isDefault" size="small" type="success">默认</n-tag>
                                </n-space>
                                <n-text depth="3" style="font-size: 12px">
                                  {{ group.symbol }} · {{ group.strategyCount || 12 }}种策略
                                </n-text>
                              </n-space>
                            </n-radio>
                          </n-card>
                        </n-gi>
                      </n-grid>
                    </n-radio-group>
                  </div>
                  <n-empty v-else description="暂无策略模板，请先到【策略模板】页面添加/创建">
                    <template #extra>
                      <n-button type="primary" size="small" @click="router.push('/toogo/strategy')">去策略模板</n-button>
                    </template>
                  </n-empty>
                </n-spin>
              </div>
            </n-tab-pane>

            <!-- 官方策略模板 -->
            <n-tab-pane name="official" tab="🔥 官方策略">
              <div style="padding: 12px 0">
                <n-spin :show="loadingOfficialStrategies">
                  <div v-if="officialStrategyGroups.length > 0">
                    <n-radio-group v-model:value="selectedOfficialGroupId" name="officialStrategy">
                      <n-grid :cols="2" :x-gap="12" :y-gap="12">
                        <n-gi v-for="group in officialStrategyGroups" :key="group.id">
                          <n-card
                            hoverable
                            size="small"
                            class="official-strategy-card"
                            :class="{ 'strategy-card-selected': selectedOfficialGroupId === group.id }"
                            @click="selectedOfficialGroupId = group.id; formData.strategySource = 'official'"
                          >
                            <n-radio :value="group.id" style="width: 100%">
                              <n-space vertical :size="4">
                                <n-space align="center">
                                  <n-text strong>{{ group.groupName }}</n-text>
                                  <n-tag size="small" type="success">官方</n-tag>
                                </n-space>
                                <n-text depth="3" style="font-size: 12px">
                                  {{ group.symbol }} · {{ group.strategyCount || 12 }}种策略
                                </n-text>
                              </n-space>
                            </n-radio>
                          </n-card>
                        </n-gi>
                      </n-grid>
                    </n-radio-group>
                    <n-text depth="3" style="font-size: 12px; display: block; margin-top: 8px">
                      提示：创建时会自动将所选官方策略复制到“我的策略”，并使用复制后的策略组创建机器人。
                    </n-text>
                  </div>
                  <n-empty v-else description="暂无官方策略模板">
                    <template #extra>
                      <n-button type="primary" size="small" @click="router.push('/toogo/strategy/official')">去官方策略页</n-button>
                    </template>
                  </n-empty>
                </n-spin>
              </div>
            </n-tab-pane>

            <!-- 盈利排行策略 -->
            <n-tab-pane name="ranking" tab="🏆 排行榜">
              <div style="padding: 24px; text-align: center">
                <n-text depth="3">盈利排行策略开发中，敬请期待</n-text>
              </div>
            </n-tab-pane>
          </n-tabs>

          <!-- 当前选择 -->
          <n-divider style="margin: 8px 0" />
          <n-space align="center" size="small">
            <n-text depth="3" style="font-size: 12px">已选：</n-text>
            <n-tag v-if="selectedStrategyGroupId" type="success" size="small">
              {{ getSelectedGroupName() }}
            </n-tag>
            <n-tag v-else size="small">未选择</n-tag>
          </n-space>
        </n-card>

         <!-- 策略模板说明 -->
         <n-card v-if="selectedStrategyGroupId" title="策略说明" size="small" :bordered="true" style="margin-bottom: 16px">
           <n-alert type="info" :bordered="false">
             <template #header>💡 智能策略匹配</template>
             <n-text depth="3" style="font-size: 12px">
               机器人运行时会根据实时市场状态（趋势/震荡/高波动/低波动）和风险偏好（激进/平衡/保守）自动从策略组中选择匹配的策略模板，无需手动选择。
             </n-text>
           </n-alert>
         </n-card>

         <!-- 高级设置（折叠） -->
         <n-collapse style="margin-bottom: 16px">
           <!-- 市场状态与风险偏好映射 -->
           <n-collapse-item title="🎯 市场状态与风险偏好映射" name="marketMapping">
             <template #header-extra>
               <n-space :size="4">
                 <n-tag v-for="market in marketStateMapping" :key="market.key" size="small" :bordered="false">
                   {{ market.icon }} {{ getRiskLabel(formData.marketRiskMapping[market.key]) }}
                 </n-tag>
               </n-space>
             </template>
             <n-text depth="3" style="font-size: 12px; margin-bottom: 12px; display: block">
               机器人运行时会根据市场状态自动选择对应的风险偏好，并从策略模板中加载交易参数
             </n-text>
             <n-grid :cols="4" :x-gap="12" :y-gap="12">
               <n-gi v-for="market in marketStateMapping" :key="market.key">
                 <div class="mapping-item-card">
                   <div class="mapping-header">
                     <n-tag :type="market.tagType" size="small">{{ market.icon }}</n-tag>
                     <span class="market-name">{{ market.label }}</span>
                   </div>
                   <div class="mapping-arrow">↓</div>
                   <n-select
                     v-model:value="formData.marketRiskMapping[market.key]"
                     :options="riskPreferenceSelectOptions"
                     size="small"
                     style="width: 100%"
                   />
                 </div>
               </n-gi>
             </n-grid>
           </n-collapse-item>

           <!-- 核心交易逻辑 -->
           <n-collapse-item title="🤖 核心交易逻辑" name="tradeLogic">
             <template #header-extra>
               <n-space :size="8">
                 <n-tag :type="formData.autoTradeEnabled ? 'success' : 'default'" size="small">
                   自动下单: {{ formData.autoTradeEnabled ? '开' : '关' }}
                 </n-tag>
                 <n-tag :type="formData.dualSidePosition ? 'info' : 'warning'" size="small">
                   双向开单: {{ formData.dualSidePosition ? '开' : '关' }}
                 </n-tag>
                 <n-tag :type="formData.autoCloseEnabled ? 'success' : 'default'" size="small">
                   自动平仓: {{ formData.autoCloseEnabled ? '开' : '关' }}
                 </n-tag>
               </n-space>
             </template>
             
             <!-- 方向判断说明 -->
             <n-card :bordered="true" size="small" style="margin-bottom: 12px; background: linear-gradient(135deg, #e8f4fd 0%, #f0f9ff 100%);">
               <n-space align="center" justify="space-between">
                 <n-space vertical :size="4">
                   <n-text strong style="font-size: 14px">📊 方向判断</n-text>
                   <n-text depth="3" style="font-size: 11px">根据时间窗口的时间和波动点数分析实时行情数据，在窗口时间内保持最高价和最低价</n-text>
                 </n-space>
                 <n-tag type="info" :bordered="false">自动执行</n-tag>
               </n-space>
               <n-divider style="margin: 10px 0" />
               <n-grid :cols="2" :x-gap="16">
                 <n-gi>
                   <div class="direction-rule long">
                     <span class="rule-icon">📈</span>
                     <div class="rule-content">
                       <span class="rule-title">做多预警</span>
                       <span class="rule-desc">实时价格 - 窗口最低价 = 波动值</span>
                       <span class="rule-note">（价格必须上升时触发）</span>
                     </div>
                   </div>
                 </n-gi>
                 <n-gi>
                   <div class="direction-rule short">
                     <span class="rule-icon">📉</span>
                     <div class="rule-content">
                       <span class="rule-title">做空预警</span>
                       <span class="rule-desc">窗口最高价 - 实时价格 = 波动值</span>
                       <span class="rule-note">（价格必须下降时触发）</span>
                     </div>
                   </div>
                 </n-gi>
               </n-grid>
               <n-text depth="3" style="font-size: 10px; margin-top: 8px; display: block; text-align: center">
                 ⚠️ 方向对冲时不给方向预警（多空同时满足条件时不触发）
               </n-text>
             </n-card>

             <!-- 全自动下单和平仓 -->
             <n-grid :cols="3" :x-gap="16">
               <n-gi>
                 <n-card :bordered="true" size="small" :class="{ 'feature-card-active': formData.autoTradeEnabled }">
                   <n-space align="center" justify="space-between">
                     <n-space vertical :size="4">
                       <n-text strong>🤖 全自动下单</n-text>
                       <n-text depth="3" style="font-size: 11px">根据方向信号和策略模板自动开仓</n-text>
                     </n-space>
                     <n-switch v-model:value="formData.autoTradeEnabled" size="large" />
                   </n-space>
                   <n-divider style="margin: 10px 0" />
                   <n-space vertical :size="2">
                     <n-text depth="3" style="font-size: 11px">• 每个方向只能有一单</n-text>
                     <n-text depth="3" style="font-size: 11px">• 杠杆和保证金从策略模板获取</n-text>
                   </n-space>
                 </n-card>
               </n-gi>
               <n-gi>
                 <n-card :bordered="true" size="small" :class="{ 'feature-card-active': formData.dualSidePosition }">
                   <n-space align="center" justify="space-between">
                     <n-space vertical :size="4">
                       <n-text strong>↔️ 双向开单</n-text>
                       <n-text depth="3" style="font-size: 11px">允许同时持有多单和空单</n-text>
                     </n-space>
                     <n-switch v-model:value="formData.dualSidePosition" size="large" />
                   </n-space>
                   <n-divider style="margin: 10px 0" />
                   <n-space vertical :size="2">
                     <n-text depth="3" style="font-size: 11px">• 开启：可同时持多单和空单</n-text>
                     <n-text depth="3" style="font-size: 11px">• 关闭：同时只能有一个持仓</n-text>
                   </n-space>
                 </n-card>
               </n-gi>
               <n-gi>
                 <n-card :bordered="true" size="small" :class="{ 'feature-card-active': formData.autoCloseEnabled }">
                   <n-space align="center" justify="space-between">
                     <n-space vertical :size="4">
                       <n-text strong>📉 全自动平仓</n-text>
                       <n-text depth="3" style="font-size: 11px">达到止损/止盈条件自动平仓</n-text>
                     </n-space>
                     <n-switch v-model:value="formData.autoCloseEnabled" size="large" />
                   </n-space>
                   <n-divider style="margin: 10px 0" />
                   <n-space vertical :size="2">
                     <n-text depth="3" style="font-size: 11px">• 止损：|盈亏|/(保证金×止损%) ≥ 100%时平仓</n-text>
                     <n-text depth="3" style="font-size: 11px">• 启动止盈：盈亏/保证金 ≥ 启动%时开启止盈按钮</n-text>
                     <n-text depth="3" style="font-size: 11px">• 止盈回撤：(最高盈利-当前盈亏)/最高盈利 ≥ 回撤%</n-text>
                   </n-space>
                 </n-card>
               </n-gi>
             </n-grid>
           </n-collapse-item>

           <!-- 策略开关 -->
           <n-collapse-item title="⚙️ 策略开关" name="switches">
             <template #header-extra>
               <n-space :size="8">
                 <n-tag :type="formData.autoMarketState ? 'success' : 'default'" size="small">市场分析</n-tag>
                 <n-tag :type="formData.useMonitorSignal ? 'success' : 'default'" size="small">方向信号</n-tag>
                 <n-tag :type="formData.autoStartTakeProfit ? 'success' : 'default'" size="small">自动止盈</n-tag>
               </n-space>
             </template>
             
             <n-grid :cols="4" :x-gap="16">
               <n-gi>
                 <div class="auto-setting-item">
                   <n-space align="center" justify="space-between">
                     <n-space vertical :size="0">
                       <n-text>自动分析市场</n-text>
                       <n-text depth="3" style="font-size: 10px">智能识别趋势/震荡/高低波动</n-text>
                     </n-space>
                     <n-switch v-model:value="formData.autoMarketState" />
                   </n-space>
                 </div>
               </n-gi>
               <n-gi>
                 <div class="auto-setting-item">
                   <n-space align="center" justify="space-between">
                     <n-space vertical :size="0">
                       <n-text>方向信号</n-text>
                       <n-text depth="3" style="font-size: 10px">窗口价格分析生成信号</n-text>
                     </n-space>
                     <n-switch v-model:value="formData.useMonitorSignal" />
                   </n-space>
                 </div>
               </n-gi>
               <n-gi>
                 <div class="auto-setting-item">
                   <n-space align="center" justify="space-between">
                     <n-space vertical :size="0">
                       <n-text>自动启动止盈</n-text>
                       <n-text depth="3" style="font-size: 10px">盈利达标后自动开启止盈</n-text>
                     </n-space>
                     <n-switch v-model:value="formData.autoStartTakeProfit" />
                   </n-space>
                 </div>
               </n-gi>
            </n-grid>
           </n-collapse-item>
         </n-collapse>

        <!-- 定时开关设置 -->
        <n-card title="定时开关" size="small" :bordered="true" style="margin-bottom: 16px">
          <n-grid :cols="2" :x-gap="24">
            <n-gi>
              <n-form-item label="定时启动" label-placement="left">
                <n-date-picker
                  v-model:value="formData.scheduleStart"
                  type="datetime"
                  clearable
                  placeholder="选择启动时间（可选）"
                  style="width: 100%"
                />
              </n-form-item>
            </n-gi>
            <n-gi>
              <n-form-item label="定时停止" label-placement="left">
                <n-date-picker
                  v-model:value="formData.scheduleStop"
                  type="datetime"
                  clearable
                  placeholder="选择停止时间（可选）"
                  style="width: 100%"
                />
              </n-form-item>
            </n-gi>
          </n-grid>
          <n-text depth="3" style="font-size: 12px">
            设置定时启动后，机器人将在指定时间自动启动；设置定时停止后，机器人将在指定时间自动暂停。不设置则立即生效。
          </n-text>
        </n-card>

        <!-- 提交按钮 -->
        <n-space justify="center" style="margin-top: 24px">
          <n-button size="large" @click="$router.back()">取消</n-button>
          <n-button type="primary" size="large" @click="handleSubmit" :loading="submitLoading">
            创建机器人
          </n-button>
        </n-space>
      </n-form>
    </n-card>

  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { useMessage } from 'naive-ui';
import { http } from '@/utils/http/axios';

const router = useRouter();
const route = useRoute();
const message = useMessage();

const formRef = ref();
const submitLoading = ref(false);
const apiConfigOptions = ref<any[]>([]);
const apiConfigMap = ref<Record<number, any>>({});
const hasUserStrategy = ref(false); // 是否有用户自定义策略
const userStrategies = ref<any[]>([]); // 用户自定义策略列表

// 策略模板相关
const strategyTabName = ref('my');
const myStrategyGroups = ref<any[]>([]);
const selectedStrategyGroupId = ref<number | null>(null);
const officialStrategyGroups = ref<any[]>([]);
const selectedOfficialGroupId = ref<number | null>(null);
// 不再需要存储具体策略模板，运行时自动匹配
// const selectedStrategyTemplate = ref<any>(null);
const loadingMyStrategies = ref(false);
const loadingOfficialStrategies = ref(false);

// 市场状态映射配置
const marketStateMapping = [
  { key: 'trend', label: '趋势市场', icon: '📈', tagType: 'success' as const, description: '市场方向明确，趋势强劲' },
  { key: 'volatile', label: '震荡市场', icon: '📊', tagType: 'warning' as const, description: '价格在区间内波动，方向不明' },
  { key: 'high_vol', label: '高波动市场', icon: '⚡', tagType: 'error' as const, description: '价格剧烈波动，风险较高' },
  { key: 'low_vol', label: '低波动市场', icon: '😴', tagType: 'info' as const, description: '价格变化缓慢，波动率低' },
];

// 风险偏好选项
const riskPreferenceSelectOptions = [
  { label: '🛡️ 保守', value: 'conservative' },
  { label: '⚖️ 平衡', value: 'balanced' },
  { label: '🚀 激进', value: 'aggressive' },
];

// 获取风险偏好标签缩写
const getRiskLabel = (value: string) => {
  const map: Record<string, string> = {
    conservative: '保守',
    balanced: '平衡',
    aggressive: '激进',
  };
  return map[value] || value;
};

const formData = ref({
  robotName: '',
  apiConfigId: null as number | null,
  maxProfitTarget: 1000,
  maxLossAmount: 500,
  // 策略来源: my-我的策略, official-官方推荐
  strategySource: 'my',
  // 自动化设置（直接使用策略模板参数，不再动态计算）
  autoMarketState: true,      // 自动分析市场状态
  useMonitorSignal: true,     // 方向信号开关
  autoStartTakeProfit: true,  // 自动启动止盈回撤开关
  // 全自动交易开关（参数从策略模板获取）
  autoTradeEnabled: true,
  autoCloseEnabled: true,
  dualSidePosition: true,  // 双向开单：默认开启
  // 市场状态与风险偏好映射
  marketRiskMapping: {
    trend: 'balanced',        // 📈 趋势市场 → 平衡
    volatile: 'balanced',     // 📊 震荡市场 → 平衡
    high_vol: 'aggressive',   // ⚡ 高波动市场 → 激进
    low_vol: 'conservative',  // 😴 低波动市场 → 保守
  } as Record<string, string>,
  // 定时开关设置
  scheduleStart: null as number | null,
  scheduleStop: null as number | null,
  // 默认配置（将从策略模板覆盖）
  exchange: '',
  symbol: 'BTCUSDT',
  orderType: 'market',
  marginMode: 'isolated',
  leverage: 20, // 默认20倍杠杆
  marginPercent: 50, // 默认50%保证金比例
  stopLossPercent: 5,
  autoStartRetreatPercent: 3,
  profitRetreatPercent: 30,
  marketState: 'trend',
  riskPreference: 'balanced',
});

// 账户余额（已废弃，不再使用）
// const accountBalance = ref(20);

const rules = {
  robotName: { required: true, message: '请输入机器人名称', trigger: 'blur' },
  apiConfigId: { required: true, type: 'number', message: '请选择API配置', trigger: 'change' },
  maxProfitTarget: { required: true, type: 'number', message: '请设置最大盈利目标', trigger: 'change' },
  maxLossAmount: { required: true, type: 'number', message: '请设置最大亏损额', trigger: 'change' },
  symbol: { required: true, message: '请选择交易对', trigger: 'change' },
  leverage: { required: true, type: 'number', message: '请设置杠杆倍数', trigger: 'change' },
  marginPercent: { required: true, type: 'number', message: '请设置保证金比例', trigger: 'change' },
  stopLossPercent: { required: true, type: 'number', message: '请设置止损百分比', trigger: 'change' },
  profitRetreatPercent: { required: true, type: 'number', message: '请设置止盈回撤百分比', trigger: 'change' },
  autoStartRetreatPercent: { required: true, type: 'number', message: '请设置启动回撤百分比', trigger: 'change' },
  marketState: { required: true, message: '请选择市场状态', trigger: 'change' },
  riskPreference: { required: true, message: '请选择风险偏好', trigger: 'change' },
};

const platformLabels: Record<string, string> = {
  bitget: 'Bitget',
  binance: 'Binance (币安)',
  okx: 'OKX (欧易)',
  gateio: 'Gate.io',
};

const selectedPlatformLabel = computed(() => {
  if (!formData.value.exchange) return '请先选择API配置';
  return platformLabels[formData.value.exchange] || formData.value.exchange;
});

const tradingPairOptions = [
  { label: 'BTC/USDT', value: 'BTCUSDT' },
  { label: 'ETH/USDT', value: 'ETHUSDT' },
  { label: 'BNB/USDT', value: 'BNBUSDT' },
  { label: 'XRP/USDT', value: 'XRPUSDT' },
  { label: 'SOL/USDT', value: 'SOLUSDT' },
  { label: 'DOGE/USDT', value: 'DOGEUSDT' },
  { label: 'ADA/USDT', value: 'ADAUSDT' },
  { label: 'AVAX/USDT', value: 'AVAXUSDT' },
  { label: 'MATIC/USDT', value: 'MATICUSDT' },
  { label: 'DOT/USDT', value: 'DOTUSDT' },
];

// API配置选择变化时，自动设置交易平台
const onApiConfigChange = (id: number) => {
  const config = apiConfigMap.value[id];
  if (config) {
    formData.value.exchange = config.platform;
  }
};

// 加载API配置列表
const loadApiConfigs = async () => {
  try {
    const res = await http.request({
      url: '/trading/apiConfig/list',
      method: 'get',
      params: { page: 1, pageSize: 100 },
    });
    const list = res?.list || [];
    apiConfigOptions.value = list.map((item: any) => ({
      label: `${item.apiName} (${platformLabels[item.platform] || item.platform})`,
      value: item.id,
    }));
    // 建立映射
    list.forEach((item: any) => {
      apiConfigMap.value[item.id] = item;
    });
    // 自动选择第一个
    if (list.length > 0 && !formData.value.apiConfigId) {
      formData.value.apiConfigId = list[0].id;
      formData.value.exchange = list[0].platform;
    }
  } catch (error) {
    console.error('加载API配置失败:', error);
  }
};

// 加载用户策略模板（已废弃，不再需要）
const loadUserStrategies = async () => {
  // 不再需要加载用户策略模板，直接使用策略组
  hasUserStrategy.value = false;
};

// 加载我的策略模板组
const loadMyStrategyGroups = async () => {
  loadingMyStrategies.value = true;
  try {
    const res = await http.request({
      url: '/strategy/group/list',
      method: 'get',
      params: { page: 1, pageSize: 100, isOfficial: 0 },
    });
    myStrategyGroups.value = res?.list || [];
    hasUserStrategy.value = myStrategyGroups.value.length > 0;
    // 自动选择默认或第一个，并加载策略模板
    const defaultGroup = myStrategyGroups.value.find((g: any) => g.isDefault);
    if (defaultGroup) {
      selectedStrategyGroupId.value = defaultGroup.id;
      await loadStrategyTemplate(defaultGroup.id);
    } else if (myStrategyGroups.value.length > 0) {
      selectedStrategyGroupId.value = myStrategyGroups.value[0].id;
      await loadStrategyTemplate(myStrategyGroups.value[0].id);
    }
  } catch (error) {
    console.error('加载我的策略失败:', error);
  } finally {
    loadingMyStrategies.value = false;
  }
};

// 加载官方策略模板组
// 已移除：创建机器人页不展示“官方策略模板”标签（官方策略请到“策略模板”页面操作）

// 获取选中的我的策略组名称
const getSelectedMyGroupName = () => {
  const group = myStrategyGroups.value.find((g: any) => g.id === selectedStrategyGroupId.value);
  return group?.groupName || '';
};

// 获取选中的官方策略组名称
const getSelectedOfficialGroupName = () => {
  const group = officialStrategyGroups.value.find((g: any) => g.id === selectedOfficialGroupId.value);
  return group?.groupName || '';
};

// 当前选择展示名称（兼容“我的/官方”）
const getSelectedGroupName = () => {
  if (formData.value.strategySource === 'official') {
    return getSelectedOfficialGroupName();
  }
  return getSelectedMyGroupName();
};

// 加载策略模板参数（已废弃，不再需要加载具体策略模板）
// 机器人运行时会自动匹配策略模板，创建时只需要选择策略组
const loadStrategyTemplate = async (groupId: number) => {
  // 不再加载具体策略模板，只记录策略组ID
  // 运行时系统会根据市场状态自动匹配
};

// 加载官方策略模板组
const loadOfficialStrategyGroups = async () => {
  loadingOfficialStrategies.value = true;
  try {
    const res = await http.request({
      url: '/strategy/group/list',
      method: 'get',
      params: { page: 1, pageSize: 100, isOfficial: 1 },
    });
    officialStrategyGroups.value = res?.list || [];
  } catch (error) {
    console.error('加载官方策略失败:', error);
  } finally {
    loadingOfficialStrategies.value = false;
  }
};

// 复制官方策略到“我的策略”，返回复制后的策略组ID
const copyOfficialToMy = async (officialGroupId: number) => {
  const res = await http.request({
    url: '/strategy/group/copyFromOfficial',
    method: 'post',
    data: { officialGroupId },
  });
  // 后端返回 { id: number }
  const id = Number(res?.id || 0);
  return id;
};

// 已移除策略模板参数相关函数，运行时自动匹配

const handleSubmit = async () => {
  try {
    await formRef.value?.validate();
  } catch (error) {
    return;
  }

  if (!formData.value.apiConfigId) {
    message.error('请先添加API配置');
    return;
  }

  // 验证必须选择策略组
  const isOfficialSelected = formData.value.strategySource === 'official';
  if (isOfficialSelected) {
    if (!selectedOfficialGroupId.value) {
      message.error('请选择官方策略组');
      return;
    }
  } else if (!selectedStrategyGroupId.value) {
    message.error('请选择策略组');
    return;
  }

  submitLoading.value = true;
  try {
    // 确定策略组ID：
    // - 我的策略：直接使用选中的策略组
    // - 官方策略：创建时自动复制到“我的策略”，并使用复制后的策略组创建机器人
    let strategyGroupId = selectedStrategyGroupId.value || 0;
    if (isOfficialSelected) {
      const officialId = selectedOfficialGroupId.value || 0;
      const copiedId = await copyOfficialToMy(officialId);
      if (!copiedId) {
        message.error('复制官方策略失败，请重试');
        submitLoading.value = false;
        return;
      }
      strategyGroupId = copiedId;
      // 尝试刷新我的策略列表，方便用户后续查看/复用
      await loadMyStrategyGroups();
    }
    
    if (strategyGroupId <= 0) {
      message.error('策略组选择失败，请重新选择');
      submitLoading.value = false;
      return;
    }

     // 机器人运行时会自动匹配策略模板，创建时只需要传递策略组ID和映射关系
     // 交易参数（杠杆、保证金、止损、止盈等）运行时从策略模板加载，不需要传递
     const data: Record<string, any> = {
       // 基础信息（必需）
       robotName: formData.value.robotName,
       apiConfigId: formData.value.apiConfigId,
       maxProfitTarget: formData.value.maxProfitTarget,
       maxLossAmount: formData.value.maxLossAmount,
       maxRuntime: 0, // 最大运行时长，默认0表示不限制
       
       // ⭐ 策略配置（必需）
       strategyGroupId: strategyGroupId,  // 策略组ID（运行时根据此ID查询策略模板）
       marketRiskMapping: formData.value.marketRiskMapping,  // 市场状态映射（运行时根据此映射匹配风险偏好）
       
       // 交易基础配置（必需）
       exchange: formData.value.exchange,  // 交易所（从API配置获取）
       symbol: formData.value.symbol,  // 交易对
       
       // 自动化开关（必需）
       autoTradeEnabled: formData.value.autoTradeEnabled ? 1 : 0,
       autoCloseEnabled: formData.value.autoCloseEnabled ? 1 : 0,
       dualSidePosition: formData.value.dualSidePosition ? 1 : 0,  // 双向开单
       autoMarketState: formData.value.autoMarketState ? 1 : 0,
       useMonitorSignal: formData.value.useMonitorSignal ? 1 : 0,
       
       // 定时开关设置（可选）
       scheduleStart: formData.value.scheduleStart ? new Date(formData.value.scheduleStart).toISOString() : '',
       scheduleStop: formData.value.scheduleStop ? new Date(formData.value.scheduleStop).toISOString() : '',
       
       // 备注
       remark: `策略组ID: ${strategyGroupId}`,
       
       // ❌ 以下字段不再传递，运行时从策略模板加载：
       // - leverage, marginPercent, stopLossPercent, profitRetreatPercent, autoStartRetreatPercent
       // - orderType, marginMode, marketState, riskPreference
     };

    await http.request({
      url: '/trading/robot/create',
      method: 'post',
      data,
    });
    message.success('机器人创建成功！');
    router.push('/toogo/robot');
  } catch (error: any) {
    const errorMsg = error?.message || error?.data?.message || '创建失败';
    message.error(errorMsg);
  } finally {
    submitLoading.value = false;
  }
};

// 已移除：创建机器人页不再展示/计算“预计消耗算力”，也不再读取算力消耗比例配置

// 已删除：加载账户余额和BTC价格的代码（不再需要）

onMounted(async () => {
  loadApiConfigs();
  loadUserStrategies();
  // 已删除：loadBtcPrice() - 不再需要
  await loadMyStrategyGroups();
  await loadOfficialStrategyGroups();
  
  // 处理从策略页面跳转过来的参数
  const preSelectedGroupId = route.query.strategyGroupId;
  if (preSelectedGroupId) {
    selectedStrategyGroupId.value = Number(preSelectedGroupId);
    formData.value.strategySource = 'my';
    strategyTabName.value = 'my';
  }
});
</script>

<style scoped lang="less">
.robot-create-page {
  padding: 16px;
  max-width: 1000px;
  margin: 0 auto;
}

.auto-setting-item {
  padding: 12px;
  background: #f8f9fa;
  border-radius: 8px;
  transition: all 0.3s ease;

  &:hover {
    background: #f0f2f5;
  }

  &.disabled {
    opacity: 0.6;
    background: #fafafa;
  }
}

.strategy-card-selected {
  border: 2px solid #18a058 !important;
  background: #f6ffed;
}

.official-strategy-card {
  transition: all 0.2s ease;
  height: 100%;
  
  &:hover {
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
    transform: translateY(-1px);
  }
  
  :deep(.n-card-body) {
    padding: 12px 16px;
  }
}

.feature-card-active {
  border: 2px solid #18a058 !important;
  background: linear-gradient(135deg, #f6ffed 0%, #e8f5e9 100%);
  box-shadow: 0 2px 8px rgba(24, 160, 88, 0.15);
}

:deep(.n-card.feature-card-active) {
  border-color: #18a058 !important;
}

/* 方向判断规则样式 */
.direction-rule {
  display: flex;
  align-items: center;
  padding: 10px 12px;
  border-radius: 8px;
  background: #fff;
  
  .rule-icon {
    font-size: 20px;
    margin-right: 10px;
  }
  
  .rule-content {
    display: flex;
    flex-direction: column;
    
    .rule-title {
      font-weight: 600;
      font-size: 13px;
      color: #333;
    }
    
    .rule-desc {
      font-size: 11px;
      color: #666;
      margin-top: 2px;
    }
    
    .rule-note {
      font-size: 10px;
      color: #999;
      margin-top: 2px;
      font-style: italic;
    }
  }
  
  &.long {
    border-left: 3px solid #18a058;
    
    .rule-title {
      color: #18a058;
    }
  }
  
  &.short {
    border-left: 3px solid #d03050;
    
    .rule-title {
      color: #d03050;
    }
  }
}

/* 其他样式 */
.reverse-rule {
  text-align: center;
  padding: 8px;
  background: #fafafa;
  border-radius: 6px;
}

/* 市场状态映射卡片 */
.mapping-item-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 12px;
  background: rgba(128, 128, 128, 0.05);
  border-radius: 8px;
  transition: background 0.2s;
  
  &:hover {
    background: rgba(128, 128, 128, 0.1);
  }
  
  .mapping-header {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 4px;
    
    .market-name {
      font-weight: 500;
      font-size: 13px;
    }
  }
  
  .mapping-arrow {
    color: #999;
    font-size: 14px;
    margin: 4px 0;
  }
}
</style>
