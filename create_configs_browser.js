/**
 * 在浏览器控制台运行此脚本来创建BTCUSDT和ETHUSDT的波动率配置
 * 
 * 使用方法：
 * 1. 打开波动率配置页面
 * 2. 打开浏览器开发者工具（F12）
 * 3. 切换到Console标签
 * 4. 复制粘贴下面的代码并执行
 */

// 确保API可用（需要从页面中获取）
if (typeof ToogoVolatilityConfigApi === 'undefined') {
  console.error('请确保在波动率配置页面运行此脚本');
  console.log('或者手动导入API：');
  console.log('import { ToogoVolatilityConfigApi } from "@/api/toogo";');
} else {
  // BTCUSDT配置
  const btcConfig = {
    symbol: 'BTCUSDT',
    highVolatilityThreshold: 2.0,    // 高波动阈值：2.0%
    lowVolatilityThreshold: 0.4,     // 低波动阈值：0.4%
    trendStrengthThreshold: 0.35,    // 趋势阈值：0.35
    weight1m: 0.10,                  // 1分钟权重：10%
    weight5m: 0.20,                  // 5分钟权重：20%
    weight15m: 0.25,                 // 15分钟权重：25%
    weight30m: 0.25,                 // 30分钟权重：25%
    weight1h: 0.20,                  // 1小时权重：20%
    isActive: 1,                     // 启用状态
  };

  // ETHUSDT配置
  const ethConfig = {
    symbol: 'ETHUSDT',
    highVolatilityThreshold: 2.5,    // 高波动阈值：2.5%
    lowVolatilityThreshold: 0.5,     // 低波动阈值：0.5%
    trendStrengthThreshold: 0.35,    // 趋势阈值：0.35
    weight1m: 0.10,                  // 1分钟权重：10%
    weight5m: 0.20,                  // 5分钟权重：20%
    weight15m: 0.25,                 // 15分钟权重：25%
    weight30m: 0.25,                 // 30分钟权重：25%
    weight1h: 0.20,                  // 1小时权重：20%
    isActive: 1,                     // 启用状态
  };

  // 创建配置的函数
  async function createConfigs() {
    try {
      console.log('开始创建BTCUSDT配置...');
      const btcRes = await ToogoVolatilityConfigApi.create(btcConfig);
      if (btcRes.code === 0) {
        console.log('✅ BTCUSDT配置创建成功！');
      } else {
        console.error('❌ BTCUSDT配置创建失败：', btcRes.msg);
      }

      console.log('开始创建ETHUSDT配置...');
      const ethRes = await ToogoVolatilityConfigApi.create(ethConfig);
      if (ethRes.code === 0) {
        console.log('✅ ETHUSDT配置创建成功！');
      } else {
        console.error('❌ ETHUSDT配置创建失败：', ethRes.msg);
      }

      console.log('\n配置创建完成！请刷新页面查看。');
    } catch (error) {
      console.error('创建配置时出错：', error);
    }
  }

  // 执行创建
  createConfigs();
}

