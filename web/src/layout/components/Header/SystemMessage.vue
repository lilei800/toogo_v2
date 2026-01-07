<template>
  <n-tabs v-model:value="currentTab" type="line" justify-content="space-evenly">
    <n-tab-pane v-for="(item, index) in visibleMessages" :key="item.key" :name="index">
      <template #tab>
        <div>
          <span>{{ item.name }}</span>
          <n-badge
            v-bind="item.badgeProps"
            :value="item.list.filter((message) => !message.isRead).length"
            :max="99"
            show-zero
          />
        </div>
      </template>
      <n-spin :show="loading">
        <n-empty v-show="item.list.length === 0" description="无数据" :show-icon="false">
          <template #extra>
            <n-button
              v-if="visibleMessages[currentTab]?.key === 9"
              size="small"
              type="primary"
              @click="goSupportChat"
            >
              联系客服
            </n-button>
            <n-button v-else size="small" @click="handleLoadMore"> 查看更多</n-button>
          </template>
        </n-empty>

        <message-list :list="item.list" @read="handleRead" />
      </n-spin>
    </n-tab-pane>
  </n-tabs>
  <n-space v-if="showAction" justify="center" size="large" class="flex border-t">
    <n-button class="act-btn" size="small" @click="handleClear">清空</n-button>
    <n-button class="act-btn" size="small" @click="handleAllRead">全部已读</n-button>
    <n-button
      v-if="visibleMessages[currentTab]?.key === 9"
      class="act-btn"
      size="small"
      type="primary"
      @click="goSupportChat"
    >
      去对话
    </n-button>
    <n-button class="act-btn" size="small" @click="handleLoadMore">查看更多</n-button>
  </n-space>
</template>

<script lang="ts" setup>
  import { computed, ref, watch } from 'vue';
  import MessageList from './MessageList.vue';
  import { notificationStoreWidthOut } from '@/store/modules/notification';
  import { ReadAll, UpRead } from '@/api/apply/notice';
  import { useRouter } from 'vue-router';

  interface Props {
    /** 只显示这些 tab key（例如：[1,2,3]） */
    onlyKeys?: number[];
    /** 排除这些 tab key（例如：[9]） */
    excludeKeys?: number[];
  }

  const props = withDefaults(defineProps<Props>(), {
    onlyKeys: undefined,
    excludeKeys: undefined,
  });

  const router = useRouter();
  const notificationStore = notificationStoreWidthOut();
  const loading = ref(false);
  const currentTab = ref(0);

  const visibleMessages = computed(() => {
    let list = notificationStore.getMessages;
    if (props.onlyKeys && props.onlyKeys.length > 0) {
      list = list.filter((m) => props.onlyKeys!.includes(m.key));
    }
    if (props.excludeKeys && props.excludeKeys.length > 0) {
      list = list.filter((m) => !props.excludeKeys!.includes(m.key));
    }
    return list;
  });

  watch(
    () => visibleMessages.value.length,
    (len) => {
      if (len <= 0) {
        currentTab.value = 0;
        return;
      }
      if (currentTab.value >= len) {
        currentTab.value = 0;
      }
    },
    { immediate: true },
  );

  const showAction = computed(() => visibleMessages.value[currentTab.value]?.list.length > 0);

  function handleRead(index: number) {
    loading.value = true;
    const message = visibleMessages.value[currentTab.value].list[index];
    const wasUnread = !message.isRead;
    UpRead({ id: message.id })
      .then(() => {
        message.isRead = true;
        if (wasUnread) {
          switch (message.type) {
            case 1:
              notificationStore.notifyUnread--;
              break;
            case 2:
              notificationStore.noticeUnread--;
              break;
            case 3:
              notificationStore.letterUnread--;
              break;
            case 9:
              notificationStore.customerServiceUnread--;
              break;
          }
        }
        if (message.type === 9) {
          goSupportChat();
        }
      })
      .finally(() => {
        loading.value = false;
      });
  }

  function handleAllRead() {
    loading.value = true;
    ReadAll({ type: visibleMessages.value[currentTab.value].key })
      .then(() => {
        visibleMessages.value[currentTab.value].list.forEach((item) =>
          Object.assign(item, { isRead: true }),
        );
        switch (visibleMessages.value[currentTab.value].key) {
          case 1:
            notificationStore.notifyUnread = 0;
            break;
          case 2:
            notificationStore.noticeUnread = 0;
            break;
          case 3:
            notificationStore.letterUnread = 0;
            break;
          case 9:
            notificationStore.customerServiceUnread = 0;
            break;
        }
      })
      .finally(() => {
        loading.value = false;
      });
  }

  function handleClear() {
    visibleMessages.value[currentTab.value].list = [];
    switch (visibleMessages.value[currentTab.value].key) {
      case 1:
        notificationStore.notifyUnread = 0;
        break;
      case 2:
        notificationStore.noticeUnread = 0;
        break;
      case 3:
        notificationStore.letterUnread = 0;
        break;
      case 9:
        notificationStore.customerServiceUnread = 0;
        break;
    }
  }

  function goSupportChat() {
    router.push({ name: 'SupportChatClient' });
  }

  function handleLoadMore() {
    if (visibleMessages.value[currentTab.value].key === 9) {
      goSupportChat();
      return;
    }
    router.push({
      name: 'home_message',
      query: {
        type: visibleMessages.value[currentTab.value].key,
      },
    });
  }
</script>
<style scoped>
  .act-btn {
    margin-top: 8px;
  }
</style>
