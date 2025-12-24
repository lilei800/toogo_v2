import { h } from 'vue';
import { NTag } from 'naive-ui';
import { BasicColumn } from '@/components/Table';
import { sexMap, statusMap, TableItem } from '@/api/table/list';
import { TableImg } from '@/components/TableImg';

export const columns: BasicColumn[] = [
  {
    title: 'id',
    key: 'id',
    width: 100,
  },
  {
    title: '名称',
    key: 'name',
    width: 160,
    search: true,
    searchProps: {
      placeholder: '请输入名称',
      style: { width: '150px' },
    },
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
    width: 180,
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
  },
];
