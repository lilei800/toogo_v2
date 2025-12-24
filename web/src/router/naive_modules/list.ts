import { RouteRecordRaw } from 'vue-router';
import { Layout } from '@/router/base';
import { TableOutlined } from '@vicons/antd';
import { renderIcon } from '@/utils/index';

const routes: Array<RouteRecordRaw> = [
  {
    path: 'list',
    name: 'List',
    redirect: '/list/basic-list',
    component: ParentLayout,
    meta: {
      title: '列表页面',
      icon: renderIcon(TableOutlined),
      sort: 3,
    },
    children: [
      {
        path: 'basic-list/:id?',
        name: 'basic-list',
        meta: {
          title: '基础列表',
        },
        component: () => import('@/views/list/basicList/index.vue'),
      },
      {
        path: 'basic-info/:id?',
        name: 'BasicInfo',
        meta: {
          title: '基础详情',
          hidden: true,
          activeMenu: 'basic-list',
          keepAlive: true,
        },
        component: () => import('@/views/list/basicList/info.vue'),
      },
    ],
  },
];

export default routes;
