import { h } from 'vue';
import { NAvatar, NTag } from 'naive-ui';
import { BasicColumn } from '@/components/Table';
import { sexMap, statusMap } from '@/api/table/list';

export const columns: BasicColumn[] = [
  {
    type: 'selection',
    key: 'selection',
  },
  {
    title: '用户名',
    key: 'username',
    width: 140,
  },
  {
    title: '头像',
    key: 'avatar',
    render(row) {
      return h(NAvatar, {
        alt: '这是一个图片说明',
        width: 32,
        src: row.avatar,
        round: true,
      });
    },
    width: 100,
  },
  {
    title: '登录账号',
    key: 'account',
    width: 120,
  },
  {
    title: '手机号',
    key: 'mobile',
    width: 120,
  },
  {
    title: '邮箱',
    key: 'email',
    width: 200,
  },
  {
    title: '性别',
    key: 'sex',
    width: 120,
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
    title: '角色',
    key: 'role',
    width: 120,
    render(row) {
      return h(
        NTag,
        {
          type: 'info',
        },
        {
          default: () => row.role,
        },
      );
    },
  },
  {
    title: '状态',
    key: 'status',
    width: 120,
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
