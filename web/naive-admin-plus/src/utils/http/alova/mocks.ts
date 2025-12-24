// 这里按需导入 mock 文件，只有在这里导入并导出，才会执行 mock 拦截
// 可根据实际开发情况配置，如手动导入 mock 文件

const modules = import.meta.glob(['/mock/**/*.ts', '!/mock/**/util.ts'], { eager: true });

export default Object.keys(modules).map((key) => {
  return modules[key].default || {};
});
