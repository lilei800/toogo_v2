<template>
  <div class="support-chat-container">
    <div class="n-layout-page-header">
      <n-card :bordered="false" title="联系客服">
        <n-space vertical>
          <n-text>一对一在线客服：你 ↔ 官方客服</n-text>
          <n-text depth="3" style="font-size: 13px"> 进入页面会自动创建/复用一个未关闭会话 </n-text>
        </n-space>
      </n-card>
    </div>

    <!-- 错误提示 -->
    <n-alert
      v-if="error"
      type="error"
      :title="errorTitle"
      closable
      @close="error = null"
      style="margin-bottom: 16px"
    >
      {{ error }}
      <template #footer>
        <n-space>
          <n-button size="small" type="primary" @click="reloadAll">重试</n-button>
          <n-text depth="3" style="font-size: 12px">
            提示：请确保后端服务已启动且数据库表已创建
          </n-text>
        </n-space>
      </template>
    </n-alert>

    <n-card :bordered="false" class="proCard chat-card" size="small">
      <template #header>
        <n-space align="center" justify="space-between" style="width: 100%">
          <n-space align="center">
            <n-icon size="20" color="#18a058">
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24">
                <path
                  fill="currentColor"
                  d="M12 12c2.21 0 4-1.79 4-4s-1.79-4-4-4-4 1.79-4 4 1.79 4 4 4zm0 2c-2.67 0-8 1.34-8 4v2h16v-2c0-2.66-5.33-4-8-4z"
                />
              </svg>
            </n-icon>
            <n-text strong style="font-size: 15px">会话</n-text>
            <n-text v-if="session" type="info">#{{ session.id }}</n-text>
            <n-text v-else depth="3">初始化中...</n-text>
            <n-tag
              v-if="session"
              size="small"
              :type="statusTagType(session.status)"
              :bordered="false"
            >
              {{ statusLabel(session.status) }}
            </n-tag>
          </n-space>
          <n-button size="small" :loading="loading" @click="reloadAll" quaternary circle>
            <template #icon>
              <n-icon
                ><svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24">
                  <path
                    fill="currentColor"
                    d="M17.65 6.35A7.958 7.958 0 0 0 12 4c-4.42 0-7.99 3.58-7.99 8s3.57 8 7.99 8c3.73 0 6.84-2.55 7.73-6h-2.08A5.99 5.99 0 0 1 12 18c-3.31 0-6-2.69-6-6s2.69-6 6-6c1.66 0 3.14.69 4.22 1.78L13 11h7V4l-2.35 2.35z"
                  /></svg
              ></n-icon>
            </template>
          </n-button>
        </n-space>
      </template>

      <n-spin :show="loading && !session">
        <n-scrollbar ref="scrollbarRef" style="height: 520px" @scroll="handleScroll">
          <div class="chat" v-if="messages.length > 0">
            <div v-for="m in messages" :key="m.id" class="msg" :class="msgClass(m)">
              <div class="bubble">
                <div class="meta">
                  <n-space align="center" :size="4">
                    <n-text depth="3" style="font-size: 11px">
                      {{ msgSenderLabel(m) }}
                    </n-text>
                    <n-text depth="3" style="font-size: 11px">
                      {{ m.createdAt || '' }}
                    </n-text>
                  </n-space>
                </div>
                <div class="content">{{ m.content }}</div>
              </div>
            </div>
          </div>
          <n-empty v-else description="暂无消息" style="margin-top: 100px">
            <template #icon>
              <n-icon size="48" color="#ccc">
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24">
                  <path
                    fill="currentColor"
                    d="M20 2H4c-1.1 0-2 .9-2 2v18l4-4h14c1.1 0 2-.9 2-2V4c0-1.1-.9-2-2-2zm0 14H6l-2 2V4h16v12z"
                  />
                </svg>
              </n-icon>
            </template>
          </n-empty>
        </n-scrollbar>
      </n-spin>

      <div class="composer">
        <n-input
          v-model:value="draft"
          type="textarea"
          :autosize="{ minRows: 3, maxRows: 6 }"
          placeholder="请输入您的问题..."
          :disabled="!canSend || loading"
          @keydown.enter.ctrl="handleSend"
          @keydown.enter.meta="handleSend"
        />
        <n-space justify="space-between" align="center" style="margin-top: 12px">
          <n-text depth="3" style="font-size: 12px">
            <n-text v-if="canSend">Ctrl+Enter 快速发送</n-text>
            <n-text v-else type="warning">{{ canSendTip }}</n-text>
          </n-text>
          <n-space>
            <n-button size="small" @click="draft = ''" :disabled="!draft" quaternary>清空</n-button>
            <n-button
              size="small"
              type="primary"
              @click="handleSend"
              :disabled="!canSend || !draftTrim || loading"
              :loading="sending"
            >
              <template #icon>
                <n-icon
                  ><svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24">
                    <path fill="currentColor" d="M2.01 21L23 12 2.01 3 2 10l15 2-15 2z" /></svg
                ></n-icon>
              </template>
              发送
            </n-button>
          </n-space>
        </n-space>
      </div>
    </n-card>
  </div>
</template>

<script lang="ts" setup>
  import { computed, onBeforeUnmount, onMounted, ref, nextTick } from 'vue';
  import { useMessage } from 'naive-ui';
  import { addOnMessage, removeOnMessage, WebSocketMessage } from '@/utils/websocket/index';
  import { SocketEnum } from '@/enums/socketEnum';
  import * as ClientSupportApi from '@/api/client/supportChat';

  const msg = useMessage();

  const session = ref<ClientSupportApi.SupportSession | null>(null);
  const messages = ref<ClientSupportApi.SupportMessage[]>([]);
  const draft = ref('');
  const loading = ref(false);
  const sending = ref(false);
  const error = ref<string | null>(null);
  const errorTitle = ref('初始化失败');
  const scrollbarRef = ref();
  const draftTrim = computed(() => draft.value.trim());

  const canSend = computed(() => !!session.value && session.value.status !== 3);
  const canSendTip = computed(() => {
    if (!session.value) return '会话未初始化';
    if (session.value.status === 3) return '会话已结束';
    if (session.value.status === 1) return '等待客服接线中...';
    return '';
  });

  function statusLabel(s: number) {
    if (s === 1) return '等待接线';
    if (s === 2) return '进行中';
    return '已结束';
  }
  function statusTagType(s: number) {
    if (s === 1) return 'warning';
    if (s === 2) return 'success';
    return 'default';
  }

  function msgClass(m: ClientSupportApi.SupportMessage) {
    // 1 = user
    return m.senderRole === 1 ? 'me' : 'other';
  }
  function msgSenderLabel(m: ClientSupportApi.SupportMessage) {
    if (m.senderRole === 1) return '我';
    if (m.senderRole === 2) return '客服';
    return '系统';
  }

  function scrollToBottom() {
    nextTick(() => {
      if (scrollbarRef.value) {
        scrollbarRef.value.scrollTo({
          top: scrollbarRef.value.$el.scrollHeight,
          behavior: 'smooth',
        });
      }
    });
  }

  function handleScroll() {
    // 可以在这里添加滚动加载更多消息的逻辑
  }

  async function loadMessages() {
    if (!session.value) return;
    try {
      const res = await ClientSupportApi.MessageList({
        sessionId: session.value.id,
        page: 1,
        pageSize: 200,
      });
      const list = res?.list || [];
      messages.value = [...list].reverse();
      scrollToBottom();
    } catch (e: any) {
      console.error('加载消息失败:', e);
      msg.warning('加载消息失败：' + (e?.message || '未知错误'));
    }
  }

  async function reloadAll() {
    loading.value = true;
    error.value = null;
    try {
      const res = await ClientSupportApi.Start();
      session.value = res?.session || null;
      await loadMessages();
      msg.success('会话初始化成功');
    } catch (e: any) {
      console.error('初始化会话失败:', e);
      errorTitle.value = '初始化会话失败';

      // 根据错误类型给出不同的提示
      if (e?.message?.includes('404')) {
        error.value =
          '后端接口不存在 (404)。请检查：\n1. 后端服务是否已启动\n2. 数据库表是否已创建（support_session, support_message 等）\n3. 路由是否已注册';
      } else if (e?.message?.includes('401') || e?.message?.includes('登录')) {
        error.value = '未登录或登录已过期，请先登录';
      } else if (e?.message?.includes('500')) {
        error.value = '服务器内部错误，请检查后端日志';
      } else {
        error.value = e?.message || '未知错误，请检查网络连接和后端服务状态';
      }

      session.value = null;
      messages.value = [];
    } finally {
      loading.value = false;
    }
  }

  async function handleSend() {
    if (!session.value) {
      msg.warning('会话未初始化');
      return;
    }
    const content = draftTrim.value;
    if (!content) {
      msg.warning('请输入消息内容');
      return;
    }

    sending.value = true;
    try {
      const res = await ClientSupportApi.Send({ sessionId: session.value.id, content });
      if (res?.message) {
        messages.value.push(res.message);
        draft.value = '';
        scrollToBottom();
      }
    } catch (e: any) {
      console.error('发送消息失败:', e);
      msg.error('发送失败：' + (e?.message || '未知错误'));
    } finally {
      sending.value = false;
    }
  }

  // WS：实时追加消息/更新会话状态
  const onWsMessage = (m: WebSocketMessage) => {
    const data = m.data as ClientSupportApi.SupportMessage;
    if (!session.value || data?.sessionId !== session.value.id) return;

    // 避免重复添加（如果本地已经有这条消息）
    const exists = messages.value.some((msg) => msg.id === data.id);
    if (!exists) {
      messages.value.push(data);
      scrollToBottom();

      // 如果是客服发来的消息，显示通知
      if (data.senderRole === 2) {
        msg.info('客服回复了您的消息', { duration: 3000 });
      }
    }
  };

  const onWsSessionUpdated = (m: WebSocketMessage) => {
    const data = m.data as Partial<ClientSupportApi.SupportSession>;
    if (!session.value || data?.id !== session.value.id) return;

    const oldStatus = session.value.status;
    session.value = { ...session.value, ...(data as any) };

    // 状态变化提示
    if (session.value && oldStatus !== session.value.status) {
      if (session.value.status === 2) {
        msg.success('客服已接线，可以开始对话了', { duration: 3000 });
      } else if (session.value.status === 3) {
        msg.warning('会话已关闭', { duration: 3000 });
      }
    }
  };

  onMounted(async () => {
    await reloadAll();
    addOnMessage(SocketEnum.EventSupportMessage, onWsMessage);
    addOnMessage(SocketEnum.EventSupportSessionUpdated, onWsSessionUpdated);
  });

  onBeforeUnmount(() => {
    removeOnMessage(SocketEnum.EventSupportMessage);
    removeOnMessage(SocketEnum.EventSupportSessionUpdated);
  });
</script>

<style scoped lang="less">
  .support-chat-container {
    padding-bottom: 20px;
  }

  .chat-card {
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  }

  .chat {
    padding: 16px 12px;
    min-height: 400px;
  }

  .msg {
    display: flex;
    margin: 16px 0;
    animation: slideIn 0.3s ease-out;
  }

  @keyframes slideIn {
    from {
      opacity: 0;
      transform: translateY(10px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  .msg.me {
    justify-content: flex-end;
  }

  .msg.other {
    justify-content: flex-start;
  }

  .bubble {
    max-width: 75%;
    padding: 12px 16px;
    border-radius: 12px;
    background: rgba(0, 0, 0, 0.04);
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.06);
    transition: all 0.2s ease;
  }

  .bubble:hover {
    box-shadow: 0 2px 6px rgba(0, 0, 0, 0.1);
  }

  .msg.me .bubble {
    background: linear-gradient(135deg, rgba(24, 160, 88, 0.15) 0%, rgba(24, 160, 88, 0.08) 100%);
    border-bottom-right-radius: 4px;
  }

  .msg.other .bubble {
    background: rgba(0, 0, 0, 0.04);
    border-bottom-left-radius: 4px;
  }

  .meta {
    margin-bottom: 8px;
    opacity: 0.7;
  }

  .content {
    white-space: pre-wrap;
    word-break: break-word;
    line-height: 1.6;
    font-size: 14px;
  }

  .composer {
    margin-top: 16px;
    padding-top: 16px;
    border-top: 1px solid rgba(0, 0, 0, 0.06);
  }

  :deep(.n-input) {
    border-radius: 8px;
  }

  :deep(.n-input__textarea-el) {
    font-size: 14px;
    line-height: 1.6;
  }

  :deep(.n-card-header) {
    padding: 16px 20px;
  }

  :deep(.n-scrollbar-content) {
    padding-right: 8px;
  }
</style>
