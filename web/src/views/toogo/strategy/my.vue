<template>
  <div class="my-strategy">
    <!-- 新手引导 -->
    <n-card class="guide-card" :bordered="false" style="margin-bottom: 16px" v-if="!hasStrategy">
      <n-space align="center" justify="space-between">
        <n-space align="center" :size="16">
          <div class="guide-icon">
            <n-icon size="40" color="#f0a020">
              <BulbOutlined />
            </n-icon>
          </div>
          <div>
            <n-text strong style="font-size: 16px">还没有策略模板？</n-text>
            <n-text depth="3" style="display: block; margin-top: 4px">
              您可以从"官方策略模板"一键添加，或者自己创建专属策略
            </n-text>
          </div>
        </n-space>
        <n-space>
          <n-button @click="$router.push('/toogo/strategy/official')">
            <template #icon><n-icon><StarOutlined /></n-icon></template>
            浏览官方策略
          </n-button>
          <n-button type="primary" @click="openCreateModal">
            <template #icon><n-icon><PlusOutlined /></n-icon></template>
            创建我的策略
          </n-button>
        </n-space>
      </n-space>
    </n-card>

    <n-card title="📋 我的策略模板" :bordered="false" class="proCard">
      <template #header-extra>
        <n-space>
          <n-button @click="$router.push('/toogo/strategy/official')" quaternary>
            <template #icon><n-icon><StarOutlined /></n-icon></template>
            从官方添加
          </n-button>
          <n-button type="primary" @click="openCreateModal">
            <template #icon><n-icon><PlusOutlined /></n-icon></template>
            创建策略模板
          </n-button>
        </n-space>
      </template>

      <!-- 筛选 -->
      <n-form inline label-placement="left" :show-feedback="false" style="margin-bottom: 16px">
        <n-form-item label="交易所">
          <n-select v-model:value="filterExchange" :options="exchangeOptions" placeholder="全部" clearable style="width: 120px" />
        </n-form-item>
        <n-form-item label="交易对">
          <n-input v-model:value="filterSymbol" placeholder="搜索交易对" clearable style="width: 140px" />
        </n-form-item>
        <n-form-item>
          <n-space>
            <n-button type="primary" @click="loadGroups">
              <template #icon><n-icon><SearchOutlined /></n-icon></template>
              查询
            </n-button>
            <n-button @click="resetFilter">重置</n-button>
          </n-space>
        </n-form-item>
      </n-form>

      <!-- 策略模板表格 -->
      <n-spin :show="loading">
        <n-data-table
          :columns="columns"
          :data="groupList"
          :row-key="(row: any) => row.id"
          :pagination="false"
          :bordered="false"
          striped
        />

        <n-empty v-if="!loading && groupList.length === 0" description="暂无策略模板">
          <template #extra>
            <n-space vertical align="center">
              <n-text depth="3">从官方策略一键添加，快速开始交易</n-text>
              <n-button type="primary" @click="$router.push('/toogo/strategy/official')">
                浏览官方策略
              </n-button>
            </n-space>
          </template>
        </n-empty>
      </n-spin>
    </n-card>

    <!-- 创建/编辑模板弹窗 -->
    <n-modal v-model:show="showModal" preset="card" :title="editingGroup ? '编辑策略模板' : '创建策略模板'" style="width: 600px">
      <n-alert type="info" style="margin-bottom: 16px">
        创建后可以为模板添加12种策略（4种市场状态 × 3种风险偏好），机器人会根据市场自动匹配最优策略
      </n-alert>
      
      <n-form ref="formRef" :model="formData" :rules="rules" label-placement="left" label-width="100">
        <n-form-item label="模板名称" path="groupName">
          <n-input v-model:value="formData.groupName" placeholder="如：BTC-USDT保守策略" maxlength="50" />
        </n-form-item>
        <n-form-item label="模板标识" path="groupKey">
          <n-input v-model:value="formData.groupKey" placeholder="唯一标识，如：my_btc_usdt" :disabled="!!editingGroup" />
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
          <n-input v-model:value="formData.description" type="textarea" :rows="3" placeholder="描述此策略模板..." maxlength="500" />
        </n-form-item>
      </n-form>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showModal = false">取消</n-button>
          <n-button type="primary" @click="handleSubmit" :loading="submitLoading">
            {{ editingGroup ? '保存' : '创建' }}
          </n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<script lang="ts" setup>
import { ref, computed, onMounted, h } from 'vue';
import { useRouter } from 'vue-router';
import { useMessage, NButton, NSpace, NTag, NPopconfirm } from 'naive-ui';
import { 
  PlusOutlined, SearchOutlined, EditOutlined, DeleteOutlined, 
  BulbOutlined, StarOutlined, EyeOutlined 
} from '@vicons/antd';
import { http } from '@/utils/http/axios';

const router = useRouter();
const message = useMessage();

// 状态
const loading = ref(false);
const groupList = ref<any[]>([]);
const showModal = ref(false);
const editingGroup = ref<any>(null);
const formRef = ref<any>(null);
const submitLoading = ref(false);

// 筛选
const filterExchange = ref<string | null>(null);
const filterSymbol = ref('');

// 是否有策略
const hasStrategy = computed(() => groupList.value.length > 0);

// 表单数据
const formData = ref({
  groupName: '',
  groupKey: '',
  exchange: 'bitget',
  symbol: 'BTC-USDT',
  orderType: 'market',
  marginMode: 'isolated',
  description: '',
});

// 选项
const exchangeOptions = [
  { label: 'Bitget', value: 'bitget' },
  { label: 'Binance', value: 'binance' },
  { label: 'OKX', value: 'okx' },
  { label: 'Gate.io', value: 'gateio' },
];

const symbolOptions = [
  { label: 'BTC/USDT', value: 'BTC-USDT' },
  { label: 'ETH/USDT', value: 'ETH-USDT' },
  { label: 'BNB/USDT', value: 'BNB-USDT' },
  { label: 'SOL/USDT', value: 'SOL-USDT' },
  { label: 'XRP/USDT', value: 'XRP-USDT' },
  { label: 'DOGE/USDT', value: 'DOGE-USDT' },
];

const orderTypeOptions = [
  { label: '市价单', value: 'market' },
  { label: '限价单', value: 'limit' },
];

const marginModeOptions = [
  { label: '逐仓', value: 'isolated' },
  { label: '全仓', value: 'cross' },
];

// 表单规则
const rules = {
  groupName: { required: true, message: '请输入模板名称', trigger: 'blur' },
  groupKey: { required: true, message: '请输入模板标识', trigger: 'blur' },
  exchange: { required: true, message: '请选择交易平台', trigger: 'change' },
  symbol: { required: true, message: '请选择交易对', trigger: 'change' },
};

// 辅助函数
function getExchangeLabel(exchange: string) {
  const labels: Record<string, string> = { bitget: 'Bitget', binance: 'Binance', okx: 'OKX', gateio: 'Gate.io' };
  return labels[exchange] || exchange || 'Bitget';
}

function getTradeTypeLabel(tradeType: string) {
  if (!tradeType || tradeType === 'perpetual' || tradeType === 'market') return '永续合约';
  if (tradeType === 'delivery') return '交割合约';
  if (tradeType === 'spot') return '现货';
  return tradeType;
}

function resetFilter() {
  filterExchange.value = null;
  filterSymbol.value = '';
  loadGroups();
}

// 表格列定义
const columns = [
  {
    title: '策略组名称',
    key: 'groupName',
    minWidth: 180,
    render: (row: any) => {
      return h(NSpace, { align: 'center', size: 8, wrap: false }, () => [
        h('span', { style: { fontWeight: 500 } }, row.groupName),
        row.isDefault ? h(NTag, { type: 'success', size: 'small' }, () => '默认') : null,
        row.fromOfficialId ? h(NTag, { type: 'warning', size: 'small' }, () => '官方') : null,
      ]);
    },
  },
  {
    title: '交易所',
    key: 'exchange',
    width: 90,
    render: (row: any) => getExchangeLabel(row.exchange),
  },
  {
    title: '交易对',
    key: 'symbol',
    width: 110,
    render: (row: any) => h(NTag, { type: 'warning', size: 'small' }, () => row.symbol || 'BTC-USDT'),
  },
  {
    title: '交易类型',
    key: 'tradeType',
    width: 90,
    render: (row: any) => getTradeTypeLabel(row.tradeType),
  },
  {
    title: '订单类型',
    key: 'orderType',
    width: 80,
    render: (row: any) => h(NTag, { type: row.orderType === 'market' ? 'warning' : 'default', size: 'small' }, () => row.orderType === 'market' || !row.orderType ? '市价' : '限价'),
  },
  {
    title: '保证金',
    key: 'marginMode',
    width: 70,
    render: (row: any) => h(NTag, { type: 'info', size: 'small' }, () => row.marginMode === 'isolated' || !row.marginMode ? '逐仓' : '全仓'),
  },
  {
    title: '策略数量',
    key: 'strategyCount',
    width: 100,
    render: (row: any) => h(NTag, { type: (row.strategyCount || 0) >= 12 ? 'success' : 'warning', size: 'small' }, () => `${row.strategyCount || 0}/12`),
  },
  {
    title: '创建时间',
    key: 'createdAt',
    width: 160,
    render: (row: any) => row.createdAt ? row.createdAt.substring(0, 16) : '-',
  },
  {
    title: '状态',
    key: 'isActive',
    width: 80,
    render: (row: any) => h(NTag, { type: row.isActive !== 0 ? 'success' : 'default', size: 'small' }, () => row.isActive !== 0 ? '启用' : '禁用'),
  },
  {
    title: '操作',
    key: 'actions',
    width: 280,
    fixed: 'right' as const,
    render: (row: any) => {
      return h(NSpace, { size: 4, wrap: false }, () => [
        h(NButton, { 
          size: 'small', 
          quaternary: true,
          onClick: () => viewStrategies(row)
        }, { 
          default: () => '查看',
          icon: () => h(EyeOutlined)
        }),
        h(NButton, { 
          size: 'small', 
          quaternary: true,
          onClick: () => handleEdit(row)
        }, { 
          default: () => '编辑',
          icon: () => h(EditOutlined)
        }),
        h(NPopconfirm, {
          onPositiveClick: () => handleDelete(row)
        }, {
          trigger: () => h(NButton, { 
            size: 'small', 
            quaternary: true,
            type: 'error'
          }, { 
            default: () => '删除',
            icon: () => h(DeleteOutlined)
          }),
          default: () => '确定删除此模板吗？'
        }),
        !row.isDefault ? h(NButton, { 
          size: 'small', 
          onClick: () => setAsDefault(row)
        }, () => '设为默认') : null,
        h(NButton, { 
          size: 'small', 
          type: 'primary',
          onClick: () => createRobotWithStrategy(row)
        }, () => '创建机器人'),
      ]);
    },
  },
];

// 加载我的策略模板
async function loadGroups() {
  loading.value = true;
  try {
    const res = await http.request({
      url: '/strategy/group/list',
      method: 'get',
      params: {
        page: 1,
        pageSize: 100,
        isOfficial: 0, // 只查询非官方的（我的）
        exchange: filterExchange.value || undefined,
        symbol: filterSymbol.value || undefined,
      },
    });
    // 按创建时间倒序排列（最新的排第一）
    const list = res?.list || [];
    groupList.value = list.sort((a: any, b: any) => {
      const timeA = new Date(a.createdAt).getTime();
      const timeB = new Date(b.createdAt).getTime();
      return timeB - timeA;
    });
  } catch (error) {
    console.error('加载模板失败', error);
  } finally {
    loading.value = false;
  }
}

// 打开创建弹窗
function openCreateModal() {
  editingGroup.value = null;
  formData.value = {
    groupName: '',
    groupKey: '',
    exchange: 'bitget',
    symbol: 'BTC-USDT',
    orderType: 'market',
    marginMode: 'isolated',
    description: '',
  };
  showModal.value = true;
}

// 编辑
function handleEdit(row: any) {
  editingGroup.value = row;
  formData.value = {
    groupName: row.groupName,
    groupKey: row.groupKey,
    exchange: row.exchange,
    symbol: row.symbol,
    orderType: row.orderType,
    marginMode: row.marginMode,
    description: row.description || '',
  };
  showModal.value = true;
}

// 删除
async function handleDelete(row: any) {
  try {
    await http.request({ url: '/strategy/group/delete', method: 'post', data: { id: row.id } });
    message.success('删除成功');
    loadGroups();
  } catch (error: any) {
    message.error(error.message || '删除失败');
  }
}

// 设为默认
async function setAsDefault(row: any) {
  try {
    await http.request({ 
      url: '/strategy/group/setDefault', 
      method: 'post', 
      data: { id: row.id } 
    });
    message.success('已设为默认策略模板');
    loadGroups();
  } catch (error: any) {
    message.error(error.message || '设置失败');
  }
}

// 查看策略
function viewStrategies(row: any) {
  router.push({ path: '/toogo/strategy/list', query: { groupId: row.id, groupName: row.groupName } });
}

// 使用此策略模板创建机器人
function createRobotWithStrategy(row: any) {
  router.push({ path: '/toogo/robot/create', query: { strategyGroupId: row.id, strategyGroupName: row.groupName } });
}

// 提交表单
async function handleSubmit() {
  try {
    await formRef.value?.validate();
  } catch {
    return;
  }

  submitLoading.value = true;
  try {
    const url = editingGroup.value ? '/strategy/group/update' : '/strategy/group/create';
    const data = editingGroup.value ? { ...formData.value, id: editingGroup.value.id } : formData.value;
    await http.request({ url, method: 'post', data });
    message.success(editingGroup.value ? '更新成功' : '创建成功');
    showModal.value = false;
    loadGroups();
  } catch (error: any) {
    message.error(error.message || '操作失败');
  } finally {
    submitLoading.value = false;
  }
}

onMounted(() => {
  loadGroups();
});
</script>

<style scoped lang="less">
.my-strategy {
  .guide-card {
    background: linear-gradient(135deg, #fff9e6 0%, #fff3cc 100%);
    border: 1px solid #f0a020;
    
    .guide-icon {
      width: 60px;
      height: 60px;
      background: #fff;
      border-radius: 12px;
      display: flex;
      align-items: center;
      justify-content: center;
      box-shadow: 0 2px 8px rgba(240, 160, 32, 0.2);
    }
  }
}
</style>

