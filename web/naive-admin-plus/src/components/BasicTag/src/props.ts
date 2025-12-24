import { tagProps } from 'naive-ui';
import { VNode, PropType } from 'vue';

/**
 * 定义标签的类型
 */
export type TagType = 'default' | 'primary' | 'success' | 'info' | 'warning' | 'error';

/**
 * 定义一个 TagOptions 接口
 * @label {string} 标签 - 用于标识信息项的名称或标题
 * @key {string | number} 键 - 信息项的唯一标识符，可以是字符串或数字
 */
export interface TagOptions {
  label: string | VNode;
  type?: TagType;
  color?: { color?: string; borderColor?: string; textColor?: string };
  key: string | number;
  icon?: VNode;
  prefix?: VNode;
  suffix?: VNode;
  closable?: boolean;
  checked?: boolean;
  checkable?: boolean;
}

/**
 * 定义基础的标签属性
 * @property {...tagProps} - 继承自 naiveui 的标签属性
 * @property {String} size - 标签尺寸，可选值为'small'、'medium'、'large'、'tiny'，默认为'small'
 * @property {[String,Number]} value - 标签的值，可以是字符串或数字，默认为undefined
 * @property {Array} mapGather - 关联数据集合，是一个 TagOptions 对象数组，默认为空对象
 */
export const basicProps = {
  ...tagProps,
  options: {
    type: Array as PropType<TagOptions[]>,
    default: () => [],
  },
  color: {
    type: Object as PropType<{ color?: string; borderColor?: string; textColor?: string }>,
  },
  type: {
    type: String as PropType<TagType>,
    default: 'default',
  },
  closable: {
    type: Boolean,
    default: false,
  },
  checkable: {
    type: Boolean,
    default: false,
  },
  click: {
    type: Boolean,
    default: false,
  },
};
