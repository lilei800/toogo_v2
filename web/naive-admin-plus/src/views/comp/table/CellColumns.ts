import { h } from 'vue';
import { NTag } from 'naive-ui';
import { BasicColumn } from '@/components/Table';
import { statusMap, TableItem } from '@/api/table/list';
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
    width: 80,
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
    width: 180,
  },
  {
    title: '城市',
    key: 'city',
    editComponent: 'NInput',
    edit: true,
    width: 180,
  },
  {
    title: '状态',
    key: 'status',
    width: 160,
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
