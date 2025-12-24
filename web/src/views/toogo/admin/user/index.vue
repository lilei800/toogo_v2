<template>
  <div class="user-admin-page">
    <n-card title="Toogo用户管理">
      <template #header-extra>
        <n-space>
          <n-select
            v-if="activeTab === 'users'"
            v-model:value="searchParams.vipLevel"
            :options="vipOptions"
            style="width: 100px"
            clearable
            placeholder="VIP等级"
            @update:value="loadData"
          />
          <n-select
            v-if="activeTab === 'users'"
            v-model:value="searchParams.isAgent"
            :options="agentOptions"
            style="width: 100px"
            clearable
            placeholder="代理商"
            @update:value="loadData"
          />
          <n-select
            v-if="activeTab === 'users'"
            v-model:value="searchParams.agentStatus"
            :options="agentStatusOptions"
            style="width: 120px"
            clearable
            placeholder="代理商状态"
            @update:value="loadData"
          />
          <n-select
            v-else
            v-model:value="searchParams.vipLevel"
            :options="vipOptions"
            style="width: 100px"
            clearable
            placeholder="VIP等级"
            @update:value="loadData"
          />
          <n-button @click="loadData">刷新</n-button>
        </n-space>
      </template>

      <n-tabs v-model:value="activeTab" type="line" animated @update:value="handleTabChange">
        <n-tab-pane name="users" tab="用户列表">
          <n-data-table
            :columns="columnsUsers"
            :data="list"
            :loading="loading"
            :pagination="pagination"
            @update:page="handlePageChange"
          />
        </n-tab-pane>
        <n-tab-pane name="agent_apply" tab="代理申请（待审批）">
          <n-data-table
            :columns="columnsAgentApply"
            :data="list"
            :loading="loading"
            :pagination="pagination"
            @update:page="handlePageChange"
          />
        </n-tab-pane>
      </n-tabs>
    </n-card>

    <!-- 详情弹窗 -->
    <n-modal v-model:show="showDetailModal" title="用户详情" preset="card" style="width: 900px">
      <n-descriptions :column="2" label-placement="left" bordered v-if="currentRow">
        <n-descriptions-item label="用户名">{{ currentRow.username }}</n-descriptions-item>
        <n-descriptions-item label="VIP等级">V{{ currentRow.vipLevel }}</n-descriptions-item>
        <n-descriptions-item label="是否代理商">
          <n-tag :type="currentRow.isAgent ? 'success' : 'default'" size="small">
            {{ currentRow.isAgent ? '是' : '否' }}
          </n-tag>
        </n-descriptions-item>
        <n-descriptions-item label="代理商状态">
          <n-tag :type="getAgentStatusType(currentRow.agentStatus)" size="small">
            {{ getAgentStatusText(currentRow.agentStatus) }}
          </n-tag>
        </n-descriptions-item>
        <n-descriptions-item label="层级解锁">
          <n-tag :type="currentRow.agentUnlockLevel === 1 ? 'success' : 'warning'" size="small">
            {{ currentRow.agentUnlockLevel === 1 ? '无限级佣金' : '仅一级佣金' }}
          </n-tag>
        </n-descriptions-item>
        <n-descriptions-item label="订阅返佣">{{ currentRow.subscribeRate || 0 }}%</n-descriptions-item>
        <n-descriptions-item label="余额账户">
          <n-space align="center">
            <n-text type="success">{{ (currentRow.balance || 0).toFixed(2) }} USDT</n-text>
          </n-space>
        </n-descriptions-item>
        <n-descriptions-item label="积分账户">
          <n-space align="center">
            <n-text type="warning">{{ (currentRow.giftPower || 0).toFixed(2) }} 积分</n-text>
          </n-space>
        </n-descriptions-item>
        <n-descriptions-item label="邀请码">{{ currentRow.inviteCode }}</n-descriptions-item>
        <n-descriptions-item label="邀请人ID">{{ currentRow.inviterId || '-' }}</n-descriptions-item>
        <n-descriptions-item label="直推人数">{{ currentRow.inviteCount }}</n-descriptions-item>
        <n-descriptions-item label="团队人数">{{ currentRow.teamCount }}</n-descriptions-item>
        <n-descriptions-item label="消耗算力">{{ currentRow.totalConsumePower }}</n-descriptions-item>
        <n-descriptions-item label="团队消耗">{{ currentRow.teamConsumePower }}</n-descriptions-item>
        <n-descriptions-item label="当前套餐">{{ currentRow.currentPlanId || '免费版' }}</n-descriptions-item>
        <n-descriptions-item label="套餐到期">{{ currentRow.planExpireTime || '-' }}</n-descriptions-item>
        <n-descriptions-item label="机器人限制">{{ currentRow.robotLimit }}</n-descriptions-item>
        <n-descriptions-item label="运行中机器人">{{ currentRow.activeRobotCount }}</n-descriptions-item>
        <n-descriptions-item label="算力折扣">{{ currentRow.powerDiscount }}%</n-descriptions-item>
        <n-descriptions-item label="代理商申请时间">{{ currentRow.agentApplyAt || '-' }}</n-descriptions-item>
        <n-descriptions-item label="创建时间">{{ currentRow.createdAt }}</n-descriptions-item>
      </n-descriptions>

      <!-- 代理商申请备注 -->
      <template v-if="currentRow.agentApplyRemark">
        <n-divider>申请备注</n-divider>
        <n-text>{{ currentRow.agentApplyRemark }}</n-text>
      </template>

      <n-divider />

      <n-space>
        <n-button type="primary" @click="editVipLevel">调整VIP等级</n-button>
        <n-button type="info" @click="openAgentModal">管理代理商</n-button>
        <n-button 
          v-if="currentRow.agentStatus === 1" 
          type="success" 
          @click="openApproveModal(true)"
        >
          审批通过
        </n-button>
        <n-button 
          v-if="currentRow.agentStatus === 1" 
          type="error" 
          @click="openApproveModal(false)"
        >
          拒绝申请
        </n-button>
        <n-button type="warning" @click="editRobotLimit">调整机器人限制</n-button>
        <n-button type="success" @click="openRechargeBalanceModal">充值余额</n-button>
        <n-button type="warning" @click="openRechargePointsModal">充值积分</n-button>
        <n-button type="info" @click="openRechargePowerModal">充值算力</n-button>
      </n-space>
    </n-modal>

    <!-- 调整VIP等级弹窗 -->
    <n-modal v-model:show="showVipModal" title="调整VIP等级" preset="dialog">
      <n-form-item label="VIP等级">
        <n-select v-model:value="editVipForm.vipLevel" :options="allVipOptions" />
      </n-form-item>
      <template #action>
        <n-button type="primary" @click="saveVipLevel" :loading="saveLoading">保存</n-button>
      </template>
    </n-modal>

    <!-- 管理代理商弹窗 -->
    <n-modal v-model:show="showAgentModal" title="管理代理商" preset="dialog" style="width: 500px">
      <n-form :model="editAgentForm" label-placement="left" label-width="100">
        <n-form-item label="是否代理商">
          <n-switch v-model:value="editAgentForm.isAgent" />
        </n-form-item>
        <n-form-item label="代理商状态" v-if="editAgentForm.isAgent">
          <n-select v-model:value="editAgentForm.agentStatus" :options="agentStatusOptionsEdit" />
        </n-form-item>
        <n-form-item label="层级解锁" v-if="editAgentForm.isAgent">
          <n-select v-model:value="editAgentForm.agentUnlockLevel" :options="unlockLevelOptions" />
        </n-form-item>
        <n-form-item label="订阅返佣比例" v-if="editAgentForm.isAgent">
          <n-input-number 
            v-model:value="editAgentForm.subscribeRate" 
            :min="0" 
            :max="100" 
            :precision="2"
            style="width: 100%"
          >
            <template #suffix>%</template>
          </n-input-number>
        </n-form-item>
      </n-form>
      <template #action>
        <n-space>
          <n-button @click="showAgentModal = false">取消</n-button>
          <n-button type="primary" @click="saveAgent" :loading="saveLoading">保存</n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- 审批代理商弹窗 -->
    <n-modal v-model:show="showApproveModal" :title="approveForm.approved ? '审批通过' : '拒绝申请'" preset="dialog" style="width: 500px">
      <n-form :model="approveForm" label-placement="left" label-width="100">
        <template v-if="approveForm.approved">
          <n-form-item label="订阅返佣比例">
            <n-input-number 
              v-model:value="approveForm.subscribeRate" 
              :min="0" 
              :max="100" 
              :precision="2"
              style="width: 100%"
            >
              <template #suffix>%</template>
            </n-input-number>
          </n-form-item>
          <n-form-item label="层级解锁">
            <n-select v-model:value="approveForm.agentUnlockLevel" :options="unlockLevelOptions" />
          </n-form-item>
        </template>
        <template v-else>
          <n-form-item label="拒绝原因">
            <n-input 
              v-model:value="approveForm.rejectReason" 
              type="textarea" 
              placeholder="请输入拒绝原因"
              :rows="3"
            />
          </n-form-item>
        </template>
      </n-form>
      <template #action>
        <n-space>
          <n-button @click="showApproveModal = false">取消</n-button>
          <n-button :type="approveForm.approved ? 'success' : 'error'" @click="submitApprove" :loading="approveLoading">
            {{ approveForm.approved ? '确认通过' : '确认拒绝' }}
          </n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- 充值算力弹窗 -->
    <n-modal v-model:show="showRechargePowerModal" title="手动充值算力" preset="dialog" style="width: 450px">
      <n-form ref="rechargePowerFormRef" :model="rechargePowerForm" :rules="rechargePowerRules" label-placement="left" label-width="80">
        <n-form-item label="目标用户">
          <n-text>{{ currentRow?.username }} (会员ID: {{ currentRow?.memberId }})</n-text>
        </n-form-item>
        <n-form-item label="当前算力">
          <n-text type="info">{{ currentRow?.power || 0 }}</n-text>
        </n-form-item>
        <n-form-item label="充值数量" path="amount">
          <n-input-number
            v-model:value="rechargePowerForm.amount"
            :min="0.01"
            :precision="2"
            placeholder="请输入充值算力数量"
            style="width: 100%"
          />
        </n-form-item>
        <n-form-item label="备注" path="remark">
          <n-input
            v-model:value="rechargePowerForm.remark"
            type="textarea"
            placeholder="请输入充值备注（可选）"
            :rows="2"
          />
        </n-form-item>
      </n-form>
      <template #action>
        <n-space>
          <n-button @click="showRechargePowerModal = false">取消</n-button>
          <n-button type="primary" @click="submitRechargePower" :loading="rechargeLoading">确认充值</n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- 充值余额弹窗 -->
    <n-modal v-model:show="showRechargeBalanceModal" title="手动充值余额" preset="dialog" style="width: 450px">
      <n-form ref="rechargeBalanceFormRef" :model="rechargeBalanceForm" :rules="rechargeRules" label-placement="left" label-width="80">
        <n-form-item label="目标用户">
          <n-text>{{ currentRow?.username }} (会员ID: {{ currentRow?.memberId }})</n-text>
        </n-form-item>
        <n-form-item label="当前余额">
          <n-text type="success">{{ (currentRow?.balance || 0).toFixed(2) }} USDT</n-text>
        </n-form-item>
        <n-form-item label="充值金额" path="amount">
          <n-input-number
            v-model:value="rechargeBalanceForm.amount"
            :min="0.01"
            :precision="2"
            placeholder="请输入充值余额金额"
            style="width: 100%"
          >
            <template #suffix>USDT</template>
          </n-input-number>
        </n-form-item>
        <n-form-item label="备注" path="remark">
          <n-input
            v-model:value="rechargeBalanceForm.remark"
            type="textarea"
            placeholder="请输入充值备注（可选）"
            :rows="2"
          />
        </n-form-item>
      </n-form>
      <template #action>
        <n-space>
          <n-button @click="showRechargeBalanceModal = false">取消</n-button>
          <n-button type="primary" @click="submitRechargeBalance" :loading="rechargeBalanceLoading">确认充值</n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- 充值积分弹窗 -->
    <n-modal v-model:show="showRechargePointsModal" title="手动充值积分" preset="dialog" style="width: 450px">
      <n-form ref="rechargePointsFormRef" :model="rechargePointsForm" :rules="rechargeRules" label-placement="left" label-width="80">
        <n-form-item label="目标用户">
          <n-text>{{ currentRow?.username }} (会员ID: {{ currentRow?.memberId }})</n-text>
        </n-form-item>
        <n-form-item label="当前积分">
          <n-text type="warning">{{ (currentRow?.giftPower || 0).toFixed(2) }} 积分</n-text>
        </n-form-item>
        <n-form-item label="充值数量" path="amount">
          <n-input-number
            v-model:value="rechargePointsForm.amount"
            :min="0.01"
            :precision="2"
            placeholder="请输入充值积分数量"
            style="width: 100%"
          >
            <template #suffix>积分</template>
          </n-input-number>
        </n-form-item>
        <n-form-item label="备注" path="remark">
          <n-input
            v-model:value="rechargePointsForm.remark"
            type="textarea"
            placeholder="请输入充值备注（可选）"
            :rows="2"
          />
        </n-form-item>
      </n-form>
      <template #action>
        <n-space>
          <n-button @click="showRechargePointsModal = false">取消</n-button>
          <n-button type="primary" @click="submitRechargePoints" :loading="rechargePointsLoading">确认充值</n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, h } from 'vue';
import { useMessage } from 'naive-ui';
import { NButton, NTag, NSpace, NAvatar, NText } from 'naive-ui';
import { ToogoUserApi, ToogoCommissionApi } from '@/api/toogo';
import toogoLogo from '@/assets/images/logo.png';

const message = useMessage();

// 后台用户表里常见的“系统默认头像”（不希望在 Toogo 用户管理里展示，统一替换为 Toogo LOGO）
const DEFAULT_AVATAR_URLS = new Set<string>([
  'https://gmycos.facms.cn/hotgo/attachment/2023-02-09/cqdq8er9nfkchdopav.png',
]);

const resolveAvatarSrc = (avatar: any) => {
  const s = String(avatar || '').trim();
  if (!s) return toogoLogo;
  if (s === 'null' || s === 'undefined' || s === '0') return toogoLogo;
  if (DEFAULT_AVATAR_URLS.has(s)) return toogoLogo;
  return s;
};

const list = ref<any[]>([]);
const loading = ref(false);
const showDetailModal = ref(false);
const showVipModal = ref(false);
const showAgentModal = ref(false);
const showRechargePowerModal = ref(false);
const showRechargeBalanceModal = ref(false);
const showRechargePointsModal = ref(false);
const showApproveModal = ref(false);
const saveLoading = ref(false);
const rechargeLoading = ref(false);
const rechargeBalanceLoading = ref(false);
const rechargePointsLoading = ref(false);
const approveLoading = ref(false);
const currentRow = ref<any>(null);
const rechargePowerFormRef = ref<any>(null);
const rechargeBalanceFormRef = ref<any>(null);
const rechargePointsFormRef = ref<any>(null);

const activeTab = ref<'users' | 'agent_apply'>('users');

const searchParams = ref({
  vipLevel: null,
  isAgent: -1,
  agentStatus: null,
  page: 1,
  perPage: 10,
});

const pagination = ref({
  page: 1,
  pageSize: 10,
  itemCount: 0,
});

const editVipForm = ref({ vipLevel: 1 });
const editAgentForm = ref({ 
  isAgent: false, 
  agentStatus: 0,
  agentUnlockLevel: 0,
  subscribeRate: 0,
});
const approveForm = ref({
  approved: true,
  subscribeRate: 50,
  agentUnlockLevel: 0,
  rejectReason: '',
});
const rechargePowerForm = ref({ amount: 0, remark: '' });
const rechargeBalanceForm = ref({ amount: 0, remark: '' });
const rechargePointsForm = ref({ amount: 0, remark: '' });

const rechargePowerRules = {
  amount: {
    required: true,
    type: 'number',
    min: 0.01,
    message: '请输入有效的充值数量',
    trigger: ['blur', 'change'],
  },
};

const rechargeRules = {
  amount: {
    required: true,
    type: 'number',
    min: 0.01,
    message: '请输入有效的充值数量',
    trigger: ['blur', 'change'],
  },
};

const vipOptions = Array.from({ length: 10 }, (_, i) => ({ label: `V${i + 1}`, value: i + 1 }));
const allVipOptions = vipOptions;
const agentOptions = [
  { label: '全部', value: -1 },
  { label: '是', value: 1 },
  { label: '否', value: 0 },
];
const agentStatusOptions = [
  { label: '未申请', value: 0 },
  { label: '待审批', value: 1 },
  { label: '已通过', value: 2 },
  { label: '已拒绝', value: 3 },
];
const agentStatusOptionsEdit = [
  { label: '未申请', value: 0 },
  { label: '待审批', value: 1 },
  { label: '已通过', value: 2 },
  { label: '已拒绝', value: 3 },
];
const unlockLevelOptions = [
  { label: '仅一级佣金', value: 0 },
  { label: '无限级佣金', value: 1 },
];
const agentLevelOptions = ref<any[]>([]);

// 获取代理商状态文字
const getAgentStatusText = (status: number) => {
  const map: Record<number, string> = {
    0: '未申请',
    1: '待审批',
    2: '已通过',
    3: '已拒绝',
  };
  return map[status] || '未知';
};

// 获取代理商状态类型（用于标签颜色）
const getAgentStatusType = (status: number) => {
  const map: Record<number, string> = {
    0: 'default',
    1: 'warning',
    2: 'success',
    3: 'error',
  };
  return map[status] || 'default';
};

const columnsUsers = [
  { title: 'ID', key: 'id', width: 60 },
  {
    title: '用户',
    key: 'username',
    render: (row: any) => h('div', { style: { display: 'flex', alignItems: 'center', gap: '8px' } }, [
      h(NAvatar, { src: resolveAvatarSrc(row.avatar), fallbackSrc: toogoLogo, round: true, size: 32 }),
      h('span', {}, row.username),
    ]),
  },
  {
    title: 'VIP等级',
    key: 'vipLevel',
    render: (row: any) => h(NTag, { type: row.vipLevel > 5 ? 'warning' : 'default', size: 'small' }, { default: () => `V${row.vipLevel}` }),
  },
  {
    title: '代理商',
    key: 'isAgent',
    render: (row: any) => h(NTag, { type: row.isAgent ? 'success' : 'default', size: 'small' }, { default: () => row.isAgent ? '是' : '否' }),
  },
  {
    title: '代理商状态',
    key: 'agentStatus',
    width: 100,
    render: (row: any) => h(NTag, { type: getAgentStatusType(row.agentStatus), size: 'small' }, { default: () => getAgentStatusText(row.agentStatus) }),
  },
  {
    title: '佣金比例',
    key: 'commissionRate',
    width: 120,
    render: (row: any) => (row.isAgent ? `${row.subscribeRate || 0}%` : '-'),
  },
  { title: '直推', key: 'inviteCount', width: 60 },
  { title: '团队', key: 'teamCount', width: 60 },
  { title: '消耗算力', key: 'totalConsumePower', width: 80, render: (row: any) => (row.totalConsumePower || 0).toFixed(2) },
  { title: '机器人', key: 'robotLimit', width: 80, render: (row: any) => `${row.activeRobotCount}/${row.robotLimit}` },
  {
    title: '状态',
    key: 'status',
    width: 60,
    render: (row: any) => h(NTag, { type: row.status === 1 ? 'success' : 'error', size: 'small' }, { default: () => row.status === 1 ? '正常' : '禁用' }),
  },
  {
    title: '操作',
    key: 'actions',
    width: 150,
    render: (row: any) => h(NSpace, { size: 'small' }, { 
      default: () => [
        h(NButton, { size: 'small', onClick: () => viewDetail(row) }, { default: () => '详情' }),
        row.agentStatus === 1 ? h(NButton, { size: 'small', type: 'success', onClick: () => { currentRow.value = row; openApproveModal(true); } }, { default: () => '审批' }) : null,
      ].filter(Boolean)
    }),
  },
];

const columnsAgentApply = [
  { title: 'ID', key: 'id', width: 60 },
  {
    title: '用户',
    key: 'username',
    render: (row: any) => h('div', { style: { display: 'flex', alignItems: 'center', gap: '8px' } }, [
      h(NAvatar, { src: resolveAvatarSrc(row.avatar), fallbackSrc: toogoLogo, round: true, size: 32 }),
      h('span', {}, row.username),
    ]),
  },
  {
    title: '当前佣金比例',
    key: 'subscribeRate',
    width: 120,
    render: (row: any) => row.subscribeRate ? `${row.subscribeRate}%` : '-',
  },
  {
    title: '当前身份',
    key: 'identity',
    width: 100,
    render: (row: any) => {
      const isAgent = row.isAgent === 1 && row.agentStatus === 2;
      return h(NTag, { type: isAgent ? 'success' : 'default', size: 'small' }, { default: () => (isAgent ? '代理' : '用户') });
    },
  },
  {
    title: '申请时间',
    key: 'agentApplyAt',
    width: 160,
    render: (row: any) => row.agentApplyAt || '-',
  },
  {
    title: '申请备注',
    key: 'agentApplyRemark',
    render: (row: any) => h(NText, { depth: 3 }, { default: () => row.agentApplyRemark || '-' }),
  },
  {
    title: '操作',
    key: 'actions',
    width: 220,
    render: (row: any) => h(NSpace, { size: 'small' }, {
      default: () => [
        h(NButton, { size: 'small', onClick: () => viewDetail(row) }, { default: () => '详情' }),
        h(NButton, { size: 'small', type: 'success', onClick: () => { currentRow.value = row; openApproveModal(true); } }, { default: () => '通过' }),
        h(NButton, { size: 'small', type: 'error', onClick: () => { currentRow.value = row; openApproveModal(false); } }, { default: () => '拒绝' }),
      ],
    }),
  },
];

const handleTabChange = (tab: 'users' | 'agent_apply') => {
  // 切换 Tab 时重置分页与筛选口径：
  pagination.value.page = 1;
  searchParams.value.page = 1;

  if (tab === 'agent_apply') {
    // 代理申请：固定只看待审批
    searchParams.value.agentStatus = 1;
    searchParams.value.isAgent = -1;
  } else {
    // 用户列表：恢复为可筛选（默认不过滤）
    searchParams.value.agentStatus = null;
  }
  loadData();
};

const loadData = async () => {
  loading.value = true;
  try {
    const res = await ToogoUserApi.list(searchParams.value);
    list.value = res?.list || [];
    pagination.value.itemCount = res?.totalCount || 0;
  } catch (error) {
    console.error('加载失败:', error);
  } finally {
    loading.value = false;
  }
};

const loadAgentLevels = async () => {
  try {
    const res = await ToogoCommissionApi.agentLevelList({ status: 1 });
    agentLevelOptions.value = (res?.list || []).map((item: any) => ({
      label: item.levelName,
      value: item.level,
    }));
  } catch (error) {
    console.error('加载代理商等级失败:', error);
  }
};

const handlePageChange = (page: number) => {
  searchParams.value.page = page;
  loadData();
};

const viewDetail = (row: any) => {
  currentRow.value = row;
  showDetailModal.value = true;
};

const editVipLevel = () => {
  editVipForm.value.vipLevel = currentRow.value.vipLevel;
  showVipModal.value = true;
};

// 打开代理商管理弹窗
const openAgentModal = () => {
  editAgentForm.value = {
    isAgent: currentRow.value.isAgent === 1,
    agentStatus: currentRow.value.agentStatus || 0,
    agentUnlockLevel: currentRow.value.agentUnlockLevel || 0,
    subscribeRate: currentRow.value.subscribeRate || 0,
  };
  showAgentModal.value = true;
};

// 打开审批弹窗
const openApproveModal = (approved: boolean) => {
  approveForm.value = {
    approved,
    subscribeRate: 50,
    agentUnlockLevel: 0,
    rejectReason: '',
  };
  showApproveModal.value = true;
};

const editRobotLimit = () => {
  message.info('调整机器人限制功能开发中');
};

const saveVipLevel = async () => {
  message.info('保存VIP等级功能开发中');
  showVipModal.value = false;
};

// 保存代理商设置
const saveAgent = async () => {
  if (!currentRow.value?.memberId) {
    message.error('未选择用户');
    return;
  }

  saveLoading.value = true;
  try {
    await ToogoCommissionApi.updateAgent({
      memberId: currentRow.value.memberId,
      isAgent: editAgentForm.value.isAgent ? 1 : 0,
      agentStatus: editAgentForm.value.agentStatus,
      agentUnlockLevel: editAgentForm.value.agentUnlockLevel,
      subscribeRate: editAgentForm.value.subscribeRate,
    });
    message.success('保存成功');
    showAgentModal.value = false;
    showDetailModal.value = false;
    loadData();
  } catch (error: any) {
    message.error(error?.message || '保存失败');
  } finally {
    saveLoading.value = false;
  }
};

// 提交审批
const submitApprove = async () => {
  if (!currentRow.value?.memberId) {
    message.error('未选择用户');
    return;
  }

  if (approveForm.value.approved) {
    if (approveForm.value.subscribeRate <= 0) {
      message.error('请设置佣金比例');
      return;
    }
  }

  approveLoading.value = true;
  try {
    await ToogoCommissionApi.approveAgent({
      memberId: currentRow.value.memberId,
      approved: approveForm.value.approved,
      subscribeRate: approveForm.value.subscribeRate,
      agentUnlockLevel: approveForm.value.agentUnlockLevel,
      rejectReason: approveForm.value.rejectReason,
    });
    message.success(approveForm.value.approved ? '审批通过' : '已拒绝申请');
    showApproveModal.value = false;
    showDetailModal.value = false;
    loadData();
  } catch (error: any) {
    message.error(error?.message || '操作失败');
  } finally {
    approveLoading.value = false;
  }
};

// 打开充值算力弹窗
const openRechargePowerModal = () => {
  rechargePowerForm.value = { amount: 0, remark: '' };
  showRechargePowerModal.value = true;
};

// 打开充值余额弹窗
const openRechargeBalanceModal = () => {
  rechargeBalanceForm.value = { amount: 0, remark: '' };
  showRechargeBalanceModal.value = true;
};

// 打开充值积分弹窗
const openRechargePointsModal = () => {
  rechargePointsForm.value = { amount: 0, remark: '' };
  showRechargePointsModal.value = true;
};

// 提交充值算力
const submitRechargePower = async () => {
  try {
    await rechargePowerFormRef.value?.validate();
  } catch (error) {
    return;
  }

  if (!currentRow.value?.memberId) {
    message.error('未选择用户');
    return;
  }

  rechargeLoading.value = true;
  try {
    const res = await ToogoUserApi.adminRechargePower({
      userId: currentRow.value.memberId, // ✅ 修改为 memberId
      amount: rechargePowerForm.value.amount,
      remark: rechargePowerForm.value.remark || `管理员手动充值 ${rechargePowerForm.value.amount} 算力`,
    });
    message.success(`充值成功！用户 ${currentRow.value.username} 新的算力余额为: ${res?.newBalance || rechargePowerForm.value.amount}`);
    showRechargePowerModal.value = false;
    showDetailModal.value = false;
    loadData(); // 刷新列表
  } catch (error: any) {
    message.error(error?.message || '充值失败');
  } finally {
    rechargeLoading.value = false;
  }
};

// 提交充值余额
const submitRechargeBalance = async () => {
  try {
    await rechargeBalanceFormRef.value?.validate();
  } catch (error) {
    return;
  }

  if (!currentRow.value?.memberId) {
    message.error('未选择用户');
    return;
  }

  rechargeBalanceLoading.value = true;
  try {
    const res = await ToogoUserApi.adminRechargeBalance({
      userId: currentRow.value.memberId, // ✅ 修改为 memberId
      amount: rechargeBalanceForm.value.amount,
      remark: rechargeBalanceForm.value.remark || `管理员手动充值 ${rechargeBalanceForm.value.amount} USDT`,
    });
    // 更新 currentRow 的余额
    if (currentRow.value && res?.newBalance !== undefined) {
      currentRow.value.balance = res.newBalance;
    }
    message.success(`充值成功！用户 ${currentRow.value.username} 新的余额为: ${res?.newBalance || rechargeBalanceForm.value.amount} USDT`);
    showRechargeBalanceModal.value = false;
    loadData(); // 刷新列表
  } catch (error: any) {
    message.error(error?.message || '充值失败');
  } finally {
    rechargeBalanceLoading.value = false;
  }
};

// 提交充值积分
const submitRechargePoints = async () => {
  try {
    await rechargePointsFormRef.value?.validate();
  } catch (error) {
    return;
  }

  if (!currentRow.value?.memberId) {
    message.error('未选择用户');
    return;
  }

  rechargePointsLoading.value = true;
  try {
    const res = await ToogoUserApi.adminRechargePoints({
      userId: currentRow.value.memberId, // ✅ 修改为 memberId
      amount: rechargePointsForm.value.amount,
      remark: rechargePointsForm.value.remark || `管理员手动充值 ${rechargePointsForm.value.amount} 积分`,
    });
    // 更新 currentRow 的积分
    if (currentRow.value && res?.newBalance !== undefined) {
      currentRow.value.giftPower = res.newBalance;
    }
    message.success(`充值成功！用户 ${currentRow.value.username} 新的积分余额为: ${res?.newBalance || rechargePointsForm.value.amount}`);
    showRechargePointsModal.value = false;
    loadData(); // 刷新列表
  } catch (error: any) {
    message.error(error?.message || '充值失败');
  } finally {
    rechargePointsLoading.value = false;
  }
};

onMounted(() => {
  // 默认进入“用户列表”
  handleTabChange(activeTab.value);
  loadData();
  loadAgentLevels();
});
</script>

<style scoped lang="less">
.user-admin-page {
  padding: 16px;
}
</style>

