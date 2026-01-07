<template>
  <div class="strategy-group">
    <!-- 新手引导卡片 -->
    <n-card class="guide-card" :bordered="false" style="margin-bottom: 16px">
      <n-space align="center" justify="space-between">
        <n-space align="center" :size="16">
          <div class="guide-icon">
            <n-icon size="40" color="#f0a020">
              <BulbOutlined />
            </n-icon>
          </div>
          <div>
            <n-text strong style="font-size: 16px">新手入门</n-text>
            <n-text depth="3" style="display: block; margin-top: 4px">
              不知道如何配置策略？直接使用官方推荐策略，一键创建机器人开始交易！
            </n-text>
          </div>
        </n-space>
        <n-button type="primary" size="large" @click="quickCreateRobot">
          <template #icon
            ><n-icon><RocketOutlined /></n-icon
          ></template>
          一键创建机器人（使用官方策略）
        </n-button>
      </n-space>
    </n-card>

    <!-- 官方推荐策略区域 -->
    <n-card
      title="🔥 官方推荐策略"
      :bordered="false"
      style="margin-bottom: 16px"
      class="official-section"
    >
      <template #header-extra>
        <n-tag type="warning">专业团队调优，适合新手</n-tag>
      </template>

      <n-grid :cols="3" :x-gap="16" :y-gap="16" v-if="officialGroups.length > 0">
        <n-gi v-for="group in officialGroups" :key="group.id">
          <n-card hoverable class="official-card">
            <template #header>
              <n-space align="center">
                <n-tag type="success" size="small">官方</n-tag>
                <span class="group-name">{{ group.groupName }}</span>
              </n-space>
            </template>
            <template #header-extra>
              <n-tag type="info" size="small">{{ group.symbol }}</n-tag>
            </template>

            <n-space vertical :size="8">
              <n-text depth="3" style="font-size: 13px">
                {{
                  group.description ||
                  '专业团队精心调优的策略组合，包含12种市场状态和风险偏好组合。'
                }}
              </n-text>

              <n-space>
                <n-tag :bordered="false" size="small">{{ getExchangeLabel(group.exchange) }}</n-tag>
                <n-tag :bordered="false" size="small" type="success"
                  >{{ group.strategyCount || 12 }}种策略</n-tag
                >
              </n-space>
            </n-space>

            <template #footer>
              <n-space justify="space-between" style="width: 100%">
                <n-button size="small" quaternary @click="viewStrategies(group)">
                  <template #icon
                    ><n-icon><EyeOutlined /></n-icon
                  ></template>
                  查看详情
                </n-button>
                <n-button type="primary" size="small" @click="useOfficialStrategy(group)">
                  <template #icon
                    ><n-icon><ThunderboltOutlined /></n-icon
                  ></template>
                  立即使用
                </n-button>
              </n-space>
            </template>
          </n-card>
        </n-gi>
      </n-grid>

      <n-empty v-else description="暂无官方策略模板" />
    </n-card>

    <!-- 我的策略模板 -->
    <n-card title="📋 我的策略模板" :bordered="false" class="proCard">
      <template #header-extra>
        <n-button type="primary" @click="openCreateModal">
          <template #icon>
            <n-icon><PlusOutlined /></n-icon>
          </template>
          创建策略模板
        </n-button>
      </template>

      <!-- 筛选 -->
      <n-space style="margin-bottom: 16px">
        <n-select
          v-model:value="filterExchange"
          :options="exchangeOptions"
          placeholder="交易平台"
          clearable
          style="width: 140px"
        />
        <n-input
          v-model:value="filterSymbol"
          placeholder="搜索交易对"
          clearable
          style="width: 160px"
        />
        <n-button @click="loadGroups">
          <template #icon
            ><n-icon><SearchOutlined /></n-icon
          ></template>
          搜索
        </n-button>
      </n-space>

      <!-- 模板列表 -->
      <n-data-table
        :columns="columns"
        :data="userGroups"
        :loading="loading"
        :pagination="pagination"
        :row-key="(row) => row.id"
      />

      <n-empty v-if="userGroups.length === 0 && !loading" description="您还没有自定义策略模板">
        <template #extra>
          <n-space vertical align="center">
            <n-text depth="3">建议先使用官方推荐策略，熟悉后再创建自定义策略</n-text>
            <n-button type="primary" @click="openCreateModal">创建我的第一个策略模板</n-button>
          </n-space>
        </template>
      </n-empty>
    </n-card>

    <!-- 创建/编辑模板弹窗 -->
    <n-modal
      v-model:show="showModal"
      preset="card"
      :title="editingGroup ? '编辑策略模板' : '创建策略模板'"
      style="width: 600px"
    >
      <n-alert type="info" style="margin-bottom: 16px">
        创建后可以为模板添加12种策略（4种市场状态 × 3种风险偏好），机器人会根据市场自动匹配最优策略
      </n-alert>

      <n-form
        ref="formRef"
        :model="formData"
        :rules="rules"
        label-placement="left"
        label-width="100"
      >
        <n-form-item label="模板名称" path="groupName">
          <n-input
            v-model:value="formData.groupName"
            placeholder="如：BTC-USDT保守策略"
            maxlength="50"
          />
        </n-form-item>
        <n-form-item label="模板标识" path="groupKey">
          <n-input
            v-model:value="formData.groupKey"
            placeholder="唯一标识，如：my_btc_usdt"
            :disabled="!!editingGroup"
          />
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
          <n-input
            v-model:value="formData.description"
            type="textarea"
            :rows="3"
            placeholder="描述此策略模板..."
            maxlength="500"
          />
        </n-form-item>
        <n-form-item label="排序">
          <n-input-number v-model:value="formData.sort" :min="1" :max="9999" style="width: 100%" />
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

    <!-- 初始化策略弹窗 -->
    <n-modal v-model:show="showInitModal" preset="card" title="初始化策略" style="width: 500px">
      <n-alert type="info" style="margin-bottom: 16px">
        将为模板 <strong>{{ initGroup?.groupName }}</strong> 自动创建12种策略组合（4种市场状态 ×
        3种风险偏好）
      </n-alert>
      <n-space vertical>
        <n-checkbox v-model:checked="initOptions.useDefault">使用官方推荐参数</n-checkbox>
        <n-text depth="3">勾选后将使用经过调优的官方参数，否则使用默认参数</n-text>
      </n-space>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showInitModal = false">取消</n-button>
          <n-button type="primary" @click="handleInitStrategies" :loading="initLoading">
            初始化12种策略
          </n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- 使用官方策略弹窗 -->
    <n-modal v-model:show="showUseModal" preset="card" title="使用官方策略" style="width: 550px">
      <n-alert type="success" style="margin-bottom: 16px">
        <template #header>即将使用 {{ selectedOfficialGroup?.groupName }}</template>
        此策略包含12种优化参数组合，机器人会根据市场状态自动选择最优策略
      </n-alert>

      <n-descriptions :column="2" label-placement="left" bordered>
        <n-descriptions-item label="交易对">
          <n-tag type="info">{{ selectedOfficialGroup?.symbol }}</n-tag>
        </n-descriptions-item>
        <n-descriptions-item label="交易平台">
          {{ getExchangeLabel(selectedOfficialGroup?.exchange) }}
        </n-descriptions-item>
        <n-descriptions-item label="策略数量">
          <n-tag type="success">{{ selectedOfficialGroup?.strategyCount || 12 }}种</n-tag>
        </n-descriptions-item>
        <n-descriptions-item label="策略来源">
          <n-tag type="warning">官方推荐</n-tag>
        </n-descriptions-item>
      </n-descriptions>

      <n-divider />

      <n-space vertical :size="12">
        <n-text strong>选择使用方式：</n-text>
        <n-radio-group v-model:value="useMode">
          <n-space vertical>
            <n-radio value="create">
              <n-space align="center">
                <span>创建新机器人</span>
                <n-tag size="small" type="info">推荐新手</n-tag>
              </n-space>
            </n-radio>
            <n-radio value="apply" :disabled="robotOptions.length === 0">
              <n-space align="center">
                <span>应用到已有机器人</span>
                <n-text v-if="robotOptions.length === 0" depth="3" style="font-size: 12px"
                  >（暂无可用机器人）</n-text
                >
              </n-space>
            </n-radio>
          </n-space>
        </n-radio-group>

        <n-form-item v-if="useMode === 'apply'" label="选择机器人" style="margin-top: 8px">
          <n-select
            v-model:value="selectedRobotId"
            :options="robotOptions"
            placeholder="请选择机器人"
          />
        </n-form-item>
      </n-space>

      <template #footer>
        <n-space justify="end">
          <n-button @click="showUseModal = false">取消</n-button>
          <n-button type="primary" @click="confirmUseOfficial" :loading="useLoading">
            {{ useMode === 'create' ? '创建机器人' : '应用策略' }}
          </n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<script lang="ts" setup>
  import { ref, h, computed, onMounted } from 'vue';
  import { useRouter } from 'vue-router';
  import { useMessage, useDialog, NButton, NTag, NSpace, NIcon, NPopconfirm } from 'naive-ui';
  import {
    PlusOutlined,
    SearchOutlined,
    EditOutlined,
    DeleteOutlined,
    SettingOutlined,
    UnorderedListOutlined,
    BulbOutlined,
    RocketOutlined,
    ThunderboltOutlined,
    EyeOutlined,
  } from '@vicons/antd';
  import { http } from '@/utils/http/axios';

  const router = useRouter();
  const message = useMessage();
  const dialog = useDialog();

  // 状态
  const loading = ref(false);
  const groupList = ref<any[]>([]);
  const showModal = ref(false);
  const showInitModal = ref(false);
  const showUseModal = ref(false);
  const editingGroup = ref<any>(null);
  const initGroup = ref<any>(null);
  const selectedOfficialGroup = ref<any>(null);
  const formRef = ref<any>(null);
  const submitLoading = ref(false);
  const initLoading = ref(false);
  const useLoading = ref(false);

  // 使用模式
  const useMode = ref<'create' | 'apply'>('create');
  const selectedRobotId = ref<number | null>(null);
  const robotOptions = ref<any[]>([]);

  // 筛选
  const filterExchange = ref<string | null>(null);
  const filterSymbol = ref('');

  // 分页
  const pagination = ref({
    page: 1,
    pageSize: 10,
    showSizePicker: true,
    pageSizes: [10, 20, 50],
    onChange: (page: number) => {
      pagination.value.page = page;
      loadGroups();
    },
    onUpdatePageSize: (pageSize: number) => {
      pagination.value.pageSize = pageSize;
      pagination.value.page = 1;
      loadGroups();
    },
  });

  // 计算属性：分离官方和用户策略
  // 注意：非官方 != 我的模板。我的模板必须有 userId（避免 public(0) 的非官方被误归类）
  const officialGroups = computed(() =>
    groupList.value.filter((g) => Number(g?.isOfficial) === 1 || g?.isOfficial === true),
  );
  const userGroups = computed(() =>
    groupList.value.filter((g) => {
      const isOfficial = Number(g?.isOfficial) === 1 || g?.isOfficial === true;
      if (isOfficial) return false;
      const uid = Number((g as any)?.userId ?? (g as any)?.user_id ?? 0);
      return uid > 0;
    }),
  );

  // 表单数据
  const formData = ref({
    groupName: '',
    groupKey: '',
    exchange: 'binance',
    symbol: 'BTC-USDT',
    orderType: 'market',
    marginMode: 'isolated',
    description: '',
    sort: 100,
  });

  // 初始化选项
  const initOptions = ref({
    useDefault: true,
  });

  // 选项
  const exchangeOptions = [
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

  // 表格列（用户自定义模板）
  const columns = [
    { title: 'ID', key: 'id', width: 60 },
    {
      title: '模板名称',
      key: 'groupName',
      render: (row: any) => h('span', { style: { fontWeight: 600 } }, row.groupName),
    },
    {
      title: '交易对',
      key: 'symbol',
      render: (row: any) => h(NTag, { type: 'info', size: 'small' }, () => row.symbol),
    },
    {
      title: '平台',
      key: 'exchange',
      render: (row: any) => getExchangeLabel(row.exchange),
    },
    {
      title: '策略数',
      key: 'strategyCount',
      render: (row: any) =>
        h(
          NTag,
          { type: row.strategyCount >= 12 ? 'success' : 'warning', size: 'small' },
          () => `${row.strategyCount || 0}种`,
        ),
    },
    {
      title: '状态',
      key: 'isActive',
      render: (row: any) =>
        h(NTag, { type: row.isActive ? 'success' : 'default', size: 'small' }, () =>
          row.isActive ? '启用' : '禁用',
        ),
    },
    { title: '排序', key: 'sort', width: 80 },
    {
      title: '操作',
      key: 'actions',
      width: 280,
      render: (row: any) =>
        h(NSpace, {}, () => [
          h(
            NButton,
            { size: 'small', type: 'primary', onClick: () => viewStrategies(row) },
            {
              default: () => '查看策略',
              icon: () => h(NIcon, null, () => h(UnorderedListOutlined)),
            },
          ),
          row.strategyCount < 12 &&
            h(
              NButton,
              { size: 'small', type: 'warning', onClick: () => openInitModal(row) },
              { default: () => '初始化', icon: () => h(NIcon, null, () => h(SettingOutlined)) },
            ),
          h(
            NButton,
            { size: 'small', quaternary: true, onClick: () => handleEdit(row) },
            { icon: () => h(NIcon, null, () => h(EditOutlined)) },
          ),
          h(
            NPopconfirm,
            { onPositiveClick: () => handleDelete(row) },
            {
              trigger: () =>
                h(
                  NButton,
                  { size: 'small', quaternary: true, type: 'error' },
                  { icon: () => h(NIcon, null, () => h(DeleteOutlined)) },
                ),
              default: () => '确定删除此模板及其所有策略吗？',
            },
          ),
        ]),
    },
  ];

  // 辅助函数
  function getExchangeLabel(exchange: string) {
    const labels: Record<string, string> = { binance: 'Binance', okx: 'OKX', gateio: 'Gate.io' };
    return labels[exchange] || exchange;
  }

  // 加载模板列表
  async function loadGroups() {
    loading.value = true;
    try {
      const res = await http.request({
        url: '/trading/strategy/group/list',
        method: 'get',
        params: {
          page: pagination.value.page,
          pageSize: pagination.value.pageSize,
          exchange: filterExchange.value || undefined,
          symbol: filterSymbol.value || undefined,
          isActive: 1,
        },
      });
      groupList.value = res?.list || [];
    } catch (error) {
      console.error('加载模板失败', error);
    } finally {
      loading.value = false;
    }
  }

  // 加载机器人列表
  async function loadRobots() {
    try {
      const res = await http.request({ url: '/trading/robot/list', method: 'get' });
      robotOptions.value = (res?.list || [])
        .filter((r: any) => r.status !== 2) // 排除运行中的机器人
        .map((r: any) => ({
          label: `${r.robotName} (${r.symbol})`,
          value: r.id,
        }));
    } catch (error) {
      console.error('加载机器人失败', error);
    }
  }

  // 快速创建机器人（使用官方策略）
  function quickCreateRobot() {
    // 找到官方BTC-USDT策略
    const btcGroup =
      officialGroups.value.find((g) => g.symbol === 'BTC-USDT') || officialGroups.value[0];
    if (btcGroup) {
      router.push({
        path: '/toogo/robot/create',
        query: {
          useOfficial: '1',
          strategyGroupId: btcGroup.id,
          symbol: btcGroup.symbol,
        },
      });
    } else {
      router.push('/toogo/robot/create');
    }
  }

  // 使用官方策略
  function useOfficialStrategy(group: any) {
    selectedOfficialGroup.value = group;
    useMode.value = 'create';
    selectedRobotId.value = null;
    loadRobots();
    showUseModal.value = true;
  }

  // 确认使用官方策略
  async function confirmUseOfficial() {
    if (useMode.value === 'create') {
      // 跳转到创建机器人页面
      router.push({
        path: '/toogo/robot/create',
        query: {
          useOfficial: '1',
          strategyGroupId: selectedOfficialGroup.value.id,
          symbol: selectedOfficialGroup.value.symbol,
        },
      });
      showUseModal.value = false;
    } else {
      // 应用到已有机器人
      if (!selectedRobotId.value) {
        message.error('请选择机器人');
        return;
      }

      useLoading.value = true;
      try {
        // 获取官方策略组的平衡型策略（推荐）
        const res = await http.request({
          url: '/trading/strategy/template/list',
          method: 'get',
          params: { groupId: selectedOfficialGroup.value.id },
        });
        const strategies = res?.list || [];
        // 优先选择平衡型-趋势市场策略
        const balancedStrategy =
          strategies.find(
            (s: any) => s.riskPreference === 'balanced' && s.marketState === 'trend',
          ) || strategies[0];

        if (balancedStrategy) {
          await http.request({
            url: '/trading/strategy/template/apply',
            method: 'post',
            data: { strategyId: balancedStrategy.id, robotId: selectedRobotId.value },
          });
          message.success('策略应用成功！机器人将使用官方推荐策略');
          showUseModal.value = false;
        } else {
          message.error('未找到可用策略');
        }
      } catch (error: any) {
        message.error(error.message || '应用失败');
      } finally {
        useLoading.value = false;
      }
    }
  }

  // 打开创建弹窗
  function openCreateModal() {
    editingGroup.value = null;
    formData.value = {
      groupName: '',
      groupKey: '',
      exchange: 'binance',
      symbol: 'BTC-USDT',
      orderType: 'market',
      marginMode: 'isolated',
      description: '',
      sort: 100,
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
      sort: row.sort,
    };
    showModal.value = true;
  }

  // 删除
  async function handleDelete(row: any) {
    try {
      await http.request({
        url: '/trading/strategy/group/delete',
        method: 'post',
        data: { id: row.id },
      });
      message.success('删除成功');
      loadGroups();
    } catch (error: any) {
      message.error(error.message || '删除失败');
    }
  }

  // 查看策略
  function viewStrategies(row: any) {
    router.push({
      path: '/toogo/strategy/list',
      query: { groupId: row.id, groupName: row.groupName },
    });
  }

  // 打开初始化弹窗
  function openInitModal(row: any) {
    initGroup.value = row;
    initOptions.value.useDefault = true;
    showInitModal.value = true;
  }

  // 初始化策略
  async function handleInitStrategies() {
    initLoading.value = true;
    try {
      await http.request({
        url: '/trading/strategy/group/initStrategies',
        method: 'post',
        data: {
          groupId: initGroup.value.id,
          useDefault: initOptions.value.useDefault,
        },
      });
      message.success('成功初始化12种策略');
      showInitModal.value = false;
      loadGroups();
    } catch (error: any) {
      message.error(error.message || '初始化失败');
    } finally {
      initLoading.value = false;
    }
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
      const url = editingGroup.value
        ? '/trading/strategy/group/update'
        : '/trading/strategy/group/create';
      const data = editingGroup.value
        ? { ...formData.value, id: editingGroup.value.id }
        : formData.value;
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
  .strategy-group {
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

    .official-section {
      background: linear-gradient(135deg, #f6ffed 0%, #e6f7e6 100%);
      border: 1px solid #52c41a;
    }

    .official-card {
      border: 1px solid #b7eb8f;
      transition: all 0.3s;

      &:hover {
        border-color: #52c41a;
        box-shadow: 0 4px 12px rgba(82, 196, 26, 0.15);
      }

      .group-name {
        font-weight: 600;
        font-size: 15px;
      }

      :deep(.n-card__footer) {
        padding: 12px 16px;
        border-top: 1px solid #f0f0f0;
      }
    }
  }
</style>
