import { h } from 'vue';
import { TableImg } from '@/components/TableImg';
import { BasicColumn } from '@/components/Table';
import { NTag } from 'naive-ui';
import { sexMap, statusMap, TableItem } from '@/api/table/list';

export const columns: BasicColumn<TableItem>[] = [
  {
    type: 'selection',
    width: 60,
    key: 'selection',
  },
  {
    title: 'id',
    key: 'id',
    width: 100,
    sortOrder: 'descend',
    sorter: (row1: TableItem, row2: TableItem) => row1.id - row2.id,
  },
  {
    title: '名称',
    key: 'name',
    width: 160,
    search: true,
    searchType: 'text',
    searchProps: {
      placeholder: '请输入名称',
      style: { width: '150px' },
    },
    sorter: 'default',
    sortOrder: 'descend',
  },
  {
    title: '头像',
    key: 'avatar',
    width: 160,
    render(row: TableItem) {
      return h(TableImg, {
        imageProps: {
          width: 32,
          height: 32,
          //@ts-ignore
          style: 'border-radius: 50%',
        },
        imgList: [row.avatar],
        spaceProps: {
          justify: 'space-between',
        },
      });
    },
    print: false,
  },
  {
    title: '性别',
    key: 'sex',
    width: 160,
    search: true,
    searchType: 'select',
    searchProps: {
      placeholder: '请选择性别',
      options: [
        {
          label: '男',
          value: 'male',
        },
        {
          label: '女',
          value: 'female',
        },
      ],
      style: { width: '150px' },
    },
    render(record) {
      return h(
        NTag,
        {
          size: 'small',
          type: record.sex === 'male' ? 'info' : 'error',
        },
        {
          default: () => sexMap[record.sex],
        },
      );
    },
  },
  {
    title: '邮箱',
    key: 'email',
    width: 190,
    print: false,
  },
  {
    title: '城市',
    key: 'city',
    width: 180,
  },
  {
    title: '状态',
    key: 'status',
    width: 160,
    sortOrder: 'descend',
    sorter: 'default',
    search: true,
    searchType: 'select',
    searchProps: {
      placeholder: '请选择状态',
      options: [
        {
          label: '拒绝',
          value: 'refuse',
        },
        {
          label: '取消',
          value: 'close',
        },
        {
          label: '成功',
          value: 'success',
        },
      ],
      style: { width: '150px' },
    },
    render(record) {
      return h(
        NTag,
        {
          size: 'small',
          type:
            record.status === 'close'
              ? 'default'
              : record.status === 'refuse'
              ? 'error'
              : 'success',
        },
        {
          default: () => statusMap[record.status],
        },
      );
    },
  },
  {
    title: '创建时间',
    width: 320,
    key: 'createDate',
    align: 'center',
    print: false,
  },
];
