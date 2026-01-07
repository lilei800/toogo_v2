<template>
  <div class="password-reset-page">
    <n-card title="重置密码" style="max-width: 420px; margin: 10vh auto">
      <n-alert v-if="!token" type="error" style="margin-top: 12px">
        链接无效或缺少 token，请检查邮件链接是否完整。
      </n-alert>

      <n-alert v-else type="info" style="margin-top: 12px">
        新密码长度 6~18 位。提交成功后将跳转到登录页。
      </n-alert>

      <n-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-placement="left"
        size="large"
        style="margin-top: 16px"
      >
        <n-form-item path="password" label="新密码">
          <n-input
            v-model:value="form.password"
            type="password"
            show-password-on="click"
            placeholder="请输入新密码"
            :disabled="!token"
          />
        </n-form-item>
        <n-form-item path="confirm" label="确认密码">
          <n-input
            v-model:value="form.confirm"
            type="password"
            show-password-on="click"
            placeholder="请再次输入新密码"
            :disabled="!token"
          />
        </n-form-item>
        <n-space vertical :size="12">
          <n-button type="primary" block :loading="loading" :disabled="!token" @click="handleSubmit"
            >提交</n-button
          >
          <n-button block secondary @click="goLogin">返回登录</n-button>
        </n-space>
      </n-form>
    </n-card>
  </div>
</template>

<script setup lang="ts">
  import { computed, ref } from 'vue';
  import { useRoute, useRouter } from 'vue-router';
  import { useMessage } from 'naive-ui';
  import { aesEcb } from '@/utils/encrypt';
  import { PasswordReset } from '@/api/system/user';
  import { ResultEnum } from '@/enums/httpEnum';

  const route = useRoute();
  const router = useRouter();
  const message = useMessage();
  const loading = ref(false);
  const formRef = ref();

  const token = computed(() => (route.query?.token as string) || '');

  const form = ref({
    password: '',
    confirm: '',
  });

  const rules = {
    password: { required: true, message: '请输入新密码', trigger: 'blur' },
    confirm: { required: true, message: '请输入确认密码', trigger: 'blur' },
  };

  function goLogin() {
    router.replace('/login');
  }

  async function handleSubmit(e: Event) {
    e.preventDefault();
    if (!token.value) {
      message.error('链接无效或已过期');
      return;
    }
    formRef.value.validate(async (errors: any) => {
      if (errors) {
        message.error('请填写完整信息');
        return;
      }
      if (form.value.password !== form.value.confirm) {
        message.error('两次输入的密码不一致');
        return;
      }
      loading.value = true;
      message.loading('提交中...');
      try {
        const { code, message: msg } = await PasswordReset({
          token: token.value,
          password: aesEcb.encrypt(form.value.password),
        });
        message.destroyAll();
        if (code === ResultEnum.SUCCESS) {
          message.success('密码已重置，请重新登录');
          router.replace('/login');
        } else {
          message.error(msg || '重置失败');
        }
      } finally {
        loading.value = false;
      }
    });
  }
</script>

<style scoped lang="less">
  .password-reset-page {
    padding: 16px;
  }
</style>
