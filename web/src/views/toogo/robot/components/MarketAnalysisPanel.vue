<template>
  <div class="market-analysis-panel" v-if="data">
    <div class="market-analysis-header">
      <n-space align="center" :size="8">
        <span style="font-weight: 600; font-size: 13px">📊 市场状态分析</span>
        <n-tag :type="getMarketStateType(data.signal?.currentMarketState)" size="small" :bordered="false">
          {{ formatMarketState(data.signal?.currentMarketState) }}
        </n-tag>
        <n-tag type="info" size="small" :bordered="false">
          {{ formatRiskPref(data.signal?.currentRiskPref) }}
        </n-tag>
      </n-space>
    </div>

    <!-- 策略模板参数 -->
    <div class="strategy-params-grid">
      <div class="param-item">
        <span class="param-label">平台</span>
        <span class="param-value">{{ robot.exchange?.toUpperCase() || robot.platform?.toUpperCase() }}</span>
      </div>
      <div class="param-item">
        <span class="param-label">货币对</span>
        <span class="param-value">{{ robot.symbol || robot.tradingPair }}</span>
      </div>
      <div class="param-item">
        <span class="param-label">交易类型</span>
        <span class="param-value">永续合约</span>
      </div>
      <div class="param-item">
        <span class="param-label">订单类型</span>
        <span class="param-value">市价单</span>
      </div>
      <div class="param-item">
        <span class="param-label">保证金模式</span>
        <span class="param-value">逐仓</span>
      </div>
      <div class="param-item highlight">
        <span class="param-label">时间窗口</span>
        <span class="param-value">{{ formatWindowTime(data.signal?.strategyWindow) }}</span>
      </div>
      <div class="param-item highlight">
        <span class="param-label">波动点数</span>
        <span class="param-value">{{ data.signal?.strategyThreshold?.toFixed(1) || '--' }}</span>
      </div>
      <div class="param-item highlight">
        <span class="param-label">杠杆倍数</span>
        <span class="param-value">{{ data.config?.leverage || '--' }}x</span>
      </div>
      <div class="param-item highlight">
        <span class="param-label">保证金比例</span>
        <span class="param-value">{{ data.config?.marginPercent?.toFixed(1) || '--' }}%</span>
      </div>
      <div class="param-item">
        <span class="param-label">止损比例</span>
        <span class="param-value error">{{ data.config?.stopLossPercent?.toFixed(1) || '--' }}%</span>
      </div>
      <div class="param-item">
        <span class="param-label">启动回撤</span>
        <span class="param-value success">{{ robot.profitActivatePercent?.toFixed(1) || '--' }}%</span>
      </div>
      <div class="param-item">
        <span class="param-label">止盈回撤</span>
        <span class="param-value success">{{ data.config?.takeProfitPercent?.toFixed(1) || '--' }}%</span>
      </div>
    </div>

    <div class="strategy-update-time">
      <n-text depth="3" style="font-size: 11px">
        策略参数根据市场状态实时调整 | 更新于 {{ formatUpdateTime(data.lastUpdate) }}
      </n-text>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Robot } from '../composables/useRobotList';
import type { AnalysisData } from '../composables/useRobotStatus';
import {
  formatMarketState,
  formatRiskPref,
  getMarketStateType,
  formatWindowTime,
  formatUpdateTime
} from '../composables/useRobotStatus';

defineProps<{
  robot: Robot;
  data: AnalysisData | undefined;
}>();
</script>

<style scoped>
.market-analysis-panel {
  background: var(--n-color-embedded);
  border-radius: 8px;
  padding: 12px;
  margin-bottom: 12px;
}

.market-analysis-header {
  margin-bottom: 10px;
  padding-bottom: 8px;
  border-bottom: 1px solid var(--n-border-color);
}

.strategy-params-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 8px;
}

.param-item {
  display: flex;
  flex-direction: column;
  padding: 6px 8px;
  background: var(--n-color);
  border-radius: 4px;
}

.param-item.highlight {
  background: rgba(var(--primary-color-rgb), 0.1);
}

.param-label {
  font-size: 11px;
  color: var(--n-text-color-3);
  margin-bottom: 2px;
}

.param-value {
  font-size: 13px;
  font-weight: 600;
}

.param-value.error {
  color: var(--error-color);
}

.param-value.success {
  color: var(--success-color);
}

.strategy-update-time {
  margin-top: 8px;
  text-align: right;
}
</style>

