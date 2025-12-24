import type { App } from 'vue';
import { createPinia } from 'pinia';

const store = createPinia();

export function setupStore(app: App<Element>) {
  app.use(store);
}

export function resetStore() {
  const context = import.meta.glob('./modules/*.ts', { eager: true });
  Object.keys(context).forEach((key) => {
    // @ts-ignore
    const store = context[key].default;
    if (store && store.$reset) {
      store.$reset();
    }
  });
}

export { store };
