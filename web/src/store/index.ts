import type { App } from 'vue';
import { createPinia } from 'pinia';

const store = createPinia();

export function setupStore(app: App<Element>) {
  app.use(store);
}

// 重置所有 store 状态（退出登录时调用）
export function resetStore() {
  // 简化实现：直接清除相关的 localStorage
  const keysToRemove = [
    'APP-PROJECT-SETTING',
    'APP-DESIGN-SETTING',
    'APP-TABS-VIEW',
    'APP-LOCALES-SETTING',
  ];
  keysToRemove.forEach((key) => {
    try {
      localStorage.removeItem(key);
    } catch (e) {
      // ignore
    }
  });
}

export { store };
