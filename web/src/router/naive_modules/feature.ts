import { RouteRecordRaw } from 'vue-router';
import { Layout, ParentLayout } from '@/router/base';
import { ControlOutlined } from '@vicons/antd';
import { renderIcon } from '@/utils';

const routes: Array<RouteRecordRaw> = [
  {
    path: 'feature',
    name: 'Feature',
    component: ParentLayout,
    redirect: '/feature/authority',
    meta: {
      title: '功能示例',
      icon: renderIcon(ControlOutlined),
      sort: 8,
    },
    children: [
      {
        path: 'authority',
        name: 'Authority',
        component: () => import('@/views/feature/authority/authority.vue'),
        meta: {
          title: '权限判断',
        },
      },
      {
        path: 'download',
        name: 'Download',
        component: () => import('@/views/feature/download/download.vue'),
        meta: {
          title: '文件下载',
        },
      },
      {
        path: 'context-menus',
        name: 'ContextMenus',
        component: () => import('@/views/feature/context-menus/context-menus.vue'),
        meta: {
          title: '右键菜单',
        },
      },
      {
        path: 'copy',
        name: 'copy',
        component: () => import('@/views/feature/copy/copy.vue'),
        meta: {
          title: '剪贴板',
        },
      },
      {
        path: 'print',
        name: 'print',
        component: () => import('@/views/feature/print/print.vue'),
        meta: {
          title: '打印',
        },
      },
      {
        path: 'scrollbar',
        name: 'scrollbar',
        component: () => import('@/views/feature/scrollbar/scrollbar.vue'),
        meta: {
          title: '滚动条',
        },
      },
      {
        path: 'excel',
        name: 'Excel',
        meta: {
          title: 'Excel',
        },
        component: ParentLayout,
        children: [
          {
            path: 'choiceExport',
            name: 'choiceExport',
            component: () => import('@/views/feature/excel/choiceExport.vue'),
            meta: {
              title: 'Format',
            },
          },
          {
            path: 'jsonExport',
            name: 'jsonExport',
            component: () => import('@/views/feature/excel/jsonExport.vue'),
            meta: {
              title: 'Json',
            },
          },
        ],
      },
      {
        path: 'tagsAction',
        name: 'TagsAction',
        meta: {
          title: '多页签操作',
        },
        component: () => import('@/views/feature/tags/tagsAction.vue'),
      },
    ],
  },
];

export default routes;
