# 🚀 TooGo量化交易系统迁移到HotGo v2.0 完整指南

## 📊 迁移进度

### ✅ 已完成

1. **API接口文件（5个）** - ✅ 已转换为TypeScript
   - `api/trading/api-config.ts` - API配置接口
   - `api/trading/proxy-config.ts` - 代理配置接口
   - `api/trading/robot.ts` - 机器人接口
   - `api/trading/order.ts` - 订单接口
   - `api/trading/monitor.ts` - 监控接口

2. **Payment API文件（3个）** - ✅ 已创建
   - `api/payment/deposit.ts` - 充值接口
   - `api/payment/withdraw.ts` - 提现接口
   - `api/payment/balance.ts` - 余额接口

### 🚧 待创建文件清单

#### Trading页面组件（5个）
```
web/src/views/trading/
├── api-config/
│   └── index.vue          - API配置管理页面
├── proxy-config/
│   └── index.vue          - 代理配置页面
├── robot/
│   ├── index.vue          - 机器人列表页面
│   ├── create.vue         - 创建机器人向导
│   └── detail.vue         - 机器人详情页面
```

#### Payment页面组件（4个）
```
web/src/views/payment/
├── deposit/
│   └── index.vue          - 充值页面
├── withdraw/
│   └── index.vue          - 提现页面
├── balance/
│   └── index.vue          - 余额页面
└── admin/
    └── withdraw-audit.vue - 提现审核页面
```

#### 路由配置（2个）
```
web/src/router/routes/modules/
├── trading.ts             - Trading路由配置
└── payment.ts             - Payment路由配置
```

---

## 📝 页面组件创建模板

### Trading页面组件架构

所有Trading页面都遵循以下架构：

```vue
<template>
  <div>
    <n-card :bordered="false" class="proCard">
      <!-- 页面头部 -->
      <BasicForm
        @register="register"
        @submit="handleSubmit"
        @reset="handleReset"
      />
      
      <!-- 数据表格 -->
      <BasicTable
        :columns="columns"
        :request="loadDataTable"
        :row-key="(row) => row.id"
        ref="actionRef"
        :actionColumn="actionColumn"
        @update:checked-row-keys="onCheckedRow"
        :scroll-x="1200"
      >
        <!-- 自定义插槽 -->
      </BasicTable>
    </n-card>
  </div>
</template>

<script lang="ts" setup>
import { ref, reactive, h } from 'vue';
import { NButton, NTag, useDialog, useMessage } from 'naive-ui';
import { BasicTable, TableAction } from '@/components/Table';
import { BasicForm, FormSchema, useForm } from '@/components/Form/index';
import { columns } from './columns';
import { schemas } from './schemas';

// API imports
import { getDataList, deleteData } from '@/api/trading/xxx';

// ... 组件逻辑
</script>
```

### Payment页面组件架构

Payment页面使用卡片布局和表单：

```vue
<template>
  <div>
    <n-grid :cols="24" :x-gap="12">
      <!-- 左侧表单 -->
      <n-gi :span="8">
        <n-card title="创建订单">
          <n-form ref="formRef" :model="formValue" :rules="rules">
            <!-- 表单项 -->
          </n-form>
        </n-card>
      </n-gi>
      
      <!-- 右侧展示 -->
      <n-gi :span="16">
        <n-card title="订单列表">
          <n-data-table
            :columns="columns"
            :data="data"
            :pagination="pagination"
          />
        </n-card>
      </n-gi>
    </n-grid>
  </div>
</template>

<script lang="ts" setup>
// ... Payment组件逻辑
</script>
```

---

## 🎨 HotGo v2.0 特性说明

### 1. 使用的UI组件库

**Naive UI** - 主要组件：

```typescript
// 常用组件导入
import {
  NCard,           // 卡片容器
  NButton,         // 按钮
  NForm,           // 表单
  NFormItem,       // 表单项
  NInput,          // 输入框
  NInputNumber,    // 数字输入
  NSelect,         // 下拉选择
  NSwitch,         // 开关
  NTag,            // 标签
  NDataTable,      // 数据表格
  NGrid,           // 栅格布局
  NGi,             // 栅格项
  NSpace,          // 间距
  NStatistic,      // 统计数值
  NDescriptions,   // 描述列表
  NModal,          // 对话框
  useMessage,      // 消息提示
  useDialog,       // 对话框hooks
} from 'naive-ui';
```

### 2. 封装的基础组件

HotGo v2.0 提供了增强组件：

```typescript
// 表格组件
import { BasicTable, TableAction } from '@/components/Table';

// 表单组件
import { BasicForm, useForm } from '@/components/Form';

// 弹窗组件
import { BasicModal, useModal } from '@/components/Modal';
```

### 3. 组合式API模式

```typescript
// 标准的组合式API结构
<script lang="ts" setup>
import { ref, reactive, computed, onMounted } from 'vue';

// 响应式状态
const loading = ref(false);
const formValue = reactive({
  name: '',
  amount: 0
});

// 计算属性
const totalAmount = computed(() => {
  return formValue.amount * 1.1;
});

// 方法
const handleSubmit = async () => {
  loading.value = true;
  try {
    // API调用
  } finally {
    loading.value = false;
  }
};

// 生命周期
onMounted(() => {
  // 初始化逻辑
});
</script>
```

---

## 🛣️ 路由配置说明

### Trading路由配置示例

```typescript
// web/src/router/routes/modules/trading.ts
import { RouteRecordRaw } from 'vue-router';
import { Layout } from '@/router/constant';
import { WalletOutlined } from '@vicons/antd';
import { renderIcon } from '@/utils/index';

const routes: Array<RouteRecordRaw> = [
  {
    path: '/trading',
    name: 'Trading',
    redirect: '/trading/robot',
    component: Layout,
    meta: {
      title: '量化交易',
      icon: renderIcon(WalletOutlined),
      sort: 3,
    },
    children: [
      {
        path: 'api-config',
        name: 'trading_api_config',
        meta: {
          title: 'API配置',
        },
        component: () => import('@/views/trading/api-config/index.vue'),
      },
      {
        path: 'proxy-config',
        name: 'trading_proxy_config',
        meta: {
          title: '代理配置',
        },
        component: () => import('@/views/trading/proxy-config/index.vue'),
      },
      {
        path: 'robot',
        name: 'trading_robot',
        meta: {
          title: '机器人管理',
        },
        component: () => import('@/views/trading/robot/index.vue'),
      },
      {
        path: 'robot/create',
        name: 'trading_robot_create',
        meta: {
          title: '创建机器人',
          hidden: true,
        },
        component: () => import('@/views/trading/robot/create.vue'),
      },
      {
        path: 'robot/detail/:id',
        name: 'trading_robot_detail',
        meta: {
          title: '机器人详情',
          hidden: true,
        },
        component: () => import('@/views/trading/robot/detail.vue'),
      },
    ],
  },
];

export default routes;
```

### Payment路由配置示例

```typescript
// web/src/router/routes/modules/payment.ts
import { RouteRecordRaw } from 'vue-router';
import { Layout } from '@/router/constant';
import { CreditCardOutlined } from '@vicons/antd';
import { renderIcon } from '@/utils/index';

const routes: Array<RouteRecordRaw> = [
  {
    path: '/payment',
    name: 'Payment',
    redirect: '/payment/balance',
    component: Layout,
    meta: {
      title: 'USDT管理',
      icon: renderIcon(CreditCardOutlined),
      sort: 4,
    },
    children: [
      {
        path: 'balance',
        name: 'payment_balance',
        meta: {
          title: '我的余额',
        },
        component: () => import('@/views/payment/balance/index.vue'),
      },
      {
        path: 'deposit',
        name: 'payment_deposit',
        meta: {
          title: 'USDT充值',
        },
        component: () => import('@/views/payment/deposit/index.vue'),
      },
      {
        path: 'withdraw',
        name: 'payment_withdraw',
        meta: {
          title: 'USDT提现',
        },
        component: () => import('@/views/payment/withdraw/index.vue'),
      },
      {
        path: 'admin/withdraw-audit',
        name: 'payment_admin_withdraw_audit',
        meta: {
          title: '提现审核',
          permissions: ['admin.payment.withdraw.audit'],
        },
        component: () => import('@/views/payment/admin/withdraw-audit.vue'),
      },
    ],
  },
];

export default routes;
```

---

## 📋 详细实施步骤

### 第1步：创建所有页面文件

#### 方案A：使用代码生成器（推荐）

如果后端已经有完整的数据表和API，可以使用HotGo的代码生成器：

1. 访问 `http://localhost:3000/system/gen`
2. 选择数据表
3. 配置生成选项
4. 一键生成前后端代码

#### 方案B：手动创建（完全控制）

逐个创建页面文件，完全按照业务需求定制。

### 第2步：创建路由配置文件

```bash
# 创建路由文件
touch D:\go\src\hotgo_v2\web\src\router\routes\modules\trading.ts
touch D:\go\src\hotgo_v2\web\src\router\routes\modules\payment.ts
```

### 第3步：安装额外依赖

```bash
cd D:\go\src\hotgo_v2\web

# 安装二维码组件（用于payment充值）
pnpm install qrcode.vue
```

### 第4步：配置图标

```typescript
// 在需要的地方导入图标
import {
  WalletOutlined,
  RobotOutlined,
  DollarOutlined,
  ApiOutlined,
  SettingOutlined,
} from '@vicons/antd';
```

### 第5步：测试验证

```bash
# 启动前端开发服务器
cd D:\go\src\hotgo_v2\web
pnpm run dev

# 访问
# http://localhost:3000
```

---

## 🔧 关键配置点

### 1. API Base URL配置

编辑 `web/.env.development`:

```env
# API地址
VITE_GLOB_API_URL=http://127.0.0.1:8000
VITE_GLOB_API_URL_PREFIX=/admin
```

### 2. 代理配置（开发环境）

编辑 `web/vite.config.ts`:

```typescript
export default defineConfig({
  server: {
    proxy: {
      '/admin': {
        target: 'http://127.0.0.1:8000',
        changeOrigin: true,
      },
    },
  },
});
```

### 3. 权限配置

如果需要权限控制，在路由meta中配置：

```typescript
meta: {
  title: '提现审核',
  permissions: ['admin.payment.withdraw.audit'],
}
```

---

## 📚 参考资源

### HotGo v2.0 官方文档

- **GitHub**: https://github.com/bufanyun/hotgo
- **在线演示**: https://hotgo.facms.cn/admin
- **本地文档**: `D:\go\src\hotgo_v2\docs\guide-zh-CN\README.md`

### Naive UI 文档

- **官网**: https://www.naiveui.com
- **组件库**: https://www.naiveui.com/zh-CN/os-theme/components/button

### Vue 3 文档

- **官网**: https://cn.vuejs.org
- **Composition API**: https://cn.vuejs.org/guide/extras/composition-api-faq.html

---

## 🎯 快速开发建议

### 1. 复用现有组件

HotGo v2.0 提供了大量可复用组件，查看：
```
D:\go\src\hotgo_v2\web\src\views\system\
```

参考这些页面的实现方式。

### 2. 使用代码片段

创建VS Code代码片段加速开发：

```json
{
  "Vue3 Setup Component": {
    "prefix": "vue3-setup",
    "body": [
      "<template>",
      "  <div>",
      "    <n-card :bordered=\"false\" title=\"${1:Title}\">",
      "      ${2:Content}",
      "    </n-card>",
      "  </div>",
      "</template>",
      "",
      "<script lang=\"ts\" setup>",
      "import { ref, reactive, onMounted } from 'vue';",
      "import { NCard, useMessage } from 'naive-ui';",
      "",
      "const message = useMessage();",
      "",
      "onMounted(() => {",
      "  // 初始化",
      "});",
      "</script>"
    ]
  }
}
```

### 3. 统一错误处理

使用HotGo提供的错误处理：

```typescript
import { useMessage } from 'naive-ui';

const message = useMessage();

try {
  await apiCall();
  message.success('操作成功');
} catch (error) {
  message.error(error.message || '操作失败');
}
```

---

## ✅ 验收标准

### 功能完整性

- [ ] 所有API接口能正常调用
- [ ] 所有页面能正常访问
- [ ] 表单验证正常工作
- [ ] 数据CRUD操作正常
- [ ] 路由跳转正常

### 用户体验

- [ ] 页面加载无明显卡顿
- [ ] 操作反馈及时（loading、message）
- [ ] 错误提示友好
- [ ] 界面美观，符合Naive UI规范

### 代码质量

- [ ] TypeScript类型定义完整
- [ ] 无明显的console.error
- [ ] 代码格式符合ESLint规范
- [ ] 组件拆分合理

---

## 🚀 下一步计划

### 完成页面创建

我将为您创建所有9个页面组件（5个Trading + 4个Payment），每个页面都是完整可用的Vue3组件。

### 创建路由配置

配置Trading和Payment的路由，使其能在侧边栏显示并正常访问。

### 集成测试

启动服务，测试所有功能是否正常工作。

---

**准备好了吗？我可以立即开始创建所有页面组件！** 🎉

