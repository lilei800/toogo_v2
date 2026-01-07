<template>
  <div class="commission-page">
    <!-- 统计卡片 -->
    <n-grid cols="1 s:2 m:2 l:4 xl:4 2xl:4" :x-gap="12" :y-gap="12" responsive="screen">
      <n-gi>
        <n-card :bordered="false">
          <n-statistic
            label="今日佣金"
            :value="stat.todayCommission?.toFixed(4) || '0.0000'"
            suffix="USDT"
          />
        </n-card>
      </n-gi>
      <n-gi>
        <n-card :bordered="false">
          <n-statistic
            label="本周佣金"
            :value="stat.weekCommission?.toFixed(4) || '0.0000'"
            suffix="USDT"
          />
        </n-card>
      </n-gi>
      <n-gi>
        <n-card :bordered="false">
          <n-statistic
            label="本月佣金"
            :value="stat.monthCommission?.toFixed(4) || '0.0000'"
            suffix="USDT"
          />
        </n-card>
      </n-gi>
      <n-gi>
        <n-card :bordered="false">
          <n-statistic
            label="累计佣金"
            :value="stat.totalCommission?.toFixed(4) || '0.0000'"
            suffix="USDT"
          />
        </n-card>
      </n-gi>
    </n-grid>

    <!-- 分类统计 -->
    <n-card title="佣金分类" :bordered="false" class="proCard mt-3">
      <n-grid cols="1 s:2 m:2 l:2 xl:2 2xl:2" :x-gap="12" :y-gap="12" responsive="screen">
        <n-gi>
          <n-statistic
            label="邀请奖励"
            :value="stat.inviteReward?.toFixed(2) || '0.00'"
            suffix="积分"
          >
            <template #prefix>
              <n-icon color="#18a058"><PeopleOutline /></n-icon>
            </template>
          </n-statistic>
        </n-gi>
        <n-gi>
          <n-statistic
            label="订阅佣金"
            :value="stat.subscribeCommission?.toFixed(4) || '0.0000'"
            suffix="USDT"
          >
            <template #prefix>
              <n-icon color="#2080f0"><CardOutline /></n-icon>
            </template>
          </n-statistic>
        </n-gi>
      </n-grid>
    </n-card>

    <!-- 佣金记录 -->
    <n-card title="佣金明细" :bordered="false" class="proCard mt-3">
      <template #header-extra>
        <n-space>
          <n-select
            v-model:value="searchParams.commissionType"
            :options="typeOptions"
            style="width: 120px"
            clearable
            placeholder="佣金类型"
            @update:value="loadList"
          />
        </n-space>
      </template>
      <n-data-table
        :columns="columns"
        :data="list"
        :loading="loading"
        :pagination="pagination"
        striped
        :bordered="false"
        @update:page="handlePageChange"
      />
    </n-card>
  </div>
</template>

<script setup lang="ts">
  import { ref, onMounted, h } from 'vue';
  import { useMessage } from 'naive-ui';
  import { NTag } from 'naive-ui';
  import { PeopleOutline, CardOutline } from '@vicons/ionicons5';
  import { ToogoCommissionApi } from '@/api/toogo';

  const message = useMessage();

  const stat = ref<any>({});
  const list = ref<any[]>([]);
  const loading = ref(false);

  const searchParams = ref({
    commissionType: null,
    page: 1,
    perPage: 10,
  });

  const pagination = ref({
    page: 1,
    pageSize: 10,
    showSizePicker: true,
    pageSizes: [10, 20, 50],
    itemCount: 0,
  });

  const typeOptions = [
    { label: '邀请奖励', value: 'invite_reward' },
    { label: '订阅佣金', value: 'subscribe' },
  ];

  const columns = [
    { title: 'ID', key: 'id', width: 60 },
    {
      title: '佣金类型',
      key: 'commissionType',
      render: (row: any) => {
        const types: any = {
          invite_reward: { text: '邀请奖励', type: 'success' },
          subscribe: { text: '订阅佣金', type: 'info' },
        };
        const type = types[row.commissionType] || { text: row.commissionType, type: 'default' };
        return h(NTag, { type: type.type, size: 'small' }, { default: () => type.text });
      },
    },
    { title: '来源用户', key: 'fromUsername' },
    {
      title: '基数',
      key: 'baseAmount',
      render: (row: any) => {
        // 邀请奖励为积分，无基数概念
        if (row.commissionType === 'invite_reward') return '-';
        return (row.baseAmount || 0).toFixed(4);
      },
    },
    {
      title: '比例',
      key: 'commissionRate',
      render: (row: any) => {
        if (row.commissionType === 'invite_reward') return '-';
        return ((row.commissionRate || 0) * 100).toFixed(2) + '%';
      },
    },
    {
      title: '金额',
      key: 'commissionAmount',
      render: (row: any) => {
        const unit = row.settleType === 'power' ? '积分' : 'USDT';
        const fixed = unit === '积分' ? 2 : 4;
        return (row.commissionAmount || 0).toFixed(fixed) + ' ' + unit;
      },
    },
    { title: '时间', key: 'createdAt' },
  ];

  const loadData = async () => {
    try {
      const statRes = await ToogoCommissionApi.stat();
      stat.value = statRes || {};
    } catch (error) {
      console.error('加载数据失败:', error);
    }
  };

  const loadList = async () => {
    loading.value = true;
    try {
      const res = await ToogoCommissionApi.logList(searchParams.value);
      list.value = res?.list || [];
      pagination.value.itemCount = res?.totalCount || 0;
    } catch (error) {
      console.error('加载佣金列表失败:', error);
    } finally {
      loading.value = false;
    }
  };

  const handlePageChange = (page: number) => {
    searchParams.value.page = page;
    loadList();
  };

  onMounted(() => {
    loadData();
    loadList();
  });
</script>

<style scoped lang="less">
  .commission-page {
    padding: 16px;
  }
</style>
