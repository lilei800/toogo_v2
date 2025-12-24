import { RouteRecordRaw } from 'vue-router';
import { Layout } from '@/router/constant';
import { renderIcon } from '@/utils';
import { WalletOutlined } from '@vicons/antd';

/**
 * Trading 量化交易路由
 */
const routes: Array<RouteRecordRaw> = [
  {
    path: '/trading',
    name: 'Trading',
    redirect: '/trading/robot',
    component: Layout,
    meta: {
      title: '量化交易',
      icon: renderIcon(WalletOutlined),
      sort: 3,
      isRoot: true,
      activeMenu: 'trading',
    },
    children: [
      {
        path: 'api-config',
        name: 'trading_api_config',
        meta: {
          title: 'API配置',
          activeMenu: 'trading_api_config',
        },
        component: () => import('@/views/trading/api-config/index.vue'),
      },
      {
        path: 'proxy-config',
        name: 'trading_proxy_config',
        meta: {
          title: '代理配置',
          activeMenu: 'trading_proxy_config',
        },
        component: () => import('@/views/trading/proxy-config/index.vue'),
      },
      {
        path: 'robot',
        name: 'trading_robot',
        meta: {
          title: '机器人管理',
          activeMenu: 'trading_robot',
        },
        component: () => import('@/views/trading/robot/index.vue'),
      },
      {
        path: 'robot/create',
        name: 'trading_robot_create',
        meta: {
          title: '创建机器人',
          hidden: true,
          activeMenu: 'trading_robot',
        },
        component: () => import('@/views/trading/robot/create.vue'),
      },
      {
        path: 'robot/detail/:id',
        name: 'trading_robot_detail',
        meta: {
          title: '机器人详情',
          hidden: true,
          activeMenu: 'trading_robot',
        },
        component: () => import('@/views/trading/robot/detail.vue'),
      },
    ],
  },
];

export default routes;

