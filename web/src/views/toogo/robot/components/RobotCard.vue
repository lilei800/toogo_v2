<template>
  <n-card
    class="robot-card"
    :class="{ running: robot.status === 2 }"
    :bordered="false"
    hoverable
    size="small"
  >
    <!-- 头部 -->
    <template #header>
      <n-space align="center" :size="8">
        <n-tag :type="getStatusType(robot.status)" size="small">
          {{ getStatusText(robot.status) }}
        </n-tag>
        <n-text strong>{{ robot.robotName }}</n-text>
      </n-space>
    </template>
    <template #header-extra>
      <n-space :size="6">
        <n-tag size="small">{{
          robot.exchange?.toUpperCase() || robot.platform?.toUpperCase()
        }}</n-tag>
        <n-tag size="small" type="info">{{ robot.symbol || robot.tradingPair }}</n-tag>
      </n-space>
    </template>

    <!-- 运行中机器人 -->
    <template v-if="robot.status === 2">
      <slot name="running-content" :robot="robot"></slot>
    </template>

    <!-- 非运行状态显示核心配置 -->
    <template v-else>
      <div class="stopped-robot-info">
        <!-- 核心交易参数 -->
        <n-grid :cols="4" :x-gap="8" :y-gap="8">
          <n-gi>
            <div class="param-box">
              <n-text depth="3" class="param-label">杠杆</n-text>
              <div class="param-value">{{ robot.leverage || '--' }}x</div>
            </div>
          </n-gi>
          <n-gi>
            <div class="param-box">
              <n-text depth="3" class="param-label">保证金</n-text>
              <div class="param-value">{{ robot.marginRatio || robot.marginPercent || '--' }}%</div>
            </div>
          </n-gi>
          <n-gi>
            <div class="param-box">
              <n-text depth="3" class="param-label">止损</n-text>
              <div class="param-value error">{{ robot.stopLossPercent || '--' }}%</div>
            </div>
          </n-gi>
          <n-gi>
            <div class="param-box">
              <n-text depth="3" class="param-label">止盈回撤</n-text>
              <div class="param-value success"
                >{{ robot.takeProfitRetracePercent || robot.profitRetreatPercent || '--' }}%</div
              >
            </div>
          </n-gi>
        </n-grid>

        <!-- 统计数据 -->
        <n-divider style="margin: 10px 0" />
        <n-grid :cols="2" :x-gap="16">
          <n-gi>
            <div class="stat-box">
              <n-text depth="3" class="stat-label">累计盈亏</n-text>
              <n-text
                :type="(robot.totalPnl || 0) >= 0 ? 'success' : 'error'"
                strong
                class="stat-value"
              >
                {{ (robot.totalPnl || 0) >= 0 ? '+' : '' }}{{ (robot.totalPnl || 0).toFixed(2) }} U
              </n-text>
            </div>
          </n-gi>
          <n-gi>
            <div class="stat-box">
              <n-text depth="3" class="stat-label">消耗算力</n-text>
              <div class="stat-value">{{ robot.consumedPower?.toFixed(2) || '0.00' }}</div>
            </div>
          </n-gi>
        </n-grid>
      </div>
    </template>

    <!-- 操作按钮 -->
    <template #action>
      <n-space justify="space-between" style="width: 100%">
        <n-space align="center">
          <n-button
            v-if="robot.status === 2"
            type="warning"
            size="small"
            @click="$emit('stop', robot)"
          >
            <template #icon><n-icon :component="PauseCircleOutlined" /></template>
            暂停
          </n-button>
          <n-button
            v-else-if="robot.status === 1"
            type="primary"
            size="small"
            @click="$emit('start', robot)"
          >
            <template #icon><n-icon :component="PlayCircleOutlined" /></template>
            启动
          </n-button>
          <n-button
            v-else-if="robot.status === 3"
            type="primary"
            size="small"
            @click="$emit('start', robot)"
          >
            <template #icon><n-icon :component="PlayCircleOutlined" /></template>
            重启
          </n-button>
          <!-- 定时启动倒计时 -->
          <ScheduleCountdown
            v-if="robot.status !== 2"
            :schedule-time="robot.scheduleStart"
            :robot-status="robot.status"
            mode="start"
          />
          <!-- 定时停止倒计时 -->
          <ScheduleCountdown
            v-if="robot.status === 2"
            :schedule-time="robot.scheduleStop"
            :robot-status="robot.status"
            mode="stop"
          />
        </n-space>
        <n-space>
          <n-button size="small" @click="$emit('detail', robot)">详情</n-button>
          <n-button size="small" @click="$emit('config', robot)">
            <template #icon><n-icon :component="SettingOutlined" /></template>
            配置
          </n-button>
          <n-button
            type="error"
            ghost
            size="small"
            @click="$emit('delete', robot)"
            :disabled="robot.status === 2"
            >删除</n-button
          >
        </n-space>
      </n-space>
    </template>
  </n-card>
</template>

<script setup lang="ts">
  import { PauseCircleOutlined, PlayCircleOutlined, SettingOutlined } from '@vicons/antd';
  import { getStatusType, getStatusText, type Robot } from '../composables/useRobotList';
  import ScheduleCountdown from './ScheduleCountdown.vue';

  defineProps<{
    robot: Robot;
  }>();

  defineEmits<{
    (e: 'start', robot: Robot): void;
    (e: 'stop', robot: Robot): void;
    (e: 'detail', robot: Robot): void;
    (e: 'config', robot: Robot): void;
    (e: 'delete', robot: Robot): void;
  }>();
</script>

<style scoped lang="less">
  .robot-card {
    transition: all 0.3s;
  }

  .robot-card.running {
    border-left: 3px solid var(--success-color);
  }

  .robot-card:hover {
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  }

  .stopped-robot-info {
    .param-box {
      text-align: center;
      padding: 6px;
      background: #f8f9fa;
      border-radius: 6px;

      .param-label {
        font-size: 11px;
        display: block;
      }

      .param-value {
        font-size: 14px;
        font-weight: 600;
        margin-top: 2px;

        &.error {
          color: var(--error-color);
        }
        &.success {
          color: var(--success-color);
        }
      }
    }

    .stat-box {
      .stat-label {
        font-size: 11px;
        display: block;
      }

      .stat-value {
        font-size: 15px;
        font-weight: 600;
      }
    }
  }
</style>
