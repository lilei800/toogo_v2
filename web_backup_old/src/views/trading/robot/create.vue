<template>
  <div>
    <n-card :bordered="false" title="创建交易机器人">
      <n-steps :current="current" :status="currentStatus">
        <n-step title="基础设置" />
        <n-step title="风险偏好" />
        <n-step title="市场行情" />
        <n-step title="下单设置" />
        <n-step title="自动平仓" />
      </n-steps>

      <n-card :bordered="false" style="margin-top: 20px;">
        <!-- 步骤1：基础设置 -->
        <div v-show="current === 1">
          <n-form ref="formRef1" :model="formValue" :rules="rules" label-placement="left" label-width="150">
            <n-form-item label="选择API接口" path="apiConfigId">
              <n-select
                v-model:value="formValue.apiConfigId"
                :options="apiConfigOptions"
                placeholder="请选择API接口"
              />
            </n-form-item>
            <n-form-item label="机器人名称" path="name">
              <n-input v-model:value="formValue.name" placeholder="请输入机器人名称" />
            </n-form-item>
            <n-form-item label="最大盈利目标 (USDT)" path="maxProfit">
              <n-input-number v-model:value="formValue.maxProfit" :min="0" style="width: 100%;" />
            </n-form-item>
            <n-form-item label="最大亏损额 (USDT)" path="maxLoss">
              <n-input-number v-model:value="formValue.maxLoss" :min="0" style="width: 100%;" />
            </n-form-item>
            <n-form-item label="最大运行时长 (秒)" path="maxRuntime">
              <n-input-number v-model:value="formValue.maxRuntime" :min="0" style="width: 100%;" />
            </n-form-item>
          </n-form>
        </div>

        <!-- 步骤2：风险偏好 -->
        <div v-show="current === 2">
          <n-form ref="formRef2" :model="formValue" label-placement="left" label-width="150">
            <n-form-item label="风险偏好">
              <n-radio-group v-model:value="formValue.riskPreference">
                <n-space vertical>
                  <n-radio value="conservative">保守型</n-radio>
                  <n-radio value="balanced">平衡型</n-radio>
                  <n-radio value="aggressive">激进型</n-radio>
                  <n-radio value="auto">自动选择</n-radio>
                </n-space>
              </n-radio-group>
            </n-form-item>
          </n-form>
        </div>

        <!-- 步骤3：市场行情 -->
        <div v-show="current === 3">
          <n-form ref="formRef3" :model="formValue" label-placement="left" label-width="150">
            <n-form-item label="市场行情">
              <n-radio-group v-model:value="formValue.marketState">
                <n-space vertical>
                  <n-radio value="trend">趋势市场</n-radio>
                  <n-radio value="oscillation">震荡市场</n-radio>
                  <n-radio value="high_volatility">高波动</n-radio>
                  <n-radio value="low_volatility">低波动</n-radio>
                  <n-radio value="auto">自动判断</n-radio>
                </n-space>
              </n-radio-group>
            </n-form-item>
          </n-form>
        </div>

        <!-- 步骤4：下单设置 -->
        <div v-show="current === 4">
          <n-form ref="formRef4" :model="formValue" label-placement="left" label-width="180">
            <n-form-item label="交易对" path="symbol">
              <n-input v-model:value="formValue.symbol" placeholder="例如: BTC_USDT" />
            </n-form-item>
            <n-form-item label="订单类型">
              <n-select v-model:value="formValue.orderType" :options="orderTypeOptions" />
            </n-form-item>
            <n-form-item label="保证金模式">
              <n-select v-model:value="formValue.marginMode" :options="marginModeOptions" />
            </n-form-item>
            <n-form-item label="杠杆倍数">
              <n-input-number v-model:value="formValue.leverage" :min="1" :max="125" style="width: 100%;" />
            </n-form-item>
            <n-form-item label="使用保证金比例 (%)">
              <n-input-number v-model:value="formValue.marginPercent" :min="1" :max="100" style="width: 100%;" />
            </n-form-item>
            <n-form-item label="启用反向下单策略">
              <n-switch v-model:value="formValue.enableReverse" />
            </n-form-item>
          </n-form>
        </div>

        <!-- 步骤5：自动平仓 -->
        <div v-show="current === 5">
          <n-form ref="formRef5" :model="formValue" label-placement="left" label-width="220">
            <n-form-item label="单个订单最大止损百分比 (%)">
              <n-input-number v-model:value="formValue.stopLossPercent" :min="0" :max="100" style="width: 100%;" />
            </n-form-item>
            <n-form-item label="单个订单止盈回撤百分比 (%)">
              <n-input-number v-model:value="formValue.takeProfitCallbackPercent" :min="0" :max="100" style="width: 100%;" />
            </n-form-item>
            <n-form-item label="启动止盈回撤的百分比 (%)">
              <n-input-number v-model:value="formValue.activateCallbackPercent" :min="0" :max="100" style="width: 100%;" />
            </n-form-item>
          </n-form>
        </div>
      </n-card>

      <!-- 操作按钮 -->
      <n-space justify="end" style="margin-top: 20px;">
        <n-button v-if="current > 1" @click="handlePrevious">
          上一步
        </n-button>
        <n-button v-if="current < 5" type="primary" @click="handleNext">
          下一步
        </n-button>
        <n-button v-if="current === 5" type="primary" :loading="submitting" @click="handleSubmit">
          创建机器人
        </n-button>
        <n-button @click="handleCancel">
          取消
        </n-button>
      </n-space>
    </n-card>
  </div>
</template>

<script lang="ts" setup>
import { ref, reactive, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import {
  NCard,
  NSteps,
  NStep,
  NForm,
  NFormItem,
  NInput,
  NInputNumber,
  NSelect,
  NRadioGroup,
  NRadio,
  NSwitch,
  NSpace,
  NButton,
  useMessage,
} from 'naive-ui';
import { createRobot } from '@/api/trading/robot';
import { getApiConfigList } from '@/api/trading/api-config';

const router = useRouter();
const message = useMessage();

const current = ref(1);
const currentStatus = ref<'process' | 'finish' | 'error' | 'wait'>('process');
const submitting = ref(false);

const formValue = reactive({
  apiConfigId: null,
  name: '',
  maxProfit: 1000,
  maxLoss: 500,
  maxRuntime: 86400,
  riskPreference: 'balanced',
  marketState: 'auto',
  symbol: 'BTC_USDT',
  orderType: 'market',
  marginMode: 'isolated',
  leverage: 10,
  marginPercent: 50,
  enableReverse: false,
  stopLossPercent: 2,
  takeProfitCallbackPercent: 1,
  activateCallbackPercent: 3,
});

const rules = {
  apiConfigId: { required: true, message: '请选择API接口', trigger: 'change' },
  name: { required: true, message: '请输入机器人名称', trigger: 'blur' },
  maxProfit: { required: true, type: 'number', message: '请输入最大盈利目标', trigger: 'blur' },
  maxLoss: { required: true, type: 'number', message: '请输入最大亏损额', trigger: 'blur' },
  maxRuntime: { required: true, type: 'number', message: '请输入最大运行时长', trigger: 'blur' },
  symbol: { required: true, message: '请输入交易对', trigger: 'blur' },
};

const apiConfigOptions = ref([]);
const orderTypeOptions = [
  { label: '市价单', value: 'market' },
  { label: '限价单', value: 'limit' },
];
const marginModeOptions = [
  { label: '逐仓', value: 'isolated' },
  { label: '全仓', value: 'cross' },
];

// 加载API配置列表
const loadApiConfigs = async () => {
  try {
    const res = await getApiConfigList();
    apiConfigOptions.value = (res.list || []).map((item: any) => ({
      label: `${item.platformName} - ${item.name}`,
      value: item.id,
    }));
  } catch (error: any) {
    message.error(error.message || '加载API配置失败');
  }
};

// 下一步
const handleNext = () => {
  if (current.value < 5) {
    current.value++;
  }
};

// 上一步
const handlePrevious = () => {
  if (current.value > 1) {
    current.value--;
  }
};

// 提交
const handleSubmit = async () => {
  submitting.value = true;
  try {
    await createRobot(formValue);
    message.success('机器人创建成功');
    router.push('/trading/robot');
  } catch (error: any) {
    message.error(error.message || '创建失败');
  } finally {
    submitting.value = false;
  }
};

// 取消
const handleCancel = () => {
  router.back();
};

onMounted(() => {
  loadApiConfigs();
});
</script>

