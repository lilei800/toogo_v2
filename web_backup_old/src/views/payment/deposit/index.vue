<template>
  <div>
    <n-grid :cols="24" :x-gap="12">
      <!-- 左侧：创建充值订单 -->
      <n-gi :span="8">
        <n-card title="💰 创建充值订单" :bordered="false">
          <n-form ref="formRef" :model="formValue" :rules="rules">
            <n-form-item label="充值金额" path="amount">
              <n-input-number
                v-model:value="formValue.amount"
                :min="10"
                :step="10"
                style="width: 100%;"
              >
                <template #suffix>USDT</template>
              </n-input-number>
            </n-form-item>
            <n-form-item label="网络类型" path="network">
              <n-select v-model:value="formValue.network" :options="networkOptions" />
            </n-form-item>
            <n-form-item>
              <n-button type="primary" block :loading="creating" @click="handleCreate">
                创建充值订单
              </n-button>
            </n-form-item>
          </n-form>
        </n-card>
      </n-gi>

      <!-- 右侧：订单展示和列表 -->
      <n-gi :span="16">
        <!-- 支付信息卡片 -->
        <n-card v-if="paymentInfo" title="📱 扫码充值" :bordered="false" style="margin-bottom: 12px;">
          <n-grid :cols="2" :x-gap="12">
            <n-gi>
              <div style="text-align: center;">
                <qrcode-vue :value="paymentInfo.payAddress" :size="200" level="H" />
              </div>
            </n-gi>
            <n-gi>
              <n-descriptions :column="1">
                <n-descriptions-item label="充值地址">
                  <n-text code>{{ paymentInfo.payAddress }}</n-text>
                  <n-button text type="primary" @click="copyText(paymentInfo.payAddress)">
                    复制
                  </n-button>
                </n-descriptions-item>
                <n-descriptions-item label="应付金额">
                  <n-text type="success" strong>{{ paymentInfo.payAmount }} USDT</n-text>
                </n-descriptions-item>
                <n-descriptions-item label="网络">
                  {{ paymentInfo.network }}
                </n-descriptions-item>
                <n-descriptions-item label="订单号">
                  <n-text code>{{ paymentInfo.orderSn }}</n-text>
                </n-descriptions-item>
              </n-descriptions>
            </n-gi>
          </n-grid>
        </n-card>

        <!-- 订单列表 -->
        <n-card title="📋 充值记录" :bordered="false">
          <n-data-table
            :columns="columns"
            :data="dataList"
            :pagination="pagination"
            :loading="loading"
            :scroll-x="1200"
          />
        </n-card>
      </n-gi>
    </n-grid>
  </div>
</template>

<script lang="ts" setup>
import { ref, reactive, onMounted, h } from 'vue';
import { NGrid, NGi, NCard, NForm, NFormItem, NInputNumber, NSelect, NButton, NText, NDescriptions, NDescriptionsItem, NDataTable, NTag, NSpace, useMessage } from 'naive-ui';
import QrcodeVue from 'qrcode.vue';
import { createDeposit, depositList, checkDeposit, cancelDeposit } from '@/api/payment/deposit';

const message = useMessage();

const formValue = reactive({ amount: 100, network: 'TRC20' });
const rules = {
  amount: { required: true, type: 'number', min: 10, message: '最小充值10 USDT', trigger: 'blur' },
  network: { required: true, message: '请选择网络', trigger: 'change' },
};

const networkOptions = [
  { label: 'TRC20 (推荐)', value: 'TRC20' },
  { label: 'ERC20', value: 'ERC20' },
  { label: 'BEP20', value: 'BEP20' },
];

const paymentInfo = ref<any>(null);
const creating = ref(false);
const dataList = ref([]);
const loading = ref(false);
const pagination = reactive({ page: 1, pageSize: 10, pageCount: 1 });

const columns = [
  { title: 'ID', key: 'id', width: 80 },
  { title: '订单号', key: 'orderSn', width: 180 },
  { title: '金额', key: 'amount', width: 120 },
  { title: '网络', key: 'network', width: 100 },
  { title: '状态', key: 'status', width: 100, render(row: any) {
    const types: any = { 1: 'warning', 2: 'info', 3: 'success', 4: 'default', 5: 'error' };
    const texts: any = { 1: '待支付', 2: '确认中', 3: '已完成', 4: '已过期', 5: '已取消' };
    return h(NTag, { type: types[row.status] }, { default: () => texts[row.status] });
  }},
  { title: '创建时间', key: 'createdAt', width: 180 },
  { title: '操作', key: 'actions', width: 180, render(row: any) {
    return h(NSpace, {}, {
      default: () => [
        row.status === 1 && h(NButton, { size: 'small', onClick: () => handleCheck(row) }, { default: () => '刷新' }),
        row.status === 1 && h(NButton, { size: 'small', type: 'error', onClick: () => handleCancel(row) }, { default: () => '取消' }),
      ]
    });
  }},
];

const handleCreate = async () => {
  creating.value = true;
  try {
    const res = await createDeposit(formValue);
    paymentInfo.value = res;
    message.success('订单创建成功');
    loadData();
  } catch (error: any) {
    message.error(error.message || '创建失败');
  } finally {
    creating.value = false;
  }
};

const loadData = async () => {
  loading.value = true;
  try {
    const res = await depositList({ page: pagination.page, pageSize: pagination.pageSize });
    dataList.value = res.list || [];
    pagination.pageCount = Math.ceil((res.total || 0) / pagination.pageSize);
  } catch (error: any) {
    message.error(error.message || '加载失败');
  } finally {
    loading.value = false;
  }
};

const handleCheck = async (row: any) => {
  try {
    await checkDeposit({ id: row.id });
    message.success('状态已更新');
    loadData();
  } catch (error: any) {
    message.error(error.message || '刷新失败');
  }
};

const handleCancel = async (row: any) => {
  try {
    await cancelDeposit({ id: row.id });
    message.success('订单已取消');
    loadData();
  } catch (error: any) {
    message.error(error.message || '取消失败');
  }
};

const copyText = (text: string) => {
  navigator.clipboard.writeText(text);
  message.success('已复制到剪贴板');
};

onMounted(() => {
  loadData();
});
</script>

