import { h } from 'vue';
import { BasicColumn } from '@/components/Table';
import { TableItem } from '@/api/table/list';
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
    editComponent: 'NInput',
    editRow: true,
    // 默认必填校验
    editRule: true,
    edit: true,
    width: 200,
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
    title: '邮箱',
    key: 'email',
    editComponent: 'NInput',
    edit: true,
    editRow: true,
    width: 180,
  },
  {
    title: '城市',
    key: 'city',
    editComponent: 'NInput',
    edit: true,
    editRow: true,
    width: 180,
  },
  {
    title: '状态',
    key: 'status',
    editRow: true,
    editComponent: 'NSelect',
    editComponentProps: {
      options: [
        {
          label: '取消',
          value: 'close',
        },
        {
          label: '拒绝',
          value: 'refuse',
        },
        {
          label: '成功',
          value: 'success',
        },
      ],
    },
    edit: true,
    width: 200,
    ellipsis: false,
  },
  {
    title: '创建时间',
    width: 320,
    key: 'createDate',
    align: 'center',
    editRow: true,
    edit: true,
    editComponent: 'NDatePicker',
    editComponentProps: {
      type: 'datetime',
    },
  },
];
