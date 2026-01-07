<template>
  <div>
    <div class="n-layout-page-header">
      <n-card :bordered="false" title="客服工作台">
        一对一会话：客户 ↔ 官方客服。未接线阶段不会泄露客户消息内容给其他客服。
      </n-card>
    </div>

    <n-grid :cols="24" :x-gap="12" :y-gap="12">
      <!-- 左：会话列表 -->
      <n-grid-item :span="7">
        <n-card :bordered="false" class="proCard" size="small">
          <div class="toolbar">
            <n-space align="center">
              <n-switch v-model:value="online" @update:value="onToggleOnline" size="small" />
              <n-text depth="3">{{ online ? '在线' : '离线' }}</n-text>
            </n-space>
            <n-space>
              <n-button size="small" type="primary" @click="handleAcceptNext" :disabled="!online">
                下一单
              </n-button>
              <n-button size="small" @click="loadSessions">刷新</n-button>
            </n-space>
          </div>

          <n-tabs type="segment" size="small" v-model:value="statusTab" @update:value="onChangeTab">
            <n-tab name="1">待接</n-tab>
            <n-tab name="2">进行中</n-tab>
            <n-tab name="3">已关闭</n-tab>
          </n-tabs>

          <n-scrollbar style="height: 620px; margin-top: 10px">
            <n-list clickable>
              <n-list-item
                v-for="s in sessions"
                :key="s.id"
                @click="selectSession(s)"
                :class="{ active: currentSession?.id === s.id }"
              >
                <div class="session-item">
                  <div class="session-title">
                    <n-ellipsis :line-clamp="1">会话 #{{ s.id }} · 用户 {{ s.userId }}</n-ellipsis>
                    <n-badge
                      v-if="(s.unreadAgent || 0) > 0"
                      :value="s.unreadAgent"
                      type="error"
                      :max="99"
                    />
                  </div>
                  <n-text depth="3" class="session-sub">
                    <n-ellipsis :line-clamp="1">
                      {{ s.lastMsgAt ? `${s.lastMsgAt} · ` : '' }}{{ s.lastMsg || '' }}
                    </n-ellipsis>
                  </n-text>
                </div>
              </n-list-item>
            </n-list>
          </n-scrollbar>
        </n-card>
      </n-grid-item>

      <!-- 右：聊天窗口 -->
      <n-grid-item :span="17">
        <n-card :bordered="false" class="proCard" size="small">
          <template #header>
            <n-space align="center" justify="space-between" style="width: 100%">
              <n-space align="center">
                <n-text strong>当前会话：</n-text>
                <n-text v-if="currentSession"
                  >#{{ currentSession.id }}（用户 {{ currentSession.userId }}）</n-text
                >
                <n-text v-else depth="3">请选择左侧会话</n-text>
                <n-tag
                  v-if="currentSession"
                  size="small"
                  :type="statusTagType(currentSession.status)"
                >
                  {{ statusLabel(currentSession.status) }}
                </n-tag>
              </n-space>
              <n-space>
                <n-button size="small" @click="openCannedDrawer" :disabled="!currentSession"
                  >常用语</n-button
                >
                <n-button size="small" @click="openTransfer" :disabled="!canTransfer"
                  >转接</n-button
                >
                <n-button size="small" type="error" @click="handleClose" :disabled="!canClose"
                  >关闭</n-button
                >
              </n-space>
            </n-space>
          </template>

          <n-scrollbar style="height: 520px">
            <div class="chat">
              <div v-for="m in messages" :key="m.id" class="msg" :class="msgClass(m)">
                <div class="bubble">
                  <div class="meta">
                    <n-text depth="3" style="font-size: 12px">
                      {{ msgSenderLabel(m) }} · {{ m.createdAt || '' }}
                    </n-text>
                  </div>
                  <div class="content">{{ m.content }}</div>
                </div>
              </div>
            </div>
          </n-scrollbar>

          <div class="composer">
            <n-input
              v-model:value="draft"
              type="textarea"
              :autosize="{ minRows: 2, maxRows: 5 }"
              placeholder="输入消息..."
              :disabled="!canSend"
            />
            <n-space justify="end" style="margin-top: 10px">
              <n-button size="small" @click="draft = ''" :disabled="!draft">清空</n-button>
              <n-button
                size="small"
                type="primary"
                @click="handleSend"
                :disabled="!canSend || !draftTrim"
              >
                发送
              </n-button>
            </n-space>
          </div>
        </n-card>
      </n-grid-item>
    </n-grid>

    <!-- 常用语 -->
    <n-drawer v-model:show="cannedDrawer" placement="right" :width="420">
      <n-drawer-content title="常用语">
        <n-space justify="space-between" align="center" style="margin-bottom: 10px">
          <n-button size="small" type="primary" @click="openCannedEdit()">新增</n-button>
          <n-button size="small" @click="loadCanned">刷新</n-button>
        </n-space>

        <n-list clickable>
          <n-list-item v-for="c in canned" :key="c.id" @click="useCanned(c)">
            <n-thing>
              <template #header>
                <n-ellipsis :line-clamp="1">{{ c.title }}</n-ellipsis>
              </template>
              <template #description>
                <n-ellipsis :line-clamp="2">{{ c.content }}</n-ellipsis>
              </template>
            </n-thing>
            <template #suffix>
              <n-space>
                <n-button size="tiny" @click.stop="openCannedEdit(c)">编辑</n-button>
                <n-button size="tiny" type="error" @click.stop="deleteCanned(c)">删除</n-button>
              </n-space>
            </template>
          </n-list-item>
        </n-list>
      </n-drawer-content>
    </n-drawer>

    <!-- 常用语编辑 -->
    <n-modal v-model:show="cannedModal">
      <n-card style="width: 520px" title="常用语" :bordered="false" size="small">
        <n-form :model="cannedForm" label-width="80">
          <n-form-item label="标题">
            <n-input v-model:value="cannedForm.title" placeholder="例如：欢迎语" />
          </n-form-item>
          <n-form-item label="内容">
            <n-input
              v-model:value="cannedForm.content"
              type="textarea"
              :autosize="{ minRows: 3, maxRows: 6 }"
              placeholder="常用回复内容"
            />
          </n-form-item>
          <n-form-item label="排序">
            <n-input-number v-model:value="cannedForm.sort" :min="0" />
          </n-form-item>
        </n-form>
        <template #footer>
          <n-space justify="end">
            <n-button size="small" @click="cannedModal = false">取消</n-button>
            <n-button
              size="small"
              type="primary"
              @click="saveCanned"
              :disabled="!cannedForm.title || !cannedForm.content"
            >
              保存
            </n-button>
          </n-space>
        </template>
      </n-card>
    </n-modal>

    <!-- 转接 -->
    <n-modal v-model:show="transferModal">
      <n-card style="width: 520px" title="转接会话" :bordered="false" size="small">
        <n-form label-width="90">
          <n-form-item label="目标客服">
            <n-select
              v-model:value="transferTo"
              :options="agentOptions"
              filterable
              placeholder="选择目标客服"
            />
          </n-form-item>
        </n-form>
        <template #footer>
          <n-space justify="end">
            <n-button size="small" @click="transferModal = false">取消</n-button>
            <n-button
              size="small"
              type="primary"
              @click="handleTransfer"
              :disabled="!transferTo || !currentSession"
            >
              转接
            </n-button>
          </n-space>
        </template>
      </n-card>
    </n-modal>
  </div>
</template>

<script lang="ts" setup>
  import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue';
  import { useDialog, useMessage } from 'naive-ui';
  import {
    addOnMessage,
    removeOnMessage,
    sendMsg,
    WebSocketMessage,
  } from '@/utils/websocket/index';
  import { SocketEnum } from '@/enums/socketEnum';
  import * as SupportApi from '@/api/supportChat';
  import { GetMemberOption } from '@/api/org/user';

  const message = useMessage();
  const dialog = useDialog();

  const online = ref(true);
  const statusTab = ref<'1' | '2' | '3'>('1');

  const sessions = ref<SupportApi.SupportSession[]>([]);
  const currentSession = ref<SupportApi.SupportSession | null>(null);
  const messages = ref<SupportApi.SupportMessage[]>([]);

  const draft = ref('');
  const draftTrim = computed(() => draft.value.trim());

  const cannedDrawer = ref(false);
  const canned = ref<SupportApi.SupportCannedReply[]>([]);
  const cannedModal = ref(false);
  const cannedForm = reactive<Partial<SupportApi.SupportCannedReply>>({
    id: 0,
    title: '',
    content: '',
    sort: 0,
    status: 1,
  });

  const transferModal = ref(false);
  const transferTo = ref<number | null>(null);
  const agentOptions = ref<Array<{ label: string; value: number }>>([]);

  const canSend = computed(() => !!currentSession.value && currentSession.value.status === 2);
  const canClose = computed(
    () => !!currentSession.value && [1, 2].includes(currentSession.value.status),
  );
  const canTransfer = computed(() => !!currentSession.value && currentSession.value.status === 2);

  function statusLabel(s: number) {
    if (s === 1) return '待接';
    if (s === 2) return '进行中';
    return '已关闭';
  }
  function statusTagType(s: number) {
    if (s === 1) return 'warning';
    if (s === 2) return 'success';
    return 'default';
  }

  function msgClass(m: SupportApi.SupportMessage) {
    // 2 = agent
    return m.senderRole === 2 ? 'me' : 'other';
  }
  function msgSenderLabel(m: SupportApi.SupportMessage) {
    if (m.senderRole === 2) return '我(客服)';
    if (m.senderRole === 1) return `用户#${m.senderId}`;
    return '系统';
  }

  async function loadSessions() {
    const res = await SupportApi.SessionList({
      status: Number(statusTab.value),
      page: 1,
      pageSize: 50,
    });
    sessions.value = res?.list || [];
  }

  async function loadMessages(sessionId: number) {
    const res = await SupportApi.MessageList({ sessionId, page: 1, pageSize: 200 });
    const list = res?.list || [];
    // 后端按 id desc，前端按时间正序显示
    messages.value = [...list].reverse();
  }

  async function acceptSession(sessionId: number) {
    const res = await SupportApi.Accept({ sessionId });
    const session = res?.session;
    if (!session) return;
    statusTab.value = '2';
    await loadSessions();
    currentSession.value = session;
    await loadMessages(session.id);
    await loadSessions();
  }

  async function selectSession(s: SupportApi.SupportSession) {
    if (s.status === 1) {
      dialog.warning({
        title: '接线',
        content: `该会话尚未接线，是否接线会话 #${s.id}？`,
        positiveText: '接线',
        negativeText: '取消',
        onPositiveClick: async () => {
          try {
            await acceptSession(s.id);
            message.success('接线成功');
          } catch (e: any) {
            message.error(e?.message || '接线失败');
          }
        },
      });
      return;
    }
    currentSession.value = s;
    try {
      await loadMessages(s.id);
      // 看消息列表会触发后端清零 unread_agent，我们再刷新一遍会话列表拿到最新未读
      await loadSessions();
    } catch (e: any) {
      message.error(e?.message || '加载消息失败');
    }
  }

  async function handleAcceptNext() {
    try {
      const res = await SupportApi.AcceptNext();
      const session = res?.session;
      if (session) {
        statusTab.value = '2';
        await loadSessions();
        currentSession.value = session;
        await loadMessages(session.id);
        await loadSessions();
        message.success('接线成功');
      }
    } catch (e: any) {
      message.error(e?.message || '接线失败');
    }
  }

  async function handleClose() {
    if (!currentSession.value) return;
    dialog.warning({
      title: '关闭会话',
      content: `确定关闭会话 #${currentSession.value.id} 吗？`,
      positiveText: '确定',
      negativeText: '取消',
      onPositiveClick: async () => {
        await SupportApi.Close({ sessionId: currentSession.value!.id });
        message.success('已关闭');
        await loadSessions();
        if (currentSession.value) {
          currentSession.value.status = 3;
        }
      },
    });
  }

  async function handleSend() {
    if (!currentSession.value) return;
    const content = draftTrim.value;
    if (!content) return;
    const res = await SupportApi.Send({ sessionId: currentSession.value.id, content });
    if (res?.message) {
      messages.value.push(res.message);
    }
    draft.value = '';
    await loadSessions();
  }

  function openCannedDrawer() {
    cannedDrawer.value = true;
    loadCanned();
  }

  async function loadCanned() {
    const res = await SupportApi.CannedList({ page: 1, pageSize: 200 });
    canned.value = res?.list || [];
  }

  function useCanned(c: SupportApi.SupportCannedReply) {
    draft.value = c.content || '';
    cannedDrawer.value = false;
  }

  function openCannedEdit(c?: SupportApi.SupportCannedReply) {
    cannedForm.id = c?.id || 0;
    cannedForm.title = c?.title || '';
    cannedForm.content = c?.content || '';
    cannedForm.sort = c?.sort ?? 0;
    cannedForm.status = c?.status ?? 1;
    cannedModal.value = true;
  }

  async function saveCanned() {
    await SupportApi.CannedEdit({
      id: cannedForm.id,
      title: cannedForm.title,
      content: cannedForm.content,
      sort: cannedForm.sort,
      status: 1,
    });
    cannedModal.value = false;
    await loadCanned();
    message.success('已保存');
  }

  function deleteCanned(c: SupportApi.SupportCannedReply) {
    dialog.warning({
      title: '删除常用语',
      content: `确定删除「${c.title}」吗？`,
      positiveText: '确定',
      negativeText: '取消',
      onPositiveClick: async () => {
        await SupportApi.CannedDelete({ id: c.id });
        await loadCanned();
        message.success('已删除');
      },
    });
  }

  async function openTransfer() {
    transferTo.value = null;
    transferModal.value = true;
    await loadAgentOptions();
  }

  async function loadAgentOptions() {
    const res = await GetMemberOption();
    const list = Array.isArray(res) ? res : [];
    agentOptions.value = list.map((m: any) => ({
      value: m.id,
      label: m.realName ? `${m.realName}（${m.username}）` : m.username,
    }));
  }

  async function handleTransfer() {
    if (!currentSession.value || !transferTo.value) return;
    const res = await SupportApi.Transfer({
      sessionId: currentSession.value.id,
      toAgentId: transferTo.value,
    });
    const session = res?.session;
    if (session) {
      message.success('已转接');
      transferModal.value = false;
      // 转接后当前客服不再拥有该会话，刷新列表并清空当前会话
      await loadSessions();
      currentSession.value = null;
      messages.value = [];
    }
  }

  async function onToggleOnline(v: boolean) {
    try {
      await SupportApi.AgentOnline({ online: v });
    } catch (_) {}
  }

  function onChangeTab() {
    currentSession.value = null;
    messages.value = [];
    loadSessions();
  }

  // WS handlers
  const onWsSessionUpdated = (_m: WebSocketMessage) => {
    loadSessions();
  };
  const onWsMessage = (m: WebSocketMessage) => {
    const msg = m.data as SupportApi.SupportMessage;
    if (currentSession.value && msg?.sessionId === currentSession.value.id) {
      messages.value.push(msg);
    } else {
      loadSessions();
    }
  };

  onMounted(async () => {
    await onToggleOnline(true);
    // 加入客服在线组：用于接收“待接队列变化”
    sendMsg(SocketEnum.EventJoin, { id: 'support_agents' });

    addOnMessage(SocketEnum.EventSupportSessionUpdated, onWsSessionUpdated);
    addOnMessage(SocketEnum.EventSupportMessage, onWsMessage);

    await loadSessions();
  });

  onBeforeUnmount(() => {
    removeOnMessage(SocketEnum.EventSupportSessionUpdated);
    removeOnMessage(SocketEnum.EventSupportMessage);
    sendMsg(SocketEnum.EventQuit, { id: 'support_agents' }, false);
    onToggleOnline(false);
  });
</script>

<style scoped lang="less">
  .toolbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 10px;
  }

  :deep(.n-list-item.active) {
    background: rgba(24, 160, 88, 0.08);
  }

  .session-item {
    width: 100%;
  }
  .session-title {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 10px;
  }
  .session-sub {
    display: block;
    margin-top: 6px;
  }

  .chat {
    padding: 8px 4px;
  }
  .msg {
    display: flex;
    margin: 10px 0;
  }
  .msg.me {
    justify-content: flex-end;
  }
  .msg.other {
    justify-content: flex-start;
  }
  .bubble {
    max-width: 70%;
    padding: 10px 12px;
    border-radius: 10px;
    background: rgba(0, 0, 0, 0.04);
  }
  .msg.me .bubble {
    background: rgba(24, 160, 88, 0.12);
  }
  .meta {
    margin-bottom: 6px;
  }
  .content {
    white-space: pre-wrap;
    word-break: break-word;
  }

  .composer {
    margin-top: 12px;
  }
</style>
