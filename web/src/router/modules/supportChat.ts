import { RouteRecordRaw } from 'vue-router';
import { Layout } from '@/router/constant';
import { renderIcon } from '@/utils/index';
import { CustomerServiceOutlined } from '@vicons/antd';

const routes: RouteRecordRaw[] = [
  {
    path: '/supportChat',
    name: 'SupportChatRoot',
    component: Layout,
    meta: {
      title: '客服',
      icon: renderIcon(CustomerServiceOutlined),
      sort: 98,
    },
    children: [
      {
        path: 'workbench',
        name: 'support_chat_workbench',
        component: () => import('@/views/supportChat/index.vue'),
        meta: {
          title: '客服工作台',
          keepAlive: true,
        },
      },
      {
        path: 'client',
        name: 'support_chat_client',
        component: () => import('@/views/supportChat/client.vue'),
        meta: {
          title: '联系客服(客户端)',
          keepAlive: true,
          // 客户端入口也需要登录；如果你们未来要游客可用，可在这里加 ignoreAuth
        },
      },
    ],
  },
];

export default routes;
