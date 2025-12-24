import { RouteRecordRaw } from 'vue-router';
import { Layout } from '@/router/base';
import { CubeOutline } from '@vicons/ionicons5';
import { renderIcon } from '@/utils';

const routes: Array<RouteRecordRaw> = [
  {
    path: '/search',
    name: 'Search',
    component: Layout,
    redirect: '/search/article',
    meta: {
      title: '搜索页面',
      icon: renderIcon(CubeOutline),
      sort: 4,
    },
    children: [
      {
        path: 'article',
        name: 'SearchArticle',
        meta: {
          title: '文章页面',
        },
        component: () => import('@/views/search/search-article/search-article.vue'),
      },
      {
        path: 'video',
        name: 'SearchVideo',
        meta: {
          title: '视频页面',
        },
        component: () => import('@/views/search/search-video/search-video.vue'),
      },
      {
        path: 'make',
        name: 'SearchMake',
        meta: {
          title: '预约页面',
        },
        component: () => import('@/views/search/search-make/search-make.vue'),
      },
    ],
  },
];

export default routes;
