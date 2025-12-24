/**
 * 创建BTCUSDT和ETHUSDT的波动率配置
 * 使用方法：在浏览器控制台运行，或者通过API调用
 */

// BTCUSDT配置（比特币，市值最大，流动性最好，波动相对稳定）
const BTCUSDTConfig = {
  symbol: 'BTCUSDT',
  highVolatilityThreshold: 2.0,    // 高波动阈值：2.0%（比特币波动相对稳定）
  lowVolatilityThreshold: 0.4,    // 低波动阈值：0.4%（比特币波动较小）
  trendStrengthThreshold: 0.35,    // 趋势阈值：0.35
  weight1m: 0.10,                  // 1分钟权重：10%
  weight5m: 0.20,                  // 5分钟权重：20%
  weight15m: 0.25,                 // 15分钟权重：25%
  weight30m: 0.25,                 // 30分钟权重：25%
  weight1h: 0.20,                  // 1小时权重：20%
  isActive: 1,                     // 启用状态
};

// ETHUSDT配置（以太坊，市值第二，流动性好，波动可能比BTC稍大）
const ETHUSDTConfig = {
  symbol: 'ETHUSDT',
  highVolatilityThreshold: 2.5,    // 高波动阈值：2.5%（以太坊波动稍大）
  lowVolatilityThreshold: 0.5,     // 低波动阈值：0.5%（以太坊波动稍大）
  trendStrengthThreshold: 0.35,    // 趋势阈值：0.35
  weight1m: 0.10,                  // 1分钟权重：10%
  weight5m: 0.20,                  // 5分钟权重：20%
  weight15m: 0.25,                 // 15分钟权重：25%
  weight30m: 0.25,                 // 30分钟权重：25%
  weight1h: 0.20,                  // 1小时权重：20%
  isActive: 1,                     // 启用状态
};

// 配置说明
console.log('BTCUSDT配置说明：');
console.log('- 高波动阈值：2.0%（比特币波动相对稳定）');
console.log('- 低波动阈值：0.4%（比特币波动较小）');
console.log('- 适合：市值最大，流动性最好，波动相对稳定的主流币种');
console.log('\nETHUSDT配置说明：');
console.log('- 高波动阈值：2.5%（以太坊波动稍大）');
console.log('- 低波动阈值：0.5%（以太坊波动稍大）');
console.log('- 适合：市值第二，流动性好，波动可能比BTC稍大的主流币种');

// 导出配置（如果是在Node.js环境中）
if (typeof module !== 'undefined' && module.exports) {
  module.exports = {
    BTCUSDTConfig,
    ETHUSDTConfig,
  };
}

