<template>
  <n-card :bordered="false" class="proCard">
    <BasicTable
      ref="actionRef"
      title="编辑单元格"
      titleTooltip="表头列显示了编辑图标代表可编辑"
      :columns="columns"
      :request="loadDataTable"
      :row-key="(row) => row.id"
      @edit-end="editEnd"
      @edit-change="onEditChange"
      @update:checked-row-keys="onCheckedRow"
      :scroll-x="1360"
      :row-props="rowProps"
    />
  </n-card>
</template>

<script lang="ts" setup>
  import { reactive, ref } from 'vue';
  import { BasicTable } from '@/components/Table';
  import { getTableList } from '@/api/table/list';
  import { columns } from './CellColumns';

  const actionRef = ref();
  const params = reactive({
    pageSize: 5,
    name: 'NaiveAdmin',
  });

  function rowProps(rows) {
    return {
      style: 'cursor: pointer;',
      onclick: function () {
        console.log('row点击事件触发');
        console.log(rows);
      },
    };
  }

  function onEditChange({ column, value, record }) {
    if (column.key === 'id') {
      record.editValueRefs.name4.value = `${value}`;
    }
    console.log(column, value, record);
  }

  const loadDataTable = async (res) => {
    return await getTableList({ ...params, ...res });
  };

  function onCheckedRow(rowKeys) {
    console.log(rowKeys);
  }

  function editEnd({ record, index, key, value }) {
    console.log('record', record);
    console.log('index', index);
    console.log('key', key);
    console.log('value', value);
  }
</script>

<style lang="less" scoped></style>
