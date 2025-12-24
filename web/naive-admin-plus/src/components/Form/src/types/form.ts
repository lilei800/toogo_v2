import { ComponentType } from './index';
import type { CSSProperties } from 'vue';
import type { GridItemProps, GridProps } from 'naive-ui/lib/grid';
import type { ButtonProps } from 'naive-ui/lib/button';
import type { CardProps } from 'naive-ui';

export interface RenderReturnParams {
  schema: FormSchema;
  values: Recordable;
  model: Recordable;
  field: string;
}

export interface componentProps {
  options?: any[];
  placeholder?: string;
  showButton?: boolean;
  onInput?: () => Recordable;
  onChange?: () => Recordable;
}

export interface componentSlotsRenderOptions {
  field: string;
  model: Recordable;
  formAction: FormActionType;
}

export type componentSlotsType = (opt: componentSlotsRenderOptions) => JSX.Element;

export interface FormSchema {
  field: string;
  label?: string;
  labelWidth?: number | string | undefined;
  labelMessage?: string;
  labelMessageStyle?: object | string;
  defaultValue?: any;
  component?: ComponentType;
  componentProps?: object | ((opt: { model: Recordable; formAction: FormActionType }) => object);
  componentSlots?: Record<string, componentSlotsType>;
  slot?: string;
  rules?: object | object[];
  giProps?: GridItemProps;
  isFull?: boolean;
  suffix?: string;
  showFeedback?: boolean;
  showLabel?: boolean;
  requireMarkPlacement?: string;
  hidden?: boolean | ((renderCallbackParams: RenderReturnParams) => boolean);
}

export interface FormProps {
  model?: Recordable;
  labelWidth?: number | string | undefined;
  schemas?: FormSchema[];
  group?: FormGroupRow[];
  inline?: boolean;
  layout?: string;
  size?: string;
  labelPlacement?: string;
  isFull?: boolean;
  showActionButtonGroup?: boolean;
  showResetButton?: boolean;
  resetButtonOptions?: Partial<ButtonProps>;
  showSubmitButton?: boolean;
  showAdvancedButton?: boolean;
  submitButtonOptions?: Partial<ButtonProps>;
  submitButtonText?: string;
  resetButtonText?: string;
  gridProps?: GridProps;
  giProps?: GridItemProps;
  resetFunc?: () => Promise<void>;
  submitFunc?: () => Promise<void>;
  submitOnReset?: boolean;
  baseGridStyle?: CSSProperties;
  draggable?: boolean;
  collapsed?: boolean;
  collapsedRows?: number;
  requireMarkPlacement?: string;
  isEnterSubmit?: boolean;
}

export interface FormActionType {
  submit: () => Promise<any>;
  setProps: (formProps: Partial<FormProps>) => Promise<void>;
  setSchema: (schemaProps: Partial<FormSchema[]>) => Promise<void>;
  setGroupSchema: (schemaProps: Partial<FormGroupRow[]>) => Promise<void>;
  setFieldsValue: (values: Recordable) => void;
  clearValidate: (name?: string | string[]) => Promise<void>;
  getFieldsValue: () => Recordable;
  resetFields: () => Promise<void>;
  validate: (
    names?: Function | string[],
    rulesFun?: (rule: { key: string }) => boolean | undefined,
  ) => Promise<any>;
  setLoading: (status: boolean) => void;
  updateSchema: (schemaProps: Partial<FormSchema> | Partial<FormSchema>[]) => Promise<void>;
  /**
   * @deprecated : 此方法存在缺陷，暂不建议使用此方法
   */
  updateGroupSchema: (key: string, propertyPath: string, newValue: any) => Promise<void>;
}

export type CardSlots = {
  cover?: Function;
  headerExtra?: Function;
  header?: Function;
  default?: Function;
  footer?: Function;
  action?: Function;
};

export interface FormGroupRow {
  key?: string;
  title: string;
  cardProps?: CardProps;
  cardSlots?: CardSlots;
  columns: FormSchema[];
}

export type RegisterFn = (formInstance: FormActionType) => void;

export type UseFormReturnType = [RegisterFn, FormActionType];
