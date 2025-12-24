import type { RouteRecordRaw } from 'vue-router';
import { Layout } from '@/router/base';
import { renderIcon } from '@/utils/index';
import { LanguageOutline } from '@vicons/ionicons5';
import { t } from '@/hooks/web/useI18n';

const routes: Array<RouteRecordRaw> = [
  {
    path: 'i18n',
    name: 'I18n',
    redirect: '/I18n/basic',
    component: ParentLayout,
    meta: {
      title: t('common.i18nText'),
      icon: renderIcon(LanguageOutline),
      alwaysShow: true,
      sort: 13,
      activeMenu: 'i18nBasic',
    },
    children: [
      {
        path: 'basic',
        name: 'i18nBasic',
        meta: {
          title: t('common.i18nText'),
          icon: renderIcon(LanguageOutline),
        },
        component: () => import('@/views/i18n/index.vue'),
      },
    ],
  },
];

export default routes;
