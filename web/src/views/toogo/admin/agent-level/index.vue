<template>
  <div class="agent-level-admin-page">
    <n-card title="代理商等级配置">
      <template #header-extra>
        <n-button type="primary" @click="openEditModal()">新增等级</n-button>
      </template>

      <n-data-table :columns="columns" :data="list" :loading="loading" :pagination="false" />
    </n-card>

    <!-- 编辑弹窗 -->
    <n-modal
      v-model:show="showEditModal"
      :title="editForm.id ? '编辑代理商等级' : '新增代理商等级'"
      preset="card"
      style="width: 700px"
    >
      <n-form
        ref="formRef"
        :model="editForm"
        :rules="rules"
        label-placement="left"
        label-width="140"
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
              <n-input v-model:value="editForm.levelName" placeholder="如：初级代理" />
            </n-form-item>
          </n-gi>
          <n-gi :span="2">
            <n-divider>升级条件</n-divider>
          </n-gi>
          <n-gi>
            <n-form-item label="团队人数要求">
              <n-input-number
                v-model:value="editForm.requireTeamCount"
                :min="0"
                style="width: 100%"
              />
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="团队订阅金额">
              <n-input-number
                v-model:value="editForm.requireTeamSubscribe"
                :min="0"
                :precision="2"
                style="width: 100%"
              />
            </n-form-item>
          </n-gi>
          <n-gi :span="2">
            <n-divider>订阅佣金比例 (%)</n-divider>
          </n-gi>
          <n-gi>
            <n-form-item label="一级订阅佣金">
              <n-input-number
                v-model:value="editForm.subscribeRate1"
                :min="0"
                :max="1"
                :precision="2"
                :step="0.01"
                style="width: 100%"
              >
                <template #suffix>{{ (editForm.subscribeRate1 * 100).toFixed(0) }}%</template>
              </n-input-number>
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="二级订阅佣金">
              <n-input-number
                v-model:value="editForm.subscribeRate2"
                :min="0"
                :max="1"
                :precision="2"
                :step="0.01"
                style="width: 100%"
              >
                <template #suffix>{{ (editForm.subscribeRate2 * 100).toFixed(0) }}%</template>
              </n-input-number>
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="三级订阅佣金">
              <n-input-number
                v-model:value="editForm.subscribeRate3"
                :min="0"
                :max="1"
                :precision="2"
                :step="0.01"
                style="width: 100%"
              >
                <template #suffix>{{ (editForm.subscribeRate3 * 100).toFixed(0) }}%</template>
              </n-input-number>
            </n-form-item>
          </n-gi>
          <n-gi :span="2">
            <n-divider>算力消耗佣金比例 (%)</n-divider>
          </n-gi>
          <n-gi>
            <n-form-item label="一级算力佣金">
              <n-input-number
                v-model:value="editForm.powerRate1"
                :min="0"
                :max="1"
                :precision="2"
                :step="0.01"
                style="width: 100%"
              >
                <template #suffix>{{ (editForm.powerRate1 * 100).toFixed(0) }}%</template>
              </n-input-number>
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="二级算力佣金">
              <n-input-number
                v-model:value="editForm.powerRate2"
                :min="0"
                :max="1"
                :precision="2"
                :step="0.01"
                style="width: 100%"
              >
                <template #suffix>{{ (editForm.powerRate2 * 100).toFixed(0) }}%</template>
              </n-input-number>
            </n-form-item>
          </n-gi>
          <n-gi>
            <n-form-item label="三级算力佣金">
              <n-input-number
                v-model:value="editForm.powerRate3"
                :min="0"
                :max="1"
                :precision="2"
                :step="0.01"
                style="width: 100%"
              >
                <template #suffix>{{ (editForm.powerRate3 * 100).toFixed(0) }}%</template>
              </n-input-number>
            </n-form-item>
          </n-gi>
          <n-gi :span="2">
            <n-form-item label="等级描述">
              <n-input v-model:value="editForm.description" type="textarea" :rows="2" />
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
  import { ToogoCommissionApi } from '@/api/toogo';

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
    requireTeamCount: 0,
    requireTeamSubscribe: 0,
    subscribeRate1: 0.1,
    subscribeRate2: 0.05,
    subscribeRate3: 0.02,
    powerRate1: 0.05,
    powerRate2: 0.02,
    powerRate3: 0.01,
    description: '',
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
      render: (row: any) => h(NTag, { type: 'info' }, { default: () => `Lv.${row.level}` }),
    },
    { title: '等级名称', key: 'levelName' },
    { title: '团队人数', key: 'requireTeamCount' },
    { title: '团队订阅', key: 'requireTeamSubscribe' },
    {
      title: '一级订阅',
      key: 'subscribeRate1',
      render: (row: any) => `${(row.subscribeRate1 * 100).toFixed(0)}%`,
    },
    {
      title: '二级订阅',
      key: 'subscribeRate2',
      render: (row: any) => `${(row.subscribeRate2 * 100).toFixed(0)}%`,
    },
    {
      title: '三级订阅',
      key: 'subscribeRate3',
      render: (row: any) => `${(row.subscribeRate3 * 100).toFixed(0)}%`,
    },
    {
      title: '一级算力',
      key: 'powerRate1',
      render: (row: any) => `${(row.powerRate1 * 100).toFixed(0)}%`,
    },
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
      const res = await ToogoCommissionApi.agentLevelList({ page: 1, perPage: 100 });
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
        requireTeamCount: 0,
        requireTeamSubscribe: 0,
        subscribeRate1: 0.1,
        subscribeRate2: 0.05,
        subscribeRate3: 0.02,
        powerRate1: 0.05,
        powerRate2: 0.02,
        powerRate3: 0.01,
        description: '',
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
      await ToogoCommissionApi.agentLevelEdit(editForm.value);
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
  .agent-level-admin-page {
    padding: 16px;
  }
</style>
