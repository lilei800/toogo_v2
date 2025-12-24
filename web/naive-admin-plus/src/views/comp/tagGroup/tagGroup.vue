<template>
  <PageWrapper
    title="标签组"
    content="BasicTag 基于 Tag 组件扩展，在你写 render 的时候，能够让你少写一些代码"
    showFooter
  >
    <n-card :bordered="false" class="mt-3" title="基本使用">
      <BasicTag :options="basicOptions" @close="handleClose" closable />
    </n-card>

    <n-card :bordered="false" class="mt-3" title="自定义：前缀 / 后缀 / Icon">
      <BasicTag :options="options" />
    </n-card>

    <n-card :bordered="false" class="mt-3" title="可选择Tag">
      <BasicTag :options="checkableOptions" checkable @update:checked="handleChange" />
    </n-card>
  </PageWrapper>
</template>

<script lang="ts" setup>
  import { ref, h } from 'vue';
  import { BasicTag } from '@/components/BasicTag/index';
  import { LeftOutlined, RightOutlined, FlagOutlined } from '@vicons/antd';
  import { TagOptions } from '@/components/BasicTag/src/props';
  import { NIcon } from 'naive-ui';

  const basicOptions = ref<TagOptions[]>([
    {
      label: '预约中',
      type: 'info',
      key: 1,
    },
    {
      label: '已到店',
      type: 'warning',
      key: 2,
    },
    {
      label: '已服务',
      type: 'success',
      key: 3,
      closable: true,
    },
  ]);

  const checkableOptions = ref<TagOptions[]>([
    {
      label: '预约中',
      key: 1,
      checked: false,
    },
    {
      label: '已到店',
      key: 2,
      checked: true,
    },
    {
      label: '已服务',
      key: 3,
      checked: false,
    },
  ]);

  const options = ref<TagOptions[]>([
    {
      label: '预约中',
      type: 'info',
      key: 1,
      prefix: h(NIcon, { component: LeftOutlined }),
    },
    {
      label: '已到店',
      type: 'warning',
      icon: h(NIcon, { component: FlagOutlined, class: 'inline-block' }),
      key: 2,
    },
    {
      label: '已服务',
      type: 'success',
      suffix: h(NIcon, { component: RightOutlined }),
      key: 3,
    },
  ]);

  function handleClose(_: MouseEvent, option: TagOptions, index: number) {
    console.log('index', index);
    console.log('option', option);
    basicOptions.value.splice(index, 1);
  }

  function handleChange(value: boolean, option: TagOptions, index) {
    console.log('option', option);
    checkableOptions.value[index].checked = value;
  }
</script>
