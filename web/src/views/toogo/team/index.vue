<template>
  <div class="team-page">
    <!-- 统计 -->
    <n-grid cols="1 s:2 m:2 l:2 xl:2 2xl:2" :x-gap="12" :y-gap="12" responsive="screen">
      <n-gi>
        <n-card :bordered="false">
          <n-statistic label="推广人数" :value="teamStat.directCount || 0" />
        </n-card>
      </n-gi>
      <n-gi>
        <n-card :bordered="false">
          <n-statistic label="推广总人数" :value="teamStat.totalCount || 0" />
        </n-card>
      </n-gi>
    </n-grid>

    <!-- 自助申请代理 -->
    <n-card :bordered="false" class="proCard mt-3" title="代理申请">
      <template #header-extra>
        <n-space align="center" :size="8">
        <n-tag v-if="agentInfo?.agentStatus === 1" type="warning" size="small">审核中</n-tag>
        <n-tag v-else-if="agentInfo?.agentStatus === 2" type="success" size="small">已通过</n-tag>
        <n-tag v-else-if="agentInfo?.agentStatus === 3" type="error" size="small">已拒绝</n-tag>
          <n-tag v-if="agentInfo?.agentStatus === 2" :type="agentInfo?.agentUnlockLevel === 1 ? 'success' : 'default'" size="small">
            {{ agentInfo?.agentUnlockLevel === 1 ? '已解锁' : '未解锁' }}
          </n-tag>
        </n-space>
      </template>
      <n-space align="center" justify="space-between">
        <n-text depth="3">提交代理申请（需管理员审核）。申请订阅返佣比例将作为审批参考。</n-text>
        <n-button
          type="primary"
          :disabled="agentInfo?.agentStatus === 1 || agentInfo?.agentStatus === 2"
          @click="openSelfApplyModal"
        >
          我要申请代理
        </n-button>
      </n-space>
    </n-card>

    <!-- 推广成员 -->
    <n-card title="推广成员" :bordered="false" class="proCard mt-3">
      <template #header-extra>
          <n-input
            v-model:value="keyword"
            clearable
            placeholder="搜索用户名（如：dong）"
            style="width: 220px"
          />
      </template>

      <n-data-table
        :columns="columns"
        :data="filteredTeamList"
        :loading="loading"
        :pagination="pagination"
        striped
        :bordered="false"
        @update:page="handlePageChange"
      >
        <template #empty>
          <n-empty :description="keyword ? '未找到匹配的推广成员' : '暂无推广成员'" />
        </template>
      </n-data-table>
    </n-card>

    <!-- 我申请代理 -->
    <n-modal v-model:show="showSelfApplyModal" preset="card" style="width: 560px" title="我要申请代理">
      <n-space vertical :size="12">
        <n-form ref="selfApplyFormRef" :model="selfApplyForm" :rules="selfApplyRules" label-width="130">
          <n-form-item label="申请订阅返佣(%)" path="subscribeRate">
            <n-input-number v-model:value="selfApplyForm.subscribeRate" :min="0.01" :max="100" :precision="2" style="width: 100%" />
          </n-form-item>
          <n-form-item label="申请理由" path="remark">
            <n-input v-model:value="selfApplyForm.remark" type="textarea" :rows="3" placeholder="请输入申请理由" />
          </n-form-item>
        </n-form>
        <n-space justify="end">
          <n-button @click="showSelfApplyModal = false">取消</n-button>
          <n-button type="primary" :loading="selfApplyLoading" @click="submitSelfApply">提交申请</n-button>
        </n-space>
      </n-space>
    </n-modal>

    <!-- 帮下级申请代理 -->
    <n-modal v-model:show="showSubApplyModal" preset="card" style="width: 560px" title="帮成员申请代理">
      <n-space vertical :size="12">
        <n-text depth="3">申请订阅返佣比例不能超过您的比例：{{ agentInfo?.subscribeRate ?? 0 }}%</n-text>
        <n-form ref="subApplyFormRef" :model="subApplyForm" :rules="subApplyRules" label-width="130">
          <n-form-item label="目标用户" path="subUsername">
            <n-text strong>{{ subApplyForm.subUsername }}</n-text>
          </n-form-item>
          <n-form-item label="申请订阅返佣(%)" path="subscribeRate">
            <n-input-number
              v-model:value="subApplyForm.subscribeRate"
              :min="0.01"
              :max="agentInfo?.subscribeRate ?? 100"
              :precision="2"
              style="width: 100%"
            />
          </n-form-item>
          <n-form-item label="申请理由" path="remark">
            <n-input v-model:value="subApplyForm.remark" type="textarea" :rows="3" placeholder="请输入申请理由" />
          </n-form-item>
        </n-form>
        <n-space justify="end">
          <n-button @click="showSubApplyModal = false">取消</n-button>
          <n-button type="primary" :loading="subApplyLoading" @click="submitSubApply">提交申请</n-button>
        </n-space>
      </n-space>
    </n-modal>

    <!-- 设置下级代理佣金比例 -->
    <n-modal v-model:show="showSetRateModal" preset="card" style="width: 520px" title="设置下级代理佣金比例">
      <n-space vertical :size="12">
        <n-text depth="3">
          上级可设置范围：订阅 ≤ {{ agentInfo?.subscribeRate ?? 0 }}%
        </n-text>
        <n-form ref="setRateFormRef" :model="setRateForm" :rules="setRateRules" label-width="120">
          <n-form-item label="下级账号" path="subUsername">
            <n-text strong>{{ setRateForm.subUsername }}</n-text>
          </n-form-item>
          <n-form-item label="订阅佣金(%)" path="subscribeRate">
            <n-input-number
              v-model:value="setRateForm.subscribeRate"
              :min="0"
              :max="agentInfo?.subscribeRate ?? 100"
              :precision="2"
              style="width: 100%"
            />
          </n-form-item>
        </n-form>
        <n-space justify="end">
          <n-button @click="showSetRateModal = false">取消</n-button>
          <n-button type="primary" :loading="savingRate" @click="submitSetRate">保存</n-button>
        </n-space>
      </n-space>
    </n-modal>

    <!-- 佣金统计 -->
    <n-card title="佣金收入" :bordered="false" class="proCard mt-3">
      <n-grid cols="1 s:2 m:2 l:4 xl:4 2xl:4" :x-gap="12" :y-gap="12" responsive="screen">
        <n-gi>
          <n-statistic label="今日佣金" :value="commissionStat.todayCommission?.toFixed(4) || '0.0000'" suffix="USDT" />
        </n-gi>
        <n-gi>
          <n-statistic label="本周佣金" :value="commissionStat.weekCommission?.toFixed(4) || '0.0000'" suffix="USDT" />
        </n-gi>
        <n-gi>
          <n-statistic label="本月佣金" :value="commissionStat.monthCommission?.toFixed(4) || '0.0000'" suffix="USDT" />
        </n-gi>
        <n-gi>
          <n-statistic label="累计佣金" :value="commissionStat.totalCommission?.toFixed(4) || '0.0000'" suffix="USDT" />
        </n-gi>
      </n-grid>
      <n-divider />
      <n-button type="primary" @click="$router.push('/toogo/commission')">查看佣金明细</n-button>
    </n-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, h, computed } from 'vue';
import type { FormInst, FormRules } from 'naive-ui';
import { useMessage, NTag, NAvatar, NButton, NText, NSpace, NForm, NFormItem, NInputNumber, NModal } from 'naive-ui';
import { ToogoUserApi, ToogoCommissionApi } from '@/api/toogo';

const message = useMessage();

const teamStat = ref<any>({});
const teamList = ref<any[]>([]);
const commissionStat = ref<any>({});
const loading = ref(false);
const keyword = ref('');

const agentInfo = ref<any>(null);
const subAgentMap = ref<Map<number, any>>(new Map());

const showSetRateModal = ref(false);
const savingRate = ref(false);
const setRateFormRef = ref<FormInst | null>(null);
const setRateForm = ref({
  subUserId: 0,
  subUsername: '',
  subscribeRate: 0,
});

const setRateRules: FormRules = {
  subscribeRate: [
    { required: true, type: 'number', message: '请输入订阅佣金比例', trigger: ['input', 'blur'] },
  ],
};

// 自助申请代理
const showSelfApplyModal = ref(false);
const selfApplyLoading = ref(false);
const selfApplyFormRef = ref<FormInst | null>(null);
const selfApplyForm = ref({ subscribeRate: 10, remark: '' });
const selfApplyRules: FormRules = {
  subscribeRate: [{ required: true, type: 'number', message: '请输入申请订阅返佣比例', trigger: ['input', 'blur'] }],
  remark: [{ required: true, message: '请输入申请理由', trigger: ['input', 'blur'] }],
};

// 帮下级申请代理
const showSubApplyModal = ref(false);
const subApplyLoading = ref(false);
const subApplyFormRef = ref<FormInst | null>(null);
const subApplyForm = ref({ subUserId: 0, subUsername: '', subscribeRate: 0, remark: '' });
const subApplyRules: FormRules = {
  subscribeRate: [{ required: true, type: 'number', message: '请输入申请订阅返佣比例', trigger: ['input', 'blur'] }],
  remark: [{ required: true, message: '请输入申请理由', trigger: ['input', 'blur'] }],
};

const searchParams = ref({
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

const filteredTeamList = computed(() => {
  const k = (keyword.value || '').trim().toLowerCase();
  if (!k) return teamList.value;
  return (teamList.value || []).filter((x: any) => String(x?.username || '').toLowerCase().includes(k));
});

const columns = [
  {
    title: '用户',
    key: 'username',
    render: (row: any) => {
      return h('div', { style: { display: 'flex', alignItems: 'center', gap: '8px' } }, [
        h(NAvatar, { src: row.avatar, round: true, size: 32 }),
        h('span', {}, row.username),
      ]);
    },
  },
  {
    title: 'VIP等级',
    key: 'vipLevel',
    render: (row: any) => {
      return h(NTag, { type: row.vipLevel > 1 ? 'warning' : 'default', size: 'small' }, { default: () => `V${row.vipLevel}` });
    },
  },
  {
    title: '身份',
    key: 'identity',
    width: 80,
    render: (row: any) => {
      const isAgent = row.isAgent === 1 && row.agentStatus === 2;
      return h(NTag, { type: isAgent ? 'success' : 'default', size: 'small' }, { default: () => (isAgent ? '代理' : '用户') });
    },
  },
  {
    title: '当前佣金比例',
    key: 'subscribeRate',
    width: 120,
    render: (row: any) => {
      return row.subscribeRate ? `${row.subscribeRate}%` : '-';
    },
  },
  {
    title: '推广人数',
    key: 'inviteCount',
    width: 100,
    render: (row: any) => {
      return row.inviteCount || 0;
    },
  },
  {
    title: '当前订阅',
    key: 'currentPlanId',
    width: 120,
    render: (row: any) => {
      return row.currentPlanId ? `套餐${row.currentPlanId}` : '无';
    },
  },
  {
    title: '订阅到期时间',
    key: 'planExpireTime',
    width: 160,
    render: (row: any) => {
      if (!row.planExpireTime) return '-';
      const date = new Date(row.planExpireTime);
      return date.toLocaleString('zh-CN', { 
        year: 'numeric', 
        month: '2-digit', 
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit'
      });
    },
  },
  {
    title: '机器人数量',
    key: 'robotCount',
    width: 120,
    render: (row: any) => {
      const active = row.activeRobotCount || 0;
      const limit = row.robotLimit || 0;
      return `${active}/${limit}`;
    },
  },
  { title: '注册时间', key: 'registerTime' },
  {
    title: '操作',
    key: 'action',
    render: (row: any) => {
      const canSet = !!agentInfo.value?.canSetSubRate;
      const agentApproved = agentInfo.value?.isAgent === 1 && agentInfo.value?.agentStatus === 2;
      const isDirect = row.level === 1;
      const sub = subAgentMap.value.get(row.memberId);
      const isSubAgent = sub?.isAgent === 1;
      const subApprovedAgent = row.isAgent === 1 && row.agentStatus === 2;
      const subPending = row.agentStatus === 1;

      // 1) 已解锁且下级已是代理：允许设置比例
      if (canSet && isDirect && isSubAgent && subApprovedAgent) {
        return h(
          NButton,
          {
            size: 'small',
            type: 'primary',
            onClick: () => openSetRateModal(row),
          },
          { default: () => '设置比例' }
        );
      }

      // 2) 直属下级非代理：允许帮TA提交代理申请（仅已通过代理可用）
      if (agentApproved && isDirect && !subApprovedAgent) {
        if (subPending) {
          return h(NTag, { type: 'warning', size: 'small' }, { default: () => '已申请' });
        }
        return h(
          NButton,
          {
            size: 'small',
            type: 'info',
            onClick: () => openSubApplyModal(row),
          },
          { default: () => '帮TA申请代理' }
        );
      }

      return h(NText, { depth: 3 }, { default: () => '-' });
    },
  },
];

const buildSubAgentMap = (subAgents: any[] | undefined) => {
  const m = new Map<number, any>();
  (subAgents || []).forEach((x) => {
    if (x?.userId) m.set(Number(x.userId), x);
  });
  subAgentMap.value = m;
};

const loadData = async () => {
  try {
    const [statRes, commRes, agentRes] = await Promise.all([
      ToogoUserApi.teamStat(),
      ToogoCommissionApi.stat(),
      ToogoCommissionApi.getAgentInfo(),
    ]);
    teamStat.value = statRes || {};
    commissionStat.value = commRes || {};
    agentInfo.value = agentRes || null;
    buildSubAgentMap(agentRes?.subAgents);
  } catch (error) {
    console.error('加载数据失败:', error);
  }
};

const loadTeamList = async () => {
  loading.value = true;
  try {
    const res = await ToogoUserApi.teamList(searchParams.value);
    teamList.value = res?.list || [];
    pagination.value.itemCount = res?.totalCount || 0;
  } catch (error) {
    console.error('加载团队列表失败:', error);
  } finally {
    loading.value = false;
  }
};

const handlePageChange = (page: number) => {
  searchParams.value.page = page;
  loadTeamList();
};

const openSetRateModal = (row: any) => {
  const sub = subAgentMap.value.get(row.memberId);
  setRateForm.value.subUserId = row.memberId;
  setRateForm.value.subUsername = row.username || String(row.memberId);
  setRateForm.value.subscribeRate = Number(sub?.subscribeRate ?? 0);
  showSetRateModal.value = true;
};

const openSelfApplyModal = () => {
  selfApplyForm.value.subscribeRate = 10;
  selfApplyForm.value.remark = '';
  showSelfApplyModal.value = true;
};

const submitSelfApply = async () => {
  try {
    await selfApplyFormRef.value?.validate?.();
  } catch {
    return;
  }
  selfApplyLoading.value = true;
  try {
    await ToogoCommissionApi.applyAgent({
      remark: selfApplyForm.value.remark,
      subscribeRate: Number(selfApplyForm.value.subscribeRate),
    });
    message.success('申请已提交，请等待管理员审核');
    showSelfApplyModal.value = false;
    await loadData();
    await loadTeamList();
  } catch (e: any) {
    message.error(e?.message || '提交失败');
  } finally {
    selfApplyLoading.value = false;
  }
};

const openSubApplyModal = (row: any) => {
  subApplyForm.value.subUserId = row.memberId;
  subApplyForm.value.subUsername = row.username || String(row.memberId);
  subApplyForm.value.subscribeRate = Math.min(Number(agentInfo.value?.subscribeRate ?? 0), 10) || 10;
  subApplyForm.value.remark = '';
  showSubApplyModal.value = true;
};

const submitSubApply = async () => {
  if (!(agentInfo.value?.isAgent === 1 && agentInfo.value?.agentStatus === 2)) {
    message.error('您尚未成为代理，无法代下级申请');
    return;
  }
  const maxSub = Number(agentInfo.value?.subscribeRate ?? 0);
  if (Number(subApplyForm.value.subscribeRate) > maxSub) {
    message.error(`申请订阅返佣比例不能超过您的比例：${maxSub}%`);
    return;
  }
  try {
    await subApplyFormRef.value?.validate?.();
  } catch {
    return;
  }
  subApplyLoading.value = true;
  try {
    await ToogoCommissionApi.applyAgentForSub({
      subUserId: subApplyForm.value.subUserId,
      remark: subApplyForm.value.remark,
      subscribeRate: Number(subApplyForm.value.subscribeRate),
    });
    message.success('已提交下级代理申请');
    showSubApplyModal.value = false;
    await loadTeamList();
  } catch (e: any) {
    message.error(e?.message || '提交失败');
  } finally {
    subApplyLoading.value = false;
  }
};

const submitSetRate = async () => {
  if (!agentInfo.value?.canSetSubRate) {
    message.error('您尚未解锁无限层级佣金，无法设置下级比例');
    return;
  }

  const maxSub = Number(agentInfo.value?.subscribeRate ?? 0);
  if (Number(setRateForm.value.subscribeRate) > maxSub) {
    message.error(`订阅佣金比例不能超过上级：${maxSub}%`);
    return;
  }

  try {
    const ok = await setRateFormRef.value?.validate?.();
    if (ok === false) return;
  } catch {
    return;
  }

  savingRate.value = true;
  try {
    await ToogoCommissionApi.setSubAgentRate({
      subUserId: setRateForm.value.subUserId,
      subscribeRate: Number(setRateForm.value.subscribeRate),
    });
    message.success('已保存下级佣金比例');
    showSetRateModal.value = false;
    // 重新拉取代理信息（刷新下级比例）
    const agentRes = await ToogoCommissionApi.getAgentInfo();
    agentInfo.value = agentRes || null;
    buildSubAgentMap(agentRes?.subAgents);
  } catch (e: any) {
    message.error(e?.message || '保存失败');
  } finally {
    savingRate.value = false;
  }
};

onMounted(() => {
  loadData();
  loadTeamList();
});
</script>

<style scoped lang="less">
.team-page {
  padding: 16px;
}
</style>

