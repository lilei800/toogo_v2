import { RouteRecordRaw } from 'vue-router';
import { Layout } from '@/router/base';
import { ProfileOutlined } from '@vicons/antd';
import { renderIcon } from '@/utils/index';

const routes: Array<RouteRecordRaw> = [
  {
    path: 'form',
    name: 'Form',
    redirect: '/form/basic-form',
    component: ParentLayout,
    meta: {
      title: '表单页面',
      icon: renderIcon(ProfileOutlined),
      sort: 2,
    },
    children: [
      {
        path: 'basic-form',
        name: 'BasicForm',
        meta: {
          title: '基础表单',
          keepAlive: true,
        },
        component: () => import('@/views/form/basicForm/index.vue'),
      },
      {
        path: 'advanced-form',
        name: 'form-advanced-form',
        meta: {
          title: '高级表单',
        },
        component: () => import('@/views/form/advancedForm/advancedForm.vue'),
      },
      {
        path: 'step-form',
        name: 'form-step-form',
        meta: {
          title: '分步表单',
        },
        component: () => import('@/views/form/stepForm/stepForm.vue'),
      },
      {
        path: 'detail',
        name: 'form-detail',
        meta: {
          title: '表单详情',
        },
        component: () => import('@/views/form/detail/index.vue'),
      },
    ],
  },
];

export default routes;
