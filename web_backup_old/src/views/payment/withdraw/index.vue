<template>
  <div>
    <n-grid :cols="24" :x-gap="12">
      <n-gi :span="8">
        <n-card title="💸 申请提现" :bordered="false">
          <n-form ref="formRef" :model="formValue" :rules="rules">
            <n-form-item label="提现金额" path="amount">
              <n-input-number v-model:value="formValue.amount" :min="10" style="width: 100%;">
                <template #suffix>USDT</template>
              </n-input-number>
            </n-form-item>
            <n-form-item label="提现地址" path="toAddress">
              <n-input v-model:value="formValue.toAddress" type="textarea" :rows="3" placeholder="请输入USDT地址" />
            </n-form-item>
            <n-form-item label="网络类型" path="network">
              <n-select v-model:value="formValue.network" :options="networkOptions" />
            </n-form-item>
            <n-form-item>
              <n-alert type="warning" title="提现提示" style="margin-bottom: 12px;">
                最小提现: 10 USDT<br/>
                手续费: 1 USDT<br/>
                到账时间: 1-24小时
              </n-alert>
            </n-form-item>
            <n-form-item>
              <n-button type="primary" block :loading="submitting" @click="handleSubmit">
                申请提现
              </n-button>
            </n-form-item>
          </n-form>
        </n-card>
      </n-gi>

      <n-gi :span="16">
        <n-card title="📋 提现记录" :bordered="false">
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
import { NGrid, NGi, NCard, NForm, NFormItem, NInput, NInputNumber, NSelect, NButton, NAlert, NDataTable, NTag, NText, useMessage } from 'naive-ui';
import { applyWithdraw, withdrawList } from '@/api/payment/withdraw';

const message = useMessage();

const formValue = reactive({ amount: 100, toAddress: '', network: 'TRC20' });
const rules = {
  amount: { required: true, type: 'number', min: 10, message: '最小提现10 USDT', trigger: 'blur' },
  toAddress: { required: true, message: '请输入提现地址', trigger: 'blur' },
  network: { required: true, message: '请选择网络', trigger: 'change' },
};

const networkOptions = [
  { label: 'TRC20', value: 'TRC20' },
  { label: 'ERC20', value: 'ERC20' },
  { label: 'BEP20', value: 'BEP20' },
];

const submitting = ref(false);
const dataList = ref([]);
const loading = ref(false);
const pagination = reactive({ page: 1, pageSize: 10, pageCount: 1 });

const columns = [
  { title: 'ID', key: 'id', width: 80 },
  { title: '订单号', key: 'orderSn', width: 180 },
  { title: '金额', key: 'amount', width: 120 },
  { title: '手续费', key: 'fee', width: 100 },
  { title: '实际到账', key: 'actualAmount', width: 120, render(row: any) {
    return h(NText, { type: 'success', strong: true }, { default: () => `${row.actualAmount} USDT` });
  }},
  { title: '提现地址', key: 'toAddress', width: 200, ellipsis: { tooltip: true } },
  { title: '状态', key: 'status', width: 100, render(row: any) {
    const types: any = { 1: 'warning', 2: 'info', 3: 'success', 4: 'error' };
    const texts: any = { 1: '待审核', 2: '处理中', 3: '已完成', 4: '已拒绝' };
    return h(NTag, { type: types[row.status] }, { default: () => texts[row.status] });
  }},
  { title: '创建时间', key: 'createdAt', width: 180 },
];

const handleSubmit = async () => {
  submitting.value = true;
  try {
    await applyWithdraw(formValue);
    message.success('提现申请已提交');
    formValue.amount = 100;
    formValue.toAddress = '';
    loadData();
  } catch (error: any) {
    message.error(error.message || '申请失败');
  } finally {
    submitting.value = false;
  }
};

const loadData = async () => {
  loading.value = true;
  try {
    const res = await withdrawList({ page: pagination.page, pageSize: pagination.pageSize });
    dataList.value = res.list || [];
    pagination.pageCount = Math.ceil((res.total || 0) / pagination.pageSize);
  } catch (error: any) {
    message.error(error.message || '加载失败');
  } finally {
    loading.value = false;
  }
};

onMounted(() => {
  loadData();
});
</script>

