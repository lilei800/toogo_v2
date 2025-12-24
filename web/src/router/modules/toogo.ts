import { RouteRecordRaw } from 'vue-router';
import { Layout, ParentLayout } from '@/router/constant';
import { renderIcon } from '@/utils/index';
import {
  DashboardOutlined,
  AppstoreOutlined,
  SafetyCertificateOutlined,
  SettingOutlined,
  CodeOutlined,
  BarChartOutlined,
  ControlOutlined,
  RobotOutlined,
  ApiOutlined,
  WalletOutlined,
  UserOutlined,
  TeamOutlined,
  ProjectOutlined,
  CreditCardOutlined,
  RiseOutlined,
  StarOutlined,
  DollarOutlined,
  FileTextOutlined,
  MonitorOutlined,
  DesktopOutlined,
  ToolOutlined,
  PartitionOutlined,
  ApartmentOutlined,
  MenuOutlined,
  LockOutlined,
  DatabaseOutlined,
  BookOutlined,
  ClockCircleOutlined,
  StopOutlined,
  GlobalOutlined,
  CloudUploadOutlined,
  NotificationOutlined,
  BugOutlined,
  BlockOutlined,
  ThunderboltOutlined,
  SwapOutlined,
  LineChartOutlined,
  InfoCircleOutlined,
} from '@vicons/antd';

// 完整的路由配置
const routes: RouteRecordRaw[] = [
  // ============ 仪表盘 ============
  {
    path: '/dashboard',
    name: 'Dashboard',
    component: Layout,
    redirect: '/dashboard/console',
    meta: {
      title: '仪表盘',
      icon: renderIcon(DashboardOutlined),
      sort: 1,
    },
    children: [
      {
        path: 'console',
        name: 'dashboard_console',
        component: () => import('@/views/dashboard/console/console.vue'),
        meta: {
          title: '主控台',
          icon: renderIcon(DesktopOutlined),
        },
      },
      {
        path: 'workplace',
        name: 'dashboard_workplace',
        component: () => import('@/views/dashboard/workplace/workplace.vue'),
        meta: {
          title: '工作台',
          icon: renderIcon(ToolOutlined),
        },
      },
    ],
  },
  // ============ 量化交易 - 用户菜单 ============
  // 注意：此配置为静态备用，实际菜单应通过后台"权限管理-菜单管理"动态配置
  {
    path: '/toogo',
    name: 'ToogoRoot',
    component: Layout,
    redirect: '/toogo/dashboard',
    meta: {
      title: '量化交易',
      icon: renderIcon(BarChartOutlined),
      sort: 2,
      permissions: [],
    },
    children: [
      // 兼容历史/菜单配置：有些环境把组件路径写成 /toogo/dashboard/index
      // 实际路由为 /toogo/dashboard（组件文件名 index.vue 与 URL 无关）
      {
        path: 'dashboard/index',
        name: 'ToogoDashboardIndexRedirect',
        redirect: '/toogo/dashboard',
        meta: {
          hidden: true,
        },
      },
      {
        path: 'dashboard',
        name: 'ToogoDashboard',
        component: () => import('@/views/toogo/dashboard/index.vue'),
        meta: {
          title: '控制台',
          icon: renderIcon(DashboardOutlined),
        },
      },
      {
        path: 'robot',
        name: 'ToogoRobot',
        component: () => import('@/views/toogo/robot/index.vue'),
        meta: {
          title: '我的机器人',
          icon: renderIcon(RobotOutlined),
        },
      },
      {
        path: 'robot/create',
        name: 'ToogoRobotCreate',
        component: () => import('@/views/toogo/robot/create.vue'),
        meta: {
          title: '创建机器人',
          hidden: true,
          activeMenu: '/toogo/robot',
        },
      },
      {
        path: 'strategy',
        name: 'ToogoStrategy',
        component: () => import('@/views/toogo/strategy/index.vue'),
        meta: {
          title: '策略模板',
          icon: renderIcon(SettingOutlined),
        },
      },
      {
        path: 'strategy/list',
        name: 'StrategyList',
        component: () => import('@/views/toogo/strategy/list.vue'),
        meta: {
          title: '策略列表',
          hidden: true,
          activeMenu: '/toogo/strategy',
        },
      },
      {
        path: 'api',
        name: 'ToogoApi',
        component: () => import('@/views/toogo/api/index.vue'),
        meta: {
          title: 'API配置',
          icon: renderIcon(ApiOutlined),
        },
      },
      {
        path: 'wallet',
        name: 'ToogoWallet',
        component: ParentLayout,
        redirect: '/toogo/wallet/overview',
        meta: {
          title: '钱包中心',
          icon: renderIcon(WalletOutlined),
        },
        children: [
          {
            path: 'overview',
            name: 'WalletOverview',
            component: () => import('@/views/toogo/finance/index.vue'),
            meta: {
              title: '资产总览',
              icon: renderIcon(DashboardOutlined),
            },
          },
          {
            path: 'subscription',
            name: 'WalletSubscription',
            component: () => import('@/views/toogo/subscription/index.vue'),
            meta: {
              title: '订阅套餐',
              icon: renderIcon(CreditCardOutlined),
            },
          },
          {
            path: 'order-history',
            name: 'WalletOrderHistory',
            component: () => import('@/views/toogo/wallet/order-history.vue'),
            meta: {
              title: '交易明细',
              icon: renderIcon(FileTextOutlined),
            },
          },
        ],
      },
      {
        path: 'promote',
        name: 'ToogoPromote',
        component: ParentLayout,
        redirect: '/toogo/promote/team',
        meta: {
          title: '推广中心',
          icon: renderIcon(TeamOutlined),
        },
        children: [
          {
            path: 'team',
            name: 'PromoteTeam',
            component: () => import('@/views/toogo/team/index.vue'),
            meta: {
              title: '我的团队',
              icon: renderIcon(TeamOutlined),
            },
          },
          {
            path: 'commission',
            name: 'PromoteCommission',
            component: () => import('@/views/toogo/commission/index.vue'),
            meta: {
              title: '佣金记录',
              icon: renderIcon(RiseOutlined),
            },
          },
        ],
      },
    ],
  },
  // ============ 量化管理 ============
  {
    path: '/toogo-admin',
    name: 'ToogoAdmin',
    component: Layout,
    redirect: '/toogo-admin/user',
    meta: {
      title: '量化管理',
      icon: renderIcon(ControlOutlined),
      sort: 3,
    },
    children: [
      {
        path: 'user',
        name: 'ToogoAdminUser',
        component: () => import('@/views/toogo/admin/user/index.vue'),
        meta: {
          title: '用户管理',
          icon: renderIcon(UserOutlined),
        },
      },
      {
        path: 'plan',
        name: 'ToogoAdminPlan',
        component: () => import('@/views/toogo/admin/plan/index.vue'),
        meta: {
          title: '套餐管理',
          icon: renderIcon(CreditCardOutlined),
        },
      },
      {
        path: 'vip-level',
        name: 'ToogoAdminVipLevel',
        component: () => import('@/views/toogo/admin/vip-level/index.vue'),
        meta: {
          title: 'VIP等级',
          icon: renderIcon(StarOutlined),
        },
      },
      {
        path: 'agent-level',
        name: 'ToogoAdminAgentLevel',
        component: () => import('@/views/toogo/admin/agent-level/index.vue'),
        meta: {
          title: '代理等级',
          icon: renderIcon(TeamOutlined),
        },
      },
      {
        path: 'strategy',
        name: 'ToogoAdminStrategy',
        component: () => import('@/views/toogo/admin/strategy/index.vue'),
        meta: {
          title: '官方策略管理',
          icon: renderIcon(PartitionOutlined),
        },
      },
      {
        path: 'withdraw',
        name: 'ToogoAdminWithdraw',
        component: () => import('@/views/toogo/admin/withdraw/index.vue'),
        meta: {
          title: '提现管理',
          icon: renderIcon(WalletOutlined),
        },
      },
      {
        path: 'deposit',
        name: 'ToogoAdminDeposit',
        component: () => import('@/views/toogo/admin/deposit/index.vue'),
        meta: {
          title: '充值管理',
          icon: renderIcon(DollarOutlined),
        },
      },
      {
        path: 'subscription',
        name: 'ToogoAdminSubscription',
        component: () => import('@/views/toogo/admin/subscription/index.vue'),
        meta: {
          title: '订阅记录',
          icon: renderIcon(FileTextOutlined),
        },
      },
      {
        path: 'commission-log',
        name: 'ToogoAdminCommissionLog',
        component: () => import('@/views/toogo/admin/commission-log/index.vue'),
        meta: {
          title: '佣金记录',
          icon: renderIcon(RiseOutlined),
        },
      },
      {
        path: 'config',
        name: 'ToogoAdminConfig',
        component: () => import('@/views/toogo/admin/config/index.vue'),
        meta: {
          title: '系统配置',
          icon: renderIcon(SettingOutlined),
        },
      },
      {
        path: 'robot',
        name: 'ToogoAdminRobot',
        component: () => import('@/views/toogo/admin/robot/index.vue'),
        meta: {
          title: '机器人监控',
          icon: renderIcon(RobotOutlined),
        },
      },
      {
        path: 'power',
        name: 'ToogoAdminPower',
        component: () => import('@/views/toogo/admin/power/index.vue'),
        meta: {
          title: '算力配置',
          icon: renderIcon(ThunderboltOutlined),
        },
      },
      {
        path: 'symbol',
        name: 'ToogoAdminSymbol',
        component: () => import('@/views/toogo/admin/symbol/index.vue'),
        meta: {
          title: '交易对管理',
          icon: renderIcon(SwapOutlined),
        },
      },
      {
        path: 'volatility-config',
        name: 'ToogoAdminVolatilityConfig',
        component: () => import('@/views/toogo/admin/volatility-config/index.vue'),
        meta: {
          title: '波动率配置',
          icon: renderIcon(LineChartOutlined),
        },
      },
    ],
  },
  // ============ 组织管理 ============
  {
    path: '/org',
    name: 'Org',
    component: Layout,
    redirect: '/org/user',
    meta: {
      title: '组织管理',
      icon: renderIcon(AppstoreOutlined),
      sort: 20,
    },
    children: [
      {
        path: 'user',
        name: 'user',
        component: () => import('@/views/org/user/user.vue'),
        meta: {
          title: '后台用户',
          icon: renderIcon(UserOutlined),
        },
      },
      {
        path: 'dept',
        name: 'org_dept',
        component: () => import('@/views/org/dept/dept.vue'),
        meta: {
          title: '部门管理',
          icon: renderIcon(ApartmentOutlined),
        },
      },
      {
        path: 'post',
        name: 'org_post',
        component: () => import('@/views/org/post/post.vue'),
        meta: {
          title: '岗位管理',
          icon: renderIcon(TeamOutlined),
        },
      },
    ],
  },
  // ============ 权限管理 ============
  {
    path: '/permission',
    name: 'Permission',
    component: Layout,
    redirect: '/permission/menu',
    meta: {
      title: '权限管理',
      icon: renderIcon(SafetyCertificateOutlined),
      sort: 40,
    },
    children: [
      {
        path: 'menu',
        name: 'permission_menu',
        component: () => import('@/views/permission/menu/menu.vue'),
        meta: {
          title: '菜单权限',
          icon: renderIcon(MenuOutlined),
        },
      },
      {
        path: 'role',
        name: 'permission_role',
        component: () => import('@/views/permission/role/role.vue'),
        meta: {
          title: '角色权限',
          icon: renderIcon(LockOutlined),
        },
      },
    ],
  },
  // ============ 系统设置 ============
  {
    path: '/system',
    name: 'System',
    component: Layout,
    redirect: '/system/config',
    meta: {
      title: '系统设置',
      icon: renderIcon(SettingOutlined),
      sort: 120,
    },
    children: [
      {
        path: 'config',
        name: 'system_config',
        component: () => import('@/views/system/config/system.vue'),
        meta: {
          title: '配置管理',
          icon: renderIcon(SettingOutlined),
        },
      },
      {
        path: 'dict',
        name: 'system_dict',
        component: () => import('@/views/system/dict/index.vue'),
        meta: {
          title: '字典管理',
          icon: renderIcon(BookOutlined),
        },
      },
      {
        path: 'cron',
        name: 'system_cron',
        component: () => import('@/views/system/cron/index.vue'),
        meta: {
          title: '定时任务',
          icon: renderIcon(ClockCircleOutlined),
        },
      },
      {
        path: 'blacklist',
        name: 'system_blacklist',
        component: () => import('@/views/system/blacklist/index.vue'),
        meta: {
          title: '黑名单',
          icon: renderIcon(StopOutlined),
        },
      },
    ],
  },
  // ============ 系统监控 ============
  {
    path: '/monitor',
    name: 'Monitors',
    component: Layout,
    redirect: '/monitor/online',
    meta: {
      title: '系统监控',
      icon: renderIcon(BarChartOutlined),
      sort: 110,
    },
    children: [
      {
        path: 'online',
        name: 'monitor_online',
        component: () => import('@/views/monitor/online/index.vue'),
        meta: {
          title: '在线用户',
          icon: renderIcon(UserOutlined),
        },
      },
      {
        path: 'serve-monitor',
        name: 'monitor_serve_monitor',
        component: () => import('@/views/monitor/serve-monitor/index.vue'),
        meta: {
          title: '服务监控',
          icon: renderIcon(MonitorOutlined),
        },
      },
      {
        path: 'serve-log',
        name: 'monitor_serve_log',
        component: () => import('@/views/monitor/serve-log/index.vue'),
        meta: {
          title: '服务日志',
          icon: renderIcon(FileTextOutlined),
        },
      },
      {
        path: 'netconn',
        name: 'monitor_netconn',
        component: () => import('@/views/monitor/netconn/index.vue'),
        meta: {
          title: '在线服务',
          icon: renderIcon(GlobalOutlined),
        },
      },
    ],
  },
  // ============ 系统应用 ============
  {
    path: '/apply',
    name: 'Applys',
    component: Layout,
    redirect: '/apply/provinces',
    meta: {
      title: '系统应用',
      icon: renderIcon(AppstoreOutlined),
      sort: 100,
    },
    children: [
      {
        path: 'provinces',
        name: 'apply_provinces',
        component: () => import('@/views/apply/provinces/list.vue'),
        meta: {
          title: '地区编码',
          icon: renderIcon(GlobalOutlined),
        },
      },
      {
        path: 'attachment',
        name: 'apply_attachment',
        component: () => import('@/views/apply/attachment/index.vue'),
        meta: {
          title: '附件管理',
          icon: renderIcon(CloudUploadOutlined),
        },
      },
      {
        path: 'notice',
        name: 'apply_notice',
        component: () => import('@/views/apply/notice/index.vue'),
        meta: {
          title: '通知公告',
          icon: renderIcon(NotificationOutlined),
        },
      },
    ],
  },
  // ============ 开发工具 ============
  {
    path: '/develop',
    name: 'Develops',
    component: Layout,
    redirect: '/develop/code',
    meta: {
      title: '开发工具',
      icon: renderIcon(CodeOutlined),
      sort: 210,
    },
    children: [
      {
        path: 'code',
        name: 'develop_code',
        component: () => import('@/views/develop/code/index.vue'),
        meta: {
          title: '代码生成',
          icon: renderIcon(CodeOutlined),
        },
      },
      {
        path: 'addons',
        name: 'develop_addons',
        component: () => import('@/views/develop/addons/index.vue'),
        meta: {
          title: '插件管理',
          icon: renderIcon(BlockOutlined),
        },
      },
    ],
  },
  // ============ 关于 ============
  {
    path: '/about',
    name: 'about',
    component: Layout,
    redirect: '/about/index',
    meta: {
      title: '关于',
      icon: renderIcon(ProjectOutlined),
      sort: 9000,
    },
    children: [
      {
        path: 'index',
        name: 'about_index',
        component: () => import('@/views/about/index.vue'),
        meta: {
          title: '关于详情',
          icon: renderIcon(InfoCircleOutlined),
        },
      },
    ],
  },
  // ============ 个人中心 ============
  {
    path: '/home',
    name: 'home',
    component: Layout,
    redirect: '/home/account',
    meta: {
      title: '个人中心',
      hidden: true,
    },
    children: [
      {
        path: 'account',
        name: 'home_account',
        component: () => import('@/views/home/account/account.vue'),
        meta: {
          title: '个人设置',
        },
      },
      {
        path: 'message',
        name: 'home_message',
        component: () => import('@/views/home/message/message.vue'),
        meta: {
          title: '我的消息',
        },
      },
    ],
  },
];

export default routes;


