import { RouteRecordRaw } from 'vue-router';
import { Layout, ParentLayout } from '@/router/base';
import { MenuOutlined } from '@vicons/antd';
import { ReorderTwo, ReorderThreeSharp, ReorderFour } from '@vicons/ionicons5';
import { renderIcon } from '@/utils/index';

const routes: Array<RouteRecordRaw> = [
  {
    path: '/multilevel',
    name: 'multilevel',
    component: Layout,
    redirect: '/multilevel/level_2/level_2_1/level_2_2',
    meta: {
      title: '多级菜单',
      icon: renderIcon(MenuOutlined),
      sort: 12,
    },
    children: [
      {
        path: 'level_2',
        name: 'level2',
        meta: {
          title: '二级菜单',
          icon: renderIcon(ReorderTwo),
        },
        component: ParentLayout,
        children: [
          {
            path: 'level_2_1',
            name: `Level21`,
            meta: {
              title: '三级菜单',
              keepAlive: true,
              icon: renderIcon(ReorderThreeSharp),
            },
            component: ParentLayout,
            children: [
              {
                path: 'level_2_2',
                name: `Level22`,
                meta: {
                  title: '四级菜单',
                  keepAlive: true,
                  icon: renderIcon(ReorderFour),
                },
                component: () => import('@/views/multistage/one.vue'),
              },
            ],
          },
        ],
      },
    ],
  },
];

export default routes;
