<template>
  <div class="robot-page">
    <!-- 统计概览 -->
    <n-grid
      cols="1 s:3 m:3 l:5 xl:5 2xl:5"
      :x-gap="12"
      :y-gap="12"
      responsive="screen"
      class="mb-3"
    >
      <n-gi>
        <n-card :bordered="false" size="small">
          <n-statistic label="总机器人" :value="total">
            <template #prefix><n-icon :component="RobotOutlined" /></template>
          </n-statistic>
        </n-card>
      </n-gi>
      <n-gi>
        <n-card :bordered="false" size="small" class="running-card">
          <n-statistic label="运行中" :value="runningCount">
            <template #prefix
              ><n-icon :component="PlayCircleOutlined" color="var(--success-color)"
            /></template>
          </n-statistic>
        </n-card>
      </n-gi>
      <n-gi>
        <n-card :bordered="false" size="small">
          <n-statistic label="今日盈亏">
            <template #default>
              <n-text :type="todayPnl >= 0 ? 'success' : 'error'" strong>
                {{ todayPnl >= 0 ? '+' : '' }}{{ todayPnl.toFixed(2) }}
              </n-text>
            </template>
          </n-statistic>
        </n-card>
      </n-gi>
      <n-gi>
        <n-card :bordered="false" size="small">
          <n-statistic label="累计盈亏">
            <template #default>
              <n-text :type="totalPnl >= 0 ? 'success' : 'error'" strong>
                {{ totalPnl >= 0 ? '+' : '' }}{{ totalPnl.toFixed(2) }}
              </n-text>
            </template>
          </n-statistic>
        </n-card>
      </n-gi>
      <n-gi>
        <n-card :bordered="false" size="small">
          <n-statistic label="消耗算力" :value="totalPower.toFixed(2)">
            <template #prefix><n-icon :component="ThunderboltOutlined" /></template>
          </n-statistic>
        </n-card>
      </n-gi>
    </n-grid>

    <!-- 操作栏 -->
    <n-card :bordered="false" size="small" class="mb-3">
      <n-space justify="space-between" align="center">
        <n-space align="center">
          <n-select
            v-model:value="searchParams.status"
            :options="statusOptions"
            placeholder="状态筛选"
            style="width: 120px"
            clearable
            size="small"
          />
          <n-select
            v-model:value="searchParams.platform"
            :options="platformOptions"
            placeholder="平台筛选"
            style="width: 120px"
            clearable
            size="small"
          />
          <n-button size="small" @click="loadData">
            <template #icon><n-icon :component="ReloadOutlined" /></template>
            刷新
          </n-button>
        </n-space>
        <n-button type="primary" @click="router.push('/toogo/robot/create')">
          <template #icon><n-icon :component="PlusOutlined" /></template>
          创建机器人
        </n-button>
      </n-space>
    </n-card>

    <!-- 机器人列表 -->
    <n-grid
      cols="1 s:1 m:2 l:2 xl:2 2xl:3"
      :x-gap="16"
      :y-gap="16"
      responsive="screen"
      v-if="robotList.length > 0"
    >
      <n-gi v-for="robot in robotList" :key="robot.id">
        <RobotCard
          :robot="robot"
          @start="handleStartRobot"
          @stop="handleStopRobot"
          @detail="viewDetail"
          @config="openRiskConfig"
          @delete="handleDeleteRobot"
        >
          <template #running-content="{ robot: r }">
            <!-- 连接状态 -->
            <div class="status-section">
              <n-space :size="12" align="center" justify="center">
                <div class="status-indicator-item">
                  <span class="indicator-dot" :class="getConnectionStatus(r.id).class"></span>
                  <span class="indicator-label">{{ getConnectionStatus(r.id).text }}</span>
                </div>
              </n-space>
            </div>

            <!-- 市场分析面板 -->
            <MarketAnalysisPanel :robot="r" :data="analysisData[r.id]" />

            <!-- 信号预警面板 -->
            <SignalAlertPanel :data="analysisData[r.id]" :ticker="tickerData[r.id]" />

            <!-- 账户信息 -->
            <div class="account-section" v-if="analysisData[r.id]?.account">
              <n-grid :cols="2" :x-gap="12">
                <n-gi>
                  <div class="account-item">
                    <n-text depth="3" style="font-size: 11px">可用余额</n-text>
                    <n-text strong style="font-size: 14px; color: var(--primary-color)">
                      {{ analysisData[r.id]?.account?.availableBalance?.toFixed(2) || '--' }} U
                    </n-text>
                  </div>
                </n-gi>
                <n-gi>
                  <div class="account-item">
                    <n-text depth="3" style="font-size: 11px">未实现盈亏</n-text>
                    <n-text
                      :type="
                        (analysisData[r.id]?.account?.unrealizedPnl || 0) >= 0 ? 'success' : 'error'
                      "
                      strong
                      style="font-size: 14px"
                    >
                      {{ (analysisData[r.id]?.account?.unrealizedPnl || 0) >= 0 ? '+' : '' }}
                      {{ analysisData[r.id]?.account?.unrealizedPnl?.toFixed(2) || '0.00' }} U
                    </n-text>
                  </div>
                </n-gi>
              </n-grid>
            </div>

            <!-- 持仓信息 -->
            <PositionPanel
              :positions="positionData[r.id]"
              @close="(pos) => closePosition(r, pos)"
            />
          </template>
        </RobotCard>
      </n-gi>
    </n-grid>

    <!-- 空状态 -->
    <n-card v-else :bordered="false">
      <n-empty description="暂无机器人，创建一个开始自动交易吧！" size="large">
        <template #extra>
          <n-button type="primary" size="large" @click="router.push('/toogo/robot/create')">
            <template #icon><n-icon :component="PlusOutlined" /></template>
            创建机器人
          </n-button>
        </template>
      </n-empty>
    </n-card>
  </div>
</template>

<script setup lang="ts">
  import { useRouter } from 'vue-router';
  import { useMessage } from 'naive-ui';
  import {
    RobotOutlined,
    PlayCircleOutlined,
    ThunderboltOutlined,
    ReloadOutlined,
    PlusOutlined,
  } from '@vicons/antd';

  // 组件
  import { RobotCard, MarketAnalysisPanel, SignalAlertPanel, PositionPanel } from './components';

  // Composables
  import { useRobotList, statusOptions, platformOptions } from './composables/useRobotList';
  import { useRobotStatus } from './composables/useRobotStatus';

  // API
  import { closePosition as closePositionApi } from '@/api/toogo/robot';

  const router = useRouter();
  const message = useMessage();

  // 使用机器人列表逻辑
  const {
    robotList,
    total,
    searchParams,
    runningCount,
    todayPnl,
    totalPnl,
    totalPower,
    loadData,
    handleStartRobot,
    handleStopRobot,
    handleDeleteRobot,
  } = useRobotList();

  // 使用机器人状态监控
  const { analysisData, tickerData, positionData, getConnectionStatus } = useRobotStatus(robotList);

  // 查看详情
  const viewDetail = (robot: any) => {
    router.push(`/toogo/robot/detail/${robot.id}`);
  };

  // 打开风险配置
  const openRiskConfig = (robot: any) => {
    // TODO: 打开风险配置弹窗
    message.info('风险配置功能开发中');
  };

  // 平仓
  const closePosition = async (robot: any, position: any) => {
    try {
      const res = await closePositionApi({
        robotId: robot.id,
        positionSide: position.positionSide,
      });
      if (res.code === 0) {
        message.success('平仓成功');
      } else {
        message.error(res.message || '平仓失败');
      }
    } catch (error) {
      message.error('平仓失败');
    }
  };
</script>

<style scoped>
  .robot-page {
    padding: 16px;
  }

  .mb-3 {
    margin-bottom: 12px;
  }

  .running-card {
    border-left: 3px solid var(--success-color);
  }

  .status-section {
    margin-bottom: 12px;
    padding: 8px;
    background: var(--n-color-embedded);
    border-radius: 6px;
  }

  .status-indicator-item {
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .indicator-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
  }

  .indicator-dot.success {
    background: var(--success-color);
    box-shadow: 0 0 6px var(--success-color);
  }

  .indicator-dot.warning {
    background: var(--warning-color);
    animation: pulse 1.5s infinite;
  }

  @keyframes pulse {
    0%,
    100% {
      opacity: 1;
    }
    50% {
      opacity: 0.5;
    }
  }

  .indicator-label {
    font-size: 12px;
    color: var(--n-text-color-2);
  }

  .account-section {
    margin-bottom: 12px;
    padding: 10px;
    background: var(--n-color-embedded);
    border-radius: 6px;
  }

  .account-item {
    display: flex;
    flex-direction: column;
  }
</style>
