<template>
  <div class="toogo-user-list">
    <n-space vertical>
      <!-- 搜索栏 -->
      <n-space>
        <n-input
          v-model:value="searchParams.username"
          placeholder="用户名"
          clearable
          style="width: 150px"
          @keyup.enter="loadData"
        />
        <n-select
          v-model:value="searchParams.vipLevel"
          :options="vipOptions"
          style="width: 120px"
          clearable
          placeholder="VIP等级"
        />
        <n-select
          v-model:value="searchParams.isAgent"
          :options="agentOptions"
          style="width: 120px"
          clearable
          placeholder="代理商"
        />
        <n-button type="primary" @click="loadData">搜索</n-button>
        <n-button @click="resetSearch">重置</n-button>
      </n-space>

      <!-- 数据表格 -->
      <n-data-table
        :columns="columns"
        :data="list"
        :loading="loading"
        :pagination="pagination"
        @update:page="handlePageChange"
        @update:page-size="handlePageSizeChange"
      />
    </n-space>

    <!-- 详情弹窗 -->
    <n-modal
      v-model:show="showDetailModal"
      title="Toogo用户详情"
      preset="card"
      style="width: 700px"
    >
      <n-descriptions :column="2" label-placement="left" bordered v-if="currentRow">
        <n-descriptions-item label="用户ID">{{ currentRow.memberId }}</n-descriptions-item>
        <n-descriptions-item label="用户名">{{ currentRow.username }}</n-descriptions-item>
        <n-descriptions-item label="VIP等级">
          <n-tag :type="currentRow.vipLevel > 5 ? 'warning' : 'info'"
            >V{{ currentRow.vipLevel }}</n-tag
          >
        </n-descriptions-item>
        <n-descriptions-item label="是否代理商">
          <n-tag :type="currentRow.isAgent ? 'success' : 'default'">{{
            currentRow.isAgent ? '是' : '否'
          }}</n-tag>
        </n-descriptions-item>
        <n-descriptions-item label="代理商等级">{{
          currentRow.agentLevel || '-'
        }}</n-descriptions-item>
        <n-descriptions-item label="邀请码">{{ currentRow.inviteCode || '-' }}</n-descriptions-item>
        <n-descriptions-item label="邀请人ID">{{
          currentRow.inviterId || '-'
        }}</n-descriptions-item>
        <n-descriptions-item label="直推人数">{{
          currentRow.inviteCount || 0
        }}</n-descriptions-item>
        <n-descriptions-item label="团队人数">{{ currentRow.teamCount || 0 }}</n-descriptions-item>
        <n-descriptions-item label="消耗算力">{{
          (currentRow.totalPowerConsume || 0).toFixed(2)
        }}</n-descriptions-item>
        <n-descriptions-item label="团队消耗">{{
          (currentRow.teamTotalPowerConsume || 0).toFixed(2)
        }}</n-descriptions-item>
        <n-descriptions-item label="机器人限制">{{
          currentRow.robotLimit || 1
        }}</n-descriptions-item>
        <n-descriptions-item label="算力折扣"
          >{{ currentRow.powerDiscount || 100 }}%</n-descriptions-item
        >
        <n-descriptions-item label="创建时间">{{ currentRow.createdAt }}</n-descriptions-item>
        <n-descriptions-item label="更新时间">{{ currentRow.updatedAt }}</n-descriptions-item>
      </n-descriptions>

      <n-divider />

      <n-space>
        <n-button type="primary" @click="editVipLevel">调整VIP等级</n-button>
        <n-button type="info" @click="editAgentStatus">设置代理商</n-button>
        <n-button type="warning" @click="editRobotLimit">调整机器人限制</n-button>
      </n-space>
    </n-modal>

    <!-- 调整VIP等级弹窗 -->
    <n-modal v-model:show="showVipModal" title="调整VIP等级" preset="dialog" style="width: 400px">
      <n-form-item label="VIP等级">
        <n-select v-model:value="editForm.vipLevel" :options="allVipOptions" />
      </n-form-item>
      <template #action>
        <n-space>
          <n-button @click="showVipModal = false">取消</n-button>
          <n-button type="primary" @click="saveVipLevel" :loading="saveLoading">保存</n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- 设置代理商弹窗 -->
    <n-modal v-model:show="showAgentModal" title="设置代理商" preset="dialog" style="width: 400px">
      <n-form>
        <n-form-item label="是否代理商">
          <n-switch v-model:value="editForm.isAgent" />
        </n-form-item>
        <n-form-item label="代理商等级" v-if="editForm.isAgent">
          <n-input-number v-model:value="editForm.agentLevel" :min="1" :max="10" />
        </n-form-item>
      </n-form>
      <template #action>
        <n-space>
          <n-button @click="showAgentModal = false">取消</n-button>
          <n-button type="primary" @click="saveAgentStatus" :loading="saveLoading">保存</n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- 调整机器人限制弹窗 -->
    <n-modal
      v-model:show="showRobotModal"
      title="调整机器人限制"
      preset="dialog"
      style="width: 400px"
    >
      <n-form-item label="机器人数量限制">
        <n-input-number v-model:value="editForm.robotLimit" :min="1" :max="100" />
      </n-form-item>
      <template #action>
        <n-space>
          <n-button @click="showRobotModal = false">取消</n-button>
          <n-button type="primary" @click="saveRobotLimit" :loading="saveLoading">保存</n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
  import { ref, onMounted, h } from 'vue';
  import { useMessage } from 'naive-ui';
  import { NButton, NTag, NAvatar } from 'naive-ui';
  import { http } from '@/utils/http/axios';

  const message = useMessage();

  const list = ref<any[]>([]);
  const loading = ref(false);
  const showDetailModal = ref(false);
  const showVipModal = ref(false);
  const showAgentModal = ref(false);
  const showRobotModal = ref(false);
  const saveLoading = ref(false);
  const currentRow = ref<any>(null);

  const searchParams = ref({
    username: '',
    vipLevel: null as number | null,
    isAgent: null as number | null,
    page: 1,
    pageSize: 10,
  });

  const pagination = ref({
    page: 1,
    pageSize: 10,
    itemCount: 0,
    showSizePicker: true,
    pageSizes: [10, 20, 50, 100],
  });

  const editForm = ref({
    vipLevel: 1,
    isAgent: false,
    agentLevel: 1,
    robotLimit: 1,
  });

  // 选项
  const vipOptions = Array.from({ length: 10 }, (_, i) => ({ label: `V${i + 1}`, value: i + 1 }));
  const allVipOptions = [{ label: '无', value: 0 }, ...vipOptions];
  const agentOptions = [
    { label: '是', value: 1 },
    { label: '否', value: 0 },
  ];

  // 表格列
  const columns = [
    { title: 'ID', key: 'memberId', width: 80 },
    {
      title: '用户',
      key: 'username',
      render: (row: any) =>
        h('div', { style: { display: 'flex', alignItems: 'center', gap: '8px' } }, [
          h(NAvatar, {
            src: row.avatar,
            round: true,
            size: 32,
            fallbackSrc: 'https://via.placeholder.com/32',
          }),
          h('span', {}, row.username || `用户${row.memberId}`),
        ]),
    },
    {
      title: 'VIP等级',
      key: 'vipLevel',
      width: 90,
      render: (row: any) =>
        h(
          NTag,
          {
            type: row.vipLevel > 5 ? 'warning' : row.vipLevel > 0 ? 'info' : 'default',
            size: 'small',
          },
          { default: () => (row.vipLevel > 0 ? `V${row.vipLevel}` : '无') },
        ),
    },
    {
      title: '代理商',
      key: 'isAgent',
      width: 80,
      render: (row: any) =>
        h(
          NTag,
          { type: row.isAgent ? 'success' : 'default', size: 'small' },
          { default: () => (row.isAgent ? '是' : '否') },
        ),
    },
    { title: '直推', key: 'inviteCount', width: 60 },
    { title: '团队', key: 'teamCount', width: 60 },
    {
      title: '消耗算力',
      key: 'totalPowerConsume',
      width: 100,
      render: (row: any) => (row.totalPowerConsume || 0).toFixed(2),
    },
    {
      title: '机器人',
      key: 'robotLimit',
      width: 80,
      render: (row: any) => `${row.activeRobots || 0}/${row.robotLimit || 1}`,
    },
    {
      title: '操作',
      key: 'actions',
      width: 100,
      render: (row: any) =>
        h(
          NButton,
          { size: 'small', type: 'primary', onClick: () => viewDetail(row) },
          { default: () => '详情' },
        ),
    },
  ];

  // 加载数据
  const loadData = async () => {
    loading.value = true;
    try {
      const params: any = {
        page: searchParams.value.page,
        pageSize: searchParams.value.pageSize,
      };
      if (searchParams.value.username) params.username = searchParams.value.username;
      if (searchParams.value.vipLevel !== null) params.vipLevel = searchParams.value.vipLevel;
      if (searchParams.value.isAgent !== null) params.isAgent = searchParams.value.isAgent;

      const res = await http.request({
        url: '/toogoUser/list',
        method: 'get',
        params,
      });

      list.value = res?.list || [];
      pagination.value.itemCount = res?.page?.totalCount || 0;
    } catch (error: any) {
      console.error('加载Toogo用户失败:', error);
      message.error(error?.message || '加载失败');
    } finally {
      loading.value = false;
    }
  };

  // 重置搜索
  const resetSearch = () => {
    searchParams.value = {
      username: '',
      vipLevel: null,
      isAgent: null,
      page: 1,
      pageSize: 10,
    };
    loadData();
  };

  // 分页
  const handlePageChange = (page: number) => {
    searchParams.value.page = page;
    pagination.value.page = page;
    loadData();
  };

  const handlePageSizeChange = (pageSize: number) => {
    searchParams.value.pageSize = pageSize;
    pagination.value.pageSize = pageSize;
    searchParams.value.page = 1;
    pagination.value.page = 1;
    loadData();
  };

  // 查看详情
  const viewDetail = (row: any) => {
    currentRow.value = row;
    showDetailModal.value = true;
  };

  // 编辑VIP等级
  const editVipLevel = () => {
    editForm.value.vipLevel = currentRow.value.vipLevel || 0;
    showVipModal.value = true;
  };

  // 保存VIP等级
  const saveVipLevel = async () => {
    saveLoading.value = true;
    try {
      await http.request({
        url: '/toogoUser/updateVip',
        method: 'post',
        params: {
          memberId: currentRow.value.memberId,
          vipLevel: editForm.value.vipLevel,
        },
      });
      message.success('VIP等级已更新');
      showVipModal.value = false;
      loadData();
    } catch (error: any) {
      message.error(error?.message || '更新失败');
    } finally {
      saveLoading.value = false;
    }
  };

  // 编辑代理商状态
  const editAgentStatus = () => {
    editForm.value.isAgent = !!currentRow.value.isAgent;
    editForm.value.agentLevel = currentRow.value.agentLevel || 1;
    showAgentModal.value = true;
  };

  // 保存代理商状态
  const saveAgentStatus = async () => {
    saveLoading.value = true;
    try {
      await http.request({
        url: '/toogoUser/updateAgent',
        method: 'post',
        params: {
          memberId: currentRow.value.memberId,
          isAgent: editForm.value.isAgent ? 1 : 0,
          agentLevel: editForm.value.isAgent ? editForm.value.agentLevel : 0,
        },
      });
      message.success('代理商设置已更新');
      showAgentModal.value = false;
      loadData();
    } catch (error: any) {
      message.error(error?.message || '更新失败');
    } finally {
      saveLoading.value = false;
    }
  };

  // 编辑机器人限制
  const editRobotLimit = () => {
    editForm.value.robotLimit = currentRow.value.robotLimit || 1;
    showRobotModal.value = true;
  };

  // 保存机器人限制
  const saveRobotLimit = async () => {
    saveLoading.value = true;
    try {
      await http.request({
        url: '/toogoUser/updateRobotLimit',
        method: 'post',
        params: {
          memberId: currentRow.value.memberId,
          robotLimit: editForm.value.robotLimit,
        },
      });
      message.success('机器人限制已更新');
      showRobotModal.value = false;
      loadData();
    } catch (error: any) {
      message.error(error?.message || '更新失败');
    } finally {
      saveLoading.value = false;
    }
  };

  onMounted(() => {
    loadData();
  });
</script>

<style scoped lang="less">
  .toogo-user-list {
    padding: 12px 0;
  }
</style>
