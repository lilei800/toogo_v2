import { RouteRecordRaw } from 'vue-router';
import { Layout } from '@/router/base';
import { CheckCircleOutlined } from '@vicons/antd';
import { renderIcon } from '@/utils/index';

const routes: Array<RouteRecordRaw> = [
  {
    path: 'result',
    name: 'Result',
    redirect: '/result/success',
    component: ParentLayout,
    meta: {
      title: '结果页面',
      icon: renderIcon(CheckCircleOutlined),
      sort: 5,
    },
    children: [
      {
        path: 'success',
        name: 'result-success',
        meta: {
          title: '成功页面',
        },
        component: () => import('@/views/result/success.vue'),
      },
      {
        path: 'fail',
        name: 'result-fail',
        meta: {
          title: '失败页面',
        },
        component: () => import('@/views/result/fail.vue'),
      },
      {
        path: 'info',
        name: 'result-info',
        meta: {
          title: '信息页面',
        },
        component: () => import('@/views/result/info.vue'),
      },
    ],
  },
];

export default routes;
