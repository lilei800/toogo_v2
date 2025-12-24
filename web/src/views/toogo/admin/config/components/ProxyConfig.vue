<template>
  <div class="proxy-config">
    <n-card title="代理配置" :bordered="false" style="margin-bottom: 16px">
      <n-alert type="info" style="margin-bottom: 16px">
        <template #header>使用说明</template>
        <div style="line-height: 1.8">
          <p><strong>代理配置仅用于API对接交易所平台</strong></p>
          <p>当您的服务器无法直接访问交易所API时，需要配置代理。代理配置仅在以下场景使用：</p>
          <ul style="margin: 8px 0; padding-left: 20px">
            <li>用户API配置测试连接</li>
            <li>交易机器人执行交易操作</li>
            <li>订单查询、账户查询等API调用</li>
          </ul>
          <p style="margin-top: 8px; color: #666">
            <strong>注意：</strong>公共行情数据（无需API Key）不使用此代理配置
          </p>
        </div>
      </n-alert>
      <n-form
        ref="formRef"
        :model="formData"
        :rules="rules"
        label-placement="left"
        label-width="140px"
      >
        <n-form-item label="启用代理" path="enabled">
          <n-switch v-model:value="formData.enabled" :checked-value="1" :unchecked-value="0" />
          <span class="form-desc">启用后将通过代理访问交易所API（仅用于API对接平台）</span>
        </n-form-item>

        <n-form-item label="代理类型" path="proxyType">
          <n-select
            v-model:value="formData.proxyType"
            :options="proxyTypeOptions"
            style="width: 200px"
            :disabled="formData.enabled === 0"
          />
          <span class="form-desc">选择代理协议类型</span>
        </n-form-item>

        <n-form-item label="代理地址" path="proxyAddress">
          <n-input
            v-model:value="formData.proxyAddress"
            placeholder="127.0.0.1:33211"
            style="width: 300px"
            :disabled="formData.enabled === 0"
          />
          <span class="form-desc">格式：IP:端口，HTTP示例：127.0.0.1:33210（已验证可用），SOCKS5示例：127.0.0.1:33211</span>
        </n-form-item>

        <n-form-item label="需要认证" path="authEnabled">
          <n-switch
            v-model:value="formData.authEnabled"
            :checked-value="1"
            :unchecked-value="0"
            :disabled="formData.enabled === 0"
          />
          <span class="form-desc">代理服务器是否需要用户名密码认证</span>
        </n-form-item>

        <template v-if="formData.authEnabled === 1">
          <n-form-item label="用户名" path="username">
            <n-input
              v-model:value="formData.username"
              placeholder="请输入用户名"
              style="width: 300px"
              :disabled="formData.enabled === 0"
            />
          </n-form-item>

          <n-form-item label="密码" path="password">
            <n-input
              v-model:value="formData.password"
              type="password"
              placeholder="请输入密码（留空则不修改）"
              show-password-on="click"
              style="width: 300px"
              :disabled="formData.enabled === 0"
            />
            <span class="form-desc">留空则不修改现有密码</span>
          </n-form-item>
        </template>

        <n-form-item>
          <n-space>
            <n-button type="primary" @click="handleSave" :loading="saveLoading">保存配置</n-button>
            <n-button @click="handleTest" :loading="testLoading" :disabled="formData.enabled === 0">
              <template #icon>
                <n-icon><ApiOutlined /></n-icon>
              </template>
              测试连接
            </n-button>
          </n-space>
        </n-form-item>
      </n-form>
    </n-card>

    <!-- 测试结果 -->
    <n-card v-if="testResult" title="测试结果" :bordered="false">
      <n-result
        :status="testResult.success ? 'success' : 'error'"
        :title="testResult.success ? '连接成功' : '连接失败'"
        :description="testResult.message"
      >
        <template v-if="testResult.success && testResult.externalIp">
          <n-descriptions label-placement="left" bordered :column="1" style="margin-top: 20px">
            <n-descriptions-item label="外网IP">
              {{ testResult.externalIp }}
            </n-descriptions-item>
            <n-descriptions-item label="延迟">
              {{ testResult.latency }}ms
            </n-descriptions-item>
          </n-descriptions>
        </template>
      </n-result>
    </n-card>

    <!-- 配置说明 -->
    <n-card title="配置说明" :bordered="false" style="margin-top: 16px">
      <n-alert type="info" :bordered="false">
        <n-ul style="margin: 0; padding-left: 20px">
          <li><strong>全局配置</strong>：此代理配置为项目全局配置，所有用户共享</li>
          <li><strong>使用场景</strong>：代理配置仅用于API对接交易所平台（如Bitget、Binance、OKX等）</li>
          <li><strong>适用场景</strong>：用户API配置测试、交易机器人执行、订单查询等需要API Key的操作</li>
          <li><strong>不适用场景</strong>：公共行情数据（无需API Key）不使用此代理配置</li>
          <li>如果您的服务器无法直接访问交易所API，请配置代理</li>
          <li>支持 SOCKS5 和 HTTP 两种代理类型</li>
          <li><strong>推荐使用HTTP代理</strong>：HTTP代理（127.0.0.1:33210）已验证可用</li>
          <li><strong>注意</strong>：如果SOCKS5代理测试失败，提示"可能不是SOCKS5代理"，说明该端口可能是HTTP代理，请改用HTTP代理类型</li>
          <li>代理地址格式：IP:端口，HTTP示例：127.0.0.1:33210（已验证可用），SOCKS5示例：127.0.0.1:33211（如果测试失败，请检查代理类型）</li>
          <li>如果代理服务器需要认证，请开启"需要认证"并填写用户名和密码</li>
          <li>保存配置后，系统会自动使用代理访问交易所API</li>
          <li>建议先测试连接，确认代理配置正确后再启用</li>
        </n-ul>
      </n-alert>
    </n-card>
  </div>
</template>

<script lang="ts" setup>
import { ref, reactive, onMounted } from 'vue';
import { useMessage } from 'naive-ui';
import { ApiOutlined } from '@vicons/antd';
import { http } from '@/utils/http/axios';

const message = useMessage();
const emit = defineEmits(['save']);

const formRef = ref<any>(null);
const saveLoading = ref(false);
const testLoading = ref(false);
const testResult = ref<any>(null);

// 代理类型选项
const proxyTypeOptions = [
  { label: 'SOCKS5', value: 'socks5' },
  { label: 'HTTP', value: 'http' },
];

// 表单数据
const formData = reactive({
  enabled: 0,
  proxyType: 'http', // HTTP代理已验证可用，使用HTTP作为默认
  proxyAddress: '127.0.0.1:33210', // HTTP代理端口（已验证可用）
  authEnabled: 0,
  username: '',
  password: '',
});

// 表单验证规则
const rules = {
  proxyAddress: {
    required: true,
    message: '请输入代理地址',
    trigger: 'blur',
  },
  proxyType: {
    required: true,
    message: '请选择代理类型',
    trigger: 'change',
  },
  username: {
    validator: (rule: any, value: string) => {
      if (formData.authEnabled === 1 && !value) {
        return new Error('启用认证时，用户名不能为空');
      }
      return true;
    },
    trigger: 'blur',
  },
};

// 加载配置
const loadConfig = async () => {
  try {
    const res = await http.request({
      url: '/trading/proxyConfig/get',
      method: 'get',
    });

    console.log('加载代理配置响应:', res);

    // http.request 在成功时返回 data（即 TradingProxyConfigModel），失败时抛出异常
    // res 直接就是配置数据对象
    const configData = res;
    
    if (configData) {
      // 确保 enabled 正确设置（1 或 0，不使用 || 0 避免将 1 转为 0）
      // 注意：enabled 可能是 0 或 1，需要正确判断
      if (configData.enabled !== undefined && configData.enabled !== null) {
        formData.enabled = configData.enabled;
      } else {
        formData.enabled = 0;
      }
      
      formData.proxyType = configData.proxyType || 'http';
      formData.proxyAddress = configData.proxyAddress || '127.0.0.1:33210';
      
      if (configData.authEnabled !== undefined && configData.authEnabled !== null) {
        formData.authEnabled = configData.authEnabled;
      } else {
        formData.authEnabled = 0;
      }
      
      formData.username = configData.username || '';
      // 密码不返回，保持为空
      formData.password = '';
      
      console.log('代理配置已加载:', {
        enabled: formData.enabled,
        proxyType: formData.proxyType,
        proxyAddress: formData.proxyAddress,
        authEnabled: formData.authEnabled,
      });
    }
  } catch (error) {
    console.error('加载代理配置失败', error);
    // 加载失败时，保持当前表单数据不变，不重置为默认值
  }
};

// 保存配置
const handleSave = async () => {
  try {
    await formRef.value?.validate();

    saveLoading.value = true;

    const data = {
      enabled: formData.enabled,
      proxyType: formData.proxyType,
      proxyAddress: formData.proxyAddress,
      authEnabled: formData.authEnabled,
      username: formData.username,
      password: formData.password, // 如果为空，后端不会更新密码
    };

    const res = await http.request({
      url: '/trading/proxyConfig/save',
      method: 'post',
      data,
    });

    console.log('保存代理配置响应:', res);

    // http.request在成功时返回data，失败时抛出异常
    // 如果能执行到这里，说明请求成功
    message.success('保存成功');
    emit('save');
    formData.password = '';
    
    // 延迟一下再重新加载配置，确保数据库已更新
    setTimeout(() => {
      loadConfig();
    }, 100);
  } catch (error: any) {
    console.error('保存代理配置异常:', error);
    console.error('错误详情:', JSON.stringify(error, null, 2));
    
    let errorMsg = '保存失败';
    
    // 尝试从不同位置提取错误信息
    if (error?.message && error.message !== '请求失败') {
      errorMsg = error.message;
    } else if (error?.response?.data?.message) {
      errorMsg = error.response.data.message;
    } else if (error?.response?.data?.msg) {
      errorMsg = error.response.data.msg;
    } else if (error?.response?.data?.error) {
      errorMsg = error.response.data.error;
    } else if (error?.code) {
      errorMsg = `请求失败 (${error.code})`;
    } else if (typeof error === 'string') {
      errorMsg = error;
    } else if (error?.toString && error.toString() !== '[object Object]') {
      errorMsg = error.toString();
    }
    
    // 如果错误信息为空或只是默认值，显示更详细的提示
    if (!errorMsg || errorMsg === '保存失败' || errorMsg === '请求失败') {
      if (error?.response?.status === 404) {
        errorMsg = '接口不存在，请检查后端服务是否正常运行';
      } else if (error?.response?.status === 500) {
        errorMsg = '服务器内部错误，请查看后端日志';
      } else if (error?.code === 'ECONNABORTED') {
        errorMsg = '请求超时，请检查网络连接';
      } else if (!error?.response) {
        errorMsg = '网络请求失败，请检查网络连接或后端服务状态';
      } else {
        errorMsg = `保存失败 (状态码: ${error?.response?.status || '未知'})`;
      }
    }
    
    message.error(errorMsg);
  } finally {
    saveLoading.value = false;
  }
};

// 测试连接
const handleTest = async () => {
  try {
    await formRef.value?.validate();

    if (formData.enabled === 0) {
      message.warning('请先启用代理');
      return;
    }

    testLoading.value = true;
    testResult.value = null;

    const data = {
      proxyType: formData.proxyType,
      proxyAddress: formData.proxyAddress,
      authEnabled: formData.authEnabled,
      username: formData.username,
      password: formData.password,
    };

    const res = await http.request({
      url: '/trading/proxyConfig/test',
      method: 'post',
      data,
    });

    if (res.code === 0) {
      testResult.value = {
        success: res.success,
        message: res.message,
        externalIp: res.externalIp,
        latency: res.latency,
      };

      if (res.success) {
        message.success('连接测试成功');
      } else {
        message.error('连接测试失败：' + res.message);
      }
    } else {
      message.error(res.message || '测试失败');
    }
  } catch (error: any) {
    testResult.value = {
      success: false,
      message: error.message || '测试失败',
    };
    message.error(error.message || '测试失败');
  } finally {
    testLoading.value = false;
  }
};

onMounted(() => {
  loadConfig();
});
</script>

<style scoped lang="less">
.proxy-config {
  .form-desc {
    margin-left: 12px;
    color: #888;
    font-size: 12px;
  }

  :deep(.n-form-item) {
    margin-bottom: 24px;
  }
}
</style>

