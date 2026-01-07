<template>
  <div class="signal-alert-panel" v-if="data">
    <!-- 信号头部 -->
    <div class="signal-alert-header">
      <div class="signal-badge" :class="getSignalBadgeClass(data.signal?.direction)">
        <span class="badge-icon">{{ getSignalIcon(data.signal?.direction) }}</span>
        <span class="badge-text">{{ getSignalDirectionText(data.signal?.direction) }}</span>
        <span class="badge-strength" v-if="data.signal?.signalStrength > 0">
          {{ (data.signal?.signalStrength || 0).toFixed(0) }}%
        </span>
      </div>
    </div>

    <!-- 价格面板 -->
    <div class="signal-price-row">
      <div class="price-current">
        <span class="price-label">当前</span>
        <span class="price-value" :class="getPriceChangeClass(ticker?.change24h)">
          {{ formatPrice(ticker?.lastPrice || data.signal?.currentPrice) }}
        </span>
        <span class="price-change" :class="getPriceChangeClass(ticker?.change24h)">
          {{ formatPriceChange(ticker?.change24h) }}
        </span>
      </div>
      <div class="price-window">
        <div class="window-item low">
          <span>窗口低</span>
          <strong>{{ formatPrice(data.signal?.windowMinPrice) }}</strong>
        </div>
        <div class="window-item high">
          <span>窗口高</span>
          <strong>{{ formatPrice(data.signal?.windowMaxPrice) }}</strong>
        </div>
      </div>
    </div>

    <!-- 触发条件 -->
    <div class="signal-triggers" v-if="data.signal?.signalThreshold">
      <div class="trigger long">
        <span class="trigger-name">📈 做多</span>
        <span class="trigger-price">≤ {{ formatPrice(longTriggerPrice) }}</span>
        <span class="trigger-dist">距{{ formatPriceUsdt(data.signal?.distanceFromMin) }}</span>
      </div>
      <div class="trigger short">
        <span class="trigger-name">📉 做空</span>
        <span class="trigger-price">≥ {{ formatPrice(shortTriggerPrice) }}</span>
        <span class="trigger-dist">距{{ formatPriceUsdt(data.signal?.distanceFromMax) }}</span>
      </div>
    </div>

    <!-- 价格曲线 -->
    <div class="signal-chart" v-if="data.priceWindow?.length > 1">
      <svg class="chart-svg" viewBox="0 0 640 120" preserveAspectRatio="none">
        <path :d="chartFillPath" class="chart-fill" />
        <path :d="chartLinePath" class="chart-line" />
        <line x1="0" :y1="baselineY" x2="640" :y2="baselineY" class="chart-baseline" />
      </svg>
    </div>

    <!-- 信号说明 -->
    <div v-if="data.signal?.reason" class="signal-reason">
      {{ data.signal.reason }}
    </div>
  </div>
</template>

<script setup lang="ts">
  import { computed } from 'vue';
  import type { AnalysisData, TickerData } from '../composables/useRobotStatus';
  import {
    formatPrice,
    formatPriceUsdt,
    formatPriceChange,
    getPriceChangeClass,
  } from '../composables/useRobotStatus';

  const props = defineProps<{
    data: AnalysisData | undefined;
    ticker: TickerData | undefined;
  }>();

  // 计算属性
  const longTriggerPrice = computed(() => {
    if (!props.data?.signal) return 0;
    return props.data.signal.windowMinPrice + props.data.signal.signalThreshold;
  });

  const shortTriggerPrice = computed(() => {
    if (!props.data?.signal) return 0;
    return props.data.signal.windowMaxPrice - props.data.signal.signalThreshold;
  });

  // 图表路径计算
  const chartLinePath = computed(() => {
    if (!props.data?.priceWindow?.length) return '';
    const points = props.data.priceWindow;
    const minPrice = Math.min(...points.map((p) => p.price));
    const maxPrice = Math.max(...points.map((p) => p.price));
    const priceRange = maxPrice - minPrice || 1;

    const width = 640;
    const height = 100;
    const padding = 10;

    return points
      .map((p, i) => {
        const x = (i / (points.length - 1)) * width;
        const y = padding + ((maxPrice - p.price) / priceRange) * (height - padding * 2);
        return `${i === 0 ? 'M' : 'L'} ${x} ${y}`;
      })
      .join(' ');
  });

  const chartFillPath = computed(() => {
    if (!chartLinePath.value) return '';
    return chartLinePath.value + ' L 640 110 L 0 110 Z';
  });

  const baselineY = computed(() => {
    if (!props.data?.priceWindow?.length) return 60;
    const points = props.data.priceWindow;
    const minPrice = Math.min(...points.map((p) => p.price));
    const maxPrice = Math.max(...points.map((p) => p.price));
    const avgPrice = (minPrice + maxPrice) / 2;
    const priceRange = maxPrice - minPrice || 1;
    return 10 + ((maxPrice - avgPrice) / priceRange) * 80;
  });

  // 信号相关函数
  function getSignalBadgeClass(direction: string | undefined): string {
    switch (direction) {
      case 'LONG':
        return 'long';
      case 'SHORT':
        return 'short';
      default:
        return 'neutral';
    }
  }

  function getSignalIcon(direction: string | undefined): string {
    switch (direction) {
      case 'LONG':
        return '📈';
      case 'SHORT':
        return '📉';
      default:
        return '⏳';
    }
  }

  function getSignalDirectionText(direction: string | undefined): string {
    switch (direction) {
      case 'LONG':
        return '做多信号';
      case 'SHORT':
        return '做空信号';
      default:
        return '监控中';
    }
  }
</script>

<style scoped>
  .signal-alert-panel {
    background: var(--n-color-embedded);
    border-radius: 8px;
    padding: 12px;
    margin-bottom: 12px;
  }

  .signal-alert-header {
    margin-bottom: 10px;
  }

  .signal-badge {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 6px 12px;
    border-radius: 20px;
    font-weight: 600;
  }

  .signal-badge.long {
    background: rgba(var(--success-color-rgb), 0.15);
    color: var(--success-color);
  }

  .signal-badge.short {
    background: rgba(var(--error-color-rgb), 0.15);
    color: var(--error-color);
  }

  .signal-badge.neutral {
    background: rgba(var(--warning-color-rgb), 0.15);
    color: var(--warning-color);
  }

  .badge-icon {
    font-size: 16px;
  }

  .badge-text {
    font-size: 14px;
  }

  .badge-strength {
    font-size: 12px;
    opacity: 0.8;
  }

  .signal-price-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 10px;
  }

  .price-current {
    display: flex;
    flex-direction: column;
  }

  .price-label {
    font-size: 11px;
    color: var(--n-text-color-3);
  }

  .price-value {
    font-size: 20px;
    font-weight: bold;
  }

  .price-value.up {
    color: var(--success-color);
  }

  .price-value.down {
    color: var(--error-color);
  }

  .price-change {
    font-size: 12px;
  }

  .price-change.up {
    color: var(--success-color);
  }

  .price-change.down {
    color: var(--error-color);
  }

  .price-window {
    display: flex;
    gap: 16px;
  }

  .window-item {
    display: flex;
    flex-direction: column;
    align-items: center;
  }

  .window-item span {
    font-size: 11px;
    color: var(--n-text-color-3);
  }

  .window-item strong {
    font-size: 14px;
  }

  .window-item.low strong {
    color: var(--success-color);
  }

  .window-item.high strong {
    color: var(--error-color);
  }

  .signal-triggers {
    display: flex;
    justify-content: space-between;
    gap: 12px;
    margin-bottom: 10px;
  }

  .trigger {
    flex: 1;
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 12px;
    border-radius: 6px;
    font-size: 12px;
  }

  .trigger.long {
    background: rgba(var(--success-color-rgb), 0.1);
  }

  .trigger.short {
    background: rgba(var(--error-color-rgb), 0.1);
  }

  .trigger-name {
    font-weight: 600;
  }

  .trigger-price {
    flex: 1;
  }

  .trigger-dist {
    color: var(--n-text-color-3);
  }

  .signal-chart {
    height: 120px;
    margin-bottom: 8px;
  }

  .chart-svg {
    width: 100%;
    height: 100%;
  }

  .chart-line {
    fill: none;
    stroke: var(--primary-color);
    stroke-width: 2;
  }

  .chart-fill {
    fill: rgba(var(--primary-color-rgb), 0.1);
  }

  .chart-baseline {
    stroke: var(--n-border-color);
    stroke-width: 1;
    stroke-dasharray: 4 4;
  }

  .signal-reason {
    font-size: 12px;
    color: var(--n-text-color-2);
    padding: 8px;
    background: var(--n-color);
    border-radius: 4px;
  }
</style>
