<template>
  <div class="official-strategy-admin">
    <n-card title="官方策略组管理" :bordered="false">
      <template #header-extra>
        <n-space>
          <n-button type="primary" @click="openCreateModal">
            <template #icon
              ><n-icon><PlusOutlined /></n-icon
            ></template>
            创建官方策略组
          </n-button>
        </n-space>
      </template>

      <n-alert type="info" style="margin-bottom: 16px" :bordered="false">
        <template #header>
          <n-space align="center">
            <n-icon size="18">🔥</n-icon>
            <span>官方策略组管理说明</span>
          </n-space>
        </template>
        此页面用于创建和维护官方策略组。每个策略组包含12种策略模板（4种市场状态×3种风险偏好），会在"策略管理-官方策略模板"中展示，供所有用户添加使用。
      </n-alert>

      <!-- 筛选 -->
      <n-form inline label-placement="left" :show-feedback="false" style="margin-bottom: 16px">
        <n-form-item label="交易所">
          <n-select
            v-model:value="filterExchange"
            :options="exchangeOptions"
            placeholder="全部"
            clearable
            style="width: 120px"
          />
        </n-form-item>
        <n-form-item label="交易对">
          <n-input
            v-model:value="filterSymbol"
            placeholder="搜索交易对"
            clearable
            style="width: 140px"
          />
        </n-form-item>
        <n-form-item>
          <n-space>
            <n-button type="primary" @click="loadData">
              <template #icon
                ><n-icon><SearchOutlined /></n-icon
              ></template>
              查询
            </n-button>
            <n-button @click="resetFilter">重置</n-button>
          </n-space>
        </n-form-item>
      </n-form>

      <!-- 策略组表格 -->
      <n-spin :show="loading">
        <n-data-table
          :columns="columns"
          :data="groupList"
          :row-key="(row: any) => row.id"
          :pagination="pagination"
          @update:page="handlePageChange"
          striped
        />
      </n-spin>
    </n-card>

    <!-- 创建/编辑策略组弹窗 -->
    <n-modal
      v-model:show="showModal"
      preset="card"
      :title="editingGroup ? '编辑官方策略组' : '创建官方策略组'"
      style="width: 600px"
    >
      <n-alert type="info" style="margin-bottom: 16px">
        创建后可以批量初始化12种策略（4种市场状态 × 3种风险偏好），机器人会根据市场自动匹配最优策略
      </n-alert>

      <n-form
        ref="formRef"
        :model="formData"
        :rules="rules"
        label-placement="left"
        label-width="100"
      >
        <n-form-item label="策略组名称" path="groupName">
          <n-input
            v-model:value="formData.groupName"
            placeholder="如：🔥 BTC-USDT 官方策略 V6.0"
            maxlength="50"
          />
        </n-form-item>
        <n-form-item label="策略组标识" path="groupKey">
          <n-input
            v-model:value="formData.groupKey"
            placeholder="唯一标识，如：official_btc_usdt_v6"
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
        <n-grid :cols="2" :x-gap="16">
          <n-gi>
            <n-form-item label="官方" path="isOfficial">
              <!-- 创建官方策略组：官方标识固定为1；编辑时允许切换 -->
              <n-switch
                v-model:value="formData.isOfficial"
                :checked-value="1"
                :unchecked-value="0"
                :disabled="!editingGroup"
              />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="状态" path="isActive">
              <n-switch v-model:value="formData.isActive" :checked-value="1" :unchecked-value="0" />
            </n-form-item>
          </n-gi>
        </n-grid>
        <n-form-item label="描述">
          <n-input
            v-model:value="formData.description"
            type="textarea"
            :rows="3"
            placeholder="描述此官方策略组..."
            maxlength="500"
          />
        </n-form-item>
        <n-form-item label="排序">
          <n-input-number v-model:value="formData.sort" :min="0" style="width: 100%" />
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
  import { ref, computed, onMounted, h } from 'vue';
  import { useRouter } from 'vue-router';
  import { useMessage, NButton, NSpace, NTag, NPopconfirm, NIcon } from 'naive-ui';
  import {
    PlusOutlined,
    SearchOutlined,
    EditOutlined,
    DeleteOutlined,
    CopyOutlined,
    EyeOutlined,
    ThunderboltOutlined,
    UnorderedListOutlined,
    SettingOutlined,
  } from '@vicons/antd';
  import { http } from '@/utils/http/axios';

  const router = useRouter();
  const message = useMessage();

  // 状态
  const loading = ref(false);
  const groupList = ref<any[]>([]);
  const showModal = ref(false);
  const showInitModal = ref(false);
  const editingGroup = ref<any>(null);
  const initGroup = ref<any>(null);
  const formRef = ref<any>(null);
  const submitLoading = ref(false);
  const initLoading = ref(false);
  const cloningMap = ref<Record<string, boolean>>({});

  // 筛选
  const filterExchange = ref<string | null>(null);
  const filterSymbol = ref('');

  // 分页
  const pagination = ref({
    page: 1,
    pageSize: 10,
    itemCount: 0,
    showSizePicker: true,
    pageSizes: [10, 20, 50],
  });

  // 初始化选项
  const initOptions = ref({
    useDefault: true,
  });

  // 表单数据
  const formData = ref({
    groupName: '',
    groupKey: '',
    exchange: 'binance',
    symbol: 'BTCUSDT',
    orderType: 'market',
    marginMode: 'isolated',
    description: '',
    isOfficial: 1,
    isActive: 1,
    sort: 1,
  });

  // 表单验证规则
  const rules = {
    groupName: { required: true, message: '请输入策略组名称', trigger: 'blur' },
    groupKey: { required: true, message: '请输入策略组标识', trigger: 'blur' },
    exchange: { required: true, message: '请选择交易平台', trigger: 'change' },
    symbol: { required: true, message: '请输入交易对', trigger: 'blur' },
  };

  // 选项
  const exchangeOptions = [
    { label: 'Binance', value: 'binance' },
    { label: 'OKX', value: 'okx' },
    { label: 'Gate', value: 'gate' },
  ];

  const symbolOptions = [
    { label: 'BTCUSDT', value: 'BTCUSDT' },
    { label: 'ETHUSDT', value: 'ETHUSDT' },
    { label: 'BNBUSDT', value: 'BNBUSDT' },
    { label: 'SOLUSDT', value: 'SOLUSDT' },
  ];

  const orderTypeOptions = [
    { label: '市价单', value: 'market' },
    { label: '限价单', value: 'limit' },
  ];

  const marginModeOptions = [
    { label: '逐仓', value: 'isolated' },
    { label: '全仓', value: 'crossed' },
  ];

  // 表格列定义
  const columns = [
    { title: 'ID', key: 'id', width: 60 },
    {
      title: '策略组名称',
      key: 'groupName',
      width: 250,
      ellipsis: {
        tooltip: true,
      },
    },
    {
      title: '交易所',
      key: 'exchange',
      width: 100,
      render: (row: any) => {
        const map: any = { binance: 'Binance', okx: 'OKX', gate: 'Gate' };
        return map[row.exchange] || row.exchange;
      },
    },
    { title: '交易对', key: 'symbol', width: 120 },
    {
      title: '策略数量',
      key: 'strategyCount',
      width: 160,
      render: (row: any) => {
        const count = row.strategyCount || 0;
        const type = count === 12 ? 'success' : count > 0 ? 'warning' : 'default';
        const cloning = !!cloningMap.value[String(row.id)];
        return h(NSpace, { align: 'center', size: 6 }, () => [
          h(NTag, { type, size: 'small' }, { default: () => `${count}/12` }),
          h(
            NPopconfirm,
            { onPositiveClick: () => handleClone(row) },
            {
              trigger: () =>
                h(
                  NButton,
                  { size: 'tiny', quaternary: true, type: 'primary', loading: cloning, disabled: cloning },
                  {
                    default: () => '复制',
                    icon: () => h(NIcon, null, () => h(CopyOutlined)),
                  },
                ),
              default: () => '确定复制此策略组（含12套策略模板）吗？',
            },
          ),
        ]);
      },
    },
    {
      title: '官方',
      key: 'isOfficial',
      width: 80,
      render: (row: any) => {
        const isOfficial = Number(row.isOfficial) === 1;
        return h(
          NTag,
          { type: isOfficial ? 'warning' : 'default', size: 'small' },
          { default: () => (isOfficial ? '官方' : '非官方') },
        );
      },
    },
    {
      title: '状态',
      key: 'isActive',
      width: 80,
      render: (row: any) => {
        const isActive = Number(row.isActive) === 1;
        return h(
          NTag,
          { type: isActive ? 'success' : 'default', size: 'small' },
          { default: () => (isActive ? '启用' : '禁用') },
        );
      },
    },
    { title: '排序', key: 'sort', width: 70 },
    {
      title: '操作',
      key: 'actions',
      width: 280,
      fixed: 'right' as const,
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
              {
                default: () => '初始化',
                icon: () => h(NIcon, null, () => h(SettingOutlined)),
              },
            ),
          h(
            NButton,
            { size: 'small', quaternary: true, onClick: () => handleEdit(row) },
            {
              icon: () => h(NIcon, null, () => h(EditOutlined)),
            },
          ),
          h(
            NPopconfirm,
            { onPositiveClick: () => handleDelete(row) },
            {
              trigger: () =>
                h(
                  NButton,
                  { size: 'small', quaternary: true, type: 'error' },
                  {
                    icon: () => h(NIcon, null, () => h(DeleteOutlined)),
                  },
                ),
              default: () => '确定删除此策略组及其所有策略吗？',
            },
          ),
        ]),
    },
  ];

  // 获取交易所标签
  function getExchangeLabel(exchange: string) {
    const map: any = { binance: 'Binance', okx: 'OKX', gate: 'Gate' };
    return map[exchange] || exchange;
  }

  // 加载数据
  async function loadData() {
    loading.value = true;
    try {
      const res = await http.request({
        url: '/strategy/group/list',
        method: 'get',
        params: {
          page: pagination.value.page,
          pageSize: pagination.value.pageSize,
          exchange: filterExchange.value,
          symbol: filterSymbol.value,
          // 显示“公共策略组”（user_id=0/NULL）：包含官方/非官方，避免切成非官方后列表消失
          nonPersonal: 1,
        },
      });
      groupList.value = res?.list || [];
      pagination.value.itemCount = res?.totalCount || res?.total || 0;
    } catch (error: any) {
      message.error(error.message || '加载失败');
    } finally {
      loading.value = false;
    }
  }

  // 重置筛选
  function resetFilter() {
    filterExchange.value = null;
    filterSymbol.value = '';
    pagination.value.page = 1;
    loadData();
  }

  // 分页变化
  function handlePageChange(page: number) {
    pagination.value.page = page;
    loadData();
  }

  // 打开创建弹窗
  function openCreateModal() {
    editingGroup.value = null;
    formData.value = {
      groupName: '',
      groupKey: '',
      exchange: 'binance',
      symbol: 'BTCUSDT',
      orderType: 'market',
      marginMode: 'isolated',
      description: '',
      isOfficial: 1,
      isActive: 1,
      sort: 1,
    };
    showModal.value = true;
  }

  // 编辑
  function handleEdit(row: any) {
    editingGroup.value = row;
    formData.value = {
      ...row,
      isOfficial: Number(row.isOfficial) === 1 ? 1 : 0,
      isActive: Number(row.isActive) === 1 ? 1 : 0,
    };
    showModal.value = true;
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

      const data: any = { ...formData.value };
      // 创建“官方策略组管理”页面的记录时，默认创建为官方策略组（与按钮文案一致）
      if (!editingGroup.value) {
        data.isOfficial = 1;
      }
      // 官方策略：归属系统用户（user_id=0）
      if (Number(data.isOfficial) === 1) {
        data.userId = 0;
      }

      if (editingGroup.value) {
        data.id = editingGroup.value.id;
      }

      await http.request({ url, method: 'post', data });
      message.success(editingGroup.value ? '更新成功' : '创建成功');
      showModal.value = false;
      loadData();
    } catch (error: any) {
      message.error(error.message || '保存失败');
    } finally {
      submitLoading.value = false;
    }
  }

  // 删除
  async function handleDelete(row: any) {
    try {
      await http.request({
        url: '/strategy/group/delete',
        method: 'post',
        data: { id: row.id },
      });
      message.success('删除成功');
      loadData();
    } catch (error: any) {
      message.error(error.message || '删除失败');
    }
  }

  // 复制策略组（含策略模板）
  async function handleClone(row: any) {
    const k = String(row?.id ?? '');
    if (k && cloningMap.value[k]) return;
    try {
      if (k) cloningMap.value[k] = true;
      await http.request({
        url: '/strategy/group/clone',
        method: 'post',
        data: { id: row.id },
      });
      message.success('复制成功');
      loadData();
    } catch (error: any) {
      message.error(error.message || '复制失败');
    } finally {
      if (k) cloningMap.value[k] = false;
    }
  }

  // 查看策略
  function viewStrategies(row: any) {
    // 管理员页面跳转时，允许修改和删除官方策略（不传递readonly参数）
    // 使用 Vue Router 进行跳转（history 模式）
    router.push({
      path: '/toogo/strategy/list',
      query: {
        groupId: row.id,
        groupName: row.groupName,
        back: '/toogo-admin/strategy',
      },
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
    if (!initGroup.value) return;

    initLoading.value = true;
    try {
      await http.request({
        url: '/strategy/group/init',
        method: 'post',
        data: {
          groupId: initGroup.value.id,
          useDefault: initOptions.value.useDefault,
        },
      });
      message.success('策略初始化成功，已生成12种策略模板');
      showInitModal.value = false;
      loadData();
    } catch (error: any) {
      message.error(error.message || '初始化失败');
    } finally {
      initLoading.value = false;
    }
  }

  // 初始化
  onMounted(() => {
    loadData();
  });
</script>

<style scoped lang="less">
  .official-strategy-admin {
    :deep(.n-data-table) {
      .n-data-table-td {
        padding: 12px 8px;
      }
    }
  }
</style>
