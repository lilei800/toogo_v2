<template>
  <div class="tableAction">
    <n-flex :size="getFlexSize" align="center" justify="center">
      <template v-for="(action, index) in getActions" :key="action.label">
        <template v-if="!action.isConfirm">
          <n-button v-bind="action">
            {{ action.label }}
            <template #icon v-if="action.icon">
              <component :is="action.icon" />
            </template>
          </n-button>
        </template>
        <template v-else>
          <n-popconfirm v-bind="action">
            <template #trigger>
              <n-button v-bind="action">
                {{ action.label }}
                <template #icon v-if="action.icon">
                  <component :is="action.icon" />
                </template>
              </n-button>
            </template>
            {{ action.confirmContent }}
          </n-popconfirm>
        </template>
        <n-divider vertical v-if="style !== 'button' && index < getActions.length - 1" />
      </template>
      <n-divider
        vertical
        v-if="style !== 'button' && getActions.length && dropDownActions && getDropdownList.length"
      />
      <n-dropdown
        v-if="dropDownActions && getDropdownList.length"
        trigger="hover"
        :options="getDropdownList"
        @select="select"
      >
        <slot name="more"></slot>
        <n-button icon-placement="right" v-bind="getMoreProps" v-if="!$slots.more">
          <template #icon v-if="getMoreProps.icon">
            <component :is="getMoreProps.icon" />
          </template>
          <div class="flex items-center">
            <span>{{ dropDownProps.label }}</span>
            <n-icon size="14" class="ml-1" v-if="!getMoreProps.icon">
              <DownOutlined />
            </n-icon>
          </div>
        </n-button>
      </n-dropdown>
    </n-flex>
  </div>
</template>

<script lang="ts" setup>
  import { PropType, computed, toRaw } from 'vue';
  import { ActionItem } from '@/components/Table';
  import { usePermission } from '@/hooks/web/usePermission';
  import { isBoolean, isFunction } from '@/utils/is';
  import { DownOutlined } from '@vicons/antd';
  import { SelectProps } from 'naive-ui';

  const { hasPermission } = usePermission();

  const props = defineProps({
    actions: {
      type: Array as PropType<ActionItem[]>,
      default: null,
      required: true,
    },
    dropDownActions: {
      type: Array as PropType<ActionItem[]>,
      default: null,
    },
    style: {
      type: String as PropType<String>,
      default: 'button',
    },
    gap: {
      type: Number as PropType<Number>,
      default: 0,
    },
    select: {
      type: Function as PropType<SelectProps>,
      default: () => {},
    },
    dropDownProps: {
      type: Object as PropType<ActionItem>,
      default: () => {
        return {
          label: '更多',
        };
      },
    },
  });

  const actionType =
    props.style === 'button' ? 'default' : props.style === 'text' ? 'primary' : 'default';
  const actionText =
    props.style === 'button' ? undefined : props.style === 'text' ? true : undefined;

  const getMoreProps: any = computed(() => {
    const { dropDownProps } = props;
    return {
      text: actionText,
      type: actionType,
      size: 'small',
      ...dropDownProps,
    };
  });

  const getFlexSize = computed(() => {
    return props.style === 'text' ? [props.gap, 0] : [5, props.gap];
  });

  const getDropdownList = computed(() => {
    return (toRaw(props.dropDownActions) || [])
      .filter((action) => {
        return hasPermission(action.auth as string[]) && isIfShow(action);
      })
      .map((action) => {
        const { popConfirm } = action;
        return {
          size: 'small',
          text: actionText,
          type: actionType,
          ...action,
          ...popConfirm,
          onConfirm: popConfirm?.confirm,
          onCancel: popConfirm?.cancel,
        } as any;
      });
  });

  function isIfShow(action: ActionItem): boolean {
    const ifShow = action.ifShow;

    let isIfShow = true;

    if (isBoolean(ifShow)) {
      isIfShow = ifShow;
    }
    if (isFunction(ifShow)) {
      isIfShow = ifShow(action);
    }
    return isIfShow;
  }

  const getActions = computed(() => {
    return (toRaw(props.actions) || [])
      .filter((action) => {
        return hasPermission(action.auth as string[]) && isIfShow(action);
      })
      .map((action: any) => {
        const { popConfirm } = action;
        //需要展示什么风格，自己修改一下参数
        return {
          size: 'small',
          text: actionText,
          type: actionType,
          ...action,
          ...(popConfirm || {}),
          onConfirm: popConfirm?.confirm,
          onCancel: popConfirm?.cancel,
          enable: !!popConfirm,
        };
      });
  });
</script>
