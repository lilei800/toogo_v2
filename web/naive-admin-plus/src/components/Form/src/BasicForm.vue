<template>
  <!-- @ts-expect-error Type Error! -->
  <DefineTemplate v-slot="{ data }">
    <n-grid v-bind="getGrid">
      <n-gi
        v-bind="schema.giProps"
        v-for="schema in data"
        :key="schema.field"
        :span="getHidden(schema) ? 0 : schema.giProps?.span || 1"
      >
        <n-form-item
          v-bind="getFormItem(schema)"
          :label="schema.label"
          :path="schema.field"
          :showFeedback="schema.showFeedback"
          v-if="!getHidden(schema)"
        >
          <!--标签名右侧温馨提示-->
          <template #label v-if="schema.labelMessage">
            {{ schema.label }}
            <n-tooltip trigger="hover" :style="schema.labelMessageStyle">
              <template #trigger>
                <n-icon size="18" class="text-gray-400 cursor-pointer">
                  <QuestionCircleOutlined />
                </n-icon>
              </template>
              {{ schema.labelMessage }}
            </n-tooltip>
          </template>

          <!--判断插槽-->
          <template v-if="schema.slot">
            <slot
              :name="schema.slot"
              :model="formModel"
              :field="schema.field"
              :value="formModel[schema.field]"
            ></slot>
          </template>

          <!--NCheckbox-->
          <template v-else-if="schema.component === 'NCheckbox'">
            <n-checkbox-group
              v-bind="getSpecComponentProps(schema)"
              v-model:value="formModel[schema.field]"
            >
              <n-space>
                <n-checkbox
                  v-for="item in getSpecComponentProps(schema)?.options"
                  :key="item.value"
                  :value="item.value"
                  :label="item.label"
                />
              </n-space>
            </n-checkbox-group>
          </template>

          <!--NRadioGroup-->
          <template v-else-if="schema.component === 'NRadioGroup'">
            <n-radio-group
              v-bind="getSpecComponentProps(schema)"
              v-model:value="formModel[schema.field]"
            >
              <n-space>
                <n-radio
                  v-for="item in getSpecComponentProps(schema)?.options"
                  :key="item.value"
                  :value="item.value"
                >
                  {{ item.label }}
                </n-radio>
              </n-space>
            </n-radio-group>
          </template>

          <!-- BasicSelect -->
          <template v-else-if="schema.component === 'BasicSelect'">
            <BasicSelect v-model:value="formModel[schema.field]" v-bind="getComponentProps(schema)">
              <template v-for="(slot, name) in schema.componentSlots" :key="name" #[name]="scope">
                <component
                  :is="
                    slot({
                      model: formModel,
                      field: schema.field,
                      formAction: formActionType,
                      ...scope,
                    })
                  "
                />
              </template>
            </BasicSelect>
          </template>

          <!-- ImageUpload -->
          <template v-else-if="schema.component === 'ImageUpload'">
            <ImageUpload
              v-model:value="formModel[schema.field]"
              v-bind="getSpecComponentProps(schema)"
            />
          </template>

          <!-- valueFormat NDatePicker -->
          <template
            v-else-if="schema.component === 'NDatePicker' && getComponentProps(schema)?.valueFormat"
          >
            <n-date-picker
              v-model:formatted-value="formModel[schema.field]"
              v-bind="getComponentProps(schema)"
              clearable
            />
          </template>

          <!-- valueFormat NTimePicker -->
          <template
            v-else-if="schema.component === 'NTimePicker' && getComponentProps(schema)?.valueFormat"
          >
            <n-time-picker
              v-model:formatted-value="formModel[schema.field]"
              v-bind="getComponentProps(schema)"
              clearable
            />
          </template>

          <!--动态渲染表单组件-->
          <component
            v-else
            :is="getComponent(schema)"
            v-model:value="formModel[schema.field]"
            :class="{ isFull: schema.isFull != false && getProps.isFull }"
          />
          <!--组件后面的内容-->
          <template v-if="schema.suffix">
            <slot
              :name="schema.suffix"
              :model="formModel"
              :field="schema.field"
              :value="formModel[schema.field]"
            ></slot>
          </template>
        </n-form-item>
      </n-gi>
      <!--提交 重置 展开 收起 按钮-->
      <n-gi
        :suffix="isInline ? true : false"
        :span="isInline ? 1 : 24"
        #="{ overflow }"
        v-if="!props.group.length && getProps.showActionButtonGroup"
      >
        <n-space
          align="center"
          :justify="isInline ? 'end' : 'start'"
          :style="{ 'margin-left': `${isInline ? 12 : getProps.labelWidth}px` }"
        >
          <n-button
            v-if="getProps.showSubmitButton"
            v-bind="getSubmitBtnOptions"
            @click="handleSubmit"
            :loading="loading"
            >{{ getProps.submitButtonText }}
          </n-button>
          <n-button v-if="getProps.showResetButton" v-bind="getResetBtnOptions" @click="resetFields"
            >{{ getProps.resetButtonText }}
          </n-button>
          <n-button
            type="primary"
            text
            icon-placement="right"
            v-if="isInline && getProps.showAdvancedButton"
            @click="unfoldToggle"
          >
            <template #icon>
              <n-icon size="14" class="unfold-icon" v-if="overflow">
                <DownOutlined />
              </n-icon>
              <n-icon size="14" class="unfold-icon" v-else>
                <UpOutlined />
              </n-icon>
            </template>
            {{ overflow ? '展开' : '收起' }}
          </n-button>
        </n-space>
      </n-gi>
    </n-grid>
  </DefineTemplate>

  <n-form v-bind="getBindValue" :model="formModel" ref="formElRef">
    <template v-if="props.group && props.group.length">
      <n-card
        v-for="pitem in getGroupSchemas"
        :key="pitem.title"
        :title="pitem.title"
        class="mb-3"
        size="small"
        content-style="padding: 20px 10px 0 10px"
        hoverable
        :segmented="{
          content: true,
        }"
        v-bind="pitem.cardProps || {}"
      >
        <ReuseTemplate v-bind="{ data: pitem.columns }" :key="pitem.title" />
        <template #cover v-if="pitem.cardSlots?.cover">
          <Render :ref="`render-cover`" :value="pitem.cardSlots?.cover?.()" />
        </template>
        <template #header v-if="pitem.cardSlots?.header">
          <Render :ref="`render-header`" :value="pitem.cardSlots?.header?.()" />
        </template>
        <template #header-extra v-if="pitem.cardSlots?.headerExtra">
          <Render :ref="`render-header-extra`" :value="pitem.cardSlots?.headerExtra?.()" />
        </template>
        <template #footer v-if="pitem.cardSlots?.footer">
          <Render :ref="`render-footer`" :value="pitem.cardSlots?.footer?.()" />
        </template>
        <template #action v-if="pitem.cardSlots?.action">
          <Render :ref="`render-action`" :value="pitem.cardSlots?.action?.()" />
        </template>
      </n-card>
      <n-space align="center" :justify="isInline ? 'end' : 'start'">
        <n-button
          v-if="getProps.showSubmitButton"
          v-bind="getSubmitBtnOptions"
          @click="handleSubmit"
          :loading="loading"
          >{{ getProps.submitButtonText }}
        </n-button>
        <n-button v-if="getProps.showResetButton" v-bind="getResetBtnOptions" @click="resetFields"
          >{{ getProps.resetButtonText }}
        </n-button>
      </n-space>
    </template>
    <template v-else>
      <ReuseTemplate v-bind="{ data: getSchema }" />
    </template>
  </n-form>
</template>

<script lang="ts" setup>
  import type { Ref } from 'vue';
  import { computed, h, onMounted, reactive, ref, unref, useAttrs, watch } from 'vue';
  import { useFormEvents } from './hooks/useFormEvents';
  import { useFormValues } from './hooks/useFormValues';

  import { basicProps } from './props';
  import { DownOutlined, QuestionCircleOutlined, UpOutlined } from '@vicons/antd';
  import { BasicSelect } from '@/components/Select';
  import { ImageUpload } from '@/components/Upload';
  import type { GridProps } from 'naive-ui/lib/grid';
  import {
    componentSlotsType,
    FormActionType,
    FormGroupRow,
    FormProps,
    FormSchema,
  } from './types/form';

  import { isArray, isBoolean, isFunction, isObject } from '@/utils/is';
  import { deepMerge } from '@/utils';
  import { createReusableTemplate, onKeyStroke } from '@vueuse/core';
  import { isNil, uniqBy } from 'lodash-es';
  import {
    NAutoComplete,
    NCascader,
    NCheckbox,
    NColorPicker,
    NDatePicker,
    NInput,
    NInputNumber,
    NRadio,
    NRadioGroup,
    NRate,
    NSelect,
    NSlider,
    NSwitch,
    NTimePicker,
    NTreeSelect,
    NUpload,
  } from 'naive-ui';

  const props = defineProps({ ...basicProps });
  const emit = defineEmits(['reset', 'submit', 'register', 'advanced', 'change']);
  const attrs = useAttrs();

  const defaultFormModel = ref<Recordable>({});
  const formModel = reactive<Recordable>({});
  const propsRef = ref<Partial<FormProps>>({});
  const schemaRef = ref<Nullable<FormSchema[]>>(null);
  const groupSchemaRef = ref<Nullable<FormGroupRow[]>>(null);
  const formElRef = ref<Nullable<FormActionType>>(null);
  const gridCollapsed = ref(!props.collapsed);
  const loading = ref(false);
  const isUpdateDefaultRef = ref(false);

  const [DefineTemplate, ReuseTemplate] = createReusableTemplate();

  watch(
    formModel,
    () => {
      emit('change', formModel);
    },
    {
      deep: true,
    },
  );

  const getSubmitBtnOptions = computed(() => {
    return Object.assign(
      {
        size: props.size,
        type: 'primary',
      },
      props.submitButtonOptions,
    );
  });

  const getResetBtnOptions = computed(() => {
    return Object.assign(
      {
        size: props.size,
        type: 'default',
      },
      props.resetButtonOptions,
    );
  });

  const componentMap = {
    NInput: NInput,
    NInputNumber: NInputNumber,
    NSelect: NSelect,
    NTreeSelect: NTreeSelect,
    NRadio: NRadio,
    NRadioGroup: NRadioGroup,
    NCheckbox: NCheckbox,
    NDatePicker: NDatePicker,
    NTimePicker: NTimePicker,
    NAutoComplete: NAutoComplete,
    NCascader: NCascader,
    NSwitch: NSwitch,
    NUpload: NUpload,
    NSlider: NSlider,
    NRate: NRate,
    NColorPicker: NColorPicker,
    BasicSelect: BasicSelect,
    ImageUpload: ImageUpload,
  };

  function getComponent(schema: FormSchema) {
    const slots = {};
    if (schema.componentSlots) {
      Object.keys(schema.componentSlots).forEach((key: string) => {
        slots[key] = () =>
          (schema.componentSlots as Record<string, componentSlotsType>)[key]({
            model: formModel, // 传入表单数据
            field: schema.field, // 传入字段名
            formAction: formActionType, // 传入表单操作
          });
      });
    }

    return h(componentMap[schema.component as string], getComponentProps(schema), slots);
  }

  function getComponentProps(schema) {
    const compProps = handleComponentProps(schema);
    return {
      clearable: true,
      ...compProps,
    };
  }

  function getSpecComponentProps(schema) {
    const compProps = handleComponentProps(schema);
    return {
      ...compProps,
    };
  }

  function getHidden(schema): boolean {
    const hidden = schema.hidden;
    const field = schema.field;
    if (isBoolean(hidden)) return hidden;

    if (isFunction(hidden)) {
      const values = getFieldsValue();
      const status = hidden({ schema, values, model: formModel, field });
      return status;
    }
    return false;
  }

  const getProps = computed((): FormProps => {
    const formProps = { ...props, ...unref(propsRef) } as FormProps;
    const rulesObj = {
      rules: {},
    };
    const schemas = formProps.schemas || [];
    schemas.forEach((item) => {
      if (item.rules && isArray(item.rules)) {
        rulesObj.rules[item.field] = item.rules;
      }
    });
    // 组装分组表单
    if (props.group && props.group.length) {
      props.group.forEach((item) => {
        item.columns.forEach((colItem) => {
          if (colItem.rules && isArray(colItem.rules)) {
            rulesObj.rules[colItem.field] = colItem.rules;
          }
        });
      });
    }
    return { ...formProps, ...unref(rulesObj) };
  });

  const isInline = computed(() => {
    const { layout } = unref(getProps);
    return layout === 'inline';
  });

  const getCollapsed = computed(() => {
    const { showAdvancedButton } = unref(getProps);
    return isInline.value && showAdvancedButton != false ? gridCollapsed.value : false;
  });

  const getGrid = computed((): GridProps => {
    const { gridProps } = unref(getProps);
    return {
      ...gridProps,
      collapsed: getCollapsed.value,
      responsive: 'screen',
    };
  });

  const getBindValue = computed(() => ({ ...attrs, ...props, ...unref(getProps) } as Recordable));

  const getGroupSchemas = computed((): FormGroupRow[] => {
    const groupSchemas = unref(groupSchemaRef) || unref(getProps).group;
    return groupSchemas as FormGroupRow[];
  });

  const getSchema = computed((): FormSchema[] => {
    const schemas: FormSchema[] = unref(schemaRef) || (unref(getProps).schemas as any);
    for (const schema of schemas) {
      const { defaultValue } = schema;
      // handle date type
      // dateItemType.includes(component as string)
      if (defaultValue) {
        schema.defaultValue = defaultValue;
      }
    }
    return schemas as FormSchema[];
  });

  function handleComponentProps(schema) {
    let { componentProps = {} } = schema ?? {};
    if (isFunction(componentProps)) {
      componentProps = componentProps({ model: formModel, formAction: formActionType }) ?? {};
    }

    return componentProps;
  }

  const { handleFormValues, initDefault } = useFormValues({
    defaultFormModel,
    getSchema,
    formModel,
    getGroupSchemas,
  });

  const { handleSubmit, validate, resetFields, getFieldsValue, clearValidate, setFieldsValue } =
    useFormEvents({
      emit,
      getProps,
      formModel,
      getSchema,
      getGroupSchemas,
      formElRef: formElRef as Ref<FormActionType>,
      defaultFormModel,
      loading,
      handleFormValues,
    });

  function getFormItem(values: FormSchema) {
    return values as any;
  }

  function unfoldToggle() {
    gridCollapsed.value = !gridCollapsed.value;
    emit('advanced', gridCollapsed.value);
  }

  async function setProps(formProps: Partial<FormProps>): Promise<void> {
    propsRef.value = deepMerge(unref(propsRef) || {}, formProps);
  }

  async function setSchema(schemaProps: FormSchema[]): Promise<void> {
    schemaRef.value = schemaProps;
  }

  async function updateSchema(schemaProps: FormSchema | FormSchema[]): Promise<void> {
    let updateData: Partial<FormSchema>[] = [];
    if (isObject(schemaProps)) {
      updateData.push(schemaProps as FormSchema);
    }

    if (isArray(schemaProps)) {
      updateData = [...schemaProps];
    }

    const hasField = updateData.every((item) => Reflect.has(item, 'field') && item.field);
    if (!hasField) {
      throw new Error(
        'All children of the form Schema array that need to be updated must contain the `field` field',
      );
    }

    const schema: FormSchema[] = [];
    const updatedSchema: FormSchema[] = [];
    unref(getSchema).forEach((val) => {
      let updateItem = updateData.find((item) => item.field === val.field);

      if (updateItem) {
        const newSchema = deepMerge(val, updateItem);
        updatedSchema.push(newSchema as FormSchema);
        schema.push(newSchema as FormSchema);
      } else {
        schema.push(val);
      }
    });

    _setDefaultValue(updatedSchema);
    await setSchema(uniqBy(schema, 'field'));
  }

  /**
   * 设置表单默认值
   * @param data
   */
  function _setDefaultValue(data: FormSchema | FormSchema[]) {
    let schemas: FormSchema[] = [];
    if (isObject(data)) {
      schemas.push(data as FormSchema);
    }

    if (isArray(data)) {
      schemas = [...data];
    }

    const obj: Recordable = {};
    const currentFieldsValue = getFieldsValue();
    schemas.forEach((item) => {
      if (
        Reflect.has(item, 'field') &&
        item.field &&
        !isNil(item.defaultValue) &&
        (!(item.field in currentFieldsValue) || isNil(currentFieldsValue[item.field]))
      ) {
        obj[item.field] = item.defaultValue;
      }
    });

    Object.keys(obj).length && setFieldsValue(obj);
  }

  async function setGroupSchema(schemaProps: FormGroupRow[]): Promise<void> {
    groupSchemaRef.value = schemaProps;
  }

  /**
   * group schemas比较特殊，层级有点多，所以新增一个 key 值，来标识具体是哪一个 group
   * @deprecated 此方法存在缺陷，将于后续版本中升级
   */
  async function updateGroupSchema(
    key: string,
    propertyPath: string,
    newValue: any,
  ): Promise<void> {
    let group = unref(getGroupSchemas).map((item) => {
      if (item.key === key) {
        const updatedItem = { ...item };
        const properties = propertyPath.split('.');
        let current: any = updatedItem;

        for (let i = 0; i < properties.length - 1; i++) {
          const prop = properties[i];
          if (prop === 'columns') {
            const field = properties[i + 1];
            current = current[prop].find((col: FormSchema) => col.field === field);
            i++;
          } else {
            current = current[prop];
          }
        }

        const lastProp = properties[properties.length - 1];

        if (isFunction(newValue) && lastProp === 'componentProps') {
          newValue = newValue({ model: formModel, formAction: formActionType }) ?? {};
        }

        current[lastProp] = newValue;

        return updatedItem;
      }

      return item;
    });

    await setGroupSchema(group);
  }

  //更新loading状态
  function setLoading(status: boolean): void {
    loading.value = status;
  }

  const formActionType: Partial<FormActionType> = {
    getFieldsValue,
    setFieldsValue,
    resetFields,
    validate,
    clearValidate,
    setProps,
    setSchema,
    setGroupSchema,
    setLoading,
    submit: handleSubmit,
    updateSchema,
    updateGroupSchema,
  };

  watch(
    () => getSchema.value,
    (schema) => {
      if (unref(isUpdateDefaultRef)) {
        return;
      }
      if (schema?.length) {
        initDefault();
        isUpdateDefaultRef.value = true;
      }
    },
  );

  onMounted(() => {
    initDefault();
    emit('register', formActionType);
    // 回车 Submit fix Enter Event
    getProps.value.isEnterSubmit &&
      onKeyStroke('Enter', handleSubmit, { target: formElRef.value as any, dedupe: true });
  });

  defineExpose({
    ...formActionType,
  });
</script>

<style lang="less" scoped>
  .isFull {
    width: 100%;
    justify-content: flex-start;
  }

  .unfold-icon {
    display: flex;
    align-items: center;
    height: 100%;
    margin-left: -3px;
  }
</style>
