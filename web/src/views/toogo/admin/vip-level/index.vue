<template>
  <div class="vip-level-admin-page">
    <n-card title="VIP等级配置">
      <template #header-extra>
        <n-button type="primary" @click="openEditModal()">新增等级</n-button>
      </template>

      <n-data-table :columns="columns" :data="list" :loading="loading" :pagination="false" />
    </n-card>

    <!-- 编辑弹窗 -->
    <n-modal
      v-model:show="showEditModal"
      :title="editForm.id ? '编辑VIP等级' : '新增VIP等级'"
      preset="card"
      style="width: 600px"
    >
      <n-form
        ref="formRef"
        :model="editForm"
        :rules="rules"
        label-placement="left"
        label-width="120"
      >
        <n-grid :cols="2" :x-gap="16">
          <n-gi>
            <n-form-item label="等级" path="level">
              <n-input-number
                v-model:value="editForm.level"
                :min="1"
                :max="10"
                style="width: 100%"
              />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="等级名称" path="levelName">
              <n-input v-model:value="editForm.levelName" placeholder="如：青铜会员" />
            </n-form-item>
          </n-gi>
          <n-gi :span="2">
            <n-divider>升级条件 (满足任一条件即可升级)</n-divider>
          </n-gi>
          <n-gi>
            <n-form-item label="邀请人数">
              <n-input-number
                v-model:value="editForm.requireInviteCount"
                :min="0"
                style="width: 100%"
              />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="消耗算力">
              <n-input-number
                v-model:value="editForm.requireConsumePower"
                :min="0"
                :precision="2"
                style="width: 100%"
              />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="团队消耗">
              <n-input-number
                v-model:value="editForm.requireTeamConsume"
                :min="0"
                :precision="2"
                style="width: 100%"
              />
            </n-form-item>
          </n-gi>
          <n-gi :span="2">
            <n-divider>等级权益</n-divider>
          </n-gi>
          <n-gi>
            <n-form-item label="算力折扣%">
              <n-input-number
                v-model:value="editForm.powerDiscount"
                :min="0"
                :max="30"
                :precision="2"
                style="width: 100%"
              />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="邀请奖励算力">
              <n-input-number
                v-model:value="editForm.inviteRewardPower"
                :min="0"
                :precision="2"
                style="width: 100%"
              />
            </n-form-item>
          </n-gi>
          <n-gi :span="2">
            <n-form-item label="等级描述">
              <n-input v-model:value="editForm.description" type="textarea" :rows="2" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="图标">
              <n-input v-model:value="editForm.icon" placeholder="图标URL或名称" />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="排序">
              <n-input-number v-model:value="editForm.sort" :min="0" style="width: 100%" />
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
  import { ref, onMounted, h } from 'vue';
  import { useMessage } from 'naive-ui';
  import { NButton, NTag, NSpace } from 'naive-ui';
  import { ToogoUserApi } from '@/api/toogo';

  const message = useMessage();

  const list = ref<any[]>([]);
  const loading = ref(false);
  const showEditModal = ref(false);
  const saveLoading = ref(false);
  const formRef = ref();

  const editForm = ref({
    id: 0,
    level: 1,
    levelName: '',
    requireInviteCount: 0,
    requireConsumePower: 0,
    requireTeamConsume: 0,
    powerDiscount: 0,
    inviteRewardPower: 30,
    description: '',
    icon: '',
    sort: 0,
    status: 1,
  });

  const rules = {
    level: { required: true, type: 'number', message: '请设置等级', trigger: 'change' },
    levelName: { required: true, message: '请输入等级名称', trigger: 'blur' },
  };

  const columns = [
    {
      title: '等级',
      key: 'level',
      render: (row: any) => h(NTag, { type: 'warning' }, { default: () => `V${row.level}` }),
    },
    { title: '等级名称', key: 'levelName' },
    { title: '邀请人数', key: 'requireInviteCount' },
    { title: '消耗算力', key: 'requireConsumePower' },
    { title: '团队消耗', key: 'requireTeamConsume' },
    { title: '算力折扣', key: 'powerDiscount', render: (row: any) => `${row.powerDiscount}%` },
    { title: '邀请奖励', key: 'inviteRewardPower' },
    {
      title: '状态',
      key: 'status',
      render: (row: any) =>
        h(
          NTag,
          { type: row.status === 1 ? 'success' : 'error', size: 'small' },
          { default: () => (row.status === 1 ? '启用' : '禁用') },
        ),
    },
    {
      title: '操作',
      key: 'actions',
      width: 100,
      render: (row: any) =>
        h(NButton, { size: 'small', onClick: () => openEditModal(row) }, { default: () => '编辑' }),
    },
  ];

  const loadData = async () => {
    loading.value = true;
    try {
      const res = await ToogoUserApi.vipLevelList({ page: 1, perPage: 100 });
      list.value = res?.list || [];
    } catch (error) {
      console.error('加载失败:', error);
    } finally {
      loading.value = false;
    }
  };

  const openEditModal = (row?: any) => {
    if (row) {
      editForm.value = { ...row };
    } else {
      editForm.value = {
        id: 0,
        level: (list.value.length || 0) + 1,
        levelName: '',
        requireInviteCount: 0,
        requireConsumePower: 0,
        requireTeamConsume: 0,
        powerDiscount: 0,
        inviteRewardPower: 30,
        description: '',
        icon: '',
        sort: 0,
        status: 1,
      };
    }
    showEditModal.value = true;
  };

  const handleSave = async () => {
    try {
      await formRef.value?.validate();
    } catch (error) {
      return;
    }

    saveLoading.value = true;
    try {
      await ToogoUserApi.vipLevelEdit(editForm.value);
      message.success('保存成功');
      showEditModal.value = false;
      loadData();
    } catch (error: any) {
      message.error(error.message || '保存失败');
    } finally {
      saveLoading.value = false;
    }
  };

  onMounted(() => {
    loadData();
  });
</script>

<style scoped lang="less">
  .vip-level-admin-page {
    padding: 16px;
  }
</style>
