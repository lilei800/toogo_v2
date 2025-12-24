import { RouteRecordRaw } from 'vue-router';
import { Layout } from '@/router/base';
import { ConstructSharp } from '@vicons/ionicons5';
import { renderIcon } from '@/utils/index';

const routes: Array<RouteRecordRaw> = [
  {
    path: 'advanced',
    name: 'advanced',
    redirect: '/advanced/table',
    component: ParentLayout,
    meta: {
      title: '高级组件',
      icon: renderIcon(ConstructSharp),
      alwaysShow: true,
      sort: 13,
      activeMenu: 'advanced_table',
    },
    children: [
      {
        path: 'table',
        name: `advanced_table`,
        meta: {
          title: '高级表格',
          icon: renderIcon(ConstructSharp),
          affix: true,
        },
        component: () => import('@/views/advanced/table/table.vue'),
      },
    ],
  },
];

export default routes;
