<template>
  <div class="toogo-api">
    <!-- API开通教程提示 -->
    <n-alert type="info" style="margin-bottom: 16px" closable>
      <template #icon>
        <n-icon><InfoCircleOutlined /></n-icon>
      </template>
      首次使用？请先在交易所开通API密钥。
      <n-button text type="primary" @click="showGuideDrawer = true" style="margin-left: 8px">
        查看开通教程 →
      </n-button>
    </n-alert>

    <!-- 重要提醒：合约资金与权限 -->
    <n-alert type="warning" style="margin-bottom: 16px">
      <template #icon>
        <n-icon><WarningOutlined /></n-icon>
      </template>
      <div>
        <n-text strong>添加API前请确认：</n-text>
        <ul style="margin: 6px 0 0 18px; padding: 0">
          <li>
            <n-text strong>1、合约账户内有资金</n-text>
            （尤其 Gate：若提示 <n-text code>USER_NOT_FOUND</n-text>，请先从现货划转任意资金到 USDT
            Futures 以创建合约账户）
          </li>
          <li>
            <n-text strong>2、API权限</n-text>
            必须开通：<n-text code>合约读写</n-text> + <n-text code>资金只读</n-text>
          </li>
          <li>
            <n-text strong>3、添加完成后请先点击"测试"</n-text>
            ，测试通过后再使用
          </li>
          <li>
            <n-text strong>4、每个机器人只能绑定一个API接口</n-text>
            ，可以在交易所添加多个子账号API对应多机器人
          </li>
        </ul>
      </div>
    </n-alert>

    <n-card title="API密钥管理" :bordered="false" class="proCard">
      <template #header-extra>
        <n-space>
          <n-button tertiary type="info" @click="showGuideDrawer = true">
            <template #icon>
              <n-icon><QuestionCircleOutlined /></n-icon>
            </template>
            开通教程
          </n-button>
          <n-button type="primary" @click="openAddModal">
            <template #icon>
              <n-icon><PlusOutlined /></n-icon>
            </template>
            添加API
          </n-button>
        </n-space>
      </template>

      <n-data-table
        :columns="columns"
        :data="apiList"
        :loading="loading"
        :pagination="pagination"
        :row-key="(row) => row.id"
      />
    </n-card>

    <!-- 添加/编辑API弹窗 -->
    <n-modal
      v-model:show="showAddModal"
      preset="dialog"
      :title="editingApi ? '编辑API' : '添加API'"
      :positive-text="editingApi ? '保存' : '添加'"
      negative-text="取消"
      @positive-click="handleSubmit"
      @negative-click="handleCancel"
      style="width: 600px"
    >
      <n-form
        ref="formRef"
        :model="formData"
        :rules="rules"
        label-placement="left"
        label-width="100"
      >
        <n-form-item label="交易所" path="exchange">
          <n-select
            v-model:value="formData.exchange"
            :options="exchangeOptions"
            placeholder="请选择交易所"
          />
        </n-form-item>

        <n-form-item label="API名称" path="name">
          <n-input v-model:value="formData.name" placeholder="请输入API名称" />
        </n-form-item>

        <n-form-item label="状态" path="status">
          <n-switch v-model:value="formData.status" :checked-value="1" :unchecked-value="2">
            <template #checked>正常</template>
            <template #unchecked>禁用</template>
          </n-switch>
        </n-form-item>

        <n-form-item label="API Key" path="apiKey">
          <n-input
            v-model:value="formData.apiKey"
            :placeholder="editingApi ? '留空则不修改' : '请输入API Key'"
          />
        </n-form-item>

        <n-form-item label="Secret Key" path="secretKey">
          <n-input
            v-model:value="formData.secretKey"
            type="password"
            :placeholder="editingApi ? '留空则不修改' : '请输入Secret Key'"
            show-password-on="click"
          />
        </n-form-item>

        <n-form-item v-if="needPassphrase" label="Passphrase" path="passphrase">
          <n-input
            v-model:value="formData.passphrase"
            type="password"
            :placeholder="editingApi ? '留空则不修改' : '请输入Passphrase（OKX必填）'"
            show-password-on="click"
          />
        </n-form-item>

        <n-form-item label="设为默认" path="isDefault">
          <n-switch v-model:value="formData.isDefault" />
        </n-form-item>
      </n-form>
    </n-modal>

    <!-- 测试连接结果 -->
    <n-modal
      v-model:show="showTestResult"
      preset="dialog"
      :title="testResult.success ? '连接成功' : '连接失败'"
      positive-text="确定"
    >
      <div class="test-result">
        <n-result
          :status="testResult.success ? 'success' : 'error'"
          :title="testResult.success ? '连接成功' : '连接失败'"
          :description="testResult.message"
        />
        <div v-if="testResult.success && testResult.balance" class="balance-info">
          <n-descriptions label-placement="left" bordered :column="1">
            <n-descriptions-item label="可用余额(USDT)">
              {{ testResult.balance.available }}
            </n-descriptions-item>
            <n-descriptions-item label="总资产(USDT)">
              {{ testResult.balance.total }}
            </n-descriptions-item>
          </n-descriptions>
        </div>
      </div>
    </n-modal>

    <!-- API开通教程抽屉 -->
    <n-drawer v-model:show="showGuideDrawer" :width="720" placement="right">
      <n-drawer-content title="交易所API开通教程" closable>
        <n-tabs type="line" animated>
          <!-- Binance教程 -->
          <n-tab-pane name="binance" tab="Binance (币安)">
            <div class="guide-content">
              <n-alert type="warning" style="margin-bottom: 16px">
                <template #icon>
                  <n-icon><WarningOutlined /></n-icon>
                </template>
                <strong>重要提示：</strong
                >请务必启用"现货交易"权限，禁用"提现"权限，并绑定IP白名单以提升安全性。
              </n-alert>

              <n-steps vertical :current="1">
                <n-step title="登录Binance账户">
                  <div class="step-content">
                    <p>1. 访问 <n-text code>https://www.binance.com</n-text> 并登录您的账户</p>
                    <p>2. 点击右上角头像，选择「API管理」</p>
                  </div>
                </n-step>

                <n-step title="创建API密钥">
                  <div class="step-content">
                    <p>1. 点击「创建API」按钮</p>
                    <p>2. 选择「系统生成」方式（推荐）</p>
                    <p>3. 设置API密钥备注名称（如：ToogoAI交易）</p>
                    <p>4. 完成安全验证（邮箱/手机验证码）</p>
                  </div>
                </n-step>

                <n-step title="配置API权限">
                  <div class="step-content">
                    <n-space vertical>
                      <n-alert type="success" size="small">
                        <template #icon
                          ><n-icon><CheckCircleOutlined /></n-icon
                        ></template>
                        <strong>必须启用：</strong>启用现货及杠杆交易
                      </n-alert>
                      <n-alert type="error" size="small">
                        <template #icon
                          ><n-icon><CloseCircleOutlined /></n-icon
                        ></template>
                        <strong>禁止启用：</strong>启用提现（确保资金安全）
                      </n-alert>
                      <p style="margin-top: 8px">其他权限保持默认关闭即可</p>
                    </n-space>
                  </div>
                </n-step>

                <n-step title="绑定IP白名单（推荐）">
                  <div class="step-content">
                    <p>1. 在API管理页面点击「编辑限制」</p>
                    <p>2. 选择「限制访问受信任的IP」</p>
                    <p>3. 添加以下IP地址：</p>
                    <n-code :code="serverIPs.join('\n')" language="text" style="margin: 8px 0" />
                    <n-alert type="info" size="small" style="margin-top: 8px">
                      IP白名单可以有效防止API密钥被盗用
                    </n-alert>
                  </div>
                </n-step>

                <n-step title="保存密钥信息">
                  <div class="step-content">
                    <p>1. 复制 <n-text strong>API Key</n-text>（公钥）</p>
                    <p>2. 复制 <n-text strong>Secret Key</n-text>（私钥，仅显示一次）</p>
                    <n-alert type="error" style="margin-top: 8px">
                      <template #icon
                        ><n-icon><LockOutlined /></n-icon
                      ></template>
                      <strong>安全提示：</strong>Secret
                      Key只会显示一次，请妥善保管，切勿泄露给他人！
                    </n-alert>
                  </div>
                </n-step>

                <n-step title="添加到系统">
                  <div class="step-content">
                    <p>1. 返回本页面，点击「添加API」按钮</p>
                    <p>2. 选择交易所：Binance</p>
                    <p>3. 填入API Key和Secret Key</p>
                    <p>4. 点击「测试」验证连接</p>
                  </div>
                </n-step>
              </n-steps>
            </div>
          </n-tab-pane>

          <!-- OKX教程 -->
          <n-tab-pane name="okx" tab="OKX (欧易)">
            <div class="guide-content">
              <n-alert type="warning" style="margin-bottom: 16px">
                <template #icon>
                  <n-icon><WarningOutlined /></n-icon>
                </template>
                <strong>重要提示：</strong>OKX需要设置Passphrase（API密码），请务必记住此密码！
              </n-alert>

              <n-steps vertical :current="1">
                <n-step title="登录OKX账户">
                  <div class="step-content">
                    <p>1. 访问 <n-text code>https://www.okx.com</n-text> 并登录您的账户</p>
                    <p>2. 点击右上角头像，选择「API」</p>
                  </div>
                </n-step>

                <n-step title="创建API密钥">
                  <div class="step-content">
                    <p>1. 点击「创建V5 API密钥」（推荐使用V5版本）</p>
                    <p>2. 设置API密钥名称（如：ToogoAI）</p>
                    <p>3. 设置Passphrase（API密码，至少8位，建议使用复杂密码）</p>
                    <n-alert type="info" size="small" style="margin-top: 8px">
                      <strong>注意：</strong
                      >Passphrase是您自己设置的密码，请务必记住，丢失后无法找回！
                    </n-alert>
                  </div>
                </n-step>

                <n-step title="配置API权限">
                  <div class="step-content">
                    <n-space vertical>
                      <n-alert type="success" size="small">
                        <template #icon
                          ><n-icon><CheckCircleOutlined /></n-icon
                        ></template>
                        <strong>必须启用：</strong>交易权限（Trade）
                      </n-alert>
                      <n-alert type="success" size="small">
                        <template #icon
                          ><n-icon><CheckCircleOutlined /></n-icon
                        ></template>
                        <strong>建议启用：</strong>读取权限（Read）
                      </n-alert>
                      <n-alert type="error" size="small">
                        <template #icon
                          ><n-icon><CloseCircleOutlined /></n-icon
                        ></template>
                        <strong>禁止启用：</strong>提现权限（Withdraw）
                      </n-alert>
                    </n-space>
                  </div>
                </n-step>

                <n-step title="绑定IP白名单（推荐）">
                  <div class="step-content">
                    <p>1. 在创建API时选择「绑定IP地址」</p>
                    <p>2. 输入以下IP地址（多个IP用换行分隔）：</p>
                    <n-code :code="serverIPs.join('\n')" language="text" style="margin: 8px 0" />
                  </div>
                </n-step>

                <n-step title="保存密钥信息">
                  <div class="step-content">
                    <p>需要保存以下三个信息：</p>
                    <p>1. <n-text strong>API Key</n-text>（公钥）</p>
                    <p>2. <n-text strong>Secret Key</n-text>（私钥，仅显示一次）</p>
                    <p>3. <n-text strong>Passphrase</n-text>（API密码，您设置的）</p>
                    <n-alert type="error" style="margin-top: 8px">
                      <template #icon
                        ><n-icon><LockOutlined /></n-icon
                      ></template>
                      <strong>安全提示：</strong>所有密钥信息都必须妥善保管，切勿泄露！
                    </n-alert>
                  </div>
                </n-step>

                <n-step title="添加到系统">
                  <div class="step-content">
                    <p>1. 返回本页面，点击「添加API」按钮</p>
                    <p>2. 选择交易所：OKX</p>
                    <p>3. 填入API Key、Secret Key和Passphrase</p>
                    <p>4. 点击「测试」验证连接</p>
                  </div>
                </n-step>
              </n-steps>
            </div>
          </n-tab-pane>

          <!-- Gate.io教程 -->
          <n-tab-pane name="gate" tab="Gate.io">
            <div class="guide-content">
              <n-alert type="info" style="margin-bottom: 16px">
                <template #icon>
                  <n-icon><InfoCircleOutlined /></n-icon>
                </template>
                <strong>提示：</strong>Gate.io不需要Passphrase，相对简单。
              </n-alert>

              <n-steps vertical :current="1">
                <n-step title="登录Gate.io账户">
                  <div class="step-content">
                    <p>1. 访问 <n-text code>https://www.gate.io</n-text> 并登录您的账户</p>
                    <p>2. 点击右上角头像，选择「API KEYs」</p>
                  </div>
                </n-step>

                <n-step title="创建API密钥">
                  <div class="step-content">
                    <p>1. 点击「创建API KEY」按钮</p>
                    <p>2. 设置API备注名称（如：ToogoAI）</p>
                    <p>3. 完成安全验证（邮箱/手机验证）</p>
                  </div>
                </n-step>

                <n-step title="配置API权限">
                  <div class="step-content">
                    <n-space vertical>
                      <n-alert type="success" size="small">
                        <template #icon
                          ><n-icon><CheckCircleOutlined /></n-icon
                        ></template>
                        <strong>必须启用：</strong>现货交易权限（Spot Trading）
                      </n-alert>
                      <n-alert type="success" size="small">
                        <template #icon
                          ><n-icon><CheckCircleOutlined /></n-icon
                        ></template>
                        <strong>建议启用：</strong>查看权限（Read Only）
                      </n-alert>
                      <n-alert type="error" size="small">
                        <template #icon
                          ><n-icon><CloseCircleOutlined /></n-icon
                        ></template>
                        <strong>禁止启用：</strong>提现权限（Withdrawal）
                      </n-alert>
                    </n-space>
                  </div>
                </n-step>

                <n-step title="绑定IP白名单（推荐）">
                  <div class="step-content">
                    <p>1. 在「IP白名单」选项中启用</p>
                    <p>2. 添加以下IP地址：</p>
                    <n-code :code="serverIPs.join('\n')" language="text" style="margin: 8px 0" />
                  </div>
                </n-step>

                <n-step title="保存密钥信息">
                  <div class="step-content">
                    <p>需要保存以下两个信息：</p>
                    <p>1. <n-text strong>API Key</n-text></p>
                    <p>2. <n-text strong>Secret Key</n-text>（仅显示一次）</p>
                    <n-alert type="error" style="margin-top: 8px">
                      <template #icon
                        ><n-icon><LockOutlined /></n-icon
                      ></template>
                      <strong>安全提示：</strong>密钥信息务必保密，丢失后需重新创建！
                    </n-alert>
                  </div>
                </n-step>

                <n-step title="添加到系统">
                  <div class="step-content">
                    <p>1. 返回本页面，点击「添加API」按钮</p>
                    <p>2. 选择交易所：Gate.io</p>
                    <p>3. 填入API Key和Secret Key</p>
                    <p>4. 点击「测试」验证连接</p>
                  </div>
                </n-step>
              </n-steps>
            </div>
          </n-tab-pane>
        </n-tabs>

        <!-- 通用安全提示 -->
        <n-divider style="margin: 24px 0" />
        <n-card title="安全注意事项" size="small" :bordered="false" style="background: #f8f9fa">
          <n-space vertical>
            <n-alert type="warning" size="small">
              <template #icon
                ><n-icon><SafetyOutlined /></n-icon
              ></template>
              <strong>1. 权限最小化原则：</strong>只开启必要的交易权限，禁用提现等高风险权限
            </n-alert>
            <n-alert type="warning" size="small">
              <template #icon
                ><n-icon><SafetyOutlined /></n-icon
              ></template>
              <strong>2. IP白名单保护：</strong>强烈建议绑定IP白名单，防止密钥被盗用
            </n-alert>
            <n-alert type="warning" size="small">
              <template #icon
                ><n-icon><SafetyOutlined /></n-icon
              ></template>
              <strong>3. 定期更换密钥：</strong>建议每3-6个月更换一次API密钥
            </n-alert>
            <n-alert type="warning" size="small">
              <template #icon
                ><n-icon><SafetyOutlined /></n-icon
              ></template>
              <strong>4. 监控异常活动：</strong>定期检查交易记录，发现异常立即禁用API
            </n-alert>
            <n-alert type="error" size="small">
              <template #icon
                ><n-icon><LockOutlined /></n-icon
              ></template>
              <strong>5. 密钥保密：</strong>切勿将API密钥分享给任何人或在公开场合展示
            </n-alert>
          </n-space>
        </n-card>
      </n-drawer-content>
    </n-drawer>
  </div>
</template>

<script lang="ts" setup>
  import { ref, computed, onMounted, h } from 'vue';
  import { useMessage, useDialog, NButton, NTag, NSpace, NIcon, NPopconfirm } from 'naive-ui';
  import {
    PlusOutlined,
    EditOutlined,
    DeleteOutlined,
    CheckCircleOutlined,
    CloseCircleOutlined,
    ApiOutlined,
    InfoCircleOutlined,
    QuestionCircleOutlined,
    WarningOutlined,
    LockOutlined,
    SafetyOutlined,
  } from '@vicons/antd';
  import { http } from '@/utils/http/axios';

  const message = useMessage();
  const dialog = useDialog();

  // 状态
  const loading = ref(false);
  const apiList = ref<any[]>([]);
  const showAddModal = ref(false);
  const showTestResult = ref(false);
  const showGuideDrawer = ref(false);
  const editingApi = ref<any>(null);
  const formRef = ref<any>(null);

  // 服务器IP地址（示例，实际应从配置获取）
  const serverIPs = ref(['# 请联系客服获取最新的服务器IP地址', '# 或在系统设置中查看']);

  // 测试结果
  const testResult = ref({
    success: false,
    message: '',
    balance: null as any,
  });

  // 交易所选项
  const exchangeOptions = [
    { label: 'Binance (币安)', value: 'binance' },
    { label: 'OKX (欧易)', value: 'okx' },
    { label: 'Gate.io', value: 'gate' },
  ];

  // 是否需要Passphrase
  const needPassphrase = computed(() => {
    return ['okx'].includes(formData.value.exchange);
  });

  // 表单数据
  const formData = ref({
    exchange: '',
    name: '',
    status: 1,
    apiKey: '',
    secretKey: '',
    passphrase: '',
    isDefault: false,
  });

  // 表单校验规则
  const rules = computed(() => {
    const isEdit = !!editingApi.value;
    return {
      exchange: { required: true, message: '请选择交易所', trigger: 'change' },
      name: { required: true, message: '请输入API名称', trigger: 'blur' },
      status: { required: true, type: 'number', trigger: 'change' },
      apiKey: { required: !isEdit, message: '请输入API Key', trigger: 'blur' },
      secretKey: { required: !isEdit, message: '请输入Secret Key', trigger: 'blur' },
      passphrase: {
        required: !isEdit && needPassphrase.value,
        message: '请输入Passphrase',
        trigger: 'blur',
      },
    };
  });

  // 分页配置
  const pagination = ref({
    page: 1,
    pageSize: 10,
    showSizePicker: true,
    pageSizes: [10, 20, 50],
    onChange: (page: number) => {
      pagination.value.page = page;
      loadApiList();
    },
    onUpdatePageSize: (pageSize: number) => {
      pagination.value.pageSize = pageSize;
      pagination.value.page = 1;
      loadApiList();
    },
  });

  // 表格列配置
  const columns = [
    { title: 'ID', key: 'id', width: 60 },
    { title: 'API名称', key: 'apiName', width: 150 },
    {
      title: '交易所',
      key: 'platform',
      width: 120,
      render: (row: any) => {
        const colors: Record<string, string> = {
          binance: '#F0B90B',
          okx: '#121212',
          gate: '#17E6A1',
        };
        const labels: Record<string, string> = {
          binance: 'Binance',
          okx: 'OKX',
          gate: 'Gate.io',
        };
        return h(
          NTag,
          { color: { color: colors[row.platform] || '#666', textColor: '#fff' } },
          () => labels[row.platform] || row.platform,
        );
      },
    },
    {
      title: 'API Key',
      key: 'apiKey',
      width: 200,
      render: (row: any) => {
        const key = row.apiKey || '';
        if (key.length > 12) {
          return key.substring(0, 8) + '****' + key.substring(key.length - 4);
        }
        return key;
      },
    },
    {
      title: '状态',
      key: 'status',
      width: 80,
      render: (row: any) => {
        return row.status === 1
          ? h(NTag, { type: 'success' }, () => '正常')
          : h(NTag, { type: 'error' }, () => '禁用');
      },
    },
    {
      title: '默认',
      key: 'isDefault',
      width: 80,
      render: (row: any) => {
        return row.isDefault === 1
          ? h(NIcon, { color: '#18a058', size: 20 }, () => h(CheckCircleOutlined))
          : '-';
      },
    },
    { title: '创建时间', key: 'createdAt', width: 180 },
    {
      title: '操作',
      key: 'actions',
      width: 220,
      render: (row: any) => {
        return h(NSpace, {}, () => [
          h(
            NButton,
            { size: 'small', quaternary: true, type: 'primary', onClick: () => handleTest(row) },
            { default: () => '测试', icon: () => h(NIcon, null, () => h(ApiOutlined)) },
          ),
          h(
            NButton,
            { size: 'small', quaternary: true, type: 'info', onClick: () => handleEdit(row) },
            { default: () => '编辑', icon: () => h(NIcon, null, () => h(EditOutlined)) },
          ),
          h(
            NPopconfirm,
            {
              onPositiveClick: () => handleDelete(row),
            },
            {
              trigger: () =>
                h(
                  NButton,
                  { size: 'small', quaternary: true, type: 'error', disabled: !!row.robotName },
                  { default: () => '删除', icon: () => h(NIcon, null, () => h(DeleteOutlined)) },
                ),
              default: () => '确定删除此API吗？',
            },
          ),
        ]);
      },
    },
  ];

  // 加载API列表
  async function loadApiList() {
    loading.value = true;
    try {
      const res = await http.request({
        url: '/trading/apiConfig/list',
        method: 'get',
        params: {
          page: pagination.value.page,
          pageSize: pagination.value.pageSize,
        },
      });
      apiList.value = res?.list || [];
      pagination.value.itemCount = res?.totalCount || 0;
    } catch (error) {
      console.error('加载API列表失败', error);
    } finally {
      loading.value = false;
    }
  }

  // 测试API连接
  async function handleTest(row: any) {
    try {
      const res = await http.request({
        url: '/trading/apiConfig/test',
        method: 'post',
        data: { id: row.id },
      });
      testResult.value = {
        success: res?.success ?? true,
        message: res?.message || 'API连接正常',
        balance: res?.balance ? { available: res.balance, total: res.balance } : null,
      };
      showTestResult.value = true;
    } catch (error: any) {
      // 安全地获取错误信息
      let errorMessage = '连接测试失败';
      if (error) {
        if (typeof error === 'string') {
          errorMessage = error;
        } else if (error.message) {
          errorMessage = error.message;
        } else if (error.msg) {
          errorMessage = error.msg;
        } else if (error.data?.message) {
          errorMessage = error.data.message;
        }
      }
      testResult.value = {
        success: false,
        message: errorMessage,
        balance: null,
      };
      showTestResult.value = true;
    }
  }

  // 编辑API
  function handleEdit(row: any) {
    editingApi.value = row;
    formData.value = {
      exchange: row.platform,
      name: row.apiName,
      status: row.status || 1, // 兼容历史脏数据：0 视为正常
      apiKey: '', // 不回填脱敏后的 key，避免误写回去
      secretKey: '', // 不回显密钥
      passphrase: '',
      isDefault: row.isDefault === 1,
    };
    showAddModal.value = true;
  }

  function openAddModal() {
    resetForm();
    showAddModal.value = true;
  }

  function handleCancel() {
    showAddModal.value = false;
    resetForm();
  }

  // 删除API
  async function handleDelete(row: any) {
    try {
      await http.request({
        url: '/trading/apiConfig/delete',
        method: 'post',
        data: { id: row.id },
      });
      message.success('删除成功');
      loadApiList();
    } catch (error: any) {
      message.error(error.message || '删除失败');
    }
  }

  // 提交表单
  async function handleSubmit() {
    try {
      await formRef.value?.validate();
      const url = editingApi.value ? '/trading/apiConfig/update' : '/trading/apiConfig/create';
      const data: any = editingApi.value
        ? {
            id: editingApi.value.id,
            platform: formData.value.exchange,
            apiName: formData.value.name,
            isDefault: formData.value.isDefault ? 1 : 0,
            status: formData.value.status,
          }
        : {
            platform: formData.value.exchange,
            apiName: formData.value.name,
            apiKey: formData.value.apiKey,
            secretKey: formData.value.secretKey,
            passphrase: formData.value.passphrase,
            isDefault: formData.value.isDefault ? 1 : 0,
            status: formData.value.status,
          };

      // 编辑时：密钥留空则不修改（避免把脱敏/空值写回去）
      if (editingApi.value) {
        if (formData.value.apiKey) data.apiKey = formData.value.apiKey;
        if (formData.value.secretKey) data.secretKey = formData.value.secretKey;
        if (formData.value.passphrase) data.passphrase = formData.value.passphrase;
      }

      await http.request({
        url,
        method: 'post',
        data,
      });

      message.success(editingApi.value ? '更新成功' : '添加成功');
      showAddModal.value = false;
      resetForm();
      loadApiList();
    } catch (error: any) {
      if (error?.message) {
        message.error(error.message);
      }
    }
    return false;
  }

  // 重置表单
  function resetForm() {
    editingApi.value = null;
    formData.value = {
      exchange: '',
      name: '',
      status: 1,
      apiKey: '',
      secretKey: '',
      passphrase: '',
      isDefault: false,
    };
  }

  onMounted(() => {
    loadApiList();
  });
</script>

<style scoped lang="less">
  .toogo-api {
    .test-result {
      padding: 20px 0;

      .balance-info {
        margin-top: 20px;
      }
    }

    .guide-content {
      padding: 8px 0;

      .step-content {
        padding: 12px 0;
        line-height: 1.8;

        p {
          margin: 8px 0;
          color: #666;

          &:first-child {
            margin-top: 0;
          }
        }

        :deep(.n-alert) {
          margin: 8px 0;
        }

        :deep(.n-code) {
          max-height: 120px;
          overflow-y: auto;
        }
      }

      :deep(.n-steps) {
        .n-step {
          padding-bottom: 24px;

          &:last-child {
            padding-bottom: 0;
          }
        }

        .n-step-indicator {
          .n-step-indicator__inner {
            font-weight: 600;
          }
        }

        .n-step__title {
          font-weight: 600;
          font-size: 15px;
          color: #333;
        }
      }
    }
  }
</style>
