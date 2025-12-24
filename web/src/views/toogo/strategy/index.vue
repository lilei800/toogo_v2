<template>
  <div class="strategy-page">
    <!-- 页面标题 -->
    <n-card :bordered="false" size="small" class="mb-4">
      <n-space justify="space-between" align="center">
        <n-space align="center" :size="12">
          <n-text strong style="font-size: 18px">📋 策略模板</n-text>
          <n-text depth="3" style="font-size: 13px">机器人运行时根据市场状态自动匹配策略</n-text>
        </n-space>
        <n-button type="primary" @click="openCreateModal" v-if="activeTab === 'my'">
          <template #icon><n-icon :component="PlusOutlined" /></template>
          创建策略组
        </n-button>
      </n-space>
    </n-card>

    <!-- Tab切换 -->
    <n-card :bordered="false">
      <n-tabs v-model:value="activeTab" type="segment" animated>
        <!-- 我的策略 -->
        <n-tab-pane name="my" tab="📂 我的策略">
          <div class="tab-content">
            <n-spin :show="loadingMy">
              <div v-if="myStrategies.length > 0">
                <n-grid :cols="3" :x-gap="16" :y-gap="16">
                  <n-gi v-for="group in myStrategies" :key="group.id">
                    <n-card hoverable size="small" class="strategy-card">
                      <template #header>
                        <n-space align="center">
                          <n-text strong>{{ group.groupName }}</n-text>
                          <n-tag v-if="group.isDefault" type="success" size="small">默认</n-tag>
                        </n-space>
                      </template>
                      <template #header-extra>
                        <n-dropdown :options="getCardActions(group)" @select="(key) => handleAction(key, group)">
                          <n-button quaternary circle size="small">
                            <template #icon><n-icon :component="MoreOutlined" /></template>
                          </n-button>
                        </n-dropdown>
                      </template>
                      
                      <n-space vertical :size="8">
                        <n-space align="center" :size="8">
                          <n-tag type="info" size="small">{{ getExchangeLabel(group.exchange) }}</n-tag>
                          <n-tag size="small">{{ group.symbol }}</n-tag>
                          <n-tag :type="(group.strategyCount || 0) === 12 ? 'success' : (group.strategyCount || 0) > 0 ? 'warning' : 'default'" size="small">
                            {{ group.strategyCount || 0 }}/12种策略
                          </n-tag>
                        </n-space>
                        <n-space :size="4">
                          <n-tag size="tiny" :bordered="false">{{ group.orderType === 'market' ? '市价单' : '限价单' }}</n-tag>
                          <n-tag size="tiny" :bordered="false">{{ group.marginMode === 'isolated' ? '逐仓' : '全仓' }}</n-tag>
                        </n-space>
                        <n-text depth="3" style="font-size: 12px">
                          {{ group.description || '包含趋势/震荡/高低波动市场的多种策略配置' }}
                        </n-text>
                      </n-space>
                      
                      <template #action>
                        <n-space justify="space-between">
                          <n-space :size="8">
                            <n-button size="small" @click="viewStrategies(group)">查看策略</n-button>
                            <n-button size="small" @click="editGroup(group)">修改</n-button>
                            <n-button size="small" type="warning" @click="openInitModal(group)" v-if="(group.strategyCount || 0) < 12">
                              初始化
                            </n-button>
                          </n-space>
                          <n-button type="primary" size="small" @click="createRobotWithStrategy(group)">
                            使用此策略
                          </n-button>
                        </n-space>
                      </template>
                    </n-card>
                  </n-gi>
                </n-grid>
              </div>
              <n-empty v-else description="暂无策略组，可从官方策略添加或自行创建">
                <template #extra>
                  <n-space>
                    <n-button @click="activeTab = 'official'">浏览官方策略</n-button>
                    <n-button type="primary" @click="openCreateModal">创建策略组</n-button>
                  </n-space>
                </template>
              </n-empty>
            </n-spin>
          </div>
        </n-tab-pane>

        <!-- 官方策略 -->
        <n-tab-pane name="official" tab="⭐ 官方策略">
          <div class="tab-content">
            <n-spin :show="loadingOfficial">
              <div v-if="officialStrategies.length > 0">
                <n-grid :cols="3" :x-gap="16" :y-gap="16">
                  <n-gi v-for="group in officialStrategies" :key="group.id">
                    <n-card hoverable size="small" class="strategy-card official">
                      <template #header>
                        <n-space align="center">
                          <n-tag type="warning" size="small">官方</n-tag>
                          <n-text strong>{{ group.groupName }}</n-text>
                        </n-space>
                      </template>
                      
                      <n-space vertical :size="8">
                        <n-space align="center" :size="8">
                          <n-tag type="info" size="small">{{ getExchangeLabel(group.exchange) }}</n-tag>
                          <n-tag size="small">{{ group.symbol }}</n-tag>
                          <n-tag type="success" size="small">{{ group.strategyCount || 12 }}种策略</n-tag>
                        </n-space>
                        <n-space :size="4">
                          <n-tag size="tiny" :bordered="false">{{ group.orderType === 'market' ? '市价单' : '限价单' }}</n-tag>
                          <n-tag size="tiny" :bordered="false">{{ group.marginMode === 'isolated' ? '逐仓' : '全仓' }}</n-tag>
                        </n-space>
                        <n-text depth="3" style="font-size: 12px">
                          {{ group.description || '官方精选策略模板，经过验证和优化' }}
                        </n-text>
                      </n-space>
                      
                      <template #action>
                        <n-space justify="space-between">
                          <n-button size="small" @click="viewStrategies(group)">查看详情</n-button>
                          <n-button type="primary" size="small" @click="addToMy(group)">
                            添加到我的
                          </n-button>
                        </n-space>
                      </template>
                    </n-card>
                  </n-gi>
                </n-grid>
              </div>
              <n-empty v-else description="暂无官方策略" />
            </n-spin>
          </div>
        </n-tab-pane>

        <!-- 盈利排行 -->
        <n-tab-pane name="ranking" tab="🏆 盈利排行">
          <div class="tab-content">
            <n-empty description="盈利排行策略功能开发中，敬请期待">
              <template #icon>
                <n-icon :size="48" :component="TrophyOutlined" color="#faad14" />
              </template>
              <template #extra>
                <n-text depth="3">即将推出：跟随高盈利用户的策略配置</n-text>
              </template>
            </n-empty>
          </div>
        </n-tab-pane>
      </n-tabs>
    </n-card>

    <!-- 创建/编辑策略组弹窗 -->
    <n-modal v-model:show="showCreateModal" preset="card" :title="editingGroup ? '编辑策略组' : '创建策略组'" style="width: 600px">
      <n-alert type="info" style="margin-bottom: 16px">
        创建后可以批量初始化12种策略（4种市场状态 × 3种风险偏好），机器人会根据市场自动匹配最优策略
      </n-alert>
      
      <n-form ref="formRef" :model="formData" :rules="formRules" label-placement="left" label-width="100">
        <n-form-item label="策略组名称" path="groupName">
          <n-input v-model:value="formData.groupName" placeholder="如：BTC-USDT 高频策略 V1.0" maxlength="50" />
        </n-form-item>
        <n-form-item label="策略组标识" path="groupKey">
          <n-input v-model:value="formData.groupKey" placeholder="唯一标识，如：my_btc_usdt_v1（留空自动生成）" :disabled="!!editingGroup" />
        </n-form-item>
        <n-grid :cols="2" :x-gap="16">
          <n-gi>
            <n-form-item label="交易平台" path="exchange">
              <n-select v-model:value="formData.exchange" :options="exchangeOptions" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="交易对" path="symbol">
              <n-select v-model:value="formData.symbol" :options="symbolOptions" filterable tag />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="订单类型" path="orderType">
              <n-select v-model:value="formData.orderType" :options="orderTypeOptions" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="保证金模式" path="marginMode">
              <n-select v-model:value="formData.marginMode" :options="marginModeOptions" />
            </n-form-item>
          </n-gi>
        </n-grid>
        <n-form-item label="描述">
          <n-input v-model:value="formData.description" type="textarea" :rows="3" placeholder="描述此策略组..." maxlength="500" />
        </n-form-item>
        <n-form-item label="排序">
          <n-input-number v-model:value="formData.sort" :min="0" style="width: 100%" />
        </n-form-item>
      </n-form>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showCreateModal = false">取消</n-button>
          <n-button type="primary" @click="handleSubmit" :loading="submitLoading">
            {{ editingGroup ? '保存' : '创建' }}
          </n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- 初始化策略弹窗 -->
    <n-modal v-model:show="showInitModal" preset="card" title="批量初始化策略" style="width: 500px">
      <n-space vertical>
        <n-alert type="warning">
          将为"{{ initGroup?.groupName }}"生成12种策略模板（4种市场状态 × 3种风险偏好）
        </n-alert>
        <n-form label-placement="left" label-width="120">
          <n-form-item label="使用默认参数">
            <n-switch v-model:value="initOptions.useDefault" />
          </n-form-item>
        </n-form>
        <n-text depth="3" style="font-size: 13px">
          默认参数包括：合理的杠杆范围、止损比例、止盈回撤等，适合大多数情况。
        </n-text>
      </n-space>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showInitModal = false">取消</n-button>
          <n-button type="primary" @click="handleInitStrategies" :loading="initLoading">
            开始初始化
          </n-button>
        </n-space>
      </template>
    </n-modal>

  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, h } from 'vue';
import { useRouter } from 'vue-router';
import { useMessage, NTag } from 'naive-ui';
import { http } from '@/utils/http/axios';
import {
  PlusOutlined,
  MoreOutlined,
  TrophyOutlined,
  EditOutlined,
  DeleteOutlined,
  CopyOutlined,
} from '@vicons/antd';

const router = useRouter();
const message = useMessage();

// Tab状态
const activeTab = ref('my');

// 数据
const myStrategies = ref<any[]>([]);
const officialStrategies = ref<any[]>([]);
const currentGroup = ref<any>(null);

// 加载状态
const loadingMy = ref(false);
const loadingOfficial = ref(false);
const submitLoading = ref(false);
const initLoading = ref(false);

// 弹窗
const showCreateModal = ref(false);
const showInitModal = ref(false);
const formRef = ref();
const editingGroup = ref<any>(null);
const initGroup = ref<any>(null);

// 初始化选项
const initOptions = ref({
  useDefault: true,
});

// 策略组表单
const formData = ref({
  groupName: '',
  groupKey: '',
  exchange: 'bitget',
  symbol: 'BTCUSDT',
  orderType: 'market',
  marginMode: 'isolated',
  description: '',
  sort: 100,
});

const formRules = {
  groupName: { required: true, message: '请输入策略组名称', trigger: 'blur' },
  exchange: { required: true, message: '请选择交易平台', trigger: 'change' },
  symbol: { required: true, message: '请选择交易对', trigger: 'change' },
};

// 选项
const exchangeOptions = [
  { label: 'Bitget', value: 'bitget' },
  { label: 'Binance', value: 'binance' },
  { label: 'OKX', value: 'okx' },
  { label: 'Gate', value: 'gate' },
];

const symbolOptions = [
  { label: 'BTCUSDT', value: 'BTCUSDT' },
  { label: 'ETHUSDT', value: 'ETHUSDT' },
  { label: 'BNBUSDT', value: 'BNBUSDT' },
  { label: 'SOLUSDT', value: 'SOLUSDT' },
  { label: 'XRPUSDT', value: 'XRPUSDT' },
  { label: 'DOGEUSDT', value: 'DOGEUSDT' },
];

const orderTypeOptions = [
  { label: '市价单', value: 'market' },
  { label: '限价单', value: 'limit' },
];

const marginModeOptions = [
  { label: '逐仓', value: 'isolated' },
  { label: '全仓', value: 'crossed' },
];

// 卡片操作菜单
const getCardActions = (group: any) => [
  { label: '查看策略', key: 'view', icon: () => h('span', {}, '📋') },
  { label: '编辑', key: 'edit', icon: () => h('span', {}, '✏️') },
  { type: 'divider', key: 'd1' },
  { label: group.isDefault ? '取消默认' : '设为默认', key: 'default', icon: () => h('span', {}, '⭐') },
  { type: 'divider', key: 'd2' },
  { label: '删除', key: 'delete', icon: () => h('span', {}, '🗑️') },
];

// 工具函数
const getExchangeLabel = (exchange: string) => {
  const map: any = { bitget: 'Bitget', binance: 'Binance', okx: 'OKX', gate: 'Gate' };
  return map[exchange] || exchange || 'Bitget';
};

const formatMarketState = (state: string) => {
  const map: Record<string, string> = { trend: '趋势', range: '震荡', high_vol: '高波动', low_vol: '低波动' };
  return map[state] || state;
};

const formatRiskPref = (pref: string) => {
  const map: Record<string, string> = { conservative: '保守', balanced: '平衡', aggressive: '激进' };
  return map[pref] || pref;
};

// 生成策略组标识
const generateGroupKey = () => {
  const timestamp = Date.now().toString(36);
  return `my_${formData.value.symbol.toLowerCase()}_${timestamp}`;
};

// 加载我的策略
const loadMyStrategies = async () => {
  loadingMy.value = true;
  try {
    const res = await http.request({
      url: '/strategy/group/list',
      method: 'get',
      params: { page: 1, pageSize: 100, isOfficial: 0 },
    });
    myStrategies.value = res?.list || [];
  } catch (error) {
    console.error('加载我的策略失败:', error);
  } finally {
    loadingMy.value = false;
  }
};

// 加载官方策略
const loadOfficialStrategies = async () => {
  loadingOfficial.value = true;
  try {
    const res = await http.request({
      url: '/strategy/group/list',
      method: 'get',
      params: { page: 1, pageSize: 100, isOfficial: 1 },
    });
    officialStrategies.value = res?.list || [];
  } catch (error) {
    console.error('加载官方策略失败:', error);
  } finally {
    loadingOfficial.value = false;
  }
};

// 打开创建策略组弹窗
const openCreateModal = () => {
  editingGroup.value = null;
  formData.value = {
    groupName: '',
    groupKey: '',
    exchange: 'bitget',
    symbol: 'BTCUSDT',
    orderType: 'market',
    marginMode: 'isolated',
    description: '',
    sort: 100,
  };
  showCreateModal.value = true;
};

// 提交策略组
const handleSubmit = async () => {
  try {
    await formRef.value?.validate();
  } catch {
    return;
  }

  // 如果没有输入groupKey，自动生成
  if (!formData.value.groupKey) {
    formData.value.groupKey = generateGroupKey();
  }

  submitLoading.value = true;
  try {
    const url = editingGroup.value ? '/strategy/group/update' : '/strategy/group/create';
    const data = { ...formData.value };
    
    if (editingGroup.value) {
      (data as any).id = editingGroup.value.id;
    }

    await http.request({ url, method: 'post', data });
    message.success(editingGroup.value ? '更新成功' : '创建成功');
    showCreateModal.value = false;
    await loadMyStrategies();
  } catch (error: any) {
    message.error(error.message || '保存失败');
  } finally {
    submitLoading.value = false;
  }
};

// 打开初始化策略弹窗
const openInitModal = (group: any) => {
  initGroup.value = group;
  initOptions.value.useDefault = true;
  showInitModal.value = true;
};

// 初始化策略
const handleInitStrategies = async () => {
  if (!initGroup.value) return;

  initLoading.value = true;
  try {
    await http.request({
      url: '/strategy/group/initStrategies',
      method: 'post',
      data: {
        groupId: initGroup.value.id,
        useDefault: initOptions.value.useDefault,
      },
    });
    message.success('策略初始化成功，已生成12种策略模板');
    showInitModal.value = false;
    await loadMyStrategies();
  } catch (error: any) {
    message.error(error.message || '初始化失败');
  } finally {
    initLoading.value = false;
  }
};

// 查看策略列表
// 查看策略列表（跳转到策略管理页面）
const viewStrategies = (group: any) => {
  const isOfficial = group.isOfficial === 1 || group.isOfficial === true;
  router.push({
    path: '/toogo/strategy/list',
    query: {
      groupId: group.id,
      groupName: group.groupName,
      readonly: isOfficial ? '1' : '0',
    },
  });
};

// 添加官方策略到我的策略
const addToMy = async (group: any) => {
  try {
    await http.request({
      url: '/strategy/group/copyFromOfficial',
      method: 'post',
      data: { officialGroupId: group.id },
    });
    message.success('添加成功！');
    await loadMyStrategies();
    activeTab.value = 'my';
  } catch (error: any) {
    const errorMsg = error?.message || '添加失败';
    if (errorMsg.includes('已存在')) {
      message.warning('该策略已存在于我的策略中');
    } else {
      message.error(errorMsg);
    }
  }
};

// 使用策略创建机器人
const createRobotWithStrategy = (group: any) => {
  router.push({ path: '/toogo/robot/create', query: { strategyGroupId: group.id } });
};

// 编辑策略组
const editGroup = (group: any) => {
  editingGroup.value = group;
  formData.value = { ...group };
  showCreateModal.value = true;
};

// 处理卡片菜单操作
const handleAction = async (key: string, group: any) => {
  switch (key) {
    case 'view':
      viewStrategies(group);
      break;
    case 'edit':
      editGroup(group);
      break;
    case 'default':
      await toggleDefault(group);
      break;
    case 'delete':
      await deleteGroup(group);
      break;
  }
};

// 切换默认状态
const toggleDefault = async (group: any) => {
  try {
    await http.request({
      url: '/strategy/group/setDefault',
      method: 'post',
      data: { id: group.id },
    });
    message.success(group.isDefault ? '已取消默认' : '已设为默认');
    await loadMyStrategies();
  } catch (error: any) {
    message.error(error?.message || '操作失败');
  }
};

// 删除策略组
const deleteGroup = async (group: any) => {
  try {
    await http.request({
      url: '/strategy/group/delete',
      method: 'post',
      data: { id: group.id },
    });
    message.success('删除成功');
    await loadMyStrategies();
  } catch (error: any) {
    message.error(error?.message || '删除失败');
  }
};

onMounted(() => {
  loadMyStrategies();
  loadOfficialStrategies();
});
</script>

<style scoped lang="less">
.strategy-page {
  padding: 16px;
}

.tab-content {
  padding: 16px 0;
  min-height: 300px;
}

.strategy-card {
  transition: all 0.3s;
  
  &:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  }
  
  &.official {
    border-top: 3px solid #faad14;
  }
}

.mb-4 {
  margin-bottom: 16px;
}
</style>
