import { RouteRecordRaw } from 'vue-router';
import { Layout, ParentLayout } from '@/router/base';
import {
  WalletOutlined,
  FormOutlined,
  InstagramOutlined,
  AlertOutlined,
  TableOutlined,
  UploadOutlined,
  BorderOutlined,
  PicRightOutlined,
  DragOutlined,
  AlignCenterOutlined,
  QrcodeOutlined,
  SelectOutlined,
  CheckSquareOutlined,
  AlignLeftOutlined,
  CodeOutlined,
  TagsOutlined,
  CreditCardOutlined,
  FileTextOutlined,
} from '@vicons/antd';
import {
  CalendarOutline,
  CardOutline,
  RadioButtonOn,
  FileTrayOutline,
  TimeOutline,
  LockClosedOutline,
  EllipseOutline,
} from '@vicons/ionicons5';
import { renderIcon } from '@/utils';
import { h } from 'vue';
import { NBadge } from 'naive-ui';

const routeName = 'comp';

const routes: Array<RouteRecordRaw> = [
  {
    path: 'comp',
    name: routeName,
    component: ParentLayout,
    redirect: '/comp/table/basic',
    meta: {
      title: '组件示例',
      extra: () => h(NBadge, { dot: true, type: 'success' }),
      icon: renderIcon(WalletOutlined),
      sort: 9,
    },
    children: [
      {
        path: 'picfiles',
        name: `PicFiles`,
        meta: {
          title: '文件管理',
          icon: renderIcon(FileTextOutlined),
          extra: () => h(NBadge, { dot: true, type: 'success' }),
        },
        component: () => import('@/views/comp/picfiles/index.vue'),
      },
      {
        path: 'collapseCard',
        name: `BasicCard`,
        meta: {
          title: '折叠卡片',
          icon: renderIcon(CreditCardOutlined),
        },
        component: () => import('@/views/comp/collapseCard/collapseCard.vue'),
      },
      {
        path: 'prosearch',
        name: `${routeName}_prosearch`,
        meta: {
          title: '搜索表单',
          icon: renderIcon(CodeOutlined),
        },
        component: () => import('@/views/comp/proSearchGroup/proSearchGroup.vue'),
      },
      {
        path: 'table',
        name: `${routeName}_table`,
        redirect: '/comp/table/basic',
        component: ParentLayout,
        meta: {
          title: '表格示例',
          icon: renderIcon(TableOutlined),
        },
        children: [
          {
            path: 'basic',
            name: `${routeName}_table_basic`,
            meta: {
              title: '基础表格',
            },
            component: () => import('@/views/comp/table/basic.vue'),
          },
          {
            path: 'editCell',
            name: `${routeName}_table_editCell`,
            meta: {
              title: '编辑单元格',
            },
            component: () => import('@/views/comp/table/editCell.vue'),
          },
          {
            path: 'editRow',
            name: `${routeName}_table_editRow`,
            meta: {
              title: '编辑行',
            },
            component: () => import('@/views/comp/table/editRow.vue'),
          },
          {
            path: 'manualTable',
            name: `${routeName}_table_manualTable`,
            meta: {
              title: '手动表格',
            },
            component: () => import('@/views/comp/table/manualTable.vue'),
          },
        ],
      },
      {
        path: 'form',
        name: `${routeName}_form`,
        redirect: '/comp/form/basic',
        component: ParentLayout,
        meta: {
          title: '表单示例',
          icon: renderIcon(FormOutlined),
        },
        children: [
          {
            path: 'basic',
            name: `${routeName}_form_basic`,
            meta: {
              title: '基础表单',
            },
            component: () => import('@/views/comp/form/basic.vue'),
          },
          {
            path: 'group',
            name: `${routeName}_form_group`,
            meta: {
              title: '分组表单',
            },
            component: () => import('@/views/comp/form/group.vue'),
          },
          {
            path: 'useForm',
            name: `useForm`,
            meta: {
              title: 'useForm',
            },
            component: () => import('@/views/comp/form/useForm.vue'),
          },
        ],
      },
      {
        path: 'modal',
        name: `${routeName}_modal`,
        meta: {
          title: '弹窗表单',
          icon: renderIcon(BorderOutlined),
        },
        component: () => import('@/views/comp/modal/index.vue'),
      },
      {
        path: 'richtext',
        name: `richtext`,
        meta: {
          title: '富文本',
          icon: renderIcon(PicRightOutlined),
        },
        component: () => import('@/views/comp/richtext/vue-quill.vue'),
      },
      {
        path: 'drag',
        name: `Drag`,
        meta: {
          title: '拖拽',
          icon: renderIcon(DragOutlined),
        },
        component: () => import('@/views/comp/drag/index.vue'),
      },
      {
        path: 'region',
        name: `Region`,
        meta: {
          title: '地区',
          icon: renderIcon(AlignCenterOutlined),
        },
        component: () => import('@/views/comp/region/index.vue'),
      },
      {
        path: 'cropper',
        name: `Cropper`,
        meta: {
          title: '图片裁剪',
          icon: renderIcon(InstagramOutlined),
        },
        component: () => import('@/views/comp/cropper/index.vue'),
      },
      {
        path: 'qrcode',
        name: `Qrcode`,
        meta: {
          title: '二维码',
          icon: renderIcon(QrcodeOutlined),
        },
        component: () => import('@/views/comp/qrcode/index.vue'),
      },
      {
        path: 'password',
        name: `Password`,
        meta: {
          title: '密码强度',
          icon: renderIcon(LockClosedOutline),
        },
        component: () => import('@/views/comp/password/index.vue'),
      },
      {
        path: 'select',
        name: `Select`,
        meta: {
          title: '选择器',
          icon: renderIcon(SelectOutlined),
        },
        component: () => import('@/views/comp/select/BasicSelect.vue'),
      },
      {
        path: 'tableselect',
        name: `TableSelect`,
        meta: {
          title: '表格选择器',
          icon: renderIcon(SelectOutlined),
        },
        component: () => import('@/views/comp/tableSelect/tableSelect.vue'),
      },
      {
        path: 'paragraph',
        name: `${routeName}_paragraph`,
        meta: {
          title: '排印扩展',
          icon: renderIcon(CalendarOutline),
        },
        component: () => import('@/views/comp/paragraph/basic.vue'),
      },
      {
        path: 'checkcard',
        name: `${routeName}_checkcard`,
        meta: {
          title: '可选卡片',
          icon: renderIcon(CardOutline),
        },
        component: () => import('@/views/comp/checkCard/basic.vue'),
      },
      {
        path: 'checkButton',
        name: `${routeName}_checkButton`,
        meta: {
          title: '可选按钮',
          icon: renderIcon(RadioButtonOn),
        },
        component: () => import('@/views/comp/checkButton/checkButton.vue'),
      },
      {
        path: 'loading',
        name: `${routeName}_loading`,
        meta: {
          title: 'Loading',
          icon: renderIcon(TimeOutline),
        },
        component: () => import('@/views/comp/loading/basic.vue'),
      },
      {
        path: 'iconpicker',
        name: `${routeName}_iconpicker`,
        meta: {
          title: '图标选择器',
          icon: renderIcon(InstagramOutlined),
        },
        component: () => import('@/views/comp/iconpicker/iconpicker.vue'),
      },
      {
        path: 'modalform',
        name: `${routeName}_modalform`,
        meta: {
          title: '弹窗表单',
          icon: renderIcon(AlertOutlined),
        },
        component: () => import('@/views/comp/modalForm/modalForm.vue'),
      },
      {
        path: 'drawerform',
        name: `${routeName}_drawerform`,
        meta: {
          title: '抽屉表单',
          icon: renderIcon(FileTrayOutline),
        },
        component: () => import('@/views/comp/drawerForm/drawerForm.vue'),
      },
      {
        path: 'formTable',
        name: `${routeName}_formTable`,
        meta: {
          title: '表单表格',
          icon: renderIcon(TableOutlined),
        },
        component: () => import('@/views/comp/formTable/formTable.vue'),
      },
      {
        path: 'upload',
        name: `${routeName}_upload`,
        meta: {
          title: '上传图片',
          icon: renderIcon(UploadOutlined),
        },
        component: () => import('@/views/comp/upload/upload.vue'),
      },
      {
        path: 'ncurd',
        name: `${routeName}_ncurd`,
        meta: {
          title: 'NCURD',
          icon: renderIcon(TableOutlined),
        },
        component: () => import('@/views/comp/ncurd/ncurd.vue'),
      },
      {
        path: 'procheckbox',
        name: `${routeName}_procheckbox`,
        meta: {
          title: '远程复选框',
          icon: renderIcon(CheckSquareOutlined),
        },
        component: () => import('@/views/comp/proCheckbox/proCheckbox.vue'),
      },
      {
        path: 'proradio',
        name: `${routeName}_proradio`,
        meta: {
          title: '远程单选框',
          icon: renderIcon(EllipseOutline),
        },
        component: () => import('@/views/comp/proRadio/proRadio.vue'),
      },
      {
        path: 'procascader',
        name: `${routeName}_proCascader`,
        meta: {
          title: '省市区联动',
          icon: renderIcon(AlignLeftOutlined),
        },
        component: () => import('@/views/comp/proCascader/proCascader.vue'),
      },
      {
        path: 'codeInput',
        name: `${routeName}_codeInput`,
        meta: {
          title: '验证码输入框',
          icon: renderIcon(CodeOutlined),
        },
        component: () => import('@/views/comp/proCode/proCode.vue'),
      },
      {
        path: 'tagGroup',
        name: `TagGroup`,
        meta: {
          title: '标签组',
          icon: renderIcon(TagsOutlined),
          extra: () => h(NBadge, { dot: true, type: 'success' }),
        },
        component: () => import('@/views/comp/tagGroup/tagGroup.vue'),
      },
    ],
  },
];

export default routes;
