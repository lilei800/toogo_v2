SET NAMES utf8mb4;

-- 更新V8策略组描述（缩短到500字符以内）
UPDATE hg_trading_strategy_group SET description = 'Toogo AI量化团队基于半年K线数据和市场波动算法精心打造的BTC-USDT官方推荐策略V8.0版本。策略参数基于2024年6个月BTCUSDT历史K线数据回测优化，涵盖趋势、震荡、高波动、低波动四种市场状态。支持Binance/Bitget/OKX/Gate多交易所，包含12种智能策略，覆盖保守、平衡、激进三种风险偏好。保守型：低杠杆(2-4x)，小止损(2-5%)，适合新手。平衡型：中杠杆(5-8x)，平衡止损止盈，适合有经验的交易者。激进型：高杠杆(10-20x)，大止损(5-10%)，适合专业交易者。' WHERE group_key = 'official_btc_usdt_v8';

-- 更新V12策略组描述（缩短到500字符以内）
UPDATE hg_trading_strategy_group SET description = 'Toogo AI量化团队基于12个月K线数据回测精心打造的BTC-USDT盈利策略V12.0版本。策略参数基于2023-2024年12个月BTCUSDT历史K线数据回测优化，所有策略均经过严格回测验证。回测结果：保守型策略胜率52%-68%，平均日收益0.6%-1.5%。平衡型策略胜率55%-66%，平均日收益1.1%-2.8%。激进型策略胜率51%-63%，平均日收益1.8%-4.2%。支持Binance/Bitget/OKX/Gate多交易所，包含12种经过回测验证的盈利策略组合。' WHERE group_key = 'official_btc_usdt_v12';

