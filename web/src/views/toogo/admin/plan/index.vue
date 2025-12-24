<template>
  <div class="plan-admin-page">
    <n-card title="套餐管理">
      <template #header-extra>
        <n-button type="primary" @click="openEditModal()">新增套餐</n-button>
      </template>
      <n-data-table
        :columns="columns"
        :data="list"
        :loading="loading"
        :pagination="pagination"
        @update:page="handlePageChange"
      />
    </n-card>

    <!-- 编辑弹窗 -->
    <n-modal v-model:show="showEditModal" :title="editForm.id ? '编辑套餐' : '新增套餐'" preset="card" style="width: 700px">
      <n-form ref="formRef" :model="editForm" :rules="rules" label-placement="left" label-width="120">
        <n-grid :cols="2" :x-gap="16">
          <n-gi>
            <n-form-item label="套餐名称" path="planName">
              <n-input v-model:value="editForm.planName" placeholder="如：A套餐" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="套餐代码" path="planCode">
              <n-input v-model:value="editForm.planCode" placeholder="如：A" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="机器人限制" path="robotLimit">
              <n-input-number v-model:value="editForm.robotLimit" :min="1" :max="10000" style="width: 100%" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="排序">
              <n-input-number v-model:value="editForm.sort" :min="0" style="width: 100%" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="是否限制购买">
              <n-switch v-model:value="editForm.purchaseLimitEnabled" :checked-value="1" :unchecked-value="0" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="限制次数">
              <n-input-number
                v-model:value="editForm.purchaseLimit"
                :min="1"
                :precision="0"
                style="width: 100%"
                placeholder="1表示仅可购买1次"
                :disabled="editForm.purchaseLimitEnabled !== 1"
              />
            </n-form-item>
          </n-gi>
          <n-gi :span="2">
            <n-divider>价格设置 (USDT)</n-divider>
          </n-gi>
          <n-gi :span="2">
            <n-form-item label="价格方案">
              <n-dynamic-input
                v-model:value="priceList"
                :min="1"
                :key-field="'periodType'"
                :item-style="{ marginBottom: '12px' }"
                @create="handlePriceCreate"
              >
                <template #default="{ value, index }">
                  <n-space vertical :size="8" style="width: 100%">
                    <n-space style="width: 100%" align="center">
                      <n-select
                        v-model:value="value.periodType"
                        :options="periodTypeOptions"
                        placeholder="选择周期"
                        style="width: 150px"
                        @update:value="(val) => handlePeriodChange(index, val)"
                      />
                      <n-input-number
                        v-model:value="value.price"
                        :min="0"
                        :precision="2"
                        placeholder="价格"
                        style="flex: 1"
                      />
                      <n-text depth="3" style="width: 40px">USDT</n-text>
                      <n-button
                        :type="editForm.defaultPeriod === value.periodType ? 'primary' : 'default'"
                        size="small"
                        @click="editForm.defaultPeriod = value.periodType"
                        style="margin-right: 8px; min-width: 60px"
                      >
                        {{ editForm.defaultPeriod === value.periodType ? '推荐' : '设推荐' }}
                      </n-button>
                      <n-button
                        quaternary
                        type="error"
                        size="small"
                        @click="handlePriceRemove(index)"
                        :disabled="priceList.length <= 1"
                      >
                        删除
                      </n-button>
                    </n-space>
                  </n-space>
                </template>
              </n-dynamic-input>
              <n-button
                quaternary
                type="primary"
                size="small"
                @click="handlePriceAdd"
                style="margin-top: 8px"
              >
                添加价格方案
              </n-button>
            </n-form-item>
          </n-gi>
          <n-gi :span="2">
            <n-divider>赠送积分（已停用）</n-divider>
          </n-gi>
          <n-gi>
            <n-form-item label="月付赠送">
              <n-input-number v-model:value="editForm.giftPowerMonthly" :min="0" style="width: 100%" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="季付赠送">
              <n-input-number v-model:value="editForm.giftPowerQuarterly" :min="0" style="width: 100%" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="半年赠送">
              <n-input-number v-model:value="editForm.giftPowerHalfYear" :min="0" style="width: 100%" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="年付赠送">
              <n-input-number v-model:value="editForm.giftPowerYearly" :min="0" style="width: 100%" />
            </n-form-item>
          </n-gi>
          <n-gi :span="2">
            <n-form-item label="套餐描述">
              <n-input v-model:value="editForm.description" type="textarea" :rows="2" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="推荐套餐">
              <n-switch v-model:value="editForm.isDefault" :checked-value="1" :unchecked-value="0" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="状态">
              <n-switch v-model:value="editForm.status" :checked-value="1" :unchecked-value="0" />
            </n-form-item>
          </n-gi>
        </n-grid>
      </n-form>
      <template #action>
        <n-space justify="end">
          <n-button @click="showEditModal = false">取消</n-button>
          <n-button type="primary" @click="handleSave" :loading="saveLoading">保存</n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, h, watch, nextTick } from 'vue';
import { useMessage, useDialog } from 'naive-ui';
import { NButton, NTag, NSpace } from 'naive-ui';
import { ToogoSubscriptionApi } from '@/api/toogo';

const message = useMessage();
const dialog = useDialog();

const list = ref<any[]>([]);
const loading = ref(false);
const showEditModal = ref(false);
const saveLoading = ref(false);
const formRef = ref();

const pagination = ref({
  page: 1,
  pageSize: 10,
  showSizePicker: true,
  pageSizes: [10, 20, 50],
  itemCount: 0,
});

const editForm = ref({
  id: 0,
  planName: '',
  planCode: '',
  robotLimit: 1,
  purchaseLimitEnabled: 0,
  purchaseLimit: 0, // 0为不限（保存时由 purchaseLimitEnabled 决定）
  priceDaily: 0,
  priceMonthly: 0,
  priceQuarterly: 0,
  priceHalfYear: 0,
  priceYearly: 0,
  defaultPeriod: '', // 推荐价格方案
  giftPowerMonthly: 0,
  giftPowerQuarterly: 0,
  giftPowerHalfYear: 0,
  giftPowerYearly: 0,
  description: '',
  isDefault: 0,
  sort: 0,
  status: 1,
});

// 价格列表（动态）
const priceList = ref<Array<{ periodType: string; price: number }>>([
  { periodType: 'monthly', price: 0 },
]);

// 周期类型选项
const periodTypeOptions = [
  { label: '日付', value: 'daily' },
  { label: '月付', value: 'monthly' },
  { label: '季付', value: 'quarterly' },
  { label: '半年付', value: 'half_year' },
  { label: '年付', value: 'yearly' },
];

// 获取已使用的周期类型
const getUsedPeriodTypes = () => {
  return priceList.value.map(item => item.periodType);
};

// 处理价格创建
const handlePriceCreate = () => {
  return { periodType: 'monthly', price: 0 };
};

// 存储每个价格项的旧周期值，用于恢复
const oldPeriodValues = ref<Map<number, string>>(new Map());

// 处理周期变化
const handlePeriodChange = (index: number, newPeriodType: string) => {
  // 检查是否已存在该周期
  const existingIndex = priceList.value.findIndex((item, idx) => idx !== index && item.periodType === newPeriodType);
  if (existingIndex !== -1) {
    message.warning('该周期已存在，请选择其他周期');
    // 恢复原值
    const oldValue = oldPeriodValues.value.get(index);
    if (oldValue) {
      nextTick(() => {
        priceList.value[index].periodType = oldValue;
      });
    }
    return;
  }
  // 保存当前值作为下次的旧值
  oldPeriodValues.value.set(index, newPeriodType);
};

// 添加价格方案
const handlePriceAdd = () => {
  // 找到第一个未使用的周期
  const usedTypes = getUsedPeriodTypes();
  const availableType = periodTypeOptions.find(opt => !usedTypes.includes(opt.value));
  if (availableType) {
    const newIndex = priceList.value.length;
    priceList.value.push({ periodType: availableType.value, price: 0 });
    // 保存新项的初始周期值
    oldPeriodValues.value.set(newIndex, availableType.value);
  } else {
    message.warning('所有周期类型都已添加');
  }
};

// 删除价格方案
const handlePriceRemove = (index: number) => {
  if (priceList.value.length <= 1) {
    message.warning('至少需要保留一个价格方案');
    return;
  }
  priceList.value.splice(index, 1);
  // 清理旧值记录并重新索引
  oldPeriodValues.value.clear();
  priceList.value.forEach((item, idx) => {
    oldPeriodValues.value.set(idx, item.periodType);
  });
};

// 同步价格列表到表单
const syncPriceListToForm = () => {
  // 重置所有价格
  editForm.value.priceDaily = 0;
  editForm.value.priceMonthly = 0;
  editForm.value.priceQuarterly = 0;
  editForm.value.priceHalfYear = 0;
  editForm.value.priceYearly = 0;
  
  // 根据价格列表设置对应字段
  priceList.value.forEach(item => {
    switch (item.periodType) {
      case 'daily':
        editForm.value.priceDaily = item.price || 0;
        break;
      case 'monthly':
        editForm.value.priceMonthly = item.price || 0;
        break;
      case 'quarterly':
        editForm.value.priceQuarterly = item.price || 0;
        break;
      case 'half_year':
        editForm.value.priceHalfYear = item.price || 0;
        break;
      case 'yearly':
        editForm.value.priceYearly = item.price || 0;
        break;
    }
  });
  
  // 如果默认周期不在价格列表中，设置为第一个价格方案
  if (priceList.value.length > 0) {
    const hasDefaultPeriod = priceList.value.some(item => item.periodType === editForm.value.defaultPeriod);
    if (!hasDefaultPeriod) {
      editForm.value.defaultPeriod = priceList.value[0].periodType;
    }
  }
};

// 从表单同步到价格列表
const syncFormToPriceList = () => {
  const prices: Array<{ periodType: string; price: number }> = [];
  
  // 添加所有价格 > 0 的方案
  if (editForm.value.priceDaily > 0) prices.push({ periodType: 'daily', price: editForm.value.priceDaily || 0 });
  if (editForm.value.priceMonthly > 0) prices.push({ periodType: 'monthly', price: editForm.value.priceMonthly || 0 });
  if (editForm.value.priceQuarterly > 0) prices.push({ periodType: 'quarterly', price: editForm.value.priceQuarterly || 0 });
  if (editForm.value.priceHalfYear > 0) prices.push({ periodType: 'half_year', price: editForm.value.priceHalfYear || 0 });
  if (editForm.value.priceYearly > 0) prices.push({ periodType: 'yearly', price: editForm.value.priceYearly || 0 });
  
  // 如果推荐周期对应的价格为 0，也要添加到价格列表中（确保推荐周期可见）
  const recommendedPeriod = editForm.value.defaultPeriod;
  if (recommendedPeriod && !prices.some(p => p.periodType === recommendedPeriod)) {
    let recommendedPrice = 0;
    switch (recommendedPeriod) {
      case 'daily':
        recommendedPrice = editForm.value.priceDaily || 0;
        break;
      case 'monthly':
        recommendedPrice = editForm.value.priceMonthly || 0;
        break;
      case 'quarterly':
        recommendedPrice = editForm.value.priceQuarterly || 0;
        break;
      case 'half_year':
        recommendedPrice = editForm.value.priceHalfYear || 0;
        break;
      case 'yearly':
        recommendedPrice = editForm.value.priceYearly || 0;
        break;
    }
    prices.push({ periodType: recommendedPeriod, price: recommendedPrice });
  }
  
  // 如果没有任何价格，至少保留一个月付
  if (prices.length === 0) {
    prices.push({ periodType: 'monthly', price: 0 });
  }
  
  priceList.value = prices;
  // 同步旧值记录
  oldPeriodValues.value.clear();
  priceList.value.forEach((item, idx) => {
    oldPeriodValues.value.set(idx, item.periodType);
  });
  
  // 如果默认周期不在价格列表中，设置为第一个价格方案
  if (priceList.value.length > 0) {
    const hasDefaultPeriod = priceList.value.some(item => item.periodType === editForm.value.defaultPeriod);
    if (!hasDefaultPeriod) {
      editForm.value.defaultPeriod = priceList.value[0].periodType;
    }
  }
};

// 监听价格列表变化，同步到表单和旧值记录
watch(priceList, () => {
  syncPriceListToForm();
  // 更新旧值记录
  oldPeriodValues.value.clear();
  priceList.value.forEach((item, idx) => {
    oldPeriodValues.value.set(idx, item.periodType);
  });
}, { deep: true });

const rules = {
  planName: { required: true, message: '请输入套餐名称', trigger: 'blur' },
  planCode: { required: true, message: '请输入套餐代码', trigger: 'blur' },
  robotLimit: { required: true, type: 'number', message: '请设置机器人限制', trigger: 'change' },
};

const columns = [
  { title: 'ID', key: 'id', width: 60 },
  { title: '套餐名称', key: 'planName' },
  { title: '代码', key: 'planCode' },
  { title: '机器人限制', key: 'robotLimit' },
  {
    title: '是否限购',
    key: 'purchaseLimit',
    width: 100,
    render: (row: any) =>
      h(NTag, { type: Number(row.purchaseLimit) > 0 ? 'warning' : 'default', size: 'small' }, { default: () => Number(row.purchaseLimit) > 0 ? '限制' : '不限' }),
  },
  {
    title: '限购次数',
    key: 'purchaseLimitCount',
    width: 100,
    render: (row: any) => (Number(row.purchaseLimit) > 0 ? Number(row.purchaseLimit) : '-'),
  },
  { title: '日付', key: 'priceDaily' },
  { title: '月付', key: 'priceMonthly' },
  { title: '季付', key: 'priceQuarterly' },
  { title: '半年付', key: 'priceHalfYear' },
  { title: '年付', key: 'priceYearly' },
  {
    title: '推荐',
    key: 'isDefault',
    render: (row: any) => h(NTag, { type: row.isDefault ? 'success' : 'default', size: 'small' }, { default: () => row.isDefault ? '是' : '否' }),
  },
  {
    title: '状态',
    key: 'status',
    render: (row: any) => h(NTag, { type: row.status === 1 ? 'success' : 'error', size: 'small' }, { default: () => row.status === 1 ? '启用' : '禁用' }),
  },
  {
    title: '操作',
    key: 'actions',
    width: 150,
    render: (row: any) => h(NSpace, {}, {
      default: () => [
        h(NButton, { size: 'small', onClick: () => openEditModal(row) }, { default: () => '编辑' }),
        h(NButton, { size: 'small', type: 'error', onClick: () => handleDelete(row) }, { default: () => '删除' }),
      ],
    }),
  },
];

const loadData = async () => {
  loading.value = true;
  try {
    const res = await ToogoSubscriptionApi.planList({ page: pagination.value.page, perPage: pagination.value.pageSize });
    list.value = res?.list || [];
    pagination.value.itemCount = res?.totalCount || 0;
  } catch (error) {
    console.error('加载失败:', error);
  } finally {
    loading.value = false;
  }
};

const handlePageChange = (page: number) => {
  pagination.value.page = page;
  loadData();
};

const openEditModal = (row?: any) => {
  if (row) {
    // 保存原始的 defaultPeriod 值（使用数据库中的值，不设默认）
    const savedDefaultPeriod = row.defaultPeriod || '';
    
    editForm.value = { 
      ...row,
      defaultPeriod: savedDefaultPeriod,
      purchaseLimitEnabled: Number(row.purchaseLimit) > 0 ? 1 : 0,
      purchaseLimit: Number(row.purchaseLimit) || 0,
    };
    // 从表单数据同步到价格列表
    syncFormToPriceList();
    // 恢复 defaultPeriod（防止被 watch 触发的 syncPriceListToForm 重置）
    editForm.value.defaultPeriod = savedDefaultPeriod;
  } else {
    editForm.value = {
      id: 0,
      planName: '',
      planCode: '',
      robotLimit: 1,
      purchaseLimitEnabled: 0,
      purchaseLimit: 0,
      priceDaily: 0,
      priceMonthly: 0,
      priceQuarterly: 0,
      priceHalfYear: 0,
      priceYearly: 0,
      defaultPeriod: '', // 推荐价格方案
      giftPowerMonthly: 0,
      giftPowerQuarterly: 0,
      giftPowerHalfYear: 0,
      giftPowerYearly: 0,
      description: '',
      isDefault: 0,
      sort: 0,
      status: 1,
    };
    // 重置价格列表
    priceList.value = [{ periodType: 'monthly', price: 0 }];
  }
  showEditModal.value = true;
};

const handleSave = async () => {
  try {
    await formRef.value?.validate();
  } catch (error) {
    return;
  }

  // 限购逻辑：只有选择“限制购买”时才允许输入次数；否则强制置 0
  if (editForm.value.purchaseLimitEnabled !== 1) {
    editForm.value.purchaseLimit = 0;
  } else {
    if (!editForm.value.purchaseLimit || Number(editForm.value.purchaseLimit) < 1) {
      message.error('已开启限制购买，请输入限制次数（>=1）');
      return;
    }
  }

  // 验证价格列表
  if (priceList.value.length === 0) {
    message.error('请至少添加一个价格方案');
    return;
  }

  // 检查是否有重复的周期类型
  const periodTypes = priceList.value.map(item => item.periodType);
  const uniquePeriodTypes = new Set(periodTypes);
  if (periodTypes.length !== uniquePeriodTypes.size) {
    message.error('价格方案中存在重复的周期类型');
    return;
  }

  // 验证推荐周期是否在价格列表中
  const hasRecommendedPeriod = priceList.value.some(item => item.periodType === editForm.value.defaultPeriod);
  if (!hasRecommendedPeriod) {
    message.warning('推荐周期不在价格方案中，已自动设置为第一个价格方案');
    editForm.value.defaultPeriod = priceList.value[0].periodType;
  }

  // 同步价格列表到表单
  syncPriceListToForm();

  saveLoading.value = true;
  try {
    await ToogoSubscriptionApi.planEdit(editForm.value);
    message.success('保存成功');
    showEditModal.value = false;
    loadData();
  } catch (error: any) {
    message.error(error.message || '保存失败');
  } finally {
    saveLoading.value = false;
  }
};

const handleDelete = (row: any) => {
  dialog.error({
    title: '确认删除',
    content: `确定要删除套餐 "${row.planName}" 吗？`,
    positiveText: '确定',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await ToogoSubscriptionApi.planDelete({ id: row.id });
        message.success('删除成功');
        loadData();
      } catch (error: any) {
        message.error(error.message || '删除失败');
      }
    },
  });
};

onMounted(() => {
  loadData();
});
</script>

<style scoped lang="less">
.plan-admin-page {
  padding: 16px;
}
</style>

