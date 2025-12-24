import type { ExtractPublicPropTypes, PropType } from 'vue';
import { cardProps, collapseTransitionProps } from 'naive-ui';

export const cardExtendProps = {
  // 标题显示分割条
  showDivider: {
    type: Boolean,
    default: false,
  },
  // title 提示文字
  tooltip: {
    type: String as PropType<string>,
  },
  // 触发展开的区域
  triggers: {
    type: Array as PropType<Array<'main' | 'arrow'>>,
    default: () => ['main', 'arrow'],
  },
  // 是否显示展开收起
  showCollapse: {
    type: Boolean,
    default: true,
  },
  // 展开文字
  spreadText: {
    type: String,
    default: '展开',
  },
  // 收起文字
  foldText: {
    type: String,
    default: '收起',
  },
} as const;

export const basicCardProps = {
  ...cardProps,
  ...cardExtendProps,
  ...collapseTransitionProps,
} as const;

export type BasicCardProps = ExtractPublicPropTypes<typeof basicCardProps>;
export type CardExtendProps = ExtractPublicPropTypes<typeof cardExtendProps>;
