<template>
  <BasicModal ref="modalRef" @register="modalRegister" @on-ok="okModal">
    <template #default>
      <div class="pt-4">
        <n-form
          :rules="rules"
          :model="model"
          label-placement="left"
          label-width="auto"
          require-mark-placement="right-hanging"
          :style="{ maxWidth: '640px' }"
        >
          <n-form-item label="选择数据" path="dataType">
            <n-select v-model:value="model.dataType" :options="dataTypeOptions" />
          </n-form-item>
          <n-form-item label="选择字段" path="fields">
            <n-table size="small">
              <thead>
                <tr>
                  <th :width="15" align="center"></th>
                  <th>
                    <n-checkbox
                      v-model:checked="checkAll"
                      :indeterminate="!checkAll && checkPart"
                      @update:checked="handleToggleAll"
                    >
                      全选
                    </n-checkbox>
                  </th>
                  <th :width="80" align="center"></th>
                </tr>
              </thead>
              <tbody>
                <n-checkbox-group
                  v-model:value="model.fields"
                  style="display: contents"
                  @update:value="handleToggleSelection"
                >
                  <Draggable
                    :list="columnList"
                    style="display: contents"
                    animation="300"
                    item-key="print-key"
                  >
                    <template #item="{ element }">
                      <tr>
                        <td align="center" style="vertical-align: bottom">
                          <n-button text size="small">
                            <n-icon :size="17" class="cursor-move">
                              <DragOutlined />
                            </n-icon>
                          </n-button>
                        </td>
                        <td>
                          <n-checkbox :value="element.key" class="w-full">
                            {{ element.title }}
                          </n-checkbox>
                        </td>
                        <td>
                          <n-input-number
                            :show-button="false"
                            size="small"
                            placeholder="列宽"
                            v-model:value="element.width"
                          />
                        </td>
                      </tr>
                    </template>
                  </Draggable>
                </n-checkbox-group>
              </tbody>
            </n-table>
          </n-form-item>
        </n-form>
      </div>
    </template>
  </BasicModal>
</template>

<script lang="ts" setup>
  import { ref, onMounted } from 'vue';
  import { useMessage } from 'naive-ui';
  import { DragOutlined } from '@vicons/antd';
  import { useTableContext } from '../../hooks/useTableContext';
  import printJS from 'print-js';
  import 'print-js/dist/print.css';
  import { BasicModal, useModal } from '@/components/Modal';
  import { createSerial, filterColumns } from './helper';
  import { cloneDeep } from 'lodash-es';
  import Draggable from 'vuedraggable';

  const rules = {
    dataType: [{ required: true, message: '请选择数据类型', trigger: 'blur' }],
  };

  const dataTypeOptions = [
    { label: '当前页数据', value: 'current' },
    { label: '选中数据', value: 'checked' },
  ];

  const message = useMessage();
  const modalRef = ref<InstanceType<typeof BasicModal>>();
  const checkAll = ref<boolean>(true);
  const checkPart = ref<boolean>(false);
  const { getBindValues, checkedRowKeys } = useTableContext();

  const columnList = ref(simpItems(cloneDeep(getBindValues.value.columns)));

  const model = ref<{ dataType: string; fields: string[] }>({
    dataType: 'current',
    fields: [],
  });

  const [modalRegister, { openModal, closeModal, setSubLoading }] = useModal({
    title: '打印',
    subBtuText: '确定',
    width: 450,
  });

  function simpItems(list) {
    if (!Array.isArray(list) || list.length === 0) return [];
    list.unshift({
      width: 50,
      key: 'serial',
      title: '序号',
    });
    try {
      return filterColumns(list).map((item) => ({
        width: item.width || null,
        key: item.key,
        title: item.title,
      }));
    } catch (error) {
      console.error('Filter error:', error);
      return [];
    }
  }

  function okModal() {
    const isChecked = model.value.dataType === 'checked';
    if (model.value.fields.length === 0) {
      message.error('打印字段不能为空');
      setSubLoading(false);
      return;
    }
    if (isChecked && checkedRowKeys.value.length === 0) {
      message.error('请在列表勾选数据后在打印');
      setSubLoading(false);
      return;
    }
    const { data, printConfig } = getBindValues.value;
    const rowKey = printConfig?.rowKey;
    const printColumn = columnList.value
      .filter((item) => model.value.fields.includes(item.key as string))
      .map((item) => ({
        field: item.key,
        displayName: item.title,
        columnSize: `${item.width}`,
      }));
    setTimeout(() => {
      printJS({
        header: '自定义打印的顶部区域',
        documentTitle: '基础列表 - Naive Admin Plus',
        headerStyle: 'font-weight: 400;margin-bottom: 30px;text-align: center;',
        printable: createSerial(data, isChecked ? checkedRowKeys.value : [], rowKey),
        properties: printColumn,
        type: 'json',
        gridHeaderStyle: 'border: 1px solid #000;text-align:center', // 设置表头样式
        gridStyle: 'border: 1px solid #000;text-align:center', // 设置表格体样式
        ...getBindValues.value.printConfig,
      });
      setSubLoading(false);
    }, 1000);
  }

  function handleToggleAll(value: boolean) {
    checkPart.value = false;
    checkAll.value = value;
    if (value) {
      model.value.fields = columnList.value.map((item) => item.key);
    } else {
      model.value.fields = [];
    }
  }

  function handleToggleSelection(values: string[]) {
    model.value.fields = values;
    checkAll.value = values.length === columnList.value.length;
    checkPart.value = values.length < columnList.value.length;
  }

  onMounted(() => {
    model.value.fields = columnList.value.map((item) => item.key);
  });

  defineExpose({
    openModal,
    closeModal,
  });
</script>
