import { RouteRecordRaw } from 'vue-router';
import { Layout } from '@/router/base';
import {
  LayoutOutlined,
  CalendarOutlined,
  CodeOutlined,
  ProfileOutlined,
  InsertRowAboveOutlined,
  TableOutlined,
  ScheduleOutlined,
} from '@vicons/antd';
import { FileTrayOutline, LaptopOutline, LogoBuffer, ReorderFour } from '@vicons/ionicons5';
import { renderIcon } from '@/utils';

const routes: Array<RouteRecordRaw> = [
  {
    path: '/common',
    name: 'common',
    component: Layout,
    redirect: '/common/curd',
    meta: {
      title: '常用模板',
      icon: renderIcon(LayoutOutlined),
      sort: 10,
    },
    children: [
      {
        path: 'general',
        name: 'layout_general',
        meta: {
          title: '常规布局',
          icon: renderIcon(ProfileOutlined),
        },
        component: () => import('@/views/layout/general/general.vue'),
      },
      {
        path: 'around',
        name: 'layout_around',
        meta: {
          title: '左右布局',
          icon: renderIcon(CodeOutlined),
        },
        component: () => import('@/views/layout/around/around.vue'),
      },
      {
        path: 'upand',
        name: 'layout_upand',
        meta: {
          title: '上下布局',
          icon: renderIcon(CalendarOutlined),
        },
        component: () => import('@/views/layout/upand/upand.vue'),
      },
      {
        path: 'curd',
        name: 'common_curd',
        meta: {
          title: 'Curd页面',
          icon: renderIcon(InsertRowAboveOutlined),
        },
        component: () => import('@/views/common/curd/curd.vue'),
      },
      {
        path: 'tabtable',
        name: 'common_tabtable',
        meta: {
          title: 'Tab表格',
          icon: renderIcon(TableOutlined),
        },
        component: () => import('@/views/common/tabTable/tabTable.vue'),
      },
      {
        path: 'detail',
        name: 'common_detail',
        meta: {
          title: '详情页',
          icon: renderIcon(ScheduleOutlined),
        },
        component: () => import('@/views/common/detail/detail.vue'),
      },
      {
        path: 'popupform',
        name: 'common_popupform',
        meta: {
          title: '弹窗表单',
          icon: renderIcon(LaptopOutline),
        },
        component: () => import('@/views/common/popupForm/popupForm.vue'),
      },
      {
        path: 'drawerform',
        name: 'common_drawerform',
        meta: {
          title: '抽屉表单',
          icon: renderIcon(FileTrayOutline),
        },
        component: () => import('@/views/common/drawerForm/drawerForm.vue'),
      },
      {
        path: 'packetform',
        name: 'common_packetform',
        meta: {
          title: '分组表单',
          icon: renderIcon(LogoBuffer),
        },
        component: () => import('@/views/common/packetForm/packetForm.vue'),
      },
      {
        path: 'dynamicform',
        name: 'common_dynamicform',
        meta: {
          title: '动态表单',
          icon: renderIcon(ReorderFour),
        },
        component: () => import('@/views/common/dynamicForm/dynamicForm.vue'),
      },
    ],
  },
];

export default routes;
