import { RouteRecordRaw } from 'vue-router';
import { Layout } from '@/router/base';
import { OptionsSharp } from '@vicons/ionicons5';
import { renderIcon } from '@/utils/index';

const routes: Array<RouteRecordRaw> = [
  {
    path: 'system',
    name: 'System',
    redirect: '/system/menu',
    component: ParentLayout,
    meta: {
      title: '系统管理',
      icon: renderIcon(OptionsSharp),
      sort: 1,
    },
    children: [
      {
        path: 'user',
        name: 'system_user',
        meta: {
          title: '用户管理',
          keepAlive: true,
        },
        component: () => import('@/views/system/user/user.vue'),
      },
      {
        path: 'menu',
        name: 'system_menu',
        meta: {
          title: '菜单管理',
          keepAlive: true,
        },
        component: () => import('@/views/system/menu/index.vue'),
      },
      {
        path: 'menu/table',
        name: 'system_menu_table',
        meta: {
          title: '菜单管理2',
        },
        component: () => import('@/views/system/menu/table.vue'),
      },
      {
        path: 'role',
        name: 'system_role',
        meta: {
          title: '角色管理',
          keepAlive: true,
        },
        component: () => import('@/views/system/role/role.vue'),
      },
      {
        path: 'dictionary',
        name: 'system_dictionary',
        meta: {
          title: '字典管理',
        },
        component: () => import('@/views/system/dictionary/dictionary.vue'),
      },
    ],
  },
];

export default routes;
