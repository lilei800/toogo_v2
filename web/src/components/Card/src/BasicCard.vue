<template>
  <n-card v-bind="nCardProps">
    <template #header>
      <div
        class="flex items-center"
        :class="{ 'cursor-pointer': triggers.includes('main') }"
        @click="handleExpand('main')"
      >
        <span v-if="title" class="mr-1 card-title" :class="{ 'card-title-divider': showDivider }">
          {{ title }}
        </span>
        <n-tooltip trigger="hover" v-if="tooltip">
          <template #trigger>
            <n-icon size="18" class="text-gray-400 cursor-pointer">
              <QuestionCircleOutlined />
            </n-icon>
          </template>
          {{ tooltip }}
        </n-tooltip>
      </div>
    </template>
    <template #header-extra>
      <div :class="{ 'cursor-pointer': triggers.includes('arrow') }" @click="handleExpand('arrow')">
        <slot name="collapse" v-if="$slots.collapse" :expanded="openCollapse"></slot>
        <n-flex v-if="showCollapse && !$slots.collapse">
          <div class="flex items-center handle-collapse">
            <span class="handle-collapse-text">{{ openCollapse ? foldText : spreadText }}</span>
            <n-icon :size="15">
              <component :is="openCollapse ? UpOutlined : DownOutlined" />
            </n-icon>
          </div>
        </n-flex>
      </div>
    </template>
    <template
      v-for="item in Object.keys($slots).filter((item) => item !== 'default')"
      :key="item"
      #[item]="data"
    >
      <slot v-bind="data || {}" :name="item"></slot>
    </template>
    <n-collapse-transition v-bind="nCollapseProps" :show="openCollapse">
      <slot name="default"></slot>
    </n-collapse-transition>
  </n-card>
</template>

<script lang="ts" setup>
  import { collapseTransitionProps, NCard } from 'naive-ui';
  import { basicCardProps, cardExtendProps } from './props';
  import { QuestionCircleOutlined, DownOutlined, UpOutlined } from '@vicons/antd';
  import { ref, watch } from 'vue';
  import { useOmitProps } from './useOmitProps';

  const props = defineProps({
    ...basicCardProps,
  });

  const nCardProps = useOmitProps(props, {
    ...cardExtendProps,
    ...collapseTransitionProps,
    title: props.title,
  });

  const nCollapseProps = useOmitProps(props, { ...cardExtendProps });

  const openCollapse = ref(false);

  watch(
    () => props.show,
    (value) => {
      openCollapse.value = value ?? true;
    },
    { immediate: true },
  );

  function handleExpand(area: 'main' | 'arrow') {
    const { triggers = [] } = props;
    if (triggers.includes(area)) {
      openCollapse.value = !openCollapse.value;
    }
  }
</script>

<style lang="less" scoped>
  .basic-card-header-trigger,
  .handle-collapse-trigger {
    cursor: pointer;
  }
  .handle-collapse {
    &:hover {
      color: var(--n-color-target);
    }
  }
  .handle-collapse-text {
    margin-right: 5px;
    font-size: var(--n-font-size);
  }
  .card-title {
    position: relative;
  }
  .card-title-divider {
    &::before {
      content: '';
      position: absolute;
      left: -10px;
      top: 50%;
      transform: translateY(-50%);
      width: 4px;
      height: 20px;
      background-color: var(--n-color-target);
      border-radius: 10px;
    }
  }
</style>
