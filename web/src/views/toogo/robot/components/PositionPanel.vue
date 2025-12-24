<template>
  <div class="position-section">
    <n-text depth="3" style="font-size: 12px; display: block; margin-bottom: 8px">当前持仓</n-text>
    
    <template v-if="positions && positions.length > 0">
      <div v-for="pos in positions" :key="pos.symbol + pos.positionSide" class="position-item">
        <n-grid :cols="5" :x-gap="8" align="center">
          <n-gi>
            <n-tag :type="pos.positionSide === 'LONG' ? 'success' : 'error'" size="small">
              {{ pos.positionSide === 'LONG' ? '做多' : '做空' }}
            </n-tag>
          </n-gi>
          <n-gi>
            <n-text depth="3" style="font-size: 11px">数量</n-text>
            <div style="font-size: 13px">{{ Math.abs(pos.positionAmt).toFixed(4) }}</div>
          </n-gi>
          <n-gi>
            <n-text depth="3" style="font-size: 11px">开仓价</n-text>
            <div style="font-size: 13px">{{ pos.entryPrice?.toFixed(2) || '--' }}</div>
          </n-gi>
          <n-gi>
            <n-text depth="3" style="font-size: 11px">未实现盈亏</n-text>
            <n-text :type="pos.unrealizedPnl >= 0 ? 'success' : 'error'" strong>
              {{ pos.unrealizedPnl >= 0 ? '+' : '' }}{{ pos.unrealizedPnl.toFixed(2) }} U
            </n-text>
          </n-gi>
          <n-gi style="text-align: right">
            <n-button size="tiny" type="warning" @click="$emit('close', pos)">平仓</n-button>
          </n-gi>
        </n-grid>
      </div>
    </template>
    
    <n-empty v-else size="small" description="暂无持仓" />
  </div>
</template>

<script setup lang="ts">
import type { PositionData } from '../composables/useRobotStatus';

defineProps<{
  positions: PositionData[] | undefined;
}>();

defineEmits<{
  (e: 'close', position: PositionData): void;
}>();
</script>

<style scoped>
.position-section {
  margin-top: 12px;
  padding: 12px;
  background: var(--n-color-embedded);
  border-radius: 8px;
}

.position-item {
  padding: 10px;
  background: var(--n-color);
  border-radius: 6px;
  margin-bottom: 8px;
}

.position-item:last-child {
  margin-bottom: 0;
}
</style>

