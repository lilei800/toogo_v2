<template>
  <div>
    <n-card :bordered="false" title="代理配置">
      <n-form ref="formRef" :model="formValue" label-placement="left" label-width="150">
        <n-form-item label="启用代理">
          <n-switch v-model:value="formValue.enabled" />
        </n-form-item>
        <n-form-item label="代理类型">
          <n-select v-model:value="formValue.proxyType" :options="proxyTypeOptions" />
        </n-form-item>
        <n-form-item label="代理地址">
          <n-input v-model:value="formValue.proxyHost" placeholder="127.0.0.1" />
        </n-form-item>
        <n-form-item label="代理端口">
          <n-input-number v-model:value="formValue.proxyPort" :min="1" :max="65535" style="width: 100%;" />
        </n-form-item>
        <n-form-item label="用户名">
          <n-input v-model:value="formValue.proxyUser" />
        </n-form-item>
        <n-form-item label="密码">
          <n-input v-model:value="formValue.proxyPass" type="password" show-password-on="click" />
        </n-form-item>
        <n-form-item>
          <n-space>
            <n-button type="primary" :loading="saving" @click="handleSave">保存配置</n-button>
            <n-button :loading="testing" @click="handleTest">测试连接</n-button>
          </n-space>
        </n-form-item>
      </n-form>
    </n-card>
  </div>
</template>

<script lang="ts" setup>
import { ref, reactive, onMounted } from 'vue';
import { NCard, NForm, NFormItem, NSwitch, NSelect, NInput, NInputNumber, NSpace, NButton, useMessage } from 'naive-ui';
import { getProxyConfig, saveProxyConfig, testProxyConfig } from '@/api/trading/proxy-config';

const message = useMessage();

const saving = ref(false);
const testing = ref(false);

const formValue = reactive({
  enabled: false,
  proxyType: 'socks5',
  proxyHost: '127.0.0.1',
  proxyPort: 10808,
  proxyUser: '',
  proxyPass: '',
});

const proxyTypeOptions = [
  { label: 'SOCKS5', value: 'socks5' },
  { label: 'HTTP', value: 'http' },
  { label: 'HTTPS', value: 'https' },
];

const loadConfig = async () => {
  try {
    const res = await getProxyConfig();
    if (res) {
      Object.assign(formValue, res);
    }
  } catch (error: any) {
    message.error(error.message || '加载配置失败');
  }
};

const handleSave = async () => {
  saving.value = true;
  try {
    await saveProxyConfig(formValue);
    message.success('保存成功');
  } catch (error: any) {
    message.error(error.message || '保存失败');
  } finally {
    saving.value = false;
  }
};

const handleTest = async () => {
  testing.value = true;
  try {
    await testProxyConfig();
    message.success('代理连接测试成功');
  } catch (error: any) {
    message.error(error.message || '连接测试失败');
  } finally {
    testing.value = false;
  }
};

onMounted(() => {
  loadConfig();
});
</script>

