<template>
  <div class="robot-page">
    <!-- 统计概览 -->
    <n-grid cols="2 s:2 m:2 l:4 xl:4 2xl:4" :x-gap="8" :y-gap="8" responsive="screen" class="mb-2">
      <n-gi>
        <n-card :bordered="false" size="small" content-style="padding: 8px 12px;">
          <n-statistic label="总机器人" :value="total">
            <template #prefix><n-icon :component="RobotOutlined" /></template>
          </n-statistic>
        </n-card>
      </n-gi>
      <n-gi>
        <n-card :bordered="false" size="small" class="running-card" content-style="padding: 8px 12px;">
          <n-statistic label="运行中" :value="runningCount">
            <template #prefix><n-icon :component="PlayCircleOutlined" color="var(--success-color)" /></template>
          </n-statistic>
        </n-card>
      </n-gi>
      <n-gi>
        <n-card :bordered="false" size="small" content-style="padding: 8px 12px;">
          <n-statistic label="今日净盈亏">
            <template #default>
              <n-text :type="todayNetPnl >= 0 ? 'success' : 'error'" strong>
                {{ todayNetPnl >= 0 ? '+' : '' }}{{ todayNetPnl.toFixed(2) }}
              </n-text>
            </template>
          </n-statistic>
        </n-card>
      </n-gi>
      <n-gi>
        <n-card :bordered="false" size="small" content-style="padding: 8px 12px;">
          <n-statistic label="累计净盈亏">
            <template #default>
              <n-text :type="totalNetPnl >= 0 ? 'success' : 'error'" strong>
                {{ totalNetPnl >= 0 ? '+' : '' }}{{ totalNetPnl.toFixed(2) }}
              </n-text>
            </template>
          </n-statistic>
        </n-card>
      </n-gi>
    </n-grid>

    <!-- 操作栏 -->
    <n-card :bordered="false" size="small" class="mb-3 filter-toolbar-card">
      <n-space justify="space-between" align="center" :wrap="false" :size="16">
        <!-- 左侧筛选区 -->
        <n-space align="center" :size="12" :wrap="false">
          <n-tag :bordered="false" size="small" style="font-weight: 600;">
            筛选条件
          </n-tag>
          <n-select 
            v-model:value="searchParams.status" 
            :options="statusOptions" 
            placeholder="选择状态" 
            style="min-width: 140px"
            clearable 
            size="small"
            :consistent-menu-width="false"
            @update:value="loadData"
          />
          <n-select 
            v-model:value="searchParams.platform" 
            :options="platformOptions" 
            placeholder="选择交易所" 
            style="min-width: 140px"
            clearable 
            size="small"
            :consistent-menu-width="false"
            @update:value="loadData"
          />
          <n-button size="small" secondary @click="loadData">
            <template #icon><n-icon :component="ReloadOutlined" /></template>
            刷新
          </n-button>
        </n-space>
        
        <!-- 右侧操作区 -->
        <n-space align="center" :size="12" :wrap="false">
          <n-tag :bordered="false" size="small" style="font-weight: 600;">
            视图模式
          </n-tag>
          <n-radio-group v-model:value="robotListViewMode" size="small" @update:value="persistRobotListViewMode">
            <n-radio-button value="card">卡片</n-radio-button>
            <n-radio-button value="table">列表</n-radio-button>
          </n-radio-group>
          <n-divider vertical />
          <n-button type="primary" size="small" @click="router.push('/toogo/robot/create')">
            <template #icon><n-icon :component="PlusOutlined" /></template>
            创建机器人
          </n-button>
        </n-space>
      </n-space>
    </n-card>

    <!-- 机器人列表 -->
    <template v-if="robotList.length > 0">
      <!-- 列表视图 -->
      <n-card v-if="robotListViewMode === 'table'" :bordered="false" size="small" class="mb-3 robot-list-table-card">
        <n-data-table
          :columns="robotTableColumns"
          :data="robotList"
          :loading="loading"
          :row-key="(row) => row.id"
          size="medium"
          striped
          class="robot-list-table"
        />
      </n-card>

      <!-- 卡片视图（原样保留） -->
      <n-grid v-else cols="1 s:1 m:2 l:2 xl:2 2xl:3" :x-gap="16" :y-gap="16" responsive="screen">
        <n-gi v-for="robot in robotList" :key="robot.id">
        <n-card 
          class="robot-card" 
          :class="{ 'running': robot.status === 2 }" 
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
              <!-- 运行中：API连接状态 -->
              <template v-if="robot.status === 2">
                <span class="header-connection" :class="getConnectionStatus(robot.id).class">
                  <span class="conn-dot"></span>
                  <span class="conn-text">{{ getConnectionStatus(robot.id).text }}</span>
                </span>
              </template>
            </n-space>
          </template>
          <template #header-extra>
            <n-space :size="6">
              <n-tag size="small">{{ robot.exchange?.toUpperCase() || robot.platform?.toUpperCase() }}</n-tag>
              <n-tag size="small" type="info">{{ robot.symbol || robot.tradingPair }}</n-tag>
            </n-space>
          </template>

          <!-- 运行中机器人 -->
          <template v-if="robot.status === 2">
            <!-- 账户+策略参数整合面板 -->
            <div class="account-strategy-panel" v-if="analysisData[robot.id]">
              <!-- 统计信息 -->
              <div class="account-stats-row">
                <div class="stat-item">
                  <span class="stat-label">账户权益</span>
                  <span class="stat-value primary">{{ (analysisData[robot.id]?.account?.accountEquity || analysisData[robot.id]?.account?.totalBalance)?.toFixed(2) || '--' }}</span>
                </div>
                <div class="stat-divider"></div>
                <div class="stat-item">
                  <span class="stat-label">可用余额</span>
                  <span class="stat-value primary">{{ analysisData[robot.id]?.account?.availableBalance?.toFixed(2) || '--' }}</span>
                </div>
                <div class="stat-divider"></div>
                <div class="stat-item">
                  <span class="stat-label">净盈亏</span>
                  <span
                    class="stat-value"
                    :class="getRunningSessionNetPnl(robot.id) === null ? '' : (Number(getRunningSessionNetPnl(robot.id)) >= 0 ? 'success' : 'error')"
                  >
                    {{
                      getRunningSessionNetPnl(robot.id) === null
                        ? '--'
                        : ((Number(getRunningSessionNetPnl(robot.id)) >= 0 ? '+' : '') + Number(getRunningSessionNetPnl(robot.id)).toFixed(6))
                    }}
                  </span>
                </div>
                <div class="stat-divider"></div>
                <div class="stat-item">
                  <span class="stat-label">运行</span>
                  <span class="stat-value">{{ formatRuntime(getRobotRuntimeSeconds(robot)) }}</span>
                </div>
              </div>
            </div>

            <!-- 方向信号预警（三列布局：机器人 | 预警按钮 | 图表） -->
            <div class="signal-alert-panel" v-if="analysisData[robot.id]">
              <div class="signal-three-column">
                <!-- 第一列：机器人动画 -->
                <div class="column-robot">
                  <div class="mini-robot-scene" :class="getRobotMoodClass(robot)">
                    <svg class="mini-robot" viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg">
                      <defs>
                        <linearGradient id="mini-metal" x1="0%" y1="0%" x2="100%" y2="100%">
                          <stop offset="0%" style="stop-color: #ffffff" />
                          <stop offset="50%" style="stop-color: #e0e0e0" />
                          <stop offset="100%" style="stop-color: #b0b0b0" />
                        </linearGradient>
                        <radialGradient :id="'eye-glow-' + robot.id" cx="50%" cy="50%" r="50%">
                          <stop offset="0%" :style="'stop-color:' + getRobotEyeColor(robot)" />
                          <stop offset="100%" :style="'stop-color:' + getRobotEyeColor(robot)" />
                        </radialGradient>
                      </defs>
                      <ellipse cx="50" cy="92" rx="18" ry="4" fill="rgba(0,0,0,0.15)" class="robot-shadow" />
                      <g class="tail-group">
                        <path d="M68 68 Q 78 60 82 50" stroke="#ccc" stroke-width="2" fill="none" stroke-linecap="round" class="tail" />
                        <circle cx="82" cy="50" r="2" :fill="getRobotEyeColor(robot)" class="tail-tip" />
                      </g>
                      <rect x="35" y="55" width="30" height="25" rx="5" fill="url(#mini-metal)" class="body" />
                      <rect x="40" y="60" width="20" height="15" rx="3" fill="#111" class="body-screen" />
                      <text x="50" y="70" text-anchor="middle" :fill="getRobotEyeColor(robot)" font-size="4" class="logo-text">Toogo</text>
                      <rect x="38" y="78" width="5" height="10" rx="2" fill="#bbb" />
                      <rect x="57" y="78" width="5" height="10" rx="2" fill="#bbb" />
                      <g class="head-group">
                        <path d="M32 38 L 28 25 L 38 30 Z" fill="#ccc" class="ear left-ear" />
                        <path d="M68 38 L 72 25 L 62 30 Z" fill="#ccc" class="ear right-ear" />
                        <rect x="30" y="30" width="40" height="28" rx="8" fill="url(#mini-metal)" class="head-shell" />
                        <rect x="34" y="34" width="32" height="20" rx="5" fill="#111" class="face-screen" />
                        <path d="M40 40 L 45 40" :stroke="getRobotEyeColor(robot)" stroke-width="1" fill="none" class="eyebrow left-eyebrow" />
                        <path d="M55 40 L 60 40" :stroke="getRobotEyeColor(robot)" stroke-width="1" fill="none" class="eyebrow right-eyebrow" />
                        <circle cx="43" cy="45" r="3" :fill="'url(#eye-glow-' + robot.id + ')'" class="eye left-eye" :style="'filter: drop-shadow(0 0 2px ' + getRobotEyeColor(robot) + ')'" />
                        <circle cx="57" cy="45" r="3" :fill="'url(#eye-glow-' + robot.id + ')'" class="eye right-eye" :style="'filter: drop-shadow(0 0 2px ' + getRobotEyeColor(robot) + ')'" />
                        <path d="M45 51 Q 50 54 55 51" stroke="#00f3ff" stroke-width="1" fill="none" class="mouth-smile" />
                      </g>
                    </svg>
                    <div class="robot-mood-text">{{ getRobotMoodText(robot) }}</div>
                  </div>
                </div>
                
                <!-- 第二列：方向预警按钮 + 距离信息 -->
                <div class="column-signal">
                  <!-- 做多预警 -->
                  <div class="signal-block long" :class="{ active: analysisData[robot.id]?.signal?.direction === 'LONG' }">
                    <div class="signal-header">
                      <span class="signal-icon">📈</span>
                      <span class="signal-label">做多</span>
                    </div>
                    <div class="signal-trigger">距 {{ formatPriceUsdt(analysisData[robot.id]?.signal?.distanceFromMin) }}</div>
                    <div class="signal-distance">触发 {{ formatPrice(getLongTriggerPrice(analysisData[robot.id])) }}</div>
                  </div>
                  <!-- 做空预警 -->
                  <div class="signal-block short" :class="{ active: analysisData[robot.id]?.signal?.direction === 'SHORT' }">
                    <div class="signal-header">
                      <span class="signal-icon">📉</span>
                      <span class="signal-label">做空</span>
                    </div>
                    <div class="signal-trigger">距 {{ formatPriceUsdt(analysisData[robot.id]?.signal?.distanceFromMax) }}</div>
                    <div class="signal-distance">触发 {{ formatPrice(getShortTriggerPrice(analysisData[robot.id])) }}</div>
                  </div>
                  <!-- 当前价格 -->
                  <div class="current-price-block">
                    <span class="price-value" :class="getPriceChangeClass(robot.id)">
                      {{ formatPrice(tickerData[robot.id]?.markPrice || tickerData[robot.id]?.lastPrice || analysisData[robot.id]?.signal?.currentPrice) }}
                    </span>
                    <span class="price-change" :class="getPriceChangeClass(robot.id)">
                      {{ formatPriceChange(tickerData[robot.id]?.change24h) }}
                    </span>
                  </div>
                </div>
                
                <!-- 第三列：价格图表（动态最大化） -->
                <div class="column-chart" v-if="analysisData[robot.id]?.priceWindow?.length > 1">
                   <svg class="chart-svg" viewBox="0 0 640 200" preserveAspectRatio="none">
                    <!-- 价格区域填充 -->
                    <path :d="getPriceChartFillPath(analysisData[robot.id])" class="chart-fill" />
                    <!-- 价格曲线 -->
                    <path :d="getPriceChartPath(analysisData[robot.id])" class="chart-line" />
                    <!-- 基准线 -->
                    <line x1="0" :y1="getBaselineY(analysisData[robot.id])" x2="640" :y2="getBaselineY(analysisData[robot.id])" class="chart-baseline" />
                    <!-- 做空触发线 -->
                    <line v-if="analysisData[robot.id]?.signal?.signalThreshold" x1="0" :y1="getShortThresholdY(analysisData[robot.id])" x2="640" :y2="getShortThresholdY(analysisData[robot.id])" class="chart-upper" />
                    <!-- 做多触发线 -->
                    <line v-if="analysisData[robot.id]?.signal?.signalThreshold" x1="0" :y1="getLongThresholdY(analysisData[robot.id])" x2="640" :y2="getLongThresholdY(analysisData[robot.id])" class="chart-lower" />
                    
                    <!-- 最低价点 + 标签 -->
                    <circle :cx="getMinPriceX(analysisData[robot.id])" :cy="getMinPriceY(analysisData[robot.id])" r="4" class="point-min" />
                    <text :x="getMinPriceX(analysisData[robot.id])" :y="getMinPriceY(analysisData[robot.id]) + 14" class="price-label price-label-min">
                      低 {{ formatPrice(analysisData[robot.id]?.signal?.windowMinPrice) }}
                    </text>
                    
                    <!-- 最高价点 + 标签 -->
                    <circle :cx="getMaxPriceX(analysisData[robot.id])" :cy="getMaxPriceY(analysisData[robot.id])" r="4" class="point-max" />
                    <text :x="getMaxPriceX(analysisData[robot.id])" :y="getMaxPriceY(analysisData[robot.id]) - 6" class="price-label price-label-max">
                      高 {{ formatPrice(analysisData[robot.id]?.signal?.windowMaxPrice) }}
                    </text>
                    
                    <!-- 实时价点 + 标签 -->
                    <circle :cx="getCurrentPriceX(analysisData[robot.id])" :cy="getCurrentPriceY(analysisData[robot.id])" r="5" class="point-current" />
                    <text :x="getCurrentPriceX(analysisData[robot.id]) + 8" :y="getCurrentPriceY(analysisData[robot.id]) + 4" class="price-label price-label-current">
                      {{ formatPrice(analysisData[robot.id]?.signal?.currentPrice || tickerData[robot.id]?.markPrice || tickerData[robot.id]?.lastPrice) }}
                    </text>
                  </svg>
                  <div class="chart-labels">
                    <span class="label-high">高 {{ formatPrice(analysisData[robot.id]?.signal?.windowMaxPrice) }}</span>
                    <span class="label-low">低 {{ formatPrice(analysisData[robot.id]?.signal?.windowMinPrice) }}</span>
              </div>

                  <!-- 信号说明（移到图表下方，去掉背景色和边框） -->
              <div v-if="analysisData[robot.id]?.signal?.reason" class="signal-reason">
                {{ analysisData[robot.id]?.signal?.reason }}
                  </div>
                </div>
              </div>

              <!-- 运行中：窗口/波动/杠杆等整行 + 多周期播报（迁移到预警记录上方） -->
              <div class="signal-prelogs-panels">
                <!-- 市场状态与风险偏好映射（默认折叠，展开可修改；参考创建机器人页） -->
                <n-collapse
                  :expanded-names="riskMappingExpandedNames[robot.id] ?? riskMappingExpandedEmpty"
                  style="margin: 0 0 8px 0;"
                  @update:expanded-names="(names) => onRiskMappingExpanded(robot.id, names as any)"
                >
                  <n-collapse-item name="marketMapping">
                    <template #header>
                      <span style="font-size: 12px; font-weight: 500">风险偏好</span>
                    </template>
                    <template #header-extra>
                      <n-space :size="4">
                        <n-tag
                          v-for="market in marketStateMapping"
                          :key="market.key"
                          :type="market.tagType"
                          size="small"
                          :bordered="false"
                          class="risk-mapping-tag"
                        >
                          {{ market.label.replace('市场', '') }}→{{ getRiskLabel(marketRiskMappingForm[robot.id]?.[market.key]) }}
                        </n-tag>
                      </n-space>
                    </template>

                    <div style="padding: 6px 0 2px;">
                      <n-text depth="3" style="font-size: 12px; margin-bottom: 10px; display: block;">
                        机器人运行时会根据市场状态自动选择对应的风险偏好，并据此匹配策略模板
                      </n-text>

                      <n-spin :show="!!riskMappingLoading[robot.id]">
                        <n-grid :cols="4" :x-gap="12" :y-gap="12">
                          <n-gi v-for="market in marketStateMapping" :key="market.key">
                            <div class="mapping-item-card">
                              <div class="mapping-header">
                                <n-tag :type="market.tagType" size="small">{{ market.label }}</n-tag>
                              </div>
                              <div class="mapping-arrow">↓</div>
                              <n-select
                                v-model:value="marketRiskMappingForm[robot.id][market.key]"
                                :options="riskPreferenceSelectOptions"
                                size="small"
                                style="width: 100%"
                              />
                            </div>
                          </n-gi>
                        </n-grid>

                        <div style="display: flex; justify-content: flex-end; margin-top: 10px;">
                          <n-button
                            size="small"
                            type="primary"
                            :loading="!!riskMappingSaving[robot.id]"
                            @click="saveMarketRiskMapping(robot.id)"
                          >
                            保存映射
                          </n-button>
                        </div>
                      </n-spin>
                    </div>
                  </n-collapse-item>
                </n-collapse>

                <!-- 多周期市场状态实时播报 -->
                <div v-if="analysisData[robot.id]?.marketStateRealtime" class="market-realtime-panel" style="margin-top: 0;">
                  <n-space justify="space-between" align="center">
                    <n-space align="center" :size="8">
                      <n-button text size="tiny" @click="toggleMarketRealtime(robot.id)">
                        {{ marketRealtimeExpanded[robot.id] ? '收起' : '展开' }} 多周期播报
                      </n-button>
                      <n-tag
                        size="tiny"
                        :bordered="false"
                        :type="getMarketStateType(analysisData[robot.id]?.marketStateRealtime?.state)"
                      >
                        {{ formatMarketState(analysisData[robot.id]?.marketStateRealtime?.state) }}
                      </n-tag>
                      <n-tag type="warning" size="tiny" :bordered="false" style="opacity: 0.7;">
                        {{ formatRiskPref(analysisData[robot.id]?.config?.riskPreference) }}
                      </n-tag>
                      <n-text depth="3" style="font-size: 12px">
                        投票占比 {{ Math.round((analysisData[robot.id]?.marketStateRealtime?.voteRatio || 0) * 100) }}%
                      </n-text>
                    </n-space>
                    <n-text depth="3" style="font-size: 11px">
                      {{ analysisData[robot.id]?.marketStateRealtime?.updatedAt }}
                    </n-text>
                  </n-space>

                  <n-collapse-transition :show="!!marketRealtimeExpanded[robot.id]">
                    <div style="margin-top: 8px">
                      <n-text depth="3" style="font-size: 11px; display: block; margin-bottom: 6px;">
                        趋势：价格一直往一个方向走（涨或跌）
                        <br />
                        震荡：价格上下波动，但没有明显方向
                        <br />
                        高波动：价格来回跑得很快、很猛
                        <br />
                        低波动：价格动得很慢，很磨人
                      </n-text>
                      <n-space :size="6" style="flex-wrap: wrap;">
                        <n-tag
                          v-for="tf in (analysisData[robot.id]?.marketStateRealtime?.timeframes || [])"
                          :key="tf.interval"
                          size="small"
                          :bordered="false"
                          :type="getMarketStateType(tf.smoothedState)"
                        >
                          {{ tf.interval }} {{ formatMarketState(tf.smoothedState) }}
                          <n-text depth="3" style="font-size: 10px; margin-left: 4px;">V{{ (tf.v ?? 0).toFixed(1) }} D{{ (tf.d ?? 0).toFixed(1) }}</n-text>
                        </n-tag>
                      </n-space>
                    </div>
                  </n-collapse-transition>
                </div>

                <!-- 策略参数（窗口/波动/杠杆等整行）：放在预警记录行正上方 -->
                <div class="strategy-params-container" style="margin-top: 10px;">
                  <n-alert v-if="getConfigError(robot.id)" type="warning" size="small" style="margin-bottom: 8px" :show-icon="true">
                    <template #header>
                      <n-space align="center" :size="4">
                        <n-icon :component="WarningOutlined" />
                        <span>策略模板加载失败，使用数据库静态值</span>
                      </n-space>
                    </template>
                    {{ getConfigError(robot.id) }}
                  </n-alert>

                  <div class="param-text-item">
                    <span class="label">窗口:</span>
                    <span class="value highlight">{{ formatWindowTime(analysisData[robot.id]?.config?.timeWindow || analysisData[robot.id]?.signal?.strategyWindow) }}</span>
                  </div>
                  <div class="param-text-item">
                    <span class="label">波动:</span>
                    <span class="value highlight">{{ (analysisData[robot.id]?.config?.threshold || analysisData[robot.id]?.signal?.strategyThreshold)?.toFixed(1) || '--' }}U</span>
                  </div>
                  <div class="param-text-item">
                    <span class="label">杠杆:</span>
                    <span class="value">{{ getRobotLeverage(robot.id) > 0 ? `${getRobotLeverage(robot.id)}x` : '--' }}</span>
                  </div>
                  <div class="param-text-item">
                    <span class="label">保证金:</span>
                    <span class="value">{{ getRobotMarginPercent(robot.id) > 0 ? `${getRobotMarginPercent(robot.id).toFixed(0)}%` : '--' }}</span>
                  </div>
                  <div class="param-text-item">
                    <span class="label">止损:</span>
                    <span class="value error">{{ getRobotStopLossPercent(robot.id) > 0 ? `${getRobotStopLossPercent(robot.id).toFixed(1)}%` : '--' }}</span>
                  </div>
                  <div class="param-text-item">
                    <span class="label">启动止盈:</span>
                    <span class="value success">{{ getRobotAutoStartRetreat(robot.id) > 0 ? `${getRobotAutoStartRetreat(robot.id).toFixed(1)}%` : '--' }}</span>
                  </div>
                  <div class="param-text-item">
                    <span class="label">止盈回撤:</span>
                    <span class="value success">{{ getRobotProfitRetreat(robot.id) > 0 ? `${getRobotProfitRetreat(robot.id).toFixed(1)}%` : '--' }}</span>
                  </div>
                </div>
              </div>

              <!-- 预警记录（可展开） -->
              <n-collapse :default-expanded-names="[]" style="margin-top: 6px">
                <n-collapse-item name="logs">
                  <template #header>
                    <span style="font-size: 12px; font-weight: 500">预警记录 ({{ signalLogs[robot.id]?.length || 0 }})</span>
                  </template>
                  <div class="signal-logs-list" v-if="signalLogs[robot.id]?.length > 0">
                    <n-card
                      v-for="(log, idx) in signalLogs[robot.id]?.slice(0, 10)"
                      :key="idx"
                      :bordered="false"
                      size="small"
                      class="signal-log-card"
                      :class="log.signalType?.toLowerCase()"
                      content-style="padding: 10px 12px;"
                    >
                      <!-- 日志头部 -->
                      <div class="signal-log-header">
                        <n-space align="center" :size="8">
                          <n-text depth="3" style="font-size: 11px; font-family: 'JetBrains Mono', monospace;">
                            {{ formatLogTime(log.createdAt) }}
                          </n-text>
                          <n-tag 
                            :type="log.signalType?.toLowerCase() === 'long' ? 'success' : 'error'" 
                            size="small"
                            :bordered="false"
                          >
                            {{ log.signalType?.toLowerCase() === 'long' ? '做多' : '做空' }}
                          </n-tag>
                          <n-text strong style="font-size: 12px; font-family: 'JetBrains Mono', monospace;">
                            {{ log.currentPrice?.toFixed(2) }}
                          </n-text>
                          <n-tag 
                            :type="log.isProcessed ? 'default' : 'info'" 
                            size="small"
                            :bordered="false"
                          >
                            {{ log.isProcessed ? '已读' : '未读' }}
                          </n-tag>
                        </n-space>
                      </div>
                      
                      <!-- 信号详情 -->
                      <div class="signal-log-details" v-if="log.windowMinPrice || log.windowMaxPrice || log.threshold || log.marketState">
                        <n-divider style="margin: 10px 0;" />
                        <n-grid :cols="4" :x-gap="12" :y-gap="8">
                          <n-gi v-if="log.windowMinPrice && log.windowMaxPrice">
                            <div class="detail-item">
                              <n-text depth="3" style="font-size: 11px;">价格窗口</n-text>
                              <n-text strong style="font-size: 12px; font-family: 'JetBrains Mono', monospace; margin-left: 6px;">
                                {{ log.windowMinPrice?.toFixed(1) }} ~ {{ log.windowMaxPrice?.toFixed(1) }}
                              </n-text>
                            </div>
                          </n-gi>
                          <n-gi v-if="log.threshold">
                            <div class="detail-item">
                              <n-text depth="3" style="font-size: 11px;">波动阈值</n-text>
                              <n-text strong style="font-size: 12px; font-family: 'JetBrains Mono', monospace; margin-left: 6px;">
                                {{ log.threshold?.toFixed(1) }} USDT
                              </n-text>
                            </div>
                          </n-gi>
                          <n-gi v-if="log.marketState">
                            <div class="detail-item">
                              <n-text depth="3" style="font-size: 11px;">市场状态</n-text>
                              <n-tag 
                                :type="getMarketStateType(log.marketState)" 
                                size="small"
                                :bordered="false"
                                style="margin-left: 6px;"
                              >
                                {{ formatMarketState(log.marketState) }}
                              </n-tag>
                            </div>
                          </n-gi>
                        </n-grid>
                      </div>
                      
                      <!-- 信号原因 -->
                      <div class="signal-log-reason" v-if="log.reason">
                        <n-divider style="margin: 10px 0;" />
                        <n-text depth="3" style="font-size: 12px; line-height: 1.5;">
                          {{ log.reason }}
                        </n-text>
                      </div>
                    </n-card>
                  </div>
                  <n-empty v-else description="暂无预警记录" size="small" style="padding: 20px 0;" />
                </n-collapse-item>
              </n-collapse>

              <!-- 交易执行日志（按需加载：展开时才请求，节省资源） -->
              <n-collapse
                :default-expanded-names="[]"
                :expanded-names="executionExpandedNames[robot.id] ?? executionExpandedEmpty"
                @update:expanded-names="(names) => onExecutionExpanded(robot.id, names)"
                style="margin-top: 6px"
              >
                <n-collapse-item name="execution-logs">
                  <template #header>
                    <n-space align="center" :size="8">
                      <span style="font-size: 12px; font-weight: 500">订单日志 ({{ executionLogs[robot.id]?.length || 0 }})</span>
                    </n-space>
                  </template>
                  <!-- header-extra 的“刷新”按钮会造成理解成本，且现在已自动预取/展开自动刷新，这里移除 -->
                  <n-space v-if="executionLogs[robot.id]?.length" align="center" justify="space-between" style="margin-bottom: 8px">
                    <n-space align="center" :size="10">
                      <n-switch v-model:value="executionOnlyFailed[robot.id]" size="small" />
                      <n-text depth="3" style="font-size: 12px">只看失败</n-text>
                    </n-space>
                    <n-text depth="3" style="font-size: 11px">
                      {{ formatUpdateTime(executionLastLoadedAt[robot.id]) }}
                    </n-text>
                  </n-space>
                  <div class="execution-logs-list" v-if="getExecutionLogsForRobot(robot.id).length > 0">
                    <n-card
                      v-for="(log, idx) in getExecutionLogsForRobot(robot.id)"
                      :key="idx"
                      :bordered="false"
                      size="small"
                      class="execution-log-card"
                      :class="log.status"
                      content-style="padding: 10px 12px;"
                    >
                      <!-- 日志头部 -->
                      <div class="execution-log-header">
                        <n-space align="center" :size="8">
                          <n-text depth="3" style="font-size: 11px; font-family: 'JetBrains Mono', monospace;">
                            {{ formatLogTime(log.createdAt) }}
                          </n-text>
                          <n-tag :type="getEventTypeTag(log.eventType)" size="small" :bordered="false">
                            {{ getEventTypeText(log.eventType) }}
                          </n-tag>
                          <n-tag v-if="getLogStep(log)" size="small" :bordered="false" type="default">
                            {{ getLogStep(log) }}
                          </n-tag>
                          <n-tag 
                            :type="log.status === 'success' ? 'success' : log.status === 'failed' ? 'error' : 'warning'" 
                            size="small"
                            :bordered="false"
                          >
                            {{ log.status === 'success' ? '成功' : log.status === 'failed' ? '失败' : '进行中' }}
                          </n-tag>
                        </n-space>
                      </div>
                      
                      <!-- 日志消息 -->
                      <div class="execution-log-message" v-if="log.message">
                        <n-text depth="3" style="font-size: 12px; line-height: 1.5;">
                          {{ log.message }}
                        </n-text>
                      </div>

                      <!-- 失败原因（优先展示结构化原因） -->
                      <div class="execution-log-message" v-if="getFailureReason(log)">
                        <n-text type="error" style="font-size: 12px; line-height: 1.5;">
                          {{ getFailureReason(log) }}
                        </n-text>
                      </div>
                      
                      <!-- 提交下单参数 -->
                      <div v-if="(log.eventType === 'order_attempt' || log.eventType === 'order_submit') && getLogSubmitParams(log)" class="execution-log-details">
                        <n-divider style="margin: 10px 0;" />
                        <n-grid :cols="4" :x-gap="12" :y-gap="8">
                          <n-gi v-if="getLogSubmitParams(log).symbol">
                            <div class="detail-item">
                              <n-text depth="3" style="font-size: 11px;">交易对</n-text>
                              <n-text strong style="font-size: 12px; margin-left: 6px;">{{ getLogSubmitParams(log).symbol }}</n-text>
                            </div>
                          </n-gi>
                          <n-gi v-if="getLogSubmitParams(log).side">
                            <div class="detail-item">
                              <n-text depth="3" style="font-size: 11px;">方向</n-text>
                              <n-tag 
                                :type="getLogSubmitParams(log).side === 'BUY' ? 'success' : 'error'" 
                                size="small"
                                :bordered="false"
                                style="margin-left: 6px;"
                              >
                                {{ getLogSubmitParams(log).side === 'BUY' ? '买入' : '卖出' }}
                              </n-tag>
                            </div>
                          </n-gi>
                          <n-gi v-if="getLogSubmitParams(log).position_side">
                            <div class="detail-item">
                              <n-text depth="3" style="font-size: 11px;">持仓</n-text>
                              <n-tag 
                                :type="getLogSubmitParams(log).position_side === 'LONG' ? 'success' : 'error'" 
                                size="small"
                                :bordered="false"
                                style="margin-left: 6px;"
                              >
                                {{ getLogSubmitParams(log).position_side === 'LONG' ? '多' : '空' }}
                              </n-tag>
                            </div>
                          </n-gi>
                          <n-gi v-if="getLogSubmitParams(log).quantity">
                            <div class="detail-item">
                              <n-text depth="3" style="font-size: 11px;">数量</n-text>
                              <n-text strong style="font-size: 12px; font-family: 'JetBrains Mono', monospace; margin-left: 6px;">
                                {{ getLogSubmitParams(log).quantity?.toFixed(4) }}
                              </n-text>
                            </div>
                          </n-gi>
                          <n-gi v-if="getLogSubmitParams(log).entry_price">
                            <div class="detail-item">
                              <n-text depth="3" style="font-size: 11px;">价格</n-text>
                              <n-text strong style="font-size: 12px; font-family: 'JetBrains Mono', monospace; margin-left: 6px;">
                                {{ getLogSubmitParams(log).entry_price?.toFixed(2) }}
                              </n-text>
                            </div>
                          </n-gi>
                          <n-gi v-if="getLogSubmitParams(log).leverage">
                            <div class="detail-item">
                              <n-text depth="3" style="font-size: 11px;">杠杆</n-text>
                              <n-text strong style="font-size: 12px; margin-left: 6px;">
                                {{ getLogSubmitParams(log).leverage }}x
                              </n-text>
                            </div>
                          </n-gi>
                          <n-gi v-if="getLogSubmitParams(log).margin">
                            <div class="detail-item">
                              <n-text depth="3" style="font-size: 11px;">保证金</n-text>
                              <n-text strong style="font-size: 12px; font-family: 'JetBrains Mono', monospace; margin-left: 6px;">
                                {{ getLogSubmitParams(log).margin?.toFixed(2) }} USDT
                              </n-text>
                            </div>
                          </n-gi>
                          <n-gi v-if="getLogSubmitParams(log).margin_percent">
                            <div class="detail-item">
                              <n-text depth="3" style="font-size: 11px;">保证金比例</n-text>
                              <n-text strong style="font-size: 12px; margin-left: 6px;">
                                {{ getLogSubmitParams(log).margin_percent }}%
                              </n-text>
                            </div>
                          </n-gi>
                        </n-grid>
                      </div>
                      
                      <!-- 订单成功策略参数 -->
                      <div v-if="log.eventType === 'order_success' && getLogStrategyParams(log)" class="execution-log-details">
                        <n-divider style="margin: 10px 0;" />
                        <n-grid :cols="4" :x-gap="12" :y-gap="8">
                          <n-gi v-if="getLogStrategyParams(log).marketState">
                            <div class="detail-item">
                              <n-text depth="3" style="font-size: 11px;">市场状态</n-text>
                              <n-tag 
                                :type="getMarketStateType(getLogStrategyParams(log).marketState)" 
                                size="small"
                                :bordered="false"
                                style="margin-left: 6px;"
                              >
                                {{ formatMarketState(getLogStrategyParams(log).marketState) }}
                              </n-tag>
                            </div>
                          </n-gi>
                          <n-gi v-if="getLogStrategyParams(log).riskPreference">
                            <div class="detail-item">
                              <n-text depth="3" style="font-size: 11px;">风险偏好</n-text>
                              <n-tag 
                                type="warning" 
                                size="small"
                                :bordered="false"
                                style="margin-left: 6px;"
                              >
                                {{ formatRiskPref(getLogStrategyParams(log).riskPreference) }}
                              </n-tag>
                            </div>
                          </n-gi>
                          <n-gi v-if="getLogStrategyParams(log).leverage">
                            <div class="detail-item">
                              <n-text depth="3" style="font-size: 11px;">杠杆</n-text>
                              <n-text strong style="font-size: 12px; margin-left: 6px;">
                                {{ getLogStrategyParams(log).leverage }}x
                              </n-text>
                            </div>
                          </n-gi>
                          <n-gi v-if="getLogStrategyParams(log).marginPercent">
                            <div class="detail-item">
                              <n-text depth="3" style="font-size: 11px;">保证金</n-text>
                              <n-text strong style="font-size: 12px; margin-left: 6px;">
                                {{ getLogStrategyParams(log).marginPercent }}%
                              </n-text>
                            </div>
                          </n-gi>
                          <n-gi v-if="getLogStrategyParams(log).stopLossPercent">
                            <div class="detail-item">
                              <n-text depth="3" style="font-size: 11px;">止损</n-text>
                              <n-text strong style="font-size: 12px; color: #d03050; margin-left: 6px;">
                                {{ getLogStrategyParams(log).stopLossPercent }}%
                              </n-text>
                            </div>
                          </n-gi>
                          <n-gi v-if="getLogStrategyParams(log).autoStartRetreat">
                            <div class="detail-item">
                              <n-text depth="3" style="font-size: 11px;">启动止盈</n-text>
                              <n-text strong style="font-size: 12px; color: #18a058; margin-left: 6px;">
                                {{ getLogStrategyParams(log).autoStartRetreat }}%
                              </n-text>
                            </div>
                          </n-gi>
                          <n-gi v-if="getLogStrategyParams(log).profitRetreatPercent">
                            <div class="detail-item">
                              <n-text depth="3" style="font-size: 11px;">止盈回撤</n-text>
                              <n-text strong style="font-size: 12px; color: #18a058; margin-left: 6px;">
                                {{ getLogStrategyParams(log).profitRetreatPercent }}%
                              </n-text>
                            </div>
                          </n-gi>
                        </n-grid>
                      </div>
                    </n-card>
                  </div>
                  <!-- 原始有日志，但筛选后为空（例如开启“只看失败”但当前无失败记录） -->
                  <n-empty
                    v-else-if="executionLogs[robot.id]?.length > 0 && executionOnlyFailed[robot.id]"
                    description="暂无失败订单日志"
                    size="small"
                    style="padding: 20px 0;"
                  />
                  <n-empty v-else description="暂无订单日志" size="small" style="padding: 20px 0;" />
                </n-collapse-item>
              </n-collapse>
            </div>

            <!-- 加载分析数据 -->
            <div v-else class="analysis-loading">
              <n-spin size="small" />
              <n-text depth="3" style="font-size: 12px; margin-left: 8px">正在分析市场...</n-text>
            </div>

            <!-- 实时持仓订单列表 -->
            <div class="positions-section">
              <div class="positions-header">
                <n-space align="center" :size="8">
                <h4 class="positions-title">持仓</h4>
                <span class="positions-count">{{ positionData[robot.id]?.length || 0 }}</span>
                </n-space>
                <n-space align="center" :size="12">
                  <n-space align="center" :size="4">
                    <span style="font: var(--font-small); color: var(--text-color-3); opacity: 0.75;">自动下单</span>
                    <n-switch 
                      :value="analysisData[robot.id]?.config?.autoTradeEnabled || false" 
                      @update:value="(val) => toggleAutoTrade(robot, val)"
                      size="small"
                      class="tiny-switch"
                    />
                  </n-space>
                  <n-space align="center" :size="4">
                    <span style="font: var(--font-small); color: var(--text-color-3); opacity: 0.75;">双向开单</span>
                    <n-switch 
                      :value="analysisData[robot.id]?.config?.dualSidePosition !== false" 
                      @update:value="(val) => toggleDualSidePosition(robot, val)"
                      size="small"
                      class="tiny-switch"
                    />
                  </n-space>
                  <n-space align="center" :size="4">
                    <span style="font: var(--font-small); color: var(--text-color-3); opacity: 0.75;">自动平仓</span>
                    <n-switch 
                      :value="analysisData[robot.id]?.config?.autoCloseEnabled || false" 
                      @update:value="(val) => toggleAutoClose(robot, val)"
                      size="small"
                      class="tiny-switch"
                    />
                  </n-space>
                  <n-space align="center" :size="4">
                    <span style="font: var(--font-small); color: var(--text-color-3); opacity: 0.75;">锁定盈利</span>
                    <n-switch
                      :value="analysisData[robot.id]?.config?.profitLockEnabled !== false"
                      @update:value="(val) => toggleProfitLock(robot, val)"
                      size="small"
                      class="tiny-switch"
                    />
                  </n-space>
                </n-space>
              </div>
              <div class="positions-table-wrapper">
                <table class="positions-table" v-if="positionData[robot.id]?.length > 0">
                  <thead>
                    <tr>
                      <th class="col-info">交易信息</th>
                      <th class="col-quantity">保证金/数量</th>
                      <th class="col-price">开仓价格</th>
                      <th class="col-price">市价</th>
                      <th class="col-pl">未实现盈亏</th>
                      <th class="col-monitor">监控</th>
                      <th class="col-action">操作</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="(pos, idx) in positionData[robot.id]" :key="pos.symbol + pos.positionSide" 
                        :class="['position-row', idx % 2 === 0 ? 'row-even' : 'row-odd']">
                      <td class="col-info">
                        <div class="info-cell">
                          <div class="info-row-second">
                            <span :class="['side-tag-mini', pos.positionSide === 'LONG' ? 'long' : 'short']">
                              {{ pos.positionSide === 'LONG' ? '多' : '空' }}
                            </span>
                          </div>
                          <div class="info-row-middle">
                            <span :class="['margin-mode-tag', pos.marginType === 'crossed' ? 'crossed' : 'isolated']">
                              {{ pos.marginType === 'crossed' ? '全仓' : '逐仓' }}
                            </span>
                            <span class="leverage-tag">
                              杠杆 {{ (pos.leverage && Number(pos.leverage) > 0) ? (Number(pos.leverage) + 'x') : '--' }}
                            </span>
                          </div>
                          <div class="info-row-order" v-if="pos.orderId || pos.clientOrderId">
                            <div class="order-info-item" v-if="pos.orderId">
                              <span class="order-value">{{ pos.orderId }}</span>
                            </div>
                            <div class="order-info-item" v-if="pos.clientOrderId">
                              <span class="order-value">{{ pos.clientOrderId }}</span>
                            </div>
                            <div class="order-info-item" v-if="pos.orderType">
                              <span class="order-value">{{ pos.orderType === 'MARKET' ? '市价' : pos.orderType === 'LIMIT' ? '限价' : pos.orderType }}</span>
                            </div>
                            <div class="order-info-item" v-if="pos.orderSide">
                              <span class="order-value">{{ pos.orderSide === 'BUY' ? '买入' : pos.orderSide === 'SELL' ? '卖出' : pos.orderSide }}</span>
                            </div>
                            <div class="order-info-item" v-if="pos.orderCreateTime">
                              <span class="order-value">{{ formatTime(pos.orderCreateTime) }}</span>
                            </div>
                          </div>
                        </div>
                      </td>
                      <td class="col-quantity">
                        <div class="qty-cell">
                          <span class="margin-in-qty">
                            {{
                              ((pos.margin && pos.margin > 0 ? pos.margin : pos.isolatedMargin) || 0) > 0
                                ? (pos.margin && pos.margin > 0 ? pos.margin : pos.isolatedMargin).toFixed(2)
                                : '--'
                            }}
                          </span>
                          <span class="quantity-value">{{ Math.abs(pos.positionAmt).toFixed(4) }}</span>
                        </div>
                      </td>
                      <td class="col-price">
                        <span class="price-value">{{ pos.entryPrice?.toFixed(2) || '--' }}</span>
                      </td>
                      <td class="col-price">
                        <span class="price-value market-price">{{ pos.markPrice?.toFixed(2) || analysisData[robot.id]?.analysis?.ticker?.lastPrice?.toFixed(2) || '--' }}</span>
                      </td>
                      <td class="col-pl">
                        <div class="pnl-display" :class="pos.unrealizedPnl >= 0 ? 'profit' : 'loss'">
                          <span class="pnl-icon">{{ pos.unrealizedPnl >= 0 ? '📈' : '📉' }}</span>
                          <span class="pnl-value">
                            {{ pos.unrealizedPnl >= 0 ? '+' : '' }}{{ pos.unrealizedPnl?.toFixed(4) || '0.0000' }}
                          </span>
                        </div>
                      </td>
                      <td class="col-monitor">
                        <div class="monitor-cell">
                          <!-- ①止损进度：|未实现盈亏| / (保证金 × 止损%) 达到100%时平仓 -->
                          <div class="monitor-item">
                            <div class="monitor-label" :title="(() => {
                              const unreal = Math.abs(Number(pos.unrealizedPnl || 0));
                              const slp = ((pos.stopLossPercent ?? null) !== null && Number(pos.stopLossPercent) > 0)
                                ? Number(pos.stopLossPercent)
                                : getRobotStopLossPercent(robot.id);
                              let m = (pos.margin && pos.margin > 0 ? pos.margin : pos.isolatedMargin) || 0;
                              if (m <= 0) {
                                const qty = Math.abs(Number(pos.positionAmt ?? pos.position_amount ?? pos.quantity ?? 0));
                                const entry = Number(pos.entryPrice ?? pos.entry_price ?? 0);
                                const lev = Number(pos.leverage ?? getRobotLeverage(robot.id) ?? 0);
                                if (qty > 0 && entry > 0 && lev > 0) m = (qty * entry) / lev;
                              }
                              const threshold = (slp && slp > 0 && m > 0) ? (m * (slp / 100)).toFixed(2) : '--';
                              return '当前亏损: ' + unreal.toFixed(2) + ' USDT / 止损阈值: ' + threshold + ' USDT';
                            })()">
                              止损
                            </div>
                            <div class="progress-bar-container">
                              <div class="progress-bar progress-bar-danger" 
                                   :style="{ 
                                     width: calcStopLossProgress(pos, robot).toFixed(1) + '%',
                                     backgroundColor: calcStopLossProgress(pos, robot) >= 80 ? '#ef4444' : '#f59e0b'
                                   }"></div>
                            </div>
                            <div class="monitor-value" :class="{
                              'text-danger': calcStopLossProgress(pos, robot) >= 100,
                              'text-warning': calcStopLossProgress(pos, robot) >= 80 && calcStopLossProgress(pos, robot) < 100
                            }">
                              <span>{{ calcStopLossProgress(pos, robot) >= 100 ? '⚠️100%' : calcStopLossProgress(pos, robot).toFixed(1) + '%' }}</span>
                              <span style="color: #9ca3af; font: var(--font-tiny); margin-left: 2px;">/{{ (pos.stopLossPercent ?? null) !== null && Number(pos.stopLossPercent) > 0 ? Number(pos.stopLossPercent).toFixed(1) + '%' : (getRobotStopLossPercent(robot.id) > 0 ? getRobotStopLossPercent(robot.id).toFixed(1) + '%' : '--') }}</span>
                            </div>
                          </div>
                          <!-- ②止盈：(实时最高盈利金额 - 实时未实现盈亏) / 最高盈利金额 >= 设定的百分比时平仓 -->
                          <!-- 默认100%绿色满条，启动后从100%往回撤 -->
                          <div class="monitor-item">
                            <div class="monitor-label" :title="getTakeProfitRetreatSwitch(robot.id, pos.symbol, pos.positionSide, pos) ? ('回撤: ' + (((pos.maxProfitReached || 0) - (pos.unrealizedPnl || 0)) / (pos.maxProfitReached || 1) * 100).toFixed(1) + '% / 设定: ' + (((pos.profitRetreatPercent ?? null) !== null && Number(pos.profitRetreatPercent) > 0) ? Number(pos.profitRetreatPercent) + '%' : (getRobotProfitRetreat(robot.id) !== null && getRobotProfitRetreat(robot.id) > 0 ? getRobotProfitRetreat(robot.id) + '%' : '--'))) : '未启动止盈，默认100%'">
                              回撤止盈
                            </div>
                            <div class="progress-bar-container">
                              <div class="progress-bar progress-bar-success" 
                                   :style="{ 
                                     width: calcProfitRetreatProgress(pos, robot).toFixed(1) + '%',
                                     backgroundColor: calcProfitRetreatProgress(pos, robot) <= 20 ? '#ef4444' : '#22c55e',
                                     opacity: 0.7
                                   }"></div>
                            </div>
                            <div class="monitor-value" style="font-weight: normal;">
                              <template v-if="!getTakeProfitRetreatSwitch(robot.id, pos.symbol, pos.positionSide, pos)">
                                <span style="color: #22c55e;">100%</span>
                                <span style="color: #9ca3af; font: var(--font-tiny); margin-left: 2px;">/{{ (pos.profitRetreatPercent ?? null) !== null && Number(pos.profitRetreatPercent) > 0 ? Number(pos.profitRetreatPercent).toFixed(1) + '%' : (getRobotProfitRetreat(robot.id) > 0 ? getRobotProfitRetreat(robot.id).toFixed(1) + '%' : '--') }}</span>
                              </template>
                              <template v-else>
                                <span :style="{ color: calcProfitRetreatProgress(pos, robot) <= 20 ? '#ef4444' : (calcProfitRetreatProgress(pos, robot) <= 50 ? '#f59e0b' : '#22c55e') }">{{ calcProfitRetreatProgress(pos, robot) <= 0 ? '⚠️0%' : calcProfitRetreatProgress(pos, robot).toFixed(1) + '%' }}</span>
                                <span style="color: #9ca3af; font: var(--font-tiny); margin-left: 2px;">/{{ (pos.profitRetreatPercent ?? null) !== null && Number(pos.profitRetreatPercent) > 0 ? Number(pos.profitRetreatPercent).toFixed(1) + '%' : (getRobotProfitRetreat(robot.id) > 0 ? getRobotProfitRetreat(robot.id).toFixed(1) + '%' : '--') }}</span>
                              </template>
                            </div>
                          </div>
                          <!-- ③启动止盈：未实现盈亏/保证金 达到设定%时开启止盈 -->
                          <!-- 说明：一旦“止盈开关”已启动（不可关闭原则），这里不再随盈利回落显示“回退血条”，避免误导 -->
                          <div v-if="!getTakeProfitRetreatSwitch(robot.id, pos.symbol, pos.positionSide, pos)" class="monitor-item">
                            <div class="monitor-label" :title="'盈利: ' + ((pos.unrealizedPnl || 0) / (pos.margin || 1) * 100).toFixed(1) + '% / 启动: ' + (((pos.autoStartRetreatPercent ?? null) !== null && Number(pos.autoStartRetreatPercent) > 0) ? Number(pos.autoStartRetreatPercent) + '%' : (getRobotAutoStartRetreat(robot.id) !== null && getRobotAutoStartRetreat(robot.id) > 0 ? getRobotAutoStartRetreat(robot.id) + '%' : '--'))">
                              启动止盈
                            </div>
                            <div class="progress-bar-container">
                              <div class="progress-bar progress-bar-warning" 
                                   :style="{ 
                                     width: calcStartProfitProgress(pos, robot).toFixed(1) + '%',
                                     backgroundColor: calcStartProfitProgress(pos, robot) >= 100 ? '#22c55e' : '#f59e0b'
                                   }"></div>
                            </div>
                            <div class="monitor-value" :class="{ 
                              'text-success': calcStartProfitProgress(pos, robot) >= 100,
                              'text-warning': calcStartProfitProgress(pos, robot) >= 80 && calcStartProfitProgress(pos, robot) < 100
                            }">
                              <span>{{ calcStartProfitProgress(pos, robot) >= 100 ? '✓已满足' : calcStartProfitProgress(pos, robot).toFixed(1) + '%' }}</span>
                              <span style="color: #9ca3af; font: var(--font-tiny); margin-left: 2px;">/{{ (pos.autoStartRetreatPercent ?? null) !== null && Number(pos.autoStartRetreatPercent) > 0 ? Number(pos.autoStartRetreatPercent).toFixed(1) + '%' : (getRobotAutoStartRetreat(robot.id) > 0 ? getRobotAutoStartRetreat(robot.id).toFixed(1) + '%' : '--') }}</span>
                            </div>
                          </div>
                          <div v-else class="monitor-item">
                            <div class="monitor-label" title="止盈开关已启动（不可关闭），启动止盈条件不再回退显示">
                              启动止盈
                            </div>
                            <div class="progress-bar-container">
                              <div class="progress-bar" :style="{ width: '100%', backgroundColor: '#22c55e' }"></div>
                            </div>
                            <div class="monitor-value text-success">
                              <span>✓已启动</span>
                            </div>
                          </div>
                          <!-- 启动止盈开关（不可关闭原则：启用后直到平仓前不允许关闭） -->
                          <div class="monitor-item monitor-switch">
                            <div class="monitor-label" style="width: 60px; min-width: 60px; font: var(--font-tiny);">启动止盈开关</div>
                            <div class="monitor-value" style="flex: 1; display: flex; align-items: center; justify-content: flex-end;">
                              <n-switch 
                                :value="getTakeProfitRetreatSwitch(robot.id, pos.symbol, pos.positionSide, pos)"
                                :disabled="true"
                                size="small"
                                style="--n-switch-width: 32px; --n-switch-height: 16px;"
                              />
                            </div>
                          </div>
                          <!-- 最高盈利 -->
                          <div class="monitor-item monitor-max-profit">
                            <div class="monitor-label" style="width: 52px; min-width: 52px; font: var(--font-small);">最高盈利</div>
                            <div class="monitor-value" style="font: var(--font-small); font-weight: 500; flex: 1; text-align: right;">
                              <span v-if="pos.maxProfitReached > 0" style="color: #22c55e; font-weight: 600;">
                                {{ pos.maxProfitReached.toFixed(4) }}
                              </span>
                              <span v-else style="color: var(--text-color-3);">--</span>
                              <span v-if="pos.maxProfitReached > 0" style="font: var(--font-tiny);"> USDT</span>
                            </div>
                          </div>
                        </div>
                      </td>
                      <td class="col-action">
                        <n-button 
                          size="small" 
                          @click="closePosition(robot, pos)"
                          class="close-position-btn"
                        >
                          平仓
                        </n-button>
                      </td>
                    </tr>
                  </tbody>
                </table>
                <div v-else class="empty-state">
                  <div class="empty-icon">📊</div>
                  <div class="empty-text">暂无持仓订单</div>
                  <div class="empty-hint">等待开仓信号...</div>
                </div>
              </div>
            </div>
          </template>

          <!-- 非运行状态显示 -->
          <template v-else>
            <!-- 未启动提示卡片 -->
            <n-card :bordered="false" size="small" style="margin-bottom: 12px; background: linear-gradient(135deg, rgba(24, 160, 88, 0.08) 0%, rgba(24, 160, 88, 0.03) 100%); border: 1px solid rgba(24, 160, 88, 0.2);">
              <n-space align="center" :size="16">
                <div style="width: 48px; height: 48px; border-radius: 50%; background: rgba(24, 160, 88, 0.1); display: flex; align-items: center; justify-content: center; flex-shrink: 0;">
                  <n-icon :component="PlayCircleOutlined" :size="28" style="color: var(--primary-color);" />
                </div>
                <div style="flex: 1; min-width: 0;">
                  <n-text strong style="font-size: 15px; color: var(--text-color-1); display: block;">
                    机器人未启动
                  </n-text>
                  <n-space :size="16" style="margin-top: 8px; flex-wrap: wrap;">
                    <!-- 交易所和交易对 -->
                    <n-space :size="4" align="center">
                      <n-text depth="3" style="font-size: 12px;">交易所：</n-text>
                      <n-tag size="small" type="info" :bordered="false">
                        {{ robot.exchange ? robot.exchange.toUpperCase() : '--' }}
                      </n-tag>
                    </n-space>
                    <n-space :size="4" align="center">
                      <n-text depth="3" style="font-size: 12px;">交易对：</n-text>
                      <n-tag size="small" type="warning" :bordered="false">
                        {{ robot.symbol || '--' }}
                      </n-tag>
                    </n-space>
                    <!-- 策略组信息 -->
                    <n-space v-if="getStrategyGroupName(robot)" :size="4" align="center">
                      <n-text depth="3" style="font-size: 12px;">策略组：</n-text>
                      <n-tag size="small" type="success" :bordered="false">
                        {{ getStrategyGroupName(robot) }}
                      </n-tag>
                    </n-space>
                    <!-- 创建时间 -->
                    <n-space :size="4" align="center">
                      <n-text depth="3" style="font-size: 12px;">创建时间：</n-text>
                      <n-text depth="3" style="font-size: 12px;">{{ robot.createdAt || '--' }}</n-text>
                    </n-space>
                  </n-space>
                </div>
              </n-space>
            </n-card>

            <!-- 错误提示（如果有错误，必须显示在最显眼的位置） -->
            <n-alert v-if="getConfigError(robot.id)" type="error" size="small" style="margin-bottom: 12px" :show-icon="true">
              <template #header>
                <n-space align="center" :size="4">
                  <n-icon :component="WarningOutlined" />
                  <span style="font-weight: 600;">配置错误</span>
                </n-space>
              </template>
              {{ getConfigError(robot.id) }}
            </n-alert>
          </template>

          <!-- 操作按钮 -->
          <template #action>
            <n-space justify="space-between" style="width: 100%">
              <n-space align="center">
                <n-button v-if="robot.status === 2" type="warning" size="small" @click="stopRobot(robot)">
                  <template #icon><n-icon :component="PauseCircleOutlined" /></template>
                  停止运行
                </n-button>
                <n-button v-else-if="robot.status === 1" type="primary" size="small" @click="startRobot(robot)">
                  <template #icon><n-icon :component="PlayCircleOutlined" /></template>
                  启动运行
                </n-button>
                <n-button v-else-if="robot.status === 3" type="primary" size="small" @click="startRobot(robot)">
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
                <!-- 订阅到期倒计时（仅显示提醒） -->
                <n-tag v-if="robot.status === 2 && subscriptionInfo.hasSubscription && subscriptionInfo.planExpireTime" 
                       :type="getSubscriptionCountdownType()" 
                       size="small">
                  订阅到期: {{ getSubscriptionCountdown() }}
                </n-tag>
              </n-space>
              <n-space>
                <n-button size="small" @click="viewDetail(robot)">查看详情</n-button>
                <n-button type="error" ghost size="small" @click="deleteRobot(robot)" :disabled="robot.status === 2">删除</n-button>
              </n-space>
            </n-space>
          </template>
        </n-card>
        </n-gi>
      </n-grid>
    </template>

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

    <!-- 详情弹窗 -->
    <n-modal v-model:show="showDetailModal" title="机器人详情" preset="card" style="width: 960px">
      <n-tabs v-if="currentRobot" type="segment" animated>
        <n-tab-pane name="info" tab="基本信息">
          <!-- 基础信息 -->
          <n-descriptions :column="3" label-placement="left" bordered size="small" class="mb-3">
            <n-descriptions-item label="机器人名称">
              <n-text strong>{{ currentRobot.robotName }}</n-text>
            </n-descriptions-item>
            <n-descriptions-item label="状态">
              <n-tag :type="getStatusType(currentRobot.status)" size="small">{{ getStatusText(currentRobot.status) }}</n-tag>
            </n-descriptions-item>
            <n-descriptions-item label="创建时间">{{ currentRobot.createdAt }}</n-descriptions-item>
          </n-descriptions>

          <!-- 交易配置 -->
          <n-card title="交易配置" size="small" :bordered="false" class="mb-3">
            <n-grid :cols="5" :x-gap="12">
              <n-gi>
                <n-statistic label="交易平台">
                  <template #default>
                    <n-tag type="info" size="small">{{ (currentRobot.platform || currentRobot.exchange || '-').toUpperCase() }}</n-tag>
                  </template>
                </n-statistic>
              </n-gi>
              <n-gi>
                <n-statistic label="交易对">
                  <template #default>
                    <n-text type="warning" strong>{{ currentRobot.tradingPair || currentRobot.symbol || '-' }}</n-text>
                  </template>
                </n-statistic>
              </n-gi>
              <n-gi>
                <n-statistic label="交易类型" :value="currentRobot.tradeType === 'perpetual' || !currentRobot.tradeType ? '永续合约' : currentRobot.tradeType" />
              </n-gi>
              <n-gi>
                <n-statistic label="订单类型" :value="currentRobot.orderType === 'market' || !currentRobot.orderType ? '市价单' : '限价单'" />
              </n-gi>
              <n-gi>
                <n-statistic label="保证金模式" :value="currentRobot.marginMode === 'isolated' || !currentRobot.marginMode ? '逐仓' : '全仓'" />
              </n-gi>
            </n-grid>
          </n-card>

          <!-- 策略参数（重新设计：统一显示逻辑） -->
          <n-card title="策略参数" size="small" :bordered="false" class="mb-3">
            <!-- 错误提示 -->
            <n-alert v-if="currentRobot && getConfigError(currentRobot.id)" type="error" size="small" style="margin-bottom: 12px" :show-icon="true">
              {{ getConfigError(currentRobot.id) }}
              </n-alert>
            
            <!-- 市场状态和风险偏好 -->
            <n-space :size="12" style="margin-bottom: 16px;">
              <n-tag :type="getMarketStateType(getDetailMarketState())" size="medium">
                市场状态: {{ formatMarketState(getDetailMarketState()) || '--' }}
              </n-tag>
              <n-tag type="warning" size="medium">
                风险偏好: {{ formatRiskPref(getDetailRiskPreference()) || '--' }}
              </n-tag>
              <n-tag type="success" size="medium">
                策略组: {{ getDetailStrategyGroupName() || '--' }}
              </n-tag>
              <n-tag type="info" size="medium">
                策略: {{ getDetailStrategyName() || '--' }}
              </n-tag>
            </n-space>
            
            <!-- 详细的策略参数 -->
            <n-grid :cols="4" :x-gap="12" :y-gap="12">
                <n-gi>
                <n-statistic label="时间窗口">
                    <template #default>
                    <n-text type="info" strong>{{ getDetailTimeWindow() !== null ? formatWindowTime(getDetailTimeWindow()) : '--' }}</n-text>
                    </template>
                  </n-statistic>
                </n-gi>
                <n-gi>
                <n-statistic label="波动值">
                    <template #default>
                    <n-text type="info" strong>{{ getDetailThreshold() !== null ? `${getDetailThreshold().toFixed(1)} USDT` : '--' }}</n-text>
                    </template>
                  </n-statistic>
                </n-gi>
                <n-gi>
                  <n-statistic label="杠杆倍数">
                    <template #default>
                      <n-text type="primary" strong>{{ getDetailLeverage() !== null ? `${getDetailLeverage()}x` : '--' }}</n-text>
                    </template>
                  </n-statistic>
                </n-gi>
                <n-gi>
                <n-statistic label="保证金比例">
                  <template #default>
                    <n-text type="primary" strong>{{ getDetailMarginPercent() !== null ? `${getDetailMarginPercent()}%` : '--' }}</n-text>
                  </template>
                </n-statistic>
                </n-gi>
                <n-gi>
                  <n-statistic label="止损比例">
                    <template #default>
                    <n-text type="error" strong>{{ getDetailStopLossPercent() !== null ? `${getDetailStopLossPercent()}%` : '--' }}</n-text>
                  </template>
                </n-statistic>
              </n-gi>
                <n-gi>
                  <n-statistic label="启动止盈">
                    <template #default>
                    <n-text type="success" strong>{{ getDetailAutoStartRetreat() !== null ? `${getDetailAutoStartRetreat()}%` : '--' }}</n-text>
                    </template>
                  </n-statistic>
                </n-gi>
                <n-gi>
                  <n-statistic label="止盈回撤">
                    <template #default>
                    <n-text type="success" strong>{{ getDetailProfitRetreat() !== null ? `${getDetailProfitRetreat()}%` : '--' }}</n-text>
                    </template>
                  </n-statistic>
                </n-gi>
                <n-gi>
                <n-statistic label="最大盈利目标">
                    <template #default>
                    <n-text type="success">{{ currentRobot.maxProfitTarget || 0 }} USDT</n-text>
                    </template>
                  </n-statistic>
                </n-gi>
                <n-gi>
                <n-statistic label="最大亏损限制">
                    <template #default>
                    <n-text type="error">{{ currentRobot.maxLossAmount || 0 }} USDT</n-text>
                    </template>
                  </n-statistic>
                </n-gi>
                <n-gi>
                  <n-statistic label="总盈亏">
                    <template #default>
                      <n-text :type="(currentRobot.totalPnl || 0) >= 0 ? 'success' : 'error'" strong>
                        {{ (currentRobot.totalPnl || 0) >= 0 ? '+' : '' }}{{ (currentRobot.totalPnl || 0).toFixed(4) }} USDT
                      </n-text>
                    </template>
                  </n-statistic>
                </n-gi>
                <n-gi>
                  <n-statistic label="消耗算力" :value="(currentRobot.consumedPower || 0).toFixed(4)" />
                </n-gi>
              </n-grid>
          </n-card>

          <!-- 策略配置 -->
          <n-card size="small" :bordered="false">
            <template #header>
              <n-space align="center">
                <span>当前策略配置</span>
                <n-tag v-if="currentRobotStrategy?.groupName" type="info" size="small">
                  {{ currentRobotStrategy.groupName }}
                </n-tag>
                <n-tag v-else type="default" size="small">默认策略</n-tag>
                <n-button type="primary" size="tiny" @click="openStrategySelector" :disabled="currentRobot.status === 2">
                  切换策略模板
                </n-button>
              </n-space>
            </template>
            
            <!-- 策略模板信息 -->
            <n-alert v-if="currentRobotStrategy" type="info" size="small" style="margin-bottom: 12px" :show-icon="false">
              <template #header>
                <n-space align="center" :size="8">
                  <n-icon :component="SettingOutlined" />
                  <span>策略模板：{{ currentRobotStrategy.groupName || '默认策略' }}</span>
                  <n-tag size="tiny" type="warning" v-if="currentRobotStrategy.isOfficial">官方</n-tag>
                </n-space>
              </template>
            </n-alert>
            
            <!-- 策略组信息和映射关系 -->
              <n-space vertical :size="10">
                <!-- 策略组信息 -->
                <div v-if="getStrategyGroupName(currentRobot)" style="padding: 10px; background: rgba(24, 160, 88, 0.05); border-radius: 6px;">
                  <n-text depth="3" style="font-size: 11px; display: block; margin-bottom: 4px;">策略组</n-text>
                  <n-text strong style="font-size: 14px; color: var(--primary-color);">
                    {{ getStrategyGroupName(currentRobot) }}
                  </n-text>
                </div>
                <!-- 市场状态映射 -->
                <div v-if="getMarketRiskMapping(currentRobot)" style="padding: 10px; background: rgba(24, 160, 88, 0.05); border-radius: 6px;">
                <n-text depth="3" style="font-size: 11px; display: block; margin-bottom: 6px;">市场状态映射关系</n-text>
                  <div style="display: flex; flex-wrap: wrap; gap: 6px;">
                    <n-tag v-for="(risk, market) in getMarketRiskMapping(currentRobot)" :key="market" size="small" type="info" :bordered="false">
                      {{ formatMarketState(market) }} → {{ formatRiskPref(risk) }}
                    </n-tag>
                  </div>
                </div>
                <!-- 自动化开关 -->
                <n-space :size="12" :wrap="true">
                  <div class="strategy-tag">
                    <n-text depth="3" style="font-size: 12px">自动市场分析</n-text>
                    <n-tag :type="currentRobot.autoMarketState === 1 ? 'success' : 'default'" size="small">
                      {{ currentRobot.autoMarketState === 1 ? '开启' : '关闭' }}
                    </n-tag>
                  </div>
                  <div class="strategy-tag">
                    <n-text depth="3" style="font-size: 12px">信号监控</n-text>
                    <n-tag :type="currentRobot.useMonitorSignal !== 0 ? 'success' : 'default'" size="small">
                      {{ currentRobot.useMonitorSignal !== 0 ? '开启' : '关闭' }}
                    </n-tag>
                  </div>
                <div class="strategy-tag">
                  <n-text depth="3" style="font-size: 12px">自动下单</n-text>
                  <n-tag :type="analysisData[currentRobot.id]?.config?.autoTradeEnabled ? 'success' : 'default'" size="small">
                    {{ analysisData[currentRobot.id]?.config?.autoTradeEnabled ? '开启' : '关闭' }}
                  </n-tag>
                </div>
                <div class="strategy-tag">
                  <n-text depth="3" style="font-size: 12px">自动平仓</n-text>
                  <n-tag :type="analysisData[currentRobot.id]?.config?.autoCloseEnabled ? 'success' : 'default'" size="small">
                    {{ analysisData[currentRobot.id]?.config?.autoCloseEnabled ? '开启' : '关闭' }}
                  </n-tag>
                </div>
              </n-space>
            </n-space>
          </n-card>
        </n-tab-pane>
        <n-tab-pane name="positions" tab="实时持仓">
          <n-data-table :columns="positionColumns" :data="currentPositions" :loading="positionLoading" size="small" />
        </n-tab-pane>
        <n-tab-pane name="orders" tab="当前挂单">
          <n-data-table :columns="openOrderColumns" :data="currentOpenOrders" :loading="orderLoading" size="small" />
        </n-tab-pane>
        <n-tab-pane name="history" tab="成交明细">
          <n-data-table :columns="historyColumnsSimple" :data="orderHistory" :loading="historyLoading" size="small" />
          <div style="margin-top: 12px; text-align: center;">
            <n-text depth="3" style="font-size: 12px;">
              查看完整历史订单，请前往
              <n-button text type="primary" size="small" @click="goToOrderHistory">钱包 → 历史订单</n-button>
            </n-text>
          </div>
        </n-tab-pane>
        <n-tab-pane name="power" tab="算力消耗">
          <n-data-table :columns="powerColumns" :data="powerConsumeList" size="small" />
        </n-tab-pane>
      </n-tabs>
    </n-modal>

    <!-- 策略模板选择弹窗 -->
    <n-modal v-model:show="showStrategySelector" preset="card" title="切换策略模板" style="width: 650px">
      <n-alert type="info" style="margin-bottom: 16px" :show-icon="false">
        为机器人 <strong>{{ currentRobot?.robotName }}</strong> 选择策略模板
      </n-alert>

      <n-spin :show="loadingStrategyGroups">
        <!-- 我的策略模板 -->
        <n-card v-if="myStrategyGroups.length > 0" title="我的策略模板" size="small" class="mb-3" :bordered="false">
          <n-radio-group v-model:value="selectedGroupId" name="myGroups">
            <n-space vertical :size="8">
              <n-card 
                v-for="group in myStrategyGroups" 
                :key="group.id"
                size="small"
                hoverable 
                :class="{ 'selected-strategy': selectedGroupId === group.id }"
                @click="selectedGroupId = group.id"
              >
                <n-radio :value="group.id" style="width: 100%">
                  <n-space align="center">
                    <n-text strong>{{ group.groupName }}</n-text>
                    <n-tag v-if="group.isDefault" type="success" size="tiny">默认</n-tag>
                    <n-tag v-if="group.fromOfficial" type="warning" size="tiny">源自官方</n-tag>
                    <n-text depth="3" style="font-size: 12px">{{ group.symbol }} · {{ group.strategyCount || 12 }}种策略</n-text>
                  </n-space>
                </n-radio>
              </n-card>
            </n-space>
          </n-radio-group>
        </n-card>

        <!-- 官方策略模板 -->
        <n-card v-if="officialStrategyGroups.length > 0" title="官方策略模板" size="small" :bordered="false">
          <template #header-extra>
            <n-text depth="3" style="font-size: 12px">选择后自动添加到我的策略</n-text>
          </template>
          <n-radio-group v-model:value="selectedGroupId" name="officialGroups">
            <n-space vertical :size="8">
              <n-card 
                v-for="group in officialStrategyGroups" 
                :key="'official_' + group.id"
                size="small"
                hoverable 
                :class="{ 'selected-strategy': selectedGroupId === group.id }"
                @click="selectOfficialGroup(group)"
              >
                <n-radio :value="group.id" style="width: 100%">
                  <n-space align="center">
                    <n-tag type="warning" size="tiny">官方</n-tag>
                    <n-text strong>{{ group.groupName }}</n-text>
                    <n-text depth="3" style="font-size: 12px">{{ group.symbol }} · {{ group.strategyCount || 12 }}种策略</n-text>
                  </n-space>
                </n-radio>
              </n-card>
            </n-space>
          </n-radio-group>
        </n-card>

        <n-empty v-if="myStrategyGroups.length === 0 && officialStrategyGroups.length === 0" description="暂无可用策略模板">
          <template #extra>
            <n-button type="primary" @click="goToStrategy">去添加策略模板</n-button>
          </template>
        </n-empty>
      </n-spin>

      <template #footer>
        <n-space justify="space-between" style="width: 100%">
          <n-button quaternary @click="goToStrategy">管理我的策略</n-button>
          <n-space>
            <n-button @click="showStrategySelector = false">取消</n-button>
            <n-button type="primary" @click="applyStrategyGroup" :loading="applyingStrategy" :disabled="!selectedGroupId">
              应用此策略模板
            </n-button>
          </n-space>
        </n-space>
      </template>
    </n-modal>

  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted, h, nextTick } from 'vue';
import { useRouter } from 'vue-router';
import { useMessage, useDialog, NTag, NButton, NSpace, NPopconfirm, NCollapseTransition, NProgress } from 'naive-ui';
import { ToogoRobotApi, ToogoExchangeApi, ToogoStrategyApi, ToogoWalletApi, ToogoSubscriptionApi } from '@/api/toogo';
import { http } from '@/utils/http/axios';
import { addOnMessage, removeOnMessage, sendMsg, WebSocketMessage } from '@/utils/websocket/index';
import { SocketEnum } from '@/enums/socketEnum';
import ScheduleCountdown from './components/ScheduleCountdown.vue';
import {
  RobotOutlined,
  PlusOutlined,
  ReloadOutlined,
  PauseCircleOutlined,
  PlayCircleOutlined,
  ThunderboltOutlined,
  SettingOutlined,
  BarChartOutlined,
  WarningOutlined,
  InfoCircleOutlined,
} from '@vicons/antd';
const router = useRouter();
const message = useMessage();
const dialog = useDialog();

// ===== 机器人列表视图（卡片/列表）=====
type RobotListViewMode = 'card' | 'table';
const ROBOT_LIST_VIEW_MODE_KEY = 'toogo_robot_list_view_mode_v1';
const robotListViewMode = ref<RobotListViewMode>((() => {
  try {
    const v = window?.localStorage?.getItem(ROBOT_LIST_VIEW_MODE_KEY);
    if (v === 'table' || v === 'card') return v;
  } catch (_) {
    // ignore
  }
  return 'card';
})());
const persistRobotListViewMode = () => {
  try {
    window?.localStorage?.setItem(ROBOT_LIST_VIEW_MODE_KEY, robotListViewMode.value);
  } catch (_) {
    // ignore
  }
};

const robotList = ref<any[]>([]);
const loading = ref(false);
const total = ref(0);
const showDetailModal = ref(false);

// 手动平仓提交中的防重复点击（robotId_symbol_positionSide -> bool）
const closeInFlight = ref<Record<string, boolean>>({});
// 手动平仓的“假进度条”（0-100），用于弹框里提示用户请求进行中
const closeProgress = ref<Record<string, number>>({});

const getCloseKey = (robotId: number, symbol: string, positionSide: string) =>
  `${robotId}_${symbol}_${positionSide}`;

const startCloseProgress = (key: string) => {
  closeProgress.value[key] = 8;
  const timer = window.setInterval(() => {
    const cur = closeProgress.value[key] ?? 0;
    // 缓慢逼近 90，给用户“正在处理”的反馈；成功/失败由业务逻辑终止
    if (cur < 90) {
      closeProgress.value[key] = Math.min(90, cur + Math.max(1, Math.round((90 - cur) * 0.15)));
    }
  }, 350);
  return () => {
    window.clearInterval(timer);
  };
};
const currentRobot = ref<any>(null);
const currentRobotStrategy = ref<any>(null); // 当前机器人使用的策略组信息
const currentStrategyTemplate = ref<any>(null); // 当前策略组对应的策略模板（用于显示参数）

// 策略模板选择相关
const showStrategySelector = ref(false);
const selectedGroupId = ref<number | null>(null);
const selectedIsOfficial = ref(false);
const myStrategyGroups = ref<any[]>([]);
const officialStrategyGroups = ref<any[]>([]);
const loadingStrategyGroups = ref(false);
const applyingStrategy = ref(false);

// 算力余额缓存（按用户ID）
const walletPowerMap = ref<Record<number, number>>({});

// 统计数据
const runningCount = ref(0);
const todayNetPnl = ref(0);
const totalNetPnl = ref(0);

// 运行区间数据（来自 钱包-交易明细-运行区间）
// key: robotId -> value
const runningSessionNetPnlMap = ref<Record<number, number>>({});  // 净盈亏（扣手续费）
const runningSessionPnlMap = ref<Record<number, number>>({});      // 盈亏
const runningSessionFeeMap = ref<Record<number, number>>({});      // 手续费
const runningSessionTradeCountMap = ref<Record<number, number>>({}); // 成交笔数
let runningSessionNetPnlLastLoadedAt = 0;

// 实时数据
const tickerData = ref<Record<number, any>>({});
const positionData = ref<Record<number, any[]>>({});
const robotStatusData = ref<Record<number, any>>({});  // 机器人运行状态数据
const analysisData = ref<Record<number, any>>({});  // 策略分析数据
const signalLogs = ref<Record<number, any[]>>({});  // 方向预警日志
// 启动止盈回撤开关状态（key: robotId_symbol_positionSide）
const takeProfitRetreatSwitch = ref<Record<string, boolean>>({});
const executionLogs = ref<Record<number, any[]>>({});  // 交易执行日志

// 金额格式化辅助函数（千分位 + 保留2位小数）
const formatAmount = (val: any): string => {
  if (val === undefined || val === null) return '--';
  const n = Number(val);
  if (!Number.isFinite(n)) return '--';
  return n.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
};

// 盈亏金额格式化（保留8位小数）
const formatPnlAmount = (val: any): string => {
  if (val === undefined || val === null) return '--';
  const n = Number(val);
  if (!Number.isFinite(n)) return '--';
  return n.toFixed(8);
};

// 列表视图列定义（复用现有页面数据，不额外请求后端）
const robotTableColumns: any[] = [
  {
    title: '名称',
    key: 'robotName',
    width: 150,
    align: 'center',
    ellipsis: { tooltip: true },
    render: (row: any) => h('span', { 
      style: 'font-weight: 600; font-size: 14px; color: var(--text-color-1);'
    }, row.robotName || '--'),
  },
  {
    title: '状态',
    key: 'status',
    width: 80,
    align: 'center',
    render: (row: any) => h(NTag, { 
      type: getStatusType(row.status), 
      size: 'small',
      bordered: false,
    }, { default: () => getStatusText(row.status) }),
  },
  {
    title: '交易所',
    key: 'exchange',
    width: 80,
    align: 'center',
    render: (row: any) => h(NTag, { 
      size: 'small',
      bordered: false,
    }, { default: () => String(row.exchange || row.platform || '--').toUpperCase() }),
  },
  {
    title: '货币对',
    key: 'symbol',
    width: 100,
    align: 'center',
    render: (row: any) => h(NTag, { 
      size: 'small', 
      type: 'info',
      bordered: false,
      style: 'font-family: "Consolas", "Monaco", monospace;'
    }, { default: () => row.symbol || row.tradingPair || '--' }),
  },
  {
    title: '账户权益',
    key: 'accountEquity',
    width: 95,
    align: 'center',
    render: (row: any) => {
      const a = analysisData.value[row.id]?.account;
      const v = a?.accountEquity ?? a?.totalBalance;
      return h('span', { 
        style: [
          'font-variant-numeric: tabular-nums;',
          'font-weight: 500;',
          'color: var(--text-color-2);',
          'font-family: "Consolas", "Monaco", monospace;'
        ].join(' ')
      }, formatAmount(v));
    },
  },
  {
    title: '可用余额',
    key: 'availableBalance',
    width: 95,
    align: 'center',
    render: (row: any) => {
      const v = analysisData.value[row.id]?.account?.availableBalance;
      return h('span', { 
        style: [
          'font-variant-numeric: tabular-nums;',
          'font-weight: 500;',
          'color: var(--text-color-2);',
          'font-family: "Consolas", "Monaco", monospace;'
        ].join(' ')
      }, formatAmount(v));
    },
  },
  {
    title: '保证金',
    key: 'margin',
    width: 95,
    align: 'center',
    render: (row: any) => {
      const list = positionData.value[row.id] || [];
      if (!Array.isArray(list) || list.length === 0) {
        return h('span', { style: 'color: var(--text-color-3);' }, '--');
      }
      let sum = 0;
      for (const pos of list) {
        const m = Number(pos?.margin ?? pos?.isolatedMargin ?? 0);
        if (Number.isFinite(m) && m > 0) sum += m;
      }
      return h('span', { 
        style: [
          'font-variant-numeric: tabular-nums;',
          'font-weight: 600;',
          'color: var(--warning-color);',
          'font-family: "Consolas", "Monaco", monospace;'
        ].join(' ')
      }, sum > 0 ? formatAmount(sum) : '--');
    },
  },
  {
    title: '盈亏',
    key: 'totalPnl',
    width: 140,
    align: 'center',
    render: (row: any) => {
      const v = getRunningSessionPnl(row.id);
      if (v === null || v === undefined) return h('span', { style: 'color: var(--text-color-3);' }, '--');
      const n = Number(v);
      const isPositive = n >= 0;
      const color = isPositive ? '#18a058' : '#d03050';
      
      return h(
        'span',
        {
          style: [
            `color: ${color}`,
            'font-weight: 600',
            'font-size: 13px',
            'font-variant-numeric: tabular-nums',
            'font-family: "Consolas", "Monaco", monospace',
          ].join(';'),
        },
        `${n >= 0 ? '+' : ''}${formatPnlAmount(n)}`
      );
    },
  },
  {
    title: '手续费',
    key: 'totalFee',
    width: 130,
    align: 'center',
    render: (row: any) => {
      const v = getRunningSessionFee(row.id);
      if (v === null || v === undefined) return h('span', { style: 'color: var(--text-color-3);' }, '--');
      
      return h(
        'span',
        {
          style: [
            'color: #f0a020',
            'font-weight: 500',
            'font-size: 13px',
            'font-variant-numeric: tabular-nums',
            'font-family: "Consolas", "Monaco", monospace',
          ].join(';'),
        },
        formatPnlAmount(v)
      );
    },
  },
  {
    title: '成交笔数',
    key: 'tradeCount',
    width: 70,
    align: 'center',
    render: (row: any) => {
      const v = getRunningSessionTradeCount(row.id);
      if (v === null || v === undefined) return h('span', { style: 'color: var(--text-color-3);' }, '--');
      
      return h(
        'span',
        {
          style: [
            'font-weight: 500',
            'font-size: 13px',
            'font-variant-numeric: tabular-nums',
          ].join(';'),
        },
        String(v)
      );
    },
  },
  {
    title: '净盈亏',
    key: 'netPnl',
    width: 150,
    align: 'center',
    render: (row: any) => {
      const v = getRunningSessionNetPnl(row.id);
      if (v === null || v === undefined) return h('span', { style: 'color: var(--text-color-3);' }, '--');
      const n = Number(v);
      const isPositive = n >= 0;
      const color = isPositive ? '#18a058' : '#d03050';
      
      return h(
        'span',
        {
          style: [
            `color: ${color}`,
            'font-weight: 700',
            'font-size: 14px',
            'font-variant-numeric: tabular-nums',
            'font-family: "Consolas", "Monaco", monospace;',
          ].join(';'),
        },
        `${n >= 0 ? '+' : ''}${formatPnlAmount(n)}`
      );
    },
  },
  {
    title: '方向 / 未实现盈亏',
    key: 'positions',
    minWidth: 180,
    align: 'center',
    render: (row: any) => {
      const list = positionData.value[row.id] || [];
      if (!Array.isArray(list) || list.length === 0) {
        return h('span', { style: 'color: var(--text-color-3);' }, '--');
      }
      const show = list.slice(0, 2);
      const rows = show.map((pos: any) => {
        const pnl = Number(pos?.unrealizedPnl || 0);
        const isLong = String(pos?.positionSide).toUpperCase() === 'LONG';
        const isPnlPositive = pnl >= 0;
        
        const pnlColor = isPnlPositive ? '#18a058' : '#d03050';
        
        return h(
          'div',
          { style: 'display: flex; align-items: center; justify-content: center; gap: 4px; padding: 1px 0;' },
          [
            h(
              NTag,
              { 
                type: isLong ? 'success' : 'error',
                size: 'small',
                style: 'font-weight: 600; min-width: 32px; text-align: center;'
              },
              { default: () => (isLong ? '多' : '空') }
            ),
            h(
              'span',
              {
                style: [
                  `color: ${pnlColor}`,
                  'font-weight: 700',
                  'font-size: 12px',
                  'font-variant-numeric: tabular-nums',
                  'font-family: "Consolas", "Monaco", monospace',
                ].join(';'),
              },
              `${pnl >= 0 ? '+' : ''}${formatPnlAmount(pnl)}`
            ),
          ]
        );
      });
      if (list.length > show.length) {
        rows.push(h('div', { 
          style: [
            'color: var(--text-color-3);',
            'font-size: 10px;',
            'margin-top: 2px;',
            'text-align: center;',
          ].join(' ')
        }, `... 另有 ${list.length - show.length} 个持仓`));
      }
      return h('div', { style: 'display: flex; flex-direction: column; align-items: center; gap: 4px; padding: 2px 0;' }, rows);
    },
  },
  {
    title: '实时预警',
    key: 'signal',
    minWidth: 110,
    align: 'center',
    render: (row: any) => {
      const sig = analysisData.value[row.id]?.signal;
      const dir = String(sig?.direction || '').toUpperCase();
      if (!dir) {
        return h('span', { style: 'color: var(--text-color-3); font-size: 12px;' }, '暂无信号');
      }
      if (dir !== 'LONG' && dir !== 'SHORT') {
        return h(NTag, { size: 'small', bordered: false }, { default: () => dir });
      }

      const isLong = dir === 'LONG';
      const tagType = isLong ? 'success' : 'error';
      const label = isLong ? '做多' : '做空';
      const dist = isLong ? sig?.distanceFromMin : sig?.distanceFromMax;
      const distNum = Number(dist ?? 0);
      const distText = dist === undefined || dist === null ? '--' : formatAmount(distNum);
      
      return h('div', { style: 'display: flex; flex-direction: column; align-items: center; gap: 4px; padding: 2px 0;' }, [
        h(NTag, { 
          size: 'small',
          type: tagType,
          bordered: false,
        }, { default: () => label }),
        h('span', { 
          style: [
            'color: var(--text-color-3);',
            'font-size: 10px;',
            'font-variant-numeric: tabular-nums;',
            'font-family: "Consolas", "Monaco", monospace;',
            'text-align: center;',
          ].join(' ')
        }, `距${distText}U`),
      ]);
    },
  },
  {
    title: '运行时长',
    key: 'runtime',
    width: 120,
    align: 'center',
    render: (row: any) => {
      const runtime = formatRuntime(getRobotRuntimeSeconds(row));
      return h('span', {
        style: [
          'font-weight: 600;',
          'color: var(--primary-color);',
          'font-variant-numeric: tabular-nums;',
          'font-family: "Consolas", "Monaco", monospace;',
          'font-size: 13px;',
        ].join(' ')
      }, runtime);
    },
  },
  {
    title: '操作',
    key: 'actions',
    width: 90,
    align: 'center',
    render: (row: any) => {
      return h(
        'div',
        { style: 'display: flex; gap: 6px; justify-content: center; flex-wrap: wrap;' },
        [
          h(
            NButton,
            { 
              size: 'small',
              type: 'info',
              secondary: true,
              style: 'font-weight: 500;',
              onClick: () => viewDetail(row)
            },
            { default: () => '详情' }
          ),
        ]
      );
    },
  },
];

// 用于“运行时长”等纯前端展示的每秒刷新（不依赖后端推送频率）
const nowTick = ref<number>(Date.now());
let nowTickTimer: any = null;

// ===== 市场状态与风险偏好映射（每个机器人独立）=====
// 注意：使用 reactive Record，避免动态 key 响应式丢失
const riskMappingExpandedNames = reactive<Record<number, string[]>>({});
const riskMappingExpandedEmpty: string[] = [];
const riskMappingLoading = reactive<Record<number, boolean>>({});
const riskMappingSaving = reactive<Record<number, boolean>>({});
const riskMappingLoadedAt = reactive<Record<number, number>>({});
const marketRiskMappingForm = reactive<Record<number, Record<string, string>>>({});

// 市场状态映射配置（与创建机器人页保持一致）
const marketStateMapping = [
  { key: 'trend', label: '趋势市场', icon: '📈', tagType: 'success' as const },
  { key: 'volatile', label: '震荡市场', icon: '📊', tagType: 'warning' as const },
  { key: 'high_vol', label: '高波动市场', icon: '⚡', tagType: 'error' as const },
  { key: 'low_vol', label: '低波动市场', icon: '😴', tagType: 'info' as const },
];

// 风险偏好选项
const riskPreferenceSelectOptions = [
  { label: '🛡️ 保守', value: 'conservative' },
  { label: '⚖️ 平衡', value: 'balanced' },
  { label: '🚀 激进', value: 'aggressive' },
];

const getRiskLabel = (value: string | undefined) => {
  const map: Record<string, string> = { conservative: '保守', balanced: '平衡', aggressive: '激进' };
  return map[String(value || '').toLowerCase()] || (value || '--');
};

const ensureMarketRiskMappingLoaded = async (robotId: number) => {
  if (!robotId) return;
  if (riskMappingLoading[robotId]) return;
  // 5分钟内不重复拉取，除非用户主动保存后再展开
  const last = riskMappingLoadedAt[robotId] || 0;
  if (last > 0 && Date.now() - last < 5 * 60_000) return;

  riskMappingLoading[robotId] = true;
  try {
    const res: any = await ToogoRobotApi.getRiskConfig({ robotId });
    const cfg = res?.marketRiskMapping || res?.MarketRiskMapping || res?.config?.marketRiskMapping || res?.config?.MarketRiskMapping || {};
    const mapping = cfg || {};
    if (!marketRiskMappingForm[robotId]) {
      marketRiskMappingForm[robotId] = {};
    }
    // 保证四个 key 都存在（后端也要求）
    marketRiskMappingForm[robotId].trend = String(mapping.trend || marketRiskMappingForm[robotId].trend || 'balanced');
    marketRiskMappingForm[robotId].volatile = String(mapping.volatile || marketRiskMappingForm[robotId].volatile || 'balanced');
    marketRiskMappingForm[robotId].high_vol = String(mapping.high_vol || marketRiskMappingForm[robotId].high_vol || 'aggressive');
    marketRiskMappingForm[robotId].low_vol = String(mapping.low_vol || marketRiskMappingForm[robotId].low_vol || 'conservative');
    riskMappingLoadedAt[robotId] = Date.now();
  } catch (e: any) {
    console.warn('[riskMapping] load failed:', e);
  } finally {
    riskMappingLoading[robotId] = false;
  }
};

const onRiskMappingExpanded = (robotId: number, names: string[]) => {
  if (!robotId) return;
  riskMappingExpandedNames[robotId] = Array.isArray(names) ? names : [];
  if ((riskMappingExpandedNames[robotId] || []).includes('marketMapping')) {
    ensureMarketRiskMappingLoaded(robotId);
  }
};

const saveMarketRiskMapping = async (robotId: number) => {
  if (!robotId) return;
  if (riskMappingSaving[robotId]) return;
  const mapping = marketRiskMappingForm[robotId] || {};
  // 前端兜底：补齐 required keys
  const payload = {
    trend: mapping.trend || 'balanced',
    volatile: mapping.volatile || 'balanced',
    high_vol: mapping.high_vol || 'aggressive',
    low_vol: mapping.low_vol || 'conservative',
  };
  riskMappingSaving[robotId] = true;
  try {
    await ToogoRobotApi.saveRiskConfig({
      robotId,
      config: {
        marketRiskMapping: payload,
      },
    });
    message.success('保存成功');
    riskMappingLoadedAt[robotId] = Date.now();
  } catch (e: any) {
    message.error(e?.message || '保存失败');
  } finally {
    riskMappingSaving[robotId] = false;
  }
};

// ===== 交易执行日志（折叠面板 + 懒加载）=====
// 注意：这里使用 reactive 的 Record，避免模板对动态 key 的索引访问出现响应式丢失，导致“展开不起作用”。
const executionExpandedNames = reactive<Record<number, string[]>>({});
const executionOnlyFailed = reactive<Record<number, boolean>>({});
const executionLastLoadedAt = reactive<Record<number, number>>({});
// 稳定的空数组引用：避免模板里使用 `|| []` 每次渲染都创建新数组，导致 Collapse 受控状态不生效
const executionExpandedEmpty: string[] = [];

const refreshExecutionLogs = async (robotId: number, force: boolean = true) => {
  if (!robotId) return;
  if (!force) {
    const last = executionLastLoadedAt[robotId] || 0;
    // 2秒内不重复刷新，避免短时间重复触发
    if (last > 0 && Date.now() - last < 2000) return;
  }
  await loadExecutionLogs(robotId, 50);
};

const onExecutionExpanded = (robotId: number, names: string[]) => {
  if (!robotId) return;
  executionExpandedNames[robotId] = Array.isArray(names) ? names : [];
  // 展开后才加载（按需）
  if ((executionExpandedNames[robotId] || []).includes('execution-logs')) {
    // 不阻塞 UI 更新，避免“点了但没展开”的错觉
    refreshExecutionLogs(robotId, false);
  }
};

// 切换订单日志展开状态（用于点击"最近失败"标签）
const getExecutionLogsForRobot = (robotId: number): any[] => {
  const logs = executionLogs.value[robotId] || [];
  if (executionOnlyFailed[robotId]) {
    return logs.filter((l: any) => {
      const s = String(l?.status || '').toLowerCase();
      return s === 'failed' || s === 'error' || s === 'fail' || s === 'failed.'; // 兼容不同写法
    });
  }
  return logs;
};

// 多周期播报面板展开状态（默认折叠）
const marketRealtimeExpanded = ref<Record<number, boolean>>({});
const toggleMarketRealtime = (robotId: number) => {
  marketRealtimeExpanded.value[robotId] = !marketRealtimeExpanded.value[robotId];
};

// 详情弹窗数据
const currentPositions = ref<any[]>([]);
const currentOpenOrders = ref<any[]>([]);
const orderHistory = ref<any[]>([]);
const powerConsumeList = ref<any[]>([]);
const positionLoading = ref(false);
const orderLoading = ref(false);
const historyLoading = ref(false);

// 订阅信息（仅用于显示倒计时）
const subscriptionInfo = ref<any>({});
const subscriptionCountdownTimer = ref<any>(null); // 倒计时更新定时器

// 定时器
let refreshTimer: any = null;
let fastRefreshTimer: any = null;  // 快速刷新（价格数据）
let orderRefreshTimer: any = null; // 订单数据刷新（10秒）
let wsFallbackCounter = 0;         // WS兜底计数器（降低HTTP轮询频率）

const searchParams = ref({
  status: null,
  platform: null,
  page: 1,
  pageSize: 20,
  perPage: 20,
});

const statusOptions = [
  { label: '全部状态', value: null },
  { label: '未启动', value: 1 },
  { label: '运行中', value: 2 },
  { label: '已暂停', value: 3 },
  { label: '已停用', value: 4 },
];

const platformOptions = [
  { label: '全部交易所', value: null },
  { label: 'Binance', value: 'binance' },
  { label: 'OKX', value: 'okx' },
  { label: 'Gate.io', value: 'gate' },
];

// 持仓列表列定义
const positionColumns = [
  { title: '交易对', key: 'symbol' },
  {
    title: '方向',
    key: 'positionSide',
    render: (row: any) => h(NTag, { type: row.positionSide === 'LONG' ? 'success' : 'error', size: 'small' }, { default: () => row.positionSide === 'LONG' ? '多' : '空' }),
  },
  { title: '数量', key: 'positionAmt', render: (row: any) => Math.abs(row.positionAmt).toFixed(4) },
  { title: '开仓价', key: 'entryPrice', render: (row: any) => row.entryPrice.toFixed(4) },
  { title: '标记价', key: 'markPrice', render: (row: any) => row.markPrice.toFixed(4) },
  {
    title: '未实现盈亏',
    key: 'unrealizedPnl',
    render: (row: any) => h('span', { style: { color: row.unrealizedPnl >= 0 ? 'var(--success-color)' : 'var(--error-color)' } }, `${row.unrealizedPnl >= 0 ? '+' : ''}${row.unrealizedPnl.toFixed(4)} USDT`),
  },
  { title: '杠杆', key: 'leverage', render: (row: any) => `${row.leverage}x` },
  {
    title: '市场状态',
    key: 'marketState',
    width: 100,
    render: (row: any) => {
      if (!row.marketState) return h('span', { style: { color: '#999' } }, '--');
      const marketStateMap: any = {
        trend: { text: '趋势', type: 'success' },
        volatile: { text: '震荡', type: 'warning' },
        high_vol: { text: '高波动', type: 'error' },
        low_vol: { text: '低波动', type: 'info' },
      };
      const state = marketStateMap[row.marketState] || { text: row.marketState, type: 'default' };
      return h(NTag, { type: state.type, size: 'small' }, { default: () => state.text });
    },
  },
  {
    title: '风险偏好',
    key: 'riskPreference',
    width: 100,
    render: (row: any) => {
      if (!row.riskPreference) return h('span', { style: { color: '#999' } }, '--');
      const riskMap: any = {
        conservative: { text: '保守', type: 'info' },
        balanced: { text: '平衡', type: 'warning' },
        aggressive: { text: '激进', type: 'error' },
      };
      const risk = riskMap[row.riskPreference] || { text: row.riskPreference, type: 'default' };
      return h(NTag, { type: risk.type, size: 'small' }, { default: () => risk.text });
    },
  },
  {
    title: '策略参数',
    key: 'strategyParams',
    width: 200,
    render: (row: any) => {
      const params: string[] = [];
      if (row.stopLossPercent || row.stop_loss_percent) {
        params.push(`止损: ${(row.stopLossPercent || row.stop_loss_percent)}%`);
      }
      if (row.autoStartRetreatPercent || row.auto_start_retreat_percent) {
        params.push(`启动: ${(row.autoStartRetreatPercent || row.auto_start_retreat_percent)}%`);
      }
      if (row.profitRetreatPercent || row.profit_retreat_percent) {
        params.push(`回撤: ${(row.profitRetreatPercent || row.profit_retreat_percent)}%`);
      }
      if (params.length === 0) return h('span', { style: { color: '#999' } }, '--');
      return h('div', { style: 'font-size: 11px; line-height: 1.4;' }, params.map(p => h('div', p)));
    },
  },
  {
    title: '操作',
    key: 'actions',
    render: (row: any) => h(NButton, { size: 'small', type: 'warning', onClick: () => closePositionInModal(row) }, { default: () => '平仓' }),
  },
];

// 当前挂单列定义
const openOrderColumns = [
  { title: '订单ID', key: 'orderId', width: 160, ellipsis: { tooltip: true } },
  { title: '交易对', key: 'symbol', width: 100 },
  {
    title: '方向',
    key: 'side',
    width: 80,
    render: (row: any) => {
      const sideMap: any = { BUY: '买入', SELL: '卖出' };
      return h(NTag, { type: row.side === 'BUY' ? 'success' : 'error', size: 'small' }, { default: () => sideMap[row.side] || row.side });
    },
  },
  {
    title: '类型',
    key: 'type',
    width: 80,
    render: (row: any) => {
      const typeMap: any = { MARKET: '市价', LIMIT: '限价' };
      return typeMap[row.type] || row.type;
    },
  },
  {
    title: '委托价',
    key: 'price',
    width: 100,
    render: (row: any) => (row.price && row.price > 0 ? row.price.toFixed(2) : '市价'),
  },
  { title: '数量', key: 'quantity', width: 100, render: (row: any) => row.quantity?.toFixed(4) || '--' },
  { title: '已成交', key: 'filledQty', width: 100, render: (row: any) => row.filledQty?.toFixed(4) || '0' },
  {
    title: '状态',
    key: 'status',
    width: 80,
    render: (row: any) => {
      const statusMap: any = {
        NEW: { text: '待成交', type: 'warning' },
        OPEN: { text: '待成交', type: 'warning' },
        PARTIALLY_FILLED: { text: '部分成交', type: 'info' },
      };
      const status = statusMap[row.status] || { text: row.status, type: 'default' };
      return h(NTag, { type: status.type, size: 'small' }, { default: () => status.text });
    },
  },
  {
    title: '操作',
    key: 'actions',
    width: 80,
    render: (row: any) =>
      h(
        NPopconfirm,
        {
          onPositiveClick: () => cancelOrder(row.orderId),
        },
        {
          trigger: () => h(NButton, { size: 'small', type: 'error' }, { default: () => '撤单' }),
          default: () => '确定要撤销此订单吗？',
        }
      ),
  },
];

// 成交明细列定义
const historyColumns = [
  { title: '成交ID', key: 'orderId', width: 160, ellipsis: { tooltip: true } },
  {
    title: '方向',
    key: 'side',
    width: 80,
    render: (row: any) => {
      const sideMap: any = { BUY: '买入', SELL: '卖出' };
      const type = row.side === 'BUY' ? 'success' : 'error';
      return h(NTag, { type, size: 'small' }, { default: () => sideMap[row.side] || row.side });
    },
  },
  {
    title: '类型',
    key: 'type',
    width: 80,
    render: (row: any) => {
      // 成交明细类型：开仓/平仓
      const typeMap: any = { MARKET: '市价', LIMIT: '限价', '开仓': '开仓', '平仓': '平仓' };
      return typeMap[row.type] || row.type || '--';
    },
  },
  {
    title: '流动性',
    key: 'tradeScope',
    width: 80,
    render: (row: any) => {
      const scopeMap: any = { maker: 'Maker', taker: 'Taker' };
      const type = row.tradeScope === 'maker' ? 'success' : 'warning';
      return h(NTag, { type, size: 'small' }, { default: () => scopeMap[row.tradeScope] || row.tradeScope || '--' });
    },
  },
  {
    title: '成交价',
    key: 'price',
    width: 100,
    render: (row: any) => (row.price && row.price > 0 ? row.price.toFixed(2) : '--'),
  },
  {
    title: '成交数量',
    key: 'quantity',
    width: 100,
    render: (row: any) => row.quantity?.toFixed(4) || '--',
  },
  {
    title: '手续费',
    key: 'fee',
    width: 120,
    render: (row: any) => {
      if (row.fee === undefined || row.fee === null) return '--';
      const fee = parseFloat(row.fee);
      const feeCoin = row.feeCoin || 'USDT';
      return h('span', { style: { color: '#d03050' } }, fee.toFixed(6) + ' ' + feeCoin);
    },
  },
  {
    title: '成交时间',
    key: 'createTime',
    width: 160,
    render: (row: any) => {
      if (!row.createTime) return '--';
      // Bitget 返回的是毫秒时间戳
      const ts = row.createTime > 9999999999 ? row.createTime : row.createTime * 1000;
      return new Date(ts).toLocaleString();
    },
  },
];

// 简化版成交明细列（使用历史订单数据）
const historyColumnsSimple = [
  {
    title: '订单号',
    key: 'orderSn',
    width: 140,
    ellipsis: { tooltip: true },
    render: (row: any) => {
      return h('div', { style: 'font-family: monospace; font-size: 12px;' }, row.orderSn || '--');
    },
  },
  {
    title: '方向',
    key: 'direction',
    width: 70,
    render: (row: any) => {
      const type = row.direction === 'long' ? 'success' : 'error';
      return h(NTag, { type, size: 'small' }, { default: () => row.directionText || row.direction || '--' });
    },
  },
  {
    title: '开仓价',
    key: 'openPrice',
    width: 100,
    render: (row: any) => row.openPrice ? row.openPrice.toFixed(2) : '--',
  },
  {
    title: '平仓价',
    key: 'closePrice',
    width: 100,
    render: (row: any) => row.closePrice ? row.closePrice.toFixed(2) : '--',
  },
  {
    title: '数量',
    key: 'quantity',
    width: 90,
    render: (row: any) => row.quantity ? row.quantity.toFixed(4) : '--',
  },
  {
    title: '盈亏',
    key: 'realizedProfit',
    width: 120,
    render: (row: any) => {
      const profit = row.realizedProfit || 0;
      const color = profit >= 0 ? '#18a058' : '#d03050';
      return h('span', { style: `color: ${color}; font-weight: 600;` }, 
        profit >= 0 ? `+${profit.toFixed(2)}` : profit.toFixed(2)
      );
    },
  },
  {
    title: '持仓时长',
    key: 'holdDurationText',
    width: 100,
  },
  {
    title: '平仓原因',
    key: 'closeReasonText',
    width: 90,
    render: (row: any) => {
      if (!row.closeReasonText) return '--';
      const typeMap: any = {
        止损: 'error',
        止盈: 'success',
        手动: 'info',
        超时: 'warning',
      };
      return h(NTag, { type: typeMap[row.closeReasonText] || 'default', size: 'small' }, 
        { default: () => row.closeReasonText }
      );
    },
  },
  {
    title: '平仓时间',
    key: 'closeTime',
    width: 160,
    render: (row: any) => {
      if (!row.closeTime || row.closeTime === '' || row.closeTime === '2006-01-02 15:00:00' || row.closeTime === '2006-01-02 15:00') {
        return '--';
      }
      return row.closeTime;
    },
  },
];

// 算力消耗列定义（积分不参与消耗，只从算力账户扣除）
const powerColumns = [
  { title: '订单号', key: 'orderSn' },
  { title: '盈利金额', key: 'profitAmount', render: (row: any) => `${row.profitAmount?.toFixed(4)} USDT` },
  { title: '消耗比例', key: 'consumeRate', render: (row: any) => `${(row.consumeRate * 100).toFixed(1)}%` },
  { title: '消耗算力', key: 'consumePower', render: (row: any) => row.consumePower?.toFixed(4) },
  { title: '扣除算力', key: 'fromPower', render: (row: any) => row.fromPower?.toFixed(4) },
  { title: '时间', key: 'createdAt' },
];

// 跳转到历史订单页面
const goToOrderHistory = () => {
  router.push('/toogo/wallet/order-history');
};

const getStatusType = (status: number) => {
  const types: any = { 1: 'default', 2: 'success', 3: 'warning', 4: 'error' };
  return types[status] || 'default';
};

const getStatusText = (status: number) => {
  const texts: any = { 1: '未启动', 2: '运行中', 3: '已暂停', 4: '已停用' };
  return texts[status] || '未知';
};

// 获取连接状态
const getConnectionStatus = (robotId: number) => {
  const status = robotStatusData.value[robotId];
  const hasTicker = tickerData.value[robotId]?.lastPrice;
  
  if (!status && !hasTicker) {
    return { text: '连接中...', class: 'connecting' };
  }
  if (hasTicker) {
    return { text: 'API已连接', class: 'connected' };
  }
  if (status?.connectionError) {
    return { text: '连接失败', class: 'disconnected' };
  }
  return { text: '连接中...', class: 'connecting' };
};

// ==================== 机器人动画状态函数 ====================
// 10种动画状态：happy, thinking, confused, tired, excited, focused, sad, conservative, balanced, aggressive

// 获取机器人情绪类名（根据状态和分析数据自动切换）
const getRobotMoodClass = (robot: any): string => {
  const data = analysisData.value[robot.id];
  const signal = data?.signal;
  const config = data?.config;
  
  // 1. 根据信号方向
  if (signal?.direction === 'LONG') return 'mood-excited';  // 做多信号 - 兴奋
  if (signal?.direction === 'SHORT') return 'mood-focused'; // 做空信号 - 专注
  
  // 2. 根据风险偏好（只使用实时数据，不使用数据库字段作为后备）
  const riskPref = config?.riskPreference;
  if (riskPref === 'conservative') return 'mood-conservative'; // 保守型
  if (riskPref === 'aggressive') return 'mood-aggressive';     // 激进型
  if (riskPref === 'balanced') return 'mood-balanced';         // 平衡型
  
  // 3. 根据市场状态
  const marketState = config?.marketState || signal?.currentMarketState;
  if (marketState === 'trend') return 'mood-happy';      // 趋势市场 - 开心
  if (marketState === 'range') return 'mood-thinking';   // 震荡市场 - 思考
  if (marketState === 'high_vol') return 'mood-excited'; // 高波动 - 兴奋
  if (marketState === 'low_vol') return 'mood-tired';    // 低波动 - 疲惫
  
  // 4. 根据连接状态
  const connStatus = getConnectionStatus(robot.id);
  if (connStatus.class === 'disconnected') return 'mood-sad';       // 断开 - 失落
  if (connStatus.class === 'connecting') return 'mood-thinking';    // 连接中 - 思考
  
  return 'mood-balanced'; // 默认 - 平衡型
};

// 获取机器人眼睛颜色（根据状态变化）
const getRobotEyeColor = (robot: any): string => {
  const moodClass = getRobotMoodClass(robot);
  const colorMap: Record<string, string> = {
    'mood-happy': '#00f3ff',
    'mood-thinking': '#ff9f43',
    'mood-confused': '#00f3ff',
    'mood-tired': '#666666',
    'mood-excited': '#ff0055',
    'mood-focused': '#ff3333',
    'mood-sad': '#4a69bd',
    'mood-conservative': '#4cd137',
    'mood-balanced': '#00a8ff',
    'mood-aggressive': '#e84118',
  };
  return colorMap[moodClass] || '#00f3ff';
};

// 获取机器人情绪文字
const getRobotMoodText = (robot: any): string => {
  const moodClass = getRobotMoodClass(robot);
  const textMap: Record<string, string> = {
    'mood-happy': '🎉 市场趋势明朗',
    'mood-thinking': '🤔 分析市场中...',
    'mood-confused': '❓ 信号不明确',
    'mood-tired': '😴 市场波动较低',
    'mood-excited': '🔥 捕捉到信号!',
    'mood-focused': '🎯 专注做空信号',
    'mood-sad': '😢 连接已断开',
    'mood-conservative': '🛡️ 保守策略运行中',
    'mood-balanced': '⚖️ 平衡策略运行中',
    'mood-aggressive': '🚀 激进策略运行中',
  };
  return textMap[moodClass] || '监控中...';
};

// ==================== 窗口价格曲线相关函数 ====================
const CHART_WIDTH = 640;
const CHART_HEIGHT = 200;

// 获取信号方向文本
const getSignalDirectionText = (direction: string) => {
  if (direction === 'LONG') return '做多';
  if (direction === 'SHORT') return '做空';
  if (direction === 'NEUTRAL') return '监控中';
  return '等待';
};

// 获取信号徽章样式类
const getSignalBadgeClass = (direction: string) => {
  if (direction === 'LONG') return 'badge-long';
  if (direction === 'SHORT') return 'badge-short';
  return 'badge-neutral';
};

// 获取信号图标
const getSignalIcon = (direction: string) => {
  if (direction === 'LONG') return '📈';
  if (direction === 'SHORT') return '📉';
  return '⏳';
};

// 获取市场状态标签类型
const getMarketStateType = (state: string | undefined): 'success' | 'warning' | 'error' | 'info' | 'default' => {
  if (state === 'trend') return 'success';
  if (state === 'range') return 'warning';      // 添加 range 类型映射（震荡-警告色）
  if (state === 'volatile') return 'warning';
  if (state === 'high_vol') return 'error';
  if (state === 'low_vol') return 'info';
  return 'default';
};

// 获取详情弹窗的市场状态（优先使用实时数据）
const getDetailMarketState = () => {
  if (!currentRobot.value) return '';
  const robotId = currentRobot.value.id;
  // 只使用实时分析数据，不使用数据库字段作为后备
  return analysisData.value[robotId]?.config?.marketState 
    || analysisData.value[robotId]?.signal?.currentMarketState 
    || '';
};

// 获取详情弹窗的风险偏好（只使用实时数据）
const getDetailRiskPreference = () => {
  if (!currentRobot.value) return '';
  const robotId = currentRobot.value.id;
  // 只使用实时分析数据，不使用数据库字段作为后备
  return analysisData.value[robotId]?.config?.riskPreference 
    || '';
};

// 获取详情弹窗的策略组名称
const getDetailStrategyGroupName = () => {
  if (!currentRobot.value) return '';
  const robotId = currentRobot.value.id;
  return analysisData.value[robotId]?.config?.strategyGroupName || '';
};

// 获取详情弹窗的策略模板名称
const getDetailStrategyName = () => {
  if (!currentRobot.value) return '';
  const robotId = currentRobot.value.id;
  return analysisData.value[robotId]?.config?.strategyName || '';
};

// ===== 获取策略参数的辅助函数（重新设计：统一逻辑）=====
// 【重新设计】简化逻辑：直接从 analysisData.config 获取策略参数，不再使用后备值

// 检查是否有错误信息
const hasConfigError = (robotId: number): boolean => {
  return !!analysisData.value[robotId]?.config?.errorMessage;
};

// 获取错误信息
const getConfigError = (robotId: number): string | null => {
  return analysisData.value[robotId]?.config?.errorMessage || null;
};

// 从 currentStrategy JSON 中解析市场状态映射
const getMarketRiskMapping = (robot: any): Record<string, string> | null => {
  // 【重要】映射关系存储在 remark 字段中（JSON格式），创建时保存的独立映射关系
  if (!robot.remark) return null;
  try {
    const mapping = typeof robot.remark === 'string' 
      ? JSON.parse(robot.remark) 
      : robot.remark;
    // 验证是否是有效的映射关系对象
    if (mapping && typeof mapping === 'object' && !Array.isArray(mapping)) {
      return mapping;
    }
    return null;
  } catch (e) {
    console.debug('解析映射关系失败:', e, 'remark:', robot.remark);
    return null;
  }
};

// 获取策略组名称
const getStrategyGroupName = (robot: any): string | null => {
  if (robot.strategyGroupName) return robot.strategyGroupName;
  return null;
};

// 详情弹窗专用：获取策略参数（统一从 config 获取，不再区分启动状态）
const getDetailLeverage = () => {
  if (!currentRobot.value) return null;
  return analysisData.value[currentRobot.value.id]?.config?.leverage ?? null;
};

const getDetailMarginPercent = () => {
  if (!currentRobot.value) return null;
  return analysisData.value[currentRobot.value.id]?.config?.marginPercent ?? null;
};

const getDetailStopLossPercent = () => {
  if (!currentRobot.value) return null;
  return analysisData.value[currentRobot.value.id]?.config?.stopLossPercent ?? null;
};

// 获取详情页启动止盈百分比
const getDetailAutoStartRetreat = () => {
  if (!currentRobot.value) return null;
  return analysisData.value[currentRobot.value.id]?.config?.autoStartRetreat ?? null;
};

// 获取详情页止盈回撤百分比
const getDetailProfitRetreat = () => {
  if (!currentRobot.value) return null;
  return analysisData.value[currentRobot.value.id]?.config?.takeProfitPercent ?? null;
};

const getDetailTimeWindow = () => {
  if (!currentRobot.value) return null;
  return analysisData.value[currentRobot.value.id]?.config?.timeWindow ?? null;
};

const getDetailThreshold = () => {
  if (!currentRobot.value) return null;
  return analysisData.value[currentRobot.value.id]?.config?.threshold ?? null;
};

// 列表页专用：获取策略参数（统一从 config 获取，不使用后备值）
const getRobotLeverage = (robotId: number) => {
  return analysisData.value[robotId]?.config?.leverage ?? 0;
};

const getRobotMarginPercent = (robotId: number) => {
  return analysisData.value[robotId]?.config?.marginPercent ?? 0;
};

// 获取机器人止损百分比（从机器人详情配置获取）
const getRobotStopLossPercent = (robotId: number) => {
  return analysisData.value[robotId]?.config?.stopLossPercent ?? 0;
};

// 获取机器人启动止盈百分比（从机器人详情配置获取）
const getRobotAutoStartRetreat = (robotId: number) => {
  return analysisData.value[robotId]?.config?.autoStartRetreat ?? 0;
};

// 获取机器人止盈回撤百分比（从机器人详情配置获取）
const getRobotProfitRetreat = (robotId: number) => {
  return analysisData.value[robotId]?.config?.takeProfitPercent ?? 0;
};

// 格式化价格（2位小数）
const formatPrice = (price: number | undefined) => {
  if (typeof price !== 'number' || isNaN(price)) return '--';
  return price.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
};

// 格式化价格（带USDT单位）
const formatPriceUsdt = (price: number | undefined) => {
  if (typeof price !== 'number' || isNaN(price)) return '--';
  return price.toFixed(4) + ' USDT';
};

// 格式化价格涨跌
const formatPriceChange = (change: number | undefined) => {
  if (typeof change !== 'number' || isNaN(change)) return '';
  const sign = change >= 0 ? '+' : '';
  return `${sign}${change.toFixed(2)}%`;
};

// 获取价格涨跌样式类
const getPriceChangeClass = (robotId: number) => {
  const ticker = tickerData.value[robotId];
  if (!ticker) return '';
  const change = ticker.change24h || 0;
  return change >= 0 ? 'up' : 'down';
};

// 格式化时间窗口（秒转分钟/秒）
const formatWindowTime = (seconds: number | undefined) => {
  if (typeof seconds !== 'number' || isNaN(seconds) || seconds <= 0) return '--';
  if (seconds >= 60) {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return secs > 0 ? `${mins}分${secs}秒` : `${mins}分钟`;
  }
  return `${seconds}秒`;
};

// 格式化运行时长
const formatRuntime = (seconds: number | undefined) => {
  if (!seconds || seconds <= 0) return '0秒';
  
  const days = Math.floor(seconds / 86400);
  const hours = Math.floor((seconds % 86400) / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  
  if (days > 0) {
    return `${days}天${hours}时`;
  } else if (hours > 0) {
    return `${hours}时${minutes}分`;
  } else if (minutes > 0) {
    return `${minutes}分`;
  } else {
    return `${seconds}秒`;
  }
};

// 获取运行时长（秒）：优先用 startTime 本地计算，其次用后端 runtimeSeconds，再兜底机器人列表字段
const getRobotRuntimeSeconds = (robot: any): number => {
  const rid = Number(robot?.id || 0);
  const cfg = rid ? analysisData.value?.[rid]?.config : null;

  // 1) 优先用 startTime 本地计算（保证每秒变化）
  const startTimeStr = String(cfg?.startTime || robot?.startTime || '').trim();
  if (startTimeStr) {
    // 后端曾出现把 Go layout "2006-01-02 15:04:05" 原样返回的情况；这种不应参与计算
    if (startTimeStr.includes('2006-01-02') || startTimeStr.includes('20060102')) {
      // ignore
    } else {
    // 后端常见格式 "YYYY-MM-DD HH:mm:ss"；在部分浏览器需要替换成 ISO
    const t = new Date(startTimeStr.replace(' ', 'T')).getTime();
    if (!isNaN(t)) {
      // 兜底：异常年份（例如 1970/2006 模板/默认值）不参与计算
      const y = new Date(t).getFullYear();
      if (y >= 2015 && y <= 2100) {
      const secs = Math.floor((nowTick.value - t) / 1000);
      return secs > 0 ? secs : 0;
      }
    }
    }
  }

  // 2) 后端实时分析里带的 runtimeSeconds
  const rs = Number(cfg?.runtimeSeconds ?? 0);
  if (rs > 0) return rs;

  // 3) 机器人列表字段（DB字段，可能不实时）
  const rr = Number(robot?.runtimeSeconds ?? 0);
  if (rr > 0) return rr;
  return 0;
};


// 格式化市场状态（兼容多种格式）
const formatMarketState = (state: string | undefined) => {
  const stateMap: Record<string, string> = {
    // 新格式（小写）
    'trend': '趋势',
    'range': '震荡',      // 添加 range 映射
    'volatile': '震荡',
    'high_vol': '高波动',
    'low_vol': '低波动',
    // 旧格式（大写）
    'STRONG_UPTREND': '强势上涨',
    'MILD_UPTREND': '温和上涨',
    'RANGING': '震荡整理',
    'MILD_DOWNTREND': '温和下跌',
    'STRONG_DOWNTREND': '强势下跌',
    'HIGH_VOLATILITY': '高波动',
    'LOW_VOLATILITY': '低波动',
  };
  return stateMap[state || ''] || state || '--';
};

// 格式化风险偏好
const formatRiskPref = (pref: string | undefined) => {
  const prefMap: Record<string, string> = {
    'aggressive': '🚀 激进',
    'balanced': '⚖️ 平衡',
    'conservative': '🛡️ 保守',
  };
  return prefMap[pref || ''] || pref || '--';
};


// 格式化日志时间（必须健壮：后端可能返回 "YYYY-MM-DD HH:mm:ss" 字符串，部分浏览器 new Date() 解析会失败）
const formatLogTime = (time: string | number | undefined) => {
  if (!time) return '--';
  try {
    let input: any = time;
    if (typeof input === 'string') {
      // 兼容 "2025-12-26 16:37:20" → "2025-12-26T16:37:20"
      // 以及后端可能返回的 gtime.String() 格式
      input = input.trim().replace(' ', 'T');
    }
    const date = new Date(input);
    if (isNaN(date.getTime())) return '--';
  return date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit', second: '2-digit' });
  } catch (e) {
    return '--';
  }
};

// 格式化持仓时间（日期+时间）
const formatTime = (time: string | number | undefined) => {
  if (!time || time === 0) return '--';
  // 如果是字符串，检查是否是有效的时间格式
  if (typeof time === 'string') {
    // 如果字符串看起来像是Go的时间格式化模板（2006-01-02 15:04:05），返回--
    if (time.includes('2006-01-02') || time.includes('20060102')) {
      return '--';
    }
  }
  const date = new Date(typeof time === 'number' ? time : time);
  // 检查日期是否有效（Invalid Date 或时间戳为0）
  if (isNaN(date.getTime()) || date.getTime() === 0) {
    return '--';
  }
  return date.toLocaleString('zh-CN', { 
    month: '2-digit', 
    day: '2-digit', 
    hour: '2-digit', 
    minute: '2-digit', 
    second: '2-digit' 
  });
};

// 格式化风险偏好（用于持仓列表显示）
const formatRiskPreference = (pref: string | undefined): string => {
  if (!pref) return '--';
  const prefMap: Record<string, string> = {
    'conservative': '保守',
    'balanced': '平衡',
    'aggressive': '激进',
  };
  return prefMap[pref.toLowerCase()] || pref;
};

// ①、启动止盈进度计算：
// 当前盈利百分比 = 未实现盈亏 / 保证金 × 100%
// 触发条件：当前盈利百分比 >= 设定的启动止盈百分比时，自动启动止盈回撤
// 血条进度 = (当前盈利百分比 / 设定启动止盈百分比) × 100%
// 【重要】一旦止盈回撤已启动，血条锁定在100%
const calcStartProfitProgress = (pos: any, robot: any) => {
  // 【完全由后端控制】只使用后端计算的值，不再进行前端计算
  // 如果后端没有返回数据，返回0（表示未计算或未满足条件）
  if (pos && ('takeProfitStartProgress' in pos) && pos.takeProfitStartProgress !== undefined && pos.takeProfitStartProgress !== null) {
    const v = Number(pos.takeProfitStartProgress);
    if (!Number.isNaN(v)) {
      return Math.max(0, Math.min(100, v));
    }
  }
  // 后端未返回数据时，返回0（不显示进度）
  return 0;
};

// ①、止损进度计算（完全由后端控制）
// 【完全由后端控制】只使用后端计算的值，不再进行前端计算
const calcStopLossProgress = (pos: any, robot: any) => {
  // 如果后端没有返回数据，返回0（表示未计算或未满足条件）
  if (pos && ('stopLossProgress' in pos) && pos.stopLossProgress !== undefined && pos.stopLossProgress !== null) {
    const v = Number(pos.stopLossProgress);
    if (!Number.isNaN(v)) {
      return Math.max(0, Math.min(100, v));
    }
  }
  // 后端未返回数据时，返回0（不显示进度）
  return 0;
};

// ②、止盈回撤进度计算（完全由后端控制）
// 【完全由后端控制】只使用后端计算的值，不再进行前端计算
const calcProfitRetreatProgress = (pos: any, robot: any) => {
  // 如果后端没有返回数据，返回100%（默认满条，表示安全状态）
  if (pos && ('takeProfitRetreatBar' in pos) && pos.takeProfitRetreatBar !== undefined && pos.takeProfitRetreatBar !== null) {
    const v = Number(pos.takeProfitRetreatBar);
    if (!Number.isNaN(v)) {
      return Math.max(0, Math.min(100, v));
    }
  }
  // 后端未返回数据时，返回100%（默认满条）
  return 100;
};

// 格式化更新时间（相对时间）
const formatUpdateTime = (time: string | number | undefined) => {
  if (!time) return '刚刚';
  const now = Date.now();
  const updateTime = typeof time === 'number' ? time : new Date(time).getTime();
  const diff = Math.floor((now - updateTime) / 1000); // 秒
  
  if (diff < 5) return '刚刚';
  if (diff < 60) return `${diff}秒前`;
  if (diff < 3600) return `${Math.floor(diff / 60)}分钟前`;
  return new Date(updateTime).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' });
};

// 加载方向预警日志（只显示有价值的信号：long/short）
const loadSignalLogs = async (robotId: number) => {
  try {
    const res = await ToogoRobotApi.signalLogs({ robotId, limit: 10 });
    if (res?.list) {
      // API已经过滤只返回有价值的信号（long/short）
      signalLogs.value[robotId] = res.list;
    }
  } catch (error: any) {
    console.debug('加载方向预警日志失败:', error);
    signalLogs.value[robotId] = [];
  }
};

// 加载交易执行日志
// - 列表页会“预取最近几条”，用于展示数量/最近失败
// - 展开面板会加载更多
const loadExecutionLogs = async (robotId: number, limit: number = 50) => {
  try {
    const res = await ToogoRobotApi.executionLogs({ robotId, limit });
    if (res?.list) {
      executionLogs.value[robotId] = res.list;
    }
    executionLastLoadedAt[robotId] = Date.now();
    if (executionOnlyFailed[robotId] === undefined) {
      executionOnlyFailed[robotId] = false;
    }
  } catch (error: any) {
    console.debug('加载交易执行日志失败:', error);
    executionLogs.value[robotId] = [];
  }
};

// 预取订单日志（避免必须手动展开才能看到“最近失败/日志数量”）
const preloadExecutionLogs = async () => {
  const robots = robotList.value || [];
  if (!robots.length) return;
  // 控制并发，避免接口风暴
  const concurrency = robots.length <= 8 ? 4 : 2;
  let idx = 0;
  const worker = async () => {
    while (idx < robots.length) {
      const i = idx++;
      const r = robots[i];
      if (!r?.id) continue;
      // 已加载过就跳过
      if (Array.isArray(executionLogs.value[r.id]) && executionLogs.value[r.id].length > 0) continue;
      await loadExecutionLogs(r.id, 8);
    }
  };
  await Promise.all(Array.from({ length: concurrency }, () => worker()));
};

const normalizeEventType = (eventType: any) => {
  // 兼容：有些库里 event_type 可能带尾逗号（例如 close_take_profit，）
  return String(eventType || '')
    .trim()
    .replace(/[，,]+$/g, '');
};

// 获取事件类型标签
const getEventTypeTag = (eventType: string): 'success' | 'warning' | 'error' | 'info' | 'default' | 'primary' => {
  const t = normalizeEventType(eventType);
  const typeMap: Record<string, 'success' | 'warning' | 'error' | 'info' | 'default' | 'primary'> = {
    'order_submit': 'warning',
    'order_attempt': 'warning',
    'order_success': 'success',
    'order_failed': 'error',
    'close_manual': 'warning',
    'close_stop_loss': 'warning',
    'close_take_profit': 'warning',
  };
  return typeMap[t] || 'default';
};

// 获取事件类型文本
const getEventTypeText = (eventType: string) => {
  const t = normalizeEventType(eventType);
  const textMap: Record<string, string> = {
    'order_submit': '提交下单',
    'order_attempt': '提交下单',
    'order_success': '下单成功',
    'order_failed': '下单失败',
    'close_manual': '手动平仓',
    'close_stop_loss': '止损平仓',
    'close_take_profit': '止盈平仓',
  };
  return textMap[t] || t;
};

// 获取日志策略参数
const getLogStrategyParams = (log: any): any => {
  if (!log.eventData) return null;
  try {
    const data = typeof log.eventData === 'string' ? JSON.parse(log.eventData) : log.eventData;
    return {
      marketState: data?.market_state || null,
      riskPreference: data?.risk_preference || null,
      leverage: data?.leverage || null,
      marginPercent: data?.margin_percent || null,
      stopLossPercent: data?.stop_loss_percent || null,
      autoStartRetreat: data?.auto_start_retreat || null,
      profitRetreatPercent: data?.profit_retreat_percent || null,
      timeWindow: data?.time_window || null,
      threshold: data?.threshold || null,
      exchangeOrderId: data?.exchange_order_id || null,
      avgPrice: data?.avg_price || null,
      filledQty: data?.filled_qty || null,
    };
  } catch (e) {
    return null;
  }
};

// 获取日志提交参数（提交下单时的具体内容）
const getLogSubmitParams = (log: any): any => {
  if (!log.eventData) return null;
  try {
    const data = typeof log.eventData === 'string' ? JSON.parse(log.eventData) : log.eventData;
    return {
      symbol: data?.symbol || null,
      side: data?.side || null,
      position_side: data?.position_side || null,
      type: data?.type || null,
      quantity: data?.quantity || null,
      entry_price: data?.entry_price || null,
      leverage: data?.leverage || null,
      margin: data?.margin || null,
      margin_percent: data?.margin_percent || null,
      market_state: data?.market_state || null,
      risk_preference: data?.risk_preference || null,
    };
  } catch (e) {
    return null;
  }
};

// 获取失败原因（容错：字段可能不存在/格式不一致，绝不抛异常）
// 说明：
// - 后端 API 当前返回的 ExecutionLogItem 里没有 failure_reason 字段，因此这里优先从 eventData/message 里提取。
const getFailureReason = (log: any): string => {
  try {
    if (!log) return '';
    const status = String(log.status || '').toLowerCase();
    if (status !== 'failed' && status !== 'error' && status !== 'fail' && status !== 'failed.') return '';

    // 兼容：如果未来后端补了字段
    const direct =
      (log.failureReason ?? log.failure_reason ?? log.failureMessage ?? log.failure_message ?? '').toString().trim();
    if (direct) return direct;

    // 从 eventData 里提取
    const raw = log.eventData;
    if (raw) {
      const data = typeof raw === 'string' ? JSON.parse(raw) : raw;
      const reason =
        (data?.failure_reason ??
          data?.failureReason ??
          data?.reason ??
          data?.error ??
          data?.err ??
          '').toString().trim();
      if (reason) return reason;
    }

    // 兜底：使用 message（但避免重复显示与 message 相同的内容）
    const msg = String(log.message || '').trim();
    return msg;
  } catch {
    return '';
  }
};

// 获取日志阶段（用于把一条失败日志拆到更明确的步骤）
// - 后端常在 eventData.step 写入失败发生在哪一步
// - 平仓类日志可能用 eventData.close_type
const getLogStep = (log: any): string => {
  try {
    if (!log) return '';
    const raw = log.eventData;
    if (!raw) return '';
    const data = typeof raw === 'string' ? JSON.parse(raw) : raw;

    const step = String(data?.step || '').trim();
    if (step) {
      const stepMap: Record<string, string> = {
        robot_check: '机器人检查',
        signal_check: '信号检查',
        auto_trade_check: '开关检查',
        position_check: '持仓检查',
        lock_check: '下单锁',
        risk_check: '风控检查',
        submit_order: '提交订单',
        create_order: '创建订单',
        wait_fill: '等待成交',
        sync_order: '同步订单',
        save_order: '保存订单',
        exchange: '交易所',
        system: '系统',
      };
      return stepMap[step] || step;
    }

    const closeType = String(data?.close_type || '').trim();
    if (closeType) {
      const closeMap: Record<string, string> = {
        manual: '手动平仓',
        stop_loss: '止损',
        take_profit: '止盈',
      };
      return closeMap[closeType] || closeType;
    }

    return '';
  } catch {
    return '';
  }
};

// 计算价格曲线路径
const getPriceChartPath = (analysis: any) => {
  const priceWindow = analysis?.priceWindow;
  if (!priceWindow || priceWindow.length < 2) return '';
  
  const prices = priceWindow.map((p: any) => p.price);
  const minPrice = Math.min(...prices);
  const maxPrice = Math.max(...prices);
  const threshold = analysis?.signal?.signalThreshold || 0;
  
  // 计算图表边界
  const chartMin = Math.min(minPrice, minPrice - threshold * 0.1);
  const chartMax = Math.max(maxPrice, maxPrice + threshold * 0.1);
  const range = chartMax - chartMin || 1;
  
  const step = CHART_WIDTH / (priceWindow.length - 1);
  
  let path = '';
  priceWindow.forEach((point: any, index: number) => {
    const x = index * step;
    const y = CHART_HEIGHT - ((point.price - chartMin) / range) * CHART_HEIGHT;
    path += index === 0 ? `M ${x} ${y}` : ` L ${x} ${y}`;
  });
  
  return path;
};

// 计算价格曲线填充路径
const getPriceChartFillPath = (analysis: any) => {
  const priceWindow = analysis?.priceWindow;
  if (!priceWindow || priceWindow.length < 2) return '';
  
  const prices = priceWindow.map((p: any) => p.price);
  const minPrice = Math.min(...prices);
  const maxPrice = Math.max(...prices);
  const threshold = analysis?.signal?.signalThreshold || 0;
  
  const chartMin = Math.min(minPrice, minPrice - threshold * 0.1);
  const chartMax = Math.max(maxPrice, maxPrice + threshold * 0.1);
  const range = chartMax - chartMin || 1;
  
  const step = CHART_WIDTH / (priceWindow.length - 1);
  const basePrice = priceWindow[0].price;
  const baseY = CHART_HEIGHT - ((basePrice - chartMin) / range) * CHART_HEIGHT;
  
  let path = '';
  priceWindow.forEach((point: any, index: number) => {
    const x = index * step;
    const y = CHART_HEIGHT - ((point.price - chartMin) / range) * CHART_HEIGHT;
    path += index === 0 ? `M ${x} ${y}` : ` L ${x} ${y}`;
  });
  
  // 闭合路径
  const lastX = (priceWindow.length - 1) * step;
  path += ` L ${lastX} ${baseY} L 0 ${baseY} Z`;
  
  return path;
};

// 获取做多阈值线Y坐标
const getLongThresholdY = (analysis: any) => {
  const priceWindow = analysis?.priceWindow;
  if (!priceWindow || priceWindow.length < 2) return CHART_HEIGHT;
  
  const prices = priceWindow.map((p: any) => p.price);
  const minPrice = Math.min(...prices);
  const maxPrice = Math.max(...prices);
  const threshold = analysis?.signal?.signalThreshold || 0;
  
  const chartMin = Math.min(minPrice, minPrice - threshold * 0.1);
  const chartMax = Math.max(maxPrice, maxPrice + threshold * 0.1);
  const range = chartMax - chartMin || 1;
  
  const buyThresholdPrice = minPrice + threshold;
  return CHART_HEIGHT - ((buyThresholdPrice - chartMin) / range) * CHART_HEIGHT;
};

// 获取做空阈值线Y坐标
const getShortThresholdY = (analysis: any) => {
  const priceWindow = analysis?.priceWindow;
  if (!priceWindow || priceWindow.length < 2) return 0;
  
  const prices = priceWindow.map((p: any) => p.price);
  const minPrice = Math.min(...prices);
  const maxPrice = Math.max(...prices);
  const threshold = analysis?.signal?.signalThreshold || 0;
  
  const chartMin = Math.min(minPrice, minPrice - threshold * 0.1);
  const chartMax = Math.max(maxPrice, maxPrice + threshold * 0.1);
  const range = chartMax - chartMin || 1;
  
  const sellThresholdPrice = maxPrice - threshold;
  return CHART_HEIGHT - ((sellThresholdPrice - chartMin) / range) * CHART_HEIGHT;
};

// ==================== 垂直图表函数（三列布局用） ====================
const VCHART_WIDTH = 200;
const VCHART_HEIGHT = 140;

// 垂直图表：价格曲线路径
const getPriceChartPathVertical = (analysis: any) => {
  const priceWindow = analysis?.priceWindow;
  if (!priceWindow || priceWindow.length < 2) return '';
  
  const prices = priceWindow.map((p: any) => p.price);
  const minPrice = Math.min(...prices);
  const maxPrice = Math.max(...prices);
  const threshold = analysis?.signal?.signalThreshold || 0;
  
  const chartMin = Math.min(minPrice, minPrice - threshold * 0.1);
  const chartMax = Math.max(maxPrice, maxPrice + threshold * 0.1);
  const range = chartMax - chartMin || 1;
  
  const step = VCHART_WIDTH / (priceWindow.length - 1);
  let path = '';
  priceWindow.forEach((point: any, index: number) => {
    const x = index * step;
    const y = VCHART_HEIGHT - ((point.price - chartMin) / range) * VCHART_HEIGHT;
    path += index === 0 ? `M ${x} ${y}` : ` L ${x} ${y}`;
  });
  return path;
};

// 垂直图表：价格曲线填充路径
const getPriceChartFillPathVertical = (analysis: any) => {
  const priceWindow = analysis?.priceWindow;
  if (!priceWindow || priceWindow.length < 2) return '';
  
  const prices = priceWindow.map((p: any) => p.price);
  const minPrice = Math.min(...prices);
  const maxPrice = Math.max(...prices);
  const threshold = analysis?.signal?.signalThreshold || 0;
  
  const chartMin = Math.min(minPrice, minPrice - threshold * 0.1);
  const chartMax = Math.max(maxPrice, maxPrice + threshold * 0.1);
  const range = chartMax - chartMin || 1;
  
  const step = VCHART_WIDTH / (priceWindow.length - 1);
  let path = '';
  priceWindow.forEach((point: any, index: number) => {
    const x = index * step;
    const y = VCHART_HEIGHT - ((point.price - chartMin) / range) * VCHART_HEIGHT;
    path += index === 0 ? `M ${x} ${y}` : ` L ${x} ${y}`;
  });
  
  const lastX = (priceWindow.length - 1) * step;
  path += ` L ${lastX} ${VCHART_HEIGHT} L 0 ${VCHART_HEIGHT} Z`;
  return path;
};

// 垂直图表：做多阈值线Y坐标
const getLongThresholdYVertical = (analysis: any) => {
  const priceWindow = analysis?.priceWindow;
  if (!priceWindow || priceWindow.length < 2) return VCHART_HEIGHT;
  
  const prices = priceWindow.map((p: any) => p.price);
  const minPrice = Math.min(...prices);
  const maxPrice = Math.max(...prices);
  const threshold = analysis?.signal?.signalThreshold || 0;
  
  const chartMin = Math.min(minPrice, minPrice - threshold * 0.1);
  const chartMax = Math.max(maxPrice, maxPrice + threshold * 0.1);
  const range = chartMax - chartMin || 1;
  
  const buyThresholdPrice = minPrice + threshold;
  return VCHART_HEIGHT - ((buyThresholdPrice - chartMin) / range) * VCHART_HEIGHT;
};

// 垂直图表：做空阈值线Y坐标
const getShortThresholdYVertical = (analysis: any) => {
  const priceWindow = analysis?.priceWindow;
  if (!priceWindow || priceWindow.length < 2) return 0;
  
  const prices = priceWindow.map((p: any) => p.price);
  const minPrice = Math.min(...prices);
  const maxPrice = Math.max(...prices);
  const threshold = analysis?.signal?.signalThreshold || 0;
  
  const chartMin = Math.min(minPrice, minPrice - threshold * 0.1);
  const chartMax = Math.max(maxPrice, maxPrice + threshold * 0.1);
  const range = chartMax - chartMin || 1;
  
  const sellThresholdPrice = maxPrice - threshold;
  return VCHART_HEIGHT - ((sellThresholdPrice - chartMin) / range) * VCHART_HEIGHT;
};

// 获取做多触发价格
const getLongTriggerPrice = (analysis: any) => {
  const priceWindow = analysis?.priceWindow;
  if (!priceWindow || priceWindow.length < 2) return 0;
  
  const prices = priceWindow.map((p: any) => p.price);
  const minPrice = Math.min(...prices);
  const threshold = analysis?.signal?.signalThreshold || 0;
  
  return minPrice + threshold;
};

// 获取做空触发价格
const getShortTriggerPrice = (analysis: any) => {
  const priceWindow = analysis?.priceWindow;
  if (!priceWindow || priceWindow.length < 2) return 0;
  
  const prices = priceWindow.map((p: any) => p.price);
  const maxPrice = Math.max(...prices);
  const threshold = analysis?.signal?.signalThreshold || 0;
  
  return maxPrice - threshold;
};

// 获取最低价X坐标
const getMinPriceX = (analysis: any) => {
  const priceWindow = analysis?.priceWindow;
  if (!priceWindow || priceWindow.length < 2) return 0;
  
  const prices = priceWindow.map((p: any) => p.price);
  const minPrice = Math.min(...prices);
  const minIndex = prices.indexOf(minPrice);
  const step = CHART_WIDTH / (priceWindow.length - 1);
  return minIndex * step;
};

// 获取最低价Y坐标
const getMinPriceY = (analysis: any) => {
  const priceWindow = analysis?.priceWindow;
  if (!priceWindow || priceWindow.length < 2) return CHART_HEIGHT;
  
  const prices = priceWindow.map((p: any) => p.price);
  const minPrice = Math.min(...prices);
  const maxPrice = Math.max(...prices);
  const threshold = analysis?.signal?.signalThreshold || 0;
  
  const chartMin = Math.min(minPrice, minPrice - threshold * 0.1);
  const chartMax = Math.max(maxPrice, maxPrice + threshold * 0.1);
  const range = chartMax - chartMin || 1;
  
  return CHART_HEIGHT - ((minPrice - chartMin) / range) * CHART_HEIGHT;
};

// 获取最高价X坐标
const getMaxPriceX = (analysis: any) => {
  const priceWindow = analysis?.priceWindow;
  if (!priceWindow || priceWindow.length < 2) return 0;
  
  const prices = priceWindow.map((p: any) => p.price);
  const maxPrice = Math.max(...prices);
  const maxIndex = prices.indexOf(maxPrice);
  const step = CHART_WIDTH / (priceWindow.length - 1);
  return maxIndex * step;
};

// 获取最高价Y坐标
const getMaxPriceY = (analysis: any) => {
  const priceWindow = analysis?.priceWindow;
  if (!priceWindow || priceWindow.length < 2) return 0;
  
  const prices = priceWindow.map((p: any) => p.price);
  const minPrice = Math.min(...prices);
  const maxPrice = Math.max(...prices);
  const threshold = analysis?.signal?.signalThreshold || 0;
  
  const chartMin = Math.min(minPrice, minPrice - threshold * 0.1);
  const chartMax = Math.max(maxPrice, maxPrice + threshold * 0.1);
  const range = chartMax - chartMin || 1;
  
  return CHART_HEIGHT - ((maxPrice - chartMin) / range) * CHART_HEIGHT;
};

// 获取当前价Y坐标
const getCurrentPriceY = (analysis: any) => {
  const priceWindow = analysis?.priceWindow;
  if (!priceWindow || priceWindow.length < 2) return CHART_HEIGHT / 2;
  
  const prices = priceWindow.map((p: any) => p.price);
  const minPrice = Math.min(...prices);
  const maxPrice = Math.max(...prices);
  const currentPrice = prices[prices.length - 1];
  const threshold = analysis?.signal?.signalThreshold || 0;
  
  const chartMin = Math.min(minPrice, minPrice - threshold * 0.1);
  const chartMax = Math.max(maxPrice, maxPrice + threshold * 0.1);
  const range = chartMax - chartMin || 1;
  
  return CHART_HEIGHT - ((currentPrice - chartMin) / range) * CHART_HEIGHT;
};

// 获取当前价X坐标
const getCurrentPriceX = (analysis: any) => {
  const priceWindow = analysis?.priceWindow;
  if (!priceWindow || priceWindow.length < 2) return CHART_WIDTH;
  
  const step = CHART_WIDTH / (priceWindow.length - 1);
  return (priceWindow.length - 1) * step;
};

// 获取基准线Y坐标（起始价）
const getBaselineY = (analysis: any) => {
  const priceWindow = analysis?.priceWindow;
  if (!priceWindow || priceWindow.length < 2) return CHART_HEIGHT / 2;
  
  const prices = priceWindow.map((p: any) => p.price);
  const minPrice = Math.min(...prices);
  const maxPrice = Math.max(...prices);
  const basePrice = priceWindow[0].price;
  const threshold = analysis?.signal?.signalThreshold || 0;
  
  const chartMin = Math.min(minPrice, minPrice - threshold * 0.1);
  const chartMax = Math.max(maxPrice, maxPrice + threshold * 0.1);
  const range = chartMax - chartMin || 1;
  
  return CHART_HEIGHT - ((basePrice - chartMin) / range) * CHART_HEIGHT;
};


const loadData = async () => {
  loading.value = true;
  try {
    const res = await ToogoRobotApi.list(searchParams.value);
    robotList.value = res?.list || [];
    total.value = res?.totalCount || 0;

    // 计算统计数据
    runningCount.value = robotList.value.filter((r: any) => r.status === 2).length;
    
    // 顶部统计：今日净盈亏（成交流水）+ 累计净盈亏（运行区间）
    await loadNetPnlSummary();

    // 加载运行区间净盈亏（用于机器人列表“净盈亏”展示）
    await loadRunningSessionNetPnl(true);

    // 加载运行中机器人的实时数据（首次加载，使用后端最新状态）
    await loadRealtimeData(true);

    // WS订阅：订阅当前运行中机器人
    updateWsSubscription();
    updateWsPositionsSubscription();
    
    // 初始加载持仓数据
    await loadPositionData();
    
    // 加载算力余额
    await loadWalletPower();
    
    // 加载所有机器人的方向预警日志（executionLogs 改为按需加载）
    for (const robot of robotList.value) {
      loadSignalLogs(robot.id);
      // 初始化折叠状态，避免 :expanded-names 传入 undefined 导致受控组件永远折叠
      if (executionExpandedNames[robot.id] === undefined) {
        executionExpandedNames[robot.id] = [];
      }
      if (executionOnlyFailed[robot.id] === undefined) {
        executionOnlyFailed[robot.id] = false;
      }
    // 初始化“市场状态与风险偏好映射”折叠状态 + 默认映射（展开时会自动从后端加载覆盖）
    if (riskMappingExpandedNames[robot.id] === undefined) {
      riskMappingExpandedNames[robot.id] = [];
    }
    if (!marketRiskMappingForm[robot.id]) {
      marketRiskMappingForm[robot.id] = {
        trend: 'balanced',
        volatile: 'balanced',
        high_vol: 'aggressive',
        low_vol: 'conservative',
      };
    }
    }

    // 预取最近几条订单日志：让列表页能直接显示“订单日志数量/最近失败”
    // 不影响展开后查看更多日志（展开会再拉一次更多）
    preloadExecutionLogs();
  } catch (error) {
    console.error('加载机器人列表失败:', error);
  } finally {
    loading.value = false;
  }
};

// 获取运行中区间数据（盈亏、手续费、成交笔数、净盈亏）：按 robotId 映射
const loadRunningSessionNetPnl = async (force: boolean = false) => {
  const now = Date.now();
  if (!force && runningSessionNetPnlLastLoadedAt > 0 && now-runningSessionNetPnlLastLoadedAt < 30_000) {
    return;
  }
  const runningRobots = robotList.value.filter((r: any) => r.status === 2);
  if (runningRobots.length === 0) {
    runningSessionNetPnlMap.value = {};
    runningSessionPnlMap.value = {};
    runningSessionFeeMap.value = {};
    runningSessionTradeCountMap.value = {};
    runningSessionNetPnlLastLoadedAt = now;
    return;
  }
  runningSessionNetPnlLastLoadedAt = now;

  try {
    const pageSize = Math.min(200, Math.max(20, runningRobots.length * 2));
    const res: any = await ToogoWalletApi.runSessionSummary({
      page: 1,
      pageSize,
      isRunning: 1,
    });
    const list: any[] = res?.list || [];
    const map: Record<number, { net: number; pnl: number; fee: number; tradeCount: number; id: number }> = {};
    for (const row of list) {
      const rid = Number(row?.robotId ?? row?.robot_id ?? 0);
      if (!rid) continue;
      const rowId = Number(row?.id ?? 0);
      const pnl = Number(row?.totalPnl ?? 0);
      const fee = Number(row?.totalFee ?? 0);
      const tradeCount = Number(row?.tradeCount ?? 0);
      const net = Number(row?.netPnl ?? row?.net_pnl ?? (pnl - fee));
      const existing = map[rid];
      // 取最新的一条（id 大）
      if (!existing || rowId > existing.id) {
        map[rid] = { net, pnl, fee, tradeCount, id: rowId };
      }
    }
    const outNet: Record<number, number> = {};
    const outPnl: Record<number, number> = {};
    const outFee: Record<number, number> = {};
    const outTradeCount: Record<number, number> = {};
    for (const [k, v] of Object.entries(map)) {
      const key = Number(k);
      outNet[key] = Number(v.net) || 0;
      outPnl[key] = Number(v.pnl) || 0;
      outFee[key] = Number(v.fee) || 0;
      outTradeCount[key] = Number(v.tradeCount) || 0;
    }
    runningSessionNetPnlMap.value = outNet;
    runningSessionPnlMap.value = outPnl;
    runningSessionFeeMap.value = outFee;
    runningSessionTradeCountMap.value = outTradeCount;
  } catch (e) {
    // 失败不影响主页面：保留旧值
    console.warn('[runSessionNetPnl] load failed:', e);
  }
};

const getRunningSessionNetPnl = (robotId: number): number | null => {
  if (!robotId) return null;
  const v = runningSessionNetPnlMap.value[robotId];
  if (typeof v !== 'number' || Number.isNaN(v)) return null;
  return v;
};

const getRunningSessionPnl = (robotId: number): number | null => {
  if (!robotId) return null;
  const v = runningSessionPnlMap.value[robotId];
  if (typeof v !== 'number' || Number.isNaN(v)) return null;
  return v;
};

const getRunningSessionFee = (robotId: number): number | null => {
  if (!robotId) return null;
  const v = runningSessionFeeMap.value[robotId];
  if (typeof v !== 'number' || Number.isNaN(v)) return null;
  return v;
};

const getRunningSessionTradeCount = (robotId: number): number | null => {
  if (!robotId) return null;
  const v = runningSessionTradeCountMap.value[robotId];
  if (typeof v !== 'number' || Number.isNaN(v)) return null;
  return v;
};

// 顶部统计：今日净盈亏（成交流水-今日） + 累计净盈亏（运行区间-全量）
const loadNetPnlSummary = async () => {
  try {
    // 获取今天的日期（开始时间）
    const today = new Date();
    today.setHours(0, 0, 0, 0);
    const todayStart = today.toISOString().slice(0, 19).replace('T', ' ');
    
    // 今日净盈亏：数据源=钱包-交易明细-成交流水-今日净盈亏(扣手续费)
    const todayRes = await ToogoWalletApi.tradeHistory({
      startTime: todayStart,
      page: 1,
      pageSize: 1 // 只需要统计数据，不需要列表
    });
    if (todayRes && todayRes.summary) {
      todayNetPnl.value = Number(todayRes.summary.totalNetPnl) || 0;
    }
    
    // 累计净盈亏：数据源=钱包-交易明细-运行区间-净盈亏(扣手续费)（全量汇总）
    const totalRes = await ToogoWalletApi.runSessionSummary({
      page: 1,
      pageSize: 1 // 只需要统计数据，不需要列表
    });
    if (totalRes && totalRes.summary) {
      totalNetPnl.value = Number(totalRes.summary.totalNetPnl) || 0;
    }
  } catch (error) {
    console.error('加载净盈亏数据失败:', error);
    // 加载失败时使用默认值
    todayNetPnl.value = 0;
    totalNetPnl.value = 0;
  }
};

// 加载用户算力余额
const loadWalletPower = async () => {
  try {
    // 获取所有不重复的用户ID
    const userIds = [...new Set(robotList.value.map((r: any) => r.userId))];
    
    for (const userId of userIds) {
      try {
        const res = await http.request({
          url: '/toogo/wallet/overview',
          method: 'get',
          params: { userId },
        });
        if (res && res.totalPower !== undefined) {
          walletPowerMap.value[userId] = res.totalPower;
        }
      } catch (e) {
        console.error('加载用户算力失败:', userId, e);
        // 单个用户加载失败不影响其他
      }
    }
  } catch (error) {
    console.error('加载算力余额失败:', error);
  }
};

// 记录上次信号方向，用于检测方向变化
const lastSignalDirection = ref<Record<number, string>>({});

// ============ WebSocket：机器人实时分析推送（替代高频轮询）============
const wsSubscribedRobotIds = ref<string>(''); // 逗号分隔
const wsOnRealtimePush = (message: WebSocketMessage) => {
  const payload = message.data;
  const list = payload?.list || [];
  applyBatchRobotAnalysisList(list, false);
};

const updateWsSubscription = () => {
  const runningRobots = robotList.value.filter((r: any) => r.status === 2);
  const robotIds = runningRobots.map((r: any) => r.id).join(',');
  if (!robotIds) return;
  if (robotIds === wsSubscribedRobotIds.value) return;
  wsSubscribedRobotIds.value = robotIds;
  sendMsg(SocketEnum.EventToogoRobotRealtimeSubscribe, { robotIds, intervalMs: 1000 });
};

const unsubscribeWs = () => {
  wsSubscribedRobotIds.value = '';
  sendMsg(SocketEnum.EventToogoRobotRealtimeUnsubscribe, {});
};

// ============ WebSocket：机器人持仓实时推送（positions snapshot）============
const wsPositionsSubscribedRobotIds = ref<string>(''); // 逗号分隔
const wsOnPositionsPush = (message: WebSocketMessage) => {
  const payload = message.data;
  const list = payload?.list || [];
  if (!list || list.length === 0) return;

  for (const item of list) {
    const robotId = item?.robotId;
    if (!robotId) continue;
    const robot = robotList.value.find((r: any) => r.id === robotId);
    if (!robot) continue;
    if (item?.error) {
      console.warn(`[WS][positions] robotId=${robotId} 获取持仓失败:`, item.error);
    }
    // 重要：错误/限流/超时不等于“无持仓”，避免推空导致闪烁/丢失
    const positions = item?.list || [];
    const isStale = !!item?.stale;
    if (item?.error && isStale && (!positions || positions.length === 0)) {
      // 失败且没有可复用快照：忽略本帧（保留前端已有仓位）
      continue;
    }
    applyRobotPositionsSnapshot(robot, positions, isStale ? 'WS推送(stale)' : 'WS推送');
  }
};

const updateWsPositionsSubscription = () => {
  // 注意：交易所可能存在“机器人未运行但仍有残留仓位/挂单”的情况（尤其重启后）。
  // 为了避免页面看不到风险敞口，这里订阅列表中的全部机器人；当数量较多时自动降频，避免打爆交易所。
  const robots = robotList.value || [];
  const robotIds = robots.map((r: any) => r.id).join(',');
  if (!robotIds) return;
  if (robotIds === wsPositionsSubscribedRobotIds.value) return;
  wsPositionsSubscribedRobotIds.value = robotIds;

  // 小规模：500ms 更实时；大规模：2000ms 降频保护
  const intervalMs = robots.length <= 5 ? 500 : 2000;
  console.log('[WS][positions] subscribe', { robotIds, intervalMs });
  sendMsg(SocketEnum.EventToogoRobotPositionsSubscribe, { robotIds, intervalMs });
};

const unsubscribeWsPositions = () => {
  wsPositionsSubscribedRobotIds.value = '';
  sendMsg(SocketEnum.EventToogoRobotPositionsUnsubscribe, {});
};

// ============ WebSocket：机器人挂单实时推送（open orders snapshot）============
// 说明：
// - 详情弹窗展示“当前挂单”时，优先用 WS 快照/增量更新，避免 60s REST 刷新导致闪烁
// - REST 仍可作为首次加载/断线兜底
const wsOrdersSubscribedRobotIds = ref<string>(''); // 逗号分隔（当前实现=详情弹窗只订阅当前机器人）
const applyRobotOrdersSnapshot = (robotId: number, orders: any[], _source: string) => {
  if (!showDetailModal.value || !currentRobot.value) return;
  if (currentRobot.value.id !== robotId) return;
  currentOpenOrders.value = orders || [];
  // console.debug(`[WS][orders] apply snapshot robotId=${robotId} source=${source} count=${(orders || []).length}`);
};

const wsOnOrdersPush = (message: WebSocketMessage) => {
  const payload = message.data;
  const list = payload?.list || [];
  if (!list || list.length === 0) return;

  for (const item of list) {
    const robotId = item?.robotId;
    if (!robotId) continue;
    const orders = item?.list || [];
    const isStale = !!item?.stale;
    if (item?.error) {
      console.warn(`[WS][orders] robotId=${robotId} 获取挂单失败:`, item.error);
      // 失败且没有可复用快照：忽略本帧，避免清空
      if (isStale && (!orders || orders.length === 0)) continue;
    }
    applyRobotOrdersSnapshot(robotId, orders, isStale ? 'WS推送(stale)' : 'WS推送');
  }
};

// 交易所私有WS触发的挂单增量（无需订阅）
const wsOnOrdersDelta = (message: WebSocketMessage) => {
  const data = message.data || {};
  const robotId = data?.robotId;
  const delta = data?.list || [];
  if (!robotId || !Array.isArray(delta) || delta.length === 0) return;
  if (!showDetailModal.value || !currentRobot.value || currentRobot.value.id !== robotId) return;

  // 增量合并：isOpen=false 则移除，其他 upsert
  const cur = Array.isArray(currentOpenOrders.value) ? [...currentOpenOrders.value] : [];
  const idxById = new Map<string, number>();
  for (let i = 0; i < cur.length; i++) {
    const id = cur[i]?.orderId;
    if (id) idxById.set(String(id), i);
  }
  for (const d of delta) {
    const id = d?.orderId;
    if (!id) continue;
    const key = String(id);
    const isOpen = d?.isOpen;
    if (isOpen === false) {
      const idx = idxById.get(key);
      if (idx !== undefined) {
        cur.splice(idx, 1);
        // 重新构建索引（订单数量通常很小，简单做即可）
        idxById.clear();
        for (let i = 0; i < cur.length; i++) {
          const oid = cur[i]?.orderId;
          if (oid) idxById.set(String(oid), i);
        }
      }
      continue;
    }
    const idx = idxById.get(key);
    if (idx === undefined) {
      cur.unshift(d);
      // 更新索引（同上，重建）
      idxById.clear();
      for (let i = 0; i < cur.length; i++) {
        const oid = cur[i]?.orderId;
        if (oid) idxById.set(String(oid), i);
      }
    } else {
      cur[idx] = { ...(cur[idx] || {}), ...(d || {}) };
    }
  }
  currentOpenOrders.value = cur;
};

const updateWsOrdersSubscription = () => {
  // 弹窗关闭/无选中机器人：自动退订，避免后台持续推送
  if (!showDetailModal.value || !currentRobot.value) {
    if (wsOrdersSubscribedRobotIds.value) {
      unsubscribeWsOrders();
    }
    return;
  }
  const robotIds = String(currentRobot.value.id);
  if (!robotIds) return;
  if (robotIds === wsOrdersSubscribedRobotIds.value) return;
  wsOrdersSubscribedRobotIds.value = robotIds;
  // 详情弹窗更需要实时：500ms
  sendMsg(SocketEnum.EventToogoRobotOrdersSubscribe, { robotIds, intervalMs: 500 });
};

const unsubscribeWsOrders = () => {
  wsOrdersSubscribedRobotIds.value = '';
  sendMsg(SocketEnum.EventToogoRobotOrdersUnsubscribe, {});
};

// 交易所私有WS触发的“持仓即时刷新”（无需订阅）
const wsOnPositionsDelta = (message: WebSocketMessage) => {
  const data = message.data || {};
  const robotId = data?.robotId;
  if (!robotId) return;
  const robot = robotList.value.find((r: any) => r.id === robotId);
  if (!robot) return;
  if (data?.error && data?.stale && (!data?.list || data?.list.length === 0)) {
    // 不用错误空数组覆盖 UI
    return;
  }
  applyRobotPositionsSnapshot(robot, data?.list || [], data?.stale ? 'WS增量(stale)' : 'WS增量');
};

// ============ WebSocket：交易关键事件（平仓成功/订单状态变更等）============
const wsOnTradeEvent = (message: WebSocketMessage) => {
  const data = message.data || {};
  const type = data?.type;
  const robotId = data?.robotId;

  // 平仓成功（手动/自动）：立刻从机器人列表持仓里移除（不等交易所/同步延迟）
  if (type === 'close_success' && robotId) {
    const symbol = data?.symbol;
    const positionSide = data?.positionSide;
    if (symbol && positionSide) {
      const closeKey = `${robotId}_${symbol}_${positionSide}`;
      recentManualCloseAt.value[closeKey] = Date.now(); // 复用抑制窗口，防止旧快照把已平仓仓位“渲染回来”
    }

    // 解除“新开仓保护窗”：平仓后应允许空快照立即清空
    positionOpenProtectionUntil.value[robotId] = 0;
    positionEmptyStreak.value[robotId] = 0;

    // 立即移除该仓位（如果能定位到 symbol/side）
    if (symbol && positionSide) {
      const before = positionData.value[robotId] || [];
      positionData.value[robotId] = before.filter((p: any) => !(p.symbol === symbol && p.positionSide === positionSide));

      // 若该机器人已无持仓，清理本地开关/最高盈利缓存
      if ((positionData.value[robotId] || []).length === 0) {
        const keysToDelete = Object.keys(takeProfitRetreatSwitch.value).filter(key => key.startsWith(`${robotId}_`));
        keysToDelete.forEach(key => delete takeProfitRetreatSwitch.value[key]);

        const maxProfitCache = getMaxProfitCache();
        const profitKeysToDelete = Object.keys(maxProfitCache).filter(key => key.startsWith(`${robotId}_`));
        profitKeysToDelete.forEach(key => delete maxProfitCache[key]);
        if (profitKeysToDelete.length > 0) {
          saveMaxProfitCache(maxProfitCache);
        }
      }
    }
  }

  // 平仓成功：如果当前正在查看详情弹窗，则立即刷新订单相关数据（挂单/成交明细）
  // 改为 WS 优先：挂单列表靠 orders/delta + orders/push 自动更新；成交明细仍由数据库更新（可按需刷新）
  if ((type === 'close_success' || type === 'order_delta') && robotId && currentRobot.value?.id === robotId && showDetailModal.value) {
    updateWsOrdersSubscription();
  }
};

// 持仓空结果防抖：连续多次为空才认为已平仓（避免后端偶发空数据导致闪烁）
const positionEmptyStreak = ref<Record<number, number>>({});
// 最近一次“非空持仓快照”的时间戳：用于抑制 OKX/网络抖动导致的瞬时空结果清空
const positionLastNonEmptyAt = ref<Record<number, number>>({});
// 新开仓保护窗：用于彻底抑制“新开仓首分钟空快照导致的消失/再出现”
// - key=robotId -> timestamp(ms)
const positionOpenProtectionUntil = ref<Record<number, number>>({});
// 单仓位缺失计数：用于 Gate 双向持仓时“偶发少一侧快照”导致的行消失/再出现闪烁
// key = `${robotId}_${symbol}_${positionSide}` => streak
const positionMissingStreak = ref<Record<string, number>>({});
// 单仓位缺失开始时间：用于“按时间保留”策略，避免 Gate 偶发缺一侧时行消失约1分钟又回来
// key = `${robotId}_${symbol}_${positionSide}` => firstMissingAt(ms)
const positionMissingSince = ref<Record<string, number>>({});
// Gate 双向持仓“缺一侧快照”主动修复：缺失时对该机器人立即补拉一次 positions，而不是长时间保留/按轮询次数等待
const positionRepairInFlight = ref<Record<number, boolean>>({});
const positionRepairLastAt = ref<Record<number, number>>({});

const repairRobotPositionsOnce = async (robot: any, reason: string) => {
  if (!robot?.id) return;
  const now = Date.now();
  if (positionRepairInFlight.value[robot.id]) return;
  const lastAt = positionRepairLastAt.value[robot.id] || 0;
  // 冷却：避免 Gate 偶发抖动导致接口风暴
  if (now - lastAt < 8000) return;
  positionRepairLastAt.value[robot.id] = now;
  positionRepairInFlight.value[robot.id] = true;
  try {
    const posRes = await ToogoRobotApi.positions({ robotId: robot.id });
    applyRobotPositionsSnapshot(robot, posRes?.list || [], `补拉(${reason})`);
  } catch {
    // ignore
  } finally {
    positionRepairInFlight.value[robot.id] = false;
  }
};

// 手动平仓后的“短暂抑制窗口”：避免交易所/缓存短时间仍返回旧持仓，导致页面又把已平仓的单子渲染回来
// key = `${robotId}_${symbol}_${positionSide}` => timestamp(ms)
const recentManualCloseAt = ref<Record<string, number>>({});

// 统一合并 batchRobotAnalysis 的结果（HTTP轮询 & WS推送共用）
const applyBatchRobotAnalysisList = (list: any[], isInitialLoad: boolean = false) => {
  if (!list || list.length === 0) return;

  for (const item of list) {
    const robotId = item.robotId;

    // 更新连接状态
    robotStatusData.value[robotId] = {
      connected: item.connected,
      connectionError: item.connectionError,
      lastUpdate: new Date().toISOString(),
    };

    if (item.connected && item.ticker) {
      // 更新实时行情数据
      tickerData.value[robotId] = {
        symbol: item.ticker.symbol,
        lastPrice: item.ticker.lastPrice,
        markPrice: item.ticker.markPrice,
        high24h: item.ticker.high24h,
        low24h: item.ticker.low24h,
        volume24h: item.ticker.volume24h,
        change24h: item.ticker.change24h,
        changePercent: item.ticker.changePercent,
      };

      // 检测信号方向变化，有方向信号时刷新日志（executionLogs 改为按需加载）
      const currentDirection = item.signal?.direction?.toUpperCase();
      const prevDirection = lastSignalDirection.value[robotId];
      if (currentDirection && (currentDirection === 'LONG' || currentDirection === 'SHORT')) {
        if (currentDirection !== prevDirection) {
          loadSignalLogs(robotId);
        }
      }
      lastSignalDirection.value[robotId] = currentDirection || '';

      // 更新分析数据（首次加载时以服务端为准；非首次保留用户刚改的开关状态）
      const existingConfig = analysisData.value[robotId]?.config;
      const existingAccount = analysisData.value[robotId]?.account;
      const newConfig = { ...item.config };
      if (!isInitialLoad && existingConfig) {
        if (existingConfig.autoTradeEnabled !== undefined && existingConfig.autoTradeEnabled !== item.config?.autoTradeEnabled) {
          newConfig.autoTradeEnabled = existingConfig.autoTradeEnabled;
        }
        if (existingConfig.autoCloseEnabled !== undefined && existingConfig.autoCloseEnabled !== item.config?.autoCloseEnabled) {
          newConfig.autoCloseEnabled = existingConfig.autoCloseEnabled;
        }
        if (existingConfig.profitLockEnabled !== undefined && existingConfig.profitLockEnabled !== item.config?.profitLockEnabled) {
          newConfig.profitLockEnabled = existingConfig.profitLockEnabled;
        }
      }

      // 【防闪烁】账户数据保护：如果新数据的账户权益为0或无效，但已有有效旧数据，则保留旧数据
      let newAccount = item.account;
      const newEquity = item.account?.accountEquity ?? item.account?.totalBalance ?? 0;
      const oldEquity = existingAccount?.accountEquity ?? existingAccount?.totalBalance ?? 0;
      if (newEquity === 0 && oldEquity > 0) {
        // 新数据权益为0但旧数据有值，保留旧数据避免闪烁
        // console.debug(`[防闪烁] robotId=${robotId} 账户权益为0，保留旧数据: ${oldEquity}`);
        newAccount = existingAccount;
      }

      analysisData.value[robotId] = {
        market: item.market,
        marketStateRealtime: item.marketStateRealtime,
        risk: item.risk,
        signal: item.signal,
        account: newAccount,
        config: newConfig,
        priceWindow: item.priceWindow,
        signalHistory: item.signalHistory,
        lastUpdate: Date.now(),
      };

      // 【重要】根据账户保证金状态判断是否已平仓（使用防抖机制，避免闪烁）
      // 注意：如果后端未返回 usedMargin 字段，不能用 `|| 0` 推断为0，否则会导致持仓被误清空并闪烁
      const usedMargin = item?.account?.usedMargin;
      const hasLocalPosition = positionData.value[robotId]?.length > 0;
      // ===== 新开仓保护窗 =====
      // 后端 usedMargin>0 往往比 positions 列表更早出现；此时可能还拿不到持仓快照，容易出现“先出现后消失”的闪烁
      // 保护策略：只要检测到 usedMargin>0，就开启一段保护窗，在窗口内忽略空快照清空逻辑
      if (usedMargin !== undefined && usedMargin !== null && usedMargin > 0) {
        const protectMs = 15_000;
        const until = Date.now() + protectMs;
        positionOpenProtectionUntil.value[robotId] = Math.max(positionOpenProtectionUntil.value[robotId] || 0, until);
      }
      if (usedMargin !== undefined && usedMargin !== null && usedMargin === 0 && hasLocalPosition) {
        // OKX/交易所同步可能短暂返回 usedMargin=0（尤其开仓/重连窗口），若最近出现过非空持仓则不清空
        const lastNonEmpty = positionLastNonEmptyAt.value[robotId] || 0;
        const protectMs = 10_000;
        if (lastNonEmpty > 0 && Date.now() - lastNonEmpty < protectMs) {
          // 只计数但不触发清空
          positionEmptyStreak.value[robotId] = (positionEmptyStreak.value[robotId] || 0) + 1;
        } else {
        // 【防闪烁】使用防抖机制：累加空结果计数，连续3次为空才真正清空
        const streak = (positionEmptyStreak.value[robotId] || 0) + 1;
        positionEmptyStreak.value[robotId] = streak;
        if (streak >= 3) {
          console.log(`[WS推送] robotId=${robotId} 连续${streak}次无持仓，清空持仓列表`);
          positionData.value[robotId] = [];
          const keysToDelete = Object.keys(takeProfitRetreatSwitch.value).filter(key => key.startsWith(`${robotId}_`));
          keysToDelete.forEach(key => delete takeProfitRetreatSwitch.value[key]);
        }
        }
      } else if (hasLocalPosition) {
        // 有持仓时重置空结果计数
        positionEmptyStreak.value[robotId] = 0;
      }
    } else {
      // 未连接：仍保存推送里的 marketStateRealtime（若有）
      if (analysisData.value[robotId]) {
        analysisData.value[robotId].marketStateRealtime = item.marketStateRealtime;
      }
    }
  }
};

// 加载实时数据（轻量化，每秒刷新）
const loadRealtimeData = async (isInitialLoad: boolean = false) => {
  const runningRobots = robotList.value.filter((r: any) => r.status === 2);
  if (runningRobots.length === 0) return;

  // 批量获取机器人分析数据（包含实时价格、信号、市场分析等）
  const robotIds = runningRobots.map((r: any) => r.id).join(',');
  
  try {
    const batchRes = await ToogoExchangeApi.batchRobotAnalysis({ robotIds });
    if (batchRes?.list) {
      for (const item of batchRes.list) {
        const robotId = item.robotId;
        
        // 更新连接状态
        robotStatusData.value[robotId] = {
          connected: item.connected,
          connectionError: item.connectionError,
          lastUpdate: new Date().toISOString()
        };
        
        if (item.connected && item.ticker) {
          // 更新实时行情数据
          tickerData.value[robotId] = {
            symbol: item.ticker.symbol,
            lastPrice: item.ticker.lastPrice,
            markPrice: item.ticker.markPrice,
            high24h: item.ticker.high24h,
            low24h: item.ticker.low24h,
            volume24h: item.ticker.volume24h,
            change24h: item.ticker.change24h,
            changePercent: item.ticker.changePercent,
          };
          
          // 检测信号方向变化，有方向信号时刷新日志（executionLogs 改为按需加载）
          const currentDirection = item.signal?.direction?.toUpperCase();
          const prevDirection = lastSignalDirection.value[robotId];
          if (currentDirection && (currentDirection === 'LONG' || currentDirection === 'SHORT')) {
            // 有方向信号，且方向变化时刷新日志
            if (currentDirection !== prevDirection) {
              loadSignalLogs(robotId);
            }
          }
          lastSignalDirection.value[robotId] = currentDirection || '';
          
          // 更新分析数据（包含价格窗口用于实时图表）
          // 【重要】首次加载时使用后端最新状态，实时刷新时保留用户刚修改的状态
          const existingConfig = analysisData.value[robotId]?.config;
          const newConfig = { ...item.config };
          
          // 只有在非首次加载时，才保留用户修改的开关状态
          // 首次加载时（页面刷新），应该使用后端返回的最新状态
          if (!isInitialLoad && existingConfig) {
            // 保留用户修改的开关状态（如果存在且与后端不同，说明用户刚修改过）
            if (existingConfig.autoTradeEnabled !== undefined && 
                existingConfig.autoTradeEnabled !== item.config?.autoTradeEnabled) {
              newConfig.autoTradeEnabled = existingConfig.autoTradeEnabled;
            }
            if (existingConfig.autoCloseEnabled !== undefined && 
                existingConfig.autoCloseEnabled !== item.config?.autoCloseEnabled) {
              newConfig.autoCloseEnabled = existingConfig.autoCloseEnabled;
            }
            if (existingConfig.profitLockEnabled !== undefined &&
                existingConfig.profitLockEnabled !== item.config?.profitLockEnabled) {
              newConfig.profitLockEnabled = existingConfig.profitLockEnabled;
            }
          }
          
          analysisData.value[robotId] = {
            market: item.market,
            marketStateRealtime: item.marketStateRealtime,
            risk: item.risk,
            signal: item.signal,
            account: item.account,
            config: newConfig,
            priceWindow: item.priceWindow,
            signalHistory: item.signalHistory,
            lastUpdate: Date.now(), // 记录更新时间
          };
          
          // 【重要】根据账户保证金状态判断是否已平仓（使用防抖机制，避免闪烁）
          // 注意：如果后端未返回 usedMargin 字段，不能用 `|| 0` 推断为0，否则会导致持仓被误清空并闪烁
          const usedMargin = item?.account?.usedMargin;
          const hasLocalPosition = positionData.value[robotId]?.length > 0;
          
          if (usedMargin !== undefined && usedMargin !== null && usedMargin === 0 && hasLocalPosition) {
            // 【防闪烁】使用防抖机制：累加空结果计数，连续3次为空才真正清空
            const protectMs = 10_000;
            const lastNonEmpty = positionLastNonEmptyAt.value[robotId] || 0;
            const openProtectUntil = positionOpenProtectionUntil.value[robotId] || 0;
            if (openProtectUntil > 0 && Date.now() < openProtectUntil) {
              // 新开仓保护窗内：忽略空快照，不清空也不计数
              continue;
            }
            if (lastNonEmpty > 0 && Date.now() - lastNonEmpty < protectMs) {
              positionEmptyStreak.value[robotId] = (positionEmptyStreak.value[robotId] || 0) + 1;
              continue;
            }
            const streak = (positionEmptyStreak.value[robotId] || 0) + 1;
            positionEmptyStreak.value[robotId] = streak;
            if (streak >= 3) {
              console.log(`[HTTP轮询] robotId=${robotId} 连续${streak}次无持仓，清空持仓列表`);
              positionData.value[robotId] = [];
              const keysToDelete = Object.keys(takeProfitRetreatSwitch.value).filter(key => key.startsWith(`${robotId}_`));
              keysToDelete.forEach(key => delete takeProfitRetreatSwitch.value[key]);
            }
          } else if (usedMargin !== undefined && usedMargin !== null && usedMargin > 0) {
            // 有持仓时重置空结果计数
            positionEmptyStreak.value[robotId] = 0;
            if (!hasLocalPosition) {
              // 后端显示有持仓，但前端没有数据 → 需要加载持仓
              console.log(`[HTTP轮询] robotId=${robotId} 检测到新开仓，触发持仓加载`);
              loadPositionData();
            }
          }
          
          // 【重要】持仓盈亏/最高盈利/止盈止损进度统一以“后端返回”为准（OKX 的持仓数量口径可能是合约张数，前端用 positionAmt 自算会偏差）。
          // 这里仅用实时行情更新 markPrice（用于展示），避免覆盖后端的 unrealizedPnl/maxProfitReached 导致前后端口径不一致。
          // 【防闪烁优化】使用原地更新而非 .map() 创建新数组，避免触发 Vue 完整重渲染
          // 永续合约展示：优先用 markPrice（更接近交易所口径），缺失时再用 lastPrice
          const currentPrice = (item.ticker?.markPrice || item.ticker?.lastPrice);
          const positions = positionData.value[robotId];
          if (currentPrice && positions?.length > 0) {
            // 【原地更新】直接修改数组中的对象属性，保持引用不变
            for (let i = 0; i < positions.length; i++) {
              const pos = positions[i];

              // 【原地更新】仅更新 markPrice（展示用）
              pos.markPrice = currentPrice;
            }
          }
        }
      }
    }
  } catch (error: any) {
    // 静默失败，避免控制台刷屏
    if (Math.random() < 0.1) {  // 只打印10%的错误
      console.warn('实时数据刷新失败:', error.message);
    }
  }
};

// 获取订阅信息（仅用于显示倒计时）
const loadSubscriptionInfo = async () => {
  try {
    const res = await ToogoSubscriptionApi.mySubscription();
    subscriptionInfo.value = res || {};
  } catch (error) {
    console.error('加载订阅信息失败:', error);
  }
};

// 获取订阅到期倒计时（仅显示）
const getSubscriptionCountdown = () => {
  if (!subscriptionInfo.value.planExpireTime) return '--';
  const expireTime = new Date(subscriptionInfo.value.planExpireTime);
  const now = new Date();
  const diff = expireTime.getTime() - now.getTime();
  
  if (diff <= 0) return '已到期';
  
  const days = Math.floor(diff / (1000 * 60 * 60 * 24));
  const hours = Math.floor((diff % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60));
  const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));
  
  if (days > 0) {
    return `${days}天${hours}小时`;
  } else if (hours > 0) {
    return `${hours}小时${minutes}分钟`;
  } else {
    return `${minutes}分钟`;
  }
};

// 获取订阅倒计时标签类型（仅显示）
const getSubscriptionCountdownType = () => {
  if (!subscriptionInfo.value.planExpireTime) return 'default';
  const expireTime = new Date(subscriptionInfo.value.planExpireTime);
  const now = new Date();
  const diff = expireTime.getTime() - now.getTime();
  const hours = diff / (1000 * 60 * 60);
  
  if (diff <= 0) return 'error';
  if (hours <= 24) return 'warning';
  return 'success';
};

// 启动倒计时更新定时器（仅用于显示）
const startCountdownTimer = () => {
  // 每60秒刷新一次订阅信息，确保倒计时准确
  subscriptionCountdownTimer.value = setInterval(async () => {
    await loadSubscriptionInfo();
  }, 60000);
};

const startRobot = async (robot: any) => {
  try {
    await ToogoRobotApi.start({ id: robot.id });
    message.success(robot?.status === 3 ? '机器人已重启' : '机器人已启动');
    loadData();
  } catch (error: any) {
    message.error(error.message || '启动失败');
  }
};

const stopRobot = async (robot: any) => {
  // 检查该机器人是否有未平仓订单
  const positions = positionData.value[robot.id] || [];
  const hasOpenPositions = positions.length > 0 && positions.some((p: any) => parseFloat(p.positionAmt) !== 0);
  
  const contentMsg = hasOpenPositions 
    ? `确定要停止机器人 "${robot.robotName}" 吗？检测到有 ${positions.length} 个未平仓订单，需要手动处理！`
    : `确定要停止机器人 "${robot.robotName}" 吗？`;
  
  dialog.warning({
    title: '确认停止',
    content: contentMsg,
    positiveText: '确定',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await ToogoRobotApi.stop({ id: robot.id });
        message.success('机器人已停止');
        loadData();
      } catch (error: any) {
        message.error(error.message || '停止失败');
      }
    },
  });
};

// 快速切换自动下单开关
const toggleAutoTrade = async (robot: any, newValue: boolean) => {
  const newStatus = newValue ? 1 : 0;
  // 立即更新UI状态（乐观更新）
  const oldValue = analysisData.value[robot.id]?.config?.autoTradeEnabled;
  if (!analysisData.value[robot.id]) {
    analysisData.value[robot.id] = {};
  }
  if (!analysisData.value[robot.id].config) {
    analysisData.value[robot.id].config = {};
  }
  analysisData.value[robot.id].config.autoTradeEnabled = newValue;
  
  try {
    await ToogoRobotApi.update({
      id: robot.id,
      autoTradeEnabled: newStatus,
    });
    message.success(newValue ? '已开启全自动下单' : '已关闭全自动下单');
    // 不立即刷新实时数据，避免覆盖用户修改的状态
    // 等待下一次自动刷新时会从后端获取最新状态
  } catch (error: any) {
    // 如果失败，恢复开关状态
    if (analysisData.value[robot.id]?.config) {
      analysisData.value[robot.id].config.autoTradeEnabled = oldValue;
    }
    message.error(error.message || '切换失败');
  }
};

// 获取启动止盈回撤开关状态
// 【关键】以后端状态为准，确保前后端一致
const getTakeProfitRetreatSwitch = (robotId: number, symbol: string, positionSide: string, pos?: any): boolean => {
  // 【完全由后端控制】只使用后端返回的开关状态，不再使用本地状态
  // 如果后端没有返回数据，返回false（表示未启动）
  if (pos !== undefined && pos !== null) {
    return pos.takeProfitEnabled || false;
  }
  // 后端状态不可用时，返回false（默认未启动）
  return false;
};
  
// 快速切换自动平仓开关
const toggleAutoClose = async (robot: any, newValue: boolean) => {
  const newStatus = newValue ? 1 : 0;
  // 立即更新UI状态（乐观更新）
  const oldValue = analysisData.value[robot.id]?.config?.autoCloseEnabled;
  if (!analysisData.value[robot.id]) {
    analysisData.value[robot.id] = {};
  }
  if (!analysisData.value[robot.id].config) {
    analysisData.value[robot.id].config = {};
  }
  analysisData.value[robot.id].config.autoCloseEnabled = newValue;
  
  try {
    await ToogoRobotApi.update({
      id: robot.id,
      autoCloseEnabled: newStatus,
    });
    message.success(newValue ? '已开启全自动平仓' : '已关闭全自动平仓');
    // 不立即刷新实时数据，避免覆盖用户修改的状态
    // 等待下一次自动刷新时会从后端获取最新状态
  } catch (error: any) {
    // 如果失败，恢复开关状态
    if (analysisData.value[robot.id]?.config) {
      analysisData.value[robot.id].config.autoCloseEnabled = oldValue;
    }
    message.error(error.message || '切换失败');
  }
};

// 锁定盈利开关：止盈开关已启动时禁止自动开新仓（后端执行判断）
const toggleProfitLock = async (robot: any, newValue: boolean) => {
  const newStatus = newValue ? 1 : 0;
  const oldValue = analysisData.value[robot.id]?.config?.profitLockEnabled;
  if (!analysisData.value[robot.id]) {
    analysisData.value[robot.id] = {};
  }
  if (!analysisData.value[robot.id].config) {
    analysisData.value[robot.id].config = {};
  }
  analysisData.value[robot.id].config.profitLockEnabled = newValue;

  try {
    await ToogoRobotApi.update({
      id: robot.id,
      profitLockEnabled: newStatus,
    });
    message.success(newValue ? '已开启锁定盈利' : '已关闭锁定盈利');
  } catch (error: any) {
    if (analysisData.value[robot.id]?.config) {
      analysisData.value[robot.id].config.profitLockEnabled = oldValue;
    }
    message.error(error.message || '切换失败');
  }
};

// 快速切换双向开单开关
const toggleDualSidePosition = async (robot: any, newValue: boolean) => {
  const newStatus = newValue ? 1 : 0;
  // 立即更新UI状态（乐观更新）
  const oldValue = analysisData.value[robot.id]?.config?.dualSidePosition;
  if (!analysisData.value[robot.id]) {
    analysisData.value[robot.id] = {};
  }
  if (!analysisData.value[robot.id].config) {
    analysisData.value[robot.id].config = {};
  }
  analysisData.value[robot.id].config.dualSidePosition = newValue;
  
  try {
    await ToogoRobotApi.update({
      id: robot.id,
      dualSidePosition: newStatus,
    });
    message.success(newValue ? '已开启双向开单（可同时持多单和空单）' : '已关闭双向开单（同时只能有一个持仓）');
  } catch (error: any) {
    // 如果失败，恢复开关状态
    if (analysisData.value[robot.id]?.config) {
      analysisData.value[robot.id].config.dualSidePosition = oldValue;
    }
    message.error(error.message || '切换失败');
  }
};

const deleteRobot = (robot: any) => {
  dialog.error({
    title: '确认删除',
    content: `确定要删除机器人 "${robot.robotName}" 吗？此操作不可恢复！`,
    positiveText: '确定删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await ToogoRobotApi.delete({ id: robot.id });
        message.success('机器人已删除');
        loadData();
      } catch (error: any) {
        message.error(error.message || '删除失败');
      }
    },
  });
};

// 手动平仓 (列表中)
const closePosition = async (robot: any, position: any) => {
  const key = getCloseKey(robot.id, position.symbol, position.positionSide);
  // 若该仓位正在平仓中，直接拦截，避免重复弹框/重复请求
  if (closeInFlight.value[key]) {
    message.warning('平仓请求处理中，请勿重复点击');
    return;
  }
  
  const d = dialog.warning({
    title: '确认平仓',
    content: () => {
      const p = closeProgress.value[key] ?? 0;
      return h('div', { style: 'display:flex;flex-direction:column;gap:10px;' }, [
        h('div', null, `确定要平仓 ${position.positionSide === 'LONG' ? '多' : '空'} ${Math.abs(position.positionAmt).toFixed(4)} 吗？`),
        closeInFlight.value[key]
          ? h('div', { style: 'margin-top:2px;' }, [
              h(NProgress, {
                type: 'line',
                percentage: p,
                height: 10,
                indicatorPlacement: 'inside',
              } as any),
              h('div', { style: 'margin-top:6px;color:#9ca3af;font-size:12px;' }, '请求提交中，请勿重复点击…'),
            ])
          : null,
      ]);
    },
    positiveText: '确定平仓',
    negativeText: '取消',
    onPositiveClick: async () => {
      // 二次防抖：Naive Dialog 在异步期间仍可能被重复触发
      if (closeInFlight.value[key]) {
        return false;
      }
      closeInFlight.value[key] = true;
      d.loading = true; // 让确认按钮进入 loading（Naive UI 内置防重复）

      const stop = startCloseProgress(key);
      try {
        await ToogoRobotApi.closePosition({
          robotId: robot.id,
          symbol: position.symbol,
          positionSide: position.positionSide,
          quantity: Math.abs(position.positionAmt),
        });
        message.success('平仓成功');
        closeProgress.value[key] = 100;

        // ========= 关键优化：立即更新前端持仓视图（不等待下一轮同步/轮询） =========
        recentManualCloseAt.value[key] = Date.now();

        // 立即从当前持仓列表移除该仓位（用户体验：平仓后立刻消失）
        const before = positionData.value[robot.id] || [];
        positionData.value[robot.id] = before.filter(
          (p: any) => !(p.symbol === position.symbol && p.positionSide === position.positionSide)
        );

        // 重置空结果计数（避免“连续3次空结果才清空”的延迟）
        positionEmptyStreak.value[robot.id] = 0;

        // 若该机器人已无持仓，清理开关与最高盈利缓存
        if ((positionData.value[robot.id] || []).length === 0) {
          const keysToDelete = Object.keys(takeProfitRetreatSwitch.value).filter(key => key.startsWith(`${robot.id}_`));
          keysToDelete.forEach(key => delete takeProfitRetreatSwitch.value[key]);

          const maxProfitCache = getMaxProfitCache();
          const profitKeysToDelete = Object.keys(maxProfitCache).filter(key => key.startsWith(`${robot.id}_`));
          profitKeysToDelete.forEach(key => delete maxProfitCache[key]);
          if (profitKeysToDelete.length > 0) {
            saveMaxProfitCache(maxProfitCache);
          }
        }
        
        // 刷新实时数据和持仓数据
        await loadRealtimeData();
        await loadPositionData();
        // 如果打开了详情页面，刷新订单数据
        if (currentRobot.value && showDetailModal.value) {
          await refreshOrderData();
        }
        // 让用户感知“完成”，再关闭弹框，避免“没反应”
        window.setTimeout(() => {
          d.destroy();
        }, 300);
      } catch (error: any) {
        message.error(error.message || '平仓失败');
        // 返回 false：保持弹框不关闭，允许用户再次点击确认重试
        return false;
      } finally {
        stop();
        d.loading = false;
        closeInFlight.value[key] = false;
      }
    },
  });
};

// 手动平仓 (详情弹窗中)
const closePositionInModal = async (position: any) => {
  if (!currentRobot.value) return;
  await closePosition(currentRobot.value, position);
  loadDetailData(currentRobot.value);
};

// 撤销挂单
const cancelOrder = async (orderId: string) => {
  if (!currentRobot.value) {
    message.error('请先选择机器人');
    return;
  }
  try {
    await ToogoRobotApi.cancelOrder({ robotId: currentRobot.value.id, orderId });
    message.success('撤单成功');
    // 刷新订单数据（挂单和成交明细）
    await refreshOrderData();
  } catch (error: any) {
    message.error('撤单失败: ' + (error.message || '未知错误'));
  }
};

// 查看详情
const viewDetail = async (robot: any) => {
  currentRobot.value = robot;
  showDetailModal.value = true;
  
  // 加载机器人详情数据
  loadDetailData(robot);
  
  // 尝试解析并加载策略组信息
  if (robot.currentStrategy) {
    try {
      const strategyData = typeof robot.currentStrategy === 'string' 
        ? JSON.parse(robot.currentStrategy) 
        : robot.currentStrategy;
      
      // 如果有groupId或group_id，加载策略组信息和对应的策略模板
      const groupId = strategyData.groupId || strategyData.group_id; // 支持两种格式（兼容旧数据）
      if (groupId) {
        try {
          const groupList = await ToogoStrategyApi.groupList({ page: 1, pageSize: 1, id: groupId });
          if (groupList?.list && groupList.list.length > 0) {
            currentRobotStrategy.value = groupList.list[0];
            
            // 加载策略组对应的策略模板（使用实时市场状态和风险偏好）
            // 注意：不使用数据库字段或硬编码默认值作为后备，如果实时数据不可用则不加载模板
            try {
              const robotId = robot.id;
              const marketState = analysisData.value[robotId]?.config?.marketState 
                || analysisData.value[robotId]?.signal?.currentMarketState;
              const riskPreference = analysisData.value[robotId]?.config?.riskPreference;
              
              // 如果实时数据不可用，不加载模板（避免使用错误的默认值）
              if (!marketState || !riskPreference) {
                console.debug('实时市场状态或风险偏好不可用，跳过加载策略模板');
                return;
              }
              
              const templateRes = await ToogoStrategyApi.templateList({ 
                groupId: groupId, 
                marketState: marketState,
                riskPreference: riskPreference,
                pageSize: 1 
              });
              if (templateRes?.list && templateRes.list.length > 0) {
                currentStrategyTemplate.value = templateRes.list[0];
              }
            } catch (err) {
              console.debug('加载策略模板失败:', err);
            }
          }
        } catch (err) {
          console.debug('加载策略组信息失败:', err);
        }
      } else {
        // 如果没有group_id，尝试查找我的策略列表里的默认策略组
        try {
          const myGroupList = await ToogoStrategyApi.groupList({ page: 1, pageSize: 100, isOfficial: 0 });
          const defaultGroup = myGroupList?.list?.find((g: any) => g.isDefault === 1);
          if (defaultGroup) {
            currentRobotStrategy.value = defaultGroup;
          }
        } catch (err) {
          console.debug('查找我的默认策略组失败:', err);
        }
      }
    } catch (err) {
      console.debug('解析策略配置失败:', err);
    }
  } else {
    // 如果没有currentStrategy，也尝试查找我的策略列表里的默认策略组
    try {
      const myGroupList = await ToogoStrategyApi.groupList({ page: 1, pageSize: 100, isOfficial: 0 });
      const defaultGroup = myGroupList?.list?.find((g: any) => g.isDefault === 1);
      if (defaultGroup) {
        currentRobotStrategy.value = defaultGroup;
      }
    } catch (err) {
      console.debug('查找我的默认策略组失败:', err);
    }
  }
};


// 刷新订单数据（单独提取，用于定时刷新）
const refreshOrderData = async () => {
  if (!currentRobot.value) return;
  
  const robotId = currentRobot.value.id;
  
  // 加载当前挂单
  try {
    const orderRes = await ToogoRobotApi.orders({ robotId });
    currentOpenOrders.value = orderRes?.list || [];
  } catch (error) {
    console.error('刷新挂单失败:', error);
  }

  // 加载成交明细（从数据库读取，数据库数据会自动更新）
  historyLoading.value = true;
  try {
    const historyRes = await ToogoWalletApi.orderHistory({ 
      robotId: robotId, 
      status: 2, // 只显示已平仓的订单
      page: 1,
      pageSize: 50 // 增加显示数量，显示更多历史订单
    });
    orderHistory.value = historyRes?.list || [];
  } catch (error) {
    console.error('刷新成交明细失败:', error);
  } finally {
    historyLoading.value = false;
  }
};

// 加载详情数据
const loadDetailData = async (robot: any) => {
  // 加载持仓
  positionLoading.value = true;
  try {
    const posRes = await ToogoRobotApi.positions({ robotId: robot.id });
    currentPositions.value = posRes?.list || [];
  } catch (error) {
    console.error('加载持仓失败:', error);
  } finally {
    positionLoading.value = false;
  }

  // 加载当前挂单
  orderLoading.value = true;
  try {
    const orderRes = await ToogoRobotApi.orders({ robotId: robot.id });
    currentOpenOrders.value = orderRes?.list || [];
  } catch (error) {
    console.error('加载挂单失败:', error);
  } finally {
    orderLoading.value = false;
  }

  // 加载成交明细（从数据库读取，数据库数据会自动更新）
  historyLoading.value = true;
  try {
    const historyRes = await ToogoWalletApi.orderHistory({ 
      robotId: robot.id, 
      status: 2, // 只显示已平仓的订单
      page: 1,
      pageSize: 50 // 增加显示数量，显示更多历史订单
    });
    orderHistory.value = historyRes?.list || [];
  } catch (error) {
    console.error('加载成交明细失败:', error);
  } finally {
    historyLoading.value = false;
  }

  // 加载算力消耗
  try {
    const powerRes = await ToogoStrategyApi.powerConsumeList({ robotId: robot.id, page: 1, perPage: 20 });
    powerConsumeList.value = powerRes?.list || [];
  } catch (error) {
    console.error('加载算力消耗失败:', error);
  }
};

// 开始定时刷新
const startRefresh = () => {
  // 【WebSocket优先】实时分析（含市场状态平滑播报）走WS推送；
  // 持仓也走WS推送；这里只保留“订阅更新 + 低频HTTP兜底”，避免WS断线导致界面长时间停滞
  fastRefreshTimer = setInterval(() => {
    // 机器人运行状态变化时（启动/停止/崩溃）需要更新订阅范围
    updateWsSubscription();
    updateWsPositionsSubscription();
    // 详情弹窗打开时：挂单走 WS（避免60s REST 刷新带来的闪烁）
    updateWsOrdersSubscription();

    wsFallbackCounter++;
    // 兜底：每30秒拉一次（仅用于WS断线或漏推时纠偏）
    if (wsFallbackCounter % 15 === 0) {
      loadRealtimeData();
      loadPositionData();
    }
  }, 2000);
  
  // 详情页挂单改为 WS；这里保留一个更低频的兜底（防止极端情况下 WS/对账都失效）
  orderRefreshTimer = setInterval(() => {
    if (currentRobot.value && showDetailModal.value) {
      // 仅兜底刷新成交明细/挂单（DB为主），频率降低，减少闪烁风险
      refreshOrderData();
    }
  }, 5 * 60 * 1000);
  
  // 慢速刷新：每30秒更新日志数据（executionLogs 改为按需加载）
  refreshTimer = setInterval(() => {
    // 刷新运行区间净盈亏（低频，来自数据库汇总）
    loadRunningSessionNetPnl(false);

    // 刷新所有机器人的方向预警日志
    for (const robot of robotList.value) {
      loadSignalLogs(robot.id);
    }
  }, 30000);
};

// 停止定时刷新
const stopRefresh = () => {
  if (refreshTimer) {
    clearInterval(refreshTimer);
    refreshTimer = null;
  }
  if (fastRefreshTimer) {
    clearInterval(fastRefreshTimer);
    fastRefreshTimer = null;
  }
  if (orderRefreshTimer) {
    clearInterval(orderRefreshTimer);
    orderRefreshTimer = null;
  }
};

// 加载持仓数据（单独提取，避免影响实时价格刷新）
// 使用 localStorage 持久化最高盈利，防止页面刷新后丢失
const MAX_PROFIT_STORAGE_KEY = 'toogo_max_profit_cache';

// 从 localStorage 获取最高盈利缓存
const getMaxProfitCache = (): Record<string, number> => {
  try {
    const cached = localStorage.getItem(MAX_PROFIT_STORAGE_KEY);
    return cached ? JSON.parse(cached) : {};
  } catch {
    return {};
  }
};

// 保存最高盈利到 localStorage
const saveMaxProfitCache = (cache: Record<string, number>) => {
  try {
    localStorage.setItem(MAX_PROFIT_STORAGE_KEY, JSON.stringify(cache));
  } catch {
    // ignore
  }
};

// 比较两个持仓是否相同（用于避免不必要的重新渲染）
// 检查持仓结构是否相同（不比较价格和盈亏，这些会频繁变化）
const isPositionStructureEqual = (oldPos: any, newPos: any): boolean => {
  if (!oldPos || !newPos) return false;
  // 只比较结构性字段，不比较价格和盈亏（这些会频繁变化导致闪烁）
  // 【注意】开关状态变化不应该被视为结构变化，应该原地更新
  return (
    oldPos.symbol === newPos.symbol &&
    oldPos.positionSide === newPos.positionSide &&
    Math.abs((oldPos.quantity || 0) - (newPos.quantity || 0)) < 0.0001 &&
    Math.abs((oldPos.entryPrice || 0) - (newPos.entryPrice || 0)) < 0.01
    // 移除 takeProfitEnabled 的比较，允许开关状态变化时原地更新
  );
};

// 将“持仓快照”合并到本地状态（HTTP轮询 & WS推送共用）
const applyRobotPositionsSnapshot = (robot: any, positions: any[], source: string) => {
  const list = positions || [];
      
  if (list.length > 0) {
        // 有数据：重置空结果计数
        positionEmptyStreak.value[robot.id] = 0;
        positionLastNonEmptyAt.value[robot.id] = Date.now();
        // 有非空快照后，继续维持一小段保护窗，避免“刚出现下一秒又空”导致闪
        {
          const protectMs = 15_000;
          const until = Date.now() + protectMs;
          positionOpenProtectionUntil.value[robot.id] = Math.max(positionOpenProtectionUntil.value[robot.id] || 0, until);
        }
        // 【防闪烁优化】使用原地更新策略：只更新变化的字段，保持对象引用稳定
        const oldPositions = positionData.value[robot.id] || [];
        const newPositions: any[] = [];
        const seenKeys = new Set<string>();
        
        // 获取最高盈利缓存（用于清理和检测新仓位）
        const maxProfitCache = getMaxProfitCache();
        let cacheNeedsUpdate = false;

    for (const newPos of list) {
      // 规范化（防止 Gate/OKX 偶发返回 long/short 导致 key 抖动、方向判断错误）
      if (newPos && newPos.positionSide != null) {
        newPos.positionSide = String(newPos.positionSide).toUpperCase().trim();
      }
      if (newPos && newPos.symbol != null) {
        // 统一 symbol key：兼容 BTC/USDT、BTC_USDT、BTCUSDT 等格式，避免 key 抖动导致“行消失又回来”
        newPos.symbol = String(newPos.symbol)
          .toUpperCase()
          .trim()
          .replaceAll('/', '')
          .replaceAll('_', '')
          .replaceAll('-', '');
      }

      // 过滤“0数量”持仓（平仓后引擎/交易所可能短暂返回 PositionAmt=0 的残留对象）
      // Gate 刚开仓时可能出现非常小的抖动值（例如 0.00009999），若阈值过大会把“第二个新仓位”误过滤导致行闪烁。
      const amt = Number(newPos.positionAmt ?? 0);
      if (!Number.isFinite(amt) || Math.abs(amt) <= 1e-12) {
        continue;
      }

      // 手动平仓抑制：5秒内忽略刚手动平仓的持仓回流
      const closeKey = `${robot.id}_${newPos.symbol}_${newPos.positionSide}`;
      const closedAt = recentManualCloseAt.value[closeKey] || 0;
      if (closedAt > 0 && Date.now() - closedAt < 5000) {
        continue;
      }

      seenKeys.add(`${newPos.symbol}_${newPos.positionSide}`);
      // 重置“单仓位缺失计数”
      positionMissingStreak.value[closeKey] = 0;
      delete positionMissingSince.value[closeKey];

          // 查找对应的旧持仓
          const oldPos = oldPositions.find(
            (p: any) => p.symbol === newPos.symbol && p.positionSide === newPos.positionSide
          );
          
      // 检测是否是新仓位（没有对应的旧持仓）
          const isNewPosition = !oldPos;
          const cacheKey = `${robot.id}_${newPos.symbol}_${newPos.positionSide}`;
          
      // 最高盈利就是订单的未实现盈亏的最大值：新仓位清缓存，从0开始
          let finalMaxProfit = 0;
          if (isNewPosition) {
            delete maxProfitCache[cacheKey];
            cacheNeedsUpdate = true;
            finalMaxProfit = newPos.maxProfitReached || 0;
          } else {
            const backendMaxProfit = newPos.maxProfitReached || 0;
            const oldMaxProfit = oldPos?.maxProfitReached || 0;
            finalMaxProfit = Math.max(backendMaxProfit, oldMaxProfit);
          }
          
      // 如果当前未实现盈亏为正且大于记录的最高盈利，更新最高盈利（只增不减）
          const currentProfit = newPos.unrealizedPnl || 0;
          if (currentProfit > 0 && currentProfit > finalMaxProfit) {
            finalMaxProfit = currentProfit;
          }
          
      // 【完全由后端控制】开关状态完全由后端控制，不再同步到本地
          
          // 如果持仓结构相同（只是价格/盈亏变化），原地更新字段，保持对象引用
          // 【修复】确保后端推送的血条字段（takeProfitStartProgress、stopLossProgress等）被正确保留
          if (oldPos && isPositionStructureEqual(oldPos, newPos)) {
            // 保留后端计算的血条字段和开关状态（优先使用后端数据）
            const backendProgressFields = {
              takeProfitStartProgress: newPos.takeProfitStartProgress,
              takeProfitRetreatBar: newPos.takeProfitRetreatBar,
              stopLossProgress: newPos.stopLossProgress,
              realTimeProfitPercent: newPos.realTimeProfitPercent,
              takeProfitRetreatPercent: newPos.takeProfitRetreatPercent,
              // 【重要】确保开关状态被正确保留（后端控制）
              takeProfitEnabled: newPos.takeProfitEnabled,
            };
            // 【重要】当开关状态变化时（特别是从false变为true），需要强制触发Vue响应式更新
            const oldTakeProfitEnabled = oldPos.takeProfitEnabled || false;
            const newTakeProfitEnabled = newPos.takeProfitEnabled || false;
            const switchStateChanged = oldTakeProfitEnabled !== newTakeProfitEnabled;
            
            Object.assign(oldPos, {
              ...newPos,
              maxProfitReached: finalMaxProfit,
              // 确保后端推送的血条字段和开关状态被保留（即使为0/false也要保留，表示后端已计算）
              ...backendProgressFields,
            });
            
            // 【修复】如果开关状态发生变化（特别是启动止盈），强制触发Vue响应式更新
            if (switchStateChanged && newTakeProfitEnabled) {
              // 使用Vue的响应式更新机制，确保UI能立即反映变化
              // 通过重新赋值触发响应式更新
              oldPos.takeProfitEnabled = true;
            }
            
        newPositions.push(oldPos);
          } else {
            // 新创建持仓对象时，确保包含所有后端字段（包括开关状态和血条字段）
            newPositions.push({
            ...newPos,
          maxProfitReached: finalMaxProfit,
            // 确保后端推送的开关状态和血条字段被包含
            takeProfitEnabled: newPos.takeProfitEnabled,
            takeProfitStartProgress: newPos.takeProfitStartProgress,
            takeProfitRetreatBar: newPos.takeProfitRetreatBar,
            stopLossProgress: newPos.stopLossProgress,
            realTimeProfitPercent: newPos.realTimeProfitPercent,
            takeProfitRetreatPercent: newPos.takeProfitRetreatPercent,
            });
          }
        }

        // Gate 双向持仓：有时快照会短暂少返回一侧（例如只返回 LONG 或只返回 SHORT）
        // 为避免 UI 行“消失又回来”闪烁：对“未出现在本次快照”的旧仓位做缺失保护。
        // 说明：之前按“连续缺失>=3次移除”，若轮询约20s，则会出现“消失约1分钟”的体感。
        // 这里改为“按时间保留”（默认2分钟内都保留），更贴近真实交易状态。
        for (const oldPos of oldPositions) {
          if (!oldPos) continue;
          const sym = String(oldPos.symbol || '')
            .toUpperCase()
            .trim()
            .replaceAll('/', '')
            .replaceAll('_', '')
            .replaceAll('-', '');
          const ps = String(oldPos.positionSide || '').toUpperCase().trim();
          if (!sym || !ps) continue;
          const k = `${sym}_${ps}`;
          if (seenKeys.has(k)) continue; // 本次已出现

          const fullKey = `${robot.id}_${sym}_${ps}`;
          const last = positionMissingStreak.value[fullKey] || 0;
          const next = last + 1;
          positionMissingStreak.value[fullKey] = next;

          // 保护策略（更优体验）：
          // - 首次缺失：立刻对该机器人“补拉一次 positions”（单机器人、带冷却），多数情况下几百毫秒内就能补齐缺失的一侧
          // - 短窗口保留：在补拉结果回来前保留旧行，避免闪烁
          // - 超过短窗口仍未出现：才认为确实已无该仓位，清理本地状态并移除（避免平仓后拖很久）
          const now = Date.now();
          const firstMissingAt = positionMissingSince.value[fullKey] || 0;
          if (firstMissingAt === 0) {
            positionMissingSince.value[fullKey] = now;
            // 异步补拉，不阻塞本次渲染
            setTimeout(() => repairRobotPositionsOnce(robot, `missing:${sym}_${ps}`), 200);
            newPositions.push(oldPos);
            continue;
          }
          const keepMs = 25_000;
          if (now-firstMissingAt < keepMs) {
            newPositions.push(oldPos);
          } else {
            // 连续缺失：清理该仓位相关的本地状态（止盈开关/最高盈利）
            const switchKey = `${robot.id}_${sym}_${ps}`;
            delete takeProfitRetreatSwitch.value[switchKey];
            delete maxProfitCache[switchKey];
            cacheNeedsUpdate = true;
            delete positionMissingSince.value[fullKey];
          }
        }
        
        // 【优化】局部刷新：只更新血条相关字段，保持对象引用不变，实现局部刷新
        // 检查持仓列表结构是否变化（增删）
        const oldKeys = oldPositions.map((p: any) => `${p.symbol}_${p.positionSide}`).sort().join(',');
        const newKeys = newPositions.map((p: any) => `${p.symbol}_${p.positionSide}`).sort().join(',');
        const structureChanged = oldKeys !== newKeys;
        
        if (structureChanged) {
          // 结构变化（增删持仓），需要替换整个数组
          positionData.value[robot.id] = [...newPositions];
        } else {
          // 结构未变化，只更新血条相关字段，实现局部刷新
          // 保持对象引用不变，只更新属性，Vue会自动检测并只更新相关DOM
          let needRefresh = false; // 标记是否需要触发局部刷新
          for (const newPos of newPositions) {
            const oldPos = oldPositions.find((p: any) => 
              p.symbol === newPos.symbol && p.positionSide === newPos.positionSide
            );
            if (oldPos) {
              // 检测止损血条是否达到100%
              const oldStopLossProgress = oldPos.stopLossProgress || 0;
              const newStopLossProgress = newPos.stopLossProgress || 0;
              const stopLossReached100 = oldStopLossProgress < 100 && newStopLossProgress >= 100;
              
              // 检测回撤止盈血条是否达到0%（需要先判断是否已启动止盈）
              const isTakeProfitEnabled = getTakeProfitRetreatSwitch(robot.id, newPos.symbol, newPos.positionSide, newPos);
              const oldRetreatBar = oldPos.takeProfitRetreatBar !== undefined && oldPos.takeProfitRetreatBar !== null ? Number(oldPos.takeProfitRetreatBar) : 100;
              const newRetreatBar = newPos.takeProfitRetreatBar !== undefined && newPos.takeProfitRetreatBar !== null ? Number(newPos.takeProfitRetreatBar) : 100;
              const retreatBarReached0 = isTakeProfitEnabled && oldRetreatBar > 0 && newRetreatBar <= 0;
              
              // 如果止损血条达到100%或回撤止盈血条达到0%，标记需要刷新
              if (stopLossReached100 || retreatBarReached0) {
                needRefresh = true;
              }
              
              // 只更新血条相关字段和开关状态，保持对象引用不变
              // 这样Vue可以检测到属性变化并只更新相关的DOM部分（血条列）
              oldPos.takeProfitStartProgress = newPos.takeProfitStartProgress;
              oldPos.takeProfitRetreatBar = newPos.takeProfitRetreatBar;
              oldPos.stopLossProgress = newPos.stopLossProgress;
              oldPos.realTimeProfitPercent = newPos.realTimeProfitPercent;
              oldPos.takeProfitRetreatPercent = newPos.takeProfitRetreatPercent;
              oldPos.takeProfitEnabled = newPos.takeProfitEnabled;
              // 同时更新未实现盈亏和最高盈利（这些也会影响血条显示）
              oldPos.unrealizedPnl = newPos.unrealizedPnl;
              if (newPos.maxProfitReached !== undefined && newPos.maxProfitReached !== null) {
                oldPos.maxProfitReached = Math.max(oldPos.maxProfitReached || 0, newPos.maxProfitReached);
              }
            }
          }
          
          // 如果检测到止损血条达到100%或回撤止盈血条达到0%，触发局部刷新（类似启动止盈血条刷新）
          if (needRefresh) {
            // 延迟执行，避免频繁刷新，同时确保Vue响应式更新完成
            setTimeout(() => {
              repairRobotPositionsOnce(robot, `血条触发刷新:止损100%或回撤0%`);
            }, 200);
          }
          
          // 不需要替换数组，Vue会自动检测对象属性的变化并局部更新DOM
        }
        
    // 保存更新后的缓存
        if (cacheNeedsUpdate) {
          saveMaxProfitCache(maxProfitCache);
        }

    // ===== 详情弹窗联动：让“实时持仓”tab 直接跟随 WS/HTTP 的持仓快照 =====
    if (showDetailModal.value && currentRobot.value?.id === robot.id) {
      // positionData 可能被替换或原地更新，这里统一从 positionData 取最终展示值
      currentPositions.value = positionData.value[robot.id] || [];
    }
    return;
  }

  // 空列表：
  // - 默认：连续3次为空才清空（避免偶发空数据造成闪烁）
  // - 若刚手动平仓，则立即清空（避免“平了还显示”）
  const hadLocalPosition = (positionData.value[robot.id] || []).length > 0;
  // 新开仓保护窗内：直接忽略空快照（不清空也不计数），彻底消除“新开仓闪一下”
  const openProtectUntil = positionOpenProtectionUntil.value[robot.id] || 0;
  if (openProtectUntil > 0 && Date.now() < openProtectUntil) {
    return;
  }
  const hasRecentManualClose = Object.keys(recentManualCloseAt.value).some((k) => {
    if (!k.startsWith(`${robot.id}_`)) return false;
    const ts = recentManualCloseAt.value[k] || 0;
    return ts > 0 && Date.now() - ts < 5000;
  });

  // 如果最近保护窗内出现过非空持仓，则认为本次空结果大概率是同步抖动，不清空（除非明确手动平仓）
  const lastNonEmpty = positionLastNonEmptyAt.value[robot.id] || 0;
  const protectMs = 10_000;
  const recentlyNonEmpty = lastNonEmpty > 0 && Date.now() - lastNonEmpty < protectMs;

  if (hadLocalPosition && hasRecentManualClose) {
    console.log(`[${source}] robotId=${robot.id} 手动平仓后空列表快速清空`);
          positionData.value[robot.id] = [];
    positionEmptyStreak.value[robot.id] = 0;
    positionLastNonEmptyAt.value[robot.id] = 0;
    positionOpenProtectionUntil.value[robot.id] = 0;

    // 详情弹窗联动：持仓被清空时同步清空详情页
    if (showDetailModal.value && currentRobot.value?.id === robot.id) {
      currentPositions.value = [];
    }

          const keysToDelete = Object.keys(takeProfitRetreatSwitch.value).filter(key => key.startsWith(`${robot.id}_`));
          keysToDelete.forEach(key => delete takeProfitRetreatSwitch.value[key]);
          
          const maxProfitCache = getMaxProfitCache();
          const profitKeysToDelete = Object.keys(maxProfitCache).filter(key => key.startsWith(`${robot.id}_`));
          profitKeysToDelete.forEach(key => delete maxProfitCache[key]);
          if (profitKeysToDelete.length > 0) {
            saveMaxProfitCache(maxProfitCache);
    }
    return;
  }

  if (hadLocalPosition && recentlyNonEmpty) {
    // 抑制抖动：不累加 streak，不清空，等待下一次快照再决策
    return;
  }

  const streak = (positionEmptyStreak.value[robot.id] || 0) + 1;
  positionEmptyStreak.value[robot.id] = streak;
  if (streak >= 3 && hadLocalPosition) {
    console.log(`[${source}] robotId=${robot.id} 连续${streak}次无持仓，清空持仓列表`);
    positionData.value[robot.id] = [];
    positionLastNonEmptyAt.value[robot.id] = 0;
    positionOpenProtectionUntil.value[robot.id] = 0;

    // 详情弹窗联动：持仓被清空时同步清空详情页
    if (showDetailModal.value && currentRobot.value?.id === robot.id) {
      currentPositions.value = [];
    }
    const keysToDelete = Object.keys(takeProfitRetreatSwitch.value).filter(key => key.startsWith(`${robot.id}_`));
    keysToDelete.forEach(key => delete takeProfitRetreatSwitch.value[key]);
    const maxProfitCache = getMaxProfitCache();
    const profitKeysToDelete = Object.keys(maxProfitCache).filter(key => key.startsWith(`${robot.id}_`));
    profitKeysToDelete.forEach(key => delete maxProfitCache[key]);
    if (profitKeysToDelete.length > 0) {
      saveMaxProfitCache(maxProfitCache);
          }
        }
};

// 加载持仓数据（兜底：WS断线时使用）
const loadPositionData = async () => {
  const runningRobots = robotList.value.filter((r: any) => r.status === 2);
  if (runningRobots.length === 0) return;
  
  // 批量加载所有运行中机器人的持仓数据
  const promises = runningRobots.map(async (robot) => {
    try {
      const posRes = await ToogoRobotApi.positions({ robotId: robot.id });
      applyRobotPositionsSnapshot(robot, posRes?.list || [], 'HTTP轮询');
    } catch (error: any) {
      // 静默失败，避免控制台刷屏
      if (Math.random() < 0.1) {
      console.warn(`加载机器人 ${robot.id} 持仓失败:`, error);
    }
  }
  });
  
  await Promise.all(promises);
};

// 打开策略模板选择器
const openStrategySelector = async () => {
  selectedGroupId.value = currentRobot.value?.strategyGroupId || null;
  selectedIsOfficial.value = false;
  await loadStrategyGroups();
  showStrategySelector.value = true;
};

// 加载策略模板组
const loadStrategyGroups = async () => {
  loadingStrategyGroups.value = true;
  try {
    // 加载我的策略模板
    const myRes = await ToogoStrategyApi.groupList({ page: 1, pageSize: 100, isOfficial: 0 });
    myStrategyGroups.value = myRes?.list || [];
    
    // 加载官方策略模板
    const officialRes = await ToogoStrategyApi.groupList({ page: 1, pageSize: 100, isOfficial: 1 });
    officialStrategyGroups.value = officialRes?.list || [];
  } catch (error) {
    console.error('加载策略模板失败:', error);
  } finally {
    loadingStrategyGroups.value = false;
  }
};

// 选择官方策略模板（标记为官方）
const selectOfficialGroup = (group: any) => {
  selectedGroupId.value = group.id;
  selectedIsOfficial.value = true;
};

// 应用策略模板
const applyStrategyGroup = async () => {
  if (!selectedGroupId.value || !currentRobot.value) return;
  
  applyingStrategy.value = true;
  try {
    let groupIdToApply = selectedGroupId.value;
    
    // 如果选择的是官方策略，先复制到我的策略
    if (selectedIsOfficial.value) {
      try {
        const copyRes = await ToogoStrategyApi.copyFromOfficial({ officialGroupId: selectedGroupId.value });
        groupIdToApply = copyRes?.id || selectedGroupId.value;
      } catch (copyError: any) {
        const errorMsg = copyError?.message || copyError?.data?.message || '';
        if (errorMsg.includes('已存在')) {
          const myGroup = myStrategyGroups.value.find((g: any) => g.fromOfficialId === selectedGroupId.value);
          if (myGroup) groupIdToApply = myGroup.id;
        } else {
          throw copyError;
        }
      }
    }
    
    // 获取策略模板中的第一个策略作为默认配置应用到机器人
    try {
      const templateRes = await ToogoStrategyApi.templateList({ groupId: groupIdToApply, pageSize: 1 });
      const firstStrategy = templateRes?.list?.[0];
      
      if (firstStrategy) {
        let configData: any = {};
        try {
          configData = typeof firstStrategy.configJson === 'string' 
            ? JSON.parse(firstStrategy.configJson || '{}') 
            : (firstStrategy.configJson || {});
        } catch {
          configData = {};
        }
        
        const updateData = {
          id: currentRobot.value.id,
          robotName: currentRobot.value.robotName,
          exchange: configData.exchange || firstStrategy.exchange || 'binance',
          symbol: configData.symbol || firstStrategy.symbol || 'BTC-USDT',
          tradeType: configData.tradeType || firstStrategy.tradeType || 'perpetual',
          orderType: configData.orderType || firstStrategy.orderType || 'market',
          marginMode: configData.marginMode || firstStrategy.marginMode || 'isolated',
          maxProfitTarget: currentRobot.value.maxProfitTarget || 100,
          maxLossAmount: currentRobot.value.maxLossAmount || 50,
          maxRuntime: currentRobot.value.maxRuntime || 0,
          riskPreference: firstStrategy.riskPreference || 'balanced',
          autoRiskPreference: 1,
          marketState: firstStrategy.marketState || 'trend',
          autoMarketState: 1,
          leverage: configData.leverage || firstStrategy.leverageMin || 10,
          marginPercent: configData.marginPercent || firstStrategy.marginPercentMin || 10,
          useMonitorSignal: 1,
          stopLossPercent: configData.stopLossPercent || firstStrategy.stopLossPercent || 5,
          profitRetreatPercent: configData.profitRetreatPercent || firstStrategy.profitRetreatPercent || 30,
          autoStartRetreatPercent: configData.autoStartRetreatPercent || firstStrategy.autoStartRetreatPercent || 3,
          remark: `策略模板: ${groupIdToApply}`,
        };
        
        await ToogoRobotApi.update(updateData);
        
        // 如果机器人正在运行，通知后端重新加载策略配置
        if (currentRobot.value.status === 2) {
          try {
            await ToogoRobotApi.reloadStrategy({ id: currentRobot.value.id });
            message.success('策略模板已应用！运行中的机器人配置已更新');
          } catch (reloadErr: any) {
            message.warning('策略已更新，但运行中配置刷新失败，建议重启机器人');
          }
        } else {
          message.success('策略模板已应用！配置已更新');
        }
        
        // 立即更新本地显示数据
        Object.assign(currentRobot.value, {
          platform: updateData.exchange,
          exchange: updateData.exchange,
          tradingPair: updateData.symbol,
          symbol: updateData.symbol,
          tradeType: updateData.tradeType,
          orderType: updateData.orderType,
          marginMode: updateData.marginMode,
          riskPreference: updateData.riskPreference,
          marketState: updateData.marketState,
          leverage: updateData.leverage,
          marginPercent: updateData.marginPercent,
          marginRatio: updateData.marginPercent,
          stopLossPercent: updateData.stopLossPercent,
          profitRetreatPercent: updateData.profitRetreatPercent,
          autoStartRetreatPercent: updateData.autoStartRetreatPercent,
          takeProfitRetracePercent: updateData.profitRetreatPercent, // 兼容旧字段名
          autoRiskPreference: updateData.autoRiskPreference,
          autoMarketState: updateData.autoMarketState,
          useMonitorSignal: updateData.useMonitorSignal,
        });
        
      } else {
        message.warning('策略模板中没有可用策略');
      }
    } catch (updateError: any) {
      throw new Error(updateError?.message || '更新机器人配置失败');
    }
    
    showStrategySelector.value = false;
    loadData();
  } catch (error: any) {
    const errorMsg = error?.message || error?.data?.message || '应用失败';
    message.error(errorMsg);
  } finally {
    applyingStrategy.value = false;
  }
};

// 跳转到策略管理页
const goToStrategy = () => {
  showStrategySelector.value = false;
  router.push('/toogo/strategy/my');
};


onMounted(async () => {
  // 每秒刷新一次“当前时间”，用于运行时长等纯UI计算
  nowTickTimer = setInterval(() => {
    nowTick.value = Date.now();
  }, 1000);

  // 监听WS推送（机器人批量实时分析）
  addOnMessage(SocketEnum.EventToogoRobotRealtimePush, wsOnRealtimePush);
  // 监听WS推送（机器人持仓快照）
  addOnMessage(SocketEnum.EventToogoRobotPositionsPush, wsOnPositionsPush);
  // 监听WS推送（机器人挂单快照）
  addOnMessage(SocketEnum.EventToogoRobotOrdersPush, wsOnOrdersPush);
  // 监听WS推送（机器人挂单增量）
  addOnMessage(SocketEnum.EventToogoRobotOrdersDelta, wsOnOrdersDelta);
  // 监听WS推送（持仓增量：由交易所私有WS触发的即时刷新）
  addOnMessage(SocketEnum.EventToogoRobotPositionsDelta, wsOnPositionsDelta);
  // 监听WS推送（交易关键事件）
  addOnMessage(SocketEnum.EventToogoRobotTradeEvent, wsOnTradeEvent);
  loadData();
  loadStrategyGroups();
  startRefresh();
  // 加载订阅信息用于显示倒计时
  await loadSubscriptionInfo();
  startCountdownTimer();
});

onUnmounted(() => {
  if (nowTickTimer) {
    clearInterval(nowTickTimer);
    nowTickTimer = null;
  }

  removeOnMessage(SocketEnum.EventToogoRobotRealtimePush);
  removeOnMessage(SocketEnum.EventToogoRobotPositionsPush);
  removeOnMessage(SocketEnum.EventToogoRobotOrdersPush);
  removeOnMessage(SocketEnum.EventToogoRobotOrdersDelta);
  removeOnMessage(SocketEnum.EventToogoRobotPositionsDelta);
  removeOnMessage(SocketEnum.EventToogoRobotTradeEvent);
  unsubscribeWs();
  unsubscribeWsPositions();
  unsubscribeWsOrders();
  stopRefresh();
  // 清理倒计时定时器
  if (subscriptionCountdownTimer.value) {
    clearInterval(subscriptionCountdownTimer.value);
    subscriptionCountdownTimer.value = null;
  }
});
</script>

<style lang="less" scoped>
.robot-page {
  padding: 16px;
  
  /* ========== 夜间模式适配变量 ========== */
  --robot-bg: #ffffff;
  --robot-border: rgba(0, 0, 0, 0.08);
  --robot-text-primary: rgba(0, 0, 0, 0.88);
  --robot-text-secondary: #666666;
  --robot-text-tertiary: #888888;
  --robot-hover-bg: rgba(0, 0, 0, 0.02);
  --robot-panel-bg: rgba(0, 0, 0, 0.02);
  --robot-chart-bg: rgba(255, 255, 255, 0.6);
  --robot-chart-line: #3b82f6;
  --robot-chart-grid: rgba(0, 0, 0, 0.1);
  
  /* 自动适配深色模式 */
  html.dark & {
    --robot-bg: rgba(255, 255, 255, 0.05);
    --robot-border: rgba(255, 255, 255, 0.12);
    --robot-text-primary: rgba(255, 255, 255, 0.9);
    --robot-text-secondary: rgba(255, 255, 255, 0.65);
    --robot-text-tertiary: rgba(255, 255, 255, 0.45);
    --robot-hover-bg: rgba(255, 255, 255, 0.05);
    --robot-panel-bg: rgba(255, 255, 255, 0.03);
    --robot-chart-bg: rgba(0, 0, 0, 0.3);
    --robot-chart-line: #60a5fa;
    --robot-chart-grid: rgba(255, 255, 255, 0.1);
  }
}

.mb-3 {
  margin-bottom: 12px;
}

/* ==================== 筛选工具栏优化 ==================== */
.filter-toolbar-card {
  :deep(.n-card__content) {
    padding: 12px 16px;
  }

  :deep(.n-select) {
    .n-base-selection {
      transition: all 0.2s ease;
      
      &:hover {
        border-color: var(--primary-color);
      }
    }
  }

  :deep(.n-radio-group) {
    .n-radio-button {
      transition: all 0.2s ease;
    }
  }
}

/* ==================== 列表视图表格样式优化 ==================== */
.robot-list-table-card {
  :deep(.n-card__content) {
    padding: 0;
  }
}

.robot-list-table {
  /* 表头样式优化 */
  :deep(.n-data-table-thead) {
    .n-data-table-th {
      font-weight: 600;
      font-size: 13px;
      padding: 10px 8px;
      background: var(--table-header-color);
    }
  }

  /* 表格行样式优化 */
  :deep(.n-data-table-tbody) {
    .n-data-table-tr {
      transition: background-color 0.2s ease;
      
      &:hover {
        background: var(--table-header-color);
      }
    }

    .n-data-table-td {
      padding: 10px 8px;
      font-size: 13px;
      line-height: 1.6;
      vertical-align: middle;
    }
  }

  /* 固定列样式 */
  :deep(.n-data-table-td--fixed-right) {
    background: inherit;
  }
}

.robot-card {
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  border-radius: 12px;
  overflow: hidden;

  &.running {
    border: 1px solid var(--success-color);
    box-shadow: 0 0 0 1px rgba(24, 160, 88, 0.1);
  }
  
  &:hover {
    transform: translateY(-2px);
    box-shadow: 0 8px 16px rgba(0, 0, 0, 0.08);
  }
}

/* ==================== 多周期市场状态播报面板 ==================== */
.market-realtime-panel {
  padding: 8px 10px;
  background: rgba(24, 160, 88, 0.06);
  border: 1px solid rgba(24, 160, 88, 0.12);
  border-radius: 8px;
}

/* ==================== 头部连接状态样式 ==================== */
.header-connection {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 2px 8px;
  border-radius: 10px;
  font-size: 10px;
  font-weight: 500;
  
  .conn-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
  }
  
  &.connected {
    background: rgba(87, 202, 34, 0.15);
    color: rgb(87, 202, 34);
    .conn-dot {
      background: rgb(87, 202, 34);
      box-shadow: 0 0 4px rgba(87, 202, 34, 0.6);
      animation: pulse-green 2s infinite;
    }
  }
  
  &.connecting {
    background: rgba(245, 158, 11, 0.15);
    color: rgb(245, 158, 11);
    .conn-dot {
      background: rgb(245, 158, 11);
      animation: pulse-yellow 1s infinite;
    }
  }
  
  &.disconnected {
    background: rgba(239, 68, 68, 0.15);
    color: rgb(239, 68, 68);
    .conn-dot {
      background: rgb(239, 68, 68);
    }
  }
}

/* ==================== 三列布局信号面板 ==================== */
.signal-three-column {
  display: grid;
  grid-template-columns: 130px auto 1fr;
  gap: 16px;
  padding: 12px;
  background: var(--card-color);
  border: 1px solid rgba(0,0,0,0.05);
  border-radius: 8px;
  min-height: 150px;
  align-items: center;
  
  @media (max-width: 768px) {
    grid-template-columns: 1fr;
    grid-template-rows: auto auto auto;
    gap: 12px;
  }
}

/* 第一列：机器人动画 */
.column-robot {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
}

/* 第二列：方向预警按钮 */
.column-signal {
  display: flex;
  flex-direction: column;
  gap: 8px;
  justify-content: center;
  height: 100%;
  min-width: 130px;
  
  .signal-block {
    padding: 8px 12px;
    border-radius: 6px;
    transition: all 0.2s ease;
    background: rgba(0,0,0,0.02);
    border: 1px solid transparent;
    
    &.long {
      .signal-header { color: #9ca3af; }
      .signal-trigger { color: #9ca3af; }
      
      &.active {
        background: rgba(16, 185, 129, 0.15);
        border-color: #10b981;
        box-shadow: 0 0 10px rgba(16, 185, 129, 0.2);
        transform: translateY(-1px);
        
        .signal-header { color: #059669; }
        .signal-trigger { color: #10b981; font-weight: 600; }
      }
    }
    
    &.short {
      .signal-header { color: #9ca3af; }
      .signal-trigger { color: #9ca3af; }
      
      &.active {
        background: rgba(239, 68, 68, 0.15);
        border-color: #ef4444;
        box-shadow: 0 0 10px rgba(239, 68, 68, 0.2);
        transform: translateY(-1px);
        
        .signal-header { color: #dc2626; }
        .signal-trigger { color: #ef4444; font-weight: 600; }
      }
    }
    
    .signal-header {
      display: flex;
      align-items: center;
      gap: 4px;
      font-weight: 500;
      font-size: 12px;
      margin-bottom: 2px;
      
      .signal-icon { font-size: 14px; }
    }
    
    .signal-trigger {
      font-size: 13px;
      font-weight: 600;
      font-family: 'JetBrains Mono', monospace;
    }
    
    .signal-distance {
      font-size: 10px;
      color: #888;
      margin-top: 2px;
    }
  }
  
  .current-price-block {
    display: flex;
    align-items: baseline;
    gap: 6px;
    padding: 6px 12px;
    background: rgba(0, 0, 0, 0.02);
    border-radius: 6px;
    
    .price-value {
      font-size: 16px;
      font-weight: 600;
      font-family: 'JetBrains Mono', monospace;
      
      &.up { color: #10b981; }
      &.down { color: #ef4444; }
    }
    
    .price-change {
      font-size: 11px;
      font-weight: 400;
      
      &.up { color: #10b981; }
      &.down { color: #ef4444; }
    }
  }
}

/* 第三列：价格图表 */
.column-chart {
  display: flex;
  flex-direction: column;
  justify-content: center;
  padding: 6px;
  height: 100%;
  min-width: 200px;
  
  .chart-svg {
    width: 100%;
    height: 120px;
    display: block;
  }
  
  .chart-svg-vertical {
     width: 100%;
     height: 100%;
     min-height: 100px;
     display: block;
  }
  
  .chart-line { fill: none; stroke: #3b82f6; stroke-width: 1.5; }
  .chart-fill { fill: #3b82f6; fill-opacity: 0.15; }
  .chart-baseline { stroke: #3b82f6; stroke-width: 1; stroke-dasharray: 3,3; opacity: 0.3; }
  .chart-upper { stroke: #ef4444; stroke-width: 1; stroke-dasharray: 2,2; }
  .chart-lower { stroke: #10b981; stroke-width: 1; stroke-dasharray: 2,2; }
  
  .chart-labels {
    display: flex;
    justify-content: space-between;
    font-size: 10px;
    font-weight: 400;
    margin-top: 4px;
    
    .label-high { color: #ef4444; }
    .label-low { color: #10b981; }
  }
}

.mini-robot-scene {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  
  .mini-robot {
    width: 120px;
    height: 120px;
    transition: all 0.3s ease;
    
    @media (max-width: 640px) {
      width: 100px;
      height: 100px;
    }
  }
  
  .robot-mood-text {
    font-size: 11px;
    color: var(--robot-text-secondary);
    font-weight: 400;
    text-align: center;
    max-width: 120px;
    line-height: 1.3;
  }
  
  .robot-shadow {
    transition: all 0.3s ease;
  }
  
  .tail-group {
    transition: all 0.3s ease;
  }
  
  .head-group {
    transition: all 0.3s ease;
  }
  
  .body-screen {
    transition: all 0.5s ease;
  }
  
  /* 动画状态样式 */
  &.mood-happy {
    .mini-robot { animation: mini-jump 0.6s infinite alternate ease-in-out; }
    .tail-group { animation: mini-wag 0.3s infinite ease-in-out; }
    .eye { display: none; }
  }
  
  &.mood-thinking {
    .head-group { transform: rotate(8deg); transform-origin: center bottom; }
    .left-eyebrow { transform: rotate(-8deg) translateY(-1px); }
    .right-eyebrow { transform: rotate(8deg) translateY(1px); }
  }
  
  &.mood-confused {
    .head-group { animation: mini-shake 1.5s infinite ease-in-out; }
    .right-ear { transform: rotate(15deg) translate(2px, -2px); }
  }
  
  &.mood-tired {
    .mini-robot { filter: grayscale(0.4); animation: mini-float 4s ease-in-out infinite; }
    .head-group { transform: translateY(3px) rotate(-3deg); }
  }
  
  &.mood-excited {
    .mini-robot { animation: mini-vibrate 0.15s infinite; }
    .tail-group { animation: mini-wag 0.15s infinite; }
    .body-screen { animation: mini-rainbow 1s infinite; }
  }
  
  &.mood-focused {
    .eye { fill: #ff3333 !important; }
    .eyebrow { transform: rotate(10deg) translateY(2px); }
  }
  
  &.mood-sad {
    .mini-robot { transform: scale(0.95); }
    .head-group { transform: translateY(5px) rotate(-5deg); }
    .tail-group { transform: rotate(20deg); }
  }
  
  &.mood-conservative {
    .eye { fill: #4cd137 !important; }
    .mini-robot { animation: mini-float 3s ease-in-out infinite; }
  }
  
  &.mood-balanced {
    .eye { fill: #00a8ff !important; }
    .head-group { animation: mini-balance 3s infinite ease-in-out; }
  }
  
  &.mood-aggressive {
    .mini-robot { animation: mini-vibrate 0.1s infinite; transform: translateY(-3px); }
    .eye { fill: #e84118 !important; filter: drop-shadow(0 0 3px #e84118) !important; }
    .eyebrow { transform: rotate(15deg) translateY(3px); }
  }
}

/* 机器人动画关键帧 */
@keyframes mini-jump {
  0% { transform: translateY(0); }
  100% { transform: translateY(-5px); }
}

@keyframes mini-wag {
  0%, 100% { transform: rotate(-5deg); }
  50% { transform: rotate(5deg); }
}

@keyframes mini-shake {
  0%, 100% { transform: rotate(0); }
  25% { transform: rotate(-5deg); }
  75% { transform: rotate(5deg); }
}

@keyframes mini-float {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-3px); }
}

@keyframes mini-vibrate {
  0%, 100% { transform: translateX(0); }
  25% { transform: translateX(-1px); }
  75% { transform: translateX(1px); }
}

@keyframes mini-rainbow {
  0% { fill: #111; }
  25% { fill: #1a1a3e; }
  50% { fill: #111; }
  75% { fill: #1e1e3e; }
  100% { fill: #111; }
}

@keyframes mini-balance {
  0%, 100% { transform: rotate(0); }
  25% { transform: rotate(-3deg); }
  75% { transform: rotate(3deg); }
}

.status-section {
  background: linear-gradient(135deg, var(--card-color) 0%, var(--hover-color) 100%);
  padding: 10px 16px;
  border-radius: 8px;
  margin-bottom: 12px;
  border: 1px solid var(--border-color);

  .status-indicator-item {
    display: flex;
    align-items: center;
    gap: 6px;

    .indicator-dot {
      width: 8px;
      height: 8px;
      border-radius: 50%;
      
      &.connected {
        background-color: rgb(87, 202, 34);
        box-shadow: 0 0 8px rgba(87, 202, 34, 0.6);
        animation: pulse-green 2s infinite;
      }
      
      &.connecting {
        background-color: rgb(240, 160, 32);
        animation: pulse-yellow 1s infinite;
      }
      
      &.disconnected {
        background-color: rgb(208, 48, 80);
      }
    }

    .indicator-label {
      font-size: 12px;
      color: var(--text-color-2);
    }
  }
}

@keyframes pulse-green {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.6; }
}

@keyframes pulse-yellow {
  0%, 100% { transform: scale(1); }
  50% { transform: scale(1.2); }
}

/* ==================== 方向信号预警面板 ==================== */
// 市场状态分析块
.market-analysis-panel {
  background: var(--robot-bg);
  border: 1px solid var(--robot-border);
  border-radius: 10px;
  padding: 12px;
  margin-bottom: 12px;
  
  .market-analysis-header {
    margin-bottom: 12px;
    padding-bottom: 8px;
    border-bottom: 1px solid rgba(0,0,0,0.05);
  }
  
  .strategy-params-grid {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 8px 12px;
    margin-bottom: 8px;
    
    &.expanded {
      grid-template-columns: repeat(3, 1fr);
      gap: 10px;
    }
    
    .param-item {
      display: flex;
      flex-direction: column;
      gap: 2px;
      
      .param-label {
        font-size: 11px;
        color: var(--robot-text-tertiary);
        font-weight: 500;
      }
      
      .param-value {
        font-size: 13px;
        font-weight: 600;
        color: var(--robot-text-primary);
        font-family: 'JetBrains Mono', 'Consolas', monospace;
        
        &.error {
          color: #ef4444;
        }
        
        &.success {
          color: #10b981;
        }
        
        &.warning {
          color: #f59e0b;
        }
        
        &.primary {
          color: #6366f1;
        }
      }
      
      &.highlight {
        background: rgba(0,0,0,0.02);
        border-radius: 6px;
        padding: 6px 8px;
        
        .param-label {
          color: #6366f1;
        }
        
        .param-value {
          color: #6366f1;
          font-size: 14px;
        }
      }
    }
  }
  
  .strategy-update-time {
    text-align: center;
    padding-top: 6px;
    border-top: 1px solid rgba(0,0,0,0.05);
  }
}

/* 账户统计行样式 */
.account-stats-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 16px;
  background: linear-gradient(to right, rgba(99, 102, 241, 0.05), rgba(16, 185, 129, 0.05));
  border-radius: 8px;
  margin-bottom: 12px;
  border: 1px solid rgba(0, 0, 0, 0.03);

  .stat-item {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 2px;
    flex: 1;

    .stat-label {
      font-size: 11px;
      color: var(--robot-text-tertiary);
      font-weight: 500;
    }

    .stat-value {
      font-family: 'JetBrains Mono', monospace;
      font-weight: 700;
      font-size: 15px;
      color: var(--robot-text-primary);

      &.primary { color: var(--primary-color); }
      &.warning { color: #f59e0b; }
      &.success { color: #10b981; }
      &.error { color: #ef4444; }
    }
  }

  .stat-divider {
    width: 1px;
    height: 24px;
    background-color: rgba(0, 0, 0, 0.06);
    margin: 0 8px;
  }
}

.market-analysis-collapse {
  margin-bottom: 8px;
  
  :deep(.n-collapse-item__header) {
    padding: 8px 12px;
    background: linear-gradient(135deg, #e8f4fd 0%, #f0f9ff 100%);
    border-radius: 8px;
  }
  
  :deep(.n-collapse-item__content-inner) {
    padding: 12px;
    background: rgba(99, 102, 241, 0.03);
    border-radius: 0 0 8px 8px;
  }
}

.signal-alert-panel {
  background: var(--robot-bg);
  border-radius: 10px;
  padding: 12px;
  margin-bottom: 12px;
  border: 1px solid var(--robot-border);
  
  .signal-alert-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 10px;
  }
  
  .signal-badge {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 6px 14px;
    border-radius: 16px;
    font-weight: 600;
    
    &.badge-long {
      background: linear-gradient(135deg, #10b981 0%, #059669 100%);
      color: white;
      box-shadow: 0 2px 6px rgba(16, 185, 129, 0.35);
    }
    
    &.badge-short {
      background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
      color: white;
      box-shadow: 0 2px 6px rgba(239, 68, 68, 0.35);
    }
    
    &.badge-neutral {
      background: linear-gradient(135deg, #6b7280 0%, #4b5563 100%);
      color: white;
    }
    
    .badge-icon { font-size: 16px; }
    .badge-text { font-size: 13px; }
    .badge-strength {
      background: rgba(255, 255, 255, 0.2);
      padding: 1px 6px;
      border-radius: 8px;
      font-size: 11px;
    }
  }
  
  .signal-tags {
    display: flex;
    gap: 5px;
  }
  
  .signal-price-row {
    display: flex;
    gap: 10px;
    margin-bottom: 10px;
    padding: 8px;
    background: var(--robot-panel-bg);
    border-radius: 8px;
    
    .price-current {
      display: flex;
      flex-direction: column;
      align-items: center;
      padding: 6px 12px;
      background: rgba(99, 102, 241, 0.1);
      border-radius: 6px;
      min-width: 90px;
      
      .price-label { font-size: 10px; color: #888; }
      .price-value {
        font-size: 16px;
        font-weight: 700;
        font-family: 'JetBrains Mono', monospace;
        &.up { color: #10b981; }
        &.down { color: #ef4444; }
      }
      .price-change {
        font-size: 10px;
        &.up { color: #10b981; }
        &.down { color: #ef4444; }
      }
    }
    
    .price-window {
      flex: 1;
      display: grid;
      grid-template-columns: repeat(4, 1fr);
      gap: 6px;
      
      .window-item {
        display: flex;
        flex-direction: column;
        align-items: center;
        font-size: 10px;
        color: var(--robot-text-secondary);
        
        strong {
          font-size: 11px;
          font-family: 'JetBrains Mono', monospace;
          color: var(--robot-text-primary);
        }
        
        &.low strong { color: #10b981; }
        &.high strong { color: #ef4444; }
        &.highlight strong { color: #6366f1; }
      }
    }
  }
  
  .signal-triggers {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 8px;
    margin-bottom: 8px;
    
    .trigger {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 6px 10px;
      border-radius: 6px;
      font-size: 11px;
      
      &.long {
        background: rgba(16, 185, 129, 0.12);
        border: 1px solid rgba(16, 185, 129, 0.25);
        .trigger-name { color: #059669; }
        .trigger-price { color: #10b981; font-weight: 700; font-family: 'JetBrains Mono', monospace; }
        .trigger-dist { color: #10b981; font-size: 9px; opacity: 0.8; }
      }
      
      &.short {
        background: rgba(239, 68, 68, 0.12);
        border: 1px solid rgba(239, 68, 68, 0.25);
        .trigger-name { color: #dc2626; }
        .trigger-price { color: #ef4444; font-weight: 700; font-family: 'JetBrains Mono', monospace; }
        .trigger-dist { color: #ef4444; font-size: 9px; opacity: 0.8; }
      }
    }
  }
  
  .signal-chart {
    background: var(--robot-chart-bg);
    backdrop-filter: blur(8px);
    border-radius: 6px;
    padding: 8px;
    margin-bottom: 6px;
    border: 1px solid var(--robot-border);
    
    .chart-svg {
      width: 100%;
      height: 120px;
      display: block;
    }
    
    .chart-line { fill: none; stroke: var(--robot-chart-line); stroke-width: 2; filter: drop-shadow(0 1px 2px rgba(59, 130, 246, 0.3)); }
    .chart-fill { fill: var(--robot-chart-line); opacity: 0.15; }
    .chart-baseline { stroke: var(--robot-chart-line); stroke-width: 1.5; stroke-dasharray: 4,4; opacity: 0.5; }
    .chart-upper { stroke: #ef4444; stroke-width: 1.5; stroke-dasharray: 4,3; opacity: 0.7; }
    .chart-lower { stroke: #10b981; stroke-width: 1.5; stroke-dasharray: 4,3; opacity: 0.7; }
    .point-min { fill: var(--robot-chart-line); stroke: var(--robot-bg); stroke-width: 2; }
    .point-max { fill: #ef4444; stroke: var(--robot-bg); stroke-width: 2; }
    .point-current { fill: #10b981; stroke: var(--robot-bg); stroke-width: 2.5; }
    
    /* 价格标签样式 */
    .price-label {
      font-size: 10px;
      font-weight: normal;
      text-anchor: middle;
    }
    .price-label-min { fill: var(--robot-chart-line); }
    .price-label-max { fill: #ef4444; }
    .price-label-current { 
      fill: #10b981; 
      font-weight: normal;
      text-anchor: start;
    }
    
    /* 图例说明 */
    .chart-legend {
      display: flex;
      justify-content: center;
      gap: 12px;
      margin-top: 4px;
      font-size: 10px;
      
      .legend-item {
        display: flex;
        align-items: center;
        gap: 2px;
      }
      .legend-max { color: #ef4444; }
      .legend-current { color: #10b981; }
      .legend-min { color: #3b82f6; }
    }
  }
  
  .signal-reason {
    padding: 4px 0 0 0;
    background: transparent;
    border: none;
    font-size: 11px;
    font-weight: 400;
    color: var(--robot-text-secondary);
    line-height: 1.5;
    margin-top: 6px;
    text-align: center;
  }
  
  .signal-logs-list {
    max-height: 400px;
    overflow-y: auto;
    overflow-x: hidden;
    padding-right: 4px;
    display: flex;
    flex-direction: column;
    gap: 8px;
    
    &::-webkit-scrollbar {
      width: 6px;
    }
    &::-webkit-scrollbar-thumb {
      background-color: rgba(0, 0, 0, 0.2);
      border-radius: 3px;
      
      &:hover {
        background-color: rgba(0, 0, 0, 0.3);
      }
    }
    &::-webkit-scrollbar-track {
      background-color: rgba(0, 0, 0, 0.05);
    }
    
    .signal-log-card {
      transition: all 0.2s ease;
      border-radius: 8px;
      border: 1px solid rgba(0, 0, 0, 0.06);
      background: var(--card-color);
      
      &:hover {
        box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
        transform: translateY(-1px);
      }
      
      &.long {
        border-left: 3px solid #18a058;
        background: linear-gradient(to right, rgba(24, 160, 88, 0.03), var(--card-color));
      }
      
      &.short {
        border-left: 3px solid #d03050;
        background: linear-gradient(to right, rgba(208, 48, 80, 0.03), var(--card-color));
      }
      
      .signal-log-header {
        display: flex;
        align-items: center;
        margin-bottom: 8px;
      }
      
      .signal-log-details {
        margin-top: 8px;
        
        .detail-item {
          display: flex;
          align-items: center;
          flex-wrap: wrap;
          min-height: 24px;
        }
        
        // 响应式布局：小屏幕时改为2列
        @media (max-width: 768px) {
          :deep(.n-grid) {
            grid-template-columns: repeat(2, 1fr) !important;
          }
        }
      }
      
      .signal-log-reason {
        margin-top: 8px;
        padding: 8px 10px;
        background: rgba(0, 0, 0, 0.02);
        border-radius: 6px;
        border-left: 2px solid rgba(0, 0, 0, 0.1);
        line-height: 1.5;
      }
    }
  }
  
  .execution-logs-list {
    max-height: 400px;
    overflow-y: auto;
    overflow-x: hidden;
    padding-right: 4px;
    display: flex;
    flex-direction: column;
    gap: 8px;
    
    &::-webkit-scrollbar {
      width: 6px;
    }
    &::-webkit-scrollbar-thumb {
      background-color: rgba(0, 0, 0, 0.2);
      border-radius: 3px;
      
      &:hover {
        background-color: rgba(0, 0, 0, 0.3);
      }
    }
    &::-webkit-scrollbar-track {
      background-color: rgba(0, 0, 0, 0.05);
    }
    
    .execution-log-card {
      transition: all 0.2s ease;
      border-radius: 8px;
      border: 1px solid rgba(0, 0, 0, 0.06);
      background: var(--card-color);
      
      &:hover {
        box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
        transform: translateY(-1px);
      }
      
      &.success {
        border-left: 3px solid #18a058;
        background: linear-gradient(to right, rgba(24, 160, 88, 0.03), var(--card-color));
      }
      
      &.failed {
        border-left: 3px solid #d03050;
        background: linear-gradient(to right, rgba(208, 48, 80, 0.03), var(--card-color));
      }
      
      &.pending {
        border-left: 3px solid #f0a020;
        background: linear-gradient(to right, rgba(240, 160, 32, 0.03), var(--card-color));
      }
      
      .execution-log-header {
        display: flex;
        align-items: center;
        margin-bottom: 8px;
      }
      
      .execution-log-message {
        padding: 8px 10px;
        background: rgba(0, 0, 0, 0.02);
        border-radius: 6px;
        margin-top: 8px;
        border-left: 2px solid rgba(0, 0, 0, 0.1);
        line-height: 1.5;
      }
      
      .execution-log-details {
        margin-top: 8px;
        
        .detail-item {
          display: flex;
          align-items: center;
          flex-wrap: wrap;
          min-height: 24px;
        }
        
        // 响应式布局：小屏幕时改为2列
        @media (max-width: 768px) {
          :deep(.n-grid) {
            grid-template-columns: repeat(2, 1fr) !important;
          }
        }
      }
    }
  }
}

.analysis-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
  background-color: var(--card-color);
  border-radius: 8px;
  margin-bottom: 12px;
}

/* 账户+策略参数整合面板 */
.account-strategy-panel {
  background: var(--robot-bg);
  border-radius: 8px;
  margin-bottom: 12px;
  padding: 12px;
  border: 1px solid var(--robot-border);
  
  .stat-value {
      font-family: 'JetBrains Mono', monospace;
      font-weight: 500;
      font-size: 14px;
      
      &.primary { color: var(--primary-color); font-weight: 600; }
      &.warning { color: #f59e0b; font-weight: 600; }
  }

  .param-box {
    text-align: center;
    padding: 6px 4px;
    background: var(--robot-panel-bg);
    border-radius: 6px;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    transition: all 0.2s;
    
    &:hover {
        background: var(--robot-hover-bg);
    }
    
    .label {
      font-size: 10px;
      color: var(--robot-text-tertiary);
      margin-bottom: 2px;
      transform: scale(0.9);
    }
    
    .value {
      font-size: 12px;
      font-weight: 600;
      font-family: 'JetBrains Mono', monospace;
      color: var(--robot-text-primary);
    }
    
    /* Variations */
    &.highlight {
        background: rgba(99, 102, 241, 0.08);
        .value { color: #6366f1; }
    }
    &.error {
        background: rgba(239, 68, 68, 0.08);
        .value { color: #ef4444; }
    }
    &.success {
        background: rgba(16, 185, 129, 0.08);
        .value { color: #10b981; }
    }
    &.warning {
        background: rgba(245, 158, 11, 0.08);
        .value { color: #f59e0b; }
    }
  }
  
  .switch-item {
      cursor: pointer;
      transition: all 0.2s;
      &:hover { 
          transform: translateY(-1px);
      }
  }
}

/* 策略参数文本列表样式 */
.strategy-params-container {
  display: flex;
  flex-wrap: wrap;
  gap: 6px 12px;
  padding: 0 4px;
  margin-bottom: 8px;

  .param-text-item {
    display: flex;
    align-items: center;
    gap: 4px;
    font-size: 12px;
    line-height: 1.5;

    .label {
      color: var(--robot-text-tertiary);
      font-size: 11px;
      font-weight: 400;
    }

    .value {
      font-family: 'JetBrains Mono', monospace;
      font-weight: 500;
      font-size: 12px;
      color: var(--robot-text-primary);

      &.highlight { color: #6366f1; font-weight: 600; }
      &.error { color: #ef4444; font-weight: 600; }
      &.success { color: #10b981; font-weight: 600; }
      &.warning { color: #f59e0b; font-weight: 600; }
    }
  }
}

/* ==================== 统一字体规范 ==================== */
:deep(.robot-card.running) {
  /* 标题文字 */
  --font-title: 500 12px/1.5 -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
  /* 正文文字 */
  --font-body: 400 12px/1.5 -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
  /* 辅助说明文字 */
  --font-secondary: 400 11px/1.4 -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
  /* 数值文字（价格、数量等） */
  --font-mono: 600 12px/1.4 'JetBrains Mono', 'Courier New', monospace;
  /* 小标签文字 */
  --font-small: 400 10px/1.3 -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
  /* 微小注释 */
  --font-tiny: 400 10px/1.2 -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
}

/* 市场状态与风险偏好映射（机器人列表页） */
.signal-prelogs-panels {
  margin-top: 6px;
}

.mapping-item-card {
  background: rgba(99, 102, 241, 0.035);
  border: 1px solid rgba(99, 102, 241, 0.12);
  border-radius: 8px;
  padding: 10px 10px 12px;
}

.mapping-header {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 6px;
}

.mapping-header .market-name {
  font-size: 12px;
  color: var(--robot-text-primary);
  font-weight: 600;
}

.mapping-arrow {
  text-align: center;
  color: rgba(0, 0, 0, 0.35);
  font-size: 12px;
  margin: 6px 0;
}

/* 风险偏好标签简洁样式 */
.risk-mapping-tag {
  font-weight: 400;
  font-size: 11px;
  background-color: transparent !important;
  color: #888888 !important;
  opacity: 1;
}

.account-section {
  background: linear-gradient(135deg, rgba(32, 128, 240, 0.08) 0%, rgba(24, 160, 88, 0.08) 100%);
  padding: 12px 16px;
  border-radius: 8px;
  margin-bottom: 12px;
  border: 1px solid var(--border-color);

  .account-item {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 4px;
  }
}

.config-section {
  background-color: var(--card-color);
  padding: 8px 12px;
  border-radius: 6px;
  margin-bottom: 12px;
  display: flex;
  align-items: center;
  
  .quick-switch {
    cursor: pointer;
    transition: all 0.2s ease;
    
    &:hover {
      transform: scale(1.05);
    }
    
    &:active {
      transform: scale(0.98);
    }
    
    .clickable-tag {
      cursor: pointer;
      transition: all 0.2s ease;
      
      &:hover {
        box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
      }
    }
  }
  
  :deep(.n-tag) {
    &:not(.clickable-tag) {
      cursor: default;
    }
  }
}

.ticker-section {
  background-color: var(--card-color);
  padding: 12px;
  border-radius: 8px;
  margin-bottom: 12px;

  .ticker-item {
    text-align: center;
  }

  .price {
    font-size: 18px;
    font-weight: bold;

    &.up { color: var(--success-color); }
    &.down { color: var(--error-color); }
  }
}

.position-section {
  background-color: var(--card-color);
  padding: 12px;
  border-radius: 8px;

  .position-item {
    padding: 10px 12px;
    background-color: var(--body-color);
    border-radius: 6px;
    margin-bottom: 8px;

    &:last-child {
      margin-bottom: 0;
    }
  }
}

/* 实时持仓订单列表样式 - 参考 toogo 项目 */
.positions-section {
  margin-bottom: 12px;
  border: 1px solid var(--robot-border);
  border-radius: 8px;
  background: var(--robot-bg);
  overflow: hidden;
}

.positions-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: linear-gradient(135deg, rgba(99, 102, 241, 0.08) 0%, rgba(99, 102, 241, 0.03) 100%);
  border-bottom: 2px solid rgba(99, 102, 241, 0.15);
  border-radius: 8px 8px 0 0;
  margin-bottom: 0;
}

.positions-title {
  font-size: 14px;
  font-weight: 700;
  margin: 0;
  color: var(--text-color-1);
  line-height: 1.2;
  letter-spacing: 0.3px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.positions-title::before {
  content: '📊';
  font-size: 16px;
}

.positions-count {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 28px;
  height: 24px;
  padding: 0 10px;
  background: linear-gradient(135deg, rgba(99, 102, 241, 0.9) 0%, rgba(99, 102, 241, 0.7) 100%);
  color: white;
  border-radius: 12px;
  font-size: 11px;
  font-weight: 700;
  box-shadow: 0 2px 4px rgba(99, 102, 241, 0.3);
  transition: all 0.2s ease;
}

.positions-count:hover {
  transform: scale(1.05);
  box-shadow: 0 4px 8px rgba(99, 102, 241, 0.4);
}

.tiny-switch {
  transform: scale(0.75);
  opacity: 0.7;
}

.positions-table-wrapper {
  border: 1px solid rgba(0, 0, 0, 0.06);
  border-radius: 0 0 8px 8px;
  background: var(--card-color);
  overflow-x: auto;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
}

.positions-table-wrapper::-webkit-scrollbar {
  height: 10px;
}

.positions-table-wrapper::-webkit-scrollbar-track {
  background: rgba(0, 0, 0, 0.05);
  border-radius: 5px;
}

.positions-table-wrapper::-webkit-scrollbar-thumb {
  background: rgba(0, 0, 0, 0.2);
  border-radius: 5px;
  
  &:hover {
    background: rgba(0, 0, 0, 0.3);
  }
}

.positions-table {
  width: 100%;
  border-collapse: separate;
  border-spacing: 0;
  font-size: 11px;
  table-layout: fixed;
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.positions-table thead {
  background: linear-gradient(135deg, rgba(99, 102, 241, 0.08) 0%, rgba(99, 102, 241, 0.03) 100%);
  backdrop-filter: blur(10px);
}

.positions-table th {
  padding: 10px 8px;
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  color: var(--text-color-2);
  white-space: nowrap;
  text-align: center;
  border-bottom: 2px solid rgba(99, 102, 241, 0.15);
  line-height: 1.3;
  position: sticky;
  top: 0;
  z-index: 10;
}

.positions-table td {
  padding: 10px 8px;
  vertical-align: middle;
  border-bottom: 1px solid rgba(0, 0, 0, 0.04);
  white-space: nowrap;
  color: var(--text-color-1);
  text-align: center;
  overflow: hidden;
  text-overflow: ellipsis;
  transition: all 0.2s ease;
}

.positions-table tr:last-child td {
  border-bottom: none;
}

.position-row {
  transition: all 0.2s ease;
  position: relative;
}

.position-row.row-even {
  background: var(--card-color);
}

.position-row.row-odd {
  background: rgba(255, 255, 255, 0.5);
}

.position-row:hover {
  background: linear-gradient(135deg, rgba(99, 102, 241, 0.1) 0%, rgba(99, 102, 241, 0.05) 100%) !important;
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(99, 102, 241, 0.15);
}

/* 列宽定义 - 紧凑模式 */
.col-info { width: 90px; min-width: 90px; }
.col-quantity { width: 95px; min-width: 95px; }
.col-price { width: 60px; min-width: 60px; }
.col-pl { width: 80px; min-width: 80px; }
.col-monitor { width: 208px; min-width: 208px; }
.col-action { 
  width: 60px; 
  min-width: 60px; 
  text-align: center;
  padding: 8px 4px !important;
}

/* 平仓按钮样式 - 黄色系 */
.close-position-btn {
  background: linear-gradient(135deg, #fbbf24 0%, #f59e0b 100%) !important;
  border: none !important;
  color: #ffffff !important;
  font-weight: 600 !important;
  border-radius: 6px !important;
  box-shadow: 0 2px 4px rgba(245, 158, 11, 0.3) !important;
  transition: all 0.3s ease !important;
  padding: 6px 14px !important;
  min-width: 52px !important;
  height: auto !important;
  line-height: 1.4 !important;
}

.close-position-btn:hover {
  background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%) !important;
  transform: translateY(-1px) !important;
  box-shadow: 0 4px 8px rgba(245, 158, 11, 0.4) !important;
}

.close-position-btn:active {
  transform: translateY(0) !important;
  box-shadow: 0 2px 4px rgba(245, 158, 11, 0.3) !important;
}

/* 交易信息列样式 */
.info-cell {
  display: flex;
  flex-direction: column;
  gap: 1px;
  align-items: center;
}
.info-row-top {
  display: flex;
  align-items: center;
  gap: 2px;
  justify-content: center;
}
.info-row-second {
  display: flex;
  align-items: center;
  gap: 2px;
  justify-content: center;
}
.info-row-middle {
  display: flex;
  align-items: center;
  gap: 4px;
  justify-content: center;
}
.info-row-order {
  margin-top: 4px;
  padding-top: 4px;
  border-top: 1px solid rgba(0, 0, 0, 0.05);
  font-size: 9px;
  line-height: 1.4;
  width: 100%;
  text-align: center;
}
.order-info-item {
  display: flex;
  align-items: center;
  margin-bottom: 2px;
  justify-content: center;
}
.order-info-item:last-child {
  margin-bottom: 0;
}
.order-label {
  color: var(--text-color-3);
  font-weight: 500;
  min-width: 50px;
  font-size: 9px;
}
.order-value {
  color: var(--text-color-2);
  font-family: monospace;
  font-size: 9px;
  word-break: break-all;
  text-align: center;
}
.market-state-text {
  font-size: 9px;
  color: var(--text-color-2);
  font-weight: 400;
}
.symbol-text {
  font-weight: 700;
  color: var(--text-color-1);
  font-size: 11px;
  letter-spacing: 0.3px;
  font-family: 'JetBrains Mono', monospace;
}
.side-tag-mini {
  font-size: 9px;
  padding: 2px 6px;
  border-radius: 4px;
  font-weight: 700;
  letter-spacing: 0.5px;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
  transition: all 0.2s ease;
}
.side-tag-mini.long { 
  color: #10b981; 
  background: linear-gradient(135deg, rgba(16, 185, 129, 0.15) 0%, rgba(16, 185, 129, 0.1) 100%);
  border: 1px solid rgba(16, 185, 129, 0.3);
}
.side-tag-mini.short { 
  color: #ef4444; 
  background: linear-gradient(135deg, rgba(239, 68, 68, 0.15) 0%, rgba(239, 68, 68, 0.1) 100%);
  border: 1px solid rgba(239, 68, 68, 0.3);
}

/* 保证金模式标签 */
.margin-mode-tag {
  font-size: 9px;
  padding: 2px 6px;
  border-radius: 4px;
  font-weight: 600;
  margin-left: 4px;
  letter-spacing: 0.3px;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
  transition: all 0.2s ease;
}

/* 杠杆标签（来自交易所持仓数据 pos.leverage） */
.leverage-tag {
  font-size: 9px;
  padding: 2px 6px;
  border-radius: 4px;
  font-weight: 600;
  letter-spacing: 0.3px;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
  transition: all 0.2s ease;
  color: #3b82f6;
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.15) 0%, rgba(59, 130, 246, 0.1) 100%);
  border: 1px solid rgba(59, 130, 246, 0.28);
}
.margin-mode-tag.isolated { 
  color: #f59e0b; 
  background: linear-gradient(135deg, rgba(245, 158, 11, 0.15) 0%, rgba(245, 158, 11, 0.1) 100%);
  border: 1px solid rgba(245, 158, 11, 0.3);
}
.margin-mode-tag.crossed { 
  color: #3b82f6; 
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.15) 0%, rgba(59, 130, 246, 0.1) 100%);
  border: 1px solid rgba(59, 130, 246, 0.3);
}

.info-time {
  font-size: 10px;
  color: var(--text-color-3);
  font-family: 'JetBrains Mono', monospace;
}

/* 数字显示样式 */
.qty-cell {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
}
.margin-in-qty {
  font-family: 'JetBrains Mono', monospace;
  font-weight: 700;
  color: var(--text-color-1);
  font-size: 11px;
}
.quantity-value {
  font-family: 'JetBrains Mono', monospace;
  font-weight: 600;
  color: var(--text-color-1);
  font-size: 11px;
}

.price-value {
  font-family: 'JetBrains Mono', monospace;
  font-weight: 500;
  color: var(--text-color-1);
  font-size: 11px;
}

.price-value.market-price {
  color: var(--primary-color);
  font-weight: 600;
}

/* 保证金信息行样式 */
.info-row-margin {
  display: flex;
  align-items: center;
  gap: 4px;
  justify-content: center;
  margin-top: 4px;
  padding-top: 4px;
  border-top: 1px solid rgba(0, 0, 0, 0.05);
  font-size: 9px;
}

.margin-label {
  color: var(--text-color-3);
  font-weight: 500;
}

.margin-info {
  color: var(--text-color-1);
  font-family: 'JetBrains Mono', monospace;
  font-weight: 600;
  font-size: 10px;
}

/* 盈亏显示优化 */
.pnl-display {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  border-radius: 6px;
  font-weight: 600;
  transition: all 0.3s ease;
}

.pnl-display.profit {
  background: linear-gradient(135deg, rgba(16, 185, 129, 0.1) 0%, rgba(16, 185, 129, 0.05) 100%);
  color: #10b981;
  border: 1px solid rgba(16, 185, 129, 0.2);
}

.pnl-display.loss {
  background: linear-gradient(135deg, rgba(239, 68, 68, 0.1) 0%, rgba(239, 68, 68, 0.05) 100%);
  color: #ef4444;
  border: 1px solid rgba(239, 68, 68, 0.2);
}

.pnl-icon {
  font-size: 14px;
  filter: drop-shadow(0 1px 2px rgba(0, 0, 0, 0.1));
}

.pnl-value {
  font-family: 'JetBrains Mono', monospace;
  font-size: 11px;
  font-weight: 700;
}

/* 监控列样式 */
.monitor-cell {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 6px 8px;
  background: rgba(0, 0, 0, 0.02);
  border-radius: 6px;
  border: 1px solid rgba(0, 0, 0, 0.04);
}

.monitor-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 2px 0;
}

.monitor-label {
  width: 45px;
  min-width: 45px;
  max-width: 45px;
  font-size: 10px;
  color: var(--text-color-2);
  white-space: nowrap;
  cursor: help;
  transition: color 0.2s;
  flex-shrink: 0;
  font-weight: 500;
}

.monitor-label:hover {
  color: var(--primary-color);
}

.progress-bar-container {
  flex: 1;
  height: 6px;
  min-width: 40px;
  max-width: 100px;
  background: rgba(148, 163, 184, 0.2);
  border-radius: 3px;
  overflow: hidden;
  box-shadow: inset 0 1px 2px rgba(0, 0, 0, 0.1);
}

.progress-bar {
  height: 100%;
  border-radius: 3px;
  transition: all 0.6s cubic-bezier(0.4, 0, 0.2, 1);
  box-shadow: 0 0 6px rgba(0, 0, 0, 0.15);
  position: relative;
}

.progress-bar::after {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.3), transparent);
  animation: shimmer 2s infinite;
}

@keyframes shimmer {
  0% { transform: translateX(-100%); }
  100% { transform: translateX(100%); }
}

.progress-bar-danger {
  background: linear-gradient(90deg, #f59e0b, #ef4444, #dc2626);
  box-shadow: 0 0 4px rgba(239, 68, 68, 0.4);
}

.progress-bar-success {
  background: linear-gradient(90deg, #22c55e, #16a34a);
  box-shadow: 0 0 4px rgba(34, 197, 94, 0.3);
}

.progress-bar-warning {
  background: linear-gradient(90deg, #fbbf24, #f59e0b, #d97706);
  box-shadow: 0 0 4px rgba(245, 158, 11, 0.3);
}

.monitor-value {
  width: auto;
  min-width: 55px;
  max-width: none;
  font-size: 10px;
  text-align: right;
  color: var(--text-color-1);
  flex-shrink: 0;
  white-space: nowrap;
  padding-left: 4px;
  font-family: 'JetBrains Mono', monospace;
  font-weight: 500;
}

.monitor-value.text-danger {
  color: #ef4444;
  font-weight: bold;
  animation: pulse-danger 1.5s ease-in-out infinite;
}

.monitor-value.text-warning {
  color: #f59e0b;
  font-weight: bold;
}

.monitor-value.text-success {
  color: #22c55e;
  font-weight: bold;
}

.monitor-value.text-disabled {
  color: #9ca3af;
  font-size: 8px;
}

/* 开关监控项样式 */
.monitor-item.monitor-switch {
  padding: 2px 0;
}

.monitor-item.monitor-switch .monitor-label {
  font-size: 9px;
  color: var(--text-color-2);
}

.monitor-item.monitor-switch .monitor-value {
  display: flex;
  align-items: center;
  justify-content: flex-end;
}

/* 禁用状态的监控项 */
.monitor-item.monitor-disabled {
  opacity: 0.5;
}

.monitor-item.monitor-disabled .monitor-label {
  color: #9ca3af;
}

.progress-bar-disabled {
  background: #9ca3af !important;
}

@keyframes pulse-danger {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.6;
  }
}

.monitor-max-profit {
  border-top: 1px solid rgba(148, 163, 184, 0.2);
  padding-top: 4px;
  margin-top: 2px;
}

/* 方向标签 */
.side-tag {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
}

.side-tag.long {
  background: rgba(16, 185, 129, 0.15);
  color: #10b981;
}

.side-tag.short {
  background: rgba(239, 68, 68, 0.15);
  color: #ef4444;
}

/* 盈亏显示 */
.pnl-display {
  display: flex;
  align-items: center;
  gap: 2px;
  padding: 2px 4px;
  border-radius: 4px;
  font-weight: 600;
}

.pnl-display.profit {
  background: linear-gradient(135deg, rgba(16, 185, 129, 0.15), rgba(16, 185, 129, 0.05));
  border: 1px solid rgba(16, 185, 129, 0.3);
}

.pnl-display.loss {
  background: linear-gradient(135deg, rgba(239, 68, 68, 0.15), rgba(239, 68, 68, 0.05));
  border: 1px solid rgba(239, 68, 68, 0.3);
}

.pnl-display .pnl-icon {
  font-size: 10px;
}

.pnl-display .pnl-value {
  font-size: 10px;
  font-weight: 700;
}

.pnl-display.profit .pnl-value {
  color: #10b981;
}

.pnl-display.loss .pnl-value {
  color: #ef4444;
}

.pnl-display .pnl-unit {
  font-size: 8px;
  opacity: 0.7;
}

.pnl-value {
  font-weight: 600;
}

.pnl-value.profit {
  color: #10b981;
}

.pnl-value.loss {
  color: #ef4444;
}

/* 空状态 */
.positions-table-wrapper .empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  background: linear-gradient(135deg, rgba(99, 102, 241, 0.03) 0%, rgba(99, 102, 241, 0.01) 100%);
  border-radius: 8px;
  border: none;
  margin: 12px 0;
  color: var(--text-color-3);
}

.empty-icon {
  font-size: 32px;
  margin-bottom: 10px;
  opacity: 0.5;
  animation: float 3s ease-in-out infinite;
  filter: drop-shadow(0 2px 4px rgba(0, 0, 0, 0.1));
}

.empty-text {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-color-2);
  margin-bottom: 8px;
}

.empty-hint {
  font-size: 12px;
  color: var(--text-color-3);
  font-style: italic;
}

@keyframes float {
  0%, 100% { transform: translateY(0px); }
  50% { transform: translateY(-10px); }
}

/* 算法逻辑说明面板样式 */
.algorithm-collapse {
  background: var(--robot-bg);
  border-radius: 8px;
  
  .algo-section {
    background: rgba(99, 102, 241, 0.05);
    border-radius: 8px;
    padding: 8px 10px;
    margin-bottom: 8px;
    border: 1px solid rgba(99, 102, 241, 0.15);
    
    &:last-child { margin-bottom: 0; }
    
    .algo-title {
      font-size: 12px;
      font-weight: 600;
      color: var(--robot-text-primary);
      margin-bottom: 4px;
      line-height: 1.3;
    }
    
    .algo-desc {
      font-size: 11px;
      color: var(--robot-text-secondary);
      margin-bottom: 6px;
      line-height: 1.3;
    }
    
    .algo-warning {
      font-size: 10px;
      color: #f59e0b;
      text-align: center;
      margin-top: 6px;
      padding: 2px 6px;
      background: rgba(245, 158, 11, 0.1);
      border-radius: 4px;
      line-height: 1.3;
    }
    
    .algo-rules {
      display: grid;
      grid-template-columns: 1fr 1fr;
      gap: 8px;
      
      .rule-item {
        display: flex;
        align-items: flex-start;
        gap: 6px;
        padding: 6px;
        border-radius: 6px;
        
        &.long {
          background: linear-gradient(135deg, rgba(16, 185, 129, 0.08) 0%, rgba(16, 185, 129, 0.02) 100%);
          border: 1px solid rgba(16, 185, 129, 0.2);
        }
        
        &.short {
          background: linear-gradient(135deg, rgba(239, 68, 68, 0.08) 0%, rgba(239, 68, 68, 0.02) 100%);
          border: 1px solid rgba(239, 68, 68, 0.2);
        }
        
        .rule-icon {
          font-size: 18px;
        }
        
        .rule-content {
          display: flex;
          flex-direction: column;
          gap: 2px;
          
          .rule-name {
            font-size: 12px;
            font-weight: 600;
          }
          
          .rule-formula {
            font-size: 11px;
            color: var(--robot-text-primary);
            font-family: 'JetBrains Mono', monospace;
            background: var(--robot-panel-bg);
            padding: 2px 6px;
            border-radius: 3px;
          }
          
          .rule-note {
            font-size: 10px;
            color: var(--robot-text-tertiary);
          }
        }
      }
    }
    
    .algo-list {
      display: flex;
      flex-direction: column;
      gap: 4px;
      
      .algo-item {
        font-size: 11px;
        color: var(--robot-text-secondary);
        padding: 4px 6px;
        background: var(--robot-panel-bg);
        border-radius: 4px;
        line-height: 1.3;
        
        strong {
          color: #6366f1;
        }
        
        &.formula {
          display: flex;
          flex-wrap: wrap;
          align-items: center;
          gap: 4px;
          
          .item-label {
            font-weight: 500;
            color: #333;
            min-width: 70px;
          }
          
          .item-formula {
            font-family: 'JetBrains Mono', monospace;
            font-size: 10px;
            background: rgba(99, 102, 241, 0.08);
            padding: 2px 6px;
            border-radius: 3px;
            color: #6366f1;
          }
          
          .item-action {
            font-weight: 500;
            color: #059669;
          }
          
          .item-note {
            font-size: 10px;
            color: #888;
            width: 100%;
            padding-left: 70px;
            margin-top: 2px;
          }
        }
      }
    }
  }
}

.selected-strategy {
  border-color: var(--primary-color) !important;
}

.strategy-tag {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  padding: 8px 12px;
  background-color: var(--card-color);
  border-radius: 6px;
}
</style>


