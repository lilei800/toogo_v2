import { toRaw, unref } from 'vue';
import { defineStore } from 'pinia';
import { RouteRecordRaw } from 'vue-router';
import { store } from '@/store';
import { asyncRoutes, constantRouter } from '@/router/index';
import { generatorDynamicRouter } from '@/router/create';
import { useProjectSetting } from '@/hooks/setting/useProjectSetting';
import { Layout, ParentLayout } from '@/router/constant';

/**
 * 前端定义的必要隐藏路由（BACK模式下后端可能没有配置这些）
 * 这些路由必须被注册以支持页面内部跳转
 * 格式：{ parentPath: string, childRoute: RouteRecordRaw }
 */
const frontendHiddenChildRoutes: Array<{
  parentPath: string;
  childRoute: RouteRecordRaw;
}> = [
  // 兼容历史：部分环境/旧菜单把控制台写成 /toogo/dashboard/index
  // 但前端实际路由为 /toogo/dashboard
  {
    parentPath: '/toogo',
    childRoute: {
      path: 'dashboard/index',
      name: 'ToogoDashboardIndexRedirect',
      redirect: '/toogo/dashboard',
      meta: {
        title: '控制台',
        hidden: true,
        ignoreAuth: true,
      },
    },
  },
  // 策略列表页面（从策略模板进入）
  {
    parentPath: '/toogo',
    childRoute: {
      path: 'strategy/list',
      name: 'StrategyList',
      component: () => import('@/views/toogo/strategy/list.vue'),
      meta: {
        title: '策略列表',
        hidden: true,
        activeMenu: '/toogo/strategy/my',
        ignoreAuth: true,
      },
    },
  },
  // 创建机器人页面（作为 /toogo/robot 的子路由）
  {
    parentPath: '/toogo/robot',
    childRoute: {
      path: 'create',
      name: 'ToogoRobotCreate',
      component: () => import('@/views/toogo/robot/create.vue'),
      meta: {
        title: '创建机器人',
        hidden: true,
        activeMenu: '/toogo/robot',
        ignoreAuth: true,
      },
    },
  },
];

/**
 * 检查某个路由（通过name或path）是否已存在于路由列表中
 */
function isRouteExists(routes: RouteRecordRaw[], name?: string, path?: string): boolean {
  for (const route of routes) {
    // 检查name匹配
    if (name && route.name === name) {
      return true;
    }
    // 检查path匹配（支持相对路径和绝对路径）
    if (path) {
      const routePath = route.path as string;
      if (routePath === path || routePath === '/' + path || path === '/' + routePath) {
        return true;
      }
    }
    // 递归检查子路由
    if (route.children && route.children.length > 0) {
      if (isRouteExists(route.children as RouteRecordRaw[], name, path)) {
        return true;
      }
    }
  }
  return false;
}

/**
 * 查找父路由（递归查找，支持嵌套路由）
 * 支持多种路径格式匹配
 */
function findParentRoute(routes: RouteRecordRaw[], parentPath: string): RouteRecordRaw | null {
  // 规范化路径：确保以/开头
  const normalizedParentPath = parentPath.startsWith('/') ? parentPath : '/' + parentPath;
  
  for (const route of routes) {
    const routePath = route.path as string;
    if (!routePath) continue;
    
    // 规范化路由路径
    const normalizedRoutePath = routePath.startsWith('/') ? routePath : '/' + routePath;
    
    // 精确匹配路径
    if (normalizedRoutePath === normalizedParentPath) {
      return route;
    }
    
    // 递归查找子路由
    if (route.children && route.children.length > 0) {
      const found = findParentRoute(route.children as RouteRecordRaw[], parentPath);
      if (found) {
        return found;
      }
    }
    
    // 如果当前路由是父路由的一部分（如 /toogo 匹配 /toogo/robot）
    // 需要确保路径匹配且不是完全相同的路径
    if (normalizedRoutePath !== normalizedParentPath && normalizedParentPath.startsWith(normalizedRoutePath + '/')) {
      // 检查是否有子路由精确匹配
      if (route.children && route.children.length > 0) {
        const found = findParentRoute(route.children as RouteRecordRaw[], parentPath);
        if (found) {
          return found;
        }
      }
      // 如果没有找到精确匹配的子路由，返回当前路由作为父路由
      // 这样可以支持将子路由添加到父路由的children中
      return route;
    }
  }
  return null;
}

/**
 * 合并前端隐藏路由到路由列表
 * 用于确保BACK模式下，一些必须的隐藏路由能被正确注册
 * 策略：先尝试将子路由合并到已有的父路由中，如果父路由不存在则创建新的
 */
function mergeFrontendHiddenRoutes(routes: RouteRecordRaw[]): RouteRecordRaw[] {
  const mergedRoutes = [...routes];
  
  // 按父路径分组
  const groupedByParent = new Map<string, RouteRecordRaw[]>();
  for (const item of frontendHiddenChildRoutes) {
    if (!groupedByParent.has(item.parentPath)) {
      groupedByParent.set(item.parentPath, []);
    }
    groupedByParent.get(item.parentPath)!.push(item.childRoute);
  }
  
  // 处理每个父路径
  groupedByParent.forEach((childRoutes, parentPath) => {
    const existingParent = findParentRoute(mergedRoutes, parentPath);
    
    for (const childRoute of childRoutes) {
      // 检查子路由是否已存在
      const childName = childRoute.name as string;
      const childPath = childRoute.path as string;
      if (isRouteExists(mergedRoutes, childName, childPath)) {
        continue; // 跳过已存在的路由
      }
      
      if (existingParent) {
        // 父路由存在，将子路由添加到父路由的children中
        if (!existingParent.children) {
          existingParent.children = [];
        }
        // 检查children中是否已存在
        const existsInChildren = existingParent.children.some(
          (r: any) => r.name === childName || r.path === childPath
        );
        if (!existsInChildren) {
          existingParent.children.push(childRoute);
        }
      } else {
        // 父路由不存在，尝试查找父路径的父路由（如 /toogo/robot 找不到，尝试找 /toogo）
        const parentParts = parentPath.split('/').filter(p => p);
        if (parentParts.length > 1) {
          // 尝试查找上一级父路由
          const grandParentPath = '/' + parentParts.slice(0, -1).join('/');
          const grandParent = findParentRoute(mergedRoutes, grandParentPath);
          if (grandParent) {
            // 创建中间父路由
            const middleParentPath = '/' + parentParts.join('/');
            let middleParent = grandParent.children?.find(
              (r: any) => r.path === middleParentPath || r.path === parentParts[parentParts.length - 1]
            );
            if (!middleParent) {
              // 创建中间父路由
              middleParent = {
                path: middleParentPath,
                name: middleParentPath.replace(/\//g, '_').replace(/^_/, ''),
                component: ParentLayout,
                meta: { title: '隐藏路由', hidden: true },
                children: [],
              };
              if (!grandParent.children) {
                grandParent.children = [];
              }
              grandParent.children.push(middleParent);
            }
            if (!middleParent.children) {
              middleParent.children = [];
            }
            const existsInMiddleChildren = middleParent.children.some(
              (r: any) => r.name === childName || r.path === childPath
            );
            if (!existsInMiddleChildren) {
              middleParent.children.push(childRoute);
            }
            continue;
          }
        }
        // 如果还是找不到，创建新的父路由
        const newParent: RouteRecordRaw = {
          path: parentPath,
          name: `${parentPath.replace(/\//g, '_').replace(/^_/, '')}_hidden`,
          component: Layout,
          meta: { title: '隐藏路由', hidden: true },
          children: [childRoute],
        };
        mergedRoutes.push(newParent);
      }
    }
  });
  
  return mergedRoutes;
}
interface TreeHelperConfig {
  id: string;
  children: string;
  pid: string;
}

const DEFAULT_CONFIG: TreeHelperConfig = {
  id: 'id',
  children: 'children',
  pid: 'pid',
};

const getConfig = (config: Partial<TreeHelperConfig>) => Object.assign({}, DEFAULT_CONFIG, config);

export interface IAsyncRouteState {
  menus: RouteRecordRaw[];
  routers: RouteRecordRaw[];
  addRouters: RouteRecordRaw[];
  keepAliveComponents: string[];
  isDynamicAddedRoute: boolean;
}

function filter<T = any>(
  tree: T[],
  func: (n: T) => boolean,
  config: Partial<TreeHelperConfig> = {},
): T[] {
  config = getConfig(config);
  const children = config.children as string;

  function listFilter(list: T[]) {
    return list
      .map((node: any) => ({ ...node }))
      .filter((node) => {
        node[children] = node[children] && listFilter(node[children]);
        return func(node) || (node[children] && node[children].length);
      });
  }

  return listFilter(tree);
}

export const useAsyncRouteStore = defineStore('app-async-route', {
  state: (): IAsyncRouteState => ({
    menus: [],
    routers: constantRouter,
    addRouters: <RouteRecordRaw[]>[],
    keepAliveComponents: ['ParentView'],
    isDynamicAddedRoute: false,
  }),
  getters: {
    getRouters() {
      const routers: RouteRecordRaw[] = this.addRouters;
      return toRaw(routers);
    },
    getMenus(): RouteRecordRaw[] {
      return this.menus;
    },
    getIsDynamicAddedRoute(): boolean {
      return this.isDynamicAddedRoute;
    },
  },
  actions: {
    //从缓存列表删除路由
    removeKeepAliveComponents(compNames) {
      if (!compNames || !compNames.length) return;
      this.keepAliveComponents = this.keepAliveComponents.filter(
        (item) => !compNames.includes(item),
      );
    },
    setDynamicAddedRoute(added: boolean) {
      this.isDynamicAddedRoute = added;
    },
    // 设置动态路由
    setRouters(routers: RouteRecordRaw[]) {
      this.addRouters = routers;
      this.routers = constantRouter.concat(routers);
    },
    setMenus(menus) {
      // 设置动态路由
      this.menus = menus;
    },
    setKeepAliveComponents(compNames) {
      // 设置需要缓存的组件
      this.keepAliveComponents = compNames;
    },
    async generateRoutes(data) {
      // 兼容：后端动态菜单拉取失败时，不应导致 accessedRouters 为 undefined 而崩溃
      let accessedRouters: RouteRecordRaw[] = [];
      const permissionsList = data.permissions || [];
      const routeFilter = (route) => {
        const { meta } = route;
        const { permissions, authEvery, ignoreAuth } = meta || {};
        // 如果设置了 ignoreAuth，直接允许访问
        if (ignoreAuth === true) return true;
        if (!permissions) return true;
        // 新增 authEvery 判断逻辑
        if (authEvery === true) {
          return permissions.every((permission) =>
            permissionsList.some((item) => item.value === permission),
          );
        }
        return permissionsList.some((item) => permissions.includes(item.value));
      };
      const { getPermissionMode } = useProjectSetting();
      const permissionMode = unref(getPermissionMode);
      if (permissionMode === 'BACK') {
        // 动态获取菜单
        try {
          accessedRouters = (await generatorDynamicRouter()) || [];
          // 合并前端定义的隐藏路由（后端可能没有配置这些）
          accessedRouters = mergeFrontendHiddenRoutes(accessedRouters);
        } catch (error) {
          // 后端菜单接口异常时，回退到前端静态路由，避免登录后白屏/跳回登录页
          console.log('[generateRoutes] generatorDynamicRouter failed, fallback to frontend routes:', error);
          accessedRouters = filter(asyncRoutes, routeFilter);
        }
      } else {
        try {
          //过滤账户是否拥有某一个权限，并将菜单从加载列表移除
          accessedRouters = filter(asyncRoutes, routeFilter);
        } catch (error) {
          console.log(error);
          accessedRouters = [];
        }
      }
      accessedRouters = (accessedRouters || []).filter(routeFilter);
      // const clapRouters = clapMultipleRoute(cloneDeep(accessedRouters));
      this.setRouters(accessedRouters);
      this.setMenus(accessedRouters);
      return toRaw(accessedRouters);
    },
  },
});

// Need to be used outside the setup
export function useAsyncRouteStoreWidthOut() {
  return useAsyncRouteStore(store);
}
