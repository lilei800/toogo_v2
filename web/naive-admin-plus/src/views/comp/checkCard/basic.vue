<template>
  <PageWrapper title="可选列表" content="基于 List 组件扩展" showFooter>
    <n-card :bordered="false" class="mt-3 proCard" title="单选" content-style="padding-top: 0;">
      <CheckCard :options="listProps1" @change="checkCardChange" />
    </n-card>
    <n-card :bordered="false" class="mt-3 proCard" title="多选" content-style="padding-top: 0;">
      <CheckCard
        :options="listProps2"
        multiple
        @change="checkCardChange"
        v-model:value="checkValue"
      />
      <div class="mt-3">
        <n-space>
          <n-button @click="checkValue = [1]">选择第一个</n-button>
          <n-button @click="checkValue = [1, 3]">选择全部</n-button>
        </n-space>
      </div>
    </n-card>
    <n-card :bordered="false" class="mt-3 proCard" title="可悬浮" content-style="padding-top: 0;">
      <CheckCard :options="listProps3" multiple hoverable @change="checkCardChange" />
      <div class="mt-3">
        <n-space>
          <n-button @click="add">添加一个</n-button>
        </n-space>
      </div>
    </n-card>
    <n-card
      :bordered="false"
      class="mt-3 proCard"
      title="选项不可用"
      content-style="padding-top: 0;"
    >
      <CheckCard :options="listProps4" multiple hoverable @change="checkCardChange" />
    </n-card>
  </PageWrapper>
</template>

<script lang="ts" setup>
  import { ref, h } from 'vue';
  import { CheckCard, OptionsConfig } from '@/components/CheckCard/index';
  import { NButton } from 'naive-ui';
  import HeaderExtra from './HeaderExtra.vue';

  const options = [
    {
      label: '菜单',
      key: 'menu',
    },
    {
      label: '列表',
      key: 'list',
    },
    {
      label: '表单',
      key: 'form',
    },
  ];

  const listProps1 = ref<OptionsConfig[]>([
    {
      title: '纯前端Pro版本',
      description: '使用Vue3、Vite5.x、Pinia构建、经典且赖看风格',
      value: 1,
      avatar: 'https://assets.naiveadmin.com/images/logo.png',
      headerExtra: h(HeaderExtra, {
        options,
      }),
    },
    {
      title: '纯前端Plus版本',
      description: '使用Vue3、Vite5.x、Pinia构建、新颖卡片间隔简约风',
      value: 2,
      avatar: 'https://assets.naiveadmin.com/images/logo.png',
    },
    {
      title: '前后端Full版本',
      description: '支持配套Java/Php语言可选、前后端配套版本，快速开发业务',
      value: 3,
      avatar: 'https://assets.naiveadmin.com/images/logo.png',
    },
  ]);

  const listProps2 = ref<OptionsConfig[]>([
    {
      title: '纯前端Pro版本',
      description: '使用Vue3、Vite5.x、Pinia构建、经典且赖看风格',
      value: 1,
      avatar: 'https://assets.naiveadmin.com/images/logo.png',
    },
    {
      title: '纯前端Plus版本',
      description: '使用Vue3、Vite5.x、Pinia构建、新颖卡片间隔简约风',
      value: 2,
      avatar: 'https://assets.naiveadmin.com/images/logo.png',
    },
    {
      title: '前后端Full版本',
      description: '支持配套Java/Php语言可选、前后端配套版本，快速开发业务',
      value: 3,
      avatar: 'https://assets.naiveadmin.com/images/logo.png',
    },
  ]);

  const listProps3 = ref<OptionsConfig[]>([
    {
      title: '纯前端Pro版本',
      description: '使用Vue3、Vite5.x、Pinia构建、经典且赖看风格',
      value: 1,
      avatar: 'https://assets.naiveadmin.com/images/logo.png',
    },
    {
      title: '纯前端Plus版本',
      description: '使用Vue3、Vite5.x、Pinia构建、新颖卡片间隔简约风',
      value: 2,
      avatar: 'https://assets.naiveadmin.com/images/logo.png',
    },
    {
      title: '前后端Full版本',
      description: '支持配套Java/Php语言可选、前后端配套版本，快速开发业务',
      value: 3,
      avatar: 'https://assets.naiveadmin.com/images/logo.png',
    },
  ]);

  const listProps4 = ref<OptionsConfig[]>([
    {
      title: '纯前端Pro版本',
      description: '使用Vue3、Vite5.x、Pinia构建、经典且赖看风格',
      value: 1,
      avatar: 'https://assets.naiveadmin.com/images/logo.png',
    },
    {
      title: '纯前端Plus版本',
      description: '使用Vue3、Vite5.x、Pinia构建、新颖卡片间隔简约风',
      value: 2,
      avatar: 'https://assets.naiveadmin.com/images/logo.png',
      disabled: true,
    },
    {
      title: '前后端Full版本',
      description: '支持配套Java/Php语言可选、前后端配套版本，快速开发业务',
      value: 3,
      avatar: 'https://assets.naiveadmin.com/images/logo.png',
    },
  ]);

  const checkValue = ref([1, 3]);

  function checkCardChange(value: any) {
    console.log(value);
  }

  function add() {
    const len = listProps3.value.length;
    listProps3.value.push({
      title: `纯前端Plus版本 （新添加-${len}）`,
      description: '使用Vue3、Vite5.x、Pinia构建、新颖卡片间隔简约风',
      value: len + 1,
      avatar: 'https://assets.naiveadmin.com/images/logo.png',
    });
  }
</script>
