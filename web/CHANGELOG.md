# CHANGELOG

## Pending

## 3.9.4

- 💎 优化 `图标选择` 组件
- 💎 设置 `打包压缩配置` 为 `gzip`
- `依赖升级`

## 3.9.3

- 🌟 新增 `TableAction` 支持 `按钮参数` 配置
- 🌟 新增 `BasicForm` 支持 `NColorPicker` 组件
- 🌟 新增 `ProSearchGroup` 支持 `Enter` 提交事件
- 🌟 新增 `ProSearchGroup` 支持 `搜索按钮纯图标` 配置
- 🌟 新增 `ImageUpload` 支持 `hideUploadTitle` 参数
- 💎 优化 `主题皮肤` 模式，禁用导航风格切换

## 3.9.2

- 🌟 新增 `ProSearchGroup` 组件支持隐藏选择器
- 🌟 新增 `ImageUpload` 组件拖拽结束回调 `change` 事件
- 💎 优化 `CheckCard` 组件不传图片场景
- 💎 优化 `Paragraph` 组件样式
- 🐛 修复 `BasicUpload` 组件回显逻辑
- 🐛 修复 `切换页面` 闪现横向滚动条
- 🐛 修复 `loadingBar` 失效
- `依赖升级`

## 3.9.1

- 🌟 新增 `BasicTable` 分页样式小屏幕采用精简样式
- 💎 优化 `移动端` 折叠菜单体验问题
- 💎 优化 `核心布局` CSS 样式
- `依赖升级`

## 3.9.0

- 🌟 新增 `PicFile` 文件管理组件
- 🌟 新增 `meta.authEvery` 验证所有权限
- 🌟 新增 `小屏幕` 侧边栏菜单支持
- 💎 优化 `小屏幕` 菜单以及样式
- 🐛 修复 `主题皮肤` 内容区域层级显示问题
- 🐛 修复 `mock打印请求` 配置不生效
- `依赖升级`

## 3.8.0

- 🌟 新增 `全屏水印` 功能显示配置
- 🌟 新增 `主题皮肤` 支持9款不同主题皮肤
- 🌟 新增 `顶部透明` 开关设置，适配主题皮肤
- 🌟 新增 `侧边栏透明` 开关设置，适配主题皮肤
- 🌟 新增 `容器透明` 开关设置，适配主题皮肤
- 🌟 新增 `多页签` 开关设置,显示图标
- 🌟 新增 `面包屑` 开关设置,显示图标
- 🌟 新增 `BasicCard` 折叠卡片组件
- 🌟 新增 `BasicForm` 新增组件原始插槽的支持
- 💎 优化 `header` 圆角样式，适配主题皮肤
- 💎 优化 `分栏菜单` 分割样式
- `依赖升级`

## 3.7.2

- 🌟 新增 `usePagination` 方法，可用于自定义分页逻辑
- 🌟 新增 `IconPicker` 组件支持分页
- `依赖升级`

## 3.7.1

- 🌟 新增 `TableAction` 组件 新增 `gap` 参数
- 🌟 新增 `useLocalSetting` 方法，获取局部环境变量
- 💎 优化 `BasicTable` 组件 设置图标间距
- `依赖升级`

## 3.7.0

- 🌟 新增 `BasicTable` 组件 支持打印功能
- 🌟 新增 `BasicTable` 组件 支持表格边框设置开关
- 💎 优化 `BasicTable` 组件 右侧图标样式
- 💎 优化 `BasicTable` 组件 内部样式
- 💎 优化 `alova` 错误处理
- 💎 优化 `类型` 定义
- `依赖升级`


## 3.6.0

- 🌟 新增 `BasicTag` 组件
- 🌟 新增 `alovaJs` 请求工具（推荐）
- 🌟 新增 `@alova/mock` 请求方案
- 🌟 新增 `@faker-js/faker` 模拟数据方案
- 🌟 新增 `多语言` 切换演示 `demo`
- 🌟 新增 `loggerMock` 环境变量，控制 `mock` 打印日志
- ⚠️ 移除 `vite-plugin-mock` 方案
- ⚠️ 调整 `imgUrl` 环境变量为 `fileUrl`
- ⚠️ 调整 `prodMock` 环境变量为 `useMock`
- ⚠️ 调整 `多语言` 菜单配置，只保留一个多语言菜单
- 🐛 修复 `表格编辑` 刷新属性丢失问题
- 💎 优化 `多页签样式` 
- 💎 优化 `基础列表` 筛选/编辑示例 
- 💎 优化 `搜索页面` 跳转以及显示问题
- `依赖升级`

## 3.5.0

- 🌟 新增 `BasicTable` 组件 可通过 `immediate` 参数，手动控制加载数据
- 🌟 新增 `BasicForm` 组件支持 `updateSchema` 方法
- 🌟 新增 `BasicForm` 组件支持 `updateGroupSchema` 方法
- 🌟 新增 `BasicForm` 组件 `componentProps` 参数支持函数，可满足特定需求
- 💎 优化 `login` 页面深色模式样式
- 💎 优化 `demo` 示例警告
- 💎 优化 `reset config` 数据不准确
- `依赖升级`


## 3.4.2

- 💎 升级 `axios` 到 `1.7.1` 采用 `AbortController` 取消请求的方案
- 💎 优化 `BasicColumn` 类型，支持泛型传入
- 🐞 修复 `BasicTable` 切换页数量，未重置勾选数据
- 💎 优化 调整 `vite/client` 的加载方式
- `依赖升级`


## 3.4.1

- 💎 优化 `深色-浅色` 左侧菜单恢复浅色风格
- `依赖升级`


## 3.4.0

- 🌟 新增 `深色主题` 切换组件，并在顶部右侧展示
- 🌟 新增 `深色主题` 切换过度样式
- 💎 优化 `骨架` 过度样式
- 💎 优化 `多页签` 过度样式
- 💎 优化 `code` 类型定义
- `依赖升级`


## 3.3.9

- 🌟 新增 `ProSearchGroup` 组件，支持 `inputValue` 和 `selectValue` 参数
- 💎 优化 `BasicSelect` 组件逻辑，支持响应式 `options`
- 💎 优化 `BasicColumn` 类型定义
- `依赖升级`


## 3.3.8

- 🌟 新增 `TogglePage` 登录页切换模版（示例）
- 💎 升级 `vite` 版本 `5.x` 注：`nodeJs` 版本需要 `18` 或者 `20` 以上版本
- `依赖升级`

## 3.3.7

- 💎 优化 `defineComponent` 语法 为 `setup`
- `依赖升级`

## 3.3.6

- 🌟 新增 `BasicTable` 支持 `downloadCsv` 方法，导出 `csv` 文件
- 💎 优化 `BasicForm` 分组表单 `setFieldsValue` 方法
- `依赖升级`

## 3.3.5

- 🌟 新增 `v-debounce` 全局指令
- 🌟 新增 `v-throttle` 全局指令
- 💎 优化 `横向菜单` 显示内容，支持收起溢出的菜单
- `依赖升级`

## 3.3.4

- 🐞 修复 `BasicForm` 的 `isEnterSubmit` 功能，导致表单无法回车
- 🌟 新增 `ProRadio` 支持再次点击取消选中
- 🌟 新增 `ProSearchGroup` 搜索表单组件
- 🌟 新增 `BasicForm` 支持 `group` 配置表单组，优先级高于 `schemas`
- 🌟 新增 `BasicForm` 支持 `change` 事件，当表单值发生变化执行
- `依赖升级`

## 3.3.3

- 🌟 新增 `BasicForm` 支持 `validate` 方法单独验证
- 💎 优化 `BasicForm` 提交按钮加载效果
- 💎 优化 `setLoadingSub` 改名为 `setLoading`
- 🐞 修复 `菜单折叠后` 层级问题
- 🐞 修复 `BasicTable` 使用插槽报错问题
- `依赖升级`

## 3.3.2

- 🌟 新增 `TableSelect` 结合 `Form` 组件使用示例
- 🌟 新增 `TableSelect` 导出 `getTableRef` 和 `updateTableSelectedKeys` 方法
- 🐞 修复 `I18n` 搜索异常
- `依赖升级`

## 3.3.1

- 🐞 修复 `多页签` 异常初始化
- 🌟 新增 `BasicTable` 导出 `setLoading`
- 🐞 修复 `面包屑` 跳转警告
- 🐞 修复 `面包屑` 多语言显示异常

## 3.3.0

- 🌟 新增 `ProCascader` 省市区联动组件
- 🌟 新增 `ProCode` 验证码组件
- 🌟 新增 `TableSelect` 支持 `formValues` 配置查询表单参数
- 🐞 修复 `分页配置` 不正确，导致 `mock` 返回数据异常 
- 💎 优化 `TypeScript` 类型
- `依赖升级`

## 3.2.0

- 🐞 修复 `getTreeValues` 方法
- 🐞 修复 `搜索页面` 跳转异常
- 🌟 新增 `ProCheckbox` 远程复选框组件
- 🌟 新增 `ProRadio` 远程单选框组件
- 🌟 新增 `BasicForm` 支持 `isEnterSubmit` 配置，监听回车事件提交
- 🌟 新增 `BasicTable` 支持配置 `search` 搜索，支持多种 `searchType` 可选
- `依赖升级`

## 3.1.0

- 🌟 新增 `scrollBar` 指令
- 🌟 新增 `resetStore` 退出登录，执行重置逻辑
- 🌟 新增 `resetRouter` 退出登录，执行重置逻辑
- 🌟 新增 `Cropper` 组件，新增 `uploadSuccess` 事件回调
- `依赖升级`


## 3.0.1

- 💎 优化 `propTypes` 定义
- 💎 优化 `TypeScript` 类型
- `依赖升级`

## 3.0.0-beta.4

- 🌟 新增 `borderRadius` 项目配置
- 💎 优化 `useECharts` 监听菜单收缩变化
- 💎 优化 `title` 动态获取显示
- `依赖升级`

## 3.0.0-beta.3 

- 🌟 适配 `多页签` 功能
- 🌟 新增 `setSchema` 动态设置示例
- 💎 优化 `setSchema` 方法参数类型
- `依赖升级`

## 3.0.0-beta.2 

- 🌟 新增 `BasicForm` 支持 `ImageUpload` 组件
- 🌟 新增 `BasicTable` 支持 `helpMessage` 配置
- 🌟 新增 `sendFormData` `get` `put` `delete` 请求方法
- 💎 优化 `Code And TypeScript`
- 💎 优化 `hasPermission` 权限判断方法，兼容两种格式
- 🐞 修复 `permissionMode` `BACK` 模式加载异常
- `依赖升级`


## 3.0.0-beta.1

### ✨ Features

🔥🔥🔥 Naive Admin Plus `3.0` 发布、工匠精神、持续打磨

#### ✨ 功能

- 🌟 `分栏菜单`
- 🌟 `常用菜单收藏夹（个性化DIY比多页签好用）`
- 🌟 `菜单栏按需隐藏`
- 🌟 `多语言支持`
- 🌟 `界面风格`
- 🌟 `配置持久化`
- 🌟 `多级菜单缓存`
- 🌟 `精细化配置重新设计`
- 🌟 `重写 ProLayout 组件，个性化定制 So easy`
- 🌟 `菜单显示位置个性化增强（全新模式DIY）`
- 🌟 `表单组件增强（可动态控制显示）`
- 🌟 `表格组件增强（可动态控制显示列）`

#### 🏷️ 组件

- 🌟 `排印扩展`
- 🌟 `可选卡片`
- 🌟 `可选按钮`
- 🌟 `Loading`
- 🌟 `图标选择器`
- 🌟 `弹窗表单`
- 🌟 `抽屉表单`
- 🌟 `表单表格`
- 🌟 `高级表格`
- 🌟 `图片上传`

#### 💻 页面

- 🌟 `全新登录页面设计（登录/注册）`
- 🌟 `全新主控台页面设计`
- 🌟 `认证页面（4个版本）`
- 🌟 `布局页面（常规/左右/上下）`
- 🌟 `常用页面-Curd`
- 🌟 `常用页面-Tab表格`
- 🌟 `常用页面-详情页面`
- 🌟 `常用页面-弹窗表单`
- 🌟 `常用页面-抽屉表单`
- 🌟 `常用页面-分组表单`
- 🌟 `常用页面-动态表单`
- 🌟 `持续更新中...`
