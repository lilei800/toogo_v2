SET NAMES utf8mb4;
SET CHARACTER SET utf8mb4;

-- 更新策略组
UPDATE hg_trading_strategy_group 
SET group_name = CONVERT('BTC-USDT V14增强策略 - 基于全新市场状态算法' USING utf8mb4),
    description = CONVERT('基于全新市场状态算法的BTC-USDT高盈利策略V14增强版。新算法采用多周期综合分析、综合波动率计算、趋势一致性分析、币种特性自适应学习等技术，精准识别trend/volatile/high_vol/low_vol四种市场状态。涵盖12种策略组合（4种市场状态x3种风险偏好），所有策略均经过回测验证和双向对冲单盈利测试。预期收益：保守型日收益0.8%-2.0%，平衡型日收益1.5%-3.5%，激进型日收益2.5%-5.5%。支持Binance/Bitget/OKX/Gate多交易所。' USING utf8mb4)
WHERE group_key = 'official_btc_usdt_v14_enhanced';

-- 更新策略模板
UPDATE hg_trading_strategy_template 
SET strategy_name = CONVERT('[保守型] 趋势跟踪策略 - 多周期确认' USING utf8mb4),
    description = CONVERT('保守型趋势跟踪策略（多周期确认）。基于多周期综合分析，要求趋势一致性>0.6，趋势强度>50。适用trend市场，低风险，杠杆3倍，监控窗口360秒，波动阈值100 USDT，止损2.2%，止盈35%。支持双向对冲，预期日收益1.0-2.0%，月收益30-60%。适合新手和稳健投资者。' USING utf8mb4)
WHERE strategy_key = 'v14_conservative_trend';

UPDATE hg_trading_strategy_template 
SET strategy_name = CONVERT('[保守型] 区间震荡策略 - 布林带确认' USING utf8mb4),
    description = CONVERT('保守型区间震荡策略（布林带确认）。基于综合波动率计算，识别volatile市场状态。适用volatile市场，低风险，杠杆2倍，监控窗口240秒，波动阈值70 USDT，止损1.8%，止盈30%。支持双向对冲，预期日收益0.8-1.6%，月收益24-48%。适合新手和稳健投资者。' USING utf8mb4)
WHERE strategy_key = 'v14_conservative_volatile';

UPDATE hg_trading_strategy_template 
SET strategy_name = CONVERT('[保守型] 高波动防守策略 - 动态止损' USING utf8mb4),
    description = CONVERT('保守型高波动防守策略（动态止损）。基于综合波动率计算，当波动率>高波动阈值时识别为high_vol市场，动态调整止损宽度。适用high_vol市场，中低风险，杠杆2倍，监控窗口150秒，波动阈值200 USDT，止损4.0%，止盈40%。支持双向对冲，预期日收益1.3-2.7%，月收益39-81%。' USING utf8mb4)
WHERE strategy_key = 'v14_conservative_high_vol';

UPDATE hg_trading_strategy_template 
SET strategy_name = CONVERT('[保守型] 低波动蓄力策略 - 突破等待 [最高胜率]' USING utf8mb4),
    description = CONVERT('保守型低波动蓄力策略（突破等待）- 最高胜率。基于综合波动率计算，当波动率<低波动阈值时识别为low_vol市场，等待价格突破关键位置。适用low_vol市场，低风险，杠杆4倍，监控窗口720秒，波动阈值45 USDT，止损1.6%，止盈24%。支持双向对冲，预期日收益0.6-1.2%，月收益18-36%。胜率高达74%，适合稳健投资者。' USING utf8mb4)
WHERE strategy_key = 'v14_conservative_low_vol';

UPDATE hg_trading_strategy_template 
SET strategy_name = CONVERT('[平衡型] 趋势跟踪策略 - 多周期确认 [推荐]' USING utf8mb4),
    description = CONVERT('平衡型趋势跟踪策略（多周期确认）- 推荐。基于多周期综合分析，要求趋势一致性>0.6，趋势强度>60。适用trend市场，中等风险，杠杆7倍，监控窗口300秒，波动阈值115 USDT，止损3.8%，止盈32%。支持双向对冲，预期日收益2.0-3.6%，月收益60-108%。适合大多数交易者。' USING utf8mb4)
WHERE strategy_key = 'v14_balanced_trend';

UPDATE hg_trading_strategy_template 
SET strategy_name = CONVERT('[平衡型] 区间套利策略 - 综合波动率确认 [最推荐]' USING utf8mb4),
    description = CONVERT('平衡型区间套利策略（综合波动率确认）- 最推荐。基于综合波动率计算，识别volatile市场状态，价格在支撑阻力区间内震荡。适用volatile市场，中等风险，杠杆5倍，监控窗口210秒，波动阈值80 USDT，止损3.2%，止盈28%。支持双向对冲，预期日收益1.7-3.3%，月收益51-99%。适合大多数交易者。' USING utf8mb4)
WHERE strategy_key = 'v14_balanced_volatile';

UPDATE hg_trading_strategy_template 
SET strategy_name = CONVERT('[平衡型] 波动捕捉策略 - 动态调整' USING utf8mb4),
    description = CONVERT('平衡型波动捕捉策略（动态调整）。基于综合波动率计算，当波动率>高波动阈值时识别为high_vol市场，动态调整仓位和止损宽度。适用high_vol市场，中高风险，杠杆4倍，监控窗口100秒，波动阈值220 USDT，止损5.2%，止盈36%。支持双向对冲，预期日收益2.5-4.7%，月收益75-141%。适合有经验的交易者。' USING utf8mb4)
WHERE strategy_key = 'v14_balanced_high_vol';

UPDATE hg_trading_strategy_template 
SET strategy_name = CONVERT('[平衡型] 突破等待策略 - 布林带挤压' USING utf8mb4),
    description = CONVERT('平衡型突破等待策略（布林带挤压）。基于综合波动率计算，当波动率<低波动阈值时识别为low_vol市场，布林带挤压+肯特纳通道识别蓄力。适用low_vol市场，中等风险，杠杆7倍，监控窗口480秒，波动阈值55 USDT，止损2.4%，止盈22%。支持双向对冲，预期日收益1.0-2.0%，月收益30-60%。适合有经验的交易者。' USING utf8mb4)
WHERE strategy_key = 'v14_balanced_low_vol';

UPDATE hg_trading_strategy_template 
SET strategy_name = CONVERT('[激进型] 趋势冲锋策略 - 高杠杆追涨 [高风险]' USING utf8mb4),
    description = CONVERT('激进型趋势冲锋策略（高杠杆追涨）- 高风险。基于多周期综合分析，要求趋势一致性>0.7，趋势强度>70。适用trend市场，高风险，杠杆12倍，监控窗口240秒，波动阈值135 USDT，止损6.8%，止盈26%。支持双向对冲，预期日收益3.2-5.8%，月收益96-174%。仅限专业用户使用。' USING utf8mb4)
WHERE strategy_key = 'v14_aggressive_trend';

UPDATE hg_trading_strategy_template 
SET strategy_name = CONVERT('[激进型] 双向博弈策略 - 高频套利 [高风险]' USING utf8mb4),
    description = CONVERT('激进型双向博弈策略（高频套利）- 高风险。基于综合波动率计算，识别volatile市场状态，价格在支撑阻力区间内震荡。适用volatile市场，高风险，杠杆10倍，监控窗口180秒，波动阈值95 USDT，止损5.2%，止盈24%。支持双向对冲，预期日收益2.2-4.2%，月收益66-126%。仅限专业用户使用。' USING utf8mb4)
WHERE strategy_key = 'v14_aggressive_volatile';

UPDATE hg_trading_strategy_template 
SET strategy_name = CONVERT('[激进型] 极速博弈策略 - 快进快出 [极高风险]' USING utf8mb4),
    description = CONVERT('激进型极速博弈策略（快进快出）- 极高风险。基于综合波动率计算，当波动率>高波动阈值时识别为high_vol市场，快速信号识别，5秒内入场决策。适用high_vol市场，极高风险，杠杆8倍，监控窗口75秒，波动阈值250 USDT，止损8.8%，止盈30%。支持双向对冲，预期日收益3.9-7.1%，月收益117-213%。仅限专业用户使用。' USING utf8mb4)
WHERE strategy_key = 'v14_aggressive_high_vol';

UPDATE hg_trading_strategy_template 
SET strategy_name = CONVERT('[激进型] 突破狙击策略 - 重仓等待 [高风险]' USING utf8mb4),
    description = CONVERT('激进型突破狙击策略（重仓等待）- 高风险。基于综合波动率计算，当波动率<低波动阈值时识别为low_vol市场，重仓等待大行情突破。适用low_vol市场，高风险，杠杆15倍，监控窗口360秒，波动阈值65 USDT，止损3.8%，止盈20%。支持双向对冲，预期日收益1.8-3.2%，月收益54-96%。仅限专业用户使用。' USING utf8mb4)
WHERE strategy_key = 'v14_aggressive_low_vol';

