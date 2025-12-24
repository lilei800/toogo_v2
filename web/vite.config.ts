import type { UserConfig, ConfigEnv } from 'vite';
import { loadEnv } from 'vite';
import { resolve } from 'path';
import { wrapperEnv } from './build/utils';
import { createVitePlugins } from './build/vite/plugin';
import { OUTPUT_DIR } from './build/constant';
import { createProxy } from './build/vite/proxy';
import pkg from './package.json';
import { format } from 'date-fns';
const { dependencies, devDependencies, name, version } = pkg;

const __APP_INFO__ = {
  pkg: { dependencies, devDependencies, name, version },
  lastBuildTime: format(new Date(), 'yyyy-MM-dd HH:mm:ss'),
};

function pathResolve(dir: string) {
  return resolve(process.cwd(), '.', dir);
}

export default ({ command, mode }: ConfigEnv): UserConfig => {
  const root = process.cwd();
  const env = loadEnv(mode, root);
  const viteEnv = wrapperEnv(env);
  const { VITE_PUBLIC_PATH, VITE_PORT, VITE_PROXY } = viteEnv;
  const isBuild = command === 'build';
  const defaultBackend = 'http://127.0.0.1:8000';
  return {
    base: VITE_PUBLIC_PATH,
    esbuild: {},
    resolve: {
      alias: [
        {
          find: 'vue-i18n',
          replacement: 'vue-i18n/dist/vue-i18n.cjs.js',
        },
        {
          find: /\/#\//,
          replacement: pathResolve('types') + '/',
        },
        {
          find: '@',
          replacement: pathResolve('src') + '/',
        },
      ],
    },
    plugins: createVitePlugins(viteEnv, isBuild),
    define: {
      __APP_INFO__: JSON.stringify(__APP_INFO__),
    },
    css: {
      preprocessorOptions: {
        less: {
          modifyVars: {},
          javascriptEnabled: true,
          additionalData: `@import "src/styles/var.less";`,
        },
      },
    },
    server: {
      host: true,
      port: VITE_PORT,
      // 说明：
      // - 环境变量 VITE_PROXY 在某些机器上可能缺失/错误，导致 proxy ECONNREFUSED，页面空白。
      // - hotgo_v2 后端默认监听 :8000（manifest/config/config.yaml），这里给本机开发加一个兜底：
      //   强制把 /admin 与 /socket 代理到 http://127.0.0.1:8000。
      // - 注意：后端路由配置中 admin 的 prefix 是 /admin，所以代理时不能去掉 /admin 前缀
      proxy: (() => {
        const proxy = createProxy(VITE_PROXY);
        proxy['/admin'] = {
          target: defaultBackend,
          changeOrigin: true,
          ws: true,
          // 不要去掉 /admin 前缀，因为后端路由需要这个前缀
          // rewrite: (path) => path.replace(/^\/admin/, ''),
        };
        proxy['/socket'] = {
          target: defaultBackend,
          changeOrigin: true,
          ws: true,
          rewrite: (path) => path.replace(/^\/socket/, '/socket'),
        };
        return proxy;
      })(),
    },
    optimizeDeps: {
      include: ['@vicons/ionicons5', '@vicons/antd'],
      exclude: [],
    },
    build: {
      target: 'es2015',
      cssTarget: 'chrome80',
      outDir: OUTPUT_DIR,
      // minify: 'terser',
      /**
       * 当 minify 为 minify 或 terser 打开注释
       */
      // terserOptions: {
      //   compress: {
      //     keep_infinity: true,
      //     drop_console: VITE_DROP_CONSOLE,
      //   },
      // },
      reportCompressedSize: false,
      chunkSizeWarningLimit: 2000,
      // 构建分包策略
      rollupOptions: {
        output: {
          manualChunks: {
            'naive-ui': ['naive-ui'],
            'lodash-es': ['lodash-es'],
            'vue-router': ['vue-router'],
            'vue-quill': ['@vueup/vue-quill'],
            'vicons-antd': ['@vicons/antd'],
            'vicons-ionicons5': ['@vicons/ionicons5'],
            vuedraggable: ['vuedraggable'],
            echarts: ['echarts'],
            vueuse: ['@vueuse/core'],
            vue: ['vue'],
            pinia: ['pinia'],
            xlsx: ['xlsx'],
          },
        },
      },
    },
  };
};
