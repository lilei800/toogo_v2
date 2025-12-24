<template>
  <NFlex>
    <component
      :is="NTag"
      v-for="(item, index) in options"
      v-bind="filterItemProps(item)"
      :closable="getClosable(item)"
      :key="`${index}_${id}`"
      :type="item.type ?? type"
      :checkable="checkable || item.checkable"
      @close="handleClose($event, item, index)"
      :checked="item.checked"
      @update:checked="handleChecked($event, item, index)"
      class="flex items-center"
      :class="{ 'cursor-pointer': props.click }"
      @click="handleClick(item, index)"
    >
      <div class="flex items-center">
        <template v-if="item.prefix">
          <component :is="item.prefix" />
        </template>
        <template v-if="item.icon">
          <component :is="item.icon" />
        </template>
        <component v-if="isObject(item.label)" :is="item.label" />
        <template v-else>{{ item.label }}</template>
        <template v-if="item.suffix">
          <component :is="item.suffix" />
        </template>
      </div>
    </component>
  </NFlex>
</template>

<script lang="ts" setup>
  import { NTag, NFlex } from 'naive-ui';
  import { basicProps, TagOptions } from './props';
  import { omit, isObject } from 'lodash-es';
  import { useAttrs, useId } from 'vue';
  import { isUnDef } from '@/utils/is';

  const id = useId();

  const props = defineProps({
    ...basicProps,
  });

  const attrs = useAttrs();

  const emits = defineEmits(['close', 'update:checked', 'click']);

  function getClosable(item: TagOptions) {
    return isUnDef(item.closable) ? item.closable : props.closable;
  }

  function filterItemProps(item: TagOptions) {
    return omit({ item, ...attrs }, ['prefix', 'suffix', 'icon']);
  }

  function handleClose($event: MouseEvent, item: TagOptions, index: number) {
    emits('close', $event, item, index);
  }

  function handleChecked(value: boolean, item: TagOptions, index: number) {
    emits('update:checked', value, item, index);
  }

  function handleClick(item: TagOptions, index: number) {
    emits('click', item, index);
  }
</script>
