<template>
  <div>
    <n-card :bordered="false" class="proCard">
      <n-space vertical :size="16">
        <!-- 顶部操作栏 -->
        <n-space>
          <n-button type="primary" @click="handleCreate">
            <template #icon>
              <n-icon><PlusOutlined /></n-icon>
            </template>
            创建机器人
          </n-button>
          <n-button @click="handleRefresh">
            <template #icon>
              <n-icon><ReloadOutlined /></n-icon>
            </template>
            刷新
          </n-button>
        </n-space>

        <!-- 机器人卡片列表 -->
        <n-grid :cols="4" :x-gap="12" :y-gap="12">
          <n-gi v-for="robot in robotList" :key="robot.id">
            <n-card hoverable>
              <template #header>
                <n-space align="center" justify="space-between">
                  <n-text strong>{{ robot.name }}</n-text>
                  <n-tag :type="getStatusType(robot.status)">
                    {{ getStatusText(robot.status) }}
                  </n-tag>
                </n-space>
              </template>

              <n-space vertical :size="8">
                <n-descriptions :column="1" size="small">
                  <n-descriptions-item label="交易所">
                    {{ robot.apiConfigName }}
                  </n-descriptions-item>
                  <n-descriptions-item label="交易对">
                    {{ robot.symbol }}
                  </n-descriptions-item>
                  <n-descriptions-item label="风险偏好">
                    {{ robot.riskPreference }}
                  </n-descriptions-item>
                  <n-descriptions-item label="总盈亏">
                    <n-text :type="robot.totalPnl >= 0 ? 'success' : 'error'" strong>
                      {{ robot.totalPnl }} USDT
                    </n-text>
                  </n-descriptions-item>
                  <n-descriptions-item label="持仓">
                    多: {{ robot.longCount }} / 空: {{ robot.shortCount }}
                  </n-descriptions-item>
                </n-descriptions>
              </n-space>

              <template #footer>
                <n-space justify="space-between">
                  <n-space>
                    <n-button
                      v-if="robot.status === 1"
                      size="small"
                      type="success"
                      @click="handleStart(robot)"
                    >
                      启动
                    </n-button>
                    <n-button
                      v-if="robot.status === 2"
                      size="small"
                      type="warning"
                      @click="handlePause(robot)"
                    >
                      暂停
                    </n-button>
                    <n-button
                      v-if="robot.status === 3"
                      size="small"
                      type="success"
                      @click="handleStart(robot)"
                    >
                      继续
                    </n-button>
                    <n-button size="small" @click="handleView(robot)">
                      详情
                    </n-button>
                  </n-space>
                  <n-button size="small" type="error" @click="handleDelete(robot)">
                    删除
                  </n-button>
                </n-space>
              </template>
            </n-card>
          </n-gi>
        </n-grid>

        <!-- 分页 -->
        <n-pagination
          v-model:page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :page-count="pagination.pageCount"
          show-size-picker
          :page-sizes="[12, 24, 48]"
          @update:page="loadData"
          @update:page-size="loadData"
        />
      </n-space>
    </n-card>
  </div>
</template>

<script lang="ts" setup>
import { ref, reactive, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import {
  NCard,
  NSpace,
  NButton,
  NIcon,
  NGrid,
  NGi,
  NTag,
  NText,
  NDescriptions,
  NDescriptionsItem,
  NPagination,
  useMessage,
  useDialog,
} from 'naive-ui';
import { PlusOutlined, ReloadOutlined } from '@vicons/antd';
import {
  getRobotList,
  startRobot,
  pauseRobot,
  deleteRobot,
} from '@/api/trading/robot';

const router = useRouter();
const message = useMessage();
const dialog = useDialog();

const robotList = ref([]);
const loading = ref(false);
const pagination = reactive({
  page: 1,
  pageSize: 12,
  pageCount: 1,
});

// 加载数据
const loadData = async () => {
  loading.value = true;
  try {
    const res = await getRobotList({
      page: pagination.page,
      pageSize: pagination.pageSize,
    });
    robotList.value = res.list || [];
    pagination.pageCount = Math.ceil((res.total || 0) / pagination.pageSize);
  } catch (error: any) {
    message.error(error.message || '加载失败');
  } finally {
    loading.value = false;
  }
};

// 创建机器人
const handleCreate = () => {
  router.push('/trading/robot/create');
};

// 查看详情
const handleView = (robot: any) => {
  router.push(`/trading/robot/detail/${robot.id}`);
};

// 启动机器人
const handleStart = async (robot: any) => {
  try {
    await startRobot({ id: robot.id });
    message.success('启动成功');
    loadData();
  } catch (error: any) {
    message.error(error.message || '启动失败');
  }
};

// 暂停机器人
const handlePause = async (robot: any) => {
  try {
    await pauseRobot({ id: robot.id });
    message.success('已暂停');
    loadData();
  } catch (error: any) {
    message.error(error.message || '暂停失败');
  }
};

// 删除机器人
const handleDelete = (robot: any) => {
  dialog.warning({
    title: '确认删除',
    content: `确定要删除机器人"${robot.name}"吗？`,
    positiveText: '确定',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await deleteRobot({ id: robot.id });
        message.success('删除成功');
        loadData();
      } catch (error: any) {
        message.error(error.message || '删除失败');
      }
    },
  });
};

// 刷新
const handleRefresh = () => {
  loadData();
};

// 获取状态类型
const getStatusType = (status: number) => {
  const types: any = {
    1: 'default',
    2: 'success',
    3: 'warning',
    4: 'error',
  };
  return types[status] || 'default';
};

// 获取状态文本
const getStatusText = (status: number) => {
  const texts: any = {
    1: '未启动',
    2: '运行中',
    3: '已暂停',
    4: '已停用',
  };
  return texts[status] || '未知';
};

onMounted(() => {
  loadData();
});
</script>

<style scoped>
.proCard {
  min-height: calc(100vh - 200px);
}
</style>

