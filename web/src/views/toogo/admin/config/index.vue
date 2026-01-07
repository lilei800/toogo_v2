<template>
  <div class="toogo-config-page">
    <n-card title="Toogo系统配置" :bordered="false">
      <template #header-extra>
        <!-- 代理配置为独立保存逻辑（见 ProxyConfig 组件内按钮），避免误点这里导致“开关不生效”的错觉 -->
        <n-button
          v-if="activeGroup !== 'proxy'"
          type="primary"
          @click="handleSave"
          :loading="saveLoading"
        >
          <template #icon>
            <n-icon><SaveOutline /></n-icon>
          </template>
          保存配置
        </n-button>
      </template>

      <n-tabs type="card" animated v-model:value="activeGroup">
        <n-tab-pane v-for="group in groups" :key="group.key" :name="group.key" :tab="group.label">
          <div class="config-group">
            <!-- 代理配置特殊处理 -->
            <ProxyConfig v-if="group.key === 'proxy'" @save="handleProxySave" />
            <!-- 其他配置 -->
            <n-form
              v-else
              ref="formRef"
              :model="formData"
              label-placement="left"
              label-width="180px"
            >
              <template v-for="item in getGroupConfigs(group.key)" :key="item.key">
                <!-- 字符串类型 -->
                <n-form-item v-if="item.type === 'string'" :label="item.name">
                  <n-input
                    v-model:value="formData[`${item.group}_${item.key}`]"
                    :placeholder="item.description"
                    style="max-width: 400px"
                  />
                  <span class="config-desc">{{ item.description }}</span>
                </n-form-item>

                <!-- 数字类型 -->
                <n-form-item v-else-if="item.type === 'number'" :label="item.name">
                  <n-input-number
                    v-model:value="formData[`${item.group}_${item.key}`]"
                    :placeholder="item.description"
                    :precision="getNumberPrecision(item.key)"
                    style="max-width: 200px"
                  />
                  <span class="config-desc">{{ item.description }}</span>
                </n-form-item>

                <!-- 布尔类型 -->
                <n-form-item v-else-if="item.type === 'boolean'" :label="item.name">
                  <n-switch
                    v-model:value="formData[`${item.group}_${item.key}`]"
                    :checked-value="'1'"
                    :unchecked-value="'0'"
                  />
                  <span class="config-desc">{{ item.description }}</span>
                </n-form-item>

                <!-- 选择类型 -->
                <n-form-item v-else-if="item.type === 'select'" :label="item.name">
                  <n-select
                    v-model:value="formData[`${item.group}_${item.key}`]"
                    :options="item.options || []"
                    style="max-width: 300px"
                  />
                  <span class="config-desc">{{ item.description }}</span>
                </n-form-item>

                <!-- JSON类型 -->
                <n-form-item v-else-if="item.type === 'json'" :label="item.name">
                  <n-input
                    v-model:value="formData[`${item.group}_${item.key}`]"
                    type="textarea"
                    :rows="4"
                    :placeholder="item.description"
                    style="max-width: 600px"
                  />
                  <span class="config-desc">{{ item.description }}</span>
                </n-form-item>

                <!-- 默认文本 -->
                <n-form-item v-else :label="item.name">
                  <n-input
                    v-model:value="formData[`${item.group}_${item.key}`]"
                    :placeholder="item.description"
                    style="max-width: 400px"
                  />
                  <span class="config-desc">{{ item.description }}</span>
                </n-form-item>
              </template>
            </n-form>
          </div>
        </n-tab-pane>
      </n-tabs>
    </n-card>
  </div>
</template>

<script lang="ts" setup>
  import { ref, reactive, onMounted } from 'vue';
  import { useMessage } from 'naive-ui';
  import { SaveOutline } from '@vicons/ionicons5';
  import { http } from '@/utils/http/axios';
  import ProxyConfig from './components/ProxyConfig.vue';

  const message = useMessage();
  const saveLoading = ref(false);
  const activeGroup = ref('basic');

  // 配置分组
  const groups = ref([
    { key: 'basic', label: '基础配置' },
    { key: 'register', label: '注册配置' },
    { key: 'power', label: '算力配置' },
    { key: 'commission', label: '佣金配置' },
    { key: 'withdraw', label: '提现配置' },
    { key: 'invite', label: '邀请配置' },
    { key: 'robot', label: '机器人配置' },
    { key: 'proxy', label: '代理配置' },
  ]);

  // 配置列表
  const configList = ref<any[]>([]);

  // 表单数据
  const formData = reactive<Record<string, any>>({});

  // 原始数据(用于对比变化)
  const originalData = reactive<Record<string, any>>({});

  // 获取指定分组的配置
  const getGroupConfigs = (groupKey: string) => {
    return configList.value.filter((item) => item.group === groupKey);
  };

  // 获取数字精度
  const getNumberPrecision = (key: string) => {
    if (key.includes('rate') || key.includes('percent')) {
      return 4;
    }
    return 2;
  };

  // 加载配置
  const loadConfig = async () => {
    try {
      const res = await http.request({
        url: '/toogo/config/list',
        method: 'get',
      });

      if (res.code === 0) {
        configList.value = res.data?.list || [];

        // 初始化表单数据
        configList.value.forEach((item) => {
          const key = `${item.group}_${item.key}`;
          formData[key] = item.value;
          originalData[key] = item.value;
        });
      } else {
        // 使用默认配置数据
        configList.value = getDefaultConfigs();
        configList.value.forEach((item) => {
          const key = `${item.group}_${item.key}`;
          formData[key] = item.value;
          originalData[key] = item.value;
        });
      }
    } catch (error) {
      // 使用默认配置数据
      configList.value = getDefaultConfigs();
      configList.value.forEach((item) => {
        const key = `${item.group}_${item.key}`;
        formData[key] = item.value;
        originalData[key] = item.value;
      });
    }
  };

  // 获取默认配置(用于演示)
  const getDefaultConfigs = () => {
    return [
      // 基础配置
      {
        group: 'basic',
        key: 'site_name',
        value: 'Toogo.Ai',
        type: 'string',
        name: '站点名称',
        description: '系统名称',
      },
      {
        group: 'basic',
        key: 'site_logo',
        value: '/images/toogo-logo.png',
        type: 'string',
        name: '站点Logo',
        description: 'Logo图片路径',
      },
      {
        group: 'basic',
        key: 'site_description',
        value: '全自动虚拟货币量化交易系统',
        type: 'string',
        name: '站点描述',
        description: '',
      },
      {
        group: 'basic',
        key: 'contact_email',
        value: 'support@toogo.ai',
        type: 'string',
        name: '联系邮箱',
        description: '客服邮箱',
      },
      {
        group: 'basic',
        key: 'contact_telegram',
        value: '',
        type: 'string',
        name: 'Telegram群组',
        description: 'Telegram群组链接',
      },

      // 注册配置
      {
        group: 'register',
        key: 'register_enabled',
        value: '1',
        type: 'boolean',
        name: '开放注册',
        description: '是否允许新用户注册',
      },
      {
        group: 'register',
        key: 'need_invite_code',
        value: '1',
        type: 'boolean',
        name: '需要邀请码',
        description: '注册时必须填写邀请码',
      },
      {
        group: 'register',
        key: 'invite_code_required',
        value: '1',
        type: 'boolean',
        name: '邀请码必填',
        description: '邀请码为空时禁止注册',
      },
      {
        group: 'register',
        key: 'email_verify',
        value: '0',
        type: 'boolean',
        name: '邮箱验证',
        description: '注册时需要验证邮箱',
      },
      {
        group: 'register',
        key: 'captcha_enabled',
        value: '1',
        type: 'boolean',
        name: '验证码',
        description: '注册时显示验证码',
      },
      {
        group: 'register',
        key: 'default_role',
        value: 'user',
        type: 'select',
        name: '默认角色',
        description: '新用户默认角色',
        options: [
          { label: '普通用户', value: 'user' },
          { label: 'VIP用户', value: 'vip' },
        ],
      },
      {
        group: 'register',
        key: 'default_vip_level',
        value: '0',
        type: 'number',
        name: '默认VIP等级',
        description: '新用户默认VIP等级(0-10)',
      },
      {
        group: 'register',
        key: 'free_trial_days',
        value: '7',
        type: 'number',
        name: '免费试用天数',
        description: '新用户免费试用期',
      },
      {
        group: 'register',
        key: 'free_robot_limit',
        value: '1',
        type: 'number',
        name: '免费机器人数',
        description: '免费试用期可用机器人数量',
      },
      {
        group: 'register',
        key: 'register_gift_power',
        value: '50',
        type: 'number',
        name: '注册赠送积分',
        description: '新用户注册赠送的积分',
      },
      {
        group: 'register',
        key: 'register_notice',
        value: '',
        type: 'string',
        name: '注册公告',
        description: '显示在注册页面的公告信息',
      },

      // 算力配置
      {
        group: 'power',
        key: 'consume_rate',
        value: '0.10',
        type: 'number',
        name: '算力消耗比例',
        description: '云机器人消耗比例(默认10%)',
      },
      {
        group: 'power',
        key: 'min_consume',
        value: '0.01',
        type: 'number',
        name: '最小消耗算力',
        description: '单笔最小消耗',
      },
      {
        group: 'power',
        key: 'exchange_rate',
        value: '1.00',
        type: 'number',
        name: 'USDT兑算力比率',
        description: '1 USDT = ? 算力',
      },
      {
        group: 'power',
        key: 'gift_power_priority',
        value: '1',
        type: 'boolean',
        name: '优先消耗积分',
        description: '优先使用积分',
      },

      // 佣金配置
      {
        group: 'commission',
        key: 'subscribe_rate_1',
        value: '0.10',
        type: 'number',
        name: '订阅一级佣金',
        description: '普通用户订阅一级佣金比例',
      },
      {
        group: 'commission',
        key: 'subscribe_rate_2',
        value: '0.05',
        type: 'number',
        name: '订阅二级佣金',
        description: '普通用户订阅二级佣金比例',
      },
      {
        group: 'commission',
        key: 'power_rate_1',
        value: '0.05',
        type: 'number',
        name: '算力一级佣金',
        description: '普通用户算力消耗一级佣金比例',
      },
      {
        group: 'commission',
        key: 'power_rate_2',
        value: '0.02',
        type: 'number',
        name: '算力二级佣金',
        description: '普通用户算力消耗二级佣金比例',
      },
      {
        group: 'commission',
        key: 'agent_enabled',
        value: '1',
        type: 'boolean',
        name: '启用代理商体系',
        description: '',
      },

      // 提现配置
      {
        group: 'withdraw',
        key: 'min_amount',
        value: '10',
        type: 'number',
        name: '最低提现金额',
        description: 'USDT',
      },
      {
        group: 'withdraw',
        key: 'fee_rate',
        value: '0.02',
        type: 'number',
        name: '提现手续费比例',
        description: '',
      },
      {
        group: 'withdraw',
        key: 'daily_limit',
        value: '10000',
        type: 'number',
        name: '每日提现限额',
        description: 'USDT',
      },
      {
        group: 'withdraw',
        key: 'auto_audit_limit',
        value: '100',
        type: 'number',
        name: '自动审核限额',
        description: '低于此金额自动通过审核',
      },
      {
        group: 'withdraw',
        key: 'withdraw_enabled',
        value: '1',
        type: 'boolean',
        name: '开放提现',
        description: '',
      },

      // 邀请配置
      {
        group: 'invite',
        key: 'code_expire_hours',
        value: '24',
        type: 'number',
        name: '邀请码有效期',
        description: '小时',
      },
      {
        group: 'invite',
        key: 'register_reward',
        value: '30',
        type: 'number',
        name: '邀请奖励算力',
        description: '邀请人和被邀请人各获得',
      },
      {
        group: 'invite',
        key: 'max_daily_invites',
        value: '10',
        type: 'number',
        name: '每日邀请上限',
        description: '每人每天最多生成的邀请码数量',
      },
      {
        group: 'invite',
        key: 'invite_poster_enabled',
        value: '1',
        type: 'boolean',
        name: '启用邀请海报',
        description: '',
      },

      // 机器人配置
      {
        group: 'robot',
        key: 'default_leverage',
        value: '10',
        type: 'number',
        name: '默认杠杆倍数',
        description: '',
      },
      {
        group: 'robot',
        key: 'max_leverage',
        value: '125',
        type: 'number',
        name: '最大杠杆倍数',
        description: '',
      },
      {
        group: 'robot',
        key: 'default_margin_percent',
        value: '30',
        type: 'number',
        name: '默认保证金比例',
        description: '%',
      },
      {
        group: 'robot',
        key: 'default_stop_loss',
        value: '10',
        type: 'number',
        name: '默认止损比例',
        description: '%',
      },
      {
        group: 'robot',
        key: 'default_profit_retreat',
        value: '18',
        type: 'number',
        name: '默认止盈回撤比例',
        description: '%',
      },
      {
        group: 'robot',
        key: 'default_start_retreat',
        value: '8',
        type: 'number',
        name: '默认启动回撤比例',
        description: '%',
      },
      {
        group: 'robot',
        key: 'auto_order_enabled',
        value: '1',
        type: 'boolean',
        name: '允许自动下单',
        description: '全局开关',
      },
      {
        group: 'robot',
        key: 'auto_close_enabled',
        value: '1',
        type: 'boolean',
        name: '允许自动平仓',
        description: '全局开关',
      },
    ];
  };

  // 保存配置
  const handleSave = async () => {
    saveLoading.value = true;

    try {
      // 收集变化的配置
      const changedItems: any[] = [];

      configList.value.forEach((item) => {
        const key = `${item.group}_${item.key}`;
        if (formData[key] !== originalData[key]) {
          changedItems.push({
            group: item.group,
            key: item.key,
            value: String(formData[key]),
          });
        }
      });

      if (changedItems.length === 0) {
        message.info('没有需要保存的修改');
        saveLoading.value = false;
        return;
      }

      // 调用API保存
      const res = await http.request({
        url: '/toogo/config/update',
        method: 'post',
        data: { items: changedItems },
      });

      if (res.code === 0) {
        message.success('保存成功');
        // 更新原始数据
        changedItems.forEach((item) => {
          const key = `${item.group}_${item.key}`;
          originalData[key] = item.value;
        });
      } else {
        message.error(res.message || '保存失败');
      }
    } catch (error) {
      message.error('保存失败');
      console.error(error);
    } finally {
      saveLoading.value = false;
    }
  };

  // 处理代理配置保存
  const handleProxySave = () => {
    message.success('代理配置已保存');
  };

  onMounted(() => {
    loadConfig();
  });
</script>

<style scoped>
  .toogo-config-page {
    padding: 16px;
  }

  .config-group {
    padding: 20px 0;
  }

  .config-desc {
    margin-left: 16px;
    color: #888;
    font-size: 13px;
  }

  :deep(.n-form-item) {
    margin-bottom: 24px;
  }

  :deep(.n-tabs-tab) {
    padding: 12px 20px !important;
  }
</style>
