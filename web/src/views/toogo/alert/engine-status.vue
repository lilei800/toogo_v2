<template>
  <div class="engine-status-page">
    <n-grid :cols="3" :x-gap="16" :y-gap="16">
      <!-- 引擎状态卡片 -->
      <n-gi>
        <n-card title="引擎状态">
          <n-statistic label="运行状态">
            <template #prefix>
              <n-icon :color="engineStatus.running ? '#18a058' : '#d03050'">
                <div
                  :class="engineStatus.running ? 'status-dot running' : 'status-dot stopped'"
                ></div>
              </n-icon>
            </template>
            {{ engineStatus.running ? '运行中' : '已停止' }}
          </n-statistic>
          <n-divider />
          <n-space vertical>
            <n-statistic label="活跃机器人" :value="engineStatus.activeRobots" />
            <n-statistic label="行情订阅数" :value="engineStatus.activeSubscriptions" />
          </n-space>
          <template #action>
            <n-button @click="refreshStatus" :loading="loading">刷新状态</n-button>
          </template>
        </n-card>
      </n-gi>

      <!-- 市场分析卡片 -->
      <n-gi>
        <n-card title="实时市场分析">
          <n-form inline label-placement="left">
            <n-form-item label="交易所">
              <n-select
                v-model:value="analysisForm.platform"
                :options="platformOptions"
                style="width: 120px"
              />
            </n-form-item>
            <n-form-item label="交易对">
              <n-input v-model:value="analysisForm.symbol" style="width: 120px" />
            </n-form-item>
            <n-form-item>
              <n-button type="primary" @click="fetchMarketAnalysis" :loading="analysisLoading">
                查询
              </n-button>
            </n-form-item>
          </n-form>

          <n-divider />

          <template v-if="marketAnalysis">
            <n-descriptions :column="2" label-placement="left">
              <n-descriptions-item label="当前价格">
                {{ marketAnalysis.currentPrice?.toFixed(2) }}
              </n-descriptions-item>
              <n-descriptions-item label="市场状态">
                <n-tag :type="getStateType(marketAnalysis.marketState)">
                  {{ getStateText(marketAnalysis.marketState) }}
                </n-tag>
              </n-descriptions-item>
              <n-descriptions-item label="置信度">
                {{ (marketAnalysis.marketStateConf * 100).toFixed(1) }}%
              </n-descriptions-item>
              <n-descriptions-item label="趋势强度">
                <span :style="{ color: marketAnalysis.trendStrength > 0 ? '#18a058' : '#d03050' }">
                  {{ marketAnalysis.trendStrength?.toFixed(4) }}
                </span>
              </n-descriptions-item>
              <n-descriptions-item label="波动率">
                {{ marketAnalysis.volatility?.toFixed(4) }}
              </n-descriptions-item>
              <n-descriptions-item label="支撑位">
                {{ marketAnalysis.supportLevel?.toFixed(2) }}
              </n-descriptions-item>
              <n-descriptions-item label="阻力位">
                {{ marketAnalysis.resistanceLevel?.toFixed(2) }}
              </n-descriptions-item>
            </n-descriptions>

            <n-divider>技术指标</n-divider>
            <n-descriptions :column="2" label-placement="left" v-if="marketAnalysis.indicators">
              <n-descriptions-item label="趋势评分">
                {{ marketAnalysis.indicators.trendScore?.toFixed(2) }}
              </n-descriptions-item>
              <n-descriptions-item label="动量评分">
                {{ marketAnalysis.indicators.momentumScore?.toFixed(2) }}
              </n-descriptions-item>
              <n-descriptions-item label="波动评分">
                {{ marketAnalysis.indicators.volatilityScore?.toFixed(2) }}
              </n-descriptions-item>
              <n-descriptions-item label="综合评分">
                {{ marketAnalysis.indicators.compositeScore?.toFixed(2) }}
              </n-descriptions-item>
            </n-descriptions>
          </template>
          <n-empty v-else description="请查询市场分析数据" />
        </n-card>
      </n-gi>

      <!-- 方向信号卡片 -->
      <n-gi>
        <n-card title="实时方向信号">
          <template v-if="directionSignal">
            <n-descriptions :column="1" label-placement="left">
              <n-descriptions-item label="方向">
                <n-tag :type="getDirectionType(directionSignal.direction)" size="large">
                  {{ getDirectionText(directionSignal.direction) }}
                </n-tag>
              </n-descriptions-item>
              <n-descriptions-item label="信号强度">
                <n-progress
                  type="line"
                  :percentage="directionSignal.strength"
                  :status="directionSignal.strength > 70 ? 'success' : 'warning'"
                />
              </n-descriptions-item>
              <n-descriptions-item label="置信度">
                <n-progress
                  type="line"
                  :percentage="directionSignal.confidence"
                  :status="directionSignal.confidence > 70 ? 'success' : 'warning'"
                />
              </n-descriptions-item>
              <n-descriptions-item label="建议操作">
                <n-tag :type="getActionType(directionSignal.action)">
                  {{ getActionText(directionSignal.action) }}
                </n-tag>
              </n-descriptions-item>
              <n-descriptions-item label="入场价">
                {{ directionSignal.entryPrice?.toFixed(2) }}
              </n-descriptions-item>
              <n-descriptions-item label="止损价">
                <span style="color: #d03050">{{ directionSignal.stopLoss?.toFixed(2) }}</span>
              </n-descriptions-item>
              <n-descriptions-item label="止盈目标">
                <span style="color: #18a058">{{ directionSignal.takeProfit1?.toFixed(2) }}</span>
              </n-descriptions-item>
              <n-descriptions-item label="信号原因">
                {{ directionSignal.reason }}
              </n-descriptions-item>
            </n-descriptions>
          </template>
          <n-empty v-else description="请先查询市场分析" />
        </n-card>
      </n-gi>
    </n-grid>

    <!-- 多周期数据 -->
    <n-card title="多周期分析数据" style="margin-top: 16px" v-if="marketAnalysis?.timeframeData">
      <n-data-table :columns="timeframeColumns" :data="timeframeTableData" :bordered="false" />
    </n-card>
  </div>
</template>

<script lang="ts" setup>
  import { ref, reactive, onMounted, computed } from 'vue';
  import { getEngineStatus, getMarketAnalysis, getDirectionSignal } from '@/api/trading/alert';

  const loading = ref(false);
  const analysisLoading = ref(false);

  const engineStatus = reactive({
    running: false,
    activeRobots: 0,
    activeSubscriptions: 0,
  });

  const analysisForm = reactive({
    platform: 'binance',
    symbol: 'BTCUSDT',
  });

  const marketAnalysis = ref<any>(null);
  const directionSignal = ref<any>(null);

  const platformOptions = [
    { label: 'Binance', value: 'binance' },
    { label: 'OKX', value: 'okx' },
    { label: 'Gate', value: 'gate' },
  ];

  const getStateType = (state: string) => {
    const map: Record<string, 'success' | 'warning' | 'error' | 'info'> = {
      trend: 'success',
      volatile: 'warning',
      high_vol: 'error',
      low_vol: 'info',
    };
    return map[state] || 'default';
  };

  const getStateText = (state: string) => {
    const map: Record<string, string> = {
      trend: '趋势',
      volatile: '震荡',
      high_vol: '高波动',
      low_vol: '低波动',
    };
    return map[state] || state;
  };

  const getDirectionType = (dir: string) => {
    const map: Record<string, 'success' | 'error' | 'default'> = {
      LONG: 'success',
      SHORT: 'error',
      NEUTRAL: 'default',
    };
    return map[dir] || 'default';
  };

  const getDirectionText = (dir: string) => {
    const map: Record<string, string> = {
      LONG: '📈 做多',
      SHORT: '📉 做空',
      NEUTRAL: '➖ 中性',
    };
    return map[dir] || dir;
  };

  const getActionType = (action: string) => {
    const map: Record<string, 'success' | 'error' | 'warning' | 'info' | 'default'> = {
      OPEN_LONG: 'success',
      OPEN_SHORT: 'error',
      CLOSE_LONG: 'warning',
      CLOSE_SHORT: 'warning',
      HOLD: 'info',
      WAIT: 'default',
    };
    return map[action] || 'default';
  };

  const getActionText = (action: string) => {
    const map: Record<string, string> = {
      OPEN_LONG: '开多',
      OPEN_SHORT: '开空',
      CLOSE_LONG: '平多',
      CLOSE_SHORT: '平空',
      HOLD: '持有',
      WAIT: '等待',
    };
    return map[action] || action;
  };

  const timeframeColumns = [
    { title: '周期', key: 'interval', width: 80 },
    { title: '趋势', key: 'trend', width: 80 },
    { title: '强度', key: 'strength', width: 100 },
    { title: 'MA5', key: 'ma5', width: 100 },
    { title: 'MA10', key: 'ma10', width: 100 },
    { title: 'MA20', key: 'ma20', width: 100 },
    { title: 'MACD', key: 'macd', width: 100 },
    { title: 'RSI', key: 'rsi', width: 80 },
    { title: 'ATR', key: 'atr', width: 100 },
    { title: '形态', key: 'pattern', width: 120 },
  ];

  const timeframeTableData = computed(() => {
    if (!marketAnalysis.value?.timeframeData) return [];
    const intervals = ['1m', '5m', '15m', '30m', '1h'];
    return intervals.map((interval) => {
      const data = marketAnalysis.value.timeframeData[interval] || {};
      return {
        interval,
        trend: data.trend || '-',
        strength: data.strength?.toFixed(4) || '-',
        ma5: data.ma5?.toFixed(2) || '-',
        ma10: data.ma10?.toFixed(2) || '-',
        ma20: data.ma20?.toFixed(2) || '-',
        macd: data.macd?.toFixed(4) || '-',
        rsi: data.rsi?.toFixed(2) || '-',
        atr: data.atr?.toFixed(4) || '-',
        pattern: data.pattern || '-',
      };
    });
  });

  const refreshStatus = async () => {
    loading.value = true;
    try {
      const res = await getEngineStatus();
      Object.assign(engineStatus, res);
    } catch (error) {
      console.error(error);
    } finally {
      loading.value = false;
    }
  };

  const fetchMarketAnalysis = async () => {
    analysisLoading.value = true;
    try {
      const [analysisRes, signalRes] = await Promise.all([
        getMarketAnalysis(analysisForm.platform, analysisForm.symbol),
        getDirectionSignal(analysisForm.platform, analysisForm.symbol),
      ]);
      marketAnalysis.value = analysisRes;
      directionSignal.value = signalRes;
    } catch (error) {
      console.error(error);
    } finally {
      analysisLoading.value = false;
    }
  };

  onMounted(() => {
    refreshStatus();
  });
</script>

<style scoped>
  .engine-status-page {
    padding: 16px;
  }

  .status-dot {
    width: 12px;
    height: 12px;
    border-radius: 50%;
    display: inline-block;
    margin-right: 8px;
  }

  .status-dot.running {
    background-color: #18a058;
    animation: pulse 1.5s ease-in-out infinite;
  }

  .status-dot.stopped {
    background-color: #d03050;
  }

  @keyframes pulse {
    0% {
      box-shadow: 0 0 0 0 rgba(24, 160, 88, 0.4);
    }
    70% {
      box-shadow: 0 0 0 10px rgba(24, 160, 88, 0);
    }
    100% {
      box-shadow: 0 0 0 0 rgba(24, 160, 88, 0);
    }
  }
</style>
