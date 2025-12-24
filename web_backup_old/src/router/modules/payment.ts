import { RouteRecordRaw } from 'vue-router';
import { Layout } from '@/router/constant';
import { renderIcon } from '@/utils';
import { DollarOutlined } from '@vicons/antd';

/**
 * Payment USDT管理路由
 */
const routes: Array<RouteRecordRaw> = [
  {
    path: '/payment',
    name: 'Payment',
    redirect: '/payment/balance',
    component: Layout,
    meta: {
      title: 'USDT管理',
      icon: renderIcon(DollarOutlined),
      sort: 4,
      isRoot: true,
      activeMenu: 'payment',
    },
    children: [
      {
        path: 'balance',
        name: 'payment_balance',
        meta: {
          title: '我的余额',
          activeMenu: 'payment_balance',
        },
        component: () => import('@/views/payment/balance/index.vue'),
      },
      {
        path: 'deposit',
        name: 'payment_deposit',
        meta: {
          title: 'USDT充值',
          activeMenu: 'payment_deposit',
        },
        component: () => import('@/views/payment/deposit/index.vue'),
      },
      {
        path: 'withdraw',
        name: 'payment_withdraw',
        meta: {
          title: 'USDT提现',
          activeMenu: 'payment_withdraw',
        },
        component: () => import('@/views/payment/withdraw/index.vue'),
      },
      {
        path: 'admin/withdraw-audit',
        name: 'payment_admin_withdraw_audit',
        meta: {
          title: '提现审核',
          activeMenu: 'payment_admin_withdraw_audit',
          permissions: ['admin.payment.withdraw.audit'],
        },
        component: () => import('@/views/payment/admin/withdraw-audit.vue'),
      },
    ],
  },
];

export default routes;

