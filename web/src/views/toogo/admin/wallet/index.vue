<template>
  <div class="toogo-admin-wallet">
    <n-card title="💰 用户钱包管理" :bordered="false">
      <template #header-extra>
        <n-button type="primary" @click="loadWalletList">
          <template #icon
            ><n-icon><ReloadOutlined /></n-icon
          ></template>
          刷新
        </n-button>
      </template>

      <!-- 搜索栏 -->
      <n-form inline :show-feedback="false" style="margin-bottom: 16px">
        <n-form-item label="用户名">
          <n-input
            v-model:value="searchForm.username"
            placeholder="搜索用户名"
            clearable
            style="width: 150px"
          />
        </n-form-item>
        <n-form-item label="手机号">
          <n-input
            v-model:value="searchForm.mobile"
            placeholder="搜索手机号"
            clearable
            style="width: 150px"
          />
        </n-form-item>
        <n-form-item>
          <n-space>
            <n-button type="primary" @click="handleSearch">
              <template #icon
                ><n-icon><SearchOutlined /></n-icon
              ></template>
              搜索
            </n-button>
            <n-button @click="handleReset">重置</n-button>
          </n-space>
        </n-form-item>
      </n-form>

      <!-- 用户钱包表格 -->
      <n-data-table
        :columns="columns"
        :data="walletList"
        :loading="loading"
        :pagination="pagination"
        :row-key="(row: any) => row.userId"
        striped
        @update:page="handlePageChange"
        @update:page-size="handlePageSizeChange"
      />
    </n-card>

    <!-- 充值弹窗 -->
    <n-modal
      v-model:show="showRechargeModal"
      preset="card"
      title="💰 手动充值"
      style="width: 520px"
    >
      <n-alert type="info" style="margin-bottom: 16px">
        为用户 <strong>{{ rechargeForm.username }}</strong> (ID: {{ rechargeForm.userId }})
        进行手动充值
      </n-alert>

      <n-form
        ref="rechargeFormRef"
        :model="rechargeForm"
        :rules="rechargeRules"
        label-placement="left"
        label-width="100"
      >
        <n-form-item label="当前余额">
          <n-space>
            <n-tag type="primary" size="large"
              >💵 余额: {{ rechargeForm.currentBalance?.toFixed(2) || 0 }} U</n-tag
            >
            <n-tag type="success" size="large"
              >⚡ 算力: {{ rechargeForm.currentPower?.toFixed(2) || 0 }}</n-tag
            >
            <n-tag type="warning" size="large"
              >🎁 积分: {{ rechargeForm.currentGiftPower?.toFixed(2) || 0 }}</n-tag
            >
          </n-space>
        </n-form-item>

        <n-form-item label="充值类型" path="accountType">
          <n-radio-group v-model:value="rechargeForm.accountType" size="large">
            <n-radio-button value="power">
              <n-space align="center" :size="4">
                <span>⚡</span>
                <span>算力</span>
              </n-space>
            </n-radio-button>
            <n-radio-button value="gift_power">
              <n-space align="center" :size="4">
                <span>🎁</span>
                <span>积分</span>
              </n-space>
            </n-radio-button>
            <n-radio-button value="balance">
              <n-space align="center" :size="4">
                <span>💵</span>
                <span>余额(USDT)</span>
              </n-space>
            </n-radio-button>
          </n-radio-group>
        </n-form-item>

        <n-form-item label="充值金额" path="amount">
          <n-input-number
            v-model:value="rechargeForm.amount"
            :min="0.01"
            :precision="2"
            style="width: 100%"
            size="large"
          >
            <template #prefix>
              <span v-if="rechargeForm.accountType === 'balance'">$</span>
              <span v-else>⚡</span>
            </template>
            <template #suffix>
              {{ rechargeForm.accountType === 'balance' ? 'USDT' : '算力' }}
            </template>
          </n-input-number>
        </n-form-item>

        <n-form-item label="快捷金额">
          <n-space>
            <n-button
              v-for="amt in [10, 50, 100, 500, 1000, 5000]"
              :key="amt"
              @click="rechargeForm.amount = amt"
              size="small"
              :type="rechargeForm.amount === amt ? 'primary' : 'default'"
            >
              {{ amt }}
            </n-button>
          </n-space>
        </n-form-item>

        <n-form-item label="备注">
          <n-input
            v-model:value="rechargeForm.remark"
            type="textarea"
            :rows="2"
            placeholder="充值原因/备注（可选）"
          />
        </n-form-item>
      </n-form>

      <template #footer>
        <n-space justify="end">
          <n-button @click="showRechargeModal = false">取消</n-button>
          <n-button type="primary" :loading="rechargeLoading" @click="handleRecharge">
            <template #icon
              ><n-icon><WalletOutlined /></n-icon
            ></template>
            确认充值
          </n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<script lang="ts" setup>
  import { ref, h, onMounted } from 'vue';
  import { useMessage, NButton, NSpace, NTag } from 'naive-ui';
  import { SearchOutlined, ReloadOutlined, WalletOutlined } from '@vicons/antd';
  import { http } from '@/utils/http/axios';

  const message = useMessage();

  // ==================== 搜索 ====================
  const searchForm = ref({
    username: '',
    mobile: '',
  });

  function handleSearch() {
    pagination.value.page = 1;
    loadWalletList();
  }

  function handleReset() {
    searchForm.value = { username: '', mobile: '' };
    pagination.value.page = 1;
    loadWalletList();
  }

  // ==================== 表格 ====================
  const loading = ref(false);
  const walletList = ref<any[]>([]);
  const pagination = ref({
    page: 1,
    pageSize: 15,
    itemCount: 0,
    showSizePicker: true,
    pageSizes: [10, 15, 20, 50, 100],
  });

  const columns = [
    { title: '用户ID', key: 'userId', width: 80 },
    { title: '用户名', key: 'username', width: 120, ellipsis: { tooltip: true } },
    { title: '手机号', key: 'mobile', width: 130 },
    {
      title: '余额(USDT)',
      key: 'balance',
      width: 120,
      render: (row: any) =>
        h(
          'span',
          { style: 'color: #2080f0; font-weight: 600; font-family: monospace' },
          (row.balance || 0).toFixed(2),
        ),
    },
    {
      title: '算力',
      key: 'power',
      width: 120,
      render: (row: any) =>
        h(
          'span',
          { style: 'color: #18a058; font-weight: 600; font-family: monospace' },
          (row.power || 0).toFixed(2),
        ),
    },
    {
      title: '积分',
      key: 'giftPower',
      width: 100,
      render: (row: any) =>
        h(
          'span',
          { style: 'color: #f0a020; font-weight: 600; font-family: monospace' },
          (row.giftPower || 0).toFixed(2),
        ),
    },
    {
      title: '总可用算力',
      key: 'totalPower',
      width: 120,
      render: (row: any) =>
        h(NTag, { type: 'success', size: 'small' }, () => (row.totalPower || 0).toFixed(2)),
    },
    {
      title: '佣金',
      key: 'commission',
      width: 100,
      render: (row: any) =>
        h(
          'span',
          { style: 'color: #8b5cf6; font-family: monospace' },
          (row.commission || 0).toFixed(2),
        ),
    },
    {
      title: 'VIP等级',
      key: 'vipLevel',
      width: 80,
      render: (row: any) =>
        row.vipLevel > 0
          ? h(NTag, { type: 'warning', size: 'small' }, () => `V${row.vipLevel}`)
          : h('span', { style: 'color: #999' }, '--'),
    },
    {
      title: '操作',
      key: 'actions',
      width: 100,
      fixed: 'right',
      render: (row: any) =>
        h(NSpace, { size: 'small' }, () => [
          h(
            NButton,
            {
              type: 'primary',
              size: 'small',
              onClick: () => openRechargeModal(row),
            },
            () => '充值',
          ),
        ]),
    },
  ];

  async function loadWalletList() {
    loading.value = true;
    try {
      const res = await http.request({
        url: '/toogo/wallet/userList',
        method: 'get',
        params: {
          username: searchForm.value.username || undefined,
          mobile: searchForm.value.mobile || undefined,
          page: pagination.value.page,
          perPage: pagination.value.pageSize,
        },
      });
      if (res.code === 0) {
        walletList.value = res.data?.list || [];
        pagination.value.itemCount = res.data?.totalCount || 0;
      } else {
        message.error(res.message || '加载失败');
      }
    } catch (error: any) {
      console.error('加载用户钱包列表失败', error);
      message.error(error.message || '加载失败');
    } finally {
      loading.value = false;
    }
  }

  function handlePageChange(page: number) {
    pagination.value.page = page;
    loadWalletList();
  }

  function handlePageSizeChange(pageSize: number) {
    pagination.value.pageSize = pageSize;
    pagination.value.page = 1;
    loadWalletList();
  }

  // ==================== 充值功能 ====================
  const showRechargeModal = ref(false);
  const rechargeLoading = ref(false);
  const rechargeFormRef = ref<any>(null);
  const rechargeForm = ref({
    userId: 0,
    username: '',
    currentBalance: 0,
    currentPower: 0,
    currentGiftPower: 0,
    accountType: 'power',
    amount: 100,
    remark: '',
  });

  const rechargeRules = {
    accountType: { required: true, message: '请选择充值类型', trigger: 'change' },
    amount: {
      required: true,
      type: 'number',
      min: 0.01,
      message: '请输入充值金额',
      trigger: 'blur',
    },
  };

  function openRechargeModal(row: any) {
    rechargeForm.value = {
      userId: row.userId,
      username: row.username || row.mobile || `用户${row.userId}`,
      currentBalance: row.balance || 0,
      currentPower: row.power || 0,
      currentGiftPower: row.giftPower || 0,
      accountType: 'power',
      amount: 100,
      remark: '',
    };
    showRechargeModal.value = true;
  }

  async function handleRecharge() {
    rechargeLoading.value = true;
    try {
      const res = await http.request({
        url: '/toogo/wallet/adminRecharge',
        method: 'post',
        data: {
          userId: rechargeForm.value.userId,
          accountType: rechargeForm.value.accountType,
          amount: rechargeForm.value.amount,
          remark: rechargeForm.value.remark || '管理员手动充值',
        },
      });
      if (res.code === 0) {
        const typeMap: Record<string, string> = {
          power: '算力',
          gift_power: '积分',
          balance: '余额',
        };
        message.success(
          `充值成功！${typeMap[rechargeForm.value.accountType]}：${res.data?.beforeAmount?.toFixed(
            2,
          )} → ${res.data?.afterAmount?.toFixed(2)}`,
        );
        showRechargeModal.value = false;
        loadWalletList();
      } else {
        message.error(res.message || '充值失败');
      }
    } catch (error: any) {
      message.error(error.message || '充值失败');
    } finally {
      rechargeLoading.value = false;
    }
  }

  onMounted(() => {
    loadWalletList();
  });
</script>

<style lang="less" scoped>
  .toogo-admin-wallet {
    :deep(.n-data-table) {
      .n-data-table-td {
        padding: 8px 12px;
      }
    }
  }
</style>
