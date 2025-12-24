import { FormSchema } from '@/components/Form';

export const schemas: FormSchema[] = [
  {
    field: 'name',
    component: 'NInput',
    label: '名称',
    rules: [{ required: true, message: '请输入名称', trigger: ['blur'] }],
  },
  {
    field: 'sex',
    component: 'NSelect',
    label: '性别',
    componentProps: {
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
    },
  },
  {
    field: 'email',
    component: 'NInput',
    label: '邮箱',
    componentProps: {
      placeholder: '请输入邮箱',
    },
  },
  {
    field: 'city',
    component: 'NInput',
    label: '城市',
    componentProps: {
      placeholder: '请输入城市',
    },
  },
  {
    field: 'status',
    component: 'NSelect',
    label: '状态',
    componentProps: {
      placeholder: '请选择性别',
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
    },
  },
];
