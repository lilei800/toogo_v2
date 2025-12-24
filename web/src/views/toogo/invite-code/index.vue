<template>
  <div class="invite-code-page">
    <n-space vertical :size="16">
      <!-- 永久邀请码 - 仅高级代理可见 -->
      <n-card title="🔐 永久邀请码" hoverable v-if="toogoUserInfo.agentUnlockLevel === 1">
        <template #header-extra>
          <n-tag type="success" size="small">永久有效</n-tag>
          <n-tag type="warning" size="small" style="margin-left: 4px;">高级代理专属</n-tag>
        </template>
        <n-space vertical :size="16">
          <n-alert type="info" :bordered="false">
            <template #header>永久邀请码说明</template>
            <ul style="margin: 8px 0; padding-left: 20px;">
              <li>格式：4位字母 + 4位数字（不含数字4）</li>
              <li>永久有效，无需刷新</li>
              <li>用于基础用户系统的邀请注册</li>
              <li>可触发 Toogo 推广奖励机制</li>
              <li><n-text type="warning" strong>高级代理专属功能</n-text></li>
            </ul>
          </n-alert>
          
          <n-space align="center" :size="16">
            <n-text strong>我的永久邀请码：</n-text>
            <n-text code style="font-size: 28px; padding: 12px 24px; letter-spacing: 2px;">
              {{ baseUserInfo.inviteCode || '------' }}
            </n-text>
            <n-button type="primary" size="large" @click="copyPermanentCode">
              <template #icon>
                <n-icon><CopyOutline /></n-icon>
              </template>
              复制邀请码
            </n-button>
            <n-button size="large" @click="copyPermanentLink">
              <template #icon>
                <n-icon><CopyOutline /></n-icon>
              </template>
              复制注册链接
            </n-button>
          </n-space>
        </n-space>
      </n-card>

      <!-- 非高级代理提示 -->
      <n-card title="🔐 永久邀请码" hoverable v-else>
        <template #header-extra>
          <n-tag type="warning" size="small">高级代理专属</n-tag>
        </template>
        <n-alert type="warning" :bordered="false">
          <template #header>升级提示</template>
          <n-space vertical :size="12">
            <n-text>永久邀请码是高级代理专属功能，享有以下特权：</n-text>
            <ul style="margin: 8px 0; padding-left: 20px;">
              <li>永久有效的邀请码，无需定期刷新</li>
              <li>更高的推广佣金比例</li>
              <li>专属推广资源和支持</li>
            </ul>
            <n-text depth="3">如需升级为高级代理，请联系客服或满足升级条件。</n-text>
            <n-space>
              <n-button type="primary" @click="$router.push('/toogo/agent')">
                <template #icon>
                  <n-icon><PeopleOutline /></n-icon>
                </template>
                了解代理体系
              </n-button>
              <n-button @click="$router.push('/toogo/subscription')">
                查看订阅套餐
              </n-button>
            </n-space>
          </n-space>
        </n-alert>
      </n-card>

      <!-- 动态邀请码 (Toogo) -->
      <n-card title="⚡ 动态邀请码 (Toogo)" hoverable>
        <template #header-extra>
          <n-tag type="warning" size="small">24小时有效</n-tag>
        </template>
        <n-space vertical :size="16">
          <n-alert type="warning" :bordered="false">
            <template #header>动态邀请码说明</template>
            <ul style="margin: 8px 0; padding-left: 20px;">
              <li>格式：2位字母 + 4位数字（不含数字4）</li>
              <li>有效期：24小时，过期后需刷新</li>
              <li>专用于 Toogo 量化交易系统推广</li>
              <li>支持推广奖励和佣金返利</li>
            </ul>
          </n-alert>

          <n-space align="center" :size="16">
            <n-text strong>我的动态邀请码：</n-text>
            <n-text code style="font-size: 28px; padding: 12px 24px; letter-spacing: 2px;">
              {{ toogoUserInfo.inviteCode || '------' }}
            </n-text>
            <n-button type="primary" size="large" @click="copyDynamicCode">
              <template #icon>
                <n-icon><CopyOutline /></n-icon>
              </template>
              复制邀请码
            </n-button>
            <n-button size="large" @click="copyDynamicLink">
              <template #icon>
                <n-icon><CopyOutline /></n-icon>
              </template>
              复制注册链接
            </n-button>
            <n-button type="info" size="large" @click="refreshDynamicCode" :loading="refreshing">
              <template #icon>
                <n-icon><RefreshOutline /></n-icon>
              </template>
              刷新邀请码
            </n-button>
          </n-space>

          <n-space align="center" v-if="toogoUserInfo.inviteCodeExpire">
            <n-text depth="3">过期时间：</n-text>
            <n-text type="warning">{{ formatExpireTime(toogoUserInfo.inviteCodeExpire) }}</n-text>
          </n-space>
        </n-space>
      </n-card>

      <!-- 推广奖励说明 -->
      <n-card title="🎁 推广奖励说明">
        <n-grid :cols="2" :x-gap="16" :y-gap="16">
          <n-gi>
            <n-card title="注册奖励" size="small" :bordered="false" embedded>
              <n-space vertical :size="8">
                <n-text>• 成功邀请新用户注册</n-text>
                <n-text type="success" strong>双方各获得 30 算力奖励</n-text>
              </n-space>
            </n-card>
          </n-gi>
          <n-gi>
            <n-card title="订阅奖励" size="small" :bordered="false" embedded>
              <n-space vertical :size="8">
                <n-text>• 被邀请人订阅套餐</n-text>
                <n-text type="success" strong>邀请人获得对应算力奖励</n-text>
              </n-space>
            </n-card>
          </n-gi>
        </n-grid>
      </n-card>

      <!-- 快速入口 -->
      <n-card title="📊 推广数据">
        <n-space :size="16">
          <n-button type="primary" @click="$router.push('/toogo/team')">
            <template #icon>
              <n-icon><PeopleOutline /></n-icon>
            </template>
            查看我的团队
          </n-button>
          <n-button type="primary" @click="$router.push('/toogo/commission')">
            <template #icon>
              <n-icon><WalletOutline /></n-icon>
            </template>
            查看佣金明细
          </n-button>
        </n-space>
      </n-card>
    </n-space>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useMessage } from 'naive-ui';
import { CopyOutline, RefreshOutline, PeopleOutline, WalletOutline } from '@vicons/ionicons5';
import { ToogoUserApi } from '@/api/toogo';
import { getUserInfo } from '@/api/system/user';

const message = useMessage();

const baseUserInfo = ref<any>({});
const toogoUserInfo = ref<any>({});
const refreshing = ref(false);

// 加载数据
const loadData = async () => {
  try {
    const [baseRes, toogoRes] = await Promise.all([
      getUserInfo(),
      ToogoUserApi.info(),
    ]);
    baseUserInfo.value = baseRes || {};
    toogoUserInfo.value = toogoRes || {};
  } catch (error) {
    console.error('加载数据失败:', error);
    message.error('加载数据失败');
  }
};

// 格式化过期时间
const formatExpireTime = (time: string) => {
  if (!time) return '';
  const date = new Date(time);
  const now = new Date();
  const diff = date.getTime() - now.getTime();
  
  if (diff < 0) {
    return '已过期';
  }
  
  const hours = Math.floor(diff / (1000 * 60 * 60));
  const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));
  
  return `${date.toLocaleString()} (剩余 ${hours}小时${minutes}分钟)`;
};

// 复制到剪贴板的通用方法
const copyToClipboard = (text: string) => {
  const input = document.createElement('input');
  input.value = text;
  input.style.position = 'fixed';
  input.style.opacity = '0';
  document.body.appendChild(input);
  input.select();
  
  try {
    const success = document.execCommand('copy');
    document.body.removeChild(input);
    return success;
  } catch (err) {
    document.body.removeChild(input);
    return false;
  }
};

// 复制永久邀请码
const copyPermanentCode = () => {
  const code = baseUserInfo.value?.inviteCode || '';
  if (!code) {
    message.error('邀请码为空');
    return;
  }
  
  if (copyToClipboard(code)) {
    message.success('永久邀请码已复制');
  } else {
    message.error('复制失败，请手动复制');
  }
};

// 复制永久邀请码注册链接
const copyPermanentLink = () => {
  const code = baseUserInfo.value?.inviteCode || '';
  if (!code) {
    message.error('邀请码为空');
    return;
  }
  const link = `${window.location.origin}/register?inviteCode=${encodeURIComponent(code)}`;
  
  if (copyToClipboard(link)) {
    message.success('注册链接已复制');
  } else {
    message.error('复制失败，请手动复制');
  }
};

// 复制动态邀请码
const copyDynamicCode = () => {
  const code = toogoUserInfo.value?.inviteCode || '';
  if (!code) {
    message.error('邀请码为空');
    return;
  }
  
  if (copyToClipboard(code)) {
    message.success('动态邀请码已复制');
  } else {
    message.error('复制失败，请手动复制');
  }
};

// 复制动态邀请码注册链接
const copyDynamicLink = () => {
  const code = toogoUserInfo.value?.inviteCode || '';
  if (!code) {
    message.error('邀请码为空');
    return;
  }
  const link = `${window.location.origin}/register?inviteCode=${encodeURIComponent(code)}`;
  
  if (copyToClipboard(link)) {
    message.success('注册链接已复制');
  } else {
    message.error('复制失败，请手动复制');
  }
};

// 刷新动态邀请码
const refreshDynamicCode = async () => {
  refreshing.value = true;
  try {
    const res = await ToogoUserApi.refreshInviteCode();
    toogoUserInfo.value.inviteCode = res?.inviteCode;
    toogoUserInfo.value.inviteCodeExpire = res?.inviteCodeExpire;
    message.success('动态邀请码已刷新');
  } catch (error) {
    message.error('刷新失败');
  } finally {
    refreshing.value = false;
  }
};

onMounted(() => {
  loadData();
});
</script>

<style scoped lang="less">
.invite-code-page {
  padding: 16px;
  max-width: 1200px;
  margin: 0 auto;

  :deep(.n-card__header) {
    font-size: 18px;
    font-weight: 600;
  }

  :deep(.n-alert) {
    ul {
      li {
        margin: 4px 0;
        line-height: 1.8;
      }
    }
  }
}
</style>
