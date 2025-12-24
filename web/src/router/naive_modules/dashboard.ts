import { RouteRecordRaw } from 'vue-router';
import { Layout } from '@/router/base';
import {
  DashboardOutlined,
  HomeOutlined,
  DesktopOutlined,
  FundProjectionScreenOutlined,
} from '@vicons/antd';
import { renderIcon } from '@/utils/index';

const routes: Array<RouteRecordRaw> = [
  {
    path: 'dashboard',
    name: 'dashboard',
    redirect: '/dashboard/console',
    component: ParentLayout,
    meta: {
      title: 'Dashboard',
      icon: renderIcon(DashboardOutlined),
      permissions: ['dashboard_console', 'dashboard_console', 'dashboard_workplace'],
      sort: 0,
    },
    children: [
      {
        path: 'console',
        name: 'dashboard_console',
        meta: {
          title: '主控制台',
          permissions: ['dashboard_console'],
          icon: renderIcon(HomeOutlined),
          affix: true,
        },
        component: () => import('@/views/dashboard/console/console.vue'),
      },
      {
        path: 'monitor',
        name: `dashboard_monitor`,
        meta: {
          title: '监控台',
          icon: renderIcon(FundProjectionScreenOutlined),
        },
        component: () => import('@/views/dashboard/monitor/monitor.vue'),
      },
      {
        path: 'workplace',
        name: `dashboard_workplace`,
        meta: {
          title: '工作台',
          icon: renderIcon(DesktopOutlined),
        },
        component: () => import('@/views/dashboard/workplace/workplace.vue'),
      },
    ],
  },
];

export default routes;
