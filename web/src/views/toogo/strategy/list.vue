<template>
  <div class="strategy-list">
    <n-card :bordered="false" class="proCard">
      <template #header>
        <n-space align="center">
          <n-button text @click="goBack">
            <template #icon><n-icon><ArrowLeftOutlined /></n-icon></template>
          </n-button>
          <span>{{ groupName }} - 策略管理</span>
          <n-tag v-if="isReadonly" type="warning" size="small">只读模式</n-tag>
        </n-space>
      </template>
      <template #header-extra>
        <n-space v-if="!isReadonly">
          <n-button @click="initAllStrategies" :disabled="strategies.length >= 12">
            <template #icon><n-icon><ThunderboltOutlined /></n-icon></template>
            批量初始化
          </n-button>
          <n-button type="primary" @click="openCreateDrawer">
            <template #icon><n-icon><PlusOutlined /></n-icon></template>
            添加策略
          </n-button>
        </n-space>
        <n-text v-else depth="3">官方策略模板（预览）</n-text>
      </template>

      <!-- 策略矩阵视图 -->
      <n-alert v-if="strategies.length < 12 && !isReadonly" type="warning" style="margin-bottom: 16px">
        当前仅有 {{ strategies.length }}/12 种策略，建议点击"批量初始化"补全所有市场状态和风险偏好组合。
      </n-alert>

      <n-tabs type="segment" v-model:value="activeMarket" animated style="margin-bottom: 16px">
        <n-tab-pane name="all" tab="全部" />
        <n-tab-pane name="trend" tab="📈 趋势市场" />
        <n-tab-pane name="range" tab="📊 震荡市场" />
        <n-tab-pane name="high_vol" tab="🔥 高波动" />
        <n-tab-pane name="low_vol" tab="💤 低波动" />
      </n-tabs>

      <n-grid :cols="3" :x-gap="16" :y-gap="16">
        <n-gi v-for="strategy in filteredStrategies" :key="strategy.id">
          <n-card hoverable class="strategy-card">
            <template #header>
              <n-space align="center" justify="space-between" style="width: 100%">
                <n-space align="center">
                  <span class="strategy-name">{{ strategy.strategyName }}</span>
                </n-space>
                <n-tag :type="getRiskType(strategy.riskPreference)" size="small">
                  {{ getRiskLabel(strategy.riskPreference) }}
                </n-tag>
              </n-space>
            </template>

            <n-space vertical :size="8">
              <n-space>
                <n-tag size="small" :bordered="false">{{ getMarketLabel(strategy.marketState) }}</n-tag>
              </n-space>

              <n-descriptions :column="2" size="small" label-placement="left">
                <n-descriptions-item label="时间窗口">
                  <n-text type="info">{{ strategy.monitorWindow || 0 }}秒</n-text>
                </n-descriptions-item>
                <n-descriptions-item label="波动值">
                  <n-text type="warning">{{ strategy.volatilityThreshold || 0 }}U</n-text>
                </n-descriptions-item>
                <n-descriptions-item label="杠杆">
                  <n-text type="info">{{ strategy.leverage || 0 }}x</n-text>
                </n-descriptions-item>
                <n-descriptions-item label="保证金">
                  <n-text type="info">{{ strategy.marginPercent || 0 }}%</n-text>
                </n-descriptions-item>
                <n-descriptions-item label="止损">
                  <n-text type="error">{{ strategy.stopLossPercent || 0 }}%</n-text>
                </n-descriptions-item>
                <n-descriptions-item label="启动止盈">
                  <n-text type="success">{{ strategy.autoStartRetreatPercent || 0 }}%</n-text>
                </n-descriptions-item>
                <n-descriptions-item label="止盈回撤">
                  <n-text type="warning">{{ strategy.profitRetreatPercent || 0 }}%</n-text>
                </n-descriptions-item>
              </n-descriptions>

              <n-text v-if="strategy.description" depth="3" style="font-size: 11px; line-height: 1.4">
                {{ strategy.description }}
              </n-text>
            </n-space>

            <template #footer>
              <n-space v-if="!isReadonly">
                <n-button size="small" quaternary @click="editStrategy(strategy)">
                  <template #icon><n-icon><EditOutlined /></n-icon></template>
                </n-button>
                <n-popconfirm @positive-click="deleteStrategy(strategy)">
                  <template #trigger>
                    <n-button size="small" quaternary type="error">
                      <template #icon><n-icon><DeleteOutlined /></n-icon></template>
                    </n-button>
                  </template>
                  确定删除此策略吗？
                </n-popconfirm>
              </n-space>
              <n-text v-else depth="3" style="font-size: 12px">💡 添加到我的策略后可修改</n-text>
            </template>
          </n-card>
        </n-gi>

        <n-gi v-if="filteredStrategies.length === 0" :span="3">
          <n-empty :description="`暂无${activeMarket === 'all' ? '' : getMarketLabel(activeMarket)}策略`">
            <template #extra v-if="!isReadonly">
              <n-button type="primary" @click="openCreateDrawer">添加策略</n-button>
            </template>
          </n-empty>
        </n-gi>
      </n-grid>
    </n-card>

    <!-- 创建/编辑策略抽屉 -->
    <n-drawer v-model:show="showDrawer" :width="700" placement="right">
      <n-drawer-content :title="editingStrategy ? '编辑策略' : '添加策略'" closable>
        <n-form ref="formRef" :model="formData" :rules="rules" label-placement="left" label-width="110">
          <!-- 策略标识（策略名称、市场状态、风险偏好） -->
          <n-grid :cols="2" :x-gap="16" style="margin-bottom: 16px">
            <n-gi :span="2">
              <n-form-item label="策略名称" path="strategyName">
                <n-input v-model:value="formData.strategyName" placeholder="如：趋势市场-保守型" />
              </n-form-item>
            </n-gi>
            <n-gi>
              <n-form-item label="市场状态" path="marketState">
                <n-select v-model:value="formData.marketState" :options="marketOptions" />
              </n-form-item>
            </n-gi>
            <n-gi>
              <n-form-item label="风险偏好" path="riskPreference">
                <n-select v-model:value="formData.riskPreference" :options="riskOptions" />
              </n-form-item>
            </n-gi>
          </n-grid>

          <n-card title="杠杆与仓位" size="small" :bordered="true" style="margin-bottom: 16px">
            <n-grid :cols="2" :x-gap="16">
              <n-gi>
                <n-form-item label="杠杆倍数" path="leverage">
                  <n-input-number v-model:value="formData.leverage" :min="1" :max="125" :show-button="false" style="width: 100%">
                    <template #suffix>x</template>
                  </n-input-number>
                </n-form-item>
              </n-gi>
              <n-gi>
                <n-form-item label="保证金比例" path="marginPercent">
                  <n-input-number v-model:value="formData.marginPercent" :min="1" :max="100" :precision="2" :show-button="false" style="width: 100%">
                    <template #suffix>%</template>
                  </n-input-number>
                </n-form-item>
              </n-gi>
            </n-grid>
          </n-card>

          <n-card title="行情监控" size="small" :bordered="true" style="margin-bottom: 16px">
            <n-grid :cols="2" :x-gap="16">
              <n-gi>
                <n-form-item label="监控窗口" path="monitorWindow">
                  <n-input-number v-model:value="formData.monitorWindow" :min="60" :max="3600" :step="60" :show-button="false" style="width: 100%">
                    <template #suffix>秒</template>
                  </n-input-number>
                </n-form-item>
              </n-gi>
              <n-gi>
                <n-form-item label="波动阈值" path="volatilityThreshold">
                  <n-input-number v-model:value="formData.volatilityThreshold" :min="0.1" :max="500" :precision="2" :show-button="false" style="width: 100%">
                    <template #suffix>U</template>
                  </n-input-number>
                </n-form-item>
              </n-gi>
            </n-grid>
          </n-card>

          <n-card title="止损止盈" size="small" :bordered="true" style="margin-bottom: 16px">
            <n-grid :cols="3" :x-gap="16">
              <n-gi>
                <n-form-item label="止损" path="stopLossPercent">
                  <n-input-number v-model:value="formData.stopLossPercent" :min="0.5" :max="50" :precision="2" :show-button="false" style="width: 100%">
                    <template #suffix>%</template>
                  </n-input-number>
                </n-form-item>
              </n-gi>
              <n-gi>
                <n-form-item label="启动止盈" path="autoStartRetreatPercent">
                  <n-input-number v-model:value="formData.autoStartRetreatPercent" :min="0.5" :max="100" :precision="2" :show-button="false" style="width: 100%">
                    <template #suffix>%</template>
                  </n-input-number>
                </n-form-item>
              </n-gi>
              <n-gi>
                <n-form-item label="止盈回撤" path="profitRetreatPercent">
                  <n-input-number v-model:value="formData.profitRetreatPercent" :min="5" :max="100" :precision="2" :show-button="false" style="width: 100%">
                    <template #suffix>%</template>
                  </n-input-number>
                </n-form-item>
              </n-gi>
            </n-grid>
          </n-card>


          <n-form-item label="描述">
            <n-input v-model:value="formData.description" type="textarea" :rows="2" placeholder="描述此策略..." maxlength="500" />
          </n-form-item>
        </n-form>

        <template #footer>
          <n-space justify="end">
            <n-button @click="showDrawer = false">取消</n-button>
            <n-button type="primary" @click="handleSubmit" :loading="submitLoading">
              {{ editingStrategy ? '保存' : '添加' }}
            </n-button>
          </n-space>
        </template>
      </n-drawer-content>
    </n-drawer>

  </div>
</template>

<script lang="ts" setup>
import { ref, computed, onMounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useMessage, useDialog } from 'naive-ui';
import { PlusOutlined, ArrowLeftOutlined, EditOutlined, DeleteOutlined, ThunderboltOutlined, InfoCircleOutlined } from '@vicons/antd';
import { http } from '@/utils/http/axios';

const route = useRoute();
const router = useRouter();
const message = useMessage();
const dialog = useDialog();

// 路由参数
const groupId = computed(() => Number(route.query.groupId) || 0);
const groupName = computed(() => (route.query.groupName as string) || '策略管理');
// 只有在明确传递readonly=1时才只读，管理员从官方策略组管理页面进入时允许修改
const isReadonly = computed(() => route.query.readonly === '1');

// 状态
const strategies = ref<any[]>([]);
const activeMarket = ref('all');
const showDrawer = ref(false);
const editingStrategy = ref<any>(null);
const formRef = ref<any>(null);
const submitLoading = ref(false);

// 表单数据
const formData = ref({
  strategyName: '',
  marketState: 'trend',
  riskPreference: 'balanced',
  monitorWindow: 300,
  volatilityThreshold: 100,
  leverage: 8,           // 杠杆倍数
  marginPercent: 12,     // 保证金比例
  stopLossPercent: 5,
  autoStartRetreatPercent: 3,
  profitRetreatPercent: 25,
  description: '',
});

// 选项
const marketOptions = [
  { label: '📈 趋势市场', value: 'trend' },
  { label: '📊 震荡市场', value: 'volatile' },
  { label: '🔥 高波动', value: 'high_vol' },
  { label: '💤 低波动', value: 'low_vol' },
];

const riskOptions = [
  { label: '🛡️ 保守型', value: 'conservative' },
  { label: '⚖️ 平衡型', value: 'balanced' },
  { label: '🚀 激进型', value: 'aggressive' },
];

// 表单规则
const rules = {
  strategyName: { required: true, message: '请输入策略名称', trigger: 'blur' },
  marketState: { required: true, message: '请选择市场状态', trigger: 'change' },
  riskPreference: { required: true, message: '请选择风险偏好', trigger: 'change' },
};

// 筛选后的策略
const filteredStrategies = computed(() => {
  if (activeMarket.value === 'all') return strategies.value;
  return strategies.value.filter((s) => s.marketState === activeMarket.value);
});

// 辅助函数
function getMarketLabel(value: string) {
  const map: Record<string, string> = {
    trend: '📈 趋势',
    range: '📊 震荡',
    volatile: '📊 震荡',
    high_vol: '🔥 高波动',
    low_vol: '💤 低波动',
  };
  return map[value] || value;
}

function getRiskLabel(value: string) {
  const map: Record<string, string> = {
    conservative: '🛡️ 保守型',
    balanced: '⚖️ 平衡型',
    aggressive: '🚀 激进型',
  };
  return map[value] || value;
}

function getRiskType(value: string): any {
  const map: Record<string, string> = {
    conservative: 'success',
    balanced: 'info',
    aggressive: 'error',
  };
  return map[value] || 'default';
}

// 返回
function goBack() {
  try {
    router.push('/toogo/strategy');
  } catch (error) {
    console.error('路由跳转失败:', error);
    window.location.href = '/toogo/strategy';
  }
}

// 加载策略列表
async function loadStrategies() {
  if (!groupId.value) {
    console.warn('groupId为空，无法加载策略');
    return;
  }
  try {
    console.log('加载策略列表，groupId:', groupId.value);
    const res = await http.request({
      url: '/strategy/template/list',
      method: 'get',
      params: { groupId: groupId.value, pageSize: 100 },
    });
    console.log('策略列表响应:', res);
    strategies.value = res?.list || [];
    console.log('策略数量:', strategies.value.length);
  } catch (error: any) {
    console.error('加载策略失败:', error);
    // 不抛出错误，避免页面崩溃
    strategies.value = [];
  }
}


// 批量初始化
function initAllStrategies() {
  if (isReadonly.value) return;
  dialog.warning({
    title: '批量初始化策略',
    content: '将为此模板创建12种策略组合（4种市场状态 × 3种风险偏好），已存在的组合将跳过。是否继续？',
    positiveText: '确定',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await http.request({
          url: '/strategy/group/initStrategies',
          method: 'post',
          data: { groupId: groupId.value, useDefault: true },
        });
        message.success('初始化成功');
        loadStrategies();
      } catch (error: any) {
        message.error(error?.response?.data?.message || error?.message || '初始化失败');
      }
    },
  });
}

// 打开创建抽屉
function openCreateDrawer() {
  editingStrategy.value = null;
  formData.value = {
    strategyName: '',
    marketState: 'trend',
    riskPreference: 'balanced',
    monitorWindow: 300,
    volatilityThreshold: 100,
    leverage: 8,
    marginPercent: 12,
    stopLossPercent: 5,
    autoStartRetreatPercent: 3,
    profitRetreatPercent: 25,
    description: '',
  };
  showDrawer.value = true;
}

// 编辑策略
function editStrategy(strategy: any) {
  editingStrategy.value = strategy;
  
  console.log('编辑策略 - 原始数据:', strategy);
  
  // 解析 config_json
  let configJson: any = {};
  try {
    configJson = strategy.configJson ? (typeof strategy.configJson === 'string' ? JSON.parse(strategy.configJson) : strategy.configJson) : {};
  } catch (error) {
    console.warn('解析 configJson 失败:', error);
  }

  console.log('编辑策略 - configJson:', configJson);

  // 多层级读取策略：优先 strategy 直接字段 → configJson → 默认值
  // 统一市场状态：range 转换为 volatile（兼容旧数据）
  let marketState = strategy.marketState || 'trend';
  if (marketState === 'range') {
    marketState = 'volatile';
  }
  
  formData.value = {
    strategyName: strategy.strategyName || '',
    marketState: marketState,
    riskPreference: strategy.riskPreference || 'balanced',
    monitorWindow: strategy.monitorWindow || configJson.monitorWindow || 300,
    volatilityThreshold: parseFloat(strategy.volatilityThreshold) || configJson.volatilityThreshold || 100,
    // 杠杆和保证金：优先从 strategy 直接字段读取，然后 configJson，最后默认值
    leverage: strategy.leverage || configJson.leverage || strategy.leverageMin || 8,
    marginPercent: strategy.marginPercent || configJson.marginPercent || parseFloat(strategy.marginPercentMin) || 12,
    // 止损止盈：优先从 strategy 直接字段读取
    stopLossPercent: parseFloat(strategy.stopLossPercent) || configJson.stopLossPercent || 5,
    autoStartRetreatPercent: parseFloat(strategy.autoStartRetreatPercent) || configJson.autoStartRetreatPercent || 3,
    profitRetreatPercent: parseFloat(strategy.profitRetreatPercent) || configJson.profitRetreatPercent || 25,
    description: strategy.description || '',
  };
  
  console.log('编辑策略 - formData:', formData.value);
  
  showDrawer.value = true;
}

// 删除策略
async function deleteStrategy(strategy: any) {
  if (isReadonly.value) return;
  try {
    await http.request({ url: '/strategy/template/delete', method: 'post', data: { id: strategy.id } });
    message.success('删除成功');
    loadStrategies();
  } catch (error: any) {
    message.error(error?.response?.data?.message || error?.message || '删除失败');
  }
}

// 应用策略

// 提交表单
async function handleSubmit(confirmed = false) {
  try {
    await formRef.value?.validate();
  } catch {
    return;
  }

  submitLoading.value = true;
  try {
    // 构建完整的 config_json（包含策略模板参数）
    const configJson = JSON.stringify({
      // 杠杆和保证金
      leverage: formData.value.leverage,
      marginPercent: formData.value.marginPercent,
      // 止损止盈
      stopLossPercent: formData.value.stopLossPercent,
      autoStartRetreatPercent: formData.value.autoStartRetreatPercent,
      profitRetreatPercent: formData.value.profitRetreatPercent,
      // 行情监控
      monitorWindow: formData.value.monitorWindow,
      volatilityThreshold: formData.value.volatilityThreshold,
    });

    const submitData = {
      groupId: groupId.value,
      strategyKey: `${groupId.value}_${formData.value.marketState}_${formData.value.riskPreference}`,
      strategyName: formData.value.strategyName,
      marketState: formData.value.marketState,
      riskPreference: formData.value.riskPreference,
      monitorWindow: formData.value.monitorWindow,
      volatilityThreshold: formData.value.volatilityThreshold,
      // 保留范围字段为兼容性，但使用推荐值填充
      leverageMin: formData.value.leverage,
      leverageMax: formData.value.leverage,
      marginPercentMin: formData.value.marginPercent,
      marginPercentMax: formData.value.marginPercent,
      stopLossPercent: formData.value.stopLossPercent,
      autoStartRetreatPercent: formData.value.autoStartRetreatPercent,
      profitRetreatPercent: formData.value.profitRetreatPercent,
      configJson,
      description: formData.value.description,
    };

    const url = editingStrategy.value ? '/strategy/template/update' : '/strategy/template/create';
    const data = editingStrategy.value ? { ...submitData, id: editingStrategy.value.id, confirmed } : submitData;

    await http.request({ url, method: 'post', data });
    message.success(editingStrategy.value ? '更新成功' : '添加成功');
    showDrawer.value = false;
    loadStrategies();
  } catch (error: any) {
    const errorMsg = error?.response?.data?.message || error?.message || '操作失败';
    
    // 检查是否是需要确认的错误（包含"绑定"关键字，且是编辑操作，且未确认）
    if (errorMsg.includes('绑定') && editingStrategy.value && !confirmed) {
      // 显示确认对话框（阻止 finally 中的 submitLoading 重置，让对话框保持显示）
      submitLoading.value = false; // 先重置 loading，让对话框可以正常显示
      dialog.warning({
        title: '确认修改策略模板',
        content: errorMsg,
        positiveText: '确认修改',
        negativeText: '取消',
        onPositiveClick: async () => {
          // 用户确认后，重新提交并传递 confirmed=true
          await handleSubmit(true);
        },
      });
      return; // 提前返回，避免执行 finally 中的代码
    } else {
      message.error(errorMsg);
    }
  } finally {
    submitLoading.value = false;
  }
}

onMounted(() => {
  loadStrategies();
});
</script>

<style scoped lang="less">
.strategy-list {
  .strategy-card {
    :deep(.n-card-header) {
      padding: 12px 16px;
    }
    :deep(.n-card__content) {
      padding: 12px 16px;
    }
    :deep(.n-card__footer) {
      padding: 12px 16px;
      border-top: 1px solid #f0f0f0;
    }

    .strategy-name {
      font-weight: 600;
      font-size: 14px;
    }
  }
}
</style>

