/**
 * 定时倒计时调试工具
 * 在浏览器控制台运行此脚本，查看机器人数据中是否包含定时字段
 */

// 1. 查看当前页面的机器人列表数据
console.log('========== 定时倒计时调试 ==========');
console.log('');

// 获取 Vue 实例（需要根据实际情况调整）
const app = document.querySelector('#app').__vueParentComponent;

// 方法1：检查响应数据
console.log('📊 方法1：打开 Network 面板');
console.log('1. 刷新页面');
console.log('2. 找到 /toogo/robot/list 请求');
console.log('3. 查看 Response，检查是否包含 scheduleStart 和 scheduleStop');
console.log('');

// 方法2：在控制台直接查看
console.log('📊 方法2：查看 localStorage 或页面数据');
console.log('执行以下代码查看机器人数据：');
console.log(`
// 假设你的机器人数据在 robotList 变量中
if (typeof robotList !== 'undefined') {
  console.log('机器人列表:', robotList);
  robotList.forEach((robot, index) => {
    console.log(\`机器人[\${index}]:\`, {
      id: robot.id,
      name: robot.robotName,
      status: robot.status,
      scheduleStart: robot.scheduleStart,  // 检查这个
      scheduleStop: robot.scheduleStop     // 检查这个
    });
  });
} else {
  console.log('未找到 robotList 变量');
}
`);
console.log('');

// 方法3：检查组件数据
console.log('📊 方法3：使用 Vue DevTools');
console.log('1. 安装 Vue DevTools 浏览器扩展');
console.log('2. 打开 DevTools → Vue 标签');
console.log('3. 找到 RobotList 或类似组件');
console.log('4. 查看 robotList 数据中是否包含 scheduleStart/scheduleStop');
console.log('');

// 方法4：直接检查 API 响应
console.log('📊 方法4：手动调用 API');
console.log('在控制台执行：');
console.log(`
fetch('/toogo/robot/list?page=1&perPage=10', {
  headers: {
    'Content-Type': 'application/json',
    // 添加你的认证 token
  }
})
.then(res => res.json())
.then(data => {
  console.log('API 响应:', data);
  if (data.data && data.data.list) {
    data.data.list.forEach((robot, index) => {
      console.log(\`机器人[\${index}]:\`, {
        id: robot.id,
        name: robot.robotName || robot.robot_name,
        status: robot.status,
        scheduleStart: robot.scheduleStart || robot.schedule_start,
        scheduleStop: robot.scheduleStop || robot.schedule_stop
      });
    });
  }
})
.catch(err => console.error('API 错误:', err));
`);
console.log('');

console.log('========== 常见问题检查 ==========');
console.log('');
console.log('❓ 问题1：字段不存在');
console.log('   → 检查后端是否返回了 scheduleStart/scheduleStop');
console.log('   → 检查字段名是否正确（驼峰 vs 下划线）');
console.log('');
console.log('❓ 问题2：字段值为 null');
console.log('   → 机器人可能没有设置定时');
console.log('   → 创建机器人时需要设置 scheduleStart 或 scheduleStop');
console.log('');
console.log('❓ 问题3：倒计时组件不显示');
console.log('   → 检查显示条件：v-if="robot.status === 2"');
console.log('   → 检查 scheduleStop 是否有值');
console.log('   → 检查时间是否未来时间（已过期的不显示）');
console.log('');
console.log('========== 快速测试 ==========');
console.log('');
console.log('🧪 创建测试机器人：');
console.log('1. 进入创建机器人页面');
console.log('2. 设置定时停止时间为 1 小时后');
console.log('3. 创建并启动机器人');
console.log('4. 返回列表页查看是否显示倒计时');
console.log('');
console.log('=====================================');


