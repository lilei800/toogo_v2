import { RouteRecordRaw } from 'vue-router';
import { Layout } from '@/router/base';
import { KeyOutlined } from '@vicons/antd';
import { LogInOutline } from '@vicons/ionicons5';
import { renderIcon } from '@/utils';

const hsot = location.origin;

const routes: Array<RouteRecordRaw> = [
  {
    path: 'authentication',
    name: 'Authentication',
    component: ParentLayout,
    meta: {
      title: '认证模板',
      icon: renderIcon(KeyOutlined),
      sort: 11,
    },
    children: [
      {
        path: 'login-v1',
        name: `${hsot}/login-v1`,
        meta: {
          title: '登录版本1',
          icon: renderIcon(LogInOutline),
        },
        component: () => import('@/views/authentication/v1/login.vue'),
      },
      {
        path: 'login-v2',
        name: `${hsot}/login-v2`,
        meta: {
          title: '登录版本2',
          icon: renderIcon(LogInOutline),
        },
        component: () => import('@/views/authentication/v2/login.vue'),
      },
      {
        path: 'login-v2',
        name: `${hsot}/login-v3`,
        meta: {
          title: '登录版本3',
          icon: renderIcon(LogInOutline),
        },
        component: () => import('@/views/authentication/v3/login.vue'),
      },
      {
        path: 'login-v4',
        name: `${hsot}/login-v4`,
        meta: {
          title: '登录版本4',
          icon: renderIcon(LogInOutline),
        },
        component: () => import('@/views/authentication/v4/login.vue'),
      },
    ],
  },
];

export default routes;
