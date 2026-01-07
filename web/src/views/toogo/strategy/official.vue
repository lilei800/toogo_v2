<template>
  <div class="official-strategy">
    <!-- 顶部说明 -->
    <n-card class="intro-card" :bordered="false" style="margin-bottom: 16px">
      <n-space align="center" justify="space-between">
        <n-space align="center" :size="16">
          <div class="intro-icon">
            <n-icon size="40" color="#52c41a">
              <SafetyCertificateOutlined />
            </n-icon>
          </div>
          <div>
            <n-text strong style="font-size: 16px">官方策略模板</n-text>
            <n-text depth="3" style="display: block; margin-top: 4px">
              由专业团队精心调优，经过大量历史数据回测验证，适合大多数市场环境
            </n-text>
          </div>
        </n-space>
        <n-tag type="success" size="large">
          <template #icon
            ><n-icon><CheckCircleOutlined /></n-icon
          ></template>
          免费使用
        </n-tag>
      </n-space>
    </n-card>

    <!-- 筛选区域 -->
    <n-card :bordered="false" style="margin-bottom: 16px">
      <n-form inline label-placement="left" :show-feedback="false">
        <n-form-item label="交易对">
          <n-select
            v-model:value="filterSymbol"
            :options="symbolOptions"
            placeholder="全部"
            clearable
            style="width: 140px"
          />
        </n-form-item>
        <n-form-item>
          <n-space>
            <n-button type="primary" @click="loadOfficialGroups">
              <template #icon
                ><n-icon><SearchOutlined /></n-icon
              ></template>
              查询
            </n-button>
            <n-button @click="resetFilter">重置</n-button>
          </n-space>
        </n-form-item>
      </n-form>
    </n-card>

    <n-card title="🔥 官方策略模板" :bordered="false" class="proCard">
      <template #header-extra>
        <n-text depth="3">共 {{ officialGroups.length }} 套模板，点击"添加"即可使用</n-text>
      </template>

      <n-spin :show="loading">
        <n-data-table
          :columns="columns"
          :data="filteredGroups"
          :row-key="(row: any) => row.id"
          :pagination="false"
          :bordered="false"
          striped
        />

        <n-empty v-if="!loading && filteredGroups.length === 0" description="暂无匹配的策略模板" />
      </n-spin>
    </n-card>

    <!-- 添加成功提示弹窗 -->
    <n-modal v-model:show="showSuccessModal" preset="card" title="添加成功" style="width: 450px">
      <n-result
        status="success"
        title="策略模板已添加"
        :description="`${addedGroupName} 已添加到您的策略模板中`"
      >
        <template #footer>
          <n-space justify="center">
            <n-button @click="showSuccessModal = false">继续浏览</n-button>
            <n-button type="primary" @click="goToMy">
              <template #icon
                ><n-icon><ArrowRightOutlined /></n-icon
              ></template>
              去我的策略查看
            </n-button>
          </n-space>
        </template>
      </n-result>
    </n-modal>
  </div>
</template>

<script lang="ts" setup>
  import { ref, computed, onMounted, h } from 'vue';
  import { useRouter } from 'vue-router';
  import { useMessage, NButton, NSpace, NTag } from 'naive-ui';
  import {
    SearchOutlined,
    SafetyCertificateOutlined,
    CheckCircleOutlined,
    ArrowRightOutlined,
    EyeOutlined,
  } from '@vicons/antd';
  import { http } from '@/utils/http/axios';

  const router = useRouter();
  const message = useMessage();

  // 筛选条件
  const filterSymbol = ref<string | null>(null);

  const symbolOptions = [
    { label: 'BTC-USDT', value: 'BTCUSDT' },
    { label: 'ETH-USDT', value: 'ETHUSDT' },
  ];

  // 状态
  const loading = ref(false);
  const officialGroups = ref<any[]>([]);
  const addingId = ref<number | null>(null);
  const addedIds = ref<number[]>([]);
  const showSuccessModal = ref(false);
  const addedGroupName = ref('');

  // 筛选后的数据
  const filteredGroups = computed(() => {
    let result = [...officialGroups.value];
    if (filterSymbol.value) {
      result = result.filter((g) => g.symbol === filterSymbol.value);
    }
    return result;
  });

  // 表格列定义
  const columns = [
    {
      title: '策略组名称',
      key: 'groupName',
      minWidth: 220,
      ellipsis: false,
      render: (row: any) => {
        return h(NSpace, { align: 'center', size: 8, wrap: false }, () => [
          h('span', { style: { fontWeight: 500, whiteSpace: 'nowrap' } }, row.groupName),
          isLatestVersion(row) ? h(NTag, { type: 'success', size: 'small' }, () => '推荐') : null,
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
      render: (row: any) =>
        h(NTag, { type: 'warning', size: 'small' }, () => row.symbol || 'BTC-USDT'),
    },
    {
      title: '交易类型',
      key: 'tradeType',
      width: 90,
      render: (row: any) => getTradeTypeLabel(row.tradeType || row.orderType),
    },
    {
      title: '订单类型',
      key: 'orderType',
      width: 80,
      render: (row: any) =>
        h(NTag, { type: row.orderType === 'market' ? 'warning' : 'default', size: 'small' }, () =>
          row.orderType === 'market' || !row.orderType ? '市价' : '限价',
        ),
    },
    {
      title: '保证金',
      key: 'marginMode',
      width: 70,
      render: (row: any) =>
        h(NTag, { type: 'info', size: 'small' }, () =>
          row.marginMode === 'isolated' || !row.marginMode ? '逐仓' : '全仓',
        ),
    },
    {
      title: '策略数量',
      key: 'strategyCount',
      width: 100,
      render: (row: any) =>
        h(NTag, { type: 'info', size: 'small' }, () => `${row.strategyCount || 12}种`),
    },
    {
      title: '创建时间',
      key: 'createdAt',
      width: 160,
      render: (row: any) => (row.createdAt ? row.createdAt.substring(0, 16) : '-'),
    },
    {
      title: '状态',
      key: 'status',
      width: 80,
      render: (row: any) => {
        if (isAdded(row.id)) {
          return h(NTag, { type: 'success', size: 'small' }, () => '已添加');
        }
        return h(NTag, { size: 'small' }, () => '可添加');
      },
    },
    {
      title: '操作',
      key: 'actions',
      width: 200,
      fixed: 'right' as const,
      render: (row: any) => {
        return h(NSpace, { size: 8 }, () => [
          h(
            NButton,
            {
              size: 'small',
              quaternary: true,
              onClick: () => viewStrategies(row),
            },
            {
              default: () => '查看详情',
              icon: () => h(EyeOutlined),
            },
          ),
          h(
            NButton,
            {
              size: 'small',
              type: 'primary',
              disabled: isAdded(row.id),
              loading: addingId.value === row.id,
              onClick: () => addToMy(row),
            },
            () => (isAdded(row.id) ? '已添加' : '添加'),
          ),
        ]);
      },
    },
  ];

  // 辅助函数
  function getExchangeLabel(exchange: string) {
    const labels: Record<string, string> = { binance: 'Binance', okx: 'OKX', gateio: 'Gate.io' };
    return labels[exchange] || exchange || 'Bitget';
  }

  function getTradeTypeLabel(tradeType: string) {
    if (!tradeType || tradeType === 'perpetual' || tradeType === 'market') return '永续合约';
    if (tradeType === 'delivery') return '交割合约';
    if (tradeType === 'spot') return '现货';
    return tradeType;
  }

  function isLatestVersion(group: any): boolean {
    const key = group.groupKey || '';
    if (key.includes('12')) return true;
    const version = key.includes('v3') ? 3 : key.includes('v2') ? 2 : 1;
    const maxVersion = Math.max(
      ...officialGroups.value.map((g) => {
        const k = g.groupKey || '';
        return k.includes('v3') ? 3 : k.includes('v2') ? 2 : 1;
      }),
    );
    return version === maxVersion;
  }

  function isAdded(groupId: number) {
    return addedIds.value.includes(groupId);
  }

  function resetFilter() {
    filterSymbol.value = null;
    loadOfficialGroups();
  }

  // 加载官方策略模板
  async function loadOfficialGroups() {
    loading.value = true;
    try {
      const res = await http.request({
        url: '/strategy/group/list',
        method: 'get',
        params: {
          page: 1,
          pageSize: 100,
          isOfficial: 1,
          isActive: 1, // 只展示启用的
        },
      });
      // 按创建时间倒序排列（最新的排第一）
      const list = res?.list || [];
      officialGroups.value = list.sort((a: any, b: any) => {
        const timeA = new Date(a.createdAt).getTime();
        const timeB = new Date(b.createdAt).getTime();
        return timeB - timeA;
      });
    } catch (error) {
      console.error('加载官方策略失败', error);
    } finally {
      loading.value = false;
    }

    // 加载用户已添加的策略
    loadUserAddedGroups();
  }

  // 加载用户已添加的策略ID
  async function loadUserAddedGroups() {
    try {
      const res = await http.request({
        url: '/strategy/group/list',
        method: 'get',
        params: { page: 1, pageSize: 100, isOfficial: 0, isActive: 1 },
      });
      const myGroups = res?.list || [];
      addedIds.value = myGroups
        .filter((g: any) => g.fromOfficialId)
        .map((g: any) => g.fromOfficialId);
    } catch (error) {
      console.error('加载用户策略失败', error);
    }
  }

  // 添加到我的策略
  async function addToMy(group: any) {
    addingId.value = group.id;
    try {
      await http.request({
        url: '/strategy/group/copyFromOfficial',
        method: 'post',
        data: { officialGroupId: group.id },
      });
      addedIds.value.push(group.id);
      addedGroupName.value = group.groupName;
      showSuccessModal.value = true;
    } catch (error: any) {
      message.error(error.message || '添加失败');
    } finally {
      addingId.value = null;
    }
  }

  // 查看策略详情
  function viewStrategies(row: any) {
    router.push({
      path: '/toogo/strategy/list',
      query: { groupId: row.id, groupName: row.groupName, readonly: '1' },
    });
  }

  // 跳转到我的策略
  function goToMy() {
    showSuccessModal.value = false;
    router.push('/toogo/strategy/my');
  }

  onMounted(() => {
    loadOfficialGroups();
  });
</script>

<style scoped lang="less">
  .official-strategy {
    .intro-card {
      background: linear-gradient(135deg, #f6ffed 0%, #d9f7be 100%);
      border: 1px solid #52c41a;

      .intro-icon {
        width: 60px;
        height: 60px;
        background: #fff;
        border-radius: 12px;
        display: flex;
        align-items: center;
        justify-content: center;
        box-shadow: 0 2px 8px rgba(82, 196, 26, 0.2);
      }
    }
  }
</style>
