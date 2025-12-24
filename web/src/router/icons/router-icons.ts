import { renderIcon } from '@/utils/index';
import * as AntdIcons from '@vicons/antd';
import * as Ionicons5 from '@vicons/ionicons5';

// 前端路由图标映射表 - 动态注册所有图标
export const constantRouterIcon: Record<string, any> = {};

// 动态注册所有 Ant Design 图标
for (const [name, icon] of Object.entries(AntdIcons)) {
  constantRouterIcon[name] = renderIcon(icon);
}

// 动态注册所有 Ionicons5 图标
for (const [name, icon] of Object.entries(Ionicons5)) {
  constantRouterIcon[name] = renderIcon(icon);
}
